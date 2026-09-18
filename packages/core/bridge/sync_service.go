package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/syncpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OutboxItem struct {
	ID            int64           `json:"id"`
	EntityType    string          `json:"entity_type"`
	EntityID      string          `json:"entity_id"`
	Action        string          `json:"action"`
	Payload       json.RawMessage `json:"payload"`
	CreatedAt     time.Time       `json:"created_at"`
	CorrelationID *string         `json:"correlation_id"`
}

type SyncService struct {
	ctx        context.Context
	db         *sql.DB
	cloudURL   string
	tenantSlug string
	grpcAddr   string
	storeID    int32
}

func NewSyncService(ctx context.Context, db *sql.DB, cloudURL, slug string, storeID int32) *SyncService {
	grpcTarget := extractSyncGRPCTarget(cloudURL)
	if storeID <= 0 {
		storeID = 1
	}

	return &SyncService{
		ctx:        ctx,
		db:         db,
		cloudURL:   cloudURL,
		tenantSlug: slug,
		grpcAddr:   grpcTarget,
		storeID:    storeID,
	}
}

func extractSyncGRPCTarget(cloudURL string) string {
	if cloudURL == "" {
		return "nembus.nashrms.com:50051"
	}
	if strings.Contains(cloudURL, ":") && !strings.HasPrefix(cloudURL, "http") {
		return cloudURL
	}
	parsed, err := url.Parse(cloudURL)
	if err != nil {
		return "nembus.nashrms.com:50051"
	}
	host := parsed.Hostname()
	if host == "" {
		host = "nembus.nashrms.com"
	}
	port := parsed.Port()
	if port == "" {
		port = "50051"
	}
	return fmt.Sprintf("%s:%s", host, port)
}

// PerformFullSync runs outbox push to cloud and delta pull from cloud
func (s *SyncService) PerformFullSync() map[string]interface{} {
	if s.db == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Database connection is not initialized",
		}
	}

	// 1. Drain local outbox queue to cloud via gRPC StreamPush
	pushResult := s.DrainOutboxGRPC()

	// 2. Fetch incremental delta updates from cloud via gRPC StreamPull
	pullResult := s.FetchDeltaGRPC()

	// 3. Get current outbox statistics
	statusResult := s.GetSyncQueueStatus()

	return map[string]interface{}{
		"success":       pushResult["success"] == true || pullResult["success"] == true,
		"pushed_count":  pushResult["pushed_count"],
		"push_errors":   pushResult["errors"],
		"pulled_count":  pullResult["pulled_count"],
		"pull_errors":   pullResult["errors"],
		"queue_status":  statusResult,
		"synced_at":     time.Now().Format(time.RFC3339),
	}
}

