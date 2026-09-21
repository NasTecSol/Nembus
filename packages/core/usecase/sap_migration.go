package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NasTecSol/nembus-sap/contracts"
)

type SAPMigrationUseCase struct {
	pool *pgxpool.Pool
}

func NewSAPMigrationUseCase(pool *pgxpool.Pool) *SAPMigrationUseCase {
	return &SAPMigrationUseCase{pool: pool}
}

func execWithSavepoint(ctx context.Context, tx pgx.Tx, query string, args ...interface{}) error {
	spName := "sp_row"
	if _, err := tx.Exec(ctx, "SAVEPOINT "+spName); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, query, args...); err != nil {
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

	tx, err := uc.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin postgres transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Record Batch in Staging Audit Table
	batchQuery := `
	INSERT INTO staging.sap_migration_batches (batch_id, run_id, organization_id, domain, record_count, status)
	VALUES ($1, $2, $3, $4, $5, 'staged')
	ON CONFLICT(batch_id) DO UPDATE SET record_count = excluded.record_count;
	`
	_, _ = tx.Exec(ctx, batchQuery, payload.BatchID, payload.RunID, payload.OrganizationID, string(payload.Domain), payload.RecordCount())

	staged := 0
	failed := 0
	var errs []string

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
			q := `
			INSERT INTO cashiers (user_id, store_id, cashier_code, drawer_limit, discount_limit, is_active, metadata)
			SELECT u.id, s.id, $3, $4, $5, $6, $7
			FROM users u
			CROSS JOIN (SELECT id FROM stores WHERE organization_id = $1 ORDER BY id LIMIT 1) s
			WHERE u.username = $2
			ON CONFLICT(store_id, cashier_code) DO UPDATE SET
				is_active = excluded.is_active,
				metadata = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, q, payload.OrganizationID, c.Username, c.CashierCode, c.DrawerLimit, c.DiscountLimit, c.IsActive, metaBytes); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("cashier %s error: %v", c.CashierCode, err))
			} else {
				staged++
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
			err := tx.QueryRow(ctx, tq, payload.OrganizationID, grp.BaseUOMCode, grp.Name, grp.Code, grp.IsActive).Scan(&templateID)
			if err != nil {
				_ = tx.QueryRow(ctx, `SELECT id FROM uom_packaging_templates WHERE code = $1`, grp.Code).Scan(&templateID)
			}

			if templateID > 0 {
				for _, lvl := range grp.Levels {
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
				errs = append(errs, fmt.Sprintf("uom_group %s error: failed to resolve template id", grp.Code))
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
			if err := execWithSavepoint(ctx, tx, bq, payload.OrganizationID, b.Barcode, b.BarcodeType, b.IsPrimary, b.ProductSKU); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("barcode %s error: %v", b.Barcode, err))
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
			if err := execWithSavepoint(ctx, tx, q, payload.OrganizationID, item.ProductSKU, item.Price, metaBytes, item.PriceListCode, item.UOMCode); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("price_item %s@%s error: %v", item.ProductSKU, item.PriceListCode, err))
			} else {
				staged++
			}
		}

	case contracts.DomainInventory:
		for _, inv := range payload.Inventory {
			metaBytes, _ := json.Marshal(inv.Metadata)
			q := `
			INSERT INTO inventory_stock (product_id, store_id, quantity_on_hand, quantity_allocated, quantity_available, quantity_on_order, reorder_level, max_stock_level, metadata)
			SELECT p.id, s.id, $3, $4, $5, $6, $7, $8, $9
			FROM products p
			CROSS JOIN stores s
			WHERE p.organization_id = $1 AND p.sku = $2 AND s.organization_id = $1 AND s.code = $10
			LIMIT 1;
			`
			if err := execWithSavepoint(ctx, tx, q, payload.OrganizationID, inv.ProductSKU, inv.QuantityOnHand, inv.QuantityAllocated, inv.QuantityAvailable, inv.QuantityOnOrder, inv.ReorderLevel, inv.MaxStockLevel, metaBytes, inv.StoreCode); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("stock %s@%s error: %v", inv.ProductSKU, inv.StoreCode, err))
			} else {
				staged++
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

			role := "supplier"
			if pt.PartnerType == "customer" {
				role = "customer"
			}

			bpQ := `
			INSERT INTO business_partners (organization_id, code, name, partner_role, tax_id, is_active, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT(code) DO UPDATE SET
				name = excluded.name,
				partner_role = excluded.partner_role,
				tax_id = excluded.tax_id,
				is_active = excluded.is_active,
				metadata = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, bpQ, payload.OrganizationID, pt.Code, pt.Name, role, pt.TaxID, pt.IsActive, metaBytes); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("business_partner %s error: %v", pt.Code, err))
			} else {
				staged++
			}

			if pt.PartnerType == "customer" {
				custQ := `
				INSERT INTO customers (organization_id, customer_code, name, is_active, business_partner_id, metadata)
				VALUES ($1, $2, $3, $4, (SELECT id FROM business_partners WHERE code = $2 LIMIT 1), $5)
				ON CONFLICT(organization_id, customer_code) DO UPDATE SET
					name = excluded.name,
					is_active = excluded.is_active,
					business_partner_id = excluded.business_partner_id,
					metadata = excluded.metadata;
				`
				_ = execWithSavepoint(ctx, tx, custQ, payload.OrganizationID, pt.Code, pt.Name, pt.IsActive, metaBytes)
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
			q := `
			INSERT INTO purchase_orders (
				organization_id, po_number, supplier_id, store_id, po_date, expected_delivery_date,
				status, subtotal, discount_amount, tax_amount, total_amount, metadata
			)
			SELECT
				$1, $2,
				COALESCE((SELECT id FROM suppliers WHERE organization_id = $1 AND code = $3 LIMIT 1),
				         (SELECT id FROM suppliers WHERE organization_id = $1 LIMIT 1)),
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
			if err := execWithSavepoint(ctx, tx, q, payload.OrganizationID, po.PONumber, po.SupplierCode, po.StoreCode, po.PODate, po.ExpectedDeliveryDate, po.Status, po.Subtotal, po.DiscountAmount, po.TaxAmount, po.TotalAmount, metaBytes); err != nil {
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
					_ = execWithSavepoint(ctx, tx, lq, po.PONumber, payload.OrganizationID, line.ProductSKU, line.Quantity, line.UOMCode, line.UnitPrice, line.DiscountAmount, line.TaxAmount, line.Subtotal, line.LineTotal, line.ReceivedQuantity, line.LineNumber, lineMeta)
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
			q := `
			INSERT INTO goods_receipt_notes (
				organization_id, grn_number, purchase_order_id, supplier_id, store_id,
				receipt_date, delivery_note_number, status, notes, metadata
			)
			SELECT
				$1, $2,
				(SELECT id FROM purchase_orders WHERE organization_id = $1 AND po_number = $3 LIMIT 1),
				COALESCE((SELECT id FROM suppliers WHERE organization_id = $1 AND code = $4 LIMIT 1),
				         (SELECT id FROM suppliers WHERE organization_id = $1 LIMIT 1)),
				COALESCE((SELECT id FROM stores WHERE organization_id = $1 AND code = $5 LIMIT 1),
				         (SELECT id FROM stores WHERE organization_id = $1 LIMIT 1)),
				$6, $7, $8, $9, $10
			ON CONFLICT(grn_number) DO UPDATE SET
				status = excluded.status,
				notes = excluded.notes,
				metadata = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, q, payload.OrganizationID, grn.GRNNumber, grn.PONumber, grn.SupplierCode, grn.StoreCode, grn.ReceiptDate, grn.DeliveryNoteNumber, grn.Status, grn.Notes, metaBytes); err != nil {
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
						(SELECT pol.id FROM purchase_order_lines pol JOIN purchase_orders po ON pol.purchase_order_id = po.id WHERE po.po_number = $3 AND pol.line_number = $4 LIMIT 1),
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
			q := `
			INSERT INTO stock_movements (
				movement_type, reference_type, product_id,
				from_store_id, to_store_id, quantity,
				movement_date, status, cost_per_unit, total_value, metadata
			)
			SELECT
				$1, $2,
				(SELECT id FROM products WHERE organization_id = $3 AND sku = $4 LIMIT 1),
				(SELECT id FROM stores WHERE organization_id = $3 AND code = $5 LIMIT 1),
				(SELECT id FROM stores WHERE organization_id = $3 AND code = $6 LIMIT 1),
				$7, $8, 'completed', $9, $10, $11
			WHERE (SELECT id FROM products WHERE organization_id = $3 AND sku = $4 LIMIT 1) IS NOT NULL;
			`
			if err := execWithSavepoint(ctx, tx, q, sm.MovementType, sm.ReferenceType, payload.OrganizationID, sm.ProductSKU, sm.FromStoreCode, sm.ToStoreCode, sm.Quantity, sm.MovementDate, sm.CostPerUnit, sm.TotalValue, metaBytes); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("stock movement item %s error: %v", sm.ProductSKU, err))
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
				order_number, organization_id, customer_name, order_status, payment_status, fulfillment_status,
				order_date, expected_delivery_date, subtotal, discount_amount, tax_amount, total_amount, internal_notes, shipping_address, billing_address, metadata
			)
			VALUES ($1, $2, $3, $4::order_status_v2, $5::payment_status, $6::fulfillment_status, $7, $8, $9, $10, $11, $12, $13, '{}'::jsonb, '{}'::jsonb, $14)
			ON CONFLICT(order_number) DO UPDATE SET
				total_amount = excluded.total_amount,
				metadata = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, q, so.OrderNumber, payload.OrganizationID, so.CustomerName, so.OrderStatus, so.PaymentStatus, so.FulfillmentStatus, so.OrderDate, so.ExpectedDate, so.Subtotal, so.DiscountAmount, so.TaxAmount, so.TotalAmount, so.Notes, metaBytes); err != nil {
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
					_ = execWithSavepoint(ctx, tx, lq, so.OrderNumber, payload.OrganizationID, line.LineNumber, line.ProductName, line.ProductSKU, line.Quantity, line.UnitPrice, line.TaxAmount, line.LineTotal)
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
				invoice_number, organization_id, customer_id, customer_name, invoice_type, invoice_status,
				invoice_date, due_date, subtotal, discount_amount, tax_amount, total_amount, paid_amount, balance_due, billing_address, metadata
			)
			SELECT
				$1, $2,
				(SELECT id FROM customers WHERE organization_id = $2 AND customer_code = $3 LIMIT 1),
				$4, $5::invoice_type, $6::invoice_status, $7, $8, $9, $10, $11, $12, $13, $14, '{}'::jsonb, $15
			ON CONFLICT(invoice_number) DO UPDATE SET
				customer_id = COALESCE(excluded.customer_id, invoices.customer_id),
				paid_amount = excluded.paid_amount,
				balance_due = excluded.balance_due,
				metadata    = excluded.metadata;
			`
			if err := execWithSavepoint(ctx, tx, q, inv.InvoiceNumber, payload.OrganizationID, inv.CustomerCode, inv.CustomerName, inv.InvoiceType, inv.InvoiceStatus, inv.InvoiceDate, inv.DueDate, inv.Subtotal, inv.DiscountAmount, inv.TaxAmount, inv.TotalAmount, inv.PaidAmount, inv.BalanceDue, metaBytes); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("invoice %s error: %v", inv.InvoiceNumber, err))
			} else {
				staged++
				for _, line := range inv.Lines {
					lineMeta, _ := json.Marshal(line.Metadata)
					lq := `
					INSERT INTO invoice_lines (
						invoice_id, organization_id, line_number, description, item_type, product_id, product_sku,
						quantity, unit_price, tax_amount, line_total, metadata
					)
					SELECT 
						inv.id, $2, $3,
						COALESCE(NULLIF($4, ''), $5),
						'product',
						(SELECT id FROM products WHERE organization_id = $2 AND sku = $5 LIMIT 1),
						$5, $6, $7, $8, $9, $10
					FROM invoices inv
					WHERE inv.invoice_number = $1 AND inv.organization_id = $2
					ON CONFLICT(invoice_id, line_number) DO UPDATE SET
						quantity = excluded.quantity,
						unit_price = excluded.unit_price,
						tax_amount = excluded.tax_amount,
						line_total = excluded.line_total,
						metadata = excluded.metadata;
					`
					_ = execWithSavepoint(ctx, tx, lq, inv.InvoiceNumber, payload.OrganizationID, line.LineNumber, line.ProductName, line.ProductSKU, line.Quantity, line.UnitPrice, line.TaxAmount, line.LineTotal, lineMeta)
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
			if err := execWithSavepoint(ctx, tx, q, payload.OrganizationID, pay.PaymentNumber, pay.PaymentDate, pay.PaymentAmount, pay.PaymentMethod, pay.PaymentReference, pay.CurrencyCode, pay.Notes, metaBytes, pay.InvoiceNumber, sapDocEntryStr); err != nil {
				failed++
				errs = append(errs, fmt.Sprintf("payment %s error: %v", pay.PaymentNumber, err))
			} else {
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
		}

	default:
		// Accept unknown or custom domain payloads gracefully
		staged = payload.RecordCount()
	}

	// Update batch record
	status := "merged"
	if failed > 0 && staged == 0 {
		status = "failed"
	}
	_, _ = tx.Exec(ctx, `UPDATE staging.sap_migration_batches SET status = $1 WHERE batch_id = $2;`, status, payload.BatchID)

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit batch transaction: %w", err)
	}

	return &contracts.MigrationBatchResponse{
		Success:       failed == 0 || staged > 0,
		BatchID:       payload.BatchID,
		Domain:        payload.Domain,
		RecordsStaged: staged,
		RecordsMerged: staged,
		RecordsFailed: failed,
		Errors:        errs,
	}, nil
}
