package sync

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/grpc/syncpb"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Every entity has its own source timestamp; the legacy ZATCA timestamp is
// deliberately not imported because older fetches advanced it on failed writes.
func (s *SyncService) fetchDeltaGRPC() {
	var storeID int32
	if err := s.masterPool.QueryRow(s.ctx, `SELECT store_id FROM local_device_config LIMIT 1`).Scan(&storeID); err != nil || storeID <= 0 {
		log.Printf("[gRPC SYNC] Delta fetch requires a valid local store: %v", err)
		return
	}
	conn, err := grpc.NewClient(s.grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[gRPC SYNC] Delta connection error: %v", err)
		return
	}
	defer conn.Close()
	client := syncpb.NewSyncServiceClient(conn)
	for _, entity := range syncpb.PullEntityTypes() {
		if err := s.pullEntity(client, storeID, entity); err != nil {
			log.Printf("[gRPC SYNC] Delta fetch failed for %s: %v", entity, err)
			metadata, _ := json.Marshal(map[string]any{"direction": "pull", "status": "error", "last_error": err.Error(), "last_attempt_at": time.Now().UTC()})
			_, recordErr := s.masterPool.Exec(s.ctx, `INSERT INTO sync_watermarks(entity_type,store_id,metadata)
				VALUES($1,$2,$3::jsonb) ON CONFLICT(entity_type,store_id) DO UPDATE
				SET metadata=COALESCE(sync_watermarks.metadata,'{}'::jsonb)||EXCLUDED.metadata`, entity, storeID, string(metadata))
			if recordErr != nil {
				log.Printf("[gRPC SYNC] Could not record %s failure: %v", entity, recordErr)
			}
		}
	}
}

func (s *SyncService) pullEntity(client syncpb.SyncServiceClient, storeID int32, entity string) error {
	since := time.Unix(0, 0).UTC()
	err := s.masterPool.QueryRow(s.ctx, `SELECT last_sync_at FROM sync_watermarks WHERE entity_type=$1 AND store_id=$2`, entity, storeID).Scan(&since)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("read watermark: %w", err)
	}
	total := 0
	for {
		ctx, cancel := context.WithTimeout(s.ctx, 60*time.Second)
		stream, err := client.StreamPull(ctx, &syncpb.PullRequest{TenantSlug: s.tenantSlug, StoreId: storeID,
			Since: timestamppb.New(since), EntityTypes: []string{entity}, Limit: 200})
		if err != nil {
			cancel()
			return err
		}
		var events []*syncpb.SyncEvent
		next := since
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
			if event.EntityType != entity || event.Sha256 != hex.EncodeToString(hash[:]) || event.EventTime == nil || event.EventTime.CheckValid() != nil || !event.EventTime.AsTime().After(since) {
				cancel()
				return fmt.Errorf("invalid delta payload, checksum or timestamp for %s", entity)
			}
			if event.EventTime.AsTime().After(next) {
				next = event.EventTime.AsTime()
			}
			events = append(events, event)
		}
		cancel()
		if err := s.saveDeltaPage(storeID, entity, next, events, total); err != nil {
			return err
		}
		total += len(events)
		if len(events) == 0 {
			log.Printf("[gRPC SYNC] %s: applied %d updates; watermark %s", entity, total, next.Format(time.RFC3339Nano))
			return nil
		}
		since = next
	}
}

// Data and its checkpoint commit together. Failed pages remain retryable.
func (s *SyncService) saveDeltaPage(storeID int32, entity string, since time.Time, events []*syncpb.SyncEvent, previousCount int) error {
	tx, err := s.masterPool.Begin(s.ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(s.ctx)
	if len(events) > 0 {
		query, err := deltaUpsertSQL(s.ctx, tx, entity)
		if err != nil {
			return err
		}
		for _, event := range events {
			if _, err := tx.Exec(s.ctx, query, string(event.PayloadJson)); err != nil {
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
	}
	data, _ := json.Marshal(metadata)
	_, err = tx.Exec(s.ctx, `INSERT INTO sync_watermarks(entity_type,store_id,last_sync_at,metadata)
		VALUES($1,$2,$3,$4::jsonb) ON CONFLICT(entity_type,store_id) DO UPDATE SET
		last_sync_at=EXCLUDED.last_sync_at,
		metadata=COALESCE(sync_watermarks.metadata,'{}'::jsonb)||EXCLUDED.metadata`, entity, storeID, since, string(data))
	if err != nil {
		return err
	}
	return tx.Commit(s.ctx)
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
	return fmt.Sprintf(`INSERT INTO %s (%s) SELECT %s FROM json_populate_record(NULL::%s,$1::json)
		ON CONFLICT(id) DO UPDATE SET %s`, table, cols, cols, table, strings.Join(updates, ",")), nil
}
