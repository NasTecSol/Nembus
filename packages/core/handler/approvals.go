package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/gin-gonic/gin"
)

type ApprovalsHandler struct{ uc *usecase.ApprovalsUseCase }

type ApprovalPolicyRequest struct {
	Action     string          `json:"action" example:"post"`
	Enabled    bool            `json:"is_enabled" example:"true"`
	Requires   bool            `json:"requires_approval" example:"true"`
	Permission string          `json:"required_permission" example:"goods_receipt_notes:approve"`
	Rules      json.RawMessage `json:"rules,omitempty" swaggertype:"object"`
}

type ApprovalDecisionRequest struct {
	Notes string `json:"notes,omitempty" example:"Reviewed and approved"`
}

func NewApprovalsHandler(uc *usecase.ApprovalsUseCase) *ApprovalsHandler {
	return &ApprovalsHandler{uc: uc}
}

func (h *ApprovalsHandler) setup(c *gin.Context) (int32, int32, bool) {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found"})
		return 0, 0, false
	}
	org, ok := middleware.GetOrganizationIDFromContext(c)
	if !ok || org <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "organization claim missing"})
		return 0, 0, false
	}
	uid, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user claim missing"})
		return 0, 0, false
	}
	n, err := strconv.ParseInt(uid, 10, 32)
	if err != nil || n <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user claim"})
		return 0, 0, false
	}
	h.uc.SetRepository(repo)
	return org, int32(n), true
}

// Submit creates an approval request for a domain action.
// @Summary Submit an approval request
// @Tags approvals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param body body usecase.ApprovalInput true "Approval request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /api/approvals/submit [post]
func (h *ApprovalsHandler) Submit(c *gin.Context) {
	org, user, ok := h.setup(c)
	if !ok {
		return
	}
	var in usecase.ApprovalInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if in.OrganizationID != org {
		c.JSON(http.StatusForbidden, gin.H{"error": "organization does not match token"})
		return
	}
	in.RequestedBy = user
	req, err := h.uc.Submit(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
}
// List returns approval requests for the current organization.
// @Summary List approval requests
// @Tags approvals
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param domain query string false "Filter by domain"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /api/approvals [get]
func (h *ApprovalsHandler) List(c *gin.Context) {
	org, _, ok := h.setup(c)
	if !ok {
		return
	}
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 32)
	var domain, status *string
	if v := c.Query("domain"); v != "" {
		domain = &v
	}
	if v := c.Query("status"); v != "" {
		status = &v
	}
	items, err := h.uc.List(c.Request.Context(), org, domain, status, int32(page), int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}
func (h *ApprovalsHandler) Decide(c *gin.Context, approve bool) {
	org, user, ok := h.setup(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid approval id"})
		return
	}
	var body struct {
		Notes string `json:"notes"`
	}
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	req, err := h.uc.Decide(c.Request.Context(), usecase.ApprovalDecision{RequestID: id, OrganizationID: org, UserID: user, Notes: body.Notes}, approve)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, req)
}
// Approve approves and applies a pending request.
// @Summary Approve an approval request
// @Tags approvals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param id path int true "Approval request ID"
// @Param body body ApprovalDecisionRequest false "Decision notes"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /api/approvals/{id}/approve [post]
func (h *ApprovalsHandler) Approve(c *gin.Context) { h.Decide(c, true) }

// Reject rejects a pending request.
// @Summary Reject an approval request
// @Tags approvals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param id path int true "Approval request ID"
// @Param body body ApprovalDecisionRequest false "Decision notes"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /api/approvals/{id}/reject [post]
func (h *ApprovalsHandler) Reject(c *gin.Context) { h.Decide(c, false) }

// SavePolicy creates or updates a policy for a domain.
// @Summary Save an approval policy
// @Tags approvals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param domain path string true "Approval domain"
// @Param body body ApprovalPolicyRequest true "Policy configuration"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/approvals/policies/{domain} [put]
func (h *ApprovalsHandler) SavePolicy(c *gin.Context) {
	org, user, ok := h.setup(c)
	if !ok {
		return
	}
	repo := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	allowed, authErr := repo.CheckUserHasPermissionCode(c.Request.Context(), repository.CheckUserHasPermissionCodeParams{UserID: user, Code: "approvals:manage"})
	if authErr != nil || !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "approvals:manage permission required"})
		return
	}
	var in ApprovalPolicyRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if in.Action == "" || in.Permission == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action and required_permission are required"})
		return
	}
	if len(in.Rules) == 0 {
		in.Rules = json.RawMessage(`{}`)
	}
	if !json.Valid(in.Rules) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rules must be valid JSON"})
		return
	}
	p, err := repo.UpsertApprovalPolicy(c.Request.Context(), repository.UpsertApprovalPolicyParams{OrganizationID: org, Domain: c.Param("domain"), Action: in.Action, IsEnabled: in.Enabled, RequiresApproval: in.Requires, RequiredPermission: in.Permission, Rules: in.Rules})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// GetPolicy returns the policy for a domain and action.
