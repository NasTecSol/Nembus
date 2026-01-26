package handler

import (
	"net/http"
	"strconv"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"

	"github.com/gin-gonic/gin"
)

// OrganizationHandler holds the use case
type OrganizationHandler struct {
	useCase *usecase.OrganizationUseCase
}

// NewOrganizationHandler creates a new handler instance
func NewOrganizationHandler(uc *usecase.OrganizationUseCase) *OrganizationHandler {
	return &OrganizationHandler{
		useCase: uc,
	}
}

// getRepositoryFromContext extracts repository from Gin context
func (h *OrganizationHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateOrganization handles POST /organizations
// @Summary      Create a new organization
// @Description  Create a new organization with required name and code
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        organization      body      CreateOrganizationRequest  true  "Organization data"
// @Success      201  {object}  OrganizationResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/organizations [post]
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	// Get repository from context and set it on use case
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return // Error already handled in getRepositoryFromContext
	}
	h.useCase.SetRepository(repo)

	// Bind JSON input
	var req CreateOrganizationRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Call UseCase
	org, err := h.useCase.CreateOrganization(
		c.Request.Context(),
		req.Name,
		req.Code,
		req.LegalName,
		req.TaxID,
		req.CurrencyCode,
		req.FiscalYearVariant,
		req.IsActive,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond with success and organization data
	c.JSON(http.StatusCreated, org)
}

// GetOrganization handles GET /organizations/:id
// @Summary      Get organization by ID
// @Description  Retrieve a specific organization by its ID
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        id            path      string  true  "Organization ID"
// @Success      200  {object}  OrganizationResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/organizations/{id} [get]
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	// Get repository from context and set it on use case
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	org, err := h.useCase.GetOrganization(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, org)
}

// GetOrganizationByCode handles GET /organizations/code/:code
// @Summary      Get organization by code
// @Description  Retrieve a specific organization by its code
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        code            path      string  true  "Organization code"
// @Success      200  {object}  OrganizationResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/organizations/code/{code} [get]
func (h *OrganizationHandler) GetOrganizationByCode(c *gin.Context) {
	// Get repository from context and set it on use case
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	code := c.Param("code")
	org, err := h.useCase.GetOrganizationByCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, org)
}

// ListOrganizations handles GET /organizations
// @Summary      List all organizations
// @Description  Retrieve a list of all organizations for the tenant
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        limit         query     int     false "Limit number of results"
// @Param        offset        query     int     false "Offset for pagination"
// @Param        is_active     query     bool    false "Filter by active status"
// @Success      200  {array}   OrganizationResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/organizations [get]
func (h *OrganizationHandler) ListOrganizations(c *gin.Context) {
	// Get repository from context and set it on use case
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")
	isActiveStr := c.Query("is_active")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		limit = 100
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil {
		offset = 0
	}

	var isActive *bool
	if isActiveStr != "" {
		active, err := strconv.ParseBool(isActiveStr)
		if err == nil {
			isActive = &active
		}
	}

	orgs, err := h.useCase.ListOrganizations(c.Request.Context(), int32(limit), int32(offset), isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orgs)
}

// UpdateOrganization handles PUT /organizations/:id
// @Summary      Update an organization
// @Description  Update an existing organization by its ID
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        id            path      string  true  "Organization ID"
// @Param        organization      body      UpdateOrganizationRequest  true  "Organization data"
// @Success      200  {object}  OrganizationResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/organizations/{id} [put]
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	// Get repository from context and set it on use case
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	// Bind JSON input
	var req UpdateOrganizationRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Call UseCase
	org, err := h.useCase.UpdateOrganization(
		c.Request.Context(),
		id,
		req.Name,
		req.LegalName,
		req.TaxID,
		req.CurrencyCode,
		req.FiscalYearVariant,
		req.IsActive,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, org)
}

// DeleteOrganization handles DELETE /organizations/:id
// @Summary      Delete an organization
// @Description  Delete an organization by its ID
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        id            path      string  true  "Organization ID"
// @Success      200  {object}  SuccessResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/organizations/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	// Get repository from context and set it on use case
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	// Call UseCase
	err := h.useCase.DeleteOrganization(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "organization deleted successfully"})
}
