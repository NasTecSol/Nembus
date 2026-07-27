package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterGoodsReceiptNoteRoutes registers goods receipt note routes under /api/goods-receipt-notes.
func RegisterGoodsReceiptNoteRoutes(r *gin.RouterGroup, h *handler.GoodsReceiptNotesHandler) {
	grn := r.Group("/goods-receipt-notes")
	{
		grn.POST("", h.CreateGoodsReceiptNote)
		grn.GET("/:id", h.GetGoodsReceiptNote)
		grn.POST("/:id/post", h.PostGoodsReceiptNote)
	}
}
