package usecase

import "fmt"

// CatalogPricePolicyRules mirrors approval_policies.rules for catalog.price_change.
type CatalogPricePolicyRules struct {
	RequiresApproval            bool
	PriceDropRequiresApproval   bool
	CostIncreaseThresholdPct    float64
	BelowCostRequiresApproval   bool
}

func ParseCatalogPricePolicy(rules map[string]any, requiresApproval bool) CatalogPricePolicyRules {
	out := CatalogPricePolicyRules{
		RequiresApproval:          requiresApproval,
		PriceDropRequiresApproval: true,
		CostIncreaseThresholdPct:  10.0,
		BelowCostRequiresApproval: true,
	}
	if !requiresApproval || rules == nil {
		return out
	}
	if v, ok := rules["price_drop_requires_approval"].(bool); ok {
		out.PriceDropRequiresApproval = v
	}
	if v, ok := rules["cost_increase_threshold_pct"].(float64); ok {
		out.CostIncreaseThresholdPct = v
	}
	if v, ok := rules["below_cost_requires_approval"].(bool); ok {
		out.BelowCostRequiresApproval = v
	}
	return out
}

func (r CatalogPricePolicyRules) CostIncreaseNeedsReview(currentCost, newCost float64) (bool, string) {
	if !r.RequiresApproval || currentCost <= 0 || newCost <= currentCost {
		return false, ""
	}
	pct := ((newCost - currentCost) / currentCost) * 100.0
	if pct < r.CostIncreaseThresholdPct {
		return false, ""
	}
	return true, fmt.Sprintf("Cost price increased by %.1f%% (from %.2f to %.2f, threshold: %.1f%%)", pct, currentCost, newCost, r.CostIncreaseThresholdPct)
}

func (r CatalogPricePolicyRules) PriceChangeNeedsReview(oldPrice, newPrice, costPrice float64, priceExists bool) (bool, string) {
	if !r.RequiresApproval || !priceExists || oldPrice <= 0 {
		return false, ""
	}
	if r.PriceDropRequiresApproval && newPrice < oldPrice {
		return true, fmt.Sprintf("Selling price dropped from %.2f to %.2f", oldPrice, newPrice)
	}
	if r.BelowCostRequiresApproval && costPrice > 0 && newPrice < costPrice {
		return true, fmt.Sprintf("Selling price %.2f is below cost price %.2f", newPrice, costPrice)
	}
	return false, ""
}

// PosTxEligibleForSapOutbox mirrors StreamSapOutbox SQL eligibility for pos_transactions.
func PosTxEligibleForSapOutbox(sapDocNum, sapSyncState string, sapForcePost bool) bool {
	if sapDocNum != "" {
		return false
	}
	if sapForcePost {
		return true
	}
	if sapSyncState == "" {
		return true
	}
	switch sapSyncState {
	case "held_stock", "adjustment_created_pending_force":
		return false
	default:
		return true
	}
}
