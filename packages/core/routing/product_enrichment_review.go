package router

import (
	"github.com/NasTecSol/nembus-core/handler"
	"github.com/gin-gonic/gin"
)

func RegisterProductEnrichmentReviewRoutes(r *gin.RouterGroup, h *handler.ProductEnrichmentReviewHandler) {
	suggestions := r.Group("/product-enrichment/suggestions")
	{
		suggestions.GET("", h.ListSuggestions)
		suggestions.GET("/:id", h.GetSuggestion)
		suggestions.POST("/:id/approve", h.ApproveSuggestion)
		suggestions.POST("/:id/reject", h.RejectSuggestion)
	}
}
