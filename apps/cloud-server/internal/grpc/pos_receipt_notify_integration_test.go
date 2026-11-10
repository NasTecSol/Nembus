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
		CREATE TABLE customers (id integer PRIMARY KEY, phone text, email text);
		INSERT INTO customers (id, phone, email) VALUES (4, '966500000001', '');
		CREATE TABLE carts (id uuid PRIMARY KEY, guest_phone varchar(50), guest_email varchar(255));
		INSERT INTO carts VALUES ('00000000-0000-0000-0000-000000000001', '+923216061056', '');
		CREATE TABLE sales_orders_v2 (
			id uuid PRIMARY KEY,
			customer_phone varchar(50),
			customer_email varchar(50),
			source_cart_id uuid REFERENCES carts(id)
		);
		INSERT INTO sales_orders_v2 VALUES ('00000000-0000-0000-0000-000000000002', '', '', '00000000-0000-0000-0000-000000000001');
		CREATE TABLE pos_transactions (
			id bigint PRIMARY KEY,
			store_id integer NOT NULL REFERENCES stores(id),
			customer_id integer REFERENCES customers(id),
			sales_order_id uuid REFERENCES sales_orders_v2(id),
			source_cart_id uuid REFERENCES carts(id),
			transaction_number text NOT NULL,
			status text NOT NULL DEFAULT 'completed',
			metadata jsonb DEFAULT '{}'::jsonb
		);
		INSERT INTO pos_transactions (id, store_id, sales_order_id, source_cart_id, transaction_number, status)
		VALUES (11, 3, '00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'POS-11', 'completed');
		INSERT INTO pos_transactions (id, store_id, customer_id, transaction_number, status)
		VALUES (9, 3, 4, 'POS-9', 'completed');
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
	noContact := []byte(`{"id":12,"transaction_number":"POS-12","status":"completed","store_id":3}`)
	if _, err := pool.Exec(ctx, `INSERT INTO pos_transactions (id, store_id, transaction_number, status) VALUES (12, 3, 'POS-12', 'completed')`); err != nil {
		t.Fatal(err)
	}
	if err := EnqueuePosReceiptNotification(ctx, pool, noContact); err != nil {
		t.Fatal(err)
	}
	guestPayload := []byte(`{"id":11,"transaction_number":"POS-11","status":"completed","store_id":3,"sales_order_id":"00000000-0000-0000-0000-000000000002","source_cart_id":"00000000-0000-0000-0000-000000000001"}`)
	if err := EnqueuePosReceiptNotification(ctx, pool, guestPayload); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM notification_events`).Scan(&count); err != nil || count != 2 {
		t.Fatalf("expected two receipt events, count=%d err=%v", count, err)
	}
	var key, status string
	var org int32
	var payloadPhone string
	if err := pool.QueryRow(ctx, `
		SELECT idempotency_key, status, organization_id, payload->>'customer_phone'
		FROM notification_events
		WHERE idempotency_key = 'pos.receipt:9:POS-9'
	`).Scan(&key, &status, &org, &payloadPhone); err != nil {
		t.Fatal(err)
	}
	if key != "pos.receipt:9:POS-9" || status != "pending" || org != 7 || payloadPhone != "966500000001" {
		t.Fatalf("event key=%s status=%s org=%d phone=%s", key, status, org, payloadPhone)
	}
	if err := pool.QueryRow(ctx, `
		SELECT payload->>'customer_phone' FROM notification_events WHERE idempotency_key = 'pos.receipt:11:POS-11'
	`).Scan(&payloadPhone); err != nil || payloadPhone != "+923216061056" {
		t.Fatalf("guest cart phone: %s err=%v", payloadPhone, err)
	}
}
