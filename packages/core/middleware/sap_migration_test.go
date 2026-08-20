package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func runSAPMigrationAuthTest(t *testing.T, claims jwt.MapClaims, requestedTenant, requestedOrganization string, m2m bool) int {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("is_m2m", m2m)
		c.Set("client_id", "sap-agent")
		registeredOrganizationID, _ := claimInt32(claims["organization_id"])
		c.Set("m2m_registered_organization_id", registeredOrganizationID)
		c.Set(string(ClaimsKey), claims)
		c.Set("m2m_requested_tenant", requestedTenant)
		c.Next()
	}, SAPMigrationAuthMiddleware())
	r.POST("/migration", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodPost, "/migration", nil)
	if requestedTenant != "" {
		req.Header.Set("x-tenant-id", requestedTenant)
	}
	if requestedOrganization != "" {
		req.Header.Set("x-organization-id", requestedOrganization)
	}
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)
	return response.Code
}

func TestSAPMigrationAuthRequiresBoundMachineClaims(t *testing.T) {
	valid := jwt.MapClaims{"tenant_slug": "tenant-a", "organization_id": float64(1), "token_type": "machine"}
	for _, tc := range []struct {
		name   string
		claims jwt.MapClaims
		m2m    bool
		want   int
	}{
		{name: "valid machine", claims: valid, m2m: true, want: http.StatusOK},
		{name: "missing tenant slug", claims: jwt.MapClaims{"organization_id": float64(1)}, m2m: true, want: http.StatusUnauthorized},
		{name: "missing organization", claims: jwt.MapClaims{"tenant_slug": "tenant-a"}, m2m: true, want: http.StatusUnauthorized},
		{name: "malformed organization", claims: jwt.MapClaims{"tenant_slug": "tenant-a", "organization_id": "one"}, m2m: true, want: http.StatusUnauthorized},
		{name: "ordinary user token", claims: jwt.MapClaims{"tenant_slug": "tenant-a", "organization_id": float64(1), "token_type": "machine"}, m2m: false, want: http.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := runSAPMigrationAuthTest(t, tc.claims, "", "", tc.m2m); got != tc.want {
				t.Fatalf("status = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestSAPMigrationAuthRejectsConsistencyMismatches(t *testing.T) {
	claims := jwt.MapClaims{"tenant_slug": "tenant-a", "organization_id": float64(5), "token_type": "machine"}
	if got := runSAPMigrationAuthTest(t, claims, "tenant-b", "5", true); got != http.StatusForbidden {
		t.Fatalf("tenant mismatch status = %d, want %d", got, http.StatusForbidden)
	}
	if got := runSAPMigrationAuthTest(t, claims, "tenant-a", "6", true); got != http.StatusForbidden {
		t.Fatalf("organization mismatch status = %d, want %d", got, http.StatusForbidden)
	}
	if got := runSAPMigrationAuthTest(t, claims, "tenant-a", "not-an-id", true); got != http.StatusBadRequest {
		t.Fatalf("malformed organization header status = %d, want %d", got, http.StatusBadRequest)
	}
}
