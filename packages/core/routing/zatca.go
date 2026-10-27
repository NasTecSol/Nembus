package router

import (
	"github.com/NasTecSol/nembus-core/handler"
	"github.com/gin-gonic/gin"
)

// RegisterZatcaRoutes registers all ZATCA and Sync API endpoints.
func RegisterZatcaRoutes(rg *gin.RouterGroup, h *handler.ZatcaHandler) {
	zatca := rg.Group("/zatca")
	{
		zatca.GET("/status", h.GetZatcaStatus)
		zatca.GET("/configs", h.GetConfigsDelta)          // Delta-Fetch (Pull: Cloud -> POS)
		zatca.POST("/sync/push", h.ReceivePushSync)       // Outbox (Push: POS -> Cloud)
		zatca.POST("/onboard", h.OnboardDevice)           // Device Onboarding (FATOORA CSID)
		zatca.GET("/terminals/:id", h.GetTerminalZatcaConfig) // Terminal ZATCA status
	}
}

