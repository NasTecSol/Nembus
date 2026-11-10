package router

import (
	"github.com/NasTecSol/nembus-core/handler"
	"github.com/gin-gonic/gin"
)

func RegisterRFQRoutes(r *gin.RouterGroup, h *handler.RFQHandler) {
	rfqs := r.Group("/rfqs")
	rfqs.GET("", h.List)
	rfqs.POST("", h.Create)
	rfqs.GET("/:id", h.Get)
	rfqs.POST("/:id/publish", h.Publish)
	rfqs.POST("/:id/close", h.Close)
	rfqs.GET("/:id/responses", h.Responses)
}

func RegisterPublicRFQRoutes(r *gin.RouterGroup, h *handler.RFQHandler) {
	rfqs := r.Group("/rfqs")
	rfqs.GET("/:token", h.PublicGet)
	rfqs.POST("/:token/responses", h.PublicSubmit)
}
