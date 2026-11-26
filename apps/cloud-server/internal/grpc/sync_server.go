package grpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/syncpb"
	"github.com/NasTecSol/nembus-core/middleware/manager"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/telemetry"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const syncInboxDDL = `
CREATE TABLE IF NOT EXISTS sync_inbox (
	outbox_uuid UUID PRIMARY KEY,
	device_id   VARCHAR(100) NOT NULL,
	store_id    INTEGER NOT NULL,
	batch_id    VARCHAR(100),
	entity_type VARCHAR(50) NOT NULL,
	entity_id   TEXT NOT NULL,
	applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE IF EXISTS sync_inbox ALTER COLUMN entity_id TYPE TEXT USING entity_id::text;
CREATE INDEX IF NOT EXISTS idx_sync_inbox_applied_at ON sync_inbox (applied_at);
CREATE INDEX IF NOT EXISTS idx_sync_inbox_batch_id ON sync_inbox (batch_id) WHERE batch_id IS NOT NULL;
`

// SyncServer implements syncpb.SyncServiceServer for gRPC bidirectional push, change_log pull, and persistent session streaming.
type SyncServer struct {
	syncpb.UnimplementedSyncServiceServer

	tenantManager    *manager.Manager
	masterPool       *pgxpool.Pool
	poolResolver     func(ctx context.Context, slug string) (*pgxpool.Pool, error)
	changeHub        *ChangeNotifyHub
	requireTokenAuth bool
}

// NewSyncServer returns a new initialized SyncServer instance.
func NewSyncServer(tm *manager.Manager, masterPool *pgxpool.Pool) *SyncServer {
	s := &SyncServer{
		tenantManager: tm,
		masterPool:    masterPool,
	}
	s.changeHub = NewChangeNotifyHub(s.getTenantPool)
	return s
}

// SetPoolResolver overrides tenant pool resolution (used in integration tests).
func (s *SyncServer) SetPoolResolver(fn func(ctx context.Context, slug string) (*pgxpool.Pool, error)) {
	s.poolResolver = fn
	if s.changeHub != nil {
		s.changeHub.poolResolver = s.getTenantPool
	}
}

// SetRequireTokenAuth enforces terminal sync token authentication on incoming sync streams.
func (s *SyncServer) SetRequireTokenAuth(require bool) {
	s.requireTokenAuth = require
}

// ChangeHub returns the server's PostgreSQL LISTEN/NOTIFY fan-out hub.
func (s *SyncServer) ChangeHub() *ChangeNotifyHub {
	if s.changeHub == nil {
		s.changeHub = NewChangeNotifyHub(s.getTenantPool)
	}
	return s.changeHub
}

func (s *SyncServer) getTenantPool(ctx context.Context, slug string) (*pgxpool.Pool, error) {
	if s.poolResolver != nil {
		return s.poolResolver(ctx, slug)
	}
	if s.tenantManager == nil {
		return nil, fmt.Errorf("tenant manager is not configured")
	}
	return s.tenantManager.GetPool(ctx, slug)
}

func extractPgSQLState(err error) string {
	if err == nil {
		return ""
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && len(pgErr.Code) == 5 {
		return pgErr.Code
	}
	return ""
}

func computePayloadSHA256(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

// StreamPush handles real-time streaming of local outbox items or atomic correlation batches from POS to Cloud.
func (s *SyncServer) StreamPush(stream syncpb.SyncService_StreamPushServer) error {
	ctx := stream.Context()
	var currentBatchID string
	var currentBatch []*syncpb.SyncEvent
	var (
		authChecked    bool
		authTerminalID int32
		authStoreID    int32
	)

	for {
		event, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Printf("[gRPC SyncServer] Error receiving push stream: %v", err)
			return err
		}

		if !authChecked && event.TenantSlug != "" {
			var authErr error
			authTerminalID, authStoreID, authErr = s.verifyStreamAuthIdentity(ctx, event.TenantSlug, event.DeviceId, event.StoreId)
			if authErr != nil {
				return authErr
			}
			authChecked = true
		}

		// Legacy single-event mode when batch_id is empty or batch_size <= 0
		if strings.TrimSpace(event.BatchId) == "" || event.BatchSize <= 0 {
			ack := s.handleSinglePushEventWithAuth(ctx, event, authTerminalID, authStoreID)
			if err := stream.Send(ack); err != nil {
				log.Printf("[gRPC SyncServer] Error sending SyncAck: %v", err)
				return err
			}
			continue
		}

		// Phase 2 Correlation Batch mode
		if currentBatchID != "" && event.BatchId != currentBatchID {
			// Previous batch was incomplete; reject it before starting the new batch
			ack := &syncpb.SyncAck{
				Id:           event.Id,
				BatchId:      currentBatchID,
				EntityKey:    syncpb.ResolveEntityKey(event.EntityKey, event.EntityId),
				Success:      false,
				ErrorMessage: fmt.Sprintf("incomplete batch %s: received %d events before new batch %s", currentBatchID, len(currentBatch), event.BatchId),
				ProcessedAt:  timestamppb.Now(),
			}
			if err := stream.Send(ack); err != nil {
				return err
			}
			currentBatchID = ""
			currentBatch = nil
		}

		if currentBatchID == "" {
			currentBatchID = event.BatchId
			currentBatch = make([]*syncpb.SyncEvent, 0, int(event.BatchSize))
		}
		currentBatch = append(currentBatch, event)

		if int32(len(currentBatch)) >= event.BatchSize {
			batchToApply := currentBatch
			currentBatchID = ""
			currentBatch = nil

			ack := s.processPushBatchWithAuth(ctx, batchToApply, authTerminalID, authStoreID)
			if err := stream.Send(ack); err != nil {
				log.Printf("[gRPC SyncServer] Error sending batch SyncAck: %v", err)
				return err
			}
		}
	}
}

func (s *SyncServer) handleSinglePushEvent(ctx context.Context, event *syncpb.SyncEvent) *syncpb.SyncAck {
	return s.handleSinglePushEventWithAuth(ctx, event, 0, 0)
}

func (s *SyncServer) handleSinglePushEventWithAuth(ctx context.Context, event *syncpb.SyncEvent, authTerminalID, authStoreID int32) *syncpb.SyncAck {
	started := time.Now()
	ctx = telemetry.WithTenantSlug(ctx, event.TenantSlug)
	syncStatus := "ok"
	defer func() {
		telemetry.RecordSyncBatch(ctx, "push", event.TenantSlug, event.EntityType, syncStatus, 1, float64(time.Since(started).Microseconds())/1000.0)
	}()

	entityKey := syncpb.ResolveEntityKey(event.EntityKey, event.EntityId)
	calculatedSha := computePayloadSHA256(event.PayloadJson)

	if authStoreID > 0 && event.StoreId > 0 && event.StoreId != authStoreID {
		syncStatus = "error"
		return &syncpb.SyncAck{
			Id:           event.Id,
			OutboxUuid:   event.OutboxUuid,
			EntityType:   event.EntityType,
			EntityId:     event.EntityId,
			EntityKey:    entityKey,
			Success:      false,
			ErrorMessage: fmt.Sprintf("store_id mismatch: event store %d does not match authenticated terminal store %d", event.StoreId, authStoreID),
			Sha256:       calculatedSha,
			ProcessedAt:  timestamppb.Now(),
		}
	}

	if event.Sha256 != "" && event.Sha256 != calculatedSha {
		syncStatus = "error"
		log.Printf("[gRPC SyncServer] SHA-256 mismatch for item %d (entity: %s): expected %s, got %s",
			event.Id, event.EntityType, event.Sha256, calculatedSha)
		return &syncpb.SyncAck{
			Id:           event.Id,
			OutboxUuid:   event.OutboxUuid,
			EntityType:   event.EntityType,
			EntityId:     event.EntityId,
			EntityKey:    entityKey,
			Success:      false,
			ErrorMessage: "SHA-256 checksum mismatch",
			Sha256:       calculatedSha,
			ProcessedAt:  timestamppb.Now(),
		}
	}

	if err := s.ingestSyncEventWithAuth(ctx, event, authTerminalID, authStoreID); err != nil {
		syncStatus = "error"
		log.Printf("[gRPC SyncServer] Failed to ingest item %d (%s): %v", event.Id, event.EntityType, err)
		return &syncpb.SyncAck{
			Id:           event.Id,
			OutboxUuid:   event.OutboxUuid,
			EntityType:   event.EntityType,
			EntityId:     event.EntityId,
			EntityKey:    entityKey,
			Success:      false,
			ErrorMessage: err.Error(),
			SqlState:     extractPgSQLState(err),
			Sha256:       calculatedSha,
			ProcessedAt:  timestamppb.Now(),
		}
	}

	return &syncpb.SyncAck{
		Id:          event.Id,
		OutboxUuid:  event.OutboxUuid,
		EntityType:  event.EntityType,
		EntityId:    event.EntityId,
		EntityKey:   entityKey,
		Success:     true,
		Sha256:      calculatedSha,
		ProcessedAt: timestamppb.Now(),
	}
}

// processPushBatch validates and applies an entire correlation batch inside one atomic PostgreSQL transaction
// with idempotent sync_inbox deduplication keyed by outbox_uuid.
func (s *SyncServer) processPushBatch(ctx context.Context, events []*syncpb.SyncEvent) *syncpb.SyncAck {
	return s.processPushBatchWithAuth(ctx, events, 0, 0)
}

func (s *SyncServer) processPushBatchWithAuth(ctx context.Context, events []*syncpb.SyncEvent, authTerminalID, authStoreID int32) *syncpb.SyncAck {
	if len(events) == 0 {
		return &syncpb.SyncAck{
			Success:      false,
			ErrorMessage: "empty correlation batch",
			ProcessedAt:  timestamppb.Now(),
		}
	}
	started := time.Now()
	first := events[0]
	last := events[len(events)-1]
	firstKey := syncpb.ResolveEntityKey(first.EntityKey, first.EntityId)
	lastKey := syncpb.ResolveEntityKey(last.EntityKey, last.EntityId)
	batchID := first.BatchId
	lastSha := computePayloadSHA256(last.PayloadJson)

	ctx = telemetry.WithTenantSlug(ctx, first.TenantSlug)
	batchStatus := "error"
	defer func() {
		telemetry.RecordSyncBatch(ctx, "push", first.TenantSlug, first.EntityType, batchStatus, int64(len(events)), float64(time.Since(started).Microseconds())/1000.0)
	}()

	// 1. Validate batch sequencing, store authorization, and SHA-256 checksums upfront before opening DB transaction
	for idx, ev := range events {
		expectedSeq := int32(idx + 1)
		evKey := syncpb.ResolveEntityKey(ev.EntityKey, ev.EntityId)
		if ev.SeqInBatch != 0 && ev.SeqInBatch != expectedSeq {
			return &syncpb.SyncAck{
				Id:           ev.Id,
				BatchId:      batchID,
				OutboxUuid:   ev.OutboxUuid,
				EntityType:   ev.EntityType,
				EntityId:     ev.EntityId,
				EntityKey:    evKey,
				Success:      false,
				FailedSeq:    expectedSeq,
				ErrorMessage: fmt.Sprintf("invalid seq_in_batch: expected %d, got %d", expectedSeq, ev.SeqInBatch),
				ProcessedAt:  timestamppb.Now(),
			}
		}
		if authStoreID > 0 && ev.StoreId > 0 && ev.StoreId != authStoreID {
			return &syncpb.SyncAck{
				Id:           ev.Id,
				BatchId:      batchID,
				OutboxUuid:   ev.OutboxUuid,
				EntityType:   ev.EntityType,
				EntityId:     ev.EntityId,
				EntityKey:    evKey,
				Success:      false,
				FailedSeq:    expectedSeq,
				ErrorMessage: fmt.Sprintf("store_id mismatch: event store %d does not match authenticated terminal store %d", ev.StoreId, authStoreID),
				ProcessedAt:  timestamppb.Now(),
			}
		}
		calcSha := computePayloadSHA256(ev.PayloadJson)
		if ev.Sha256 != "" && ev.Sha256 != calcSha {
			return &syncpb.SyncAck{
				Id:           ev.Id,
				BatchId:      batchID,
				OutboxUuid:   ev.OutboxUuid,
				EntityType:   ev.EntityType,
				EntityId:     ev.EntityId,
				EntityKey:    evKey,
				Success:      false,
				FailedSeq:    expectedSeq,
				ErrorMessage: "SHA-256 checksum mismatch",
				Sha256:       calcSha,
				ProcessedAt:  timestamppb.Now(),
			}
		}
	}

	if first.TenantSlug == "" {
		return &syncpb.SyncAck{
			Id:           last.Id,
			BatchId:      batchID,
			OutboxUuid:   first.OutboxUuid,
			EntityType:   first.EntityType,
			EntityId:     first.EntityId,
			EntityKey:    firstKey,
			Success:      false,
			FailedSeq:    1,
			ErrorMessage: "tenant_slug is required",
			ProcessedAt:  timestamppb.Now(),
		}
	}

	pool, err := s.getTenantPool(ctx, first.TenantSlug)
	if err != nil {
		return &syncpb.SyncAck{
			Id:           last.Id,
			BatchId:      batchID,
			OutboxUuid:   first.OutboxUuid,
			EntityType:   first.EntityType,
			EntityId:     first.EntityId,
			EntityKey:    firstKey,
			Success:      false,
			FailedSeq:    1,
			ErrorMessage: fmt.Sprintf("failed to get tenant pool: %v", err),
			ProcessedAt:  timestamppb.Now(),
		}
	}

	if _, err := pool.Exec(ctx, syncInboxDDL); err != nil {
		log.Printf("[gRPC SyncServer] sync_inbox DDL check warning: %v", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return &syncpb.SyncAck{
			Id:           last.Id,
			BatchId:      batchID,
			OutboxUuid:   first.OutboxUuid,
			EntityType:   first.EntityType,
			EntityId:     first.EntityId,
			EntityKey:    firstKey,
			Success:      false,
			FailedSeq:    1,
			SqlState:     extractPgSQLState(err),
			ErrorMessage: fmt.Sprintf("begin batch transaction: %v", err),
			ProcessedAt:  timestamppb.Now(),
		}
	}
	defer func() { _ = tx.Rollback(ctx) }()

	type postCommitHook struct {
		entityType string
		payload    []byte
	}
	var hooks []postCommitHook

	for idx, ev := range events {
		seqInBatch := int32(idx + 1)
		if ev.SeqInBatch > 0 {
			seqInBatch = ev.SeqInBatch
		}
		evKey := syncpb.ResolveEntityKey(ev.EntityKey, ev.EntityId)
		effectiveStoreID := ev.StoreId
		if authStoreID > 0 {
			effectiveStoreID = authStoreID
		}

		// Idempotent sync_inbox check keyed by outbox_uuid
		outboxUUID := strings.TrimSpace(ev.OutboxUuid)
		if outboxUUID != "" {
			if _, parseErr := uuid.Parse(outboxUUID); parseErr != nil {
				return &syncpb.SyncAck{
					Id:           ev.Id,
					BatchId:      batchID,
					OutboxUuid:   outboxUUID,
					EntityType:   ev.EntityType,
					EntityId:     ev.EntityId,
					EntityKey:    evKey,
					Success:      false,
					FailedSeq:    seqInBatch,
					ErrorMessage: fmt.Sprintf("invalid outbox_uuid %q: %v", outboxUUID, parseErr),
					ProcessedAt:  timestamppb.Now(),
				}
			}
			deviceID := ev.DeviceId
			if deviceID == "" {
				deviceID = fmt.Sprintf("store-%d", effectiveStoreID)
			}
			var insertedUUID string
			inboxErr := tx.QueryRow(ctx, `
				INSERT INTO sync_inbox (outbox_uuid, device_id, store_id, batch_id, entity_type, entity_id)
				VALUES ($1::uuid, $2, $3, $4, $5, $6)
				ON CONFLICT (outbox_uuid) DO NOTHING
				RETURNING outbox_uuid::text
			`, outboxUUID, deviceID, effectiveStoreID, batchID, ev.EntityType, evKey).Scan(&insertedUUID)
			if errors.Is(inboxErr, pgx.ErrNoRows) {
				// Duplicate outbox_uuid already applied in sync_inbox — skip idempotently!
				continue
			}
			if inboxErr != nil {
				return &syncpb.SyncAck{
					Id:           ev.Id,
					BatchId:      batchID,
					OutboxUuid:   outboxUUID,
					EntityType:   ev.EntityType,
					EntityId:     ev.EntityId,
					EntityKey:    evKey,
					Success:      false,
					FailedSeq:    seqInBatch,
					SqlState:     extractPgSQLState(inboxErr),
					ErrorMessage: fmt.Sprintf("sync_inbox check failed: %v", inboxErr),
					ProcessedAt:  timestamppb.Now(),
				}
			}
		}

		if !isSupportedPushEntity(ev.EntityType) {
			return &syncpb.SyncAck{
				Id:           ev.Id,
				BatchId:      batchID,
				OutboxUuid:   ev.OutboxUuid,
				EntityType:   ev.EntityType,
				EntityId:     ev.EntityId,
				EntityKey:    evKey,
				Success:      false,
				FailedSeq:    seqInBatch,
				ErrorMessage: fmt.Sprintf("unsupported sync entity type: %s", ev.EntityType),
				ProcessedAt:  timestamppb.Now(),
			}
		}

		if err := s.upsertEntityJSONWithAuthDB(ctx, tx, pool, ev.EntityType, ev.Action, ev.PayloadJson, false, effectiveStoreID, authTerminalID); err != nil {
			log.Printf("[gRPC SyncServer] Batch %s failed at seq %d (%s key %s): %v",
				batchID, seqInBatch, ev.EntityType, evKey, err)
			return &syncpb.SyncAck{
				Id:           ev.Id,
				BatchId:      batchID,
				OutboxUuid:   ev.OutboxUuid,
				EntityType:   ev.EntityType,
				EntityId:     ev.EntityId,
				EntityKey:    evKey,
				Success:      false,
				FailedSeq:    seqInBatch,
				SqlState:     extractPgSQLState(err),
				ErrorMessage: err.Error(),
				ProcessedAt:  timestamppb.Now(),
			}
		}

		if ev.EntityType == "pos_transactions" || ev.EntityType == "pos_transaction_lines" {
			hooks = append(hooks, postCommitHook{entityType: ev.EntityType, payload: ev.PayloadJson})
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return &syncpb.SyncAck{
			Id:           last.Id,
			BatchId:      batchID,
			OutboxUuid:   last.OutboxUuid,
			EntityType:   last.EntityType,
			EntityId:     last.EntityId,
			EntityKey:    lastKey,
			Success:      false,
			FailedSeq:    int32(len(events)),
			SqlState:     extractPgSQLState(err),
			ErrorMessage: fmt.Sprintf("commit batch transaction: %v", err),
			ProcessedAt:  timestamppb.Now(),
		}
	}

	// Execute post-commit hooks and watermark updates now that all batch rows are committed
	for _, h := range hooks {
		if h.entityType == "pos_transactions" {
			enqueuePosReceiptAfterSync(ctx, pool, h.payload)
		} else if h.entityType == "pos_transaction_lines" {
			EnqueuePosReceiptAfterLineSync(ctx, pool, h.payload)
		}
	}
	s.recordPushWatermark(ctx, pool, last)
	batchStatus = "ok"

	return &syncpb.SyncAck{
		Id:          last.Id,
		BatchId:     batchID,
		OutboxUuid:  last.OutboxUuid,
		EntityType:  last.EntityType,
		EntityId:    last.EntityId,
		EntityKey:   lastKey,
		Success:     true,
		Sha256:      lastSha,
		ProcessedAt: timestamppb.Now(),
	}
}

func isSupportedPushEntity(entityType string) bool {
	switch entityType {
	case "pos_transactions", "pos_transaction_lines", "pos_payments", "cashier_sessions",
		"customers",
		"sales_returns", "sales_return_lines", "stock_movements",
		"carts", "sales_orders_v2", "sales_order_lines_v2", "draft_cart_templates", "restaurant_orders", "restaurant_order_items",
		"kiosk_sessions", "stock_counts", "stock_count_lines", "waste_logs", "user_biometric_templates":
		return true
	default:
		return false
	}
}

// ingestSyncEvent processes and persists a single incoming entity payload (legacy or non-batched).
func (s *SyncServer) ingestSyncEvent(ctx context.Context, event *syncpb.SyncEvent) error {
	return s.ingestSyncEventWithAuth(ctx, event, 0, 0)
}

func (s *SyncServer) ingestSyncEventWithAuth(ctx context.Context, event *syncpb.SyncEvent, authTerminalID, authStoreID int32) error {
	if event.TenantSlug == "" {
		return fmt.Errorf("tenant_slug is required")
	}
	if authStoreID > 0 && event.StoreId > 0 && event.StoreId != authStoreID {
		return fmt.Errorf("store_id mismatch: event store %d does not match authenticated terminal store %d", event.StoreId, authStoreID)
	}
	effectiveStoreID := event.StoreId
	if authStoreID > 0 {
		effectiveStoreID = authStoreID
	}
	entityKey := syncpb.ResolveEntityKey(event.EntityKey, event.EntityId)

	pool, err := s.getTenantPool(ctx, event.TenantSlug)
	if err != nil {
		return fmt.Errorf("failed to get tenant pool: %w", err)
	}

	if !isSupportedPushEntity(event.EntityType) {
		return fmt.Errorf("unsupported sync entity type: %s", event.EntityType)
	}

	outboxUUID := strings.TrimSpace(event.OutboxUuid)
	if outboxUUID != "" {
		if _, parseErr := uuid.Parse(outboxUUID); parseErr == nil {
			_, _ = pool.Exec(ctx, syncInboxDDL)
			tx, txErr := pool.Begin(ctx)
			if txErr != nil {
				return txErr
			}
			defer func() { _ = tx.Rollback(ctx) }()

			deviceID := event.DeviceId
			if deviceID == "" {
				deviceID = fmt.Sprintf("store-%d", effectiveStoreID)
			}
			var insertedUUID string
			inboxErr := tx.QueryRow(ctx, `
				INSERT INTO sync_inbox (outbox_uuid, device_id, store_id, batch_id, entity_type, entity_id)
				VALUES ($1::uuid, $2, $3, NULLIF($4, ''), $5, $6)
				ON CONFLICT (outbox_uuid) DO NOTHING
				RETURNING outbox_uuid::text
			`, outboxUUID, deviceID, effectiveStoreID, event.BatchId, event.EntityType, entityKey).Scan(&insertedUUID)
			if errors.Is(inboxErr, pgx.ErrNoRows) {
				return nil
			}
			if inboxErr != nil {
				return inboxErr
			}
			if err := s.upsertEntityJSONWithAuthDB(ctx, tx, pool, event.EntityType, event.Action, event.PayloadJson, false, effectiveStoreID, authTerminalID); err != nil {
				return err
			}
			if err := tx.Commit(ctx); err != nil {
				return err
			}
			if event.EntityType == "pos_transactions" {
				enqueuePosReceiptAfterSync(ctx, pool, event.PayloadJson)
			}
			if event.EntityType == "pos_transaction_lines" {
				EnqueuePosReceiptAfterLineSync(ctx, pool, event.PayloadJson)
			}
			s.recordPushWatermark(ctx, pool, event)
			return nil
		}
	}

	var legacyDB repository.DBTX
	if pool != nil {
		legacyDB = pool
	}
	if err := s.upsertEntityJSONWithAuthDB(ctx, legacyDB, pool, event.EntityType, event.Action, event.PayloadJson, true, effectiveStoreID, authTerminalID); err != nil {
		log.Printf("[gRPC SyncServer] Ingestion error for %s (action=%s, key %s): %v", event.EntityType, event.Action, entityKey, err)
		return err
	}
	log.Printf("[gRPC SyncServer] Successfully ingested %s entity key %s (action=%s, Correlation: %s) for store %d",
		event.EntityType, entityKey, event.Action, event.CorrelationId, effectiveStoreID)

	s.recordPushWatermark(ctx, pool, event)
	return nil
}

func (s *SyncServer) recordPushWatermark(ctx context.Context, pool *pgxpool.Pool, event *syncpb.SyncEvent) {
	if pool == nil || event == nil {
		return
	}
	metadataJSON, _ := json.Marshal(map[string]interface{}{
		"last_event_id":  event.Id,
		"action":         event.Action,
		"correlation_id": event.CorrelationId,
	})
	_, err := pool.Exec(ctx, `
		INSERT INTO sync_watermarks (entity_type, store_id, last_sync_at, metadata)
		VALUES ($1, $2, NOW(), $3)
		ON CONFLICT (entity_type, store_id) DO UPDATE SET
			last_sync_at = EXCLUDED.last_sync_at,
			metadata     = EXCLUDED.metadata;
	`, event.EntityType, event.StoreId, string(metadataJSON))
	if err != nil {
		log.Printf("[gRPC SyncServer] Watermark update warning: %v", err)
	}
}

func getTableUpdateClauseDB(ctx context.Context, db repository.DBTX, table string) (string, error) {
	if db == nil {
		return "", fmt.Errorf("read columns for %s: database pool is nil", table)
	}
	rows, err := db.Query(ctx, `
		SELECT column_name 
		FROM information_schema.columns 
		WHERE table_schema = current_schema() AND table_name = $1
		  AND column_name NOT IN ('id', 'created_at') AND is_generated = 'NEVER'
		ORDER BY ordinal_position;
	`, table)
	if err != nil {
		return "", fmt.Errorf("read columns for %s: %w", table, err)
	}
	defer rows.Close()

	var sets []string
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return "", err
		}
		sets = append(sets, fmt.Sprintf("%s = EXCLUDED.%s", col, col))
	}

	if err := rows.Err(); err != nil {
		return "", err
	}
	clause := strings.Join(sets, ", ")
	if clause == "" {
		clause = "id = EXCLUDED.id"
	}
	return clause, nil
}

func getTableUpdateClause(ctx context.Context, pool *pgxpool.Pool, table string) (string, error) {
	if pool == nil {
		return "", fmt.Errorf("read columns for %s: database pool is nil", table)
	}
	return getTableUpdateClauseDB(ctx, pool, table)
}

// upsertEntityJSON preserves FK checks. The client sends checkout parents
// before POS rows; bypassing constraints would leave orphaned sales behind.
func (s *SyncServer) upsertEntityJSON(ctx context.Context, pool *pgxpool.Pool, entityType, action string, payload []byte) error {
	var db repository.DBTX
	if pool != nil {
		db = pool
	}
	return s.upsertEntityJSONWithDB(ctx, db, pool, entityType, action, payload, true)
}

func (s *SyncServer) upsertEntityJSONWithDB(ctx context.Context, db repository.DBTX, pool *pgxpool.Pool, entityType, action string, payload []byte, runPostCommitHooks bool) error {
	return s.upsertEntityJSONWithStoreDB(ctx, db, pool, entityType, action, payload, runPostCommitHooks, 0)
}

func (s *SyncServer) upsertEntityJSONWithStoreDB(ctx context.Context, db repository.DBTX, pool *pgxpool.Pool, entityType, action string, payload []byte, runPostCommitHooks bool, authorizedStoreID int32) error {
	return s.upsertEntityJSONWithAuthDB(ctx, db, pool, entityType, action, payload, runPostCommitHooks, authorizedStoreID, 0)
}

func (s *SyncServer) upsertEntityJSONWithAuthDB(ctx context.Context, db repository.DBTX, pool *pgxpool.Pool, entityType, action string, payload []byte, runPostCommitHooks bool, authorizedStoreID, authenticatedTerminalID int32) error {
	if entityType == "sales_orders_v2" && action == "POS_CHECKOUT" {
		if tx, ok := db.(pgx.Tx); ok {
			return s.applyPOSCheckoutInTx(ctx, tx, payload)
		}
		return s.applyPOSCheckout(ctx, pool, payload)
	}
	if len(payload) == 0 {
		return fmt.Errorf("empty sync payload for %s", entityType)
	}
	if entityType == "customers" {
		if action != "INSERT" {
			return fmt.Errorf("unsupported customer sync action: %s", action)
		}
		return upsertCreatedCustomerDB(ctx, db, payload)
	}
	validTables := map[string]bool{
		"pos_transactions": true, "pos_transaction_lines": true, "pos_payments": true,
		"sales_returns": true, "sales_return_lines": true, "stock_movements": true,
		"cashier_sessions": true, "carts": true, "sales_orders_v2": true, "sales_order_lines_v2": true,
		"draft_cart_templates": true, "restaurant_orders": true, "restaurant_order_items": true,
		"kiosk_sessions": true, "stock_counts": true, "stock_count_lines": true, "waste_logs": true,
		"notification_events": true, "user_biometric_templates": true,
	}
	if !validTables[entityType] {
		return fmt.Errorf("unsupported entity type for upsert: %s", entityType)
	}
	if err := validateEntityOwnership(ctx, db, entityType, payload, authorizedStoreID, authenticatedTerminalID); err != nil {
		return err
	}
	updateClause, err := getTableUpdateClauseDB(ctx, db, entityType)
	if err != nil {
		return err
	}
	if entityType == "user_biometric_templates" {
		if err := usecase.ApplyUserBiometricTemplateFromSync(ctx, db, payload); err != nil {
			return fmt.Errorf("failed to upsert %s (action=%s): %w", entityType, action, err)
		}
		return nil
	}
	if entityType == "sales_orders_v2" {
		// Legacy header events must not undo a completed inventory transition.
		updateClause = strings.ReplaceAll(updateClause, "fulfillment_status = EXCLUDED.fulfillment_status",
			"fulfillment_status = CASE WHEN sales_orders_v2.fulfillment_status = 'fulfilled' THEN sales_orders_v2.fulfillment_status ELSE EXCLUDED.fulfillment_status END")
	}
	var applyRecipeStock bool
	var recipeMovement stockMovementSyncPayload
	if entityType == "stock_movements" {
		if err := json.Unmarshal(payload, &recipeMovement); err == nil &&
			(recipeMovement.MovementType == "recipe_deduction" || recipeMovement.MovementType == "recipe_reversal") {
			if authorizedStoreID > 0 && (recipeMovement.FromStoreID == nil || *recipeMovement.FromStoreID != authorizedStoreID) {
				return fmt.Errorf("unauthorized from_store_id for terminal store %d", authorizedStoreID)
			}
			if _, parseErr := uuid.Parse(recipeMovement.ID); parseErr == nil {
				var existingID string
				chkErr := db.QueryRow(ctx, `SELECT id::text FROM stock_movements WHERE id = $1::uuid`, recipeMovement.ID).Scan(&existingID)
				if errors.Is(chkErr, pgx.ErrNoRows) {
					applyRecipeStock = true
				} else if chkErr != nil {
					return fmt.Errorf("check existing stock_movement %s: %w", recipeMovement.ID, chkErr)
				}
			}
		}
	}
	conflictWhere := entityConflictOwnershipWhere(entityType)
	query := fmt.Sprintf(`
		INSERT INTO %s
		SELECT * FROM json_populate_record(NULL::%s, $1::json)
		ON CONFLICT (id) DO UPDATE SET %s%s;
	`, entityType, entityType, updateClause, conflictWhere)
	tag, err := db.Exec(ctx, query, string(payload))
	if err != nil {
		// Return the original constraint and SQLSTATE. A DO NOTHING fallback
		// can hide rejected updates, or replace the cause with SQLSTATE 25P02.
		return fmt.Errorf("failed to upsert %s (action=%s): %w", entityType, action, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("ownership conflict on %s: row belongs to another terminal or store", entityType)
	}
	if applyRecipeStock {
		if err := applyRecipeStockMovementSync(ctx, db, recipeMovement); err != nil {
			return fmt.Errorf("apply cloud stock update for %s %s: %w", recipeMovement.MovementType, recipeMovement.ID, err)
		}
	}
	if runPostCommitHooks && pool != nil {
		if entityType == "pos_transactions" {
			enqueuePosReceiptAfterSync(ctx, pool, payload)
		}
		if entityType == "pos_transaction_lines" {
			EnqueuePosReceiptAfterLineSync(ctx, pool, payload)
		}
	}
	return nil
}

func entityConflictOwnershipWhere(entityType string) string {
	switch entityType {
	case "cashier_sessions":
		return " WHERE cashier_sessions.pos_terminal_id = EXCLUDED.pos_terminal_id"
	case "pos_transactions":
		return " WHERE pos_transactions.store_id = EXCLUDED.store_id AND (pos_transactions.pos_terminal_id IS NULL OR pos_transactions.pos_terminal_id = EXCLUDED.pos_terminal_id)"
	case "pos_transaction_lines":
		return " WHERE pos_transaction_lines.transaction_id = EXCLUDED.transaction_id"
	case "pos_payments":
		return " WHERE pos_payments.transaction_id = EXCLUDED.transaction_id"
	case "sales_returns":
		return " WHERE sales_returns.store_id = EXCLUDED.store_id"
	case "sales_return_lines":
		return " WHERE sales_return_lines.return_id = EXCLUDED.return_id"
	case "restaurant_orders":
		return " WHERE restaurant_orders.store_id = EXCLUDED.store_id AND (restaurant_orders.pos_terminal_id IS NULL OR restaurant_orders.pos_terminal_id = EXCLUDED.pos_terminal_id)"
	case "restaurant_order_items":
		return " WHERE restaurant_order_items.order_id = EXCLUDED.order_id"
	case "kiosk_sessions":
		return " WHERE kiosk_sessions.store_id = EXCLUDED.store_id AND kiosk_sessions.pos_terminal_id = EXCLUDED.pos_terminal_id"
	case "stock_counts":
		return " WHERE stock_counts.store_id = EXCLUDED.store_id"
	case "stock_count_lines":
		return " WHERE stock_count_lines.stock_count_id = EXCLUDED.stock_count_id"
	case "waste_logs":
		return " WHERE waste_logs.store_id = EXCLUDED.store_id"
	default:
		return ""
	}
}

func validateEntityOwnership(ctx context.Context, db repository.DBTX, entityType string, payload []byte, authorizedStoreID, authenticatedTerminalID int32) error {
	if authorizedStoreID <= 0 && authenticatedTerminalID <= 0 {
		return nil
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil
	}
	payloadStoreID := int32(jsonInt64(raw["store_id"]))
	payloadTerminalID := int32(jsonInt64(raw["pos_terminal_id"]))
	transactionID := jsonEntityIDString(raw["transaction_id"])
	returnID := jsonEntityIDString(raw["return_id"])
	orderID := jsonEntityIDString(raw["order_id"])
	stockCountID := jsonEntityIDString(raw["stock_count_id"])

	switch entityType {
	case "cashier_sessions", "pos_transactions", "restaurant_orders", "kiosk_sessions":
		if authenticatedTerminalID > 0 && payloadTerminalID > 0 && payloadTerminalID != authenticatedTerminalID {
			return fmt.Errorf("ownership violation on %s: payload pos_terminal_id %d does not match authenticated terminal %d",
				entityType, payloadTerminalID, authenticatedTerminalID)
		}
	}

	switch entityType {
	case "pos_transactions", "restaurant_orders", "kiosk_sessions", "sales_returns", "stock_counts", "waste_logs":
		if authorizedStoreID > 0 && payloadStoreID > 0 && payloadStoreID != authorizedStoreID {
			return fmt.Errorf("ownership violation on %s: payload store_id %d does not match authorized store %d",
				entityType, payloadStoreID, authorizedStoreID)
		}
	}

	if db == nil {
		return nil
	}

	if entityType == "cashier_sessions" && authorizedStoreID > 0 && payloadTerminalID > 0 {
		var hasTerminalsTable bool
		if err := db.QueryRow(ctx, `SELECT to_regclass('pos_terminals') IS NOT NULL`).Scan(&hasTerminalsTable); err == nil && hasTerminalsTable {
			var termStoreID int32
			if err := db.QueryRow(ctx, `SELECT store_id FROM pos_terminals WHERE id = $1`, payloadTerminalID).Scan(&termStoreID); err == nil && termStoreID > 0 && termStoreID != authorizedStoreID {
				return fmt.Errorf("ownership violation on cashier_sessions: terminal %d belongs to store %d, not authorized store %d",
					payloadTerminalID, termStoreID, authorizedStoreID)
			}
		}
	}

	switch entityType {
	case "pos_transaction_lines", "pos_payments":
		if transactionID != "" {
			var parentStoreID, parentTerminalID int32
			err := db.QueryRow(ctx, `SELECT COALESCE(store_id, 0), COALESCE(pos_terminal_id, 0) FROM pos_transactions WHERE id::text = $1`, transactionID).
				Scan(&parentStoreID, &parentTerminalID)
			if err == nil {
				if authorizedStoreID > 0 && parentStoreID > 0 && parentStoreID != authorizedStoreID {
					return fmt.Errorf("ownership violation on %s: parent transaction %s belongs to store %d, not authorized store %d",
						entityType, transactionID, parentStoreID, authorizedStoreID)
				}
				if authenticatedTerminalID > 0 && parentTerminalID > 0 && parentTerminalID != authenticatedTerminalID {
					return fmt.Errorf("ownership violation on %s: parent transaction %s belongs to terminal %d, not authenticated terminal %d",
						entityType, transactionID, parentTerminalID, authenticatedTerminalID)
				}
			}
		}
	case "sales_return_lines":
		if returnID != "" {
			var parentStoreID int32
			err := db.QueryRow(ctx, `SELECT COALESCE(store_id, 0) FROM sales_returns WHERE id::text = $1`, returnID).
				Scan(&parentStoreID)
			if err == nil && authorizedStoreID > 0 && parentStoreID > 0 && parentStoreID != authorizedStoreID {
				return fmt.Errorf("ownership violation on sales_return_lines: parent return %s belongs to store %d, not authorized store %d",
					returnID, parentStoreID, authorizedStoreID)
			}
		}
	case "restaurant_order_items":
		if orderID != "" {
			var parentStoreID, parentTerminalID int32
			err := db.QueryRow(ctx, `SELECT COALESCE(store_id, 0), COALESCE(pos_terminal_id, 0) FROM restaurant_orders WHERE id::text = $1`, orderID).
				Scan(&parentStoreID, &parentTerminalID)
			if err == nil {
				if authorizedStoreID > 0 && parentStoreID > 0 && parentStoreID != authorizedStoreID {
					return fmt.Errorf("ownership violation on restaurant_order_items: parent order %s belongs to store %d, not authorized store %d",
						orderID, parentStoreID, authorizedStoreID)
				}
				if authenticatedTerminalID > 0 && parentTerminalID > 0 && parentTerminalID != authenticatedTerminalID {
					return fmt.Errorf("ownership violation on restaurant_order_items: parent order %s belongs to terminal %d, not authenticated terminal %d",
						orderID, parentTerminalID, authenticatedTerminalID)
				}
			}
		}
	case "stock_count_lines":
		if stockCountID != "" {
			var parentStoreID int32
			err := db.QueryRow(ctx, `SELECT COALESCE(store_id, 0) FROM stock_counts WHERE id::text = $1`, stockCountID).
				Scan(&parentStoreID)
			if err == nil && authorizedStoreID > 0 && parentStoreID > 0 && parentStoreID != authorizedStoreID {
				return fmt.Errorf("ownership violation on stock_count_lines: parent stock_count %s belongs to store %d, not authorized store %d",
					stockCountID, parentStoreID, authorizedStoreID)
			}
		}
	}
	return nil
}

// A customer ID is referenced by carts, orders, and transactions. Never
// overwrite a different cloud customer merely because the ID matches.
func upsertCreatedCustomer(ctx context.Context, pool *pgxpool.Pool, payload []byte) error {
	return upsertCreatedCustomerDB(ctx, pool, payload)
}

func upsertCreatedCustomerDB(ctx context.Context, db repository.DBTX, payload []byte) error {
	var incoming struct {
		ID             string `json:"id"`
		OrganizationID int32  `json:"organization_id"`
		CustomerCode   string `json:"customer_code"`
	}
	if err := json.Unmarshal(payload, &incoming); err != nil {
		return fmt.Errorf("invalid customer sync payload: %w", err)
	}
	incoming.ID = strings.TrimSpace(incoming.ID)
	if _, err := uuid.Parse(incoming.ID); err != nil || incoming.OrganizationID <= 0 || incoming.CustomerCode == "" {
		return fmt.Errorf("customer sync requires id, organization_id, and customer_code")
	}
	var insertedID string
	err := db.QueryRow(ctx, `INSERT INTO customers
		SELECT * FROM json_populate_record(NULL::customers, $1::json)
		ON CONFLICT (id) DO NOTHING RETURNING id::text`, string(payload)).Scan(&insertedID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to insert customer %s: %w", incoming.ID, err)
	}
	var organizationID int32
	var customerCode string
	if err := db.QueryRow(ctx, `SELECT organization_id, customer_code FROM customers WHERE id = $1::uuid`, incoming.ID).
		Scan(&organizationID, &customerCode); err != nil {
		return fmt.Errorf("read existing customer %s: %w", incoming.ID, err)
	}
	if organizationID != incoming.OrganizationID || customerCode != incoming.CustomerCode {
		return fmt.Errorf("customer ID %s already belongs to another cloud customer; resolve the ID collision before syncing dependent sales", incoming.ID)
	}
	return nil
}

// applyPOSCheckout saves the complete order before the fulfillment transition.
// Stock changes and fulfilled status commit together; retries never reset it.
func (s *SyncServer) applyPOSCheckout(ctx context.Context, pool *pgxpool.Pool, payload []byte) error {
	if pool == nil {
		return fmt.Errorf("database pool is nil")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := s.applyPOSCheckoutInTx(ctx, tx, payload); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *SyncServer) applyPOSCheckoutInTx(ctx context.Context, tx pgx.Tx, payload []byte) error {
	var checkout struct {
		Order map[string]json.RawMessage `json:"order"`
		Lines []json.RawMessage          `json:"lines"`
	}
	if err := json.Unmarshal(payload, &checkout); err != nil {
		return fmt.Errorf("invalid checkout: %w", err)
	}
	var orderID, status string
	if err := json.Unmarshal(checkout.Order["id"], &orderID); err != nil {
		return fmt.Errorf("checkout order ID is required")
	}
	if err := json.Unmarshal(checkout.Order["fulfillment_status"], &status); err != nil || status != "fulfilled" {
		return fmt.Errorf("checkout must be fulfilled locally")
	}
	if len(checkout.Lines) == 0 {
		return fmt.Errorf("checkout has no order lines")
	}
	for _, line := range checkout.Lines {
		var ref struct {
			OrderID string `json:"sales_order_id"`
		}
		if err := json.Unmarshal(line, &ref); err != nil || ref.OrderID != orderID {
			return fmt.Errorf("checkout line does not belong to order %s", orderID)
		}
	}
	orderClause, err := getTableUpdateClauseDB(ctx, tx, "sales_orders_v2")
	if err != nil {
		return err
	}
	lineClause, err := getTableUpdateClauseDB(ctx, tx, "sales_order_lines_v2")
	if err != nil {
		return err
	}
	// Also serialize simultaneous first inserts, when there is no row to lock.
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1, 0))", orderID); err != nil {
		return err
	}
	var existingStatus string
	err = tx.QueryRow(ctx, "SELECT fulfillment_status::text FROM sales_orders_v2 WHERE id=$1::uuid FOR UPDATE", orderID).Scan(&existingStatus)
	if err != nil && err != pgx.ErrNoRows {
		return err
	}
	if err == nil && existingStatus == "fulfilled" {
		return nil
	}
	var triggerEnabled bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1 FROM pg_trigger WHERE tgrelid='sales_orders_v2'::regclass
		AND tgname='trg_deduct_inventory_on_fulfillment' AND tgenabled IN ('O','A')
	) AND NOT EXISTS (
		SELECT 1 FROM pg_trigger WHERE tgrelid=to_regclass('pos_transaction_lines')
		AND tgname='trg_deduct_inventory_on_pos_transaction' AND tgenabled IN ('O','A')
	) AND current_setting('session_replication_role') <> 'replica'`).Scan(&triggerEnabled); err != nil {
		return err
	}
	if !triggerEnabled {
		return fmt.Errorf("cloud inventory requires the fulfillment trigger enabled and the legacy POS deduction trigger disabled")
	}
	checkout.Order["fulfillment_status"] = json.RawMessage(`"unfulfilled"`)
	orderJSON, err := json.Marshal(checkout.Order)
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`INSERT INTO sales_orders_v2 SELECT * FROM json_populate_record(NULL::sales_orders_v2,$1::json)
		ON CONFLICT(id) DO UPDATE SET %s`, orderClause)
	if _, err := tx.Exec(ctx, query, string(orderJSON)); err != nil {
		return fmt.Errorf("save checkout order: %w", err)
	}
	query = fmt.Sprintf(`INSERT INTO sales_order_lines_v2 SELECT * FROM json_populate_record(NULL::sales_order_lines_v2,$1::json)
		ON CONFLICT(id) DO UPDATE SET %s WHERE sales_order_lines_v2.sales_order_id=EXCLUDED.sales_order_id`, lineClause)
	for _, line := range checkout.Lines {
		tag, err := tx.Exec(ctx, query, string(line))
		if err != nil {
			return fmt.Errorf("save checkout line: %w", err)
		}
		if tag.RowsAffected() != 1 {
			return fmt.Errorf("checkout line ID belongs to another order")
		}
	}
	// The existing trigger performs deduction with all lines present. Its writes
	// roll back with this transaction if any part of fulfillment fails.
	if _, err := tx.Exec(ctx, "UPDATE sales_orders_v2 SET fulfillment_status='fulfilled' WHERE id=$1::uuid", orderID); err != nil {
		return fmt.Errorf("fulfill checkout: %w", err)
	}
	return nil
}

