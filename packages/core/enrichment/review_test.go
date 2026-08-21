package enrichment

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestCanApproveSuggestionEnforcesWholeSuggestionPolicy(t *testing.T) {
	current := reviewTestSnapshot()
	current.Category = &CategoryIdentity{ID: 3, Code: "CAT-3", Name: "Category 3"}
	row := reviewTestRecord(t, current, &BrandProposal{
		Action: ActionMatchExisting, TargetID: int32Ptr(7), Confidence: 0.9,
	}, &CategoryProposal{Action: ActionKeepExisting, Confidence: 1}, &DescriptionProposal{
		Action: ActionProposeNew, Value: "A safe description", Confidence: 0.8,
	})
	fingerprint, _ := FingerprintSnapshot(current)
	valid := CanApproveSuggestion(row, current, fingerprint, &ReviewCanonicalTarget{ID: 7, Code: "BRAND-7", Name: "Brand 7"}, nil)
	if !valid.Approvable || len(valid.BlockingReasons) != 0 {
		t.Fatalf("expected valid whole suggestion, got %+v", valid)
	}

	row.ProposedBrand, _ = json.Marshal(&BrandProposal{Action: ActionProposeNew, CanonicalName: "New Brand", Confidence: 0.9})
	blocked := CanApproveSuggestion(row, current, fingerprint, nil, nil)
	if blocked.Approvable || !contains(blocked.BlockingReasons, "brand_propose_new_requires_review_resolution") {
		t.Fatalf("expected brand PROPOSE_NEW block, got %+v", blocked)
	}

	current.Brand = &BrandIdentity{ID: 8, Code: "STRUCTURED", Name: "Structured"}
	row = reviewTestRecord(t, current, &BrandProposal{Action: ActionMatchExisting, TargetID: int32Ptr(7), Confidence: 0.9}, nil, nil)
	fingerprint, _ = FingerprintSnapshot(current)
	precedence := CanApproveSuggestion(row, current, fingerprint, &ReviewCanonicalTarget{ID: 7, Code: "BRAND-7"}, nil)
	if precedence.Approvable || !contains(precedence.BlockingReasons, "structured_brand_precedence_violation") {
		t.Fatalf("expected structured brand precedence block, got %+v", precedence)
	}
}

func TestReviewStaleReasonsAndOperationalChanges(t *testing.T) {
	prior := reviewTestSnapshot()
	row := reviewTestRecord(t, prior, nil, nil, nil)
	fingerprint, _ := FingerprintSnapshot(prior)

	nameChanged := prior
	nameChanged.SourceItemName = "Changed item"
	changedFingerprint, _ := FingerprintSnapshot(nameChanged)
	analysis := CanApproveSuggestion(row, nameChanged, changedFingerprint, nil, nil)
	if !analysis.Stale || !contains(analysis.StaleReasons, "source_item_name_changed") {
		t.Fatalf("expected source name stale result, got %+v", analysis)
	}

	// Inventory and pricing are not fields in the Stage 2A snapshot, so they
	// cannot alter the canonical fingerprint.
	unchangedFingerprint, err := FingerprintSnapshot(prior)
	if err != nil || unchangedFingerprint != fingerprint {
		t.Fatalf("expected operationally unrelated state to preserve fingerprint")
	}
}

func TestReviewServiceApproveIsAtomicAndDoesNotApplyProduct(t *testing.T) {
	current := reviewTestSnapshot()
	row := reviewTestRecord(t, current, nil, nil, &DescriptionProposal{Action: ActionProposeNew, Value: "Description", Confidence: 0.9})
	store := &fakeReviewStore{row: row, current: current, tx: &fakeReviewTx{approved: row}}
	service := NewReviewService(store)
	detail, err := service.ApproveSuggestion(context.Background(), 1, 1, 5)
	if err != nil {
		t.Fatalf("approve failed: %v", err)
	}
	if detail.ReviewState.Status != SuggestionStatusApproved || detail.ReviewState.AppliedAt != nil {
		t.Fatalf("expected approved but unapplied state, got %+v", detail.ReviewState)
	}
	if store.tx.audit.NewStatus != SuggestionStatusApproved || !store.tx.committed {
		t.Fatalf("expected atomic approval audit and commit, got %+v", store.tx)
	}
	if store.current.Description != current.Description || store.current.Brand != current.Brand || store.current.Category != current.Category {
		t.Fatalf("review service mutated authoritative product snapshot")
	}

	store.tx = &fakeReviewTx{approved: row, auditErr: errors.New("audit unavailable")}
	if _, err := service.ApproveSuggestion(context.Background(), 1, 1, 5); err == nil {
		t.Fatal("expected audit failure")
	}
	if store.tx.committed {
		t.Fatal("audit failure must prevent transaction commit")
	}
}

