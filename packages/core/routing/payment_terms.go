package router

import (
	"github.com/NasTecSol/nembus-core/handler"
	"github.com/gin-gonic/gin"
)

func RegisterPaymentTermRoutes(r *gin.RouterGroup, h *handler.PaymentTermsHandler) {
	group := r.Group("/payment-terms")
	{
		group.POST("", h.CreatePaymentTerm)
		group.GET("", h.ListPaymentTerms)
		group.GET("/:id", h.GetPaymentTerm)
		group.PUT("/:id", h.UpdatePaymentTerm)
		group.DELETE("/:id", h.DeletePaymentTerm)
	}
}
