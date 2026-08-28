package usecase

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// This is a source-level regression test because SAP migration tests have no
// live database seam. It locks the actual product upsert contract: INSERT
// keeps new products incomplete, while ON CONFLICT preserves only absent
// brand/description values from the existing organization-local row.
func TestSAPProductUpsertPreservesAbsentEnrichmentValues(t *testing.T) {
	query := strings.ToLower(sapProductUpsertQuery)

	for _, fragment := range []string{
		"on conflict(organization_id, sku) do update set",
		"description = case",
		"when nullif(btrim(excluded.description), '') is not null then excluded.description",
		"else products.description",
		"brand_id = case",
		"when excluded.brand_id is not null then excluded.brand_id",
		"else products.brand_id",
		"category_id = excluded.category_id",
		"product_type = excluded.product_type",
		"is_active = excluded.is_active",
		"is_sellable = excluded.is_sellable",
		"track_inventory = excluded.track_inventory",
		"$1, $2, $3, $4",
		"(select id from brands where code = $6)",
	} {
		if !strings.Contains(query, fragment) {
			t.Errorf("SAP product upsert is missing contract fragment %q", fragment)
		}
	}

	for _, forbidden := range []string{
		"description = excluded.description,",
		"brand_id = excluded.brand_id,",
	} {
		if strings.Contains(query, forbidden) {
			t.Errorf("SAP product upsert still unconditionally overwrites %s", forbidden)
		}
	}
}

type barcodeTestRow struct {
	scan func(...any) error
}

func (r barcodeTestRow) Scan(dest ...any) error {
	return r.scan(dest...)
}

// barcodeTestTx is deliberately small: persistSAPBarcode only needs QueryRow
// and Exec, while embedding pgx.Tx keeps the test independent of a database.
type barcodeTestTx struct {
	pgx.Tx

	targetID        int
	targetErr       error
	ownerProductID  int
	ownerOrgID      int
	ownerSKU        string
	ownerErr        error
	upsertAffected  int64
	upsertCount     int
	ownerQueryCount int
}

func (tx *barcodeTestTx) QueryRow(_ context.Context, query string, _ ...any) pgx.Row {
	if strings.Contains(strings.ToLower(query), "where pb.barcode") {
		tx.ownerQueryCount++
		return barcodeTestRow{scan: func(dest ...any) error {
			if tx.ownerErr != nil {
				return tx.ownerErr
			}
			*dest[0].(*int) = tx.ownerProductID
			*dest[1].(*int) = tx.ownerOrgID
			*dest[2].(*string) = tx.ownerSKU
			return nil
		}}
	}

	return barcodeTestRow{scan: func(dest ...any) error {
		if tx.targetErr != nil {
			return tx.targetErr
		}
		*dest[0].(*int) = tx.targetID
		return nil
	}}
}

func (tx *barcodeTestTx) Exec(_ context.Context, query string, _ ...any) (pgconn.CommandTag, error) {
	if strings.Contains(strings.ToLower(query), "insert into product_barcodes") {
		tx.upsertCount++
		return pgconn.NewCommandTag(fmt.Sprintf("INSERT 0 %d", tx.upsertAffected)), nil
	}
	return pgconn.NewCommandTag("OK"), nil
}

func TestPersistSAPBarcodeAllowsMultipleDistinctBarcodes(t *testing.T) {
	tx := &barcodeTestTx{targetID: 7, upsertAffected: 1}
	for _, barcode := range []string{"111", "222", "333"} {
		if err := persistSAPBarcode(context.Background(), tx, 42, barcode, "EAN13", false, "SKU-A"); err != nil {
			t.Fatalf("persistSAPBarcode(%q) error: %v", barcode, err)
		}
	}
	if tx.upsertCount != 3 {
		t.Fatalf("upsert count = %d, want 3 distinct barcode writes", tx.upsertCount)
	}
}

func TestPersistSAPBarcodeRepeatedSameProductIsIdempotent(t *testing.T) {
	tx := &barcodeTestTx{targetID: 7, upsertAffected: 1}
	for range 2 {
		if err := persistSAPBarcode(context.Background(), tx, 42, "111", "EAN13", true, "SKU-A"); err != nil {
			t.Fatalf("repeated same-product barcode persist error: %v", err)
		}
	}
	if tx.upsertCount != 2 {
		t.Fatalf("upsert count = %d, want 2 idempotent writes", tx.upsertCount)
	}
	if tx.ownerQueryCount != 0 {
		t.Fatalf("owner lookup count = %d, want 0 for successful same-product writes", tx.ownerQueryCount)
	}
}

func TestPersistSAPBarcodeBlocksCrossProductConflict(t *testing.T) {
	tx := &barcodeTestTx{
		targetID:       7,
		upsertAffected: 0,
		ownerProductID: 3,
		ownerOrgID:     42,
		ownerSKU:       "SKU-OWNER",
	}
	err := persistSAPBarcode(context.Background(), tx, 42, "111", "EAN13", true, "SKU-TARGET")
	if err == nil {
		t.Fatal("cross-product barcode ownership conflict was accepted")
	}
	for _, fragment := range []string{"111", "ownership conflict", "SKU-OWNER", "SKU-TARGET"} {
		if !strings.Contains(err.Error(), fragment) {
			t.Errorf("conflict error %q does not contain %q", err, fragment)
		}
	}
}

func TestPersistSAPBarcodeRejectsMissingTargetProduct(t *testing.T) {
	tx := &barcodeTestTx{targetErr: pgx.ErrNoRows}
	err := persistSAPBarcode(context.Background(), tx, 42, "111", "EAN13", true, "MISSING-SKU")
	if err == nil || !strings.Contains(err.Error(), "MISSING-SKU") {
		t.Fatalf("missing target product error = %v, want a useful target-product error", err)
	}
	if tx.upsertCount != 0 {
		t.Fatalf("upsert count = %d, want 0 when target product is missing", tx.upsertCount)
	}
}

func TestSAPBarcodeUpsertChecksOwnershipBeforeUpdating(t *testing.T) {
	query := strings.Join(strings.Fields(strings.ToLower(sapBarcodeUpsertQuery)), " ")
	for _, fragment := range []string{
		"insert into product_barcodes (product_id, barcode, barcode_type, is_primary)",
		"on conflict (barcode) do update set is_primary = excluded.is_primary",
		"where product_barcodes.product_id = excluded.product_id",
	} {
		if !strings.Contains(query, fragment) {
			t.Errorf("barcode upsert is missing contract fragment %q", fragment)
		}
	}
	if strings.Contains(query, "product_id = excluded.product_id,") {
		t.Fatal("barcode upsert still changes product ownership")
	}
	if !strings.Contains(strings.ToLower(sapBarcodeProductLookupQuery), "organization_id = $1") {
		t.Fatal("barcode target product lookup is not organization-scoped")
	}
}
