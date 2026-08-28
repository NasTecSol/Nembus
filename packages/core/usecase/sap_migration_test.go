package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/NasTecSol/nembus-sap/mappings"
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

type priceTestRow struct {
	err error
	id  int
}

func (r priceTestRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*dest[0].(*int) = r.id
	return nil
}

type priceTestTx struct {
	pgx.Tx

	productID, priceListID, uomID    int
	productErr, priceListErr, uomErr error
	updateAffected, insertAffected   int64
	productArgs                      []any
	writes                           []string
}

func (tx *priceTestTx) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	lower := strings.ToLower(query)
	switch {
	case strings.Contains(lower, "from products"):
		tx.productArgs = append([]any(nil), args...)
		return priceTestRow{id: tx.productID, err: tx.productErr}
	case strings.Contains(lower, "from price_lists"):
		return priceTestRow{id: tx.priceListID, err: tx.priceListErr}
	default:
		return priceTestRow{id: tx.uomID, err: tx.uomErr}
	}
}

func (tx *priceTestTx) Exec(_ context.Context, query string, _ ...any) (pgconn.CommandTag, error) {
	tx.writes = append(tx.writes, query)
	lower := strings.ToLower(query)
	switch {
	case strings.Contains(lower, "update product_prices"):
		return pgconn.NewCommandTag(fmt.Sprintf("UPDATE 0 %d", tx.updateAffected)), nil
	case strings.Contains(lower, "insert into product_prices"):
		return pgconn.NewCommandTag(fmt.Sprintf("INSERT 0 %d", tx.insertAffected)), nil
	default:
		return pgconn.NewCommandTag("OK"), nil
	}
}

func testPriceItem() mappings.CanonicalPriceListItem {
	return mappings.CanonicalPriceListItem{
		PriceListCode: "PL-10",
		ProductSKU:    "SKU-001",
		UOMCode:       "BOX",
		Price:         45.25,
		CurrencyCode:  "SAR",
		Metadata:      map[string]interface{}{"sap_uom_entry": int64(5)},
	}
}

func TestPersistSAPPriceItemInsertsResolvedPrice(t *testing.T) {
	tx := &priceTestTx{productID: 11, priceListID: 22, uomID: 33, insertAffected: 1}
	if err := persistSAPPriceItem(context.Background(), tx, 42, testPriceItem()); err != nil {
		t.Fatalf("persistSAPPriceItem() error = %v", err)
	}
	if len(tx.writes) != 6 { // SAVEPOINT, UPDATE, RELEASE, SAVEPOINT, INSERT, RELEASE
		t.Fatalf("write count = %d, want update and insert writes plus savepoints", len(tx.writes))
	}
	if !strings.Contains(strings.ToLower(tx.writes[1]), "update product_prices") || !strings.Contains(strings.ToLower(tx.writes[4]), "insert into product_prices") {
		t.Fatalf("writes do not contain update-then-insert persistence: %v", tx.writes)
	}
}

func TestPersistSAPPriceItemRerunUpdatesExistingBusinessKey(t *testing.T) {
	tx := &priceTestTx{productID: 11, priceListID: 22, uomID: 33, updateAffected: 1}
	for range 2 {
		if err := persistSAPPriceItem(context.Background(), tx, 42, testPriceItem()); err != nil {
			t.Fatalf("persistSAPPriceItem() error = %v", err)
		}
	}
	for _, query := range tx.writes {
		if strings.Contains(strings.ToLower(query), "insert into product_prices") {
			t.Fatal("idempotent replay inserted a duplicate price row")
		}
	}
}

func TestPersistSAPPriceItemRejectsMissingProduct(t *testing.T) {
	tx := &priceTestTx{productErr: pgx.ErrNoRows}
	err := persistSAPPriceItem(context.Background(), tx, 42, testPriceItem())
	if err == nil || !strings.Contains(err.Error(), "SKU-001") {
		t.Fatalf("missing product error = %v, want product SKU context", err)
	}
	if len(tx.writes) != 0 {
		t.Fatalf("writes = %d, want no writes for missing product", len(tx.writes))
	}
}

