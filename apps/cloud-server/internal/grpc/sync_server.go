package grpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/syncpb"
	"github.com/NasTecSol/nembus-core/middleware/manager"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SyncServer implements syncpb.SyncServiceServer for gRPC bidirectional push and delta pull streaming.
type SyncServer struct {
	syncpb.UnimplementedSyncServiceServer

	tenantManager *manager.Manager
	masterPool    *pgxpool.Pool
}

// NewSyncServer returns a new initialized SyncServer instance.
func NewSyncServer(tm *manager.Manager, masterPool *pgxpool.Pool) *SyncServer {
	return &SyncServer{
		tenantManager: tm,
		masterPool:    masterPool,
	}
}

// StreamPush handles real-time streaming of local outbox items from client to cloud with SHA-256 verification.
func (s *SyncServer) StreamPush(stream syncpb.SyncService_StreamPushServer) error {
	ctx := stream.Context()

	for {
		event, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Printf("[gRPC SyncServer] Error receiving push stream: %v", err)
			return err
		}

		// Calculate SHA-256 checksum over PayloadJson
		hasher := sha256.New()
		hasher.Write(event.PayloadJson)
		calculatedSha := hex.EncodeToString(hasher.Sum(nil))

		// Checksum verification
		if event.Sha256 != "" && event.Sha256 != calculatedSha {
			log.Printf("[gRPC SyncServer] SHA-256 mismatch for item %d (entity: %s): expected %s, got %s",
				event.Id, event.EntityType, event.Sha256, calculatedSha)
			_ = stream.Send(&syncpb.SyncAck{
				Id:           event.Id,
				EntityType:   event.EntityType,
				EntityId:     event.EntityId,
				Success:      false,
				ErrorMessage: "SHA-256 checksum mismatch",
				Sha256:       calculatedSha,
				ProcessedAt:  timestamppb.Now(),
			})
			continue
		}

		// Ingest entity into database
		err = s.ingestSyncEvent(ctx, event)
		if err != nil {
			log.Printf("[gRPC SyncServer] Failed to ingest item %d (%s): %v", event.Id, event.EntityType, err)
			_ = stream.Send(&syncpb.SyncAck{
				Id:           event.Id,
				EntityType:   event.EntityType,
				EntityId:     event.EntityId,
				Success:      false,
				ErrorMessage: err.Error(),
				Sha256:       calculatedSha,
				ProcessedAt:  timestamppb.Now(),
			})
			continue
		}

		// Send successful acknowledgment
		ack := &syncpb.SyncAck{
			Id:          event.Id,
			EntityType:  event.EntityType,
			EntityId:    event.EntityId,
			Success:     true,
			Sha256:      calculatedSha,
			ProcessedAt: timestamppb.Now(),
		}
		if err := stream.Send(ack); err != nil {
			log.Printf("[gRPC SyncServer] Error sending SyncAck: %v", err)
			return err
		}
	}
}

// ingestSyncEvent processes and persists incoming entity payloads across POS, Wholesale, Restaurant, and Inventory verticals.
func (s *SyncServer) ingestSyncEvent(ctx context.Context, event *syncpb.SyncEvent) error {
	if event.TenantSlug == "" {
		return fmt.Errorf("tenant_slug is required")
	}

	pool, err := s.tenantManager.GetPool(ctx, event.TenantSlug)
	if err != nil {
		return fmt.Errorf("failed to get tenant pool: %w", err)
	}

	// Dynamic ingestion routing by entity type
	switch event.EntityType {
	case "pos_transactions", "pos_transaction_lines", "pos_payments", "cashier_sessions",
		"carts", "sales_orders_v2", "sales_order_lines_v2", "draft_cart_templates", "restaurant_orders", "restaurant_order_items",
		"kiosk_sessions", "stock_counts", "stock_count_lines", "waste_logs":
		if err := s.upsertEntityJSON(ctx, pool, event.EntityType, event.Action, event.PayloadJson); err != nil {
			log.Printf("[gRPC SyncServer] Ingestion error for %s (action=%s, ID %d): %v", event.EntityType, event.Action, event.EntityId, err)
			return err
		}
		log.Printf("[gRPC SyncServer] Successfully ingested %s entity ID %d (action=%s, Correlation: %s) for store %d",
			event.EntityType, event.EntityId, event.Action, event.CorrelationId, event.StoreId)
	default:
		return fmt.Errorf("unsupported sync entity type: %s", event.EntityType)
	}

	// Insert into raw sync log / store outbox log for processing
	metadataJSON, _ := json.Marshal(map[string]interface{}{
		"last_event_id":  event.Id,
		"action":         event.Action,
		"correlation_id": event.CorrelationId,
	})

	_, err = pool.Exec(ctx, `
		INSERT INTO sync_watermarks (entity_type, store_id, last_sync_at, metadata)
		VALUES ($1, $2, NOW(), $3)
		ON CONFLICT (entity_type, store_id) DO UPDATE SET
			last_sync_at = EXCLUDED.last_sync_at,
			metadata     = EXCLUDED.metadata;
	`, event.EntityType, event.StoreId, string(metadataJSON))

	if err != nil {
		log.Printf("[gRPC SyncServer] Watermark update warning: %v", err)
	}

	return nil
}

