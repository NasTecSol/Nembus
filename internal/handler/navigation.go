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

// NavigationHandler holds the use case
type NavigationHandler struct {
	useCase *usecase.NavigationUseCase
}

// NewNavigationHandler creates a new handler instance
func NewNavigationHandler(uc *usecase.NavigationUseCase) *NavigationHandler {
	return &NavigationHandler{
		useCase: uc,
	}
}

// getRepositoryFromContext extracts repository from Gin context
func (h *NavigationHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// GetUserNavigation handles GET /api/navigation/user/:user_id
// @Summary      Get user navigation
// @Description  Returns the complete navigation structure for a specific user including modules, menus, submenus with permissions and UI settings
// @Tags         navigation
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header   string  true  "Tenant identifier"
// @Param        Authorization  header   string  true  "Bearer token"
// @Param        user_id        path     int     true  "User ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/navigation/user/{user_id} [get]
func (h *NavigationHandler) GetUserNavigation(c *gin.Context) {
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

	// Call usecase
	resp := h.useCase.GetUserNavigation(c.Request.Context(), int32(userID))

	// Respond with the response from use case
	c.JSON(resp.StatusCode, resp)
}
