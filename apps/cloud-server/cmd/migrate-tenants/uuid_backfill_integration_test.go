package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func detUUID(table string, oldID int64) uuid.UUID {
	return utils.DetUUIDFromLegacyID(table, oldID)
}

// migratedTables lists all 15 POS-originated push-synced tables migrated from SERIAL -> UUID.
var migratedTables = []string{
	"user_biometric_templates",
	"cashier_sessions",
	"customers",
	"kiosk_sessions",
	"stock_counts",
	"stock_count_lines",
	"pos_transactions",
	"pos_transaction_lines",
	"pos_payments",
	"restaurant_orders",
	"restaurant_order_items",
	"waste_logs",
	"sales_returns",
	"sales_return_lines",
	"stock_movements",
}

// all29FKChecks enumerates every referencing FK column across all 20 tables
// pointing to the 15 migrated tables.
type fkVerificationCase struct {
	childTable    string
	fkColumn      string
	parentTable   string
	rowWhere      string
	parentOldID   int64
	nullRowWhere  string // optional row that must preserve NULL
}

var all29FKChecks = []fkVerificationCase{
	// 1-9: external tables referencing customers(id)
	{childTable: "carts", fkColumn: "customer_id", parentTable: "customers", rowWhere: "cart_number = 'CART-101'", parentOldID: 101, nullRowWhere: "cart_number = 'CART-102'"},
	{childTable: "draft_cart_templates", fkColumn: "customer_id", parentTable: "customers", rowWhere: "template_name = 'Weekly Order'", parentOldID: 101},
	{childTable: "sales_orders_v2", fkColumn: "customer_id", parentTable: "customers", rowWhere: "order_number = 'SO2-101'", parentOldID: 101, nullRowWhere: "order_number = 'SO2-102'"},
	{childTable: "invoices", fkColumn: "customer_id", parentTable: "customers", rowWhere: "invoice_number = 'INV-101'", parentOldID: 101, nullRowWhere: "invoice_number = 'INV-102'"},
	{childTable: "quotes", fkColumn: "customer_id", parentTable: "customers", rowWhere: "quote_number = 'QT-101'", parentOldID: 101},
	{childTable: "sales_analytics", fkColumn: "customer_id", parentTable: "customers", rowWhere: "revenue = 100", parentOldID: 101},
	{childTable: "sales_orders", fkColumn: "customer_id", parentTable: "customers", rowWhere: "order_number = 'SO-101'", parentOldID: 101},
	{childTable: "customer_addresses", fkColumn: "customer_id", parentTable: "customers", rowWhere: "address_line = '123 King Fahd Rd'", parentOldID: 101},
	{childTable: "customer_payments", fkColumn: "customer_id", parentTable: "customers", rowWhere: "payment_number = 'CPAY-101'", parentOldID: 101},

	// 10: stock_count_lines
	{childTable: "stock_count_lines", fkColumn: "stock_count_id", parentTable: "stock_counts", rowWhere: "expected_quantity = 100", parentOldID: 101},

	// 11-12: pos_transactions
	{childTable: "pos_transactions", fkColumn: "cashier_session_id", parentTable: "cashier_sessions", rowWhere: "transaction_number = 'POS-101'", parentOldID: 101},
	{childTable: "pos_transactions", fkColumn: "customer_id", parentTable: "customers", rowWhere: "transaction_number = 'POS-101'", parentOldID: 101, nullRowWhere: "transaction_number = 'POS-102'"},

	// 13-14: pos_transaction_lines & pos_payments
	{childTable: "pos_transaction_lines", fkColumn: "transaction_id", parentTable: "pos_transactions", rowWhere: "quantity = 2 AND unit_price = 50", parentOldID: 101},
	{childTable: "pos_payments", fkColumn: "transaction_id", parentTable: "pos_transactions", rowWhere: "ecr_reference_id = 'ECR-101'", parentOldID: 101},

	// 15-17: restaurant_carts
	{childTable: "restaurant_carts", fkColumn: "customer_id", parentTable: "customers", rowWhere: "cart_number = 'RCART-101'", parentOldID: 101, nullRowWhere: "cart_number = 'RCART-102'"},
	{childTable: "restaurant_carts", fkColumn: "kiosk_session_id", parentTable: "kiosk_sessions", rowWhere: "cart_number = 'RCART-101'", parentOldID: 101, nullRowWhere: "cart_number = 'RCART-102'"},
	{childTable: "restaurant_carts", fkColumn: "converted_to_order_id", parentTable: "restaurant_orders", rowWhere: "cart_number = 'RCART-101'", parentOldID: 101, nullRowWhere: "cart_number = 'RCART-102'"},

	// 18-21: restaurant_orders
	{childTable: "restaurant_orders", fkColumn: "cashier_session_id", parentTable: "cashier_sessions", rowWhere: "order_number = 'ORD-101'", parentOldID: 101, nullRowWhere: "order_number = 'ORD-102'"},
	{childTable: "restaurant_orders", fkColumn: "customer_id", parentTable: "customers", rowWhere: "order_number = 'ORD-101'", parentOldID: 101, nullRowWhere: "order_number = 'ORD-102'"},
	{childTable: "restaurant_orders", fkColumn: "pos_transaction_id", parentTable: "pos_transactions", rowWhere: "order_number = 'ORD-101'", parentOldID: 101, nullRowWhere: "order_number = 'ORD-102'"},
	{childTable: "restaurant_orders", fkColumn: "kiosk_session_id", parentTable: "kiosk_sessions", rowWhere: "order_number = 'ORD-101'", parentOldID: 101},

	// 22-23: restaurant_order_items & restaurant_order_status_history
	{childTable: "restaurant_order_items", fkColumn: "order_id", parentTable: "restaurant_orders", rowWhere: "quantity = 2 AND status = 'preparing'", parentOldID: 101},
	{childTable: "restaurant_order_status_history", fkColumn: "restaurant_order_id", parentTable: "restaurant_orders", rowWhere: "to_status = 'preparing'", parentOldID: 101},

	// 24: waste_logs
	{childTable: "waste_logs", fkColumn: "order_id", parentTable: "restaurant_orders", rowWhere: "reason = 'Spilled'", parentOldID: 101, nullRowWhere: "reason = 'Expired'"},

	// 25-27: sales_returns
	{childTable: "sales_returns", fkColumn: "cashier_session_id", parentTable: "cashier_sessions", rowWhere: "return_number = 'RET-101'", parentOldID: 101, nullRowWhere: "return_number = 'RET-102'"},
	{childTable: "sales_returns", fkColumn: "customer_id", parentTable: "customers", rowWhere: "return_number = 'RET-101'", parentOldID: 101, nullRowWhere: "return_number = 'RET-102'"},
	{childTable: "sales_returns", fkColumn: "original_transaction_id", parentTable: "pos_transactions", rowWhere: "return_number = 'RET-101'", parentOldID: 101, nullRowWhere: "return_number = 'RET-102'"},

	// 28-29: sales_return_lines
	{childTable: "sales_return_lines", fkColumn: "return_id", parentTable: "sales_returns", rowWhere: "condition = 'good'", parentOldID: 101},
	{childTable: "sales_return_lines", fkColumn: "original_line_id", parentTable: "pos_transaction_lines", rowWhere: "condition = 'good'", parentOldID: 101, nullRowWhere: "condition = 'damaged'"},
}

type legacySnapshot struct {
	preCounts map[string]int64
}

