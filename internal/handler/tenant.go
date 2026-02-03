package handler

import (
	"net/http"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"
	"NEMBUS/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TenantHandler holds the use case
type TenantHandler struct {
	useCase *usecase.TenantUseCase
}

// NewTenantHandler creates a new handler instance
func NewTenantHandler(uc *usecase.TenantUseCase) *TenantHandler {
	return &TenantHandler{
		useCase: uc,
	}
}

// getRepositoryFromContext extracts repository from Gin context
func (h *TenantHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateTenant handles POST /tenants
// @Summary      Create tenant
// @Description  Create a new tenant
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer token"
// @Param        tenant  body  CreateTenantRequest  true  "Tenant data"
// @Success      201  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/tenants [post]
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req repository.CreateTenantParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"invalid request body",
			nil,
		))
		return
	}

	resp := h.useCase.CreateTenant(c.Request.Context(), req)
	c.JSON(resp.StatusCode, resp)
}

// GetTenantBySlug handles GET /tenants/:slug
// @Summary      Get tenant by slug
// @Description  Get active tenant by slug
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        slug  path  string  true  "Tenant slug"
// @Success      200  {object}  TenantResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/tenants/{slug} [get]
func (h *TenantHandler) GetTenantBySlug(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	slug := c.Param("slug")
	resp := h.useCase.GetTenantBySlug(c.Request.Context(), slug)

	c.JSON(resp.StatusCode, resp)
}

// GetTenantBySlugAny handles GET /tenants/:slug/any
// @Summary      Get tenant by slug (any status)
// @Description  Get tenant by slug regardless of active status
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        slug  path  string  true  "Tenant slug"
// @Success      200  {object}  TenantResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/tenants/{slug}/any [get]
func (h *TenantHandler) GetTenantBySlugAny(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	slug := c.Param("slug")
	resp := h.useCase.GetTenantBySlugAny(c.Request.Context(), slug)

	c.JSON(resp.StatusCode, resp)
}

// ListActiveTenants handles GET /tenants
// @Summary      List active tenants
// @Description  List all active tenants
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   TenantResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/tenants [get]
func (h *TenantHandler) ListActiveTenants(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListActiveTenants(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListAllTenants handles GET /tenants/all
// @Summary      List all tenants
// @Description  Admin endpoint to list all tenants
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   TenantResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/tenants/all [get]
func (h *TenantHandler) ListAllTenants(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListAllTenants(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// UpdateTenant handles PUT /tenants/:id
// @Summary      Update tenant
// @Description  Update tenant fields
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Tenant ID"
// @Param        tenant  body  UpdateTenantRequest  true  "Tenant data"
// @Success      200  {object}  TenantResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/tenants/{id} [put]
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"invalid tenant id",
			nil,
		))
		return
	}

	var req repository.UpdateTenantParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"invalid request body",
			nil,
		))
		return
	}

	resp := h.useCase.UpdateTenant(c.Request.Context(), id, req)
	c.JSON(resp.StatusCode, resp)
}

// DeactivateTenant handles PUT /tenants/:id/deactivate
// @Summary      Deactivate tenant
// @Description  Soft deactivate tenant
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Tenant ID"
// @Success      200  {object}  TenantResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/tenants/{id}/deactivate [put]
func (h *TenantHandler) DeactivateTenant(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"invalid tenant id",
			nil,
		))
		return
	}

	resp := h.useCase.DeactivateTenant(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}
