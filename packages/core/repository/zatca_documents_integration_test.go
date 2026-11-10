package repository

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
)

// Temporary tables isolate these fixtures from application data. Use a dedicated
// PostgreSQL instance; no schema migrations or application credentials are needed.
func TestZatcaDocumentLocationsPostgres(t *testing.T) {
	dsn := os.Getenv("NEMBUS_ZATCA_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set NEMBUS_ZATCA_TEST_DATABASE_URL to run PostgreSQL document filter tests")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `
CREATE TEMP TABLE stores (id int PRIMARY KEY, organization_id int);
CREATE TEMP TABLE pos_terminals (id int PRIMARY KEY, store_id int);
CREATE TEMP TABLE cashier_sessions (id int PRIMARY KEY, pos_terminal_id int);
CREATE TEMP TABLE sales_orders_v2 (id uuid PRIMARY KEY, organization_id int, store_id int, pos_terminal_id int);
CREATE TEMP TABLE pos_transactions (id int PRIMARY KEY, store_id int, pos_terminal_id int, cashier_session_id int);
CREATE TEMP TABLE invoices (id uuid PRIMARY KEY, organization_id int, store_id int, sales_order_id uuid);
CREATE TEMP TABLE sales_returns (id int PRIMARY KEY, store_id int, cashier_session_id int);
CREATE TEMP TABLE zatca_document_chain (
 id bigint PRIMARY KEY, entity_type text, entity_id text, device_config_id int DEFAULT 99,
 organization_id int DEFAULT 7, zatca_uuid uuid DEFAULT '00000000-0000-4000-8000-000000000099',
 icv bigint DEFAULT 1, pih text DEFAULT 'pih', xml_hash text DEFAULT 'hash',
 zatca_status text DEFAULT 'reported', zatca_response jsonb DEFAULT '{"reportingStatus":"REPORTED"}',
 qr_code_base64 text, signed_xml text, submitted_at timestamptz, cleared_at timestamptz,
 created_at timestamp, updated_at timestamp
);
INSERT INTO stores VALUES (3,7),(4,7),(5,9);
INSERT INTO pos_terminals VALUES (8,3),(9,3),(10,4),(11,5);
INSERT INTO cashier_sessions VALUES (20,8),(21,9),(22,10);
INSERT INTO sales_orders_v2 VALUES
 ('00000000-0000-4000-8000-000000000001',7,3,8),
 ('00000000-0000-4000-8000-000000000002',7,4,10),
 ('00000000-0000-4000-8000-000000000003',7,3,NULL),
 ('00000000-0000-4000-8000-000000000004',9,5,11);
INSERT INTO pos_transactions VALUES (1,3,9,21),(2,3,NULL,20),(3,5,11,NULL);
INSERT INTO invoices VALUES
 ('00000000-0000-4000-8000-000000000001',7,3,'00000000-0000-4000-8000-000000000001'),
 ('00000000-0000-4000-8000-000000000002',7,4,NULL);
-- Return terminal is where the return was processed, not the original sale.
INSERT INTO sales_returns VALUES (1,3,21),(2,4,NULL),(3,3,22);
INSERT INTO zatca_document_chain (id,entity_type,entity_id) VALUES
 (1,'sales_order_v2','00000000-0000-4000-8000-000000000001'),
 (2,'sales_order_v2','00000000-0000-4000-8000-000000000002'),
 (3,'pos_transaction','1'),(4,'pos_transaction','2'),
 (5,'invoice','00000000-0000-4000-8000-000000000001'),
 (6,'invoice','00000000-0000-4000-8000-000000000002'),
 (7,'sales_return','1'),(8,'sales_return','2'),
 (9,'sales_order_v2','00000000-0000-4000-8000-000000000003'),
 (10,'pos_transaction','deleted-or-invalid'),(11,'unknown','1'),
 (12,'sales_return','3'),
 (13,'sales_order_v2','00000000-0000-4000-8000-000000000004'),
 (14,'pos_transaction','3');
INSERT INTO zatca_document_chain (id,organization_id,entity_type,entity_id)
 VALUES (15,9,'sales_order_v2','00000000-0000-4000-8000-000000000004');
`)
	if err != nil {
		t.Fatal(err)
	}
	repo := New(tx)
	for _, tc := range []struct {
		name                 string
		org, store, terminal int32
		limit                int32
		offset, total        int64
		ids                  []int64
	}{
		{"all", 7, 0, 0, 50, 0, 14, []int64{14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}},
		{"store", 7, 3, 0, 50, 0, 7, []int64{12, 9, 7, 5, 4, 3, 1}},
		{"terminal", 7, 0, 8, 50, 0, 3, []int64{5, 4, 1}},
		{"both", 7, 3, 9, 50, 0, 2, []int64{7, 3}},
		{"other store same device", 7, 4, 0, 50, 0, 3, []int64{8, 6, 2}},
		{"mismatch", 7, 4, 8, 50, 0, 0, []int64{}},
		{"foreign store", 7, 5, 0, 50, 0, 0, []int64{}},
		{"foreign terminal", 7, 0, 11, 50, 0, 0, []int64{}},
		{"other organization", 9, 5, 11, 50, 0, 1, []int64{15}},
		{"page two", 7, 0, 8, 1, 1, 3, []int64{4}},
		{"out of range", 7, 0, 8, 1, 8, 3, []int64{}},
		{"not found", 7, 999, 0, 50, 0, 0, []int64{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rows, total, err := repo.ListZatcaDocumentsPage(ctx, tc.org, tc.limit, tc.offset, ZatcaDocumentFilters{StoreID: tc.store, PosTerminalID: tc.terminal})
			if err != nil {
				t.Fatal(err)
			}
			ids := make([]int64, 0, len(rows))
			for _, row := range rows {
				ids = append(ids, row.ID)
				if tc.store != 0 && (!row.StoreID.Valid || row.StoreID.Int32 != tc.store) {
					t.Fatalf("incorrect store: %+v", row)
				}
				if tc.terminal != 0 && (!row.PosTerminalID.Valid || row.PosTerminalID.Int32 != tc.terminal) {
					t.Fatalf("incorrect terminal: %+v", row)
				}
				if row.OrganizationID != tc.org || row.ZatcaStatus.ZatcaDocStatus != ZatcaDocStatusReported {
					t.Fatalf("saved document changed: %+v", row)
				}
				if (row.ID == 10 || row.ID == 11 || row.ID == 13 || row.ID == 14) && (row.StoreID.Valid || row.PosTerminalID.Valid) {
					t.Fatalf("unknown source inherited location: %+v", row)
				}
				if row.ID == 12 && row.PosTerminalID.Valid {
					t.Fatal("return inherited a terminal from a different store")
				}
			}
			if total != tc.total || !reflect.DeepEqual(ids, tc.ids) {
				t.Fatalf("got IDs %v total %d, want %v total %d", ids, total, tc.ids, tc.total)
			}
		})
	}
	if _, _, err := repo.ListZatcaDocumentsPage(ctx, 0, 50, 0, ZatcaDocumentFilters{}); err == nil {
		t.Fatal("missing organization accepted")
	}
}