func TestUUIDBackfillAndUpgradePaths(t *testing.T) {
	rawDSN := resolveBackfillTestDSN()
	if rawDSN == "" {
		t.Skip("set NEMBUS_BACKFILL_TEST_DATABASE_URL or NEMBUS_SYNC_TEST_DATABASE_URL to run PostgreSQL backfill integration tests")
	}

	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("failed to locate repository root: %v", err)
	}

	migrationsDir := filepath.Join(repoRoot, "packages", "core", "db", "migrations")
	initialSchemaPath := filepath.Join(migrationsDir, "20260918111505_initial_schema.sql")
	initialBaseSQL, m1UUIDSQL, err := loadInitialMigrationSections(initialSchemaPath)
	if err != nil {
		t.Fatalf("read initial migration sections: %v", err)
	}
	if strings.TrimSpace(m1UUIDSQL) == "" {
		m1File := filepath.Join(migrationsDir, "20261121120000_m1_uuid_push_entities.sql")
		if raw, readErr := os.ReadFile(m1File); readErr == nil {
			m1UUIDSQL = strings.TrimSpace(string(raw))
		}
	}
	if strings.TrimSpace(m1UUIDSQL) == "" {
		if os.Getenv("NEMBUS_REQUIRE_UUID_MIGRATION") == "1" {
			t.Fatalf("NEMBUS_REQUIRE_UUID_MIGRATION=1 but M1 UUID migration missing (expected marker in %s or file 20261121120000_m1_uuid_push_entities.sql)", initialSchemaPath)
		}
		t.Skipf("pending M1 migration: no M1 UUID SQL in %s or 20261121120000_m1_uuid_push_entities.sql", initialSchemaPath)
	}

	baseURL, err := dsnToURL(rawDSN, "")
	if err != nil {
		t.Fatalf("parse backfill test DSN: %v", err)
	}
	if err := validateSafeTestBaseURL(baseURL); err != nil {
		t.Fatalf("unsafe test database DSN refused: %v", err)
	}

	atlasBin := resolveAtlasBinary()
	schemaDir := filepath.Join(repoRoot, "packages", "core", "db", "schema")
	preDeclarativePath := filepath.Join(repoRoot, "packages", "core", "db", "patches", "pre_declarative.sql")
	viewsPath := filepath.Join(repoRoot, "packages", "core", "db", "views_functions", "90_views_functions.sql")
	posMigrationsDir := filepath.Join(repoRoot, "apps", "pos-client", "migrations")

	t.Run("PathA_VersionedMigration", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		dbURL, pool := createDisposableDB(t, ctx, baseURL, "patha")
		if err := applySQLText(ctx, dbURL, initialBaseSQL); err != nil {
			t.Fatalf("apply initial schema (pre-M1 base): %v", err)
		}

		snap := seedLegacyDataset(t, ctx, pool, true)

		if err := applySQLText(ctx, dbURL, m1UUIDSQL); err != nil {
			t.Fatalf("apply M1 UUID section: %v", err)
		}

		verifyMigratedDatabase(t, ctx, pool, snap, true)
	})

	t.Run("PathB_DeclarativeSchemaApply", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		// Verify pre_declarative.sql is a safe no-op on an empty initialized schema.
		emptyURL, _ := createDisposableDB(t, ctx, baseURL, "empty")
		if err := applySQLText(ctx, emptyURL, initialBaseSQL); err != nil {
			t.Fatalf("apply initial schema to empty DB: %v", err)
		}
		if err := applySQLScript(ctx, emptyURL, preDeclarativePath); err != nil {
			t.Fatalf("pre_declarative.sql failed on empty initialized DB: %v", err)
		}

		dbURL, pool := createDisposableDB(t, ctx, baseURL, "pathb")
		if err := applySQLText(ctx, dbURL, initialBaseSQL); err != nil {
			t.Fatalf("apply initial schema: %v", err)
		}

		snap := seedLegacyDataset(t, ctx, pool, true)

		// Apply pre_declarative.sql twice to prove idempotency.
		if err := applySQLScript(ctx, dbURL, preDeclarativePath); err != nil {
			t.Fatalf("first pre_declarative.sql apply failed: %v", err)
		}
		if err := applySQLScript(ctx, dbURL, preDeclarativePath); err != nil {
			t.Fatalf("second (idempotent) pre_declarative.sql apply failed: %v", err)
		}

		devURL := os.Getenv("ATLAS_DEV_URL")
		if devURL == "" {
			devURL, _ = createDisposableDB(t, ctx, baseURL, "bdev")
		} else {
			parsedDev, err := dsnToURL(devURL, "")
			if err != nil {
				t.Fatalf("parse ATLAS_DEV_URL: %v", err)
			}
			if err := validateSafeTestBaseURL(parsedDev); err != nil {
				t.Fatalf("unsafe ATLAS_DEV_URL refused: %v", err)
			}
			devURL = parsedDev
		}

		if err := executeAtlasDeclarative(atlasBin, dbURL, schemaDir, devURL, false); err != nil {
			t.Fatalf("executeAtlasDeclarative failed: %v", err)
		}
		if err := applySQLScript(ctx, dbURL, viewsPath); err != nil {
			t.Fatalf("apply 90_views_functions.sql failed: %v", err)
		}

		verifyMigratedDatabase(t, ctx, pool, snap, true)
	})

	t.Run("PathC_POSSimulation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		posM1Path := filepath.Join(posMigrationsDir, "20261121120000_m1_uuid_push_entities.sql")
		posM1SQL, err := os.ReadFile(posM1Path)
		if err != nil {
			t.Fatalf("read POS M1 UUID migration %s: %v", posM1Path, err)
		}
		posM1SQLText := stripGooseUpSQL(string(posM1SQL))
		if strings.TrimSpace(posM1SQLText) == "" {
			t.Fatalf("POS M1 UUID migration file empty: %s", posM1Path)
		}

		dbURL, pool := createDisposableDB(t, ctx, baseURL, "pathc")
		if err := applySQLText(ctx, dbURL, initialBaseSQL); err != nil {
			t.Fatalf("apply initial schema: %v", err)
		}

		// Install legacy POS extensions (sync_queue.entity_id BIGINT + legacy triggers) before seeding.
		if _, err := pool.Exec(ctx, legacyPOSExtensionsDDL); err != nil {
			t.Fatalf("install legacy POS extensions DDL: %v", err)
		}

		snap := seedLegacyDataset(t, ctx, pool, false)

		// Verify legacy sync_queue captured integer IDs while change_log remained suppressed.
		var legacyOutboxCount int64
		if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM sync_queue WHERE entity_type = 'pos_transactions' AND entity_id = 101`).Scan(&legacyOutboxCount); err != nil {
			t.Fatalf("query legacy sync_queue: %v", err)
		}
		if legacyOutboxCount != 1 {
			t.Fatalf("expected 1 legacy sync_queue row for pos_transactions=101, got %d", legacyOutboxCount)
		}
		var clCount int64
		if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM change_log`).Scan(&clCount); err != nil {
			t.Fatalf("query change_log on POS DB: %v", err)
		}
		if clCount != 0 {
			t.Fatalf("expected change_log suppression when local_device_config exists, got %d rows", clCount)
		}

		// Seed one non-migrated integer outbox row in legacy sync_queue to verify mixed-key preservation.
		if _, err := pool.Exec(ctx, `INSERT INTO sync_queue (entity_type, entity_id, action, payload, status) VALUES ('products', 101, 'UPDATE', '{"id":101}'::jsonb, 'pending')`); err != nil {
			t.Fatalf("seed non-migrated sync_queue row: %v", err)
		}
		snap.preCounts["sync_queue"]++

		// Apply M1 UUID section from synced POS initial migration + updated 99999999999999_pos_extensions.sql.
		posM1Body := posM1SQLText
		if _, err := pool.Exec(ctx, posM1Body); err != nil {
			t.Fatalf("apply POS M1 UUID section: %v", err)
		}
		posExtPath := filepath.Join(posMigrationsDir, "99999999999999_pos_extensions.sql")
		if err := applyGooseUpFile(ctx, pool, posExtPath); err != nil {
			t.Fatalf("apply %s: %v", posExtPath, err)
		}

		verifyMigratedDatabase(t, ctx, pool, snap, false)

		// Verify sync_queue.entity_id was upgraded to TEXT and backfilled deterministically.
		var syncQueueType string
		if err := pool.QueryRow(ctx, `
			SELECT data_type FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = 'sync_queue' AND column_name = 'entity_id'
		`).Scan(&syncQueueType); err != nil {
			t.Fatalf("inspect sync_queue.entity_id type: %v", err)
		}
		if syncQueueType != "text" {
			t.Fatalf("expected sync_queue.entity_id data_type='text', got %q", syncQueueType)
		}

		var gotTxOutboxID string
		if err := pool.QueryRow(ctx, `SELECT entity_id FROM sync_queue WHERE entity_type = 'pos_transactions' ORDER BY id ASC LIMIT 1`).Scan(&gotTxOutboxID); err != nil {
			t.Fatalf("query migrated sync_queue pos_transactions row: %v", err)
		}
		wantTxUUID := detUUID("pos_transactions", 101).String()
		if gotTxOutboxID != wantTxUUID {
			t.Errorf("sync_queue pos_transactions entity_id = %q, want %q", gotTxOutboxID, wantTxUUID)
		}

		var gotProdOutboxID string
		if err := pool.QueryRow(ctx, `SELECT entity_id FROM sync_queue WHERE entity_type = 'products' ORDER BY id ASC LIMIT 1`).Scan(&gotProdOutboxID); err != nil {
			t.Fatalf("query unmigrated sync_queue products row: %v", err)
		}
		if gotProdOutboxID != "101" {
			t.Errorf("sync_queue products entity_id = %q, want %q", gotProdOutboxID, "101")
		}

		// Fresh INSERT into pos_transactions must trigger trg_enqueue_sync_outbox() with a valid RFC 4122 UUIDv4 text.
		var freshTxID uuid.UUID
		err = pool.QueryRow(ctx, `
			INSERT INTO pos_transactions (store_id, cashier_id, cashier_session_id, transaction_number, subtotal, total_amount, amount_paid, status)
			VALUES (101, 101, $1, 'POS-FRESH-999', 25.00, 25.00, 25.00, 'completed')
			RETURNING id
		`, detUUID("cashier_sessions", 101)).Scan(&freshTxID)
		if err != nil {
			t.Fatalf("fresh POS transaction insert failed: %v", err)
		}
		assertValidUUIDv4(t, "fresh pos_transactions.id", freshTxID)

		var enqueuedEntityID string
		if err := pool.QueryRow(ctx, `
			SELECT entity_id FROM sync_queue
			WHERE entity_type = 'pos_transactions' AND entity_id = $1
		`, freshTxID.String()).Scan(&enqueuedEntityID); err != nil {
			t.Fatalf("expected fresh pos_transactions outbox row for %s: %v", freshTxID, err)
		}
		if enqueuedEntityID != freshTxID.String() {
			t.Errorf("enqueued sync_queue.entity_id = %q, want %q", enqueuedEntityID, freshTxID.String())
		}
	})
}

