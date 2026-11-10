package grpc

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/syncpb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const changeLogSchemaDDL = `
CREATE TABLE IF NOT EXISTS change_log (
	seq         BIGSERIAL PRIMARY KEY,
	entity_type VARCHAR(50) NOT NULL,
	entity_id   BIGINT NOT NULL,
	op          VARCHAR(10) NOT NULL CHECK (op IN ('INSERT', 'UPDATE', 'DELETE')),
	payload     JSONB,
	changed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_change_log_entity_seq ON change_log (entity_type, entity_id, seq DESC);

CREATE TABLE IF NOT EXISTS change_log_meta (
	id                  SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
	compacted_up_to_seq BIGINT NOT NULL DEFAULT 0,
	min_retained_seq    BIGINT NOT NULL DEFAULT 0,
	updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
INSERT INTO change_log_meta (id, compacted_up_to_seq, min_retained_seq)
VALUES (1, 0, 0)
ON CONFLICT (id) DO NOTHING;
`

func ensureChangeLogSchema(ctx context.Context, pool *pgxpool.Pool) {
	if pool == nil {
		return
	}
	if _, err := pool.Exec(ctx, changeLogSchemaDDL); err != nil {
		log.Printf("[gRPC SyncServer] change_log DDL check warning: %v", err)
	}
}

func buildPullEntityRankCaseSQL() string {
	var sb strings.Builder
	sb.WriteString("CASE entity_type ")
	for idx, entity := range syncpb.PullEntityTypes() {
		sb.WriteString(fmt.Sprintf("WHEN '%s' THEN %d ", entity, idx))
	}
	sb.WriteString(fmt.Sprintf("ELSE %d END", len(syncpb.PullEntityTypes())))
	return sb.String()
}

