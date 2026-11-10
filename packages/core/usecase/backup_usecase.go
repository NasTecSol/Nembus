package usecase

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/backuppb"
	"github.com/NasTecSol/nembus-core/middleware/manager"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ─── DTOs ─────────────────────────────────────────────────────────────────────

// BackupStatusOutput is returned by GetBackupStatus.
type BackupStatusOutput struct {
	TenantSlug string `json:"tenant_slug"`
	InProgress bool   `json:"in_progress"`
	BytesSent  uint64 `json:"bytes_sent,omitempty"`
	StartedAt  string `json:"started_at,omitempty"`
}

// BackupJobOutput is returned after a server-side backup job is triggered.
type BackupJobOutput struct {
	TenantSlug string `json:"tenant_slug"`
	JobID      string `json:"job_id"`
	FilePath   string `json:"file_path"`
	StartedAt  string `json:"started_at"`
	Status     string `json:"status"` // "running" | "done" | "failed"
}

// BackupGRPCInfoOutput tells clients how to reach the gRPC stream directly.
type BackupGRPCInfoOutput struct {
	GRPCAddress string `json:"grpc_address"`
	TenantSlug  string `json:"tenant_slug"`
	Note        string `json:"note"`
}

// ─── serverJob tracks a background server-side backup ─────────────────────────

type serverJob struct {
	id         string
	tenantSlug string
	filePath   string
	startedAt  time.Time
	status     string // "running" | "done" | "failed"
	errMsg     string
}

// ─── UseCase ──────────────────────────────────────────────────────────────────

// BackupUseCase manages backup operations for tenants.
// It does NOT use an SQLC repository because backup data comes from the
// gRPC server state (in-memory progress map) and from pg_dump (subprocess).
type BackupUseCase struct {
	tenantManager *manager.Manager
	grpcAddr      string // e.g. "localhost:50051"

	mu   sync.RWMutex
	jobs map[string]*serverJob // key: jobID
}

// NewBackupUseCase creates a new BackupUseCase.
// grpcAddr is the address of the local gRPC backup server (e.g. ":50051").
func NewBackupUseCase(tm *manager.Manager, grpcAddr string) *BackupUseCase {
	return &BackupUseCase{
		tenantManager: tm,
		grpcAddr:      grpcAddr,
		jobs:          make(map[string]*serverJob),
	}
}

// SetRepository is a no-op that satisfies the optional convention used by handlers.
// The backup usecase does not read from SQLC queries.
func (uc *BackupUseCase) SetRepository(_ *repository.Queries) {}

// ─── GetBackupStatus ──────────────────────────────────────────────────────────

// GetBackupStatus queries the gRPC service for the current backup progress of a tenant.
func (uc *BackupUseCase) GetBackupStatus(ctx context.Context, tenantSlug string) *repository.Response {
	if tenantSlug == "" {
		return utils.NewResponse(utils.CodeBadReq, "tenant_slug is required", nil)
	}

	conn, err := grpc.NewClient(uc.grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("cannot connect to gRPC server: %v", err), nil)
	}
	defer conn.Close()

	client := backuppb.NewBackupServiceClient(conn)
	resp, err := client.BackupStatus(ctx, &backuppb.BackupStatusRequest{TenantSlug: tenantSlug})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("BackupStatus RPC failed: %v", err), nil)
	}

	out := BackupStatusOutput{
		TenantSlug: tenantSlug,
		InProgress: resp.InProgress,
		BytesSent:  resp.BytesSent,
		StartedAt:  resp.StartedAt,
	}
	return utils.NewResponse(utils.CodeOK, "backup status fetched", out)
}

// ─── GetGRPCInfo ─────────────────────────────────────────────────────────────

// GetGRPCInfo returns connection details for a POS terminal / offline app to
// initiate a gRPC streaming backup directly.
func (uc *BackupUseCase) GetGRPCInfo(ctx context.Context, tenantSlug string) *repository.Response {
	if tenantSlug == "" {
		return utils.NewResponse(utils.CodeBadReq, "tenant_slug is required", nil)
	}

	// Validate the tenant exists before handing out the address.
	if _, err := uc.tenantManager.GetTenantDSN(ctx, tenantSlug); err != nil {
		return utils.NewResponse(utils.CodeNotFound, fmt.Sprintf("tenant not found: %v", err), nil)
	}

	out := BackupGRPCInfoOutput{
		GRPCAddress: uc.grpcAddr,
		TenantSlug:  tenantSlug,
		Note:        "Connect to GRPCAddress and call BackupService.StreamBackup with your JWT token to receive the backup stream.",
	}
	return utils.NewResponse(utils.CodeOK, "gRPC connection info", out)
}

// ─── TriggerServerBackup ──────────────────────────────────────────────────────

