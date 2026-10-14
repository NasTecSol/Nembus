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
	return DeviceIdentity{StoreID: *storeID, TerminalID: terminalID, Name: name}, nil
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
		_, err = tx.Exec(ctx, `INSERT INTO local_device_config(device_name,store_id,pos_terminal_id) VALUES($1,$2,$3)`, device.Name, device.StoreID, device.TerminalID)
	} else if err == nil {
		_, err = tx.Exec(ctx, `UPDATE local_device_config SET device_name=$1,store_id=$2,pos_terminal_id=$3 WHERE id=$4`, device.Name, device.StoreID, device.TerminalID, id)
	}
	if err != nil {
		return fmt.Errorf("save local device identity: %w", err)
	}
	return tx.Commit(ctx)
}
