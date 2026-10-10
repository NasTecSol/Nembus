package sync

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"log"
	"net/url"
	"strings"
	"sync"
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
	syncMu     sync.Mutex
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
		ctx:        context.Background(),
		masterPool: pool,
		cloudURL:   cloudURL,
		tenantSlug: slug,
		grpcAddr:   grpcTarget,
	}
}

func extractGRPCTarget(target string) string {
	if target == "" {
		return "localhost:50051"
	}
	// If already in host:port format without scheme (e.g. "nembus.nashrms.com:50051")
	if strings.Contains(target, ":") && !strings.Contains(target, "://") {
		return target
	}
	// If target has no scheme, prefix with http:// so url.Parse can extract hostname
	if !strings.Contains(target, "://") {
		target = "http://" + target
	}
	parsed, err := url.Parse(target)
	if err != nil {
		return "localhost:50051"
	}
	host := parsed.Hostname()
	if host == "" {
		host = "localhost"
	}
	port := parsed.Port()
	if port == "" {
		port = "50051"
	}
	return fmt.Sprintf("%s:%s", host, port)
}

func (s *SyncService) Start() {
	// Immediate sync on startup, then periodic loop
	go s.performSync()

	ticker := time.NewTicker(15 * time.Second)
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
	log.Printf("🔄 gRPC Sync Service started for tenant [%s] (target: %s, interval: 15s)", s.tenantSlug, s.grpcAddr)
}

// SyncNow triggers immediate outbox draining and delta fetch
func (s *SyncService) SyncNow() {
	go s.performSync()
}

func (s *SyncService) performSync() {
	// Startup, the timer and SyncNow can overlap. Only one worker may drain
	// the queue, otherwise the same failure consumes several retries at once.
	if !s.syncMu.TryLock() {
		return
	}
	defer s.syncMu.Unlock()
	if s.masterPool == nil {
		return
	}

	// 1. gRPC Push Strategy (Local Outbox -> Cloud)
	s.drainOutboxGRPC()

	// 2. gRPC Pull Strategy (Cloud -> Local Watermarks Delta)
	s.fetchDeltaGRPC()
}

// drainOutboxGRPC streams pending sync_queue items to Cloud via gRPC StreamPush with SHA-256 checksums
const pendingOutboxSQL = `
	SELECT q.id, q.entity_type, q.entity_id, q.action, q.payload, q.created_at, q.correlation_id
	FROM sync_queue q
	WHERE q.status = 'pending'
	  AND NOT EXISTS (
		SELECT 1 FROM sync_queue parent
		WHERE q.entity_type IN ('pos_transaction_lines', 'pos_payments')
		  AND parent.entity_type = 'pos_transactions'
		  AND parent.entity_id = (q.payload->>'transaction_id')::bigint
		  AND parent.status IN ('pending', 'failed')
	  )
	ORDER BY q.id ASC
	LIMIT 50;
`

func (s *SyncService) drainOutboxGRPC() {
	// Child events stay pending without consuming retries while their parent
	// is pending or failed. They become eligible after its success is saved.
	rows, err := s.masterPool.Query(s.ctx, pendingOutboxSQL)
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
	if err := rows.Err(); err != nil {
		log.Printf("[gRPC SYNC] Failed to read local sync_queue: %v", err)
		return
	}
	rows.Close()

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
		// Checkout creates a UUID sales order and cart locally, but the outbox
		// only contains the resulting POS rows. Send those referenced parents
		// first, including for older failed entries being retried.
		parents, err := s.transactionParents(item)
		if err != nil {
			if errors.Is(err, errCheckoutNotReady) {
				continue // Checkout is still completing locally; no retry is consumed.
			}
			s.recordOutboxFailure([]OutboxItem{item}, err.Error())
			continue
		}
		parentFailed := false
		for _, parent := range parents {
			if err := pushOutboxItem(stream, parent, storeID, s.tenantSlug); err != nil {
				s.recordOutboxFailure([]OutboxItem{item}, fmt.Sprintf("parent %s: %v", parent.EntityType, err))
				parentFailed = true
				break
			}
		}
		if parentFailed {
			continue
		}
		if err := pushOutboxItem(stream, item, storeID, s.tenantSlug); err != nil {
			log.Printf("[gRPC SYNC] Failed to push %s item %d: %v", item.EntityType, item.ID, err)
			s.recordOutboxFailure([]OutboxItem{item}, err.Error())
			continue
		}
		syncedIDs = append(syncedIDs, item.ID)
	}

	_ = stream.CloseSend()

	if len(syncedIDs) > 0 {
		_, _ = s.masterPool.Exec(s.ctx, `
			UPDATE sync_queue
			SET status = 'synced', synced_at = NOW(), last_error = NULL
			WHERE id = ANY($1);
		`, syncedIDs)
		log.Printf("[gRPC SYNC] Pushed %d items upstream to Cloud via gRPC stream", len(syncedIDs))
	}
}

