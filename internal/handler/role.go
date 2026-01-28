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
		PermissionID int32       `json:"permission_id" binding:"required"`
		Scope        *string     `json:"scope,omitempty"`
		Metadata     interface{} `json:"metadata"`
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

	resp := h.useCase.AssignPermissionToRole(
		c.Request.Context(),
		int32(roleID),
		req.PermissionID,
		req.Scope,
		metadataBytes,
	)
	c.JSON(resp.StatusCode, resp)
}

// RemovePermissionFromRole handles DELETE /api/roles/:id/permissions/:permission_id
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
