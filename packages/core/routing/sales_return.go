package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterSalesReturnRoutes registers sales return routes under /api/pos/returns.
func RegisterSalesReturnRoutes(r *gin.RouterGroup, h *handler.SalesReturnHandler) {
	returns := r.Group("/pos/returns")
	{
		returns.POST("", h.ProcessReturn)
	}
}
