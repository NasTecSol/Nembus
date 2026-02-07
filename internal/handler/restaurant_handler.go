package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"
	"NEMBUS/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type RestaurantHandler struct {
	useCase *usecase.RestaurantUseCase
}

func NewRestaurantHandler(uc *usecase.RestaurantUseCase) *RestaurantHandler {
	return &RestaurantHandler{useCase: uc}
}

func (h *RestaurantHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// === Tables ===

// ListTables handles GET /api/restaurant/stores/:store_id/tables
// @Summary      List restaurant tables
// @Description  Returns all tables for a given store.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        store_id     path      int     true   "Store ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/stores/{store_id}/tables [get]
func (h *RestaurantHandler) ListTables(c *gin.Context) {
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

	resp := h.useCase.ListTables(c.Request.Context(), int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// GetTable handles GET /api/restaurant/tables/:id
// @Summary      Get restaurant table
// @Description  Returns a single restaurant table by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Table ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/tables/{id} [get]
func (h *RestaurantHandler) GetTable(c *gin.Context) {
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

	resp := h.useCase.GetTable(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// CreateTable handles POST /api/restaurant/tables
// @Summary      Create restaurant table
// @Description  Creates a new table in a store.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        body         body      CreateRestaurantTableRequest  true  "Table data"
// @Success      201          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/tables [post]
func (h *RestaurantHandler) CreateTable(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateRestaurantTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	params := repository.CreateRestaurantTableParams{
		StoreID:     req.StoreID,
		TableNumber: req.TableNumber,
		TableName:   pgtype.Text{String: req.TableName, Valid: req.TableName != ""},
		Section:     pgtype.Text{String: req.Section, Valid: req.Section != ""},
		Capacity:    pgtype.Int4{Int32: req.Capacity, Valid: true},
		IsActive:    pgtype.Bool{Bool: req.IsActive, Valid: true},
		Metadata:    []byte(req.Metadata),
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.CreateTable(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// UpdateTable handles PUT /api/restaurant/tables/:id
// @Summary      Update restaurant table
// @Description  Updates an existing restaurant table.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Table ID"
// @Param        body         body      CreateRestaurantTableRequest  true  "Updated table data"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/tables/{id} [put]
func (h *RestaurantHandler) UpdateTable(c *gin.Context) {
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

	var req CreateRestaurantTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	params := repository.UpdateRestaurantTableParams{
		ID:          int32(id),
		TableNumber: req.TableNumber,
		TableName:   pgtype.Text{String: req.TableName, Valid: req.TableName != ""},
		Section:     pgtype.Text{String: req.Section, Valid: req.Section != ""},
		Capacity:    pgtype.Int4{Int32: req.Capacity, Valid: true},
		IsActive:    pgtype.Bool{Bool: req.IsActive, Valid: true},
		Metadata:    []byte(req.Metadata),
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.UpdateTable(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// DeleteTable handles DELETE /api/restaurant/tables/:id
// @Summary      Delete restaurant table
// @Description  Deletes a restaurant table by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Table ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/tables/{id} [delete]
func (h *RestaurantHandler) DeleteTable(c *gin.Context) {
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

	resp := h.useCase.DeleteTable(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// === Menu Categories ===

// ListMenuCategories handles GET /api/restaurant/stores/:store_id/menu-categories
// @Summary      List menu categories
// @Description  Returns all menu categories for a given store.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        store_id     path      int     true   "Store ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/stores/{store_id}/menu-categories [get]
func (h *RestaurantHandler) ListMenuCategories(c *gin.Context) {
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

	resp := h.useCase.ListMenuCategories(c.Request.Context(), int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// GetMenuCategory handles GET /api/restaurant/menu-categories/:id
// @Summary      Get menu category
// @Description  Returns a single menu category by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Category ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/menu-categories/{id} [get]
func (h *RestaurantHandler) GetMenuCategory(c *gin.Context) {
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

	resp := h.useCase.GetMenuCategory(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// CreateMenuCategory handles POST /api/restaurant/menu-categories
// @Summary      Create menu category
// @Description  Creates a new menu category.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        body         body      CreateMenuCategoryRequest  true  "Category data"
// @Success      201          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/menu-categories [post]
func (h *RestaurantHandler) CreateMenuCategory(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateMenuCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	params := repository.CreateMenuCategoryParams{
		StoreID:          req.StoreID,
		ParentCategoryID: pgtype.Int4{Int32: utils.DerefInt32(req.ParentCategoryID), Valid: req.ParentCategoryID != nil},
		Name:             req.Name,
		Code:             req.Code,
		Description:      pgtype.Text{String: req.Description, Valid: req.Description != ""},
		CategoryLevel:    pgtype.Int4{Int32: req.CategoryLevel, Valid: true},
		DisplayOrder:     pgtype.Int4{Int32: req.DisplayOrder, Valid: true},
		Icon:             pgtype.Text{String: req.Icon, Valid: req.Icon != ""},
		ImageUrl:         pgtype.Text{String: req.ImageUrl, Valid: req.ImageUrl != ""},
		IsActive:         pgtype.Bool{Bool: req.IsActive, Valid: true},
		Metadata:         []byte(req.Metadata),
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.CreateMenuCategory(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// UpdateMenuCategory handles PUT /api/restaurant/menu-categories/:id
// @Summary      Update menu category
// @Description  Updates an existing menu category.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Category ID"
// @Param        body         body      CreateMenuCategoryRequest  true  "Updated category data"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/menu-categories/{id} [put]
func (h *RestaurantHandler) UpdateMenuCategory(c *gin.Context) {
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

	var req CreateMenuCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	params := repository.UpdateMenuCategoryParams{
		ID:               int32(id),
		ParentCategoryID: pgtype.Int4{Int32: utils.DerefInt32(req.ParentCategoryID), Valid: req.ParentCategoryID != nil},
		Name:             req.Name,
		Code:             req.Code,
		Description:      pgtype.Text{String: req.Description, Valid: req.Description != ""},
		CategoryLevel:    pgtype.Int4{Int32: req.CategoryLevel, Valid: true},
		DisplayOrder:     pgtype.Int4{Int32: req.DisplayOrder, Valid: true},
		Icon:             pgtype.Text{String: req.Icon, Valid: req.Icon != ""},
		ImageUrl:         pgtype.Text{String: req.ImageUrl, Valid: req.ImageUrl != ""},
		IsActive:         pgtype.Bool{Bool: req.IsActive, Valid: true},
		Metadata:         []byte(req.Metadata),
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.UpdateMenuCategory(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// DeleteMenuCategory handles DELETE /api/restaurant/menu-categories/:id
// @Summary      Delete menu category
// @Description  Deletes a menu category by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Category ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/menu-categories/{id} [delete]
func (h *RestaurantHandler) DeleteMenuCategory(c *gin.Context) {
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

	resp := h.useCase.DeleteMenuCategory(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// === Menu Items ===

// ListMenuItems handles GET /api/restaurant/menu-categories/:category_id/items
// @Summary      List menu items
// @Description  Returns all menu items for a given category.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        category_id  path      int     true   "Category ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/menu-categories/{category_id}/items [get]
func (h *RestaurantHandler) ListMenuItems(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	categoryID, err := strconv.ParseInt(c.Param("category_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid category_id", nil))
		return
	}

	resp := h.useCase.ListMenuItems(c.Request.Context(), int32(categoryID))
	c.JSON(resp.StatusCode, resp)
}

// GetMenuItem handles GET /api/restaurant/menu-items/:id
// @Summary      Get menu item
// @Description  Returns a single menu item by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Menu Item ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/menu-items/{id} [get]
func (h *RestaurantHandler) GetMenuItem(c *gin.Context) {
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

	resp := h.useCase.GetMenuItem(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// CreateMenuItem handles POST /api/restaurant/menu-items
// @Summary      Create menu item
// @Description  Creates a new menu item.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        body         body      CreateMenuItemRequest  true  "Menu item data"
// @Success      201          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/menu-items [post]
func (h *RestaurantHandler) CreateMenuItem(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateMenuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	basePrice, _ := repo.ParseNumeric(c.Request.Context(), req.BasePrice)

	params := repository.CreateMenuItemParams{
		StoreID:            req.StoreID,
		MenuCategoryID:     req.MenuCategoryID,
		ProductID:          pgtype.Int4{Int32: utils.DerefInt32(req.ProductID), Valid: req.ProductID != nil},
		RecipeID:           pgtype.Int4{Int32: utils.DerefInt32(req.RecipeID), Valid: req.RecipeID != nil},
		Name:               req.Name,
		ShortName:          pgtype.Text{String: req.ShortName, Valid: req.ShortName != ""},
		Description:        pgtype.Text{String: req.Description, Valid: req.Description != ""},
		ImageUrl:           pgtype.Text{String: req.ImageUrl, Valid: req.ImageUrl != ""},
		BasePrice:          basePrice,
		PreparationTimeMin: pgtype.Int4{Int32: req.PreparationTimeMin, Valid: true},
		TaxCategoryID:      pgtype.Int4{Int32: utils.DerefInt32(req.TaxCategoryID), Valid: req.TaxCategoryID != nil},
		IsAvailable:        pgtype.Bool{Bool: req.IsAvailable, Valid: true},
		IsActive:           pgtype.Bool{Bool: req.IsActive, Valid: true},
		DisplayOrder:       pgtype.Int4{Int32: req.DisplayOrder, Valid: true},
		Metadata:           []byte(req.Metadata),
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.CreateMenuItem(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// UpdateMenuItem handles PUT /api/restaurant/menu-items/:id
// @Summary      Update menu item
// @Description  Updates an existing menu item.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Menu Item ID"
// @Param        body         body      CreateMenuItemRequest  true  "Updated item data"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/menu-items/{id} [put]
func (h *RestaurantHandler) UpdateMenuItem(c *gin.Context) {
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

	var req CreateMenuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	basePrice, _ := repo.ParseNumeric(c.Request.Context(), req.BasePrice)

	params := repository.UpdateMenuItemParams{
		ID:                 int32(id),
		MenuCategoryID:     req.MenuCategoryID,
		ProductID:          pgtype.Int4{Int32: utils.DerefInt32(req.ProductID), Valid: req.ProductID != nil},
		RecipeID:           pgtype.Int4{Int32: utils.DerefInt32(req.RecipeID), Valid: req.RecipeID != nil},
		Name:               req.Name,
		ShortName:          pgtype.Text{String: req.ShortName, Valid: req.ShortName != ""},
		Description:        pgtype.Text{String: req.Description, Valid: req.Description != ""},
		ImageUrl:           pgtype.Text{String: req.ImageUrl, Valid: req.ImageUrl != ""},
		BasePrice:          basePrice,
		PreparationTimeMin: pgtype.Int4{Int32: req.PreparationTimeMin, Valid: true},
		TaxCategoryID:      pgtype.Int4{Int32: utils.DerefInt32(req.TaxCategoryID), Valid: req.TaxCategoryID != nil},
		IsAvailable:        pgtype.Bool{Bool: req.IsAvailable, Valid: true},
		IsActive:           pgtype.Bool{Bool: req.IsActive, Valid: true},
		DisplayOrder:       pgtype.Int4{Int32: req.DisplayOrder, Valid: true},
		Metadata:           []byte(req.Metadata),
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.UpdateMenuItem(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// DeleteMenuItem handles DELETE /api/restaurant/menu-items/:id
// @Summary      Delete menu item
// @Description  Deletes a menu item by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Menu Item ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/menu-items/{id} [delete]
func (h *RestaurantHandler) DeleteMenuItem(c *gin.Context) {
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

	resp := h.useCase.DeleteMenuItem(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// === Modifiers ===

// ListModifiers handles GET /api/restaurant/menu-items/:item_id/modifiers
// @Summary      List item modifiers
// @Description  Returns all modifiers for a given menu item.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        item_id      path      int     true   "Menu Item ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/menu-items/{item_id}/modifiers [get]
func (h *RestaurantHandler) ListModifiers(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	itemID, err := strconv.ParseInt(c.Param("item_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid item_id", nil))
		return
	}

	resp := h.useCase.ListModifiers(c.Request.Context(), int32(itemID))
	c.JSON(resp.StatusCode, resp)
}

// GetModifier handles GET /api/restaurant/modifiers/:id
// @Summary      Get item modifier
// @Description  Returns a single item modifier by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Modifier ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/modifiers/{id} [get]
func (h *RestaurantHandler) GetModifier(c *gin.Context) {
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

	resp := h.useCase.GetModifier(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// CreateModifier handles POST /api/restaurant/modifiers
// @Summary      Create item modifier
// @Description  Creates a new modifier for a menu item.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        body         body      CreateMenuItemModifierRequest  true  "Modifier data"
// @Success      201          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/modifiers [post]
func (h *RestaurantHandler) CreateModifier(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateMenuItemModifierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	priceAdj, _ := repo.ParseNumeric(c.Request.Context(), req.PriceAdjustment)

	params := repository.CreateMenuItemModifierParams{
		MenuItemID:      req.MenuItemID,
		ModifierName:    req.ModifierName,
		ModifierType:    req.ModifierType,
		PriceAdjustment: priceAdj,
		IsActive:        pgtype.Bool{Bool: req.IsActive, Valid: true},
		DisplayOrder:    pgtype.Int4{Int32: req.DisplayOrder, Valid: true},
		Metadata:        []byte(req.Metadata),
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.CreateModifier(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// UpdateModifier handles PUT /api/restaurant/modifiers/:id
// @Summary      Update item modifier
// @Description  Updates an existing item modifier.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Modifier ID"
// @Param        body         body      CreateMenuItemModifierRequest  true  "Updated modifier data"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/modifiers/{id} [put]
func (h *RestaurantHandler) UpdateModifier(c *gin.Context) {
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

	var req CreateMenuItemModifierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	priceAdj, _ := repo.ParseNumeric(c.Request.Context(), req.PriceAdjustment)

	params := repository.UpdateMenuItemModifierParams{
		ID:              int32(id),
		ModifierName:    req.ModifierName,
		ModifierType:    req.ModifierType,
		PriceAdjustment: priceAdj,
		IsActive:        pgtype.Bool{Bool: req.IsActive, Valid: true},
		DisplayOrder:    pgtype.Int4{Int32: req.DisplayOrder, Valid: true},
		Metadata:        []byte(req.Metadata),
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.UpdateModifier(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// DeleteModifier handles DELETE /api/restaurant/modifiers/:id
// @Summary      Delete item modifier
// @Description  Deletes an item modifier by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Modifier ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/modifiers/{id} [delete]
func (h *RestaurantHandler) DeleteModifier(c *gin.Context) {
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

	resp := h.useCase.DeleteModifier(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// === Menu ===

// GetFullMenu handles GET /api/restaurant/stores/:store_id/menu
// @Summary      Get full restaurant menu
// @Description  Returns the full menu for a store, with optional category filtering.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id           header    string  true   "Tenant identifier"
// @Param        store_id              path      int     true   "Store ID"
// @Param        category_id           query     int     false  "Filter by category ID"
// @Param        include_unavailable   query     bool    false  "Include unavailable items (default false)"
// @Success      200                   {object}  SuccessResponse
// @Failure      400                   {object}  ErrorResponse
// @Failure      401                   {object}  ErrorResponse
// @Failure      500                   {object}  ErrorResponse
// @Router       /api/restaurant/stores/{store_id}/menu [get]
func (h *RestaurantHandler) GetFullMenu(c *gin.Context) {
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

	var categoryID *int32
	if s := c.Query("category_id"); s != "" {
		id, _ := strconv.ParseInt(s, 10, 32)
		cid := int32(id)
		categoryID = &cid
	}
	includeUnavail := c.Query("include_unavailable") == "true"

	resp := h.useCase.GetFullMenu(c.Request.Context(), int32(storeID), categoryID, includeUnavail)
	c.JSON(resp.StatusCode, resp)
}

// === KDS ===

// GetKdsOrders handles GET /api/restaurant/stores/:store_id/kds
// @Summary      Get KDS orders
// @Description  Returns active orders for the Kitchen Display System.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string    true   "Tenant identifier"
// @Param        store_id     path      int       true   "Store ID"
// @Param        statuses     query     string    false  "Comma-separated statuses (default: pending,confirmed,preparing)"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/stores/{store_id}/kds [get]
func (h *RestaurantHandler) GetKdsOrders(c *gin.Context) {
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

	statusesStr := c.Query("statuses")
	var statuses []string
	if statusesStr != "" {
		statuses = strings.Split(statusesStr, ",")
	} else {
		statuses = []string{"pending", "confirmed", "preparing"}
	}

	resp := h.useCase.GetKdsOrders(c.Request.Context(), int32(storeID), statuses)
	c.JSON(resp.StatusCode, resp)
}

// === Orders ===

// CreateOrder handles POST /api/restaurant/orders
// @Summary      Create restaurant order
// @Description  Creates a new restaurant order (table-side, waiter, or counter).
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        body         body      CreateRestaurantOrderRequest  true  "Order data"
// @Success      201          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/orders [post]
func (h *RestaurantHandler) CreateOrder(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateRestaurantOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	subtotal, _ := repo.ParseNumeric(c.Request.Context(), req.Subtotal)
	discount, _ := repo.ParseNumeric(c.Request.Context(), req.DiscountAmount)
	tax, _ := repo.ParseNumeric(c.Request.Context(), req.TaxAmount)
	total, _ := repo.ParseNumeric(c.Request.Context(), req.TotalAmount)
	paid, _ := repo.ParseNumeric(c.Request.Context(), req.AmountPaid)
	change, _ := repo.ParseNumeric(c.Request.Context(), req.ChangeGiven)

	params := repository.CreateRestaurantOrderParams{
		StoreID:          req.StoreID,
		TableID:          pgtype.Int4{Int32: utils.DerefInt32(req.TableID), Valid: req.TableID != nil},
		CashierID:        pgtype.Int4{Int32: utils.DerefInt32(req.CashierID), Valid: req.CashierID != nil},
		CashierSessionID: pgtype.Int4{Int32: utils.DerefInt32(req.CashierSessionID), Valid: req.CashierSessionID != nil},
		CustomerID:       pgtype.Int4{Int32: utils.DerefInt32(req.CustomerID), Valid: req.CustomerID != nil},
		OrderNumber:      req.OrderNumber,
		OrderSource:      req.OrderSource,
		Status:           req.Status,
		Subtotal:         subtotal,
		DiscountAmount:   discount,
		TaxAmount:        tax,
		TotalAmount:      total,
		AmountPaid:       paid,
		ChangeGiven:      change,
		Notes:            pgtype.Text{String: req.Notes, Valid: req.Notes != ""},
		PosTransactionID: pgtype.Int4{Int32: utils.DerefInt32(req.PosTransactionID), Valid: req.PosTransactionID != nil},
		Metadata:         []byte(req.Metadata),
		OrderedAt:        pgtype.Timestamp{Time: time.Now(), Valid: true},
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.CreateOrder(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// UpdateOrder handles PUT /api/restaurant/orders/:id
// @Summary      Update restaurant order
// @Description  Updates an existing restaurant order.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Order ID"
// @Param        body         body      CreateRestaurantOrderRequest  true  "Updated order data"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/orders/{id} [put]
func (h *RestaurantHandler) UpdateOrder(c *gin.Context) {
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

	var req CreateRestaurantOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	subtotal, _ := repo.ParseNumeric(c.Request.Context(), req.Subtotal)
	discount, _ := repo.ParseNumeric(c.Request.Context(), req.DiscountAmount)
	tax, _ := repo.ParseNumeric(c.Request.Context(), req.TaxAmount)
	total, _ := repo.ParseNumeric(c.Request.Context(), req.TotalAmount)
	paid, _ := repo.ParseNumeric(c.Request.Context(), req.AmountPaid)
	change, _ := repo.ParseNumeric(c.Request.Context(), req.ChangeGiven)

	params := repository.UpdateRestaurantOrderParams{
		ID:               int32(id),
		TableID:          pgtype.Int4{Int32: utils.DerefInt32(req.TableID), Valid: req.TableID != nil},
		CashierID:        pgtype.Int4{Int32: utils.DerefInt32(req.CashierID), Valid: req.CashierID != nil},
		CashierSessionID: pgtype.Int4{Int32: utils.DerefInt32(req.CashierSessionID), Valid: req.CashierSessionID != nil},
		CustomerID:       pgtype.Int4{Int32: utils.DerefInt32(req.CustomerID), Valid: req.CustomerID != nil},
		Status:           req.Status,
		Subtotal:         subtotal,
		DiscountAmount:   discount,
		TaxAmount:        tax,
		TotalAmount:      total,
		AmountPaid:       paid,
		ChangeGiven:      change,
		Notes:            pgtype.Text{String: req.Notes, Valid: req.Notes != ""},
		PosTransactionID: pgtype.Int4{Int32: utils.DerefInt32(req.PosTransactionID), Valid: req.PosTransactionID != nil},
		Metadata:         []byte(req.Metadata),
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.UpdateOrder(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// DeleteOrder handles DELETE /api/restaurant/orders/:id
// @Summary      Delete restaurant order
// @Description  Deletes a restaurant order by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Order ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/orders/{id} [delete]
func (h *RestaurantHandler) DeleteOrder(c *gin.Context) {
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

	resp := h.useCase.DeleteOrder(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// GetOrder handles GET /api/restaurant/orders/:id
// @Summary      Get restaurant order
// @Description  Returns a single restaurant order with its items by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Order ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/orders/{id} [get]
func (h *RestaurantHandler) GetOrder(c *gin.Context) {
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

	resp := h.useCase.GetOrder(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// UpdateOrderStatus handles PATCH /api/restaurant/orders/:id/status
// @Summary      Update order status
// @Description  Updates the status of a restaurant order.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Order ID"
// @Param        body         body      UpdateStatusRequest  true  "Status data"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/orders/{id}/status [patch]
func (h *RestaurantHandler) UpdateOrderStatus(c *gin.Context) {
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

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.UpdateOrderStatus(c.Request.Context(), int32(id), req.Status)
	c.JSON(resp.StatusCode, resp)
}

// SettleOrder handles POST /api/restaurant/orders/:id/settle
// @Summary      Settle restaurant order
// @Description  Marks a restaurant order as paid and links it to a POS transaction.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Order ID"
// @Param        body         body      SettleOrderRequest  true  "Settlement data"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/orders/{id}/settle [post]
func (h *RestaurantHandler) SettleOrder(c *gin.Context) {
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

	var req SettleOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.SettleOrder(c.Request.Context(), int32(id), req.PosTransactionID)
	c.JSON(resp.StatusCode, resp)
}

// CreateOnlineOrder handles POST /api/restaurant/orders/online
// @Summary      Create online order
// @Description  Creates a new online restaurant order.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        body         body      CreateOnlineOrderRequest  true  "Online order data"
// @Success      201          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/orders/online [post]
func (h *RestaurantHandler) CreateOnlineOrder(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateOnlineOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	items := make([]repository.CreateRestaurantOrderItemParams, len(req.Items))
	for i, item := range req.Items {
		qty, _ := repo.ParseNumeric(c.Request.Context(), item.Quantity)
		price, _ := repo.ParseNumeric(c.Request.Context(), item.UnitPrice)
		modTotal, _ := repo.ParseNumeric(c.Request.Context(), item.ModifiersTotal)
		disc, _ := repo.ParseNumeric(c.Request.Context(), item.DiscountAmount)
		tax, _ := repo.ParseNumeric(c.Request.Context(), item.TaxAmount)
		sub, _ := repo.ParseNumeric(c.Request.Context(), item.Subtotal)

		items[i] = repository.CreateRestaurantOrderItemParams{
			MenuItemID:        item.MenuItemID,
			Quantity:          qty,
			UnitPrice:         price,
			ModifiersSnapshot: []byte(item.ModifiersSnapshot),
			ModifiersTotal:    modTotal,
			DiscountAmount:    disc,
			TaxAmount:         tax,
			Subtotal:          sub,
			LineNumber:        pgtype.Int4{Int32: item.LineNumber, Valid: true},
			Notes:             pgtype.Text{String: item.Notes, Valid: item.Notes != ""},
			Status:            pgtype.Text{String: item.Status, Valid: item.Status != ""},
			Metadata:          []byte(item.Metadata),
		}
		if item.ModifiersSnapshot == "" {
			items[i].ModifiersSnapshot = []byte("[]")
		}
		if item.Metadata == "" {
			items[i].Metadata = []byte("{}")
		}
	}

	resp := h.useCase.CreateOnlineOrder(c.Request.Context(), req.StoreID, req.CustomerID, items)
	c.JSON(resp.StatusCode, resp)
}

// === Order Items ===

// GetOrderItem handles GET /api/restaurant/order-items/:id
// @Summary      Get order item
// @Description  Returns a single restaurant order item by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Order Item ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/order-items/{id} [get]
func (h *RestaurantHandler) GetOrderItem(c *gin.Context) {
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

	resp := h.useCase.GetOrderItem(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// UpdateOrderItem handles PUT /api/restaurant/order-items/:id
// @Summary      Update order item
// @Description  Updates an existing restaurant order item.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Order Item ID"
// @Param        body         body      CreateRestaurantOrderItemRequest  true  "Updated item data"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/order-items/{id} [put]
func (h *RestaurantHandler) UpdateOrderItem(c *gin.Context) {
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

	var req CreateRestaurantOrderItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	qty, _ := repo.ParseNumeric(c.Request.Context(), req.Quantity)
	price, _ := repo.ParseNumeric(c.Request.Context(), req.UnitPrice)
	modTotal, _ := repo.ParseNumeric(c.Request.Context(), req.ModifiersTotal)
	disc, _ := repo.ParseNumeric(c.Request.Context(), req.DiscountAmount)
	tax, _ := repo.ParseNumeric(c.Request.Context(), req.TaxAmount)
	sub, _ := repo.ParseNumeric(c.Request.Context(), req.Subtotal)

	params := repository.UpdateRestaurantOrderItemParams{
		ID:                int32(id),
		MenuItemID:        req.MenuItemID,
		Quantity:          qty,
		UnitPrice:         price,
		ModifiersSnapshot: []byte(req.ModifiersSnapshot),
		ModifiersTotal:    modTotal,
		DiscountAmount:    disc,
		TaxAmount:         tax,
		Subtotal:          sub,
		LineNumber:        pgtype.Int4{Int32: req.LineNumber, Valid: true},
		Notes:             pgtype.Text{String: req.Notes, Valid: req.Notes != ""},
		Status:            pgtype.Text{String: req.Status, Valid: req.Status != ""},
		Metadata:          []byte(req.Metadata),
	}
	if req.ModifiersSnapshot == "" {
		params.ModifiersSnapshot = []byte("[]")
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.UpdateOrderItem(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// DeleteOrderItem handles DELETE /api/restaurant/order-items/:id
// @Summary      Delete order item
// @Description  Deletes a restaurant order item by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Order Item ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/order-items/{id} [delete]
func (h *RestaurantHandler) DeleteOrderItem(c *gin.Context) {
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

	resp := h.useCase.DeleteOrderItem(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// === Waste ===

// CreateWasteLog handles POST /api/restaurant/waste-logs
// @Summary      Create waste log
// @Description  Logs wasted ingredients or menu items.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        body         body      CreateWasteLogRequest  true  "Waste data"
// @Success      201          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/waste-logs [post]
func (h *RestaurantHandler) CreateWasteLog(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateWasteLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	qty, _ := repo.ParseNumeric(c.Request.Context(), req.Quantity)
	uCost, _ := repo.ParseNumeric(c.Request.Context(), req.UnitCost)
	tCost, _ := repo.ParseNumeric(c.Request.Context(), req.TotalCost)

	params := repository.CreateWasteLogParams{
		StoreID:     req.StoreID,
		ProductID:   pgtype.Int4{Int32: utils.DerefInt32(req.ProductID), Valid: req.ProductID != nil},
		MenuItemID:  pgtype.Int4{Int32: utils.DerefInt32(req.MenuItemID), Valid: req.MenuItemID != nil},
		RecipeID:    pgtype.Int4{Int32: utils.DerefInt32(req.RecipeID), Valid: req.RecipeID != nil},
		WasteSource: req.WasteSource,
		Quantity:    qty,
		UomID:       pgtype.Int4{Int32: utils.DerefInt32(req.UomID), Valid: req.UomID != nil},
		UnitCost:    uCost,
		TotalCost:   tCost,
		Reason:      pgtype.Text{String: req.Reason, Valid: req.Reason != ""},
		LoggedBy:    pgtype.Int4{Int32: utils.DerefInt32(req.LoggedBy), Valid: req.LoggedBy != nil},
		OrderID:     pgtype.Int4{Int32: utils.DerefInt32(req.OrderID), Valid: req.OrderID != nil},
		WastedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
		Metadata:    []byte(req.Metadata),
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.CreateWasteLog(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// GetWasteLog handles GET /api/restaurant/waste-logs/:id
// @Summary      Get waste log
// @Description  Returns a single waste log entry by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Waste Log ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/waste-logs/{id} [get]
func (h *RestaurantHandler) GetWasteLog(c *gin.Context) {
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

	resp := h.useCase.GetWasteLog(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// UpdateWasteLog handles PUT /api/restaurant/waste-logs/:id
// @Summary      Update waste log
// @Description  Updates an existing waste log entry.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Waste Log ID"
// @Param        body         body      CreateWasteLogRequest  true  "Updated waste data"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/waste-logs/{id} [put]
func (h *RestaurantHandler) UpdateWasteLog(c *gin.Context) {
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

	var req CreateWasteLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	qty, _ := repo.ParseNumeric(c.Request.Context(), req.Quantity)
	uCost, _ := repo.ParseNumeric(c.Request.Context(), req.UnitCost)
	tCost, _ := repo.ParseNumeric(c.Request.Context(), req.TotalCost)

	params := repository.UpdateWasteLogParams{
		ID:          int32(id),
		ProductID:   pgtype.Int4{Int32: utils.DerefInt32(req.ProductID), Valid: req.ProductID != nil},
		MenuItemID:  pgtype.Int4{Int32: utils.DerefInt32(req.MenuItemID), Valid: req.MenuItemID != nil},
		RecipeID:    pgtype.Int4{Int32: utils.DerefInt32(req.RecipeID), Valid: req.RecipeID != nil},
		WasteSource: req.WasteSource,
		Quantity:    qty,
		UomID:       pgtype.Int4{Int32: utils.DerefInt32(req.UomID), Valid: req.UomID != nil},
		UnitCost:    uCost,
		TotalCost:   tCost,
		Reason:      pgtype.Text{String: req.Reason, Valid: req.Reason != ""},
		LoggedBy:    pgtype.Int4{Int32: utils.DerefInt32(req.LoggedBy), Valid: req.LoggedBy != nil},
		OrderID:     pgtype.Int4{Int32: utils.DerefInt32(req.OrderID), Valid: req.OrderID != nil},
		WastedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
		Metadata:    []byte(req.Metadata),
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.UpdateWasteLog(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// DeleteWasteLog handles DELETE /api/restaurant/waste-logs/:id
// @Summary      Delete waste log
// @Description  Deletes a waste log entry by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Waste Log ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/waste-logs/{id} [delete]
func (h *RestaurantHandler) DeleteWasteLog(c *gin.Context) {
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

	resp := h.useCase.DeleteWasteLog(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// GetWasteReport handles GET /api/restaurant/stores/:store_id/waste-report
// @Summary      Get waste report
// @Description  Returns a waste report for a store within a date range.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        store_id      path      int     true   "Store ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/restaurant/stores/{store_id}/waste-report [get]
func (h *RestaurantHandler) GetWasteReport(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeID, _ := strconv.ParseInt(c.Param("store_id"), 10, 32)

	resp := h.useCase.GetWasteReport(c.Request.Context(), int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// === Recipes ===

// ListRecipes handles GET /api/restaurant/stores/:store_id/recipes
// @Summary      List recipes
// @Description  Returns all recipes for a given organization.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        store_id     path      int     true   "Store ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/stores/{store_id}/recipes [get]
func (h *RestaurantHandler) ListRecipes(c *gin.Context) {
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

	store, err := repo.GetStore(c.Request.Context(), int32(storeID))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewResponse(utils.CodeNotFound, "store not found", nil))
		return
	}

	resp := h.useCase.ListRecipes(c.Request.Context(), store.OrganizationID)
	c.JSON(resp.StatusCode, resp)
}

// GetRecipe handles GET /api/restaurant/recipes/:id
// @Summary      Get recipe
// @Description  Returns a single recipe with ingredients and cost by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Recipe ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/recipes/{id} [get]
func (h *RestaurantHandler) GetRecipe(c *gin.Context) {
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

	resp := h.useCase.GetRecipe(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// CreateRecipe handles POST /api/restaurant/recipes
// @Summary      Create recipe
// @Description  Creates a new recipe header.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        body         body      CreateRecipeRequest  true  "Recipe data"
// @Success      201          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/recipes [post]
func (h *RestaurantHandler) CreateRecipe(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateRecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	yieldQty, _ := repo.ParseNumeric(c.Request.Context(), req.YieldQuantity)

	params := repository.CreateRecipeParams{
		OrganizationID:      req.OrganizationID,
		RecipeCode:          req.RecipeCode,
		RecipeName:          req.RecipeName,
		Description:         pgtype.Text{String: req.Description, Valid: req.Description != ""},
		FinishedProductID:   pgtype.Int4{Int32: utils.DerefInt32(req.FinishedProductID), Valid: req.FinishedProductID != nil},
		YieldQuantity:       yieldQty,
		YieldUomID:          pgtype.Int4{Int32: utils.DerefInt32(req.YieldUomID), Valid: req.YieldUomID != nil},
		PreparationSteps:    pgtype.Text{String: req.PreparationSteps, Valid: req.PreparationSteps != ""},
		PreparationTimeMin:  pgtype.Int4{Int32: req.PreparationTimeMin, Valid: true},
		CookingTimeMin:      pgtype.Int4{Int32: req.CookingTimeMin, Valid: true},
		IsActive:            pgtype.Bool{Bool: req.IsActive, Valid: true},
		Metadata:            []byte(req.Metadata),
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.CreateRecipe(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// UpdateRecipe handles PUT /api/restaurant/recipes/:id
// @Summary      Update recipe
// @Description  Updates an existing recipe header.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Recipe ID"
// @Param        body         body      CreateRecipeRequest  true  "Updated recipe data"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/recipes/{id} [put]
func (h *RestaurantHandler) UpdateRecipe(c *gin.Context) {
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

	var req CreateRecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	yieldQty, _ := repo.ParseNumeric(c.Request.Context(), req.YieldQuantity)

	params := repository.UpdateRecipeParams{
		ID:                  int32(id),
		RecipeCode:          req.RecipeCode,
		RecipeName:          req.RecipeName,
		Description:         pgtype.Text{String: req.Description, Valid: req.Description != ""},
		FinishedProductID:   pgtype.Int4{Int32: utils.DerefInt32(req.FinishedProductID), Valid: req.FinishedProductID != nil},
		YieldQuantity:       yieldQty,
		YieldUomID:          pgtype.Int4{Int32: utils.DerefInt32(req.YieldUomID), Valid: req.YieldUomID != nil},
		PreparationSteps:    pgtype.Text{String: req.PreparationSteps, Valid: req.PreparationSteps != ""},
		PreparationTimeMin:  pgtype.Int4{Int32: req.PreparationTimeMin, Valid: true},
		CookingTimeMin:      pgtype.Int4{Int32: req.CookingTimeMin, Valid: true},
		IsActive:            pgtype.Bool{Bool: req.IsActive, Valid: true},
		Metadata:            []byte(req.Metadata),
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.UpdateRecipe(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// DeleteRecipe handles DELETE /api/restaurant/recipes/:id
// @Summary      Delete recipe
// @Description  Deletes a recipe and its ingredients by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Recipe ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/recipes/{id} [delete]
func (h *RestaurantHandler) DeleteRecipe(c *gin.Context) {
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

	resp := h.useCase.DeleteRecipe(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// AddRecipeIngredient handles POST /api/restaurant/recipes/:id/ingredients
// @Summary      Add recipe ingredient
// @Description  Adds an ingredient line to a recipe.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Recipe ID"
// @Param        body         body      CreateRecipeIngredientRequest  true  "Ingredient data"
// @Success      201          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/recipes/{id}/ingredients [post]
func (h *RestaurantHandler) AddRecipeIngredient(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, _ := strconv.ParseInt(c.Param("id"), 10, 32)
	var req CreateRecipeIngredientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	qty, _ := repo.ParseNumeric(c.Request.Context(), req.Quantity)

	params := repository.CreateRecipeIngredientParams{
		RecipeID:         int32(id),
		ProductID:        req.ProductID,
		ProductVariantID: pgtype.Int4{Int32: utils.DerefInt32(req.ProductVariantID), Valid: req.ProductVariantID != nil},
		Quantity:         qty,
		UomID:            pgtype.Int4{Int32: utils.DerefInt32(req.UomID), Valid: req.UomID != nil},
		IsOptional:       pgtype.Bool{Bool: req.IsOptional, Valid: true},
		IsByproduct:      pgtype.Bool{Bool: req.IsByproduct, Valid: true},
		LineNumber:       pgtype.Int4{Int32: req.LineNumber, Valid: true},
		Metadata:         []byte(req.Metadata),
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.AddRecipeIngredient(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// === Recipe Ingredients ===

// GetRecipeIngredient handles GET /api/restaurant/recipe-ingredients/:id
// @Summary      Get recipe ingredient
// @Description  Returns a single recipe ingredient line by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Ingredient ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/recipe-ingredients/{id} [get]
func (h *RestaurantHandler) GetRecipeIngredient(c *gin.Context) {
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

	resp := h.useCase.GetRecipeIngredient(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// UpdateRecipeIngredient handles PUT /api/restaurant/recipe-ingredients/:id
// @Summary      Update recipe ingredient
// @Description  Updates an existing recipe ingredient line.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Ingredient ID"
// @Param        body         body      CreateRecipeIngredientRequest  true  "Updated ingredient data"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/recipe-ingredients/{id} [put]
func (h *RestaurantHandler) UpdateRecipeIngredient(c *gin.Context) {
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

	var req CreateRecipeIngredientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	qty, _ := repo.ParseNumeric(c.Request.Context(), req.Quantity)

	params := repository.UpdateRecipeIngredientParams{
		ID:               int32(id),
		ProductID:        req.ProductID,
		ProductVariantID: pgtype.Int4{Int32: utils.DerefInt32(req.ProductVariantID), Valid: req.ProductVariantID != nil},
		Quantity:         qty,
		UomID:            pgtype.Int4{Int32: utils.DerefInt32(req.UomID), Valid: req.UomID != nil},
		IsOptional:       pgtype.Bool{Bool: req.IsOptional, Valid: true},
		IsByproduct:      pgtype.Bool{Bool: req.IsByproduct, Valid: true},
		LineNumber:       pgtype.Int4{Int32: req.LineNumber, Valid: true},
		Metadata:         []byte(req.Metadata),
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.UpdateRecipeIngredient(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// DeleteRecipeIngredient handles DELETE /api/restaurant/recipe-ingredients/:id
// @Summary      Delete recipe ingredient
// @Description  Deletes a recipe ingredient line by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Ingredient ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/recipe-ingredients/{id} [delete]
func (h *RestaurantHandler) DeleteRecipeIngredient(c *gin.Context) {
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

	resp := h.useCase.DeleteRecipeIngredient(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// === Kiosk ===

// CreateKioskSession handles POST /api/restaurant/kiosk/sessions
// @Summary      Create kiosk session
// @Description  Starts a new self-service kiosk session.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        body         body      CreateKioskSessionRequest  true  "Session data"
// @Success      201          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/kiosk/sessions [post]
func (h *RestaurantHandler) CreateKioskSession(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateKioskSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	params := repository.CreateKioskSessionParams{
		PosTerminalID: req.PosTerminalID,
		StoreID:       req.StoreID,
		SessionToken:  req.SessionToken,
		Status:        pgtype.Text{String: req.Status, Valid: req.Status != ""},
		OpenedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
		Metadata:      []byte(req.Metadata),
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.CreateKioskSession(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// UpdateKioskSession handles PUT /api/restaurant/kiosk/sessions/:id
// @Summary      Update kiosk session
// @Description  Updates an existing kiosk session.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Session ID"
// @Param        body         body      CreateKioskSessionRequest  true  "Updated session data"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/kiosk/sessions/{id} [put]
func (h *RestaurantHandler) UpdateKioskSession(c *gin.Context) {
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

	var req CreateKioskSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	params := repository.UpdateKioskSessionParams{
		ID:       int32(id),
		Status:   pgtype.Text{String: req.Status, Valid: req.Status != ""},
		Metadata: []byte(req.Metadata),
	}
	if req.Metadata == "" {
		params.Metadata = []byte("{}")
	}

	resp := h.useCase.UpdateKioskSession(c.Request.Context(), params)
	c.JSON(resp.StatusCode, resp)
}

// DeleteKioskSession handles DELETE /api/restaurant/kiosk/sessions/:id
// @Summary      Delete kiosk session
// @Description  Deletes a kiosk session by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Session ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/kiosk/sessions/{id} [delete]
func (h *RestaurantHandler) DeleteKioskSession(c *gin.Context) {
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

	resp := h.useCase.DeleteKioskSession(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// GetKioskSession handles GET /api/restaurant/kiosk/sessions/:token
// @Summary      Get kiosk session by token
// @Description  Returns kiosk session details by session token.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        token        path      string  true   "Session Token"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/kiosk/sessions/{token} [get]
func (h *RestaurantHandler) GetKioskSession(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	token := c.Param("token")
	resp := h.useCase.GetKioskSession(c.Request.Context(), token)
	c.JSON(resp.StatusCode, resp)
}

// GetKioskSessionByID handles GET /api/restaurant/kiosk/sessions/id/:id
// @Summary      Get kiosk session by ID
// @Description  Returns kiosk session details by ID.
// @Tags         restaurant
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true   "Tenant identifier"
// @Param        id           path      int     true   "Session ID"
// @Success      200          {object}  SuccessResponse
// @Failure      400          {object}  ErrorResponse
// @Failure      401          {object}  ErrorResponse
// @Failure      404          {object}  ErrorResponse
// @Failure      500          {object}  ErrorResponse
// @Router       /api/restaurant/kiosk/sessions/id/{id} [get]
func (h *RestaurantHandler) GetKioskSessionByID(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, _ := strconv.ParseInt(c.Param("id"), 10, 32)
	resp := h.useCase.GetKioskSessionByID(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}
