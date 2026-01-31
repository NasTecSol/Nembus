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

// RoleHandler holds the use case
type RoleHandler struct {
	useCase *usecase.RoleUseCase
}

// NewRoleHandler creates a new handler instance
func NewRoleHandler(uc *usecase.RoleUseCase) *RoleHandler {
	return &RoleHandler{useCase: uc}
}

// getRepositoryFromContext extracts repository from Gin context
func (h *RoleHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateRole handles POST /api/roles
// @Summary      Create a new role
// @Description  Create a new role with optional metadata and system flag
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        role          body      CreateRoleRequest  true  "Role data"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/roles [post]
func (h *RoleHandler) CreateRole(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req struct {
		Name         string      `json:"name" binding:"required"`
		Code         string      `json:"code" binding:"required"`
		Description  *string     `json:"description,omitempty"`
		IsSystemRole bool        `json:"is_system_role"`
		IsActive     bool        `json:"is_active"`
		Metadata     interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"invalid request body",
			nil,
		))
		return
	}

	var metadataBytes []byte
	if req.Metadata != nil {
		b, err := json.Marshal(req.Metadata)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewResponse(
				utils.CodeError,
				"failed to process metadata",
				nil,
			))
			return
		}
		metadataBytes = b
	}

	resp := h.useCase.CreateRole(
		c.Request.Context(),
		req.Name,
		req.Code,
		req.Description,
		req.IsSystemRole,
		req.IsActive,
		metadataBytes,
	)

	c.JSON(resp.StatusCode, resp)
}

// GetRole handles GET /api/roles/:id
// @Summary      Get role by ID
// @Description  Retrieve a specific role by its ID
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Role ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/roles/{id} [get]
func (h *RoleHandler) GetRole(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid role id", nil))
		return
	}

	resp := h.useCase.GetRole(c.Request.Context(), int32(roleID))
	c.JSON(resp.StatusCode, resp)
}

// GetRoleByCode handles GET /api/roles/code/:code
// @Summary      Get role by code
// @Description  Retrieve a specific role by its unique code
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        code          path      string  true  "Role code"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/roles/code/{code} [get]
func (h *RoleHandler) GetRoleByCode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	code := c.Param("code")
	resp := h.useCase.GetRoleByCode(c.Request.Context(), code)
	c.JSON(resp.StatusCode, resp)
}

// ListRoles handles GET /api/roles
// @Summary      List all roles
// @Description  Retrieve a list of all roles
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/roles [get]
func (h *RoleHandler) ListRoles(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListRoles(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListActiveRoles handles GET /api/roles/active
// @Summary      List active roles
// @Description  Retrieve a list of all active roles
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/roles/active [get]
func (h *RoleHandler) ListActiveRoles(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListActiveRoles(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ListNonSystemRoles handles GET /api/roles/non-system
// @Summary      List non-system roles
// @Description  Retrieve a list of roles that are not system roles
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/roles/non-system [get]
func (h *RoleHandler) ListNonSystemRoles(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.ListNonSystemRoles(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// UpdateRole handles PUT /api/roles/:id
// @Summary      Update a role
// @Description  Update role details including name, description and active status
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int               true  "Role ID"
// @Param        role          body      UpdateRoleRequest true  "Role data"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/roles/{id} [put]
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid role id", nil))
		return
	}

	var req struct {
		Name        string      `json:"name" binding:"required"`
		Description *string     `json:"description,omitempty"`
		IsActive    bool        `json:"is_active"`
		Metadata    interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"invalid request body",
			nil,
		))
		return
	}

	var metadataBytes []byte
	if req.Metadata != nil {
		b, err := json.Marshal(req.Metadata)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewResponse(
				utils.CodeError,
				"failed to process metadata",
				nil,
			))
			return
		}
		metadataBytes = b
	}

	resp := h.useCase.UpdateRole(
		c.Request.Context(),
		int32(roleID),
		req.Name,
		req.Description,
		req.IsActive,
		metadataBytes,
	)
	c.JSON(resp.StatusCode, resp)
}

// DeleteRole handles DELETE /api/roles/:id
// @Summary      Delete a role
// @Description  Delete a role by ID (non-system roles only)
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Role ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid role id", nil))
		return
	}

	resp := h.useCase.DeleteRole(c.Request.Context(), int32(roleID))
	c.JSON(resp.StatusCode, resp)
}

