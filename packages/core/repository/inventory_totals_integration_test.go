package repository

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

// All fixtures are temporary tables on a dedicated connection and rolled back.
// search_path prevents this test from reading or writing application tables.
func TestInventoryTotalsPostgres(t *testing.T) {
	dsn := os.Getenv("NEMBUS_INVENTORY_TOTALS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set NEMBUS_INVENTORY_TOTALS_TEST_DATABASE_URL for PostgreSQL tests")
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
SET LOCAL search_path = pg_temp;
CREATE TEMP TABLE organizations (id int, currency_code text);
CREATE TEMP TABLE stores (id int, organization_id int, name text);
CREATE TEMP TABLE storage_locations (id int, store_id int);
CREATE TEMP TABLE products (id int, organization_id int, base_uom_id int, track_inventory bool, cost_price numeric);
CREATE TEMP TABLE inventory_stock (store_id int, storage_location_id int, product_id int, product_variant_id int, quantity_on_hand numeric, quantity_available numeric);
CREATE TEMP TABLE price_lists (id int, currency_code text, is_active bool, valid_from date, valid_to date);
CREATE TEMP TABLE product_prices (id int, product_id int, product_variant_id int, price_list_id int, uom_id int, price numeric, min_quantity numeric, max_quantity numeric, is_active bool, valid_from date, valid_to date, updated_at timestamp);
INSERT INTO organizations VALUES (1,'SAR'),(2,'SAR'),(3,'SAR');
INSERT INTO stores VALUES (10,1,'First'),(11,1,'Second'),(12,1,'Empty'),(20,2,'Other organization');
INSERT INTO storage_locations VALUES (100,10),(101,10),(110,11);
INSERT INTO products VALUES (1,1,100,true,2),(2,1,100,true,3.5),(3,1,100,true,NULL),(4,1,100,false,10),(5,2,100,true,99),(6,1,100,true,4);
INSERT INTO inventory_stock VALUES
 (10,100,1,NULL,5,4),(10,101,1,7,2,1),(10,100,2,NULL,1.5,1.5),
 (10,100,3,NULL,3,3),(10,100,4,NULL,100,100),(10,100,5,NULL,100,100),
 (11,110,1,NULL,10,8),(11,110,2,NULL,-1,-1),(11,110,6,NULL,2,2),
 (20,NULL,5,NULL,1000,1000),(10,100,1,NULL,NULL,NULL);
INSERT INTO price_lists VALUES (1,'SAR',true,NULL,NULL),(2,'USD',true,NULL,NULL),(3,'SAR',true,NULL,CURRENT_DATE-1);
INSERT INTO product_prices VALUES
 (1,1,NULL,1,100,9,1,NULL,true,NULL,NULL,CURRENT_TIMESTAMP-interval '1 day'),
 (2,1,NULL,1,100,10,1,NULL,true,NULL,NULL,CURRENT_TIMESTAMP),
 (3,1,7,1,100,12,1,NULL,true,NULL,NULL,CURRENT_TIMESTAMP),
 (4,2,NULL,1,100,20,1,NULL,true,NULL,NULL,CURRENT_TIMESTAMP),
 (5,6,NULL,1,200,50,1,NULL,true,NULL,NULL,CURRENT_TIMESTAMP),
 (6,1,NULL,1,100,999,1,NULL,false,NULL,NULL,CURRENT_TIMESTAMP),
 (7,1,NULL,1,100,999,1,NULL,true,CURRENT_DATE+1,NULL,CURRENT_TIMESTAMP),
 (8,1,NULL,1,100,999,1,NULL,true,NULL,CURRENT_DATE-1,CURRENT_TIMESTAMP),
 (9,1,NULL,1,100,1,10,NULL,true,NULL,NULL,CURRENT_TIMESTAMP);
`)
	if err != nil {
		t.Fatal(err)
	}
	repo := New(tx)
	type totals struct {
		TotalUnits  float64           `json:"total_units"`
		Available   float64           `json:"total_available_units"`
		Cost        float64           `json:"total_cost_value"`
		Sale        float64           `json:"total_sale_value"`
		MissingCost float64           `json:"units_missing_cost_price"`
		MissingSale float64           `json:"units_missing_sale_price"`
		Products    int64             `json:"product_count"`
		Stores      []json.RawMessage `json:"stores"`
	}
	for _, tc := range []struct {
		name                                                   string
		org, store, location                                   int32
		units, available, cost, sale, missingCost, missingSale float64
		products                                               int64
		stores                                                 int
	}{
		{"all stores", 1, 0, 0, 22.5, 18.5, 43.75, 184, 3, 5, 4, 3},
		{"one store", 1, 10, 0, 11.5, 9.5, 19.25, 104, 3, 3, 3, 1},
		{"one location", 1, 10, 100, 9.5, 8.5, 15.25, 80, 3, 3, 3, 1},
		{"variant price", 1, 10, 101, 2, 1, 4, 24, 0, 0, 1, 1},
		{"negative units", 1, 11, 0, 11, 9, 24.5, 80, 0, 2, 3, 1},
		{"empty store", 1, 12, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		{"empty org", 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			row, err := repo.GetInventoryTotals(ctx, GetInventoryTotalsParams{OrganizationID: tc.org, StoreID: tc.store, StorageLocationID: tc.location, PriceListID: 1})
			if err != nil {
				t.Fatal(err)
			}
			if !row.OrganizationExists || !row.StoreExists || !row.LocationExists || !row.PriceListValid {
				t.Fatalf("invalid scope: %+v", row)
			}
			var got totals
			if err := json.Unmarshal(row.Data, &got); err != nil {
				t.Fatal(err)
			}
			if got.TotalUnits != tc.units || got.Available != tc.available || got.Cost != tc.cost || got.Sale != tc.sale || got.MissingCost != tc.missingCost || got.MissingSale != tc.missingSale || got.Products != tc.products || got.Stores == nil || len(got.Stores) != tc.stores {
				t.Fatalf("unexpected totals: %s", row.Data)
			}
		})
	}
	for _, tc := range []struct {
		name    string
		params  GetInventoryTotalsParams
		invalid string
	}{
		{"wrong organization", GetInventoryTotalsParams{OrganizationID: 999, PriceListID: 1}, "org"},
		{"foreign store", GetInventoryTotalsParams{OrganizationID: 1, StoreID: 20, PriceListID: 1}, "store"},
		{"foreign location", GetInventoryTotalsParams{OrganizationID: 1, StoreID: 10, StorageLocationID: 110, PriceListID: 1}, "location"},
		{"missing location", GetInventoryTotalsParams{OrganizationID: 1, StoreID: 10, StorageLocationID: 999, PriceListID: 1}, "location"},
		{"different currency", GetInventoryTotalsParams{OrganizationID: 1, PriceListID: 2}, "price"},
		{"expired list", GetInventoryTotalsParams{OrganizationID: 1, PriceListID: 3}, "price"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			row, err := repo.GetInventoryTotals(ctx, tc.params)
			if err != nil {
				t.Fatal(err)
			}
			flags := map[string]bool{"org": row.OrganizationExists, "store": row.StoreExists, "location": row.LocationExists, "price": row.PriceListValid}
			if flags[tc.invalid] {
				t.Fatalf("invalid scope accepted: %+v", row)
			}
		})
	}
	if _, err := tx.Exec(ctx, `ALTER TABLE pg_temp.products DROP COLUMN cost_price`); err != nil {
		t.Fatal(err)
	}
	row, err := repo.GetInventoryTotals(ctx, GetInventoryTotalsParams{OrganizationID: 1, PriceListID: 1})
	if err != nil {
		t.Fatal(err)
	}
	var missing totals
	if err := json.Unmarshal(row.Data, &missing); err != nil {
		t.Fatal(err)
	}
	if missing.Cost != 0 || missing.MissingCost != 24.5 {
		t.Fatalf("missing cost column not handled: %s", row.Data)
	}
}
