package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

// ApprovalInput is a domain-neutral snapshot. Domain use cases own the meaning
// of Payload and must apply it only through a registered Accommodator.
type ApprovalInput struct {
	OrganizationID int32           `json:"organization_id" binding:"required"`
	Domain         string          `json:"domain" binding:"required"`
	Action         string          `json:"action" binding:"required"`
	EntityType     string          `json:"entity_type" binding:"required"`
	EntityID       *string         `json:"entity_id,omitempty"`
	RequestedBy    int32           `json:"requested_by" binding:"required"`
	Summary        string          `json:"summary" binding:"required"`
	Payload        json.RawMessage `json:"payload" binding:"required"`
	Notes          *string         `json:"notes,omitempty"`
}

type ApprovalDecision struct {
	RequestID      int64
	OrganizationID int32
	UserID         int32
	Notes          string
}

// DomainAccommodator commits/rejects a domain proposal. Implementations are
// registered by the owning module; payload interpretation stays out of this package.
type DomainAccommodator interface {
	Accommodate(context.Context, *repository.Queries, repository.ApprovalRequest, int32) error
	Reject(context.Context, *repository.Queries, repository.ApprovalRequest, string) error
}

type ApprovalsUseCase struct {
	repo          *repository.Queries
	accommodators map[string]DomainAccommodator
}

func NewApprovalsUseCase() *ApprovalsUseCase {
	return &ApprovalsUseCase{accommodators: make(map[string]DomainAccommodator)}
}

func (uc *ApprovalsUseCase) SetRepository(repo *repository.Queries) { uc.repo = repo }

func (uc *ApprovalsUseCase) RegisterAccommodator(domain string, adapter DomainAccommodator) {
	uc.accommodators[strings.ToLower(strings.TrimSpace(domain))] = adapter
}

func (uc *ApprovalsUseCase) Submit(ctx context.Context, in ApprovalInput) (repository.ApprovalRequest, error) {
	if uc.repo == nil {
		return repository.ApprovalRequest{}, errors.New("repository is not configured")
	}
	in.Domain, in.Action = strings.ToLower(strings.TrimSpace(in.Domain)), strings.ToLower(strings.TrimSpace(in.Action))
	if in.OrganizationID <= 0 || in.Domain == "" || in.Action == "" || in.EntityType == "" || strings.TrimSpace(in.Summary) == "" || len(in.Payload) == 0 || !json.Valid(in.Payload) {
		return repository.ApprovalRequest{}, errors.New("organization, domain, action, entity type, summary, and valid JSON payload are required")
	}
	policy, err := uc.repo.GetApprovalPolicy(ctx, repository.GetApprovalPolicyParams{OrganizationID: in.OrganizationID, Domain: in.Domain, Action: in.Action})
	if err != nil {
		return repository.ApprovalRequest{}, fmt.Errorf("approval policy is not configured for %s.%s", in.Domain, in.Action)
	}
	if !policy.IsEnabled || !policy.RequiresApproval {
		return repository.ApprovalRequest{}, errors.New("approval is not required for this action; execute it through its domain use case")
	}
	entityID, notes := pgtype.Text{}, pgtype.Text{}
	if in.EntityID != nil {
		entityID = pgtype.Text{String: *in.EntityID, Valid: true}
	}
	if in.Notes != nil {
		notes = pgtype.Text{String: *in.Notes, Valid: true}
	}
	requestedBy := pgtype.Int4{}
	if in.RequestedBy > 0 {
		requestedBy = pgtype.Int4{Int32: in.RequestedBy, Valid: true}
	}
	request, err := uc.repo.CreateApprovalRequest(ctx, repository.CreateApprovalRequestParams{
		OrganizationID:     in.OrganizationID,
		Domain:             in.Domain,
		Action:             in.Action,
		EntityType:         in.EntityType,
		EntityID:           entityID,
		RequestedBy:        requestedBy,
		RequiredPermission: policy.RequiredPermission,
		Summary:            in.Summary,
		Payload:            in.Payload,
		SubmissionNotes:    notes,
		Metadata:           json.RawMessage(`{}`),
	})
	if err != nil {
		return repository.ApprovalRequest{}, err
	}
	_, err = uc.repo.CreateApprovalRequestLog(ctx, repository.CreateApprovalRequestLogParams{
		ApprovalRequestID: request.ID,
		UserID:            requestedBy,
		Action:            "submitted",
		Notes:             notes,
		Metadata:          json.RawMessage(`{}`),
	})
	return request, err
}

