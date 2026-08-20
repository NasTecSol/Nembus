package enrichment

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestEnrichmentWorkerCompletesValidatedProviderResult(t *testing.T) {
	snapshot := workerSnapshot()
	fingerprint, err := FingerprintSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	store := &workerStore{rows: []EnrichmentExecutionSuggestion{{ID: 7, OrganizationID: 2, ProductID: 3, SourceItemCode: "INV00006", SourceDataFingerprint: fingerprint, Status: SuggestionStatusPending}}, snapshot: snapshot}
	provider := &workerProvider{result: EnrichmentResult{SourceItemCode: "INV00006", Provider: "openai", Model: "gpt-test", Proposals: validWorkerProposals()}}
	worker := NewEnrichmentWorker(store, provider, EnrichmentExecutionConfig{Timeout: time.Second, MaxAttempts: 3, Now: func() time.Time { return time.Unix(100, 0) }}, nil)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if provider.calls != 1 || store.completed != 1 || store.rows[0].Status != SuggestionStatusInReview {
		t.Fatalf("expected one successful completion, calls=%d completed=%d status=%s", provider.calls, store.completed, store.rows[0].Status)
	}
	if len(store.completion.ProposedBrand) == 0 || store.completion.Provider != "openai" {
		t.Fatalf("validated proposals and trusted metadata were not persisted")
	}
}

func TestEnrichmentWorkerRetriesThenFailsAtMaxAttempts(t *testing.T) {
	snapshot := workerSnapshot()
	fingerprint, _ := FingerprintSnapshot(snapshot)
	store := &workerStore{rows: []EnrichmentExecutionSuggestion{{ID: 8, OrganizationID: 2, ProductID: 3, SourceItemCode: "INV00006", SourceDataFingerprint: fingerprint, Status: SuggestionStatusPending}}, snapshot: snapshot}
	provider := &workerProvider{err: &ProviderError{Class: ProviderErrorRetryable, Code: "http_429", Err: errors.New("rate limited")}}
	now := time.Unix(100, 0)
	worker := NewEnrichmentWorker(store, provider, EnrichmentExecutionConfig{Timeout: time.Second, MaxAttempts: 2, Now: func() time.Time { return now }}, nil)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.retry == nil || store.rows[0].Status != SuggestionStatusRetryable || store.retry.ErrorCode != "http_429" {
		t.Fatalf("expected durable retry, retry=%+v status=%s", store.retry, store.rows[0].Status)
	}
	store.rows[0].Status = SuggestionStatusRetryable
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.failed == nil || store.rows[0].Status != SuggestionStatusFailed {
		t.Fatalf("expected terminal failure at max attempts, failure=%+v status=%s", store.failed, store.rows[0].Status)
	}
}

func TestEnrichmentWorkerDoesNotCallProviderForStaleSource(t *testing.T) {
	snapshot := workerSnapshot()
	store := &workerStore{rows: []EnrichmentExecutionSuggestion{{ID: 9, OrganizationID: 2, ProductID: 3, SourceItemCode: "INV00006", SourceDataFingerprint: "old-fingerprint", Status: SuggestionStatusPending}}, snapshot: snapshot}
	provider := &workerProvider{}
	worker := NewEnrichmentWorker(store, provider, EnrichmentExecutionConfig{Timeout: time.Second, MaxAttempts: 3}, nil)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if provider.calls != 0 || store.failed == nil || store.failed.ErrorCode != "stale_source" {
		t.Fatalf("stale source must not reach provider, calls=%d failure=%+v", provider.calls, store.failed)
	}
}

type workerProvider struct {
	result EnrichmentResult
	err    error
	calls  int
}

func (p *workerProvider) Enrich(_ context.Context, _ EnrichmentRequest) (EnrichmentResult, error) {
	p.calls++
	return p.result, p.err
}

type workerStore struct {
	rows       []EnrichmentExecutionSuggestion
	snapshot   EnrichmentSourceSnapshot
	completed  int
	completion EnrichmentCompletion
	retry      *EnrichmentRetry
	failed     *EnrichmentFailure
}

func (s *workerStore) ListDueEnrichmentSuggestions(_ context.Context, _ int, _ time.Time) ([]EnrichmentExecutionSuggestion, error) {
	return append([]EnrichmentExecutionSuggestion(nil), s.rows...), nil
}

func (s *workerStore) ClaimEnrichmentSuggestion(_ context.Context, organizationID, id int32) (EnrichmentExecutionSuggestion, error) {
	for i := range s.rows {
		if s.rows[i].OrganizationID == organizationID && s.rows[i].ID == id {
			if s.rows[i].Status != SuggestionStatusPending && s.rows[i].Status != SuggestionStatusRetryable {
				return EnrichmentExecutionSuggestion{}, errors.New("not claimable")
			}
			s.rows[i].Status = SuggestionStatusProcessing
			s.rows[i].AttemptCount++
			return s.rows[i], nil
		}
	}
	return EnrichmentExecutionSuggestion{}, errors.New("missing suggestion")
}

func (s *workerStore) CompleteEnrichmentSuggestion(_ context.Context, input EnrichmentCompletion) error {
	s.completed++
	s.completion = input
	s.rows[0].Status = SuggestionStatusInReview
	return nil
}

func (s *workerStore) MarkEnrichmentRetryable(_ context.Context, input EnrichmentRetry) error {
	s.retry = &input
	s.rows[0].Status = SuggestionStatusRetryable
	return nil
}

func (s *workerStore) MarkEnrichmentFailed(_ context.Context, input EnrichmentFailure) error {
	s.failed = &input
	s.rows[0].Status = SuggestionStatusFailed
	return nil
}

func (s *workerStore) LoadSAPProductEnrichmentSnapshot(_ context.Context, _ int32, _ string) (EnrichmentSourceSnapshot, error) {
	return s.snapshot, nil
}

func (s *workerStore) ListBrandCandidates(context.Context, int) ([]BrandCandidate, error) {
	return []BrandCandidate{{ID: 10, Code: "PANTENE", Name: "Pantene"}}, nil
}

func (s *workerStore) ListCategoryCandidates(context.Context, int) ([]CategoryCandidate, error) {
	return []CategoryCandidate{{ID: 20, Code: "HAIR", Name: "Hair Care"}}, nil
}

func workerSnapshot() EnrichmentSourceSnapshot {
	return EnrichmentSourceSnapshot{OrganizationID: 2, ProductID: 3, SourceSystem: SourceSystemSAP, SourceItemCode: "INV00006", SourceItemName: "شامبو بانتين صحي ونظيف 24*400 مل", ProductType: ProductTypeStandard}
}

func validWorkerProposals() ProposalSet {
	return ProposalSet{
		Brand:                &BrandProposal{Action: ActionMatchExisting, TargetID: int32Pointer(10), TargetCode: "PANTENE", Confidence: 0.9},
		Category:             &CategoryProposal{Action: ActionNoMatch, Confidence: 0.2},
		Description:          &DescriptionProposal{Action: ActionProposeNew, Value: "شامبو بانتين", Confidence: 0.8},
		UnsupportedSemantics: []UnsupportedSemantic{{SemanticType: "packaging", Key: "size_text", Value: json.RawMessage(`"24*400 مل"`), Confidence: 0.8}},
	}
}