// StreamPull handles Cloud -> Local Terminal delta updates based on sync watermarks (legacy).
func (s *SyncServer) StreamPull(req *syncpb.PullRequest, stream syncpb.SyncService_StreamPullServer) error {
	if req.TenantSlug == "" {
		return status.Error(codes.InvalidArgument, "tenant_slug is required")
	}

	ctx := telemetry.WithTenantSlug(stream.Context(), req.TenantSlug)
	if err := s.verifyStreamAuth(ctx, req.TenantSlug, "", req.StoreId); err != nil {
		return err
	}

	pool, err := s.getTenantPool(ctx, req.TenantSlug)
	if err != nil {
		return status.Errorf(codes.NotFound, "tenant DB pool unavailable: %v", err)
	}

	var sinceTime time.Time
	if req.Since != nil {
		sinceTime = req.Since.AsTime()
	} else {
		sinceTime = time.Unix(0, 0)
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > syncpb.PullPageLimit {
		limit = syncpb.PullPageLimit
	}
	afterKey, hasCursorMeta := resolvePullAfterKey(ctx, req)
	if afterKey != "" || hasCursorMeta {
		if len(req.EntityTypes) != 1 || !isValidPullCursor(req.EntityTypes[0], afterKey) {
			return status.Error(codes.InvalidArgument, "pull cursor requires a valid entity key and exactly one entity")
		}
		if hasCursorMeta {
			if err := stream.SendHeader(metadata.Pairs(syncpb.PullCursorVersionKey, "1")); err != nil {
				return err
			}
		}
	}

	// Delta entity target categories
	targetEntities := req.EntityTypes
	if len(targetEntities) == 0 {
		targetEntities = syncpb.PullEntityTypes()
	}
	for _, entityType := range targetEntities {
		if !syncpb.IsPullEntity(entityType) {
			return status.Errorf(codes.InvalidArgument, "unsupported pull entity: %s", entityType)
		}
	}

	for _, entityType := range targetEntities {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		entityStart := time.Now()
		err := s.streamEntityDeltas(ctx, pool, req, entityType, sinceTime, limit, stream)
		durMs := float64(time.Since(entityStart).Microseconds()) / 1000.0
		if err != nil {
			telemetry.RecordSyncBatch(ctx, "pull", req.TenantSlug, entityType, "error", 1, durMs)
			log.Printf("[gRPC SyncServer] Error streaming deltas for %s: %v", entityType, err)
			return status.Errorf(codes.Internal, "stream %s: %v", entityType, err)
		}
		telemetry.RecordSyncBatch(ctx, "pull", req.TenantSlug, entityType, "ok", 1, durMs)
	}

	return nil
}

func resolvePullAfterKey(ctx context.Context, req *syncpb.PullRequest) (string, bool) {
	afterKey := strings.TrimSpace(req.GetAfterKey())
	var hasCursorMeta bool
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(syncpb.PullCursorAfterKey); len(vals) > 0 {
			hasCursorMeta = true
			if afterKey == "" {
				afterKey = strings.TrimSpace(vals[0])
			}
		} else if vals := md.Get(syncpb.PullCursorKey); len(vals) > 0 {
			hasCursorMeta = true
			if afterKey == "" {
				afterKey = strings.TrimSpace(vals[0])
			}
		}
	}
	return afterKey, hasCursorMeta
}

