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
	entity_id   BIGINT NOT NULL,
	applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
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
	var authChecked bool

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
			if err := s.verifyStreamAuth(ctx, event.TenantSlug, event.DeviceId, event.StoreId); err != nil {
				return err
			}
			authChecked = true
		}

		// Legacy single-event mode when batch_id is empty or batch_size <= 0
		if strings.TrimSpace(event.BatchId) == "" || event.BatchSize <= 0 {
			ack := s.handleSinglePushEvent(ctx, event)
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

			ack := s.processPushBatch(ctx, batchToApply)
			if err := stream.Send(ack); err != nil {
				log.Printf("[gRPC SyncServer] Error sending batch SyncAck: %v", err)
				return err
			}
		}
	}
}

func (s *SyncServer) handleSinglePushEvent(ctx context.Context, event *syncpb.SyncEvent) *syncpb.SyncAck {
	calculatedSha := computePayloadSHA256(event.PayloadJson)
	if event.Sha256 != "" && event.Sha256 != calculatedSha {
		log.Printf("[gRPC SyncServer] SHA-256 mismatch for item %d (entity: %s): expected %s, got %s",
			event.Id, event.EntityType, event.Sha256, calculatedSha)
		return &syncpb.SyncAck{
			Id:           event.Id,
			OutboxUuid:   event.OutboxUuid,
			EntityType:   event.EntityType,
			EntityId:     event.EntityId,
			Success:      false,
			ErrorMessage: "SHA-256 checksum mismatch",
			Sha256:       calculatedSha,
			ProcessedAt:  timestamppb.Now(),
		}
	}

	if err := s.ingestSyncEvent(ctx, event); err != nil {
		log.Printf("[gRPC SyncServer] Failed to ingest item %d (%s): %v", event.Id, event.EntityType, err)
		return &syncpb.SyncAck{
			Id:           event.Id,
			OutboxUuid:   event.OutboxUuid,
			EntityType:   event.EntityType,
			EntityId:     event.EntityId,
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
		Success:     true,
		Sha256:      calculatedSha,
		ProcessedAt: timestamppb.Now(),
	}
}

// processPushBatch validates and applies an entire correlation batch inside one atomic PostgreSQL transaction
// with idempotent sync_inbox deduplication keyed by outbox_uuid.
func (s *SyncServer) processPushBatch(ctx context.Context, events []*syncpb.SyncEvent) *syncpb.SyncAck {
	if len(events) == 0 {
		return &syncpb.SyncAck{
			Success:      false,
			ErrorMessage: "empty correlation batch",
			ProcessedAt:  timestamppb.Now(),
		}
	}
	first := events[0]
	last := events[len(events)-1]
	batchID := first.BatchId
	lastSha := computePayloadSHA256(last.PayloadJson)

	// 1. Validate batch sequencing and SHA-256 checksums upfront before opening DB transaction
	for idx, ev := range events {
		expectedSeq := int32(idx + 1)
		if ev.SeqInBatch != 0 && ev.SeqInBatch != expectedSeq {
			return &syncpb.SyncAck{
				Id:           ev.Id,
				BatchId:      batchID,
				OutboxUuid:   ev.OutboxUuid,
				EntityType:   ev.EntityType,
				EntityId:     ev.EntityId,
				Success:      false,
				FailedSeq:    expectedSeq,
				ErrorMessage: fmt.Sprintf("invalid seq_in_batch: expected %d, got %d", expectedSeq, ev.SeqInBatch),
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
					Success:      false,
					FailedSeq:    seqInBatch,
					ErrorMessage: fmt.Sprintf("invalid outbox_uuid %q: %v", outboxUUID, parseErr),
					ProcessedAt:  timestamppb.Now(),
				}
			}
			deviceID := ev.DeviceId
			if deviceID == "" {
				deviceID = fmt.Sprintf("store-%d", ev.StoreId)
			}
			var insertedUUID string
			inboxErr := tx.QueryRow(ctx, `
				INSERT INTO sync_inbox (outbox_uuid, device_id, store_id, batch_id, entity_type, entity_id)
				VALUES ($1::uuid, $2, $3, $4, $5, $6)
				ON CONFLICT (outbox_uuid) DO NOTHING
				RETURNING outbox_uuid::text
			`, outboxUUID, deviceID, ev.StoreId, batchID, ev.EntityType, ev.EntityId).Scan(&insertedUUID)
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
				Success:      false,
				FailedSeq:    seqInBatch,
				ErrorMessage: fmt.Sprintf("unsupported sync entity type: %s", ev.EntityType),
				ProcessedAt:  timestamppb.Now(),
			}
		}

		if err := s.upsertEntityJSONWithDB(ctx, tx, pool, ev.EntityType, ev.Action, ev.PayloadJson, false); err != nil {
			log.Printf("[gRPC SyncServer] Batch %s failed at seq %d (%s ID %d): %v",
				batchID, seqInBatch, ev.EntityType, ev.EntityId, err)
			return &syncpb.SyncAck{
				Id:           ev.Id,
				BatchId:      batchID,
				OutboxUuid:   ev.OutboxUuid,
				EntityType:   ev.EntityType,
				EntityId:     ev.EntityId,
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

	return &syncpb.SyncAck{
		Id:          last.Id,
		BatchId:     batchID,
		OutboxUuid:  last.OutboxUuid,
		EntityType:  last.EntityType,
		EntityId:    last.EntityId,
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
	if event.TenantSlug == "" {
		return fmt.Errorf("tenant_slug is required")
	}

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
				deviceID = fmt.Sprintf("store-%d", event.StoreId)
			}
			var insertedUUID string
			inboxErr := tx.QueryRow(ctx, `
				INSERT INTO sync_inbox (outbox_uuid, device_id, store_id, batch_id, entity_type, entity_id)
				VALUES ($1::uuid, $2, $3, NULLIF($4, ''), $5, $6)
				ON CONFLICT (outbox_uuid) DO NOTHING
				RETURNING outbox_uuid::text
			`, outboxUUID, deviceID, event.StoreId, event.BatchId, event.EntityType, event.EntityId).Scan(&insertedUUID)
			if errors.Is(inboxErr, pgx.ErrNoRows) {
				return nil
			}
			if inboxErr != nil {
				return inboxErr
			}
			if err := s.upsertEntityJSONWithDB(ctx, tx, pool, event.EntityType, event.Action, event.PayloadJson, false); err != nil {
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

	if err := s.upsertEntityJSON(ctx, pool, event.EntityType, event.Action, event.PayloadJson); err != nil {
		log.Printf("[gRPC SyncServer] Ingestion error for %s (action=%s, ID %d): %v", event.EntityType, event.Action, event.EntityId, err)
		return err
	}
	log.Printf("[gRPC SyncServer] Successfully ingested %s entity ID %d (action=%s, Correlation: %s) for store %d",
		event.EntityType, event.EntityId, event.Action, event.CorrelationId, event.StoreId)

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
	query := fmt.Sprintf(`
		INSERT INTO %s
		SELECT * FROM json_populate_record(NULL::%s, $1::json)
		ON CONFLICT (id) DO UPDATE SET %s;
	`, entityType, entityType, updateClause)
	if _, err := db.Exec(ctx, query, string(payload)); err != nil {
		// Return the original constraint and SQLSTATE. A DO NOTHING fallback
		// can hide rejected updates, or replace the cause with SQLSTATE 25P02.
		return fmt.Errorf("failed to upsert %s (action=%s): %w", entityType, action, err)
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

// A customer ID is referenced by carts, orders, and transactions. A cloned POS
// can generate the same serial ID as a cloud write made after cloning, so never
// overwrite a different cloud customer merely because the numeric ID matches.
func upsertCreatedCustomer(ctx context.Context, pool *pgxpool.Pool, payload []byte) error {
	return upsertCreatedCustomerDB(ctx, pool, payload)
}

func upsertCreatedCustomerDB(ctx context.Context, db repository.DBTX, payload []byte) error {
	var incoming struct {
		ID             int32  `json:"id"`
		OrganizationID int32  `json:"organization_id"`
		CustomerCode   string `json:"customer_code"`
	}
	if err := json.Unmarshal(payload, &incoming); err != nil {
		return fmt.Errorf("invalid customer sync payload: %w", err)
	}
	if incoming.ID <= 0 || incoming.OrganizationID <= 0 || incoming.CustomerCode == "" {
		return fmt.Errorf("customer sync requires id, organization_id, and customer_code")
	}
	var insertedID int32
	err := db.QueryRow(ctx, `INSERT INTO customers
		SELECT * FROM json_populate_record(NULL::customers, $1::json)
		ON CONFLICT (id) DO NOTHING RETURNING id`, string(payload)).Scan(&insertedID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to insert customer %d: %w", incoming.ID, err)
	}
	var organizationID int32
	var customerCode string
	if err := db.QueryRow(ctx, `SELECT organization_id, customer_code FROM customers WHERE id=$1`, incoming.ID).
		Scan(&organizationID, &customerCode); err != nil {
		return fmt.Errorf("read existing customer %d: %w", incoming.ID, err)
	}
	if organizationID != incoming.OrganizationID || customerCode != incoming.CustomerCode {
		return fmt.Errorf("customer ID %d already belongs to another cloud customer; resolve the ID collision before syncing dependent sales", incoming.ID)
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

	ctx := stream.Context()
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
	if md, ok := metadata.FromIncomingContext(ctx); ok && len(md.Get(syncpb.PullCursorKey)) > 0 {
		values := md.Get(syncpb.PullCursorKey)
		afterID, err := strconv.ParseInt(values[0], 10, 64)
		if err != nil || afterID < 0 || len(req.EntityTypes) != 1 {
			return status.Error(codes.InvalidArgument, "pull cursor requires a nonnegative ID and exactly one entity")
		}
		if err := stream.SendHeader(metadata.Pairs(syncpb.PullCursorVersionKey, "1")); err != nil {
			return err
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

		if err := s.streamEntityDeltas(ctx, pool, req, entityType, sinceTime, limit, stream); err != nil {
			log.Printf("[gRPC SyncServer] Error streaming deltas for %s: %v", entityType, err)
			return status.Errorf(codes.Internal, "stream %s: %v", entityType, err)
		}
	}

	return nil
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
	query := fmt.Sprintf(`SELECT id, row_to_json(t)::text, updated_at FROM %s t WHERE updated_at > $1 ORDER BY updated_at ASC FETCH FIRST ($2) ROWS WITH TIES`, pgx.Identifier{entityType}.Sanitize())
	args := []any{since, limit}
	if md, ok := metadata.FromIncomingContext(ctx); ok && len(md.Get(syncpb.PullCursorKey)) > 0 {
		afterID, err := strconv.ParseInt(md.Get(syncpb.PullCursorKey)[0], 10, 64)
		if err != nil || afterID < 0 {
			return status.Error(codes.InvalidArgument, "invalid pull cursor")
		}
		query = fmt.Sprintf(`SELECT id, row_to_json(t)::text, updated_at FROM %s t
			WHERE updated_at > $1 OR (updated_at = $1 AND id > $3)
			ORDER BY updated_at, id LIMIT $2`, pgx.Identifier{entityType}.Sanitize())
		args = append(args, afterID)
	}

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query %s: %w", entityType, err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var jsonStr string
		var updatedAt time.Time

		if err := rows.Scan(&id, &jsonStr, &updatedAt); err != nil {
			return fmt.Errorf("scan %s: %w", entityType, err)
		}

		payloadBytes := []byte(jsonStr)
		shaHex := computePayloadSHA256(payloadBytes)

		event := &syncpb.SyncEvent{
			Id:          id,
			EntityType:  entityType,
			EntityId:    id,
			Action:      "UPDATE",
			PayloadJson: payloadBytes,
			StoreId:     req.StoreId,
			TenantSlug:  req.TenantSlug,
			EventTime:   timestamppb.New(updatedAt),
			Sha256:      shaHex,
			IsLastChunk: true,
		}

		if err := stream.Send(event); err != nil {
			return err
		}
	}

	return rows.Err()
}
