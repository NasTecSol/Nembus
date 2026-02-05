package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"
	"NEMBUS/utils" // Assuming your NewResponse is here

	"github.com/gin-gonic/gin"
)

// UserHandler holds the use case
type UserHandler struct {
	useCase *usecase.UserUseCase
}

// NewUserHandler creates a new handler instance
func NewUserHandler(uc *usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		useCase: uc,
	}
}

// getRepositoryFromContext extracts repository from Gin context
func (h *UserHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateUser handles POST /users
// @Summary      Create a new user
// @Description  Create a new user with optional login credentials
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        user      body      CreateUserRequest  true  "User data"
// @Success      201  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	// Get repository from context and set it on use case
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return // Error already handled in getRepositoryFromContext
	}
	h.useCase.SetRepository(repo)

	// Bind JSON input
	var req struct {
		FirstName    string  `json:"first_name" binding:"required"`
		LastName     string  `json:"last_name"`
		Username     string  `json:"username" binding:"required"`
		Email        string  `json:"email" binding:"required"`
		IsActive     bool    `json:"is_active"`
		Password     *string `json:"password,omitempty"`
		EmployeeCode *string `json:"employee_code,omitempty"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Call UseCase
	response := h.useCase.CreateUser(c.Request.Context(), req.FirstName, req.LastName, req.Username, req.Email, req.IsActive, req.Password, req.EmployeeCode)

	// Respond with the response from use case
	c.JSON(response.StatusCode, response)
}

// GetUser handles GET /users/:id
// @Summary      Get user by ID
// @Description  Retrieve a specific user by their ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        id            path      string  true  "User ID"
// @Success      200  {object}  UserResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	// Get repository from context and set it on use case
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetUser(c.Request.Context(), id)
	if resp.StatusCode != utils.CodeOK {
		c.JSON(resp.StatusCode, gin.H{"error": resp.Message})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListUsers handles GET /users
// @Summary      List all users
// @Description  Retrieve a list of all users for the tenant
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        limit         query     int     false "Limit number of results"
// @Param        offset        query     int     false "Offset for pagination"
// @Success      200  {array}   UserResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Get repository from context and set it on use case
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		limit = 100
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil {
		offset = 0
	}

	resp := h.useCase.ListUsers(c.Request.Context(), int32(limit), int32(offset))

	// Return the standard response
	c.JSON(resp.StatusCode, resp)
}

// AssignRoleToUser handles POST /users/:id/roles
// @Summary      Assign role to user
// @Description  Assign a specific role to a user with optional metadata
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "User ID"
// @Param body body AssignRoleToUserRequest true "Role assignment data"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/users/addUserRoles/{id} [post]
func (h *UserHandler) AssignRoleToUser(c *gin.Context) {
	// Get repository from context and set it on use case
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	// Parse user ID from path
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"invalid user id",
			nil,
		))
		return
	}

	// Bind request body
	var req AssignRoleToUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(
			utils.CodeBadReq,
			"invalid request body",
			nil,
		))
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

	// Call usecase
	resp := h.useCase.AssignRoleToUser(
		c.Request.Context(),
		int32(userID),
		req.RoleID,
		req.StoreID, // 👈 swagger + handler now aligned
		metadataBytes,
	)

	c.JSON(resp.StatusCode, resp)
}

// UpdateUser handles PATCH /users/{id}
// @Summary      Update user details
// @Description  Update email, name, employee code, is_active, or metadata
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "User ID"
// @Param        body          body      UpdateUserRequest  true  "User update payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router /api/users/{id} [patch]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	metaBytes, _ := json.Marshal(req.Metadata)

	resp := h.useCase.UpdateUser(
		c.Request.Context(),
		int32(id),
		req.Email,
		req.FirstName,
		req.LastName,
		req.EmployeeCode,
		req.IsActive,
		metaBytes,
	)
	c.JSON(resp.StatusCode, resp)
}

// UpdateUserPassword handles POST /users/:id/password
// @Summary      Update user password
// @Description  Change a user's password
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "User ID"
// @Param        body          body      UpdateUserPasswordRequest  true  "Password payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router /api/users/{id}/password [put]
func (h *UserHandler) UpdateUserPassword(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}

	var req UpdateUserPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid password", nil))
		return
	}

	resp := h.useCase.UpdateUserPassword(c.Request.Context(), int32(id), req.NewPassword)
	c.JSON(resp.StatusCode, resp)
}

// GrantStoreAccess handles POST /users/grantStore/{id}
// @Summary      Grant store access to user
// @Description  Grant a user access to a store
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header string true "Tenant identifier"
// @Param        Authorization header string true "Bearer token"
// @Param        id             path   int    true "User ID"
// @Param        body           body   GrantStoreAccessRequest true "Store access payload"
// @Success      201 {object} SuccessResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/grantStore/{id} [post]
func (h *UserHandler) GrantStoreAccess(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}

	var req GrantStoreAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	metaBytes, _ := json.Marshal(req.Metadata)

	resp := h.useCase.GrantStoreAccess(c.Request.Context(), int32(id), req.StoreID, req.IsPrimary, metaBytes)
	c.JSON(resp.StatusCode, resp)
}

// SetUserPrimaryStore handles PUT /users/{id}/stores/primary
// @Summary      Set user's primary store
// @Description  Unset other primary stores and set new primary store
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header string true "Tenant identifier"
// @Param        Authorization header string true "Bearer token"
// @Param        id             path   int    true "User ID"
// @Success      200 {object} SuccessResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/{id}/stores/primary [put]
func (h *UserHandler) SetUserPrimaryStore(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}

	resp := h.useCase.SetUserPrimaryStore(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// GetUsersByRole handles GET /users/role/{role_id}
// @Summary      List users by role
// @Description  Retrieve all users that have a specific role
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        Authorization header string true "Bearer token"
// @Param        role_id path int true "Role ID"
// @Success      200 {object} SuccessResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router /api/users/role/{role_id} [get]
func (h *UserHandler) GetUsersByRole(c *gin.Context) {
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

	resp := h.useCase.GetUsersByRole(c.Request.Context(), int32(roleID))
	c.JSON(resp.StatusCode, resp)
}

// RevokeRole handles DELETE /users/revokeRole/{user_id}/{role_id}
// @Summary      Revoke a role from user
// @Description  Remove a specific role from a user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header string true "Tenant identifier"
// @Param        Authorization header string true "Bearer token"
// @Param        user_id        path   int    true "User ID"
// @Param        role_id        path   int    true "Role ID"
// @Success      200 {object} SuccessResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/revokeRole/{user_id}/{role_id} [delete]
func (h *UserHandler) RevokeRole(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}

	var req struct {
		RoleID int32 `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body", nil))
		return
	}

	resp := h.useCase.RevokeRole(c.Request.Context(), int32(userID), req.RoleID)
	c.JSON(resp.StatusCode, resp)
}

