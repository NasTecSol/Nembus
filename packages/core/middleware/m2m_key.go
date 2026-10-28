package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

// M2MAPIKey is the shared key used by clients to request M2M tokens.
const M2MAPIKey = "96a4f0d46e6b9ec8604766288a1f16767ef4c53c361c8dc3dfeb418a9cb7637c"

// M2MAPIKeyMiddleware authenticates token issuance without a user JWT.
func M2MAPIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		expected := M2MAPIKey
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
