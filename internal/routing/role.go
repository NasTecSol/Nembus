package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterRoleRoutes(r *gin.RouterGroup, h *handler.RoleHandler) {
	roles := r.Group("/roles")
	{
		// CRUD
		roles.POST("", h.CreateRole)
		roles.GET("", h.ListRoles)
		roles.GET("/:id", h.GetRole)
		roles.GET("/code/:code", h.GetRoleByCode)
		roles.PUT("/:id", h.UpdateRole)
		roles.DELETE("/:id", h.DeleteRole)

		// Filters / status
		roles.GET("/active", h.ListActiveRoles)
		roles.GET("/non-system", h.ListNonSystemRoles)
		roles.PATCH("/:id/active", h.ToggleRoleActive)

		// Role permissions
		roles.POST("/:id/permissions", h.AssignPermissionToRole)
		roles.GET("/:id/permissions", h.GetRolePermissions)
		roles.DELETE("/:id/permissions/:permission_id", h.RemovePermissionFromRole)
		roles.GET("/:id/permissions/:permission_id/check", h.CheckRoleHasPermission)
	}
}
