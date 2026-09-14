package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterComboBundleRoutes registers all routes for combo bundles and promotional deals.
func RegisterComboBundleRoutes(r *gin.RouterGroup, h *handler.ComboBundleHandler) {
	bundles := r.Group("/combo-bundles")
	{
		bundles.POST("", h.CreateComboBundle)
		bundles.GET("", h.ListComboBundles)
		bundles.GET("/:id", h.GetComboBundle)
		bundles.PUT("/:id", h.UpdateComboBundle)
		bundles.PATCH("/:id/toggle", h.ToggleComboBundleActive)
		bundles.DELETE("/:id", h.DeleteComboBundle)

		// Items sub-routes
		bundles.POST("/:id/items", h.AddComboBundleItem)
		bundles.GET("/:id/items", h.ListComboBundleItems)
		bundles.PUT("/:id/items/batch", h.BatchSetComboBundleItems)

		// Promotion Bridge
		bundles.POST("/:id/create-promotion", h.CreatePromotionFromCombo)
	}

	// Direct item operations
	items := r.Group("/combo-bundle-items")
	{
		items.PUT("/:item_id", h.UpdateComboBundleItem)
		items.DELETE("/:item_id", h.DeleteComboBundleItem)
	}

	// Store-specific route for POS / restaurant kiosks
	stores := r.Group("/stores/:store_id")
	{
		stores.GET("/combo-bundles", h.ListStoreComboBundles)
	}

	// Restaurant alias for backwards compatibility
	restaurant := r.Group("/restaurant")
	{
		restaurant.GET("/stores/:store_id/combo-bundles", h.ListStoreComboBundles)
	}
}
