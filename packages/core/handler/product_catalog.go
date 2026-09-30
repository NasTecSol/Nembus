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
// @Description  Returns all master products with their base details, UOMs, conversions, pricing, variants, and barcodes nested. Supports optional category filter and search filter.
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true   "Tenant identifier"
// @Param        Authorization    header    string  true   "Bearer token"
// @Param        organization_id  query     int     true   "Organization ID"
// @Param        product_id       query     int     false  "Optional filter by product ID"
// @Param        category_id      query     int     false  "Optional filter by category ID"
// @Param        q                query     string  false  "Optional search query (SKU, product name, description, brand, category, barcode, variant)"
// @Param        search           query     string  false  "Alias for search query"
// @Param        limit            query     int     false  "Page size (default 100)"
// @Param        offset           query     int     false  "Page offset (default 0)"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
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

	productIDStr := c.Query("product_id")
	if productIDStr == "" {
		productIDStr = c.Query("id")
	}
	if productIDStr != "" {
		resp := h.useCase.GetMasterProductCatalogByID(
			c.Request.Context(),
			orgIDStr,
			productIDStr,
		)
		c.JSON(resp.StatusCode, resp)
		return
	}

	searchQuery := c.Query("q")
	if searchQuery == "" {
		searchQuery = c.Query("query")
	}
	if searchQuery == "" {
		searchQuery = c.Query("search")
	}
	if searchQuery == "" {
		searchQuery = c.Query("search_term")
	}

	categoryIDStr := c.Query("category_id")

	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit <= 0 {
		limit = 100
	}
	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil || offset < 0 {
		offset = 0
	}

	if searchQuery != "" {
		resp := h.useCase.SearchMasterProductCatalog(
			c.Request.Context(),
			orgIDStr,
			searchQuery,
			int32(limit),
			int32(offset),
		)
		c.JSON(resp.StatusCode, resp)
		return
	}

	if categoryIDStr != "" && categoryIDStr != "0" {
		resp := h.useCase.GetMasterProductCatalogByCategory(
			c.Request.Context(),
			orgIDStr,
			categoryIDStr,
			int32(limit),
			int32(offset),
		)
		c.JSON(resp.StatusCode, resp)
		return
	}

	resp := h.useCase.GetMasterProductCatalog(
		c.Request.Context(),
		orgIDStr,
		int32(limit),
		int32(offset),
	)
	c.JSON(resp.StatusCode, resp)
}

// GetMasterProductCatalogByCategory handles GET /api/products/master-catalog/category/:categoryID
// @Summary      Master product catalog by category
// @Description  Returns master product catalog with full details (variants, pricing, barcodes, UOMs, conversions, inventory) filtered by a specific category with pagination.
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true   "Tenant identifier"
// @Param        Authorization    header    string  true   "Bearer token"
// @Param        categoryID       path      int     true   "Category ID"
// @Param        organization_id  query     int     true   "Organization ID"
// @Param        limit            query     int     false  "Page size (default 100)"
// @Param        offset           query     int     false  "Page offset (default 0)"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/products/master-catalog/category/{categoryID} [get]
func (h *ProductCatalogHandler) GetMasterProductCatalogByCategory(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	categoryIDStr := c.Param("categoryID")
	if categoryIDStr == "" {
		categoryIDStr = c.Param("id")
	}

	orgIDStr := c.Query("organization_id")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil))
		return
	}

	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit <= 0 {
		limit = 100
	}
	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil || offset < 0 {
		offset = 0
	}

	resp := h.useCase.GetMasterProductCatalogByCategory(
		c.Request.Context(),
		orgIDStr,
		categoryIDStr,
		int32(limit),
		int32(offset),
	)
	c.JSON(resp.StatusCode, resp)
}

// SearchMasterProductCatalog handles GET /api/products/master-catalog/search
// @Summary      Search master product catalog
// @Description  Searches master product catalog by SKU, product name, description, brand, category, variant, or barcode with nested details (UOMs, conversions, pricing, variants, barcodes, inventory) and pagination.
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true   "Tenant identifier"
// @Param        Authorization    header    string  true   "Bearer token"
// @Param        organization_id  query     int     true   "Organization ID"
// @Param        q                query     string  false  "Search query (SKU, product name, description, brand, category, barcode, variant)"
// @Param        query            query     string  false  "Alias for search query"
// @Param        search           query     string  false  "Alias for search query"
// @Param        search_term      query     string  false  "Alias for search query"
// @Param        limit            query     int     false  "Page size (default 100)"
// @Param        offset           query     int     false  "Page offset (default 0)"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/products/master-catalog/search [get]
func (h *ProductCatalogHandler) SearchMasterProductCatalog(c *gin.Context) {
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

	searchQuery := c.Query("q")
	if searchQuery == "" {
		searchQuery = c.Query("query")
	}
	if searchQuery == "" {
		searchQuery = c.Query("search")
	}
	if searchQuery == "" {
		searchQuery = c.Query("search_term")
	}

	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit <= 0 {
		limit = 100
	}
	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil || offset < 0 {
		offset = 0
	}

	resp := h.useCase.SearchMasterProductCatalog(
		c.Request.Context(),
		orgIDStr,
		searchQuery,
		int32(limit),
		int32(offset),
	)
	c.JSON(resp.StatusCode, resp)
}

// GetMasterProductCatalogByID handles GET /api/products/master-catalog/product/:productID
// @Summary      Master product catalog by product ID
// @Description  Returns master product catalog with full details (variants, pricing, barcodes, UOMs, conversions, inventory) for a single product.
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true   "Tenant identifier"
// @Param        Authorization    header    string  true   "Bearer token"
// @Param        productID        path      int     true   "Product ID"
// @Param        organization_id  query     int     true   "Organization ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/products/master-catalog/product/{productID} [get]
func (h *ProductCatalogHandler) GetMasterProductCatalogByID(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	productIDStr := c.Param("productID")
	if productIDStr == "" {
		productIDStr = c.Param("product_id")
	}
	if productIDStr == "" {
		productIDStr = c.Param("id")
	}

	orgIDStr := c.Query("organization_id")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil))
		return
	}

	resp := h.useCase.GetMasterProductCatalogByID(
		c.Request.Context(),
		orgIDStr,
		productIDStr,
	)
	c.JSON(resp.StatusCode, resp)
}


