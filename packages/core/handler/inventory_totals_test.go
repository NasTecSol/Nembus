package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseInventoryTotalsParams(t *testing.T) {
	for _, tc := range []struct{ name, query, wantError string }{
		{"all stores", "organization_id=1", ""},
		{"one store", "organization_id=1&store_id=2", ""},
		{"one location", "organization_id=1&store_id=2&storage_location_id=3&price_list_id=4", ""},
		{"missing org", "", "organization_id"},
		{"negative store", "organization_id=1&store_id=-1", "store_id"},
		{"invalid store", "organization_id=1&store_id=abc", "store_id"},
		{"overflow store", "organization_id=1&store_id=2147483648", "store_id"},
		{"negative location", "organization_id=1&storage_location_id=-1", "storage_location_id"},
		{"location needs store", "organization_id=1&storage_location_id=3", "store_id is required"},
		{"invalid list", "organization_id=1&price_list_id=0", "price_list_id"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("GET", "/api/inventory-stock/totals?"+tc.query, nil)
			p, err := parseInventoryTotalsParams(c)
			if tc.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("expected %s, got %v", tc.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if p.OrganizationID != 1 {
				t.Fatalf("unexpected organization: %+v", p)
			}
			if tc.name == "all stores" && (p.StoreID != 0 || p.StorageLocationID != 0 || p.PriceListID != 1) {
				t.Fatalf("unexpected defaults: %+v", p)
			}
			if tc.name == "one location" && (p.StoreID != 2 || p.StorageLocationID != 3 || p.PriceListID != 4) {
				t.Fatalf("lost filters: %+v", p)
			}
		})
	}
}
