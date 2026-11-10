package syncpb

// A timestamp and row ID form a bounded pull cursor, including timestamp ties.
// Metadata keeps the existing protobuf wire format compatible.
const PullCursorKey = "nembus-pull-after-id"
const PullCursorVersionKey = "nembus-pull-cursor-version"
const PullPageLimit = 200

// Metadata keys for Phase 4 terminal authentication and session negotiation.
const SyncTokenMetadataKey = "x-nembus-sync-token"
const SyncDeviceIDMetadataKey = "x-nembus-device-id"
const SyncStoreIDMetadataKey = "x-nembus-store-id"

// MaxInFlightBatches bounds the number of unacknowledged correlation batches
// that a POS client may pipeline over StreamPush or StreamSession.
const MaxInFlightBatches = 10

// PullEntityTypes lists replicated entities in parent-before-child order.
// Return a fresh slice so callers cannot mutate the shared allowlist.
func PullEntityTypes() []string {
	return []string{
		"units_of_measure",
		"product_categories",
		"payment_terms",
		"users",
		"roles",
		"permissions",
		"role_permissions",
		"user_roles",
		"cashiers",
		"business_partners",
		"partner_addresses",
		"user_biometric_templates",
		"price_lists",
		"products",
		"product_variants",
		"product_barcodes",
		"customers",
		"recipes",
		"menu_items",
		"menu_modifier_groups",
		"combo_bundles",
		"menu_item_availability_schedules",
		"promotions",
		"restaurant_promotions",
		"product_prices",
		"stock_counts",
		"stock_count_lines",
		"inventory_stock",
		"stock_movements",
		"organizations",
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

// PullEntityRank returns the 0-based parent-before-child dependency rank for a pulled entity.
func PullEntityRank(entity string) int {
	for idx, candidate := range PullEntityTypes() {
		if candidate == entity {
			return idx
		}
	}
	return len(PullEntityTypes())
}

// LeafDeleteEntityTypes lists the join/leaf tables where hard DELETE operations
// are permitted and replicated as DELETE tombstones to POS terminals.
func LeafDeleteEntityTypes() []string {
	return []string{
		"role_permissions",
		"user_roles",
		"partner_addresses",
		"user_biometric_templates",
		"product_barcodes",
		"menu_item_availability_schedules",
		"product_prices",
	}
}

// IsLeafDeleteEntity reports whether hard DELETE tombstones are permitted for entity.
func IsLeafDeleteEntity(entity string) bool {
	for _, candidate := range LeafDeleteEntityTypes() {
		if candidate == entity {
			return true
		}
	}
	return false
}

// PushEntityRank returns the deterministic dependency rank for ordering events
// within a POS correlation batch before streaming to Cloud.
func PushEntityRank(entity, action string) int {
	switch entity {
	case "customers":
		return 10
	case "cashier_sessions":
		return 20
	case "carts":
		if action == "INSERT" {
			return 22
		}
		return 28
	case "cart_items":
		return 24
	case "sales_orders_v2":
		if action == "POS_CHECKOUT" {
			return 27
		}
		return 25
	case "sales_order_lines_v2":
		return 26
	case "pos_transactions":
		return 30
	case "pos_transaction_lines":
		return 40
	case "pos_payments":
		return 45
	case "sales_returns":
		return 50
	case "sales_return_lines":
		return 55
	case "zatca_invoices":
		return 80
	default:
		return 90
	}
}
