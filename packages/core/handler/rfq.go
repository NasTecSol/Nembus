package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/middleware/manager"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RFQHandler struct {
	useCase       *usecase.RFQUseCase
	tenantManager *manager.Manager
}

func NewRFQHandler(uc *usecase.RFQUseCase, tm *manager.Manager) *RFQHandler {
	return &RFQHandler{useCase: uc, tenantManager: tm}
}

func (h *RFQHandler) tenantRepo(c *gin.Context) (*repository.Queries, string, bool) {
	slug := strings.TrimSpace(c.Query("tenant"))
	if slug == "" {
		slug = strings.TrimSpace(c.GetHeader("x-tenant-id"))
	}
	if slug == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "tenant query parameter is required", nil))
		return nil, "", false
	}
	pool, err := h.tenantManager.GetPool(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewResponse(utils.CodeNotFound, "tenant not found or inactive", nil))
		return nil, "", false
	}
	return repository.New(pool), slug, true
}

func (h *RFQHandler) authenticatedRepo(c *gin.Context) (*repository.Queries, int32, bool) {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok || repo == nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "repository not found in context", nil))
		return nil, 0, false
	}
	orgID, ok := middleware.GetOrganizationIDFromContext(c)
	if !ok || orgID <= 0 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id is required in authentication context", nil))
		return nil, 0, false
	}
	return repo, orgID, true
}

// List returns the authenticated organization's RFQs.
// @Summary List RFQs
// @Tags RFQs
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param page query int false "Page number"
// @Param limit query int false "Results per page"
// @Param status query string false "RFQ status"
// @Param timeline_status query string false "Timeline status"
// @Param search query string false "Search RFQ number or title"
// @Param has_response query string false "Filter by response presence"
// @Param supplier_type query string false "Filter by supplier type"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/rfqs [get]
func (h *RFQHandler) List(c *gin.Context) {
	repo, orgID, ok := h.authenticatedRepo(c)
	if !ok {
		return
	}
	page, err := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "page must be a positive integer", nil))
		return
	}
	limit, err := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "limit must be a positive integer", nil))
		return
	}
	resp := h.useCase.List(c.Request.Context(), repo, usecase.RFQListFilter{
		OrganizationID: orgID, Status: c.Query("status"), Timeline: c.Query("timeline_status"),
		Search: c.Query("search"), HasResponse: c.Query("has_response"), SupplierType: c.Query("supplier_type"),
		Page: int32(page), Limit: int32(limit),
	})
	c.JSON(resp.StatusCode, resp)
}

// Create creates an RFQ draft and its requested product lines.
// @Summary Create RFQ draft
// @Tags RFQs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param request body usecase.CreateRFQInput true "Example: organization_id must match the authenticated organization; provide vendor_name and at least one product item."
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/rfqs [post]
func (h *RFQHandler) Create(c *gin.Context) {
	repo, orgID, ok := h.authenticatedRepo(c)
	if !ok {
		return
	}
	var in usecase.CreateRFQInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid RFQ payload", nil))
		return
	}
	var userID *int32
	if raw, exists := middleware.GetUserIDFromContext(c); exists {
		if n, err := strconv.ParseInt(raw, 10, 32); err == nil {
			v := int32(n)
			userID = &v
		}
	}
	resp := h.useCase.Create(c.Request.Context(), repo, orgID, userID, in)
	c.JSON(resp.StatusCode, resp)
}

func parseRFQID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid RFQ id", nil))
		return uuid.Nil, false
	}
	return id, true
}

// Get retrieves an RFQ and its requested lines.
// @Summary Get RFQ
// @Tags RFQs
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param id path string true "RFQ UUID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/rfqs/{id} [get]
func (h *RFQHandler) Get(c *gin.Context) {
	repo, orgID, ok := h.authenticatedRepo(c)
	if !ok {
		return
	}
	id, ok := parseRFQID(c)
	if !ok {
		return
	}
	rfq, err := repo.GetRFQByID(c.Request.Context(), id)
	if err != nil || rfq.OrganizationID != orgID {
		c.JSON(http.StatusNotFound, utils.NewResponse(utils.CodeNotFound, "RFQ not found", nil))
		return
	}
	items, err := repo.ListRFQItems(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "failed to load RFQ items", nil))
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(utils.CodeOK, "RFQ retrieved", gin.H{"rfq": rfq, "items": items}))
}

// Publish publishes an RFQ and returns the supplier link.
// @Summary Publish RFQ
// @Tags RFQs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param id path string true "RFQ UUID"
// @Param request body PublishRFQInput true "Publication expiry; example expires_at is 2026-11-30T00:00:00Z."
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/rfqs/{id}/publish [post]
func (h *RFQHandler) Publish(c *gin.Context) {
	repo, orgID, ok := h.authenticatedRepo(c)
	if !ok {
		return
	}
	id, ok := parseRFQID(c)
	if !ok {
		return
	}
	var in PublishRFQInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "expires_at in RFC3339 format is required", nil))
		return
	}
	expiresAt, err := time.Parse(time.RFC3339, in.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "expires_at must use RFC3339 format", nil))
		return
	}
	resp := h.useCase.Publish(c.Request.Context(), repo, orgID, id, expiresAt)
	if row, ok := resp.Data.(repository.Rfq); ok && resp.StatusCode == http.StatusOK {
		slug := c.GetHeader("x-tenant-id")
		if slug == "" {
			slug = c.Query("tenant")
		}
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		resp.Data = gin.H{"rfq": row, "supplier_link": fmt.Sprintf("%s://%s/api/public/rfqs/%s?tenant=%s", scheme, c.Request.Host, url.PathEscape(row.Token), url.QueryEscape(slug))}
	}
	c.JSON(resp.StatusCode, resp)
}

