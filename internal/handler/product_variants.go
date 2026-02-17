package handler

import (
	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"
	"NEMBUS/utils"
	"encoding/json"
	"net/http"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type ProductVariantHandler struct {
	useCase *usecase.ProductVariantUseCase
}

func NewProductVariantHandler(uc *usecase.ProductVariantUseCase) *ProductVariantHandler {
	return &ProductVariantHandler{useCase: uc}
}

func (h *ProductVariantHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "repository not found in context", nil))
		c.Abort()
		return nil
	}
	return repo
}

// ---------------- CREATE ----------------

// CreateProductVariant handles POST /api/product-variants
// @Summary      Create product variant
// @Description  Create a new product variant
// @Tags         product-variants
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                          true  "Tenant identifier"
// @Param        Authorization header    string                          true  "Bearer token"
// @Param        body          body      CreateProductVariantRequest true  "Product variant payload"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/product-variants [post]
func (h *ProductVariantHandler) CreateProductVariant(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateProductVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	meta, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}

	arg := repository.CreateProductVariantParams{
		ProductID:         req.ProductID,
		VariantSku:        req.VariantSku,
		VariantName:       pgtype.Text{String: req.VariantName, Valid: req.VariantName != ""},
		VariantAttributes: json.RawMessage("{}"),
		IsActive:          boolPtr(req.IsActive),
		Metadata:          meta,
	}

	if req.VariantAttributes != nil {
		arg.VariantAttributes, _ = json.Marshal(req.VariantAttributes)
	}

	resp := h.useCase.CreateProductVariant(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// ---------------- UPDATE ----------------

// UpdateProductVariant handles PUT /api/product-variants/variant/:variant_id
// @Summary      Update product variant
// @Description  Update an existing product variant
// @Tags         product-variants
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                          true  "Tenant identifier"
// @Param        Authorization header    string                          true  "Bearer token"
// @Param        id            path      int                             true  "Product variant ID"
// @Param        body          body      CreateProductVariantRequest true  "Product variant payload"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/product-variants/variant/{variant_id} [put]
func (h *ProductVariantHandler) UpdateProductVariant(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	// Get variant ID from path
	idStr := c.Param("variant_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid product variant id", nil))
		return
	}

	// Bind JSON to DTO
	var req CreateProductVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", err.Error()))
		return
	}

	// Marshal VariantAttributes to JSON
	var variantAttributesJSON json.RawMessage = json.RawMessage("{}")
	if req.VariantAttributes != nil {
		b, err := json.Marshal(req.VariantAttributes)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid variant_attributes", nil))
			return
		}
		variantAttributesJSON = b
	}

	// Marshal Metadata to JSON
	var metadataJSON json.RawMessage = json.RawMessage("{}")
	if req.Metadata != nil {
		b, err := json.Marshal(req.Metadata)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
			return
		}
		metadataJSON = b
	}

	// Prepare Update Params
	arg := repository.UpdateProductVariantParams{
		ID:                int32(id),
		VariantName:       pgtype.Text{String: req.VariantName, Valid: req.VariantName != ""},
		VariantAttributes: variantAttributesJSON,
		IsActive:          boolPtr(req.IsActive),
		Metadata:          metadataJSON,
	}

	// Call use case
	resp := h.useCase.UpdateProductVariant(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// ---------------- DELETE ----------------

// DeleteProductVariant handles DELETE /api/product-variants/variant/:variant_id
// @Summary      Delete product variant
// @Description  Delete a product variant by ID
// @Tags         product-variants
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Product variant ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/product-variants/variant/{variant_id} [delete]
func (h *ProductVariantHandler) DeleteProductVariant(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("variant_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid product variant id", nil))
		return
	}

	resp := h.useCase.DeleteProductVariant(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// ---------------- TOGGLE ACTIVE ----------------

// ToggleProductVariantActive handles PATCH /api/product-variants/variant/:variant_id/toggle-active
// @Summary      Toggle product variant active status
// @Description  Enable or disable a product variant
// @Tags         product-variants
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                             true  "Tenant identifier"
// @Param        Authorization header    string                             true  "Bearer token"
// @Param        variant_id    path      int                                true  "Product variant ID"
// @Param        body          body      ToggleProductVariantActiveRequest true "Active status payload"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/product-variants/variant/{variant_id}/toggle-active [patch]
func (h *ProductVariantHandler) ToggleProductVariantActive(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("variant_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid product variant id", nil))
		return
	}

	var req ToggleProductVariantActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	arg := repository.ToggleProductVariantActiveParams{
		ID:       int32(id),
		IsActive: pgtype.Bool{Valid: true, Bool: *req.IsActive},
	}

	resp := h.useCase.ToggleProductVariantActive(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// ---------------- GET BY ID ----------------

// GetProductVariant handles GET /api/product-variants/variant/:variant_id
// @Summary      Get product variant by ID
// @Description  Retrieve a product variant by its ID
// @Tags         product-variants
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Product variant ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/product-variants/variant/{variant_id} [get]
func (h *ProductVariantHandler) GetProductVariant(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("variant_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid product variant id", nil))
		return
	}

	resp := h.useCase.GetProductVariant(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// ---------------- GET BY SKU ----------------

// GetProductVariantBySKU handles GET /api/product-variants/by-sku
// @Summary      Get product variant by SKU
// @Description  Retrieve a product variant by SKU
// @Tags         product-variants
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        sku           query     string  true  "Variant SKU"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/product-variants/by-sku [get]
func (h *ProductVariantHandler) GetProductVariantBySKU(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	sku := c.Query("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "sku query param required", nil))
		return
	}

	resp := h.useCase.GetProductVariantBySKU(c.Request.Context(), sku)
	c.JSON(resp.StatusCode, resp)
}

// ---------------- LIST ALL ----------------

// ListProductVariants handles GET /api/product-variants
// @Summary      List all product variants
// @Description  Retrieve all product variants
// @Tags         product-variants
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200          {object}  SuccessResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/product-variants [get]
func (h *ProductVariantHandler) ListProductVariants(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListProductVariants(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ---------------- LIST BY PRODUCT ----------------

// ListProductVariantsByProduct handles GET /api/product-variants/product/:product_id
// @Summary      List product variants by product
// @Description  Retrieve all variants of a specific product
// @Tags         product-variants
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        product_id    path      int     true  "Product ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/product-variants/product/{product_id} [get]
func (h *ProductVariantHandler) ListProductVariantsByProduct(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("product_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid product id", nil))
		return
	}

	resp := h.useCase.ListProductVariantsByProduct(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// ---------------- LIST ACTIVE BY PRODUCT ----------------

// ListActiveProductVariantsByProduct handles GET /api/product-variants/active/:product_id
// @Summary      List active product variants by product
// @Description  Retrieve only active variants of a specific product
// @Tags         product-variants
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        product_id    path      int     true  "Product ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/product-variants/active/{product_id} [get]
func (h *ProductVariantHandler) ListActiveProductVariantsByProduct(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("product_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid product id", nil))
		return
	}

	resp := h.useCase.ListActiveProductVariantsByProduct(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// ---------------- SEARCH ----------------

// SearchProductVariants handles GET /api/product-variants/search
// @Summary      Search product variants
// @Description  Search product variants by SKU, name, or product name
// @Tags         product-variants
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        sku           query     string  false "Variant SKU"
// @Param        name          query     string  false "Variant Name"
// @Param        product_name  query     string  false "Product Name"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/product-variants/search [get]
func (h *ProductVariantHandler) SearchProductVariants(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	// Get query parameters
	sku := c.Query("sku") // search term for variant SKU, variant name, or product name
	limitStr := c.Query("limit")
	limit := int32(10) // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = int32(l)
		}
	}

	arg := repository.SearchProductVariantsParams{
		VariantSku: "%" + sku + "%", // because your SQL uses ILIKE $1
		Limit:      limit,
	}

	resp := h.useCase.SearchProductVariants(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}
