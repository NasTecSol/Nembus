package handler

import (
	"net/http"
	"os"

	"NEMBUS/internal/middleware"

	"github.com/gin-gonic/gin"
)

// DevHandler handles development-only endpoints
type DevHandler struct{}

// NewDevHandler creates a new dev handler instance
func NewDevHandler() *DevHandler {
	return &DevHandler{}
}

// GetDevToken generates a development token for testing
// This endpoint should only be available in development mode
// @Summary      Get development token
// @Description  Generate a JWT token for development/testing purposes. Only available in development mode.
// @Tags         dev
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string  "token"
// @Failure      403  {object}  map[string]string  "error"
// @Failure      500  {object}  map[string]string  "error"
// @Router       /dev/token [get]
func (h *DevHandler) GetDevToken(c *gin.Context) {
	// Check if we're in development mode
	env := os.Getenv("ENV")
	if env != "development" && env != "dev" {
		c.JSON(http.StatusForbidden, gin.H{"error": "dev token endpoint only available in development mode"})
		return
	}

	// Get dev user ID and login from environment or use defaults
	devUserID := os.Getenv("DEV_USER_ID")
	if devUserID == "" {
		devUserID = "00000000-0000-0000-0000-000000000000" // Default dev user ID
	}

	devUserLogin := os.Getenv("DEV_USER_LOGIN")
	if devUserLogin == "" {
		devUserLogin = "dev_user"
	}

	// Generate token
	token, err := middleware.GenerateJWTToken(devUserID, devUserLogin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate dev token", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"type":       "Bearer",
		"user_id":    devUserID,
		"user_login": devUserLogin,
		"note":       "This is a development token. Set ENV=development to use this endpoint.",
	})
}
