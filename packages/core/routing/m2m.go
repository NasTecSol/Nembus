package router

import (
	"github.com/NasTecSol/nembus-core/handler"
	"github.com/NasTecSol/nembus-core/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterM2MRoutes uses the shared key for both issuance and listing.
func RegisterM2MRoutes(api *gin.RouterGroup, h *handler.M2MHandler) {
	api.POST("/m2m/tokens", middleware.M2MAPIKeyMiddleware(), h.CreateToken)
	api.GET("/m2m/tokens", middleware.M2MAPIKeyMiddleware(), h.ListTokens)
}
