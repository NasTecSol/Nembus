package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterM2MRoutes(r *gin.RouterGroup, h *handler.M2MHandler) {
	r.POST("/m2m/tokens", h.CreateToken)
	r.GET("/m2m/tokens", h.ListTokens)
}
