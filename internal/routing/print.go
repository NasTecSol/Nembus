package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

// RegisterPrintRoutes registers the ESC/POS print endpoint.
func RegisterPrintRoutes(r *gin.RouterGroup, h *handler.PrintHandler) {
	print := r.Group("/print")
	{
		// POST /api/print/receipt – send a receipt to the thermal printer
		print.POST("/receipt", h.PrintReceipt)
	}
}
