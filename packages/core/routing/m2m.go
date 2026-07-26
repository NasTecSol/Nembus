package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

func RegisterM2MRoutes(r *gin.RouterGroup, h *handler.M2MHandler) {
	r.POST("/m2m/tokens", h.CreateToken)
	r.GET("/m2m/tokens", h.ListTokens)
}
