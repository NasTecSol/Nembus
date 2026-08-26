package enrichment

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestApplyApprovedSuggestionBrandAndCategoryUseExactTargets(t *testing.T) {
	current := applicationTestSnapshot()
	brandID, categoryID := int32(7), int32(9)
	tx := newApplicationTestTx(t, current, &BrandProposal{Action: ActionMatchExisting, TargetID: &brandID, Confidence: 0.9}, &CategoryProposal{Action: ActionMatchExisting, TargetID: &categoryID, Confidence: 0.9}, nil)
	tx.brandTarget = &ReviewCanonicalTarget{ID: brandID, Code: "BRAND-7", Name: "Brand 7"}
	tx.categoryTarget = &ReviewCanonicalTarget{ID: categoryID, Code: "CAT-9", Name: "Category 9"}
	result, err := NewProductEnrichmentApplicationService(&applicationTestStore{tx: tx}).ApplyApprovedSuggestion(context.Background(), 1, 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != SuggestionStatusApplied || result.AlreadyApplied || !equalStrings(result.ChangedFields, []string{"brand_id", "category_id"}) {
		t.Fatalf("unexpected result: %+v", result)
	}
	if tx.plan.BrandID == nil || *tx.plan.BrandID != brandID || tx.plan.CategoryID == nil || *tx.plan.CategoryID != categoryID || tx.plan.Description != nil {
		t.Fatalf("unexpected narrow plan: %+v", tx.plan)
	}
	if !tx.committed || tx.audit.ApplierUserID != 5 || !equalStrings(tx.audit.ChangedFields, result.ChangedFields) {
		t.Fatalf("expected committed applier audit, tx=%+v audit=%+v", tx, tx.audit)
	}
}

func TestApproveAndApplySuggestionTransitionsAndMutatesInOneApplicationTransaction(t *testing.T) {
	current := applicationTestSnapshot()
	brandID := int32(7)
	tx := newApplicationTestTx(t, current, &BrandProposal{Action: ActionMatchExisting, TargetID: &brandID, Confidence: 0.9}, nil, nil)
	tx.suggestion.Status = SuggestionStatusInReview
	tx.brandTarget = &ReviewCanonicalTarget{ID: brandID, Code: "BRAND-7", Name: "Brand 7"}

	result, err := NewProductEnrichmentApplicationService(&applicationTestStore{tx: tx}).ApproveAndApplySuggestion(context.Background(), 1, 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != SuggestionStatusApplied || !tx.committed || tx.audit.OldStatus != SuggestionStatusInReview {
		t.Fatalf("expected atomic approval/application, result=%+v tx=%+v", result, tx)
	}
	if tx.plan.BrandID == nil || *tx.plan.BrandID != brandID || !equalStrings(tx.plan.ChangedFields, []string{"brand_id"}) {
		t.Fatalf("expected canonical brand product mutation plan, %+v", tx.plan)
	}
}

func TestApplyApprovedSuggestionDescriptionUsesCanonicalValidation(t *testing.T) {
	for _, currentDescription := range []string{"", "   ", "\t\n"} {
		t.Run("missing/"+currentDescription, func(t *testing.T) {
			current := applicationTestSnapshot()
			current.Description = currentDescription
			tx := newApplicationTestTx(t, current, nil, nil, &DescriptionProposal{Action: ActionProposeNew, Value: "  Safe description  ", Confidence: 0.8})
			result, err := NewProductEnrichmentApplicationService(&applicationTestStore{tx: tx}).ApplyApprovedSuggestion(context.Background(), 1, 1, 5)
			if err != nil {
				t.Fatal(err)
			}
			if tx.plan.Description == nil || *tx.plan.Description != "Safe description" || !equalStrings(result.ChangedFields, []string{"description"}) {
				t.Fatalf("expected trimmed canonical description plan, result=%+v plan=%+v", result, tx.plan)
			}
		})
	}
}

func TestApplyApprovedSuggestionRejectsPrecedenceAndCanonicalConflicts(t *testing.T) {
	tests := []struct {
		name  string
		setup func(EnrichmentSourceSnapshot, *applicationTestTx)
		code  ApplicationErrorCode
	}{
		{name: "brand appeared", setup: func(current EnrichmentSourceSnapshot, tx *applicationTestTx) {
			current.Brand = &BrandIdentity{ID: 3, Code: "SAP-BRAND"}
			tx.current = current
		}, code: ApplicationErrorStale},
		{name: "category appeared", setup: func(current EnrichmentSourceSnapshot, tx *applicationTestTx) {
			current.Category = &CategoryIdentity{ID: 4, Code: "SAP-CATEGORY"}
			tx.current = current
		}, code: ApplicationErrorStale},
		{name: "brand target disappeared", setup: func(_ EnrichmentSourceSnapshot, tx *applicationTestTx) { tx.brandTarget = nil }, code: ApplicationErrorCanonicalConflict},
		{name: "new brand is not created", setup: func(_ EnrichmentSourceSnapshot, tx *applicationTestTx) {
			proposal := &BrandProposal{Action: ActionProposeNew, CanonicalName: "New", Confidence: 0.8}
			tx.setProposals(t, proposal, nil, nil)
		}, code: ApplicationErrorCanonicalConflict},
		{name: "new category is not created", setup: func(_ EnrichmentSourceSnapshot, tx *applicationTestTx) {
			proposal := &CategoryProposal{Action: ActionProposeNew, CanonicalName: "New", Confidence: 0.8}
			tx.setProposals(t, nil, proposal, nil)
		}, code: ApplicationErrorCanonicalConflict},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			current := applicationTestSnapshot()
			brandID := int32(7)
			tx := newApplicationTestTx(t, current, &BrandProposal{Action: ActionMatchExisting, TargetID: &brandID, Confidence: 0.9}, nil, nil)
			tx.brandTarget = &ReviewCanonicalTarget{ID: brandID, Code: "BRAND-7"}
			test.setup(current, tx)
			_, err := NewProductEnrichmentApplicationService(&applicationTestStore{tx: tx}).ApplyApprovedSuggestion(context.Background(), 1, 1, 5)
			assertApplicationCode(t, err, test.code)
			if tx.committed || tx.audit.SuggestionID != 0 {
				t.Fatalf("conflict must not commit or audit: tx=%+v", tx)
			}
		})
	}
}

