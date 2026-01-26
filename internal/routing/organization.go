package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterOrganizationRoutes(r *gin.RouterGroup, h *handler.OrganizationHandler) {
	org := r.Group("/organizations")
	{
		org.POST("", h.CreateOrganization)
		org.GET("", h.ListOrganizations)
		org.GET("/code/:code", h.GetOrganizationByCode)
		org.GET("/:id", h.GetOrganization)
		org.PUT("/:id", h.UpdateOrganization)
		org.DELETE("/:id", h.DeleteOrganization)
	}
}