func isValidPullCursor(entityType, afterKey string) bool {
	if afterKey == "" {
		return false
	}
	if syncpb.IsUUIDEntity(entityType) {
		_, err := uuid.Parse(afterKey)
		return err == nil
	}
	afterID, err := strconv.ParseInt(afterKey, 10, 64)
	return err == nil && afterID >= 0
}

func (s *SyncServer) streamEntityDeltas(
	ctx context.Context,
	pool *pgxpool.Pool,
	req *syncpb.PullRequest,
	entityType string,
	since time.Time,
	limit int32,
	stream syncpb.SyncService_StreamPullServer,
) error {
	// Query modified entities updated after watermark using PostgreSQL row_to_json
	// Include the entire timestamp boundary so the next request's strict >
	// predicate cannot skip rows sharing the last timestamp in this page.
	ident := pgx.Identifier{entityType}.Sanitize()
	query := fmt.Sprintf(`SELECT id::text, row_to_json(t)::text, updated_at FROM %s t WHERE updated_at > $1 ORDER BY updated_at ASC FETCH FIRST ($2) ROWS WITH TIES`, ident)
	args := []any{since, limit}
	afterKey, hasCursorMeta := resolvePullAfterKey(ctx, req)
	if afterKey != "" || hasCursorMeta {
		if syncpb.IsUUIDEntity(entityType) {
			if _, err := uuid.Parse(afterKey); err != nil {
				return status.Error(codes.InvalidArgument, "invalid UUID pull cursor")
			}
			query = fmt.Sprintf(`SELECT id::text, row_to_json(t)::text, updated_at FROM %s t
				WHERE updated_at > $1 OR (updated_at = $1 AND id > $3::uuid)
				ORDER BY updated_at, id LIMIT $2`, ident)
			args = append(args, afterKey)
		} else {
			afterID, err := strconv.ParseInt(afterKey, 10, 64)
			if err != nil || afterID < 0 {
				return status.Error(codes.InvalidArgument, "invalid pull cursor")
			}
			query = fmt.Sprintf(`SELECT id::text, row_to_json(t)::text, updated_at FROM %s t
				WHERE updated_at > $1 OR (updated_at = $1 AND id > $3::bigint)
				ORDER BY updated_at, id LIMIT $2`, ident)
			args = append(args, afterID)
		}
	}

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query %s: %w", entityType, err)
	}
	defer rows.Close()

	for rows.Next() {
		var idStr string
		var jsonStr string
		var updatedAt time.Time

		if err := rows.Scan(&idStr, &jsonStr, &updatedAt); err != nil {
			return fmt.Errorf("scan %s: %w", entityType, err)
		}

		payloadBytes := []byte(jsonStr)
		shaHex := computePayloadSHA256(payloadBytes)

		event := &syncpb.SyncEvent{
			EntityType:  entityType,
			EntityKey:   idStr,
			Action:      "UPDATE",
			PayloadJson: payloadBytes,
			StoreId:     req.StoreId,
			TenantSlug:  req.TenantSlug,
			EventTime:   timestamppb.New(updatedAt),
			Sha256:      shaHex,
			IsLastChunk: true,
		}
		if numID, ok := syncpb.ParseNumericEntityID(idStr); ok {
			event.Id = numID
			event.EntityId = numID
		}

		if err := stream.Send(event); err != nil {
			return err
		}
	}

	return rows.Err()
}

