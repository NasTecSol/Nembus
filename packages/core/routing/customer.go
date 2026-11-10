package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

func RegisterCustomerRoutes(r *gin.RouterGroup, h *handler.CustomerHandler) {
	customers := r.Group("/customers")
	{
		customers.POST("", h.CreateCustomer)

		customers.GET("", h.ListCustomers)
		customers.GET("/active", h.ListActiveCustomers)
		customers.GET("/search", h.SearchCustomers)
		customers.GET("/outstanding", h.GetCustomersWithOutstandingBalance)
		customers.GET("/:id", h.GetCustomerByID)
		customers.GET("/code/:code", h.GetCustomerByCode)
		customers.GET("/type/:customer_type", h.ListCustomersByType)
		customers.GET("/:id/credit-status", h.GetCustomerCreditStatus)

		customers.PUT("/:id", h.UpdateCustomer)
		customers.PATCH("/:id/active", h.ToggleCustomerActive)
		customers.PATCH("/:id/balance", h.UpdateCustomerBalance)

		customers.DELETE("/:id", h.DeleteCustomer)
	}
}
