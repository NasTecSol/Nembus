package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

// RegisterOrderRoutes registers Order routes under /api/orders.
func RegisterOrderRoutes(r *gin.RouterGroup, h *handler.OrderHandler) {
	orders := r.Group("/orders")
	{
		// Order header operations
		orders.GET("", h.ListOrders)
		orders.POST("", h.CreateOrder)
		orders.GET("/by-number/:order_number", h.GetOrderByNumber)
		orders.GET("/:id", h.GetOrder)
		orders.PUT("/:id", h.UpdateOrder)
		orders.PUT("/:id/status", h.UpdateOrderStatus)
		orders.PUT("/:id/payment-status", h.UpdateOrderPaymentStatus)
		orders.PUT("/:id/fulfillment-status", h.UpdateOrderFulfillmentStatus)
		orders.PUT("/:id/totals", h.UpdateOrderTotals)
		orders.PUT("/:id/delivery", h.UpdateOrderDelivery)
		orders.PUT("/:id/assign", h.AssignOrder)
		orders.POST("/:id/cancel", h.CancelOrder)
		orders.DELETE("/:id", h.DeleteOrder)

		// Order lines
		orders.GET("/:id/lines", h.ListOrderLines)
		orders.POST("/:id/lines", h.CreateOrderLine)
		orders.GET("/:id/lines/totals", h.GetOrderLineTotals)
		orders.GET("/:id/lines/margin", h.GetOrderLineMargin)

		// Status history
		orders.POST("/:id/status-history", h.CreateOrderStatusHistory)
		orders.GET("/:id/status-history", h.ListOrderStatusHistory)
	}

	// Order line specific routes
	orderLines := r.Group("/order-lines")
	{
		orderLines.GET("/:line_id", h.GetOrderLine)
		orderLines.PUT("/:line_id", h.UpdateOrderLine)
		orderLines.PATCH("/:line_id/fulfillment", h.UpdateOrderLineFulfillment)
		orderLines.PATCH("/:line_id/status", h.UpdateOrderLineStatus)
		orderLines.DELETE("/:line_id", h.DeleteOrderLine)
	}

	// Fulfillment routes
	fulfillments := r.Group("/order-fulfillments")
	{
		fulfillments.GET("/:id", h.GetOrderFulfillment)
		fulfillments.GET("/by-number/:fulfillment_number", h.GetOrderFulfillmentByNumber)
		fulfillments.PUT("/:id", h.UpdateOrderFulfillment)
		fulfillments.PUT("/:id/shipment", h.UpdateFulfillmentShipment)
		fulfillments.PUT("/:id/pick-pack", h.UpdateFulfillmentPickPack)
		fulfillments.DELETE("/:id", h.DeleteOrderFulfillment)

		fulfillments.POST("/:id/items", h.CreateOrderFulfillmentItem)
		fulfillments.GET("/:id/items", h.ListOrderFulfillmentItems)
	}

	// Fulfillment item routes
	fulfillmentItems := r.Group("/order-fulfillment-items")
	{
		fulfillmentItems.DELETE("/:item_id", h.DeleteOrderFulfillmentItem)
	}
}
