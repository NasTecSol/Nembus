package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterUOMRoutes registers Unit of Measure and product UOM conversion routes under /api.
func RegisterUOMRoutes(r *gin.RouterGroup, h *handler.UOMHandler) {
	uoms := r.Group("/uoms")
	{
		uoms.POST("", h.CreateUnitOfMeasure)
		uoms.GET("", h.ListUnitsOfMeasure)
		uoms.GET("/active", h.ListActiveUnitsOfMeasure)
		uoms.GET("/by-type", h.ListUnitsByType)
		uoms.GET("/code/:code", h.GetUnitOfMeasureByCode)
		uoms.GET("/:id", h.GetUnitOfMeasure)
		uoms.PUT("/:id", h.UpdateUnitOfMeasure)
		uoms.DELETE("/:id", h.DeleteUnitOfMeasure)
	}

	productConversions := r.Group("/products/:product_id/uom-conversions")
	{
		productConversions.POST("", h.CreateProductUOMConversion)
		productConversions.GET("", h.ListProductUOMConversions)
		productConversions.GET("/detailed", h.GetProductUOMConversionsDetailed)
		productConversions.GET("/lookup", h.GetProductUOMConversion)
	}

	conversions := r.Group("/uom-conversions")
	{
		conversions.PUT("/:id", h.UpdateProductUOMConversion)
		conversions.DELETE("/:id", h.DeleteProductUOMConversion)
	}
}

