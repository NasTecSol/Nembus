package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterInvoiceRoutes registers commercial invoice routes under /api/invoices.
func RegisterInvoiceRoutes(r *gin.RouterGroup, h *handler.InvoiceHandler) {
	invoices := r.Group("/invoices")
	{
		// Invoice header operations
		invoices.GET("", h.ListInvoices)
		invoices.POST("", h.CreateInvoice)
		invoices.GET("/stats", h.GetInvoiceStats)
		invoices.GET("/overdue", h.ListOverdueInvoices)
		invoices.GET("/customer/:customer_id", h.ListCustomerInvoices)
		invoices.GET("/by-number/:invoice_number", h.GetInvoiceByNumber)
		invoices.POST("/from-order/:order_id", h.CreateInvoiceFromOrder)
		invoices.GET("/:id", h.GetInvoice)
		invoices.PUT("/:id", h.UpdateInvoice)
		invoices.PATCH("/:id/status", h.UpdateInvoiceStatus)
		invoices.DELETE("/:id", h.DeleteInvoice)

		// Invoice lines
		invoices.GET("/:id/lines", h.ListInvoiceLines)
		invoices.POST("/:id/lines", h.CreateInvoiceLine)
		invoices.DELETE("/lines/:line_id", h.DeleteInvoiceLine)

		// Invoice payments
		invoices.GET("/:id/payments", h.ListInvoicePayments)
		invoices.POST("/:id/payments", h.CreateInvoicePayment)
	}
}