// StreamChanges streams gap-safe monotonic change_log events from Cloud to POS.
// Writers acquire pg_advisory_xact_lock(hashtext('nembus.change_log_seq')) before allocating
// seq values; StreamChanges acquires pg_advisory_xact_lock_shared inside a READ ONLY transaction
// so no in-flight writer can commit a lower seq after the reader's snapshot opens.
func (s *SyncServer) StreamChanges(req *syncpb.StreamChangesRequest, stream syncpb.SyncService_StreamChangesServer) error {
	if req.TenantSlug == "" {
		return status.Error(codes.InvalidArgument, "tenant_slug is required")
	}
	if req.FromSeq < 0 {
		return status.Error(codes.InvalidArgument, "from_seq must be non-negative")
	}

	ctx := stream.Context()
	if err := s.verifyStreamAuth(ctx, req.TenantSlug, req.DeviceId, req.StoreId); err != nil {
		return err
	}

	pool, err := s.getTenantPool(ctx, req.TenantSlug)
	if err != nil {
		return status.Errorf(codes.NotFound, "tenant DB pool unavailable: %v", err)
	}

	ensureChangeLogSchema(ctx, pool)

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return status.Errorf(codes.Internal, "begin read-only change_log tx: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Acquire shared advisory lock matching the exclusive lock in trg_record_change_log()
	// to drain all in-flight change_log writers before taking the snapshot read.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock_shared(hashtext('nembus.change_log_seq'))`); err != nil {
		return status.Errorf(codes.Internal, "acquire change_log shared barrier lock: %v", err)
	}

	var compactedUpToSeq, minRetainedSeq int64
	err = tx.QueryRow(ctx, `
		SELECT compacted_up_to_seq, min_retained_seq
		FROM change_log_meta
		WHERE id = 1
	`).Scan(&compactedUpToSeq, &minRetainedSeq)
	if err != nil && err != pgx.ErrNoRows {
		return status.Errorf(codes.Internal, "read change_log_meta: %v", err)
	}

	if req.FromSeq > 0 && req.FromSeq < minRetainedSeq {
		return status.Errorf(codes.OutOfRange, "CURSOR_EXPIRED: from_seq %d is behind min_retained_seq %d", req.FromSeq, minRetainedSeq)
	}

	var maxSeq int64
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(seq), 0) FROM change_log`).Scan(&maxSeq); err != nil {
		return status.Errorf(codes.Internal, "read max change_log seq: %v", err)
	}
	if maxSeq <= req.FromSeq {
		return nil
	}

	limit := req.Limit
	if limit <= 0 {
		limit = syncpb.PullPageLimit
	}
	if limit > 1000 {
		limit = 1000
	}

	isSnapshotMode := req.FromSeq < compactedUpToSeq

	var rows pgx.Rows
	if isSnapshotMode {
		rankExpr := buildPullEntityRankCaseSQL()
		deleteFilter := ""
		if req.FromSeq == 0 {
			deleteFilter = "AND op <> 'DELETE'"
		}
		query := fmt.Sprintf(`
			SELECT seq, entity_type, entity_id, op, COALESCE(payload::text, ''), changed_at
			FROM change_log
			WHERE seq > $1 AND seq <= $2 %s
			ORDER BY %s ASC, seq ASC
		`, deleteFilter, rankExpr)
		rows, err = tx.Query(ctx, query, req.FromSeq, maxSeq)
	} else {
		rows, err = tx.Query(ctx, `
			SELECT seq, entity_type, entity_id, op, COALESCE(payload::text, ''), changed_at
			FROM change_log
			WHERE seq > $1
			ORDER BY seq ASC
			LIMIT $2
		`, req.FromSeq, limit)
	}
	if err != nil {
		return status.Errorf(codes.Internal, "query change_log: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			seq        int64
			entityType string
			entityID   int64
			op         string
			payloadStr string
			changedAt  time.Time
		)
		if err := rows.Scan(&seq, &entityType, &entityID, &op, &payloadStr, &changedAt); err != nil {
			return status.Errorf(codes.Internal, "scan change_log row: %v", err)
		}
		if !syncpb.IsPullEntity(entityType) {
			continue
		}
		var payloadBytes []byte
		var shaHex string
		if payloadStr != "" && payloadStr != "null" {
			payloadBytes = []byte(payloadStr)
			shaHex = computePayloadSHA256(payloadBytes)
		}

		ev := &syncpb.ChangeEvent{
			Seq:            seq,
			EntityType:     entityType,
			EntityId:       entityID,
			Op:             op,
			PayloadJson:    payloadBytes,
			Sha256:         shaHex,
			ChangedAt:      timestamppb.New(changedAt),
			IsSnapshotMode: isSnapshotMode,
		}
		if isSnapshotMode {
			ev.SnapshotMaxSeq = maxSeq
		}
		if err := stream.Send(ev); err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return status.Errorf(codes.Internal, "iterate change_log rows: %v", err)
	}
	return nil
}

// CompactChangeLog removes superseded (entity_type, entity_id) history rows and purges
// expired DELETE tombstones older than tombstoneRetention (default 30 days), updating
// change_log_meta.compacted_up_to_seq and change_log_meta.min_retained_seq.
func CompactChangeLog(ctx context.Context, pool *pgxpool.Pool, tombstoneRetention time.Duration) (int64, error) {
	if pool == nil {
		return 0, fmt.Errorf("database pool is nil")
	}
	if tombstoneRetention <= 0 {
		tombstoneRetention = 30 * 24 * time.Hour
	}
	ensureChangeLogSchema(ctx, pool)

	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('nembus.change_log_seq'))`); err != nil {
		return 0, err
	}

	var maxSeq int64
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(seq), 0) FROM change_log`).Scan(&maxSeq); err != nil {
		return 0, err
	}
	if maxSeq == 0 {
		return 0, nil
	}

	// 1. Delete superseded intermediate states for the same (entity_type, entity_id)
	tag1, err := tx.Exec(ctx, `
		DELETE FROM change_log c
		WHERE c.seq <= $1
		  AND EXISTS (
			SELECT 1 FROM change_log newer
			WHERE newer.entity_type = c.entity_type
			  AND newer.entity_id = c.entity_id
			  AND newer.seq > c.seq
			  AND newer.seq <= $1
		  )
	`, maxSeq)
	if err != nil {
		return 0, fmt.Errorf("compact superseded change_log rows: %w", err)
	}
	deleted := tag1.RowsAffected()

	// 2. Purge DELETE tombstones older than tombstoneRetention and advance min_retained_seq
	cutoff := time.Now().UTC().Add(-tombstoneRetention)
	var maxExpiredTombstoneSeq int64
	_ = tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(seq), 0)
		FROM change_log
		WHERE op = 'DELETE' AND changed_at < $1
	`, cutoff).Scan(&maxExpiredTombstoneSeq)

	if maxExpiredTombstoneSeq > 0 {
		tag2, err := tx.Exec(ctx, `
			DELETE FROM change_log
			WHERE op = 'DELETE' AND seq <= $1
		`, maxExpiredTombstoneSeq)
		if err != nil {
			return 0, fmt.Errorf("purge expired DELETE tombstones: %w", err)
		}
		deleted += tag2.RowsAffected()
	}

	_, err = tx.Exec(ctx, `
		UPDATE change_log_meta
		SET compacted_up_to_seq = GREATEST(compacted_up_to_seq, $1),
		    min_retained_seq    = GREATEST(min_retained_seq, $2),
		    updated_at          = NOW()
		WHERE id = 1
	`, maxSeq, maxExpiredTombstoneSeq)
	if err != nil {
		return 0, fmt.Errorf("update change_log_meta: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return deleted, nil
}

// CleanupSyncInbox removes sync_inbox deduplication records older than retention (default 30 days).
func CleanupSyncInbox(ctx context.Context, pool *pgxpool.Pool, retention time.Duration) (int64, error) {
	if pool == nil {
		return 0, fmt.Errorf("database pool is nil")
	}
	if retention <= 0 {
		retention = 30 * 24 * time.Hour
	}
	cutoff := time.Now().UTC().Add(-retention)
	tag, err := pool.Exec(ctx, `DELETE FROM sync_inbox WHERE applied_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
