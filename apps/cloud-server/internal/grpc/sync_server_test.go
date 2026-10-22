package grpc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCheckoutParentsBeforePOS(t *testing.T) {
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
	schema := fmt.Sprintf("sync_test_%d", time.Now().UnixNano())
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
		CREATE TABLE sales_orders_v2 (id uuid PRIMARY KEY, source_cart_id uuid REFERENCES carts(id));
		CREATE TABLE cashier_sessions (id integer PRIMARY KEY);
		CREATE TABLE pos_transactions (id integer PRIMARY KEY, cashier_session_id integer REFERENCES cashier_sessions(id),
			sales_order_id uuid REFERENCES sales_orders_v2(id), source_cart_id uuid REFERENCES carts(id), total_amount numeric);
		CREATE TABLE pos_transaction_lines (id integer PRIMARY KEY, transaction_id integer REFERENCES pos_transactions(id));
		CREATE TABLE pos_payments (id integer PRIMARY KEY, transaction_id integer REFERENCES pos_transactions(id), amount numeric);
	`)
	if err != nil {
		t.Fatal(err)
	}
	s := &SyncServer{}
	upsert := func(table, payload string) error {
		return s.upsertEntityJSON(ctx, pool, table, "INSERT", []byte(payload))
	}
	if err := upsert("cashier_sessions", `{"id":1}`); err != nil {
		t.Fatal(err)
	}
	transaction := `{"id":1,"cashier_session_id":1,"sales_order_id":"00000000-0000-0000-0000-000000000002","source_cart_id":"00000000-0000-0000-0000-000000000001","total_amount":10}`
	// Reproduce the real queue error even with an already synced session.
	err = upsert("pos_transactions", transaction)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" || pgErr.ConstraintName != "pos_transactions_sales_order_id_fkey" {
		t.Fatalf("expected original sales-order FK error, got %v", err)
	}
	for _, event := range []struct{ table, payload string }{
		{"carts", `{"id":"00000000-0000-0000-0000-000000000001"}`},
		{"sales_orders_v2", `{"id":"00000000-0000-0000-0000-000000000002","source_cart_id":"00000000-0000-0000-0000-000000000001"}`},
		{"pos_transactions", transaction},
		{"pos_transaction_lines", `{"id":1,"transaction_id":1}`},
		{"pos_payments", `{"id":1,"transaction_id":1,"amount":10}`},
	} {
		for attempt := 0; attempt < 2; attempt++ {
			if err := upsert(event.table, event.payload); err != nil {
				t.Fatalf("%s: %v", event.table, err)
			}
		}
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM pos_payments WHERE transaction_id=1 AND amount=10").Scan(&count); err != nil || count != 1 {
		t.Fatalf("payment was not persisted exactly once: count=%d err=%v", count, err)
	}
	// An invalid retry must fail, rather than being acknowledged by DO NOTHING.
	if err := upsert("pos_payments", `{"id":1,"transaction_id":999,"amount":20}`); err == nil {
		t.Fatal("invalid conflicting update was silently accepted")
	}
}

func TestUpsertAcceptedEntityTypes(t *testing.T) {
	s := &SyncServer{}
	ctx := context.Background()

	// Empty payload check
	err := s.upsertEntityJSON(ctx, nil, "sales_returns", "INSERT", []byte{})
	if err == nil || !strings.Contains(err.Error(), "empty sync payload for sales_returns") {
		t.Fatalf("expected empty payload error, got %v", err)
	}

	// Unsupported table check
	err = s.upsertEntityJSON(ctx, nil, "unknown_table_foo", "INSERT", []byte(`{"id":1}`))
	if err == nil || !strings.Contains(err.Error(), "unsupported entity type for upsert: unknown_table_foo") {
		t.Fatalf("expected unsupported entity type error, got %v", err)
	}

	// Ensure sales_returns, sales_return_lines, and stock_movements are accepted as valid tables
	for _, table := range []string{"sales_returns", "sales_return_lines", "stock_movements"} {
		err := s.upsertEntityJSON(ctx, nil, table, "INSERT", []byte(`{"id":1}`))
		// With nil pool, it should pass the validTables check and fail on getTableUpdateClause
		if err != nil && strings.Contains(err.Error(), "unsupported entity type for upsert") {
			t.Errorf("entity type %s was rejected as unsupported by upsertEntityJSON", table)
		}
	}
}

func TestSyncServerUpsertReturnsAndMovements(t *testing.T) {
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
	schema := fmt.Sprintf("returns_movements_test_%d", time.Now().UnixNano())
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
		CREATE TABLE sales_returns (
			id bigint PRIMARY KEY,
			return_number text,
			status text,
			subtotal numeric,
			total_refund_amount numeric,
			created_at timestamp DEFAULT now(),
			updated_at timestamp DEFAULT now()
		);
		CREATE TABLE sales_return_lines (
			id bigint PRIMARY KEY,
			return_id bigint REFERENCES sales_returns(id),
			quantity numeric,
			refund_amount numeric,
			created_at timestamp DEFAULT now(),
			updated_at timestamp DEFAULT now()
		);
		CREATE TABLE stock_movements (
			id bigint PRIMARY KEY,
			movement_type text,
			quantity numeric,
			status text,
			created_at timestamp DEFAULT now(),
			updated_at timestamp DEFAULT now()
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	s := &SyncServer{}
	upsert := func(table, payload string) error {
		return s.upsertEntityJSON(ctx, pool, table, "INSERT", []byte(payload))
	}

	// 1. Upsert sales_returns
	retPayload := `{"id":100,"return_number":"RET-001","status":"pending","subtotal":50.00,"total_refund_amount":57.50}`
	if err := upsert("sales_returns", retPayload); err != nil {
		t.Fatalf("upsert sales_returns: %v", err)
	}

	// 2. Upsert sales_return_lines
	linePayload := `{"id":200,"return_id":100,"quantity":1.000,"refund_amount":57.50}`
	if err := upsert("sales_return_lines", linePayload); err != nil {
		t.Fatalf("upsert sales_return_lines: %v", err)
	}

	// 3. Upsert stock_movements
	smPayload := `{"id":300,"movement_type":"waste","quantity":-1.000,"status":"completed"}`
	if err := upsert("stock_movements", smPayload); err != nil {
		t.Fatalf("upsert stock_movements: %v", err)
	}

	// Verify records in database
	var retCount, lineCount, smCount int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM sales_returns WHERE id=100 AND status='pending'").Scan(&retCount); err != nil || retCount != 1 {
		t.Fatalf("sales_returns not inserted correctly: count=%d, err=%v", retCount, err)
	}
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM sales_return_lines WHERE id=200 AND return_id=100").Scan(&lineCount); err != nil || lineCount != 1 {
		t.Fatalf("sales_return_lines not inserted correctly: count=%d, err=%v", lineCount, err)
	}
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM stock_movements WHERE id=300 AND movement_type='waste'").Scan(&smCount); err != nil || smCount != 1 {
		t.Fatalf("stock_movements not inserted correctly: count=%d, err=%v", smCount, err)
	}

	// Update existing record via upsert (idempotent / on conflict)
	retUpdate := `{"id":100,"return_number":"RET-001","status":"completed","subtotal":50.00,"total_refund_amount":57.50}`
	if err := upsert("sales_returns", retUpdate); err != nil {
		t.Fatalf("upsert sales_returns update: %v", err)
	}
	var newStatus string
	if err := pool.QueryRow(ctx, "SELECT status FROM sales_returns WHERE id=100").Scan(&newStatus); err != nil || newStatus != "completed" {
		t.Fatalf("sales_returns status not updated: status=%s, err=%v", newStatus, err)
	}
}
