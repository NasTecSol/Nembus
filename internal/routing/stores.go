package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterStoreRoutes(r *gin.RouterGroup, h *handler.StoreHandler) {
	store := r.Group("/stores")
	{
		// CRUD routes
		store.POST("", h.CreateStore)       // Create a new store
		store.GET("", h.ListStores)         // List all stores (with pagination/filtering)
		
		// Specialized routes (register before generic :id routes)
		store.GET("/pos-enabled", h.ListPOSEnabledStores)          // List all POS enabled stores
		store.GET("/warehouse", h.ListWarehouseStores)             // List all warehouse stores
		store.GET("/by-parent/:parent_id", h.ListStoresByParent)   // List stores by parent store
		store.GET("/:id/hierarchy", h.GetStorageLocationHierarchy) // Get storage location hierarchy
		
		// Generic :id routes (register after specific routes)
		store.GET("/:id", h.GetStore)       // Get store by ID
		store.GET("/:id/active-session", h.GetActiveSession) // Get store active session
		store.PATCH("/:id", h.UpdateStore)  // Update store
		store.DELETE("/:id", h.DeleteStore) // Delete store by ID
	}
}