// DrainOutboxGRPC streams pending sync_queue items to Cloud via gRPC StreamPush with SHA-256 checksums
func (s *SyncService) DrainOutboxGRPC() map[string]interface{} {
	// 1. Auto-enqueue any unqueued cashiers (Priority 30)
	_, _ = s.db.ExecContext(s.ctx, `
		INSERT INTO sync_queue (entity_type, entity_id, action, payload, status, priority)
		SELECT 'cashiers', CAST(c.id AS TEXT), 'INSERT', json_object(
			'id', c.id,
			'cashier_code', COALESCE(c.cashier_code, 'CASH-' || c.id),
			'status', COALESCE(c.status, 'active'),
			'organization_id', 1
		), 'pending', 30
		FROM cashiers c
		WHERE NOT EXISTS (
			SELECT 1 FROM sync_queue sq WHERE sq.entity_type = 'cashiers' AND sq.entity_id = CAST(c.id AS TEXT)
		);
	`)

	// 2. Auto-enqueue any unqueued cashier_sessions (Priority 25)
	_, _ = s.db.ExecContext(s.ctx, `
		INSERT INTO sync_queue (entity_type, entity_id, action, payload, status, priority)
		SELECT 'cashier_sessions', CAST(cs.id AS TEXT), 'INSERT', json_object(
			'id', cs.id,
			'cashier_id', COALESCE(cs.cashier_id, 1),
			'pos_terminal_id', COALESCE(cs.pos_terminal_id, 1),
			'session_number', COALESCE(cs.session_number, 'SESS-' || cs.id),
			'status', COALESCE(cs.status, 'open'),
			'opening_balance', COALESCE(cs.opening_balance, 0),
			'opened_at', COALESCE(cs.opened_at, CURRENT_TIMESTAMP)
		), 'pending', 25
		FROM cashier_sessions cs
		WHERE NOT EXISTS (
			SELECT 1 FROM sync_queue sq WHERE sq.entity_type = 'cashier_sessions' AND sq.entity_id = CAST(cs.id AS TEXT)
		);
	`)

	// 3. Auto-enqueue any unqueued customers so customer FKs are pushed first (Priority 20)
	_, _ = s.db.ExecContext(s.ctx, `
		INSERT INTO sync_queue (entity_type, entity_id, action, payload, status, priority)
		SELECT 'customers', CAST(c.id AS TEXT), 'INSERT', json_object(
			'id', c.id,
			'name', c.name,
			'customer_code', COALESCE(c.customer_code, 'CUST-' || c.id),
			'phone', COALESCE(c.phone, ''),
			'email', COALESCE(c.email, ''),
			'customer_type', COALESCE(c.customer_type, 'retail'),
			'organization_id', 1,
			'is_active', 1
		), 'pending', 20
		FROM customers c
		WHERE NOT EXISTS (
			SELECT 1 FROM sync_queue sq WHERE sq.entity_type = 'customers' AND sq.entity_id = CAST(c.id AS TEXT)
		);
	`)

	// Reset previously failed items so user-triggered sync retries all outbox items
	_, _ = s.db.ExecContext(s.ctx, `
		UPDATE sync_queue
		SET status = 'pending', retry_count = 0
		WHERE status = 'failed';
	`)

	rows, err := s.db.QueryContext(s.ctx, `
		SELECT id, entity_type, entity_id, action, payload, created_at, correlation_id
		FROM sync_queue
		WHERE status = 'pending'
		ORDER BY priority DESC, created_at ASC
		LIMIT 50;
	`)
	if err != nil {
		log.Printf("[gRPC OUTBOX SYNC] Failed to query local sync_queue: %v", err)
		return map[string]interface{}{
			"success":      false,
			"pushed_count": 0,
			"errors":       []string{fmt.Sprintf("Failed to query sync_queue: %v", err)},
		}
	}
	defer rows.Close()

	var items []OutboxItem
	for rows.Next() {
		var item OutboxItem
		var entityIDStr string
		var createdAtStr string
		var payloadStr string

		if err := rows.Scan(&item.ID, &item.EntityType, &entityIDStr, &item.Action, &payloadStr, &createdAtStr, &item.CorrelationID); err != nil {
			continue
		}
		item.EntityID = entityIDStr

		// Normalize PostgreSQL enum types and foreign keys for cloud compatibility
		if item.EntityType == "sales_orders_v2" || item.EntityType == "pos_transactions" {
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(payloadStr), &m); err == nil {
				changed := false
				if ot, ok := m["order_type"].(string); ok && ot == "pos_sale" {
					m["order_type"] = "standard"
					changed = true
				}
				if _, ok := m["source_cart_id"]; ok {
					delete(m, "source_cart_id")
					changed = true
				}
				if changed {
					if updatedPayload, err := json.Marshal(m); err == nil {
						payloadStr = string(updatedPayload)
						_, _ = s.db.ExecContext(s.ctx, `UPDATE sync_queue SET payload = ? WHERE id = ?`, payloadStr, item.ID)
					}
				}
			}
		}

		item.Payload = json.RawMessage(payloadStr)

		if parsedTime, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			item.CreatedAt = parsedTime
		} else if parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
			item.CreatedAt = parsedTime
		} else {
			item.CreatedAt = time.Now()
		}

		items = append(items, item)
	}

	if len(items) == 0 {
		return map[string]interface{}{
			"success":      true,
			"pushed_count": 0,
			"errors":       []string{},
		}
	}

	dialCtx, dialCancel := context.WithTimeout(s.ctx, 15*time.Second)
	defer dialCancel()

	conn, err := grpc.DialContext(dialCtx, s.grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Printf("[gRPC OUTBOX SYNC] Connection failed to %s: %v", s.grpcAddr, err)
		s.recordOutboxFailure(items, err.Error())
		return map[string]interface{}{
			"success":      false,
			"pushed_count": 0,
			"errors":       []string{fmt.Sprintf("gRPC connection failed to %s: %v", s.grpcAddr, err)},
		}
	}
	defer conn.Close()

	client := syncpb.NewSyncServiceClient(conn)
	stream, err := client.StreamPush(s.ctx)
	if err != nil {
		log.Printf("[gRPC OUTBOX SYNC] StreamPush initialization error: %v", err)
		s.recordOutboxFailure(items, err.Error())
		return map[string]interface{}{
			"success":      false,
			"pushed_count": 0,
			"errors":       []string{fmt.Sprintf("StreamPush init error: %v", err)},
		}
	}

	var syncedIDs []int64
	var pushErrors []string

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
			EntityId:      item.ID,
			Action:        item.Action,
			PayloadJson:   payloadBytes,
			StoreId:       s.storeID,
			TenantSlug:    s.tenantSlug,
			EventTime:     timestamppb.New(item.CreatedAt),
			Sha256:        shaHex,
			IsLastChunk:   true,
			CorrelationId: corrID,
		}

		if err := stream.Send(event); err != nil {
			errMsg := fmt.Sprintf("Send error for item %d (%s): %v", item.ID, item.EntityType, err)
			log.Printf("[gRPC OUTBOX SYNC] %s", errMsg)
			s.recordOutboxFailure([]OutboxItem{item}, errMsg)
			pushErrors = append(pushErrors, errMsg)
			continue
		}

		ack, err := stream.Recv()
		if err != nil {
			errMsg := fmt.Sprintf("Recv error for item %d (%s): %v", item.ID, item.EntityType, err)
			log.Printf("[gRPC OUTBOX SYNC] %s", errMsg)
			s.recordOutboxFailure([]OutboxItem{item}, errMsg)
			pushErrors = append(pushErrors, errMsg)
			break
		}

		if ack.Success && ack.Sha256 == shaHex {
			syncedIDs = append(syncedIDs, ack.Id)
			_, _ = s.db.ExecContext(s.ctx, `
				UPDATE sync_queue
				SET status = 'synced', synced_at = CURRENT_TIMESTAMP
				WHERE id = ?;
			`, ack.Id)
		} else {
			errMsg := fmt.Sprintf("Server rejected item %d: %s", ack.Id, ack.ErrorMessage)
			log.Printf("[gRPC OUTBOX SYNC] %s", errMsg)
			s.recordOutboxFailure([]OutboxItem{item}, errMsg)
			pushErrors = append(pushErrors, errMsg)
		}
	}

	_ = stream.CloseSend()

	log.Printf("[gRPC OUTBOX SYNC] Pushed %d items upstream to Cloud via gRPC stream", len(syncedIDs))

	return map[string]interface{}{
		"success":      len(syncedIDs) > 0 || len(pushErrors) == 0,
		"pushed_count": len(syncedIDs),
		"errors":       pushErrors,
	}
}

