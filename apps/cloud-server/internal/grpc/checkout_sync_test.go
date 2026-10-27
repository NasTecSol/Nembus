package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCheckoutInventoryExactlyOnce(t *testing.T) {
	dsn := os.Getenv("NEMBUS_SYNC_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set NEMBUS_SYNC_TEST_DATABASE_URL for PostgreSQL tests")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := fmt.Sprintf("checkout_test_%d", time.Now().UnixNano())
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
 CREATE TABLE sales_orders_v2(id uuid PRIMARY KEY,order_number text,store_id integer,fulfillment_status text);
 CREATE TABLE sales_order_lines_v2(id uuid PRIMARY KEY,sales_order_id uuid REFERENCES sales_orders_v2(id),
 product_id integer,product_variant_id integer,quantity_ordered numeric CHECK(quantity_ordered>0),quantity_fulfilled numeric,
 uom_id integer,batch_number text);
 CREATE TABLE inventory_stock(product_id integer,product_variant_id integer,store_id integer,
 quantity_on_hand numeric,quantity_available numeric,quantity_allocated numeric,updated_at timestamp);
 CREATE TABLE product_uom_conversions(product_id integer,from_uom_id integer,conversion_factor numeric);
 CREATE TABLE product_batches(product_id integer,product_variant_id integer,store_id integer,batch_number text,quantity_available numeric,updated_at timestamp);
 CREATE TABLE stock_movements(movement_type text,reference_type text,reference_id integer,product_id integer,product_variant_id integer,
 from_store_id integer,quantity numeric,uom_id integer,batch_number text,status text,metadata jsonb);
 CREATE TABLE stock_reservations(id integer,quantity_reserved numeric,reference_type text,reference_id text,product_id integer,
 product_variant_id integer,store_id integer,status text,updated_at timestamp);
 INSERT INTO inventory_stock VALUES(1,NULL,1,10,10,0,now());
 INSERT INTO product_batches VALUES(1,NULL,1,'batch-a',10,now());
 INSERT INTO product_uom_conversions VALUES(1,2,2);
 `)
	if err != nil {
		t.Fatal(err)
	}
	// Load the actual production fulfillment function and trigger unchanged.
	source, err := os.ReadFile("../../../../packages/core/db/views_functions/90_views_functions.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(source)
	start := strings.Index(sql, "CREATE OR REPLACE FUNCTION fn_trigger_deduct_inventory_on_fulfillment()")
	end := strings.Index(sql, "CREATE OR REPLACE FUNCTION fn_trigger_deduct_inventory_on_pos_transaction()")
	if start < 0 || end <= start {
		t.Fatal("fulfillment trigger source not found")
	}
	if _, err := pool.Exec(ctx, sql[start:end]); err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"order":{"id":"00000000-0000-0000-0000-000000000001","order_number":"TEST-1","store_id":1,"fulfillment_status":"fulfilled"},"lines":[{"id":"00000000-0000-0000-0000-000000000002","sales_order_id":"00000000-0000-0000-0000-000000000001","product_id":1,"quantity_ordered":1,"quantity_fulfilled":1,"uom_id":2,"batch_number":"batch-a"}]}`)
	server := &SyncServer{}
	apply := func(p []byte) error { return server.upsertEntityJSON(ctx, pool, "sales_orders_v2", "POS_CHECKOUT", p) }
	assertStock := func(want int, movements int) {
		t.Helper()
		var stock, batch, count int
		if err := pool.QueryRow(ctx, "SELECT quantity_on_hand FROM inventory_stock").Scan(&stock); err != nil {
			t.Fatal(err)
		}
		if err := pool.QueryRow(ctx, "SELECT quantity_available FROM product_batches").Scan(&batch); err != nil {
			t.Fatal(err)
		}
		if err := pool.QueryRow(ctx, "SELECT count(*) FROM stock_movements WHERE movement_type='sale'").Scan(&count); err != nil {
			t.Fatal(err)
		}
		if stock != want || batch != want || count != movements {
			t.Fatalf("stock=%d batch=%d movements=%d; want %d/%d", stock, batch, count, want, movements)
		}
	}
	// Interrupted checkout: a later invalid line rolls back the earlier line and header.
	var bad map[string]json.RawMessage
	if err := json.Unmarshal(payload, &bad); err != nil {
		t.Fatal(err)
	}
	bad["lines"] = json.RawMessage(`[{"id":"00000000-0000-0000-0000-000000000002","sales_order_id":"00000000-0000-0000-0000-000000000001","product_id":1,"quantity_ordered":1},{"id":"00000000-0000-0000-0000-000000000003","sales_order_id":"00000000-0000-0000-0000-000000000001","product_id":1,"quantity_ordered":-1}]`)
	invalid, _ := json.Marshal(bad)
	if err := apply(invalid); err == nil {
		t.Fatal("invalid line accepted")
	}
	assertStock(10, 0)
	var orderCount int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM sales_orders_v2").Scan(&orderCount); err != nil || orderCount != 0 {
		t.Fatalf("failed checkout persisted: %d %v", orderCount, err)
	}
	// Simultaneous first deliveries must perform one transition, not five.
	var wg sync.WaitGroup
	failures := make(chan error, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); failures <- apply(payload) }()
	}
	wg.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	assertStock(8, 1)
	if err := apply(payload); err != nil {
		t.Fatal(err)
	}
	assertStock(8, 1)
	// A delayed header event cannot reset the status and enable another deduction.
	if err := server.upsertEntityJSON(ctx, pool, "sales_orders_v2", "INSERT", []byte(`{"id":"00000000-0000-0000-0000-000000000001","order_number":"TEST-1","store_id":1,"fulfillment_status":"unfulfilled"}`)); err != nil {
		t.Fatal(err)
	}
	if err := apply(payload); err != nil {
		t.Fatal(err)
	}
	assertStock(8, 1)
	var status string
	if err := pool.QueryRow(ctx, "SELECT fulfillment_status FROM sales_orders_v2").Scan(&status); err != nil || status != "fulfilled" {
		t.Fatalf("fulfillment=%s err=%v", status, err)
	}
	// A trigger error after the inventory UPDATE rolls back stock, lines and header.
	fresh := []byte(strings.ReplaceAll(strings.ReplaceAll(string(payload), "00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000004"), "00000000-0000-0000-0000-000000000002", "00000000-0000-0000-0000-000000000005"))
	if _, err := pool.Exec(ctx, "ALTER TABLE stock_movements ADD CONSTRAINT reject_new_movement CHECK(false) NOT VALID"); err != nil {
		t.Fatal(err)
	}
	if err := apply(fresh); err == nil {
		t.Fatal("fulfillment failure was acknowledged")
	}
	assertStock(8, 1)
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM sales_orders_v2 WHERE id='00000000-0000-0000-0000-000000000004'").Scan(&orderCount); err != nil || orderCount != 0 {
		t.Fatalf("fulfillment failure persisted order: %d %v", orderCount, err)
	}
	if _, err := pool.Exec(ctx, "ALTER TABLE stock_movements DROP CONSTRAINT reject_new_movement"); err != nil {
		t.Fatal(err)
	}
	// Never acknowledge a sale if the existing database lacks its deduction trigger.
	if _, err := pool.Exec(ctx, "ALTER TABLE sales_orders_v2 DISABLE TRIGGER trg_deduct_inventory_on_fulfillment"); err != nil {
		t.Fatal(err)
	}
	if err := apply(fresh); err == nil {
		t.Fatal("checkout accepted without inventory trigger")
	}
	assertStock(8, 1)
}
