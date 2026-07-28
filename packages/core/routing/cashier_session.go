package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

func RegisterCashierSessionRoutes(r *gin.RouterGroup, h *handler.CashierSessionHandler) {
	cashierSessions := r.Group("/cashier-sessions")
	{
		cashierSessions.POST("", h.OpenCashierSession)
		cashierSessions.GET("/active/:cashier_id", h.GetActiveCashierSession)
		cashierSessions.GET("/:id", h.GetSessionByID)
		cashierSessions.PUT("/:id/close", h.CloseCashierSession)
		cashierSessions.GET("/:id/summary", h.GetSessionSummary)
		cashierSessions.GET("/closed/:cashier_id", h.GetClosedCashierSessionsByDateRange)
	}
}
