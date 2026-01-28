package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterPermissionRoutes(r *gin.RouterGroup, h *handler.PermissionHandler) {
	permissions := r.Group("/permissions")
	{
		permissions.GET("/user/:user_id/submenu/:submenu_code", h.CheckUserSubmenuPermission)
	}
}
