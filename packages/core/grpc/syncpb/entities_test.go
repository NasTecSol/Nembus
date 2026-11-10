package syncpb

import "testing"

func TestPullEntityRankAndLeafDeleteAllowlist(t *testing.T) {
	entities := PullEntityTypes()
	if len(entities) != 30 {
		t.Fatalf("expected 30 PullEntityTypes, got %d", len(entities))
	}
	for idx, entity := range entities {
		if got := PullEntityRank(entity); got != idx {
			t.Fatalf("PullEntityRank(%q) = %d, want %d", entity, got, idx)
		}
	}
	if PullEntityRank("unknown_table") != len(entities) {
		t.Fatalf("expected unknown table rank %d", len(entities))
	}

	// Verify parent-before-child invariants in PullEntityRank
	parentChildPairs := [][2]string{
		{"roles", "role_permissions"},
		{"permissions", "role_permissions"},
		{"users", "user_roles"},
		{"roles", "user_roles"},
		{"business_partners", "partner_addresses"},
		{"users", "user_biometric_templates"},
		{"products", "product_variants"},
		{"product_variants", "product_barcodes"},
		{"price_lists", "product_prices"},
		{"products", "product_prices"},
		{"stock_counts", "stock_count_lines"},
	}
	for _, pair := range parentChildPairs {
		if PullEntityRank(pair[0]) >= PullEntityRank(pair[1]) {
			t.Errorf("expected parent %s (rank %d) before child %s (rank %d)",
				pair[0], PullEntityRank(pair[0]), pair[1], PullEntityRank(pair[1]))
		}
	}

	// Only the 7 join/leaf tables allow hard DELETE tombstones
	leafTables := LeafDeleteEntityTypes()
	if len(leafTables) != 7 {
		t.Fatalf("expected 7 LeafDeleteEntityTypes, got %d", len(leafTables))
	}
	for _, leaf := range leafTables {
		if !IsLeafDeleteEntity(leaf) {
			t.Errorf("expected %s to be a leaf delete entity", leaf)
		}
		if !IsPullEntity(leaf) {
			t.Errorf("leaf delete entity %s must also be in PullEntityTypes", leaf)
		}
	}
	for _, master := range []string{"products", "users", "customers", "organizations", "price_lists"} {
		if IsLeafDeleteEntity(master) {
			t.Errorf("master table %s must not allow hard DELETE tombstones", master)
		}
	}
}

func TestPushEntityRankOrder(t *testing.T) {
	ordered := []struct {
		entity string
		action string
	}{
		{"customers", "INSERT"},
		{"cashier_sessions", "INSERT"},
		{"carts", "INSERT"},
		{"cart_items", "INSERT"},
		{"sales_orders_v2", "INSERT"},
		{"sales_order_lines_v2", "INSERT"},
		{"sales_orders_v2", "POS_CHECKOUT"},
		{"carts", "UPDATE"},
		{"pos_transactions", "INSERT"},
		{"pos_transaction_lines", "INSERT"},
		{"pos_payments", "INSERT"},
		{"sales_returns", "INSERT"},
		{"sales_return_lines", "INSERT"},
		{"zatca_invoices", "INSERT"},
	}

	for i := 1; i < len(ordered); i++ {
		prev := ordered[i-1]
		curr := ordered[i]
		rankPrev := PushEntityRank(prev.entity, prev.action)
		rankCurr := PushEntityRank(curr.entity, curr.action)
		if rankPrev >= rankCurr {
			t.Errorf("expected (%s, %s) rank %d < (%s, %s) rank %d",
				prev.entity, prev.action, rankPrev, curr.entity, curr.action, rankCurr)
		}
	}
}
