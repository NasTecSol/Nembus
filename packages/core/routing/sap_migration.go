package router

import (
	"github.com/gin-gonic/gin"

	"github.com/NasTecSol/nembus-core/handler"
)

func RegisterSAPMigrationRoutes(r *gin.RouterGroup, h *handler.SAPMigrationHandler) {
	migration := r.Group("/migration")
	{
		migration.POST("/batch", h.IngestBatch)
	}
}
