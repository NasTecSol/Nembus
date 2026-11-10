package router

import (
	"github.com/NasTecSol/nembus-core/handler"
	"github.com/gin-gonic/gin"
)

// RegisterGeideaPaymentRoutes registers Geidea Local ECR terminal payment routes.
func RegisterGeideaPaymentRoutes(r *gin.RouterGroup, h *handler.GeideaPaymentHandler) {
	geidea := r.Group("/pos/payments/geidea")
	{
		geidea.POST("/purchase", h.Purchase)
		geidea.POST("/resolve-uncertain", h.ResolveUncertain)
		geidea.POST("/ping", h.Ping)
	}

	v1 := r.Group("/v1/pos/payments/geidea")
	{
		v1.POST("/purchase", h.Purchase)
		v1.POST("/resolve-uncertain", h.ResolveUncertain)
		v1.POST("/ping", h.Ping)
	}
}
