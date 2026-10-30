package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestReportedSalesOrderQRPostgres(t *testing.T) {
	dsn := os.Getenv("NEMBUS_ZATCA_PRINT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set NEMBUS_ZATCA_PRINT_TEST_DATABASE_URL for isolated receipt QR test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("zatca_print_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize()); err != nil {
		t.Fatal(err)
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		_, _ = admin.Exec(context.Background(), "DROP SCHEMA "+pgx.Identifier{schema}.Sanitize()+" CASCADE")
		admin.Close()
	})
	_, err = pool.Exec(ctx, `CREATE TABLE sales_orders_v2(id uuid PRIMARY KEY,organization_id int NOT NULL);
CREATE TABLE zatca_document_chain(id bigserial PRIMARY KEY,entity_type text,entity_id text,organization_id int,zatca_status text,zatca_response jsonb,qr_code_base64 text);`)
	if err != nil {
		t.Fatal(err)
	}
	orderID := uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO sales_orders_v2(id,organization_id) VALUES($1,1)`, orderID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO zatca_document_chain(entity_type,entity_id,organization_id,zatca_status,zatca_response,qr_code_base64)
 VALUES('sales_order_v2',$1,1,'pending','{"reportingStatus":"REPORTED"}','signed-qr')`, orderID); err != nil {
		t.Fatal(err)
	}
	repo := New(pool)
	check := func(org int32, want string) {
		t.Helper()
		got, err := repo.GetReportedSalesOrderQR(ctx, org, orderID)
		if err != nil || got != want {
			t.Fatalf("org=%d qr=%q want=%q err=%v", org, got, want, err)
		}
	}
	check(1, "") // A locally signed QR cannot print before reporting.
	if _, err := pool.Exec(ctx, `UPDATE zatca_document_chain SET zatca_status='reported',zatca_response='{"reportingStatus":"NOT_REPORTED"}'`); err != nil {
		t.Fatal(err)
	}
	check(1, "")
	if _, err := pool.Exec(ctx, `UPDATE zatca_document_chain SET zatca_response='{"reportingStatus":"REPORTED"}'`); err != nil {
		t.Fatal(err)
	}
	check(1, "signed-qr")
	if _, err := repo.GetReportedSalesOrderQR(ctx, 2, orderID); err != pgx.ErrNoRows {
		t.Fatalf("another organization accessed the QR: %v", err)
	}
}
