package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/gin-gonic/gin"
)

// InventoryStockHandler handles inventory stock endpoints.
type InventoryStockHandler struct {
	useCase *usecase.InventoryStockUseCase
}

// NewInventoryStockHandler creates a new handler instance.
func NewInventoryStockHandler(uc *usecase.InventoryStockUseCase) *InventoryStockHandler {
	return &InventoryStockHandler{useCase: uc}
}

func (h *InventoryStockHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateInventoryStockRequest represents request body for creating inventory stock.
type CreateInventoryStockRequest struct {
	ProductID         int32                  `json:"product_id" binding:"required"`
	ProductVariantID  *int32                 `json:"product_variant_id,omitempty"`
	StoreID           int32                  `json:"store_id" binding:"required"`
	StorageLocationID *int32                 `json:"storage_location_id,omitempty"`
	QuantityOnHand    *string                `json:"quantity_on_hand,omitempty"`
	QuantityAllocated *string                `json:"quantity_allocated,omitempty"`
	QuantityAvailable *string                `json:"quantity_available,omitempty"`
	QuantityOnOrder   *string                `json:"quantity_on_order,omitempty"`
	QuantityInTransit *string                `json:"quantity_in_transit,omitempty"`
	ReorderLevel      *string                `json:"reorder_level,omitempty"`
	ReorderQuantity   *string                `json:"reorder_quantity,omitempty"`
	MaxStockLevel     *string                `json:"max_stock_level,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateInventoryStockRequest represents request body for updating inventory stock.
type UpdateInventoryStockRequest struct {
	QuantityOnHand    *string                `json:"quantity_on_hand,omitempty"`
	QuantityAllocated *string                `json:"quantity_allocated,omitempty"`
	QuantityAvailable *string                `json:"quantity_available,omitempty"`
	QuantityOnOrder   *string                `json:"quantity_on_order,omitempty"`
	QuantityInTransit *string                `json:"quantity_in_transit,omitempty"`
	ReorderLevel      *string                `json:"reorder_level,omitempty"`
	ReorderQuantity   *string                `json:"reorder_quantity,omitempty"`
	MaxStockLevel     *string                `json:"max_stock_level,omitempty"`
	LastCountedAt     *string                `json:"last_counted_at,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// AdjustInventoryStockRequest represents request body for adjusting inventory stock.
type AdjustInventoryStockRequest struct {
	QuantityOnHandDelta    *string `json:"quantity_on_hand_delta,omitempty"`
	QuantityAvailableDelta *string `json:"quantity_available_delta,omitempty"`
	QuantityAllocatedDelta *string `json:"quantity_allocated_delta,omitempty"`
	QuantityOnOrderDelta   *string `json:"quantity_on_order_delta,omitempty"`
	QuantityInTransitDelta *string `json:"quantity_in_transit_delta,omitempty"`
}

// UpsertInventoryStockRequest represents request body for upserting inventory stock.
type UpsertInventoryStockRequest struct {
	ProductID         int32                  `json:"product_id" binding:"required"`
	ProductVariantID  *int32                 `json:"product_variant_id,omitempty"`
	StoreID           int32                  `json:"store_id" binding:"required"`
	StorageLocationID *int32                 `json:"storage_location_id,omitempty"`
	QuantityOnHand    *string                `json:"quantity_on_hand,omitempty"`
	QuantityAllocated *string                `json:"quantity_allocated,omitempty"`
	QuantityAvailable *string                `json:"quantity_available,omitempty"`
	QuantityOnOrder   *string                `json:"quantity_on_order,omitempty"`
	QuantityInTransit *string                `json:"quantity_in_transit,omitempty"`
	ReorderLevel      *string                `json:"reorder_level,omitempty"`
	ReorderQuantity   *string                `json:"reorder_quantity,omitempty"`
	MaxStockLevel     *string                `json:"max_stock_level,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// CreateInventoryStock handles POST /api/inventory-stock
// @Summary      Create inventory stock
// @Description  Create a new inventory stock record
// @Tags         inventory_stock
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                      true  "Tenant identifier"
// @Param        Authorization header    string                      true  "Bearer token"
// @Param        body          body      CreateInventoryStockRequest true  "Inventory stock payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/inventory-stock [post]
func (h *InventoryStockHandler) CreateInventoryStock(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateInventoryStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.CreateInventoryStock(
		c.Request.Context(),
		req.ProductID,
		req.ProductVariantID,
		req.StoreID,
		req.StorageLocationID,
		req.QuantityOnHand,
		req.QuantityAllocated,
		req.QuantityAvailable,
		req.QuantityOnOrder,
		req.QuantityInTransit,
		req.ReorderLevel,
		req.ReorderQuantity,
		req.MaxStockLevel,
		req.Metadata,
	)
	c.JSON(resp.StatusCode, resp)
}

// GetInventoryStock handles GET /api/inventory-stock/:id
// @Summary      Get inventory stock by ID
// @Description  Retrieve inventory stock by its ID
// @Tags         inventory_stock
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Inventory stock ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/inventory-stock/{id} [get]
func (h *InventoryStockHandler) GetInventoryStock(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetInventoryStock(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// GetInventoryStockByProductAndStore handles GET /api/inventory-stock/product-store
// @Summary      Get inventory stock by product and store
// @Description  Retrieve inventory stock by product, variant, store, and location
// @Tags         inventory_stock
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id        header    string  true  "Tenant identifier"
// @Param        Authorization      header    string  true  "Bearer token"
// @Param        product_id         query     string  true  "Product ID"
// @Param        product_variant_id query     string  false "Product variant ID"
// @Param        store_id           query     string  true  "Store ID"
// @Param        storage_location_id query    string  false "Storage location ID"
// @Success      200                {object}  SuccessResponse
// @Failure      400                {object}  ErrorResponse
// @Failure      401                {object}  ErrorResponse
// @Failure      404                {object}  ErrorResponse
// @Failure      500                {object}  ErrorResponse
// @Router       /api/inventory-stock/product-store [get]
func (h *InventoryStockHandler) GetInventoryStockByProductAndStore(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	productID := c.Query("product_id")
	productVariantID := c.Query("product_variant_id")
	storeID := c.Query("store_id")
	storageLocationID := c.Query("storage_location_id")

	if productID == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "product_id is required", nil))
		return
	}
	if storeID == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "store_id is required", nil))
		return
	}

	var pvID *string
	if productVariantID != "" {
		pvID = &productVariantID
	}

	var slID *string
	if storageLocationID != "" {
		slID = &storageLocationID
	}

	resp := h.useCase.GetInventoryStockByProductAndStore(c.Request.Context(), productID, pvID, storeID, slID)
	c.JSON(resp.StatusCode, resp)
}

// UpdateInventoryStock handles PUT /api/inventory-stock/:id
// @Summary      Update inventory stock
// @Description  Update an existing inventory stock record
// @Tags         inventory_stock
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                     true  "Tenant identifier"
// @Param        Authorization header    string                     true  "Bearer token"
// @Param        id            path      string                     true  "Inventory stock ID"
// @Param        body          body      UpdateInventoryStockRequest true  "Inventory stock payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/inventory-stock/{id} [put]
func (h *InventoryStockHandler) UpdateInventoryStock(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req UpdateInventoryStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.UpdateInventoryStock(
		c.Request.Context(),
		id,
		req.QuantityOnHand,
		req.QuantityAllocated,
		req.QuantityAvailable,
		req.QuantityOnOrder,
		req.QuantityInTransit,
		req.ReorderLevel,
		req.ReorderQuantity,
		req.MaxStockLevel,
		req.LastCountedAt,
		req.Metadata,
	)
	c.JSON(resp.StatusCode, resp)
}

// UpsertInventoryStock handles POST /api/inventory-stock/upsert
// @Summary      Upsert inventory stock
// @Description  Create or update inventory stock record
// @Tags         inventory_stock
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                      true  "Tenant identifier"
// @Param        Authorization header    string                      true  "Bearer token"
// @Param        body          body      UpsertInventoryStockRequest true  "Inventory stock payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/inventory-stock/upsert [post]
func (h *InventoryStockHandler) UpsertInventoryStock(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req UpsertInventoryStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.UpsertInventoryStock(
		c.Request.Context(),
		req.ProductID,
		req.ProductVariantID,
		req.StoreID,
		req.StorageLocationID,
		req.QuantityOnHand,
		req.QuantityAllocated,
		req.QuantityAvailable,
		req.QuantityOnOrder,
		req.QuantityInTransit,
		req.ReorderLevel,
		req.ReorderQuantity,
		req.MaxStockLevel,
		req.Metadata,
	)
	c.JSON(resp.StatusCode, resp)
}

// AdjustInventoryStock handles POST /api/inventory-stock/:id/adjust
// @Summary      Adjust inventory stock
// @Description  Adjust inventory stock quantities by deltas
// @Tags         inventory_stock
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                    true  "Tenant identifier"
// @Param        Authorization header    string                    true  "Bearer token"
// @Param        id            path      string                    true  "Inventory stock ID"
// @Param        body          body      AdjustInventoryStockRequest true  "Adjustment payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/inventory-stock/{id}/adjust [post]
func (h *InventoryStockHandler) AdjustInventoryStock(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req AdjustInventoryStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.AdjustInventoryStock(
		c.Request.Context(),
		id,
		req.QuantityOnHandDelta,
		req.QuantityAvailableDelta,
		req.QuantityAllocatedDelta,
		req.QuantityOnOrderDelta,
		req.QuantityInTransitDelta,
	)
	c.JSON(resp.StatusCode, resp)
}

// AdjustInventoryStockByProductAndStore handles POST /api/inventory-stock/adjust
// @Summary      Adjust inventory stock by product and store
// @Description  Adjust inventory stock quantities by product and store
// @Tags         inventory_stock
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id       header    string                    true  "Tenant identifier"
// @Param        Authorization     header    string                    true  "Bearer token"
// @Param        product_id        query     string                    true  "Product ID"
// @Param        product_variant_id query    string                    false "Product variant ID"
// @Param        store_id          query     string                    true  "Store ID"
// @Param        body              body      AdjustInventoryStockRequest true  "Adjustment payload"
// @Success      200               {object}  SuccessResponse
// @Failure      400               {object}  ErrorResponse
// @Failure      401               {object}  ErrorResponse
// @Failure      404               {object}  ErrorResponse
// @Failure      500               {object}  ErrorResponse
// @Router       /api/inventory-stock/adjust [post]
func (h *InventoryStockHandler) AdjustInventoryStockByProductAndStore(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	productID := c.Query("product_id")
	productVariantID := c.Query("product_variant_id")
	storeID := c.Query("store_id")

	if productID == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "product_id is required", nil))
		return
	}
	if storeID == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "store_id is required", nil))
		return
	}

	var req AdjustInventoryStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	var pvID *string
	if productVariantID != "" {
		pvID = &productVariantID
	}

	resp := h.useCase.AdjustInventoryStockByProductAndStore(
		c.Request.Context(),
		productID,
		pvID,
		storeID,
		req.QuantityOnHandDelta,
		req.QuantityAvailableDelta,
		req.QuantityAllocatedDelta,
		req.QuantityOnOrderDelta,
		req.QuantityInTransitDelta,
	)
	c.JSON(resp.StatusCode, resp)
}

// DeleteInventoryStock handles DELETE /api/inventory-stock/:id
// @Summary      Delete inventory stock
// @Description  Delete inventory stock by ID
// @Tags         inventory_stock
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Inventory stock ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/inventory-stock/{id} [delete]
func (h *InventoryStockHandler) DeleteInventoryStock(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.DeleteInventoryStock(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// ListInventoryStock handles GET /api/inventory-stock
// @Summary      List inventory stock
// @Description  List all inventory stock records
// @Tags         inventory_stock
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/inventory-stock [get]
func (h *InventoryStockHandler) ListInventoryStock(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListInventoryStock(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListInventoryStockByStore handles GET /api/inventory-stock/store/:store_id
// @Summary      List inventory stock by store
// @Description  List inventory stock records for a specific store
// @Tags         inventory_stock
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        store_id      path      string  true  "Store ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/inventory-stock/store/{store_id} [get]
func (h *InventoryStockHandler) ListInventoryStockByStore(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeID := c.Param("store_id")
	resp := h.useCase.ListInventoryStockByStore(c.Request.Context(), storeID)
	c.JSON(resp.StatusCode, resp)
}

// ListInventoryStockByProduct handles GET /api/inventory-stock/product/:product_id
// @Summary      List inventory stock by product
// @Description  List inventory stock records for a specific product
// @Tags         inventory_stock
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
// @Router       /api/inventory-stock/product/{product_id} [get]
func (h *InventoryStockHandler) ListInventoryStockByProduct(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	productID := c.Param("product_id")
	resp := h.useCase.ListInventoryStockByProduct(c.Request.Context(), productID)
	c.JSON(resp.StatusCode, resp)
}

// ListInventoryStockByStorageLocation handles GET /api/inventory-stock/storage-location/:storage_location_id
// @Summary      List inventory stock by storage location
// @Description  List inventory stock records for a specific storage location
// @Tags         inventory_stock
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id        header    string  true  "Tenant identifier"
// @Param        Authorization      header    string  true  "Bearer token"
// @Param        storage_location_id path      string  true  "Storage location ID"
// @Success      200                {object}  SuccessResponse
// @Failure      400                {object}  ErrorResponse
// @Failure      401                {object}  ErrorResponse
// @Failure      500                {object}  ErrorResponse
// @Router       /api/inventory-stock/storage-location/{storage_location_id} [get]
func (h *InventoryStockHandler) ListInventoryStockByStorageLocation(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storageLocationID := c.Param("storage_location_id")
	resp := h.useCase.ListInventoryStockByStorageLocation(c.Request.Context(), storageLocationID)
	c.JSON(resp.StatusCode, resp)
}

// ListInventoryStockByStoreAndLocation handles GET /api/inventory-stock/store/:store_id/location
// @Summary      List inventory stock by store and location
// @Description  List inventory stock records for a specific store and location
// @Tags         inventory_stock
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id        header    string  true  "Tenant identifier"
// @Param        Authorization      header    string  true  "Bearer token"
// @Param        store_id           path      string  true  "Store ID"
// @Param        storage_location_id query     string  false "Storage location ID"
// @Success      200                {object}  SuccessResponse
// @Failure      400                {object}  ErrorResponse
// @Failure      401                {object}  ErrorResponse
// @Failure      500                {object}  ErrorResponse
// @Router       /api/inventory-stock/store/{store_id}/location [get]
func (h *InventoryStockHandler) ListInventoryStockByStoreAndLocation(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeID := c.Param("store_id")
	storageLocationID := c.Query("storage_location_id")

	var slID *string
	if storageLocationID != "" {
		slID = &storageLocationID
	}

	resp := h.useCase.ListInventoryStockByStoreAndLocation(c.Request.Context(), storeID, slID)
	c.JSON(resp.StatusCode, resp)
}

// GetInventoryStockSummary handles GET /api/inventory-stock/store/:store_id/summary
// @Summary      Get inventory stock summary
// @Description  Get inventory stock summary for a store
// @Tags         inventory_stock
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        store_id      path      string  true  "Store ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/inventory-stock/store/{store_id}/summary [get]
func (h *InventoryStockHandler) GetInventoryStockSummary(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeID := c.Param("store_id")
	resp := h.useCase.GetInventoryStockSummary(c.Request.Context(), storeID)
	c.JSON(resp.StatusCode, resp)
}

// InventoryStoreTotals documents totals for one store. Values use on-hand base units.
type InventoryStoreTotals struct {
	StoreID               int32   `json:"store_id"`
	StoreName             string  `json:"store_name"`
	ProductCount          int64   `json:"product_count"`
	TotalUnits            float64 `json:"total_units"`
	TotalAvailableUnits   float64 `json:"total_available_units"`
	TotalCostValue        float64 `json:"total_cost_value"`
	TotalSaleValue        float64 `json:"total_sale_value"`
	UnitsMissingCostPrice float64 `json:"units_missing_cost_price"`
	UnitsMissingSalePrice float64 `json:"units_missing_sale_price"`
}

// InventoryTotalsData documents the JSON produced with PostgreSQL decimal arithmetic.
type InventoryTotalsData struct {
	OrganizationID        int32                  `json:"organization_id"`
	StoreID               *int32                 `json:"store_id"`
	StorageLocationID     *int32                 `json:"storage_location_id"`
	PriceListID           int32                  `json:"price_list_id"`
	CurrencyCode          string                 `json:"currency_code"`
	QuantityBasis         string                 `json:"quantity_basis"`
	ProductCount          int64                  `json:"product_count"`
	StoreCount            int64                  `json:"store_count"`
	TotalUnits            float64                `json:"total_units"`
	TotalAvailableUnits   float64                `json:"total_available_units"`
	TotalCostValue        float64                `json:"total_cost_value"`
	TotalSaleValue        float64                `json:"total_sale_value"`
	UnitsMissingCostPrice float64                `json:"units_missing_cost_price"`
	UnitsMissingSalePrice float64                `json:"units_missing_sale_price"`
	Stores                []InventoryStoreTotals `json:"stores"`
}

// GetInventoryTotals handles GET /api/inventory-stock/totals.
// @Summary      Inventory units and valuation across stores
// @Description  Returns combined totals and per-store totals for inventory-tracked products in one organization. total_units sums quantity_on_hand in base units, including allocated stock and signed negative quantities; available units are returned separately. Cost value is on-hand quantity times current products.cost_price. Sale value is on-hand quantity times one currently active base-UOM (or unspecified-UOM) unit price from the selected price list; variant-specific prices take precedence over product prices. Quantity tiers must apply to one unit. Prices are used as stored, without tax adjustments or discounts. Missing costs/prices contribute zero value and are reported as absolute units_missing_*_price. Price List 01 (ID 1) is the default. No pagination: totals cover every matching inventory row. Omit store/location for all stores; specify store_id for one store; specify both store_id and storage_location_id for one location. Empty stores return zero totals. The price list must use the organization's currency. On databases without products.cost_price, costs are reported as missing.
// @Tags         inventory-stock
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        organization_id query int true "Organization ID" minimum(1)
// @Param        store_id query int false "Store ID (omit or 0 for all stores)" minimum(0)
// @Param        storage_location_id query int false "Storage location ID; requires store_id (omit or 0 for all locations)" minimum(0)
// @Param        price_list_id query int false "Selling price list" default(1) minimum(1)
// @Success      200 {object} SuccessResponse{data=InventoryTotalsData}
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/inventory-stock/totals [get]
func (h *InventoryStockHandler) GetInventoryTotals(c *gin.Context) {
	params, err := parseInventoryTotalsParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	uc := usecase.NewInventoryStockUseCase()
	uc.SetRepository(repo)
	resp := uc.GetInventoryTotals(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

func parseInventoryTotalsParams(c *gin.Context) (repository.GetInventoryTotalsParams, error) {
	var params repository.GetInventoryTotalsParams
	for _, field := range []struct {
		name, fallback string
		minimum        int64
		dest           *int32
	}{
		{"organization_id", "", 1, &params.OrganizationID},
		{"store_id", "0", 0, &params.StoreID},
		{"storage_location_id", "0", 0, &params.StorageLocationID},
		{"price_list_id", "1", 1, &params.PriceListID},
	} {
		value, err := strconv.ParseInt(c.DefaultQuery(field.name, field.fallback), 10, 32)
		if err != nil || value < field.minimum {
			return params, fmt.Errorf("invalid or missing %s", field.name)
		}
		*field.dest = int32(value)
	}
	if params.StorageLocationID != 0 && params.StoreID == 0 {
		return params, fmt.Errorf("store_id is required with storage_location_id")
	}
	return params, nil
}
