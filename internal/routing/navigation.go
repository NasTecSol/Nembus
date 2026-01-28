package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterNavigationRoutes(r *gin.RouterGroup, h *handler.NavigationHandler) {
	navigation := r.Group("/navigation")
	{
		navigation.GET("/user/:user_id", h.GetUserNavigation)
	}
}
