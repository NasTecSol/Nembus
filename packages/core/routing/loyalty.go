package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterLoyaltyRoutes registers loyalty redemption rule routes and
// also adds loyalty-specific customer sub-routes under /customers/:id.
func RegisterLoyaltyRoutes(r *gin.RouterGroup, h *handler.LoyaltyHandler) {
	// ── Loyalty Redemption Rules ─────────────────────────────────────────────
	rules := r.Group("/loyalty-rules")
	{
		rules.POST("", h.CreateLoyaltyRule)

		rules.GET("", h.ListLoyaltyRules)
		rules.GET("/active", h.GetActiveLoyaltyRule)
		rules.GET("/:id", h.GetLoyaltyRule)

		rules.PUT("/:id", h.UpdateLoyaltyRule)
		rules.PATCH("/:id/active", h.ToggleLoyaltyRuleActive)

		rules.DELETE("/:id", h.DeleteLoyaltyRule)
	}

	// ── Customer Loyalty Endpoints ───────────────────────────────────────────
	// These extend the existing /customers group without touching customer.go
	customers := r.Group("/customers")
	{
		customers.GET("/:id/loyalty-balance", h.GetCustomerLoyaltyBalance)
		customers.PATCH("/:id/loyalty-points", h.AdjustCustomerLoyaltyPoints)
	}
}
