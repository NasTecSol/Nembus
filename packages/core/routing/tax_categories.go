package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterTaxCategoryRoutes registers tax category routes under /api/tax-categories.
func RegisterTaxCategoryRoutes(r *gin.RouterGroup, h *handler.TaxCategoriesHandler) {
	tc := r.Group("/tax-categories")
	{
		tc.POST("", h.CreateTaxCategory)
		tc.GET("", h.ListTaxCategories)
		tc.GET("/active", h.ListActiveTaxCategories)
		tc.GET("/code/:code", h.GetTaxCategoryByCode)

		tc.GET("/:id", h.GetTaxCategory)
		tc.PUT("/:id", h.UpdateTaxCategory)
		tc.DELETE("/:id", h.DeleteTaxCategory)
		tc.PATCH("/:id/active", h.ToggleTaxCategoryActive)
	}
}

