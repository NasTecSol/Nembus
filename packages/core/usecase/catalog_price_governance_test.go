package usecase

import (
	"strings"
	"testing"
)

func testCatalogPriceRules() CatalogPriceRules {
	return CatalogPriceRules{
		PriceDropRequiresApproval: true,
		BelowCostRequiresApproval: true,
		HasCostIncreaseThreshold:  true,
		CostIncreaseThresholdPct:  10.0,
	}
}

func TestParseCatalogPriceRules_NoHardcodedDefaults(t *testing.T) {
	for _, input := range []map[string]any{nil, {}} {
		r := ParseCatalogPriceRules(input)
		if r.PriceDropRequiresApproval || r.BelowCostRequiresApproval || r.HasCostIncreaseThreshold || r.CostIncreaseThresholdPct != 0 {
			t.Fatalf("expected zero rules for %v, got %+v", input, r)
		}
	}
}

func TestEvaluateSellPriceChange_IncreaseRequiresHold(t *testing.T) {
	rules := testCatalogPriceRules()
	hold, reason := EvaluateSellPriceChange(true, rules, 100, 120, 80, true)
	if !hold {
		t.Fatal("expected hold on price increase")
	}
	if reason == "" {
		t.Fatal("expected reason")
	}
}

func TestEvaluateSellPriceChange_IncreaseThreshold(t *testing.T) {
	rules := testCatalogPriceRules() // 10% threshold
	if hold, _ := EvaluateSellPriceChange(true, rules, 15, 16, 10, true); hold {
		t.Fatal("expected no hold for 6.7% sell price increase below 10% threshold")
	}
	if hold, _ := EvaluateSellPriceChange(true, rules, 15, 17, 10, true); !hold {
		t.Fatal("expected hold for 13.3% sell price increase at/above 10% threshold")
	}
	if hold, _ := EvaluateSellPriceChange(true, rules, 15, 16, 20, true); !hold {
		t.Fatal("expected hold when below-cost rule triggers even below increase threshold")
	}
}

func TestEvaluateSellPriceChange_DecreaseRequiresHold(t *testing.T) {
	rules := testCatalogPriceRules()
	hold, _ := EvaluateSellPriceChange(true, rules, 100, 90, 80, true)
	if !hold {
		t.Fatal("expected hold on price decrease")
	}
}

func TestEvaluateSellPriceChange_DecreaseAllowedWhenDropRuleDisabled(t *testing.T) {
	rules := CatalogPriceRules{PriceDropRequiresApproval: false, BelowCostRequiresApproval: true}
	hold, _ := EvaluateSellPriceChange(true, rules, 100, 90, 80, true)
	if hold {
		t.Fatal("expected no hold for above-cost drop when price_drop_requires_approval is false")
	}
	hold, reason := EvaluateSellPriceChange(true, rules, 100, 75, 80, true)
	if !hold || !strings.Contains(strings.ToLower(reason), "below cost") {
		t.Fatalf("expected hold with below cost reason, got hold=%v reason=%q", hold, reason)
	}
}

func TestEvaluateSellPriceChange_Unchanged(t *testing.T) {
	rules := testCatalogPriceRules()
	hold, _ := EvaluateSellPriceChange(true, rules, 100, 100, 80, true)
	if hold {
		t.Fatal("expected no hold when unchanged")
	}
}

func TestEvaluateSellPriceChange_FirstPrice(t *testing.T) {
	rules := testCatalogPriceRules()
	hold, _ := EvaluateSellPriceChange(true, rules, 0, 50, 30, false)
	if hold {
		t.Fatal("expected no hold for first price")
	}
}

func TestEvaluateSellPriceChange_BelowCostReason(t *testing.T) {
	rules := testCatalogPriceRules()
	_, reason := EvaluateSellPriceChange(true, rules, 100, 95, 100, true)
	if reason == "" || !strings.Contains(strings.ToLower(reason), "below cost") {
		t.Fatalf("expected below cost in reason, got %q", reason)
	}
}

func TestEvaluateCostChange_Threshold(t *testing.T) {
	rules := testCatalogPriceRules()
	hold, _ := EvaluateCostChange(true, rules, 100, 105)
	if hold {
		t.Fatal("expected no hold below 10% threshold")
	}
	hold, _ = EvaluateCostChange(true, rules, 100, 115)
	if !hold {
		t.Fatal("expected hold at or above 10% threshold")
	}
	hold, _ = EvaluateCostChange(true, CatalogPriceRules{}, 100, 150)
	if hold {
		t.Fatal("expected no hold when cost_increase_threshold_pct is omitted")
	}
}
