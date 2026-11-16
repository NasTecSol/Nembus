package usecase

import (
	"fmt"
	"math"
)

const catalogPriceEpsilon = 0.0001

// CatalogPriceRules holds parsed approval_policies.rules for catalog.price_change.
type CatalogPriceRules struct {
	PriceDropRequiresApproval bool
	BelowCostRequiresApproval bool
	HasCostIncreaseThreshold  bool
	CostIncreaseThresholdPct  float64
}

// ParseCatalogPriceRules reads rule map values from approval_policies.rules JSONB.
func ParseCatalogPriceRules(rules map[string]any) CatalogPriceRules {
	var out CatalogPriceRules
	if rules == nil {
		return out
	}
	if v, ok := ruleBool(rules, "price_drop_requires_approval"); ok {
		out.PriceDropRequiresApproval = v
	}
	if v, ok := ruleBool(rules, "below_cost_requires_approval"); ok {
		out.BelowCostRequiresApproval = v
	}
	if v, ok := ruleFloat(rules, "cost_increase_threshold_pct"); ok {
		out.HasCostIncreaseThreshold = true
		out.CostIncreaseThresholdPct = v
	}
	return out
}

func ruleBool(rules map[string]any, key string) (bool, bool) {
	v, ok := rules[key]
	if !ok {
		return false, false
	}
	switch t := v.(type) {
	case bool:
		return t, true
	default:
		return false, false
	}
}

func ruleFloat(rules map[string]any, key string) (float64, bool) {
	v, ok := rules[key]
	if !ok {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	default:
		return 0, false
	}
}

// EvaluateSellPriceChange decides whether a proposed sell price must go through approval
// before updating product_prices. Initial prices (no active row) are applied directly.
func EvaluateSellPriceChange(requiresApproval bool, rules CatalogPriceRules, oldPrice, newPrice, cost float64, hasActivePrice bool) (hold bool, reason string) {
	if !requiresApproval || !hasActivePrice || oldPrice <= 0 {
		return false, ""
	}
	if math.Abs(newPrice-oldPrice) <= catalogPriceEpsilon {
		return false, ""
	}

	belowCost := rules.BelowCostRequiresApproval && cost > 0 && newPrice < cost

	if newPrice > oldPrice {
		pct := ((newPrice - oldPrice) / oldPrice) * 100.0
		if rules.HasCostIncreaseThreshold && pct < rules.CostIncreaseThresholdPct {
			if belowCost {
				return true, fmt.Sprintf("Selling price %.2f is below cost price %.2f", newPrice, cost)
			}
			return false, ""
		}
		if rules.HasCostIncreaseThreshold {
			reason = fmt.Sprintf("Selling price increased by %.1f%% from %.2f to %.2f (threshold: %.1f%%)", pct, oldPrice, newPrice, rules.CostIncreaseThresholdPct)
		} else {
			reason = fmt.Sprintf("Selling price increased from %.2f to %.2f", oldPrice, newPrice)
		}
		if belowCost {
			reason = fmt.Sprintf("%s; selling price %.2f is below cost price %.2f", reason, newPrice, cost)
		}
		return true, reason
	}

	if !rules.PriceDropRequiresApproval && !belowCost {
		return false, ""
	}
	if rules.PriceDropRequiresApproval {
		reason = fmt.Sprintf("Selling price dropped from %.2f to %.2f", oldPrice, newPrice)
		if belowCost {
			reason = fmt.Sprintf("%s; selling price %.2f is below cost price %.2f", reason, newPrice, cost)
		}
		return true, reason
	}
	reason = fmt.Sprintf("Selling price dropped from %.2f to %.2f; selling price %.2f is below cost price %.2f", oldPrice, newPrice, newPrice, cost)
	return true, reason
}

// EvaluateCostChange decides whether a proposed cost must be held for approval before
// updating products.cost_price.
func EvaluateCostChange(requiresApproval bool, rules CatalogPriceRules, oldCost, newCost float64) (hold bool, reason string) {
	if !requiresApproval || !rules.HasCostIncreaseThreshold || newCost <= 0 || oldCost <= 0 || newCost <= oldCost {
		return false, ""
	}
	pct := ((newCost - oldCost) / oldCost) * 100.0
	if pct < rules.CostIncreaseThresholdPct {
		return false, ""
	}
	reason = fmt.Sprintf("Cost price increased by %.1f%% (from %.2f to %.2f, threshold: %.1f%%)",
		pct, oldCost, newCost, rules.CostIncreaseThresholdPct)
	return true, reason
}
