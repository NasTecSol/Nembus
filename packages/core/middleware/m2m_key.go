package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// M2MAPIKeyMiddleware authenticates token issuance without a user JWT.
func M2MAPIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		expected := os.Getenv("M2M_API_KEY")
		if expected == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "M2M token issuance is not configured"})
			return
		}
		provided := c.GetHeader("x-m2m-key")
		expectedHash := sha256.Sum256([]byte(expected))
		providedHash := sha256.Sum256([]byte(provided))
		if provided == "" || subtle.ConstantTimeCompare(expectedHash[:], providedHash[:]) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing x-m2m-key"})
			return
		}
		c.Next()
	}
}
