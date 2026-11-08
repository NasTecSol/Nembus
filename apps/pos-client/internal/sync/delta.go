package sync

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
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	maxDeltaPagesPerEntity = 5
	maxChangePagesPerCycle = 20
	ChangeLogWatermarkKey  = "__change_log__"
)

var errStreamChangesUnimplemented = errors.New("StreamChanges RPC is not implemented by cloud server")

// fetchChangesGRPC pulls monotonic change_log events via StreamChanges (Phase 3),
// falling back to legacy fetchDeltaGRPC only if the cloud server returns codes.Unimplemented.
func (s *SyncService) fetchChangesGRPC() {
	configCtx, cancelConfig := context.WithTimeout(s.ctx, 15*time.Second)
	storeID, deviceID, syncToken, err := ReadLocalDeviceCredentials(configCtx, s.masterPool)
	cancelConfig()
	if err != nil || storeID <= 0 {
		log.Printf("[gRPC SYNC] Change-log pull requires a valid local store: %v", err)
		return
	}

	conn, err := s.getOrCreateConn()
	if err != nil {
		log.Printf("[gRPC SYNC] Change-log connection error: %v", err)
		return
	}

	client := syncpb.NewSyncServiceClient(conn)
	if err := s.pullChangeLog(client, storeID, deviceID, syncToken); err != nil {
		if errors.Is(err, errStreamChangesUnimplemented) {
			log.Printf("[gRPC SYNC] Cloud server does not implement StreamChanges yet; falling back to StreamPull")
			s.fetchDeltaGRPC()
			return
		}
		if s.ctx.Err() != nil {
			return
		}
		log.Printf("[gRPC SYNC] Change-log pull failed: %v", err)
		metaBytes, _ := json.Marshal(map[string]any{
			"direction":       "pull",
			"status":          "error",
			"last_error":      err.Error(),
			"last_attempt_at": time.Now().UTC(),
		})
		recordCtx, cancelRecord := context.WithTimeout(s.ctx, 15*time.Second)
		_, _ = s.masterPool.Exec(recordCtx, `
			INSERT INTO sync_watermarks(entity_type, store_id, metadata)
			VALUES ($1, $2, $3::jsonb)
			ON CONFLICT (entity_type, store_id) DO UPDATE
			SET metadata = COALESCE(sync_watermarks.metadata, '{}'::jsonb) || EXCLUDED.metadata
		`, ChangeLogWatermarkKey, storeID, string(metaBytes))
		cancelRecord()
	}
}

