package repository

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/jackc/pgx/v5"
)

// Run against a dedicated database initialized with core schema and views.
// All fixture data is rolled back; never use an application database for tests.
func TestMasterProductStockCatalogPostgres(t *testing.T) {
	dsn := os.Getenv("NEMBUS_STOCK_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set NEMBUS_STOCK_TEST_DATABASE_URL to run stock catalog integration tests")
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
INSERT INTO organizations (id, name, code) VALUES
(-920001, 'Stock test', 'stock-test'), (-920002, 'Other org', 'stock-test-other');
INSERT INTO stores (id, organization_id, name, code) VALUES
(-920011, -920001, 'First', 'first'), (-920012, -920001, 'Second', 'second'),
(-920013, -920002, 'Other', 'other');
INSERT INTO storage_locations (id, store_id, name, code) VALUES
(-920021, -920011, 'First location', 'first'), (-920022, -920011, 'Second location', 'second'),
(-920023, -920012, 'Other store location', 'other');
INSERT INTO product_categories (id, name, code) VALUES
(-920031, 'Category A', 'stock-cat-a'), (-920032, 'Category B', 'stock-cat-b');
INSERT INTO products (id, organization_id, sku, name, category_id, track_inventory) VALUES
(-920101, -920001, 'stock-p1', 'Multiple locations', -920031, true),
(-920102, -920001, 'stock-p2', 'Boundary', -920031, true),
(-920103, -920001, 'stock-p3', 'No inventory', -920031, true),
(-920104, -920001, 'stock-p4', 'Category B', -920032, true),
(-920105, -920001, 'stock-p5', 'Non inventory', -920031, false),
(-920106, -920002, 'stock-p6', 'Other organization', -920031, true),
(-920107, -920001, 'stock-p7', 'Null quantity', -920031, true),
(-920108, -920001, 'stock-p8', 'Negative quantity', -920031, true),
(-920109, -920001, 'stock-p9', 'Decimal quantity', -920031, true);
INSERT INTO product_variants (id, product_id, variant_sku, variant_attributes) VALUES
(-920041, -920101, 'stock-variant-1', '{}'), (-920042, -920101, 'stock-variant-2', '{}');
INSERT INTO inventory_stock (product_id, product_variant_id, store_id, storage_location_id, quantity_available) VALUES
(-920101, -920041, -920011, -920021, 5), (-920101, -920042, -920011, -920022, 7),
(-920101, NULL, -920012, -920023, 100), (-920102, NULL, -920011, -920021, 20),
(-920104, NULL, -920011, -920021, 3), (-920105, NULL, -920011, -920021, 1),
(-920106, NULL, -920013, NULL, 1), (-920107, NULL, -920011, -920021, NULL),
(-920108, NULL, -920011, -920021, -2), (-920109, NULL, -920011, -920021, 19.5);
`)
	if err != nil {
		t.Fatal(err)
	}
	repo := New(tx)
	for _, tc := range []struct {
		name, status, threshold   string
		store, location, category int32
		want                      []int32
	}{
		{"global low", "low", "20", 0, 0, 0, []int32{-920103, -920104, -920107, -920108, -920109}},
		{"boundary non low", "non_low", "20", 0, 0, 0, []int32{-920101, -920102}},
		{"both classifications", "all", "20", 0, 0, 0, []int32{-920101, -920102, -920103, -920104, -920107, -920108, -920109}},
		{"store total", "low", "20", -920011, 0, 0, []int32{-920101, -920104, -920107, -920108, -920109}},
		{"store and category", "low", "20", -920011, 0, -920031, []int32{-920101, -920107, -920108, -920109}},
		{"store and location", "low", "20", -920011, -920022, 0, []int32{-920101}},
		{"location alone", "low", "20", 0, -920022, 0, []int32{-920101}},
		{"other store", "non_low", "20", -920012, 0, 0, []int32{-920101}},
		{"mismatched scope", "all", "20", -920012, -920021, 0, []int32{}},
		{"other organization store", "all", "20", -920013, 0, 0, []int32{}},
		{"decimal boundary", "low", "19.5", -920011, 0, 0, []int32{-920101, -920104, -920107, -920108}},
		{"zero threshold", "low", "0", 0, 0, 0, []int32{-920108}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := GetMasterProductStockCatalogParams{OrganizationID: -920001, StoreID: tc.store, StorageLocationID: tc.location, CategoryID: tc.category, StockStatus: tc.status, Limit: 100}
			if err := p.LowStockThreshold.Scan(tc.threshold); err != nil {
				t.Fatal(err)
			}
			rows, err := repo.GetMasterProductStockCatalog(ctx, p)
			if err != nil {
				t.Fatal(err)
			}
			ids := make([]int32, 0, len(rows))
			for _, row := range rows {
				ids = append(ids, row.ProductID)
				var inventory []struct {
					StoreID    int32   `json:"store_id"`
					LocationID int32   `json:"storage_location_id"`
					Available  float64 `json:"quantity_available"`
				}
				raw, err := json.Marshal(row.Inventory)
				if err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(raw, &inventory); err != nil {
					t.Fatal(err)
				}
				var sum float64
				for _, stock := range inventory {
					if tc.store != 0 && stock.StoreID != tc.store {
						t.Fatalf("inventory outside requested store: %s", raw)
					}
					if tc.location != 0 && stock.LocationID != tc.location {
						t.Fatalf("inventory outside requested location: %s", raw)
					}
					sum += stock.Available
				}
				available, _ := row.AvailableStock.Float64Value()
				threshold, _ := p.LowStockThreshold.Float64Value()
				if available.Float64 != sum || row.IsLowStock != (sum < threshold.Float64) {
					t.Fatalf("incorrect stock summary: %+v", row)
				}
			}
			sort.Slice(tc.want, func(i, j int) bool { return tc.want[i] < tc.want[j] })
			if !reflect.DeepEqual(ids, tc.want) {
				t.Fatalf("got products %v, want %v", ids, tc.want)
			}
			count, err := repo.GetMasterProductStockCatalogCount(ctx, GetMasterProductStockCatalogCountParams{
				OrganizationID: p.OrganizationID, StoreID: p.StoreID, StorageLocationID: p.StorageLocationID,
				CategoryID: p.CategoryID, StockStatus: p.StockStatus, LowStockThreshold: p.LowStockThreshold,
			})
			if err != nil || count != int64(len(tc.want)) {
				t.Fatalf("count=%d err=%v, want %d", count, err, len(tc.want))
			}
			// Pagination must preserve product ordering and count after the last page.
			for offset := 0; offset <= len(tc.want)+1; offset++ {
				p.Limit, p.Offset = 1, int32(offset)
				page, err := repo.GetMasterProductStockCatalog(ctx, p)
				if err != nil {
					t.Fatal(err)
				}
				if offset < len(tc.want) {
					if len(page) != 1 || page[0].ProductID != tc.want[offset] {
						t.Fatalf("unexpected page at offset %d: %+v", offset, page)
					}
				} else if len(page) != 0 {
					t.Fatalf("expected empty page at offset %d", offset)
				}
			}
		})
	}
}
