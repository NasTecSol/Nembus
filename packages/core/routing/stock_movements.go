package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterStockMovementRoutes registers stock movement routes under /api/stock-movements.
func RegisterStockMovementRoutes(r *gin.RouterGroup, h *handler.StockMovementHandler) {
	stock := r.Group("/stock-movements")
	{
		// Stock movement operations
		stock.POST("", h.CreateStockMovement)
		stock.GET("", h.ListStockMovements)
		stock.GET("/reference-types", h.ListStockMovementReferenceTypes)
		stock.GET("/reference-type", h.ListStockMovementsByReferenceType)
		stock.GET("/by-reference-type", h.ListStockMovementsByReferenceType)
		stock.GET("/:id", h.GetStockMovement)
		stock.PATCH("/:id/status", h.UpdateStockMovementStatus)

		// Stock movements by product
		stock.GET("/product/:productID", h.ListStockMovementsByProduct)
		stock.GET("/product/:productID/daterange", h.ListStockMovementsByProductWithDateRange)

		// Stock movements by date range
		stock.GET("/daterange", h.ListStockMovementsByDateRange)
	}
}
