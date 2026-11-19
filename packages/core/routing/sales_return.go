package router

import (
	"github.com/NasTecSol/nembus-core/handler"
	"github.com/NasTecSol/nembus-core/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterSalesReturnRoutes registers sales return routes under /api/pos/returns.
// All return operations require a supervisor authorization token in the
// X-Authorization-Token header (obtained via POST /api/auth/authorize-action).
func RegisterSalesReturnRoutes(r *gin.RouterGroup, h *handler.SalesReturnHandler) {
	returns := r.Group("/pos/returns")
	{
		returns.GET("/unposted", h.ListUnpostedSalesReturnsForSAP)
		returns.GET("/:id/lines", h.ListSalesReturnLinesForSAP)
		returns.POST("/:id/sap-sync", h.MarkSalesReturnSAPSynced)
		returns.POST("",
			middleware.RequireAuthorizationToken("sale:refund_order"),
			h.ProcessReturn,
		)
	}
}
