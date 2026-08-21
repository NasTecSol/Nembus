package router

import (
	"testing"

	"github.com/NasTecSol/nembus-core/handler"
	"github.com/gin-gonic/gin"
)

func TestProductEnrichmentApplicationRouteIsExplicitPostOnly(t *testing.T) {
	r := gin.New()
	api := r.Group("/api")
	RegisterProductEnrichmentApplicationRoutes(api, handler.NewProductEnrichmentApplicationHandler())

	routes := r.Routes()
	if len(routes) != 1 || routes[0].Method != "POST" || routes[0].Path != "/api/product-enrichment/suggestions/:id/apply" {
		t.Fatalf("unexpected application routes: %+v", routes)
	}
}
