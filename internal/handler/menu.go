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
)

type MenuHandler struct {
	useCase *usecase.MenuUseCase
}

func NewMenuHandler(uc *usecase.MenuUseCase) *MenuHandler {
	return &MenuHandler{
		useCase: uc,
	}
}

func (h *MenuHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateMenu handles POST /menus
// @Summary      Create a new menu
// @Description  Create a menu under a module with optional parent menu
// @Tags         menus
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header  string            true  "Tenant identifier"
// @Param        Authorization header  string            true  "Bearer token"
// @Param        menu  body  handler.CreateMenuRequest  true  "Menu data"
// @Success      201  {object}  handler.MenuResponse
// @Failure      400  {object}  handler.ErrorResponse
// @Failure      401  {object}  handler.ErrorResponse
// @Failure      500  {object}  handler.ErrorResponse
// @Router       /api/menus [post]
func (h *MenuHandler) CreateMenu(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req struct {
		ModuleID     int32       `json:"module_id" binding:"required"`
		ParentMenuID *int32      `json:"parent_menu_id"`
		Name         string      `json:"name" binding:"required"`
		Code         string      `json:"code" binding:"required"`
		RoutePath    *string     `json:"route_path"`
		Icon         *string     `json:"icon"`
		DisplayOrder *int32      `json:"display_order"`
		IsActive     bool        `json:"is_active"`
		Metadata     interface{} `json:"metadata"`
	}

	if err := c.BindJSON(&req); err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid request body", err.Error())
		c.JSON(resp.StatusCode, resp)
		return
	}
	metadataBytes, err := json.Marshal(req.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(
			utils.CodeError,
			"failed to process metadata",
			nil,
		))
		return
	}

	resp := h.useCase.CreateMenu(
		c.Request.Context(),
		req.ModuleID,
		req.ParentMenuID,
		req.Name,
		req.Code,
		req.RoutePath,
		req.Icon,
		req.DisplayOrder,
		req.IsActive,
		metadataBytes,
	)

	c.JSON(resp.StatusCode, resp)
}

// GetMenu handles GET /menus/{id}
// @Summary      Get menu by ID
// @Description  Retrieve a single menu by its ID
// @Tags         menus
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header  string  true  "Tenant identifier"
// @Param        Authorization header  string  true  "Bearer token"
// @Param        id             path    int     true  "Menu ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/menus/{id} [get]
func (h *MenuHandler) GetMenu(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid menu id", nil)
		c.JSON(resp.StatusCode, resp)
		return
	}

	resp := h.useCase.GetMenu(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// GetMenuByCode handles GET /menus/by-code
// @Summary      Get menu by code
// @Description  Retrieve a menu using module ID and menu code
// @Tags         menus
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header  string  true  "Tenant identifier"
// @Param        Authorization header  string  true  "Bearer token"
// @Param        module_id     query   int     true  "Module ID"
// @Param        code          query   string  true  "Menu code"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/menus/by-code [get]
func (h *MenuHandler) GetMenuByCode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	moduleID, err := strconv.ParseInt(c.Query("module_id"), 10, 32)
	if err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid module id", nil)
		c.JSON(resp.StatusCode, resp)
		return
	}

	code := c.Query("code")
	if code == "" {
		resp := utils.NewResponse(utils.CodeBadReq, "menu code is required", nil)
		c.JSON(resp.StatusCode, resp)
		return
	}

	resp := h.useCase.GetMenuByCode(c.Request.Context(), int32(moduleID), code)
	c.JSON(resp.StatusCode, resp)
}

// ListMenus handles GET /menus
// @Summary      List all menus
// @Description  Retrieve all menus across modules
// @Tags         menus
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header  string  true  "Tenant identifier"
// @Param        Authorization header  string  true  "Bearer token"
// @Success      200  {object}  SuccessResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/menus [get]
func (h *MenuHandler) ListMenus(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListMenus(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListMenusByModule handles GET /modules/{moduleId}/menus
// @Summary      List menus by module
// @Description  Retrieve all menus for a specific module
// @Tags         menus
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header  string  true  "Tenant identifier"
// @Param        Authorization header  string  true  "Bearer token"
// @Param        moduleId      path    int     true  "Module ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /api/modules/{moduleId}/menus [get]
func (h *MenuHandler) ListMenusByModule(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	moduleID, err := strconv.ParseInt(c.Param("moduleId"), 10, 32)
	if err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid module id", nil)
		c.JSON(resp.StatusCode, resp)
		return
	}

	resp := h.useCase.ListMenusByModule(c.Request.Context(), int32(moduleID))
	c.JSON(resp.StatusCode, resp)
}

// ListActiveMenusByModule handles GET /modules/{moduleId}/menus/active
// @Summary      List active menus by module
// @Description  Retrieve only active menus for a module
// @Tags         menus
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header  string  true  "Tenant identifier"
// @Param        Authorization header  string  true  "Bearer token"
// @Param        moduleId      path    int     true  "Module ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /api/modules/{moduleId}/menus/active [get]
func (h *MenuHandler) ListActiveMenusByModule(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	moduleID, err := strconv.ParseInt(c.Param("moduleId"), 10, 32)
	if err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid module id", nil)
		c.JSON(resp.StatusCode, resp)
		return
	}

	resp := h.useCase.ListActiveMenusByModule(c.Request.Context(), int32(moduleID))
	c.JSON(resp.StatusCode, resp)
}

