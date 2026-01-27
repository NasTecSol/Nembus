package router

import (
	"NEMBUS/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterImageRoutes(r *gin.RouterGroup, h *handler.ImageHandler) {
	image := r.Group("/images")
	{
		image.POST("uploadImage/:module", h.UploadModuleImage)
	}
}
