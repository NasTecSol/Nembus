package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterMenuRoutes(r *gin.RouterGroup, h *handler.MenuHandler) {
	menu := r.Group("/menus")
	{
		menu.POST("", h.CreateMenu)
		menu.GET("", h.ListMenus)
		menu.GET("/:id", h.GetMenu)
		menu.GET("/module/:moduleId", h.ListMenusByModule)
		menu.GET("/parent/:parentId", h.ListMenusByParent)
		menu.PATCH("/:id/toggle-active", h.ToggleMenuActive)
		menu.PUT("/:id", h.UpdateMenu)
		menu.DELETE("/:id", h.DeleteMenu)
	}
}
