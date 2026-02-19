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

// PermissionHandler holds the use case
type PermissionHandler struct {
	useCase *usecase.PermissionUseCase
}

// NewPermissionHandler creates a new handler instance
func NewPermissionHandler(uc *usecase.PermissionUseCase) *PermissionHandler {
	return &PermissionHandler{
		useCase: uc,
	}
}

// getRepositoryFromContext extracts repository from Gin context
func (h *PermissionHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreatePermission handles POST /api/permissions
// @Summary      Create a new permission
// @Description  Create a new permission with name, code, optional description and metadata
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        permission      body      CreatePermissionRequest  true  "Permission data"
// @Success      201  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions [post]
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req struct {
		Name        string      `json:"name" binding:"required"`
		Code        string      `json:"code" binding:"required"`
		Description *string     `json:"description,omitempty"`
		Metadata    interface{} `json:"metadata,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body", nil))
		return
	}
	var metadataBytes []byte
	if req.Metadata != nil {
		b, err := json.Marshal(req.Metadata)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "failed to process metadata", nil))
			return
		}
		metadataBytes = b
	}

	resp := h.useCase.CreatePermission(c.Request.Context(), req.Name, req.Code, req.Description, metadataBytes)
	c.JSON(resp.StatusCode, resp)
}

// GetPermission handles GET /api/permissions/:id
// @Summary      Get permission by ID
// @Description  Retrieve a specific permission by its ID
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        id              path      int     true  "Permission ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/{id} [get]
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid permission id", nil))
		return
	}

	resp := h.useCase.GetPermission(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// GetPermissionByCode handles GET /api/permissions/code/:code
// @Summary      Get permission by code
// @Description  Retrieve a specific permission by its unique code
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        code            path      string  true  "Permission code"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/code/{code} [get]
func (h *PermissionHandler) GetPermissionByCode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	code := c.Param("code")
	resp := h.useCase.GetPermissionByCode(c.Request.Context(), code)
	c.JSON(resp.StatusCode, resp)
}

// ListPermissions handles GET /api/permissions
// @Summary      List permissions
// @Description  Retrieve paginated list of permissions
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        limit           query     int     false "Limit (default 50)"
// @Param        offset          query     int     false "Offset (default 0)"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions [get]
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 32)
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 32)

	resp := h.useCase.ListPermissions(c.Request.Context(), int32(limit), int32(offset))
	c.JSON(resp.StatusCode, resp)
}

// UpdatePermission handles PUT /api/permissions/:id
// @Summary      Update a permission
// @Description  Update an existing permission by ID
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        id              path      int     true  "Permission ID"
// @Param        permission      body      UpdatePermissionRequest  true  "Permission data"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/{id} [put]
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid permission id", nil))
		return
	}

	var req struct {
		Name        *string     `json:"name,omitempty"`
		Description *string     `json:"description,omitempty"`
		Metadata    interface{} `json:"metadata,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body", nil))
		return
	}
	var metadataBytes []byte
	if req.Metadata != nil {
		b, err := json.Marshal(req.Metadata)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "failed to process metadata", nil))
			return
		}
		metadataBytes = b
	}

	resp := h.useCase.UpdatePermission(c.Request.Context(), int32(id), req.Name, req.Description, metadataBytes)
	c.JSON(resp.StatusCode, resp)
}

