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

// CheckUserSubmenuPermission handles GET /api/permissions/user/:user_id/submenu/:submenu_code
// @Summary      Check user submenu permission
// @Description  Checks if a user has access to a specific submenu by submenu code
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header   string  true  "Tenant identifier"
// @Param        Authorization  header   string  true  "Bearer token"
// @Param        user_id        path     int     true  "User ID"
// @Param        submenu_code   path     string  true  "Submenu code"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/permissions/user/{user_id}/submenu/{submenu_code} [get]
func (h *PermissionHandler) CheckUserSubmenuPermission(c *gin.Context) {
	// Get repository from context and set it on use case
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	// Get user ID from path parameter
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"invalid user ID",
			nil,
		))
		return
	}

	// Get submenu code from path parameter
	submenuCode := c.Param("submenu_code")
	if submenuCode == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"submenu code is required",
			nil,
		))
		return
	}

	// Call usecase
	resp := h.useCase.CheckUserSubmenuPermission(c.Request.Context(), int32(userID), submenuCode)

	// Respond with the response from use case
	c.JSON(resp.StatusCode, resp)
}
