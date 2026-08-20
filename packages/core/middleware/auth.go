package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const UserIDKey contextKey = "user_id"
const UserLoginKey contextKey = "user_login"
const ClaimsKey contextKey = "jwt_claims"

const tenantSlugClaim = "tenant_slug"

type M2MClient struct {
	ClientID   string `json:"client_id"`
	ClientName string `json:"client_name"`
	// TenantSlug is the canonical tenant binding for new machine credentials.
	TenantSlug string `json:"tenant_slug,omitempty"`
	// TenantID is retained only to read older registry entries. Its value is
	// also a tenant slug; it is never an organization ID.
	TenantID       string   `json:"tenant_id,omitempty"`
	OrganizationID int32    `json:"organization_id,omitempty"`
	Scopes         []string `json:"scopes"`
	IsActive       bool     `json:"is_active"`
	Token          string   `json:"token"`
}

func (c M2MClient) BoundTenantSlug() string {
	if c.TenantSlug != "" {
		return c.TenantSlug
	}
	return c.TenantID
}

type M2MRegistry struct {
	Clients []M2MClient `json:"clients"`
}

var m2mMutex sync.RWMutex

// loadM2MClients reads the JSON configuration file
func loadM2MClients() ([]M2MClient, error) {
	filePath := "config/m2m_clients.json"
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var registry M2MRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, err
	}
	return registry.Clients, nil
}

// LoadM2MClients safely loads M2M clients
func LoadM2MClients() ([]M2MClient, error) {
	m2mMutex.RLock()
	defer m2mMutex.RUnlock()
	return loadM2MClients()
}

// SaveM2MClient safely registers or updates an M2M client in the config file
func SaveM2MClient(client M2MClient) error {
	m2mMutex.Lock()
	defer m2mMutex.Unlock()

	clients, err := loadM2MClients()
	if err != nil {
		return err
	}

	updated := false
	for i, c := range clients {
		if c.ClientID == client.ClientID {
			clients[i] = client
			updated = true
			break
		}
	}

	if !updated {
		clients = append(clients, client)
	}

	registry := M2MRegistry{
		Clients: clients,
	}

	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return err
	}

	// Ensure config dir exists
	err = os.MkdirAll("config", 0755)
	if err != nil {
		return err
	}

	return os.WriteFile("config/m2m_clients.json", data, 0600)
}

