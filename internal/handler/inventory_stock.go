package handler

import (
	"net/http"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"
	"NEMBUS/utils"

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
	QuantityAllocated  *string                `json:"quantity_allocated,omitempty"`
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