func TestPersistSAPPriceItemRejectsMissingPriceList(t *testing.T) {
	tx := &priceTestTx{productID: 11, priceListErr: pgx.ErrNoRows}
	err := persistSAPPriceItem(context.Background(), tx, 42, testPriceItem())
	if err == nil || !strings.Contains(err.Error(), "PL-10") {
		t.Fatalf("missing price-list error = %v, want price-list context", err)
	}
	if len(tx.writes) != 0 {
		t.Fatalf("writes = %d, want no writes for missing price list", len(tx.writes))
	}
}

func TestPersistSAPPriceItemRejectsMissingRequiredUOM(t *testing.T) {
	tx := &priceTestTx{productID: 11, priceListID: 22, uomErr: pgx.ErrNoRows}
	err := persistSAPPriceItem(context.Background(), tx, 42, testPriceItem())
	if err == nil || !strings.Contains(err.Error(), "BOX") {
		t.Fatalf("missing UoM error = %v, want UoM context", err)
	}
	if len(tx.writes) != 0 {
		t.Fatalf("writes = %d, want no writes for missing UoM", len(tx.writes))
	}
}

func TestPersistSAPPriceItemRejectsZeroRowWrite(t *testing.T) {
	tx := &priceTestTx{productID: 11, priceListID: 22, uomID: 33}
	err := persistSAPPriceItem(context.Background(), tx, 42, testPriceItem())
	if err == nil || !strings.Contains(err.Error(), "zero rows") {
		t.Fatalf("zero-row write error = %v, want explicit persistence failure", err)
	}
}

func TestPersistSAPPriceItemScopesProductLookupToOrganization(t *testing.T) {
	tx := &priceTestTx{productID: 11, priceListID: 22, uomID: 33, insertAffected: 1}
	if err := persistSAPPriceItem(context.Background(), tx, 42, testPriceItem()); err != nil {
		t.Fatalf("persistSAPPriceItem() error = %v", err)
	}
	if !strings.Contains(strings.ToLower(sapPriceProductLookupQuery), "organization_id = $1") {
		t.Fatal("product lookup is not organization scoped")
	}
	if len(tx.productArgs) != 2 || tx.productArgs[0] != 42 {
		t.Fatalf("product lookup args = %v, want organization ID 42 and SKU", tx.productArgs)
	}
}

func TestSAPPriceItemPersistenceKeepsUOMInBusinessKey(t *testing.T) {
	query := strings.ToLower(sapPriceItemUpdateQuery)
	if !strings.Contains(query, "uom_id is not distinct from $3") {
		t.Fatal("price update does not distinguish price rows by UoM")
	}
	if strings.Contains(query, "delete from product_prices") {
		t.Fatal("price persistence still deletes rows before replay")
	}
}

type inventoryTestRow struct {
	err   error
	id    int
	count int64
}

func (r inventoryTestRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) == 2 {
		*dest[0].(*int) = r.id
		*dest[1].(*int64) = r.count
		return nil
	}
	*dest[0].(*int) = r.id
	return nil
}

type inventoryTestTx struct {
	pgx.Tx
	productID, storeID, locationID       int
	productErr, storeErr, locationErr    error
	existingID                           int
	existingCount                        int64
	updateAffected, insertAffected       int64
	productArgs, storeArgs, locationArgs []any
	writes                               []string
	writeArgs                            [][]any
}

func (tx *inventoryTestTx) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	lower := strings.ToLower(query)
	switch {
	case strings.Contains(lower, "from products"):
		tx.productArgs = append([]any(nil), args...)
		return inventoryTestRow{id: tx.productID, err: tx.productErr}
	case strings.Contains(lower, "from stores") && !strings.Contains(lower, "storage_locations"):
		tx.storeArgs = append([]any(nil), args...)
		return inventoryTestRow{id: tx.storeID, err: tx.storeErr}
	case strings.Contains(lower, "from storage_locations"):
		tx.locationArgs = append([]any(nil), args...)
		return inventoryTestRow{id: tx.locationID, err: tx.locationErr}
	case strings.Contains(lower, "from inventory_stock"):
		return inventoryTestRow{id: tx.existingID, count: tx.existingCount}
	default:
		return inventoryTestRow{err: errors.New("unexpected inventory query")}
	}
}

