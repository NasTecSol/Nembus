package enrichment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ApplicationErrorCode is provider- and transport-neutral. E2 can map these
// distinctions to HTTP semantics without coupling this deterministic engine
// to handlers or permissions.
type ApplicationErrorCode string

const (
	ApplicationErrorNotFound            ApplicationErrorCode = "not_found"
	ApplicationErrorNotApproved         ApplicationErrorCode = "not_approved"
	ApplicationErrorStale               ApplicationErrorCode = "stale"
	ApplicationErrorInvalidProposal     ApplicationErrorCode = "invalid_proposal"
	ApplicationErrorCanonicalConflict   ApplicationErrorCode = "canonical_target_conflict"
	ApplicationErrorConditionalConflict ApplicationErrorCode = "conditional_write_conflict"
	ApplicationErrorAuditFailure        ApplicationErrorCode = "audit_failure"
	ApplicationErrorPersistence         ApplicationErrorCode = "persistence"
)

// ApplicationError is deliberately free of HTTP status codes.
type ApplicationError struct {
	Code    ApplicationErrorCode
	Message string
	Err     error
}

func (e *ApplicationError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e *ApplicationError) Unwrap() error { return e.Err }

// These sentinels let a tenant store classify scoped row misses without
// making the domain layer depend on pgx.
var (
	ErrApplicationSuggestionNotFound = errors.New("application suggestion not found")
	ErrApplicationProductNotFound    = errors.New("application product not found")
	ErrApplicationTransitionConflict = errors.New("application lifecycle transition conflict")
)

// ApplyPlan is the complete E1 mutation surface. A nil field means that field
// is not reachable by the application transaction.
type ApplyPlan struct {
	BrandID       *int32
	CategoryID    *int32
	Description   *string
	ChangedFields []string
}

type ApplicationAudit struct {
	OrganizationID     int32
	SuggestionID       int32
	ProductID          int32
	ApplierUserID      int32
	OldStatus          SuggestionStatus
	NewStatus          SuggestionStatus
	ChangedFields      []string
	OldBrandID         *int32
	NewBrandID         *int32
	OldCategoryID      *int32
	NewCategoryID      *int32
	DescriptionChanged bool
}

type ApplicationResult struct {
	SuggestionID   int32
	ProductID      int32
	Status         SuggestionStatus
	AppliedAt      *time.Time
	ChangedFields  []string
	AlreadyApplied bool
}

// ProductEnrichmentApplicationStore is constructed from the request's
// tenant-local repository. It has no master/control-plane implementation.
type ProductEnrichmentApplicationStore interface {
	BeginProductEnrichmentApplication(context.Context) (ProductEnrichmentApplicationTransaction, error)
}

// ProductEnrichmentApplicationTransaction is the one-transaction E1 seam.
// Implementations must preserve the lock order: suggestion, then product.
type ProductEnrichmentApplicationTransaction interface {
	LockProductEnrichmentSuggestion(context.Context, int32, int32) (ReviewSuggestionRecord, error)
	LockProductEnrichmentSnapshot(context.Context, int32, int32) (EnrichmentSourceSnapshot, error)
	ResolveBrandApplicationTarget(context.Context, *int32, string) (*ReviewCanonicalTarget, error)
	ResolveCategoryApplicationTarget(context.Context, *int32, string) (*ReviewCanonicalTarget, error)
	ApplyProductEnrichmentFields(context.Context, int32, int32, ApplyPlan) (int64, error)
	MarkProductEnrichmentSuggestionApplied(context.Context, int32, int32) (ReviewSuggestionRecord, error)
	InsertProductEnrichmentApplicationAudit(context.Context, ApplicationAudit) error
	Commit(context.Context) error
	Rollback(context.Context) error
}

type ProductEnrichmentApplicationService struct {
	store ProductEnrichmentApplicationStore
}

func NewProductEnrichmentApplicationService(store ProductEnrichmentApplicationStore) *ProductEnrichmentApplicationService {
	return &ProductEnrichmentApplicationService{store: store}
}