func (s *SyncService) pullChangeLog(client syncpb.SyncServiceClient, storeID int32, deviceID, syncToken string) error {
	readCtx, cancelRead := context.WithTimeout(s.ctx, 15*time.Second)
	var fromSeq int64
	err := s.masterPool.QueryRow(readCtx, `
		SELECT COALESCE((metadata->>'last_seq')::bigint, 0)
		FROM sync_watermarks
		WHERE entity_type = $1 AND store_id = $2
	`, ChangeLogWatermarkKey, storeID).Scan(&fromSeq)
	cancelRead()
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("read change_log watermark: %w", err)
	}
	s.lastAppliedSeq.Store(fromSeq)

	totalApplied := 0
	expiredRetried := false

	for page := 0; page < maxChangePagesPerCycle; page++ {
		if s.ctx.Err() != nil {
			return s.ctx.Err()
		}
		ctx, cancel := context.WithTimeout(s.ctx, 60*time.Second)
		ctx = s.withSyncMetadata(ctx, storeID, deviceID, syncToken)

		stream, err := client.StreamChanges(ctx, &syncpb.StreamChangesRequest{
			TenantSlug: s.tenantSlug,
			StoreId:    storeID,
			FromSeq:    fromSeq,
			Limit:      syncpb.PullPageLimit,
			DeviceId:   deviceID,
		})
		if err != nil {
			cancel()
			if status.Code(err) == codes.Unimplemented {
				return errStreamChangesUnimplemented
			}
			if (status.Code(err) == codes.OutOfRange || strings.Contains(err.Error(), "CURSOR_EXPIRED")) && !expiredRetried {
				expiredRetried = true
				log.Printf("[gRPC SYNC] Change-log cursor %d expired on Cloud; resetting cursor to 0 for snapshot catchup", fromSeq)
				if resetErr := s.resetChangeLogWatermark(storeID); resetErr != nil {
					return resetErr
				}
				fromSeq = 0
				continue
			}
			return err
		}

		var (
			events         []*syncpb.ChangeEvent
			nextSeq        = fromSeq
			isSnapshotMode bool
			snapshotMaxSeq int64
		)

		for {
			ev, recvErr := stream.Recv()
			if recvErr == io.EOF {
				break
			}
			if recvErr != nil {
				cancel()
				if status.Code(recvErr) == codes.Unimplemented {
					return errStreamChangesUnimplemented
				}
				if (status.Code(recvErr) == codes.OutOfRange || strings.Contains(recvErr.Error(), "CURSOR_EXPIRED")) && !expiredRetried {
					expiredRetried = true
					log.Printf("[gRPC SYNC] Change-log cursor %d expired mid-stream; resetting cursor to 0 for snapshot catchup", fromSeq)
					if resetErr := s.resetChangeLogWatermark(storeID); resetErr != nil {
						return resetErr
					}
					fromSeq = 0
					events = nil
					break
				}
				return recvErr
			}

			if !syncpb.IsPullEntity(ev.EntityType) {
				cancel()
				return fmt.Errorf("unsupported change_log entity: %s", ev.EntityType)
			}
			if ev.Op != "DELETE" {
				hash := sha256.Sum256(ev.PayloadJson)
				if ev.Sha256 != "" && ev.Sha256 != hex.EncodeToString(hash[:]) {
					cancel()
					return fmt.Errorf("change_log SHA-256 mismatch for %s ID %d (seq %d)", ev.EntityType, ev.EntityId, ev.Seq)
				}
			}
			if ev.IsSnapshotMode {
				isSnapshotMode = true
				if ev.SnapshotMaxSeq > snapshotMaxSeq {
					snapshotMaxSeq = ev.SnapshotMaxSeq
				}
				if ev.Seq > nextSeq {
					nextSeq = ev.Seq
				}
			} else {
				if ev.Seq <= nextSeq {
					cancel()
					return fmt.Errorf("non-monotonic change_log seq %d (previous %d)", ev.Seq, nextSeq)
				}
				nextSeq = ev.Seq
			}
			events = append(events, ev)
		}
		cancel()

		if expiredRetried && len(events) == 0 && fromSeq == 0 {
			continue
		}

		if isSnapshotMode && snapshotMaxSeq > nextSeq {
			nextSeq = snapshotMaxSeq
		}

		if err := s.saveChangePage(storeID, nextSeq, events, totalApplied); err != nil {
			return err
		}
		s.lastAppliedSeq.Store(nextSeq)
		totalApplied += len(events)

		if len(events) == 0 || isSnapshotMode || len(events) < syncpb.PullPageLimit {
			if totalApplied > 0 {
				log.Printf("[gRPC SYNC] Change-log pull applied %d events; seq watermark = %d", totalApplied, nextSeq)
			}
			return nil
		}
		fromSeq = nextSeq
	}

	log.Printf("[gRPC SYNC] Change-log pull applied %d events; remaining pages resume next cycle (seq=%d)", totalApplied, fromSeq)
	return nil
}

func (s *SyncService) resetChangeLogWatermark(storeID int32) error {
	ctx, cancel := context.WithTimeout(s.ctx, 15*time.Second)
	defer cancel()
	metaBytes, _ := json.Marshal(map[string]any{
		"direction":       "pull",
		"status":          "cursor_expired_reset",
		"last_seq":        0,
		"last_attempt_at": time.Now().UTC(),
	})
	_, err := s.masterPool.Exec(ctx, `
		INSERT INTO sync_watermarks(entity_type, store_id, last_sync_at, metadata)
		VALUES ($1, $2, NOW(), $3::jsonb)
		ON CONFLICT (entity_type, store_id) DO UPDATE SET
			last_sync_at = NOW(),
			metadata = COALESCE(sync_watermarks.metadata, '{}'::jsonb) || EXCLUDED.metadata
	`, ChangeLogWatermarkKey, storeID, string(metaBytes))
	return err
}

func (s *SyncService) saveChangePage(storeID int32, nextSeq int64, events []*syncpb.ChangeEvent, previousCount int) error {
	ctx, cancel := context.WithTimeout(s.ctx, 60*time.Second)
	defer cancel()
	return s.saveChangePageContext(ctx, storeID, nextSeq, events, previousCount)
}

