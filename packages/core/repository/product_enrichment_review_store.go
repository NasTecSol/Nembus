package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/NasTecSol/nembus-core/enrichment"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProductEnrichmentReviewStore is request-local and always wraps the
// repository selected by TenantMiddleware. It has no master-repository path.
type ProductEnrichmentReviewStore struct {
	queries *Queries
	pool    *pgxpool.Pool
}

func NewProductEnrichmentReviewStore(queries *Queries, pool *pgxpool.Pool) *ProductEnrichmentReviewStore {
	return &ProductEnrichmentReviewStore{queries: queries, pool: pool}
}

func (s *ProductEnrichmentReviewStore) ListReviewSuggestions(ctx context.Context, organizationID int32, status enrichment.ReviewListStatus, limit, offset int32) ([]enrichment.ReviewSuggestionRecord, error) {
	if s == nil || s.queries == nil {
		return nil, fmt.Errorf("product enrichment review repository is not configured")
	}
	rows, err := s.queries.ListProductEnrichmentSuggestions(ctx, ListProductEnrichmentSuggestionsParams{
		OrganizationID: organizationID,
		Limit:          limit,
		Offset:         offset,
		Status:         pgtype.Text{String: string(status), Valid: true},
	})
	if err != nil {
		return nil, err
	}
	result := make([]enrichment.ReviewSuggestionRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, reviewSuggestionRecord(row))
	}
	return result, nil
}

func (s *ProductEnrichmentReviewStore) GetReviewSuggestion(ctx context.Context, organizationID, suggestionID int32) (enrichment.ReviewSuggestionRecord, error) {
	if s == nil || s.queries == nil {
		return enrichment.ReviewSuggestionRecord{}, fmt.Errorf("product enrichment review repository is not configured")
	}
	row, err := s.queries.GetProductEnrichmentSuggestionByID(ctx, GetProductEnrichmentSuggestionByIDParams{
		OrganizationID: organizationID,
		ID:             suggestionID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return enrichment.ReviewSuggestionRecord{}, fmt.Errorf("%w: %v", enrichment.ErrReviewSuggestionNotFound, err)
		}
		return enrichment.ReviewSuggestionRecord{}, err
	}
	return reviewSuggestionRecord(row), nil
}

func (s *ProductEnrichmentReviewStore) LoadSAPProductEnrichmentSnapshotByID(ctx context.Context, organizationID, productID int32) (enrichment.EnrichmentSourceSnapshot, error) {
	return NewProductEnrichmentStore(s.queries).LoadProductEnrichmentSnapshotByID(ctx, organizationID, productID)
}

