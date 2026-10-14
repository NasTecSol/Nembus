package syncpb

// A timestamp and row ID form a bounded pull cursor, including timestamp ties.
// Metadata keeps the existing protobuf wire format compatible.
const PullCursorKey = "nembus-pull-after-id"
const PullCursorVersionKey = "nembus-pull-cursor-version"
const PullPageLimit = 200

// PullEntityTypes lists replicated entities in parent-before-child order.
// Return a fresh slice so callers cannot mutate the shared allowlist.
func PullEntityTypes() []string {
	return []string{
		"price_lists", "products", "product_variants", "product_barcodes",
		"customers", "recipes", "menu_items", "menu_modifier_groups", "combo_bundles",
		"menu_item_availability_schedules",
		"promotions", "restaurant_promotions", "product_prices",
		"stock_counts", "stock_count_lines", "inventory_stock", "stock_movements",
		"zatca_device_configs",
	}
}

func IsPullEntity(entity string) bool {
	for _, candidate := range PullEntityTypes() {
		if candidate == entity {
			return true
		}
	}
	return false
}
