package grpc

import (
	"context"
	"crypto/subtle"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/syncpb"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const TenantSlugMetadataKey = "x-tenant-id"

func (s *SyncServer) verifyStreamAuth(ctx context.Context, tenantSlug, deviceID string, storeID int32) error {
	_, _, err := s.verifyStreamAuthIdentity(ctx, tenantSlug, deviceID, storeID)
	return err
}

func (s *SyncServer) verifyStreamAuthIdentity(ctx context.Context, tenantSlug, deviceID string, storeID int32) (int32, int32, error) {
	var rawToken string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(syncpb.SyncTokenMetadataKey); len(vals) > 0 {
			rawToken = strings.TrimSpace(vals[0])
		}
		if storeID <= 0 {
			if vals := md.Get(syncpb.SyncStoreIDMetadataKey); len(vals) > 0 {
				if parsed, err := strconv.ParseInt(vals[0], 10, 32); err == nil {
					storeID = int32(parsed)
				}
			}
		}
		if deviceID == "" {
			if vals := md.Get(syncpb.SyncDeviceIDMetadataKey); len(vals) > 0 {
				deviceID = strings.TrimSpace(vals[0])
			}
		}
	}

	if rawToken == "" {
		if s.requireTokenAuth {
			return 0, 0, status.Error(codes.Unauthenticated, "missing terminal sync token")
		}
		return 0, storeID, nil
	}

	if tenantSlug == "" {
		return 0, 0, status.Error(codes.InvalidArgument, "tenant_slug is required for token verification")
	}
	pool, err := s.getTenantPool(ctx, tenantSlug)
	if err != nil {
		return 0, 0, status.Errorf(codes.NotFound, "tenant DB pool unavailable: %v", err)
	}

	candidateHash := usecase.HashSyncToken(rawToken)
	rows, err := pool.Query(ctx, `
		SELECT id, store_id, COALESCE(sync_token_hash, '')
		FROM pos_terminals
		WHERE sync_token_hash IS NOT NULL
		  AND ($1 <= 0 OR store_id = $1)
		  AND COALESCE(is_active, true) = true
	`, storeID)
	if err != nil {
		if s.requireTokenAuth {
			return 0, 0, status.Errorf(codes.Internal, "verify terminal sync token: %v", err)
		}
		return 0, storeID, nil
	}
	defer rows.Close()

	var (
		matched           bool
		matchedTerminalID int32
		matchedStoreID    int32
	)
	for rows.Next() {
		var (
			termID     int32
			termStore  int32
			storedHash string
		)
		if err := rows.Scan(&termID, &termStore, &storedHash); err != nil {
			continue
		}
		if len(storedHash) == len(candidateHash) &&
			subtle.ConstantTimeCompare([]byte(storedHash), []byte(candidateHash)) == 1 {
			matched = true
			matchedTerminalID = termID
			matchedStoreID = termStore
			break
		}
	}
	if !matched {
		return 0, 0, status.Error(codes.Unauthenticated, "invalid or revoked terminal sync token")
	}
	return matchedTerminalID, matchedStoreID, nil
}

// ChangeNotifyHub manages per-tenant PostgreSQL LISTEN/NOTIFY ("nembus_changes")
// fan-out to connected POS StreamSession clients.
type ChangeNotifyHub struct {
	mu           sync.Mutex
	poolResolver func(ctx context.Context, slug string) (*pgxpool.Pool, error)
	subscribers  map[string]map[chan int64]struct{}
	listeners    map[string]context.CancelFunc
}

func NewChangeNotifyHub(resolver func(ctx context.Context, slug string) (*pgxpool.Pool, error)) *ChangeNotifyHub {
	return &ChangeNotifyHub{
		poolResolver: resolver,
		subscribers:  make(map[string]map[chan int64]struct{}),
		listeners:    make(map[string]context.CancelFunc),
	}
}

