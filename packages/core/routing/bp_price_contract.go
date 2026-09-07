package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterBPPriceContractRoutes registers business partner price contract routes.
func RegisterBPPriceContractRoutes(r *gin.RouterGroup, h *handler.BPPriceContractHandler) {
	contracts := r.Group("/bp-price-contracts")
	{
		contracts.POST("", h.CreateBPPriceContract)
		contracts.GET("", h.ListBPPriceContracts)
		contracts.GET("/effective", h.GetEffectiveBPPriceContract)
		contracts.GET("/partner/:partner_id", h.ListBPPriceContractsByPartner)
		contracts.GET("/:id", h.GetBPPriceContract)
		contracts.PUT("/:id", h.UpdateBPPriceContract)
		contracts.DELETE("/:id", h.DeleteBPPriceContract)
		contracts.PATCH("/:id/toggle", h.ToggleBPPriceContractActive)
	}
}
