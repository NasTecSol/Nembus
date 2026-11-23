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
	StoreID      int32
	TerminalID   *int32
	TerminalCode string
	Name         string
	SyncToken    string
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
	var terminalCode string
	if raw := fields["terminal_code"]; len(raw) > 0 && string(raw) != "null" {
		_ = json.Unmarshal(raw, &terminalCode)
		terminalCode = strings.TrimSpace(terminalCode)
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
	return DeviceIdentity{
		StoreID:      *storeID,
		TerminalID:   terminalID,
		TerminalCode: terminalCode,
		Name:         name,
		SyncToken:    syncToken,
	}, nil
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
	_, _ = tx.Exec(ctx, `ALTER TABLE IF EXISTS local_device_config ADD COLUMN IF NOT EXISTS sync_token TEXT, ADD COLUMN IF NOT EXISTS terminal_code TEXT`)
	_, _ = tx.Exec(ctx, `ALTER TABLE IF EXISTS pos_terminals ADD COLUMN IF NOT EXISTS terminal_code TEXT`)
	terminalCode := strings.TrimSpace(device.TerminalCode)
	if device.TerminalID != nil {
		var dbTerminalCode string
		err := tx.QueryRow(ctx, `SELECT COALESCE(terminal_code, '') FROM pos_terminals WHERE id=$1 AND store_id=$2`, *device.TerminalID, device.StoreID).Scan(&dbTerminalCode)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("terminal %d does not belong to store %d", *device.TerminalID, device.StoreID)
		}
		if err != nil {
			return err
		}
		if strings.TrimSpace(dbTerminalCode) != "" {
			terminalCode = strings.TrimSpace(dbTerminalCode)
		}
	}
	var id int32
	err = tx.QueryRow(ctx, `SELECT id FROM local_device_config ORDER BY id LIMIT 1 FOR UPDATE`).Scan(&id)
	if err == pgx.ErrNoRows {
		_, err = tx.Exec(ctx, `INSERT INTO local_device_config(device_name,store_id,pos_terminal_id,sync_token,terminal_code) VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,''))`,
			device.Name, device.StoreID, device.TerminalID, device.SyncToken, terminalCode)
	} else if err == nil {
		_, err = tx.Exec(ctx, `UPDATE local_device_config SET device_name=$1,store_id=$2,pos_terminal_id=$3,sync_token=COALESCE(NULLIF($4,''),sync_token),terminal_code=NULLIF($5,'') WHERE id=$6`,
			device.Name, device.StoreID, device.TerminalID, device.SyncToken, terminalCode, id)
	}
	if err != nil {
		return fmt.Errorf("save local device identity: %w", err)
	}
	// Seed __change_log__ watermark from restored backup change_log if not yet initialized
	var hasChangeLogTables bool
	_ = tx.QueryRow(ctx, `SELECT to_regclass('change_log') IS NOT NULL AND to_regclass('sync_watermarks') IS NOT NULL`).Scan(&hasChangeLogTables)
	if hasChangeLogTables {
		if _, spErr := tx.Exec(ctx, `SAVEPOINT sp_seed_changelog_watermark`); spErr == nil {
			if _, seedErr := tx.Exec(ctx, `
				INSERT INTO sync_watermarks (entity_type, store_id, last_sync_at, metadata)
				SELECT '__change_log__', $1, NOW(),
				       jsonb_build_object('direction', 'pull', 'status', 'synced', 'last_seq', COALESCE((SELECT MAX(seq) FROM change_log), 0))
				ON CONFLICT (entity_type, store_id) DO NOTHING
			`, device.StoreID); seedErr != nil {
				_, _ = tx.Exec(ctx, `ROLLBACK TO SAVEPOINT sp_seed_changelog_watermark`)
			}
			_, _ = tx.Exec(ctx, `RELEASE SAVEPOINT sp_seed_changelog_watermark`)
		}
	}
	return tx.Commit(ctx)
}

// ReadLocalTerminalCode resolves the active terminal_code from local_device_config
// and/or pos_terminals when pos_terminal_id is configured.
func ReadLocalTerminalCode(ctx context.Context, pool *pgxpool.Pool) string {
	if pool == nil {
		return ""
	}
	var terminalCode string
	err := pool.QueryRow(ctx, `
		SELECT COALESCE(
			NULLIF(TRIM(l.terminal_code), ''),
			(SELECT COALESCE(TRIM(t.terminal_code), '') FROM pos_terminals t WHERE t.id = l.pos_terminal_id),
			''
		)
		FROM local_device_config l
		ORDER BY l.id
		LIMIT 1
	`).Scan(&terminalCode)
	if err != nil {
		_ = pool.QueryRow(ctx, `
			SELECT COALESCE(TRIM(t.terminal_code), '')
			FROM local_device_config l
			JOIN pos_terminals t ON t.id = l.pos_terminal_id
			ORDER BY l.id
			LIMIT 1
		`).Scan(&terminalCode)
	}
	return strings.TrimSpace(terminalCode)
}

// EnrichDeviceConfigJSON resolves terminal_code from pos_terminals (or local_device_config)
// when terminal_id is configured and injects "terminal_code" into the JSON object.
func EnrichDeviceConfigJSON(ctx context.Context, pool *pgxpool.Pool, configJSON string) string {
	trimmed := strings.TrimSpace(configJSON)
	if trimmed == "" || pool == nil {
		return configJSON
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &fields); err != nil {
		return configJSON
	}
	identity, err := ParseDeviceIdentity(trimmed)
	var terminalCode string
	if err == nil && identity.TerminalID != nil && *identity.TerminalID > 0 {
		_ = pool.QueryRow(ctx, `SELECT COALESCE(TRIM(terminal_code), '') FROM pos_terminals WHERE id = $1`, *identity.TerminalID).Scan(&terminalCode)
	}
	if terminalCode == "" {
		terminalCode = ReadLocalTerminalCode(ctx, pool)
	}
	if terminalCode == "" && err == nil {
		terminalCode = strings.TrimSpace(identity.TerminalCode)
	}
	if terminalCode == "" {
		return configJSON
	}
	encodedCode, mErr := json.Marshal(terminalCode)
	if mErr != nil {
		return configJSON
	}
	fields["terminal_code"] = encodedCode
	enriched, mErr := json.Marshal(fields)
	if mErr != nil {
		return configJSON
	}
	return string(enriched)
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
