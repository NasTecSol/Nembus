package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/gin-gonic/gin"
)

// ProductPricingHandler handles product pricing endpoints.
type ProductPricingHandler struct {
	useCase *usecase.ProductPricingUseCase
}

// NewProductPricingHandler creates a new handler instance.
func NewProductPricingHandler(uc *usecase.ProductPricingUseCase) *ProductPricingHandler {
	return &ProductPricingHandler{useCase: uc}
}

func (h *ProductPricingHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateProductPriceRequest represents request body for creating a product price.
type CreateProductPriceRequest struct {
	ProductID        int32                  `json:"product_id" binding:"required"`
	ProductVariantID *int32                 `json:"product_variant_id,omitempty"`
	PriceListID      int32                  `json:"price_list_id" binding:"required"`
	UomID            *int32                 `json:"uom_id,omitempty"`
	Price            float64                `json:"price" binding:"required"` // <-- float64
	MinQuantity      *float64               `json:"min_quantity,omitempty"`
	MaxQuantity      *float64               `json:"max_quantity,omitempty"`
	ValidFrom        *time.Time             `json:"valid_from,omitempty"`
	ValidTo          *time.Time             `json:"valid_to,omitempty"`
	IsActive         *bool                  `json:"is_active,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateProductPriceRequest represents request body for updating a product price.
type UpdateProductPriceRequest struct {
	Price       *string                `json:"price,omitempty"`
	MinQuantity *string                `json:"min_quantity,omitempty"`
	MaxQuantity *string                `json:"max_quantity,omitempty"`
	ValidFrom   *string                `json:"valid_from,omitempty"`
	ValidTo     *string                `json:"valid_to,omitempty"`
	IsActive    *bool                  `json:"is_active,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BulkUpdatePricesRequest represents request body for bulk updating prices.
type BulkUpdatePricesRequest struct {
	PercentageChange string `json:"percentage_change" binding:"required"`
}

// CreateProductPrice handles POST /api/product-prices
// @Summary      Create product price
// @Description  Create a new product price
// @Tags         product_pricing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                    true  "Tenant identifier"
// @Param        Authorization header    string                    true  "Bearer token"
// @Param        body          body      CreateProductPriceRequest true  "Product price payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/product-prices [post]
func (h *ProductPricingHandler) CreateProductPrice(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateProductPriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.CreateProductPrice(
		c.Request.Context(),
		req.ProductID,
		req.ProductVariantID,
		req.PriceListID,
		req.UomID,
		req.Price,       // float64 (NOT &req.Price)
		req.MinQuantity, // *float64
		req.MaxQuantity, // *float64
		req.ValidFrom,   // *time.Time
		req.ValidTo,     // *time.Time
		req.IsActive,
		req.Metadata,
	)
	c.JSON(resp.StatusCode, resp)
}

// GetProductPrice handles GET /api/product-prices/:id
// @Summary      Get product price by ID
// @Description  Retrieve a product price by its ID
// @Tags         product_pricing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Product price ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/product-prices/{id} [get]
func (h *ProductPricingHandler) GetProductPrice(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetProductPrice(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// UpdateProductPrice handles PUT /api/product-prices/:id
// @Summary      Update product price
// @Description  Update an existing product price
// @Tags         product_pricing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                   true  "Tenant identifier"
// @Param        Authorization header    string                   true  "Bearer token"
// @Param        id            path      string                   true  "Product price ID"
// @Param        body          body      UpdateProductPriceRequest true  "Product price payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/product-prices/{id} [put]
func (h *ProductPricingHandler) UpdateProductPrice(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req UpdateProductPriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.UpdateProductPrice(
		c.Request.Context(),
		id,
		req.Price,
		req.MinQuantity,
		req.MaxQuantity,
		req.ValidFrom,
		req.ValidTo,
		req.IsActive,
		req.Metadata,
	)
	c.JSON(resp.StatusCode, resp)
}

// DeleteProductPrice handles DELETE /api/product-prices/:id
// @Summary      Delete product price
// @Description  Delete a product price by ID
// @Tags         product_pricing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Product price ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/product-prices/{id} [delete]
func (h *ProductPricingHandler) DeleteProductPrice(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.DeleteProductPrice(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// ListProductPrices handles GET /api/product-prices/product/:product_id
// @Summary      List product prices
// @Description  List all prices for a specific product
// @Tags         product_pricing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id       header    string  true  "Tenant identifier"
// @Param        Authorization     header    string  true  "Bearer token"
// @Param        product_id        path      string  true  "Product ID"
// @Param        product_variant_id query     string  false "Product variant ID"
// @Param        is_active         query     bool    false "Filter by active status"
// @Success      200               {object}  SuccessResponse
// @Failure      400               {object}  ErrorResponse
// @Failure      401               {object}  ErrorResponse
// @Failure      500               {object}  ErrorResponse
// @Router       /api/product-prices/product/{product_id} [get]
func (h *ProductPricingHandler) ListProductPrices(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	productID := c.Param("product_id")
	productVariantID := c.Query("product_variant_id")
	isActiveStr := c.Query("is_active")

	var isActive *bool
	if isActiveStr == "true" {
		t := true
		isActive = &t
	} else if isActiveStr == "false" {
		f := false
		isActive = &f
	}

	var pvID *string
	if productVariantID != "" {
		pvID = &productVariantID
	}

	resp := h.useCase.ListProductPrices(c.Request.Context(), productID, pvID, isActive)
	c.JSON(resp.StatusCode, resp)
}

// GetEffectivePrice handles GET /api/product-prices/effective
// @Summary      Get effective price
// @Description  Get the effective price for a product based on quantity and price list
// @Tags         product_pricing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id       header    string  true  "Tenant identifier"
// @Param        Authorization     header    string  true  "Bearer token"
// @Param        product_id        query     string  true  "Product ID"
// @Param        price_list_id     query     string  true  "Price list ID"
// @Param        quantity          query     string  true  "Quantity"
// @Param        product_variant_id query     string  false "Product variant ID"
// @Success      200               {object}  SuccessResponse
// @Failure      400               {object}  ErrorResponse
// @Failure      401               {object}  ErrorResponse
// @Failure      404               {object}  ErrorResponse
// @Failure      500               {object}  ErrorResponse
// @Router       /api/product-prices/effective [get]
func (h *ProductPricingHandler) GetEffectivePrice(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	productID := c.Query("product_id")
	priceListID := c.Query("price_list_id")
	quantity := c.Query("quantity")
	productVariantID := c.Query("product_variant_id")

	if productID == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "product_id is required", nil))
		return
	}
	if priceListID == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "price_list_id is required", nil))
		return
	}
	if quantity == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "quantity is required", nil))
		return
	}

	var pvID *string
	if productVariantID != "" {
		pvID = &productVariantID
	}

	resp := h.useCase.GetEffectivePrice(c.Request.Context(), productID, priceListID, quantity, pvID)
	c.JSON(resp.StatusCode, resp)
}

