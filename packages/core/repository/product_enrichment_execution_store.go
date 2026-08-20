package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/NasTecSol/nembus-core/enrichment"
	"github.com/jackc/pgx/v5/pgtype"
)

// ProductEnrichmentStore adapts generated Stage 2C execution queries to the
// provider-neutral worker contract. SQL concurrency and organization scoping
// remain enforced by the generated query predicates.
func (s *ProductEnrichmentStore) ListDueEnrichmentSuggestions(ctx context.Context, limit int, now time.Time) ([]enrichment.EnrichmentExecutionSuggestion, error) {
	if s == nil || s.queries == nil {
		return nil, fmt.Errorf("product enrichment repository is not configured")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("enrichment batch limit must be positive")
	}
	rows, err := s.queries.ListDueProductEnrichmentSuggestions(ctx, ListDueProductEnrichmentSuggestionsParams{
		Now: pgtype.Timestamp{Time: now, Valid: true}, BatchLimit: int32(limit),
	})
	if err != nil {
		return nil, err
	}
	result := make([]enrichment.EnrichmentExecutionSuggestion, 0, len(rows))
	for _, row := range rows {
		result = append(result, enrichment.EnrichmentExecutionSuggestion{
			ID: row.ID, OrganizationID: row.OrganizationID, ProductID: row.ProductID,
			SourceItemCode: row.SourceItemCode, SourceDataFingerprint: row.SourceDataFingerprint,
			Status: enrichment.SuggestionStatus(row.Status), AttemptCount: int(row.AttemptCount),
		})
	}
	return result, nil
}

func (s *ProductEnrichmentStore) ClaimEnrichmentSuggestion(ctx context.Context, organizationID, id int32) (enrichment.EnrichmentExecutionSuggestion, error) {
	if s == nil || s.queries == nil {
		return enrichment.EnrichmentExecutionSuggestion{}, fmt.Errorf("product enrichment repository is not configured")
	}
	row, err := s.queries.ClaimProductEnrichmentSuggestion(ctx, ClaimProductEnrichmentSuggestionParams{OrganizationID: organizationID, ID: id})
	if err != nil {
		return enrichment.EnrichmentExecutionSuggestion{}, err
	}
	return enrichment.EnrichmentExecutionSuggestion{
		ID: row.ID, OrganizationID: row.OrganizationID, ProductID: row.ProductID,
		SourceItemCode: row.SourceItemCode, SourceDataFingerprint: row.SourceDataFingerprint,
		Status: enrichment.SuggestionStatus(row.Status), AttemptCount: int(row.AttemptCount),
	}, nil
}

func (s *ProductEnrichmentStore) CompleteEnrichmentSuggestion(ctx context.Context, input enrichment.EnrichmentCompletion) error {
	if s == nil || s.queries == nil {
		return fmt.Errorf("product enrichment repository is not configured")
	}
	_, err := s.queries.CompleteProductEnrichmentSuggestion(ctx, CompleteProductEnrichmentSuggestionParams{
		ProposedBrand: input.ProposedBrand, ProposedCategory: input.ProposedCategory,
		ProposedDescription: input.ProposedDescription, UnsupportedSemantics: input.UnsupportedSemantics,
		Provider:       pgtype.Text{String: input.Provider, Valid: input.Provider != ""},
		Model:          pgtype.Text{String: input.Model, Valid: input.Model != ""},
		ModelVersion:   pgtype.Text{String: input.ModelVersion, Valid: input.ModelVersion != ""},
		OrganizationID: input.OrganizationID, ID: input.ID,
	})
	return err
}

func (s *ProductEnrichmentStore) MarkEnrichmentRetryable(ctx context.Context, input enrichment.EnrichmentRetry) error {
	if s == nil || s.queries == nil {
		return fmt.Errorf("product enrichment repository is not configured")
	}
	_, err := s.queries.RetryProductEnrichmentSuggestion(ctx, RetryProductEnrichmentSuggestionParams{
		NextAttemptAt:  pgtype.Timestamp{Time: input.NextAttemptAt, Valid: true},
		LastErrorCode:  pgtype.Text{String: input.ErrorCode, Valid: input.ErrorCode != ""},
		OrganizationID: input.OrganizationID, ID: input.ID,
	})
	return err
}

func (s *ProductEnrichmentStore) MarkEnrichmentFailed(ctx context.Context, input enrichment.EnrichmentFailure) error {
	if s == nil || s.queries == nil {
		return fmt.Errorf("product enrichment repository is not configured")
	}
	_, err := s.queries.FailProductEnrichmentSuggestion(ctx, FailProductEnrichmentSuggestionParams{
		LastErrorCode:  pgtype.Text{String: input.ErrorCode, Valid: input.ErrorCode != ""},
		OrganizationID: input.OrganizationID, ID: input.ID,
	})
	return err
}

var _ enrichment.EnrichmentExecutionStore = (*ProductEnrichmentStore)(nil)
