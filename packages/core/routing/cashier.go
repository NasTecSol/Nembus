package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

func RegisterCashierRoutes(r *gin.RouterGroup, h *handler.CashierHandler) {
	cashiers := r.Group("/cashiers")
	{
		// Create operations
		cashiers.POST("", h.CreateCashier)
		cashiers.POST("/with-defaults", h.CreateCashierWithDefaults)

		// Read operations
		cashiers.GET("/all", h.ListAllCashiers)
		cashiers.GET("/active", h.ListActiveCashiers)
		cashiers.GET("", h.ListCashiers)
		cashiers.GET("/count", h.CountCashiers)
		cashiers.GET("/count/active", h.CountActiveCashiers)
		cashiers.GET("/count/store/:store_id", h.CountCashiersByStore)
		cashiers.GET("/store/:store_id", h.ListCashiersByStore)
		cashiers.GET("/store/:store_id/active", h.ListActiveCashiersByStore)
		cashiers.GET("/store/:store_id/active-with-sessions", h.ListActiveCashiersInStore)
		cashiers.GET("/:id", h.GetCashierByID)
		cashiers.GET("/code/:code", h.GetCashierByCode)
		cashiers.GET("/user/:user_id", h.GetCashierByUserID)
		cashiers.GET("/:id/exists", h.CashierExists)
		cashiers.GET("/code/:code/exists", h.CashierCodeExists)
		cashiers.GET("/:id/limits", h.GetCashierWithLimits)

		// Update operations
		cashiers.PUT("/:id", h.UpdateCashier)
		cashiers.PATCH("/:id/limits", h.UpdateCashierLimits)
		cashiers.PATCH("/:id/drawer-limit", h.UpdateCashierDrawerLimit)
		cashiers.PATCH("/:id/discount-limit", h.UpdateCashierDiscountLimit)
		cashiers.PATCH("/:id/metadata", h.UpdateCashierMetadata)
		cashiers.PATCH("/:id/activate", h.ActivateCashier)
		cashiers.PATCH("/:id/deactivate", h.DeactivateCashier)

		// Delete operations
		cashiers.DELETE("/:id", h.DeleteCashier)
		cashiers.DELETE("/:id/soft", h.SoftDeleteCashier)
	}
}
