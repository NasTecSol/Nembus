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

// StoreHandler holds the store use case
type StoreHandler struct {
	useCase *usecase.StoreUseCase
}

// NewStoreHandler creates a new handler instance
func NewStoreHandler(uc *usecase.StoreUseCase) *StoreHandler {
	return &StoreHandler{
		useCase: uc,
	}
}

// getRepositoryFromContext extracts repository from Gin context
func (h *StoreHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// --------------------------------------------------
// Create Store
// --------------------------------------------------

// CreateStore handles POST /stores
// @Summary      Create a new store
// @Description  Create a store under tenant organization
// @Tags         stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        store         body      CreateStoreRequest  true  "Store data"
// @Success      201  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/stores [post]
func (h *StoreHandler) CreateStore(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req struct {
		Name          string      `json:"name" binding:"required"`
		Code          string      `json:"code" binding:"required"`
		StoreType     *string     `json:"store_type,omitempty"`
		ParentStoreID *int32      `json:"parent_store_id,omitempty"`
		IsWarehouse   bool        `json:"is_warehouse"`
		IsPOSEnabled  bool        `json:"is_pos_enabled"`
		Timezone      *string     `json:"timezone,omitempty"`
		IsActive      bool        `json:"is_active"`
		Metadata      interface{} `json:"metadata,omitempty"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	resp := h.useCase.CreateStore(
		c.Request.Context(),
		req.Name,
		req.Code,
		req.StoreType,
		req.ParentStoreID,
		req.IsWarehouse,
		req.IsPOSEnabled,
		req.Timezone,
		req.IsActive,
		req.Metadata,
	)

	c.JSON(resp.StatusCode, resp)
}

// --------------------------------------------------
// Get Store
// --------------------------------------------------

// GetStore handles GET /stores/:id
// @Summary      Get store by ID
// @Description  Retrieve a store by its ID
// @Tags         stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id             path      string  true  "Store ID"
// @Success      200  {object}  StoreResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/stores/{id} [get]
func (h *StoreHandler) GetStore(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetStore(c.Request.Context(), id)

	if resp.StatusCode != utils.CodeOK {
		c.JSON(resp.StatusCode, gin.H{"error": resp.Message})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// --------------------------------------------------
// List Stores
// --------------------------------------------------

// ListStores handles GET /stores
// @Summary      List stores
// @Description  Retrieve paginated list of stores
// @Tags         stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        limit          query     int     false "Limit"
// @Param        offset         query     int     false "Offset"
// @Param        is_active      query     bool    false "Filter by active status"
// @Param        store_type     query     string  false "Filter by store type"
// @Success      200  {array}   StoreResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/stores [get]
func (h *StoreHandler) ListStores(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		limit = 100
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil {
		offset = 0
	}

	var isActive *bool
	if v := c.Query("is_active"); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			isActive = &parsed
		}
	}

	var storeType *string
	if v := c.Query("store_type"); v != "" {
		storeType = &v
	}

	resp := h.useCase.ListStores(
		c.Request.Context(),
		int32(limit),
		int32(offset),
		isActive,
		storeType,
	)

	c.JSON(resp.StatusCode, resp)
}

// --------------------------------------------------
// Delete Store
// --------------------------------------------------

// DeleteStore handles DELETE /stores/:id
// @Summary      Delete store
// @Description  Delete store by ID
// @Tags         stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id             path      string  true  "Store ID"
// @Success      200  {object}  SuccessResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/stores/{id} [delete]
func (h *StoreHandler) DeleteStore(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.DeleteStore(c.Request.Context(), id)

	c.JSON(resp.StatusCode, resp)
}

// --------------------------------------------------
// List POS Enabled Stores
// --------------------------------------------------

// ListPOSEnabledStores handles GET /stores/pos-enabled
// @Summary      List POS-enabled stores
// @Description  Retrieve all active stores that have POS enabled
// @Tags         stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200  {array}   StoreResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/stores/pos-enabled [get]
func (h *StoreHandler) ListPOSEnabledStores(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListPOSEnabledStores(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// --------------------------------------------------
// List Warehouse Stores
// --------------------------------------------------

// ListWarehouseStores handles GET /stores/warehouses
// @Summary      List warehouse stores
// @Description  Retrieve all active stores marked as warehouse
// @Tags         stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200  {array}   StoreResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/stores/warehouses [get]
func (h *StoreHandler) ListWarehouseStores(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListWarehouseStores(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// --------------------------------------------------
// List Stores By Parent
// --------------------------------------------------

// ListStoresByParent handles GET /stores/parent/:parent_id
// @Summary      List stores by parent
// @Description  Retrieve stores that are children of a parent store
// @Tags         stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        parent_id      path      int     true  "Parent Store ID"
// @Param        is_active      query     bool    false "Filter by active status"
// @Success      200  {array}   StoreResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/stores/parent/{parent_id} [get]
func (h *StoreHandler) ListStoresByParent(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	parentIDStr := c.Param("parent_id")
	parentID, err := strconv.ParseInt(parentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent store ID"})
		return
	}

	var isActive *bool
	if v := c.Query("is_active"); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			isActive = &parsed
		}
	}

	resp := h.useCase.ListStoresByParent(c.Request.Context(), int32(parentID), isActive)
	c.JSON(resp.StatusCode, resp)
}

// --------------------------------------------------
// Get Storage Location Hierarchy
// --------------------------------------------------

// GetStorageLocationHierarchy handles GET /stores/:id/locations
// @Summary      Get storage location hierarchy
// @Description  Retrieve the full recursive storage location hierarchy for a store
// @Tags         stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id             path      int     true  "Store ID"
// @Param        is_active      query     bool    false "Filter locations by active status"
// @Success 200 {array} object
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/stores/{id}/locations [get]
func (h *StoreHandler) GetStorageLocationHierarchy(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeIDStr := c.Param("id")
	storeID, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store ID"})
		return
	}

	var isActive *bool
	if v := c.Query("is_active"); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			isActive = &parsed
		}
	}

	resp := h.useCase.GetStorageLocationHierarchy(c.Request.Context(), int32(storeID), isActive)
	c.JSON(resp.StatusCode, resp)
}

// --------------------------------------------------
// Update Store
// --------------------------------------------------

// UpdateStore handles PATCH /stores/:id
// @Summary      Update store
// @Description  Update store details by ID
// @Tags         stores
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id             path      string  true  "Store ID"
// @Param        store          body      UpdateStoreRequest  true  "Store fields to update"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/stores/{id} [patch]
func (h *StoreHandler) UpdateStore(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req struct {
		Name         *string     `json:"name,omitempty"`
		StoreType    *string     `json:"store_type,omitempty"`
		IsWarehouse  *bool       `json:"is_warehouse,omitempty"`
		IsPOSEnabled *bool       `json:"is_pos_enabled,omitempty"`
		Timezone     *string     `json:"timezone,omitempty"`
		IsActive     *bool       `json:"is_active,omitempty"`
		Metadata     interface{} `json:"metadata,omitempty"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	resp := h.useCase.UpdateStore(
		c.Request.Context(),
		id,
		req.Name,
		req.StoreType,
		req.IsWarehouse,
		req.IsPOSEnabled,
		req.Timezone,
		req.IsActive,
		req.Metadata,
	)

	c.JSON(resp.StatusCode, resp)
}