func TestReviewServiceRejectAllowsStaleAndProposeNew(t *testing.T) {
	prior := reviewTestSnapshot()
	row := reviewTestRecord(t, prior, &BrandProposal{Action: ActionProposeNew, CanonicalName: "New", Confidence: 0.8}, nil, nil)
	current := prior
	current.SourceItemName = "Changed"
	store := &fakeReviewStore{row: row, current: current, tx: &fakeReviewTx{rejected: row}}
	service := NewReviewService(store)
	detail, err := service.RejectSuggestion(context.Background(), 1, 1, 5)
	if err != nil {
		t.Fatalf("stale reject failed: %v", err)
	}
	if detail.ReviewState.Status != SuggestionStatusRejected || store.tx.audit.NewStatus != SuggestionStatusRejected {
		t.Fatalf("expected rejected stale suggestion, got %+v", detail)
	}
}

func TestReviewDetailProjectionDoesNotExposeRawInternalState(t *testing.T) {
	current := reviewTestSnapshot()
	current.UOM.Base = &UOMIdentity{ID: 4, Code: "EA", Name: "Each"}
	row := reviewTestRecord(t, current, nil, nil, nil)
	detail := detailFromRecord(row, current, ReviewAnalysis{})
	encoded, err := json.Marshal(detail)
	if err != nil {
		t.Fatal(err)
	}
	body := string(encoded)
	for _, forbidden := range []string{"structured_current", "metadata", "inventory", "price", "tax", "supplier", "api_key", "raw_prompt", "raw_response"} {
		if containsString(body, forbidden) {
			t.Fatalf("detail projection exposed forbidden field %q: %s", forbidden, body)
		}
	}
}

func reviewTestSnapshot() EnrichmentSourceSnapshot {
	return EnrichmentSourceSnapshot{
		OrganizationID: 1, ProductID: 95, SourceSystem: SourceSystemSAP,
		SourceItemCode: "INV00095", SourceItemName: "Test product",
		ProductType: ProductTypeStandard,
	}
}

func reviewTestRecord(t *testing.T, snapshot EnrichmentSourceSnapshot, brand *BrandProposal, category *CategoryProposal, description *DescriptionProposal) ReviewSuggestionRecord {
	t.Helper()
	structured, err := StructuredCurrent(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	fingerprint, err := FingerprintSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	brandJSON, _ := json.Marshal(brand)
	categoryJSON, _ := json.Marshal(category)
	descriptionJSON, _ := json.Marshal(description)
	return ReviewSuggestionRecord{
		ID: 1, OrganizationID: 1, ProductID: snapshot.ProductID, SourceItemCode: snapshot.SourceItemCode,
		SourceItemName: snapshot.SourceItemName, SourceDataFingerprint: fingerprint, ContractVersion: EnrichmentContractVersion,
		StructuredCurrent: structured, ProposedBrand: brandJSON, ProposedCategory: categoryJSON, ProposedDescription: descriptionJSON,
		Status: SuggestionStatusInReview,
	}
}

type fakeReviewStore struct {
	row     ReviewSuggestionRecord
	current EnrichmentSourceSnapshot
	tx      *fakeReviewTx
}

func (f *fakeReviewStore) ListReviewSuggestions(context.Context, int32, ReviewListStatus, int32, int32) ([]ReviewSuggestionRecord, error) {
	return []ReviewSuggestionRecord{f.row}, nil
}
func (f *fakeReviewStore) GetReviewSuggestion(context.Context, int32, int32) (ReviewSuggestionRecord, error) {
	return f.row, nil
}
func (f *fakeReviewStore) LoadSAPProductEnrichmentSnapshotByID(context.Context, int32, int32) (EnrichmentSourceSnapshot, error) {
	return f.current, nil
}
func (f *fakeReviewStore) ResolveBrandReviewTarget(context.Context, *int32, string) (*ReviewCanonicalTarget, error) {
	return f.tx.brandTarget, nil
}
func (f *fakeReviewStore) ResolveCategoryReviewTarget(context.Context, *int32, string) (*ReviewCanonicalTarget, error) {
	return f.tx.categoryTarget, nil
}
func (f *fakeReviewStore) BeginReviewTransaction(context.Context) (ReviewTransaction, error) {
	return f.tx, nil
}

type fakeReviewTx struct {
	approved, rejected          ReviewSuggestionRecord
	brandTarget, categoryTarget *ReviewCanonicalTarget
	audit                       ReviewAudit
	auditErr                    error
	committed                   bool
}

func (f *fakeReviewTx) Approve(context.Context, int32, int32, int32) (ReviewSuggestionRecord, error) {
	f.approved.Status = SuggestionStatusApproved
	f.approved.AppliedAt = nil
	return f.approved, nil
}
func (f *fakeReviewTx) Reject(context.Context, int32, int32, int32) (ReviewSuggestionRecord, error) {
	f.rejected.Status = SuggestionStatusRejected
	return f.rejected, nil
}
func (f *fakeReviewTx) InsertAudit(_ context.Context, audit ReviewAudit) error {
	f.audit = audit
	return f.auditErr
}
func (f *fakeReviewTx) Commit(context.Context) error   { f.committed = true; return nil }
func (f *fakeReviewTx) Rollback(context.Context) error { return nil }

func int32Ptr(value int32) *int32 { return &value }
func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func containsString(value, expected string) bool {
	for i := 0; i+len(expected) <= len(value); i++ {
		if value[i:i+len(expected)] == expected {
			return true
		}
	}
	return false
}
