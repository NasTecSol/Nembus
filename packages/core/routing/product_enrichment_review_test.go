package router

import (
	"testing"

	"github.com/NasTecSol/nembus-core/handler"
	"github.com/gin-gonic/gin"
)

func TestProductEnrichmentReviewRoutesExposeReadAndReviewOperations(t *testing.T) {
	r := gin.New()
	api := r.Group("/api")
	RegisterProductEnrichmentReviewRoutes(api, handler.NewProductEnrichmentReviewHandler())

	routes := r.Routes()
	want := map[string]bool{
		"GET /api/product-enrichment/suggestions":              false,
		"GET /api/product-enrichment/suggestions/:id":          false,
		"POST /api/product-enrichment/suggestions/:id/approve": false,
		"POST /api/product-enrichment/suggestions/:id/reject":  false,
	}
	for _, route := range routes {
		key := route.Method + " " + route.Path
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}
	for route, found := range want {
		if !found {
			t.Fatalf("missing product enrichment route %s; routes=%+v", route, routes)
		}
	}
}
