package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterPurchaseOrderRoutes registers purchase order routes under /api/purchase-orders and /api/purchase-order-lines.
func RegisterPurchaseOrderRoutes(r *gin.RouterGroup, h *handler.PurchaseOrdersHandler) {
	po := r.Group("/purchase-orders")
	{
		po.POST("", h.CreatePurchaseOrder)
		po.GET("", h.ListPurchaseOrders)
		po.GET("/:id", h.GetPurchaseOrder)
		po.PUT("/:id", h.UpdatePurchaseOrder)
		po.PATCH("/:id/status", h.UpdatePurchaseOrderStatus)
		po.POST("/:id/approve", h.ApprovePurchaseOrder)
		po.DELETE("/:id", h.DeletePurchaseOrder)

		// Nested Purchase Order Lines routes
		po.GET("/:id/lines", h.ListPurchaseOrderLines)
		po.POST("/:id/lines", h.AddPurchaseOrderLine)
		po.GET("/:id/lines/:line_id", h.GetPurchaseOrderLine)
		po.PUT("/:id/lines/:line_id", h.UpdatePurchaseOrderLine)
		po.DELETE("/:id/lines/:line_id", h.DeletePurchaseOrderLine)
	}

	// Standalone Purchase Order Lines routes
	poLines := r.Group("/purchase-order-lines")
	{
		poLines.GET("", h.ListAllPurchaseOrderLines)
		poLines.GET("/:id", h.GetPurchaseOrderLine)
		poLines.PUT("/:id", h.UpdatePurchaseOrderLine)
		poLines.DELETE("/:id", h.DeletePurchaseOrderLine)
	}
}