// RevokeAllRoles handles DELETE /users/revokeAllRoles/{user_id}
// @Summary      Revoke all roles from user
// @Description  Remove all roles assigned to a user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header string true "Tenant identifier"
// @Param        Authorization header string true "Bearer token"
// @Param        user_id        path   int    true "User ID"
// @Success      200 {object} SuccessResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/revokeAllRoles/{user_id} [delete]
func (h *UserHandler) RevokeAllRoles(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}

	resp := h.useCase.RevokeAllRoles(c.Request.Context(), int32(userID))
	c.JSON(resp.StatusCode, resp)
}

// RevokeStoreAccess handles DELETE /users/revokeStore/{user_id}/{store_id}
// @Summary      Revoke store access for a user
// @Description  Remove a specific store access from a user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        Authorization header string true "Bearer token"
// @Param        user_id path int true "User ID"
// @Param        store_id path int true "Store ID"
// @Success      200 {object} SuccessResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/revokeStore/{user_id}/{store_id} [delete]
func (h *UserHandler) RevokeStoreAccess(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}

	var req struct {
		StoreID int32 `json:"store_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body", nil))
		return
	}

	resp := h.useCase.RevokeStoreAccess(c.Request.Context(), int32(userID), req.StoreID)
	c.JSON(resp.StatusCode, resp)
}

// RevokeAllStoreAccess handles DELETE /users/revokeAllStores/{user_id}
// @Summary      Revoke all store access for a user
// @Description  Remove all store access assigned to a user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        Authorization header string true "Bearer token"
// @Param        user_id path int true "User ID"
// @Success      200 {object} SuccessResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/revokeAllStores/{user_id} [delete]
func (h *UserHandler) RevokeAllStoreAccess(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}

	resp := h.useCase.RevokeAllStoreAccess(c.Request.Context(), int32(userID))
	c.JSON(resp.StatusCode, resp)
}

// GetUserWithDetails handles GET /users/details/:id
// @Summary      Get user with roles and stores
// @Description  Retrieve detailed user info including roles and store access
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        Authorization header string true "Bearer token"
// @Param        id path int true "User ID"
// @Success      200 {object} SuccessResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/details/{id} [get]
func (h *UserHandler) GetUserWithDetails(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}

	resp := h.useCase.GetUserWithDetails(c.Request.Context(), int32(userID))
	c.JSON(resp.StatusCode, resp)
}

// ListUsersWithDetails handles GET /users/details
// @Summary      List all users with details
// @Description  Retrieve all users including roles and stores
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        Authorization header string true "Bearer token"
// @Param        limit query int false "Limit number of results"
// @Param        offset query int false "Offset for pagination"
// @Success      200 {object} SuccessResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/details [get]
func (h *UserHandler) ListUsersWithDetails(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		limit = 100
	}
	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil {
		offset = 0
	}

	resp := h.useCase.ListUsersWithDetails(c.Request.Context(), int32(limit), int32(offset))
	c.JSON(resp.StatusCode, resp)
}

// SearchUsers handles GET /users/search
// @Summary      Search users
// @Description  Search users by name, username, or email
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        Authorization header string true "Bearer token"
// @Param        q query string true "Search query"
// @Success      200 {object} SuccessResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	// Get query parameter
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "query parameter 'q' is required", nil))
		return
	}

	// Optional: pagination
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		limit = 100
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil {
		offset = 0
	}

	// Call use case with search term, limit, and offset
	resp := h.useCase.SearchUsers(c.Request.Context(), query, int32(limit), int32(offset))
	c.JSON(resp.StatusCode, resp)
}

// GetStoreUsers handles GET /store/:store_id
// @Summary      List users of a store
// @Description  Retrieve all users who have access to a specific store
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        Authorization header string true "Bearer token"
// @Param        store_id path int true "Store ID"
// @Success      200 {object} SuccessResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router /store/{store_id} [get]
func (h *UserHandler) GetStoreUsers(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeIDStr := c.Param("store_id")
	storeID, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store id", nil))
		return
	}

	resp := h.useCase.GetStoreUsers(c.Request.Context(), int32(storeID))
	c.JSON(resp.StatusCode, resp)
}

// GetUserStores handles GET /users/:id/stores
// @Summary      Get stores assigned to a user
// @Description  Retrieve all stores a user has access to
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        Authorization header string true "Bearer token"
// @Param        id path int true "User ID"
// @Success      200 {object} SuccessResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/{id}/stores [get]
func (h *UserHandler) GetUserStores(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}

	resp := h.useCase.GetUserStores(c.Request.Context(), int32(userID))
	c.JSON(resp.StatusCode, resp)
}

// GetUserPrimaryStore handles GET /users/:id/primaryStore
// @Summary      Get user's primary store
// @Description  Retrieve the primary store assigned to the user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        Authorization header string true "Bearer token"
// @Param        id path int true "User ID"
// @Success      200 {object} SuccessResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/{id}/primaryStore [get]
func (h *UserHandler) GetUserPrimaryStore(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid user id", nil))
		return
	}

	resp := h.useCase.GetUserPrimaryStore(c.Request.Context(), int32(userID))
	c.JSON(resp.StatusCode, resp)
}
