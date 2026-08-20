package enrichment

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTenantEnrichmentSupervisorIsolatesCollidingTenantIDs(t *testing.T) {
	provider := &routingProvider{}
	stores := map[string]*supervisorStore{
		"tenant-a": newSupervisorStore("A-BRAND", "A-CATEGORY"),
		"tenant-b": newSupervisorStore("B-BRAND", "B-CATEGORY"),
	}
	registry := &supervisorRegistry{tenants: []TenantRegistration{
		{Slug: "tenant-a", Active: true},
		{Slug: "tenant-b", Active: true},
	}}
	masterEnrichmentCalls := 0
	factory := func(_ context.Context, tenant TenantRegistration) (*EnrichmentWorker, error) {
		store := stores[tenant.Slug]
		if store == nil {
			masterEnrichmentCalls++
			return nil, errors.New("unknown tenant")
		}
		return NewEnrichmentWorker(store, provider, EnrichmentExecutionConfig{Timeout: time.Second, MaxAttempts: 2}, nil), nil
	}
	supervisor := NewTenantEnrichmentSupervisor(registry, factory, EnrichmentExecutionConfig{Interval: time.Second}, nil)

	if err := supervisor.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	for slug, store := range stores {
		if store.completed != 1 || store.rows[0].Status != SuggestionStatusInReview {
			t.Fatalf("tenant %s was not completed locally: completed=%d status=%s", slug, store.completed, store.rows[0].Status)
		}
		if store.completion.OrganizationID != 1 || store.completion.ID != 1 {
			t.Fatalf("tenant %s completion used non-local IDs: %+v", slug, store.completion)
		}
	}
	if masterEnrichmentCalls != 0 {
		t.Fatalf("master enrichment fallback was called %d times", masterEnrichmentCalls)
	}
	if len(provider.requests) != 2 || provider.requests[0].BrandCandidates[0].Code != "A-BRAND" || provider.requests[1].BrandCandidates[0].Code != "B-BRAND" {
		t.Fatalf("tenant candidate dictionaries crossed boundaries: %+v", provider.requests)
	}
	if provider.requests[0].CategoryCandidates[0].Code != "A-CATEGORY" || provider.requests[1].CategoryCandidates[0].Code != "B-CATEGORY" {
		t.Fatalf("tenant category dictionaries crossed boundaries: %+v", provider.requests)
	}
}

func TestTenantEnrichmentSupervisorSkipsDisabledAndDiscoversNewTenants(t *testing.T) {
	provider := &routingProvider{}
	stores := map[string]*supervisorStore{
		"tenant-a": newSupervisorStore("A-BRAND", "A-CATEGORY"),
		"tenant-c": newSupervisorStore("C-BRAND", "C-CATEGORY"),
	}
	registry := &supervisorRegistry{cycles: [][]TenantRegistration{
		{{Slug: "tenant-a", Active: true}, {Slug: "tenant-b", Active: false}},
		{{Slug: "tenant-a", Active: true}, {Slug: "tenant-c", Active: true}},
		{},
	}}
	var built []string
	factory := func(_ context.Context, tenant TenantRegistration) (*EnrichmentWorker, error) {
		built = append(built, tenant.Slug)
		return NewEnrichmentWorker(stores[tenant.Slug], provider, EnrichmentExecutionConfig{Timeout: time.Second}, nil), nil
	}
	supervisor := NewTenantEnrichmentSupervisor(registry, factory, EnrichmentExecutionConfig{Interval: time.Second}, nil)

	if err := supervisor.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := supervisor.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := supervisor.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(built) != 3 || built[0] != "tenant-a" || built[1] != "tenant-a" || built[2] != "tenant-c" {
		t.Fatalf("unexpected tenant processing sequence: %v", built)
	}
}

func TestTenantEnrichmentSupervisorContinuesAfterTenantSetupFailure(t *testing.T) {
	provider := &routingProvider{}
	bStore := newSupervisorStore("B-BRAND", "B-CATEGORY")
	registry := &supervisorRegistry{tenants: []TenantRegistration{{Slug: "tenant-a", Active: true}, {Slug: "tenant-b", Active: true}}}
	factory := func(_ context.Context, tenant TenantRegistration) (*EnrichmentWorker, error) {
		if tenant.Slug == "tenant-a" {
			return nil, errors.New("tenant database unavailable")
		}
		return NewEnrichmentWorker(bStore, provider, EnrichmentExecutionConfig{Timeout: time.Second}, nil), nil
	}
	supervisor := NewTenantEnrichmentSupervisor(registry, factory, EnrichmentExecutionConfig{Interval: time.Second}, nil)

	if err := supervisor.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if bStore.completed != 1 || bStore.rows[0].Status != SuggestionStatusInReview {
		t.Fatalf("tenant B did not run after tenant A setup failure: completed=%d status=%s", bStore.completed, bStore.rows[0].Status)
	}
}

