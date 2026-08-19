package handler

import (
	"net/http"
	"strconv"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/gin-gonic/gin"
)

// ProductCatalogHandler handles admin product catalog requests.
type ProductCatalogHandler struct {
	useCase *usecase.ProductCatalogUseCase
}

// NewProductCatalogHandler creates a new ProductCatalogHandler.
func NewProductCatalogHandler(uc *usecase.ProductCatalogUseCase) *ProductCatalogHandler {
	return &ProductCatalogHandler{useCase: uc}
}

func (h *ProductCatalogHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "repository not found in context", nil))
		c.Abort()
		return nil
	}
	return repo
}

// ListProductsWithVariants handles GET /api/products/catalog
// @Summary      Admin product catalog
// @Description  Returns all master products with their variants embedded as a JSON array. Supports optional category filter and pagination.
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true   "Tenant identifier"
// @Param        Authorization    header    string  true   "Bearer token"
// @Param        organization_id  query     int     true   "Organization ID"
// @Param        category_id      query     int     false  "Filter by category ID (omit or 0 for all)"
// @Param        limit            query     int     false  "Page size (default 20)"
// @Param        offset           query     int     false  "Page offset (default 0)"
// @Success      200  {array}   ListProductsWithVariantsResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/products/catalog [get]
func (h *ProductCatalogHandler) ListProductsWithVariants(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil))
		return
	}

	categoryIDStr := c.DefaultQuery("category_id", "")

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit <= 0 {
		limit = 20
	}
	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil || offset < 0 {
		offset = 0
	}

	resp := h.useCase.ListProductsWithVariants(
		c.Request.Context(),
		orgIDStr,
		categoryIDStr,
		int32(limit),
		int32(offset),
	)
	c.JSON(resp.StatusCode, resp)
}

// GetMasterProductCatalog handles GET /api/products/master-catalog
// @Summary      Master product catalog (detailed)
// @Description  Returns all master products with their base details, UOMs, conversions, pricing, variants, and barcodes nested.
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true   "Tenant identifier"
// @Param        Authorization    header    string  true   "Bearer token"
// @Param        organization_id  query     int     true   "Organization ID"
// @Success      200  {object}  repository.Response
// @Failure      400  {object}  repository.Response
// @Failure      401  {object}  repository.Response
// @Failure      500  {object}  repository.Response
// @Router       /api/products/master-catalog [get]
func (h *ProductCatalogHandler) GetMasterProductCatalog(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil))
		return
	}

	resp := h.useCase.GetMasterProductCatalog(
		c.Request.Context(),
		orgIDStr,
	)
	c.JSON(resp.StatusCode, resp)
}