type stockMovementSyncPayload struct {
	ID               string  `json:"id"`
	MovementType     string  `json:"movement_type"`
	ProductID        int32   `json:"product_id"`
	ProductVariantID *int32  `json:"product_variant_id"`
	FromStoreID      *int32  `json:"from_store_id"`
	FromLocationID   *int32  `json:"from_location_id"`
	ToLocationID     *int32  `json:"to_location_id"`
	Quantity         float64 `json:"quantity"`
}

func applyRecipeStockMovementSync(ctx context.Context, db repository.DBTX, sm stockMovementSyncPayload) error {
	if sm.FromStoreID == nil || *sm.FromStoreID <= 0 || sm.ProductID <= 0 || sm.Quantity <= 0 {
		return nil
	}
	var locID *int32
	if sm.MovementType == "recipe_deduction" {
		locID = sm.FromLocationID
		tag, err := db.Exec(ctx, `
			UPDATE inventory_stock
			SET
				quantity_on_hand = quantity_on_hand - $1::numeric,
				quantity_available = GREATEST(0, quantity_available - $1::numeric),
				updated_at = CURRENT_TIMESTAMP
			WHERE id = (
				SELECT ist.id
				FROM inventory_stock ist
				WHERE ist.store_id = $2
				  AND ist.product_id = $3
				  AND ist.product_variant_id IS NOT DISTINCT FROM $4::int
				ORDER BY (ist.storage_location_id IS NOT DISTINCT FROM $5::int) DESC, ist.storage_location_id NULLS FIRST, ist.id ASC
				LIMIT 1
			)
		`, sm.Quantity, *sm.FromStoreID, sm.ProductID, sm.ProductVariantID, locID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("no inventory_stock row found for store %d product %d", *sm.FromStoreID, sm.ProductID)
		}
		return nil
	}
	if sm.MovementType == "recipe_reversal" {
		locID = sm.ToLocationID
		if locID == nil {
			locID = sm.FromLocationID
		}
		tag, err := db.Exec(ctx, `
			UPDATE inventory_stock
			SET
				quantity_on_hand = quantity_on_hand + $1::numeric,
				quantity_available = GREATEST(0, (quantity_on_hand + $1::numeric) - COALESCE(quantity_allocated, 0)),
				updated_at = CURRENT_TIMESTAMP
			WHERE id = (
				SELECT ist.id
				FROM inventory_stock ist
				WHERE ist.store_id = $2
				  AND ist.product_id = $3
				  AND ist.product_variant_id IS NOT DISTINCT FROM $4::int
				ORDER BY (ist.storage_location_id IS NOT DISTINCT FROM $5::int) DESC, ist.storage_location_id NULLS FIRST, ist.id ASC
				LIMIT 1
			)
		`, sm.Quantity, *sm.FromStoreID, sm.ProductID, sm.ProductVariantID, locID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("no inventory_stock row found for store %d product %d", *sm.FromStoreID, sm.ProductID)
		}
		return nil
	}
	return nil
}
