package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterInventoryStockRoutes registers inventory stock routes under /api/inventory-stock.
func RegisterInventoryStockRoutes(r *gin.RouterGroup, h *handler.InventoryStockHandler) {
	is := r.Group("/inventory-stock")
	{
		// CRUD operations
		is.POST("", h.CreateInventoryStock)
		is.GET("", h.ListInventoryStock)
		is.GET("/:id", h.GetInventoryStock)
		is.PUT("/:id", h.UpdateInventoryStock)
		is.DELETE("/:id", h.DeleteInventoryStock)

		// Upsert
		is.POST("/upsert", h.UpsertInventoryStock)

		// Adjustments
		is.POST("/:id/adjust", h.AdjustInventoryStock)
		is.POST("/adjust", h.AdjustInventoryStockByProductAndStore)

		// Query operations
		is.GET("/product-store", h.GetInventoryStockByProductAndStore)

		// Store-specific routes
		is.GET("/store/:store_id", h.ListInventoryStockByStore)
		is.GET("/store/:store_id/location", h.ListInventoryStockByStoreAndLocation)
		is.GET("/store/:store_id/summary", h.GetInventoryStockSummary)

		// Product-specific routes
		is.GET("/product/:product_id", h.ListInventoryStockByProduct)

		// Storage location-specific routes
		is.GET("/storage-location/:storage_location_id", h.ListInventoryStockByStorageLocation)
	}
}
