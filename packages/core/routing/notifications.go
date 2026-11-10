package router

import (
	"github.com/NasTecSol/nembus-core/handler"
	"github.com/gin-gonic/gin"
)

func RegisterNotificationRoutes(r *gin.RouterGroup, h *handler.NotificationHandler) {
	h.RegisterRoutes(r)
}
