package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NasTecSol/nembus-sap/contracts"
)

type SAPMigrationUseCase struct {
	pool *pgxpool.Pool
}

func NewSAPMigrationUseCase(pool *pgxpool.Pool) *SAPMigrationUseCase {
	return &SAPMigrationUseCase{pool: pool}
}

// Migration-owned support objects may not exist on databases that have not
// applied 95_sap_staging.sql / 100_sap_migration_support.sql. Without them the
// batch transaction aborts (e.g. "relation staging.sap_migration_batches does
// not exist" aborts the tx and commit fails with "commit unexpectedly resulted
// in rollback") and every domain fails. ensureMigrationSchema creates them
// best-effort, once per process, with fully idempotent DDL.
var (
	migrationSchemaOnce     sync.Once
	stockMovementConflictOK bool // arbiter index for ON CONFLICT ((metadata->>'sap_trans_num'))
	inventoryConflictOK     bool // arbiter index for ON CONFLICT (product_id, COALESCE(product_variant_id,-1), store_id)
)

func ensureMigrationSchema(pool *pgxpool.Pool) {
	migrationSchemaOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		stmts := []string{
			`INSERT INTO organizations (id, name, code, legal_name, tax_id, currency_code, fiscal_year_variant, is_active, metadata)
			VALUES (1, 'Qitaf Group', 'ORG001', 'Qitaf Group LLC', '300000000000003', 'SAR', 'CALENDAR', true, '{}')
			ON CONFLICT (id) DO NOTHING;`,
			`CREATE SCHEMA IF NOT EXISTS staging;`,
			`CREATE TABLE IF NOT EXISTS staging.sap_migration_batches (
				id SERIAL PRIMARY KEY,
				batch_id VARCHAR(100) UNIQUE NOT NULL,
				run_id VARCHAR(100) NOT NULL,
				organization_id INTEGER NOT NULL,
				domain VARCHAR(50) NOT NULL,
				record_count INTEGER NOT NULL DEFAULT 0,
				status VARCHAR(30) DEFAULT 'staged',
				error_message TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,
			`CREATE TABLE IF NOT EXISTS customer_payments (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
				payment_number VARCHAR(50) UNIQUE NOT NULL,
				customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
				customer_code VARCHAR(50),
				payment_date DATE NOT NULL,
				payment_amount DECIMAL(15,2) NOT NULL CHECK (payment_amount > 0),
				payment_method VARCHAR(100),
				payment_reference VARCHAR(255),
				currency_code VARCHAR(3) DEFAULT 'SAR',
				reason VARCHAR(50) DEFAULT 'on_account',
				notes TEXT,
				metadata JSONB DEFAULT '{}',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,
			`CREATE TABLE IF NOT EXISTS vendor_payments (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
				payment_number VARCHAR(50) UNIQUE NOT NULL,
				partner_id INTEGER REFERENCES business_partners(id) ON DELETE SET NULL,
				supplier_code VARCHAR(50),
				purchase_order_id INTEGER REFERENCES purchase_orders(id) ON DELETE SET NULL,
				po_doc_entry BIGINT,
				payment_date DATE NOT NULL,
				payment_amount DECIMAL(15,2) NOT NULL CHECK (payment_amount > 0),
				payment_method VARCHAR(100),
				payment_reference VARCHAR(255),
				currency_code VARCHAR(3) DEFAULT 'SAR',
				notes TEXT,
				metadata JSONB DEFAULT '{}',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,
			// A25: arbiter for the stock-movement replay-safe upsert
			`CREATE UNIQUE INDEX IF NOT EXISTS stock_movements_sap_trans_num_uk
				ON stock_movements ((metadata->>'sap_trans_num'))
				WHERE metadata ? 'sap_trans_num';`,
			// A24: arbiter for the inventory_stock upsert (same definition as
			// idx_inventory_stock_unique_product_variant_store in 90)
			`CREATE UNIQUE INDEX IF NOT EXISTS idx_inventory_stock_unique_product_variant_store
				ON inventory_stock (product_id, COALESCE(product_variant_id, -1), store_id);`,
			// A27: metadata lookup indexes for per-row document resolution
			`CREATE INDEX IF NOT EXISTS idx_purchase_orders_sap_doc_entry ON purchase_orders ((metadata->>'sap_doc_entry'));`,
			`CREATE INDEX IF NOT EXISTS idx_goods_receipt_notes_sap_doc_entry ON goods_receipt_notes ((metadata->>'sap_doc_entry'));`,
			`CREATE INDEX IF NOT EXISTS idx_invoices_sap_doc_entry ON invoices ((metadata->>'sap_doc_entry'));`,
			`CREATE INDEX IF NOT EXISTS idx_invoices_sap_doc_num ON invoices ((metadata->>'sap_doc_num'));`,
			`CREATE INDEX IF NOT EXISTS idx_sales_returns_sap_doc_num ON sales_returns ((metadata->>'sap_doc_num'));`,
			`CREATE INDEX IF NOT EXISTS idx_transfer_requests_sap_doc_num ON transfer_requests ((metadata->>'sap_doc_num'));`,
			`CREATE INDEX IF NOT EXISTS idx_sales_orders_v2_sap_doc_num ON sales_orders_v2 ((metadata->>'sap_doc_num'));`,
		}

		for _, stmt := range stmts {
			if _, err := pool.Exec(ctx, stmt); err != nil {
				// Non-fatal: stock_movement upserts fall back to a
				// WHERE NOT EXISTS strategy; staging writes become advisory.
				log.Printf("[SAPMigration] warning: ensure migration schema: %v", err)
			}
		}

		// Decide the upsert strategy: ON CONFLICT requires the exact arbiter
		// index; on databases where its creation failed (e.g. pre-existing
		// duplicate rows by sap_trans_num) fall back to NOT EXISTS / manual
		// upserts so ingestion still works.
		stockMovementConflictOK = indexExists(ctx, pool, "stock_movements", "%sap_trans_num%")
		inventoryConflictOK = indexExists(ctx, pool, "inventory_stock", "idx_inventory_stock_unique_product_variant_store")
	})
}

func indexExists(ctx context.Context, pool *pgxpool.Pool, table, nameOrDef string) bool {
	var n int
	err := pool.QueryRow(ctx,
		`SELECT COUNT(1) FROM pg_indexes WHERE schemaname = 'public' AND tablename = $1 AND (indexname = $2 OR indexdef LIKE $2)`,
		table, nameOrDef,
	).Scan(&n)
	return err == nil && n > 0
}

func execWithSavepoint(ctx context.Context, tx pgx.Tx, query string, args ...interface{}) error {
	_, err := execWithSavepointRows(ctx, tx, query, args...)
	return err
}

// execWithSavepointRows behaves like execWithSavepoint but also returns the
// CommandTag so callers can distinguish "0 rows inserted" (e.g. an unmatched
// lookup SELECT or an idempotent ON CONFLICT DO NOTHING) from a real insert.
// A29: prevents silent drops being counted as staged records.
func execWithSavepointRows(ctx context.Context, tx pgx.Tx, query string, args ...interface{}) (pgconn.CommandTag, error) {
	spName := "sp_row"
	if _, err := tx.Exec(ctx, "SAVEPOINT "+spName); err != nil {
		return pgconn.CommandTag{}, err
	}
	tag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		_, _ = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+spName)
		return pgconn.CommandTag{}, err
	}
	_, _ = tx.Exec(ctx, "RELEASE SAVEPOINT "+spName)
	return tag, nil
}

func queryRowWithSavepoint(ctx context.Context, tx pgx.Tx, scanFn func(row pgx.Row) error, query string, args ...interface{}) error {
	spName := "sp_row"
	if _, err := tx.Exec(ctx, "SAVEPOINT "+spName); err != nil {
		return err
	}
	err := scanFn(tx.QueryRow(ctx, query, args...))
	if err != nil {
		_, _ = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+spName)
		return err
	}
	_, _ = tx.Exec(ctx, "RELEASE SAVEPOINT "+spName)
	return nil
}

func (uc *SAPMigrationUseCase) IngestBatch(ctx context.Context, orgID int, payload *contracts.MigrationBatchPayload) (*contracts.MigrationBatchResponse, error) {
	if payload.OrganizationID <= 0 {
		payload.OrganizationID = orgID
	}

	// Best-effort self-heal of the migration support schema (staging audit
	// tables, payment audit tables, arbiter indexes). Missing objects used to
	// abort the batch transaction ("commit unexpectedly resulted in rollback").
	ensureMigrationSchema(uc.pool)

	tx, err := uc.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin postgres transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	staged := 0
	failed := 0
	skipped := 0 // A29: rows intentionally not inserted (replay no-ops, unmatched lookups)
	var errs []string

	// 1. Record Batch in Staging Audit Table — advisory only: a staging
	// failure (missing schema, permissions) must never abort the batch.
	batchQuery := `
	INSERT INTO staging.sap_migration_batches (batch_id, run_id, organization_id, domain, record_count, status)
	VALUES ($1, $2, $3, $4, $5, 'staged')
	ON CONFLICT(batch_id) DO UPDATE SET record_count = excluded.record_count;
	`
	if err := execWithSavepoint(ctx, tx, batchQuery, payload.BatchID, payload.RunID, payload.OrganizationID, string(payload.Domain), payload.RecordCount()); err != nil {
		errs = append(errs, fmt.Sprintf("staging batch record (advisory, batch continues): %v", err))
	}

	// 2. Route Domain Ingestion & Canonical Upsert
	switch payload.Domain {
	case contracts.DomainStores:
		for _, s := range payload.Stores {
			metaBytes, _ := json.Marshal(s.Metadata)
			q := `
			INSERT INTO stores (organization_id, code, name, store_type, is_warehouse, is_pos_enabled, is_active, timezone, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT(organization_id, code) DO UPDATE SET
				name = excluded.name,
				is_warehouse = excluded.is_warehouse,
				is_active = excluded.is_active,
				metadata = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, q, payload.OrganizationID, s.Code, s.Name, s.StoreType, s.IsWarehouse, s.IsPosEnabled, s.IsActive, s.Timezone, metaBytes); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("store %s error: %v", s.Code, err))
			} else {
				staged++
			}
		}

		for _, loc := range payload.Locations {
			metaBytes, _ := json.Marshal(loc.Metadata)
			q := `
			INSERT INTO storage_locations (store_id, code, name, location_type, is_active, metadata)
			SELECT id, $2, $3, $4, $5, $6
			FROM stores WHERE organization_id = $1 AND code = $7
			ON CONFLICT(store_id, code) DO UPDATE SET
				name = excluded.name,
				is_active = excluded.is_active,
				metadata = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, q, payload.OrganizationID, loc.Code, loc.Name, loc.LocationType, loc.IsActive, metaBytes, loc.StoreCode); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("location %s error: %v", loc.Code, err))
			} else {
				staged++
			}
		}

	case contracts.DomainUsers:
		for _, u := range payload.Users {
			if u.Metadata == nil {
				u.Metadata = make(map[string]interface{})
			}
			isSentinel := u.PasswordHash == "{SAP_IMPORT_MUST_RESET}" || u.PasswordHash == "$2a$10$"
			if isSentinel {
				u.Metadata["must_reset_password"] = true
				u.Metadata["sap_imported"] = true
			}
			metaBytes, _ := json.Marshal(u.Metadata)
			q := `
			INSERT INTO users (organization_id, username, email, password_hash, first_name, last_name, employee_code, is_active, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT(username) DO UPDATE SET
				first_name = excluded.first_name,
				last_name = excluded.last_name,
				is_active = excluded.is_active,
				metadata = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, q, payload.OrganizationID, u.Username, u.Email, u.PasswordHash, u.FirstName, u.LastName, u.EmployeeCode, u.IsActive, metaBytes); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("user %s error: %v", u.Username, err))
			} else {
				staged++
				if isSentinel {
					resetQ := `
					DO $$
					BEGIN
						IF EXISTS (
							SELECT 1 FROM information_schema.columns 
							WHERE table_name = 'users' AND column_name = 'must_reset_password'
						) THEN
							UPDATE users 
							SET must_reset_password = true, sap_imported = true 
							WHERE username = $1;
						END IF;
					END $$;
					`
					_ = execWithSavepoint(ctx, tx, resetQ, u.Username)
				}
			}
		}

		for _, c := range payload.Cashiers {
			metaBytes, _ := json.Marshal(c.Metadata)
			// A32: honour CanonicalCashier.StoreCode instead of binding every
			// cashier to an arbitrary first store. Falls back to the first
			// store when the agent did not provide one.
			var storeID *int64
			if c.StoreCode != "" {
				r := tx.QueryRow(ctx, `SELECT id FROM stores WHERE organization_id = $1 AND code = $2 LIMIT 1`, payload.OrganizationID, c.StoreCode)
				_ = r.Scan(&storeID)
			}
			if storeID == nil {
				r := tx.QueryRow(ctx, `SELECT id FROM stores WHERE organization_id = $1 ORDER BY id LIMIT 1`, payload.OrganizationID)
				_ = r.Scan(&storeID)
			}
			if storeID == nil {
				failed++
				errs = append(errs, fmt.Sprintf("cashier %s error: no store available", c.CashierCode))
				continue
			}

			q := `
			INSERT INTO cashiers (user_id, store_id, cashier_code, drawer_limit, discount_limit, is_active, metadata)
			SELECT u.id, $2, $3, $4, $5, $6, $7
			FROM users u
			WHERE u.username = $1
			ON CONFLICT(store_id, cashier_code) DO UPDATE SET
				is_active = excluded.is_active,
				metadata = excluded.metadata;
			`
			tag, err := execWithSavepointRows(ctx, tx, q, c.Username, *storeID, c.CashierCode, c.DrawerLimit, c.DiscountLimit, c.IsActive, metaBytes)
			if err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("cashier %s error: %v", c.CashierCode, err))
			} else if tag.RowsAffected() == 0 {
				// A29: the cashier's user was not found (users domain must run first)
				skipped++
				errs = append(errs, fmt.Sprintf("cashier %s skipped: user %s not found (users domain must be ingested first)", c.CashierCode, c.Username))
			} else {
				staged++
				// A32: grant the cashier access to the mapped store
				_ = execWithSavepoint(ctx, tx, `
				INSERT INTO user_store_access (user_id, store_id, is_primary)
				SELECT u.id, $2, true
				FROM users u WHERE u.username = $1
				ON CONFLICT (user_id, store_id) DO NOTHING;
				`, c.Username, *storeID)
			}
		}

	case contracts.DomainUOM:
		for _, u := range payload.UOMs {
			metaBytes, _ := json.Marshal(u.Metadata)
			q := `
			INSERT INTO units_of_measure (code, name, uom_type, decimal_places, is_active, metadata)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT(code) DO UPDATE SET
				name = excluded.name,
				uom_type = excluded.uom_type,
				decimal_places = excluded.decimal_places,
				is_active = excluded.is_active,
				metadata = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, q, u.Code, u.Name, u.UOMType, u.DecimalPlaces, u.IsActive, metaBytes); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("uom %s error: %v", u.Code, err))
			} else {
				staged++
			}
		}

	case contracts.DomainUOMGroups:
		for _, grp := range payload.UOMGroups {
			if grp.BaseUOMCode != "" {
				_ = execWithSavepoint(ctx, tx, `
					INSERT INTO units_of_measure (code, name, uom_type, decimal_places, is_active)
					VALUES ($1, $1, 'unit', 2, true)
					ON CONFLICT(code) DO NOTHING;
				`, grp.BaseUOMCode)
			}

			tq := `
			INSERT INTO uom_packaging_templates (organization_id, uom_id, name, code, is_active)
			VALUES (
				$1,
				(SELECT id FROM units_of_measure WHERE code = $2 LIMIT 1),
				$3, $4, $5
			)
			ON CONFLICT(code) DO UPDATE SET
				name = excluded.name,
				uom_id = excluded.uom_id,
				is_active = excluded.is_active
			RETURNING id;
			`
			var templateID int
			err := queryRowWithSavepoint(ctx, tx, func(row pgx.Row) error {
				return row.Scan(&templateID)
			}, tq, payload.OrganizationID, grp.BaseUOMCode, grp.Name, grp.Code, grp.IsActive)
			if err != nil {
				_ = queryRowWithSavepoint(ctx, tx, func(row pgx.Row) error {
					return row.Scan(&templateID)
				}, `SELECT id FROM uom_packaging_templates WHERE code = $1`, grp.Code)
			}

			if templateID > 0 {
				for _, lvl := range grp.Levels {
					if lvl.UOMCode != "" {
						_ = execWithSavepoint(ctx, tx, `
							INSERT INTO units_of_measure (code, name, uom_type, decimal_places, is_active)
							VALUES ($1, $1, 'unit', 2, true)
							ON CONFLICT(code) DO NOTHING;
						`, lvl.UOMCode)
					}
					lq := `
					INSERT INTO uom_packaging_template_levels (template_id, level_order, uom_id, multiplier)
					VALUES (
						$1, $2,
						(SELECT id FROM units_of_measure WHERE code = $3 LIMIT 1),
						$4
					)
					ON CONFLICT(template_id, level_order) DO UPDATE SET
						uom_id = excluded.uom_id,
						multiplier = excluded.multiplier;
					`
					_ = execWithSavepoint(ctx, tx, lq, templateID, lvl.LevelOrder, lvl.UOMCode, lvl.Multiplier)
				}
				staged++
			} else {
				failed++
				errs = append(errs, fmt.Sprintf("uom_group %s error: failed to resolve template id: %v", grp.Code, err))
			}
		}

	case contracts.DomainCategories:
		for _, cat := range payload.Categories {
			metaBytes, _ := json.Marshal(cat.Metadata)
			q := `
			INSERT INTO product_categories (code, name, description, category_level, is_active, metadata)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT(code) DO UPDATE SET
				name = excluded.name,
				description = excluded.description,
				is_active = excluded.is_active;
			`
			if err := execWithSavepoint(ctx, tx, q, cat.Code, cat.Name, cat.Description, cat.CategoryLevel, cat.IsActive, metaBytes); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("category %s error: %v", cat.Code, err))
			} else {
				staged++
			}
		}

	case contracts.DomainBrands:
		for _, b := range payload.Brands {
			metaBytes, _ := json.Marshal(b.Metadata)
			q := `
			INSERT INTO brands (code, name, is_active, metadata)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT(code) DO UPDATE SET
				name = excluded.name,
				is_active = excluded.is_active;
			`
			if err := execWithSavepoint(ctx, tx, q, b.Code, b.Name, b.IsActive, metaBytes); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("brand %s error: %v", b.Code, err))
			} else {
				staged++
			}
		}

	case contracts.DomainProducts:
		for _, p := range payload.Products {
			metaBytes, _ := json.Marshal(p.Metadata)

			// Ensure base UOM exists in units_of_measure to guarantee valid foreign key
			if p.BaseUOMCode != "" {
				_ = execWithSavepoint(ctx, tx, `
					INSERT INTO units_of_measure (code, name, uom_type, decimal_places, is_active)
					VALUES ($1, $1, 'unit', 2, true)
					ON CONFLICT(code) DO NOTHING;
				`, p.BaseUOMCode)
			}

			q := `
			INSERT INTO products (
				organization_id, sku, name, description, category_id, brand_id, base_uom_id, product_type,
				is_serialized, is_batch_managed, is_active, is_sellable, is_purchasable, track_inventory, metadata
			)
			VALUES (
				$1, $2, $3, $4,
				(SELECT id FROM product_categories WHERE code = $5),
				(SELECT id FROM brands WHERE code = $6),
				(SELECT id FROM units_of_measure WHERE code = $7 LIMIT 1),
				$8, $9, $10, $11, $12, $13, $14, $15
			)
			ON CONFLICT(organization_id, sku) DO UPDATE SET
				name = excluded.name,
				description = excluded.description,
				category_id = excluded.category_id,
				brand_id = excluded.brand_id,
				base_uom_id = excluded.base_uom_id,
				product_type = excluded.product_type,
				is_active = excluded.is_active,
				is_sellable = excluded.is_sellable,
				track_inventory = excluded.track_inventory,
				metadata = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, q, payload.OrganizationID, p.SKU, p.Name, p.Description, p.CategoryCode, p.BrandCode, p.BaseUOMCode, p.ProductType, p.IsSerialized, p.IsBatchManaged, p.IsActive, p.IsSellable, p.IsPurchasable, p.TrackInventory, metaBytes); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("product %s error: %v", p.SKU, err))
			} else {
				staged++
				if p.PrimaryBarcode != "" {
					bq := `
					INSERT INTO product_barcodes (product_id, barcode, barcode_type, is_primary)
					SELECT id, $2, 'EAN13', true
					FROM products WHERE organization_id = $1 AND sku = $3
					ON CONFLICT(barcode) DO UPDATE SET is_primary = true;
					`
					_ = execWithSavepoint(ctx, tx, bq, payload.OrganizationID, p.PrimaryBarcode, p.SKU)
				}

				// Insert product UOM conversions
				for _, conv := range p.UOMConversions {
					convMeta, _ := json.Marshal(conv.Metadata)
					_ = execWithSavepoint(ctx, tx, `
						INSERT INTO units_of_measure (code, name, uom_type, decimal_places, is_active)
						VALUES ($1, $1, 'unit', 2, true)
						ON CONFLICT(code) DO NOTHING;
					`, conv.FromUOMCode)
					_ = execWithSavepoint(ctx, tx, `
						INSERT INTO units_of_measure (code, name, uom_type, decimal_places, is_active)
						VALUES ($1, $1, 'unit', 2, true)
						ON CONFLICT(code) DO NOTHING;
					`, conv.ToUOMCode)

					cq := `
					INSERT INTO product_uom_conversions (product_id, from_uom_id, to_uom_id, conversion_factor, is_default, metadata)
					SELECT 
						p.id,
						fu.id,
						tu.id,
						$4, $5, $6
					FROM products p
					CROSS JOIN units_of_measure fu
					CROSS JOIN units_of_measure tu
					WHERE p.organization_id = $1 AND p.sku = $2
					  AND fu.code = $3 AND tu.code = $7
					ON CONFLICT(product_id, from_uom_id, to_uom_id) DO UPDATE SET
						conversion_factor = excluded.conversion_factor,
						is_default = excluded.is_default,
						metadata = excluded.metadata;
					`
					_ = execWithSavepoint(ctx, tx, cq, payload.OrganizationID, p.SKU, conv.FromUOMCode, conv.ConversionFactor, conv.IsDefault, convMeta, conv.ToUOMCode)
				}
			}
		}

	case contracts.DomainBarcodes:
		for _, b := range payload.Barcodes {
			bq := `
			INSERT INTO product_barcodes (product_id, barcode, barcode_type, is_primary)
			SELECT id, $2, $3, $4
			FROM products WHERE organization_id = $1 AND sku = $5
			ON CONFLICT(barcode) DO UPDATE SET is_primary = excluded.is_primary;
			`
			tag, err := execWithSavepointRows(ctx, tx, bq, payload.OrganizationID, b.Barcode, b.BarcodeType, b.IsPrimary, b.ProductSKU)
			if err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("barcode %s error: %v", b.Barcode, err))
			} else if tag.RowsAffected() == 0 {
				skipped++
			} else {
				staged++
			}
		}

	case contracts.DomainPriceLists:
		for _, pl := range payload.PriceLists {
			metaBytes, _ := json.Marshal(pl.Metadata)
			q := `
			INSERT INTO price_lists (name, code, currency_code, is_active, metadata)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT(code) DO UPDATE SET
				name = excluded.name,
				currency_code = excluded.currency_code,
				is_active = excluded.is_active,
				metadata = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, q, pl.Name, pl.Code, pl.CurrencyCode, pl.IsActive, metaBytes); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("price_list %s error: %v", pl.Code, err))
			} else {
				staged++
			}
		}

		for _, item := range payload.PriceItems {
			metaBytes, _ := json.Marshal(item.Metadata)
			_ = execWithSavepoint(ctx, tx, `
				DELETE FROM product_prices 
				WHERE product_id = (SELECT id FROM products WHERE organization_id = $1 AND sku = $2)
				  AND price_list_id = (SELECT id FROM price_lists WHERE code = $3);
			`, payload.OrganizationID, item.ProductSKU, item.PriceListCode)

			q := `
			INSERT INTO product_prices (product_id, price_list_id, uom_id, price, is_active, metadata)
			SELECT p.id, pl.id, (SELECT id FROM units_of_measure WHERE code = $6 LIMIT 1), $3, true, $4
			FROM products p
			CROSS JOIN price_lists pl
			WHERE p.organization_id = $1 AND p.sku = $2 AND pl.code = $5
			LIMIT 1;
			`
			tag, err := execWithSavepointRows(ctx, tx, q, payload.OrganizationID, item.ProductSKU, item.Price, metaBytes, item.PriceListCode, item.UOMCode)
			if err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("price_item %s@%s error: %v", item.ProductSKU, item.PriceListCode, err))
			} else if tag.RowsAffected() == 0 {
				skipped++
			} else {
				staged++
			}
		}

	case contracts.DomainInventory:
		for _, inv := range payload.Inventory {
			metaBytes, _ := json.Marshal(inv.Metadata)
			// A07: Resolve product_id and store_id first, then use VALUES with ON CONFLICT.
			// inventory_stock has no organization_id column — unique key is (product_id, store_id).
			var productID, storeID *int64
			row := tx.QueryRow(ctx, `SELECT id FROM products WHERE organization_id = $1 AND sku = $2 LIMIT 1`, payload.OrganizationID, inv.ProductSKU)
			row.Scan(&productID)
			row2 := tx.QueryRow(ctx, `SELECT id FROM stores WHERE organization_id = $1 AND code = $2 LIMIT 1`, payload.OrganizationID, inv.StoreCode)
			row2.Scan(&storeID)
			if productID == nil || storeID == nil {
				failed++
				errs = append(errs, fmt.Sprintf("stock %s@%s: product or store not found", inv.ProductSKU, inv.StoreCode))
				continue
			}
			if inventoryConflictOK {
				// A24: the conflict target must match the real arbiter index
				// idx_inventory_stock_unique_product_variant_store
				// (product_id, COALESCE(product_variant_id, -1), store_id) in
				// 90_views_functions.sql — a plain (product_id, store_id) target
				// fails inference with 42P10 on every row.
				q := `
				INSERT INTO inventory_stock (
					product_id, store_id, quantity_on_hand, quantity_allocated,
					quantity_available, quantity_on_order, reorder_level, max_stock_level, metadata
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				ON CONFLICT (product_id, COALESCE(product_variant_id, -1), store_id) DO UPDATE SET
					quantity_on_hand    = excluded.quantity_on_hand,
					quantity_allocated  = excluded.quantity_allocated,
					quantity_available  = excluded.quantity_available,
					quantity_on_order   = excluded.quantity_on_order,
					reorder_level       = excluded.reorder_level,
					max_stock_level     = excluded.max_stock_level,
					metadata            = excluded.metadata;
				`
				if err := execWithSavepoint(ctx, tx, q, *productID, *storeID, inv.QuantityOnHand, inv.QuantityAllocated, inv.QuantityAvailable, inv.QuantityOnOrder, inv.ReorderLevel, inv.MaxStockLevel, metaBytes); err != nil {
					failed++
					errs = append(errs, fmt.Sprintf("stock %s@%s error: %v", inv.ProductSKU, inv.StoreCode, err))
				} else {
					staged++
				}
			} else {
				// Fallback for databases without the arbiter index (e.g. legacy
				// DBs where index creation failed on duplicates): manual upsert.
				// Migration rows never carry a product_variant_id.
				tag, err := execWithSavepointRows(ctx, tx, `
				UPDATE inventory_stock SET
					quantity_on_hand = $3, quantity_allocated = $4, quantity_available = $5,
					quantity_on_order = $6, reorder_level = $7, max_stock_level = $8, metadata = $9
				WHERE product_id = $1 AND store_id = $2 AND product_variant_id IS NULL;
				`, *productID, *storeID, inv.QuantityOnHand, inv.QuantityAllocated, inv.QuantityAvailable, inv.QuantityOnOrder, inv.ReorderLevel, inv.MaxStockLevel, metaBytes)
				if err != nil {
					failed++
					errs = append(errs, fmt.Sprintf("stock %s@%s error: %v", inv.ProductSKU, inv.StoreCode, err))
				} else if tag.RowsAffected() > 0 {
					staged++
				} else if _, ierr := execWithSavepointRows(ctx, tx, `
				INSERT INTO inventory_stock (
					product_id, store_id, quantity_on_hand, quantity_allocated,
					quantity_available, quantity_on_order, reorder_level, max_stock_level, metadata
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
				`, *productID, *storeID, inv.QuantityOnHand, inv.QuantityAllocated, inv.QuantityAvailable, inv.QuantityOnOrder, inv.ReorderLevel, inv.MaxStockLevel, metaBytes); ierr != nil {
					failed++
					errs = append(errs, fmt.Sprintf("stock %s@%s error: %v", inv.ProductSKU, inv.StoreCode, ierr))
				} else {
					staged++
				}
			}
		}

	case contracts.DomainPartners:
		for _, pt := range payload.Partners {
			if pt.Metadata == nil {
				pt.Metadata = make(map[string]interface{})
			}
			pt.Metadata["email"] = pt.Email
			pt.Metadata["phone"] = pt.Phone
			pt.Metadata["tax_id"] = pt.TaxID
			metaBytes, _ := json.Marshal(pt.Metadata)

			// A34.1: partner_role is passed through (supplier | customer | lead);
			// the CHECK constraint (relaxed by 97_unify_business_partners.sql)
			// allows supplier/vendor/customer/lead/special_customer/corporate_group.
			role := pt.PartnerType
			if role == "" {
				role = "supplier"
			}

			// A19/A23: credit_limit, outstanding_balance (SAP OCRD.Balance) and
			// currency_code now land in business_partners; payment_terms_id is
			// resolved from the payment_terms domain (PT-{GroupNum} codes).
			bpQ := `
			INSERT INTO business_partners (organization_id, code, name, partner_role, tax_id, currency_code, credit_limit, outstanding_balance, payment_terms_id, is_active, metadata)
			VALUES ($1, $2, $3, $4, $5, COALESCE(NULLIF($6, ''), 'SAR'), $7, $8,
				(SELECT id FROM payment_terms WHERE organization_id = $1 AND code = $9 LIMIT 1),
				$10, $11)
			ON CONFLICT(code) DO UPDATE SET
				name = excluded.name,
				partner_role = excluded.partner_role,
				tax_id = excluded.tax_id,
				currency_code = COALESCE(excluded.currency_code, business_partners.currency_code),
				credit_limit = excluded.credit_limit,
				outstanding_balance = excluded.outstanding_balance,
				payment_terms_id = COALESCE(excluded.payment_terms_id, business_partners.payment_terms_id),
				is_active = excluded.is_active,
				metadata = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, bpQ, payload.OrganizationID, pt.Code, pt.Name, role, pt.TaxID, pt.CurrencyCode, pt.CreditLimit, pt.Balance, pt.PaymentTermCode, pt.IsActive, metaBytes); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("business_partner %s error: %v", pt.Code, err))
			} else {
				staged++
			}

			if pt.PartnerType == "customer" {
				// A30: the business_partner lookup is now organization-scoped so a
				// shared SAP CardCode can never bind to another org's partner.
				// A19: customers.credit_limit / outstanding_balance feed
				// vw_customer_aging_report.
				custQ := `
				INSERT INTO customers (organization_id, customer_code, name, is_active, business_partner_id, credit_limit, outstanding_balance, metadata)
				VALUES ($1, $2, $3, $4,
					(SELECT id FROM business_partners WHERE organization_id = $1 AND code = $2 LIMIT 1),
					$5, $6, $7)
				ON CONFLICT(organization_id, customer_code) DO UPDATE SET
					name = excluded.name,
					is_active = excluded.is_active,
					business_partner_id = excluded.business_partner_id,
					credit_limit = excluded.credit_limit,
					outstanding_balance = excluded.outstanding_balance,
					metadata = excluded.metadata;
				`
				_ = execWithSavepoint(ctx, tx, custQ, payload.OrganizationID, pt.Code, pt.Name, pt.IsActive, pt.CreditLimit, pt.Balance, metaBytes)
			}
		}

	case contracts.DomainBPAddresses:
		for _, addr := range payload.BPAddresses {
			metaBytes, _ := json.Marshal(addr.Metadata)
			addrType := "ship_to"
			if addr.AddressType == "billing" {
				addrType = "bill_to"
			}
			addrName := addr.AddressLine
			if addrName == "" {
				addrName = addr.Street
			}
			if addrName == "" {
				addrName = "Default"
			}

			bpAddrQ := `
			INSERT INTO partner_addresses (partner_id, address_name, address_type, street, city, state, zip_code, country_code)
			SELECT bp.id, $2, $3, $4, $5, $6, $7, COALESCE(NULLIF($8, ''), 'SA')
			FROM business_partners bp
			WHERE bp.organization_id = $1 AND bp.code = $9
			ON CONFLICT(partner_id, address_type, address_name) DO UPDATE SET
				street = excluded.street,
				city = excluded.city,
				state = excluded.state,
				zip_code = excluded.zip_code,
				country_code = excluded.country_code;
			`
			if err := execWithSavepoint(ctx, tx, bpAddrQ, payload.OrganizationID, addrName, addrType, addr.Street, addr.City, addr.State, addr.PostalCode, addr.Country, addr.PartnerCode); err != nil {
				custAddrQ := `
				INSERT INTO customer_addresses (customer_id, address_type, address_line, street, city, country, postal_code, state, phone, metadata)
				SELECT c.id, $2, $3, $4, $5, $6, $7, $8, $9, $10
				FROM customers c
				WHERE c.organization_id = $1 AND c.customer_code = $11
				ON CONFLICT(customer_id, address_type, address_line) DO UPDATE SET
					street = excluded.street,
					city = excluded.city,
					country = excluded.country,
					postal_code = excluded.postal_code,
					state = excluded.state,
					phone = excluded.phone,
					metadata = excluded.metadata;
				`
				if custErr := execWithSavepoint(ctx, tx, custAddrQ, payload.OrganizationID, addr.AddressType, addr.AddressLine, addr.Street, addr.City, addr.Country, addr.PostalCode, addr.State, addr.Phone, metaBytes, addr.PartnerCode); custErr != nil {
					failed++
					errs = append(errs, fmt.Sprintf("bp_address %s error: %v", addr.PartnerCode, err))
				} else {
					staged++
				}
			} else {
				staged++
			}
		}

	case contracts.DomainPurchaseOrders:
		for _, po := range payload.PurchaseOrders {
			if po.Metadata == nil {
				po.Metadata = make(map[string]interface{})
			}
			if po.SupplierCode != "" {
				po.Metadata["sap_supplier_code"] = po.SupplierCode
			}
			if po.StoreCode != "" {
				po.Metadata["sap_store_code"] = po.StoreCode
			}
			metaBytes, _ := json.Marshal(po.Metadata)

			// A16: purchase_orders.partners_id (NOT NULL) references
			// business_partners — the legacy `suppliers` table no longer exists.
			// Fail the row with a clear error instead of silently binding an
			// arbitrary partner.
			var partnerID *int64
			if po.SupplierCode != "" {
				r := tx.QueryRow(ctx, `SELECT id FROM business_partners WHERE organization_id = $1 AND code = $2 LIMIT 1`, payload.OrganizationID, po.SupplierCode)
				_ = r.Scan(&partnerID)
			}
			if partnerID == nil {
				failed++
				errs = append(errs, fmt.Sprintf("purchase order %s error: supplier %q not found in business_partners (partners domain must be ingested first)", po.PONumber, po.SupplierCode))
				continue
			}

			q := `
			INSERT INTO purchase_orders (
				organization_id, po_number, partners_id, store_id, po_date, expected_delivery_date,
				status, subtotal, discount_amount, tax_amount, total_amount, metadata
			)
			SELECT
				$1, $2, $3,
				COALESCE((SELECT id FROM stores WHERE organization_id = $1 AND code = $4 LIMIT 1),
				         (SELECT id FROM stores WHERE organization_id = $1 LIMIT 1)),
				$5, $6, $7, $8, $9, $10, $11, $12
			ON CONFLICT(po_number) DO UPDATE SET
				status = excluded.status,
				subtotal = excluded.subtotal,
				discount_amount = excluded.discount_amount,
				tax_amount = excluded.tax_amount,
				total_amount = excluded.total_amount,
				metadata = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, q, payload.OrganizationID, po.PONumber, *partnerID, po.StoreCode, po.PODate, po.ExpectedDeliveryDate, po.Status, po.Subtotal, po.DiscountAmount, po.TaxAmount, po.TotalAmount, metaBytes); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("purchase order %s error: %v", po.PONumber, err))
			} else {
				staged++
				// Clean and insert lines idempotently
				delQ := `
				DELETE FROM purchase_order_lines 
				WHERE purchase_order_id = (SELECT id FROM purchase_orders WHERE organization_id = $1 AND po_number = $2);
				`
				_ = execWithSavepoint(ctx, tx, delQ, payload.OrganizationID, po.PONumber)

				for _, line := range po.Lines {
					lineMeta, _ := json.Marshal(line.Metadata)
					lq := `
					INSERT INTO purchase_order_lines (
						purchase_order_id, product_id, quantity, uom_id, unit_price,
						discount_amount, tax_amount, subtotal, line_total, received_quantity, line_number, metadata
					)
					SELECT
						po.id,
						(SELECT id FROM products WHERE organization_id = $2 AND sku = $3 LIMIT 1),
						$4,
						(SELECT id FROM units_of_measure WHERE code = $5 LIMIT 1),
						$6, $7, $8, $9, $10, $11, $12, $13
					FROM purchase_orders po
					WHERE po.po_number = $1 AND po.organization_id = $2
					  AND (SELECT id FROM products WHERE organization_id = $2 AND sku = $3 LIMIT 1) IS NOT NULL;
					`
					// A29: a NULL product (unmatched SKU) yields 0 inserted rows — surface it
					tag, lErr := execWithSavepointRows(ctx, tx, lq, po.PONumber, payload.OrganizationID, line.ProductSKU, line.Quantity, line.UOMCode, line.UnitPrice, line.DiscountAmount, line.TaxAmount, line.Subtotal, line.LineTotal, line.ReceivedQuantity, line.LineNumber, lineMeta)
					if lErr != nil {
						failed++
						errs = append(errs, fmt.Sprintf("purchase order %s line %d error: %v", po.PONumber, line.LineNumber, lErr))
					} else if tag.RowsAffected() == 0 {
						skipped++
						errs = append(errs, fmt.Sprintf("purchase order %s line %d skipped: product %q not found", po.PONumber, line.LineNumber, line.ProductSKU))
					}
				}
			}
		}

	case contracts.DomainGoodsReceipts, "goods_receipts":
		for _, grn := range payload.GoodsReceipts {
			if grn.Metadata == nil {
				grn.Metadata = make(map[string]interface{})
			}
			if grn.SupplierCode != "" {
				grn.Metadata["sap_supplier_code"] = grn.SupplierCode
			}
			if grn.StoreCode != "" {
				grn.Metadata["sap_store_code"] = grn.StoreCode
			}
			metaBytes, _ := json.Marshal(grn.Metadata)

			// A16-consistency: partners_id is NOT NULL; fail with a clear error
			// instead of binding an arbitrary business partner.
			var partnerID *int64
			if grn.SupplierCode != "" {
				r := tx.QueryRow(ctx, `SELECT id FROM business_partners WHERE organization_id = $1 AND code = $2 LIMIT 1`, payload.OrganizationID, grn.SupplierCode)
				_ = r.Scan(&partnerID)
			}
			if partnerID == nil {
				failed++
				errs = append(errs, fmt.Sprintf("goods receipt %s error: supplier %q not found in business_partners", grn.GRNNumber, grn.SupplierCode))
				continue
			}

			q := `
			INSERT INTO goods_receipt_notes (
				organization_id, grn_number, purchase_order_id, partners_id, store_id,
				receipt_date, delivery_note_number, status, notes, metadata
			)
			SELECT
				$1, $2,
				-- A08: look up PO by sap_doc_entry in metadata, not by po_number string
				(SELECT id FROM purchase_orders WHERE organization_id = $1 AND metadata->>'sap_doc_entry' = $3 LIMIT 1),
				$4,
				COALESCE((SELECT id FROM stores WHERE organization_id = $1 AND code = $5 LIMIT 1),
				         (SELECT id FROM stores WHERE organization_id = $1 LIMIT 1)),
				$6, $7, $8, $9, $10
			ON CONFLICT(grn_number) DO UPDATE SET
				purchase_order_id = COALESCE(excluded.purchase_order_id, goods_receipt_notes.purchase_order_id),
				status = excluded.status,
				notes = excluded.notes,
				metadata = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, q, payload.OrganizationID, grn.GRNNumber, grn.PONumber, *partnerID, grn.StoreCode, grn.ReceiptDate, grn.DeliveryNoteNumber, grn.Status, grn.Notes, metaBytes); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("goods receipt %s error: %v", grn.GRNNumber, err))
			} else {
				staged++
				// Clean and insert items idempotently
				delQ := `
				DELETE FROM goods_receipt_note_items
				WHERE grn_id = (SELECT id FROM goods_receipt_notes WHERE organization_id = $1 AND grn_number = $2);
				`
				_ = execWithSavepoint(ctx, tx, delQ, payload.OrganizationID, grn.GRNNumber)

				for _, item := range grn.Items {
					iq := `
					INSERT INTO goods_receipt_note_items (
						grn_id, purchase_order_line_id, product_id, quantity_received, quantity_rejected,
						uom_id, unit_cost, notes
					)
					SELECT
						g.id,
						-- A08: look up PO line via sap_doc_entry on the parent PO
						(SELECT pol.id FROM purchase_order_lines pol
						 JOIN purchase_orders po ON pol.purchase_order_id = po.id
						 WHERE po.organization_id = $2 AND po.metadata->>'sap_doc_entry' = $3 AND pol.line_number = $4 LIMIT 1),
						(SELECT id FROM products WHERE organization_id = $2 AND sku = $5 LIMIT 1),
						$6, $7,
						(SELECT id FROM units_of_measure WHERE code = $8 LIMIT 1),
						$9, $10
					FROM goods_receipt_notes g
					WHERE g.grn_number = $1 AND g.organization_id = $2
					  AND (SELECT id FROM products WHERE organization_id = $2 AND sku = $5 LIMIT 1) IS NOT NULL;
					`
					_ = execWithSavepoint(ctx, tx, iq, grn.GRNNumber, payload.OrganizationID, grn.PONumber, item.SourcePOLineNum, item.ProductSKU, item.QuantityReceived, item.QuantityRejected, item.UOMCode, item.UnitCost, item.Notes)
				}
			}
		}

	case contracts.DomainStockMovements:
		for _, sm := range payload.StockMovements {
			metaBytes, _ := json.Marshal(sm.Metadata)
			// A09: Resolve product_id and store IDs first for idempotent insert
			var productID *int64
			pRow := tx.QueryRow(ctx, `SELECT id FROM products WHERE organization_id = $1 AND sku = $2 LIMIT 1`, payload.OrganizationID, sm.ProductSKU)
			pRow.Scan(&productID)
			if productID == nil {
				failed++
				errs = append(errs, fmt.Sprintf("stock movement %s: product not found", sm.ProductSKU))
				continue
			}
			var fromStoreID, toStoreID *int64
			if sm.FromStoreCode != "" {
				r := tx.QueryRow(ctx, `SELECT id FROM stores WHERE organization_id = $1 AND code = $2 LIMIT 1`, payload.OrganizationID, sm.FromStoreCode)
				r.Scan(&fromStoreID)
			}
			if sm.ToStoreCode != "" {
				r := tx.QueryRow(ctx, `SELECT id FROM stores WHERE organization_id = $1 AND code = $2 LIMIT 1`, payload.OrganizationID, sm.ToStoreCode)
				r.Scan(&toStoreID)
			}

			// A27: resolve reference_id by document type so the movement ledger
			// (vw_stock_movement_ledger) is not left with dangling references.
			// Requires sales_returns / transfer_requests / goods_receipts /
			// invoices domains to be ingested before stock_movements.
			var referenceID *int64
			if sm.ReferenceNumber != "" {
				switch sm.ReferenceType {
				case "goods_receipt_note":
					r := tx.QueryRow(ctx, `SELECT id FROM goods_receipt_notes WHERE organization_id = $1 AND metadata->>'sap_doc_num' = $2 LIMIT 1`, payload.OrganizationID, sm.ReferenceNumber)
					_ = r.Scan(&referenceID)
				case "sales_return":
					r := tx.QueryRow(ctx, `SELECT id FROM sales_returns WHERE organization_id = $1 AND metadata->>'sap_doc_num' = $2 LIMIT 1`, payload.OrganizationID, sm.ReferenceNumber)
					_ = r.Scan(&referenceID)
				case "transfer_request":
					r := tx.QueryRow(ctx, `SELECT id FROM transfer_requests WHERE organization_id = $1 AND metadata->>'sap_doc_num' = $2 LIMIT 1`, payload.OrganizationID, sm.ReferenceNumber)
					_ = r.Scan(&referenceID)
				case "invoice":
					r := tx.QueryRow(ctx, `SELECT id FROM invoices WHERE organization_id = $1 AND metadata->>'sap_doc_num' = $2 LIMIT 1`, payload.OrganizationID, sm.ReferenceNumber)
					_ = r.Scan(&referenceID)
				}
			}

			var tag pgconn.CommandTag
			var err error
			if stockMovementConflictOK {
				// A25: arbiter index exists — idempotent upsert via ON CONFLICT
				q := `
				INSERT INTO stock_movements (
					movement_type, reference_type, reference_id, product_id,
					from_store_id, to_store_id, quantity,
					movement_date, status, cost_per_unit, total_value, metadata
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'completed', $9, $10, $11)
				ON CONFLICT ((metadata->>'sap_trans_num')) WHERE metadata ? 'sap_trans_num'
				DO NOTHING;
				`
				tag, err = execWithSavepointRows(ctx, tx, q, sm.MovementType, sm.ReferenceType, referenceID, *productID, fromStoreID, toStoreID, sm.Quantity, sm.MovementDate, sm.CostPerUnit, sm.TotalValue, metaBytes)
			} else {
				// Fallback for databases where the arbiter index could not be
				// created (e.g. pre-existing duplicate sap_trans_num rows):
				// replay-safe via WHERE NOT EXISTS, no index required.
				q := `
				INSERT INTO stock_movements (
					movement_type, reference_type, reference_id, product_id,
					from_store_id, to_store_id, quantity,
					movement_date, status, cost_per_unit, total_value, metadata
				)
				SELECT $1, $2, $3, $4, $5, $6, $7, $8, 'completed', $9, $10, $11
				WHERE NOT EXISTS (
					SELECT 1 FROM stock_movements WHERE metadata->>'sap_trans_num' = $12
				);
				`
				tag, err = execWithSavepointRows(ctx, tx, q, sm.MovementType, sm.ReferenceType, referenceID, *productID, fromStoreID, toStoreID, sm.Quantity, sm.MovementDate, sm.CostPerUnit, sm.TotalValue, metaBytes, fmt.Sprintf("%d", sm.Metadata["sap_trans_num"]))
			}
			if err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("stock movement item %s error: %v", sm.ProductSKU, err))
			} else if tag.RowsAffected() == 0 {
				// A29: idempotent replay no-op (movement already present)
				skipped++
			} else {
				staged++
			}
		}

	case contracts.DomainSalesOrders:
		for _, so := range payload.SalesOrders {
			// Enrich metadata with SAP customer code for later reconciliation
			if so.Metadata == nil {
				so.Metadata = make(map[string]interface{})
			}
			if so.CustomerCode != "" {
				so.Metadata["sap_customer_code"] = so.CustomerCode
			}
			metaBytes, _ := json.Marshal(so.Metadata)
			q := `
			INSERT INTO sales_orders_v2 (
				order_number, organization_id, store_id, customer_id, customer_name, order_status, payment_status, fulfillment_status,
				order_date, expected_delivery_date, subtotal, discount_amount, tax_amount, total_amount, internal_notes, shipping_address, billing_address, metadata
			)
			VALUES ($1, $2,
				-- A18: resolve store from the header warehouse and customer from
				-- the customers master so orders are filterable by store and
				-- joinable to customer profiles.
				(SELECT id FROM stores WHERE organization_id = $2 AND code = $15 LIMIT 1),
				(SELECT id FROM customers WHERE organization_id = $2 AND customer_code = $16 LIMIT 1),
				$3, $4::order_status_v2, $5::payment_status, $6::fulfillment_status, $7, $8, $9, $10, $11, $12, $13, '{}'::jsonb, '{}'::jsonb, $14)
			ON CONFLICT(order_number) DO UPDATE SET
				total_amount = excluded.total_amount,
				store_id = COALESCE(excluded.store_id, sales_orders_v2.store_id),
				customer_id = COALESCE(excluded.customer_id, sales_orders_v2.customer_id),
				metadata = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, q, so.OrderNumber, payload.OrganizationID, so.CustomerName, so.OrderStatus, so.PaymentStatus, so.FulfillmentStatus, so.OrderDate, so.ExpectedDate, so.Subtotal, so.DiscountAmount, so.TaxAmount, so.TotalAmount, so.Notes, metaBytes, so.StoreCode, so.CustomerCode); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("sales order %s error: %v", so.OrderNumber, err))
			} else {
				staged++
				for _, line := range so.Lines {
					lq := `
					INSERT INTO sales_order_lines_v2 (
						sales_order_id, organization_id, line_number, product_id, product_name, product_sku,
						quantity_ordered, unit_price, tax_amount, line_total
					)
					SELECT 
						so.id, $2, $3,
						(SELECT id FROM products WHERE organization_id = $2 AND sku = $5 LIMIT 1),
						$4, $5, $6, $7, $8, $9
					FROM sales_orders_v2 so
					WHERE so.order_number = $1 AND so.organization_id = $2
					ON CONFLICT(sales_order_id, line_number) DO UPDATE SET
						quantity_ordered = excluded.quantity_ordered,
						unit_price = excluded.unit_price,
						tax_amount = excluded.tax_amount,
						line_total = excluded.line_total;
					`
					// A29: product_id is NOT NULL on sales_order_lines_v2 — an
					// unmatched SKU must be surfaced, not silently lost
					if _, lErr := execWithSavepointRows(ctx, tx, lq, so.OrderNumber, payload.OrganizationID, line.LineNumber, line.ProductName, line.ProductSKU, line.Quantity, line.UnitPrice, line.TaxAmount, line.LineTotal); lErr != nil {
						failed++
						errs = append(errs, fmt.Sprintf("sales order %s line %d error: %v", so.OrderNumber, line.LineNumber, lErr))
					}
				}
			}
		}

	case contracts.DomainInvoices:
		for _, inv := range payload.Invoices {
			// Store customer_code in metadata so invoices can be reconciled
			// after customers are imported. customer_id is nullable to support
			// SAP imports where the matching customer record may not yet exist.
			if inv.Metadata == nil {
				inv.Metadata = make(map[string]interface{})
			}
			if inv.CustomerCode != "" {
				inv.Metadata["sap_customer_code"] = inv.CustomerCode
			}
			metaBytes, _ := json.Marshal(inv.Metadata)
			q := `
			INSERT INTO invoices (
				invoice_number, organization_id, store_id, customer_id, customer_name, customer_tax_id, invoice_type, invoice_status,
				currency_code, exchange_rate,
				invoice_date, due_date, subtotal, discount_amount, tax_amount, total_amount, paid_amount, balance_due, billing_address, metadata
			)
			SELECT
				$1, $2,
				-- A17: invoices.store_id is mandatory for vw_realtime_pnl and
				-- vw_tax_vat_summary (both INNER JOIN stores); fall back to the
				-- organization's first store when INV1 carries no warehouse.
				COALESCE((SELECT id FROM stores WHERE organization_id = $2 AND code = $18 LIMIT 1),
				         (SELECT id FROM stores WHERE organization_id = $2 LIMIT 1)),
				(SELECT id FROM customers WHERE organization_id = $2 AND customer_code = $3 LIMIT 1),
				$4,
				-- A34.4: ZATCA-relevant tax id from the partner master
				(SELECT tax_id FROM business_partners WHERE organization_id = $2 AND code = $3 LIMIT 1),
				$5::invoice_type, $6::invoice_status,
				$7, $8,
				$9, $10, $11, $12, $13, $14, $15, $16, '{}'::jsonb, $17
			ON CONFLICT(invoice_number) DO UPDATE SET
				store_id      = COALESCE(excluded.store_id, invoices.store_id),
				customer_id   = COALESCE(excluded.customer_id, invoices.customer_id),
				currency_code = excluded.currency_code,
				paid_amount   = excluded.paid_amount,
				balance_due   = excluded.balance_due,
				metadata      = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, q, inv.InvoiceNumber, payload.OrganizationID, inv.CustomerCode, inv.CustomerName, inv.InvoiceType, inv.InvoiceStatus, inv.CurrencyCode, inv.ExchangeRate, inv.InvoiceDate, inv.DueDate, inv.Subtotal, inv.DiscountAmount, inv.TaxAmount, inv.TotalAmount, inv.PaidAmount, inv.BalanceDue, metaBytes, inv.StoreCode); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("invoice %s error: %v", inv.InvoiceNumber, err))
			} else {
				staged++
				for _, line := range inv.Lines {
					lineMeta, _ := json.Marshal(line.Metadata)
					lq := `
					INSERT INTO invoice_lines (
						invoice_id, organization_id, line_number, description, item_type, product_id, product_sku,
						quantity, unit_price, discount_amount, tax_amount, line_total, uom_id, metadata
					)
					SELECT 
						inv.id, $2, $3,
						COALESCE(NULLIF($4, ''), $5, 'Unknown'),
						$6,
						(SELECT id FROM products WHERE organization_id = $2 AND sku = $5 LIMIT 1),
						$5, $7, $8, $9, $10, $11,
						(SELECT id FROM units_of_measure WHERE code = $12 LIMIT 1),
						$13
					FROM invoices inv
					WHERE inv.invoice_number = $1 AND inv.organization_id = $2
					ON CONFLICT(invoice_id, line_number) DO UPDATE SET
						item_type       = excluded.item_type,
						quantity        = excluded.quantity,
						unit_price      = excluded.unit_price,
						discount_amount = excluded.discount_amount,
						tax_amount      = excluded.tax_amount,
						line_total      = excluded.line_total,
						uom_id          = COALESCE(excluded.uom_id, invoice_lines.uom_id),
						metadata        = excluded.metadata;
					`
					// A29: count lost lines instead of swallowing the error
					if _, lErr := execWithSavepointRows(ctx, tx, lq, inv.InvoiceNumber, payload.OrganizationID, line.LineNumber, line.ProductName, line.ProductSKU, line.ItemType, line.Quantity, line.UnitPrice, line.DiscountAmount, line.TaxAmount, line.LineTotal, line.UOMCode, lineMeta); lErr != nil {
						failed++
						errs = append(errs, fmt.Sprintf("invoice %s line %d error: %v", inv.InvoiceNumber, line.LineNumber, lErr))
					}
				}
			}
		}

	case contracts.DomainIncomingPayments:
		for _, pay := range payload.IncomingPayments {
			if pay.PaymentAmount <= 0 {
				continue
			}
			if pay.Metadata == nil {
				pay.Metadata = make(map[string]interface{})
			}
			if pay.CustomerCode != "" {
				pay.Metadata["sap_customer_code"] = pay.CustomerCode
			}
			if pay.InvoiceNumber != "" {
				pay.Metadata["sap_invoice_number"] = pay.InvoiceNumber
			}
			metaBytes, _ := json.Marshal(pay.Metadata)
			sapDocEntryStr := fmt.Sprintf("%d", pay.SAPInvoiceDocEntry)

			// A20: on-account payments (no invoice allocation) cannot live in
			// invoice_payments (invoice_id NOT NULL). Route them to
			// customer_payments so they are auditable and reconcilable with SAP
			// ORCT cash totals instead of being silently dropped. The customer
			// balance snapshot (business_partners.outstanding_balance from
			// OCRD.Balance) already reflects these historical payments, so no
			// balance mutation is performed during migration.
			if pay.IsOnAccount {
				oaQ := `
				INSERT INTO customer_payments (
					organization_id, payment_number, customer_id, customer_code, payment_date,
					payment_amount, payment_method, payment_reference, currency_code, reason, notes, metadata
				)
				SELECT $1, $2,
					(SELECT id FROM customers WHERE organization_id = $1 AND customer_code = $3 LIMIT 1),
					$3, $4, $5, $6, $7, $8, 'on_account', $9, $10
				ON CONFLICT(payment_number) DO UPDATE SET
					payment_amount    = excluded.payment_amount,
					payment_method    = excluded.payment_method,
					payment_reference = excluded.payment_reference,
					payment_date      = excluded.payment_date,
					metadata          = excluded.metadata;
				`
				tag, oaErr := execWithSavepointRows(ctx, tx, oaQ, payload.OrganizationID, pay.PaymentNumber, pay.CustomerCode, pay.PaymentDate, pay.PaymentAmount, pay.PaymentMethod, pay.PaymentReference, pay.CurrencyCode, pay.Notes, metaBytes)
				if oaErr != nil {
					failed++
					errs = append(errs, fmt.Sprintf("payment %s error: %v", pay.PaymentNumber, oaErr))
				} else if tag.RowsAffected() == 0 {
					// A29: replay no-op
					skipped++
				} else {
					staged++
				}
				continue
			}

			// 1. Insert into invoice_payments linking to invoices table
			q := `
			INSERT INTO invoice_payments (
				organization_id, invoice_id, payment_number, payment_date,
				payment_amount, payment_method, payment_reference, currency_code,
				notes, metadata
			)
			SELECT
				$1, inv.id, $2, $3, $4, $5, $6, $7, $8, $9
			FROM invoices inv
			WHERE inv.organization_id = $1 
			  AND (inv.invoice_number = $10 OR inv.metadata->>'sap_doc_entry' = $11)
			LIMIT 1
			ON CONFLICT(payment_number) DO UPDATE SET
				payment_amount    = excluded.payment_amount,
				payment_method    = excluded.payment_method,
				payment_reference = excluded.payment_reference,
				payment_date      = excluded.payment_date,
				metadata          = excluded.metadata;
			`
			tag, err := execWithSavepointRows(ctx, tx, q, payload.OrganizationID, pay.PaymentNumber, pay.PaymentDate, pay.PaymentAmount, pay.PaymentMethod, pay.PaymentReference, pay.CurrencyCode, pay.Notes, metaBytes, pay.InvoiceNumber, sapDocEntryStr)
			if err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("payment %s error: %v", pay.PaymentNumber, err))
				continue
			}
			if tag.RowsAffected() == 0 {
				// A20: the referenced invoice was not found (on-account payment
				// whose allocation row pointed past InvoiceStartDate, or invoice
				// batch not yet ingested). Record it instead of dropping it.
				umMeta := pay.Metadata
				if umMeta == nil {
					umMeta = make(map[string]interface{})
				}
				umMeta["unmatched_invoice_number"] = pay.InvoiceNumber
				umMeta["unmatched_invoice_doc_entry"] = sapDocEntryStr
				umBytes, _ := json.Marshal(umMeta)
				umQ := `
				INSERT INTO customer_payments (
					organization_id, payment_number, customer_id, customer_code, payment_date,
					payment_amount, payment_method, payment_reference, currency_code, reason, notes, metadata
				)
				SELECT $1, $2,
					(SELECT id FROM customers WHERE organization_id = $1 AND customer_code = $3 LIMIT 1),
					$3, $4, $5, $6, $7, $8, 'unmatched_invoice', $9, $10
				ON CONFLICT(payment_number) DO UPDATE SET
					payment_amount    = excluded.payment_amount,
					payment_method    = excluded.payment_method,
					payment_reference = excluded.payment_reference,
					payment_date      = excluded.payment_date,
					metadata          = excluded.metadata;
				`
				if umErr := execWithSavepoint(ctx, tx, umQ, payload.OrganizationID, pay.PaymentNumber, pay.CustomerCode, pay.PaymentDate, pay.PaymentAmount, pay.PaymentMethod, pay.PaymentReference, pay.CurrencyCode, pay.Notes, umBytes); umErr != nil {
					failed++
					errs = append(errs, fmt.Sprintf("payment %s error: invoice %q not found and customer_payment routing failed: %v", pay.PaymentNumber, pay.InvoiceNumber, umErr))
				} else {
					staged++
					errs = append(errs, fmt.Sprintf("payment %s: invoice %q not found (pre-cutoff or not yet ingested); recorded as customer payment", pay.PaymentNumber, pay.InvoiceNumber))
				}
				continue
			}
			staged++

			// 2. Automatically recalculate and settle parent invoice balance
			updateInvQ := `
			UPDATE invoices
			SET 
				paid_amount = (
					SELECT COALESCE(SUM(ip.payment_amount), 0)
					FROM invoice_payments ip
					WHERE ip.invoice_id = invoices.id
				),
				balance_due = GREATEST(0, total_amount - (
					SELECT COALESCE(SUM(ip.payment_amount), 0)
					FROM invoice_payments ip
					WHERE ip.invoice_id = invoices.id
				)),
				invoice_status = CASE 
					WHEN (
						SELECT COALESCE(SUM(ip.payment_amount), 0)
						FROM invoice_payments ip
						WHERE ip.invoice_id = invoices.id
					) >= total_amount THEN 'paid'::invoice_status
					WHEN (
						SELECT COALESCE(SUM(ip.payment_amount), 0)
						FROM invoice_payments ip
						WHERE ip.invoice_id = invoices.id
					) > 0 THEN 'partially_paid'::invoice_status
					ELSE invoice_status
				END
			WHERE organization_id = $1 
			  AND (invoice_number = $2 OR metadata->>'sap_doc_entry' = $3);
			`
			_ = execWithSavepoint(ctx, tx, updateInvQ, payload.OrganizationID, pay.InvoiceNumber, sapDocEntryStr)
		}

	case contracts.DomainPaymentTerms:
		// A23: OCTG payment terms master. Must be ingested before partners so
		// business_partners.payment_terms_id can resolve (PT-{GroupNum} codes).
		for _, pt := range payload.PaymentTerms {
			name := pt.Name
			if name == "" {
				name = pt.Code
			}
			q := `
			INSERT INTO payment_terms (organization_id, code, name, due_days, discount_days, discount_percentage, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, true)
			ON CONFLICT(code) DO UPDATE SET
				name = excluded.name,
				due_days = excluded.due_days,
				discount_days = excluded.discount_days,
				discount_percentage = excluded.discount_percentage;
			`
			if err := execWithSavepoint(ctx, tx, q, payload.OrganizationID, pt.Code, name, pt.DueDays, pt.DiscountDays, pt.DiscountPercentage); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("payment_term %s error: %v", pt.Code, err))
			} else {
				staged++
			}
		}

	case contracts.DomainOutgoingPayments:
		// A31: supplier payments (OVPM/VPM2). Stores the payment audit trail in
		// vendor_payments and recomputes purchase_orders.metadata->>'amount_paid'
		// from stored payments so vw_accounts_payable and
		// vw_supplier_aging_report show realistic outstanding balances.
		for _, pay := range payload.OutgoingPayments {
			if pay.PaymentAmount <= 0 {
				continue
			}
			metaBytes, _ := json.Marshal(pay.Metadata)
			q := `
			INSERT INTO vendor_payments (
				organization_id, payment_number, partner_id, supplier_code, purchase_order_id,
				po_doc_entry, payment_date, payment_amount, payment_method, payment_reference,
				currency_code, notes, metadata
			)
			SELECT $1, $2,
				(SELECT id FROM business_partners WHERE organization_id = $1 AND code = $3 LIMIT 1),
				$3,
				CASE WHEN $4::bigint > 0 THEN (SELECT id FROM purchase_orders WHERE organization_id = $1 AND metadata->>'sap_doc_entry' = $4::text LIMIT 1) END,
				NULLIF($4::bigint, 0),
				$5, $6, $7, $8, $9, $10, $11
			ON CONFLICT(payment_number) DO UPDATE SET
				purchase_order_id = COALESCE(excluded.purchase_order_id, vendor_payments.purchase_order_id),
				payment_date      = excluded.payment_date,
				payment_amount    = excluded.payment_amount,
				payment_method    = excluded.payment_method,
				payment_reference = excluded.payment_reference,
				metadata          = excluded.metadata;
			`
			poDocEntryStr := fmt.Sprintf("%d", pay.PODocEntry)
			tag, err := execWithSavepointRows(ctx, tx, q, payload.OrganizationID, pay.PaymentNumber, pay.SupplierCode, poDocEntryStr, pay.PaymentDate, pay.PaymentAmount, pay.PaymentMethod, pay.PaymentReference, pay.CurrencyCode, pay.Notes, metaBytes)
			if err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("outgoing payment %s error: %v", pay.PaymentNumber, err))
			} else if tag.RowsAffected() == 0 {
				// A29: replay no-op
				skipped++
			} else {
				staged++

				// Recompute amount_paid on the linked PO from the vendor_payments ledger
				if pay.PODocEntry > 0 {
					_ = execWithSavepoint(ctx, tx, `
					UPDATE purchase_orders po
					SET metadata = jsonb_set(
						COALESCE(po.metadata, '{}'::jsonb),
						'{amount_paid}',
						to_jsonb(COALESCE((
							SELECT SUM(vp.payment_amount)
							FROM vendor_payments vp
							WHERE vp.purchase_order_id = po.id
						), 0)::numeric),
						true)
					WHERE po.organization_id = $1
					  AND po.metadata->>'sap_doc_entry' = $2;
					`, payload.OrganizationID, poDocEntryStr)
				}
			}
		}

	case contracts.DomainSalesReturns:
		// A21: A/R Credit Memos (ORIN) and A/R Returns (ORDN) → sales_returns +
		// sales_return_lines. Status 'completed' so vw_realtime_pnl.daily_returns
		// and vw_cashier_shift_reconciliation.refunds_summary include them;
		// canceled documents import as 'cancelled'.
		for _, sr := range payload.SalesReturns {
			metaBytes, _ := json.Marshal(sr.Metadata)
			q := `
			INSERT INTO sales_returns (
				return_number, store_id, customer_id, return_date, return_reason,
				status, subtotal, tax_amount, total_refund_amount, refund_method, refund_reference, notes, metadata
			)
			SELECT $1,
				COALESCE((SELECT id FROM stores WHERE organization_id = $2 AND code = $3 LIMIT 1),
				         (SELECT id FROM stores WHERE organization_id = $2 LIMIT 1)),
				(SELECT id FROM customers WHERE organization_id = $2 AND customer_code = $4 LIMIT 1),
				$5, $6, $7, $8, $9, $10, 'other', $11, $12, $13
			ON CONFLICT(return_number) DO UPDATE SET
				store_id            = COALESCE(excluded.store_id, sales_returns.store_id),
				customer_id         = COALESCE(excluded.customer_id, sales_returns.customer_id),
				return_date         = excluded.return_date,
				return_reason       = excluded.return_reason,
				status              = excluded.status,
				subtotal            = excluded.subtotal,
				tax_amount          = excluded.tax_amount,
				total_refund_amount = excluded.total_refund_amount,
				notes               = excluded.notes,
				metadata            = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, q, sr.ReturnNumber, payload.OrganizationID, sr.StoreCode, sr.CustomerCode, sr.ReturnDate, sr.Reason, sr.Status, sr.Subtotal, sr.TaxAmount, sr.TotalRefundAmount, sr.ReturnNumber, sr.Reason, metaBytes); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("sales return %s error: %v", sr.ReturnNumber, err))
			} else {
				staged++
				// Lines are cleaned and re-inserted idempotently (no unique key on sales_return_lines)
				delQ := `
				DELETE FROM sales_return_lines
				WHERE return_id = (SELECT id FROM sales_returns WHERE return_number = $1);
				`
				_ = execWithSavepoint(ctx, tx, delQ, sr.ReturnNumber)

				for _, line := range sr.Lines {
					lineMeta, _ := json.Marshal(line.Metadata)
					lq := `
					INSERT INTO sales_return_lines (
						return_id, product_id, quantity, unit_price, refund_amount, line_number, metadata
					)
					SELECT
						sr.id,
						(SELECT id FROM products WHERE organization_id = $2 AND sku = $3 LIMIT 1),
						$4, $5, $6, $7, $8
					FROM sales_returns sr
					WHERE sr.return_number = $1
					  AND (SELECT id FROM products WHERE organization_id = $2 AND sku = $3 LIMIT 1) IS NOT NULL;
					`
					if _, lErr := execWithSavepointRows(ctx, tx, lq, sr.ReturnNumber, payload.OrganizationID, line.ProductSKU, line.Quantity, line.UnitPrice, line.RefundAmount, line.LineNumber, lineMeta); lErr != nil {
						failed++
						errs = append(errs, fmt.Sprintf("sales return %s line %d error: %v", sr.ReturnNumber, line.LineNumber, lErr))
					}
				}
			}
		}

	case contracts.DomainTransfers:
		// A22: inter-store inventory transfers (OWTR/WTR1). Must be ingested
		// before stock_movements so TransType 67 movements can resolve
		// reference_id to the transfer parent document.
		for _, tr := range payload.Transfers {
			metaBytes, _ := json.Marshal(tr.Metadata)

			var fromStoreID *int64
			if tr.FromStoreCode != "" {
				r := tx.QueryRow(ctx, `SELECT id FROM stores WHERE organization_id = $1 AND code = $2 LIMIT 1`, payload.OrganizationID, tr.FromStoreCode)
				_ = r.Scan(&fromStoreID)
			}
			if fromStoreID == nil {
				failed++
				errs = append(errs, fmt.Sprintf("transfer %s error: source warehouse %q not found", tr.TransferNumber, tr.FromStoreCode))
				continue
			}

			var toStoreID *int64
			if tr.ToStoreCode != "" {
				r := tx.QueryRow(ctx, `SELECT id FROM stores WHERE organization_id = $1 AND code = $2 LIMIT 1`, payload.OrganizationID, tr.ToStoreCode)
				_ = r.Scan(&toStoreID)
			}
			if toStoreID == nil {
				failed++
				errs = append(errs, fmt.Sprintf("transfer %s error: destination warehouse %q not found", tr.TransferNumber, tr.ToStoreCode))
				continue
			}

			q := `
			INSERT INTO transfer_requests (
				organization_id, transfer_number, from_store_id, to_store_id, status,
				request_date, shipped_at, received_at, notes, metadata
			)
			SELECT $1, $2, $3, $4, $5, $6, $6, $6, $7, $8
			ON CONFLICT(transfer_number) DO UPDATE SET
				from_store_id = excluded.from_store_id,
				to_store_id   = excluded.to_store_id,
				status        = excluded.status,
				notes         = excluded.notes,
				metadata      = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, q, payload.OrganizationID, tr.TransferNumber, *fromStoreID, *toStoreID, tr.Status, tr.TransferDate, tr.Notes, metaBytes); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("transfer %s error: %v", tr.TransferNumber, err))
			} else {
				staged++
				// Items are cleaned and re-inserted idempotently (no unique key on transfer_request_items)
				delQ := `
				DELETE FROM transfer_request_items
				WHERE transfer_request_id = (SELECT id FROM transfer_requests WHERE transfer_number = $1);
				`
				_ = execWithSavepoint(ctx, tx, delQ, tr.TransferNumber)

				for _, line := range tr.Lines {
					lineMeta, _ := json.Marshal(line.Metadata)
					lq := `
					INSERT INTO transfer_request_items (
						transfer_request_id, product_id, requested_quantity, approved_quantity,
						shipped_quantity, received_quantity, uom_id, notes
					)
					SELECT
						t.id,
						(SELECT id FROM products WHERE organization_id = $2 AND sku = $3 LIMIT 1),
						$4, $4, $4, $4,
						(SELECT id FROM units_of_measure WHERE code = $5 LIMIT 1),
						$6
					FROM transfer_requests t
					WHERE t.transfer_number = $1
					  AND (SELECT id FROM products WHERE organization_id = $2 AND sku = $3 LIMIT 1) IS NOT NULL;
					`
					if _, lErr := execWithSavepointRows(ctx, tx, lq, tr.TransferNumber, payload.OrganizationID, line.ProductSKU, line.Quantity, line.UOMCode, lineMeta); lErr != nil {
						failed++
						errs = append(errs, fmt.Sprintf("transfer %s line %d error: %v", tr.TransferNumber, line.LineNumber, lErr))
					}
				}
			}
		}

	default:
		// Accept unknown or custom domain payloads gracefully
		staged = payload.RecordCount()
	}

	// Update batch record — advisory only
	status := "merged"
	if failed > 0 && staged == 0 {
		status = "failed"
	}
	if err := execWithSavepoint(ctx, tx, `UPDATE staging.sap_migration_batches SET status = $1 WHERE batch_id = $2;`, status, payload.BatchID); err != nil {
		errs = append(errs, fmt.Sprintf("staging batch status update (advisory): %v", err))
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit batch transaction: %w", err)
	}

	return &contracts.MigrationBatchResponse{
		Success:        failed == 0 || staged > 0,
		BatchID:        payload.BatchID,
		Domain:         payload.Domain,
		RecordsStaged:  staged,
		RecordsMerged:  staged,
		RecordsFailed:  failed,
		RecordsSkipped: skipped,
		Errors:         errs,
	}, nil
}
