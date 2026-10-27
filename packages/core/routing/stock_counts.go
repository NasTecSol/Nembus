package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterStockCountRoutes registers stock count routes under /api/stock-counts.
func RegisterStockCountRoutes(r *gin.RouterGroup, h *handler.StockCountsHandler) {
	sc := r.Group("/stock-counts")
	{
		// CRUD Header
		sc.POST("", h.CreateStockCount)
		sc.GET("", h.ListStockCounts)
		sc.GET("/:id", h.GetStockCount)
		sc.PUT("/:id", h.UpdateStockCount)
		sc.DELETE("/:id", h.DeleteStockCount)

		// Workflow Status Transitions
		sc.POST("/:id/start", h.StartStockCount)
		sc.POST("/:id/complete", h.CompleteStockCount)
		sc.POST("/:id/approve", h.ApproveStockCount)
		sc.POST("/:id/reconcile", h.ReconcileStockCount)

		// Variance Summary
		sc.GET("/:id/summary", h.GetStockCountSummary)

		// Line Items Operations
		sc.POST("/:id/lines", h.AddStockCountLine)
		sc.GET("/:id/lines", h.ListStockCountLines)
		sc.GET("/:id/lines/:line_id", h.GetStockCountLine)
		sc.PUT("/:id/lines/:line_id", h.UpdateStockCountLine)
		sc.DELETE("/:id/lines/:line_id", h.DeleteStockCountLine)
		sc.PUT("/:id/lines", h.BulkUpdateStockCountLines)
	}
}
