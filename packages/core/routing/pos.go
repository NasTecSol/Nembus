package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterPosRoutes registers POS product, transaction, and payment routes under /api/pos.
func RegisterPosRoutes(r *gin.RouterGroup, h *handler.PosHandler) {
	pos := r.Group("/pos")
	pos.GET("/categories", h.GetCategories)
	pos.POST("/products", h.AddProduct)
	pos.POST("/payments", h.ProcessPayment)

	payments := pos.Group("/payments")
	{
		payments.GET("/:id", h.GetPayment)
		payments.PUT("/:id", h.UpdatePayment)
		payments.DELETE("/:id", h.DeletePayment)
	}

	stores := pos.Group("/stores/:store_id")
	{
		products := stores.Group("/products")
		{
			products.GET("", h.ListProducts)
			products.GET("/search", h.SearchProduct)
			products.GET("/category/:category_id", h.GetProductsByCategory)
		}
		stores.GET("/transactions", h.ListTodayTransactions)
	}

	transactions := pos.Group("/transactions")
	{
		transactions.POST("", h.CreateTransaction)
		transactions.GET("/by-cashier-session", h.ListTransactionsByCashierSession)
		transactions.GET("/:id", h.GetTransaction)
		transactions.GET("/:id/full", h.GetTransactionFull)
		transactions.GET("/:id/payments", h.GetTransactionPayments)
		transactions.GET("/:id/payment-summary", h.GetTransactionPaymentSummary)
		transactions.POST("/:id/void", h.VoidTransaction)
	}
}
