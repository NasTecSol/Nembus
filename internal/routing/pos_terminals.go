package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

// RegisterPosTerminalsRoutes registers POS terminal routes under /api/pos.
func RegisterPosTerminalsRoutes(r *gin.RouterGroup, h *handler.PosTerminalsHandler) {
	pos := r.Group("/pos")
	// Global terminals
	terminals := pos.Group("/terminals")
	{
		terminals.GET("", h.ListPOSTerminals)
		terminals.POST("", h.CreatePOSTerminal)
		terminals.GET("/:id", h.GetPOSTerminal)
		terminals.PUT("/:id", h.UpdatePOSTerminal)
		terminals.DELETE("/:id", h.DeletePOSTerminal)
		terminals.PATCH("/:id/active", h.TogglePOSTerminalActive)
	}
	// Store-specific terminals
	stores := pos.Group("/stores/:store_id")
	storeTerminals := stores.Group("/terminals")
	{
		storeTerminals.GET("", h.ListPOSTerminalsByStore)
		storeTerminals.GET("/active", h.ListActivePOSTerminalsByStore)
		storeTerminals.GET("/code/:code", h.GetPOSTerminalByCode)
	}
}