// TriggerServerBackup starts a background backup that is saved to disk ON THE
// SERVER (not streamed to the caller).  This is useful when the admin wants to
// create a server-side snapshot that can later be moved to cloud storage.
//
// The backup is performed by connecting to the local gRPC server (so concurrency
// guards and tenant validation are respected) and writing chunks to a local file.
func (uc *BackupUseCase) TriggerServerBackup(ctx context.Context, tenantSlug, authToken string, compressed bool) *repository.Response {
	if tenantSlug == "" {
		return utils.NewResponse(utils.CodeBadReq, "tenant_slug is required", nil)
	}
	if authToken == "" {
		return utils.NewResponse(utils.CodeBadReq, "auth_token is required", nil)
	}

	jobID := fmt.Sprintf("%s_%d", tenantSlug, time.Now().UnixNano())
	ext := ".sql"
	if compressed {
		ext = ".pgdump"
	}

	backupDir := "./backups"
	if err := os.MkdirAll(backupDir, 0750); err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("cannot create backups directory: %v", err), nil)
	}

	filePath := filepath.Join(backupDir, fmt.Sprintf("backup_%s_%d%s", tenantSlug, time.Now().Unix(), ext))

	job := &serverJob{
		id:         jobID,
		tenantSlug: tenantSlug,
		filePath:   filePath,
		startedAt:  time.Now(),
		status:     "running",
	}

	uc.mu.Lock()
	uc.jobs[jobID] = job
	uc.mu.Unlock()

	// Run the backup asynchronously so the HTTP response returns immediately.
	go func() {
		err := uc.runBackupToFile(tenantSlug, authToken, compressed, filePath)
		uc.mu.Lock()
		defer uc.mu.Unlock()
		if err != nil {
			job.status = "failed"
			job.errMsg = err.Error()
			log.Printf("[BackupUseCase] server backup failed for tenant %s: %v", tenantSlug, err)
		} else {
			job.status = "done"
			log.Printf("[BackupUseCase] server backup done for tenant %s → %s", tenantSlug, filePath)
		}
	}()

	out := BackupJobOutput{
		TenantSlug: tenantSlug,
		JobID:      jobID,
		FilePath:   filePath,
		StartedAt:  job.startedAt.Format(time.RFC3339),
		Status:     "running",
	}
	return utils.NewResponse(utils.CodeCreated, "backup job started", out)
}

// runBackupToFile dials the gRPC server and writes the stream to a file.
func (uc *BackupUseCase) runBackupToFile(tenantSlug, authToken string, compressed bool, filePath string) error {
	// Use a fresh background context so the HTTP request cancellation doesn't abort the file write.
	bgCtx := context.Background()

	conn, err := grpc.NewClient(uc.grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("grpc dial: %w", err)
	}
	defer conn.Close()

	client := backuppb.NewBackupServiceClient(conn)
	stream, err := client.StreamBackup(bgCtx, &backuppb.BackupRequest{
		TenantSlug: tenantSlug,
		AuthToken:  authToken,
		Compressed: compressed,
	})
	if err != nil {
		return fmt.Errorf("StreamBackup: %w", err)
	}

	f, err := os.Create(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("stream recv: %w", err)
		}
		if len(chunk.Data) > 0 {
			if _, werr := f.Write(chunk.Data); werr != nil {
				return fmt.Errorf("write: %w", werr)
			}
		}
		if chunk.IsLast {
			break
		}
	}
	return nil
}

// ─── GetJob ───────────────────────────────────────────────────────────────────

// GetJobStatus returns the status of a previously triggered server-side backup job.
func (uc *BackupUseCase) GetJobStatus(_ context.Context, jobID string) *repository.Response {
	if jobID == "" {
		return utils.NewResponse(utils.CodeBadReq, "job_id is required", nil)
	}

	uc.mu.RLock()
	job, ok := uc.jobs[jobID]
	uc.mu.RUnlock()

	if !ok {
		return utils.NewResponse(utils.CodeNotFound, "job not found", nil)
	}

	out := BackupJobOutput{
		TenantSlug: job.tenantSlug,
		JobID:      job.id,
		FilePath:   job.filePath,
		StartedAt:  job.startedAt.Format(time.RFC3339),
		Status:     job.status,
	}
	return utils.NewResponse(utils.CodeOK, "job status fetched", out)
}

// ─── ListJobs ─────────────────────────────────────────────────────────────────

// ListJobs returns all server-side backup jobs (running, done, failed).
func (uc *BackupUseCase) ListJobs(_ context.Context) *repository.Response {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	out := make([]BackupJobOutput, 0, len(uc.jobs))
	for _, job := range uc.jobs {
		out = append(out, BackupJobOutput{
			TenantSlug: job.tenantSlug,
			JobID:      job.id,
			FilePath:   job.filePath,
			StartedAt:  job.startedAt.Format(time.RFC3339),
			Status:     job.status,
		})
	}
	return utils.NewResponse(utils.CodeOK, "backup jobs listed", out)
}
