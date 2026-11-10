package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

func RegisterOrganizationRoutes(r *gin.RouterGroup, h *handler.OrganizationHandler) {
	org := r.Group("/organizations")
	{
		org.POST("", h.CreateOrganization)
		org.GET("", h.ListOrganizations)
		org.GET("/code/:code", h.GetOrganizationByCode)
		org.GET("/:id", h.GetOrganization)
		org.GET("/:id/zatca", h.GetOrganizationZatca)
		org.PUT("/:id/zatca", h.UpdateOrganizationZatca)
		org.PUT("/:id", h.UpdateOrganization)
		org.DELETE("/:id", h.DeleteOrganization)
	}
}
