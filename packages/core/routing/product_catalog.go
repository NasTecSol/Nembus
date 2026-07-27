package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterProductCatalogRoutes registers admin product catalog routes.
func RegisterProductCatalogRoutes(r *gin.RouterGroup, h *handler.ProductCatalogHandler) {
	products := r.Group("/products")
	{
		// Admin catalog: products with embedded variants
		products.GET("/catalog", h.ListProductsWithVariants)
	}
}