// transactionParents returns the local cart and order referenced by a POS
// transaction in FK order. UUID identities stay in the JSON payload; EntityID
// is only a numeric diagnostic field in the current protocol.
var errCheckoutNotReady = errors.New("checkout is not fulfilled locally yet")

func (s *SyncService) transactionParents(item OutboxItem) ([]OutboxItem, error) {
	if item.EntityType != "pos_transactions" {
		return nil, nil
	}
	var refs struct {
		SalesOrderID *string `json:"sales_order_id"`
		SourceCartID *string `json:"source_cart_id"`
	}
	if err := json.Unmarshal(item.Payload, &refs); err != nil {
		return nil, fmt.Errorf("invalid transaction payload: %w", err)
	}
	if refs.SalesOrderID == nil && refs.SourceCartID == nil {
		return nil, nil
	}
	rows, err := s.masterPool.Query(s.ctx, `
		SELECT entity_type, payload FROM (
			SELECT 1 AS seq, 'carts' AS entity_type, row_to_json(c)::text AS payload
			FROM carts c
			WHERE c.id = $2::uuid OR c.id = (
				SELECT source_cart_id FROM sales_orders_v2 WHERE id = $1::uuid
			)
			UNION ALL
			SELECT 2 AS seq, 'sales_orders_v2', json_build_object(
				'order', row_to_json(o),
				'lines', COALESCE((SELECT json_agg(l ORDER BY l.line_number)
					FROM sales_order_lines_v2 l WHERE l.sales_order_id = o.id), '[]'::json)
			)::text
			FROM sales_orders_v2 o WHERE o.id = $1::uuid
		) parents ORDER BY seq;
	`, refs.SalesOrderID, refs.SourceCartID)
	if err != nil {
		return nil, fmt.Errorf("read transaction parents: %w", err)
	}
	defer rows.Close()
	var parents []OutboxItem
	for rows.Next() {
		parent := OutboxItem{ID: item.ID, Action: "INSERT", CreatedAt: item.CreatedAt, CorrelationID: item.CorrelationID}
		var payload string
		if err := rows.Scan(&parent.EntityType, &payload); err != nil {
			return nil, fmt.Errorf("read transaction parent: %w", err)
		}
		parent.Payload = json.RawMessage(payload)
		if parent.EntityType == "sales_orders_v2" {
			var checkout struct {
				Order struct {
					FulfillmentStatus string `json:"fulfillment_status"`
					OrderStatus       string `json:"order_status"`
				} `json:"order"`
			}
			if err := json.Unmarshal(parent.Payload, &checkout); err != nil {
				return nil, err
			}
			if checkout.Order.FulfillmentStatus != "fulfilled" || checkout.Order.OrderStatus != "fulfilled" {
				return nil, errCheckoutNotReady
			}
			parent.Action = "POS_CHECKOUT"
		}
		parents = append(parents, parent)
	}
	return parents, rows.Err()
}

func pushOutboxItem(stream syncpb.SyncService_StreamPushClient, item OutboxItem, storeID int32, tenantSlug string) error {
	hash := sha256.Sum256(item.Payload)
	checksum := hex.EncodeToString(hash[:])
	event := &syncpb.SyncEvent{
		Id: item.ID, EntityType: item.EntityType, EntityId: item.EntityID,
		Action: item.Action, PayloadJson: item.Payload, StoreId: storeID,
		TenantSlug: tenantSlug, EventTime: timestamppb.New(item.CreatedAt),
		Sha256: checksum, IsLastChunk: true,
	}
	if item.CorrelationID != nil {
		event.CorrelationId = *item.CorrelationID
	}
	if err := stream.Send(event); err != nil {
		return fmt.Errorf("send error: %w", err)
	}
	ack, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("receive error: %w", err)
	}
	if ack.Id != item.ID || ack.EntityType != item.EntityType || ack.Sha256 != checksum {
		return fmt.Errorf("acknowledgment does not match %s", item.EntityType)
	}
	if !ack.Success {
		return fmt.Errorf("server rejected: %s", ack.ErrorMessage)
	}
	return nil
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

// ResetFailedItems resets all 'failed' sync_queue entries back to 'pending' so they are
// retried on the next sync cycle. Call this after deploying a cloud-side fix that was
// causing legitimate INSERT events to be rejected (e.g. FK constraint violations).
func (s *SyncService) ResetFailedItems() (int64, error) {
	tag, err := s.masterPool.Exec(s.ctx, `
		UPDATE sync_queue
		SET status      = 'pending',
		    retry_count = 0,
		    last_error  = NULL
		WHERE status = 'failed';
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to reset failed sync_queue items: %w", err)
	}
	count := tag.RowsAffected()
	if count > 0 {
		log.Printf("[gRPC SYNC] Reset %d failed sync_queue items to pending for retry", count)
	}
	return count, nil
}
