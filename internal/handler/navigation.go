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
	useCase  *usecase.NavigationUseCase
	useCase2 *usecase.RoleUseCase
	useCase3 *usecase.UserUseCase
}

// NewNavigationHandler creates a new handler instance
func NewNavigationHandler(uc *usecase.NavigationUseCase, uc2 *usecase.RoleUseCase, uc3 *usecase.UserUseCase) *NavigationHandler {
	return &NavigationHandler{
		useCase:  uc,
		useCase2: uc2,
		useCase3: uc3,
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

// GetNavigationByRoleCode handles GET /api/navigation/role/:role_code
// @Summary      Get navigation by role
// @Description  Returns the complete navigation structure for a specific role including modules, menus, submenus with permissions, UI settings, and the number of users assigned to this role
// @Tags         navigation
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header   string  true  "Tenant identifier"
// @Param        Authorization  header   string  true  "Bearer token"
// @Param        role_code      path     string  true  "Role code"
// @Success      200  {object}  RoleNavigationResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/navigation/rolesWithUserCounts/{role_code} [get]
func (h *NavigationHandler) GetNavigationByRoleCodeWithUserCounts(c *gin.Context) {
	// Get repository from context and set it on use case
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)
	h.useCase2.SetRepository(repo)
	h.useCase3.SetRepository(repo)

	// Get role code from path parameter
	roleCode := c.Param("role_code")
	if roleCode == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"role code cannot be empty",
			nil,
		))
		return
	}

	// Step 2: Fetch role info to get role ID
	roleResp := h.useCase2.GetRoleByCode(c.Request.Context(), roleCode)
	if roleResp.StatusCode != utils.CodeOK {
		// Role not found or error
		c.JSON(roleResp.StatusCode, roleResp)
		return
	}
	// Type assert to your Role struct
	roleStruct, ok := roleResp.Data.(repository.Role)
	if !ok {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"invalid role data format",
			nil,
		))
		return
	}

	roleID := roleStruct.ID

	// Call the use case to get users with this role
	usersResp := h.useCase3.GetUsersByRole(c.Request.Context(), roleID)
	if usersResp.StatusCode != utils.CodeOK {
		c.JSON(usersResp.StatusCode, usersResp)
		return
	}
	userList := usersResp.Data
	userCount := 0
	if users, ok := userList.([]repository.User); ok {
		userCount = len(users)
	}

	// Call usecase
	resp := h.useCase.GetNavigationByRoleCode(c.Request.Context(), roleCode)

	// Step 4: Combine navigation and user info
	responseData := map[string]interface{}{
		"navigation": resp.Data,
		"user_count": userCount,
		//"users":      userList,
	}

	// Respond with the response from use case
	c.JSON(resp.StatusCode, utils.NewResponse(
		utils.CodeOK,
		"navigation fetched successfully",
		responseData,
	))
}