func (s *SyncService) recordOutboxFailure(items []OutboxItem, errMsg string) {
	for _, item := range items {
		_, _ = s.db.ExecContext(s.ctx, `
			UPDATE sync_queue
			SET retry_count = retry_count + 1,
			    last_error = ?,
			    status = CASE WHEN retry_count + 1 >= max_retries THEN 'failed' ELSE 'pending' END
			WHERE id = ?;
		`, errMsg, item.ID)
	}
}

// FetchDeltaGRPC pulls updated entities from Cloud via gRPC StreamPull using sync watermarks
func (s *SyncService) FetchDeltaGRPC() map[string]interface{} {
	var lastSyncAtStr sql.NullString
	_ = s.db.QueryRowContext(s.ctx, `
		SELECT last_zatca_sync_at FROM local_device_config LIMIT 1;
	`).Scan(&lastSyncAtStr)

	lastSyncAt := time.Unix(0, 0)
	if lastSyncAtStr.Valid && lastSyncAtStr.String != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", lastSyncAtStr.String); err == nil {
			lastSyncAt = t
		} else if t, err := time.Parse(time.RFC3339, lastSyncAtStr.String); err == nil {
			lastSyncAt = t
		}
	}

	dialCtx, dialCancel := context.WithTimeout(s.ctx, 15*time.Second)
	defer dialCancel()

	conn, err := grpc.DialContext(dialCtx, s.grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return map[string]interface{}{
			"success":      false,
			"pulled_count": 0,
			"errors":       []string{fmt.Sprintf("gRPC connection failed to %s: %v", s.grpcAddr, err)},
		}
	}
	defer conn.Close()

	client := syncpb.NewSyncServiceClient(conn)
	req := &syncpb.PullRequest{
		TenantSlug: s.tenantSlug,
		StoreId:    s.storeID,
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
		log.Printf("[gRPC DELTA SYNC] StreamPull request error: %v", err)
		return map[string]interface{}{
			"success":      false,
			"pulled_count": 0,
			"errors":       []string{fmt.Sprintf("StreamPull request error: %v", err)},
		}
	}

	pulledCount := 0
	var pullErrors []string

	for {
		event, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			pullErrors = append(pullErrors, fmt.Sprintf("Error receiving delta stream: %v", err))
			break
		}

		// Checksum verification
		hasher := sha256.New()
		hasher.Write(event.PayloadJson)
		calculatedSha := hex.EncodeToString(hasher.Sum(nil))

		if event.Sha256 != "" && event.Sha256 != calculatedSha {
			pullErrors = append(pullErrors, fmt.Sprintf("SHA-256 mismatch for %s ID %d", event.EntityType, event.EntityId))
			continue
		}

		s.applyPulledEntitySQLite(event)
		pulledCount++
	}

	if pulledCount > 0 {
		_, _ = s.db.ExecContext(s.ctx, `
			UPDATE local_device_config
			SET last_zatca_sync_at = CURRENT_TIMESTAMP
			WHERE id = 1;
		`)
		log.Printf("[gRPC DELTA SYNC] Pulled & verified %d delta updates from Cloud via gRPC", pulledCount)
	}

	return map[string]interface{}{
		"success":      pulledCount > 0 || len(pullErrors) == 0,
		"pulled_count": pulledCount,
		"errors":       pullErrors,
	}
}

