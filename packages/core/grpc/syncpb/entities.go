package syncpb

import (
	"strconv"
	"strings"
)

// A timestamp and row ID form a bounded pull cursor, including timestamp ties.
// Metadata keeps the existing protobuf wire format compatible.
const PullCursorKey = "nembus-pull-after-id"
const PullCursorAfterKey = "nembus-pull-after-key"
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
		"printing_templates",
		"store_receipt_settings",
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

// IsUUIDEntity reports whether entityType uses a UUID primary key.
//
// Categories covered:
//   - 15 M1-migrated push-synced tables (5 also in PullEntityTypes, 1 in LeafDeleteEntityTypes):
//     user_biometric_templates, cashier_sessions, stock_movements, stock_counts, stock_count_lines,
//     customers, pos_transactions, pos_transaction_lines, pos_payments, restaurant_orders,
//     restaurant_order_items, waste_logs, kiosk_sessions, sales_returns, sales_return_lines.
//   - 6 already-UUID sync-referenced tables (isSupportedPushEntity / validTables / PushEntityRank):
//     carts, cart_items, sales_orders_v2, sales_order_lines_v2, draft_cart_templates, notification_events.
//   - 18 additional already-UUID domain tables in packages/core/db/schema/*.sql:
//     draft_cart_template_items, restaurant_carts, restaurant_cart_items, order_fulfillments,
//     order_fulfillment_items, invoices, invoice_lines, invoice_payments, quotes, quote_lines,
//     user_notifications, notification_deliveries, customer_payments, vendor_payments,
//     rfqs, rfq_items, rfq_responses, rfq_response_items.
//
// All 27 remaining integer PullEntityTypes(), zatca_device_configs, and zatca_document_chain return false.
func IsUUIDEntity(entityType string) bool {
	switch entityType {
	case // 15 M1-migrated push-synced tables
		"user_biometric_templates",
		"cashier_sessions",
		"stock_movements",
		"stock_counts",
		"stock_count_lines",
		"customers",
		"pos_transactions",
		"pos_transaction_lines",
		"pos_payments",
		"restaurant_orders",
		"restaurant_order_items",
		"waste_logs",
		"kiosk_sessions",
		"sales_returns",
		"sales_return_lines",
		// 6 already-UUID sync-referenced tables
		"carts",
		"cart_items",
		"sales_orders_v2",
		"sales_order_lines_v2",
		"draft_cart_templates",
		"notification_events",
		// 18 additional already-UUID domain tables in packages/core/db/schema/*.sql
		"draft_cart_template_items",
		"restaurant_carts",
		"restaurant_cart_items",
		"order_fulfillments",
		"order_fulfillment_items",
		"invoices",
		"invoice_lines",
		"invoice_payments",
		"quotes",
		"quote_lines",
		"user_notifications",
		"notification_deliveries",
		"customer_payments",
		"vendor_payments",
		"rfqs",
		"rfq_items",
		"rfq_responses",
		"rfq_response_items":
		return true
	default:
		return false
	}
}

// ResolveEntityKey returns the canonical string entity key from entityKey if
// non-empty after trimming whitespace, or formats positive legacy entityID as
// base-10. Returns "" when neither is valid.
func ResolveEntityKey(entityKey string, entityID int64) string {
	if trimmed := strings.TrimSpace(entityKey); trimmed != "" {
		return trimmed
	}
	if entityID > 0 {
		return strconv.FormatInt(entityID, 10)
	}
	return ""
}

// ParseNumericEntityID parses key as a positive base-10 int64 entity ID.
// Returns (0, false) for empty, whitespace-only, leading '+', non-positive,
// overflowing, or non-numeric strings (such as UUIDs).
func ParseNumericEntityID(key string) (int64, bool) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" || strings.HasPrefix(trimmed, "+") {
		return 0, false
	}
	id, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
