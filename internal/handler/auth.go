package handler

import (
	"net/http"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"

	"github.com/gin-gonic/gin"
)

// AuthHandler holds the auth use case
type AuthHandler struct {
	useCase *usecase.AuthUseCase
}

// NewAuthHandler creates a new auth handler instance
func NewAuthHandler(uc *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		useCase: uc,
	}
}

// getRepositoryFromContext extracts repository from Gin context
func (h *AuthHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// Login handles POST /login
// @Summary      User login
// @Description  Authenticate user and receive JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        request      body      LoginRequest  true  "Login credentials"
// @Success      200  {object}  LoginResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	// Get repository from context and set it on use case
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return // Error already handled in getRepositoryFromContext
	}
	h.useCase.SetRepository(repo)

	// Bind JSON input
	var req struct {
		UserLogin string `json:"user_login" binding:"required"`
		Password  string `json:"password" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Call UseCase
	token, err := h.useCase.Login(c.Request.Context(), req.UserLogin, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Respond with token
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"type":  "Bearer",
	})
}
