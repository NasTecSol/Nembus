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

type ProductBarcodeHandler struct {
	useCase *usecase.ProductBarcodeUseCase
}

func NewProductBarcodeHandler(uc *usecase.ProductBarcodeUseCase) *ProductBarcodeHandler {
	return &ProductBarcodeHandler{useCase: uc}
}

func (h *ProductBarcodeHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateProductBarcode handles POST /api/product-barcodes
// @Summary      Create product barcode
// @Description  Create a new barcode for a product
// @Tags         product-barcodes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                       true  "Tenant identifier"
// @Param        Authorization header    string                       true  "Bearer token"
// @Param        body          body      CreateProductBarcodeRequest  true  "Barcode payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/product-barcodes [post]
func (h *ProductBarcodeHandler) CreateProductBarcode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateProductBarcodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	metaBytes, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}

	arg := repository.CreateProductBarcodeParams{
		ProductID:        req.ProductID,
		ProductVariantID: int4Ptr(req.ProductVariantID),
		Barcode:          req.Barcode,
		BarcodeType:      textPtr(req.BarcodeType),
		IsPrimary:        boolPtr(req.IsPrimary),
		Metadata:         metaBytes,
	}

	resp := h.useCase.CreateProductBarcode(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// GetProductBarcode handles GET /api/product-barcodes/:id
// @Summary      Get product barcode
// @Description  Get a specific product barcode by ID
// @Tags         product-barcodes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Barcode ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/product-barcodes/{id} [get]
func (h *ProductBarcodeHandler) GetProductBarcode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid barcode id", nil))
		return
	}

	resp := h.useCase.GetProductBarcode(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// GetProductByBarcode handles GET /api/product-barcodes/lookup/:barcode
// @Summary      Lookup product by barcode
// @Description  Find product information using barcode
// @Tags         product-barcodes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        barcode       path      string  true  "Barcode value"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/product-barcodes/lookup/{barcode} [get]
func (h *ProductBarcodeHandler) GetProductByBarcode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	barcode := c.Param("barcode")
	if barcode == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "barcode is required", nil))
		return
	}

	resp := h.useCase.GetProductByBarcode(c.Request.Context(), barcode)
	c.JSON(resp.StatusCode, resp)
}

// ListProductBarcodes handles GET /api/product-barcodes
// @Summary      List all product barcodes
// @Description  Get a list of all product barcodes
// @Tags         product-barcodes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/product-barcodes [get]
func (h *ProductBarcodeHandler) ListProductBarcodes(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListProductBarcodes(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListProductBarcodesByProduct handles GET /api/products/:product_id/barcodes
// @Summary      List barcodes for a product
// @Description  Get all barcodes associated with a specific product
// @Tags         product-barcodes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        product_id    path      int     true  "Product ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/products/{product_id}/barcodes [get]
func (h *ProductBarcodeHandler) ListProductBarcodesByProduct(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	productIDStr := c.Param("product_id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil))
		return
	}

	resp := h.useCase.ListProductBarcodesByProduct(c.Request.Context(), int32(productID))
	c.JSON(resp.StatusCode, resp)
}

// ListProductBarcodesByVariant handles GET /api/product-variants/:variant_id/barcodes
// @Summary      List barcodes for a product variant
// @Description  Get all barcodes associated with a specific product variant
// @Tags         product-barcodes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        variant_id    path      int     true  "Product Variant ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/product-variants/{variant_id}/barcodes [get]
func (h *ProductBarcodeHandler) ListProductBarcodesByVariant(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	variantIDStr := c.Param("variant_id")
	variantID, err := strconv.Atoi(variantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid variant_id", nil))
		return
	}

	resp := h.useCase.ListProductBarcodesByVariant(c.Request.Context(), int4Ptr(&[]int32{int32(variantID)}[0]))
	c.JSON(resp.StatusCode, resp)
}

// UpdateProductBarcode handles PUT /api/product-barcodes/:id
// @Summary      Update product barcode
// @Description  Update barcode type, primary status, or metadata
// @Tags         product-barcodes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                       true  "Tenant identifier"
// @Param        Authorization header    string                       true  "Bearer token"
// @Param        id            path      int                          true  "Barcode ID"
// @Param        body          body      UpdateProductBarcodeRequest  true  "Update payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/product-barcodes/{id} [put]
func (h *ProductBarcodeHandler) UpdateProductBarcode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid barcode id", nil))
		return
	}

	var req UpdateProductBarcodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	metaBytes, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}

	arg := repository.UpdateProductBarcodeParams{
		ID:          int32(id),
		BarcodeType: textPtr(req.BarcodeType),
		IsPrimary:   boolPtr(req.IsPrimary),
		Metadata:    metaBytes,
	}

	resp := h.useCase.UpdateProductBarcode(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// SetPrimaryBarcode handles PUT /api/products/:product_id/barcodes/primary
// @Summary      Set primary barcode
// @Description  Set a specific barcode as the primary barcode for a product
// @Tags         product-barcodes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                    true  "Tenant identifier"
// @Param        Authorization header    string                    true  "Bearer token"
// @Param        product_id    path      int                       true  "Product ID"
// @Param        body          body      SetPrimaryBarcodeRequest  true  "Primary barcode payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/products/{product_id}/barcodes/primary [put]
func (h *ProductBarcodeHandler) SetPrimaryBarcode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	productIDStr := c.Param("product_id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil))
		return
	}

	var req SetPrimaryBarcodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	arg := repository.SetPrimaryBarcodeParams{
		ProductID: int32(productID),
		ID:        req.BarcodeID,
	}

	resp := h.useCase.SetPrimaryBarcode(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// GetPrimaryBarcode handles GET /api/products/:product_id/barcodes/primary
// @Summary      Get primary barcode
// @Description  Get the primary barcode for a product
// @Tags         product-barcodes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        product_id    path      int     true  "Product ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/products/{product_id}/barcodes/primary [get]
func (h *ProductBarcodeHandler) GetPrimaryBarcode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	productIDStr := c.Param("product_id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil))
		return
	}

	resp := h.useCase.GetPrimaryBarcode(c.Request.Context(), int32(productID))
	c.JSON(resp.StatusCode, resp)
}

// DeleteProductBarcode handles DELETE /api/product-barcodes/:id
// @Summary      Delete product barcode
// @Description  Delete a product barcode
// @Tags         product-barcodes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Barcode ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/product-barcodes/{id} [delete]
func (h *ProductBarcodeHandler) DeleteProductBarcode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid barcode id", nil))
		return
	}

	resp := h.useCase.DeleteProductBarcode(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}
