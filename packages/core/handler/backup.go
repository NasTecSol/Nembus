package handler

import (
	"net/http"

	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/gin-gonic/gin"
)

// BackupHandler exposes HTTP endpoints for the backup service.
// Actual byte streaming happens over gRPC directly from the gRPC server;
// these HTTP endpoints handle management operations.
type BackupHandler struct {
	useCase *usecase.BackupUseCase
}

// NewBackupHandler constructs a BackupHandler.
func NewBackupHandler(uc *usecase.BackupUseCase) *BackupHandler {
	return &BackupHandler{useCase: uc}
}

// GetBackupStatus handles GET /api/backup/status/:tenant
// @Summary      Get backup status for a tenant
// @Description  Returns whether a gRPC backup stream is currently active for the tenant, and how many bytes have been sent so far
// @Tags         backup
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        tenant        path      string  true  "Tenant slug"
// @Success      200  {object}  BackupStatusResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/backup/status/{tenant} [get]
func (h *BackupHandler) GetBackupStatus(c *gin.Context) {
	tenant := c.Param("tenant")
	if tenant == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "tenant slug is required", nil))
		return
	}
	resp := h.useCase.GetBackupStatus(c.Request.Context(), tenant)
	c.JSON(resp.StatusCode, resp)
}

// GetGRPCInfo handles GET /api/backup/grpc-info/:tenant
// @Summary      Get gRPC connection info for backup streaming
// @Description  Returns the gRPC server address and tenant slug so a POS terminal or offline app can initiate a direct streaming backup
// @Tags         backup
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        tenant        path      string  true  "Tenant slug"
// @Success      200  {object}  BackupGRPCInfoResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/backup/grpc-info/{tenant} [get]
func (h *BackupHandler) GetGRPCInfo(c *gin.Context) {
	tenant := c.Param("tenant")
	if tenant == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "tenant slug is required", nil))
		return
	}
	resp := h.useCase.GetGRPCInfo(c.Request.Context(), tenant)
	c.JSON(resp.StatusCode, resp)
}

// TriggerServerBackup handles POST /api/backup/trigger
// @Summary      Trigger a server-side database backup
// @Description  Starts an asynchronous backup job on the server. The backup is saved to disk on the server and is NOT streamed to the caller. Use GET /api/backup/jobs/:job_id to track progress.
// @Tags         backup
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                    true  "Tenant identifier"
// @Param        Authorization header    string                    true  "Bearer token"
// @Param        body          body      TriggerBackupRequest      true  "Backup request"
// @Success      201  {object}  BackupJobResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/backup/trigger [post]
func (h *BackupHandler) TriggerServerBackup(c *gin.Context) {
	var req struct {
		TenantSlug string `json:"tenant_slug" binding:"required"`
		AuthToken  string `json:"auth_token"  binding:"required"`
		Compressed bool   `json:"compressed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body: "+err.Error(), nil))
		return
	}

	resp := h.useCase.TriggerServerBackup(c.Request.Context(), req.TenantSlug, req.AuthToken, req.Compressed)
	c.JSON(resp.StatusCode, resp)
}

// GetJobStatus handles GET /api/backup/jobs/:job_id
// @Summary      Get status of a server-side backup job
// @Description  Returns the current status (running/done/failed), file path, and start time of a triggered backup job
// @Tags         backup
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        job_id        path      string  true  "Job ID returned by /api/backup/trigger"
// @Success      200  {object}  BackupJobResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/backup/jobs/{job_id} [get]
func (h *BackupHandler) GetJobStatus(c *gin.Context) {
	jobID := c.Param("job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "job_id is required", nil))
		return
	}
	resp := h.useCase.GetJobStatus(c.Request.Context(), jobID)
	c.JSON(resp.StatusCode, resp)
}

// ListJobs handles GET /api/backup/jobs
// @Summary      List all backup jobs
// @Description  Returns all server-side backup jobs (running, done, failed) for this server instance
// @Tags         backup
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200  {array}   BackupJobResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /api/backup/jobs [get]
func (h *BackupHandler) ListJobs(c *gin.Context) {
	resp := h.useCase.ListJobs(c.Request.Context())
	c.JSON(resp.StatusCode, resp)
}

// ─── Swagger doc types ────────────────────────────────────────────────────────

// BackupStatusResponse is the response shape for GET /api/backup/status/:tenant
// swagger:model BackupStatusResponse
type BackupStatusResponse struct {
	TenantSlug string `json:"tenant_slug"`
	InProgress bool   `json:"in_progress"`
	BytesSent  uint64 `json:"bytes_sent,omitempty"`
	StartedAt  string `json:"started_at,omitempty"`
}

// BackupGRPCInfoResponse is the response shape for GET /api/backup/grpc-info/:tenant
// swagger:model BackupGRPCInfoResponse
type BackupGRPCInfoResponse struct {
	GRPCAddress string `json:"grpc_address"`
	TenantSlug  string `json:"tenant_slug"`
	Note        string `json:"note"`
}

// TriggerBackupRequest is the request body for POST /api/backup/trigger
// swagger:model TriggerBackupRequest
type TriggerBackupRequest struct {
	TenantSlug string `json:"tenant_slug" example:"store-riyadh-01"`
	AuthToken  string `json:"auth_token"  example:"eyJhbGciOiJIUzI1NiIs..."`
	Compressed bool   `json:"compressed"  example:"false"`
}

// BackupJobResponse is the response shape for job-related endpoints
// swagger:model BackupJobResponse
type BackupJobResponse struct {
	TenantSlug string `json:"tenant_slug"`
	JobID      string `json:"job_id"`
	FilePath   string `json:"file_path"`
	StartedAt  string `json:"started_at"`
	Status     string `json:"status"`
}
