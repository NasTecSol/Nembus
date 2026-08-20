package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// SAPMigrationOrganizationMiddleware validates the trusted organization claim
// against the organizations table in the already selected tenant database.
// It deliberately does not consult the master repository.
func SAPMigrationOrganizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, ok := TrustedMachineIdentityFromContext(c.Request.Context())
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Trusted migration identity required"})
			c.Abort()
			return
		}
		repo, ok := RepositoryFromContext(c.Request.Context())
		if !ok {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tenant database unavailable"})
			c.Abort()
			return
		}

		organization, err := repo.GetOrganization(c.Request.Context(), identity.OrganizationID)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusForbidden, gin.H{"error": "Organization is not valid for this tenant"})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tenant organization validation unavailable"})
			}
			c.Abort()
			return
		}
		if organization.ID != identity.OrganizationID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Organization is not valid for this tenant"})
			c.Abort()
			return
		}

		c.Next()
	}
}
