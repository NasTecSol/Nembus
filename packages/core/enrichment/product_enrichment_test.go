package enrichment

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProposalSetRejectsProductTypeInUnsupportedValue(t *testing.T) {
	proposals := ProposalSet{UnsupportedSemantics: []UnsupportedSemantic{{
		SemanticType: "classification",
		Key:          "future_hint",
		Value:        json.RawMessage(`{"product_type":"raw_material"}`),
		Confidence:   0.9,
	}}}

	if err := proposals.Validate(json.RawMessage(`{}`)); err == nil || !strings.Contains(err.Error(), "product_type") {
		t.Fatalf("expected product_type rejection, got %v", err)
	}
}

func TestPopulatedStructuredCategoryCannotBeReplaced(t *testing.T) {
	proposals := ProposalSet{Category: &CategoryProposal{
		Action:        ActionMatchExisting,
		TargetCode:    "BEV",
		CanonicalName: "Beverages",
		Confidence:    0.95,
	}}

	err := proposals.Validate(json.RawMessage(`{"category":{"code":"EXISTING"}}`))
	if err == nil || !strings.Contains(err.Error(), "KEEP_EXISTING") {
		t.Fatalf("expected category replacement rejection, got %v", err)
	}
}

func TestMatchExistingBrandRequiresCanonicalTarget(t *testing.T) {
	proposals := ProposalSet{Brand: &BrandProposal{Action: ActionMatchExisting, Confidence: 0.8}}

	err := proposals.Validate(json.RawMessage(`{}`))
	if err == nil || !strings.Contains(err.Error(), "canonical target") {
		t.Fatalf("expected canonical target rejection, got %v", err)
	}
}

func TestProposeNewRemainsReviewProposal(t *testing.T) {
	proposal := BrandProposal{
		Action:        ActionProposeNew,
		CanonicalName: "New Brand",
		Confidence:    0.7,
	}
	if err := (ProposalSet{Brand: &proposal}).Validate(json.RawMessage(`{}`)); err != nil {
		t.Fatalf("expected reviewable proposal to validate: %v", err)
	}
	if proposal.Action != ActionProposeNew || proposal.TargetID != nil || proposal.TargetCode != "" {
		t.Fatalf("PROPOSE_NEW must remain an unbound review proposal: %+v", proposal)
	}
}

func TestUnsupportedSemanticCanBeRetainedWithoutProductColumnTarget(t *testing.T) {
	proposals := ProposalSet{UnsupportedSemantics: []UnsupportedSemantic{{
		SemanticType: "capacity",
		Key:          "package_capacity",
		Value:        json.RawMessage(`"400 ml"`),
		Confidence:   0.88,
	}}}

	if err := proposals.Validate(json.RawMessage(`{}`)); err != nil {
		t.Fatalf("expected unsupported semantic to validate: %v", err)
	}
}

func TestUnsupportedSemanticRejectsSAPAuthoritativeTargets(t *testing.T) {
	for _, target := range []string{"SKU", "ItemCode", "barcode", "inventory", "warehouse", "price", "tax", "uom_conversion", "supplier", "is_active", "is_sellable", "is_purchasable", "product_type"} {
		t.Run(target, func(t *testing.T) {
			proposals := ProposalSet{UnsupportedSemantics: []UnsupportedSemantic{{
				SemanticType: "future_hint",
				Key:          target,
				Value:        json.RawMessage(`"not a target"`),
				Confidence:   0.5,
			}}}
			if err := proposals.Validate(json.RawMessage(`{}`)); err == nil {
				t.Fatalf("expected %q to be rejected", target)
			}
		})
	}
}

func TestConfidenceMustBeBetweenZeroAndOne(t *testing.T) {
	proposals := ProposalSet{Description: &DescriptionProposal{
		Action:     ActionProposeNew,
		Value:      "A description",
		Confidence: 1.1,
	}}

	if err := proposals.Validate(json.RawMessage(`{}`)); err == nil || !strings.Contains(err.Error(), "confidence") {
		t.Fatalf("expected confidence rejection, got %v", err)
	}
}