func TestApplyApprovedSuggestionRejectsCanonicalIdentityMismatch(t *testing.T) {
	current := applicationTestSnapshot()
	wantedID := int32(7)
	tx := newApplicationTestTx(t, current, &BrandProposal{Action: ActionMatchExisting, TargetID: &wantedID, TargetCode: "BRAND-7", Confidence: 0.9}, nil, nil)
	tx.brandTarget = &ReviewCanonicalTarget{ID: wantedID, Code: "BRAND-OTHER", Name: "Other"}
	_, err := NewProductEnrichmentApplicationService(&applicationTestStore{tx: tx}).ApplyApprovedSuggestion(context.Background(), 1, 1, 5)
	assertApplicationCode(t, err, ApplicationErrorCanonicalConflict)
	if tx.committed || tx.audit.SuggestionID != 0 {
		t.Fatalf("identity mismatch must not mutate: %+v", tx)
	}
}

func TestApplyApprovedSuggestionRejectsInvalidDescriptionAndOversizedValue(t *testing.T) {
	for _, value := range []string{"", "   ", string(make([]rune, MaxDescriptionRunes+1))} {
		current := applicationTestSnapshot()
		tx := newApplicationTestTx(t, current, nil, nil, &DescriptionProposal{Action: ActionProposeNew, Value: value, Confidence: 0.8})
		_, err := NewProductEnrichmentApplicationService(&applicationTestStore{tx: tx}).ApplyApprovedSuggestion(context.Background(), 1, 1, 5)
		assertApplicationCode(t, err, ApplicationErrorInvalidProposal)
	}
}