// DeletePermission handles DELETE /api/permissions/:id
// @Summary      Delete a permission
// @Description  Delete a permission by ID
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        id              path      int     true  "Permission ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/{id} [delete]
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid permission id", nil))
		return
	}

	resp := h.useCase.DeletePermission(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// AssignPermissionToMenu handles POST /api/permissions/menu/:menu_id
// @Summary      Assign permission to menu
// @Description  Assign a permission to a menu with optional metadata
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        menu_id         path      int     true  "Menu ID"
// @Param        body            body      AssignPermissionToEntityRequest  true  "Permission ID and optional metadata"
// @Success      201  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/menu/{menu_id} [post]
func (h *PermissionHandler) AssignPermissionToMenu(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	menuIDStr := c.Param("menu_id")
	menuID, err := strconv.ParseInt(menuIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid menu id", nil))
		return
	}

	var req struct {
		PermissionID int32       `json:"permission_id" binding:"required"`
		Metadata     interface{} `json:"metadata,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body", nil))
		return
	}
	var metadataBytes []byte
	if req.Metadata != nil {
		b, err := json.Marshal(req.Metadata)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "failed to process metadata", nil))
			return
		}
		metadataBytes = b
	}

	resp := h.useCase.AssignPermissionToMenu(c.Request.Context(), int32(menuID), req.PermissionID, metadataBytes)
	c.JSON(resp.StatusCode, resp)
}

// GetMenuPermissions handles GET /api/permissions/menu/:menu_id
// @Summary      Get menu permissions
// @Description  List all permissions assigned to a menu
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        menu_id         path      int     true  "Menu ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/menu/{menu_id} [get]
func (h *PermissionHandler) GetMenuPermissions(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	menuIDStr := c.Param("menu_id")
	menuID, err := strconv.ParseInt(menuIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid menu id", nil))
		return
	}

	resp := h.useCase.GetMenuPermissions(c.Request.Context(), int32(menuID))
	c.JSON(resp.StatusCode, resp)
}

// RevokePermissionFromMenu handles DELETE /api/permissions/menu/:menu_id/permission/:permission_id
// @Summary      Revoke permission from menu
// @Description  Remove a permission from a menu
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        menu_id         path      int     true  "Menu ID"
// @Param        permission_id   path      int     true  "Permission ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/menu/{menu_id}/permission/{permission_id} [delete]
func (h *PermissionHandler) RevokePermissionFromMenu(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	menuID, err := strconv.ParseInt(c.Param("menu_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid menu id", nil))
		return
	}
	permID, err := strconv.ParseInt(c.Param("permission_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid permission id", nil))
		return
	}

	resp := h.useCase.RevokePermissionFromMenu(c.Request.Context(), int32(menuID), int32(permID))
	c.JSON(resp.StatusCode, resp)
}

// AssignPermissionToModule handles POST /api/permissions/module/:module_id
// @Summary      Assign permission to module
// @Description  Assign a permission to a module with optional metadata
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        module_id       path      int     true  "Module ID"
// @Param        body            body      AssignPermissionToEntityRequest  true  "Permission ID and optional metadata"
// @Success      201  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/module/{module_id} [post]
func (h *PermissionHandler) AssignPermissionToModule(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	moduleIDStr := c.Param("module_id")
	moduleID, err := strconv.ParseInt(moduleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid module id", nil))
		return
	}

	var req struct {
		PermissionID int32       `json:"permission_id" binding:"required"`
		Metadata     interface{} `json:"metadata,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body", nil))
		return
	}
	var metadataBytes []byte
	if req.Metadata != nil {
		b, err := json.Marshal(req.Metadata)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "failed to process metadata", nil))
			return
		}
		metadataBytes = b
	}

	resp := h.useCase.AssignPermissionToModule(c.Request.Context(), int32(moduleID), req.PermissionID, metadataBytes)
	c.JSON(resp.StatusCode, resp)
}

// GetModulePermissions handles GET /api/permissions/module/:module_id
// @Summary      Get module permissions
// @Description  List all permissions assigned to a module
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        module_id       path      int     true  "Module ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/module/{module_id} [get]
func (h *PermissionHandler) GetModulePermissions(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	moduleIDStr := c.Param("module_id")
	moduleID, err := strconv.ParseInt(moduleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid module id", nil))
		return
	}

	resp := h.useCase.GetModulePermissions(c.Request.Context(), int32(moduleID))
	c.JSON(resp.StatusCode, resp)
}

// RevokePermissionFromModule handles DELETE /api/permissions/module/:module_id/permission/:permission_id
// @Summary      Revoke permission from module
// @Description  Remove a permission from a module
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        module_id       path      int     true  "Module ID"
// @Param        permission_id   path      int     true  "Permission ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/module/{module_id}/permission/{permission_id} [delete]
func (h *PermissionHandler) RevokePermissionFromModule(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	moduleID, err := strconv.ParseInt(c.Param("module_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid module id", nil))
		return
	}
	permID, err := strconv.ParseInt(c.Param("permission_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid permission id", nil))
		return
	}

	resp := h.useCase.RevokePermissionFromModule(c.Request.Context(), int32(moduleID), int32(permID))
	c.JSON(resp.StatusCode, resp)
}