func (s *ProductEnrichmentReviewStore) ResolveBrandReviewTarget(ctx context.Context, targetID *int32, targetCode string) (*enrichment.ReviewCanonicalTarget, error) {
	if s == nil || s.queries == nil {
		return nil, fmt.Errorf("product enrichment review repository is not configured")
	}
	var brand Brand
	var err error
	if targetID != nil && *targetID > 0 {
		brand, err = s.queries.GetBrandByID(ctx, *targetID)
	} else if targetCode != "" {
		brand, err = s.queries.GetBrandByCode(ctx, targetCode)
	} else {
		return nil, nil
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if !brand.IsActive.Valid || !brand.IsActive.Bool {
		return nil, nil
	}
	if targetID != nil && *targetID > 0 && brand.ID != *targetID {
		return nil, nil
	}
	if targetCode != "" && brand.Code != targetCode {
		return nil, nil
	}
	return &enrichment.ReviewCanonicalTarget{ID: brand.ID, Code: brand.Code, Name: brand.Name}, nil
}

func (s *ProductEnrichmentReviewStore) ResolveCategoryReviewTarget(ctx context.Context, targetID *int32, targetCode string) (*enrichment.ReviewCanonicalTarget, error) {
	if s == nil || s.queries == nil {
		return nil, fmt.Errorf("product enrichment review repository is not configured")
	}
	var category ProductCategory
	var err error
	if targetID != nil && *targetID > 0 {
		category, err = s.queries.GetProductCategory(ctx, *targetID)
	} else if targetCode != "" {
		category, err = s.queries.GetProductCategoryByCode(ctx, targetCode)
	} else {
		return nil, nil
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if !category.IsActive.Valid || !category.IsActive.Bool {
		return nil, nil
	}
	if targetID != nil && *targetID > 0 && category.ID != *targetID {
		return nil, nil
	}
	if targetCode != "" && category.Code != targetCode {
		return nil, nil
	}
	return &enrichment.ReviewCanonicalTarget{ID: category.ID, Code: category.Code, Name: category.Name}, nil
}

func (s *ProductEnrichmentReviewStore) BeginReviewTransaction(ctx context.Context) (enrichment.ReviewTransaction, error) {
	if s == nil || s.pool == nil {
		return nil, fmt.Errorf("tenant transaction pool is not configured")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &productEnrichmentReviewTransaction{queries: s.queries.WithTx(tx), tx: tx}, nil
}

// BeginProductEnrichmentApplication allows the review service to hand the
// same tenant-local repository to the existing application transaction.
func (s *ProductEnrichmentReviewStore) BeginProductEnrichmentApplication(ctx context.Context) (enrichment.ProductEnrichmentApplicationTransaction, error) {
	if s == nil || s.queries == nil || s.pool == nil {
		return nil, fmt.Errorf("tenant application repository is not configured")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &productEnrichmentApplicationTransaction{queries: s.queries.WithTx(tx), tx: tx}, nil
}

type productEnrichmentReviewTransaction struct {
	queries *Queries
	tx      pgx.Tx
}

func (t *productEnrichmentReviewTransaction) Approve(ctx context.Context, organizationID, suggestionID, reviewerID int32) (enrichment.ReviewSuggestionRecord, error) {
	row, err := t.queries.ApproveProductEnrichmentSuggestion(ctx, ApproveProductEnrichmentSuggestionParams{
		OrganizationID: organizationID, ID: suggestionID, ReviewerID: nullableInt4(reviewIDPtr(reviewerID)),
	})
	if err != nil {
		return enrichment.ReviewSuggestionRecord{}, err
	}
	return reviewSuggestionRecord(row), nil
}

func (t *productEnrichmentReviewTransaction) Reject(ctx context.Context, organizationID, suggestionID, reviewerID int32) (enrichment.ReviewSuggestionRecord, error) {
	row, err := t.queries.RejectProductEnrichmentSuggestion(ctx, RejectProductEnrichmentSuggestionParams{
		OrganizationID: organizationID, ID: suggestionID, ReviewerID: nullableInt4(reviewIDPtr(reviewerID)),
	})
	if err != nil {
		return enrichment.ReviewSuggestionRecord{}, err
	}
	return reviewSuggestionRecord(row), nil
}

func (t *productEnrichmentReviewTransaction) InsertAudit(ctx context.Context, audit enrichment.ReviewAudit) error {
	oldValues, err := json.Marshal(map[string]any{"status": audit.OldStatus})
	if err != nil {
		return err
	}
	newValues, err := json.Marshal(map[string]any{
		"status": audit.NewStatus, "event": audit.Event, "suggestion_id": audit.SuggestionID,
		"product_id": audit.ProductID, "organization_id": audit.OrganizationID, "reviewer_id": audit.ReviewerID,
	})
	if err != nil {
		return err
	}
	return t.queries.InsertProductEnrichmentReviewAudit(ctx, InsertProductEnrichmentReviewAuditParams{
		OrganizationID: pgtype.Int4{Int32: audit.OrganizationID, Valid: true},
		TableName:      "product_enrichment_suggestions",
		RecordID:       strconv.FormatInt(int64(audit.SuggestionID), 10),
		Action:         "UPDATE",
		OldValues:      oldValues,
		NewValues:      newValues,
		ChangedFields:  []string{"status", "reviewer_id", "reviewed_at"},
		PerformedBy:    nullableInt4(reviewIDPtr(audit.ReviewerID)),
	})
}

func (t *productEnrichmentReviewTransaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}
func (t *productEnrichmentReviewTransaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

func reviewSuggestionRecord(row ProductEnrichmentSuggestion) enrichment.ReviewSuggestionRecord {
	return enrichment.ReviewSuggestionRecord{
		ID: row.ID, OrganizationID: row.OrganizationID, ProductID: row.ProductID,
		SourceItemCode: row.SourceItemCode, SourceItemName: row.SourceItemName,
		SourceDataFingerprint: row.SourceDataFingerprint, ContractVersion: row.ContractVersion,
		StructuredCurrent: row.StructuredCurrent, ProposedBrand: row.ProposedBrand,
		ProposedCategory: row.ProposedCategory, ProposedDescription: row.ProposedDescription,
		UnsupportedSemantics: row.UnsupportedSemantics, Source: row.Source,
		Provider: pgText(row.Provider), Model: pgText(row.Model), ModelVersion: pgText(row.ModelVersion),
		Status: enrichment.SuggestionStatus(row.Status), ReviewerID: pgInt4Ptr(row.ReviewerID),
		ReviewedAt: pgTimestampPtr(row.ReviewedAt), AppliedAt: pgTimestampPtr(row.AppliedAt),
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func pgInt4Ptr(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}
	result := value.Int32
	return &result
}

func pgTimestampPtr(value pgtype.Timestamp) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func reviewIDPtr(value int32) *int32 {
	if value <= 0 {
		return nil
	}
	return &value
}

var _ enrichment.ReviewApplicationStore = (*ProductEnrichmentReviewStore)(nil)
