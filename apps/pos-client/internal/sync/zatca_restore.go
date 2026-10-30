package sync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Saved device configuration lives outside the tenant database and survives a
// re-clone. Only its serial is used to request the selected device's recovery.
func SavedZatcaSerial(raw string) string {
	var config struct {
		Settings string `json:"zatca_settings"`
	}
	var settings struct {
		Serial string `json:"device_serial"`
	}
	if json.Unmarshal([]byte(raw), &config) != nil || json.Unmarshal([]byte(config.Settings), &settings) != nil {
		return ""
	}
	return strings.TrimSpace(settings.Serial)
}

func normalizeSigningKey(key string) string { return strings.Join(strings.Fields(key), "") }

// GuardZatcaClone runs before the destructive SQL restore. A cloud download is
// never allowed to erase an unacknowledged device or a newer signing chain.
func GuardZatcaClone(ctx context.Context, pool *pgxpool.Pool, cloudURL, tenant, serial string) error {
	var exists bool
	if err := pool.QueryRow(ctx, `SELECT to_regclass('zatca_device_configs') IS NOT NULL`).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return nil
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM zatca_device_configs WHERE device_type='pos'`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	if serial == "" {
		return fmt.Errorf("save this device's completed setup before cloning again so its ZATCA identity can be recovered")
	}
	var id, org int32
	var key, certificate string
	var head int64
	err := pool.QueryRow(ctx, `SELECT id,organization_id,private_key_pem,COALESCE(production_csid,''),COALESCE((SELECT MAX(icv) FROM zatca_document_chain WHERE device_config_id=d.id),0) FROM zatca_device_configs d WHERE device_serial=$1 AND device_type='pos' LIMIT 1`, serial).Scan(&id, &org, &key, &certificate, &head)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	var pending int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM sync_queue WHERE status IN ('pending','failed') AND ((entity_type='zatca_device_configs' AND entity_id=$1) OR (entity_type='zatca_document_chain' AND (payload->>'device_config_id')::int=$1))`, id).Scan(&pending); err != nil {
		return err
	}
	if pending > 0 {
		return fmt.Errorf("ZATCA changes are not backed up yet (%d queued items); sync successfully before cloning again", pending)
	}
	cloud, err := NewZatcaCloud(ctx, cloudURL, tenant)
	if err != nil {
		return err
	}
	page, err := cloud.RecoveryPage(ctx, serial, org, 0, -1)
	if err != nil {
		return fmt.Errorf("verify cloud ZATCA backup before cloning: %w", err)
	}
	var remote struct {
		PrivateKey  string `json:"private_key_pem"`
		Certificate string `json:"production_csid"`
	}
	if json.Unmarshal(page.Device, &remote) != nil || normalizeSigningKey(key) != normalizeSigningKey(remote.PrivateKey) || (certificate != "" && certificate != remote.Certificate) || page.HeadICV < head {
		return fmt.Errorf("cloud ZATCA backup does not match this device's current credentials or chain; local database was preserved")
	}
	if head > 0 {
		last, err := cloud.RecoveryPage(ctx, serial, org, head-1, head)
		if err != nil {
			return err
		}
		var remoteDoc struct {
			Hash string `json:"xml_hash"`
			UUID string `json:"zatca_uuid"`
		}
		var localHash, localUUID string
		if len(last.Documents) != 1 || json.Unmarshal(last.Documents[0], &remoteDoc) != nil {
			return fmt.Errorf("cloud device-chain checkpoint is missing")
		}
		if err := pool.QueryRow(ctx, `SELECT xml_hash,zatca_uuid::text FROM zatca_document_chain WHERE device_config_id=$1 AND icv=$2`, id, head).Scan(&localHash, &localUUID); err != nil {
			return err
		}
		if remoteDoc.Hash != localHash || remoteDoc.UUID != localUUID {
			return fmt.Errorf("cloud device-chain checkpoint differs; local database was preserved")
		}
	}
	return nil
}

// RecoverZatcaDevice restores this physical serial over HTTPS.
func RecoverZatcaDevice(ctx context.Context, repo *repository.Queries, cloudURL, tenant, serial string, orgID int32) error {
	cloud, err := NewZatcaCloud(ctx, cloudURL, tenant)
	if err != nil {
		return err
	}
	return repo.RestoreZatcaDevice(ctx, serial, orgID, func(after, through int64) (repository.ZatcaRestorePage, error) {
		return cloud.RecoveryPage(ctx, serial, orgID, after, through)
	})
}
