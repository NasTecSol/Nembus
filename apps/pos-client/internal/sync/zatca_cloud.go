package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5"
)

var ErrZatcaBackupMissing = errors.New("no cloud ZATCA registration exists for this device")

type ZatcaCloud struct {
	baseURL, tenant, token string
	http                   *http.Client
}

func NewZatcaCloud(ctx context.Context, baseURL, tenant string) (*ZatcaCloud, error) {
	u, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("valid ZATCA cloud HTTPS URL is required")
	}
	loopback := u.Hostname() == "localhost" || net.ParseIP(u.Hostname()).IsLoopback()
	if u.Scheme != "https" && !(u.Scheme == "http" && loopback) {
		return nil, fmt.Errorf("ZATCA credentials require HTTPS; plaintext gRPC/remote HTTP is not allowed")
	}
	if tenant == "" {
		return nil, fmt.Errorf("tenant is required for ZATCA cloud sync")
	}
	c := &ZatcaCloud{baseURL: u.String(), tenant: tenant, http: &http.Client{Timeout: 45 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
	clients, err := middleware.LoadM2MClients()
	if err != nil {
		return nil, fmt.Errorf("read local M2M registrations: %w", err)
	}
	for _, client := range clients {
		if client.TenantID == tenant && client.IsActive && client.Token != "" {
			c.token = client.Token
			break
		}
	}
	if c.token == "" {
		if err := c.refreshToken(ctx); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *ZatcaCloud) refreshToken(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/m2m/tokens", nil)
	if err != nil {
		return err
	}
	req.Header.Set("x-tenant-id", c.tenant)
	req.Header.Set("x-m2m-key", middleware.M2MAPIKey)
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("fetch cloud application registration: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("fetch cloud application registration: HTTP %d", res.StatusCode)
	}
	var clients []middleware.M2MClient
	if json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&clients) != nil {
		return fmt.Errorf("invalid cloud application registration response")
	}
	for _, client := range clients {
		if client.TenantID == c.tenant && client.IsActive && client.Token != "" {
			c.token = client.Token
			if err := middleware.SaveM2MClient(client); err != nil {
				return fmt.Errorf("save local application registration: %w", err)
			}
			return nil
		}
	}
	return fmt.Errorf("no active cloud application registration for selected tenant")
}

func (c *ZatcaCloud) request(ctx context.Context, method, path string, body []byte) (json.RawMessage, error) {
	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("x-tenant-id", c.tenant)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("cloud ZATCA request failed: %w", err)
		}
		var envelope struct {
			Message string          `json:"message"`
			Data    json.RawMessage `json:"data"`
		}
		decodeErr := json.NewDecoder(io.LimitReader(res.Body, 16<<20)).Decode(&envelope)
		res.Body.Close()
		if res.StatusCode == 401 && attempt == 0 {
			if err := c.refreshToken(ctx); err != nil {
				return nil, err
			}
			continue
		}
		if res.StatusCode == 404 && envelope.Message == "no saved ZATCA registration for this serial" {
			return nil, ErrZatcaBackupMissing
		}
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("cloud ZATCA request returned HTTP %d; ensure the cloud sync service is deployed and the device identity matches", res.StatusCode)
		}
		if decodeErr != nil {
			return nil, fmt.Errorf("invalid cloud ZATCA response")
		}
		return envelope.Data, nil
	}
	return nil, fmt.Errorf("cloud application registration was rejected")
}

func (c *ZatcaCloud) Push(ctx context.Context, item OutboxItem, snapshot repository.ZatcaSyncSnapshot) error {
	var device struct {
		StoreID int32 `json:"store_id"`
	}
	if json.Unmarshal(snapshot.Device, &device) != nil || device.StoreID <= 0 {
		return fmt.Errorf("ZATCA snapshot has no store")
	}
	data, err := json.Marshal(map[string]any{"store_id": device.StoreID, "items": []any{map[string]any{"id": item.ID, "entity_id": item.EntityID, "entity_type": item.EntityType, "action": item.Action, "payload": snapshot, "created_at": item.CreatedAt}}})
	if err != nil {
		return err
	}
	raw, err := c.request(ctx, "POST", "/api/zatca/sync/push", data)
	if err != nil {
		return err
	}
	var ack struct {
		Version int     `json:"zatca_sync_version"`
		IDs     []int64 `json:"synced_ids"`
	}
	if json.Unmarshal(raw, &ack) != nil || ack.Version != 1 || len(ack.IDs) != 1 || ack.IDs[0] != item.ID {
		return fmt.Errorf("cloud did not confirm durable ZATCA persistence; update the cloud server before retrying")
	}
	return nil
}