func (s *SyncService) saveChangePageContext(ctx context.Context, storeID int32, nextSeq int64, events []*syncpb.ChangeEvent, previousCount int) error {
	tx, err := s.masterPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(rollbackCtx)
	}()

	if len(events) > 0 {
		if _, err := tx.Exec(ctx, "SET LOCAL nembus.sync_applier = 'true'"); err != nil {
			return err
		}
		upsertQueryCache := make(map[string]string)

		for _, ev := range events {
			if strings.EqualFold(ev.Op, "DELETE") {
				if err := applyLeafDeleteTombstone(ctx, tx, ev.EntityType, ev.EntityId); err != nil {
					return err
				}
				continue
			}

			query, ok := upsertQueryCache[ev.EntityType]
			if !ok {
				var qErr error
				query, qErr = deltaUpsertSQL(ctx, tx, ev.EntityType)
				if qErr != nil {
					return qErr
				}
				upsertQueryCache[ev.EntityType] = query
			}
			if _, err := tx.Exec(ctx, query, string(ev.PayloadJson)); err != nil {
				return fmt.Errorf("apply change_log %s ID %d (seq %d): %w", ev.EntityType, ev.EntityId, ev.Seq, err)
			}
		}
	}

	meta := map[string]any{
		"direction":       "pull",
		"status":          "syncing",
		"last_error":      nil,
		"last_seq":        nextSeq,
		"last_attempt_at": time.Now().UTC(),
		"records_applied": previousCount + len(events),
	}
	if len(events) == 0 || len(events) < syncpb.PullPageLimit {
		meta["status"] = "synced"
		meta["last_success_at"] = time.Now().UTC()
	}
	metaJSON, _ := json.Marshal(meta)

	_, err = tx.Exec(ctx, `
		INSERT INTO sync_watermarks(entity_type, store_id, last_sync_at, metadata)
		VALUES ($1, $2, NOW(), $3::jsonb)
		ON CONFLICT (entity_type, store_id) DO UPDATE SET
			last_sync_at = NOW(),
			metadata = COALESCE(sync_watermarks.metadata, '{}'::jsonb) || EXCLUDED.metadata
	`, ChangeLogWatermarkKey, storeID, string(metaJSON))
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func applyLeafDeleteTombstone(ctx context.Context, tx pgx.Tx, entityType string, entityID int64) error {
	if !syncpb.IsLeafDeleteEntity(entityType) {
		log.Printf("[gRPC SYNC] Ignoring hard DELETE tombstone for non-leaf table %s (id=%d)", entityType, entityID)
		return nil
	}
	if _, err := tx.Exec(ctx, "SAVEPOINT sp_change_delete"); err != nil {
		return err
	}
	table := pgx.Identifier{entityType}.Sanitize()
	_, err := tx.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE id = $1", table), entityID)
	if err != nil {
		_, _ = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT sp_change_delete")
		_, _ = tx.Exec(ctx, "RELEASE SAVEPOINT sp_change_delete")
		log.Printf("[gRPC SYNC] Skipped DELETE tombstone for %s ID %d due to local reference: %v", entityType, entityID, err)
		return nil
	}
	_, _ = tx.Exec(ctx, "RELEASE SAVEPOINT sp_change_delete")
	return nil
}

// Every entity has its own source timestamp; the legacy ZATCA timestamp is
// deliberately not imported because older fetches advanced it on failed writes.
func (s *SyncService) fetchDeltaGRPC() {
	var storeID int32
	configCtx, cancelConfig := context.WithTimeout(s.ctx, 15*time.Second)
	defer cancelConfig()
	if err := s.masterPool.QueryRow(configCtx, `SELECT store_id FROM local_device_config ORDER BY id LIMIT 1`).Scan(&storeID); err != nil || storeID <= 0 {
		log.Printf("[gRPC SYNC] Delta fetch requires a valid local store: %v", err)
		return
	}
	conn, err := s.getOrCreateConn()
	if err != nil {
		log.Printf("[gRPC SYNC] Delta connection error: %v", err)
		return
	}
	client := syncpb.NewSyncServiceClient(conn)
	for _, entity := range syncpb.PullEntityTypes() {
		if s.ctx.Err() != nil {
			return
		}
		if err := s.pullEntity(client, storeID, entity); err != nil {
			if s.ctx.Err() != nil {
				return
			}
			log.Printf("[gRPC SYNC] Delta fetch failed for %s: %v", entity, err)
			metadata, _ := json.Marshal(map[string]any{"direction": "pull", "status": "error", "last_error": err.Error(), "last_attempt_at": time.Now().UTC()})
			recordCtx, cancelRecord := context.WithTimeout(s.ctx, 15*time.Second)
			_, recordErr := s.masterPool.Exec(recordCtx, `INSERT INTO sync_watermarks(entity_type,store_id,metadata)
				VALUES($1,$2,$3::jsonb) ON CONFLICT(entity_type,store_id) DO UPDATE
				SET metadata=COALESCE(sync_watermarks.metadata,'{}'::jsonb)||EXCLUDED.metadata`, entity, storeID, string(metadata))
			cancelRecord()
			if recordErr != nil {
				log.Printf("[gRPC SYNC] Could not record %s failure: %v", entity, recordErr)
			}
		}
	}
}

