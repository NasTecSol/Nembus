package enrichment

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestEvaluateEligibilityMVPGaps(t *testing.T) {
	cases := []struct {
		name         string
		input        EnrichmentEligibilityInput
		wantEligible bool
		wantGap      EnrichmentGap
	}{
		{
			name: "missing brand",
			input: EnrichmentEligibilityInput{
				SourceSystem: SourceSystemSAP, OrganizationID: 1, ProductID: 2,
				SourceItemCode: "A", SourceItemName: "A", ProductType: ProductTypeStandard,
				Description: "desc", Category: &CategoryIdentity{ID: 10},
			},
			wantEligible: true, wantGap: GapMissingBrand,
		},
		{
			name: "empty description",
			input: EnrichmentEligibilityInput{
				SourceSystem: SourceSystemSAP, OrganizationID: 1, ProductID: 2,
				SourceItemCode: "A", SourceItemName: "A", ProductType: ProductTypeFixedAsset,
				Brand: &BrandIdentity{ID: 10}, Category: &CategoryIdentity{ID: 11},
			},
			wantEligible: true, wantGap: GapMissingDescription,
		},
		{
			name: "missing category",
			input: EnrichmentEligibilityInput{
				SourceSystem: SourceSystemSAP, OrganizationID: 1, ProductID: 2,
				SourceItemCode: "A", SourceItemName: "A", ProductType: ProductTypeRawMaterial,
				Brand: &BrandIdentity{Code: "BRAND"}, Description: "desc",
			},
			wantEligible: true, wantGap: GapMissingCategory,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			decision := EvaluateEligibility(tc.input)
			if decision.Eligible != tc.wantEligible {
				t.Fatalf("eligible = %v, want %v: %+v", decision.Eligible, tc.wantEligible, decision)
			}
			if len(decision.Gaps) != 1 || decision.Gaps[0] != tc.wantGap {
				t.Fatalf("gaps = %#v, want %#v", decision.Gaps, []EnrichmentGap{tc.wantGap})
			}
		})
	}
}

func TestEvaluateEligibilityResolvedAndInvalidSource(t *testing.T) {
	complete := EvaluateEligibility(EnrichmentEligibilityInput{
		SourceSystem: SourceSystemSAP, OrganizationID: 1, ProductID: 2,
		SourceItemCode: "A", SourceItemName: "A", ProductType: ProductTypeFinishedGood,
		Brand: &BrandIdentity{ID: 10}, Category: &CategoryIdentity{Code: "CAT"}, Description: "desc",
	})
	if complete.Eligible || len(complete.Gaps) != 0 || len(complete.Reasons) != 1 || complete.Reasons[0] != ReasonNoEnrichmentGaps {
		t.Fatalf("expected complete product to be ineligible without gaps: %+v", complete)
	}

	for _, tc := range []struct {
		name   string
		input  EnrichmentEligibilityInput
		reason EligibilityReason
	}{
		{name: "invalid product type", input: EnrichmentEligibilityInput{SourceSystem: SourceSystemSAP, OrganizationID: 1, ProductID: 2, SourceItemCode: "A", SourceItemName: "A", ProductType: "electronics"}, reason: ReasonInvalidProductType},
		{name: "missing item code", input: EnrichmentEligibilityInput{SourceSystem: SourceSystemSAP, OrganizationID: 1, ProductID: 2, SourceItemName: "A", ProductType: ProductTypeStandard}, reason: ReasonMissingItemCode},
		{name: "missing item name", input: EnrichmentEligibilityInput{SourceSystem: SourceSystemSAP, OrganizationID: 1, ProductID: 2, SourceItemCode: "A", ProductType: ProductTypeStandard}, reason: ReasonMissingItemName},
	} {
		t.Run(tc.name, func(t *testing.T) {
			decision := EvaluateEligibility(tc.input)
			if decision.Eligible || !decision.Rejected && tc.reason == ReasonInvalidProductType {
				t.Fatalf("unexpected eligibility/rejection result: %+v", decision)
			}
			if !containsReason(decision.Reasons, tc.reason) {
				t.Fatalf("reasons = %#v, want %q", decision.Reasons, tc.reason)
			}
		})
	}
}

func TestFingerprintSnapshotIsStableAndSourceRelevant(t *testing.T) {
	base := testSnapshot()
	fingerprint, err := FingerprintSnapshot(base)
	if err != nil {
		t.Fatal(err)
	}

	reordered := base
	reordered.UOM.Conversions = []UOMConversionContext{base.UOM.Conversions[1], base.UOM.Conversions[0]}
	reorderedFingerprint, err := FingerprintSnapshot(reordered)
	if err != nil {
		t.Fatal(err)
	}
	if fingerprint != reorderedFingerprint {
		t.Fatal("conversion representation order changed the fingerprint")
	}

	changes := []struct {
		name   string
		mutate func(*EnrichmentSourceSnapshot)
	}{
		{name: "item name", mutate: func(s *EnrichmentSourceSnapshot) { s.SourceItemName = "Changed" }},
		{name: "description", mutate: func(s *EnrichmentSourceSnapshot) { s.Description = "Changed" }},
		{name: "brand", mutate: func(s *EnrichmentSourceSnapshot) { s.Brand.Code = "OTHER" }},
		{name: "category", mutate: func(s *EnrichmentSourceSnapshot) { s.Category.Code = "OTHER" }},
		{name: "product type", mutate: func(s *EnrichmentSourceSnapshot) { s.ProductType = ProductTypeFixedAsset }},
		{name: "UoM context", mutate: func(s *EnrichmentSourceSnapshot) { s.UOM.Conversions[0].ConversionFactor = "99" }},
	}
	for _, tc := range changes {
		t.Run(tc.name, func(t *testing.T) {
			changed := testSnapshot()
			tc.mutate(&changed)
			changedFingerprint, err := FingerprintSnapshot(changed)
			if err != nil {
				t.Fatal(err)
			}
			if changedFingerprint == fingerprint {
				t.Fatal("relevant source change did not change fingerprint")
			}
		})
	}

	encoded, err := json.Marshal(fingerprintSource{UOM: base.UOM})
	if err != nil {
		t.Fatal(err)
	}
	for _, excluded := range []string{"inventory", "price", "tax", "supplier", "barcode", "warehouse"} {
		if strings.Contains(string(encoded), excluded) {
			t.Fatalf("fingerprint source unexpectedly contains %q: %s", excluded, encoded)
		}
	}
}