func (tx *inventoryTestTx) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	tx.writes = append(tx.writes, query)
	lower := strings.ToLower(query)
	switch {
	case strings.Contains(lower, "update inventory_stock"):
		tx.writeArgs = append(tx.writeArgs, append([]any(nil), args...))
		return pgconn.NewCommandTag(fmt.Sprintf("UPDATE 0 %d", tx.updateAffected)), nil
	case strings.Contains(lower, "insert into inventory_stock"):
		tx.writeArgs = append(tx.writeArgs, append([]any(nil), args...))
		if tx.insertAffected > 0 {
			tx.existingID = 99
			tx.existingCount = 1
		}
		return pgconn.NewCommandTag(fmt.Sprintf("INSERT 0 %d", tx.insertAffected)), nil
	default:
		return pgconn.NewCommandTag("OK"), nil
	}
}

func testInventory() mappings.CanonicalInventoryStock {
	return mappings.CanonicalInventoryStock{
		ProductSKU:        "SKU-001",
		StoreCode:         "WH-001",
		QuantityOnHand:    10,
		QuantityAllocated: 0,
		QuantityAvailable: 10,
		QuantityOnOrder:   0,
		ReorderLevel:      2,
		MaxStockLevel:     20,
		Metadata:          map[string]interface{}{"sap_item_code": "SKU-001"},
	}
}

func TestPersistSAPInventoryInsertsFirstAbsoluteSnapshot(t *testing.T) {
	tx := &inventoryTestTx{productID: 11, storeID: 22, locationID: 33, insertAffected: 1}
	if err := persistSAPInventory(context.Background(), tx, 42, testInventory()); err != nil {
		t.Fatalf("persistSAPInventory() error = %v", err)
	}
	if tx.existingCount != 1 || tx.existingID != 99 {
		t.Fatalf("insert did not create one logical row in test state: id=%d count=%d", tx.existingID, tx.existingCount)
	}
	if len(tx.writeArgs) == 0 || tx.writeArgs[len(tx.writeArgs)-1][0] != 11 || tx.writeArgs[len(tx.writeArgs)-1][1] != 22 || tx.writeArgs[len(tx.writeArgs)-1][2] != 33 {
		t.Fatalf("insert args = %v, want product/store/MAIN-location IDs", tx.writeArgs)
	}
}

func TestPersistSAPInventoryRerunUpdatesOneLogicalRowAbsolutely(t *testing.T) {
	tx := &inventoryTestTx{productID: 11, storeID: 22, locationID: 33, insertAffected: 1}
	if err := persistSAPInventory(context.Background(), tx, 42, testInventory()); err != nil {
		t.Fatalf("first persist error = %v", err)
	}
	tx.insertAffected = 0
	tx.updateAffected = 1
	changed := testInventory()
	changed.QuantityOnHand = 7
	changed.QuantityAvailable = 7
	if err := persistSAPInventory(context.Background(), tx, 42, changed); err != nil {
		t.Fatalf("second persist error = %v", err)
	}
	if tx.existingCount != 1 {
		t.Fatalf("logical row count = %d, want 1", tx.existingCount)
	}
	insertCount := 0
	for _, query := range tx.writes {
		if strings.Contains(strings.ToLower(query), "insert into inventory_stock") {
			insertCount++
		}
	}
	if insertCount != 1 {
		t.Fatalf("inventory insert count = %d, want exactly one first-snapshot insert", insertCount)
	}
	lastArgs := tx.writeArgs[len(tx.writeArgs)-1]
	if lastArgs[4] != 7.0 || lastArgs[5] != 0.0 || lastArgs[6] != 7.0 || lastArgs[7] != 0.0 || lastArgs[8] != 0 {
		t.Fatalf("absolute update args = %v, want on-hand/available 7 and operational zeroes", lastArgs)
	}
}

