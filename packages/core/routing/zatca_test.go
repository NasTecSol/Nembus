package router

import (
	"testing"

	"github.com/NasTecSol/nembus-core/handler"
	"github.com/gin-gonic/gin"
)

func TestRegisterZatcaRoutes(t *testing.T) {
	engine := gin.New()
	RegisterZatcaRoutes(engine.Group("/api"), handler.NewZatcaHandler(nil))
	for _, path := range []string{"/api/zatca/documents", "/api/zatca/orders/:id/status"} {
		count := 0
		for _, route := range engine.Routes() {
			if route.Method == "GET" && route.Path == path {
				count++
			}
		}
		if count != 1 {
			t.Fatalf("GET %s registered %d times, want 1", path, count)
		}
	}
}
