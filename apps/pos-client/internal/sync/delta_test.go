package sync

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/syncpb"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestDeltaDataAndWatermarkCommitTogether(t *testing.T) {
	dsn := os.Getenv("NEMBUS_SYNC_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set NEMBUS_SYNC_TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := fmt.Sprintf("delta_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE") }()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	_, err = pool.Exec(ctx, `CREATE TABLE promotions(id integer PRIMARY KEY, discount_value integer NOT NULL,
		is_active boolean, created_at timestamp, updated_at timestamp);
		CREATE TABLE sync_watermarks(entity_type text,store_id integer,last_sync_at timestamptz,
		metadata jsonb, UNIQUE(entity_type,store_id));
		INSERT INTO promotions VALUES(1,10,true,'2026-01-01','2026-01-01');
		INSERT INTO sync_watermarks VALUES('inventory_stock',1,'2026-01-01','{}')`)
	if err != nil {
		t.Fatal(err)
	}
	s := NewSyncService(ctx, pool, "localhost:50051", "test")
	checkpoint := time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)
	events := []*syncpb.SyncEvent{{EntityId: 1, PayloadJson: []byte(`{"id":1,"discount_value":20,"is_active":false,"created_at":"2026-01-01","updated_at":"2026-10-10"}`)}}
	if err := s.saveDeltaPage(1, "promotions", checkpoint, events, 0); err != nil {
		t.Fatal(err)
	}
	var discount int
	var active bool
	if err := pool.QueryRow(ctx, "SELECT discount_value,is_active FROM promotions WHERE id=1").Scan(&discount, &active); err != nil || discount != 20 || active {
		t.Fatalf("promotion update: discount=%d active=%v err=%v", discount, active, err)
	}
	// A later invalid row rolls back both the earlier write and checkpoint.
	events[0].PayloadJson = []byte(`{"id":1,"discount_value":30,"is_active":true}`)
	events = append(events, &syncpb.SyncEvent{EntityId: 2, PayloadJson: []byte(`{"id":2}`)})
	if err := s.saveDeltaPage(1, "promotions", checkpoint.Add(time.Hour), events, 0); err == nil {
		t.Fatal("expected constraint failure")
	}
	var saved time.Time
	if err := pool.QueryRow(ctx, "SELECT last_sync_at FROM sync_watermarks WHERE entity_type='promotions'").Scan(&saved); err != nil || !saved.Equal(checkpoint) {
		t.Fatalf("failed page advanced checkpoint: %v %v", saved, err)
	}
	if err := pool.QueryRow(ctx, "SELECT discount_value FROM promotions WHERE id=1").Scan(&discount); err != nil || discount != 20 {
		t.Fatalf("failed page changed data: %d %v", discount, err)
	}
	if err := s.saveDeltaPage(1, "promotions", checkpoint, nil, 1); err != nil {
		t.Fatal(err)
	}
	var state string
	if err := pool.QueryRow(ctx, "SELECT metadata->>'status' FROM sync_watermarks WHERE entity_type='promotions'").Scan(&state); err != nil || state != "synced" {
		t.Fatalf("empty completion not recorded: %s %v", state, err)
	}
	if err := pool.QueryRow(ctx, "SELECT last_sync_at FROM sync_watermarks WHERE entity_type='inventory_stock'").Scan(&saved); err != nil || saved.Equal(checkpoint) {
		t.Fatalf("promotion changed inventory watermark: %v %v", saved, err)
	}
	var cursorID int64
	if err := pool.QueryRow(ctx, "SELECT (metadata->>'cursor_id')::bigint FROM sync_watermarks WHERE entity_type='promotions'").Scan(&cursorID); err != nil || cursorID != 1 {
		t.Fatalf("cursor not preserved after empty page: %d %v", cursorID, err)
	}
	// A locked local row must time out without advancing its checkpoint.
	lock, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback(ctx)
	if _, err := lock.Exec(ctx, "SELECT id FROM promotions WHERE id=1 FOR UPDATE"); err != nil {
		t.Fatal(err)
	}
	shortCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	err = s.saveDeltaPageContext(shortCtx, 1, "promotions", checkpoint.Add(time.Hour), events[:1], 0)
	cancel()
	if err == nil || shortCtx.Err() != context.DeadlineExceeded {
		t.Fatalf("blocked page did not time out: %v", err)
	}
	if err := lock.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, "SELECT last_sync_at FROM sync_watermarks WHERE entity_type='promotions'").Scan(&saved); err != nil || !saved.Equal(checkpoint) {
		t.Fatalf("blocked page advanced cursor: %v %v", saved, err)
	}
	// A large backlog yields after five pages and resumes from its saved ID.
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	syncpb.RegisterSyncServiceServer(server, &backlogSyncServer{timestamp: checkpoint.Add(time.Second)})
	go func() { _ = server.Serve(listener) }()
	defer server.Stop()
	defer listener.Close()
	conn, err := grpc.NewClient("passthrough:///backlog-test", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := syncpb.NewSyncServiceClient(conn)
	if err := s.pullEntity(client, 1, "promotions"); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, "SELECT (metadata->>'cursor_id')::bigint,metadata->>'status' FROM sync_watermarks WHERE entity_type='promotions'").Scan(&cursorID, &state); err != nil || cursorID != 1000 || state != "syncing" {
		t.Fatalf("backlog did not yield: cursor=%d state=%s err=%v", cursorID, state, err)
	}
	if err := s.pullEntity(client, 1, "promotions"); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, "SELECT (metadata->>'cursor_id')::bigint,metadata->>'status' FROM sync_watermarks WHERE entity_type='promotions'").Scan(&cursorID, &state); err != nil || cursorID != 1200 || state != "synced" {
		t.Fatalf("backlog did not resume: cursor=%d state=%s err=%v", cursorID, state, err)
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM promotions").Scan(&count); err != nil || count != 1200 {
		t.Fatalf("lost backlog rows: %d %v", count, err)
	}
}

type backlogSyncServer struct {
	syncpb.UnimplementedSyncServiceServer
	timestamp time.Time
}

func (s *backlogSyncServer) StreamPull(req *syncpb.PullRequest, stream syncpb.SyncService_StreamPullServer) error {
	md, _ := metadata.FromIncomingContext(stream.Context())
	afterID, err := strconv.ParseInt(md.Get(syncpb.PullCursorKey)[0], 10, 64)
	if err != nil {
		return err
	}
	if req.Since.AsTime().Before(s.timestamp) {
		afterID = 0
	}
	if err := stream.SendHeader(metadata.Pairs(syncpb.PullCursorVersionKey, "1")); err != nil {
		return err
	}
	for id := afterID + 1; id <= 1200 && id <= afterID+int64(req.Limit); id++ {
		payload := []byte(fmt.Sprintf(`{"id":%d,"discount_value":5,"is_active":true,"updated_at":"%s"}`, id, s.timestamp.Format(time.RFC3339Nano)))
		hash := sha256.Sum256(payload)
		if err := stream.Send(&syncpb.SyncEvent{EntityType: "promotions", EntityId: id, PayloadJson: payload, Sha256: fmt.Sprintf("%x", hash), EventTime: timestamppb.New(s.timestamp)}); err != nil {
			return err
		}
	}
	return nil
}

func TestCursorTimestampTies(t *testing.T) {
	now := time.Now()
	for _, test := range []struct {
		timestamp time.Time
		id        int64
		want      bool
	}{
		{now, 10, false}, {now, 11, true}, {now, 9, false},
		{now.Add(time.Second), 1, true}, {now.Add(-time.Second), 100, false},
	} {
		if got := cursorAfter(test.timestamp, test.id, now, 10); got != test.want {
			t.Fatalf("cursor timestamp=%s id=%d got=%v want=%v", test.timestamp, test.id, got, test.want)
		}
	}
}

func TestStopWaitsForCanceledWorker(t *testing.T) {
	s := NewSyncService(context.Background(), nil, "localhost:50051", "test")
	started := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		s.syncMu.Lock()
		defer s.syncMu.Unlock()
		close(started)
		<-s.ctx.Done()
		close(finished)
	}()
	<-started
	s.Stop()
	select {
	case <-finished:
	default:
		t.Fatal("Stop returned before worker exited")
	}
	if s.ctx.Err() == nil {
		t.Fatal("worker context was not canceled")
	}
}
