package sync

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/syncpb"
	"github.com/jackc/pgx/v5/pgxpool"
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
	// Timestamp ties must all be delivered before advancing a timestamp cursor.
	rows, err := pool.Query(ctx, `SELECT id FROM (VALUES(1,'2026-01-01'::timestamp),(2,'2026-01-01'::timestamp),
		(3,'2026-01-02'::timestamp)) AS t(id,updated_at) ORDER BY updated_at FETCH FIRST ($1) ROWS WITH TIES`, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
	}
	if rows.Err() != nil || count != 2 {
		t.Fatalf("timestamp boundary lost: count=%d err=%v", count, rows.Err())
	}
}
