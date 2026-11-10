package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestNullTaxOrderTotalsPostgres(t *testing.T) {
	dsn := os.Getenv("NEMBUS_ORDER_TOTALS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set NEMBUS_ORDER_TOTALS_TEST_DATABASE_URL for isolated PostgreSQL tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := fmt.Sprintf("order_totals_%d", time.Now().UnixNano())
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`CREATE TABLE sales_orders_v2(id uuid PRIMARY KEY, subtotal numeric(15,2) DEFAULT 0,
 tax_amount numeric(15,2) DEFAULT 0,total_amount numeric(15,2) DEFAULT 0,discount_amount numeric(15,2) DEFAULT 0,
 shipping_amount numeric(15,2) DEFAULT 0,adjustment_amount numeric(15,2) DEFAULT 0,paid_amount numeric(15,2) DEFAULT 0,
 balance_due numeric(15,2) DEFAULT 0,updated_at timestamptz DEFAULT now());
 CREATE TABLE sales_order_lines_v2(id serial PRIMARY KEY,sales_order_id uuid REFERENCES sales_orders_v2,
 line_total numeric(15,2),tax_amount numeric(15,2),discount_amount numeric(15,2));
 CREATE TABLE invoices(id serial PRIMARY KEY,sales_order_id uuid REFERENCES sales_orders_v2);
 CREATE TABLE pos_transactions(id serial PRIMARY KEY,sales_order_id uuid REFERENCES sales_orders_v2);
 CREATE TABLE zatca_document_chain(entity_type text,entity_id text);
 CREATE TABLE zatca_order_jobs(order_id uuid PRIMARY KEY REFERENCES sales_orders_v2,last_error text);
 CREATE TABLE carts(id uuid PRIMARY KEY);
 CREATE TABLE cart_items(cart_id uuid REFERENCES carts,line_total numeric(15,2),tax_amount numeric(15,2),discount_amount numeric(15,2));`)
	// Reproduce an existing tenant with the old trigger before applying the runtime upgrade.
	canonical, err := os.ReadFile(filepath.Join("..", "..", "db", "views_functions", "90_views_functions.sql"))
	if err != nil {
		t.Fatal(err)
	}
	trigger := strings.Split(string(canonical), "-- TRIGGERS FOR INVOICE CALCULATIONS")[0]
	exec(strings.ReplaceAll(trigger, "line_total - COALESCE(tax_amount, 0)", "line_total - tax_amount"))
	id := func(n int) string { return fmt.Sprintf("00000000-0000-0000-0000-%012d", n) }
	for n := 1; n <= 8; n++ {
		exec("INSERT INTO sales_orders_v2(id) VALUES($1::uuid)", id(n))
		exec("INSERT INTO zatca_order_jobs VALUES($1::uuid,'unexplained difference between order and line totals')", id(n))
		exec("INSERT INTO sales_order_lines_v2(sales_order_id,line_total,tax_amount) VALUES($1::uuid,17.83,NULL)", id(n))
	}
	assertTotals := func(n int, subtotal, tax, total, balance string) {
		t.Helper()
		var a, b, c, d string
		err := pool.QueryRow(ctx, "SELECT subtotal::text,tax_amount::text,total_amount::text,balance_due::text FROM sales_orders_v2 WHERE id=$1::uuid", id(n)).Scan(&a, &b, &c, &d)
		if err != nil {
			t.Fatal(err)
		}
		if a != subtotal || b != tax || c != total || d != balance {
			t.Fatalf("order %d totals=(%s,%s,%s,%s), want=(%s,%s,%s,%s)", n, a, b, c, d, subtotal, tax, total, balance)
		}
	}
	// Historical orders without a failing job must not be backfilled or queued.
	exec("INSERT INTO sales_orders_v2(id) VALUES($1::uuid)", id(10))
	exec("INSERT INTO sales_order_lines_v2(sales_order_id,line_total,tax_amount) VALUES($1::uuid,17.83,NULL)", id(10))
	assertTotals(1, "0.00", "0.00", "0.00", "0.00")
	exec("UPDATE sales_orders_v2 SET paid_amount=5,balance_due=-5 WHERE id=$1::uuid", id(1))
	exec("INSERT INTO zatca_document_chain VALUES('sales_order_v2',$1)", id(2))
	exec("INSERT INTO invoices(sales_order_id) VALUES($1::uuid)", id(3))
	exec("UPDATE sales_orders_v2 SET subtotal=9,total_amount=9,balance_due=9 WHERE id=$1::uuid", id(4))
	exec("INSERT INTO sales_order_lines_v2(sales_order_id,line_total,tax_amount) VALUES($1::uuid,115,15)", id(5))
	assertTotals(5, "100.00", "15.00", "115.00", "115.00")
	exec("UPDATE sales_order_lines_v2 SET tax_amount=0 WHERE sales_order_id=$1::uuid", id(6))
	exec("UPDATE sales_order_lines_v2 SET line_total=115,tax_amount=15 WHERE sales_order_id=$1::uuid", id(7))
	exec("INSERT INTO pos_transactions(id,sales_order_id) VALUES(8,$1::uuid)", id(8))
	exec("INSERT INTO zatca_document_chain VALUES('pos_transaction','8')")
	if err := RepairNullTaxOrderTotals(ctx, pool); err != nil {
		t.Fatal(err)
	}
	assertTotals(1, "17.83", "0.00", "17.83", "12.83")
	assertTotals(2, "0.00", "0.00", "0.00", "0.00")
	assertTotals(3, "0.00", "0.00", "0.00", "0.00")
	assertTotals(4, "9.00", "0.00", "9.00", "9.00")
	assertTotals(5, "117.83", "15.00", "132.83", "132.83")
	assertTotals(6, "17.83", "0.00", "17.83", "17.83")
	assertTotals(7, "100.00", "15.00", "115.00", "115.00")
	assertTotals(8, "0.00", "0.00", "0.00", "0.00")
	assertTotals(10, "0.00", "0.00", "0.00", "0.00")
	var firstUpdate, secondUpdate time.Time
	if err := pool.QueryRow(ctx, "SELECT updated_at FROM sales_orders_v2 WHERE id=$1::uuid", id(1)).Scan(&firstUpdate); err != nil {
		t.Fatal(err)
	}
	if err := RepairNullTaxOrderTotals(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, "SELECT updated_at FROM sales_orders_v2 WHERE id=$1::uuid", id(1)).Scan(&secondUpdate); err != nil {
		t.Fatal(err)
	}
	if !firstUpdate.Equal(secondUpdate) {
		t.Fatal("repeated repair rewrote an already repaired order")
	}
	var stillNull bool
	if err := pool.QueryRow(ctx, "SELECT tax_amount IS NULL FROM sales_order_lines_v2 WHERE sales_order_id=$1::uuid", id(1)).Scan(&stillNull); err != nil || !stillNull {
		t.Fatal("repair changed line tax configuration", err)
	}
	// New inserts, updates and deletes now keep the header consistent.
	exec("INSERT INTO sales_orders_v2(id) VALUES($1::uuid)", id(9))
	exec("INSERT INTO sales_order_lines_v2(sales_order_id,line_total,tax_amount) VALUES($1::uuid,17.83,NULL)", id(9))
	assertTotals(9, "17.83", "0.00", "17.83", "17.83")
	exec("UPDATE sales_order_lines_v2 SET line_total=20 WHERE sales_order_id=$1::uuid", id(9))
	assertTotals(9, "20.00", "0.00", "20.00", "20.00")
	exec("DELETE FROM sales_order_lines_v2 WHERE sales_order_id=$1::uuid", id(9))
	assertTotals(9, "0.00", "0.00", "0.00", "0.00")
	// Both the local and cloud source queries must also include untaxed amounts.
	exec("INSERT INTO carts VALUES($1::uuid)", id(1))
	exec("INSERT INTO cart_items VALUES($1::uuid,17.83,NULL,NULL)", id(1))
	for _, file := range []string{filepath.Join("..", "..", "queries", "cart_order_queries.sql"), filepath.Join("..", "..", "..", "..", "apps", "cloud-server", "queries", "cart_order_queries.sql")} {
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"GetCartTotals", "GetOrderLineTotals"} {
			query := strings.SplitN(string(source), "-- name: "+name+" :one", 2)[1]
			query = strings.SplitN(query, ";", 2)[0]
			var subtotal, tax, total string
			err := pool.QueryRow(ctx, "SELECT subtotal::text,total_tax::text,total::text FROM ("+query+") t", id(1)).Scan(&subtotal, &tax, &total)
			if err != nil {
				t.Fatal(err)
			}
			if subtotal != "17.83" || tax != "0" || total != "17.83" {
				t.Fatalf("%s in %s returned (%s,%s,%s)", name, file, subtotal, tax, total)
			}
		}
	}
}