// ApplyApprovedSuggestion deterministically applies one approved suggestion
// inside one tenant-local transaction. It performs no permission check; E2
// owns authenticated endpoint authorization.
func (s *ProductEnrichmentApplicationService) ApplyApprovedSuggestion(ctx context.Context, organizationID, suggestionID, applierUserID int32) (ApplicationResult, error) {
	if s == nil || s.store == nil {
		return ApplicationResult{}, applicationPersistence("application store is not configured", nil)
	}
	if organizationID <= 0 || suggestionID <= 0 || applierUserID <= 0 {
		return ApplicationResult{}, applicationError(ApplicationErrorInvalidProposal, "organization, suggestion, and applier IDs must be positive", nil)
	}

	tx, err := s.store.BeginProductEnrichmentApplication(ctx)
	if err != nil {
		return ApplicationResult{}, applicationPersistence("begin application transaction", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	suggestion, err := tx.LockProductEnrichmentSuggestion(ctx, organizationID, suggestionID)
	if err != nil {
		if errors.Is(err, ErrApplicationSuggestionNotFound) {
			return ApplicationResult{}, applicationError(ApplicationErrorNotFound, "enrichment suggestion not found", err)
		}
		return ApplicationResult{}, applicationPersistence("lock enrichment suggestion", err)
	}
	if suggestion.Status == SuggestionStatusApplied {
		if err := tx.Commit(ctx); err != nil {
			return ApplicationResult{}, applicationPersistence("commit already-applied result", err)
		}
		committed = true
		return applicationResultFromRecord(suggestion, nil, true), nil
	}
	if suggestion.Status != SuggestionStatusApproved {
		return ApplicationResult{}, applicationError(ApplicationErrorNotApproved, "enrichment suggestion is not approved", nil)
	}
	if suggestion.OrganizationID != organizationID || suggestion.ProductID <= 0 {
		return ApplicationResult{}, applicationError(ApplicationErrorNotFound, "enrichment suggestion is outside the selected organization", nil)
	}

	current, err := tx.LockProductEnrichmentSnapshot(ctx, organizationID, suggestion.ProductID)
	if err != nil {
		if errors.Is(err, ErrApplicationProductNotFound) {
			return ApplicationResult{}, applicationError(ApplicationErrorNotFound, "enrichment product not found", err)
		}
		return ApplicationResult{}, applicationPersistence("lock current enrichment product", err)
	}
	if err := validateApplicationSource(suggestion, current); err != nil {
		return ApplicationResult{}, err
	}

	currentFingerprint, err := FingerprintSnapshot(current)
	if err != nil {
		return ApplicationResult{}, applicationPersistence("fingerprint current enrichment product", err)
	}
	if currentFingerprint != suggestion.SourceDataFingerprint {
		return ApplicationResult{}, applicationError(ApplicationErrorStale, "enrichment suggestion source is stale", nil)
	}

	brand, category, description, unsupported, err := decodeProposals(suggestion)
	if err != nil {
		return ApplicationResult{}, applicationError(ApplicationErrorInvalidProposal, "decode approved enrichment proposal", err)
	}
	structuredCurrent, err := StructuredCurrent(current)
	if err != nil {
		return ApplicationResult{}, applicationPersistence("marshal current enrichment context", err)
	}
	proposals := ProposalSet{Brand: brand, Category: category, Description: description, UnsupportedSemantics: unsupported}
	if err := proposals.Validate(structuredCurrent); err != nil {
		return ApplicationResult{}, applicationError(ApplicationErrorInvalidProposal, "approved enrichment proposal is invalid", err)
	}

	plan, err := s.buildApplicationPlan(ctx, tx, current, brand, category, description)
	if err != nil {
		return ApplicationResult{}, err
	}
	if plan.BrandID != nil || plan.CategoryID != nil || plan.Description != nil {
		rows, err := tx.ApplyProductEnrichmentFields(ctx, organizationID, suggestion.ProductID, plan)
		if err != nil {
			return ApplicationResult{}, applicationPersistence("apply narrow product enrichment fields", err)
		}
		if rows != 1 {
			return ApplicationResult{}, applicationError(ApplicationErrorConditionalConflict, "product changed during enrichment application", nil)
		}
	}

	updated, err := tx.MarkProductEnrichmentSuggestionApplied(ctx, organizationID, suggestionID)
	if err != nil {
		if errors.Is(err, ErrApplicationTransitionConflict) {
			return ApplicationResult{}, applicationError(ApplicationErrorConditionalConflict, "enrichment suggestion changed during application", err)
		}
		return ApplicationResult{}, applicationPersistence("mark enrichment suggestion applied", err)
	}
	if updated.Status != SuggestionStatusApplied {
		return ApplicationResult{}, applicationError(ApplicationErrorConditionalConflict, "enrichment suggestion was not transitioned to applied", nil)
	}
	if err := tx.InsertProductEnrichmentApplicationAudit(ctx, ApplicationAudit{
		OrganizationID:     organizationID,
		SuggestionID:       suggestionID,
		ProductID:          suggestion.ProductID,
		ApplierUserID:      applierUserID,
		OldStatus:          SuggestionStatusApproved,
		NewStatus:          SuggestionStatusApplied,
		ChangedFields:      append([]string(nil), plan.ChangedFields...),
		OldBrandID:         identityID(current.Brand),
		NewBrandID:         plan.BrandID,
		OldCategoryID:      identityID(current.Category),
		NewCategoryID:      plan.CategoryID,
		DescriptionChanged: plan.Description != nil,
	}); err != nil {
		return ApplicationResult{}, applicationError(ApplicationErrorAuditFailure, "write enrichment application audit", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return ApplicationResult{}, applicationPersistence("commit enrichment application", err)
	}
	committed = true
	return applicationResultFromRecord(updated, plan.ChangedFields, false), nil
}

func (s *ProductEnrichmentApplicationService) buildApplicationPlan(ctx context.Context, tx ProductEnrichmentApplicationTransaction, current EnrichmentSourceSnapshot, brand *BrandProposal, category *CategoryProposal, description *DescriptionProposal) (ApplyPlan, error) {
	plan := ApplyPlan{}
	if brand != nil {
		switch brand.Action {
		case ActionKeepExisting, ActionNoMatch:
		case ActionMatchExisting:
			if current.Brand.Resolved() {
				return ApplyPlan{}, applicationError(ApplicationErrorCanonicalConflict, "structured brand takes precedence over approved brand match", nil)
			}
			target, err := tx.ResolveBrandApplicationTarget(ctx, brand.TargetID, brand.TargetCode)
			if err != nil {
				return ApplyPlan{}, applicationPersistence("revalidate approved brand target", err)
			}
			if target == nil || !canonicalTargetMatches(target, brand.TargetID, brand.TargetCode) {
				return ApplyPlan{}, applicationError(ApplicationErrorCanonicalConflict, "approved brand target is missing, inactive, or mismatched", nil)
			}
			id := target.ID
			plan.BrandID = &id
			plan.ChangedFields = append(plan.ChangedFields, "brand_id")
		case ActionProposeNew, ActionUnsupportedTarget:
			return ApplyPlan{}, applicationError(ApplicationErrorCanonicalConflict, "approved brand action cannot create or apply taxonomy", nil)
		default:
			return ApplyPlan{}, applicationError(ApplicationErrorInvalidProposal, "unsupported brand action", nil)
		}
	}
	if category != nil {
		switch category.Action {
		case ActionKeepExisting, ActionNoMatch:
		case ActionMatchExisting:
			if current.Category.Resolved() {
				return ApplyPlan{}, applicationError(ApplicationErrorCanonicalConflict, "structured category takes precedence over approved category match", nil)
			}
			target, err := tx.ResolveCategoryApplicationTarget(ctx, category.TargetID, category.TargetCode)
			if err != nil {
				return ApplyPlan{}, applicationPersistence("revalidate approved category target", err)
			}
			if target == nil || !canonicalTargetMatches(target, category.TargetID, category.TargetCode) {
				return ApplyPlan{}, applicationError(ApplicationErrorCanonicalConflict, "approved category target is missing, inactive, or mismatched", nil)
			}
			id := target.ID
			plan.CategoryID = &id
			plan.ChangedFields = append(plan.ChangedFields, "category_id")
		case ActionProposeNew, ActionUnsupportedTarget:
			return ApplyPlan{}, applicationError(ApplicationErrorCanonicalConflict, "approved category action cannot create or apply taxonomy", nil)
		default:
			return ApplyPlan{}, applicationError(ApplicationErrorInvalidProposal, "unsupported category action", nil)
		}
	}
	if description != nil {
		switch description.Action {
		case ActionKeepExisting, ActionNoMatch:
		case ActionProposeNew:
			if strings.TrimSpace(current.Description) != "" {
				return ApplyPlan{}, applicationError(ApplicationErrorCanonicalConflict, "structured description takes precedence over approved description", nil)
			}
			value, err := NormalizeProposedDescription(description.Value)
			if err != nil {
				return ApplyPlan{}, applicationError(ApplicationErrorInvalidProposal, "approved description is invalid", err)
			}
			plan.Description = &value
			plan.ChangedFields = append(plan.ChangedFields, "description")
		case ActionMatchExisting, ActionUnsupportedTarget:
			return ApplyPlan{}, applicationError(ApplicationErrorInvalidProposal, "unsupported description action", nil)
		default:
			return ApplyPlan{}, applicationError(ApplicationErrorInvalidProposal, "unsupported description action", nil)
		}
	}
	return plan, nil
}

func canonicalTargetMatches(target *ReviewCanonicalTarget, targetID *int32, targetCode string) bool {
	if target == nil || target.ID <= 0 {
		return false
	}
	if targetID == nil && strings.TrimSpace(targetCode) == "" {
		return false
	}
	if targetID != nil {
		if *targetID <= 0 || target.ID != *targetID {
			return false
		}
	}
	if code := strings.TrimSpace(targetCode); code != "" && target.Code != code {
		return false
	}
	return true
}

func validateApplicationSource(row ReviewSuggestionRecord, current EnrichmentSourceSnapshot) error {
	if current.SourceSystem != SourceSystemSAP || !current.ProductType.Valid() || strings.TrimSpace(current.SourceItemCode) == "" || strings.TrimSpace(current.SourceItemName) == "" {
		return applicationError(ApplicationErrorStale, "current product source context is no longer valid", nil)
	}
	if current.OrganizationID != row.OrganizationID || current.ProductID != row.ProductID || current.SourceItemCode != row.SourceItemCode || current.SourceItemName != row.SourceItemName {
		return applicationError(ApplicationErrorStale, "current product source identity changed", nil)
	}
	var prior EnrichmentSourceSnapshot
	if err := json.Unmarshal(row.StructuredCurrent, &prior); err != nil {
		return applicationError(ApplicationErrorInvalidProposal, "stored structured enrichment context is invalid", err)
	}
	if prior.SourceSystem != current.SourceSystem || prior.SourceItemCode != current.SourceItemCode || prior.SourceItemName != current.SourceItemName || prior.ProductType != current.ProductType {
		return applicationError(ApplicationErrorStale, "current product source context changed", nil)
	}
	return nil
}

func identityID(identity interface{ Resolved() bool }) *int32 {
	switch value := identity.(type) {
	case *BrandIdentity:
		if value != nil && value.ID > 0 {
			id := value.ID
			return &id
		}
	case *CategoryIdentity:
		if value != nil && value.ID > 0 {
			id := value.ID
			return &id
		}
	}
	return nil
}

func applicationResultFromRecord(row ReviewSuggestionRecord, changedFields []string, alreadyApplied bool) ApplicationResult {
	return ApplicationResult{SuggestionID: row.ID, ProductID: row.ProductID, Status: row.Status, AppliedAt: row.AppliedAt, ChangedFields: append([]string(nil), changedFields...), AlreadyApplied: alreadyApplied}
}

func applicationError(code ApplicationErrorCode, message string, err error) error {
	return &ApplicationError{Code: code, Message: message, Err: err}
}

func applicationPersistence(message string, err error) error {
	return applicationError(ApplicationErrorPersistence, message, err)
}
