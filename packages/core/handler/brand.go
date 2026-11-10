package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/gin-gonic/gin"
)

type BrandHandler struct {
	useCase *usecase.BrandUseCase
}

func NewBrandHandler(uc *usecase.BrandUseCase) *BrandHandler {
	return &BrandHandler{
		useCase: uc,
	}
}

// getRepositoryFromContext extracts repository from Gin context
func (h *BrandHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateBrand handles POST /brands
// @Summary      Create a new brand
// @Description  Create a new brand with required name and code
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        brand         body      CreateBrandRequest  true  "Brand data"
// @Success      201           {object}  BrandResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands [post]
func (h *BrandHandler) CreateBrand(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req struct {
		Name     string           `json:"name" binding:"required"`
		Code     string           `json:"code" binding:"required"`
		IsActive bool             `json:"is_active"`
		Metadata *json.RawMessage `json:"metadata,omitempty"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	resp := h.useCase.CreateBrand(c.Request.Context(), req.Name, req.Code, req.IsActive, req.Metadata)
	c.JSON(resp.StatusCode, resp)
}

// CreateBrandWithDefaults handles POST /brands/with-defaults
// @Summary      Create a new brand with defaults
// @Description  Create a new brand with default active status
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        brand         body      CreateBrandWithDefaultsRequest  true  "Brand data"
// @Success      201           {object}  BrandResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/with-defaults [post]
func (h *BrandHandler) CreateBrandWithDefaults(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req struct {
		Name string `json:"name" binding:"required"`
		Code string `json:"code" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	resp := h.useCase.CreateBrandWithDefaults(c.Request.Context(), req.Name, req.Code)
	c.JSON(resp.StatusCode, resp)
}

// GetBrandByID handles GET /brands/:id
// @Summary      Get brand by ID
// @Description  Retrieve a specific brand by its ID
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Brand ID"
// @Success      200           {object}  BrandResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/brands/{id} [get]
func (h *BrandHandler) GetBrandByID(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetBrandByID(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// GetBrandByCode handles GET /brands/code/:code
// @Summary      Get brand by code
// @Description  Retrieve a specific brand by its code
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        code          path      string  true  "Brand code"
// @Success      200           {object}  BrandResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/brands/code/{code} [get]
func (h *BrandHandler) GetBrandByCode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	code := c.Param("code")
	resp := h.useCase.GetBrandByCode(c.Request.Context(), code)
	c.JSON(resp.StatusCode, resp)
}

// ListAllBrands handles GET /brands/all
// @Summary      List all brands
// @Description  Retrieve a list of all brands without pagination
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {array}   BrandResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/all [get]
func (h *BrandHandler) ListAllBrands(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListAllBrands(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListActiveBrands handles GET /brands/active
// @Summary      List active brands
// @Description  Retrieve a list of only active brands
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {array}   BrandResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/active [get]
func (h *BrandHandler) ListActiveBrands(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListActiveBrands(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListBrands handles GET /brands
// @Summary      List brands with pagination
// @Description  Retrieve a list of brands with pagination
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        limit         query     int     false "Limit"
// @Param        offset        query     int     false "Offset"
// @Success      200           {array}   BrandResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands [get]
func (h *BrandHandler) ListBrands(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.ParseInt(limitStr, 10, 32)
	offset, _ := strconv.ParseInt(offsetStr, 10, 32)

	resp := h.useCase.ListBrands(c.Request.Context(), int32(limit), int32(offset))
	c.JSON(resp.StatusCode, resp)
}

// ListActiveBrandsWithPagination handles GET /brands/active/paginated
// @Summary      List active brands with pagination
// @Description  Retrieve a list of active brands with pagination
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        limit         query     int     false "Limit"
// @Param        offset        query     int     false "Offset"
// @Success      200           {array}   BrandResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/active/paginated [get]
func (h *BrandHandler) ListActiveBrandsWithPagination(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.ParseInt(limitStr, 10, 32)
	offset, _ := strconv.ParseInt(offsetStr, 10, 32)

	resp := h.useCase.ListActiveBrandsWithPagination(c.Request.Context(), int32(limit), int32(offset))
	c.JSON(resp.StatusCode, resp)
}

// CountBrands handles GET /brands/count
// @Summary      Count brands
// @Description  Get total number of brands
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  CountResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/count [get]
func (h *BrandHandler) CountBrands(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.CountBrands(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// CountActiveBrands handles GET /brands/count/active
// @Summary      Count active brands
// @Description  Get total number of active brands
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  CountResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/count/active [get]
func (h *BrandHandler) CountActiveBrands(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.CountActiveBrands(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// SearchBrands handles GET /brands/search
// @Summary      Search brands
// @Description  Search brands by name or code
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        q             query     string  true  "Search term"
// @Param        limit         query     int     false "Limit"
// @Param        offset        query     int     false "Offset"
// @Success      200           {array}   BrandResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/search [get]
func (h *BrandHandler) SearchBrands(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	searchTerm := c.Query("q")
	if searchTerm == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "search term is required", nil))
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.ParseInt(limitStr, 10, 32)
	offset, _ := strconv.ParseInt(offsetStr, 10, 32)

	resp := h.useCase.SearchBrands(c.Request.Context(), searchTerm, int32(limit), int32(offset))
	c.JSON(resp.StatusCode, resp)
}

// SearchActiveBrands handles GET /brands/search/active
// @Summary      Search active brands
// @Description  Search active brands by name or code
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        q             query     string  true  "Search term"
// @Param        limit         query     int     false "Limit"
// @Param        offset        query     int     false "Offset"
// @Success      200           {array}   BrandResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/search/active [get]
func (h *BrandHandler) SearchActiveBrands(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	searchTerm := c.Query("q")
	if searchTerm == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "search term is required", nil))
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.ParseInt(limitStr, 10, 32)
	offset, _ := strconv.ParseInt(offsetStr, 10, 32)

	resp := h.useCase.SearchActiveBrands(c.Request.Context(), searchTerm, int32(limit), int32(offset))
	c.JSON(resp.StatusCode, resp)
}

// UpdateBrand handles PUT /brands/:id
// @Summary      Update a brand
// @Description  Update an existing brand by its ID
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Brand ID"
// @Param        brand         body      UpdateBrandRequest  true  "Brand data"
// @Success      200           {object}  BrandResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/{id} [put]
func (h *BrandHandler) UpdateBrand(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req struct {
		Name     *string          `json:"name,omitempty"`
		Code     *string          `json:"code,omitempty"`
		IsActive *bool            `json:"is_active,omitempty"`
		Metadata *json.RawMessage `json:"metadata,omitempty"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	resp := h.useCase.UpdateBrand(c.Request.Context(), id, req.Name, req.Code, req.IsActive, req.Metadata)
	c.JSON(resp.StatusCode, resp)
}

// UpdateBrandName handles PATCH /brands/:id/name
// @Summary      Update brand name
// @Description  Update only the brand name
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Brand ID"
// @Param        name          body      UpdateBrandNameRequest  true  "Name data"
// @Success      200           {object}  BrandResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/brands/{id}/name [patch]
func (h *BrandHandler) UpdateBrandName(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	resp := h.useCase.UpdateBrandName(c.Request.Context(), id, req.Name)
	c.JSON(resp.StatusCode, resp)
}

// UpdateBrandCode handles PATCH /brands/:id/code
// @Summary      Update brand code
// @Description  Update only the brand code
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Brand ID"
// @Param        code          body      UpdateBrandCodeRequest  true  "Code data"
// @Success      200           {object}  BrandResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/brands/{id}/code [patch]
func (h *BrandHandler) UpdateBrandCode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	resp := h.useCase.UpdateBrandCode(c.Request.Context(), id, req.Code)
	c.JSON(resp.StatusCode, resp)
}

// UpdateBrandMetadata handles PATCH /brands/:id/metadata
// @Summary      Update brand metadata
// @Description  Update only the brand metadata
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Brand ID"
// @Param        metadata      body      UpdateBrandMetadataRequest  true  "Metadata"
// @Success      200           {object}  BrandResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/brands/{id}/metadata [patch]
func (h *BrandHandler) UpdateBrandMetadata(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req struct {
		Metadata json.RawMessage `json:"metadata" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	resp := h.useCase.UpdateBrandMetadata(c.Request.Context(), id, req.Metadata)
	c.JSON(resp.StatusCode, resp)
}

// ActivateBrand handles PATCH /brands/:id/activate
// @Summary      Activate a brand
// @Description  Activate a brand
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Brand ID"
// @Success      200           {object}  BrandResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/brands/{id}/activate [patch]
func (h *BrandHandler) ActivateBrand(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.ActivateBrand(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// DeactivateBrand handles PATCH /brands/:id/deactivate
// @Summary      Deactivate a brand
// @Description  Deactivate a brand
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Brand ID"
// @Success      200           {object}  BrandResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/brands/{id}/deactivate [patch]
func (h *BrandHandler) DeactivateBrand(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.DeactivateBrand(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// ToggleBrandStatus handles PATCH /brands/:id/toggle-status
// @Summary      Toggle brand status
// @Description  Toggle brand active status
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Brand ID"
// @Success      200           {object}  BrandResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/brands/{id}/toggle-status [patch]
func (h *BrandHandler) ToggleBrandStatus(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.ToggleBrandStatus(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// DeleteBrand handles DELETE /brands/:id
// @Summary      Delete a brand
// @Description  Hard delete a brand by its ID
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Brand ID"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/{id} [delete]
func (h *BrandHandler) DeleteBrand(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.DeleteBrand(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// DeleteBrandByCode handles DELETE /brands/code/:code
// @Summary      Delete a brand by code
// @Description  Hard delete a brand by its code
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        code          path      string  true  "Brand code"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/code/{code} [delete]
func (h *BrandHandler) DeleteBrandByCode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	code := c.Param("code")
	resp := h.useCase.DeleteBrandByCode(c.Request.Context(), code)
	c.JSON(resp.StatusCode, resp)
}

// SoftDeleteBrand handles DELETE /brands/:id/soft
// @Summary      Soft delete a brand
// @Description  Soft delete a brand by deactivating it
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Brand ID"
// @Success      200           {object}  BrandResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/brands/{id}/soft [delete]
func (h *BrandHandler) SoftDeleteBrand(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.SoftDeleteBrand(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// BrandExists handles GET /brands/:id/exists
// @Summary      Check if brand exists
// @Description  Check if a brand exists by ID
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Brand ID"
// @Success      200           {object}  ExistsResponse
// @Failure      401           {object}  ErrorResponse
// @Router       /api/brands/{id}/exists [get]
func (h *BrandHandler) BrandExists(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.BrandExists(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// BrandCodeExists handles GET /brands/code/:code/exists
// @Summary      Check if brand code exists
// @Description  Check if a brand code already exists
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        code          path      string  true  "Brand code"
// @Success      200           {object}  ExistsResponse
// @Failure      401           {object}  ErrorResponse
// @Router       /api/brands/code/{code}/exists [get]
func (h *BrandHandler) BrandCodeExists(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	code := c.Param("code")
	resp := h.useCase.BrandCodeExists(c.Request.Context(), code)
	c.JSON(resp.StatusCode, resp)
}

// GetBrandWithProductCount handles GET /brands/:id/products/count
// @Summary      Get brand with product count
// @Description  Get brand with count of associated products
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Brand ID"
// @Success      200           {object}  BrandWithProductCountResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/brands/{id}/products/count [get]
func (h *BrandHandler) GetBrandWithProductCount(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetBrandWithProductCount(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// ListBrandsWithProductCounts handles GET /brands/products/counts
// @Summary      List brands with product counts
// @Description  List all brands with their product counts
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {array}   BrandWithProductCountResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/products/counts [get]
func (h *BrandHandler) ListBrandsWithProductCounts(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListBrandsWithProductCounts(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListActiveBrandsWithProductCounts handles GET /brands/active/products/counts
// @Summary      List active brands with product counts
// @Description  List active brands with their product counts
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {array}   BrandWithProductCountResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/active/products/counts [get]
func (h *BrandHandler) ListActiveBrandsWithProductCounts(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListActiveBrandsWithProductCounts(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// GetTopBrandsByProductCount handles GET /brands/top
// @Summary      Get top brands by product count
// @Description  Get top N brands by number of products
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        limit         query     int     false "Limit"
// @Success      200           {array}   BrandWithProductCountResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/top [get]
func (h *BrandHandler) GetTopBrandsByProductCount(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.ParseInt(limitStr, 10, 32)

	resp := h.useCase.GetTopBrandsByProductCount(c.Request.Context(), int32(limit))
	c.JSON(resp.StatusCode, resp)
}

// GetBrandsWithNoProducts handles GET /brands/no-products
// @Summary      Get brands with no products
// @Description  Get brands that have no associated products
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {array}   BrandResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/no-products [get]
func (h *BrandHandler) GetBrandsWithNoProducts(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.GetBrandsWithNoProducts(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// GetInactiveBrandsWithActiveProducts handles GET /brands/inactive/active-products
// @Summary      Get inactive brands with active products
// @Description  Get inactive brands that still have active products
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {array}   BrandWithProductCountResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/inactive/active-products [get]
func (h *BrandHandler) GetInactiveBrandsWithActiveProducts(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.GetInactiveBrandsWithActiveProducts(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// BulkActivateBrands handles POST /brands/bulk/activate
// @Summary      Bulk activate brands
// @Description  Activate multiple brands by IDs
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        ids           body      BulkBrandIDsRequest  true  "Brand IDs"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/bulk/activate [post]
func (h *BrandHandler) BulkActivateBrands(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req struct {
		IDs []int32 `json:"ids" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	resp := h.useCase.BulkActivateBrands(c.Request.Context(), req.IDs)
	c.JSON(resp.StatusCode, resp)
}

// BulkDeactivateBrands handles POST /brands/bulk/deactivate
// @Summary      Bulk deactivate brands
// @Description  Deactivate multiple brands by IDs
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        ids           body      BulkBrandIDsRequest  true  "Brand IDs"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/bulk/deactivate [post]
func (h *BrandHandler) BulkDeactivateBrands(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req struct {
		IDs []int32 `json:"ids" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	resp := h.useCase.BulkDeactivateBrands(c.Request.Context(), req.IDs)
	c.JSON(resp.StatusCode, resp)
}

// BulkDeleteBrands handles POST /brands/bulk/delete
// @Summary      Bulk delete brands
// @Description  Delete multiple brands by IDs
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        ids           body      BulkBrandIDsRequest  true  "Brand IDs"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/bulk/delete [post]
func (h *BrandHandler) BulkDeleteBrands(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req struct {
		IDs []int32 `json:"ids" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	resp := h.useCase.BulkDeleteBrands(c.Request.Context(), req.IDs)
	c.JSON(resp.StatusCode, resp)
}

// GetRecentlyCreatedBrands handles GET /brands/recent/created
// @Summary      Get recently created brands
// @Description  Get brands created in the last N days
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        days          query     int     false "Number of days"
// @Success      200           {array}   BrandResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/recent/created [get]
func (h *BrandHandler) GetRecentlyCreatedBrands(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	daysStr := c.DefaultQuery("days", "7")
	days, _ := strconv.Atoi(daysStr)

	resp := h.useCase.GetRecentlyCreatedBrands(c.Request.Context(), days)
	c.JSON(resp.StatusCode, resp)
}

// GetRecentlyUpdatedBrands handles GET /brands/recent/updated
// @Summary      Get recently updated brands
// @Description  Get brands updated in the last N days
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        days          query     int     false "Number of days"
// @Success      200           {array}   BrandResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/recent/updated [get]
func (h *BrandHandler) GetRecentlyUpdatedBrands(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	daysStr := c.DefaultQuery("days", "7")
	days, _ := strconv.Atoi(daysStr)

	resp := h.useCase.GetRecentlyUpdatedBrands(c.Request.Context(), days)
	c.JSON(resp.StatusCode, resp)
}

// GetBrandsByCreationDate handles GET /brands/by-date
// @Summary      Get brands by creation date
// @Description  Get brands created between two dates
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        start_date    query     string  true  "Start date (YYYY-MM-DD)"
// @Param        end_date      query     string  true  "End date (YYYY-MM-DD)"
// @Success      200           {array}   BrandResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/by-date [get]
func (h *BrandHandler) GetBrandsByCreationDate(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "start_date and end_date are required", nil))
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid start_date format", nil))
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid end_date format", nil))
		return
	}

	resp := h.useCase.GetBrandsByCreationDate(c.Request.Context(), startDate, endDate)
	c.JSON(resp.StatusCode, resp)
}

// GetBrandMetadataByKey handles GET /brands/:id/metadata/:key
// @Summary      Get brand metadata by key
// @Description  Get a specific metadata field from a brand
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Brand ID"
// @Param        key           path      string  true  "Metadata key"
// @Success      200           {object}  MetadataResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/brands/{id}/metadata/{key} [get]
func (h *BrandHandler) GetBrandMetadataByKey(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	key := c.Param("key")

	resp := h.useCase.GetBrandMetadataByKey(c.Request.Context(), id, key)
	c.JSON(resp.StatusCode, resp)
}

// ListBrandsWithStats handles GET /brands/stats
// @Summary      List brands with statistics
// @Description  List brands with statistics
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        active_only   query     bool    false "Active only"
// @Param        search        query     string  false "Search term"
// @Param        limit         query     int     false "Limit"
// @Param        offset        query     int     false "Offset"
// @Success      200           {array}   BrandWithStatsResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/brands/stats [get]
func (h *BrandHandler) ListBrandsWithStats(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	activeOnlyStr := c.DefaultQuery("active_only", "false")
	activeOnly := activeOnlyStr == "true"

	searchStr := c.Query("search")
	var search *string
	if searchStr != "" {
		search = &searchStr
	}

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.ParseInt(limitStr, 10, 32)
	offset, _ := strconv.ParseInt(offsetStr, 10, 32)

	resp := h.useCase.ListBrandsWithStats(c.Request.Context(), activeOnly, search, int32(limit), int32(offset))
	c.JSON(resp.StatusCode, resp)
}