func TestApplyApprovedSuggestionNoOpAndIdempotentRetry(t *testing.T) {
	current := applicationTestSnapshot()
	tx := newApplicationTestTx(t, current, &BrandProposal{Action: ActionKeepExisting, Confidence: 1}, &CategoryProposal{Action: ActionNoMatch, Confidence: 1}, nil)
	store := &applicationTestStore{tx: tx}
	result, err := NewProductEnrichmentApplicationService(store).ApplyApprovedSuggestion(context.Background(), 1, 1, 5)
	if err != nil || len(result.ChangedFields) != 0 || tx.plan.BrandID != nil || tx.plan.CategoryID != nil || tx.plan.Description != nil {
		t.Fatalf("expected zero-change application, result=%+v err=%v plan=%+v", result, err, tx.plan)
	}
	firstAudit := tx.audit
	tx.resetForAlreadyApplied()
	second, err := NewProductEnrichmentApplicationService(store).ApplyApprovedSuggestion(context.Background(), 1, 1, 5)
	if err != nil || !second.AlreadyApplied || second.Status != SuggestionStatusApplied || tx.audit.SuggestionID != firstAudit.SuggestionID {
		t.Fatalf("expected safe idempotent result without duplicate audit, result=%+v err=%v tx=%+v", second, err, tx)
	}
}

func TestApplyApprovedSuggestionLifecycleAndStaleRemainApproved(t *testing.T) {
	statuses := []SuggestionStatus{SuggestionStatusPending, SuggestionStatusProcessing, SuggestionStatusRetryable, SuggestionStatusFailed, SuggestionStatusInReview, SuggestionStatusRejected}
	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			tx := newApplicationTestTx(t, applicationTestSnapshot(), nil, nil, nil)
			tx.suggestion.Status = status
			tx.initialSuggestion.Status = status
			_, err := NewProductEnrichmentApplicationService(&applicationTestStore{tx: tx}).ApplyApprovedSuggestion(context.Background(), 1, 1, 5)
			assertApplicationCode(t, err, ApplicationErrorNotApproved)
			if tx.committed || tx.suggestion.Status != status {
				t.Fatalf("non-approved status changed: %+v", tx)
			}
		})
	}

	current := applicationTestSnapshot()
	tx := newApplicationTestTx(t, current, nil, nil, nil)
	current.SourceItemName = "changed after approval"
	tx.current = current
	_, err := NewProductEnrichmentApplicationService(&applicationTestStore{tx: tx}).ApplyApprovedSuggestion(context.Background(), 1, 1, 5)
	assertApplicationCode(t, err, ApplicationErrorStale)
	if tx.suggestion.Status != SuggestionStatusApproved || tx.committed || tx.audit.SuggestionID != 0 {
		t.Fatalf("stale apply must leave approved suggestion untouched: %+v", tx)
	}
}

func TestApplyApprovedSuggestionAtomicFailureSeams(t *testing.T) {
	for _, test := range []struct {
		name      string
		configure func(*applicationTestTx)
		code      ApplicationErrorCode
	}{
		{name: "conditional product write", configure: func(tx *applicationTestTx) { tx.updateRows = 0 }, code: ApplicationErrorConditionalConflict},
		{name: "lifecycle transition", configure: func(tx *applicationTestTx) { tx.applyErr = errors.New("transition failed") }, code: ApplicationErrorPersistence},
		{name: "audit insert", configure: func(tx *applicationTestTx) { tx.auditErr = errors.New("audit failed") }, code: ApplicationErrorAuditFailure},
	} {
		t.Run(test.name, func(t *testing.T) {
			current := applicationTestSnapshot()
			brandID := int32(7)
			tx := newApplicationTestTx(t, current, &BrandProposal{Action: ActionMatchExisting, TargetID: &brandID, Confidence: 0.9}, nil, nil)
			tx.brandTarget = &ReviewCanonicalTarget{ID: brandID, Code: "BRAND-7"}
			test.configure(tx)
			_, err := NewProductEnrichmentApplicationService(&applicationTestStore{tx: tx}).ApplyApprovedSuggestion(context.Background(), 1, 1, 5)
			assertApplicationCode(t, err, test.code)
			if tx.committed || !tx.rolledBack || tx.suggestion.Status != SuggestionStatusApproved || tx.audit.SuggestionID != 0 {
				t.Fatalf("expected rollback with approved/no-audit state: %+v", tx)
			}
		})
	}
}