// ToggleRoleActive handles PATCH /api/roles/:id/active
// @Summary      Toggle role active status
// @Description  Enable or disable a role
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Role ID"
// @Param        body          body      object  true  "Active flag"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/roles/{id}/active [patch]
func (h *RoleHandler) ToggleRoleActive(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid role id", nil))
		return
	}

	var req struct {
		IsActive bool `json:"is_active" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body", nil))
		return
	}

	resp := h.useCase.ToggleRoleActive(c.Request.Context(), int32(roleID), req.IsActive)
	c.JSON(resp.StatusCode, resp)
}

// AssignPermissionToRole handles POST /api/roles/:id/permissions
// @Summary      Assign permissions to role
// @Description  Assign one or more permissions to a role with optional scope and metadata
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string                        true  "Tenant identifier"
// @Param        Authorization   header    string                        true  "Bearer token"
// @Param        id              path      int                           true  "Role ID"
// @Param permission_data body AssignPermissionToRoleRequest true "Array of permissions to assign"
// @Success      201             {object}  SuccessResponse
// @Failure      400             {object}  ErrorResponse
// @Failure      401             {object}  ErrorResponse
// @Failure      404             {object}  ErrorResponse
// @Failure      500             {object}  ErrorResponse
// @Router       /api/roles/{id}/permissions [post]
func (h *RoleHandler) AssignPermissionToRole(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid role id", nil))
		return
	}

	var req struct {
		Permissions []struct {
			PermissionID int32       `json:"permission_id" binding:"required"`
			Scope        *string     `json:"scope,omitempty"`
			Metadata     interface{} `json:"metadata"`
		} `json:"permissions" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body", nil))
		return
	}

	// Convert metadata to []byte for each permission
	var permissions []usecase.RolePermissionInput
	for _, p := range req.Permissions {
		metaBytes := []byte("{}")
		if p.Metadata != nil {
			b, err := json.Marshal(p.Metadata)
			if err != nil {
				c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "failed to process metadata", nil))
				return
			}
			metaBytes = b
		}
		permissions = append(permissions, usecase.RolePermissionInput{
			PermissionID: p.PermissionID,
			Scope:        p.Scope,
			Metadata:     metaBytes,
		})
	}

	resp := h.useCase.AssignPermissionToRole(
		c.Request.Context(),
		int32(roleID),
		permissions,
	)
	c.JSON(resp.StatusCode, resp)
}

// RemovePermissionFromRole handles DELETE /api/roles/:id/permissions/:permission_id
// @Summary      Remove permission from role
// @Description  Remove a specific permission from a role
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        id             path      int     true  "Role ID"
// @Param        permission_id  path      int     true  "Permission ID"
// @Success      200            {object}  SuccessResponse
// @Failure      400            {object}  ErrorResponse
// @Failure      401            {object}  ErrorResponse
// @Failure      404            {object}  ErrorResponse
// @Failure      500            {object}  ErrorResponse
// @Router       /api/roles/{id}/permissions/{permission_id} [delete]
func (h *RoleHandler) RemovePermissionFromRole(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid role id", nil))
		return
	}

	permIDStr := c.Param("permission_id")
	permID, err := strconv.ParseInt(permIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid permission id", nil))
		return
	}

	resp := h.useCase.RemovePermissionFromRole(c.Request.Context(), int32(roleID), int32(permID))
	c.JSON(resp.StatusCode, resp)
}

// GetRolePermissions handles GET /api/roles/:id/permissions
// @Summary      Get role permissions
// @Description  List all permissions assigned to a role
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Role ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/roles/{id}/permissions [get]
func (h *RoleHandler) GetRolePermissions(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid role id", nil))
		return
	}

	resp := h.useCase.GetRolePermissions(c.Request.Context(), int32(roleID))
	c.JSON(resp.StatusCode, resp)
}

// CheckRoleHasPermission handles GET /api/roles/:id/permissions/:permission_id/check
// @Summary      Check role permission
// @Description  Check if a role has a specific permission
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        id             path      int     true  "Role ID"
// @Param        permission_id  path      int     true  "Permission ID"
// @Success      200            {object}  SuccessResponse
// @Failure      400            {object}  ErrorResponse
// @Failure      401            {object}  ErrorResponse
// @Failure      404            {object}  ErrorResponse
// @Failure      500            {object}  ErrorResponse
// @Router       /api/roles/{id}/permissions/{permission_id}/check [get]
func (h *RoleHandler) CheckRoleHasPermission(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid role id", nil))
		return
	}

	permIDStr := c.Param("permission_id")
	permID, err := strconv.ParseInt(permIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid permission id", nil))
		return
	}

	resp := h.useCase.CheckRoleHasPermission(c.Request.Context(), int32(roleID), int32(permID))
	c.JSON(resp.StatusCode, resp)
}
