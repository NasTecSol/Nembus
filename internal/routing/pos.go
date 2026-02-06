package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

// RegisterPosRoutes registers POS product routes under /api/pos.
func RegisterPosRoutes(r *gin.RouterGroup, h *handler.PosHandler) {
	pos := r.Group("/pos")
	pos.GET("/categories", h.GetCategories)
	pos.POST("/products", h.AddProduct)
	stores := pos.Group("/stores/:store_id")
	products := stores.Group("/products")
	{
		products.GET("", h.ListProducts)
		products.GET("/search", h.SearchProduct)
		products.GET("/category/:category_id", h.GetProductsByCategory)
	}
}
