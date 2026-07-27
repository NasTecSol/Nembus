package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterTransferRequestRoutes registers transfer request routes under /api/transfer-requests.
func RegisterTransferRequestRoutes(r *gin.RouterGroup, h *handler.TransferRequestsHandler) {
	transfers := r.Group("/transfer-requests")
	{
		transfers.POST("", h.CreateTransferRequest)
		transfers.GET("", h.ListTransferRequests)
		transfers.GET("/:id", h.GetTransferRequest)
		transfers.POST("/:id/approve", h.ApproveTransferRequest)
		transfers.POST("/:id/ship", h.ShipTransferRequest)
		transfers.POST("/:id/receive", h.ReceiveTransferRequest)
	}
}
