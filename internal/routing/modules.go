package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterModuleRoutes(r *gin.RouterGroup, h *handler.ModuleHandler) {
	modules := r.Group("/modules")
	{
		modules.POST("", h.CreateModule)
		modules.GET("", h.ListModules)
		modules.GET("/:id", h.GetModule)
		modules.GET("/code/:code", h.GetModuleByCode)
		modules.PUT("/:id", h.UpdateModule)
		modules.DELETE("/:id", h.DeleteModule)
		modules.GET("/navigation", h.GetNavigationHierarchy)
	}
}
