package router

import (
	"github.com/NasTecSol/nembus-core/handler"
	"github.com/gin-gonic/gin"
)

func RegisterProductEnrichmentApplicationRoutes(r *gin.RouterGroup, h *handler.ProductEnrichmentApplicationHandler) {
	suggestions := r.Group("/product-enrichment/suggestions")
	{
		suggestions.POST("/:id/apply", h.ApplySuggestion)
	}
}
