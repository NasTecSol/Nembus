package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

func RegisterBrandRoutes(r *gin.RouterGroup, h *handler.BrandHandler) {
	brands := r.Group("/brands")
	{
		// Create operations
		brands.POST("", h.CreateBrand)
		brands.POST("/with-defaults", h.CreateBrandWithDefaults)

		// Read operations
		brands.GET("/all", h.ListAllBrands)
		brands.GET("/active", h.ListActiveBrands)
		brands.GET("", h.ListBrands)
		brands.GET("/active/paginated", h.ListActiveBrandsWithPagination)
		brands.GET("/count", h.CountBrands)
		brands.GET("/count/active", h.CountActiveBrands)
		brands.GET("/search", h.SearchBrands)
		brands.GET("/search/active", h.SearchActiveBrands)
		brands.GET("/:id", h.GetBrandByID)
		brands.GET("/code/:code", h.GetBrandByCode)
		brands.GET("/:id/exists", h.BrandExists)
		brands.GET("/code/:code/exists", h.BrandCodeExists)

		// Update operations
		brands.PUT("/:id", h.UpdateBrand)
		brands.PATCH("/:id/name", h.UpdateBrandName)
		brands.PATCH("/:id/code", h.UpdateBrandCode)
		brands.PATCH("/:id/metadata", h.UpdateBrandMetadata)
		brands.PATCH("/:id/activate", h.ActivateBrand)
		brands.PATCH("/:id/deactivate", h.DeactivateBrand)
		brands.PATCH("/:id/toggle-status", h.ToggleBrandStatus)

		// Delete operations
		brands.DELETE("/:id", h.DeleteBrand)
		brands.DELETE("/code/:code", h.DeleteBrandByCode)
		brands.DELETE("/:id/soft", h.SoftDeleteBrand)

		// Reporting & Analytics
		brands.GET("/:id/products/count", h.GetBrandWithProductCount)
		brands.GET("/products/counts", h.ListBrandsWithProductCounts)
		brands.GET("/active/products/counts", h.ListActiveBrandsWithProductCounts)
		brands.GET("/top", h.GetTopBrandsByProductCount)
		brands.GET("/no-products", h.GetBrandsWithNoProducts)
		brands.GET("/inactive/active-products", h.GetInactiveBrandsWithActiveProducts)

		// Bulk operations
		brands.POST("/bulk/activate", h.BulkActivateBrands)
		brands.POST("/bulk/deactivate", h.BulkDeactivateBrands)
		brands.POST("/bulk/delete", h.BulkDeleteBrands)

		// Audit & Metadata queries
		brands.GET("/recent/created", h.GetRecentlyCreatedBrands)
		brands.GET("/recent/updated", h.GetRecentlyUpdatedBrands)
		brands.GET("/by-date", h.GetBrandsByCreationDate)
		brands.GET("/:id/metadata/:key", h.GetBrandMetadataByKey)
		brands.GET("/stats", h.ListBrandsWithStats)
	}
}
