package grpc

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/syncpb"
	"github.com/NasTecSol/nembus-core/usecase"
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