func resolveBackfillTestDSN() string {
	for _, k := range []string{
		"NEMBUS_BACKFILL_TEST_DATABASE_URL",
		"NEMBUS_MIGRATION_TEST_DATABASE_URL",
		"NEMBUS_SYNC_TEST_DATABASE_URL",
		"NEMBUS_RESTORE_TEST_DSN",
	} {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		probe := filepath.Join(dir, "packages", "core", "db", "migrations", "20260918111505_initial_schema.sql")
		if _, err := os.Stat(probe); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not locate packages/core/db/migrations/20260918111505_initial_schema.sql from working directory")
		}
		dir = parent
	}
}

var forbiddenDBNames = []string{"masterdb", "qitaf", "prod", "production", "staging"}

func validateSafeTestBaseURL(dbURL string) error {
	u, err := url.Parse(dbURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	dbName := strings.ToLower(strings.TrimPrefix(u.Path, "/"))
	for _, f := range forbiddenDBNames {
		if dbName == f || strings.Contains(dbName, f) {
			return fmt.Errorf("refusing forbidden database name %q", dbName)
		}
	}
	host := strings.ToLower(u.Hostname())
	if host != "localhost" && host != "127.0.0.1" && host != "::1" {
		if os.Getenv("NEMBUS_ALLOW_REMOTE_TEST_DB") != "1" {
			return fmt.Errorf("refusing non-local database host %q without NEMBUS_ALLOW_REMOTE_TEST_DB=1", host)
		}
	}
	return nil
}

func createDisposableDB(t *testing.T, ctx context.Context, baseURL, suffix string) (string, *pgxpool.Pool) {
	t.Helper()

	dbName := fmt.Sprintf("nembus_uuid_backfill_test_%s_%d", suffix, time.Now().UnixNano())
	if len(dbName) > 63 {
		dbName = dbName[:63]
	}
	for _, f := range forbiddenDBNames {
		if strings.Contains(strings.ToLower(dbName), f) {
			t.Fatalf("generated unsafe db name %q", dbName)
		}
	}

	adminPool, err := pgxpool.New(ctx, baseURL)
	if err != nil {
		t.Fatalf("connect admin pool: %v", err)
	}

	createSQL := fmt.Sprintf(`CREATE DATABASE "%s" WITH ENCODING 'UTF8'`, dbName)
	if _, err := adminPool.Exec(ctx, createSQL); err != nil {
		adminPool.Close()
		t.Fatalf("create disposable database %q: %v", dbName, err)
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		adminPool.Close()
		t.Fatalf("parse baseURL: %v", err)
	}
	u.Path = "/" + dbName
	targetURL := u.String()

	testPool, err := pgxpool.New(ctx, targetURL)
	if err != nil {
		_, _ = adminPool.Exec(context.Background(), fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, dbName))
		adminPool.Close()
		t.Fatalf("connect disposable test pool %q: %v", dbName, err)
	}

	t.Cleanup(func() {
		testPool.Close()
		dropCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if _, err := adminPool.Exec(dropCtx, fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, dbName)); err != nil {
			t.Logf("warning: failed to drop disposable database %q: %v", dbName, err)
		}
		adminPool.Close()
	})

	return targetURL, testPool
}

func assertValidUUIDv4(t *testing.T, label string, u uuid.UUID) {
	t.Helper()
	if u == uuid.Nil {
		t.Errorf("%s is nil UUID", label)
		return
	}
	if u.Version() != 4 {
		t.Errorf("%s UUID version = %d, want 4 (%s)", label, u.Version(), u)
	}
	if u.Variant() != uuid.RFC4122 {
		t.Errorf("%s UUID variant = %v, want RFC4122 (%s)", label, u.Variant(), u)
	}
}

