package usecase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
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
CREATE TABLE stores(id int PRIMARY KEY,organization_id int);
CREATE TABLE sales_orders_v2(id uuid PRIMARY KEY,organization_id int,store_id int,pos_terminal_id int,order_number text);
CREATE TABLE pos_transactions(id int PRIMARY KEY,sales_order_id uuid,store_id int,pos_terminal_id int,transaction_number text);
CREATE TABLE zatca_device_configs(organization_id int,store_id int,pos_terminal_id int,device_type text,is_active bool,is_revoked bool);
CREATE TABLE zatca_document_chain(id bigserial PRIMARY KEY,entity_type text,entity_id text,organization_id int,zatca_status text,zatca_response jsonb,qr_code_base64 text);
INSERT INTO organizations(id,name,code,metadata) VALUES
 (1,'Test Store','TEST','{"zatca":{"enabled":true}}'),
 (2,'Disabled','OFF','{"zatca":{"enabled":false}}'),
 (3,'Other Organization','OTHER','{"zatca":{"enabled":true}}');
INSERT INTO stores VALUES(1,1),(2,1),(3,3);
INSERT INTO zatca_device_configs VALUES
 (1,1,8,'pos',true,false),
 (1,1,10,'pos',true,true),
 (1,1,11,'pos',false,false),
 (1,1,NULL,'cloud',true,false),
 (3,2,98,'pos',true,false);`)
	if err != nil {
		t.Fatal(err)
	}
	addOrder := func(store int, terminal any, status, qr string) string {
		t.Helper()
		id := uuid.NewString()
		if _, err := pool.Exec(ctx, `INSERT INTO sales_orders_v2 VALUES($1,1,$2,$3,$4)`, id, store, terminal, "SO-"+id); err != nil {
			t.Fatal(err)
		}
		if status != "" {
			if _, err := pool.Exec(ctx, `INSERT INTO zatca_document_chain(entity_type,entity_id,organization_id,zatca_status,zatca_response,qr_code_base64)
 VALUES('sales_order_v2',$1,1,$2,'{"reportingStatus":"REPORTED"}',$3)`, id, status, qr); err != nil {
				t.Fatal(err)
			}
		}
		return id
	}
	reportedID := addOrder(1, 8, "reported", "signed-qr")
	pendingID := addOrder(1, 8, "pending", "signed-qr")
	failedID := addOrder(1, 8, "failed", "signed-qr")
	noRegistrationID := addOrder(1, 99, "", "")
	wrongStoreID := addOrder(2, 8, "", "")
	wrongOrgDeviceID := addOrder(2, 98, "", "")
	noTerminalID := addOrder(1, nil, "", "")
	revokedID := addOrder(1, 10, "", "")
	inactiveID := addOrder(1, 11, "", "")
	cloudOnlyID := addOrder(1, 12, "", "")
	unsignedID := addOrder(1, 8, "", "")
	transactionRegisteredID := addOrder(1, 99, "", "")
	transactionUnregisteredID := addOrder(1, 8, "", "")
	if _, err := pool.Exec(ctx, `INSERT INTO pos_transactions VALUES
      (4,$1,1,8,'TXN-ACTUAL-REGISTERED'),(5,$2,2,99,'TXN-ACTUAL-UNREGISTERED')`, transactionRegisteredID, transactionUnregisteredID); err != nil {
		t.Fatal(err)
	}
	historicalID := addOrder(1, 10, "reported", "historical-qr")
	deletedDeviceID := addOrder(1, 20, "reported", "historical-qr")
	missingQRID := addOrder(1, 10, "reported", "")
	oversizedQRID := addOrder(1, 10, "reported", strings.Repeat("x", 2001))
	if _, err := pool.Exec(ctx, `INSERT INTO pos_transactions VALUES
 (1,NULL,1,99,'TXN-UNREGISTERED'),(2,$1,1,8,'TXN-PENDING'),(3,NULL,1,10,'TXN-LEGACY')`, pendingID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO zatca_document_chain(entity_type,entity_id,organization_id,zatca_status,zatca_response,qr_code_base64)
 VALUES('pos_transaction','3',1,'reported','{"reportingStatus":"REPORTED"}','legacy-qr')`); err != nil {
		t.Fatal(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	printed := make(chan []byte)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			data, _ := io.ReadAll(conn)
			conn.Close()
			select {
			case printed <- data:
			case <-ctx.Done():
				return
			}
		}
	}()
	uc := NewPrintUseCase()
	uc.SetRepository(repository.New(pool))
	port := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	for _, test := range []struct {
		name, order, number, receiptType, want, qr string
		org                                        int32
	}{
		{name: "pending", order: pendingID, want: "not_reported"},
		{name: "failed", order: failedID, want: "not_reported"},
		{name: "registered without document", order: unsignedID, want: "not_reported"},
		{name: "actual transaction registered", order: transactionRegisteredID, want: "not_reported"},
		{name: "actual transaction unregistered", order: transactionUnregisteredID, want: "not_applicable"},
		{name: "reported", order: reportedID, want: "included", qr: "signed-qr"},
		{name: "unregistered terminal", order: noRegistrationID, want: "not_applicable"},
		{name: "different store", order: wrongStoreID, want: "not_applicable"},
		{name: "different organization registration", order: wrongOrgDeviceID, want: "not_applicable"},
		{name: "missing terminal", order: noTerminalID, want: "not_applicable"},
		{name: "revoked registration", order: revokedID, want: "not_applicable"},
		{name: "inactive registration", order: inactiveID, want: "not_applicable"},
		{name: "cloud registration only", order: cloudOnlyID, want: "not_applicable"},
		{name: "reported before revocation", order: historicalID, want: "included", qr: "historical-qr"},
		{name: "reported before device deletion", order: deletedDeviceID, want: "included", qr: "historical-qr"},
		{name: "reported without saved QR", order: missingQRID, want: "unavailable"},
		{name: "oversized saved QR", order: oversizedQRID, want: "unavailable"},
		{name: "invalid explicit identifier", order: "invalid", number: reportedID, want: "unavailable"},
		{name: "sale not found", order: uuid.NewString(), want: "unavailable"},
		{name: "cross organization sale", order: reportedID, org: 3, want: "unavailable"},
		{name: "organization disabled", order: reportedID, org: 2, want: "not_applicable"},
		{name: "Z report", receiptType: "Z-REPORT", want: "not_applicable"},
		{name: "session summary", receiptType: "SESSION SUMMARY", want: "not_applicable"},
		{name: "order UUID fallback", number: reportedID, want: "included", qr: "signed-qr"},
		{name: "order number fallback", number: "SO-" + reportedID, want: "included", qr: "signed-qr"},
		{name: "standalone unregistered POS transaction", number: "TXN-UNREGISTERED", want: "not_applicable"},
		{name: "transaction linked to pending order", number: "TXN-PENDING", want: "not_reported"},
		{name: "reported legacy POS transaction", number: "TXN-LEGACY", want: "included", qr: "legacy-qr"},
	} {
		t.Run(test.name, func(t *testing.T) {
			org := test.org
			if org == 0 {
				org = 1
			}
			result := uc.PrintReceipt(ctx, PrintReceiptInput{OrgID: org,
				Printer: printing.PrinterConfig{Mode: "network", NetworkIP: "127.0.0.1", NetworkPort: port},
				Receipt: printing.ReceiptData{Type: test.receiptType, ReceiptNumber: test.number, SalesOrderID: test.order, QRCodeBase64: "browser-forgery"}})
			if result.StatusCode != 200 {
				t.Fatalf("printing blocked: %d %s", result.StatusCode, result.Message)
			}
			data := result.Data.(map[string]interface{})
			if data["qr_status"] != test.want || data["qr_code_base64"] != test.qr {
				t.Fatalf("status=%v QR=%v, want status=%s QR=%s", data["qr_status"], data["qr_code_base64"], test.want, test.qr)
			}
			if test.want == "not_applicable" && result.Message != "receipt printed successfully" {
				t.Fatalf("unexpected ZATCA warning: %s", result.Message)
			}
			select {
			case bytesSent := <-printed:
				if bytes.Contains(bytesSent, []byte("browser-forgery")) {
					t.Fatal("caller QR reached printer")
				}
				if test.qr != "" {
					if !bytes.Contains(bytesSent, []byte(test.qr)) {
						t.Fatal("saved QR missing from receipt bytes")
					}
				} else if bytes.Contains(bytesSent, []byte{printing.GS, '(', 'k'}) {
					t.Fatal("QR printed for a receipt without a reported QR")
				}
			case <-ctx.Done():
				t.Fatal("receipt bytes did not reach printer")
			}
		})
	}
}