// Subscribe registers a buffered channel to receive the latest change_log seq for tenantSlug.
func (h *ChangeNotifyHub) Subscribe(tenantSlug string) (<-chan int64, func()) {
	ch := make(chan int64, 16)
	if tenantSlug == "" {
		return ch, func() { close(ch) }
	}

	h.mu.Lock()
	subs, ok := h.subscribers[tenantSlug]
	if !ok {
		subs = make(map[chan int64]struct{})
		h.subscribers[tenantSlug] = subs
	}
	subs[ch] = struct{}{}
	if _, running := h.listeners[tenantSlug]; !running && h.poolResolver != nil {
		listenCtx, cancel := context.WithCancel(context.Background())
		h.listeners[tenantSlug] = cancel
		go h.listenTenantChanges(listenCtx, tenantSlug)
	}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if tenantSubs, exists := h.subscribers[tenantSlug]; exists {
			delete(tenantSubs, ch)
			if len(tenantSubs) == 0 {
				delete(h.subscribers, tenantSlug)
				if cancel, ok := h.listeners[tenantSlug]; ok {
					cancel()
					delete(h.listeners, tenantSlug)
				}
			}
		}
	}
	return ch, unsubscribe
}

// BroadcastSeq fans out a newly committed change_log sequence number to all active
// StreamSession subscribers for tenantSlug without blocking.
func (h *ChangeNotifyHub) BroadcastSeq(tenantSlug string, seq int64) {
	if tenantSlug == "" || seq <= 0 {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	subs := h.subscribers[tenantSlug]
	for ch := range subs {
		select {
		case ch <- seq:
		default:
			// Replace stale unread notification with the newer sequence watermark
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- seq:
			default:
			}
		}
	}
}

