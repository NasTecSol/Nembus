package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/handler"
	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestM2MTokenAuthentication(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("JWT_SECRET", "test-jwt-secret")
	r := gin.New()
	protected := r.Group("/api", middleware.JWTAuthMiddleware())
	RegisterM2MRoutes(protected, r.Group("/api"), handler.NewM2MHandler())

	for _, tc := range []struct {
		name, method, key, tenant, authorization string
		status                                   int
	}{
		{"issue without JWT", "POST", middleware.M2MAPIKey, "stg", "", http.StatusOK},
		{"missing key", "POST", "", "stg", "", http.StatusUnauthorized},
		{"wrong key", "POST", "wrong", "stg", "", http.StatusUnauthorized},
		{"JWT cannot replace key", "POST", "", "stg", "Bearer irrelevant", http.StatusUnauthorized},
		{"missing tenant", "POST", middleware.M2MAPIKey, "", "", http.StatusBadRequest},
		{"listing requires JWT", "GET", middleware.M2MAPIKey, "stg", "", http.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/api/m2m/tokens", strings.NewReader(`{"client_id":"pos-app-stg","client_name":"Staging POS","scopes":["products:read"],"years":2}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("x-m2m-key", tc.key)
			req.Header.Set("x-tenant-id", tc.tenant)
			req.Header.Set("Authorization", tc.authorization)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", w.Code, tc.status, w.Body.String())
			}
			if tc.method == "POST" && w.Header().Get("Cache-Control") != "no-store" {
				t.Error("token issuance must disable caching")
			}
			if tc.status == http.StatusOK {
				var client middleware.M2MClient
				if err := json.Unmarshal(w.Body.Bytes(), &client); err != nil {
					t.Fatal(err)
				}
				if client.ClientID != "pos-app-stg" || client.TenantID != "stg" || client.Token == "" || !client.IsActive {
					t.Fatal("incorrect issued client")
				}
				token, err := jwt.Parse(client.Token, func(token *jwt.Token) (interface{}, error) { return []byte("test-jwt-secret"), nil }, jwt.WithValidMethods([]string{"HS256"}))
				if err != nil {
					t.Fatal(err)
				}
				claims := token.Claims.(jwt.MapClaims)
				if claims["tenant_id"] != "stg" || claims["client_id"] != client.ClientID || claims["is_m2m"] != true {
					t.Fatal("incorrect claims")
				}
				clients, err := middleware.LoadM2MClients()
				if err != nil || len(clients) != 1 || clients[0].Token != client.Token {
					t.Fatal("issued token was not saved")
				}
				// Use the newly issued token to authenticate a subsequent request.
				authReq := httptest.NewRequest("GET", "/api/m2m/tokens", nil)
				authReq.Header.Set("Authorization", "Bearer "+client.Token)
				authReq.Header.Set("x-tenant-id", "stg")
				authW := httptest.NewRecorder()
				r.ServeHTTP(authW, authReq)
				if authW.Code != http.StatusOK {
					t.Fatalf("issued token rejected: %s", authW.Body.String())
				}
				var listed []middleware.M2MClient
				if err := json.Unmarshal(authW.Body.Bytes(), &listed); err != nil {
					t.Fatal(err)
				}
				if len(listed) != 1 || listed[0].Token != client.Token {
					t.Fatal("listing returned incorrect token")
				}
			}
		})
	}
}
