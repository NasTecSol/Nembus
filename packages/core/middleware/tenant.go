package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/NasTecSol/nembus-core/middleware/manager"
	"github.com/NasTecSol/nembus-core/repository"

	"github.com/gin-gonic/gin"
)

type contextKey string

const RepoKey contextKey = "tenant_repo"

// TenantMiddleware returns a Gin middleware that injects tenant-specific repository
func TenantMiddleware(tm *manager.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("x-tenant-id")
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "x-tenant-id header required"})
			c.Abort()
			return
		}

		pool, err := tm.GetPool(c.Request.Context(), tenantID)
		if err != nil {
			// Log the actual error for debugging
			log.Printf("Failed to get tenant pool for slug '%s': %v", tenantID, err)
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Tenant not found or inactive",
				"details": err.Error(),
				"slug":    tenantID,
			})
			c.Abort()
			return
		}

		// Injects the tenant-specific repository into the request context
		repo := repository.New(pool)
		ctx := context.WithValue(c.Request.Context(), RepoKey, repo)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// MasterRepositoryMiddleware returns a Gin middleware that injects the master repository
// This is used for public endpoints that need to access the master database
func MasterRepositoryMiddleware(masterRepo *repository.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Injects the master repository into the request context
		ctx := context.WithValue(c.Request.Context(), RepoKey, masterRepo)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
