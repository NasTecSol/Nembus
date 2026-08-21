package enrichment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const ReviewPermissionCode = "product_enrichment:review"

type ReviewListStatus string

const (
	ReviewStatusInReview ReviewListStatus = "in_review"
	ReviewStatusApproved ReviewListStatus = "approved"
	ReviewStatusRejected ReviewListStatus = "rejected"
)

func (s ReviewListStatus) Valid() bool {
	return s == ReviewStatusInReview || s == ReviewStatusApproved || s == ReviewStatusRejected
}

type ReviewSuggestionRecord struct {
	ID                    int32
	OrganizationID        int32
	ProductID             int32
	SourceItemCode        string
	SourceItemName        string
	SourceDataFingerprint string
	ContractVersion       string
	StructuredCurrent     json.RawMessage
	ProposedBrand         json.RawMessage
	ProposedCategory      json.RawMessage
	ProposedDescription   json.RawMessage
	UnsupportedSemantics  json.RawMessage
	Source                string
	Provider              string
	Model                 string
	ModelVersion          string
	Status                SuggestionStatus
	ReviewerID            *int32
	ReviewedAt            *time.Time
	AppliedAt             *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type ReviewCanonicalTarget struct {
	ID   int32
	Code string
	Name string
}

type ReviewAudit struct {
	SuggestionID   int32
	ProductID      int32
	OrganizationID int32
	ReviewerID     int32
	OldStatus      SuggestionStatus
	NewStatus      SuggestionStatus
	Event          string
}

type ReviewTransaction interface {
	Approve(context.Context, int32, int32, int32) (ReviewSuggestionRecord, error)
	Reject(context.Context, int32, int32, int32) (ReviewSuggestionRecord, error)
	InsertAudit(context.Context, ReviewAudit) error
	Commit(context.Context) error
	Rollback(context.Context) error
}

type ReviewStore interface {
	ListReviewSuggestions(context.Context, int32, ReviewListStatus, int32, int32) ([]ReviewSuggestionRecord, error)
	GetReviewSuggestion(context.Context, int32, int32) (ReviewSuggestionRecord, error)
	LoadSAPProductEnrichmentSnapshotByID(context.Context, int32, int32) (EnrichmentSourceSnapshot, error)
	ResolveBrandReviewTarget(context.Context, *int32, string) (*ReviewCanonicalTarget, error)
	ResolveCategoryReviewTarget(context.Context, *int32, string) (*ReviewCanonicalTarget, error)
	BeginReviewTransaction(context.Context) (ReviewTransaction, error)
}

type ReviewError struct {
	Code    string
	Message string
	Err     error
}

func (e *ReviewError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e *ReviewError) Unwrap() error { return e.Err }

const (
	ReviewErrorNotFound      = "not_found"
	ReviewErrorConflict      = "conflict"
	ReviewErrorStale         = "stale"
	ReviewErrorNotReviewable = "not_reviewable"
	ReviewErrorBadRequest    = "bad_request"
	ReviewErrorInternal      = "internal"
)

var ErrReviewSuggestionNotFound = errors.New("review suggestion not found")

type ReviewService struct {
	store ReviewStore
}

func NewReviewService(store ReviewStore) *ReviewService {
	return &ReviewService{store: store}
}

func (s *ReviewService) ListSuggestions(ctx context.Context, organizationID int32, status ReviewListStatus, limit, offset int32) ([]ReviewListItem, error) {
	if s == nil || s.store == nil {
		return nil, reviewInternal("review store is not configured", nil)
	}
	if organizationID <= 0 || !status.Valid() || limit <= 0 || offset < 0 {
		return nil, reviewBadRequest("invalid review list parameters", nil)
	}
	rows, err := s.store.ListReviewSuggestions(ctx, organizationID, status, limit, offset)
	if err != nil {
		return nil, reviewInternal("list enrichment suggestions", err)
	}
	items := make([]ReviewListItem, 0, len(rows))
	for _, row := range rows {
		item, err := listItemFromRecord(row)
		if err != nil {
			return nil, reviewInternal("decode enrichment suggestion", err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *ReviewService) GetSuggestion(ctx context.Context, organizationID, suggestionID int32) (ReviewDetail, error) {
	if s == nil || s.store == nil {
		return ReviewDetail{}, reviewInternal("review store is not configured", nil)
	}
	row, err := s.store.GetReviewSuggestion(ctx, organizationID, suggestionID)
	if err != nil {
		if errors.Is(err, ErrReviewSuggestionNotFound) {
			return ReviewDetail{}, reviewNotFound("enrichment suggestion not found", err)
		}
		return ReviewDetail{}, reviewInternal("get enrichment suggestion", err)
	}
	current, err := s.store.LoadSAPProductEnrichmentSnapshotByID(ctx, organizationID, row.ProductID)
	if err != nil {
		return ReviewDetail{}, reviewInternal("load current product enrichment context", err)
	}
	analysis, err := s.analyze(ctx, row, current)
	if err != nil {
		return ReviewDetail{}, err
	}
	return detailFromRecord(row, current, analysis), nil
}

func (s *ReviewService) ApproveSuggestion(ctx context.Context, organizationID, suggestionID, reviewerID int32) (ReviewDetail, error) {
	if s == nil || s.store == nil {
		return ReviewDetail{}, reviewInternal("review store is not configured", nil)
	}
	row, err := s.store.GetReviewSuggestion(ctx, organizationID, suggestionID)
	if err != nil {
		if errors.Is(err, ErrReviewSuggestionNotFound) {
			return ReviewDetail{}, reviewNotFound("enrichment suggestion not found", err)
		}
		return ReviewDetail{}, reviewInternal("get enrichment suggestion", err)
	}
	if row.Status != SuggestionStatusInReview {
		return ReviewDetail{}, reviewConflict("enrichment suggestion is no longer in review", nil)
	}
	current, err := s.store.LoadSAPProductEnrichmentSnapshotByID(ctx, organizationID, row.ProductID)
	if err != nil {
		return ReviewDetail{}, reviewConflict("current product state cannot be revalidated", err)
	}
	analysis, err := s.analyze(ctx, row, current)
	if err != nil {
		return ReviewDetail{}, err
	}
	if !analysis.Approval.Approvable {
		if analysis.Approval.Stale {
			return ReviewDetail{}, reviewStale("enrichment suggestion source is stale", nil)
		}
		return ReviewDetail{}, reviewNotReviewable("enrichment suggestion is not approvable", nil)
	}

	tx, err := s.store.BeginReviewTransaction(ctx)
	if err != nil {
		return ReviewDetail{}, reviewInternal("begin review transaction", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	updated, err := tx.Approve(ctx, organizationID, suggestionID, reviewerID)
	if err != nil {
		return ReviewDetail{}, reviewConflict("enrichment suggestion changed during review", err)
	}
	if err := tx.InsertAudit(ctx, ReviewAudit{
		SuggestionID: suggestionID, ProductID: row.ProductID, OrganizationID: organizationID,
		ReviewerID: reviewerID, OldStatus: SuggestionStatusInReview,
		NewStatus: SuggestionStatusApproved, Event: "product_enrichment.approved",
	}); err != nil {
		return ReviewDetail{}, reviewInternal("write review audit", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return ReviewDetail{}, reviewInternal("commit review transaction", err)
	}
	return detailFromRecord(updated, current, analysis), nil
}

func (s *ReviewService) RejectSuggestion(ctx context.Context, organizationID, suggestionID, reviewerID int32) (ReviewDetail, error) {
	if s == nil || s.store == nil {
		return ReviewDetail{}, reviewInternal("review store is not configured", nil)
	}
	row, err := s.store.GetReviewSuggestion(ctx, organizationID, suggestionID)
	if err != nil {
		if errors.Is(err, ErrReviewSuggestionNotFound) {
			return ReviewDetail{}, reviewNotFound("enrichment suggestion not found", err)
		}
		return ReviewDetail{}, reviewInternal("get enrichment suggestion", err)
	}
	if row.Status != SuggestionStatusInReview {
		return ReviewDetail{}, reviewConflict("enrichment suggestion is no longer in review", nil)
	}

	tx, err := s.store.BeginReviewTransaction(ctx)
	if err != nil {
		return ReviewDetail{}, reviewInternal("begin review transaction", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	updated, err := tx.Reject(ctx, organizationID, suggestionID, reviewerID)
	if err != nil {
		return ReviewDetail{}, reviewConflict("enrichment suggestion changed during review", err)
	}
	if err := tx.InsertAudit(ctx, ReviewAudit{
		SuggestionID: suggestionID, ProductID: row.ProductID, OrganizationID: organizationID,
		ReviewerID: reviewerID, OldStatus: SuggestionStatusInReview,
		NewStatus: SuggestionStatusRejected, Event: "product_enrichment.rejected",
	}); err != nil {
		return ReviewDetail{}, reviewInternal("write review audit", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return ReviewDetail{}, reviewInternal("commit review transaction", err)
	}
	current, currentErr := s.store.LoadSAPProductEnrichmentSnapshotByID(ctx, organizationID, row.ProductID)
	if currentErr != nil {
		return detailFromRecord(updated, EnrichmentSourceSnapshot{}, ReviewAnalysis{Approval: ApprovalAnalysis{}}), nil
	}
	analysis, _ := s.analyze(ctx, updated, current)
	return detailFromRecord(updated, current, analysis), nil
}

type ReviewListItem struct {
	SuggestionID        int32            `json:"suggestion_id"`
	ProductID           int32            `json:"product_id"`
	SourceItemCode      string           `json:"source_item_code"`
	SourceItemName      string           `json:"source_item_name"`
	Status              SuggestionStatus `json:"status"`
	Provider            string           `json:"provider,omitempty"`
	Model               string           `json:"model,omitempty"`
	ModelVersion        string           `json:"model_version,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
	ProposedBrand       *ProposalSummary `json:"proposed_brand,omitempty"`
	ProposedCategory    *ProposalSummary `json:"proposed_category,omitempty"`
	ProposedDescription *ProposalSummary `json:"proposed_description,omitempty"`
}

type ProposalSummary struct {
	Action        ProposalAction `json:"action"`
	TargetID      *int32         `json:"target_id,omitempty"`
	TargetCode    string         `json:"target_code,omitempty"`
	CanonicalName string         `json:"canonical_name,omitempty"`
	Value         string         `json:"value,omitempty"`
	Confidence    float64        `json:"confidence"`
}

type ReviewDetail struct {
	SourceIdentity       SourceIdentityView       `json:"source_identity"`
	InferenceSnapshot    InferenceSnapshotView    `json:"inference_snapshot"`
	CurrentAuthoritative CurrentAuthoritativeView `json:"current_authoritative_state"`
	ProposedBrand        *BrandProposal           `json:"proposed_brand,omitempty"`
	ProposedCategory     *CategoryProposal        `json:"proposed_category,omitempty"`
	ProposedDescription  *DescriptionProposal     `json:"proposed_description,omitempty"`
	UnsupportedSemantics []UnsupportedSemantic    `json:"unsupported_semantics,omitempty"`
	ProviderContext      ProviderContextView      `json:"provider_context"`
	ReviewState          ReviewStateView          `json:"review_state"`
	Safety               ApprovalAnalysis         `json:"safety"`
}

type SourceIdentityView struct {
	SuggestionID   int32  `json:"suggestion_id"`
	ProductID      int32  `json:"product_id"`
	SourceItemCode string `json:"source_item_code"`
	SourceItemName string `json:"source_item_name"`
}

type InferenceSnapshotView struct {
	SourceSystem string            `json:"source_system"`
	Description  string            `json:"description,omitempty"`
	ProductType  ProductType       `json:"product_type"`
	Brand        *BrandIdentity    `json:"brand,omitempty"`
	Category     *CategoryIdentity `json:"category,omitempty"`
}

type CurrentAuthoritativeView struct {
	ProductType ProductType       `json:"product_type"`
	Brand       *BrandIdentity    `json:"brand,omitempty"`
	Category    *CategoryIdentity `json:"category,omitempty"`
	Description string            `json:"description,omitempty"`
}

type ProviderContextView struct {
	Provider     string `json:"provider,omitempty"`
	Model        string `json:"model,omitempty"`
	ModelVersion string `json:"model_version,omitempty"`
}

type ReviewStateView struct {
	Status     SuggestionStatus `json:"status"`
	ReviewerID *int32           `json:"reviewer_id,omitempty"`
	ReviewedAt *time.Time       `json:"reviewed_at,omitempty"`
	AppliedAt  *time.Time       `json:"applied_at,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

type ApprovalAnalysis struct {
	Stale           bool     `json:"stale"`
	StaleReasons    []string `json:"stale_reasons,omitempty"`
	Approvable      bool     `json:"approvable"`
	BlockingReasons []string `json:"blocking_reasons,omitempty"`
}

func (s *ReviewService) analyze(ctx context.Context, row ReviewSuggestionRecord, current EnrichmentSourceSnapshot) (ReviewAnalysis, error) {
	brand, category, description, unsupported, err := decodeProposals(row)
	if err != nil {
		return ReviewAnalysis{}, reviewConflict("persisted enrichment proposal is invalid", err)
	}
	currentStructured, err := StructuredCurrent(current)
	if err != nil {
		return ReviewAnalysis{}, reviewInternal("marshal current enrichment context", err)
	}
	currentFingerprint, err := FingerprintSnapshot(current)
	if err != nil {
		return ReviewAnalysis{}, reviewInternal("fingerprint current enrichment context", err)
	}
	analysis := ReviewAnalysis{}
	if currentFingerprint != row.SourceDataFingerprint {
		analysis.Approval.Stale = true
		analysis.Approval.StaleReasons = staleReasons(row, current)
		if len(analysis.Approval.StaleReasons) == 0 {
			analysis.Approval.StaleReasons = []string{"source_data_fingerprint_changed"}
		}
	}
	proposalSet := ProposalSet{Brand: brand, Category: category, Description: description, UnsupportedSemantics: unsupported}
	if err := proposalSet.Validate(currentStructured); err != nil {
		analysis.Approval.BlockingReasons = append(analysis.Approval.BlockingReasons, "proposal_contract_invalid")
	}
	brandTarget, categoryTarget := (*ReviewCanonicalTarget)(nil), (*ReviewCanonicalTarget)(nil)
	if brand != nil && brand.Action == ActionMatchExisting {
		brandTarget, err = s.store.ResolveBrandReviewTarget(ctx, brand.TargetID, brand.TargetCode)
		if err != nil {
			return ReviewAnalysis{}, reviewInternal("revalidate brand target", err)
		}
		if brandTarget == nil {
			analysis.Approval.BlockingReasons = append(analysis.Approval.BlockingReasons, "brand_match_target_invalid")
		}
	}
	if category != nil && category.Action == ActionMatchExisting {
		categoryTarget, err = s.store.ResolveCategoryReviewTarget(ctx, category.TargetID, category.TargetCode)
		if err != nil {
			return ReviewAnalysis{}, reviewInternal("revalidate category target", err)
		}
		if categoryTarget == nil {
			analysis.Approval.BlockingReasons = append(analysis.Approval.BlockingReasons, "category_match_target_invalid")
		}
	}
	if reasons := approvalBlockingReasons(row, current, brand, category, description, brandTarget, categoryTarget); len(reasons) > 0 {
		analysis.Approval.BlockingReasons = append(analysis.Approval.BlockingReasons, reasons...)
	}
	if row.Status == SuggestionStatusInReview && !analysis.Approval.Stale && len(analysis.Approval.BlockingReasons) == 0 {
		analysis.Approval.Approvable = true
	}
	return analysis, nil
}

type ReviewAnalysis struct {
	Approval ApprovalAnalysis
}

func CanApproveSuggestion(row ReviewSuggestionRecord, current EnrichmentSourceSnapshot, currentFingerprint string, brandTarget, categoryTarget *ReviewCanonicalTarget) ApprovalAnalysis {
	brand, category, description, unsupported, err := decodeProposals(row)
	analysis := ApprovalAnalysis{}
	if err != nil {
		analysis.BlockingReasons = append(analysis.BlockingReasons, "proposal_contract_invalid")
		return analysis
	}
	structuredCurrent, err := StructuredCurrent(current)
	if err != nil {
		analysis.BlockingReasons = append(analysis.BlockingReasons, "proposal_contract_invalid")
		return analysis
	}
	if err := (ProposalSet{Brand: brand, Category: category, Description: description, UnsupportedSemantics: unsupported}).Validate(structuredCurrent); err != nil {
		analysis.BlockingReasons = append(analysis.BlockingReasons, "proposal_contract_invalid")
	}
	if currentFingerprint != row.SourceDataFingerprint {
		analysis.Stale = true
		analysis.StaleReasons = staleReasons(row, current)
	}
	analysis.BlockingReasons = append(analysis.BlockingReasons, approvalBlockingReasons(row, current, brand, category, description, brandTarget, categoryTarget)...)
	if row.Status != SuggestionStatusInReview {
		analysis.BlockingReasons = append(analysis.BlockingReasons, "status_not_in_review")
	}
	analysis.Approvable = !analysis.Stale && len(analysis.BlockingReasons) == 0
	return analysis
}

func approvalBlockingReasons(row ReviewSuggestionRecord, current EnrichmentSourceSnapshot, brand *BrandProposal, category *CategoryProposal, description *DescriptionProposal, brandTarget, categoryTarget *ReviewCanonicalTarget) []string {
	reasons := []string{}
	if current.SourceSystem != SourceSystemSAP {
		reasons = append(reasons, "source_system_not_sap")
	}
	if !current.ProductType.Valid() {
		reasons = append(reasons, "invalid_product_type")
	}
	if current.ProductID != 0 && row.ProductID != current.ProductID {
		reasons = append(reasons, "product_identity_changed")
	}
	if row.SourceItemCode != "" && current.SourceItemCode != row.SourceItemCode {
		reasons = append(reasons, "source_item_code_changed")
	}
	if brand != nil {
		if current.Brand.Resolved() && brand.Action != ActionKeepExisting {
			reasons = append(reasons, "structured_brand_precedence_violation")
		}
		if !current.Brand.Resolved() && brand.Action == ActionKeepExisting {
			reasons = append(reasons, "brand_keep_existing_without_current_brand")
		}
		if brand.Action == ActionProposeNew {
			reasons = append(reasons, "brand_propose_new_requires_review_resolution")
		}
		if brand.Action == ActionMatchExisting && brandTarget == nil {
			reasons = append(reasons, "brand_match_target_invalid")
		}
		if brand.Action == ActionUnsupportedTarget {
			reasons = append(reasons, "brand_unsupported_target")
		}
	}
	if category != nil {
		if current.Category.Resolved() && category.Action != ActionKeepExisting {
			reasons = append(reasons, "structured_category_precedence_violation")
		}
		if !current.Category.Resolved() && category.Action == ActionKeepExisting {
			reasons = append(reasons, "category_keep_existing_without_current_category")
		}
		if category.Action == ActionProposeNew {
			reasons = append(reasons, "category_propose_new_requires_review_resolution")
		}
		if category.Action == ActionMatchExisting && categoryTarget == nil {
			reasons = append(reasons, "category_match_target_invalid")
		}
		if category.Action == ActionUnsupportedTarget {
			reasons = append(reasons, "category_unsupported_target")
		}
	}
	if description != nil {
		if description.Action == ActionProposeNew {
			if strings.TrimSpace(current.Description) != "" {
				reasons = append(reasons, "description_precedence_violation")
			}
		} else if description.Action == ActionUnsupportedTarget {
			reasons = append(reasons, "description_unsupported_target")
		}
	}
	return uniqueStrings(reasons)
}

func staleReasons(row ReviewSuggestionRecord, current EnrichmentSourceSnapshot) []string {
	var prior EnrichmentSourceSnapshot
	if json.Unmarshal(row.StructuredCurrent, &prior) != nil {
		return []string{"source_data_fingerprint_changed"}
	}
	reasons := []string{}
	if prior.SourceItemName != current.SourceItemName {
		reasons = append(reasons, "source_item_name_changed")
	}
	if prior.ProductType != current.ProductType {
		reasons = append(reasons, "product_type_changed")
	}
	if !prior.Brand.Resolved() && current.Brand.Resolved() {
		reasons = append(reasons, "structured_brand_resolved")
	}
	if !prior.Category.Resolved() && current.Category.Resolved() {
		reasons = append(reasons, "structured_category_populated")
	}
	if strings.TrimSpace(prior.Description) == "" && strings.TrimSpace(current.Description) != "" {
		var proposed DescriptionProposal
		if decodeOptional(row.ProposedDescription, &proposed) == nil && proposed.Action == ActionProposeNew {
			reasons = append(reasons, "description_populated")
		}
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "source_data_fingerprint_changed")
	}
	return uniqueStrings(reasons)
}

func decodeProposals(row ReviewSuggestionRecord) (*BrandProposal, *CategoryProposal, *DescriptionProposal, []UnsupportedSemantic, error) {
	var brand BrandProposal
	var category CategoryProposal
	var description DescriptionProposal
	var unsupported []UnsupportedSemantic
	if err := decodeOptional(row.ProposedBrand, &brand); err != nil {
		return nil, nil, nil, nil, err
	}
	if err := decodeOptional(row.ProposedCategory, &category); err != nil {
		return nil, nil, nil, nil, err
	}
	if err := decodeOptional(row.ProposedDescription, &description); err != nil {
		return nil, nil, nil, nil, err
	}
	if err := decodeOptional(row.UnsupportedSemantics, &unsupported); err != nil {
		return nil, nil, nil, nil, err
	}
	var brandPtr *BrandProposal
	var categoryPtr *CategoryProposal
	var descriptionPtr *DescriptionProposal
	if hasJSONValue(row.ProposedBrand) {
		brandPtr = &brand
	}
	if hasJSONValue(row.ProposedCategory) {
		categoryPtr = &category
	}
	if hasJSONValue(row.ProposedDescription) {
		descriptionPtr = &description
	}
	return brandPtr, categoryPtr, descriptionPtr, unsupported, nil
}

func decodeOptional[T any](raw json.RawMessage, target *T) error {
	if !hasJSONValue(raw) {
		return nil
	}
	return json.Unmarshal(raw, target)
}

func hasJSONValue(raw json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(raw))
	return trimmed != "" && trimmed != "null"
}

func listItemFromRecord(row ReviewSuggestionRecord) (ReviewListItem, error) {
	brand, category, description, _, err := decodeProposals(row)
	if err != nil {
		return ReviewListItem{}, err
	}
	return ReviewListItem{
		SuggestionID: row.ID, ProductID: row.ProductID, SourceItemCode: row.SourceItemCode,
		SourceItemName: row.SourceItemName, Status: row.Status, Provider: row.Provider,
		Model: row.Model, ModelVersion: row.ModelVersion, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		ProposedBrand: proposalSummary(brand), ProposedCategory: categorySummary(category), ProposedDescription: descriptionSummary(description),
	}, nil
}

func proposalSummary(proposal *BrandProposal) *ProposalSummary {
	if proposal == nil {
		return nil
	}
	return &ProposalSummary{Action: proposal.Action, TargetID: proposal.TargetID, TargetCode: proposal.TargetCode, CanonicalName: proposal.CanonicalName, Confidence: proposal.Confidence}
}

func categorySummary(proposal *CategoryProposal) *ProposalSummary {
	if proposal == nil {
		return nil
	}
	return &ProposalSummary{Action: proposal.Action, TargetID: proposal.TargetID, TargetCode: proposal.TargetCode, CanonicalName: proposal.CanonicalName, Confidence: proposal.Confidence}
}

func descriptionSummary(proposal *DescriptionProposal) *ProposalSummary {
	if proposal == nil {
		return nil
	}
	return &ProposalSummary{Action: proposal.Action, Value: proposal.Value, Confidence: proposal.Confidence}
}

func detailFromRecord(row ReviewSuggestionRecord, current EnrichmentSourceSnapshot, analysis ReviewAnalysis) ReviewDetail {
	brand, category, description, unsupported, _ := decodeProposals(row)
	return ReviewDetail{
		SourceIdentity:       SourceIdentityView{SuggestionID: row.ID, ProductID: row.ProductID, SourceItemCode: row.SourceItemCode, SourceItemName: row.SourceItemName},
		InferenceSnapshot:    inferenceSnapshotView(row.StructuredCurrent),
		CurrentAuthoritative: CurrentAuthoritativeView{ProductType: current.ProductType, Brand: current.Brand, Category: current.Category, Description: current.Description},
		ProposedBrand:        brand, ProposedCategory: category, ProposedDescription: description, UnsupportedSemantics: unsupported,
		ProviderContext: ProviderContextView{Provider: row.Provider, Model: row.Model, ModelVersion: row.ModelVersion},
		ReviewState:     ReviewStateView{Status: row.Status, ReviewerID: row.ReviewerID, ReviewedAt: row.ReviewedAt, AppliedAt: row.AppliedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt},
		Safety:          analysis.Approval,
	}
}

func inferenceSnapshotView(raw json.RawMessage) InferenceSnapshotView {
	var snapshot EnrichmentSourceSnapshot
	if json.Unmarshal(raw, &snapshot) != nil {
		return InferenceSnapshotView{}
	}
	return InferenceSnapshotView{SourceSystem: snapshot.SourceSystem, Description: snapshot.Description, ProductType: snapshot.ProductType, Brand: snapshot.Brand, Category: snapshot.Category}
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; !ok {
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}

func reviewNotFound(message string, err error) error {
	return &ReviewError{Code: ReviewErrorNotFound, Message: message, Err: err}
}
func reviewConflict(message string, err error) error {
	return &ReviewError{Code: ReviewErrorConflict, Message: message, Err: err}
}
func reviewStale(message string, err error) error {
	return &ReviewError{Code: ReviewErrorStale, Message: message, Err: err}
}
func reviewNotReviewable(message string, err error) error {
	return &ReviewError{Code: ReviewErrorNotReviewable, Message: message, Err: err}
}
func reviewBadRequest(message string, err error) error {
	return &ReviewError{Code: ReviewErrorBadRequest, Message: message, Err: err}
}
func reviewInternal(message string, err error) error {
	return &ReviewError{Code: ReviewErrorInternal, Message: message, Err: err}
}