func getTableUpdateClause(ctx context.Context, pool *pgxpool.Pool, table string) (string, error) {
	rows, err := pool.Query(ctx, `
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

// upsertEntityJSON preserves FK checks. The client sends checkout parents
// before POS rows; bypassing constraints would leave orphaned sales behind.
func (s *SyncServer) upsertEntityJSON(ctx context.Context, pool *pgxpool.Pool, entityType, action string, payload []byte) error {
	if entityType == "sales_orders_v2" && action == "POS_CHECKOUT" {
		return s.applyPOSCheckout(ctx, pool, payload)
	}
	if len(payload) == 0 {
		return fmt.Errorf("empty sync payload for %s", entityType)
	}
	validTables := map[string]bool{
		"pos_transactions": true, "pos_transaction_lines": true, "pos_payments": true,
		"cashier_sessions": true, "carts": true, "sales_orders_v2": true, "sales_order_lines_v2": true,
		"draft_cart_templates": true, "restaurant_orders": true, "restaurant_order_items": true,
		"kiosk_sessions": true, "stock_counts": true, "stock_count_lines": true, "waste_logs": true,
	}
	if !validTables[entityType] {
		return fmt.Errorf("unsupported entity type for upsert: %s", entityType)
	}
	updateClause, err := getTableUpdateClause(ctx, pool, entityType)
	if err != nil {
		return err
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
	if _, err := pool.Exec(ctx, query, string(payload)); err != nil {
		// Return the original constraint and SQLSTATE. A DO NOTHING fallback
		// can hide rejected updates, or replace the cause with SQLSTATE 25P02.
		return fmt.Errorf("failed to upsert %s (action=%s): %w", entityType, action, err)
	}
	return nil
}

// applyPOSCheckout saves the complete order before the fulfillment transition.
// Stock changes and fulfilled status commit together; retries never reset it.
func (s *SyncServer) applyPOSCheckout(ctx context.Context, pool *pgxpool.Pool, payload []byte) error {
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
	orderClause, err := getTableUpdateClause(ctx, pool, "sales_orders_v2")
	if err != nil {
		return err
	}
	lineClause, err := getTableUpdateClause(ctx, pool, "sales_order_lines_v2")
	if err != nil {
		return err
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
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
		return tx.Commit(ctx)
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
	return tx.Commit(ctx)
}

// StreamPull handles Cloud -> Local Terminal delta updates based on sync watermarks.
func (s *SyncServer) StreamPull(req *syncpb.PullRequest, stream syncpb.SyncService_StreamPullServer) error {
	if req.TenantSlug == "" {
		return status.Error(codes.InvalidArgument, "tenant_slug is required")
	}

	ctx := stream.Context()
	pool, err := s.tenantManager.GetPool(ctx, req.TenantSlug)
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
		hasher := sha256.New()
		hasher.Write(payloadBytes)
		shaHex := hex.EncodeToString(hasher.Sum(nil))

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
