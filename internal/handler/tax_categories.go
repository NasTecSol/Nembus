package handler

import (
	"net/http"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"
	"NEMBUS/utils"

	"github.com/gin-gonic/gin"
)

// TaxCategoriesHandler handles tax category endpoints.
type TaxCategoriesHandler struct {
	useCase *usecase.TaxCategoriesUseCase
}

// NewTaxCategoriesHandler creates a new handler instance.
func NewTaxCategoriesHandler(uc *usecase.TaxCategoriesUseCase) *TaxCategoriesHandler {
	return &TaxCategoriesHandler{useCase: uc}
}

func (h *TaxCategoriesHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateTaxCategoryRequest represents the request body to create a tax category.
type CreateTaxCategoryRequest struct {
	Name        string                 `json:"name" binding:"required" example:"Standard VAT"`
	Code        string                 `json:"code" binding:"required" example:"VAT_STD"`
	TaxRate     string                 `json:"tax_rate" binding:"required" example:"15.00"` // decimal as string
	IsInclusive bool                   `json:"is_inclusive" example:"false"`
	IsActive    bool                   `json:"is_active" example:"true"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

// UpdateTaxCategoryRequest represents the request body to update a tax category.
type UpdateTaxCategoryRequest struct {
	Name        string                 `json:"name" binding:"required" example:"Standard VAT"`
	TaxRate     string                 `json:"tax_rate" binding:"required" example:"15.00"`
	IsInclusive bool                   `json:"is_inclusive" example:"false"`
	IsActive    bool                   `json:"is_active" example:"true"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

// ToggleTaxCategoryActiveRequest toggles active flag.
type ToggleTaxCategoryActiveRequest struct {
	IsActive bool `json:"is_active" binding:"required" example:"true"`
}

// CreateTaxCategory handles POST /api/tax-categories
// @Summary      Create tax category
// @Description  Create a new tax category
// @Tags         tax_categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                    true  "Tenant identifier"
// @Param        Authorization header    string                    true  "Bearer token"
// @Param        body          body      CreateTaxCategoryRequest  true  "Tax category payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/tax-categories [post]
func (h *TaxCategoriesHandler) CreateTaxCategory(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateTaxCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.CreateTaxCategory(
		c.Request.Context(),
		req.Name,
		req.Code,
		req.TaxRate,
		req.IsInclusive,
		req.IsActive,
		req.Metadata,
	)
	c.JSON(resp.StatusCode, resp)
}

// GetTaxCategory handles GET /api/tax-categories/:id
// @Summary      Get tax category by ID
// @Description  Retrieve a tax category by its ID
// @Tags         tax_categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Tax category ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/tax-categories/{id} [get]
func (h *TaxCategoriesHandler) GetTaxCategory(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetTaxCategory(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// GetTaxCategoryByCode handles GET /api/tax-categories/code/:code
// @Summary      Get tax category by code
// @Description  Retrieve a tax category by its unique code
// @Tags         tax_categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        code          path      string  true  "Tax category code"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/tax-categories/code/{code} [get]
func (h *TaxCategoriesHandler) GetTaxCategoryByCode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	code := c.Param("code")
	resp := h.useCase.GetTaxCategoryByCode(c.Request.Context(), code)
	c.JSON(resp.StatusCode, resp)
}

// ListTaxCategories handles GET /api/tax-categories
// @Summary      List tax categories
// @Description  List all tax categories
// @Tags         tax_categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/tax-categories [get]
func (h *TaxCategoriesHandler) ListTaxCategories(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListTaxCategories(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListActiveTaxCategories handles GET /api/tax-categories/active
// @Summary      List active tax categories
// @Description  List only active tax categories
// @Tags         tax_categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/tax-categories/active [get]
func (h *TaxCategoriesHandler) ListActiveTaxCategories(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListActiveTaxCategories(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// UpdateTaxCategory handles PUT /api/tax-categories/:id
// @Summary      Update tax category
// @Description  Update an existing tax category
// @Tags         tax_categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                    true  "Tenant identifier"
// @Param        Authorization header    string                    true  "Bearer token"
// @Param        id            path      string                    true  "Tax category ID"
// @Param        body          body      UpdateTaxCategoryRequest  true  "Tax category payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/tax-categories/{id} [put]
func (h *TaxCategoriesHandler) UpdateTaxCategory(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req UpdateTaxCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.UpdateTaxCategory(
		c.Request.Context(),
		id,
		req.Name,
		req.TaxRate,
		req.IsInclusive,
		req.IsActive,
		req.Metadata,
	)
	c.JSON(resp.StatusCode, resp)
}

// DeleteTaxCategory handles DELETE /api/tax-categories/:id
// @Summary      Delete tax category
// @Description  Delete a tax category by ID
// @Tags         tax_categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Tax category ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/tax-categories/{id} [delete]
func (h *TaxCategoriesHandler) DeleteTaxCategory(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.DeleteTaxCategory(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// ToggleTaxCategoryActive handles PATCH /api/tax-categories/:id/active
// @Summary      Toggle tax category active flag
// @Description  Update the active flag for a tax category
// @Tags         tax_categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                          true  "Tenant identifier"
// @Param        Authorization header    string                          true  "Bearer token"
// @Param        id            path      string                          true  "Tax category ID"
// @Param        body          body      ToggleTaxCategoryActiveRequest  true  "Active flag payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/tax-categories/{id}/active [patch]
func (h *TaxCategoriesHandler) ToggleTaxCategoryActive(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req ToggleTaxCategoryActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.ToggleTaxCategoryActive(c.Request.Context(), id, req.IsActive)
	c.JSON(resp.StatusCode, resp)
}

