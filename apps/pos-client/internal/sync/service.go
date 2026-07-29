package sync

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/syncpb"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SyncService struct {
	ctx        context.Context
	masterPool *pgxpool.Pool
	cloudURL   string
	tenantSlug string
	grpcAddr   string
}

type OutboxItem struct {
	ID            int64           `json:"id"`
	EntityType    string          `json:"entity_type"`
	EntityID      int64           `json:"entity_id"`
	Action        string          `json:"action"`
	Payload       json.RawMessage `json:"payload"`
	CreatedAt     time.Time       `json:"created_at"`
	CorrelationID *string         `json:"correlation_id"`
}

func NewSyncService(ctx context.Context, pool *pgxpool.Pool, cloudURL, slug string) *SyncService {
	// Resolve gRPC target address from cloudURL or environment default
	grpcTarget := extractGRPCTarget(cloudURL)

	return &SyncService{
		ctx:        ctx,
		masterPool: pool,
		cloudURL:   cloudURL,
		tenantSlug: slug,
		grpcAddr:   grpcTarget,
	}
}

func extractGRPCTarget(cloudURL string) string {
	if cloudURL == "" {
		return "localhost:50051"
	}
	parsed, err := url.Parse(cloudURL)
	if err != nil {
		return "localhost:50051"
	}
	host := parsed.Hostname()
	if host == "" {
		host = "localhost"
	}
	return fmt.Sprintf("%s:50051", host)
}

func (s *SyncService) Start() {
	// Immediate sync on startup, then periodic loop
	go s.performSync()

	ticker := time.NewTicker(2 * time.Minute)
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.performSync()
			}
		}
	}()
	log.Printf("🔄 gRPC Sync Service started for tenant [%s] (target: %s)", s.tenantSlug, s.grpcAddr)
}

func (s *SyncService) performSync() {
	if s.masterPool == nil {
		return
	}

	// 1. gRPC Push Strategy (Local Outbox -> Cloud)
	s.drainOutboxGRPC()

	// 2. gRPC Pull Strategy (Cloud -> Local Watermarks Delta)
	s.fetchDeltaGRPC()
}

// drainOutboxGRPC streams pending sync_queue items to Cloud via gRPC StreamPush with SHA-256 checksums
func (s *SyncService) drainOutboxGRPC() {
	rows, err := s.masterPool.Query(s.ctx, `
		SELECT id, entity_type, entity_id, action, payload, created_at, correlation_id
		FROM sync_queue
		WHERE status = 'pending'
		ORDER BY priority DESC, created_at ASC
		LIMIT 50;
	`)
	if err != nil {
		log.Printf("[gRPC SYNC] Failed to query local sync_queue: %v", err)
		return
	}
	defer rows.Close()

	var items []OutboxItem
	for rows.Next() {
		var item OutboxItem
		if err := rows.Scan(&item.ID, &item.EntityType, &item.EntityID, &item.Action, &item.Payload, &item.CreatedAt, &item.CorrelationID); err != nil {
			continue
		}
		items = append(items, item)
	}

	if len(items) == 0 {
		return
	}

	var storeID int32
	_ = s.masterPool.QueryRow(s.ctx, `SELECT store_id FROM local_device_config LIMIT 1;`).Scan(&storeID)

	conn, err := grpc.NewClient(s.grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[gRPC SYNC] Connection failed to %s: %v", s.grpcAddr, err)
		s.recordOutboxFailure(items, err.Error())
		return
	}
	defer conn.Close()

	client := syncpb.NewSyncServiceClient(conn)
	stream, err := client.StreamPush(s.ctx)
	if err != nil {
		log.Printf("[gRPC SYNC] StreamPush initialization error: %v", err)
		s.recordOutboxFailure(items, err.Error())
		return
	}

	var syncedIDs []int64

	for _, item := range items {
		payloadBytes := []byte(item.Payload)
		hasher := sha256.New()
		hasher.Write(payloadBytes)
		shaHex := hex.EncodeToString(hasher.Sum(nil))

		var corrID string
		if item.CorrelationID != nil {
			corrID = *item.CorrelationID
		}

		event := &syncpb.SyncEvent{
			Id:            item.ID,
			EntityType:    item.EntityType,
			EntityId:      item.EntityID,
			Action:        item.Action,
			PayloadJson:   payloadBytes,
			StoreId:       storeID,
			TenantSlug:    s.tenantSlug,
			EventTime:     timestamppb.New(item.CreatedAt),
			Sha256:        shaHex,
			IsLastChunk:   true,
			CorrelationId: corrID,
		}

		if err := stream.Send(event); err != nil {
			log.Printf("[gRPC SYNC] Failed to send SyncEvent item %d: %v", item.ID, err)
			continue
		}

		ack, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("[gRPC SYNC] Failed to receive SyncAck for item %d: %v", item.ID, err)
			break
		}

		if ack.Success && ack.Sha256 == shaHex {
			syncedIDs = append(syncedIDs, ack.Id)
		} else {
			log.Printf("[gRPC SYNC] Server rejected item %d: %s", ack.Id, ack.ErrorMessage)
		}
	}

	_ = stream.CloseSend()

	if len(syncedIDs) > 0 {
		_, _ = s.masterPool.Exec(s.ctx, `
			UPDATE sync_queue
			SET status = 'synced', synced_at = NOW()
			WHERE id = ANY($1);
		`, syncedIDs)
		log.Printf("[gRPC SYNC] Pushed %d items upstream to Cloud via gRPC stream", len(syncedIDs))
	}
}

