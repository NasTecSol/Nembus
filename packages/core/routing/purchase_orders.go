package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterPurchaseOrderRoutes registers purchase order routes under /api/purchase-orders.
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
	}
}
