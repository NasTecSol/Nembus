package sync

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/syncpb"
	corezatca "github.com/NasTecSol/nembus-core/zatca"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const OutboxResilienceDDL = `
ALTER TABLE IF EXISTS sync_queue
	ADD COLUMN IF NOT EXISTS error_class VARCHAR(32),
	ADD COLUMN IF NOT EXISTS next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	ADD COLUMN IF NOT EXISTS quarantined_at TIMESTAMPTZ,
	ADD COLUMN IF NOT EXISTS outbox_uuid UUID NOT NULL DEFAULT gen_random_uuid();
CREATE INDEX IF NOT EXISTS idx_sync_queue_pending_v2
	ON sync_queue(status, next_attempt_at ASC, priority DESC, id ASC)
	WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_sync_queue_correlation
	ON sync_queue(correlation_id, id ASC)
	WHERE status = 'pending';
ALTER TABLE IF EXISTS local_device_config
	ADD COLUMN IF NOT EXISTS sync_token TEXT;

CREATE OR REPLACE FUNCTION trg_enqueue_sync_outbox()
RETURNS TRIGGER AS $$
DECLARE
    entity_name VARCHAR(50);
    prio INTEGER := 5;
    corr_id VARCHAR(100);
    existing_id BIGINT;
BEGIN
    -- Prevent echo-loop re-queuing when rows are applied from Cloud StreamPull/StreamChanges
    IF current_setting('nembus.sync_applier', true) = 'true' THEN
        RETURN NEW;
    END IF;

    entity_name := TG_ARGV[0];
    IF entity_name = 'pos_transactions' OR entity_name = 'pos_payments' OR entity_name = 'sales_returns' THEN
        prio := 10;
    END IF;

    corr_id := COALESCE(NULLIF(current_setting('nembus.correlation_id', true), ''), pg_current_xact_id()::text);

    -- Coalesce same-(correlation_id, entity_type, entity_id) pending rows within the same transaction
    SELECT id INTO existing_id
    FROM sync_queue
    WHERE correlation_id = corr_id
      AND entity_type = entity_name
      AND entity_id = NEW.id
      AND status = 'pending'
    ORDER BY id DESC
    LIMIT 1;

    IF existing_id IS NOT NULL THEN
        UPDATE sync_queue
        SET payload = row_to_json(NEW)::jsonb
        WHERE id = existing_id;
    ELSE
        INSERT INTO sync_queue (entity_type, entity_id, action, payload, status, priority, correlation_id)
        VALUES (
            entity_name,
            NEW.id,
            TG_OP,
            row_to_json(NEW)::jsonb,
            'pending',
            prio,
            corr_id
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_sync_pos_transactions ON pos_transactions;
CREATE TRIGGER trg_sync_pos_transactions
    AFTER INSERT OR UPDATE ON pos_transactions
    FOR EACH ROW
    EXECUTE FUNCTION trg_enqueue_sync_outbox('pos_transactions');
`

type SyncService struct {
	ctx                  context.Context
	cancel               context.CancelFunc
	masterPool           *pgxpool.Pool
	cloudURL             string
	tenantSlug           string
	grpcAddr             string
	zatcaCloudURL        string
	outboxSchemaReady    atomic.Bool
	zatcaSchemaReady     atomic.Bool
	zatcaBackfilled      atomic.Bool
	orderAutomationReady atomic.Bool
	lastAppliedSeq       atomic.Int64
	orderSubmitter       corezatca.OrderSubmitter
	syncMu               sync.Mutex
	connMu               sync.Mutex
	conn                 *grpc.ClientConn
	sessionPullAckCh     chan int64
}

type QuarantinedOutboxItem struct {
	ID            int64      `json:"id"`
	EntityType    string     `json:"entity_type"`
	EntityID      int64      `json:"entity_id"`
	Action        string     `json:"action"`
	LastError     string     `json:"last_error"`
	ErrorClass    string     `json:"error_class"`
	RetryCount    int        `json:"retry_count"`
	CreatedAt     time.Time  `json:"created_at"`
	QuarantinedAt *time.Time `json:"quarantined_at,omitempty"`
}

type SyncQueueStatus struct {
	PendingCount     int64                   `json:"pending_count"`
	QuarantinedCount int64                   `json:"quarantined_count"`
	QuarantinedItems []QuarantinedOutboxItem `json:"quarantined_items"`
}

type OutboxItem struct {
	ID            int64           `json:"id"`
	OutboxUUID    string          `json:"outbox_uuid"`
	EntityType    string          `json:"entity_type"`
	EntityID      int64           `json:"entity_id"`
	Action        string          `json:"action"`
	Payload       json.RawMessage `json:"payload"`
	CreatedAt     time.Time       `json:"created_at"`
	CorrelationID *string         `json:"correlation_id"`
}

// CorrelationBatch groups all OutboxItems sharing a correlation_id (plus any synthesized
// parent checkout items) for atomic transaction apply on Cloud.
type CorrelationBatch struct {
	BatchID   string
	Events    []OutboxItem
	OutboxIDs []int64
}

func NewSyncService(ctx context.Context, pool *pgxpool.Pool, cloudURL, slug string) *SyncService {
	grpcTarget := extractGRPCTarget(cloudURL)
	if ctx == nil {
		ctx = context.Background()
	}
	workerCtx, cancel := context.WithCancel(ctx)

	return &SyncService{
		ctx:              workerCtx,
		cancel:           cancel,
		masterPool:       pool,
		cloudURL:         cloudURL,
		tenantSlug:       slug,
		grpcAddr:         grpcTarget,
		sessionPullAckCh: make(chan int64, 8),
	}
}

