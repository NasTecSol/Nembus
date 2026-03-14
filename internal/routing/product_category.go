package router

import (
	"NEMBUS/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterProductCategoryRoutes(r *gin.RouterGroup, h *handler.ProductCategoryHandler) {
	group := r.Group("/product-categories")
	{
		group.POST("", h.CreateProductCategory)
		group.GET("", h.ListProductCategories)
		group.GET("/hierarchy", h.GetCategoryHierarchy)
		group.GET("/:id", h.GetProductCategory)
		group.GET("/code/:code", h.GetProductCategoryByCode)
		group.GET("/:id/children", h.ListCategoryChildren)
		group.PUT("/:id", h.UpdateProductCategory)
		group.DELETE("/:id", h.DeleteProductCategory)
	}
}
