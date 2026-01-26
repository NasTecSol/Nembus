package handler

import (
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