func (uc *ApprovalsUseCase) FindPending(ctx context.Context, orgID int32, domain, action, entityType, entityID string) (*repository.ApprovalRequest, error) {
	if uc.repo == nil {
		return nil, errors.New("repository is not configured")
	}
	req, err := uc.repo.FindPendingApprovalRequestByEntity(ctx, repository.FindPendingApprovalRequestByEntityParams{
		OrganizationID: orgID,
		Domain:         strings.ToLower(strings.TrimSpace(domain)),
		Action:         strings.ToLower(strings.TrimSpace(action)),
		EntityType:     entityType,
		EntityID:       pgtype.Text{String: entityID, Valid: entityID != ""},
	})
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (uc *ApprovalsUseCase) SubmitOrUpdatePending(ctx context.Context, in ApprovalInput) (repository.ApprovalRequest, error) {
	if uc.repo == nil {
		return repository.ApprovalRequest{}, errors.New("repository is not configured")
	}
	if in.EntityID != nil && *in.EntityID != "" {
		existing, err := uc.FindPending(ctx, in.OrganizationID, in.Domain, in.Action, in.EntityType, *in.EntityID)
		if err == nil && existing != nil {
			updated, updateErr := uc.repo.UpdateApprovalRequestPayload(ctx, repository.UpdateApprovalRequestPayloadParams{
				ID:             existing.ID,
				OrganizationID: existing.OrganizationID,
				Payload:        in.Payload,
				Summary:        in.Summary,
			})
			if updateErr == nil {
				return updated, nil
			}
		}
	}
	return uc.Submit(ctx, in)
}

func (uc *ApprovalsUseCase) EnsureDefaultPolicies(ctx context.Context, orgID int32) error {
	if uc.repo == nil {
		return errors.New("repository is not configured")
	}
	defaultPolicies := []struct {
		domain     string
		action     string
		permission string
		rules      string
	}{
		{
			domain:     "catalog",
			action:     "price_change",
			permission: "catalog:pricing:approve",
			rules:      `{"price_drop_requires_approval": true, "cost_increase_threshold_pct": 10.0, "below_cost_requires_approval": true}`,
		},
		{
			domain:     "sap_invoice",
			action:     "stock_hold",
			permission: "sap:invoice:override",
			rules:      `{"hold_on_zero_stock": true, "allow_manual_reverify": true, "allow_adjustment_force": true}`,
		},
	}
	for _, p := range defaultPolicies {
		_, err := uc.repo.UpsertApprovalPolicy(ctx, repository.UpsertApprovalPolicyParams{
			OrganizationID:     orgID,
			Domain:             p.domain,
			Action:             p.action,
			IsEnabled:          true,
			RequiresApproval:   true,
			RequiredPermission: p.permission,
			Rules:              json.RawMessage(p.rules),
		})
		if err != nil {
			return fmt.Errorf("failed to ensure approval policy %s.%s: %w", p.domain, p.action, err)
		}
	}
	return nil
}


func (uc *ApprovalsUseCase) List(ctx context.Context, orgID int32, domain, status *string, page, limit int32) ([]repository.ApprovalRequest, error) {
	if uc.repo == nil {
		return nil, errors.New("repository is not configured")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	d, s := pgtype.Text{}, pgtype.Text{}
	if domain != nil && *domain != "" {
		d = pgtype.Text{String: strings.ToLower(*domain), Valid: true}
	}
	if status != nil && *status != "" {
		s = pgtype.Text{String: *status, Valid: true}
	}
	return uc.repo.ListApprovalRequests(ctx, repository.ListApprovalRequestsParams{OrganizationID: orgID, Domain: d, Status: s, OffsetCount: (page - 1) * limit, LimitCount: limit})
}

func (uc *ApprovalsUseCase) Decide(ctx context.Context, d ApprovalDecision, approve bool) (repository.ApprovalRequest, error) {
	if uc.repo == nil {
		return repository.ApprovalRequest{}, errors.New("repository is not configured")
	}
	req, err := uc.repo.GetApprovalRequest(ctx, repository.GetApprovalRequestParams{ID: d.RequestID, OrganizationID: d.OrganizationID})
	if err != nil {
		return repository.ApprovalRequest{}, errors.New("approval request not found")
	}
	if req.Status != "pending" {
		return repository.ApprovalRequest{}, errors.New("approval request is no longer pending")
	}
	if req.RequestedBy.Valid && req.RequestedBy.Int32 == d.UserID {
		return repository.ApprovalRequest{}, errors.New("requesters cannot approve or reject their own requests")
	}
	ok, err := uc.repo.CheckUserHasPermissionCode(ctx, repository.CheckUserHasPermissionCodeParams{UserID: d.UserID, Code: req.RequiredPermission})
	if err != nil {
		return repository.ApprovalRequest{}, err
	}
	if !ok {
		return repository.ApprovalRequest{}, errors.New("reviewer lacks the required permission")
	}
	adapter := uc.accommodators[req.Domain]
	if adapter == nil {
		return repository.ApprovalRequest{}, fmt.Errorf("no domain accommodator registered for %q", req.Domain)
	}
	claimed, err := uc.repo.ClaimApprovalRequest(ctx, repository.ClaimApprovalRequestParams{ID: d.RequestID, OrganizationID: d.OrganizationID})
	if err != nil {
		return repository.ApprovalRequest{}, err
	}
	if claimed != 1 {
		return repository.ApprovalRequest{}, errors.New("approval request is already being decided")
	}
	if approve {
		err = adapter.Accommodate(ctx, uc.repo, req, d.UserID)
	} else {
		err = adapter.Reject(ctx, uc.repo, req, d.Notes)
	}
	if err != nil {
		_ = uc.repo.ReleaseApprovalRequest(ctx, repository.ReleaseApprovalRequestParams{ID: d.RequestID, OrganizationID: d.OrganizationID})
		return repository.ApprovalRequest{}, fmt.Errorf("domain action failed: %w", err)
	}
	status, action := "rejected", "rejected"
	if approve {
		status, action = "approved", "approved"
	}
	updated, err := uc.repo.DecideApprovalRequest(ctx, repository.DecideApprovalRequestParams{ID: d.RequestID, OrganizationID: d.OrganizationID, Status: status, DecisionNotes: pgtype.Text{String: d.Notes, Valid: d.Notes != ""}, ReviewedBy: pgtype.Int4{Int32: d.UserID, Valid: true}})
	if err != nil {
		return repository.ApprovalRequest{}, err
	}
	_, err = uc.repo.CreateApprovalRequestLog(ctx, repository.CreateApprovalRequestLogParams{ApprovalRequestID: d.RequestID, UserID: pgtype.Int4{Int32: d.UserID, Valid: true}, Action: action, Notes: pgtype.Text{String: d.Notes, Valid: d.Notes != ""}, Metadata: json.RawMessage(`{}`)})
	return updated, err
}

func (uc *ApprovalsUseCase) Cancel(ctx context.Context, id int64, orgID, userID int32) (repository.ApprovalRequest, error) {
	if uc.repo == nil {
		return repository.ApprovalRequest{}, errors.New("repository is not configured")
	}
	updated, err := uc.repo.CancelApprovalRequest(ctx, repository.CancelApprovalRequestParams{ID: id, OrganizationID: orgID, RequestedBy: pgtype.Int4{Int32: userID, Valid: true}})
	if err != nil {
		return repository.ApprovalRequest{}, errors.New("pending request not found or requester does not own it")
	}
	_, err = uc.repo.CreateApprovalRequestLog(ctx, repository.CreateApprovalRequestLogParams{ApprovalRequestID: id, UserID: pgtype.Int4{Int32: userID, Valid: true}, Action: "cancelled", Metadata: json.RawMessage(`{}`)})
	return updated, err
}
