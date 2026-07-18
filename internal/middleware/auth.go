package middleware

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
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

			// Load clients from config
			clients, err := loadM2MClients()
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
