package sync

import (
	"strings"
	"testing"
)

func TestProductPricesReconcileNaturalKeySQL(t *testing.T) {
	requiredFragments := []string{
		"DELETE FROM product_prices p",
		"json_populate_record(NULL::product_prices, $1::json) inc",
		"p.id <> inc.id",
		"p.product_id = inc.product_id",
		"p.price_list_id = inc.price_list_id",
		"p.product_variant_id IS NOT DISTINCT FROM inc.product_variant_id",
		"p.uom_id IS NOT DISTINCT FROM inc.uom_id",
		"p.min_quantity IS NOT DISTINCT FROM inc.min_quantity",
	}
	for _, frag := range requiredFragments {
		if !strings.Contains(productPricesReconcileNaturalKeySQL, frag) {
			t.Errorf("productPricesReconcileNaturalKeySQL missing %q", frag)
		}
	}
}
