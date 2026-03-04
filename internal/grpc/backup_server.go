// Package grpc implements the gRPC BackupService server.
// It streams a live pg_dump of a tenant's database to any connecting client
// (POS terminal, offline application, etc.) as 32 KB chunks with SHA-256
// checksums per chunk so the receiver can verify data integrity.
package grpc

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"

	"NEMBUS/internal/grpc/backuppb"
	"NEMBUS/internal/middleware/manager"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	chunkSize     = 32 * 1024 // 32 KB per chunk
	maxConcurrent = 3         // maximum simultaneous backup streams
)

// progressEntry tracks a single in-flight backup.
type progressEntry struct {
	bytesSent uint64
	startedAt time.Time
}

// BackupServer implements backuppb.BackupServiceServer.
type BackupServer struct {
	backuppb.UnimplementedBackupServiceServer

	tenantManager *manager.Manager
	jwtSecret     string
	pgDumpPath    string // full path to pg_dump binary (or just "pg_dump" if in PATH)

	mu       sync.RWMutex
	progress map[string]*progressEntry // key: tenant_slug
}

// NewBackupServer constructs a BackupServer.
func NewBackupServer(tm *manager.Manager, jwtSecret, pgDumpPath string) *BackupServer {
	if pgDumpPath == "" {
		pgDumpPath = "pg_dump" // fallback to PATH lookup
	}
	return &BackupServer{
		tenantManager: tm,
		jwtSecret:     jwtSecret,
		pgDumpPath:    pgDumpPath,
		progress:      make(map[string]*progressEntry),
	}
}

// ─── StreamBackup ─────────────────────────────────────────────────────────────

// StreamBackup runs pg_dump against the named tenant's database and streams
// the output as BackupChunk messages until pg_dump finishes or the client
// disconnects.
func (s *BackupServer) StreamBackup(req *backuppb.BackupRequest, stream backuppb.BackupService_StreamBackupServer) error {
	ctx := stream.Context()

	// 1. Validate auth token (simple non-empty check; swap in full JWT validation for prod)
	if err := s.validateToken(req.AuthToken); err != nil {
		return status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// 2. Concurrency guard
	s.mu.Lock()
	if len(s.progress) >= maxConcurrent {
		s.mu.Unlock()
		return status.Errorf(codes.ResourceExhausted, "too many concurrent backups (max %d)", maxConcurrent)
	}
	s.progress[req.TenantSlug] = &progressEntry{startedAt: time.Now()}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.progress, req.TenantSlug)
		s.mu.Unlock()
	}()

	// 3. Resolve tenant DSN
	dsn, err := s.tenantManager.GetTenantDSN(ctx, req.TenantSlug)
	if err != nil {
		return status.Errorf(codes.NotFound, "tenant not found: %v", err)
	}

	// 4. Build pg_dump command
	var pgArgs []string
	if req.Compressed {
		pgArgs = []string{"--no-password", "--format=custom", dsn}
	} else {
		pgArgs = []string{"--no-password", "--format=plain", dsn}
	}

	cmd := exec.CommandContext(ctx, s.pgDumpPath, pgArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return status.Errorf(codes.Internal, "failed to open pg_dump stdout: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return status.Errorf(codes.Internal, "failed to open pg_dump stderr: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return status.Errorf(codes.Internal, "failed to start pg_dump: %v", err)
	}

	// Kill pg_dump if the client disconnects.
	go func() {
		<-ctx.Done()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()

	// 5. Stream chunks
	buf := make([]byte, chunkSize)
	var totalSent uint64

	for {
		n, readErr := stdout.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			h := sha256.Sum256(chunk)
			totalSent += uint64(n)

			// Update live progress
			s.mu.Lock()
			if e, ok := s.progress[req.TenantSlug]; ok {
				e.bytesSent = totalSent
			}
			s.mu.Unlock()

			msg := &backuppb.BackupChunk{
				Data:   append([]byte(nil), chunk...),
				Offset: totalSent,
				Sha256: fmt.Sprintf("%x", h),
				IsLast: false,
			}
			if sendErr := stream.Send(msg); sendErr != nil {
				log.Printf("[backup] send error for tenant %s: %v", req.TenantSlug, sendErr)
				return sendErr
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			_ = cmd.Process.Kill()
			return status.Errorf(codes.Internal, "read error from pg_dump: %v", readErr)
		}
	}

	// Drain stderr for logging
	if errBytes, _ := io.ReadAll(stderr); len(errBytes) > 0 {
		log.Printf("[backup] pg_dump stderr for tenant %s: %s", req.TenantSlug, errBytes)
	}

	if err := cmd.Wait(); err != nil {
		return status.Errorf(codes.Internal, "pg_dump exited with error: %v", err)
	}

	// Send the final sentinel chunk
	final := &backuppb.BackupChunk{
		Data:   nil,
		Offset: totalSent,
		Sha256: "",
		IsLast: true,
	}
	return stream.Send(final)
}

// ─── BackupStatus ─────────────────────────────────────────────────────────────

// BackupStatus returns the current progress of an in-flight backup.
func (s *BackupServer) BackupStatus(_ context.Context, req *backuppb.BackupStatusRequest) (*backuppb.BackupStatusResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.progress[req.TenantSlug]
	if !ok {
		return &backuppb.BackupStatusResponse{InProgress: false}, nil
	}
	return &backuppb.BackupStatusResponse{
		InProgress: true,
		BytesSent:  e.bytesSent,
		StartedAt:  e.startedAt.Format(time.RFC3339),
	}, nil
}

// ─── Auth helpers ─────────────────────────────────────────────────────────────

func (s *BackupServer) validateToken(token string) error {
	if token == "" {
		return fmt.Errorf("empty token")
	}
	// TODO: Replace with full JWT validation using s.jwtSecret, e.g.:
	// parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
	//     return []byte(s.jwtSecret), nil
	// })
	// if err != nil || !parsed.Valid { return fmt.Errorf("invalid JWT") }
	return nil
}
