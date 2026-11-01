package usecase

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestPriceChangeWorkflow_ApproveAppliesNewPrice(t *testing.T) {
	acc := NewCatalogPriceAccommodator()
	newPrice := 79.99
	payload := CatalogPriceChangePayload{
		ProductID:   10,
		ProductSKU:  "SKU-1",
		PriceListID: 1,
		NewPrice:    &newPrice,
	}
	raw, _ := json.Marshal(payload)
	req := repository.ApprovalRequest{Payload: raw}

	mockDB := &workflowMockDBTX{catalogPriceUpserts: 0}
	repo := repository.New(mockDB)
	if err := acc.Accommodate(context.Background(), repo, req, 99); err != nil {
		t.Fatalf("accommodate: %v", err)
	}
	if mockDB.catalogPriceUpserts != 1 {
		t.Fatalf("expected 1 price upsert, got %d", mockDB.catalogPriceUpserts)
	}
}

func TestStockHoldWorkflow_ApproveSetsReadyForPost(t *testing.T) {
	acc := NewSAPInvoiceStockAccommodator()
	payload := SAPInvoiceStockPayload{TransactionID: 42, TransactionNumber: "TX-42"}
	raw, _ := json.Marshal(payload)
	req := repository.ApprovalRequest{ID: 1, Payload: raw}

	mockDB := &stockHoldMockDB{existingMeta: json.RawMessage(`{"cashier_id":5,"sap_sync_state":"held_stock"}`)}
	repo := repository.New(mockDB)
	if err := acc.Accommodate(context.Background(), repo, req, 7); err != nil {
		t.Fatalf("accommodate: %v", err)
	}
	if len(mockDB.updatedPosMetadata) != 1 {
		t.Fatalf("expected metadata update, got %d", len(mockDB.updatedPosMetadata))
	}
	var merged map[string]any
	if err := json.Unmarshal(mockDB.updatedPosMetadata[0].Metadata, &merged); err != nil {
		t.Fatal(err)
	}
	if merged["sap_sync_state"] != "ready_for_sap_post" {
		t.Fatalf("expected ready_for_sap_post, got %v", merged["sap_sync_state"])
	}
	if merged["cashier_id"] != float64(5) {
		t.Fatalf("expected merged cashier_id preserved, got %v", merged["cashier_id"])
	}
}

func TestStockHoldWorkflow_RejectSetsForcePost(t *testing.T) {
	acc := NewSAPInvoiceStockAccommodator()
	payload := SAPInvoiceStockPayload{TransactionID: 42}
	raw, _ := json.Marshal(payload)
	req := repository.ApprovalRequest{ID: 1, Payload: raw}

	mockDB := &stockHoldMockDB{existingMeta: json.RawMessage(`{}`)}
	repo := repository.New(mockDB)
	if err := acc.Reject(context.Background(), repo, req, "adjust and post"); err != nil {
		t.Fatalf("reject: %v", err)
	}
	var merged map[string]any
	_ = json.Unmarshal(mockDB.updatedPosMetadata[0].Metadata, &merged)
	if merged["sap_sync_state"] != "ready_for_sap_post" || merged["sap_force_post"] != true {
		t.Fatalf("expected force post queue metadata, got %v", merged)
	}
}

type stockHoldMockDB struct {
	existingMeta       json.RawMessage
	updatedPosMetadata []repository.UpdatePosTransactionMetadataParams
}

func (m *stockHoldMockDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	if strings.Contains(sql, "UPDATE pos_transactions") && strings.Contains(sql, "metadata") {
		if len(args) >= 2 {
			id := args[0].(int32)
			meta, _ := args[1].(json.RawMessage)
			m.updatedPosMetadata = append(m.updatedPosMetadata, repository.UpdatePosTransactionMetadataParams{
				ID: id, Metadata: meta,
			})
		}
	}
	if strings.Contains(sql, "INSERT INTO stock_movements") {
		return pgconn.NewCommandTag("INSERT 0 1"), nil
	}
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func (m *stockHoldMockDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (m *stockHoldMockDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return &stockHoldMockRow{db: m, sql: sql}
}

func (m *stockHoldMockDB) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

type stockHoldMockRow struct {
	db  *stockHoldMockDB
	sql string
}

func (r *stockHoldMockRow) Scan(dest ...any) error {
	if !strings.Contains(r.sql, "GetPosTransaction") {
		return nil
	}
	for i, d := range dest {
		switch ptr := d.(type) {
		case *int32:
			if i == 0 {
				*ptr = 42
			}
		case *string:
			if i == 6 {
				*ptr = "TX-42"
			}
		case *time.Time:
			if i == 7 {
				*ptr = time.Now()
			}
		case *json.RawMessage:
			*ptr = append(json.RawMessage(nil), r.db.existingMeta...)
		case *pgtype.Int4, *pgtype.Text, *pgtype.Timestamp, *pgtype.Numeric, *pgtype.UUID:
			// leave zero values
		}
	}
	return nil
}
