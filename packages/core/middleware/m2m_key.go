package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// defaultM2MAPIKey is the development fallback when NEMBUS_M2M_API_KEY is unset.
const defaultM2MAPIKey = "96a4f0d46e6b9ec8604766288a1f16767ef4c53c361c8dc3dfeb418a9cb7637c"

// M2MAPIKey is the shared key used by clients to request M2M tokens (development default).
// Production deployments must set NEMBUS_M2M_API_KEY.
const M2MAPIKey = defaultM2MAPIKey

// ExpectedM2MAPIKey returns the configured M2M API key (env override in production).
func ExpectedM2MAPIKey() string {
	if v := strings.TrimSpace(os.Getenv("NEMBUS_M2M_API_KEY")); v != "" {
		return v
	}
	return defaultM2MAPIKey
}

// M2M tenant allowlist from NEMBUS_M2M_ALLOWED_TENANTS (comma-separated slugs).
// When unset, all tenants are allowed (development). In production, scope keys to permitted tenants.
func m2mAllowedTenantSet() map[string]struct{} {
	raw := strings.TrimSpace(os.Getenv("NEMBUS_M2M_ALLOWED_TENANTS"))
	if raw == "" {
		return nil
	}
	out := make(map[string]struct{})
	for part := range strings.SplitSeq(raw, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out[part] = struct{}{}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func m2mTenantAllowed(tenantID string) bool {
	allowed := m2mAllowedTenantSet()
	if allowed == nil {
		return true
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return false
	}
	_, ok := allowed[tenantID]
	return ok
}

// M2MAPIKeyMiddleware authenticates token issuance without a user JWT.
func M2MAPIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		expected := ExpectedM2MAPIKey()
		provided := c.GetHeader("x-m2m-key")
		expectedHash := sha256.Sum256([]byte(expected))
		providedHash := sha256.Sum256([]byte(provided))
		if provided == "" || subtle.ConstantTimeCompare(expectedHash[:], providedHash[:]) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing x-m2m-key"})
			return
		}
		tenantID := c.GetHeader("x-tenant-id")
		if tenantID == "" {
			tenantID = c.Query("tenant_id")
		}
		if !m2mTenantAllowed(tenantID) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "x-tenant-id is not permitted for M2M token issuance"})
			return
		}
		c.Next()
	}
}
