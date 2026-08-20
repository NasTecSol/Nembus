package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/middleware/manager"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func signStandardTestToken(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()
	if _, ok := claims["exp"]; !ok {
		claims["exp"] = time.Now().Add(time.Hour).Unix()
	}
	if _, ok := claims["iat"]; !ok {
		claims["iat"] = time.Now().Unix()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}
	return signed
}

func TestGenerateJWTTokenIsTenantBound(t *testing.T) {
	originalSecret := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", "tenant-binding-test-secret")
	defer os.Setenv("JWT_SECRET", originalSecret)

	tokenString, err := GenerateJWTToken("42", "alice", "tenant-a")
	if err != nil {
		t.Fatalf("GenerateJWTToken returned error: %v", err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("tenant-binding-test-secret"), nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("generated token could not be parsed: %v", err)
	}
	claims := token.Claims.(jwt.MapClaims)
	if claims["tenant_slug"] != "tenant-a" {
		t.Fatalf("expected tenant_slug tenant-a, got %v", claims["tenant_slug"])
	}
	if claims["user_id"] != "42" || claims["user_login"] != "alice" {
		t.Fatalf("existing user claims changed: %#v", claims)
	}
	if _, err := GenerateJWTToken("42", "alice", ""); err == nil {
		t.Fatal("expected empty tenant slug to be rejected")
	}
	if _, err := GenerateJWTToken("42", "alice", "tenant-a"); err != nil {
		t.Fatalf("trusted tenant context value should be accepted: %v", err)
	}
	if stringClaims, ok := claims["tenant_slug"].(string); !ok || stringClaims == "postgres://secret" {
		t.Fatal("tenant claim must not contain database configuration")
	}
}

func TestGenerateM2MTokenRequiresTenantAndOrganizationBinding(t *testing.T) {
	originalSecret := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", "m2m-binding-test-secret")
	defer os.Setenv("JWT_SECRET", originalSecret)

	tokenString, err := GenerateM2MToken("sap-agent", "SAP Agent", "tenant-a", 5, []string{"sap:migration"}, 1)
	if err != nil {
		t.Fatalf("GenerateM2MToken returned error: %v", err)
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("m2m-binding-test-secret"), nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("generated M2M token could not be parsed: %v", err)
	}
	claims := token.Claims.(jwt.MapClaims)
	if claims["tenant_slug"] != "tenant-a" || claims["organization_id"] != float64(5) || claims["token_type"] != "machine" {
		t.Fatalf("unexpected M2M claims: %#v", claims)
	}
	if _, err := GenerateM2MToken("sap-agent", "SAP Agent", "tenant-a", 0, nil, 1); err == nil {
		t.Fatal("expected missing organization binding to be rejected")
	}
	if _, err := GenerateM2MToken("sap-agent", "SAP Agent", "", 5, nil, 1); err == nil {
		t.Fatal("expected missing tenant binding to be rejected")
	}
}

func TestTenantSlugComesFromTrustedContext(t *testing.T) {
	ctx := withTenantSlug(context.Background(), "tenant-a")
	slug, ok := TenantSlugFromContext(ctx)
	if !ok || slug != "tenant-a" {
		t.Fatalf("expected trusted tenant-a context, got %q, %v", slug, ok)
	}

	originalSecret := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", "trusted-context-secret")
	defer os.Setenv("JWT_SECRET", originalSecret)
	tokenString, err := GenerateJWTToken("7", "alice", slug)
	if err != nil {
		t.Fatalf("failed to issue token from trusted tenant context: %v", err)
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("trusted-context-secret"), nil
	})
	if err != nil || !token.Valid || token.Claims.(jwt.MapClaims)["tenant_slug"] != slug {
		t.Fatalf("token was not bound to trusted context tenant: %v", err)
	}
}

func TestTenantBindingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "tenant-binding-middleware-secret"
	originalSecret := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", secret)
	defer os.Setenv("JWT_SECRET", originalSecret)

	tests := []struct {
		name          string
		claims        jwt.MapClaims
		header        string
		expectedCode  int
		expectHandler bool
	}{
		{
			name: "same tenant passes",
			claims: jwt.MapClaims{
				"user_id": "7", "user_login": "alice", "tenant_slug": "tenant-a",
			},
			header:        "tenant-a",
			expectedCode:  http.StatusOK,
			expectHandler: true,
		},
		{
			name: "tenant b token passes only for tenant b",
			claims: jwt.MapClaims{
				"user_id": "7", "user_login": "alice", "tenant_slug": "tenant-b",
			},
			header:        "tenant-b",
			expectedCode:  http.StatusOK,
			expectHandler: true,
		},
		{
			name: "different tenant is rejected",
			claims: jwt.MapClaims{
				"user_id": "7", "user_login": "alice", "tenant_slug": "tenant-a",
			},
			header:        "tenant-b",
			expectedCode:  http.StatusUnauthorized,
			expectHandler: false,
		},
		{
			name: "old unbound token is rejected",
			claims: jwt.MapClaims{
				"user_id": "7", "user_login": "alice",
			},
			header:        "tenant-a",
			expectedCode:  http.StatusUnauthorized,
			expectHandler: false,
		},
		{
			name: "empty tenant claim is rejected",
			claims: jwt.MapClaims{
				"user_id": "7", "user_login": "alice", "tenant_slug": "",
			},
			header:        "tenant-a",
			expectedCode:  http.StatusUnauthorized,
			expectHandler: false,
		},
		{
			name: "malformed tenant claim is rejected",
			claims: jwt.MapClaims{
				"user_id": "7", "user_login": "alice", "tenant_slug": 123,
			},
			header:        "tenant-a",
			expectedCode:  http.StatusUnauthorized,
			expectHandler: false,
		},
		{
			name: "malformed header is rejected",
			claims: jwt.MapClaims{
				"user_id": "7", "user_login": "alice", "tenant_slug": "tenant-a",
			},
			header:        " tenant-a",
			expectedCode:  http.StatusUnauthorized,
			expectHandler: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			r := gin.New()
			r.Use(JWTAuthMiddleware(), TenantBindingMiddleware())
			r.GET("/protected", func(c *gin.Context) {
				called = true
				c.Status(http.StatusOK)
			})

			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			request.Header.Set("Authorization", "Bearer "+signStandardTestToken(t, secret, tt.claims))
			if tt.header != "" {
				request.Header.Set("x-tenant-id", tt.header)
			}
			response := httptest.NewRecorder()
			r.ServeHTTP(response, request)

			if response.Code != tt.expectedCode {
				t.Fatalf("expected status %d, got %d: %s", tt.expectedCode, response.Code, response.Body.String())
			}
			if called != tt.expectHandler {
				t.Fatalf("handler called=%v, want %v", called, tt.expectHandler)
			}
		})
	}
}

func TestTenantBindingRejectsBeforeTenantPoolSelection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "tenant-pool-order-secret"
	originalSecret := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", secret)
	defer os.Setenv("JWT_SECRET", originalSecret)

	called := false
	r := gin.New()
	r.Use(JWTAuthMiddleware(), TenantBindingMiddleware(), TenantMiddleware(manager.NewManager(nil)))
	r.GET("/protected", func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+signStandardTestToken(t, secret, jwt.MapClaims{
		"user_id": "7", "user_login": "alice", "tenant_slug": "tenant-a",
	}))
	request.Header.Set("x-tenant-id", "tenant-b")
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected mismatch to be unauthorized, got %d: %s", response.Code, response.Body.String())
	}
	if called {
		t.Fatal("handler was called after cross-tenant binding mismatch")
	}
}

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
