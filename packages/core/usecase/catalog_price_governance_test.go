package usecase

import (
	"strings"
	"testing"
)

func TestEvaluateSellPriceChange_IncreaseRequiresHold(t *testing.T) {
	rules := DefaultCatalogPriceRules()
	hold, reason := EvaluateSellPriceChange(true, rules, 100, 120, 80, true)
	if !hold {
		t.Fatal("expected hold on price increase")
	}
	if reason == "" {
		t.Fatal("expected reason")
	}
}

func TestEvaluateSellPriceChange_DecreaseRequiresHold(t *testing.T) {
	rules := DefaultCatalogPriceRules()
	hold, _ := EvaluateSellPriceChange(true, rules, 100, 90, 80, true)
	if !hold {
		t.Fatal("expected hold on price decrease")
	}
}

func TestEvaluateSellPriceChange_Unchanged(t *testing.T) {
	rules := DefaultCatalogPriceRules()
	hold, _ := EvaluateSellPriceChange(true, rules, 100, 100, 80, true)
	if hold {
		t.Fatal("expected no hold when unchanged")
	}
}

func TestEvaluateSellPriceChange_FirstPrice(t *testing.T) {
	rules := DefaultCatalogPriceRules()
	hold, _ := EvaluateSellPriceChange(true, rules, 0, 50, 30, false)
	if hold {
		t.Fatal("expected no hold for first price")
	}
}

func TestEvaluateSellPriceChange_BelowCostReason(t *testing.T) {
	rules := DefaultCatalogPriceRules()
	_, reason := EvaluateSellPriceChange(true, rules, 100, 95, 100, true)
	if reason == "" || !strings.Contains(strings.ToLower(reason), "below cost") {
		t.Fatalf("expected below cost in reason, got %q", reason)
	}
}

func TestEvaluateCostChange_Threshold(t *testing.T) {
	rules := DefaultCatalogPriceRules()
	hold, _ := EvaluateCostChange(true, rules, 100, 105)
	if hold {
		t.Fatal("expected no hold below 10% threshold")
	}
	hold, _ = EvaluateCostChange(true, rules, 100, 115)
	if !hold {
		t.Fatal("expected hold at or above 10% threshold")
	}
}
