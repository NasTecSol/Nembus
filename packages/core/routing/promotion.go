package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterPromotionRoutes registers Promotion routes under /api/promotions.
func RegisterPromotionRoutes(r *gin.RouterGroup, h *handler.PromotionHandler) {
	promos := r.Group("/promotions")
	{
		// Admin CRUD
		promos.POST("", h.CreatePromotion)
		promos.GET("", h.ListAllPromotions)
		promos.GET("/active", h.ListActivePromotions)
		promos.GET("/:id", h.GetPromotion)
		promos.GET("/code/:code", h.GetPromotionByCode)
		promos.PUT("/:id", h.UpdatePromotion)
		promos.PATCH("/:id/status", h.UpdatePromotionStatus)
		promos.DELETE("/:id", h.DeletePromotion)

		// Coupon application
		promos.POST("/apply-coupon", h.ApplyCoupon)
		promos.POST("/validate-coupon", h.ValidateCoupon)
	}
}