// AssignPermissionToSubmenu handles POST /api/permissions/submenu/:submenu_id
// @Summary      Assign permission to submenu
// @Description  Assign a permission to a submenu with optional metadata
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        submenu_id      path      int     true  "Submenu ID"
// @Param        body            body      AssignPermissionToEntityRequest  true  "Permission ID and optional metadata"
// @Success      201  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/submenu/{submenu_id} [post]
func (h *PermissionHandler) AssignPermissionToSubmenu(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	submenuIDStr := c.Param("submenu_id")
	submenuID, err := strconv.ParseInt(submenuIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid submenu id", nil))
		return
	}

	var req struct {
		PermissionID int32       `json:"permission_id" binding:"required"`
		Metadata     interface{} `json:"metadata,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body", nil))
		return
	}
	var metadataBytes []byte
	if req.Metadata != nil {
		b, err := json.Marshal(req.Metadata)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "failed to process metadata", nil))
			return
		}
		metadataBytes = b
	}

	resp := h.useCase.AssignPermissionToSubmenu(c.Request.Context(), int32(submenuID), req.PermissionID, metadataBytes)
	c.JSON(resp.StatusCode, resp)
}

// GetSubmenuPermissions handles GET /api/permissions/submenu/:submenu_id
// @Summary      Get submenu permissions
// @Description  List all permissions assigned to a submenu
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        submenu_id      path      int     true  "Submenu ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/submenu/{submenu_id} [get]
func (h *PermissionHandler) GetSubmenuPermissions(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	submenuIDStr := c.Param("submenu_id")
	submenuID, err := strconv.ParseInt(submenuIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid submenu id", nil))
		return
	}

	resp := h.useCase.GetSubmenuPermissions(c.Request.Context(), int32(submenuID))
	c.JSON(resp.StatusCode, resp)
}

// RevokePermissionFromSubmenu handles DELETE /api/permissions/submenu/:submenu_id/permission/:permission_id
// @Summary      Revoke permission from submenu
// @Description  Remove a permission from a submenu
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        submenu_id      path      int     true  "Submenu ID"
// @Param        permission_id   path      int     true  "Permission ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/submenu/{submenu_id}/permission/{permission_id} [delete]
func (h *PermissionHandler) RevokePermissionFromSubmenu(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	submenuID, err := strconv.ParseInt(c.Param("submenu_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid submenu id", nil))
		return
	}
	permID, err := strconv.ParseInt(c.Param("permission_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid permission id", nil))
		return
	}

	resp := h.useCase.RevokePermissionFromSubmenu(c.Request.Context(), int32(submenuID), int32(permID))
	c.JSON(resp.StatusCode, resp)
}

// GetRolePermissionsWithScope handles GET /api/permissions/role/:role_id
// @Summary      Get role permissions with scope
// @Description  List all permissions assigned to a role with their scope
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        role_id         path      int     true  "Role ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/role/{role_id} [get]
func (h *PermissionHandler) GetRolePermissionsWithScope(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	roleIDStr := c.Param("role_id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid role id", nil))
		return
	}

	resp := h.useCase.GetRolePermissionsWithScope(c.Request.Context(), int32(roleID))
	c.JSON(resp.StatusCode, resp)
}

// RevokePermissionFromRole handles DELETE /api/permissions/role/:role_id/permission/:permission_id
// @Summary      Revoke permission from role
// @Description  Remove a permission from a role
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        role_id         path      int     true  "Role ID"
// @Param        permission_id   path      int     true  "Permission ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/role/{role_id}/permission/{permission_id} [delete]
func (h *PermissionHandler) RevokePermissionFromRole(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	roleID, err := strconv.ParseInt(c.Param("role_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid role id", nil))
		return
	}
	permID, err := strconv.ParseInt(c.Param("permission_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid permission id", nil))
		return
	}

	resp := h.useCase.RevokePermissionFromRole(c.Request.Context(), int32(roleID), int32(permID))
	c.JSON(resp.StatusCode, resp)
}

// UpdateRolePermissionScope handles PUT /api/permissions/role/:role_id/permission/:permission_id/scope
// @Summary      Update role permission scope
// @Description  Update the scope of a permission assigned to a role
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        role_id         path      int     true  "Role ID"
// @Param        permission_id   path      int     true  "Permission ID"
// @Param        body            body      UpdateRolePermissionScopeRequest  true  "Scope value"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/role/{role_id}/permission/{permission_id}/scope [put]
func (h *PermissionHandler) UpdateRolePermissionScope(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	roleID, err := strconv.ParseInt(c.Param("role_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid role id", nil))
		return
	}
	permID, err := strconv.ParseInt(c.Param("permission_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid permission id", nil))
		return
	}

	var req struct {
		Scope *string `json:"scope,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body", nil))
		return
	}

	resp := h.useCase.UpdateRolePermissionScope(c.Request.Context(), int32(roleID), int32(permID), req.Scope)
	c.JSON(resp.StatusCode, resp)
}

