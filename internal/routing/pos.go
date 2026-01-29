package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterPOSRoutes(r *gin.RouterGroup, h *handler.POSHandler) {
	pos := r.Group("/pos")
	{
		pos.GET("/products", h.GetPOSProducts)
		pos.GET("/products/barcode/:barcode", h.GetPOSProductByBarcode)
		pos.GET("/products/:id", h.GetPOSProductByID)
		pos.GET("/products/search", h.SearchPOSProducts)
		pos.GET("/categories", h.GetPOSCategories)
		pos.GET("/promotions", h.GetPOSPromotedProducts)
		pos.POST("/products", h.CreateProduct)
	}
}
