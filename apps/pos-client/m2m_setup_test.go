package main

import (
	"encoding/json"
	"github.com/NasTecSol/nembus-client/internal/config"
	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLocalM2MSetup(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("JWT_SECRET", "local-setup-test-secret")
	r := setupRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &config.Config{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/m2m/tokens", nil)
	req.Header.Set("Origin", "http://wails.localhost")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent || !strings.Contains(w.Header().Get("Access-Control-Allow-Headers"), "x-m2m-key") {
		t.Fatal("local bootstrap CORS preflight failed")
	}
	req = httptest.NewRequest(http.MethodPost, "/api/m2m/tokens", strings.NewReader(`{"client_id":"local-pos-zatca-test","client_name":"Local POS ZATCA","years":1}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tenant-id", "cloned-tenant")
	req.Header.Set("x-m2m-key", middleware.M2MAPIKey)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("local issuance failed: %d %s", w.Code, w.Body.String())
	}
	var client middleware.M2MClient
	if err := json.Unmarshal(w.Body.Bytes(), &client); err != nil {
		t.Fatal(err)
	}
	if client.Token == "" || client.TenantID != "cloned-tenant" {
		t.Fatal("missing local tenant token")
	}
	r.GET("/test-local-auth", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
		if c.GetHeader("x-tenant-id") != "cloned-tenant" {
			t.Error("incorrect authenticated tenant")
		}
		c.Status(http.StatusNoContent)
	})
	req = httptest.NewRequest(http.MethodGet, "/test-local-auth", nil)
	req.Header.Set("Authorization", "Bearer "+client.Token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("local JWT rejected: %d %s", w.Code, w.Body.String())
	}
}