func (s *SyncService) pullEntity(client syncpb.SyncServiceClient, storeID int32, entity string) error {
	log.Printf("[gRPC SYNC] Checking %s for tenant %s, store %d", entity, s.tenantSlug, storeID)
	readCtx, cancelRead := context.WithTimeout(s.ctx, 15*time.Second)
	defer cancelRead()
	since := time.Unix(0, 0).UTC()
	var afterID int64
	err := s.masterPool.QueryRow(readCtx, `SELECT last_sync_at, COALESCE((metadata->>'cursor_id')::bigint,0)
		FROM sync_watermarks WHERE entity_type=$1 AND store_id=$2`, entity, storeID).Scan(&since, &afterID)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("read watermark: %w", err)
	}
	total := 0
	for page := 0; page < maxDeltaPagesPerEntity; page++ {
		ctx, cancel := context.WithTimeout(s.ctx, 60*time.Second)
		ctx = metadata.AppendToOutgoingContext(ctx, syncpb.PullCursorKey, strconv.FormatInt(afterID, 10))
		stream, err := client.StreamPull(ctx, &syncpb.PullRequest{TenantSlug: s.tenantSlug, StoreId: storeID,
			Since: timestamppb.New(since), EntityTypes: []string{entity}, Limit: syncpb.PullPageLimit})
		if err != nil {
			cancel()
			return err
		}
		headers, err := stream.Header()
		if err != nil {
			cancel()
			return err
		}
		if versions := headers.Get(syncpb.PullCursorVersionKey); len(versions) != 1 || versions[0] != "1" {
			cancel()
			return fmt.Errorf("cloud server must be updated to support bounded delta pagination")
		}
		var events []*syncpb.SyncEvent
		next := since
		nextID := afterID
		for {
			event, recvErr := stream.Recv()
			if recvErr == io.EOF {
				break
			}
			if recvErr != nil {
				cancel()
				return recvErr
			}
			hash := sha256.Sum256(event.PayloadJson)
			if len(events) >= syncpb.PullPageLimit || event.EntityType != entity || event.Sha256 != hex.EncodeToString(hash[:]) || event.EventTime == nil || event.EventTime.CheckValid() != nil || !cursorAfter(event.EventTime.AsTime(), event.EntityId, next, nextID) {
				cancel()
				return fmt.Errorf("invalid delta payload, checksum or timestamp for %s", entity)
			}
			next = event.EventTime.AsTime()
			nextID = event.EntityId
			events = append(events, event)
		}
		cancel()
		log.Printf("[gRPC SYNC] Applying %s page %d (%d rows)", entity, page+1, len(events))
		if err := s.saveDeltaPage(storeID, entity, next, events, total); err != nil {
			return err
		}
		total += len(events)
		if len(events) == 0 {
			log.Printf("[gRPC SYNC] %s: applied %d updates; watermark %s", entity, total, next.Format(time.RFC3339Nano))
			return nil
		}
		since = next
		afterID = nextID
	}
	log.Printf("[gRPC SYNC] %s: applied %d updates; remaining pages resume next cycle", entity, total)
	return nil
}

func cursorAfter(timestamp time.Time, id int64, since time.Time, afterID int64) bool {
	return timestamp.After(since) || (timestamp.Equal(since) && id > afterID)
}

