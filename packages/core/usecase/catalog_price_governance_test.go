package usecase

import (
	"strings"
	"testing"
)

func TestParseCatalogPricePolicy_DefaultsWhenDisabled(t *testing.T) {
	r := ParseCatalogPricePolicy(nil, false)
	if r.RequiresApproval {
		t.Fatal("expected requires_approval false")
	}
}

func TestPriceChangeWorkflow_PolicyDetectsDrop(t *testing.T) {
	r := ParseCatalogPricePolicy(map[string]any{"price_drop_requires_approval": true}, true)
	ok, reason := r.PriceChangeNeedsReview(100, 80, 50, true)
	if !ok || !strings.Contains(reason, "dropped") {
		t.Fatalf("expected price drop review, got ok=%v reason=%q", ok, reason)
	}
}

func TestCostIncreaseWorkflow_PolicyDetectsThreshold(t *testing.T) {
	r := ParseCatalogPricePolicy(map[string]any{"cost_increase_threshold_pct": 10.0}, true)
	ok, _ := r.CostIncreaseNeedsReview(100, 115)
	if !ok {
		t.Fatal("expected cost increase review at 15%")
	}
	ok, _ = r.CostIncreaseNeedsReview(100, 105)
	if ok {
		t.Fatal("expected no review at 5% increase")
	}
}

func TestPosTxEligibleForSapOutbox_ForcePost(t *testing.T) {
	if !PosTxEligibleForSapOutbox("", "held_stock", true) {
		t.Fatal("sap_force_post should re-admit held_stock rows")
	}
	if PosTxEligibleForSapOutbox("", "held_stock", false) {
		t.Fatal("held_stock without force should be excluded")
	}
	if !PosTxEligibleForSapOutbox("", "ready_for_sap_post", false) {
		t.Fatal("ready_for_sap_post should be eligible")
	}
	if PosTxEligibleForSapOutbox("5001", "synced", false) {
		t.Fatal("posted doc should be excluded")
	}
}
