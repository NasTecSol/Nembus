package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterProductVariantRoutes(r *gin.RouterGroup, h *handler.ProductVariantHandler) {
	pv := r.Group("/product-variants")
	{
		// Static & nested routes first
		pv.GET("", h.ListProductVariants)
		pv.GET("/search", h.SearchProductVariants)
		pv.GET("/product/:product_id", h.ListProductVariantsByProduct)
		pv.GET("/active/:product_id", h.ListActiveProductVariantsByProduct)
		pv.GET("/by-sku", h.GetProductVariantBySKU)

		// Parameter routes last
		pv.GET("/variant/:variant_id", h.GetProductVariant)
		pv.PUT("/variant/:variant_id", h.UpdateProductVariant)
		pv.DELETE("/variant/:variant_id", h.DeleteProductVariant)
		pv.PATCH("/variant/:variant_id/toggle-active", h.ToggleProductVariantActive)

		// Create
		pv.POST("", h.CreateProductVariant)
	}

}