func (s *SyncService) applyPulledEntitySQLite(event *syncpb.SyncEvent) {
	if len(event.PayloadJson) == 0 {
		return
	}

	var rowMap map[string]interface{}
	if err := json.Unmarshal(event.PayloadJson, &rowMap); err != nil {
		return
	}

	columns := make([]string, 0, len(rowMap))
	placeholders := make([]string, 0, len(rowMap))
	values := make([]interface{}, 0, len(rowMap))
	updateClauses := make([]string, 0, len(rowMap))

	for col, val := range rowMap {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		values = append(values, val)
		if col != "id" {
			updateClauses = append(updateClauses, fmt.Sprintf("%s = excluded.%s", col, col))
		}
	}

	if len(columns) == 0 {
		return
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (id) DO UPDATE SET %s;",
		event.EntityType,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(updateClauses, ", "),
	)

	_, _ = s.db.ExecContext(s.ctx, query, values...)
}

// GetSyncQueueStatus queries counts and details of local sync queue
func (s *SyncService) GetSyncQueueStatus() map[string]interface{} {
	var pendingCount, syncedCount, failedCount int
	var lastSyncedAt sql.NullString

	_ = s.db.QueryRowContext(s.ctx, `SELECT COUNT(*) FROM sync_queue WHERE status = 'pending';`).Scan(&pendingCount)
	_ = s.db.QueryRowContext(s.ctx, `SELECT COUNT(*) FROM sync_queue WHERE status = 'synced';`).Scan(&syncedCount)
	_ = s.db.QueryRowContext(s.ctx, `SELECT COUNT(*) FROM sync_queue WHERE status = 'failed';`).Scan(&failedCount)
	_ = s.db.QueryRowContext(s.ctx, `SELECT MAX(synced_at) FROM sync_queue WHERE status = 'synced';`).Scan(&lastSyncedAt)

	return map[string]interface{}{
		"pending_count":  pendingCount,
		"synced_count":   syncedCount,
		"failed_count":   failedCount,
		"last_synced_at": lastSyncedAt.String,
	}
}

// EnqueueOutboxRecord enqueues an offline action into sync_queue
func EnqueueOutboxRecord(db *sql.DB, entityType, entityID, action string, payload map[string]interface{}, priority int, correlationID string) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if priority <= 0 {
		priority = 5
		if entityType == "pos_transactions" || entityType == "pos_payments" || entityType == "sales_orders_v2" {
			priority = 10
		}
	}

	_, err = db.Exec(`
		INSERT INTO sync_queue (entity_type, entity_id, action, payload, status, priority, correlation_id, created_at)
		VALUES (?, ?, ?, ?, 'pending', ?, ?, CURRENT_TIMESTAMP);
	`, entityType, entityID, action, string(payloadBytes), priority, correlationID)

	return err
}
