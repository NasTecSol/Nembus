package router

import (
	"github.com/NasTecSol/nembus-core/handler"
	"github.com/gin-gonic/gin"
)

func RegisterReceiptTemplateRoutes(r *gin.RouterGroup, h *handler.ReceiptTemplateHandler) {
	r.GET("/organizations/:id/receipt-templates", h.List)
	r.POST("/organizations/:id/receipt-templates", h.Create)
	r.PUT("/organizations/:id/receipt-templates/:template_id", h.Update)
	r.PUT("/organizations/:id/receipt-template-default", h.SetOrganizationDefault)
	r.PUT("/stores/:id/receipt-template-default", h.SetStoreDefault)
	r.GET("/stores/:id/receipt-template-default", h.GetStoreDefault)
}