func TestPersistSAPInventoryRejectsMissingReferencesBeforeWrite(t *testing.T) {
	tests := []struct {
		name string
		tx   *inventoryTestTx
		want string
	}{
		{name: "product", tx: &inventoryTestTx{productErr: pgx.ErrNoRows}, want: "SKU-001"},
		{name: "store", tx: &inventoryTestTx{productID: 11, storeErr: pgx.ErrNoRows}, want: "WH-001"},
		{name: "location", tx: &inventoryTestTx{productID: 11, storeID: 22, locationErr: pgx.ErrNoRows}, want: "MAIN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := persistSAPInventory(context.Background(), tt.tx, 42, testInventory())
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
			for _, query := range tt.tx.writes {
				if strings.Contains(strings.ToLower(query), "insert into inventory_stock") || strings.Contains(strings.ToLower(query), "update inventory_stock") {
					t.Fatal("missing reference still attempted an inventory write")
				}
			}
		})
	}
}

func TestPersistSAPInventoryRejectsZeroRowWritesAndDuplicates(t *testing.T) {
	zero := &inventoryTestTx{productID: 11, storeID: 22, locationID: 33}
	if err := persistSAPInventory(context.Background(), zero, 42, testInventory()); err == nil || !strings.Contains(err.Error(), "insert affected 0 rows") {
		t.Fatalf("zero-row insert error = %v, want explicit failure", err)
	}
	zeroUpdate := &inventoryTestTx{productID: 11, storeID: 22, locationID: 33, existingID: 44, existingCount: 1}
	if err := persistSAPInventory(context.Background(), zeroUpdate, 42, testInventory()); err == nil || !strings.Contains(err.Error(), "update affected 0 rows") {
		t.Fatalf("zero-row update error = %v, want explicit failure", err)
	}

	duplicate := &inventoryTestTx{productID: 11, storeID: 22, locationID: 33, existingCount: 2}
	if err := persistSAPInventory(context.Background(), duplicate, 42, testInventory()); err == nil || !strings.Contains(err.Error(), "2 existing rows") {
		t.Fatalf("duplicate ownership error = %v, want ambiguity failure", err)
	}
}

func TestPersistSAPInventoryUsesOrganizationAndStoreScopedKeys(t *testing.T) {
	tx := &inventoryTestTx{productID: 11, storeID: 22, locationID: 33, insertAffected: 1}
	if err := persistSAPInventory(context.Background(), tx, 42, testInventory()); err != nil {
		t.Fatalf("persistSAPInventory() error = %v", err)
	}
	if len(tx.productArgs) != 2 || tx.productArgs[0] != 42 || tx.productArgs[1] != "SKU-001" {
		t.Fatalf("product lookup args = %v, want organization and SKU", tx.productArgs)
	}
	if len(tx.storeArgs) != 2 || tx.storeArgs[0] != 42 || tx.storeArgs[1] != "WH-001" {
		t.Fatalf("store lookup args = %v, want organization and store code", tx.storeArgs)
	}
	if len(tx.locationArgs) != 3 || tx.locationArgs[0] != 42 || tx.locationArgs[1] != 22 || tx.locationArgs[2] != sapInventoryLocationCode {
		t.Fatalf("location lookup args = %v, want organization, resolved store, MAIN", tx.locationArgs)
	}
	for _, fragment := range []string{"product_variant_id is null", "storage_location_id", "where id = $1", "quantity_on_hand = $5"} {
		if !strings.Contains(strings.ToLower(sapInventoryUpdateQuery), fragment) {
			t.Errorf("inventory update query missing %q", fragment)
		}
	}
	locationQuery := strings.ToLower(sapInventoryLocationLookupQuery)
	for _, fragment := range []string{"join stores s on s.id = sl.store_id", "s.organization_id = $1", "s.id = $2", "sl.code = $3"} {
		if !strings.Contains(locationQuery, fragment) {
			t.Errorf("storage-location lookup missing ownership guard %q", fragment)
		}
	}
}