func TestCoordinatorCreatesIdempotentPendingSuggestion(t *testing.T) {
	store := &fakeEnrichmentStore{snapshot: testSnapshot()}
	coordinator := NewProductEnrichmentCoordinator(store)

	first, err := coordinator.EnqueueSAPProduct(context.Background(), 1, "SAP-1")
	if err != nil {
		t.Fatal(err)
	}
	if !first.Decision.Eligible || first.Suggestion.Status != SuggestionStatusPending || store.createCalls != 1 {
		t.Fatalf("unexpected first enqueue: result=%+v calls=%d", first, store.createCalls)
	}

	store.existing.Status = SuggestionStatusApproved
	second, err := coordinator.EnqueueSAPProduct(context.Background(), 1, "SAP-1")
	if err != nil {
		t.Fatal(err)
	}
	if second.Suggestion.Status != SuggestionStatusApproved || store.createCalls != 2 {
		t.Fatalf("existing lifecycle was not preserved: result=%+v calls=%d", second, store.createCalls)
	}
}

func TestCoordinatorSkipsIneligibleAndSurfacesStoreFailure(t *testing.T) {
	ineligibleStore := &fakeEnrichmentStore{snapshot: testSnapshot()}
	ineligibleStore.snapshot.Description = "already complete"
	result, err := NewProductEnrichmentCoordinator(ineligibleStore).EnqueueSAPProduct(context.Background(), 1, "SAP-1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Decision.Eligible || ineligibleStore.createCalls != 0 {
		t.Fatalf("ineligible product was enqueued: result=%+v calls=%d", result, ineligibleStore.createCalls)
	}

	failure := errors.New("store unavailable")
	failingStore := &fakeEnrichmentStore{snapshot: testSnapshot(), loadErr: failure}
	if _, err := NewProductEnrichmentCoordinator(failingStore).EnqueueSAPProduct(context.Background(), 1, "SAP-1"); err == nil || !strings.Contains(err.Error(), failure.Error()) {
		t.Fatalf("expected store failure, got %v", err)
	}
}

func TestNormalizeProposedDescription(t *testing.T) {
	normalized, err := NormalizeProposedDescription("  وصف catalog  ")
	if err != nil || normalized != "وصف catalog" {
		t.Fatalf("normalized = %q, err = %v", normalized, err)
	}
	long := strings.Repeat("a", MaxDescriptionRunes+1)
	if _, err := NormalizeProposedDescription(long); err == nil {
		t.Fatal("expected overlong description to fail")
	}
}

type fakeEnrichmentStore struct {
	snapshot    EnrichmentSourceSnapshot
	existing    PendingSuggestion
	loadErr     error
	createErr   error
	createCalls int
}

func (s *fakeEnrichmentStore) LoadSAPProductEnrichmentSnapshot(_ context.Context, organizationID int32, _ string) (EnrichmentSourceSnapshot, error) {
	if s.loadErr != nil {
		return EnrichmentSourceSnapshot{}, s.loadErr
	}
	snapshot := s.snapshot
	snapshot.OrganizationID = organizationID
	return snapshot, nil
}

func (s *fakeEnrichmentStore) CreateOrGetPendingSuggestion(_ context.Context, _ PendingSuggestionInput) (PendingSuggestion, error) {
	s.createCalls++
	if s.createErr != nil {
		return PendingSuggestion{}, s.createErr
	}
	if s.existing.ID == 0 {
		s.existing = PendingSuggestion{ID: 101, Status: SuggestionStatusPending}
	}
	return s.existing, nil
}

func testSnapshot() EnrichmentSourceSnapshot {
	return EnrichmentSourceSnapshot{
		OrganizationID: 1,
		ProductID:      2,
		SourceSystem:   SourceSystemSAP,
		SourceItemCode: "SAP-1",
		SourceItemName: "Source name",
		Description:    "",
		ProductType:    ProductTypeStandard,
		Brand:          &BrandIdentity{ID: 10, Code: "BRAND", Name: "Brand"},
		Category:       &CategoryIdentity{ID: 20, Code: "CAT", Name: "Category", Path: []string{"Category"}},
		UOM: UOMContext{Base: &UOMIdentity{ID: 30, Code: "EA", Name: "Each"}, Conversions: []UOMConversionContext{
			{From: UOMIdentity{ID: 30, Code: "EA", Name: "Each"}, To: UOMIdentity{ID: 31, Code: "BOX", Name: "Box"}, ConversionFactor: "12", IsDefault: true},
			{From: UOMIdentity{ID: 31, Code: "BOX", Name: "Box"}, To: UOMIdentity{ID: 30, Code: "EA", Name: "Each"}, ConversionFactor: "0.083333", IsDefault: false},
		}},
	}
}

func containsReason(reasons []EligibilityReason, want EligibilityReason) bool {
	for _, reason := range reasons {
		if reason == want {
			return true
		}
	}
	return false
}
