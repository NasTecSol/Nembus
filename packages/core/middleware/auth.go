package middleware

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const UserIDKey contextKey = "user_id"
const ClaimsKey contextKey = "jwt_claims"

type M2MClient struct {
	ClientID   string   `json:"client_id"`
	ClientName string   `json:"client_name"`
	TenantID   string   `json:"tenant_id"`
	Scopes     []string `json:"scopes"`
	IsActive   bool     `json:"is_active"`
	Token      string   `json:"token"`
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
func GenerateM2MToken(clientID, clientName, tenantID string, scopes []string, durationYears int) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET not configured")
	}

	expirationTime := time.Now().AddDate(durationYears, 0, 0)

	claims := jwt.MapClaims{
		"iss":         "nembus-api",
		"sub":         clientID,
		"client_id":   clientID,
		"client_name": clientName,
		"tenant_id":   tenantID,
		"scopes":      scopes,
		"is_m2m":      true,
		"exp":         expirationTime.Unix(),
		"iat":         time.Now().Unix(),
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
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

			// Dynamically set x-tenant-id header so TenantMiddleware resolves it
			c.Request.Header.Set("x-tenant-id", matchedClient.TenantID)
		} else {
			// Standard User authentication
			c.Set("is_m2m", false)
			if userID, ok := claims["user_id"].(string); ok {
				c.Set(string(UserIDKey), userID)
			}
		}

		c.Set(string(ClaimsKey), claims)

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

// GetClaimsFromContext extracts JWT claims from Gin context
func GetClaimsFromContext(c *gin.Context) (jwt.MapClaims, bool) {
	claims, exists := c.Get(string(ClaimsKey))
	if !exists {
		return nil, false
	}
	claimsMap, ok := claims.(jwt.MapClaims)
	return claimsMap, ok
}

// GenerateJWTToken generates a JWT token for a user
func GenerateJWTToken(userID string, userLogin string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET not configured")
	}

	// Set token expiration (24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create claims
	claims := jwt.MapClaims{
		"user_id":    userID,
		"user_login": userLogin,
		"exp":        expirationTime.Unix(),
		"iat":        time.Now().Unix(),
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

// GenerateDevToken generates a development token with custom user ID and login
// This is a convenience function for testing
func GenerateDevToken(userID, userLogin string) (string, error) {
	return GenerateJWTToken(userID, userLogin)
}

// GenerateAuthorizationToken mints a short-lived JWT for supervisor override actions.
// The token is scoped to a specific permission code and optionally to a cashier ID.
// Default expiry is 5 minutes — enough time for the cashier to complete the action.
func GenerateAuthorizationToken(supervisorID int32, supervisorLogin, permissionCode string, cashierID *int32) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET not configured")
	}

	expiresAt := time.Now().Add(5 * time.Minute)

	claims := jwt.MapClaims{
		"is_authorization_token": true,
		"authorized_by_user_id":  supervisorID,
		"authorized_by_login":    supervisorLogin,
		"permission_code":        permissionCode,
		"exp":                    expiresAt.Unix(),
		"iat":                    time.Now().Unix(),
	}
	if cashierID != nil {
		claims["cashier_id"] = *cashierID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// RequireAuthorizationToken returns a Gin middleware that enforces the presence of a valid
// supervisor authorization token in the X-Authorization-Token header.
// It checks:
//  1. The token is a valid JWT signed with JWT_SECRET
//  2. The claim "is_authorization_token" is true
//  3. The claim "permission_code" matches the required code
//  4. The token has not expired
//
// On success it stores "authorized_by_user_id" and "authorized_by_login" in the Gin context.
func RequireAuthorizationToken(permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("X-Authorization-Token")
		if tokenString == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "supervisor authorization required — provide X-Authorization-Token header",
			})
			c.Abort()
			return
		}

		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT_SECRET not configured"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid or expired authorization token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "malformed authorization token"})
			c.Abort()
			return
		}

		// Must be an authorization token, not a regular user session token
		isAuthToken, _ := claims["is_authorization_token"].(bool)
		if !isAuthToken {
			c.JSON(http.StatusForbidden, gin.H{"error": "token is not a supervisor authorization token"})
			c.Abort()
			return
		}

		// Permission code in the token must match the required code for this route
		tokenPermCode, _ := claims["permission_code"].(string)
		if tokenPermCode != permissionCode {
			c.JSON(http.StatusForbidden, gin.H{
				"error":             "authorization token does not cover the required permission",
				"required":          permissionCode,
				"token_covers":      tokenPermCode,
			})
			c.Abort()
			return
		}

		// Inject supervisor context for downstream handlers / audit logging
		if userID, ok := claims["authorized_by_user_id"].(float64); ok {
			c.Set("authorized_by_user_id", int32(userID))
		}
		if login, ok := claims["authorized_by_login"].(string); ok {
			c.Set("authorized_by_login", login)
		}
		if cashierID, ok := claims["cashier_id"].(float64); ok {
			c.Set("authorization_cashier_id", int32(cashierID))
		}

		c.Next()
	}
}

