package router

import (
	"github.com/NasTecSol/nembus-core/handler"
	"github.com/gin-gonic/gin"
)

// RegisterReleaseRoutes registers public and administrative routes for app updates and releases
func RegisterReleaseRoutes(r *gin.Engine, h *handler.ReleaseHandler) {
	// Release management (protected by X-Release-Secret in handler)
	apiV1 := r.Group("/api/v1")
	{
		releases := apiV1.Group("/releases")
		{
			releases.POST("/upload", h.UploadRelease)
		}

		// Client update check and download (protected by X-Nembus-App-Key in handler)
		updates := apiV1.Group("/updates")
		{
			updates.GET("/:app/latest", h.CheckLatestUpdate)
			updates.GET("/:app/download/:id", h.DownloadRelease)
		}
	}
}
