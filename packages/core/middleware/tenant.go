package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"unicode"

	"github.com/NasTecSol/nembus-core/middleware/manager"
	"github.com/NasTecSol/nembus-core/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const RepoKey contextKey = "tenant_repo"
const tenantSlugKey contextKey = "tenant_slug"
const tenantPoolKey contextKey = "tenant_pool"

// validTenantSlug validates the canonical tenant selector without imposing a
// new taxonomy. Tenant slugs are stored as VARCHAR(100) and are matched
// exactly by the tenant registry.
func validTenantSlug(slug string) bool {
	if slug == "" || len(slug) > 100 || strings.TrimSpace(slug) != slug {
		return false
	}
	for _, r := range slug {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	return true
}

// ValidTenantSlug exposes the same canonical validation to credential
// issuance and other boundary handlers without exposing tenant persistence.
func ValidTenantSlug(slug string) bool { return validTenantSlug(slug) }

// TenantSlugFromContext extracts the trusted tenant slug from a request
// context. It never falls back to client input.
func TenantSlugFromContext(ctx context.Context) (string, bool) {
	slug, ok := ctx.Value(tenantSlugKey).(string)
	return slug, ok && validTenantSlug(slug)
}

// AuthenticatedTenantSlug extracts the verified tenant binding from a request
// context. It is populated by TenantBindingMiddleware for protected requests
// and by TenantMiddleware after login tenant selection.
func AuthenticatedTenantSlug(ctx context.Context) (string, bool) {
	return TenantSlugFromContext(ctx)
}

func withTenantSlug(ctx context.Context, slug string) context.Context {
	return context.WithValue(ctx, tenantSlugKey, slug)
}

// RepositoryFromContext returns the repository created for the selected
// tenant. It is intentionally typed so request handlers cannot fall back to a
// process-global master repository.
func RepositoryFromContext(ctx context.Context) (*repository.Queries, bool) {
	repo, ok := ctx.Value(RepoKey).(*repository.Queries)
	return repo, ok && repo != nil
}

// TenantPoolFromContext returns the exact pool used to construct the request
// repository. Request-scoped use cases use this pool without mutable global
// repository setters.
func TenantPoolFromContext(ctx context.Context) (*pgxpool.Pool, bool) {
	pool, ok := ctx.Value(tenantPoolKey).(*pgxpool.Pool)
	return pool, ok && pool != nil
}

// TenantBindingMiddleware validates the signed tenant binding before
// TenantMiddleware can select a tenant database.
func TenantBindingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isM2M, _ := c.Get("is_m2m")
		if m2m, ok := isM2M.(bool); ok && m2m {
			claims, ok := GetClaimsFromContext(c)
			signedTenant, signedOK := claims[tenantSlugClaim].(string)
			if !signedOK {
				signedTenant, signedOK = claims["tenant_id"].(string)
			}
			requestedTenant, requestedExists := c.Get("m2m_requested_tenant")
			requested, requestedOK := requestedTenant.(string)
			if !ok || !signedOK || !validTenantSlug(signedTenant) ||
				(requestedExists && requestedOK && requested != "" && requested != signedTenant) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized tenant"})
				c.Abort()
				return
			}
			c.Request = c.Request.WithContext(withTenantSlug(c.Request.Context(), signedTenant))
			c.Next()
			return
		}

		claims, ok := GetClaimsFromContext(c)
		signedTenant, signedOK := claims[tenantSlugClaim].(string)
		requestedTenant := c.GetHeader("x-tenant-id")
		if !ok || !signedOK || !validTenantSlug(signedTenant) ||
			!validTenantSlug(requestedTenant) || requestedTenant != signedTenant {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized tenant"})
			c.Abort()
			return
		}

		c.Request = c.Request.WithContext(withTenantSlug(c.Request.Context(), signedTenant))
		c.Next()
	}
}

// TenantMiddleware returns a Gin middleware that injects tenant-specific repository
func TenantMiddleware(tm *manager.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, bound := TenantSlugFromContext(c.Request.Context())
		requestedTenant := c.GetHeader("x-tenant-id")
		if bound {
			if !validTenantSlug(requestedTenant) || requestedTenant != tenantID {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized tenant"})
				c.Abort()
				return
			}
		} else {
			// A request that already passed JWT authentication must have passed
			// TenantBindingMiddleware. Login is the normal path without JWT.
			if _, authenticated := c.Get(string(ClaimsKey)); authenticated {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized tenant"})
				c.Abort()
				return
			}
			tenantID = requestedTenant
		}
		if !validTenantSlug(tenantID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "x-tenant-id header required"})
			c.Abort()
			return
		}

		pool, err := tm.GetPool(c.Request.Context(), tenantID)
		if err != nil {
			// Keep infrastructure details, including connection information, out
			// of the response. The slug is only an identifier and is safe to log.
			log.Printf("Failed to get tenant pool for slug %q: %v", tenantID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found or inactive"})
			c.Abort()
			return
		}

		// Injects the tenant-specific repository into the request context
		repo := repository.New(pool)
		ctx := context.WithValue(c.Request.Context(), RepoKey, repo)
		ctx = context.WithValue(ctx, tenantPoolKey, pool)
		ctx = withTenantSlug(ctx, tenantID)
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