// enqueuePullAckSeq notifies the active StreamSession (if any) that the POS applied
// change_log events through the given sequence watermark.
func (s *SyncService) enqueuePullAckSeq(seq int64) {
	if s == nil || seq <= 0 || s.sessionPullAckCh == nil {
		return
	}
	select {
	case s.sessionPullAckCh <- seq:
	default:
		select {
		case <-s.sessionPullAckCh:
		default:
		}
		select {
		case s.sessionPullAckCh <- seq:
		default:
		}
	}
}

func extractGRPCTarget(target string) string {
	if target == "" {
		return "localhost:50051"
	}
	if strings.Contains(target, ":") && !strings.Contains(target, "://") {
		return target
	}
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

func isLoopbackOrTestTarget(target string) bool {
	if strings.HasPrefix(target, "passthrough:///") || strings.HasPrefix(target, "bufnet") {
		return true
	}
	host, _, err := net.SplitHostPort(target)
	if err != nil {
		host = target
	}
	host = strings.Trim(strings.ToLower(strings.TrimSpace(host)), "[]")
	if host == "" || host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return true
	}
	return false
}

func (s *SyncService) Start() {
	// Immediate sync on startup, persistent StreamSession listener, and 15s fallback ticker
	go s.performSync()
	go s.runSessionLoop()

	ticker := time.NewTicker(15 * time.Second)
	go func() {
		defer ticker.Stop()
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

func (s *SyncService) ConfigureZatcaSync(cloudURL string) { s.zatcaCloudURL = cloudURL }

func (s *SyncService) ConfigureOrderAutomation(submitter corezatca.OrderSubmitter) {
	s.orderSubmitter = submitter
}

// Stop cancels in-flight RPCs, closes the shared gRPC connection, and waits for the worker to leave the database.
func (s *SyncService) Stop() {
	s.cancel()
	s.syncMu.Lock()
	s.syncMu.Unlock()
	s.closeConn()
}

// getOrCreateConn returns the shared gRPC ClientConn for s.grpcAddr, creating one
// lazily or replacing it if it has entered Shutdown state.
func (s *SyncService) getOrCreateConn() (*grpc.ClientConn, error) {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	if s.conn != nil {
		if s.conn.GetState() != connectivity.Shutdown {
			return s.conn, nil
		}
		_ = s.conn.Close()
		s.conn = nil
	}

	var creds credentials.TransportCredentials
	if isLoopbackOrTestTarget(s.grpcAddr) || strings.EqualFold(os.Getenv("NEMBUS_GRPC_INSECURE"), "true") {
		creds = insecure.NewCredentials()
	} else {
		creds = credentials.NewTLS(&tls.Config{
			MinVersion: tls.VersionTLS12,
		})
	}

	conn, err := grpc.NewClient(
		s.grpcAddr,
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return nil, err
	}
	s.conn = conn
	return s.conn, nil
}

func (s *SyncService) closeConn() {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	if s.conn != nil {
		_ = s.conn.Close()
		s.conn = nil
	}
}

func (s *SyncService) withSyncMetadata(ctx context.Context, storeID int32, deviceID, syncToken string) context.Context {
	pairs := []string{
		"x-tenant-id", s.tenantSlug,
		syncpb.SyncStoreIDMetadataKey, strconv.FormatInt(int64(storeID), 10),
	}
	if deviceID != "" {
		pairs = append(pairs, syncpb.SyncDeviceIDMetadataKey, deviceID)
	}
	if syncToken != "" {
		pairs = append(pairs, syncpb.SyncTokenMetadataKey, syncToken)
	}
	return metadata.AppendToOutgoingContext(ctx, pairs...)
}

// runSessionLoop maintains a persistent bidirectional StreamSession (Phase 4)
// to receive instant ChangesAvailableNotification signals from Cloud.
func (s *SyncService) runSessionLoop() {
	attempt := 0
	for {
		if s.ctx.Err() != nil || s.masterPool == nil {
			return
		}
		credCtx, cancelCred := context.WithTimeout(s.ctx, 10*time.Second)
		storeID, deviceID, syncToken, err := ReadLocalDeviceCredentials(credCtx, s.masterPool)
		cancelCred()
		if err != nil || storeID <= 0 {
			select {
			case <-s.ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

		conn, err := s.getOrCreateConn()
		if err != nil {
			delay := ComputeBackoffDelay(attempt)
			attempt++
			select {
			case <-s.ctx.Done():
				return
			case <-time.After(delay):
				continue
			}
		}

		client := syncpb.NewSyncServiceClient(conn)
		sessionCtx, cancelSession := context.WithCancel(s.withSyncMetadata(s.ctx, storeID, deviceID, syncToken))
		stream, err := client.StreamSession(sessionCtx)
		if err != nil {
			cancelSession()
			if status.Code(err) == codes.Unimplemented {
				log.Printf("[gRPC SYNC] Cloud server does not implement StreamSession yet; relying on 15s polling")
				return
			}
			delay := ComputeBackoffDelay(attempt)
			attempt++
			select {
			case <-s.ctx.Done():
				return
			case <-time.After(delay):
				continue
			}
		}

		attempt = 0
		go func(sendStream syncpb.SyncService_StreamSessionClient) {
			for {
				select {
				case <-sessionCtx.Done():
					return
				case seq := <-s.sessionPullAckCh:
					_ = sendStream.Send(&syncpb.SessionClientMsg{
						Msg: &syncpb.SessionClientMsg_PullAckSeq{PullAckSeq: seq},
					})
				}
			}
		}(stream)

		for {
			msg, recvErr := stream.Recv()
			if recvErr != nil {
				cancelSession()
				if status.Code(recvErr) == codes.Unimplemented {
					return
				}
				break
			}
			if notif := msg.GetChangesAvailable(); notif != nil {
				if notif.LatestSeq > s.lastAppliedSeq.Load() {
					s.SyncNow()
				}
			}
		}

		delay := ComputeBackoffDelay(attempt)
		attempt++
		select {
		case <-s.ctx.Done():
			return
		case <-time.After(delay):
		}
	}
}

func (s *SyncService) ensureOutboxSchema(ctx context.Context) {
	if s.outboxSchemaReady.Load() || s.masterPool == nil {
		return
	}
	if _, err := s.masterPool.Exec(ctx, OutboxResilienceDDL); err != nil {
		log.Printf("[gRPC SYNC] Outbox resilience schema check warning: %v", err)
		return
	}
	s.outboxSchemaReady.Store(true)
}

func (s *SyncService) TenantSlug() string { return s.tenantSlug }

// SyncNow triggers immediate outbox draining and delta fetch
func (s *SyncService) SyncNow() {
	go s.performSync()
}

func (s *SyncService) performSync() {
	// Startup, the timer and SyncNow can overlap. Only one worker may drain
	// the queue, otherwise the same failure consumes several retries at once.
	if !s.syncMu.TryLock() {
		log.Printf("[gRPC SYNC] Previous cycle is still running for tenant %s", s.tenantSlug)
		return
	}
	defer s.syncMu.Unlock()
	if s.masterPool == nil || s.ctx.Err() != nil {
		return
	}
	started := time.Now()
	if s.orderSubmitter != nil {
		if !s.orderAutomationReady.Load() {
			if _, err := s.masterPool.Exec(s.ctx, corezatca.OrderAutomationDDL); err != nil {
				log.Printf("[ZATCA ORDER] Job installation failed: %v", err)
			} else {
				s.orderAutomationReady.Store(true)
			}
		}
		if s.orderAutomationReady.Load() {
			orderCtx, cancel := context.WithTimeout(s.ctx, 60*time.Second)
			if err := corezatca.ProcessOrderAutomationBatch(orderCtx, s.masterPool, s.orderSubmitter); err != nil && orderCtx.Err() == nil {
				log.Printf("[ZATCA ORDER] Automatic submission failed: %v", err)
			}
			cancel()
		}
	}
	if s.zatcaCloudURL != "" {
		if !s.zatcaSchemaReady.Load() {
			if _, err := s.masterPool.Exec(s.ctx, ZatcaOutboxDDL); err != nil {
				log.Printf("[ZATCA SYNC] Outbox installation failed: %v", err)
			} else {
				s.zatcaSchemaReady.Store(true)
			}
		}
		if s.zatcaSchemaReady.Load() && !s.zatcaBackfilled.Load() {
			var configured bool
			if err := s.masterPool.QueryRow(s.ctx, `SELECT EXISTS(SELECT 1 FROM local_device_config WHERE pos_terminal_id>0 AND store_id>0)`).Scan(&configured); err == nil && configured {
				if _, err := s.masterPool.Exec(s.ctx, backfillZatcaOutboxSQL); err != nil {
					log.Printf("[ZATCA SYNC] Initial backup enqueue failed: %v", err)
				} else {
					s.zatcaBackfilled.Store(true)
				}
			}
		}
	}
	log.Printf("[gRPC SYNC] Starting cycle for tenant %s (target %s)", s.tenantSlug, s.grpcAddr)
	defer func() { log.Printf("[gRPC SYNC] Cycle finished in %s", time.Since(started).Round(time.Millisecond)) }()

	// 1. gRPC Push Strategy (Local Outbox -> Cloud)
	s.drainOutboxGRPC()
	s.drainZatcaOutbox()

	// 2. gRPC Pull Strategy (Cloud change_log -> Local Watermarks with StreamPull fallback)
	s.fetchChangesGRPC()
}

// drainOutboxGRPC streams pending sync_queue items to Cloud via gRPC StreamPush with SHA-256 checksums.
// Note: items sharing the same correlation_id are allowed through together and ordered by PushEntityRank
// inside their CorrelationBatch, while cross-transaction children remain blocked if an earlier parent
// transaction is still pending, failed, or quarantined.
const pendingOutboxSQL = `
	SELECT q.id, COALESCE(q.outbox_uuid::text, ''), q.entity_type, q.entity_id, q.action, q.payload, q.created_at, q.correlation_id
	FROM sync_queue q
	WHERE q.status = 'pending'
	  AND COALESCE(q.next_attempt_at, q.created_at, NOW()) <= NOW()
	  AND q.entity_type NOT IN ('zatca_device_configs','zatca_document_chain')
	  AND NOT EXISTS (
		SELECT 1 FROM sync_queue parent
		WHERE (
			(q.entity_type IN ('pos_transaction_lines', 'pos_payments')
			AND parent.entity_type = 'pos_transactions'
			AND parent.entity_id = (q.payload->>'transaction_id')::bigint)
			OR
			(q.entity_type IN ('pos_transactions', 'sales_returns')
			AND parent.entity_type = 'customers'
			AND parent.entity_id = (q.payload->>'customer_id')::bigint)
			OR
			(q.entity_type = 'sales_return_lines'
			AND parent.entity_type = 'sales_returns'
			AND parent.entity_id = (q.payload->>'return_id')::bigint)
		)
		AND parent.status IN ('pending', 'failed', 'quarantined')
		AND (
			parent.status IN ('failed', 'quarantined')
			OR COALESCE(NULLIF(parent.correlation_id, ''), '__parent_none__') <> COALESCE(NULLIF(q.correlation_id, ''), '__child_none__')
		)
	  )
	ORDER BY q.id ASC
	LIMIT 100;
`

func (s *SyncService) drainOutboxGRPC() {
	ctx, cancel := context.WithTimeout(s.ctx, 60*time.Second)
	defer cancel()
	s.ensureOutboxSchema(ctx)

	rows, err := s.masterPool.Query(ctx, pendingOutboxSQL)
	if err != nil {
		log.Printf("[gRPC SYNC] Failed to query local sync_queue: %v", err)
		return
	}
	defer rows.Close()

	var items []OutboxItem
	for rows.Next() {
		var item OutboxItem
		if err := rows.Scan(&item.ID, &item.OutboxUUID, &item.EntityType, &item.EntityID, &item.Action, &item.Payload, &item.CreatedAt, &item.CorrelationID); err != nil {
			continue
		}
		if item.OutboxUUID == "" {
			item.OutboxUUID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("outbox-%d", item.ID))).String()
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

	batches, skippedItems := s.groupOutboxItemsIntoBatches(ctx, items)
	for _, skipped := range skippedItems {
		s.recordOutboxFailureErr([]OutboxItem{skipped.item}, skipped.err)
	}
	if len(batches) == 0 {
		return
	}

	log.Printf("[gRPC SYNC] Pushing %d pending outbox entries across %d correlation batch(es)", len(items), len(batches))

	storeID, deviceID, syncToken, _ := ReadLocalDeviceCredentials(ctx, s.masterPool)

	conn, err := s.getOrCreateConn()
	if err != nil {
		log.Printf("[gRPC SYNC] Connection failed to %s: %v", s.grpcAddr, err)
		s.recordOutboxFailureErr(items, err)
		return
	}

	client := syncpb.NewSyncServiceClient(conn)
	pushCtx := s.withSyncMetadata(ctx, storeID, deviceID, syncToken)
	stream, err := client.StreamPush(pushCtx)
	if err != nil {
		log.Printf("[gRPC SYNC] StreamPush initialization error: %v", err)
		s.recordOutboxFailureErr(items, err)
		return
	}

	syncedCount := s.pushCorrelationBatchesWindowed(pushCtx, stream, batches, storeID, deviceID)
	_ = stream.CloseSend()

	if syncedCount > 0 {
		log.Printf("[gRPC SYNC] Pushed %d outbox items upstream to Cloud via gRPC stream", syncedCount)
	}
}

type skippedOutboxItem struct {
	item OutboxItem
	err  error
}

// groupOutboxItemsIntoBatches groups pending OutboxItems by correlation_id, synthesizes
// any missing checkout parents (for legacy/direct-insert pos_transactions rows), and sorts
// each batch in deterministic parent-before-child dependency order (PushEntityRank).
func (s *SyncService) groupOutboxItemsIntoBatches(ctx context.Context, items []OutboxItem) ([]CorrelationBatch, []skippedOutboxItem) {
	batchMap := make(map[string]*CorrelationBatch)
	var batchOrder []string
	var skipped []skippedOutboxItem

	for _, item := range items {
		batchKey := ""
		if item.CorrelationID != nil && strings.TrimSpace(*item.CorrelationID) != "" {
			batchKey = strings.TrimSpace(*item.CorrelationID)
		} else {
			batchKey = fmt.Sprintf("single-%d", item.ID)
		}

		b, exists := batchMap[batchKey]
		if !exists {
			b = &CorrelationBatch{
				BatchID: batchKey,
			}
			batchMap[batchKey] = b
			batchOrder = append(batchOrder, batchKey)
		}
		b.Events = append(b.Events, item)
		b.OutboxIDs = append(b.OutboxIDs, item.ID)
	}

	var result []CorrelationBatch
	for _, key := range batchOrder {
		b := batchMap[key]

		// Check if batch already has an explicitly enqueued sales_orders_v2 (POS_CHECKOUT)
		// item. If it does, verify that the order is already fulfilled before allowing it
		// to be pushed to Cloud. If it is not yet fulfilled (e.g. the cashier suspended or
		// is still editing the order), hold the batch locally for this cycle (Solution B).
		batchNotReady := false
		hasCheckoutOrder := false
		for i := range b.Events {
			ev := &b.Events[i]
			if ev.EntityType == "sales_orders_v2" && ev.Action == "POS_CHECKOUT" {
				// Parse order ID and fulfillment_status from the enqueued payload.
				var checkoutPayload struct {
					Order struct {
						ID                string `json:"id"`
						FulfillmentStatus string `json:"fulfillment_status"`
					} `json:"order"`
				}
				if ev.Payload != nil {
					_ = json.Unmarshal(ev.Payload, &checkoutPayload)
				}

				// If we have access to the master pool, check the LIVE state of the order
				// in sales_orders_v2. When the outbox item was originally enqueued (at cart conversion),
				// the order might have been unfulfilled; when the cashier finishes checkout/payment,
				// sales_orders_v2 is updated to 'fulfilled'. We must inspect the live state and refresh
				// the payload so cloud receives the fulfilled order.
				if s.masterPool != nil && checkoutPayload.Order.ID != "" {
					var livePayload []byte
					var liveStatus string
					err := s.masterPool.QueryRow(ctx, `
						SELECT json_build_object(
							'order', row_to_json(o),
							'lines', COALESCE((SELECT json_agg(l ORDER BY l.line_number)
								FROM sales_order_lines_v2 l WHERE l.sales_order_id = o.id), '[]'::json)
						)::text,
						COALESCE(o.fulfillment_status::text, '')
						FROM sales_orders_v2 o
						WHERE o.id = $1::uuid
					`, checkoutPayload.Order.ID).Scan(&livePayload, &liveStatus)
					if err == nil {
						if liveStatus != "fulfilled" {
							batchNotReady = true
							break
						}
						// Live order is fulfilled locally! Refresh the payload so cloud receives it.
						ev.Payload = livePayload
						if ev.ID > 0 {
							_, _ = s.masterPool.Exec(ctx, `UPDATE sync_queue SET payload = $1::jsonb WHERE id = $2`, string(livePayload), ev.ID)
						}
						hasCheckoutOrder = true
						break
					}
				}

				if checkoutPayload.Order.FulfillmentStatus != "fulfilled" {
					// Order not yet finalised locally — hold this batch.
					batchNotReady = true
					break
				}
				hasCheckoutOrder = true
				break
			}
		}

		if batchNotReady {
			continue
		}

		if !hasCheckoutOrder {
			var synthesized []OutboxItem
			var parentErr error
			var offendingItem OutboxItem

			for _, ev := range b.Events {
				if ev.EntityType != "pos_transactions" {
					continue
				}
				parents, err := s.transactionParentsContext(ctx, ev)
				if err != nil {
					if errors.Is(err, errCheckoutNotReady) {
						batchNotReady = true
						break
					}
					parentErr = err
					offendingItem = ev
					break
				}
				for idx, p := range parents {
					p.OutboxUUID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("%s-parent-%d-%s-%s", ev.OutboxUUID, idx, p.EntityType, p.Action))).String()
					synthesized = append(synthesized, p)
				}
			}

			if batchNotReady {
				continue
			}
			if parentErr != nil {
				skipped = append(skipped, skippedOutboxItem{item: offendingItem, err: parentErr})
				continue
			}
			if len(synthesized) > 0 {
				b.Events = append(synthesized, b.Events...)
			}
		}

		OrderBatchEvents(b.Events)
		result = append(result, *b)
	}

	return result, skipped
}

// OrderBatchEvents sorts a slice of OutboxItems in parent-before-child dependency order
// using syncpb.PushEntityRank, preserving original ID order for ties.
func OrderBatchEvents(events []OutboxItem) {
	sort.SliceStable(events, func(i, j int) bool {
		rankI := syncpb.PushEntityRank(events[i].EntityType, events[i].Action)
		rankJ := syncpb.PushEntityRank(events[j].EntityType, events[j].Action)
		if rankI != rankJ {
			return rankI < rankJ
		}
		return events[i].ID < events[j].ID
	})
}

// pushCorrelationBatchesWindowed streams CorrelationBatches over StreamPush with a sliding
// window of up to syncpb.MaxInFlightBatches (10) unacknowledged batches in flight.
func (s *SyncService) pushCorrelationBatchesWindowed(
	ctx context.Context,
	stream syncpb.SyncService_StreamPushClient,
	batches []CorrelationBatch,
	storeID int32,
	deviceID string,
) int {
	if len(batches) == 0 {
		return 0
	}

	windowSlots := make(chan struct{}, syncpb.MaxInFlightBatches)
	inFlightCh := make(chan CorrelationBatch, len(batches))
	sendDone := make(chan error, 1)

	go func() {
		defer close(inFlightCh)
		for _, batch := range batches {
			select {
			case <-ctx.Done():
				sendDone <- ctx.Err()
				return
			case windowSlots <- struct{}{}:
			}

			batchSize := int32(len(batch.Events))
			var sendErr error
			for idx, item := range batch.Events {
				hash := sha256.Sum256(item.Payload)
				checksum := hex.EncodeToString(hash[:])
				event := &syncpb.SyncEvent{
					Id:          item.ID,
					OutboxUuid:  item.OutboxUUID,
					DeviceId:    deviceID,
					BatchId:     batch.BatchID,
					BatchSize:   batchSize,
					SeqInBatch:  int32(idx + 1),
					EntityType:  item.EntityType,
					EntityId:    item.EntityID,
					Action:      item.Action,
					PayloadJson: item.Payload,
					StoreId:     storeID,
					TenantSlug:  s.tenantSlug,
					EventTime:   timestamppb.New(item.CreatedAt),
					Sha256:      checksum,
					IsLastChunk: true,
				}
				if item.CorrelationID != nil {
					event.CorrelationId = *item.CorrelationID
				}
				if err := stream.Send(event); err != nil {
					sendErr = fmt.Errorf("send error: %w", err)
					break
				}
			}
			if sendErr != nil {
				<-windowSlots
				s.recordOutboxIDsFailureErr(batch.OutboxIDs, sendErr)
				sendDone <- sendErr
				return
			}
			inFlightCh <- batch
		}
		sendDone <- nil
	}()

	var syncedCount int
	for batch := range inFlightCh {
		ack, err := stream.Recv()
		<-windowSlots
		if err != nil {
			s.recordOutboxIDsFailureErr(batch.OutboxIDs, fmt.Errorf("receive error: %w", err))
			for remaining := range inFlightCh {
				<-windowSlots
				s.recordOutboxIDsFailureErr(remaining.OutboxIDs, fmt.Errorf("stream closed before ack: %w", err))
			}
			break
		}

		// Backward compatibility: if talking to an older server that sends per-event acks (ack.BatchId == "")
		if ack.BatchId == "" && len(batch.Events) > 1 {
			batchFailed := !ack.Success
			failErr := s.buildBatchAckError(ack)
			for i := 1; i < len(batch.Events); i++ {
				extraAck, recvErr := stream.Recv()
				if recvErr != nil {
					batchFailed = true
					failErr = fmt.Errorf("receive error: %w", recvErr)
					break
				}
				if !extraAck.Success && !batchFailed {
					batchFailed = true
					failErr = s.buildBatchAckError(extraAck)
				}
			}
			if batchFailed {
				s.recordOutboxIDsFailureErr(batch.OutboxIDs, failErr)
				continue
			}
			if err := s.markOutboxBatchSynced(batch.OutboxIDs); err != nil {
				log.Printf("[gRPC SYNC] Could not save batch acknowledgment for %s: %v", batch.BatchID, err)
				break
			}
			syncedCount += len(batch.OutboxIDs)
			continue
		}

		if !ack.Success {
			failErr := s.buildBatchAckError(ack)
			log.Printf("[gRPC SYNC] Batch %s rejected (failed_seq=%d, sql_state=%s): %v",
				batch.BatchID, ack.FailedSeq, ack.SqlState, failErr)
			s.recordOutboxIDsFailureErr(batch.OutboxIDs, failErr)
			continue
		}

		if err := s.markOutboxBatchSynced(batch.OutboxIDs); err != nil {
			log.Printf("[gRPC SYNC] Could not save push acknowledgment for batch %s: %v", batch.BatchID, err)
			break
		}
		syncedCount += len(batch.OutboxIDs)
	}

	<-sendDone
	return syncedCount
}

func (s *SyncService) buildBatchAckError(ack *syncpb.SyncAck) error {
	if ack == nil {
		return errors.New("nil SyncAck from server")
	}
	if ack.SqlState != "" {
		return fmt.Errorf("server rejected (SQLSTATE %s, seq %d): %s", ack.SqlState, ack.FailedSeq, ack.ErrorMessage)
	}
	return fmt.Errorf("server rejected: %s", ack.ErrorMessage)
}

func (s *SyncService) markOutboxItemSynced(id int64) error {
	return s.markOutboxBatchSynced([]int64{id})
}

func (s *SyncService) markOutboxBatchSynced(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	saveCtx, cancelSave := context.WithTimeout(s.ctx, 15*time.Second)
	defer cancelSave()
	_, err := s.masterPool.Exec(saveCtx, `
		UPDATE sync_queue
		SET status = 'synced',
		    synced_at = NOW(),
		    last_error = NULL,
		    error_class = NULL,
		    quarantined_at = NULL
		WHERE id = ANY($1);
	`, ids)
	return err
}

var errCheckoutNotReady = errors.New("checkout is not fulfilled locally yet")

// transactionParents returns checkout prerequisites for a POS transaction.
// Cart and order reference each other (converted_to_order_id / source_cart_id), so
// cloud sync sends: cart without conversion link, POS_CHECKOUT order, then cart UPDATE.
func (s *SyncService) transactionParents(item OutboxItem) ([]OutboxItem, error) {
	ctx, cancel := context.WithTimeout(s.ctx, 15*time.Second)
	defer cancel()
	return s.transactionParentsContext(ctx, item)
}

func (s *SyncService) transactionParentsContext(ctx context.Context, item OutboxItem) ([]OutboxItem, error) {
	if item.EntityType != "pos_transactions" {
		return nil, nil
	}
	var refs struct {
		ID           int64   `json:"id"`
		SalesOrderID *string `json:"sales_order_id"`
		SourceCartID *string `json:"source_cart_id"`
	}
	if err := json.Unmarshal(item.Payload, &refs); err != nil {
		return nil, fmt.Errorf("invalid transaction payload: %w", err)
	}
	if refs.SalesOrderID == nil && refs.SourceCartID == nil && refs.ID == 0 {
		return nil, nil
	}
	rows, err := s.masterPool.Query(ctx, `
		WITH pt AS (
			SELECT sales_order_id, source_cart_id
			FROM pos_transactions
			WHERE id = $3
		),
		ctx AS (
			SELECT
				COALESCE($1::uuid, (SELECT sales_order_id FROM pt)) AS order_id,
				COALESCE(
					$2::uuid,
					(SELECT source_cart_id FROM pt),
					(SELECT o.source_cart_id FROM sales_orders_v2 o WHERE o.id = COALESCE($1::uuid, (SELECT sales_order_id FROM pt)))
				) AS cart_id
		)
		SELECT entity_type, payload FROM (
			SELECT 1 AS seq, 'sales_orders_v2' AS entity_type, json_build_object(
				'order', row_to_json(o),
				'lines', COALESCE((SELECT json_agg(l ORDER BY l.line_number)
					FROM sales_order_lines_v2 l WHERE l.sales_order_id = o.id), '[]'::json)
			)::text AS payload
			FROM sales_orders_v2 o, ctx
			WHERE ctx.order_id IS NOT NULL AND o.id = ctx.order_id
			UNION ALL
			SELECT 2 AS seq, 'carts' AS entity_type, row_to_json(c)::text AS payload
			FROM carts c, ctx
			WHERE ctx.cart_id IS NOT NULL AND c.id = ctx.cart_id
		) parents ORDER BY seq;
	`, refs.SalesOrderID, refs.SourceCartID, refs.ID)
	if err != nil {
		return nil, fmt.Errorf("read transaction parents: %w", err)
	}
	defer rows.Close()

	base := OutboxItem{ID: item.ID, CreatedAt: item.CreatedAt, CorrelationID: item.CorrelationID}
	var orderPayload, cartPayload json.RawMessage
	for rows.Next() {
		var entityType, payload string
		if err := rows.Scan(&entityType, &payload); err != nil {
			return nil, fmt.Errorf("read transaction parent: %w", err)
		}
		switch entityType {
		case "sales_orders_v2":
			var checkout struct {
				Order struct {
					FulfillmentStatus string `json:"fulfillment_status"`
					OrderStatus       string `json:"order_status"`
				} `json:"order"`
			}
			if err := json.Unmarshal([]byte(payload), &checkout); err != nil {
				return nil, err
			}
			if checkout.Order.FulfillmentStatus != "fulfilled" || checkout.Order.OrderStatus != "fulfilled" {
				return nil, errCheckoutNotReady
			}
			orderPayload = json.RawMessage(payload)
		case "carts":
			cartPayload = json.RawMessage(payload)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if orderPayload == nil && cartPayload == nil {
		return nil, nil
	}

	var parents []OutboxItem
	if cartPayload != nil {
		stripped, err := stripCartConvertedOrderLink(cartPayload)
		if err != nil {
			return nil, fmt.Errorf("strip cart order link: %w", err)
		}
		parents = append(parents, OutboxItem{
			ID: item.ID, EntityType: "carts", Action: "INSERT",
			Payload: stripped, CreatedAt: base.CreatedAt, CorrelationID: base.CorrelationID,
		})
	}
	if orderPayload != nil {
		parents = append(parents, OutboxItem{
			ID: item.ID, EntityType: "sales_orders_v2", Action: "POS_CHECKOUT",
			Payload: orderPayload, CreatedAt: base.CreatedAt, CorrelationID: base.CorrelationID,
		})
	}
	if cartPayload != nil {
		parents = append(parents, OutboxItem{
			ID: item.ID, EntityType: "carts", Action: "UPDATE",
			Payload: cartPayload, CreatedAt: base.CreatedAt, CorrelationID: base.CorrelationID,
		})
	}
	return parents, nil
}

// stripCartConvertedOrderLink clears converted_to_order_id so cloud can insert the
// cart before the sales order exists. The link is restored in a follow-up UPDATE.
func stripCartConvertedOrderLink(cartPayload json.RawMessage) (json.RawMessage, error) {
	var cart map[string]json.RawMessage
	if err := json.Unmarshal(cartPayload, &cart); err != nil {
		return nil, err
	}
	cart["converted_to_order_id"] = json.RawMessage("null")
	cart["converted_at"] = json.RawMessage("null")
	return json.Marshal(cart)
}

func pushOutboxItem(stream syncpb.SyncService_StreamPushClient, item OutboxItem, storeID int32, tenantSlug string) error {
	hash := sha256.Sum256(item.Payload)
	checksum := hex.EncodeToString(hash[:])
	event := &syncpb.SyncEvent{
		Id: item.ID, OutboxUuid: item.OutboxUUID, EntityType: item.EntityType, EntityId: item.EntityID,
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
	s.recordOutboxFailureErr(items, errors.New(errMsg))
}

func (s *SyncService) recordOutboxFailureErr(items []OutboxItem, syncErr error) {
	if len(items) == 0 || syncErr == nil {
		return
	}
	ids := make([]int64, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}
	s.recordOutboxIDsFailureErr(ids, syncErr)
}

func (s *SyncService) recordOutboxIDsFailureErr(ids []int64, syncErr error) {
	if len(ids) == 0 || syncErr == nil {
		return
	}
	ctx, cancel := context.WithTimeout(s.ctx, 15*time.Second)
	defer cancel()
	s.ensureOutboxSchema(ctx)

	errClass := ClassifySyncError(syncErr)
	errMsg := syncErr.Error()

	if errClass == ErrorClassPermanent {
		_, err := s.masterPool.Exec(ctx, `
			UPDATE sync_queue
			SET retry_count    = COALESCE(retry_count, 0) + 1,
			    last_error     = $1,
			    error_class    = 'permanent',
			    status         = 'quarantined',
			    quarantined_at = COALESCE(quarantined_at, NOW())
			WHERE id = ANY($2);
		`, errMsg, ids)
		if err != nil {
			log.Printf("[gRPC SYNC] Could not quarantine outbox item(s): %v", err)
		}
		return
	}

	// Transient failure: infinite retry with exponential backoff + jitter; never burn into 'failed'.
	for _, id := range ids {
		var currentRetries int
		_ = s.masterPool.QueryRow(ctx, `SELECT COALESCE(retry_count, 0) FROM sync_queue WHERE id = $1`, id).Scan(&currentRetries)
		delay := ComputeBackoffDelay(currentRetries)
		nextAttempt := time.Now().UTC().Add(delay)
		_, err := s.masterPool.Exec(ctx, `
			UPDATE sync_queue
			SET retry_count     = COALESCE(retry_count, 0) + 1,
			    last_error      = $1,
			    error_class     = 'transient',
			    status          = 'pending',
			    next_attempt_at = $2
			WHERE id = $3;
		`, errMsg, nextAttempt, id)
		if err != nil {
			log.Printf("[gRPC SYNC] Could not record transient outbox backoff for item %d: %v", id, err)
		}
	}
}

// GetQueueStatus returns counts of pending and quarantined sync_queue entries
// along with details of quarantined items for the POS UI badge.
func (s *SyncService) GetQueueStatus(ctx context.Context) (*SyncQueueStatus, error) {
	if s.masterPool == nil {
		return &SyncQueueStatus{QuarantinedItems: []QuarantinedOutboxItem{}}, nil
	}
	if ctx == nil {
		ctx = s.ctx
	}
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	s.ensureOutboxSchema(queryCtx)

	qStatus := &SyncQueueStatus{
		QuarantinedItems: make([]QuarantinedOutboxItem, 0),
	}

	err := s.masterPool.QueryRow(queryCtx, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status IN ('quarantined', 'failed'))
		FROM sync_queue
		WHERE entity_type NOT IN ('zatca_device_configs', 'zatca_document_chain');
	`).Scan(&qStatus.PendingCount, &qStatus.QuarantinedCount)
	if err != nil {
		return nil, fmt.Errorf("read sync_queue counts: %w", err)
	}

	rows, err := s.masterPool.Query(queryCtx, `
		SELECT id, entity_type, entity_id, action, COALESCE(last_error, ''),
		       COALESCE(error_class, 'permanent'), COALESCE(retry_count, 0),
		       COALESCE(created_at, NOW()), quarantined_at
		FROM sync_queue
		WHERE status IN ('quarantined', 'failed')
		  AND entity_type NOT IN ('zatca_device_configs', 'zatca_document_chain')
		ORDER BY COALESCE(quarantined_at, created_at) DESC, id DESC
		LIMIT 50;
	`)
	if err != nil {
		return nil, fmt.Errorf("read quarantined sync_queue items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item QuarantinedOutboxItem
		if err := rows.Scan(
			&item.ID,
			&item.EntityType,
			&item.EntityID,
			&item.Action,
			&item.LastError,
			&item.ErrorClass,
			&item.RetryCount,
			&item.CreatedAt,
			&item.QuarantinedAt,
		); err != nil {
			continue
		}
		qStatus.QuarantinedItems = append(qStatus.QuarantinedItems, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return qStatus, nil
}

// RetryQuarantinedItems resets specific (or all, when ids is empty) quarantined/failed
// sync_queue items back to 'pending' with immediate eligibility.
func (s *SyncService) RetryQuarantinedItems(ctx context.Context, ids []int64) (int64, error) {
	if s.masterPool == nil {
		return 0, fmt.Errorf("sync database pool is not initialized")
	}
	if ctx == nil {
		ctx = s.ctx
	}
	execCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	s.ensureOutboxSchema(execCtx)

	var rowsAffected int64
	if len(ids) > 0 {
		tag, err := s.masterPool.Exec(execCtx, `
			UPDATE sync_queue
			SET status          = 'pending',
			    retry_count     = 0,
			    last_error      = NULL,
			    error_class     = NULL,
			    quarantined_at  = NULL,
			    next_attempt_at = NOW()
			WHERE id = ANY($1) AND status IN ('quarantined', 'failed');
		`, ids)
		if err != nil {
			return 0, fmt.Errorf("failed to retry quarantined sync_queue items: %w", err)
		}
		rowsAffected = tag.RowsAffected()
	} else {
		tag, err := s.masterPool.Exec(execCtx, `
			UPDATE sync_queue
			SET status          = 'pending',
			    retry_count     = 0,
			    last_error      = NULL,
			    error_class     = NULL,
			    quarantined_at  = NULL,
			    next_attempt_at = NOW()
			WHERE status IN ('quarantined', 'failed');
		`)
		if err != nil {
			return 0, fmt.Errorf("failed to retry all quarantined sync_queue items: %w", err)
		}
		rowsAffected = tag.RowsAffected()
	}

	if rowsAffected > 0 {
		log.Printf("[gRPC SYNC] Reset %d quarantined/failed sync_queue items to pending for immediate retry", rowsAffected)
		s.SyncNow()
	}
	return rowsAffected, nil
}

// ResetFailedItems resets legacy 'failed' sync_queue entries back to 'pending' so they are
// retried on the next sync cycle. Quarantined items remain quarantined until operator retry.
func (s *SyncService) ResetFailedItems() (int64, error) {
	s.ensureOutboxSchema(s.ctx)
	tag, err := s.masterPool.Exec(s.ctx, `
		UPDATE sync_queue
		SET status          = 'pending',
		    retry_count     = 0,
		    last_error      = NULL,
		    error_class     = NULL,
		    next_attempt_at = NOW()
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

// Ensure io import is used
var _ = io.EOF
