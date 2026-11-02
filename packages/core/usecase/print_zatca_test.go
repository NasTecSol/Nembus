package usecase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/printing"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPrintReceiptDoesNotBlockOnZatcaReportingPostgres(t *testing.T) {
	dsn := os.Getenv("NEMBUS_ZATCA_PRINT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set NEMBUS_ZATCA_PRINT_TEST_DATABASE_URL for isolated receipt test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("zatca_print_usecase_%d", time.Now().UnixNano())
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
	_, err = pool.Exec(ctx, `CREATE TABLE organizations(id int PRIMARY KEY,name text,code text,legal_name text,tax_id text,currency_code text,fiscal_year_variant text,is_active bool,metadata jsonb,created_at timestamp,updated_at timestamp);
CREATE TABLE sales_orders_v2(id uuid PRIMARY KEY,organization_id int);
CREATE TABLE zatca_document_chain(id bigserial PRIMARY KEY,entity_type text,entity_id text,organization_id int,zatca_status text,zatca_response jsonb,qr_code_base64 text);
INSERT INTO organizations(id,name,code,metadata) VALUES(1,'Test Store','TEST','{"zatca":{"enabled":true}}');`)
	if err != nil {
		t.Fatal(err)
	}
	reportedID, pendingID, failedID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	for _, orderID := range []string{reportedID, pendingID, failedID} {
		if _, err := pool.Exec(ctx, `INSERT INTO sales_orders_v2(id,organization_id) VALUES($1,1)`, orderID); err != nil {
			t.Fatal(err)
		}
	}
	for _, row := range []struct{ id, status string }{{reportedID, "reported"}, {pendingID, "pending"}, {failedID, "failed"}} {
		if _, err := pool.Exec(ctx, `INSERT INTO zatca_document_chain(entity_type,entity_id,organization_id,zatca_status,zatca_response,qr_code_base64) VALUES('sales_order_v2',$1,1,$2,'{"reportingStatus":"REPORTED"}','signed-qr')`, row.id, row.status); err != nil {
			t.Fatal(err)
		}
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	printed := make(chan []byte, 3)
	go func() {
		for range 3 {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			data, _ := io.ReadAll(conn)
			conn.Close()
			printed <- data
		}
	}()
	uc := NewPrintUseCase()
	uc.SetRepository(repository.New(pool))
	port := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	makeInput := func(orderID string) PrintReceiptInput {
		return PrintReceiptInput{OrgID: 1,
			Printer: printing.PrinterConfig{Mode: "network", NetworkIP: "127.0.0.1", NetworkPort: port},
			Receipt: printing.ReceiptData{ReceiptNumber: orderID, SalesOrderID: orderID, QRCodeBase64: "browser-forgery"}}
	}
	for _, orderID := range []string{pendingID, failedID} {
		result := uc.PrintReceipt(ctx, makeInput(orderID))
		if result.StatusCode != 200 || result.Data.(map[string]interface{})["qr_status"] != "not_reported" {
			t.Fatalf("pending or failed ZATCA reporting blocked the receipt: %d", result.StatusCode)
		}
		select {
		case data := <-printed:
			if bytes.Contains(data, []byte("signed-qr")) || bytes.Contains(data, []byte("browser-forgery")) || bytes.Contains(data, []byte{printing.GS, '(', 'k'}) {
				t.Fatal("unreported receipt printed a QR")
			}
		case <-ctx.Done():
			t.Fatal("unreported receipt did not reach the printer")
		}
	}
	if result := uc.PrintReceipt(ctx, makeInput(reportedID)); result.StatusCode != 200 || result.Data.(map[string]interface{})["qr_status"] != "included" {
		t.Fatalf("reported receipt was not printed: %d %s", result.StatusCode, result.Message)
	}
	select {
	case data := <-printed:
		if !bytes.Contains(data, []byte("signed-qr")) || bytes.Contains(data, []byte("browser-forgery")) {
			t.Fatal("printer did not receive the saved ZATCA QR exclusively")
		}
	case <-ctx.Done():
		t.Fatal("printer did not receive receipt bytes")
	}
}
