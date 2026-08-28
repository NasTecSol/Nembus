package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterCartRoutes registers Cart routes under /api/carts.
func RegisterCartRoutes(r *gin.RouterGroup, h *handler.CartHandler) {
	carts := r.Group("/carts")
	{
		// Cart header operations
		carts.POST("", h.CreateCart)
		carts.POST("/new", h.CreateNewCart)
		carts.GET("", h.ListActiveCarts)
		carts.GET("/abandoned", h.ListAbandonedCarts)
		carts.GET("/by-number/:cart_number", h.GetCartByNumber)
		carts.GET("/by-customer", h.GetActiveCartByCustomer)
		carts.GET("/by-guest", h.GetActiveCartByGuestIdentifier)
		carts.GET("/drafts/by-cashier/:cashier_id", h.GetDraftCartsByCashier)
		carts.GET("/:id", h.GetCart)
		carts.PUT("/:id", h.UpdateCart)
		carts.PUT("/:id/status", h.UpdateCartStatus)
		carts.PUT("/:id/customer", h.UpdateCartCustomer)
		carts.DELETE("/:id", h.DeleteCart)
		carts.POST("/expire", h.ExpireAbandonedCarts)

		// Cart item operations
		carts.GET("/:id/items", h.ListCartItems)
		carts.POST("/:id/items", h.AddToCart)
		carts.POST("/:id/items/raw", h.CreateCartItemRaw)
		carts.GET("/:id/items/by-product", h.GetCartItemByProduct)
		carts.DELETE("/:id/items", h.ClearCartItems)
		carts.GET("/:id/items/count", h.GetCartItemCount)

		// Cart totals and conversion
		carts.GET("/:id/totals", h.GetCartTotals)
		carts.POST("/:id/coupon", h.ApplyCouponToCart)
		carts.POST("/:id/recalculate", h.RecalculateCartTotals)
		carts.POST("/:id/checkout", h.ConvertToOrder)
		carts.POST("/:id/reopen", h.ReopenCart)
		carts.POST("/:id/merge", h.MergeGuestCartToCustomer)

		// Cart activity log
		carts.POST("/:id/activities", h.CreateCartActivity)
		carts.GET("/:id/activities", h.ListCartActivities)
	}

	// Cart item specific routes
	cartItems := r.Group("/cart-items")
	{
		cartItems.GET("/:item_id", h.GetCartItem)
		cartItems.PUT("/:item_id", h.UpdateCartItem)
		cartItems.PATCH("/:item_id/quantity", h.UpdateCartItemQuantity)
		cartItems.DELETE("/:item_id", h.DeleteCartItem)
	}
}
