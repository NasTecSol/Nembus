package handler

import (
	"net/http"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/gin-gonic/gin"
)

// PriceListsHandler handles price list endpoints.
type PriceListsHandler struct {
	useCase *usecase.PriceListsUseCase
}

// NewPriceListsHandler creates a new handler instance.
func NewPriceListsHandler(uc *usecase.PriceListsUseCase) *PriceListsHandler {
	return &PriceListsHandler{useCase: uc}
}

func (h *PriceListsHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreatePriceListRequest represents request body for creating a price list.
type CreatePriceListRequest struct {
	Name          string                 `json:"name" binding:"required" example:"Default Price List"`
	Code          string                 `json:"code" binding:"required" example:"DEFAULT"`
	PriceListType *string                `json:"price_list_type,omitempty" example:"retail"`
	CurrencyCode  *string                `json:"currency_code,omitempty" example:"USD"`
	ValidFrom     *string                `json:"valid_from,omitempty" example:"2026-01-01"`
	ValidTo       *string                `json:"valid_to,omitempty" example:"2026-12-31"`
	IsDefault     bool                   `json:"is_default" example:"true"`
	IsActive      bool                   `json:"is_active" example:"true"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

// UpdatePriceListRequest represents request body for updating a price list.
type UpdatePriceListRequest struct {
	Name          string                 `json:"name" binding:"required" example:"Updated Price List"`
	PriceListType *string                `json:"price_list_type,omitempty" example:"retail"`
	CurrencyCode  *string                `json:"currency_code,omitempty" example:"USD"`
	ValidFrom     *string                `json:"valid_from,omitempty" example:"2026-01-01"`
	ValidTo       *string                `json:"valid_to,omitempty" example:"2026-12-31"`
	IsDefault     bool                   `json:"is_default" example:"true"`
	IsActive      bool                   `json:"is_active" example:"true"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" swaggertype:"object"`
}

// TogglePriceListActiveRequest is used to toggle active flag.
type TogglePriceListActiveRequest struct {
	IsActive bool `json:"is_active" binding:"required" example:"true"`
}

// CreatePriceList handles POST /api/price-lists
// @Summary      Create price list
// @Description  Create a new price list
// @Tags         price_lists
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                  true  "Tenant identifier"
// @Param        Authorization header    string                  true  "Bearer token"
// @Param        body          body      CreatePriceListRequest  true  "Price list payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/price-lists [post]
func (h *PriceListsHandler) CreatePriceList(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreatePriceListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.CreatePriceList(
		c.Request.Context(),
		req.Name,
		req.Code,
		req.PriceListType,
		req.CurrencyCode,
		req.ValidFrom,
		req.ValidTo,
		req.IsDefault,
		req.IsActive,
		req.Metadata,
	)
	c.JSON(resp.StatusCode, resp)
}

// GetPriceList handles GET /api/price-lists/:id
// @Summary      Get price list by ID
// @Description  Retrieve a price list by its ID
// @Tags         price_lists
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Price list ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/price-lists/{id} [get]
func (h *PriceListsHandler) GetPriceList(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetPriceList(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// GetPriceListByCode handles GET /api/price-lists/code/:code
// @Summary      Get price list by code
// @Description  Retrieve a price list by its unique code
// @Tags         price_lists
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        code          path      string  true  "Price list code"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/price-lists/code/{code} [get]
func (h *PriceListsHandler) GetPriceListByCode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	code := c.Param("code")
	resp := h.useCase.GetPriceListByCode(c.Request.Context(), code)
	c.JSON(resp.StatusCode, resp)
}

// GetDefaultPriceList handles GET /api/price-lists/default
// @Summary      Get default price list
// @Description  Retrieve the default active price list
// @Tags         price_lists
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/price-lists/default [get]
func (h *PriceListsHandler) GetDefaultPriceList(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.GetDefaultPriceList(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListPriceLists handles GET /api/price-lists
// @Summary      List price lists
// @Description  List all price lists
// @Tags         price_lists
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/price-lists [get]
func (h *PriceListsHandler) ListPriceLists(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListPriceLists(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListActivePriceLists handles GET /api/price-lists/active
// @Summary      List active price lists
// @Description  List only active price lists
// @Tags         price_lists
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/price-lists/active [get]
func (h *PriceListsHandler) ListActivePriceLists(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListActivePriceLists(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListValidPriceLists handles GET /api/price-lists/valid
// @Summary      List valid price lists
// @Description  List active price lists valid for current date
// @Tags         price_lists
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/price-lists/valid [get]
func (h *PriceListsHandler) ListValidPriceLists(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListValidPriceLists(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// UpdatePriceList handles PUT /api/price-lists/:id
// @Summary      Update price list
// @Description  Update an existing price list
// @Tags         price_lists
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                 true  "Tenant identifier"
// @Param        Authorization header    string                 true  "Bearer token"
// @Param        id            path      string                 true  "Price list ID"
// @Param        body          body      UpdatePriceListRequest true  "Price list payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/price-lists/{id} [put]
func (h *PriceListsHandler) UpdatePriceList(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req UpdatePriceListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.UpdatePriceList(
		c.Request.Context(),
		id,
		req.Name,
		req.PriceListType,
		req.CurrencyCode,
		req.ValidFrom,
		req.ValidTo,
		req.IsDefault,
		req.IsActive,
		req.Metadata,
	)
	c.JSON(resp.StatusCode, resp)
}

// DeletePriceList handles DELETE /api/price-lists/:id
// @Summary      Delete price list
// @Description  Delete a price list by ID
// @Tags         price_lists
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Price list ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/price-lists/{id} [delete]
func (h *PriceListsHandler) DeletePriceList(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.DeletePriceList(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// SetDefaultPriceList handles POST /api/price-lists/{id}/set-default
// @Summary      Set default price list
// @Description  Mark the specified price list as default
// @Tags         price_lists
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Price list ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/price-lists/{id}/set-default [post]
func (h *PriceListsHandler) SetDefaultPriceList(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.SetDefaultPriceList(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// TogglePriceListActive handles PATCH /api/price-lists/{id}/active
// @Summary      Toggle price list active flag
// @Description  Update the active flag for a price list
// @Tags         price_lists
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                     true  "Tenant identifier"
// @Param        Authorization header    string                     true  "Bearer token"
// @Param        id            path      string                     true  "Price list ID"
// @Param        body          body      TogglePriceListActiveRequest true  "Active flag payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/price-lists/{id}/active [patch]
func (h *PriceListsHandler) TogglePriceListActive(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req TogglePriceListActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.TogglePriceListActive(c.Request.Context(), id, req.IsActive)
	c.JSON(resp.StatusCode, resp)
}