func (h *ChangeNotifyHub) listenTenantChanges(ctx context.Context, tenantSlug string) {
	for {
		if ctx.Err() != nil {
			return
		}
		if h.poolResolver == nil {
			return
		}
		pool, err := h.poolResolver(ctx, tenantSlug)
		if err != nil || pool == nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		conn, err := pool.Acquire(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		_, err = conn.Exec(ctx, `LISTEN nembus_changes`)
		if err != nil {
			conn.Release()
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		for {
			notification, waitErr := conn.Conn().WaitForNotification(ctx)
			if waitErr != nil {
				conn.Release()
				break
			}
			if notification != nil {
				seq, parseErr := strconv.ParseInt(strings.TrimSpace(notification.Payload), 10, 64)
				if parseErr == nil && seq > 0 {
					h.BroadcastSeq(tenantSlug, seq)
				}
			}
		}
	}
}

// StreamSession implements the Phase 4 persistent bidirectional gRPC session stream:
// - Windowed/pipelined correlation batch push with per-batch SyncAck
// - Server-pushed ChangesAvailableNotification(latest_seq) via PostgreSQL LISTEN/NOTIFY
func (s *SyncServer) StreamSession(stream syncpb.SyncService_StreamSessionServer) error {
	ctx := stream.Context()

	var (
		tenantSlug string
		deviceID   string
		storeID    int32
	)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(TenantSlugMetadataKey); len(vals) > 0 {
			tenantSlug = strings.TrimSpace(vals[0])
		}
		if vals := md.Get(syncpb.SyncDeviceIDMetadataKey); len(vals) > 0 {
			deviceID = strings.TrimSpace(vals[0])
		}
		if vals := md.Get(syncpb.SyncStoreIDMetadataKey); len(vals) > 0 {
			if parsed, err := strconv.ParseInt(vals[0], 10, 32); err == nil {
				storeID = int32(parsed)
			}
		}
	}

	var (
		authChecked    bool
		authTerminalID int32
		authStoreID    int32
	)
	if tenantSlug != "" {
		var err error
		authTerminalID, authStoreID, err = s.verifyStreamAuthIdentity(ctx, tenantSlug, deviceID, storeID)
		if err != nil {
			return err
		}
		authChecked = true
	}

	var sendMu sync.Mutex
	sendServerMsg := func(msg *syncpb.SessionServerMsg) error {
		sendMu.Lock()
		defer sendMu.Unlock()
		return stream.Send(msg)
	}

	var (
		subOnce     sync.Once
		unsubscribe func()
	)
	defer func() {
		if unsubscribe != nil {
			unsubscribe()
		}
	}()

	startSubscription := func(slug string) {
		if slug == "" {
			return
		}
		subOnce.Do(func() {
			notifyCh, unsub := s.ChangeHub().Subscribe(slug)
			unsubscribe = unsub
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case latestSeq, ok := <-notifyCh:
						if !ok {
							return
						}
						_ = sendServerMsg(&syncpb.SessionServerMsg{
							Msg: &syncpb.SessionServerMsg_ChangesAvailable{
								ChangesAvailable: &syncpb.ChangesAvailableNotification{
									LatestSeq: latestSeq,
								},
							},
						})
					}
				}
			}()
		})
	}

	if tenantSlug != "" {
		startSubscription(tenantSlug)
	}

	var currentBatchID string
	var currentBatch []*syncpb.SyncEvent
	var lastAckedSeq int64

	for {
		clientMsg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if pushEv := clientMsg.GetPushEvent(); pushEv != nil {
			if !authChecked && pushEv.TenantSlug != "" {
				var authErr error
				authTerminalID, authStoreID, authErr = s.verifyStreamAuthIdentity(ctx, pushEv.TenantSlug, pushEv.DeviceId, pushEv.StoreId)
				if authErr != nil {
					return authErr
				}
				authChecked = true
			}
			if pushEv.TenantSlug != "" {
				startSubscription(pushEv.TenantSlug)
			}

			if authStoreID > 0 && pushEv.StoreId > 0 && pushEv.StoreId != authStoreID {
				ack := &syncpb.SyncAck{
					Id:           pushEv.Id,
					BatchId:      pushEv.BatchId,
					OutboxUuid:   pushEv.OutboxUuid,
					EntityType:   pushEv.EntityType,
					EntityId:     pushEv.EntityId,
					EntityKey:    syncpb.ResolveEntityKey(pushEv.EntityKey, pushEv.EntityId),
					Success:      false,
					ErrorMessage: fmt.Sprintf("store_id mismatch: event store %d does not match authenticated terminal store %d", pushEv.StoreId, authStoreID),
					ProcessedAt:  timestamppb.Now(),
				}
				if err := sendServerMsg(&syncpb.SessionServerMsg{
					Msg: &syncpb.SessionServerMsg_PushAck{PushAck: ack},
				}); err != nil {
					return err
				}
				currentBatchID = ""
				currentBatch = nil
				continue
			}

			// Single-event mode when batch_id is empty
			if strings.TrimSpace(pushEv.BatchId) == "" || pushEv.BatchSize <= 0 {
				ack := s.handleSinglePushEventWithAuth(ctx, pushEv, authTerminalID, authStoreID)
				if err := sendServerMsg(&syncpb.SessionServerMsg{
					Msg: &syncpb.SessionServerMsg_PushAck{PushAck: ack},
				}); err != nil {
					return err
				}
				continue
			}

			// Correlation batch mode
			if currentBatchID != "" && pushEv.BatchId != currentBatchID {
				ack := &syncpb.SyncAck{
					Id:           pushEv.Id,
					BatchId:      currentBatchID,
					EntityKey:    syncpb.ResolveEntityKey(pushEv.EntityKey, pushEv.EntityId),
					Success:      false,
					ErrorMessage: fmt.Sprintf("incomplete batch %s before new batch %s", currentBatchID, pushEv.BatchId),
					ProcessedAt:  timestamppb.Now(),
				}
				if err := sendServerMsg(&syncpb.SessionServerMsg{
					Msg: &syncpb.SessionServerMsg_PushAck{PushAck: ack},
				}); err != nil {
					return err
				}
				currentBatchID = ""
				currentBatch = nil
			}

			if currentBatchID == "" {
				currentBatchID = pushEv.BatchId
				currentBatch = make([]*syncpb.SyncEvent, 0, int(pushEv.BatchSize))
			}
			currentBatch = append(currentBatch, pushEv)

			if int32(len(currentBatch)) >= pushEv.BatchSize {
				batchToApply := currentBatch
				currentBatchID = ""
				currentBatch = nil

				ack := s.processPushBatchWithAuth(ctx, batchToApply, authTerminalID, authStoreID)
				if err := sendServerMsg(&syncpb.SessionServerMsg{
					Msg: &syncpb.SessionServerMsg_PushAck{PushAck: ack},
				}); err != nil {
					return err
				}
			}
		} else if pullAckSeq := clientMsg.GetPullAckSeq(); pullAckSeq > 0 {
			if pullAckSeq > lastAckedSeq {
				lastAckedSeq = pullAckSeq
				log.Printf("[gRPC SyncServer] Terminal %s (tenant=%s store=%d) acknowledged change_log seq %d",
					deviceID, tenantSlug, storeID, pullAckSeq)
			}
		}
	}
}