// GenerateM2MToken generates a long-lived JWT token for M2M communication
func GenerateM2MToken(clientID, clientName, tenantSlug string, organizationID int32, scopes []string, durationYears int) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET not configured")
	}
	if !validTenantSlug(tenantSlug) {
		return "", errors.New("tenant slug is required")
	}
	if organizationID <= 0 {
		return "", errors.New("organization ID is required")
	}

	expirationTime := time.Now().AddDate(durationYears, 0, 0)

	claims := jwt.MapClaims{
		"iss":             "nembus-api",
		"sub":             clientID,
		"client_id":       clientID,
		"client_name":     clientName,
		"tenant_slug":     tenantSlug,
		"organization_id": organizationID,
		"token_type":      "machine",
		"scopes":          scopes,
		"is_m2m":          true,
		"exp":             expirationTime.Unix(),
		"iat":             time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// JWTAuthMiddleware validates JWT tokens from the Authorization header
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check if it starts with "Bearer "
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be in format: Bearer <token>"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Get JWT secret from environment
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT_SECRET not configured"})
			c.Abort()
			return
		}

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Check if it's an M2M Token
		if isM2M, ok := claims["is_m2m"].(bool); ok && isM2M {
			clientID, _ := claims["client_id"].(string)
			if clientID == "" {
				clientID, _ = claims["sub"].(string)
			}

			if clientID == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid M2M token: missing client_id/sub"})
				c.Abort()
				return
			}

			// Load clients from config (thread-safe)
			clients, err := LoadM2MClients()
			if err != nil {
				log.Printf("Error loading M2M clients: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
				return
			}

			var matchedClient *M2MClient
			for _, client := range clients {
				if client.ClientID == clientID {
					matchedClient = &client
					break
				}
			}

			if matchedClient == nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized client"})
				c.Abort()
				return
			}

			signedTenant, ok := claims["tenant_slug"].(string)
			if !ok {
				// Preserve authentication for older non-SAP M2M integrations. The
				// migration-specific middleware requires the new tenant_slug claim.
				signedTenant, ok = claims["tenant_id"].(string)
			}
			boundTenant := matchedClient.BoundTenantSlug()
			if !ok || !validTenantSlug(signedTenant) || !validTenantSlug(boundTenant) || signedTenant != boundTenant {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized client"})
				c.Abort()
				return
			}
			if rawOrganization, exists := claims["organization_id"]; exists {
				organizationID, valid := claimInt32(rawOrganization)
				if !valid || (matchedClient.OrganizationID > 0 && organizationID != matchedClient.OrganizationID) {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized client"})
					c.Abort()
					return
				}
			}

			if !matchedClient.IsActive {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Client is inactive"})
				c.Abort()
				return
			}

			// If client has a specific whitelisted token, check if it matches
			if matchedClient.Token != "" && matchedClient.Token != tokenString {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been rotated/revoked"})
				c.Abort()
				return
			}

			// Inject client info into context
			c.Set("is_m2m", true)
			c.Set("client_id", matchedClient.ClientID)
			c.Set("client_name", matchedClient.ClientName)
			c.Set("scopes", matchedClient.Scopes)
			c.Set("m2m_tenant_slug", signedTenant)
			c.Set("m2m_registered_organization_id", matchedClient.OrganizationID)
			if organizationID, valid := claimInt32(claims["organization_id"]); valid {
				c.Set("m2m_organization_id", organizationID)
			}

			// Preserve an explicitly supplied header for TenantBindingMiddleware to
			// compare before using the registry-bound tenant below.
			c.Set("m2m_requested_tenant", c.GetHeader("x-tenant-id"))

			// Dynamically set x-tenant-id header so existing M2M callers continue
			// to work when they omit the header.
			c.Request.Header.Set("x-tenant-id", boundTenant)
		} else {
			// Standard User authentication
			c.Set("is_m2m", false)
			userID, ok := claims["user_id"].(string)
			if !ok || userID == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}
			c.Set(string(UserIDKey), userID)
			requestContext := context.WithValue(c.Request.Context(), UserIDKey, userID)
			if userLogin, ok := claims["user_login"].(string); ok {
				c.Set(string(UserLoginKey), userLogin)
				requestContext = context.WithValue(requestContext, UserLoginKey, userLogin)
			}
			c.Request = c.Request.WithContext(requestContext)
		}

		c.Set(string(ClaimsKey), claims)

		c.Next()
	}
}

func claimInt32(value interface{}) (int32, bool) {
	var parsed int64
	switch v := value.(type) {
	case float64:
		parsed = int64(v)
		if float64(parsed) != v {
			return 0, false
		}
	case float32:
		parsed = int64(v)
		if float32(parsed) != v {
			return 0, false
		}
	case int:
		parsed = int64(v)
	case int32:
		parsed = int64(v)
	case int64:
		parsed = v
	case json.Number:
		var err error
		parsed, err = v.Int64()
		if err != nil {
			return 0, false
		}
	case string:
		var err error
		parsed, err = strconv.ParseInt(strings.TrimSpace(v), 10, 32)
		if err != nil {
			return 0, false
		}
	default:
		return 0, false
	}
	if parsed <= 0 || parsed > int64(^uint32(0)>>1) {
		return 0, false
	}
	return int32(parsed), true
}

// TrustedMachineIdentity is the only tenant/organization authority accepted
// by the SAP migration route.
type TrustedMachineIdentity struct {
	ClientID       string
	TenantSlug     string
	OrganizationID int32
}

const trustedMachineIdentityKey contextKey = "trusted_machine_identity"

func TrustedMachineIdentityFromContext(ctx context.Context) (TrustedMachineIdentity, bool) {
	identity, ok := ctx.Value(trustedMachineIdentityKey).(TrustedMachineIdentity)
	return identity, ok && validTenantSlug(identity.TenantSlug) && identity.OrganizationID > 0
}

