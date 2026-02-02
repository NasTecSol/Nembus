package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

// RegisterPosRoutes registers POS product routes under /api/pos.
func RegisterPosRoutes(r *gin.RouterGroup, h *handler.PosHandler) {
	pos := r.Group("/pos")
	stores := pos.Group("/stores/:store_id")
	products := stores.Group("/products")
	{
		// GET /api/pos/stores/:store_id/products - list products (query: category_id, search_term, include_out_of_stock)
		products.GET("", h.ListProducts)
		// GET /api/pos/stores/:store_id/products/search?q=...&limit=... - search by barcode, id, or name (must be before :category_id)
		products.GET("/search", h.SearchProduct)
		// GET /api/pos/stores/:store_id/products/category/:category_id (query: include_subcategories)
		products.GET("/category/:category_id", h.GetProductsByCategory)
	}
}
