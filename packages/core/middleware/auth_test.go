package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestJWTAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Set up temporary environment and config file for testing
	origSecret := os.Getenv("JWT_SECRET")
	jwtSecret := "test-jwt-signing-secret-key-12345"
	os.Setenv("JWT_SECRET", jwtSecret)
	defer func() {
		os.Setenv("JWT_SECRET", origSecret)
	}()

	// Backup existing config
	existingBackup := false
	if _, err := os.Stat("config/m2m_clients.json"); err == nil {
		err = os.Rename("config/m2m_clients.json", "config/m2m_clients.json.bak")
		if err != nil {
			t.Fatalf("failed to backup config: %v", err)
		}
		existingBackup = true
	}
	defer func() {
		if existingBackup {
			os.Rename("config/m2m_clients.json.bak", "config/m2m_clients.json")
		} else {
			os.Remove("config/m2m_clients.json")
		}
	}()

	// Write test registry
	testRegistry := M2MRegistry{
		Clients: []M2MClient{
			{
				ClientID:   "test-billing",
				ClientName: "Test Billing App",
				TenantID:   "tenant-test",
				Scopes:     []string{"products:read", "orders:write"},
				IsActive:   true,
			},
			{
				ClientID:   "test-inactive",
				ClientName: "Test Inactive App",
				TenantID:   "tenant-test",
				Scopes:     []string{"products:read"},
				IsActive:   false,
			},
			{
				ClientID:   "test-whitelisted-token",
				ClientName: "Test Whitelisted Token App",
				TenantID:   "tenant-test",
				Scopes:     []string{"products:read"},
				IsActive:   true,
				Token:      "will-set-this-later",
			},
		},
	}

	// Create config dir if not exists
	os.MkdirAll("config", 0755)

	// Sign a valid M2M token
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := jwt.MapClaims{
		"iss":         "nembus-api",
		"sub":         "test-billing",
		"client_id":   "test-billing",
		"client_name": "Test Billing App",
		"tenant_id":   "tenant-test",
		"scopes":      []interface{}{"products:read", "orders:write"},
		"is_m2m":      true,
		"exp":         expirationTime.Unix(),
		"iat":         time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	validM2MToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// Sign a token for inactive client
	claims["sub"] = "test-inactive"
	claims["client_id"] = "test-inactive"
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	inactiveM2MToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// Sign a token for client that is not in the JSON config
	claims["sub"] = "test-not-found"
	claims["client_id"] = "test-not-found"
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	notFoundM2MToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// Sign a token for whitelisted token app
	claims["sub"] = "test-whitelisted-token"
	claims["client_id"] = "test-whitelisted-token"
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	whitelistedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	testRegistry.Clients[2].Token = whitelistedToken

	// Save test registry
	registryData, _ := json.Marshal(testRegistry)
	err = os.WriteFile("config/m2m_clients.json", registryData, 0644)
	if err != nil {
		t.Fatalf("failed to write test registry file: %v", err)
	}

	// Create test Gin router
	r := gin.New()
	r.Use(JWTAuthMiddleware())
	r.GET("/test", func(c *gin.Context) {
		isM2M, _ := c.Get("is_m2m")
		clientID, _ := c.Get("client_id")
		clientName, _ := c.Get("client_name")
		tenantID := c.GetHeader("x-tenant-id")
		c.JSON(http.StatusOK, gin.H{
			"is_m2m":      isM2M,
			"client_id":   clientID,
			"client_name": clientName,
			"tenant_id":   tenantID,
		})
	})

	// Run test cases
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		verifyResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Valid M2M Token",
			authHeader:     "Bearer " + validM2MToken,
			expectedStatus: http.StatusOK,
			verifyResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var body map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &body)
				if body["is_m2m"] != true {
					t.Errorf("expected is_m2m to be true, got %v", body["is_m2m"])
				}
				if body["client_id"] != "test-billing" {
					t.Errorf("expected client_id test-billing, got %v", body["client_id"])
				}
				if body["tenant_id"] != "tenant-test" {
					t.Errorf("expected tenant_id tenant-test, got %v", body["tenant_id"])
				}
			},
		},
		{
			name:           "Inactive M2M Client",
			authHeader:     "Bearer " + inactiveM2MToken,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "M2M Client Not Found in Config",
			authHeader:     "Bearer " + notFoundM2MToken,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Valid Whitelisted Token M2M Client",
			authHeader:     "Bearer " + whitelistedToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid Signature M2M Token",
			authHeader:     "Bearer " + validM2MToken + "invalid",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Missing Header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Malformed Header",
			authHeader:     "InvalidFormat token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
			if tt.verifyResponse != nil {
				tt.verifyResponse(t, w)
			}
		})
	}
}
