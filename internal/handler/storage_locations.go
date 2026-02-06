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

// StorageLocationsHandler holds the storage locations use case.
type StorageLocationsHandler struct {
	useCase *usecase.StorageLocationsUseCase
}

// NewStorageLocationsHandler creates a new storage locations handler.
func NewStorageLocationsHandler(uc *usecase.StorageLocationsUseCase) *StorageLocationsHandler {
	return &StorageLocationsHandler{useCase: uc}
}

func (h *StorageLocationsHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateStorageLocation handles POST /api/storage-locations
// @Summary      Create storage location
// @Description  Creates a new storage location for a store
// @Tags         storage-locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        body          body      CreateStorageLocationRequest  true  "Storage location payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/storage-locations [post]
func (h *StorageLocationsHandler) CreateStorageLocation(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateStorageLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	input := &usecase.CreateStorageLocationInput{
		StoreID:          req.StoreID,
		Code:             req.Code,
		Name:             req.Name,
		LocationType:     req.LocationType,
		ParentLocationID: req.ParentLocationID,
		IsActive:         req.IsActive,
		Metadata:         nil,
	}
	resp := h.useCase.CreateStorageLocation(c.Request.Context(), input)
	c.JSON(resp.StatusCode, resp)
}

// GetStorageLocation handles GET /api/storage-locations/:id
// @Summary      Get storage location by ID
// @Description  Returns a single storage location by id
// @Tags         storage-locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Storage location ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/storage-locations/{id} [get]
func (h *StorageLocationsHandler) GetStorageLocation(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid id", nil))
		return
	}
	resp := h.useCase.GetStorageLocation(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// GetStorageLocationByCode handles GET /api/stores/:store_id/storage-locations/code/:code
// @Summary      Get storage location by code
// @Description  Returns a storage location by store ID and code
// @Tags         storage-locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        store_id      path      int     true  "Store ID"
// @Param        code          path      string  true  "Location code"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stores/{store_id}/storage-locations/code/{code} [get]
func (h *StorageLocationsHandler) GetStorageLocationByCode(c *gin.Context) {
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
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "code is required", nil))
		return
	}
	resp := h.useCase.GetStorageLocationByCode(c.Request.Context(), int32(storeID), code)
	c.JSON(resp.StatusCode, resp)
}

// ListStorageLocations handles GET /api/storage-locations
// @Summary      List all storage locations
// @Description  Returns all storage locations
// @Tags         storage-locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/storage-locations [get]
func (h *StorageLocationsHandler) ListStorageLocations(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListStorageLocations(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListStorageLocationsByStore handles GET /api/stores/:store_id/storage-locations
// @Summary      List storage locations by store
// @Description  Returns all storage locations for a store
// @Tags         storage-locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        store_id      path      int     true  "Store ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stores/{store_id}/storage-locations [get]
func (h *StorageLocationsHandler) ListStorageLocationsByStore(c *gin.Context) {
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
	resp := h.useCase.ListStorageLocationsByStore(c.Request.Context(), int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// ListActiveStorageLocationsByStore handles GET /api/stores/:store_id/storage-locations/active
// @Summary      List active storage locations by store
// @Description  Returns active storage locations for a store
// @Tags         storage-locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        store_id      path      int     true  "Store ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stores/{store_id}/storage-locations/active [get]
func (h *StorageLocationsHandler) ListActiveStorageLocationsByStore(c *gin.Context) {
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
	resp := h.useCase.ListActiveStorageLocationsByStore(c.Request.Context(), int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// ListStorageLocationsByParent handles GET /api/storage-locations/by-parent?parent_id=...
// @Summary      List storage locations by parent
// @Description  Returns storage locations by parent location ID (optional query: parent_id)
// @Tags         storage-locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        parent_id     query     int     false  "Parent location ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/storage-locations/by-parent [get]
func (h *StorageLocationsHandler) ListStorageLocationsByParent(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var parentID *int32
	if s := c.Query("parent_id"); s != "" {
		id, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid parent_id", nil))
			return
		}
		pid := int32(id)
		parentID = &pid
	}
	resp := h.useCase.ListStorageLocationsByParent(c.Request.Context(), parentID)
	c.JSON(resp.StatusCode, resp)
}

// ListStorageLocationsByType handles GET /api/stores/:store_id/storage-locations/type/:location_type
// @Summary      List storage locations by type
// @Description  Returns storage locations for a store filtered by location type
// @Tags         storage-locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        store_id        path      int     true  "Store ID"
// @Param        location_type   path      string  true  "Location type"
// @Success      200             {object}  SuccessResponse
// @Failure      400             {object}  ErrorResponse
// @Failure      401             {object}  ErrorResponse
// @Failure      404             {object}  ErrorResponse
// @Failure      500             {object}  ErrorResponse
// @Router       /api/stores/{store_id}/storage-locations/type/{location_type} [get]
func (h *StorageLocationsHandler) ListStorageLocationsByType(c *gin.Context) {
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
	locationType := c.Param("location_type")
	if locationType == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "location_type is required", nil))
		return
	}
	resp := h.useCase.ListStorageLocationsByType(c.Request.Context(), int32(storeID), locationType)
	c.JSON(resp.StatusCode, resp)
}

// UpdateStorageLocation handles PUT /api/storage-locations/:id
// @Summary      Update storage location
// @Description  Updates an existing storage location
// @Tags         storage-locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Storage location ID"
// @Param        body          body      UpdateStorageLocationRequest  true  "Update payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/storage-locations/{id} [put]
func (h *StorageLocationsHandler) UpdateStorageLocation(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid id", nil))
		return
	}
	var req UpdateStorageLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}
	input := &usecase.UpdateStorageLocationInput{
		ID:               int32(id),
		Name:             req.Name,
		LocationType:     req.LocationType,
		ParentLocationID: req.ParentLocationID,
		IsActive:         req.IsActive,
		Metadata:         nil,
	}
	resp := h.useCase.UpdateStorageLocation(c.Request.Context(), input)
	c.JSON(resp.StatusCode, resp)
}

// DeleteStorageLocation handles DELETE /api/storage-locations/:id
// @Summary      Delete storage location
// @Description  Deletes a storage location by id
// @Tags         storage-locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Storage location ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/storage-locations/{id} [delete]
func (h *StorageLocationsHandler) DeleteStorageLocation(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid id", nil))
		return
	}
	resp := h.useCase.DeleteStorageLocation(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// ToggleStorageLocationActive handles PATCH /api/storage-locations/:id/active
// @Summary      Toggle storage location active state
// @Description  Sets the active state of a storage location
// @Tags         storage-locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Storage location ID"
// @Param        body          body      ToggleStorageLocationActiveRequest  true  "Active state"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/storage-locations/{id}/active [patch]
func (h *StorageLocationsHandler) ToggleStorageLocationActive(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid id", nil))
		return
	}
	var req ToggleStorageLocationActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}
	resp := h.useCase.ToggleStorageLocationActive(c.Request.Context(), int32(id), req.IsActive)
	c.JSON(resp.StatusCode, resp)
}