// GetProductPriceForList handles GET /api/product-prices/price-list
// @Summary      Get product price for price list
// @Description  Get product price for a specific price list
// @Tags         product_pricing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id       header    string  true  "Tenant identifier"
// @Param        Authorization     header    string  true  "Bearer token"
// @Param        product_id        query     string  true  "Product ID"
// @Param        price_list_id     query     string  true  "Price list ID"
// @Param        product_variant_id query     string  false "Product variant ID"
// @Param        uom_id            query     string  false "UOM ID"
// @Param        quantity          query     string  false "Quantity"
// @Success      200               {object}  SuccessResponse
// @Failure      400               {object}  ErrorResponse
// @Failure      401               {object}  ErrorResponse
// @Failure      404               {object}  ErrorResponse
// @Failure      500               {object}  ErrorResponse
// @Router       /api/product-prices/price-list [get]
func (h *ProductPricingHandler) GetProductPriceForList(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	productID := c.Query("product_id")
	priceListID := c.Query("price_list_id")
	productVariantID := c.Query("product_variant_id")
	uomID := c.Query("uom_id")
	quantity := c.Query("quantity")

	if productID == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "product_id is required", nil))
		return
	}
	if priceListID == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "price_list_id is required", nil))
		return
	}

	var pvID *string
	if productVariantID != "" {
		pvID = &productVariantID
	}

	var uID *string
	if uomID != "" {
		uID = &uomID
	}

	var qty *string
	if quantity != "" {
		qty = &quantity
	}

	resp := h.useCase.GetProductPriceForList(c.Request.Context(), productID, priceListID, pvID, uID, qty)
	c.JSON(resp.StatusCode, resp)
}

