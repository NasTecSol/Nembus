package grpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/syncpb"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestGenerateAndHashSyncToken(t *testing.T) {
	raw1, hash1, err := usecase.GenerateSyncToken()
	if err != nil {
		t.Fatalf("GenerateSyncToken failed: %v", err)
	}
	if len(raw1) != 64 || len(hash1) != 64 {
		t.Fatalf("expected 64-char hex raw token and SHA-256 hash, got len(raw)=%d len(hash)=%d", len(raw1), len(hash1))
	}
	if usecase.HashSyncToken(raw1) != hash1 {
		t.Fatalf("HashSyncToken(%q) = %q, want %q", raw1, usecase.HashSyncToken(raw1), hash1)
	}
	raw2, hash2, err := usecase.GenerateSyncToken()
	if err != nil {
		t.Fatalf("second GenerateSyncToken failed: %v", err)
	}
	if raw1 == raw2 || hash1 == hash2 {
		t.Fatal("generated identical random sync tokens")
	}
}

func TestChangeNotifyHubSubscribeAndBroadcast(t *testing.T) {
	hub := NewChangeNotifyHub(nil)
	ch1, unsub1 := hub.Subscribe("tenant-a")
	defer unsub1()
	ch2, unsub2 := hub.Subscribe("tenant-a")
	defer unsub2()
	chOther, unsubOther := hub.Subscribe("tenant-b")
	defer unsubOther()

	hub.BroadcastSeq("tenant-a", 42)

	select {
	case seq := <-ch1:
		if seq != 42 {
			t.Fatalf("ch1 got seq=%d, want 42", seq)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for ch1 notification")
	}

	select {
	case seq := <-ch2:
		if seq != 42 {
			t.Fatalf("ch2 got seq=%d, want 42", seq)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for ch2 notification")
	}

	select {
	case seq := <-chOther:
		t.Fatalf("unexpected notification on tenant-b channel: %d", seq)
	default:
	}

	// Overflow coalescing: burst 50 sequence notifications without reading; channel must not block
	for seq := int64(100); seq <= 150; seq++ {
		hub.BroadcastSeq("tenant-a", seq)
	}
	var lastRead int64
	for {
		select {
		case s := <-ch1:
			lastRead = s
		default:
			goto drained
		}
	}
drained:
	if lastRead != 150 {
		t.Fatalf("expected coalesced channel to retain latest seq 150, got %d", lastRead)
	}
}

func TestProcessPushBatchValidation(t *testing.T) {
	srv := NewSyncServer(nil, nil)
	ctx := context.Background()

	// 1. Out-of-order seq_in_batch is rejected with FailedSeq
	payload := []byte(`{"id":1}`)
	sha := computePayloadSHA256(payload)
	ack := srv.processPushBatch(ctx, []*syncpb.SyncEvent{
		{
			Id:          10,
			OutboxUuid:  "00000000-0000-0000-0000-000000000010",
			BatchId:     "batch-1",
			BatchSize:   2,
			SeqInBatch:  1,
			EntityType:  "pos_transactions",
			EntityId:    1,
			PayloadJson: payload,
			Sha256:      sha,
			TenantSlug:  "demo",
		},
		{
			Id:          11,
			OutboxUuid:  "00000000-0000-0000-0000-000000000011",
			BatchId:     "batch-1",
			BatchSize:   2,
			SeqInBatch:  3, // Invalid seq_in_batch (expected 2)
			EntityType:  "pos_transaction_lines",
			EntityId:    1,
			PayloadJson: payload,
			Sha256:      sha,
			TenantSlug:  "demo",
		},
	})
	if ack.Success || ack.FailedSeq != 2 || !strings.Contains(ack.ErrorMessage, "seq_in_batch") {
		t.Fatalf("expected FailedSeq=2 seq_in_batch error, got success=%v failed_seq=%d msg=%q",
			ack.Success, ack.FailedSeq, ack.ErrorMessage)
	}

	// 2. SHA-256 mismatch is rejected at the offending event index
	ack2 := srv.processPushBatch(ctx, []*syncpb.SyncEvent{
		{
			Id:          10,
			BatchId:     "batch-2",
			BatchSize:   1,
			SeqInBatch:  1,
			EntityType:  "pos_transactions",
			EntityId:    1,
			PayloadJson: payload,
			Sha256:      "deadbeef",
			TenantSlug:  "demo",
		},
	})
	if ack2.Success || ack2.FailedSeq != 1 || !strings.Contains(ack2.ErrorMessage, "SHA-256") {
		t.Fatalf("expected SHA-256 mismatch at FailedSeq=1, got %+v", ack2)
	}
}

func TestStreamSessionNotificationsAndTokenAuth(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	syncSrv := NewSyncServer(nil, nil)
	syncSrv.SetRequireTokenAuth(true)
	syncpb.RegisterSyncServiceServer(grpcServer, syncSrv)
	go func() { _ = grpcServer.Serve(lis) }()
	defer grpcServer.Stop()
	defer lis.Close()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
	)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer conn.Close()
	client := syncpb.NewSyncServiceClient(conn)

	// 1. Unauthenticated StreamChanges call without token must return Unauthenticated
	stream, err := client.StreamChanges(context.Background(), &syncpb.StreamChangesRequest{
		TenantSlug: "demo",
		StoreId:    1,
		FromSeq:    0,
	})
	if err == nil {
		_, err = stream.Recv()
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated when requireTokenAuth=true and token is missing, got %v", err)
	}

	// 2. Disable token auth requirement and verify StreamSession delivers ChangesAvailableNotification
	syncSrv.SetRequireTokenAuth(false)
	sessionCtx, cancelSession := context.WithTimeout(
		metadata.AppendToOutgoingContext(context.Background(), "x-tenant-id", "demo"),
		5*time.Second,
	)
	defer cancelSession()

	sessStream, err := client.StreamSession(sessionCtx)
	if err != nil {
		t.Fatalf("StreamSession failed: %v", err)
	}

	// Give server goroutine a moment to register subscription, then broadcast seq 77
	time.Sleep(30 * time.Millisecond)
	syncSrv.ChangeHub().BroadcastSeq("demo", 77)

	msg, err := sessStream.Recv()
	if err != nil {
		t.Fatalf("sessStream.Recv failed: %v", err)
	}
	if msg.GetChangesAvailable() == nil || msg.GetChangesAvailable().LatestSeq != 77 {
		t.Fatalf("expected ChangesAvailable(77), got %+v", msg)
	}
}

func TestBuildPullEntityRankCaseSQL(t *testing.T) {
	sqlExpr := buildPullEntityRankCaseSQL()
	if !strings.HasPrefix(sqlExpr, "CASE entity_type ") || !strings.HasSuffix(sqlExpr, "END") {
		t.Fatalf("unexpected CASE expression: %s", sqlExpr)
	}
	for _, entity := range syncpb.PullEntityTypes() {
		if !strings.Contains(sqlExpr, "'"+entity+"'") {
			t.Errorf("CASE SQL missing entity %s", entity)
		}
	}
}

func openMultiTerminalTestPool(t *testing.T, prefix string) (*pgxpool.Pool, func()) {
	t.Helper()
	dsn := os.Getenv("NEMBUS_SYNC_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set NEMBUS_SYNC_TEST_DATABASE_URL to run PostgreSQL multi-terminal integration tests")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		_, _ = admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
		admin.Close()
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		_, _ = admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
		admin.Close()
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `
		CREATE TABLE pos_terminals (
			id integer PRIMARY KEY,
			store_id integer NOT NULL,
			sync_token_hash text,
			is_active boolean DEFAULT true
		);
		CREATE TABLE cashier_sessions (
			id uuid PRIMARY KEY,
			pos_terminal_id integer NOT NULL,
			session_number text UNIQUE,
			status text NOT NULL DEFAULT 'open'
		);
		CREATE TABLE pos_transactions (
			id uuid PRIMARY KEY,
			store_id integer NOT NULL,
			pos_terminal_id integer,
			cashier_session_id uuid REFERENCES cashier_sessions(id),
			transaction_number text UNIQUE,
			total_amount numeric NOT NULL,
			status text NOT NULL DEFAULT 'completed'
		);
		CREATE TABLE pos_transaction_lines (
			id uuid PRIMARY KEY,
			transaction_id uuid NOT NULL REFERENCES pos_transactions(id),
			quantity numeric NOT NULL,
			line_total numeric NOT NULL
		);
		CREATE TABLE pos_payments (
			id uuid PRIMARY KEY,
			transaction_id uuid NOT NULL REFERENCES pos_transactions(id),
			payment_method text NOT NULL,
			amount numeric NOT NULL
		);
		CREATE TABLE sync_watermarks (
			entity_type text NOT NULL,
			store_id integer NOT NULL,
			last_sync_at timestamptz NOT NULL DEFAULT now(),
			metadata jsonb,
			PRIMARY KEY (entity_type, store_id)
		);
	`)
	if err != nil {
		pool.Close()
		_, _ = admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
		admin.Close()
		t.Fatal(err)
	}
	cleanup := func() {
		pool.Close()
		_, _ = admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
		admin.Close()
	}
	return pool, cleanup
}