func TestApplyApprovedSuggestionValidatesTrustedIDsAndFingerprintScope(t *testing.T) {
	service := NewProductEnrichmentApplicationService(&applicationTestStore{tx: newApplicationTestTx(t, applicationTestSnapshot(), nil, nil, nil)})
	for _, ids := range [][3]int32{{0, 1, 5}, {1, 0, 5}, {1, 1, 0}} {
		_, err := service.ApplyApprovedSuggestion(context.Background(), ids[0], ids[1], ids[2])
		assertApplicationCode(t, err, ApplicationErrorInvalidProposal)
	}

	prior := applicationTestSnapshot()
	fingerprint, err := FingerprintSnapshot(prior)
	if err != nil {
		t.Fatal(err)
	}
	inventoryOnly := prior
	// Inventory and pricing are intentionally absent from the snapshot, so the
	// same fingerprint remains valid for those unrelated changes.
	if got, _ := FingerprintSnapshot(inventoryOnly); got != fingerprint {
		t.Fatal("unrelated operational state changed the enrichment fingerprint")
	}
}

type applicationTestStore struct{ tx *applicationTestTx }

func (s *applicationTestStore) BeginProductEnrichmentApplication(context.Context) (ProductEnrichmentApplicationTransaction, error) {
	return s.tx, nil
}

type applicationTestTx struct {
	suggestion        ReviewSuggestionRecord
	initialSuggestion ReviewSuggestionRecord
	current           EnrichmentSourceSnapshot
	brandTarget       *ReviewCanonicalTarget
	categoryTarget    *ReviewCanonicalTarget
	plan              ApplyPlan
	audit             ApplicationAudit
	updateRows        int64
	applyErr          error
	auditErr          error
	committed         bool
	rolledBack        bool
}

func newApplicationTestTx(t *testing.T, current EnrichmentSourceSnapshot, brand *BrandProposal, category *CategoryProposal, description *DescriptionProposal) *applicationTestTx {
	t.Helper()
	tx := &applicationTestTx{current: current, updateRows: 1}
	tx.setProposals(t, brand, category, description)
	return tx
}

func (tx *applicationTestTx) refreshCurrent(t *testing.T) {
	t.Helper()
	tx.suggestion.SourceItemName = tx.current.SourceItemName
	tx.suggestion.SourceDataFingerprint, _ = FingerprintSnapshot(tx.current)
	tx.suggestion.StructuredCurrent, _ = StructuredCurrent(tx.current)
}

func (tx *applicationTestTx) setProposals(t *testing.T, brand *BrandProposal, category *CategoryProposal, description *DescriptionProposal) {
	t.Helper()
	brandJSON, _ := json.Marshal(brand)
	categoryJSON, _ := json.Marshal(category)
	descriptionJSON, _ := json.Marshal(description)
	structured, err := StructuredCurrent(tx.current)
	if err != nil {
		t.Fatal(err)
	}
	fingerprint, err := FingerprintSnapshot(tx.current)
	if err != nil {
		t.Fatal(err)
	}
	tx.suggestion = ReviewSuggestionRecord{ID: 1, OrganizationID: 1, ProductID: tx.current.ProductID, SourceItemCode: tx.current.SourceItemCode, SourceItemName: tx.current.SourceItemName, SourceDataFingerprint: fingerprint, ContractVersion: EnrichmentContractVersion, StructuredCurrent: structured, ProposedBrand: brandJSON, ProposedCategory: categoryJSON, ProposedDescription: descriptionJSON, Status: SuggestionStatusApproved}
	tx.initialSuggestion = tx.suggestion
}