// GetPriceComparison handles GET /api/product-prices/comparison/:product_id
// @Summary      Get price comparison
// @Description  Get price comparison across all price lists for a product
// @Tags         product_pricing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        product_id    path      string  true  "Product ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/product-prices/comparison/{product_id} [get]
func (h *ProductPricingHandler) GetPriceComparison(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	productID := c.Param("product_id")
	resp := h.useCase.GetPriceComparison(c.Request.Context(), productID)
	c.JSON(resp.StatusCode, resp)
}

// ListPricesByPriceList handles GET /api/product-prices/price-list/:price_list_id
// @Summary      List prices by price list
// @Description  List all prices for a specific price list
// @Tags         product_pricing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        price_list_id path      string  true  "Price list ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/product-prices/price-list/{price_list_id} [get]
func (h *ProductPricingHandler) ListPricesByPriceList(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	priceListID := c.Param("price_list_id")
	resp := h.useCase.ListPricesByPriceList(c.Request.Context(), priceListID)
	c.JSON(resp.StatusCode, resp)
}

// GetProductWithPricing handles GET /api/product-prices/product/:product_id/with-pricing
// @Summary      Get product with pricing
// @Description  Get a product with all its pricing information
// @Tags         product_pricing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        product_id    path      string  true  "Product ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/product-prices/product/{product_id}/with-pricing [get]
func (h *ProductPricingHandler) GetProductWithPricing(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	productID := c.Param("product_id")
	resp := h.useCase.GetProductWithPricing(c.Request.Context(), productID)
	c.JSON(resp.StatusCode, resp)
}

// SearchProductsWithPrices handles GET /api/product-prices/search
// @Summary      Search products with prices
// @Description  Search products with prices for a specific price list
// @Tags         product_pricing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        organization_id query     string  true  "Organization ID"
// @Param        price_list_id   query     string  true  "Price list ID"
// @Param        search_term     query     string  true  "Search term"
// @Param        limit           query     int     false "Limit (default: 20)"
// @Param        offset          query     int     false "Offset (default: 0)"
// @Success      200            {object}  SuccessResponse
// @Failure      400            {object}  ErrorResponse
// @Failure      401            {object}  ErrorResponse
// @Failure      500            {object}  ErrorResponse
// @Router       /api/product-prices/search [get]
func (h *ProductPricingHandler) SearchProductsWithPrices(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	organizationID := c.Query("organization_id")
	priceListID := c.Query("price_list_id")
	searchTerm := c.Query("search_term")

	if organizationID == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil))
		return
	}
	if priceListID == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "price_list_id is required", nil))
		return
	}
	if searchTerm == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "search_term is required", nil))
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		limit = 20
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil {
		offset = 0
	}

	resp := h.useCase.SearchProductsWithPrices(
		c.Request.Context(),
		organizationID,
		priceListID,
		searchTerm,
		int32(limit),
		int32(offset),
	)
	c.JSON(resp.StatusCode, resp)
}

// BulkUpdatePrices handles POST /api/product-prices/price-list/:price_list_id/bulk-update
// @Summary      Bulk update prices
// @Description  Update prices in bulk for a price list by percentage
// @Tags         product_pricing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                 true  "Tenant identifier"
// @Param        Authorization header    string                 true  "Bearer token"
// @Param        price_list_id path      string                 true  "Price list ID"
// @Param        body          body      BulkUpdatePricesRequest true  "Bulk update payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/product-prices/price-list/{price_list_id}/bulk-update [post]
func (h *ProductPricingHandler) BulkUpdatePrices(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	priceListID := c.Param("price_list_id")

	var req BulkUpdatePricesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.BulkUpdatePrices(c.Request.Context(), priceListID, req.PercentageChange)
	c.JSON(resp.StatusCode, resp)
}

// ExpirePrices handles POST /api/product-prices/price-list/:price_list_id/expire
// @Summary      Expire prices
// @Description  Expire all active prices for a price list
// @Tags         product_pricing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        price_list_id path      string  true  "Price list ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/product-prices/price-list/{price_list_id}/expire [post]
func (h *ProductPricingHandler) ExpirePrices(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	priceListID := c.Param("price_list_id")
	resp := h.useCase.ExpirePrices(c.Request.Context(), priceListID)
	c.JSON(resp.StatusCode, resp)
}