// SAPMigrationAuthMiddleware narrows the already verified JWT protocol to a
// migration-capable, tenant/org-bound machine credential. It validates only
// optional consistency headers; claims remain authoritative.
func SAPMigrationAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isM2M, ok := c.Get("is_m2m")
		if !ok || isM2M != true {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "SAP migration requires machine authentication"})
			c.Abort()
			return
		}
		claims, ok := GetClaimsFromContext(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid machine authentication"})
			c.Abort()
			return
		}
		tenantSlug, tenantOK := claims[tenantSlugClaim].(string)
		organizationID, organizationOK := claimInt32(claims["organization_id"])
		tokenType, tokenTypeOK := claims["token_type"].(string)
		registeredOrganizationID, registeredOK := c.Get("m2m_registered_organization_id")
		registered, registeredValid := claimInt32(registeredOrganizationID)
		if !tenantOK || !validTenantSlug(tenantSlug) || !organizationOK || !tokenTypeOK || tokenType != "machine" || !registeredOK || !registeredValid || registered != organizationID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Machine credential is not bound to a tenant and organization"})
			c.Abort()
			return
		}

		if requested := c.GetString("m2m_requested_tenant"); requested != "" && requested != tenantSlug {
			c.JSON(http.StatusForbidden, gin.H{"error": "Tenant consistency check failed"})
			c.Abort()
			return
		}
		if requested := c.GetHeader("x-organization-id"); requested != "" {
			requestedID, valid := claimInt32(requested)
			if !valid {
				c.JSON(http.StatusBadRequest, gin.H{"error": "x-organization-id must be a positive integer"})
				c.Abort()
				return
			}
			if requestedID != organizationID {
				c.JSON(http.StatusForbidden, gin.H{"error": "Organization consistency check failed"})
				c.Abort()
				return
			}
		}

		identity := TrustedMachineIdentity{ClientID: c.GetString("client_id"), TenantSlug: tenantSlug, OrganizationID: organizationID}
		c.Request = c.Request.WithContext(context.WithValue(withTenantSlug(c.Request.Context(), tenantSlug), trustedMachineIdentityKey, identity))
		c.Next()
	}
}

// GetUserIDFromContext extracts user ID from Gin context
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get(string(UserIDKey))
	if !exists {
		return "", false
	}
	userIDStr, ok := userID.(string)
	return userIDStr, ok
}

// GetUserLoginFromContext extracts the authenticated login from Gin context.
func GetUserLoginFromContext(c *gin.Context) (string, bool) {
	userLogin, exists := c.Get(string(UserLoginKey))
	if !exists {
		return "", false
	}
	userLoginStr, ok := userLogin.(string)
	return userLoginStr, ok && userLoginStr != ""
}

// AuthenticatedUserID extracts the authenticated standard-user ID from a
// request context populated by JWTAuthMiddleware.
func AuthenticatedUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok && userID != ""
}

// AuthenticatedUserLogin extracts the authenticated standard-user login from
// a request context populated by JWTAuthMiddleware.
func AuthenticatedUserLogin(ctx context.Context) (string, bool) {
	userLogin, ok := ctx.Value(UserLoginKey).(string)
	return userLogin, ok && userLogin != ""
}

// GetClaimsFromContext extracts JWT claims from Gin context
func GetClaimsFromContext(c *gin.Context) (jwt.MapClaims, bool) {
	claims, exists := c.Get(string(ClaimsKey))
	if !exists {
		return nil, false
	}
	claimsMap, ok := claims.(jwt.MapClaims)
	return claimsMap, ok
}

// GenerateJWTToken generates a tenant-bound JWT token for a standard user.
func GenerateJWTToken(userID string, userLogin string, tenantSlug string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET not configured")
	}
	if !validTenantSlug(tenantSlug) {
		return "", errors.New("tenant slug is required")
	}

	// Set token expiration (24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create claims
	claims := jwt.MapClaims{
		"user_id":       userID,
		"user_login":    userLogin,
		tenantSlugClaim: tenantSlug,
		"exp":           expirationTime.Unix(),
		"iat":           time.Now().Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateDevToken generates a development token with custom user ID, login,
// and an explicitly selected tenant slug.
// This is a convenience function for testing
func GenerateDevToken(userID, userLogin, tenantSlug string) (string, error) {
	return GenerateJWTToken(userID, userLogin, tenantSlug)
}