// ListMenusByParent handles GET /menus/parent/{parentId}
// @Summary      List menus by parent menu
// @Description  Retrieve child menus under a parent menu
// @Tags         menus
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header  string  true  "Tenant identifier"
// @Param        Authorization header  string  true  "Bearer token"
// @Param        parentId      path    int     true  "Parent menu ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /api/menus/parent/{parentId} [get]
func (h *MenuHandler) ListMenusByParent(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	parentID64, err := strconv.ParseInt(c.Param("parentId"), 10, 32)
	if err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid parent menu id", nil)
		c.JSON(resp.StatusCode, resp)
		return
	}

	parentID := int32(parentID64)

	resp := h.useCase.ListMenusByParent(c.Request.Context(), &parentID)
	c.JSON(resp.StatusCode, resp)
}

// UpdateMenu handles PUT /menus/{id}
// @Summary      Update a menu
// @Description  Update menu details by ID
// @Tags         menus
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header  string  true  "Tenant identifier"
// @Param        Authorization header  string  true  "Bearer token"
// @Param        id            path    int     true  "Menu ID"
// @Param        menu          body    object  true  "Updated menu data"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/menus/{id} [put]
func (h *MenuHandler) UpdateMenu(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid menu id", nil)
		c.JSON(resp.StatusCode, resp)
		return
	}

	var req struct {
		ParentMenuID *int32      `json:"parent_menu_id"`
		Name         string      `json:"name"`
		RoutePath    *string     `json:"route_path"`
		Icon         *string     `json:"icon"`
		DisplayOrder *int32      `json:"display_order"`
		IsActive     bool        `json:"is_active"`
		Metadata     interface{} `json:"metadata"`
	}

	if err := c.BindJSON(&req); err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid request body", err.Error())
		c.JSON(resp.StatusCode, resp)
		return
	}

	metadataBytes, err := json.Marshal(req.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(
			utils.CodeError,
			"failed to process metadata",
			nil,
		))
		return
	}

	resp := h.useCase.UpdateMenu(
		c.Request.Context(),
		int32(id),
		req.ParentMenuID,
		req.Name,
		req.RoutePath,
		req.Icon,
		req.DisplayOrder,
		req.IsActive,
		metadataBytes,
	)

	c.JSON(resp.StatusCode, resp)
}

// ToggleMenuActive handles PATCH /menus/{id}/active
// @Summary      Toggle menu active status
// @Description  Enable or disable a menu
// @Tags         menus
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header  string  true  "Tenant identifier"
// @Param        Authorization header  string  true  "Bearer token"
// @Param        id            path    int     true  "Menu ID"
// @Param        status        body    object  true  "Active status"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /api/menus/{id}/active [patch]
func (h *MenuHandler) ToggleMenuActive(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid menu id", nil)
		c.JSON(resp.StatusCode, resp)
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}

	if err := c.BindJSON(&req); err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid request body", err.Error())
		c.JSON(resp.StatusCode, resp)
		return
	}

	resp := h.useCase.ToggleMenuActive(c.Request.Context(), int32(id), req.IsActive)
	c.JSON(resp.StatusCode, resp)
}

// DeleteMenu handles DELETE /menus/{id}
// @Summary      Delete a menu
// @Description  Delete a menu by ID
// @Tags         menus
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header  string  true  "Tenant identifier"
// @Param        Authorization header  string  true  "Bearer token"
// @Param        id            path    int     true  "Menu ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/menus/{id} [delete]
func (h *MenuHandler) DeleteMenu(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid menu id", nil)
		c.JSON(resp.StatusCode, resp)
		return
	}

	resp := h.useCase.DeleteMenu(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}
