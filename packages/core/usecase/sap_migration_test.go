package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-sap/contracts"
	"github.com/NasTecSol/nembus-sap/mappings"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestSAPMigrationStockMovements(t *testing.T) {
	ctx := context.Background()
	connStr := "postgres://postgres:root1234@localhost:5432/nembus_migration_audit_postfix_20260929?sslmode=disable"
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to db: %v", err)
	}
	defer pool.Close()

	// 1. Setup a dummy store and product
	// Apply schema changes for test
	_, err = pool.Exec(ctx, `ALTER TABLE stock_movements ADD COLUMN IF NOT EXISTS reference_uuid uuid`)
	if err != nil {
		t.Fatalf("Failed to alter table: %v", err)
	}

	var storeID int32
	err = pool.QueryRow(ctx, `
		INSERT INTO stores (organization_id, code, name, is_active, metadata)
		VALUES (1, $1, 'Test Store', true, '{}'::jsonb) RETURNING id
	`, fmt.Sprintf("ST-%d", time.Now().Unix())).Scan(&storeID)
	if err != nil {
		t.Fatalf("Failed to insert store: %v", err)
	}

	var productID int32
	err = pool.QueryRow(ctx, `
		INSERT INTO products (organization_id, sku, name, is_active)
		VALUES (1, $1, 'Test Prod', true) RETURNING id
	`, fmt.Sprintf("SKU-%d", time.Now().Unix())).Scan(&productID)
	if err != nil {
		t.Fatalf("Failed to insert product: %v", err)
	}

	var customerID int32
	custCode := fmt.Sprintf("CUST-%d", time.Now().Unix())
	err = pool.QueryRow(ctx, `
		INSERT INTO customers (organization_id, customer_code, name, is_active)
		VALUES (1, $1, 'Test Cust', true) RETURNING id
	`, custCode).Scan(&customerID)
	if err != nil {
		t.Fatalf("Failed to insert customer: %v", err)
	}

	uc := NewSAPMigrationUseCase(pool)

	// 2. Ingest an Invoice
	invDocNum := fmt.Sprintf("TEST-INV-%d", time.Now().Unix())
	batchInv := &contracts.MigrationBatchPayload{
		Domain:         contracts.DomainInvoices,
		OrganizationID: 1,
		Invoices: []mappings.CanonicalInvoice{
			{
				InvoiceNumber: invDocNum,
				InvoiceType:   "standard", 
				InvoiceStatus: "paid",
				InvoiceDate:   time.Now(),
				CustomerCode:  custCode,
				Lines: []mappings.CanonicalInvoiceLine{
					{
						ProductSKU: fmt.Sprintf("SKU-%d", time.Now().Unix()), // SKU is checked from db
						Quantity:   1,
						UnitPrice:  100,
						LineTotal:  100,
					},
				},
				Metadata: map[string]interface{}{
					"sap_doc_num": invDocNum,
				},
			},
		},
	}
	resInv, err := uc.IngestBatch(ctx, 1, batchInv)
	if err != nil {
		t.Fatalf("Failed to ingest invoice: %v", err)
	}
	if resInv.RecordsStaged == 0 {
		t.Fatalf("Invoice not staged! Errors: %v", resInv.Errors)
	}

	// Wait for ingestion commit if any background things happen (it's synchronous though)

	// Fetch the generated Invoice UUID
	var invoiceID string
	err = pool.QueryRow(ctx, `SELECT id::text FROM invoices WHERE metadata->>'sap_doc_num' = $1 LIMIT 1`, invDocNum).Scan(&invoiceID)
	if err != nil {
		t.Fatalf("Failed to fetch ingested invoice UUID: %v", err)
	}
	t.Logf("Ingested Invoice UUID: %s", invoiceID)

	// 3. Ingest Stock Movement referencing the Invoice
	transNum := fmt.Sprintf("TEST-TRANS-%d", time.Now().Unix())
	batchSm := &contracts.MigrationBatchPayload{
		Domain:         contracts.DomainStockMovements,
		OrganizationID: 1,
		StockMovements: []mappings.CanonicalStockMovement{
			{
				MovementType:    "sale",
				ReferenceType:   "invoice",
				ReferenceNumber: invDocNum, // Matches the invoice DocNum
				ProductSKU:      fmt.Sprintf("SKU-%d", time.Now().Unix()), // Will fail to link product if we use new SKU? Yes, let's use the one we created!
				FromStoreCode:   fmt.Sprintf("ST-%d", time.Now().Unix()),
				Quantity:        -1,
				MovementDate:    time.Now(),
				Metadata: map[string]interface{}{
					"sap_trans_num": transNum,
					"sap_base_ref":  invDocNum,
				},
			},
		},
	}
	// Fix SKU and StoreCode to match what we created
	var createdSKU, createdStoreCode string
	pool.QueryRow(ctx, `SELECT sku FROM products WHERE id = $1`, productID).Scan(&createdSKU)
	pool.QueryRow(ctx, `SELECT code FROM stores WHERE id = $1`, storeID).Scan(&createdStoreCode)
	
	batchSm.StockMovements[0].ProductSKU = createdSKU
	batchSm.StockMovements[0].FromStoreCode = createdStoreCode

	res, err := uc.IngestBatch(ctx, 1, batchSm)
	if err != nil {
		t.Fatalf("Failed to ingest stock movement: %v", err)
	}
	if res.RecordsStaged == 0 {
		t.Fatalf("Expected at least 1 staged stock movement, got 0. Errors: %v", res.Errors)
	}

	// 4. Verify that the stock movement has the correct reference_uuid
	var refUUID *string
	var refID *int64
	err = pool.QueryRow(ctx, `SELECT reference_uuid::text, reference_id FROM stock_movements WHERE metadata->>'sap_trans_num' = $1`, transNum).Scan(&refUUID, &refID)
	if err != nil {
		t.Fatalf("Failed to fetch inserted stock movement: %v", err)
	}

	if refUUID == nil {
		t.Fatalf("Expected reference_uuid to be populated, got nil (reference_id: %v)", refID)
	}
	if *refUUID != invoiceID {
		t.Fatalf("Expected reference_uuid to be %s, got %s", invoiceID, *refUUID)
	}

	t.Logf("Success! Stock movement reference_uuid perfectly matches Invoice ID: %s", *refUUID)
}
