package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/NasTecSol/nembus-core/enrichment"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProductEnrichmentApplicationStore is request-local and must be built from
// the tenant repository selected by the authenticated request. It has no
// master-repository fallback.
type ProductEnrichmentApplicationStore struct {
	queries *Queries
	pool    *pgxpool.Pool
}

func NewProductEnrichmentApplicationStore(queries *Queries, pool *pgxpool.Pool) *ProductEnrichmentApplicationStore {
	return &ProductEnrichmentApplicationStore{queries: queries, pool: pool}
}

func (s *ProductEnrichmentApplicationStore) BeginProductEnrichmentApplication(ctx context.Context) (enrichment.ProductEnrichmentApplicationTransaction, error) {
	if s == nil || s.queries == nil || s.pool == nil {
		return nil, fmt.Errorf("tenant application repository is not configured")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &productEnrichmentApplicationTransaction{queries: s.queries.WithTx(tx), tx: tx}, nil
}

type productEnrichmentApplicationTransaction struct {
	queries *Queries
	tx      pgx.Tx
}

func (t *productEnrichmentApplicationTransaction) LockProductEnrichmentSuggestion(ctx context.Context, organizationID, suggestionID int32) (enrichment.ReviewSuggestionRecord, error) {
	row, err := t.queries.LockProductEnrichmentSuggestionForApplication(ctx, LockProductEnrichmentSuggestionForApplicationParams{OrganizationID: organizationID, ID: suggestionID})
	if err != nil {
		if err == pgx.ErrNoRows {
			return enrichment.ReviewSuggestionRecord{}, enrichment.ErrApplicationSuggestionNotFound
		}
		return enrichment.ReviewSuggestionRecord{}, err
	}
	return reviewSuggestionRecord(row), nil
}

func (t *productEnrichmentApplicationTransaction) LockProductEnrichmentSnapshot(ctx context.Context, organizationID, productID int32) (enrichment.EnrichmentSourceSnapshot, error) {
	product, err := t.queries.LockProductForEnrichmentApplication(ctx, LockProductForEnrichmentApplicationParams{OrganizationID: organizationID, ID: productID})
	if err != nil {
		if err == pgx.ErrNoRows {
			return enrichment.EnrichmentSourceSnapshot{}, enrichment.ErrApplicationProductNotFound
		}
		return enrichment.EnrichmentSourceSnapshot{}, err
	}
	return (&ProductEnrichmentStore{queries: t.queries}).loadSnapshotFromProduct(ctx, product)
}

func (t *productEnrichmentApplicationTransaction) ResolveBrandApplicationTarget(ctx context.Context, targetID *int32, targetCode string) (*enrichment.ReviewCanonicalTarget, error) {
	if targetID != nil && *targetID <= 0 {
		return nil, nil
	}
	code := strings.TrimSpace(targetCode)
	if targetID == nil && code == "" {
		return nil, nil
	}
	var brand Brand
	var err error
	if targetID != nil {
		brand, err = t.queries.LockBrandForEnrichmentApplicationByID(ctx, *targetID)
	} else {
		brand, err = t.queries.LockBrandForEnrichmentApplicationByCode(ctx, code)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if !brand.IsActive.Valid || !brand.IsActive.Bool || (targetID != nil && brand.ID != *targetID) || (code != "" && brand.Code != code) {
		return nil, nil
	}
	return &enrichment.ReviewCanonicalTarget{ID: brand.ID, Code: brand.Code, Name: brand.Name}, nil
}

func (t *productEnrichmentApplicationTransaction) ResolveCategoryApplicationTarget(ctx context.Context, targetID *int32, targetCode string) (*enrichment.ReviewCanonicalTarget, error) {
	if targetID != nil && *targetID <= 0 {
		return nil, nil
	}
	code := strings.TrimSpace(targetCode)
	if targetID == nil && code == "" {
		return nil, nil
	}
	var category ProductCategory
	var err error
	if targetID != nil {
		category, err = t.queries.LockCategoryForEnrichmentApplicationByID(ctx, *targetID)
	} else {
		category, err = t.queries.LockCategoryForEnrichmentApplicationByCode(ctx, code)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if !category.IsActive.Valid || !category.IsActive.Bool || (targetID != nil && category.ID != *targetID) || (code != "" && category.Code != code) {
		return nil, nil
	}
	return &enrichment.ReviewCanonicalTarget{ID: category.ID, Code: category.Code, Name: category.Name}, nil
}

func (t *productEnrichmentApplicationTransaction) ApplyProductEnrichmentFields(ctx context.Context, organizationID, productID int32, plan enrichment.ApplyPlan) (int64, error) {
	return t.queries.ApplyProductEnrichmentFields(ctx, ApplyProductEnrichmentFieldsParams{
		OrganizationID: organizationID,
		ID:             productID,
		BrandID:        nullableInt4(plan.BrandID),
		CategoryID:     nullableInt4(plan.CategoryID),
		Description:    nullableText(plan.Description),
	})
}

func (t *productEnrichmentApplicationTransaction) MarkProductEnrichmentSuggestionApplied(ctx context.Context, organizationID, suggestionID int32) (enrichment.ReviewSuggestionRecord, error) {
	row, err := t.queries.MarkProductEnrichmentSuggestionApplied(ctx, MarkProductEnrichmentSuggestionAppliedParams{OrganizationID: organizationID, ID: suggestionID})
	if err != nil {
		if err == pgx.ErrNoRows {
			return enrichment.ReviewSuggestionRecord{}, enrichment.ErrApplicationTransitionConflict
		}
		return enrichment.ReviewSuggestionRecord{}, err
	}
	return reviewSuggestionRecord(row), nil
}

func (t *productEnrichmentApplicationTransaction) ApproveAndApplyProductEnrichmentSuggestion(ctx context.Context, organizationID, suggestionID int32, reviewerID *int32) (enrichment.ReviewSuggestionRecord, error) {
	row, err := t.queries.ApproveAndApplyProductEnrichmentSuggestion(ctx, ApproveAndApplyProductEnrichmentSuggestionParams{
		OrganizationID: organizationID, ID: suggestionID, ReviewerID: nullableInt4(reviewerID),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return enrichment.ReviewSuggestionRecord{}, enrichment.ErrApplicationTransitionConflict
		}
		return enrichment.ReviewSuggestionRecord{}, err
	}
	return reviewSuggestionRecord(row), nil
}

func (t *productEnrichmentApplicationTransaction) InsertProductEnrichmentApplicationAudit(ctx context.Context, audit enrichment.ApplicationAudit) error {
	oldValues := map[string]any{"status": audit.OldStatus}
	newValues := map[string]any{
		"event":               "product_enrichment.applied",
		"status":              audit.NewStatus,
		"organization_id":     audit.OrganizationID,
		"suggestion_id":       audit.SuggestionID,
		"product_id":          audit.ProductID,
		"applier_user_id":     audit.ApplierUserID,
		"description_changed": audit.DescriptionChanged,
	}
	if audit.OldBrandID != nil || audit.NewBrandID != nil {
		oldValues["brand_id"] = audit.OldBrandID
		newValues["brand_id"] = audit.NewBrandID
	}
	if audit.OldCategoryID != nil || audit.NewCategoryID != nil {
		oldValues["category_id"] = audit.OldCategoryID
		newValues["category_id"] = audit.NewCategoryID
	}
	oldJSON, err := json.Marshal(oldValues)
	if err != nil {
		return err
	}
	newJSON, err := json.Marshal(newValues)
	if err != nil {
		return err
	}
	return t.queries.InsertProductEnrichmentReviewAudit(ctx, InsertProductEnrichmentReviewAuditParams{
		OrganizationID: pgtype.Int4{Int32: audit.OrganizationID, Valid: true},
		TableName:      "product_enrichment_suggestions",
		RecordID:       strconv.FormatInt(int64(audit.SuggestionID), 10),
		Action:         "UPDATE",
		OldValues:      oldJSON,
		NewValues:      newJSON,
		ChangedFields:  append([]string(nil), audit.ChangedFields...),
		PerformedBy:    nullableInt4(reviewIDPtr(audit.ApplierUserID)),
	})
}

func (t *productEnrichmentApplicationTransaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *productEnrichmentApplicationTransaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

func nullableInt4(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *value, Valid: true}
}

func nullableText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

var _ enrichment.ProductEnrichmentApplicationStore = (*ProductEnrichmentApplicationStore)(nil)
var _ enrichment.ProductEnrichmentApplicationTransaction = (*productEnrichmentApplicationTransaction)(nil)
