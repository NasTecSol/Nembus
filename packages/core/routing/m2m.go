package router

import (
	"github.com/NasTecSol/nembus-core/handler"
	"github.com/NasTecSol/nembus-core/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterM2MRoutes uses the shared key for issuance and JWT for listing.
func RegisterM2MRoutes(protected, issuance *gin.RouterGroup, h *handler.M2MHandler) {
	issuance.POST("/m2m/tokens", middleware.M2MAPIKeyMiddleware(), h.CreateToken)
	protected.GET("/m2m/tokens", h.ListTokens)
}
