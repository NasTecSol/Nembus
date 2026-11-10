package sync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDeviceConfigIncomplete = errors.New("select a store before starting sync")

type DeviceIdentity struct {
	StoreID    int32
	TerminalID *int32
	Name       string
	SyncToken  string
}

func ParseDeviceIdentity(configJSON string) (DeviceIdentity, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(configJSON), &fields); err != nil {
		return DeviceIdentity{}, fmt.Errorf("invalid device configuration: %w", err)
	}
	parseID := func(key string) (*int32, error) {
		raw := fields[key]
		if len(raw) == 0 || string(raw) == "null" {
			return nil, nil
		}
		value := string(raw)
		if raw[0] == '"' {
			if err := json.Unmarshal(raw, &value); err != nil {
				return nil, err
			}
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, nil
		}
		id, err := strconv.ParseInt(value, 10, 32)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("%s must be a positive integer", key)
		}
		result := int32(id)
		return &result, nil
	}
	storeID, err := parseID("store_id")
	if err != nil {
		return DeviceIdentity{}, err
	}
	if storeID == nil {
		return DeviceIdentity{}, ErrDeviceConfigIncomplete
	}
	terminalID, err := parseID("terminal_id")
	if err != nil {
		return DeviceIdentity{}, err
	}
	name := "POS Terminal"
	if raw := fields["terminal_name"]; len(raw) > 0 {
		if err := json.Unmarshal(raw, &name); err != nil {
			return DeviceIdentity{}, fmt.Errorf("invalid terminal_name: %w", err)
		}
		if strings.TrimSpace(name) == "" {
			name = "POS Terminal"
		}
	}
	var syncToken string
	if raw := fields["sync_token"]; len(raw) > 0 && string(raw) != "null" {
		_ = json.Unmarshal(raw, &syncToken)
		syncToken = strings.TrimSpace(syncToken)
	}
	return DeviceIdentity{StoreID: *storeID, TerminalID: terminalID, Name: name, SyncToken: syncToken}, nil
}

// SaveDeviceIdentity uses separate statements in one transaction. Existing
// configuration and sync timestamps survive a rejected store/terminal change.
func SaveDeviceIdentity(ctx context.Context, pool *pgxpool.Pool, device DeviceIdentity) error {
	if device.StoreID <= 0 {
		return ErrDeviceConfigIncomplete
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('nembus.local_device_config'))`); err != nil {
		return err
	}
	_, _ = tx.Exec(ctx, `ALTER TABLE IF EXISTS local_device_config ADD COLUMN IF NOT EXISTS sync_token TEXT`)
	if device.TerminalID != nil {
		var matches bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pos_terminals WHERE id=$1 AND store_id=$2)`, *device.TerminalID, device.StoreID).Scan(&matches); err != nil {
			return err
		}
		if !matches {
			return fmt.Errorf("terminal %d does not belong to store %d", *device.TerminalID, device.StoreID)
		}
	}
	var id int32
	err = tx.QueryRow(ctx, `SELECT id FROM local_device_config ORDER BY id LIMIT 1 FOR UPDATE`).Scan(&id)
	if err == pgx.ErrNoRows {
		_, err = tx.Exec(ctx, `INSERT INTO local_device_config(device_name,store_id,pos_terminal_id,sync_token) VALUES($1,$2,$3,NULLIF($4,''))`,
			device.Name, device.StoreID, device.TerminalID, device.SyncToken)
	} else if err == nil {
		_, err = tx.Exec(ctx, `UPDATE local_device_config SET device_name=$1,store_id=$2,pos_terminal_id=$3,sync_token=COALESCE(NULLIF($4,''),sync_token) WHERE id=$5`,
			device.Name, device.StoreID, device.TerminalID, device.SyncToken, id)
	}
	if err != nil {
		return fmt.Errorf("save local device identity: %w", err)
	}
	// Seed __change_log__ watermark from restored backup change_log if not yet initialized
	_, _ = tx.Exec(ctx, `
		INSERT INTO sync_watermarks (entity_type, store_id, last_sync_at, metadata)
		SELECT '__change_log__', $1, NOW(),
		       jsonb_build_object('direction', 'pull', 'status', 'synced', 'last_seq', COALESCE((SELECT MAX(seq) FROM change_log), 0))
		WHERE to_regclass('public.change_log') IS NOT NULL
		  AND to_regclass('public.sync_watermarks') IS NOT NULL
		ON CONFLICT (entity_type, store_id) DO NOTHING
	`, device.StoreID)
	return tx.Commit(ctx)
}

// ReadLocalDeviceCredentials loads the configured store_id, stable device_id, and sync_token
// from local_device_config.
func ReadLocalDeviceCredentials(ctx context.Context, pool *pgxpool.Pool) (storeID int32, deviceID string, syncToken string, err error) {
	if pool == nil {
		return 0, "", "", fmt.Errorf("database pool is nil")
	}
	var terminalID int32
	var deviceName string
	err = pool.QueryRow(ctx, `
		SELECT COALESCE(store_id, 0),
		       COALESCE(pos_terminal_id, 0),
		       COALESCE(device_name, 'POS Terminal'),
		       COALESCE(sync_token, '')
		FROM local_device_config
		ORDER BY id
		LIMIT 1
	`).Scan(&storeID, &terminalID, &deviceName, &syncToken)
	if err != nil {
		// Fallback if sync_token column was not yet added
		err2 := pool.QueryRow(ctx, `
			SELECT COALESCE(store_id, 0),
			       COALESCE(pos_terminal_id, 0),
			       COALESCE(device_name, 'POS Terminal')
			FROM local_device_config
			ORDER BY id
			LIMIT 1
		`).Scan(&storeID, &terminalID, &deviceName)
		if err2 != nil {
			return 0, "", "", err2
		}
		err = nil
	}
	if terminalID > 0 {
		deviceID = fmt.Sprintf("terminal-%d", terminalID)
	} else if storeID > 0 {
		cleanName := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(deviceName)), " ", "-")
		if cleanName == "" {
			cleanName = "pos"
		}
		deviceID = fmt.Sprintf("store-%d-%s", storeID, cleanName)
	}
	return storeID, deviceID, strings.TrimSpace(syncToken), nil
}
