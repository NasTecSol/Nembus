package handler

import (
	"net/http"
	"strconv"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"
	"NEMBUS/utils"

	"github.com/gin-gonic/gin"
)

// PosHandler holds the POS use case.
type PosHandler struct {
	useCase *usecase.PosUseCase
}

// NewPosHandler creates a new POS handler.
func NewPosHandler(uc *usecase.PosUseCase) *PosHandler {
	return &PosHandler{useCase: uc}
}

func (h *PosHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// ListProducts handles GET /api/pos/stores/:store_id/products
// @Summary      List POS products for store
// @Description  Returns products with stock for a store (categories, prices, barcode). Optional filters: category_id, search_term, include_out_of_stock.
// @Tags         pos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id           header    string  true   "Tenant identifier"
// @Param        Authorization         header    string  true   "Bearer token"
// @Param        store_id              path      int     true   "Store ID"
// @Param        category_id           query     int     false  "Filter by category ID"
// @Param        search_term           query     string  false  "Filter by name, SKU, or barcode"
// @Param        include_out_of_stock  query     bool    false  "Include out-of-stock products (default false)"
// @Success      200                   {object}  SuccessResponse
// @Failure      400                   {object}  ErrorResponse
// @Failure      401                   {object}  ErrorResponse
// @Failure      404                   {object}  ErrorResponse
// @Failure      500                   {object}  ErrorResponse
// @Router       /api/pos/stores/{store_id}/products [get]
func (h *PosHandler) ListProducts(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeID, err := strconv.ParseInt(c.Param("store_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	var categoryID *int32
	if s := c.Query("category_id"); s != "" {
		id, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid category_id", nil))
			return
		}
		catID := int32(id)
		categoryID = &catID
	}
	var searchTerm *string
	if s := c.Query("search_term"); s != "" {
		searchTerm = &s
	}
	includeOutOfStock := c.Query("include_out_of_stock") == "true" || c.Query("include_out_of_stock") == "1"

	resp := h.useCase.ListProductsForStore(c.Request.Context(), int32(storeID), categoryID, searchTerm, includeOutOfStock)
	c.JSON(resp.StatusCode, resp)
}

// GetProductsByCategory handles GET /api/pos/stores/:store_id/products/category/:category_id
// @Summary      Get POS products by category
// @Description  Returns products in a category (and optionally subcategories) for a store, with stock and pricing.
// @Tags         pos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id            header    string  true   "Tenant identifier"
// @Param        Authorization          header    string  true   "Bearer token"
// @Param        store_id               path      int     true   "Store ID"
// @Param        category_id            path      int     true   "Category ID"
// @Param        include_subcategories  query     bool    false  "Include subcategories (default true)"
// @Success      200                    {object}  SuccessResponse
// @Failure      400                    {object}  ErrorResponse
// @Failure      401                    {object}  ErrorResponse
// @Failure      404                    {object}  ErrorResponse
// @Failure      500                    {object}  ErrorResponse
// @Router       /api/pos/stores/{store_id}/products/category/{category_id} [get]
func (h *PosHandler) GetProductsByCategory(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeID, err := strconv.ParseInt(c.Param("store_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}
	categoryID, err := strconv.ParseInt(c.Param("category_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid category_id", nil))
		return
	}
	includeSubcategories := c.Query("include_subcategories") != "false" && c.Query("include_subcategories") != "0"

	resp := h.useCase.GetProductsByCategory(c.Request.Context(), int32(storeID), int32(categoryID), includeSubcategories)
	c.JSON(resp.StatusCode, resp)
}

// SearchProduct handles GET /api/pos/stores/:store_id/products/search
// @Summary      Search POS product by barcode, ID, or name
// @Description  Searches by barcode (exact), product ID (exact), or name/SKU (fuzzy). Returns single product or list of matches.
// @Tags         pos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        store_id      path      int     true   "Store ID"
// @Param        q             query     string  true   "Search term (barcode, product ID, or name/SKU)"
// @Param        limit         query     int     false  "Max results (default 50)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/stores/{store_id}/products/search [get]
func (h *PosHandler) SearchProduct(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeID, err := strconv.ParseInt(c.Param("store_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "query parameter 'q' required", nil))
		return
	}
	limit := int32(50)
	if s := c.Query("limit"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 32); err == nil && n > 0 {
			limit = int32(n)
		}
	}

	resp := h.useCase.SearchProduct(c.Request.Context(), int32(storeID), q, limit)
	c.JSON(resp.StatusCode, resp)
}
