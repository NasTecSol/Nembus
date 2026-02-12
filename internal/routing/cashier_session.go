package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterCashierSessionRoutes(r *gin.RouterGroup, h *handler.CashierSessionHandler) {
	cashierSessions := r.Group("/cashier-sessions")
	{
		cashierSessions.POST("", h.OpenCashierSession)
		cashierSessions.GET("/active/:cashier_id", h.GetActiveCashierSession)
		cashierSessions.PUT("/:id/close", h.CloseCashierSession)
		cashierSessions.GET("/:id/summary", h.GetSessionSummary)
	}
}