// @Summary Get an approval policy
// @Tags approvals
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param domain path string true "Approval domain"
// @Param action query string false "Action" default(submit)
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} ErrorResponse
// @Router /api/approvals/policies/{domain} [get]
func (h *ApprovalsHandler) GetPolicy(c *gin.Context) {
	org, _, ok := h.setup(c)
	if !ok {
		return
	}
	repo := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	action := c.DefaultQuery("action", "submit")
	policy, err := repo.GetApprovalPolicy(c.Request.Context(), repository.GetApprovalPolicyParams{
		OrganizationID: org,
		Domain:         c.Param("domain"),
		Action:         action,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "approval policy not found"})
		return
	}
	c.JSON(http.StatusOK, policy)
}

// DeletePolicy removes a policy for a domain and action.
// @Summary Delete an approval policy
// @Tags approvals
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param domain path string true "Approval domain"
// @Param action query string false "Action" default(submit)
// @Success 204
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/approvals/policies/{domain} [delete]
func (h *ApprovalsHandler) DeletePolicy(c *gin.Context) {
	org, user, ok := h.setup(c)
	if !ok {
		return
	}
	repo := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	allowed, err := repo.CheckUserHasPermissionCode(c.Request.Context(), repository.CheckUserHasPermissionCodeParams{UserID: user, Code: "approvals:manage"})
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "approvals:manage permission required"})
		return
	}
	affected, err := repo.DeleteApprovalPolicy(c.Request.Context(), repository.DeleteApprovalPolicyParams{
		OrganizationID: org,
		Domain:         c.Param("domain"),
		Action:         c.DefaultQuery("action", "submit"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "approval policy not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

// Get returns one approval request.
// @Summary Get an approval request
// @Tags approvals
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param id path int true "Approval request ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} ErrorResponse
// @Router /api/approvals/{id} [get]
func (h *ApprovalsHandler) Get(c *gin.Context) {
	org, _, ok := h.setup(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid approval id"})
		return
	}
	repo := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	request, err := repo.GetApprovalRequest(c.Request.Context(), repository.GetApprovalRequestParams{ID: id, OrganizationID: org})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "approval request not found"})
		return
	}
	c.JSON(http.StatusOK, request)
}

// Delete removes an approval request.
// @Summary Delete an approval request
// @Tags approvals
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param id path int true "Approval request ID"
// @Success 204
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/approvals/{id} [delete]
func (h *ApprovalsHandler) Delete(c *gin.Context) {
	org, user, ok := h.setup(c)
	if !ok {
		return
	}
	repo := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	allowed, err := repo.CheckUserHasPermissionCode(c.Request.Context(), repository.CheckUserHasPermissionCodeParams{UserID: user, Code: "approvals:manage"})
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "approvals:manage permission required"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid approval id"})
		return
	}
	affected, err := repo.DeleteApprovalRequest(c.Request.Context(), repository.DeleteApprovalRequestParams{OrganizationID: org, ID: id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "approval request not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

// Policies lists approval policies for the current organization.
// @Summary List approval policies
// @Tags approvals
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /api/approvals/policies [get]
func (h *ApprovalsHandler) Policies(c *gin.Context) {
	org, _, ok := h.setup(c)
	if !ok {
		return
	}
	repo := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	items, err := repo.ListApprovalPolicies(c.Request.Context(), org)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// Cancel cancels a pending request owned by the current user.
// @Summary Cancel an approval request
// @Tags approvals
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param id path int true "Approval request ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /api/approvals/{id}/cancel [post]
func (h *ApprovalsHandler) Cancel(c *gin.Context) {
	org, user, ok := h.setup(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid approval id"})
		return
	}
	updated, err := h.uc.Cancel(c.Request.Context(), id, org, user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}