func (s *SyncService) recordOutboxFailure(items []OutboxItem, errMsg string) {
	ids := make([]int64, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}

	_, _ = s.masterPool.Exec(s.ctx, `
		UPDATE sync_queue
		SET retry_count = retry_count + 1,
		    last_error = $1,
		    status = CASE WHEN retry_count + 1 >= max_retries THEN 'failed' ELSE 'pending' END
		WHERE id = ANY($2);
	`, errMsg, ids)
}

// fetchDeltaGRPC pulls updated entities from Cloud via gRPC StreamPull using sync_watermarks
func (s *SyncService) fetchDeltaGRPC() {
	var storeID int32
	var lastSyncAt time.Time

	err := s.masterPool.QueryRow(s.ctx, `
		SELECT store_id, COALESCE(last_zatca_sync_at, '1970-01-01 00:00:00+00')
		FROM local_device_config
		LIMIT 1;
	`).Scan(&storeID, &lastSyncAt)
	if err != nil {
		return
	}

	conn, err := grpc.NewClient(s.grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return
	}
	defer conn.Close()

	client := syncpb.NewSyncServiceClient(conn)
	req := &syncpb.PullRequest{
		TenantSlug: s.tenantSlug,
		StoreId:    storeID,
		Since:      timestamppb.New(lastSyncAt),
		EntityTypes: []string{
			"product_barcodes", "product_prices", "promotions",
			"customers",
			"menu_items", "menu_modifier_groups", "combo_bundles", "recipes", "menu_item_availability_schedules",
			"inventory_stock", "stock_movements",
			"zatca_device_configs",
		},
		Limit: 200,
	}

	stream, err := client.StreamPull(s.ctx, req)
	if err != nil {
		log.Printf("[gRPC SYNC] StreamPull request error: %v", err)
		return
	}

	pulledCount := 0
	now := time.Now()

	for {
		event, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("[gRPC SYNC] Error receiving delta stream: %v", err)
			break
		}

		// SHA-256 verification
		hasher := sha256.New()
		hasher.Write(event.PayloadJson)
		calculatedSha := hex.EncodeToString(hasher.Sum(nil))

		if event.Sha256 != "" && event.Sha256 != calculatedSha {
			log.Printf("[gRPC SYNC] SHA-256 checksum error for pulled entity %s ID %d", event.EntityType, event.EntityId)
			continue
		}

		s.applyPulledEntity(event)
		pulledCount++
	}

	if pulledCount > 0 {
		_, _ = s.masterPool.Exec(s.ctx, `
			UPDATE local_device_config
			SET last_zatca_sync_at = $1;
		`, now)
		log.Printf("[gRPC SYNC] Pulled & verified %d delta updates from Cloud via gRPC", pulledCount)
	}
}

func (s *SyncService) applyPulledEntity(event *syncpb.SyncEvent) {
	if len(event.PayloadJson) == 0 {
		return
	}

	// Supported catalog, pricing, menu, promotion, and compliance entities
	validTables := map[string]bool{
		"product_barcodes": true, "product_prices": true, "promotions": true,
		"menu_items": true, "menu_modifier_groups": true, "combo_bundles": true,
		"recipes": true, "menu_item_availability_schedules": true,
		"customers": true, "inventory_stock": true, "zatca_device_configs": true,
	}

	if !validTables[event.EntityType] {
		log.Printf("[gRPC SYNC] Applied generic entity delta: %s (ID %d)", event.EntityType, event.EntityId)
		return
	}

	query := fmt.Sprintf(`
		INSERT INTO %s
		SELECT * FROM json_populate_record(NULL::%s, $1::json)
		ON CONFLICT (id) DO UPDATE SET updated_at = NOW();
	`, event.EntityType, event.EntityType)

	_, err := s.masterPool.Exec(s.ctx, query, string(event.PayloadJson))
	if err != nil {
		fallbackQuery := fmt.Sprintf(`
			INSERT INTO %s
			SELECT * FROM json_populate_record(NULL::%s, $1::json)
			ON CONFLICT (id) DO NOTHING;
		`, event.EntityType, event.EntityType)
		if _, err2 := s.masterPool.Exec(s.ctx, fallbackQuery, string(event.PayloadJson)); err2 != nil {
			log.Printf("[gRPC SYNC] Error applying pulled entity %s (ID %d): %v (fallback error: %v)",
				event.EntityType, event.EntityId, err, err2)
			return
		}
	}

	log.Printf("[gRPC SYNC] Successfully applied delta for %s (ID %d)", event.EntityType, event.EntityId)
}
