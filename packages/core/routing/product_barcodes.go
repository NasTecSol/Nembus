package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

func RegisterProductBarcodeRoutes(r *gin.RouterGroup, h *handler.ProductBarcodeHandler) {
	// Product barcode CRUD routes
	productBarcodes := r.Group("/product-barcodes")
	{
		productBarcodes.POST("", h.CreateProductBarcode)
		productBarcodes.GET("", h.ListProductBarcodes)
		productBarcodes.GET("/lookup/:barcode", h.GetProductByBarcode)
		productBarcodes.GET("/:id", h.GetProductBarcode)
		productBarcodes.PUT("/:id", h.UpdateProductBarcode)
		productBarcodes.DELETE("/:id", h.DeleteProductBarcode)
	}

	// Product-specific barcode routes
	products := r.Group("/products")
	{
		products.GET("/:product_id/barcodes", h.ListProductBarcodesByProduct)
		products.GET("/:product_id/barcodes/primary", h.GetPrimaryBarcode)
		products.PUT("/:product_id/barcodes/primary", h.SetPrimaryBarcode)
	}

	// Product variant-specific barcode routes
	productVariants := r.Group("/product-variants")
	{
		productVariants.GET("/:variant_id/barcodes", h.ListProductBarcodesByVariant)
	}
}
