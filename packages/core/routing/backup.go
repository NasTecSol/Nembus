package router

import (
	"github.com/NasTecSol/nembus-core/handler"

	"github.com/gin-gonic/gin"
)

// RegisterBackupRoutes wires up the HTTP backup management endpoints.
// Note: actual byte-level backup streaming happens over gRPC at :GRPC_PORT.
// These HTTP endpoints handle status queries, grpc info retrieval, and
// triggering server-side backup jobs.
func RegisterBackupRoutes(r *gin.RouterGroup, h *handler.BackupHandler) {
	backup := r.Group("/backup")
	{
		// Query the gRPC server for live streaming status of a tenant
		backup.GET("/status/:tenant", h.GetBackupStatus)

		// Return gRPC address + tenant info so POS/offline apps can stream directly
		backup.GET("/grpc-info/:tenant", h.GetGRPCInfo)

		// Server-side backup management (backup saved to server disk)
		backup.POST("/trigger", h.TriggerServerBackup)
		backup.GET("/jobs", h.ListJobs)
		backup.GET("/jobs/:job_id", h.GetJobStatus)
	}
}