func seedLegacyDataset(t *testing.T, ctx context.Context, pool *pgxpool.Pool, isCloud bool) legacySnapshot {
	t.Helper()

	execMust := func(sqlStmt string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sqlStmt, args...); err != nil {
			t.Fatalf("seed SQL failed:\n%s\nerr: %v", sqlStmt, err)
		}
	}

	// 1. Prerequisite parent tables in strict topological order.
	execMust(`INSERT INTO organizations (id, name, code, currency_code) VALUES (101, 'Test Org', 'ORG-101', 'SAR')`)
	execMust(`INSERT INTO stores (id, organization_id, name, code, is_pos_enabled) VALUES
		(101, 101, 'Main Store', 'STR-101', true),
		(102, 101, 'Branch Store', 'STR-102', true)`)
	execMust(`INSERT INTO storage_locations (id, store_id, code, name) VALUES (101, 101, 'LOC-101', 'Shelf A')`)
	execMust(`INSERT INTO pos_terminals (id, store_id, terminal_code, terminal_name) VALUES
		(101, 101, 'T01', 'Terminal 1'),
		(102, 101, 'T02', 'Terminal 2')`)
	execMust(`INSERT INTO users (id, organization_id, username, email, password_hash, first_name, last_name) VALUES
		(101, 101, 'cashier1', 'cashier1@test.local', '$2a$10$test', 'Alice', 'Cashier'),
		(102, 101, 'cashier2', 'cashier2@test.local', '$2a$10$test', 'Bob', 'Manager')`)
	execMust(`INSERT INTO cashiers (id, user_id, store_id, cashier_code, discount_limit) VALUES
		(101, 101, 101, 'CSH-101', 10.00),
		(102, 102, 101, 'CSH-102', 20.00)`)
	execMust(`INSERT INTO product_categories (id, name, code) VALUES (101, 'Beverages', 'CAT-101')`)
	execMust(`INSERT INTO brands (id, name, code) VALUES (101, 'NembusBrand', 'BRD-101')`)
	execMust(`INSERT INTO units_of_measure (id, code, name) VALUES (101, 'PCS', 'Pieces')`)
	execMust(`INSERT INTO tax_categories (id, name, code, tax_rate) VALUES (101, 'VAT 15%', 'VAT15', 15.00)`)
	execMust(`INSERT INTO price_lists (id, name, code, currency_code) VALUES (101, 'Retail', 'PL-101', 'SAR')`)
	execMust(`INSERT INTO products (id, organization_id, sku, name, category_id, brand_id, base_uom_id, tax_category_id, cost_price) VALUES
		(101, 101, 'SKU-101', 'Espresso', 101, 101, 101, 101, 5.00),
		(102, 101, 'SKU-102', 'Croissant', 101, 101, 101, 101, 3.00)`)
	execMust(`INSERT INTO product_variants (id, product_id, variant_sku, variant_name, variant_attributes) VALUES
		(101, 101, 'SKU-101-L', 'Large', '{"size":"L"}'::jsonb)`)
	execMust(`INSERT INTO business_partners (id, organization_id, code, name, partner_role, currency_code) VALUES
		(101, 101, 'BP-101', 'Coffee Supplier', 'supplier', 'SAR')`)
	execMust(`INSERT INTO purchase_orders (id, organization_id, po_number, partners_id, store_id, po_date) VALUES
		(101, 101, 'PO-101', 101, 101, CURRENT_DATE)`)
	execMust(`INSERT INTO menu_categories (id, store_id, name, code) VALUES (101, 101, 'Hot Drinks', 'MC-101')`)
	execMust(`INSERT INTO recipes (id, organization_id, recipe_code, recipe_name, finished_product_id, yield_uom_id) VALUES
		(101, 101, 'RCP-101', 'Latte Recipe', 101, 101)`)
	execMust(`INSERT INTO menu_items (id, store_id, menu_category_id, product_id, recipe_id, name, short_name, base_price, cost_price, tax_category_id, item_type) VALUES
		(101, 101, 101, 101, 101, 'Latte', 'Latte', 18.00, 5.00, 101, 'standard')`)
	execMust(`INSERT INTO restaurant_tables (id, store_id, table_number, table_name) VALUES
		(101, 101, 'T-101', 'Table 1')`)
	execMust(`INSERT INTO inventory_stock (product_id, store_id, storage_location_id, quantity_on_hand, quantity_available) VALUES
		(101, 101, 101, 100.000, 100.000)`)

	// 2. All 15 migrated tables (2+ rows each with explicit IDs 101, 102).
	execMust(`INSERT INTO user_biometric_templates (id, organization_id, user_id, finger_index, fmd_base64, quality_score) VALUES
		(101, 101, 101, 1, 'Zm1kMTAx', 5),
		(102, 101, 102, 2, 'Zm1kMTAy', 4)`)
	execMust(`INSERT INTO cashier_sessions (id, cashier_id, pos_terminal_id, session_number, opening_balance, status) VALUES
		(101, 101, 101, 'SES-LEGACY-101', 100.00, 'open'),
		(102, 102, 102, 'SES-LEGACY-102', 200.00, 'closed')`)
	execMust(`INSERT INTO customers (id, organization_id, customer_code, name, email, phone, credit_limit, outstanding_balance, loyalty_points) VALUES
		(101, 101, 'CUST-101', 'Legacy Customer 1', 'c101@test.local', '+966500000101', 5000.00, 100.00, 50.00),
		(102, 101, 'CUST-102', 'Legacy Customer 2', 'c102@test.local', '+966500000102', 2000.00, 0.00, 10.00)`)
	execMust(`INSERT INTO kiosk_sessions (id, pos_terminal_id, store_id, session_token, status) VALUES
		(101, 101, 101, 'kiosk-tok-101', 'active'),
		(102, 102, 101, 'kiosk-tok-102', 'closed')`)
	execMust(`INSERT INTO stock_counts (id, count_number, store_id, storage_location_id, count_type, status, counted_by, approved_by) VALUES
		(101, 'SC-101', 101, 101, 'full', 'approved', 101, 102),
		(102, 'SC-102', 101, 101, 'cycle', 'planned', 101, NULL)`)
	execMust(`INSERT INTO stock_count_lines (id, stock_count_id, product_id, storage_location_id, expected_quantity, system_quantity, counted_quantity, variance, uom_id) VALUES
		(101, 101, 101, 101, 100, 100, 98, -2, 101),
		(102, 102, 102, 101, 50, 50, 50, 0, 101)`)
	execMust(`INSERT INTO pos_transactions (id, store_id, cashier_id, cashier_session_id, customer_id, pos_terminal_id, transaction_number, transaction_type, subtotal, tax_amount, total_amount, total_cost, amount_paid, status, price_list_id) VALUES
		(101, 101, 101, 101, 101, 101, 'POS-101', 'sale', 100.00, 15.00, 115.00, 40.00, 115.00, 'completed', 101),
		(102, 101, 102, 102, NULL, NULL, 'POS-102', 'sale', 50.00, 7.50, 57.50, 20.00, 57.50, 'completed', NULL)`)
	execMust(`INSERT INTO pos_transaction_lines (id, transaction_id, product_id, product_variant_id, quantity, uom_id, unit_price, tax_amount, subtotal, line_total, cost_price, line_number) VALUES
		(101, 101, 101, 101, 2, 101, 50.00, 15.00, 100.00, 115.00, 20.00, 1),
		(102, 102, 102, NULL, 1, 101, 50.00, 7.50, 50.00, 57.50, 20.00, 1)`)
	execMust(`INSERT INTO pos_payments (id, transaction_id, payment_method, amount, status, ecr_reference_id) VALUES
		(101, 101, 'cash', 115.00, 'completed', 'ECR-101'),
		(102, 102, 'card', 57.50, 'completed', 'ECR-102')`)
	execMust(`INSERT INTO restaurant_orders (id, store_id, table_id, cashier_id, cashier_session_id, customer_id, order_number, order_source, status, subtotal, tax_amount, total_amount, pos_transaction_id, organization_id, pos_terminal_id, kiosk_session_id, order_type, payment_status) VALUES
		(101, 101, 101, 101, 101, 101, 'ORD-101', 'pos', 'preparing', 36.00, 5.40, 41.40, 101, 101, 101, 101, 'dine_in', 'paid'),
		(102, 101, 101, 102, NULL, NULL, 'ORD-102', 'kiosk', 'pending', 18.00, 2.70, 20.70, NULL, 101, 102, 102, 'takeaway', 'unpaid')`)
	execMust(`INSERT INTO restaurant_order_items (id, order_id, menu_item_id, menu_item_name, quantity, unit_price, subtotal, tax_amount, line_total, line_number, status) VALUES
		(101, 101, 101, 'Latte', 2, 18.00, 36.00, 5.40, 41.40, 1, 'preparing'),
		(102, 102, 101, 'Latte', 1, 18.00, 18.00, 2.70, 20.70, 1, 'pending')`)
	execMust(`INSERT INTO waste_logs (id, store_id, product_id, menu_item_id, recipe_id, waste_source, quantity, uom_id, unit_cost, total_cost, reason, logged_by, order_id) VALUES
		(101, 101, 101, 101, 101, 'kitchen', 1.000, 101, 5.00, 5.00, 'Spilled', 101, 101),
		(102, 101, 102, NULL, NULL, 'register', 2.000, 101, 3.00, 6.00, 'Expired', 102, NULL)`)
	execMust(`INSERT INTO sales_returns (id, return_number, store_id, cashier_id, cashier_session_id, customer_id, original_transaction_id, return_reason, status, subtotal, tax_amount, total_refund_amount, refund_method, approved_by) VALUES
		(101, 'RET-101', 101, 101, 101, 101, 101, 'Defective', 'completed', 50.00, 7.50, 57.50, 'cash', 102),
		(102, 'RET-102', 101, 102, NULL, NULL, NULL, 'Unlinked', 'pending', 10.00, 1.50, 11.50, 'cash', NULL)`)

	// Polymorphic stock_movements (explicit IDs 101..105) before sales_return_lines trigger fires.
	execMust(`INSERT INTO stock_movements (id, movement_type, reference_type, reference_id, product_id, from_store_id, to_store_id, quantity, uom_id, posted_by) VALUES
		(101, 'sale', 'pos_transaction', 101, 101, 101, NULL, 2.000, 101, 101),
		(102, 'count_adjustment', 'stock_count', 101, 101, NULL, 101, -2.000, 101, 101),
		(103, 'purchase_receipt', 'purchase_order', 101, 101, NULL, 101, 10.000, 101, 101),
		(104, 'kitchen_prep', 'restaurant_order', 101, 101, 101, NULL, 1.000, 101, 101),
		(105, 'stock_adjustment', NULL, NULL, 101, NULL, 101, 5.000, 101, 101)`)

	// Advance sequences for all 15 migrated tables so trigger inserts (sales_return_lines) get id >= 201.
	for _, tbl := range migratedTables {
		execMust(fmt.Sprintf(`SELECT setval(pg_get_serial_sequence('public.%s', 'id'), GREATEST(COALESCE((SELECT MAX(id) FROM %s), 1), 200))`, tbl, tbl))
	}

	// sales_return_lines triggers trg_process_sales_return -> adds 1 stock_movements row (good) and 1 waste_logs row (damaged).
	execMust(`INSERT INTO sales_return_lines (id, return_id, product_id, product_variant_id, original_line_id, quantity, unit_price, refund_amount, return_to_stock, condition, line_number) VALUES
		(101, 101, 101, 101, 101, 1.000, 50.00, 57.50, true, 'good', 1),
		(102, 102, 102, NULL, NULL, 1.000, 10.00, 11.50, false, 'damaged', 1)`)
	execMust(`SELECT setval(pg_get_serial_sequence('public.sales_return_lines', 'id'), GREATEST(COALESCE((SELECT MAX(id) FROM sales_return_lines), 1), 200))`)

	// 3. Remaining 11 external referencing tables (covering all 29 FK columns across 20 tables).
	execMust(`INSERT INTO carts (cart_number, organization_id, store_id, customer_id, guest_identifier) VALUES
		('CART-101', 101, 101, 101, NULL),
		('CART-102', 101, 101, NULL, 'guest-102')`)
	execMust(`INSERT INTO draft_cart_templates (organization_id, customer_id, template_name) VALUES
		(101, 101, 'Weekly Order')`)
	execMust(`INSERT INTO sales_orders_v2 (order_number, organization_id, store_id, customer_id, shipping_address, billing_address) VALUES
		('SO2-101', 101, 101, 101, '{}'::jsonb, '{}'::jsonb),
		('SO2-102', 101, 101, NULL, '{}'::jsonb, '{}'::jsonb)`)
	execMust(`INSERT INTO invoices (invoice_number, organization_id, store_id, customer_id, customer_name, due_date, total_amount, balance_due, billing_address) VALUES
		('INV-101', 101, 101, 101, 'Legacy Customer 1', CURRENT_DATE, 115.00, 115.00, '{}'::jsonb),
		('INV-102', 101, 101, NULL, 'Walk-in', CURRENT_DATE, 50.00, 0.00, '{}'::jsonb)`)
	execMust(`INSERT INTO quotes (quote_number, organization_id, store_id, customer_id, customer_name, valid_until, total_amount) VALUES
		('QT-101', 101, 101, 101, 'Legacy Customer 1', CURRENT_DATE, 100.00)`)
	execMust(`INSERT INTO sales_analytics (organization_id, store_id, product_id, category_id, customer_id, date, revenue, net_revenue, transactions) VALUES
		(101, 101, 101, 101, 101, CURRENT_DATE, 100.00, 100.00, 1)`)
	execMust(`INSERT INTO sales_orders (organization_id, order_number, customer_id, store_id, order_date) VALUES
		(101, 'SO-101', 101, 101, CURRENT_DATE)`)
	execMust(`INSERT INTO customer_addresses (customer_id, address_type, address_line, city, country) VALUES
		(101, 'shipping', '123 King Fahd Rd', 'Riyadh', 'SA')`)
	execMust(`INSERT INTO customer_payments (organization_id, payment_number, customer_id, customer_code, payment_date, payment_amount) VALUES
		(101, 'CPAY-101', 101, 'CUST-101', CURRENT_DATE, 50.00)`)
	execMust(`INSERT INTO restaurant_carts (cart_number, organization_id, store_id, table_id, customer_id, kiosk_session_id, converted_to_order_id) VALUES
		('RCART-101', 101, 101, 101, 101, 101, 101),
		('RCART-102', 101, 101, 101, NULL, NULL, NULL)`)
	execMust(`INSERT INTO restaurant_order_status_history (restaurant_order_id, store_id, from_status, to_status, changed_by_user_id) VALUES
		(101, 101, 'pending', 'preparing', 101)`)

	// 4. Seed sync_inbox (1 migrated + 1 unmigrated).
	execMust(`INSERT INTO sync_inbox (outbox_uuid, device_id, store_id, batch_id, entity_type, entity_id) VALUES
		('11111111-1111-4111-8111-111111111101', 'dev-t01', 101, 'batch-1', 'pos_transactions', 101),
		('11111111-1111-4111-8111-111111111102', 'dev-t01', 101, 'batch-1', 'products', 101)`)

	// 5. Capture pre-migration row counts AFTER all triggers have executed.
	tablesToCount := append([]string{}, migratedTables...)
	tablesToCount = append(tablesToCount,
		"carts", "draft_cart_templates", "sales_orders_v2", "invoices",
		"quotes", "sales_analytics", "sales_orders", "customer_addresses",
		"customer_payments", "restaurant_carts", "restaurant_order_status_history",
		"sync_inbox",
	)
	if isCloud {
		tablesToCount = append(tablesToCount, "change_log")
	} else {
		tablesToCount = append(tablesToCount, "sync_queue")
	}

	preCounts := make(map[string]int64, len(tablesToCount))
	for _, tbl := range tablesToCount {
		var cnt int64
		if err := pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s`, tbl)).Scan(&cnt); err != nil {
			t.Fatalf("capture preCount for %s: %v", tbl, err)
		}
		preCounts[tbl] = cnt
	}

	return legacySnapshot{preCounts: preCounts}
}

func verifyMigratedDatabase(t *testing.T, ctx context.Context, pool *pgxpool.Pool, snap legacySnapshot, isCloud bool) {
	t.Helper()

	// 1. information_schema.columns verification: 15 PKs (uuid + gen_random_uuid()), 29 FKs (uuid), 3 text sync/polymorphic cols.
	for _, tbl := range migratedTables {
		var dataType string
		var colDefault *string
		err := pool.QueryRow(ctx, `
			SELECT data_type, column_default
			FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1 AND column_name = 'id'
		`, tbl).Scan(&dataType, &colDefault)
		if err != nil {
			t.Fatalf("inspect %s.id column: %v", tbl, err)
		}
		if dataType != "uuid" {
			t.Errorf("%s.id data_type = %q, want 'uuid'", tbl, dataType)
		}
		if colDefault == nil || !strings.Contains(*colDefault, "gen_random_uuid") {
			t.Errorf("%s.id column_default = %v, want gen_random_uuid()", tbl, colDefault)
		}
	}

	seenFK := map[string]bool{}
	for _, fk := range all29FKChecks {
		key := fk.childTable + "." + fk.fkColumn
		if seenFK[key] {
			continue
		}
		seenFK[key] = true
		var dataType string
		if err := pool.QueryRow(ctx, `
			SELECT data_type FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2
		`, fk.childTable, fk.fkColumn).Scan(&dataType); err != nil {
			t.Fatalf("inspect FK %s: %v", key, err)
		}
		if dataType != "uuid" {
			t.Errorf("FK column %s data_type = %q, want 'uuid'", key, dataType)
		}
	}

	for _, textCol := range [][2]string{
		{"stock_movements", "reference_id"},
		{"sync_inbox", "entity_id"},
		{"change_log", "entity_id"},
	} {
		var dataType string
		if err := pool.QueryRow(ctx, `
			SELECT data_type FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2
		`, textCol[0], textCol[1]).Scan(&dataType); err != nil {
			t.Fatalf("inspect %s.%s: %v", textCol[0], textCol[1], err)
		}
		if dataType != "text" {
			t.Errorf("%s.%s data_type = %q, want 'text'", textCol[0], textCol[1], dataType)
		}
	}

	// 2. Exact 100% row-count preservation.
	for tbl, pre := range snap.preCounts {
		var post int64
		if err := pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s`, tbl)).Scan(&post); err != nil {
			t.Fatalf("count %s post-migration: %v", tbl, err)
		}
		if post != pre {
			t.Errorf("row count mismatch on %s: post=%d, pre=%d", tbl, post, pre)
		}
	}

	// 3. Deterministic UUID match for every seeded row across all 15 migrated tables.
	for _, tbl := range migratedTables {
		for _, oldID := range []int64{101, 102} {
			expected := detUUID(tbl, oldID)
			var exists bool
			if err := pool.QueryRow(ctx, fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE id = $1)`, tbl), expected).Scan(&exists); err != nil {
				t.Fatalf("check detUUID %s(%d): %v", tbl, oldID, err)
			}
			if !exists {
				t.Errorf("missing deterministic UUID %s for %s old_id=%d", expected, tbl, oldID)
			}
		}
	}

	// 4. Exact parent-child JOIN integrity across ALL 29 FK columns + NULL preservation.
	for _, fk := range all29FKChecks {
		wantParent := detUUID(fk.parentTable, fk.parentOldID)
		q := fmt.Sprintf(`
			SELECT c.%s
			FROM %s c
			JOIN %s p ON c.%s = p.id
			WHERE c.%s
		`, fk.fkColumn, fk.childTable, fk.parentTable, fk.fkColumn, fk.rowWhere)
		var gotParent uuid.UUID
		if err := pool.QueryRow(ctx, q).Scan(&gotParent); err != nil {
			t.Fatalf("FK JOIN check %s.%s -> %s (%s): %v", fk.childTable, fk.fkColumn, fk.parentTable, fk.rowWhere, err)
		}
		if gotParent != wantParent {
			t.Errorf("FK JOIN %s.%s = %s, want %s", fk.childTable, fk.fkColumn, gotParent, wantParent)
		}

		if fk.nullRowWhere != "" {
			var isNull bool
			nq := fmt.Sprintf(`SELECT %s IS NULL FROM %s WHERE %s`, fk.fkColumn, fk.childTable, fk.nullRowWhere)
			if err := pool.QueryRow(ctx, nq).Scan(&isNull); err != nil {
				t.Fatalf("NULL FK check %s.%s (%s): %v", fk.childTable, fk.fkColumn, fk.nullRowWhere, err)
			}
			if !isNull {
				t.Errorf("expected %s.%s to remain NULL where %s", fk.childTable, fk.fkColumn, fk.nullRowWhere)
			}
		}
	}

	// 5. Polymorphic stock_movements.reference_id verification.
	polyChecks := []struct {
		oldMovementID int64
		refType       string
		wantRefID     string
	}{
		{101, "pos_transaction", detUUID("pos_transactions", 101).String()},
		{102, "stock_count", detUUID("stock_counts", 101).String()},
		{103, "purchase_order", "101"},
		{104, "restaurant_order", detUUID("restaurant_orders", 101).String()},
	}
	for _, pc := range polyChecks {
		var gotRefType string
		var gotRefID *string
		if err := pool.QueryRow(ctx, `SELECT reference_type, reference_id FROM stock_movements WHERE id = $1`, detUUID("stock_movements", pc.oldMovementID)).Scan(&gotRefType, &gotRefID); err != nil {
			t.Fatalf("query stock_movements(%d): %v", pc.oldMovementID, err)
		}
		if gotRefType != pc.refType {
			t.Errorf("stock_movements(%d).reference_type = %q, want %q", pc.oldMovementID, gotRefType, pc.refType)
		}
		if gotRefID == nil || *gotRefID != pc.wantRefID {
			t.Errorf("stock_movements(%d).reference_id = %v, want %q", pc.oldMovementID, gotRefID, pc.wantRefID)
		}
	}
	var nullRefID *string
	if err := pool.QueryRow(ctx, `SELECT reference_id FROM stock_movements WHERE id = $1`, detUUID("stock_movements", 105)).Scan(&nullRefID); err != nil {
		t.Fatalf("query stock_movements(105): %v", err)
	}
	if nullRefID != nil {
		t.Errorf("stock_movements(105).reference_id = %q, want NULL", *nullRefID)
	}
	var triggerReturnRef string
	if err := pool.QueryRow(ctx, `SELECT reference_id FROM stock_movements WHERE movement_type = 'return' AND reference_type = 'sales_return' LIMIT 1`).Scan(&triggerReturnRef); err != nil {
		t.Fatalf("query trigger-generated return stock_movement: %v", err)
	}
	if triggerReturnRef != detUUID("sales_returns", 101).String() {
		t.Errorf("return stock_movement reference_id = %q, want %q", triggerReturnRef, detUUID("sales_returns", 101).String())
	}

	// 6. sync_inbox & change_log TEXT entity_id verification.
	var inboxMigratedID string
	if err := pool.QueryRow(ctx, `SELECT entity_id FROM sync_inbox WHERE outbox_uuid = '11111111-1111-4111-8111-111111111101'`).Scan(&inboxMigratedID); err != nil {
		t.Fatalf("query sync_inbox migrated row: %v", err)
	}
	if inboxMigratedID != detUUID("pos_transactions", 101).String() {
		t.Errorf("sync_inbox pos_transactions entity_id = %q, want %q", inboxMigratedID, detUUID("pos_transactions", 101).String())
	}
	var inboxProductID string
	if err := pool.QueryRow(ctx, `SELECT entity_id FROM sync_inbox WHERE outbox_uuid = '11111111-1111-4111-8111-111111111102'`).Scan(&inboxProductID); err != nil {
		t.Fatalf("query sync_inbox unmigrated row: %v", err)
	}
	if inboxProductID != "101" {
		t.Errorf("sync_inbox products entity_id = %q, want '101'", inboxProductID)
	}

	if isCloud {
		for _, clCheck := range []struct {
			entityType string
			wantID     string
		}{
			{"customers", detUUID("customers", 101).String()},
			{"stock_counts", detUUID("stock_counts", 101).String()},
			{"stock_movements", detUUID("stock_movements", 101).String()},
			{"products", "101"},
		} {
			var found bool
			if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM change_log WHERE entity_type = $1 AND entity_id = $2)`, clCheck.entityType, clCheck.wantID).Scan(&found); err != nil {
				t.Fatalf("query change_log (%s, %s): %v", clCheck.entityType, clCheck.wantID, err)
			}
			if !found {
				t.Errorf("change_log missing entity_type=%q entity_id=%q", clCheck.entityType, clCheck.wantID)
			}
		}
	}

	// 7. Query all 8 dependent views.
	views := []string{
		"vw_customer_aging_report",
		"vw_profit_margin_analysis",
		"vw_active_restaurant_orders",
		"vw_waste_daily_summary",
		"vw_realtime_pnl",
		"vw_cashier_shift_reconciliation",
		"vw_stock_movement_ledger",
		"vw_tax_vat_summary",
	}
	for _, v := range views {
		var cnt int64
		if err := pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s`, v)).Scan(&cnt); err != nil {
			t.Errorf("query view %s failed: %v", v, err)
		}
	}

	// 8. Invoke key updated functions.
	var kdsCount int64
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM fn_get_kds_orders(101)`).Scan(&kdsCount); err != nil {
		t.Errorf("fn_get_kds_orders(101) failed: %v", err)
	} else if kdsCount < 1 {
		t.Errorf("fn_get_kds_orders(101) returned %d rows, want >= 1", kdsCount)
	}

	var returnableQty float64
	if err := pool.QueryRow(ctx, `SELECT fn_get_returnable_quantity($1::uuid)::float8`, detUUID("pos_transaction_lines", 101)).Scan(&returnableQty); err != nil {
		t.Errorf("fn_get_returnable_quantity failed: %v", err)
	} else if returnableQty != 1.0 {
		t.Errorf("fn_get_returnable_quantity = %v, want 1.0", returnableQty)
	}

	var loyaltyCust uuid.UUID
	var ruleApplied string
	if err := pool.QueryRow(ctx, `SELECT rule_applied, customer_id FROM fn_calculate_loyalty_earned($1::uuid)`, detUUID("pos_transactions", 101)).Scan(&ruleApplied, &loyaltyCust); err != nil {
		t.Errorf("fn_calculate_loyalty_earned failed: %v", err)
	} else if loyaltyCust != detUUID("customers", 101) {
		t.Errorf("fn_calculate_loyalty_earned customer_id = %s, want %s", loyaltyCust, detUUID("customers", 101))
	}

	var reconcileOK bool
	var reconcileMsg string
	var linesUpdated int
	if err := pool.QueryRow(ctx, `SELECT success, message, lines_updated FROM fn_reconcile_stock_count($1::uuid)`, detUUID("stock_counts", 101)).Scan(&reconcileOK, &reconcileMsg, &linesUpdated); err != nil {
		t.Errorf("fn_reconcile_stock_count failed: %v", err)
	} else if !reconcileOK || linesUpdated != 1 {
		t.Errorf("fn_reconcile_stock_count = (%v, %q, %d), want (true, ..., 1)", reconcileOK, reconcileMsg, linesUpdated)
	}

	// 9. Subsequent INSERTs omitting id across ALL 15 migrated tables generate valid RFC 4122 UUIDv4.
	cID := detUUID("customers", 101)
	sessID := detUUID("cashier_sessions", 101)
	scID := detUUID("stock_counts", 101)
	txID := detUUID("pos_transactions", 101)
	ordID := detUUID("restaurant_orders", 101)
	retID := detUUID("sales_returns", 101)

	freshInserts := []struct {
		table string
		sql   string
		args  []any
	}{
		{"user_biometric_templates", `INSERT INTO user_biometric_templates (organization_id, user_id, finger_index, fmd_base64) VALUES (101, 101, 9, 'Zm1kOTk5') RETURNING id`, nil},
		{"cashier_sessions", `INSERT INTO cashier_sessions (cashier_id, pos_terminal_id, session_number) VALUES (101, 101, 'SES-POST-901') RETURNING id`, nil},
		{"customers", `INSERT INTO customers (organization_id, customer_code, name) VALUES (101, 'CUST-POST-901', 'Post Migration Customer') RETURNING id`, nil},
		{"kiosk_sessions", `INSERT INTO kiosk_sessions (pos_terminal_id, store_id, session_token) VALUES (101, 101, 'kiosk-post-901') RETURNING id`, nil},
		{"stock_counts", `INSERT INTO stock_counts (count_number, store_id) VALUES ('SC-POST-901', 101) RETURNING id`, nil},
		{"stock_count_lines", `INSERT INTO stock_count_lines (stock_count_id, product_id) VALUES ($1, 101) RETURNING id`, []any{scID}},
		{"pos_transactions", `INSERT INTO pos_transactions (store_id, cashier_id, cashier_session_id, customer_id, transaction_number, subtotal, total_amount, amount_paid) VALUES (101, 101, $1, $2, 'POS-POST-901', 10, 10, 10) RETURNING id`, []any{sessID, cID}},
		{"pos_transaction_lines", `INSERT INTO pos_transaction_lines (transaction_id, product_id, quantity, unit_price, subtotal) VALUES ($1, 101, 1, 10, 10) RETURNING id`, []any{txID}},
		{"pos_payments", `INSERT INTO pos_payments (transaction_id, payment_method, amount) VALUES ($1, 'cash', 10) RETURNING id`, []any{txID}},
		{"restaurant_orders", `INSERT INTO restaurant_orders (store_id, order_number) VALUES (101, 'ORD-POST-901') RETURNING id`, nil},
		{"restaurant_order_items", `INSERT INTO restaurant_order_items (order_id, menu_item_id, quantity, unit_price, subtotal) VALUES ($1, 101, 1, 18, 18) RETURNING id`, []any{ordID}},
		{"waste_logs", `INSERT INTO waste_logs (store_id, product_id, quantity) VALUES (101, 101, 1) RETURNING id`, nil},
		{"sales_returns", `INSERT INTO sales_returns (return_number, store_id) VALUES ('RET-POST-901', 101) RETURNING id`, nil},
		{"sales_return_lines", `INSERT INTO sales_return_lines (return_id, product_id, quantity, unit_price, refund_amount, return_to_stock, condition) VALUES ($1, 101, 1, 10, 10, false, 'damaged') RETURNING id`, []any{retID}},
		{"stock_movements", `INSERT INTO stock_movements (movement_type, product_id, quantity, to_store_id) VALUES ('manual', 101, 1, 101) RETURNING id`, nil},
	}

	insertedIDs := make(map[string]uuid.UUID, len(freshInserts))
	for _, fi := range freshInserts {
		var newID uuid.UUID
		if err := pool.QueryRow(ctx, fi.sql, fi.args...).Scan(&newID); err != nil {
			t.Errorf("post-migration INSERT into %s failed: %v", fi.table, err)
			continue
		}
		assertValidUUIDv4(t, fi.table+".id", newID)
		insertedIDs[fi.table] = newID
	}

	if isCloud {
		if custNewID, ok := insertedIDs["customers"]; ok {
			var found bool
			if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM change_log WHERE entity_type = 'customers' AND entity_id = $1 AND op = 'INSERT')`, custNewID.String()).Scan(&found); err != nil {
				t.Errorf("check change_log for new customer %s: %v", custNewID, err)
			} else if !found {
				t.Errorf("trg_record_change_log did not record newly inserted customer UUID %s in change_log", custNewID)
			}
		}
	}
}

const m1UUIDSectionMarker = "@nembus:m1_uuid_push_entities_begin"

func loadInitialMigrationSections(path string) (baseSQL, m1SQL string, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("read %s: %w", path, err)
	}
	s := strings.ReplaceAll(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\r", "\n")
	idx := strings.Index(s, m1UUIDSectionMarker)
	if idx < 0 {
		return strings.TrimSpace(s), "", nil
	}
	// Include the banner comment block that starts before the marker line.
	lineStart := strings.LastIndex(s[:idx], "\n-- =====")
	if lineStart < 0 {
		lineStart = idx
	} else {
		lineStart++ // skip newline before banner
	}
	return strings.TrimSpace(s[:lineStart]), strings.TrimSpace(s[lineStart:]), nil
}

func applySQLText(ctx context.Context, dbURL, sqlText string) error {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, sqlText); err != nil {
		return fmt.Errorf("exec sql: %w", err)
	}
	return nil
}

var gooseDirectiveLineRe = regexp.MustCompile(`(?m)^--\s*\+goose\s+.*(?:\r?\n|$)`)

func stripGooseUpSQL(raw string) string {
	if idx := strings.Index(raw, "-- +goose Down"); idx >= 0 {
		raw = raw[:idx]
	}
	return gooseDirectiveLineRe.ReplaceAllString(raw, "")
}

func applyGooseUpFile(ctx context.Context, pool *pgxpool.Pool, filePath string) error {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}
	sqlText := stripGooseUpSQL(string(b))
	if _, err := pool.Exec(ctx, sqlText); err != nil {
		return fmt.Errorf("exec %s: %w", filePath, err)
	}
	return nil
}

// legacyPOSExtensionsDDL installs pre-M1 POS-only outbox/device tables and triggers
// (sync_queue.entity_id BIGINT NOT NULL) to simulate an existing POS terminal before upgrade.
const legacyPOSExtensionsDDL = `
CREATE TABLE IF NOT EXISTS local_printer_configs (
    id              SERIAL PRIMARY KEY,
    printer_name    VARCHAR(100) NOT NULL,
    connection_type VARCHAR(50)  NOT NULL,
    ip_address      VARCHAR(45),
    port            INTEGER,
    is_default      BOOLEAN      DEFAULT false,
    created_at      TIMESTAMPTZ  DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sync_queue (
    id              BIGSERIAL    PRIMARY KEY,
    outbox_uuid     UUID         NOT NULL DEFAULT gen_random_uuid(),
    entity_type     VARCHAR(50)  NOT NULL,
    entity_id       BIGINT       NOT NULL,
    action          VARCHAR(20)  NOT NULL,
    payload         JSONB        NOT NULL,
    status          VARCHAR(20)  DEFAULT 'pending',
    priority        INTEGER      DEFAULT 0,
    correlation_id  VARCHAR(100),
    max_retries     INTEGER      DEFAULT 5,
    retry_count     INTEGER      DEFAULT 0,
    last_error      TEXT,
    error_class     VARCHAR(32),
    next_attempt_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    quarantined_at  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  DEFAULT NOW(),
    synced_at       TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS local_device_config (
    id                  SERIAL       PRIMARY KEY,
    device_name         VARCHAR(100) NOT NULL,
    store_id            INTEGER      REFERENCES stores(id),
    pos_terminal_id     INTEGER      REFERENCES pos_terminals(id),
    sync_token          TEXT,
    last_zatca_sync_at  TIMESTAMPTZ  DEFAULT '1970-01-01',
    zatca_enabled       BOOLEAN      DEFAULT false,
    created_at          TIMESTAMPTZ  DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION trg_enqueue_sync_outbox()
RETURNS TRIGGER AS $$
DECLARE
    entity_name VARCHAR(50);
    prio INTEGER := 5;
    corr_id VARCHAR(100);
    existing_id BIGINT;
BEGIN
    IF current_setting('nembus.sync_applier', true) = 'true' THEN
        RETURN NEW;
    END IF;
    entity_name := TG_ARGV[0];
    IF entity_name IN ('pos_transactions', 'pos_payments', 'sales_returns') THEN
        prio := 10;
    END IF;
    corr_id := COALESCE(NULLIF(current_setting('nembus.correlation_id', true), ''), pg_current_xact_id()::text);
    SELECT id INTO existing_id
    FROM sync_queue
    WHERE correlation_id = corr_id
      AND entity_type = entity_name
      AND entity_id = NEW.id
      AND status = 'pending'
    ORDER BY id DESC
    LIMIT 1;
    IF existing_id IS NOT NULL THEN
        UPDATE sync_queue SET payload = row_to_json(NEW)::jsonb WHERE id = existing_id;
    ELSE
        INSERT INTO sync_queue (entity_type, entity_id, action, payload, status, priority, correlation_id)
        VALUES (entity_name, NEW.id, TG_OP, row_to_json(NEW)::jsonb, 'pending', prio, corr_id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_sync_user_biometric_templates ON user_biometric_templates;
CREATE TRIGGER trg_sync_user_biometric_templates AFTER INSERT OR UPDATE ON user_biometric_templates FOR EACH ROW EXECUTE FUNCTION trg_enqueue_sync_outbox('user_biometric_templates');
DROP TRIGGER IF EXISTS trg_sync_pos_transactions ON pos_transactions;
CREATE TRIGGER trg_sync_pos_transactions AFTER INSERT OR UPDATE ON pos_transactions FOR EACH ROW EXECUTE FUNCTION trg_enqueue_sync_outbox('pos_transactions');
DROP TRIGGER IF EXISTS trg_sync_pos_transaction_lines ON pos_transaction_lines;
CREATE TRIGGER trg_sync_pos_transaction_lines AFTER INSERT ON pos_transaction_lines FOR EACH ROW EXECUTE FUNCTION trg_enqueue_sync_outbox('pos_transaction_lines');
DROP TRIGGER IF EXISTS trg_sync_pos_payments ON pos_payments;
CREATE TRIGGER trg_sync_pos_payments AFTER INSERT ON pos_payments FOR EACH ROW EXECUTE FUNCTION trg_enqueue_sync_outbox('pos_payments');
DROP TRIGGER IF EXISTS trg_sync_cashier_sessions ON cashier_sessions;
CREATE TRIGGER trg_sync_cashier_sessions AFTER INSERT OR UPDATE ON cashier_sessions FOR EACH ROW EXECUTE FUNCTION trg_enqueue_sync_outbox('cashier_sessions');
DROP TRIGGER IF EXISTS trg_sync_sales_returns ON sales_returns;
CREATE TRIGGER trg_sync_sales_returns AFTER INSERT OR UPDATE ON sales_returns FOR EACH ROW EXECUTE FUNCTION trg_enqueue_sync_outbox('sales_returns');
DROP TRIGGER IF EXISTS trg_sync_sales_return_lines ON sales_return_lines;
CREATE TRIGGER trg_sync_sales_return_lines AFTER INSERT ON sales_return_lines FOR EACH ROW EXECUTE FUNCTION trg_enqueue_sync_outbox('sales_return_lines');
DROP TRIGGER IF EXISTS trg_sync_stock_movements ON stock_movements;
CREATE TRIGGER trg_sync_stock_movements AFTER INSERT ON stock_movements FOR EACH ROW
    WHEN (NEW.movement_type IN ('stock_adjustment', 'adjustment', 'waste', 'damage', 'physical_count'))
    EXECUTE FUNCTION trg_enqueue_sync_outbox('stock_movements');
`

func TestUUIDBackfillHelpers(t *testing.T) {
	u1 := detUUID("customers", 101)
	u2 := detUUID("customers", 102)
	uOther := detUUID("pos_transactions", 101)
	assertValidUUIDv4(t, "detUUID(customers,101)", u1)
	assertValidUUIDv4(t, "detUUID(customers,102)", u2)
	assertValidUUIDv4(t, "detUUID(pos_transactions,101)", uOther)
	if u1 == u2 || u1 == uOther {
		t.Fatalf("detUUID collision: u1=%s u2=%s uOther=%s", u1, u2, uOther)
	}
	if u1 != detUUID("customers", 101) {
		t.Fatal("detUUID must be deterministic")
	}

	if err := validateSafeTestBaseURL("postgres://u:p@localhost:5432/postgres?sslmode=disable"); err != nil {
		t.Errorf("expected local postgres maintenance DB to be allowed, got: %v", err)
	}
	for _, bad := range []string{
		"postgres://u:p@localhost:5432/masterDB",
		"postgres://u:p@localhost:5432/qitaf",
		"postgres://u:p@localhost:5432/nembus_prod",
		"postgres://u:p@localhost:5432/staging",
		"postgres://u:p@db.example.com:5432/postgres",
	} {
		if err := validateSafeTestBaseURL(bad); err == nil {
			t.Errorf("expected validateSafeTestBaseURL(%q) to fail", bad)
		}
	}

	raw := "-- +goose Up\n-- +goose StatementBegin\nSELECT 1;\n-- +goose StatementEnd\n-- +goose Down\nDROP TABLE x;\n"
	got := strings.TrimSpace(stripGooseUpSQL(raw))
	if got != "SELECT 1;" {
		t.Errorf("stripGooseUpSQL() = %q, want %q", got, "SELECT 1;")
	}
}
