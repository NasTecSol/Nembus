package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterTenantRoutes(r *gin.RouterGroup, h *handler.TenantHandler) {
	tenants := r.Group("/tenants")
	{
		// CRUD
		tenants.POST("", h.CreateTenant)
		tenants.GET("", h.ListActiveTenants)
		tenants.GET("/all", h.ListAllTenants)
		tenants.GET("/:slug", h.GetTenantBySlug)
		tenants.PUT("/:id", h.UpdateTenant)

		// Special cases
		tenants.GET("/:slug/any", h.GetTenantBySlugAny)
		tenants.PUT("/:id/deactivate", h.DeactivateTenant)
	}
}
