package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

// RegisterStorageLocationsRoutes registers storage location routes under /api.
func RegisterStorageLocationsRoutes(r *gin.RouterGroup, h *handler.StorageLocationsHandler) {
	// Global storage locations
	locations := r.Group("/storage-locations")
	{
		locations.GET("", h.ListStorageLocations)
		locations.POST("", h.CreateStorageLocation)
		locations.GET("/by-parent", h.ListStorageLocationsByParent)
		locations.GET("/:id", h.GetStorageLocation)
		locations.PUT("/:id", h.UpdateStorageLocation)
		locations.DELETE("/:id", h.DeleteStorageLocation)
		locations.PATCH("/:id/active", h.ToggleStorageLocationActive)
	}
	// Store-specific storage locations
	stores := r.Group("/stores/:store_id")
	storeLocations := stores.Group("/storage-locations")
	{
		storeLocations.GET("", h.ListStorageLocationsByStore)
		storeLocations.GET("/active", h.ListActiveStorageLocationsByStore)
		storeLocations.GET("/code/:code", h.GetStorageLocationByCode)
		storeLocations.GET("/type/:location_type", h.ListStorageLocationsByType)
	}
}
