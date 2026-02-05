package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(r *gin.RouterGroup, h *handler.UserHandler) {
	user := r.Group("/users")
	{
		// User CRUD
		user.POST("", h.CreateUser)
		user.GET("/:id", h.GetUser)
		user.GET("", h.ListUsers)
		user.PATCH("/:id", h.UpdateUser)                // Update user details
		user.PUT("/:id/password", h.UpdateUserPassword) // Update user password

		// User Roles
		user.POST("addUserRoles/:id", h.AssignRoleToUser)
		user.GET("/role/:role_id", h.GetUsersByRole)              // Get users by role
		user.DELETE("revokeRole/:user_id/:role_id", h.RevokeRole) // Revoke specific role
		user.DELETE("revokeAllRoles/:user_id", h.RevokeAllRoles)  // Revoke all roles

		// User Stores
		user.POST("grantStore/:id", h.GrantStoreAccess)                    // Grant store access
		user.PUT("/:id/stores/primary", h.SetUserPrimaryStore)             // Set primary store
		user.DELETE("revokeStore/:user_id/:store_id", h.RevokeStoreAccess) // Revoke store access
		user.DELETE("revokeAllStores/:user_id", h.RevokeAllStoreAccess)    // Revoke all stores

		// User Details with roles/stores
		user.GET("/details/:id", h.GetUserWithDetails)
		user.GET("/details", h.ListUsersWithDetails)

		// User search
		user.GET("/search", h.SearchUsers)

		// Store specific users
		user.GET("/store/:store_id", h.GetStoreUsers)
		user.GET("/:id/stores", h.GetUserStores)
		user.GET("/:id/primaryStore", h.GetUserPrimaryStore)
	}
}