// Data and its checkpoint commit together. Failed pages remain retryable.
func (s *SyncService) saveDeltaPage(storeID int32, entity string, since time.Time, events []*syncpb.SyncEvent, previousCount int) error {
	ctx, cancel := context.WithTimeout(s.ctx, 60*time.Second)
	defer cancel()
	return s.saveDeltaPageContext(ctx, storeID, entity, since, events, previousCount)
}

func (s *SyncService) saveDeltaPageContext(ctx context.Context, storeID int32, entity string, since time.Time, events []*syncpb.SyncEvent, previousCount int) error {
	tx, err := s.masterPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(rollbackCtx)
	}()
	if len(events) > 0 {
		if _, err := tx.Exec(ctx, "SET LOCAL nembus.sync_applier = 'true'"); err != nil {
			return err
		}
		query, err := deltaUpsertSQL(ctx, tx, entity)
		if err != nil {
			return err
		}
		for _, event := range events {
			if _, err := tx.Exec(ctx, query, string(event.PayloadJson)); err != nil {
				return fmt.Errorf("apply %s ID %d: %w", entity, event.EntityId, err)
			}
		}
	}
	metadata := map[string]any{"direction": "pull", "status": "syncing", "last_error": nil,
		"last_attempt_at": time.Now().UTC(), "records_applied": previousCount + len(events)}
	if len(events) == 0 {
		metadata["status"] = "synced"
		metadata["last_success_at"] = time.Now().UTC()
	} else {
		metadata["last_entity_id"] = events[len(events)-1].EntityId
		metadata["cursor_id"] = events[len(events)-1].EntityId
	}
	data, _ := json.Marshal(metadata)
	_, err = tx.Exec(ctx, `INSERT INTO sync_watermarks(entity_type,store_id,last_sync_at,metadata)
		VALUES($1,$2,$3,$4::jsonb) ON CONFLICT(entity_type,store_id) DO UPDATE SET
		last_sync_at=EXCLUDED.last_sync_at,
		metadata=COALESCE(sync_watermarks.metadata,'{}'::jsonb)||EXCLUDED.metadata`, entity, storeID, since, string(data))
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func deltaUpsertSQL(ctx context.Context, tx pgx.Tx, entity string) (string, error) {
	if !syncpb.IsPullEntity(entity) {
		return "", fmt.Errorf("unsupported pull entity: %s", entity)
	}
	rows, err := tx.Query(ctx, `SELECT column_name FROM information_schema.columns
		WHERE table_schema=current_schema() AND table_name=$1 AND is_generated='NEVER'
		ORDER BY ordinal_position`, entity)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var columns, updates []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			return "", err
		}
		quoted := pgx.Identifier{column}.Sanitize()
		if entity == "user_biometric_templates" && column == "id" {
			continue
		}
		columns = append(columns, quoted)
		if column != "id" && column != "created_at" {
			updates = append(updates, quoted+"=EXCLUDED."+quoted)
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if len(columns) == 0 || len(updates) == 0 {
		return "", fmt.Errorf("no writable columns for %s", entity)
	}
	table := pgx.Identifier{entity}.Sanitize()
	cols := strings.Join(columns, ",")
	conflictTarget := "id"
	if entity == "user_biometric_templates" {
		conflictTarget = "user_id, finger_index"
		return fmt.Sprintf(`INSERT INTO %s (%s) SELECT %s FROM json_populate_record(NULL::%s,$1::json)
		ON CONFLICT(%s) DO UPDATE SET %s
		WHERE EXCLUDED.updated_at >= %s.updated_at`, table, cols, cols, table, conflictTarget, strings.Join(updates, ","), table), nil
	}
	if entity == "customers" {
		// A cloud delta must not replace a locally created customer whose
		// outbox item has not been acknowledged yet (including quarantined/failed items).
		return fmt.Sprintf(`INSERT INTO %s (%s) SELECT %s FROM json_populate_record(NULL::%s,$1::json)
		ON CONFLICT(id) DO UPDATE SET %s
		WHERE NOT EXISTS (SELECT 1 FROM sync_queue q WHERE q.entity_type='customers'
		AND q.entity_id=%s.id AND q.status IN ('pending','failed','quarantined'))`,
			table, cols, cols, table, strings.Join(updates, ","), table), nil
	}
	return fmt.Sprintf(`INSERT INTO %s (%s) SELECT %s FROM json_populate_record(NULL::%s,$1::json)
		ON CONFLICT(%s) DO UPDATE SET %s`, table, cols, cols, table, conflictTarget, strings.Join(updates, ",")), nil
}
