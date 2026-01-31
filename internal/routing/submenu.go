package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterSubmenuRoutes(r *gin.RouterGroup, h *handler.SubmenuHandler) {
	submenu := r.Group("/submenus")
	{
		submenu.POST("", h.CreateSubmenu)                           // Create a new submenu
		submenu.GET("", h.ListSubmenus)                             // List all submenus
		submenu.GET("/:id", h.GetSubmenu)                           // Get submenu by ID
		submenu.GET("/by-menu/:menu_id", h.ListSubmenusByMenu)      // List submenus by menu ID
		submenu.GET("/active/:menu_id", h.ListActiveSubmenusByMenu) // List active submenus by menu
		submenu.GET("/parent/:parent_id", h.ListSubmenusByParent)   // List submenus by parent submenu
		submenu.GET("/by-code", h.GetSubmenuByCode)                 // Get submenu by code (menu_id + code query)
		submenu.PATCH("/:id/toggle", h.ToggleSubmenuActive)         // Toggle submenu active status
		submenu.PUT("/:id", h.UpdateSubmenu)                        // Update submenu by ID
		submenu.DELETE("/:id", h.DeleteSubmenu)                     // Delete submenu by ID
	}
}
