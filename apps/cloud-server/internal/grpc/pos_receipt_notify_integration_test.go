package grpc

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestEnqueuePosReceiptIdempotent(t *testing.T) {
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
	schema := fmt.Sprintf("receipt_notify_%d", time.Now().UnixNano())
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
	if _, err := pool.Exec(ctx, `
		CREATE TABLE stores (id integer PRIMARY KEY, organization_id integer NOT NULL, name text);
		INSERT INTO stores (id, organization_id, name) VALUES (3, 7, 'Main');
		CREATE TABLE notification_events (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			organization_id integer NOT NULL,
			event_type_id text NOT NULL,
			entity_type text NOT NULL,
			entity_id text NOT NULL,
			idempotency_key text NOT NULL UNIQUE,
			payload jsonb NOT NULL,
			status text NOT NULL DEFAULT 'pending',
			retry_count integer DEFAULT 0,
			created_at timestamptz DEFAULT now(),
			dispatched_at timestamptz
		);
	`); err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"id":9,"transaction_number":"POS-9","status":"completed","store_id":3,"customer_id":4,"voided_at":null}`)
	if err := EnqueuePosReceiptNotification(ctx, pool, payload); err != nil {
		t.Fatal(err)
	}
	if err := EnqueuePosReceiptNotification(ctx, pool, payload); err != nil {
		t.Fatal(err)
	}
	voided := []byte(`{"id":10,"transaction_number":"POS-10","status":"voided","store_id":3}`)
	if err := EnqueuePosReceiptNotification(ctx, pool, voided); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM notification_events`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("expected one receipt event, count=%d err=%v", count, err)
	}
	var key, status string
	var org int32
	if err := pool.QueryRow(ctx, `SELECT idempotency_key, status, organization_id FROM notification_events`).Scan(&key, &status, &org); err != nil {
		t.Fatal(err)
	}
	if key != "pos.receipt:9:POS-9" || status != "pending" || org != 7 {
		t.Fatalf("event key=%s status=%s org=%d", key, status, org)
	}
}
