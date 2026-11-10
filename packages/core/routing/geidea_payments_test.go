package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NasTecSol/nembus-core/handler"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/gin-gonic/gin"
)

func TestRegisterGeideaPaymentRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handler.NewGeideaPaymentHandler(usecase.NewGeideaPaymentUseCase(nil))
	RegisterGeideaPaymentRoutes(r.Group(""), h)

	wantRoutes := map[string]bool{
		"POST /pos/payments/geidea/purchase":             false,
		"POST /pos/payments/geidea/resolve-uncertain":    false,
		"POST /pos/payments/geidea/ping":                 false,
		"POST /v1/pos/payments/geidea/purchase":          false,
		"POST /v1/pos/payments/geidea/resolve-uncertain": false,
		"POST /v1/pos/payments/geidea/ping":              false,
	}
	for _, route := range r.Routes() {
		key := route.Method + " " + route.Path
		if _, exists := wantRoutes[key]; exists {
			wantRoutes[key] = true
		}
	}
	for route, found := range wantRoutes {
		if !found {
			t.Errorf("route not registered: %s", route)
		}
	}

	for _, path := range []string{
		"/pos/payments/geidea/ping",
		"/v1/pos/payments/geidea/ping",
	} {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Errorf("POST %s returned %d, want %d: %s", path, res.Code, http.StatusOK, res.Body.String())
		}
	}
}
