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
		zatca.GET("/receipt-qr", h.GetReceiptQR)
		zatca.GET("/configs", h.GetConfigsDelta)    // Delta-Fetch (Pull: Cloud -> POS)
		zatca.POST("/sync/push", h.ReceivePushSync) // Outbox (Push: POS -> Cloud)
		zatca.POST("/onboard", h.OnboardDevice)     // Device Onboarding (FATOORA CSID)
		zatca.POST("/devices/:id/complete-onboarding", h.CompleteDeviceOnboarding)
		zatca.POST("/transactions/:id/submit", h.SubmitPosTransaction)
		zatca.POST("/transactions/:id/check", h.CheckPosTransaction)
		zatca.POST("/orders/:id/generate", h.GenerateSalesOrder)
		zatca.POST("/orders/:id/check", h.CheckSalesOrder)
		zatca.POST("/orders/:id/submit", h.SubmitSalesOrder)
		zatca.GET("/terminals/:id", h.GetTerminalZatcaConfig) // Terminal ZATCA status
	}
}