// GetUserPermissions handles GET /api/permissions/user/:user_id
// @Summary      Get user permissions
// @Description  List all permissions for a user (via roles)
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        user_id         path      int     true  "User ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/user/{user_id} [get]
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}

	resp := h.useCase.GetUserPermissions(c.Request.Context(), int32(userID))
	c.JSON(resp.StatusCode, resp)
}

// GetUserPermissionsWithScope handles GET /api/permissions/user/:user_id/with-scope
// @Summary      Get user permissions with scope
// @Description  List all permissions for a user with their scope
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        user_id         path      int     true  "User ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/user/{user_id}/with-scope [get]
func (h *PermissionHandler) GetUserPermissionsWithScope(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}

	resp := h.useCase.GetUserPermissionsWithScope(c.Request.Context(), int32(userID))
	c.JSON(resp.StatusCode, resp)
}

// CheckUserHasPermission handles GET /api/permissions/user/:user_id/check/:code
// @Summary      Check user has permission
// @Description  Check if a user has a specific permission by code
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        user_id         path      int     true  "User ID"
// @Param        code            path      string  true  "Permission code"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/user/{user_id}/check/{code} [get]
func (h *PermissionHandler) CheckUserHasPermission(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}
	code := c.Param("code")

	resp := h.useCase.CheckUserHasPermission(c.Request.Context(), int32(userID), code)
	c.JSON(resp.StatusCode, resp)
}

// GetUserAccessibleModules handles GET /api/permissions/user/:user_id/modules
// @Summary      Get user accessible modules
// @Description  List all modules accessible to a user based on permissions
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        user_id         path      int     true  "User ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/user/{user_id}/modules [get]
func (h *PermissionHandler) GetUserAccessibleModules(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}

	resp := h.useCase.GetUserAccessibleModules(c.Request.Context(), int32(userID))
	c.JSON(resp.StatusCode, resp)
}

// GetUserAccessibleMenus handles GET /api/permissions/user/:user_id/menus
// @Summary      Get user accessible menus
// @Description  List all menus accessible to a user based on permissions
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        user_id         path      int     true  "User ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/user/{user_id}/menus [get]
func (h *PermissionHandler) GetUserAccessibleMenus(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}

	resp := h.useCase.GetUserAccessibleMenus(c.Request.Context(), int32(userID))
	c.JSON(resp.StatusCode, resp)
}

// GetUserAccessibleSubmenus handles GET /api/permissions/user/:user_id/submenus
// @Summary      Get user accessible submenus
// @Description  List all submenus accessible to a user based on permissions
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        user_id         path      int     true  "User ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/user/{user_id}/submenus [get]
func (h *PermissionHandler) GetUserAccessibleSubmenus(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}

	resp := h.useCase.GetUserAccessibleSubmenus(c.Request.Context(), int32(userID))
	c.JSON(resp.StatusCode, resp)
}

// CheckUserSubmenuPermission handles GET /api/permissions/user/:user_id/submenu/:submenu_code
// @Summary      Check user submenu permission
// @Description  Checks if a user has access to a specific submenu by submenu code
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        user_id         path      int     true  "User ID"
// @Param        submenu_code    path      string  true  "Submenu code"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/user/{user_id}/submenu/{submenu_code} [get]
func (h *PermissionHandler) CheckUserSubmenuPermission(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}
	submenuCode := c.Param("submenu_code")
	if submenuCode == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "submenu code is required", nil))
		return
	}

	resp := h.useCase.CheckUserSubmenuPermission(c.Request.Context(), int32(userID), submenuCode)
	c.JSON(resp.StatusCode, resp)
}
