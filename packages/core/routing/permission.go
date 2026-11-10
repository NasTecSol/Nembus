package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

func RegisterPermissionRoutes(r *gin.RouterGroup, h *handler.PermissionHandler) {
	permissions := r.Group("/permissions")
	{
		// CRUD - POST and GET list first
		permissions.POST("", h.CreatePermission)
		permissions.GET("", h.ListPermissions)

		// Entity-specific routes (must be before /:id to avoid "role", "module" etc. matching as id)
		permissions.GET("/role/:role_id", h.GetRolePermissionsWithScope)
		permissions.DELETE("/role/:role_id/permission/:permission_id", h.RevokePermissionFromRole)
		permissions.PUT("/role/:role_id/permission/:permission_id/scope", h.UpdateRolePermissionScope)

		// Module permissions
		permissions.GET("/module/:module_id", h.GetModulePermissions)
		permissions.POST("/module/:module_id", h.AssignPermissionToModule)
		permissions.DELETE("/module/:module_id/permission/:permission_id", h.RevokePermissionFromModule)

		// Menu permissions
		permissions.GET("/menu/:menu_id", h.GetMenuPermissions)
		permissions.POST("/menu/:menu_id", h.AssignPermissionToMenu)
		permissions.DELETE("/menu/:menu_id/permission/:permission_id", h.RevokePermissionFromMenu)

		// Submenu permissions
		permissions.GET("/submenu/:submenu_id", h.GetSubmenuPermissions)
		permissions.POST("/submenu/:submenu_id", h.AssignPermissionToSubmenu)
		permissions.DELETE("/submenu/:submenu_id/permission/:permission_id", h.RevokePermissionFromSubmenu)

		// User permissions (more specific routes first)
		permissions.GET("/user/:user_id/submenu/:submenu_code", h.CheckUserSubmenuPermission)
		permissions.GET("/user/:user_id/with-scope", h.GetUserPermissionsWithScope)
		permissions.GET("/user/:user_id/check/:code", h.CheckUserHasPermission)
		permissions.GET("/user/:user_id/modules", h.GetUserAccessibleModules)
		permissions.GET("/user/:user_id/menus", h.GetUserAccessibleMenus)
		permissions.GET("/user/:user_id/submenus", h.GetUserAccessibleSubmenus)
		permissions.GET("/user/:user_id", h.GetUserPermissions)

		// CRUD by ID (last, so /role/:id etc. don't match)
		permissions.GET("/code/:code", h.GetPermissionByCode)
		permissions.GET("/:id", h.GetPermission)
		permissions.PUT("/:id", h.UpdatePermission)
		permissions.DELETE("/:id", h.DeletePermission)
	}
}
