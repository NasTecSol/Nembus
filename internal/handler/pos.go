package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"
	"NEMBUS/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type POSHandler struct {
	useCase *usecase.POSUseCase
}

func NewPOSHandler(uc *usecase.POSUseCase) *POSHandler {
	return &POSHandler{
		useCase: uc,
	}
}

func (h *POSHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// GetPOSProducts handles GET /api/pos/products
func (h *POSHandler) GetPOSProducts(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}

	storeIDStr := c.Query("store_id")
	storeID, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	var categoryID *int32
	if catIDStr := c.Query("category_id"); catIDStr != "" {
		if id, err := strconv.ParseInt(catIDStr, 10, 32); err == nil {
			id32 := int32(id)
			categoryID = &id32
		}
	}

	searchTerm := c.Query("search")
	includeOutOfStock, _ := strconv.ParseBool(c.DefaultQuery("include_out_of_stock", "false"))

	resp := h.useCase.GetPOSProducts(c.Request.Context(), repo, int32(storeID), categoryID, searchTerm, includeOutOfStock)
	c.JSON(resp.StatusCode, resp)
}

// GetPOSProductByBarcode handles GET /api/pos/products/barcode/:barcode
func (h *POSHandler) GetPOSProductByBarcode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}

	barcode := c.Param("barcode")
	storeIDStr := c.Query("store_id")
	storeID, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	resp := h.useCase.GetPOSProductByBarcode(c.Request.Context(), repo, barcode, int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// GetPOSProductByID handles GET /api/pos/products/:id
func (h *POSHandler) GetPOSProductByID(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid product id", nil))
		return
	}

	storeIDStr := c.Query("store_id")
	storeID, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	resp := h.useCase.GetPOSProductByID(c.Request.Context(), repo, int32(id), int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// SearchPOSProducts handles GET /api/pos/products/search
func (h *POSHandler) SearchPOSProducts(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}

	searchTerm := c.Query("q")
	storeIDStr := c.Query("store_id")
	storeID, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil))
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.ParseInt(limitStr, 10, 32)

	resp := h.useCase.SearchPOSProducts(c.Request.Context(), repo, searchTerm, int32(storeID), int32(limit))
	c.JSON(resp.StatusCode, resp)
}

// GetPOSCategories handles GET /api/pos/categories
func (h *POSHandler) GetPOSCategories(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}

	resp := h.useCase.GetPOSCategories(c.Request.Context(), repo)
	c.JSON(resp.StatusCode, resp)
}

// GetPOSPromotedProducts handles GET /api/pos/promotions
func (h *POSHandler) GetPOSPromotedProducts(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}

	var storeID *int32
	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		if id, err := strconv.ParseInt(storeIDStr, 10, 32); err == nil {
			id32 := int32(id)
			storeID = &id32
		}
	}

	resp := h.useCase.GetPOSPromotedProducts(c.Request.Context(), repo, storeID)
	c.JSON(resp.StatusCode, resp)
}

// CreateProduct handles POST /api/pos/products
func (h *POSHandler) CreateProduct(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}

	var req struct {
		OrganizationID       int32           `json:"organization_id" binding:"required"`
		Sku                  string          `json:"sku" binding:"required"`
		Name                 string          `json:"name" binding:"required"`
		Description          string          `json:"description"`
		CategoryID           int32           `json:"category_id"`
		BrandID              int32           `json:"brand_id"`
		BaseUomID            int32           `json:"base_uom_id"`
		ProductType          string          `json:"product_type"`
		TaxCategoryID        int32           `json:"tax_category_id"`
		IsSerialized         bool            `json:"is_serialized"`
		IsBatchManaged       bool            `json:"is_batch_managed"`
		IsActive             bool            `json:"is_active"`
		IsSellable           bool            `json:"is_sellable"`
		IsPurchasable        bool            `json:"is_purchasable"`
		AllowDecimalQuantity bool            `json:"allow_decimal_quantity"`
		TrackInventory       bool            `json:"track_inventory"`
		Metadata             json.RawMessage `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	params := repository.CreateProductParams{
		OrganizationID:       req.OrganizationID,
		Sku:                  req.Sku,
		Name:                 req.Name,
		Description:          pgtype.Text{String: req.Description, Valid: req.Description != ""},
		CategoryID:           pgtype.Int4{Int32: req.CategoryID, Valid: req.CategoryID != 0},
		BrandID:              pgtype.Int4{Int32: req.BrandID, Valid: req.BrandID != 0},
		BaseUomID:            pgtype.Int4{Int32: req.BaseUomID, Valid: req.BaseUomID != 0},
		ProductType:          pgtype.Text{String: req.ProductType, Valid: req.ProductType != ""},
		TaxCategoryID:        pgtype.Int4{Int32: req.TaxCategoryID, Valid: req.TaxCategoryID != 0},
		IsSerialized:         pgtype.Bool{Bool: req.IsSerialized, Valid: true},
		IsBatchManaged:       pgtype.Bool{Bool: req.IsBatchManaged, Valid: true},
		IsActive:             pgtype.Bool{Bool: req.IsActive, Valid: true},
		IsSellable:           pgtype.Bool{Bool: req.IsSellable, Valid: true},
		IsPurchasable:        pgtype.Bool{Bool: req.IsPurchasable, Valid: true},
		AllowDecimalQuantity: pgtype.Bool{Bool: req.AllowDecimalQuantity, Valid: true},
		TrackInventory:       pgtype.Bool{Bool: req.TrackInventory, Valid: true},
		Metadata:             req.Metadata,
	}

	resp := h.useCase.CreateProduct(c.Request.Context(), repo, params)
	c.JSON(resp.StatusCode, resp)
}