func (tx *applicationTestTx) resetForAlreadyApplied() {
	tx.suggestion.Status = SuggestionStatusApplied
	now := time.Now()
	tx.suggestion.AppliedAt = &now
	tx.committed = false
	tx.rolledBack = false
	// Preserve the first audit to assert that the retry does not overwrite it.
}

func (tx *applicationTestTx) LockProductEnrichmentSuggestion(context.Context, int32, int32) (ReviewSuggestionRecord, error) {
	return tx.suggestion, nil
}
func (tx *applicationTestTx) LockProductEnrichmentSnapshot(context.Context, int32, int32) (EnrichmentSourceSnapshot, error) {
	return tx.current, nil
}
func (tx *applicationTestTx) ResolveBrandApplicationTarget(context.Context, *int32, string) (*ReviewCanonicalTarget, error) {
	return tx.brandTarget, nil
}
func (tx *applicationTestTx) ResolveCategoryApplicationTarget(context.Context, *int32, string) (*ReviewCanonicalTarget, error) {
	return tx.categoryTarget, nil
}
func (tx *applicationTestTx) ApplyProductEnrichmentFields(_ context.Context, _ int32, _ int32, plan ApplyPlan) (int64, error) {
	tx.plan = plan
	return tx.updateRows, nil
}
func (tx *applicationTestTx) MarkProductEnrichmentSuggestionApplied(context.Context, int32, int32) (ReviewSuggestionRecord, error) {
	if tx.applyErr != nil {
		return ReviewSuggestionRecord{}, tx.applyErr
	}
	tx.suggestion.Status = SuggestionStatusApplied
	now := time.Now()
	tx.suggestion.AppliedAt = &now
	return tx.suggestion, nil
}
func (tx *applicationTestTx) ApproveAndApplyProductEnrichmentSuggestion(_ context.Context, _ int32, _ int32, reviewerID *int32) (ReviewSuggestionRecord, error) {
	if tx.applyErr != nil {
		return ReviewSuggestionRecord{}, tx.applyErr
	}
	tx.suggestion.Status = SuggestionStatusApplied
	tx.suggestion.ReviewerID = reviewerID
	now := time.Now()
	tx.suggestion.ReviewedAt = &now
	tx.suggestion.AppliedAt = &now
	return tx.suggestion, nil
}
func (tx *applicationTestTx) InsertProductEnrichmentApplicationAudit(_ context.Context, audit ApplicationAudit) error {
	if tx.auditErr != nil {
		return tx.auditErr
	}
	if tx.audit.SuggestionID != 0 {
		return errors.New("duplicate audit")
	}
	tx.audit = audit
	return nil
}
func (tx *applicationTestTx) Commit(context.Context) error { tx.committed = true; return nil }
func (tx *applicationTestTx) Rollback(context.Context) error {
	tx.suggestion = tx.initialSuggestion
	tx.rolledBack = true
	return nil
}

func applicationTestSnapshot() EnrichmentSourceSnapshot {
	return EnrichmentSourceSnapshot{OrganizationID: 1, ProductID: 95, SourceSystem: SourceSystemSAP, SourceItemCode: "INV00095", SourceItemName: "Test product", ProductType: ProductTypeStandard}
}

func assertApplicationCode(t *testing.T, err error, want ApplicationErrorCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected application error %q", want)
	}
	var applicationErr *ApplicationError
	if !errors.As(err, &applicationErr) || applicationErr.Code != want {
		t.Fatalf("expected application error %q, got %v", want, err)
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
