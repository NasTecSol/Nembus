package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterPriceListRoutes registers price list routes under /api/price-lists.
func RegisterPriceListRoutes(r *gin.RouterGroup, h *handler.PriceListsHandler) {
	pl := r.Group("/price-lists")
	{
		pl.POST("", h.CreatePriceList)
		pl.GET("", h.ListPriceLists)
		pl.GET("/active", h.ListActivePriceLists)
		pl.GET("/valid", h.ListValidPriceLists)
		pl.GET("/default", h.GetDefaultPriceList)
		pl.GET("/code/:code", h.GetPriceListByCode)

		pl.GET("/:id", h.GetPriceList)
		pl.PUT("/:id", h.UpdatePriceList)
		pl.DELETE("/:id", h.DeletePriceList)
		pl.POST("/:id/set-default", h.SetDefaultPriceList)
		pl.PATCH("/:id/active", h.TogglePriceListActive)
	}
}

