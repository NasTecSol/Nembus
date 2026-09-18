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
		// Master catalog: full product details including variants, pricing, barcodes, UOMs, conversions
		products.GET("/master-catalog", h.GetMasterProductCatalog)
		// Master catalog search: search products with full details by SKU, name, barcode, variant, etc.
		products.GET("/master-catalog/search", h.SearchMasterProductCatalog)
	}
}