func TestTenantEnrichmentSupervisorContinuesAfterTenantWorkerFailure(t *testing.T) {
	provider := &routingProvider{}
	aStore := newSupervisorStore("A-BRAND", "A-CATEGORY")
	aStore.listErr = errors.New("product_enrichment_suggestions table missing")
	bStore := newSupervisorStore("B-BRAND", "B-CATEGORY")
	registry := &supervisorRegistry{tenants: []TenantRegistration{{Slug: "tenant-a", Active: true}, {Slug: "tenant-b", Active: true}}}
	factory := func(_ context.Context, tenant TenantRegistration) (*EnrichmentWorker, error) {
		if tenant.Slug == "tenant-a" {
			return NewEnrichmentWorker(aStore, provider, EnrichmentExecutionConfig{Timeout: time.Second}, nil), nil
		}
		return NewEnrichmentWorker(bStore, provider, EnrichmentExecutionConfig{Timeout: time.Second}, nil), nil
	}
	supervisor := NewTenantEnrichmentSupervisor(registry, factory, EnrichmentExecutionConfig{Interval: time.Second}, nil)

	if err := supervisor.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if bStore.completed != 1 || bStore.rows[0].Status != SuggestionStatusInReview {
		t.Fatalf("tenant B did not run after tenant A worker failure: completed=%d status=%s", bStore.completed, bStore.rows[0].Status)
	}
}

type supervisorRegistry struct {
	tenants []TenantRegistration
	cycles  [][]TenantRegistration
	call    int
}

func (r *supervisorRegistry) ListActiveTenants(context.Context) ([]TenantRegistration, error) {
	if r.cycles == nil {
		return append([]TenantRegistration(nil), r.tenants...), nil
	}
	if r.call >= len(r.cycles) {
		return nil, nil
	}
	tenants := append([]TenantRegistration(nil), r.cycles[r.call]...)
	r.call++
	return tenants, nil
}

type routingProvider struct {
	requests []EnrichmentRequest
}

func (p *routingProvider) Enrich(_ context.Context, request EnrichmentRequest) (EnrichmentResult, error) {
	p.requests = append(p.requests, request)
	brand := &BrandProposal{Action: ActionNoMatch, Confidence: 0.2}
	if len(request.BrandCandidates) > 0 {
		candidate := request.BrandCandidates[0]
		brand = &BrandProposal{Action: ActionMatchExisting, TargetID: &candidate.ID, TargetCode: candidate.Code, Confidence: 0.9}
	}
	return EnrichmentResult{
		SourceItemCode: request.SourceItemCode,
		Provider:       "openai",
		Model:          "test-model",
		Proposals: ProposalSet{
			Brand:       brand,
			Category:    &CategoryProposal{Action: ActionNoMatch, Confidence: 0.2},
			Description: &DescriptionProposal{Action: ActionProposeNew, Value: "test description", Confidence: 0.8},
		},
	}, nil
}

type supervisorStore struct {
	rows       []EnrichmentExecutionSuggestion
	snapshot   EnrichmentSourceSnapshot
	brands     []BrandCandidate
	categories []CategoryCandidate
	listErr    error
	completed  int
	completion EnrichmentCompletion
}

func newSupervisorStore(brandCode, categoryCode string) *supervisorStore {
	snapshot := EnrichmentSourceSnapshot{
		OrganizationID: 1, ProductID: 95, SourceSystem: SourceSystemSAP,
		SourceItemCode: "INV00095", SourceItemName: "same local item", ProductType: ProductTypeStandard,
	}
	fingerprint, _ := FingerprintSnapshot(snapshot)
	return &supervisorStore{
		rows: []EnrichmentExecutionSuggestion{{ID: 1, OrganizationID: 1, ProductID: 95, SourceItemCode: snapshot.SourceItemCode, SourceDataFingerprint: fingerprint, Status: SuggestionStatusPending}},
		snapshot: snapshot,
		brands: []BrandCandidate{{ID: 10, Code: brandCode, Name: brandCode}},
		categories: []CategoryCandidate{{ID: 20, Code: categoryCode, Name: categoryCode}},
	}
}

func (s *supervisorStore) ListDueEnrichmentSuggestions(context.Context, int, time.Time) ([]EnrichmentExecutionSuggestion, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return append([]EnrichmentExecutionSuggestion(nil), s.rows...), nil
}

func (s *supervisorStore) ClaimEnrichmentSuggestion(_ context.Context, organizationID, id int32) (EnrichmentExecutionSuggestion, error) {
	if len(s.rows) != 1 || s.rows[0].OrganizationID != organizationID || s.rows[0].ID != id {
		return EnrichmentExecutionSuggestion{}, errors.New("tenant-local suggestion not found")
	}
	s.rows[0].Status = SuggestionStatusProcessing
	s.rows[0].AttemptCount++
	return s.rows[0], nil
}

func (s *supervisorStore) CompleteEnrichmentSuggestion(_ context.Context, input EnrichmentCompletion) error {
	s.completed++
	s.completion = input
	s.rows[0].Status = SuggestionStatusInReview
	return nil
}

func (*supervisorStore) MarkEnrichmentRetryable(context.Context, EnrichmentRetry) error { return nil }
func (*supervisorStore) MarkEnrichmentFailed(context.Context, EnrichmentFailure) error   { return nil }

func (s *supervisorStore) LoadSAPProductEnrichmentSnapshot(context.Context, int32, string) (EnrichmentSourceSnapshot, error) {
	return s.snapshot, nil
}

func (s *supervisorStore) ListBrandCandidates(context.Context, int) ([]BrandCandidate, error) {
	return append([]BrandCandidate(nil), s.brands...), nil
}

func (s *supervisorStore) ListCategoryCandidates(context.Context, int) ([]CategoryCandidate, error) {
	return append([]CategoryCandidate(nil), s.categories...), nil
}
