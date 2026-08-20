package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/NasTecSol/nembus-core/middleware"

	"github.com/gin-gonic/gin"
)

func TestM2MHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Set up temporary environment
	origSecret := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", "test-secret-key-12345")
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

	// Ensure config folder exists
	os.MkdirAll("config", 0755)
	initialRegistry := middleware.M2MRegistry{
		Clients: []middleware.M2MClient{
			{
				ClientID:   "other-tenant-client",
				ClientName: "Other Tenant App",
				TenantID:   "other-tenant",
				Scopes:     []string{"products:read"},
				IsActive:   true,
				Token:      "some-token",
			},
		},
	}
	initialData, _ := json.Marshal(initialRegistry)
	os.WriteFile("config/m2m_clients.json", initialData, 0644)

	// Initialize handler
	h := NewM2MHandler()

	// Initialize Gin router
	r := gin.New()
	api := r.Group("/api")
	{
		api.POST("/m2m/tokens", h.CreateToken)
		api.GET("/m2m/tokens", h.ListTokens)
	}

	t.Run("Create Token Successful", func(t *testing.T) {
		reqBody := CreateM2MRequest{
			ClientID:       "test-billing-app",
			ClientName:     "Billing Service",
			OrganizationID: 1,
			Scopes:         []string{"products:read", "orders:write"},
			Years:          2,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/m2m/tokens", bytes.NewBuffer(bodyBytes))
		req.Header.Set("x-tenant-id", "qitaf-tenant")
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var resp middleware.M2MClient
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if resp.ClientID != "test-billing-app" {
			t.Errorf("expected client_id test-billing-app, got %s", resp.ClientID)
		}
		if resp.TenantID != "qitaf-tenant" {
			t.Errorf("expected tenant_id qitaf-tenant, got %s", resp.TenantID)
		}
		if len(resp.Token) == 0 {
			t.Error("expected generated token to be non-empty")
		}
	})

	t.Run("Create Token Missing Tenant", func(t *testing.T) {
		reqBody := CreateM2MRequest{
			ClientID:   "test-billing-app",
			ClientName: "Billing Service",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/m2m/tokens", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("List Tokens Filtering by Tenant", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/m2m/tokens", nil)
		req.Header.Set("x-tenant-id", "qitaf-tenant")

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp []middleware.M2MClient
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		// Should only find the one we created for qitaf-tenant, not the other-tenant-client
		if len(resp) != 1 {
			t.Fatalf("expected 1 client, got %d", len(resp))
		}

		if resp[0].ClientID != "test-billing-app" {
			t.Errorf("expected test-billing-app, got %s", resp[0].ClientID)
		}
	})
}
