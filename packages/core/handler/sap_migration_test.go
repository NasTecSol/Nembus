package handler

import (
	"context"
	"testing"

	"github.com/NasTecSol/nembus-core/enrichment"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
)

type sapMigrationEnqueuerFake struct {
	calls int
}

func (f *sapMigrationEnqueuerFake) EnqueueSAPProduct(context.Context, int32, string) (enrichment.EnqueueResult, error) {
	f.calls++
	return enrichment.EnqueueResult{}, nil
}

func TestSAPMigrationHandlerGatesStage2AAtApplicationBoundary(t *testing.T) {
	for _, test := range []struct {
		name              string
		enrichmentEnabled bool
		wantFactoryCalls  int
	}{
		{name: "disabled eligible product", enrichmentEnabled: false, wantFactoryCalls: 0},
		{name: "disabled complete product", enrichmentEnabled: false, wantFactoryCalls: 0},
		{name: "enabled eligible product", enrichmentEnabled: true, wantFactoryCalls: 1},
		{name: "enabled complete product", enrichmentEnabled: true, wantFactoryCalls: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			h := NewSAPMigrationHandler(test.enrichmentEnabled)
			factoryCalls := 0
			h.newEnrichmentCoordinator = func(*repository.Queries) enrichment.EnrichmentEnqueuer {
				factoryCalls++
				return &sapMigrationEnqueuerFake{}
			}

			uc := usecase.NewSAPMigrationUseCase(nil)
			h.configureEnrichment(uc, nil)

			if factoryCalls != test.wantFactoryCalls {
				t.Fatalf("coordinator factory calls=%d, want %d", factoryCalls, test.wantFactoryCalls)
			}
		})
	}
}

func TestSAPMigrationHandlerDisabledGateDoesNotInspectEligibility(t *testing.T) {
	h := NewSAPMigrationHandler(false)
	factoryCalled := false
	h.newEnrichmentCoordinator = func(*repository.Queries) enrichment.EnrichmentEnqueuer {
		factoryCalled = true
		return &sapMigrationEnqueuerFake{}
	}

	// Both complete and otherwise eligible products are bypassed before the
	// provider-independent coordinator can load a snapshot or evaluate gaps.
	for _, product := range []enrichment.EnrichmentEligibilityInput{
		{
			SourceSystem: enrichment.SourceSystemSAP, OrganizationID: 1, ProductID: 2,
			SourceItemCode: "SAP-INCOMPLETE", SourceItemName: "Eligible item",
			ProductType: enrichment.ProductTypeStandard,
		},
		{
			SourceSystem: enrichment.SourceSystemSAP, OrganizationID: 1, ProductID: 3,
			SourceItemCode: "SAP-COMPLETE", SourceItemName: "Complete item",
			ProductType: enrichment.ProductTypeFinishedGood,
			Brand:       &enrichment.BrandIdentity{ID: 10}, Category: &enrichment.CategoryIdentity{ID: 20}, Description: "Complete",
		},
	} {
		decision := enrichment.EvaluateEligibility(product)
		if product.SourceItemCode == "SAP-INCOMPLETE" && !decision.Eligible {
			t.Fatalf("test product should be eligible: %+v", decision)
		}
		if product.SourceItemCode == "SAP-COMPLETE" && decision.Eligible {
			t.Fatalf("test product should be complete: %+v", decision)
		}
		h.configureEnrichment(usecase.NewSAPMigrationUseCase(nil), nil)
	}

	if factoryCalled {
		t.Fatal("disabled SAP migration configured Stage 2A for a product")
	}
}
