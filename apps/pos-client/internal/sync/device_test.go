package sync

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestParseDeviceIdentity(t *testing.T) {
	for _, raw := range []string{`{"store_id":"1","terminal_id":"7"}`, `{"store_id":1,"terminal_id":7}`} {
		device, err := ParseDeviceIdentity(raw)
		if err != nil || device.StoreID != 1 || device.TerminalID == nil || *device.TerminalID != 7 || device.Name != "POS Terminal" {
			t.Fatalf("parse %s: device=%+v err=%v", raw, device, err)
		}
	}
	for _, raw := range []string{`{}`, `{"app_language":"en"}`, `{"store_id":""}`} {
		if _, err := ParseDeviceIdentity(raw); !errors.Is(err, ErrDeviceConfigIncomplete) {
			t.Fatalf("incomplete config %s: %v", raw, err)
		}
	}
	for _, raw := range []string{`invalid`, `{"store_id":-1}`, `{"store_id":1.5}`, `{"store_id":1,"terminal_id":"bad"}`} {
		if _, err := ParseDeviceIdentity(raw); err == nil || errors.Is(err, ErrDeviceConfigIncomplete) {
			t.Fatalf("invalid config accepted: %s err=%v", raw, err)
		}
	}
}

func TestSaveDeviceIdentityAndPreserveExistingConfiguration(t *testing.T) {
	dsn := os.Getenv("NEMBUS_SYNC_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set NEMBUS_SYNC_TEST_DATABASE_URL for PostgreSQL tests")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := fmt.Sprintf("device_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE") }()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	_, err = pool.Exec(ctx, `CREATE TABLE stores(id integer PRIMARY KEY);
		CREATE TABLE pos_terminals(id integer PRIMARY KEY,store_id integer REFERENCES stores(id));
		CREATE TABLE local_device_config(id serial PRIMARY KEY,device_name text,
		store_id integer REFERENCES stores(id),pos_terminal_id integer REFERENCES pos_terminals(id),
		last_zatca_sync_at timestamptz DEFAULT '2026-01-01');
		INSERT INTO stores VALUES(1),(2); INSERT INTO pos_terminals VALUES(7,1)`)
	if err != nil {
		t.Fatal(err)
	}
	// The original parameterized multi-command Exec fails before saving anything.
	if _, err := pool.Exec(ctx, `DELETE FROM local_device_config;
		INSERT INTO local_device_config(device_name,store_id,pos_terminal_id) VALUES($1,$2,$3)`, "POS", 1, 7); err == nil {
		t.Fatal("expected original parameterized multi-statement save to fail")
	}
	device, err := ParseDeviceIdentity(`{"store_id":"1","terminal_id":"7","terminal_name":"Till 1"}`)
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveDeviceIdentity(ctx, pool, device); err != nil {
		t.Fatal(err)
	}
	var originalID, storeID, terminalID int32
	var originalTimestamp time.Time
	if err := pool.QueryRow(ctx, `SELECT id,last_zatca_sync_at FROM local_device_config`).Scan(&originalID, &originalTimestamp); err != nil {
		t.Fatal(err)
	}
	device.Name = "Till renamed"
	if err := SaveDeviceIdentity(ctx, pool, device); err != nil {
		t.Fatal(err)
	}
	device.StoreID = 2
	if err := SaveDeviceIdentity(ctx, pool, device); err == nil {
		t.Fatal("accepted a terminal belonging to another store")
	}
	device.StoreID = 999
	device.TerminalID = nil
	if err := SaveDeviceIdentity(ctx, pool, device); err == nil {
		t.Fatal("accepted missing store")
	}
	var id int32
	var timestamp time.Time
	var name string
	if err := pool.QueryRow(ctx, `SELECT id,store_id,pos_terminal_id,device_name,last_zatca_sync_at FROM local_device_config`).Scan(&id, &storeID, &terminalID, &name, &timestamp); err != nil {
		t.Fatal(err)
	}
	if id != originalID || storeID != 1 || terminalID != 7 || name != "Till renamed" || !timestamp.Equal(originalTimestamp) {
		t.Fatalf("configuration changed after failed save: id=%d store=%d terminal=%d name=%s timestamp=%v", id, storeID, terminalID, name, timestamp)
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM local_device_config").Scan(&count); err != nil || count != 1 {
		t.Fatalf("save created duplicates: %d %v", count, err)
	}
}
