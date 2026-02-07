package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterStoreRoutes(r *gin.RouterGroup, h *handler.StoreHandler) {
	store := r.Group("/stores")
	{
		// CRUD routes
		store.POST("", h.CreateStore)             // Create a new store
		store.GET("/:store_id", h.GetStore)       // Get store by ID
		store.GET("", h.ListStores)               // List all stores (with pagination/filtering)
		store.PATCH("/:store_id", h.UpdateStore)  // Update store
		store.DELETE("/:store_id", h.DeleteStore) // Delete store by ID

		// Specialized routes
		store.GET("/pos-enabled", h.ListPOSEnabledStores)                // List all POS enabled stores
		store.GET("/warehouse", h.ListWarehouseStores)                   // List all warehouse stores
		store.GET("/by-parent/:parent_id", h.ListStoresByParent)         // List stores by parent store
		store.GET("/:store_id/hierarchy", h.GetStorageLocationHierarchy) // Get storage location hierarchy
	}
}
