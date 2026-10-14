package grpc

import (
	"context"
	"errors"
	"fmt"
	"os"
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
