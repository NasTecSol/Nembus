package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SyncService struct {
	ctx        context.Context
	masterPool *pgxpool.Pool
	cloudURL   string
	tenantSlug string
	httpClient *http.Client
}

type OutboxItem struct {
	ID         int64          `json:"id"`
	EntityType string         `json:"entity_type"`
	EntityID   int64          `json:"entity_id"`
	Action     string         `json:"action"`
	Payload    json.RawMessage `json:"payload"`
	CreatedAt  time.Time      `json:"created_at"`
}

type PushSyncRequest struct {
	StoreID int32        `json:"store_id"`
	Items   []OutboxItem `json:"items"`
}

type PushSyncResponse struct {
	StatusCode int      `json:"statusCode"`
	Message    string   `json:"message"`
	Data       struct {
		SyncedIDs []int64 `json:"synced_ids"`
		Count     int     `json:"count"`
	} `json:"data"`
}

type DeltaConfig struct {
	ID             int32     `json:"id"`
	OrganizationID int32     `json:"organization_id"`
	StoreID        *int32    `json:"store_id"`
	PosTerminalID  *int32    `json:"pos_terminal_id"`
	DeviceSerial   string    `json:"device_serial"`
	DeviceType     string    `json:"device_type"`
	CsrPem         *string   `json:"csr_pem"`
	PrivateKeyPem  string    `json:"private_key_pem"`
	ComplianceCsid *string   `json:"compliance_csid"`
	ProductionCsid *string   `json:"production_csid"`
	CsidExpiry     *time.Time`json:"csid_expiry"`
	ZatcaEnv       *string   `json:"zatca_env"`
	IsActive       bool      `json:"is_active"`
	IsRevoked      bool      `json:"is_revoked"`
	RevokedAt      *time.Time`json:"revoked_at"`
	RevokedReason  *string   `json:"revoked_reason"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type DeltaConfigResponse struct {
	StatusCode int           `json:"statusCode"`
	Message    string        `json:"message"`
	Data       []DeltaConfig `json:"data"`
}

func NewSyncService(ctx context.Context, pool *pgxpool.Pool, cloudURL, slug string) *SyncService {
	return &SyncService{
		ctx:        ctx,
		masterPool: pool,
		cloudURL:   cloudURL,
		tenantSlug: slug,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
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
	log.Printf("🔄 Sync Service started for tenant [%s] (push outbox + pull delta)", s.tenantSlug)
}

func (s *SyncService) performSync() {
	if s.masterPool == nil {
		return
	}

	// 1. Transactional Outbox Push (POS -> Cloud)
	s.drainOutbox()

	// 2. Delta-Fetch Pull (Cloud -> POS)
	s.fetchDelta()
}

// drainOutbox implements the Transactional Outbox Pattern (Push).
// Selects pending items from local sync_queue and transmits them upstream to Cloud.
func (s *SyncService) drainOutbox() {
	rows, err := s.masterPool.Query(s.ctx, `
		SELECT id, entity_type, entity_id, action, payload, created_at
		FROM sync_queue
		WHERE status = 'pending'
		ORDER BY priority DESC, created_at ASC
		LIMIT 50;
	`)
	if err != nil {
		log.Printf("[SYNC] Failed to query sync_queue: %v", err)
		return
	}
	defer rows.Close()

	var items []OutboxItem
	for rows.Next() {
		var item OutboxItem
		if err := rows.Scan(&item.ID, &item.EntityType, &item.EntityID, &item.Action, &item.Payload, &item.CreatedAt); err != nil {
			continue
		}
		items = append(items, item)
	}

	if len(items) == 0 {
		return
	}

	// Get local store_id
	var storeID int32
	_ = s.masterPool.QueryRow(s.ctx, `SELECT store_id FROM local_device_config LIMIT 1;`).Scan(&storeID)

	reqPayload := PushSyncRequest{
		StoreID: storeID,
		Items:   items,
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return
	}

	url := fmt.Sprintf("%s/api/zatca/sync/push", s.cloudURL)
	req, err := http.NewRequestWithContext(s.ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if s.tenantSlug != "" {
		req.Header.Set("x-tenant-id", s.tenantSlug)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.recordOutboxFailure(items, err.Error())
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		s.recordOutboxFailure(items, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)))
		return
	}

	var pushRes PushSyncResponse
	if err := json.Unmarshal(respBody, &pushRes); err != nil {
		s.recordOutboxFailure(items, err.Error())
		return
	}

	// Mark acknowledged items as synced
	if len(pushRes.Data.SyncedIDs) > 0 {
		_, _ = s.masterPool.Exec(s.ctx, `
			UPDATE sync_queue
			SET status = 'synced', synced_at = NOW()
			WHERE id = ANY($1);
		`, pushRes.Data.SyncedIDs)
		log.Printf("[SYNC] Pushed %d outbox items to Cloud", len(pushRes.Data.SyncedIDs))
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

// fetchDelta implements the Delta-Fetch Mechanism (Pull).
// Queries Cloud API for updated ZATCA configs since last sync timestamp and upserts them locally.
func (s *SyncService) fetchDelta() {
	var storeID int32
	var lastSyncAt time.Time

	err := s.masterPool.QueryRow(s.ctx, `
		SELECT store_id, last_zatca_sync_at
		FROM local_device_config
		LIMIT 1;
	`).Scan(&storeID, &lastSyncAt)
	if err != nil {
		return
	}

	url := fmt.Sprintf("%s/api/zatca/configs?store_id=%d&since=%s",
		s.cloudURL, storeID, lastSyncAt.Format(time.RFC3339))

	req, err := http.NewRequestWithContext(s.ctx, "GET", url, nil)
	if err != nil {
		return
	}

	if s.tenantSlug != "" {
		req.Header.Set("x-tenant-id", s.tenantSlug)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var deltaRes DeltaConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&deltaRes); err != nil {
		return
	}

	if len(deltaRes.Data) == 0 {
		return
	}

	now := time.Now()
	for _, cfg := range deltaRes.Data {
		s.upsertZatcaConfigLocally(cfg)
	}

	// Update local watermark
	_, _ = s.masterPool.Exec(s.ctx, `
		UPDATE local_device_config
		SET last_zatca_sync_at = $1;
	`, now)

	log.Printf("[SYNC] Pulled %d ZATCA config updates from Cloud", len(deltaRes.Data))
}

func (s *SyncService) upsertZatcaConfigLocally(cfg DeltaConfig) {
	query := `
		INSERT INTO zatca_device_configs (
			id, organization_id, store_id, pos_terminal_id, device_serial, device_type,
			csr_pem, private_key_pem, compliance_csid, production_csid, csid_expiry,
			zatca_env, is_active, is_revoked, revoked_at, revoked_reason, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
		ON CONFLICT (organization_id, device_serial) DO UPDATE SET
			production_csid = EXCLUDED.production_csid,
			csid_expiry     = EXCLUDED.csid_expiry,
			is_active       = EXCLUDED.is_active,
			is_revoked      = EXCLUDED.is_revoked,
			revoked_at      = EXCLUDED.revoked_at,
			revoked_reason  = EXCLUDED.revoked_reason,
			updated_at      = EXCLUDED.updated_at;
	`
	_, _ = s.masterPool.Exec(s.ctx, query,
		cfg.ID, cfg.OrganizationID, cfg.StoreID, cfg.PosTerminalID, cfg.DeviceSerial, cfg.DeviceType,
		cfg.CsrPem, cfg.PrivateKeyPem, cfg.ComplianceCsid, cfg.ProductionCsid, cfg.CsidExpiry,
		cfg.ZatcaEnv, cfg.IsActive, cfg.IsRevoked, cfg.RevokedAt, cfg.RevokedReason, cfg.UpdatedAt,
	)
}
