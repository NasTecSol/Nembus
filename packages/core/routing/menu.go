package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

func RegisterMenuRoutes(r *gin.RouterGroup, h *handler.MenuHandler) {
	menu := r.Group("/menus")
	{
		menu.POST("", h.CreateMenu)
		menu.GET("", h.ListMenus)
		menu.GET("/:id", h.GetMenu)
		menu.GET("/parent/:parentId", h.ListMenusByParent)
		menu.PATCH("/:id/toggle-active", h.ToggleMenuActive)
		menu.PUT("/:id", h.UpdateMenu)
		menu.DELETE("/:id", h.DeleteMenu)
	}

	modules := r.Group("/modules")
	{
		modules.GET("/:id/menus", h.ListMenusByModule)
		modules.GET("/:id/menus/active", h.ListActiveMenusByModule)
	}
}
