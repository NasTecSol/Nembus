package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(r *gin.RouterGroup, h *handler.UserHandler) {
	user := r.Group("/users")
	{
		user.POST("", h.CreateUser)
		user.GET("/:id", h.GetUser)
		user.GET("", h.ListUsers)

		// 🔑 User Roles
		user.POST("addUserRoles/:id", h.AssignRoleToUser)
	}
}
