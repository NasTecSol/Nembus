package handler

import (
	"encoding/json"
	"net/http"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"
	"NEMBUS/utils"

	"github.com/gin-gonic/gin"
)

type ProductCategoryHandler struct {
	useCase *usecase.ProductCategoryUseCase
}

func NewProductCategoryHandler(uc *usecase.ProductCategoryUseCase) *ProductCategoryHandler {
	return &ProductCategoryHandler{
		useCase: uc,
	}
}

func (h *ProductCategoryHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateProductCategory handles POST /product-categories
// @Summary      Create a new product category
// @Description  Create a new product category
// @Tags         product-categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        category      body      CreateProductCategoryRequest  true  "Category data"
// @Success      201           {object}  ProductCategoryResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/product-categories [post]
func (h *ProductCategoryHandler) CreateProductCategory(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateProductCategoryRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	meta, _ := json.Marshal(req.Metadata)
	metaRaw := json.RawMessage(meta)

	resp := h.useCase.CreateProductCategory(c.Request.Context(), req.ParentCategoryID, req.Name, req.Code, req.Description, req.CategoryLevel, req.IsActive, &metaRaw)
	c.JSON(resp.StatusCode, resp)
}

// GetProductCategory handles GET /product-categories/:id
// @Summary      Get product category by ID
// @Description  Get product category by ID
// @Tags         product-categories
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Category ID"
// @Success      200  {object}  ProductCategoryResponse
// @Router       /api/product-categories/{id} [get]
func (h *ProductCategoryHandler) GetProductCategory(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetProductCategory(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// GetProductCategoryByCode handles GET /product-categories/code/:code
// @Summary      Get product category by code
// @Description  Get product category by code
// @Tags         product-categories
// @Accept       json
// @Produce      json
// @Param        code  path      string  true  "Category Code"
// @Success      200   {object}  ProductCategoryResponse
// @Router       /api/product-categories/code/{code} [get]
func (h *ProductCategoryHandler) GetProductCategoryByCode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	code := c.Param("code")
	resp := h.useCase.GetProductCategoryByCode(c.Request.Context(), code)
	c.JSON(resp.StatusCode, resp)
}

// ListProductCategories handles GET /product-categories
// @Summary      List root product categories
// @Description  List root product categories (where parent_category_id is NULL)
// @Tags         product-categories
// @Accept       json
// @Produce      json
// @Param        is_active  query     bool  false  "Filter by active status"
// @Success      200        {array}   ProductCategoryResponse
// @Router       /api/product-categories [get]
func (h *ProductCategoryHandler) ListProductCategories(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var isActive *bool
	if val, ok := c.GetQuery("is_active"); ok {
		b := val == "true"
		isActive = &b
	}

	resp := h.useCase.ListProductCategories(c.Request.Context(), isActive)
	c.JSON(resp.StatusCode, resp)
}

// UpdateProductCategory handles PUT /product-categories/:id
// @Summary      Update product category
// @Description  Update product category
// @Tags         product-categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id        path      int     true  "Category ID"
// @Param        category  body      UpdateProductCategoryRequest  true  "Category data"
// @Success      200       {object}  ProductCategoryResponse
// @Router       /api/product-categories/{id} [put]
func (h *ProductCategoryHandler) UpdateProductCategory(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req UpdateProductCategoryRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	var metaRaw *json.RawMessage
	if req.Metadata != nil {
		meta, _ := json.Marshal(req.Metadata)
		mr := json.RawMessage(meta)
		metaRaw = &mr
	}

	resp := h.useCase.UpdateProductCategory(c.Request.Context(), id, req.ParentCategoryID, req.Name, req.Description, req.CategoryLevel, req.IsActive, metaRaw)
	c.JSON(resp.StatusCode, resp)
}

// DeleteProductCategory handles DELETE /product-categories/:id
// @Summary      Delete product category
// @Description  Delete product category
// @Tags         product-categories
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Category ID"
// @Success      200  {object}  SuccessResponse
// @Router       /api/product-categories/{id} [delete]
func (h *ProductCategoryHandler) DeleteProductCategory(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.DeleteProductCategory(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// GetCategoryHierarchy handles GET /product-categories/hierarchy
// @Summary      Get full category hierarchy
// @Description  Get full category hierarchy using recursive query
// @Tags         product-categories
// @Accept       json
// @Produce      json
// @Param        is_active  query     bool  false  "Filter by active status"
// @Success      200        {array}   CategoryHierarchyResponse
// @Router       /api/product-categories/hierarchy [get]
func (h *ProductCategoryHandler) GetCategoryHierarchy(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var isActive *bool
	if val, ok := c.GetQuery("is_active"); ok {
		b := val == "true"
		isActive = &b
	}

	resp := h.useCase.GetCategoryHierarchy(c.Request.Context(), isActive)
	c.JSON(resp.StatusCode, resp)
}

// ListCategoryChildren handles GET /product-categories/:id/children
// @Summary      List sub-categories of a category
// @Description  List sub-categories of a category
// @Tags         product-categories
// @Accept       json
// @Produce      json
// @Param        id         path      int   true   "Parent Category ID"
// @Param        is_active  query     bool  false  "Filter by active status"
// @Success      200        {array}   ProductCategoryResponse
// @Router       /api/product-categories/{id}/children [get]
func (h *ProductCategoryHandler) ListCategoryChildren(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	var isActive *bool
	if val, ok := c.GetQuery("is_active"); ok {
		b := val == "true"
		isActive = &b
	}

	resp := h.useCase.ListCategoryChildren(c.Request.Context(), id, isActive)
	c.JSON(resp.StatusCode, resp)
}