func (c *ZatcaCloud) RecoveryPage(ctx context.Context, serial string, organizationID int32, after, through int64) (repository.ZatcaRestorePage, error) {
	values := url.Values{"device_serial": {serial}, "organization_id": {strconv.Itoa(int(organizationID))}, "after_icv": {strconv.FormatInt(after, 10)}, "through_icv": {strconv.FormatInt(through, 10)}}
	raw, err := c.request(ctx, "GET", "/api/zatca/sync/restore?"+values.Encode(), nil)
	var page repository.ZatcaRestorePage
	if err != nil {
		return page, err
	}
	if json.Unmarshal(raw, &page) != nil || len(page.Device) == 0 || page.Documents == nil || page.NextICV < after || page.HeadICV < page.NextICV || (page.HasMore && page.NextICV == after) {
		return page, fmt.Errorf("invalid ZATCA recovery page")
	}
	return page, nil
}

func (c *ZatcaCloud) VerifyCapabilities(ctx context.Context) error {
	raw, err := c.request(ctx, "GET", "/api/zatca/sync/status", nil)
	if err != nil {
		return err
	}
	var capabilities struct {
		Version        int  `json:"zatca_sync_version"`
		SecureRecovery bool `json:"device_recovery_over_https"`
	}
	if json.Unmarshal(raw, &capabilities) != nil || capabilities.Version != 1 || !capabilities.SecureRecovery {
		return fmt.Errorf("update cloud server to support durable ZATCA sync and HTTPS recovery before cloning")
	}
	return nil
}

func (s *SyncService) pushZatcaItem(ctx context.Context, item OutboxItem, cloud *ZatcaCloud) error {
	var err error
	var snapshot repository.ZatcaSyncSnapshot
	if item.EntityType == "zatca_device_configs" {
		err = s.masterPool.QueryRow(ctx, `SELECT row_to_json(d) FROM zatca_device_configs d WHERE id=$1`, item.EntityID).Scan(&snapshot.Device)
	} else {
		err = s.masterPool.QueryRow(ctx, `SELECT row_to_json(d),row_to_json(c) FROM zatca_document_chain c JOIN zatca_device_configs d ON d.id=c.device_config_id WHERE c.id=$1`, item.EntityID).Scan(&snapshot.Device, &snapshot.Document)
	}
	if err != nil {
		return fmt.Errorf("read current ZATCA snapshot: %w", err)
	}
	return cloud.Push(ctx, item, snapshot)
}

// Re-query after each acknowledgment so a whole chain can drain in one cycle.
// A failed item is attempted once per cycle; children wait without using retries.
const pendingZatcaSQL = `SELECT q.id,q.entity_type,q.entity_id,q.action,q.payload,q.created_at,q.correlation_id
FROM sync_queue q WHERE q.status='pending' AND q.entity_type IN ('zatca_device_configs','zatca_document_chain') AND NOT(q.id=ANY($1::bigint[]))
AND NOT EXISTS(SELECT 1 FROM sync_queue parent WHERE parent.status IN ('pending','failed') AND q.entity_type='zatca_document_chain' AND (
 (parent.entity_type='zatca_device_configs' AND parent.entity_id=(q.payload->>'device_config_id')::bigint)
 OR (parent.entity_type='zatca_document_chain' AND parent.payload->>'device_config_id'=q.payload->>'device_config_id' AND (parent.payload->>'icv')::bigint<(q.payload->>'icv')::bigint)))
ORDER BY CASE WHEN q.entity_type='zatca_device_configs' THEN 0 ELSE 1 END,COALESCE((q.payload->>'device_config_id')::bigint,q.entity_id),COALESCE((q.payload->>'icv')::bigint,0),q.id LIMIT 1`

func (s *SyncService) drainZatcaOutbox() {
	if s.zatcaCloudURL == "" || s.masterPool == nil {
		return
	}
	ctx, cancel := context.WithTimeout(s.ctx, 60*time.Second)
	defer cancel()
	seen := make([]int64, 0, 50)
	var cloud *ZatcaCloud
	for len(seen) < 50 && ctx.Err() == nil {
		var item OutboxItem
		err := s.masterPool.QueryRow(ctx, pendingZatcaSQL, seen).Scan(&item.ID, &item.EntityType, &item.EntityID, &item.Action, &item.Payload, &item.CreatedAt, &item.CorrelationID)
		if errors.Is(err, pgx.ErrNoRows) {
			return
		}
		if err != nil {
			log.Printf("[ZATCA SYNC] Read queue failed: %v", err)
			return
		}
		seen = append(seen, item.ID)
		if cloud == nil {
			cloud, err = NewZatcaCloud(ctx, s.zatcaCloudURL, s.tenantSlug)
			if err != nil {
				s.recordOutboxFailure([]OutboxItem{item}, err.Error())
				return
			}
		}
		if err := s.pushZatcaItem(ctx, item, cloud); err != nil {
			s.recordOutboxFailure([]OutboxItem{item}, err.Error())
			continue
		}
		_, err = s.masterPool.Exec(ctx, `UPDATE sync_queue SET status='synced',synced_at=NOW(),last_error=NULL WHERE entity_type=$1 AND entity_id=$2 AND id<=$3 AND status IN ('pending','failed')`, item.EntityType, item.EntityID, item.ID)
		if err != nil {
			log.Printf("[ZATCA SYNC] Could not save acknowledgment for item %d", item.ID)
			return
		}
	}
}
