package enrichment

import (
	"encoding/json"
	"math"
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

func TestUnsupportedSemanticTargetNormalization(t *testing.T) {
	for _, target := range []string{
		"product_type", "productType", "ProductType", "product-type",
		"ItemCode", "itemCode", "source_item_code", "sap_item_code",
		"uom_conversion", "uomConversion", "UOM-conversion", "conversion factor",
		"isActive", "is-sellable", "isPurchasable", "barcode ownership", "store inventory",
	} {
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

func TestUnsupportedSemanticPreservesNonAuthoritativeEvidence(t *testing.T) {
	for _, key := range []string{"anti_dandruff", "capacity", "dimensions", "resolution", "size_text", "model_number"} {
		t.Run(key, func(t *testing.T) {
			proposals := ProposalSet{UnsupportedSemantics: []UnsupportedSemantic{{
				SemanticType: "product_evidence",
				Key:          key,
				Value:        json.RawMessage(`"evidence"`),
				Confidence:   0.5,
			}}}
			if err := proposals.Validate(json.RawMessage(`{}`)); err != nil {
				t.Fatalf("expected %q to remain valid evidence: %v", key, err)
			}
		})
	}
}

func TestResolvedStructuredBrandOnlyKeepsExisting(t *testing.T) {
	proposals := ProposalSet{Brand: &BrandProposal{Action: ActionKeepExisting, Confidence: 0.8}}
	if err := proposals.Validate(json.RawMessage(`{"brand_id":42,"brand_code":"PANTENE"}`)); err != nil {
		t.Fatalf("expected KEEP_EXISTING for a resolved brand to validate: %v", err)
	}
}

func TestResolvedStructuredBrandCannotBeReplaced(t *testing.T) {
	for _, action := range []ProposalAction{ActionMatchExisting, ActionProposeNew, ActionNoMatch} {
		t.Run(string(action), func(t *testing.T) {
			proposal := &BrandProposal{Action: action, Confidence: 0.8}
			if action == ActionMatchExisting {
				proposal.TargetCode = "OTHER"
			} else if action == ActionProposeNew {
				proposal.CanonicalName = "Other Brand"
			}
			err := (ProposalSet{Brand: proposal}).Validate(json.RawMessage(`{"brand_id":42,"brand_code":"PANTENE"}`))
			if err == nil || !strings.Contains(err.Error(), "KEEP_EXISTING") {
				t.Fatalf("expected resolved brand precedence rejection for %s, got %v", action, err)
			}
		})
	}
}

func TestUnresolvedStructuredBrandAllowsMatchingOrReviewProposal(t *testing.T) {
	match := ProposalSet{Brand: &BrandProposal{
		Action:     ActionMatchExisting,
		TargetCode: "PANTENE",
		Confidence: 0.8,
	}}
	if err := match.Validate(json.RawMessage(`{"brand_id":null,"brand_code":""}`)); err != nil {
		t.Fatalf("expected unresolved brand match to validate: %v", err)
	}

	propose := ProposalSet{Brand: &BrandProposal{
		Action:        ActionProposeNew,
		CanonicalName: "New Brand",
		Confidence:    0.8,
	}}
	if err := propose.Validate(json.RawMessage(`{}`)); err != nil {
		t.Fatalf("expected unresolved brand proposal to remain reviewable: %v", err)
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

func TestConfidenceAcceptsBoundsAndRejectsNonFiniteValues(t *testing.T) {
	for _, confidence := range []float64{0, 1} {
		proposal := ProposalSet{Description: &DescriptionProposal{
			Action:     ActionProposeNew,
			Value:      "A description",
			Confidence: confidence,
		}}
		if err := proposal.Validate(json.RawMessage(`{}`)); err != nil {
			t.Fatalf("expected confidence %v to validate: %v", confidence, err)
		}
	}

	for _, confidence := range []float64{-0.01, 1.01, math.NaN(), math.Inf(1), math.Inf(-1)} {
		proposal := ProposalSet{Description: &DescriptionProposal{
			Action:     ActionProposeNew,
			Value:      "A description",
			Confidence: confidence,
		}}
		if err := proposal.Validate(json.RawMessage(`{}`)); err == nil || !strings.Contains(err.Error(), "confidence") {
			t.Fatalf("expected confidence %v to be rejected, got %v", confidence, err)
		}
	}
}