type PublishRFQInput struct {
	ExpiresAt string `json:"expires_at" binding:"required" example:"2026-11-30T00:00:00Z"`
}

// Close stops an RFQ from accepting supplier responses.
// @Summary Close RFQ
// @Tags RFQs
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param id path string true "RFQ UUID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/rfqs/{id}/close [post]
func (h *RFQHandler) Close(c *gin.Context) {
	repo, orgID, ok := h.authenticatedRepo(c)
	if !ok {
		return
	}
	id, ok := parseRFQID(c)
	if !ok {
		return
	}
	resp := h.useCase.Close(c.Request.Context(), repo, orgID, id)
	c.JSON(resp.StatusCode, resp)
}

// Responses lists supplier submissions and their quoted lines for an RFQ.
// @Summary List supplier responses
// @Tags RFQs
// @Produce json
// @Security BearerAuth
// @Param x-tenant-id header string true "Tenant identifier"
// @Param id path string true "RFQ UUID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/rfqs/{id}/responses [get]
func (h *RFQHandler) Responses(c *gin.Context) {
	repo, orgID, ok := h.authenticatedRepo(c)
	if !ok {
		return
	}
	id, ok := parseRFQID(c)
	if !ok {
		return
	}
	rfq, err := repo.GetRFQByID(c.Request.Context(), id)
	if err != nil || rfq.OrganizationID != orgID {
		c.JSON(http.StatusNotFound, utils.NewResponse(utils.CodeNotFound, "RFQ not found", nil))
		return
	}
	responses, err := repo.ListRFQResponses(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "failed to load supplier responses", nil))
		return
	}
	out := make([]gin.H, 0, len(responses))
	for _, response := range responses {
		items, itemErr := repo.ListRFQResponseItems(c.Request.Context(), response.ID)
		if itemErr != nil {
			c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "failed to load response items", nil))
			return
		}
		out = append(out, gin.H{"response": response, "items": items})
	}
	c.JSON(http.StatusOK, utils.NewResponse(utils.CodeOK, "supplier responses retrieved", out))
}

// PublicGet returns the published RFQ details suppliers need to prepare a quote.
// @Summary Get public RFQ details
// @Tags Public RFQs
// @Produce json
// @Param token path string true "Supplier access token"
// @Param tenant query string true "Tenant identifier"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/public/rfqs/{token} [get]
func (h *RFQHandler) PublicGet(c *gin.Context) {
	repo, _, ok := h.tenantRepo(c)
	if !ok {
		return
	}
	rfq, err := repo.GetRFQByToken(c.Request.Context(), c.Param("token"))
	if err != nil || rfq.Status != "published" {
		c.JSON(http.StatusNotFound, utils.NewResponse(utils.CodeNotFound, "RFQ link is invalid or no longer available", nil))
		return
	}
	items, err := repo.ListRFQItems(c.Request.Context(), rfq.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "failed to load RFQ items", nil))
		return
	}
	publicItems := make([]gin.H, 0, len(items))
	for _, item := range items {
		publicItems = append(publicItems, gin.H{"id": item.ID, "product_id": item.ProductID.Int32, "uom_id": item.UomID.Int32, "item_name": item.ItemName, "item_description": item.ItemDescription, "quantity": item.Quantity, "uom": item.UomName, "line_number": item.LineNumber, "notes": item.Notes})
	}
	c.JSON(http.StatusOK, utils.NewResponse(utils.CodeOK, "RFQ retrieved", gin.H{
		"rfq_number": rfq.RfqNumber, "vendor_name": rfq.VendorName, "title": rfq.Title, "description": rfq.Description,
		"timeline_status": rfq.TimelineStatus, "is_expired": rfq.IsTokenExpired,
		"time_remaining_seconds": rfq.TimeRemainingSeconds, "items": publicItems,
	}))
}

// PublicSubmit submits a supplier quotation for every requested RFQ line.
// @Summary Submit supplier quotation
// @Tags Public RFQs
// @Accept json
// @Produce json
// @Param token path string true "Supplier access token"
// @Param tenant query string true "Tenant identifier"
// @Param request body usecase.SubmitRFQResponseInput true "Copy rfq_item_id, product_id, and uom_id from the public RFQ GET response; include one quote for each requested item."
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/public/rfqs/{token}/responses [post]
func (h *RFQHandler) PublicSubmit(c *gin.Context) {
	repo, _, ok := h.tenantRepo(c)
	if !ok {
		return
	}
	rfq, err := repo.GetRFQByToken(c.Request.Context(), c.Param("token"))
	if err != nil || rfq.Status != "published" {
		c.JSON(http.StatusNotFound, utils.NewResponse(utils.CodeNotFound, "RFQ link is invalid or no longer accepting responses", nil))
		return
	}
	var in usecase.SubmitRFQResponseInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid supplier response payload", nil))
		return
	}
	resp := h.useCase.Submit(c.Request.Context(), repo, rfq, in)
	c.JSON(resp.StatusCode, resp)
}
