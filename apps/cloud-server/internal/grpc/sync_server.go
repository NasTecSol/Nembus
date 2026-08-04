package grpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/syncpb"
	"github.com/NasTecSol/nembus-core/middleware/manager"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
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
		"sales_orders_v2", "draft_cart_templates", "restaurant_orders", "restaurant_order_items",
		"kiosk_sessions", "stock_counts", "stock_count_lines", "waste_logs":
		if err := s.upsertEntityJSON(ctx, pool, event.EntityType, event.PayloadJson); err != nil {
			log.Printf("[gRPC SyncServer] Ingestion error for %s (ID %d): %v", event.EntityType, event.EntityId, err)
			return err
		}
		log.Printf("[gRPC SyncServer] Successfully ingested %s entity ID %d (Correlation: %s) for store %d",
			event.EntityType, event.EntityId, event.CorrelationId, event.StoreId)
	default:
		log.Printf("[gRPC SyncServer] Processed generic entity %s ID %d", event.EntityType, event.EntityId)
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

// upsertEntityJSON executes PostgreSQL json_populate_record upsert inside a database transaction
func (s *SyncServer) upsertEntityJSON(ctx context.Context, pool *pgxpool.Pool, entityType string, payload []byte) error {
	if len(payload) == 0 {
		return nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Disable foreign key constraint checks during sync ingestion so entities can be inserted out of order
	_, _ = tx.Exec(ctx, "SET LOCAL session_replication_role = 'replica';")

	validTables := map[string]bool{
		"pos_transactions": true, "pos_transaction_lines": true, "pos_payments": true,
		"cashier_sessions": true, "sales_orders_v2": true, "sales_order_lines_v2": true,
		"draft_cart_templates": true, "restaurant_orders": true, "restaurant_order_items": true,
		"kiosk_sessions": true, "stock_counts": true, "stock_count_lines": true, "waste_logs": true,
	}

	if !validTables[entityType] {
		return fmt.Errorf("unsupported entity type for upsert: %s", entityType)
	}

	_, _ = tx.Exec(ctx, "SAVEPOINT sp_upsert;")

	query := fmt.Sprintf(`
		INSERT INTO %s
		SELECT * FROM json_populate_record(NULL::%s, $1::json)
		ON CONFLICT (id) DO UPDATE SET updated_at = NOW();
	`, entityType, entityType)

	_, err = tx.Exec(ctx, query, string(payload))
	if err != nil {
		_, _ = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT sp_upsert;")
		fallbackQuery := fmt.Sprintf(`
			INSERT INTO %s
			SELECT * FROM json_populate_record(NULL::%s, $1::json)
			ON CONFLICT (id) DO NOTHING;
		`, entityType, entityType)
		if _, err2 := tx.Exec(ctx, fallbackQuery, string(payload)); err2 != nil {
			return fmt.Errorf("failed to upsert %s: %w (fallback: %v)", entityType, err, err2)
		}
	} else {
		_, _ = tx.Exec(ctx, "RELEASE SAVEPOINT sp_upsert;")
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

	// Delta entity target categories
	targetEntities := req.EntityTypes
	if len(targetEntities) == 0 {
		targetEntities = []string{
			"product_barcodes", "product_prices", "promotions",
			"customers",
			"menu_items", "menu_modifier_groups", "combo_bundles", "recipes", "menu_item_availability_schedules",
			"inventory_stock", "stock_movements",
			"zatca_device_configs",
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
			continue
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
	query := fmt.Sprintf(`SELECT id, row_to_json(t)::text, updated_at FROM %s t WHERE updated_at > $1 ORDER BY updated_at ASC LIMIT $2`, entityType)

	rows, err := pool.Query(ctx, query, since, limit)
	if err != nil {
		// Table might not exist in target schema, log warning and skip
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var jsonStr string
		var updatedAt time.Time

		if err := rows.Scan(&id, &jsonStr, &updatedAt); err != nil {
			continue
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

	return nil
}
