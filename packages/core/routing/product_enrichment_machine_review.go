package router

import (
	"github.com/NasTecSol/nembus-core/handler"
	"github.com/gin-gonic/gin"
)

// RegisterProductEnrichmentMachineReviewRoutes mounts the SAP Agent proxy
// contract. The caller must place this group behind JWT, SAP machine,
// tenant-binding, and organization validation middleware.
func RegisterProductEnrichmentMachineReviewRoutes(r *gin.RouterGroup, h *handler.ProductEnrichmentReviewHandler) {
	suggestions := r.Group("/product-enrichment/suggestions")
	{
		suggestions.GET("", h.ListMachineSuggestions)
		suggestions.GET("/:id", h.GetMachineSuggestion)
		suggestions.POST("/:id/approve", h.ApproveMachineSuggestion)
		suggestions.POST("/:id/reject", h.RejectMachineSuggestion)
	}
}
