package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

// RegisterProductPricingRoutes registers product pricing routes under /api/product-prices.
func RegisterProductPricingRoutes(r *gin.RouterGroup, h *handler.ProductPricingHandler) {
	pp := r.Group("/product-prices")
	{
		// CRUD operations
		pp.POST("", h.CreateProductPrice)
		pp.GET("/:id", h.GetProductPrice)
		pp.PUT("/:id", h.UpdateProductPrice)
		pp.DELETE("/:id", h.DeleteProductPrice)

		// Product-specific routes
		pp.GET("/product/:product_id", h.ListProductPrices)
		pp.GET("/product/:product_id/with-pricing", h.GetProductWithPricing)

		// Price list-specific routes
		pp.GET("/price-list/:price_list_id", h.ListPricesByPriceList)
		pp.POST("/price-list/:price_list_id/bulk-update", h.BulkUpdatePrices)
		pp.POST("/price-list/:price_list_id/expire", h.ExpirePrices)

		// Query routes
		pp.GET("/effective", h.GetEffectivePrice)
		pp.GET("/price-list", h.GetProductPriceForList)
		pp.GET("/comparison/:product_id", h.GetPriceComparison)
		pp.GET("/search", h.SearchProductsWithPrices)
	}
}
