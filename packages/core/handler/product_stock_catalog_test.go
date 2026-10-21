package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseProductStockCatalogParams(t *testing.T) {
	for _, tc := range []struct {
		name      string
		query     string
		wantError string
	}{
		{"defaults", "organization_id=1", ""},
		{"combined filters", "organization_id=1&category_id=2&store_id=3&storage_location_id=4&low_stock_threshold=20.5&stock_status=all&limit=2&offset=4", ""},
		{"zero threshold", "organization_id=1&low_stock_threshold=0&stock_status=non_low", ""},
		{"missing organization", "", "organization_id"},
		{"negative store", "organization_id=1&store_id=-1", "store_id"},
		{"invalid location", "organization_id=1&storage_location_id=abc", "storage_location_id"},
		{"overflow category", "organization_id=1&category_id=2147483648", "category_id"},
		{"invalid status", "organization_id=1&stock_status=unknown", "stock_status"},
		{"negative threshold", "organization_id=1&low_stock_threshold=-1", "low_stock_threshold"},
		{"nan threshold", "organization_id=1&low_stock_threshold=NaN", "low_stock_threshold"},
		{"infinite threshold", "organization_id=1&low_stock_threshold=Inf", "low_stock_threshold"},
		{"zero limit", "organization_id=1&limit=0", "limit"},
		{"negative offset", "organization_id=1&offset=-1", "offset"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("GET", "/api/products/master-catalog/stock?"+tc.query, nil)
			p, err := parseProductStockCatalogParams(c)
			if tc.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("expected error containing %q, got %v", tc.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tc.name == "defaults" {
				threshold, _ := p.LowStockThreshold.Float64Value()
				if p.OrganizationID != 1 || p.StockStatus != "low" || p.Limit != 100 || p.Offset != 0 || threshold.Float64 != 20 {
					t.Fatalf("unexpected defaults: %+v", p)
				}
			}
			if tc.name == "combined filters" {
				threshold, _ := p.LowStockThreshold.Float64Value()
				if p.CategoryID != 2 || p.StoreID != 3 || p.StorageLocationID != 4 || p.StockStatus != "all" || p.Limit != 2 || p.Offset != 4 || threshold.Float64 != 20.5 {
					t.Fatalf("filters were not preserved: %+v", p)
				}
			}
		})
	}
}
