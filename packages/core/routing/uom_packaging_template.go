package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterUomPackagingTemplateRoutes registers packaging template and level routes.
func RegisterUomPackagingTemplateRoutes(r *gin.RouterGroup, h *handler.UomPackagingTemplateHandler) {
	templates := r.Group("/uom-packaging-templates")
	{
		templates.POST("", h.CreateTemplate)
		templates.POST("/pipeline", h.CreateTemplatePipeline)
		templates.GET("", h.ListTemplates)
		templates.GET("/by-uom/:uom_id", h.GetTemplatesByUomID)
		templates.GET("/:id", h.GetTemplate)
		templates.GET("/:id/full", h.GetTemplateWithLevels)
		templates.PUT("/:id", h.UpdateTemplate)
		templates.DELETE("/:id", h.DeleteTemplate)

		// Levels for a template
		templates.POST("/:id/levels", h.CreateLevel)
		templates.GET("/:id/levels", h.ListLevels)
	}

	levels := r.Group("/uom-packaging-template-levels")
	{
		levels.PUT("/:level_id", h.UpdateLevel)
		levels.DELETE("/:level_id", h.DeleteLevel)
	}
}
