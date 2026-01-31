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

type SubmenuHandler struct {
	useCase *usecase.SubmenuUseCase
}

func NewSubmenuHandler(uc *usecase.SubmenuUseCase) *SubmenuHandler {
	return &SubmenuHandler{useCase: uc}
}

func (h *SubmenuHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateSubmenu
// @Summary      Create a new submenu
// @Description  Create a submenu under a menu with optional parent submenu
// @Tags         submenus
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header  string                   true  "Tenant identifier"
// @Param        Authorization header  string                   true  "Bearer token"
// @Param        submenu       body    handler.CreateSubmenuRequest  true  "Submenu data"
// @Success      201  {object}  handler.SubmenuResponse
// @Failure      400  {object}  handler.ErrorResponse
// @Failure      401  {object}  handler.ErrorResponse
// @Failure      500  {object}  handler.ErrorResponse
// @Router       /api/submenus [post]
func (h *SubmenuHandler) CreateSubmenu(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req struct {
		MenuID          int32       `json:"menu_id" binding:"required"`
		ParentSubmenuID *int32      `json:"parent_submenu_id"`
		Name            string      `json:"name" binding:"required"`
		Code            string      `json:"code" binding:"required"`
		RoutePath       *string     `json:"route_path"`
		Icon            *string     `json:"icon"`
		DisplayOrder    *int32      `json:"display_order"`
		IsActive        bool        `json:"is_active"`
		Metadata        interface{} `json:"metadata"`
	}

	if err := c.BindJSON(&req); err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid request body", err.Error())
		c.JSON(resp.StatusCode, resp)
		return
	}

	metadataBytes, err := json.Marshal(req.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "failed to process metadata", nil))
		return
	}

	resp := h.useCase.CreateSubmenu(
		c.Request.Context(),
		req.MenuID,
		req.ParentSubmenuID,
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

// GetSubmenu handles GET /submenus/{id}
// @Summary      Get submenu by ID
// @Description  Retrieve a single submenu by its ID
// @Tags         submenus
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header  string  true  "Tenant identifier"
// @Param        Authorization header  string  true  "Bearer token"
// @Param        id            path    int     true  "Submenu ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/submenus/{id} [get]
func (h *SubmenuHandler) GetSubmenu(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid submenu id", nil)
		c.JSON(resp.StatusCode, resp)
		return
	}

	resp := h.useCase.GetSubmenu(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// ListSubmenusByMenu handles GET /submenus/by-menu/{menu_id}
// @Summary      List submenus by menu
// @Description  Retrieve all submenus under a specific menu
// @Tags         submenus
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header  string  true  "Tenant identifier"
// @Param        Authorization header  string  true  "Bearer token"
// @Param        menu_id       query   int     true  "Menu ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/submenus/by-menu/{menu_id} [get]
func (h *SubmenuHandler) ListSubmenusByMenu(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	menuID, err := strconv.ParseInt(c.Param("menu_id"), 10, 32)
	if err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid menu id", nil)
		c.JSON(resp.StatusCode, resp)
		return
	}

	resp := h.useCase.ListSubmenusByMenu(c.Request.Context(), int32(menuID))
	c.JSON(resp.StatusCode, resp)
}

// UpdateSubmenu handles PUT /submenus/{id}
// @Summary      Update submenu
// @Description  Update submenu details
// @Tags         submenus
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header  string                   true  "Tenant identifier"
// @Param        Authorization header  string                   true  "Bearer token"
// @Param        id            path    int                      true  "Submenu ID"
// @Param        submenu       body    handler.UpdateSubmenuRequest  true  "Submenu data"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/submenus/{id} [put]
func (h *SubmenuHandler) UpdateSubmenu(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid submenu id", nil)
		c.JSON(resp.StatusCode, resp)
		return
	}

	var req struct {
		ParentSubmenuID *int32      `json:"parent_submenu_id"`
		Name            string      `json:"name"`
		RoutePath       *string     `json:"route_path"`
		Icon            *string     `json:"icon"`
		DisplayOrder    *int32      `json:"display_order"`
		IsActive        bool        `json:"is_active"`
		Metadata        interface{} `json:"metadata"`
	}

	if err := c.BindJSON(&req); err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid request body", err.Error())
		c.JSON(resp.StatusCode, resp)
		return
	}

	metadataBytes, err := json.Marshal(req.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "failed to process metadata", nil))
		return
	}

	resp := h.useCase.UpdateSubmenu(
		c.Request.Context(),
		int32(id),
		req.ParentSubmenuID,
		req.Name,
		req.RoutePath,
		req.Icon,
		req.DisplayOrder,
		req.IsActive,
		metadataBytes,
	)

	c.JSON(resp.StatusCode, resp)
}

