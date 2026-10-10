package sync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestTransactionParentsForExistingQueueEntry(t *testing.T) {
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
	schema := fmt.Sprintf("sync_parent_test_%d", time.Now().UnixNano())
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
	_, err = pool.Exec(ctx, `
		CREATE TABLE carts (id uuid PRIMARY KEY);
		CREATE TABLE sales_orders_v2 (id uuid PRIMARY KEY, source_cart_id uuid REFERENCES carts(id), fulfillment_status text, order_status text);
CREATE TABLE sales_order_lines_v2 (id uuid PRIMARY KEY, sales_order_id uuid REFERENCES sales_orders_v2(id), line_number integer);
		INSERT INTO carts VALUES ('00000000-0000-0000-0000-000000000001');
		INSERT INTO sales_orders_v2 VALUES ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'fulfilled', 'fulfilled');
INSERT INTO sales_order_lines_v2 VALUES ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000002', 1);
		CREATE TABLE sync_queue (id bigint PRIMARY KEY, entity_type text, entity_id bigint, action text,
			payload jsonb, created_at timestamp DEFAULT now(), correlation_id text, status text);
		INSERT INTO sync_queue (id, entity_type, entity_id, action, payload, status) VALUES
			(1, 'pos_transactions', 7, 'INSERT', '{"id":7}', 'failed'),
			(2, 'pos_transaction_lines', 8, 'INSERT', '{"id":8,"transaction_id":7}', 'pending'),
			(3, 'pos_payments', 9, 'INSERT', '{"id":9,"transaction_id":7}', 'pending'),
			(4, 'cashier_sessions', 10, 'INSERT', '{"id":10}', 'pending');
	`)
	if err != nil {
		t.Fatal(err)
	}
	s := NewSyncService(ctx, pool, "localhost:50051", "test")
	item := OutboxItem{ID: 42, EntityType: "pos_transactions", Payload: json.RawMessage(`{"id":7,"sales_order_id":"00000000-0000-0000-0000-000000000002","source_cart_id":"00000000-0000-0000-0000-000000000001"}`)}
	parents, err := s.transactionParents(item)
	if err != nil {
		t.Fatal(err)
	}
	if len(parents) != 2 || parents[0].EntityType != "carts" || parents[1].EntityType != "sales_orders_v2" {
		t.Fatalf("expected one cart before order, got %+v", parents)
	}
	for _, p := range parents {
		if p.ID != item.ID || (p.EntityType == "carts" && p.Action != "INSERT") || (p.EntityType == "sales_orders_v2" && p.Action != "POS_CHECKOUT") || !json.Valid(p.Payload) {
			t.Fatalf("invalid parent event: %+v", p)
		}
	}
	var snapshot struct {
		Order json.RawMessage   `json:"order"`
		Lines []json.RawMessage `json:"lines"`
	}
	if err := json.Unmarshal(parents[1].Payload, &snapshot); err != nil || len(snapshot.Lines) != 1 {
		t.Fatalf("missing checkout lines: %s", parents[1].Payload)
	}
	if _, err := pool.Exec(ctx, "UPDATE sales_orders_v2 SET fulfillment_status='unfulfilled'"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.transactionParents(item); !errors.Is(err, errCheckoutNotReady) {
		t.Fatalf("expected unfinished checkout to wait, got %v", err)
	}
	if _, err := pool.Exec(ctx, "UPDATE sales_orders_v2 SET fulfillment_status='fulfilled'"); err != nil {
		t.Fatal(err)
	}
	// The order's source cart must also be sent if an older POS payload omits it.
	item.Payload = json.RawMessage(`{"id":7,"sales_order_id":"00000000-0000-0000-0000-000000000002"}`)
	parents, err = s.transactionParents(item)
	if err != nil || len(parents) != 2 {
		t.Fatalf("order-only reference: parents=%+v err=%v", parents, err)
	}
	item.Payload = json.RawMessage(`{"id":7,"sales_order_id":null,"source_cart_id":null}`)
	parents, err = s.transactionParents(item)
	if err != nil || len(parents) != 0 {
		t.Fatalf("direct POS sale: parents=%+v err=%v", parents, err)
	}
	item.EntityType = "pos_payments"
	parents, err = s.transactionParents(item)
	if err != nil || len(parents) != 0 {
		t.Fatalf("payment unexpectedly has checkout parents")
	}
	eligible := func() []int64 {
		t.Helper()
		rows, err := pool.Query(ctx, pendingOutboxSQL)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var ids []int64
		for rows.Next() {
			var item OutboxItem
			if err := rows.Scan(&item.ID, &item.EntityType, &item.EntityID, &item.Action, &item.Payload, &item.CreatedAt, &item.CorrelationID); err != nil {
				t.Fatal(err)
			}
			ids = append(ids, item.ID)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		return ids
	}
	if ids := eligible(); len(ids) != 1 || ids[0] != 4 {
		t.Fatalf("children of failed transaction must wait; eligible=%v", ids)
	}
	if _, err := pool.Exec(ctx, "UPDATE sync_queue SET status='pending' WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	if ids := eligible(); len(ids) != 2 || ids[0] != 1 || ids[1] != 4 {
		t.Fatalf("children of pending transaction must wait; eligible=%v", ids)
	}
	if _, err := pool.Exec(ctx, "UPDATE sync_queue SET status='synced' WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	if ids := eligible(); len(ids) != 3 || ids[0] != 2 || ids[1] != 3 || ids[2] != 4 {
		t.Fatalf("children must become eligible in order after transaction sync; eligible=%v", ids)
	}
}