func TestMultiTerminalConcurrentOfflineSync(t *testing.T) {
	pool, cleanup := openMultiTerminalTestPool(t, "multi_term_conc")
	defer cleanup()
	ctx := context.Background()

	srv := NewSyncServer(nil, nil)
	srv.SetPoolResolver(func(ctx context.Context, slug string) (*pgxpool.Pool, error) {
		return pool, nil
	})

	type termFixture struct {
		terminalID int32
		sessionID  string
		txID       string
		lineID     string
		payID      string
		txNum      string
		batchID    string
		outboxPfx  string
	}
	fixtures := []termFixture{
		{
			terminalID: 1,
			sessionID:  "11111111-1111-4111-8111-111111111101",
			txID:       "11111111-1111-4111-8111-111111111102",
			lineID:     "11111111-1111-4111-8111-111111111103",
			payID:      "11111111-1111-4111-8111-111111111104",
			txNum:      "POS-20261121-150000-T1-A001",
			batchID:    "batch-t1",
			outboxPfx:  "11111111-1111-4111-8111-11111111120",
		},
		{
			terminalID: 2,
			sessionID:  "22222222-2222-4222-8222-222222222201",
			txID:       "22222222-2222-4222-8222-222222222202",
			lineID:     "22222222-2222-4222-8222-222222222203",
			payID:      "22222222-2222-4222-8222-222222222204",
			txNum:      "POS-20261121-150000-T2-B002",
			batchID:    "batch-t2",
			outboxPfx:  "22222222-2222-4222-8222-22222222220",
		},
	}

	var wg sync.WaitGroup
	acks := make([]*syncpb.SyncAck, len(fixtures))
	for i, fx := range fixtures {
		wg.Add(1)
		go func(idx int, f termFixture) {
			defer wg.Done()
			sessPayload := []byte(fmt.Sprintf(`{"id":%q,"pos_terminal_id":%d,"session_number":"SES-T%d-01","status":"open"}`, f.sessionID, f.terminalID, f.terminalID))
			txPayload := []byte(fmt.Sprintf(`{"id":%q,"store_id":1,"pos_terminal_id":%d,"cashier_session_id":%q,"transaction_number":%q,"total_amount":25.00,"status":"completed"}`, f.txID, f.terminalID, f.sessionID, f.txNum))
			linePayload := []byte(fmt.Sprintf(`{"id":%q,"transaction_id":%q,"quantity":1,"line_total":25.00}`, f.lineID, f.txID))
			payPayload := []byte(fmt.Sprintf(`{"id":%q,"transaction_id":%q,"payment_method":"cash","amount":25.00}`, f.payID, f.txID))

			events := []*syncpb.SyncEvent{
				{
					Id:          1,
					OutboxUuid:  f.outboxPfx + "1",
					BatchId:     f.batchID,
					BatchSize:   4,
					SeqInBatch:  1,
					EntityType:  "cashier_sessions",
					EntityKey:   f.sessionID,
					Action:      "INSERT",
					PayloadJson: sessPayload,
					Sha256:      computePayloadSHA256(sessPayload),
					StoreId:     1,
					TenantSlug:  "demo",
				},
				{
					Id:          2,
					OutboxUuid:  f.outboxPfx + "2",
					BatchId:     f.batchID,
					BatchSize:   4,
					SeqInBatch:  2,
					EntityType:  "pos_transactions",
					EntityKey:   f.txID,
					Action:      "INSERT",
					PayloadJson: txPayload,
					Sha256:      computePayloadSHA256(txPayload),
					StoreId:     1,
					TenantSlug:  "demo",
				},
				{
					Id:          3,
					OutboxUuid:  f.outboxPfx + "3",
					BatchId:     f.batchID,
					BatchSize:   4,
					SeqInBatch:  3,
					EntityType:  "pos_transaction_lines",
					EntityKey:   f.lineID,
					Action:      "INSERT",
					PayloadJson: linePayload,
					Sha256:      computePayloadSHA256(linePayload),
					StoreId:     1,
					TenantSlug:  "demo",
				},
				{
					Id:          4,
					OutboxUuid:  f.outboxPfx + "4",
					BatchId:     f.batchID,
					BatchSize:   4,
					SeqInBatch:  4,
					EntityType:  "pos_payments",
					EntityKey:   f.payID,
					Action:      "INSERT",
					PayloadJson: payPayload,
					Sha256:      computePayloadSHA256(payPayload),
					StoreId:     1,
					TenantSlug:  "demo",
				},
			}
			acks[idx] = srv.processPushBatchWithAuth(ctx, events, f.terminalID, 1)
		}(i, fx)
	}
	wg.Wait()

	for i, ack := range acks {
		if ack == nil || !ack.Success {
			t.Fatalf("terminal %d batch push failed: %+v", i+1, ack)
		}
	}

	var sessionCount, txCount, reparentedCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM cashier_sessions`).Scan(&sessionCount); err != nil || sessionCount != 2 {
		t.Fatalf("expected 2 distinct cashier_sessions, got %d (err=%v)", sessionCount, err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM pos_transactions`).Scan(&txCount); err != nil || txCount != 2 {
		t.Fatalf("expected 2 distinct pos_transactions, got %d (err=%v)", txCount, err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT count(*)
		FROM pos_transaction_lines l
		JOIN pos_transactions pt ON pt.id = l.transaction_id
		WHERE (l.id = '11111111-1111-4111-8111-111111111103'::uuid AND pt.pos_terminal_id <> 1)
		   OR (l.id = '22222222-2222-4222-8222-222222222203'::uuid AND pt.pos_terminal_id <> 2)
	`).Scan(&reparentedCount); err != nil || reparentedCount != 0 {
		t.Fatalf("expected 0 cross-reparented child rows, got %d (err=%v)", reparentedCount, err)
	}
}

func TestCrossTerminalOverwriteRejection(t *testing.T) {
	ctx := context.Background()

	// 1. Pre-DB payload ownership check rejects Terminal 2 claiming Terminal 1's pos_terminal_id
	err := validateEntityOwnership(
		ctx,
		nil,
		"pos_transactions",
		[]byte(`{"id":"11111111-1111-4111-8111-111111111102","store_id":1,"pos_terminal_id":1}`),
		1,
		2,
	)
	if err == nil || !strings.Contains(err.Error(), "ownership") {
		t.Fatalf("expected ownership violation when terminal 2 pushes pos_terminal_id=1, got %v", err)
	}

	// 2. Full DB conflict guard when Terminal 2 attempts to overwrite Terminal 1's cashier_session or pos_transaction
	pool, cleanup := openMultiTerminalTestPool(t, "cross_term_reject")
	defer cleanup()

	srv := NewSyncServer(nil, nil)
	srv.SetPoolResolver(func(ctx context.Context, slug string) (*pgxpool.Pool, error) {
		return pool, nil
	})

	const sess1ID = "11111111-1111-4111-8111-111111111101"
	const tx1ID = "11111111-1111-4111-8111-111111111102"

	sess1Payload := []byte(`{"id":"` + sess1ID + `","pos_terminal_id":1,"session_number":"SES-T1-01","status":"open"}`)
	ackSess1 := srv.handleSinglePushEventWithAuth(ctx, &syncpb.SyncEvent{
		Id:          1,
		EntityType:  "cashier_sessions",
		EntityKey:   sess1ID,
		Action:      "INSERT",
		PayloadJson: sess1Payload,
		Sha256:      computePayloadSHA256(sess1Payload),
		StoreId:     1,
		TenantSlug:  "demo",
	}, 1, 1)
	if !ackSess1.Success {
		t.Fatalf("terminal 1 session insert failed: %s", ackSess1.ErrorMessage)
	}

	tx1Payload := []byte(`{"id":"` + tx1ID + `","store_id":1,"pos_terminal_id":1,"cashier_session_id":"` + sess1ID + `","transaction_number":"POS-T1-001","total_amount":50.00,"status":"completed"}`)
	ackTx1 := srv.handleSinglePushEventWithAuth(ctx, &syncpb.SyncEvent{
		Id:          2,
		EntityType:  "pos_transactions",
		EntityKey:   tx1ID,
		Action:      "INSERT",
		PayloadJson: tx1Payload,
		Sha256:      computePayloadSHA256(tx1Payload),
		StoreId:     1,
		TenantSlug:  "demo",
	}, 1, 1)
	if !ackTx1.Success {
		t.Fatalf("terminal 1 transaction insert failed: %s", ackTx1.ErrorMessage)
	}

	// Terminal 2 attempts to overwrite Terminal 1's cashier_session ID with pos_terminal_id=2
	hijackSessPayload := []byte(`{"id":"` + sess1ID + `","pos_terminal_id":2,"session_number":"SES-HIJACK","status":"closed"}`)
	ackHijackSess := srv.handleSinglePushEventWithAuth(ctx, &syncpb.SyncEvent{
		Id:          3,
		EntityType:  "cashier_sessions",
		EntityKey:   sess1ID,
		Action:      "UPDATE",
		PayloadJson: hijackSessPayload,
		Sha256:      computePayloadSHA256(hijackSessPayload),
		StoreId:     1,
		TenantSlug:  "demo",
	}, 2, 1)
	if ackHijackSess.Success || !strings.Contains(ackHijackSess.ErrorMessage, "ownership") {
		t.Fatalf("expected ownership failure on cross-terminal cashier_session overwrite, got %+v", ackHijackSess)
	}

	// Terminal 2 attempts to overwrite Terminal 1's pos_transaction ID with pos_terminal_id=2
	hijackTxPayload := []byte(`{"id":"` + tx1ID + `","store_id":1,"pos_terminal_id":2,"cashier_session_id":"` + sess1ID + `","transaction_number":"POS-T1-001","total_amount":999.00,"status":"voided"}`)
	ackHijackTx := srv.handleSinglePushEventWithAuth(ctx, &syncpb.SyncEvent{
		Id:          4,
		EntityType:  "pos_transactions",
		EntityKey:   tx1ID,
		Action:      "UPDATE",
		PayloadJson: hijackTxPayload,
		Sha256:      computePayloadSHA256(hijackTxPayload),
		StoreId:     1,
		TenantSlug:  "demo",
	}, 2, 1)
	if ackHijackTx.Success || !strings.Contains(ackHijackTx.ErrorMessage, "ownership") {
		t.Fatalf("expected ownership failure on cross-terminal pos_transaction overwrite, got %+v", ackHijackTx)
	}

	// Verify Terminal 1's rows remain completely unmodified
	var sessTermID int32
	var sessStatus string
	if err := pool.QueryRow(ctx, `SELECT pos_terminal_id, status FROM cashier_sessions WHERE id = $1::uuid`, sess1ID).Scan(&sessTermID, &sessStatus); err != nil {
		t.Fatal(err)
	}
	if sessTermID != 1 || sessStatus != "open" {
		t.Fatalf("cashier_session was modified by cross-terminal overwrite: term=%d status=%s", sessTermID, sessStatus)
	}

	var txTermID int32
	var txStatus string
	var txTotal float64
	if err := pool.QueryRow(ctx, `SELECT pos_terminal_id, status, total_amount::float8 FROM pos_transactions WHERE id = $1::uuid`, tx1ID).Scan(&txTermID, &txStatus, &txTotal); err != nil {
		t.Fatal(err)
	}
	if txTermID != 1 || txStatus != "completed" || txTotal != 50.0 {
		t.Fatalf("pos_transaction was modified by cross-terminal overwrite: term=%d status=%s total=%v", txTermID, txStatus, txTotal)
	}
}