// ToggleSubmenuActive handles PATCH /submenus/{id}/toggle
// @Summary      Toggle submenu active status
// @Description  Activate or deactivate a submenu
// @Tags         submenus
// @Accept       json
// @Produce      json
// @Security     BearerAuth
//
//	@Param        x-tenant-id   header  string                   true  "Tenant identifier"
//
// @Param        Authorization header  string                   true  "Bearer token"
// @Param        id            path    int                      true  "Submenu ID"
// @Param        body          body    handler.ToggleSubmenuActiveRequest  true  "Active status"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/submenus/{id}/toggle [patch]
func (h *SubmenuHandler) ToggleSubmenuActive(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid submenu id", nil)
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

	resp := h.useCase.ToggleSubmenuActive(c.Request.Context(), int32(id), req.IsActive)
	c.JSON(resp.StatusCode, resp)
}

// GetSubmenuByCode handles GET /submenus/by-code
// @Summary      Get submenu by code
// @Description  Retrieve a submenu using menu ID and submenu code
// @Tags         submenus
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header  string  true  "Tenant identifier"
// @Param        Authorization header  string  true  "Bearer token"
// @Param        menu_id       query   int     true  "Menu ID"
// @Param        code          query   string  true  "Submenu code"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/submenus/by-code [get]
func (h *SubmenuHandler) GetSubmenuByCode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	menuID, err := strconv.ParseInt(c.Query("menu_id"), 10, 32)
	if err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid menu id", nil)
		c.JSON(resp.StatusCode, resp)
		return
	}

	code := c.Query("code")
	if code == "" {
		resp := utils.NewResponse(utils.CodeBadReq, "submenu code is required", nil)
		c.JSON(resp.StatusCode, resp)
		return
	}

	resp := h.useCase.GetSubmenuByCode(c.Request.Context(), int32(menuID), code)
	c.JSON(resp.StatusCode, resp)
}

// ListSubmenus handles GET /submenus
// @Summary      List all submenus
// @Description  Retrieve all submenus across all menus
// @Tags         submenus
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header  string  true  "Tenant identifier"
// @Param        Authorization header  string  true  "Bearer token"
// @Success      200  {object}  SuccessResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/submenus [get]
func (h *SubmenuHandler) ListSubmenus(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListSubmenus(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListActiveSubmenusByMenu handles GET /submenus/active/{menu_id}
// @Summary      List active submenus by menu
// @Description  Retrieve only active submenus under a specific menu
// @Tags         submenus
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header  string  true  "Tenant identifier"
// @Param        Authorization header  string  true  "Bearer token"
// @Param        menu_id       path    int     true  "Menu ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/submenus/active/{menu_id} [get]
func (h *SubmenuHandler) ListActiveSubmenusByMenu(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	menuID, err := strconv.ParseInt(c.Param("menu_id"), 10, 32)
	if err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid menu id", nil)
		c.JSON(resp.StatusCode, resp)
		return
	}

	resp := h.useCase.ListActiveSubmenusByMenu(c.Request.Context(), int32(menuID))
	c.JSON(resp.StatusCode, resp)
}

// ListSubmenusByParent handles GET /submenus/parent/{parent_id}
// @Summary      List submenus by parent submenu
// @Description  Retrieve child submenus under a parent submenu
// @Tags         submenus
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header  string  true  "Tenant identifier"
// @Param        Authorization header  string  true  "Bearer token"
// @Param        parent_id     path    int     true  "Parent Submenu ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/submenus/parent/{parent_id} [get]
func (h *SubmenuHandler) ListSubmenusByParent(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	parentID, err := strconv.ParseInt(c.Param("parent_id"), 10, 32)
	if err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid parent submenu id", nil)
		c.JSON(resp.StatusCode, resp)
		return
	}

	resp := h.useCase.ListSubmenusByParent(c.Request.Context(), &[]int32{int32(parentID)}[0])
	c.JSON(resp.StatusCode, resp)
}

// DeleteSubmenu handles DELETE /submenus/{id}
// @Summary      Delete submenu
// @Description  Remove a submenu permanently
// @Tags         submenus
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header  string  true  "Tenant identifier"
// @Param        Authorization header  string  true  "Bearer token"
// @Param        id            path    int     true  "Submenu ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/submenus/{id} [delete]
func (h *SubmenuHandler) DeleteSubmenu(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		resp := utils.NewResponse(utils.CodeBadReq, "invalid submenu id", nil)
		c.JSON(resp.StatusCode, resp)
		return
	}

	resp := h.useCase.DeleteSubmenu(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}
