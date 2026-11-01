package usecase

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NasTecSol/nembus-core/middleware/manager"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-sap/contracts"
)

type SapAgentUseCase struct {
	tenantManager *manager.Manager
	masterRepo    *repository.Queries
	migrationUC   *SAPMigrationUseCase
}

func NewSapAgentUseCase(tm *manager.Manager, masterRepo *repository.Queries, migrationUC *SAPMigrationUseCase) *SapAgentUseCase {
	return &SapAgentUseCase{
		tenantManager: tm,
		masterRepo:    masterRepo,
		migrationUC:   migrationUC,
	}
}

type TenantResolveResult struct {
	TenantSlug     string
	TenantName     string
	OrganizationID int32
	DefaultStoreID int32
}

func (uc *SapAgentUseCase) ResolveTenant(ctx context.Context, slug string) (*TenantResolveResult, error) {
	if slug == "" {
		return nil, fmt.Errorf("tenant_slug is required")
	}
	tenant, err := uc.masterRepo.GetTenantBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	pool, err := uc.tenantManager.GetPool(ctx, slug)
	if err != nil {
		return nil, err
	}

	var orgID int32
	err = pool.QueryRow(ctx, `SELECT id FROM organizations WHERE is_active = true ORDER BY id ASC LIMIT 1`).Scan(&orgID)
	if err != nil {
		orgID = 1
	}

	var storeID int32
	_ = pool.QueryRow(ctx, `SELECT id FROM stores WHERE organization_id = $1 AND is_active = true ORDER BY id ASC LIMIT 1`, orgID).Scan(&storeID)

	if err := uc.ensureDefaultApprovalPolicies(ctx, slug, orgID); err != nil {
		return nil, fmt.Errorf("ensure approval policies: %w", err)
	}

	return &TenantResolveResult{
		TenantSlug:     tenant.Slug,
		TenantName:     tenant.TenantName,
		OrganizationID: orgID,
		DefaultStoreID: storeID,
	}, nil
}

func (uc *SapAgentUseCase) ensureDefaultApprovalPolicies(ctx context.Context, tenantSlug string, orgID int32) error {
	pool, err := uc.tenantManager.GetPool(ctx, tenantSlug)
	if err != nil {
		return err
	}
	repo := repository.New(pool)
	approvalsUC := NewApprovalsUseCase()
	approvalsUC.SetRepository(repo)
	return approvalsUC.EnsureDefaultPolicies(ctx, orgID)
}

func (uc *SapAgentUseCase) IngestMigrationChunk(ctx context.Context, orgID int, gzipPayload []byte) (*contracts.MigrationBatchResponse, error) {
	reader, err := gzip.NewReader(bytes.NewReader(gzipPayload))
	if err != nil {
		return nil, fmt.Errorf("invalid gzip payload: %w", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read gzip payload: %w", err)
	}

	var payload contracts.MigrationBatchPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parse batch json: %w", err)
	}

	return uc.migrationUC.IngestBatch(ctx, orgID, &payload)
}

// SapOutboxBundle groups header + related rows for upstream SAP posting.
type SapOutboxBundle struct {
	EntityType string
	EntityID   int64
	HeaderJSON []byte
	LinesJSON  []byte
	PaymentsJSON []byte
}

func (uc *SapAgentUseCase) StreamSapOutbox(ctx context.Context, tenantSlug string, limit int) ([]SapOutboxBundle, error) {
	if limit <= 0 {
		limit = 50
	}
	pool, err := uc.tenantManager.GetPool(ctx, tenantSlug)
	if err != nil {
		return nil, err
	}

	var bundles []SapOutboxBundle

	txQuery := `
		SELECT id, store_id, cashier_id, cashier_session_id, customer_id,
		       transaction_number, transaction_date, subtotal, discount_amount,
		       tax_amount, total_amount, amount_paid, status, metadata
		FROM pos_transactions
		WHERE status = 'completed'
		  AND (metadata->>'sap_doc_num' IS NULL OR metadata->>'sap_doc_num' = '')
		  AND (
		    metadata->>'sap_sync_state' IS NULL
		    OR metadata->>'sap_sync_state' NOT IN ('held_stock', 'adjustment_created_pending_force')
		    OR COALESCE((metadata->>'sap_force_post')::boolean, false) = true
		  )
		ORDER BY id ASC
		LIMIT $1
	`
	rows, err := pool.Query(ctx, txQuery, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var header map[string]any
		var id int64
		var storeID, cashierID, cashierSessionID int32
		var customerID sql.NullInt32
		var txNum, status string
		var txDate time.Time
		var subtotal, discount, tax, total, paid float64
		var meta []byte

		if err := rows.Scan(&id, &storeID, &cashierID, &cashierSessionID, &customerID,
			&txNum, &txDate, &subtotal, &discount, &tax, &total, &paid, &status, &meta); err != nil {
			return nil, err
		}

		header = map[string]any{
			"id": id, "store_id": storeID, "cashier_id": cashierID,
			"cashier_session_id": cashierSessionID, "transaction_number": txNum,
			"transaction_date": txDate, "subtotal": subtotal, "discount_amount": discount,
			"tax_amount": tax, "total_amount": total, "amount_paid": paid, "status": status,
			"metadata": json.RawMessage(meta),
		}
		if customerID.Valid {
			header["customer_id"] = customerID.Int32
		}

		lines, _ := fetchTxLines(ctx, pool, id)
		payments, _ := fetchTxPayments(ctx, pool, id)
		headerBytes, _ := json.Marshal(header)
		linesBytes, _ := json.Marshal(lines)
		payBytes, _ := json.Marshal(payments)

		bundles = append(bundles, SapOutboxBundle{
			EntityType:   "pos_transaction",
			EntityID:     id,
			HeaderJSON:   headerBytes,
			LinesJSON:    linesBytes,
			PaymentsJSON: payBytes,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	retQuery := `
		SELECT id, return_number, store_id, cashier_id, customer_id,
		       original_transaction_id, return_date, return_reason, status,
		       subtotal, tax_amount, total_refund_amount, metadata
		FROM sales_returns
		WHERE status = 'completed'
		  AND (metadata->>'sap_doc_num' IS NULL OR metadata->>'sap_doc_num' = '')
		ORDER BY id ASC
		LIMIT $1
	`
	retRows, err := pool.Query(ctx, retQuery, limit)
	if err != nil {
		return bundles, nil
	}
	defer retRows.Close()

	for retRows.Next() {
		var id int64
		var storeID int32
		var cashierID, customerID, origTxID sql.NullInt32
		var retNum, status string
		var retDate time.Time
		var reason sql.NullString
		var subtotal, tax, refund float64
		var meta []byte

		if err := retRows.Scan(&id, &retNum, &storeID, &cashierID, &customerID,
			&origTxID, &retDate, &reason, &status, &subtotal, &tax, &refund, &meta); err != nil {
			return nil, err
		}

		header := map[string]any{
			"id": id, "return_number": retNum, "store_id": storeID,
			"return_date": retDate, "status": status,
			"subtotal": subtotal, "tax_amount": tax, "total_refund_amount": refund,
			"metadata": json.RawMessage(meta),
		}
		if cashierID.Valid {
			header["cashier_id"] = cashierID.Int32
		}
		if customerID.Valid {
			header["customer_id"] = customerID.Int32
		}
		if origTxID.Valid {
			header["original_transaction_id"] = origTxID.Int32
		}
		if reason.Valid {
			header["return_reason"] = reason.String
		}

		lines, _ := fetchReturnLines(ctx, pool, id)
		headerBytes, _ := json.Marshal(header)
		linesBytes, _ := json.Marshal(lines)

		bundles = append(bundles, SapOutboxBundle{
			EntityType: "sales_return",
			EntityID:   id,
			HeaderJSON: headerBytes,
			LinesJSON:  linesBytes,
		})
	}

	return bundles, retRows.Err()
}

func fetchTxLines(ctx context.Context, pool *pgxpool.Pool, txID int64) ([]map[string]any, error) {
	query := `
		SELECT l.id, l.transaction_id, l.product_id, p.sku, p.name,
		       l.quantity, l.uom_id, l.unit_price, l.discount_amount,
		       l.tax_amount, l.subtotal, l.line_total, COALESCE(l.line_number, 0)
		FROM pos_transaction_lines l
		JOIN products p ON p.id = l.product_id
		WHERE l.transaction_id = $1
		ORDER BY l.id ASC
	`
	rows, err := pool.Query(ctx, query, txID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]any
	for rows.Next() {
		var id, transactionID int64
		var productID int32
		var sku, name string
		var qty, unitPrice, disc, tax, sub, lineTotal float64
		var uomID sql.NullInt32
		var lineNum int

		if err := rows.Scan(&id, &transactionID, &productID, &sku, &name,
			&qty, &uomID, &unitPrice, &disc, &tax, &sub, &lineTotal, &lineNum); err != nil {
			return nil, err
		}
		line := map[string]any{
			"id": id, "transaction_id": transactionID, "product_id": productID,
			"product_sku": sku, "product_name": name, "quantity": qty,
			"unit_price": unitPrice, "discount_amount": disc, "tax_amount": tax,
			"subtotal": sub, "line_total": lineTotal, "line_number": lineNum,
		}
		if uomID.Valid {
			line["uom_id"] = uomID.Int32
		}
		out = append(out, line)
	}
	return out, rows.Err()
}

func fetchTxPayments(ctx context.Context, pool *pgxpool.Pool, txID int64) ([]map[string]any, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, transaction_id, payment_method, amount, reference_number, payment_date
		FROM pos_payments WHERE transaction_id = $1`, txID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]any
	for rows.Next() {
		var id, transactionID int64
		var method string
		var amount float64
		var ref sql.NullString
		var payDate time.Time
		if err := rows.Scan(&id, &transactionID, &method, &amount, &ref, &payDate); err != nil {
			return nil, err
		}
		p := map[string]any{
			"id": id, "transaction_id": transactionID,
			"payment_method": method, "amount": amount, "payment_date": payDate,
		}
		if ref.Valid {
			p["reference"] = ref.String
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func fetchReturnLines(ctx context.Context, pool *pgxpool.Pool, returnID int64) ([]map[string]any, error) {
	rows, err := pool.Query(ctx, `
		SELECT l.id, l.return_id, l.product_id, p.sku, p.name,
		       l.original_line_id, l.quantity, l.unit_price, l.refund_amount,
		       l.return_to_stock, COALESCE(l.line_number, 0)
		FROM sales_return_lines l
		JOIN products p ON p.id = l.product_id
		WHERE l.return_id = $1
		ORDER BY l.id ASC`, returnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]any
	for rows.Next() {
		var id, retID int64
		var productID int32
		var sku, name string
		var origLine sql.NullInt32
		var qty, unitPrice, refund float64
		var returnToStock bool
		var lineNum int

		if err := rows.Scan(&id, &retID, &productID, &sku, &name,
			&origLine, &qty, &unitPrice, &refund, &returnToStock, &lineNum); err != nil {
			return nil, err
		}
		line := map[string]any{
			"id": id, "return_id": retID, "product_id": productID,
			"product_sku": sku, "product_name": name, "quantity": qty,
			"unit_price": unitPrice, "refund_amount": refund,
			"return_to_stock": returnToStock, "line_number": lineNum,
		}
		if origLine.Valid {
			line["original_line_id"] = origLine.Int32
		}
		out = append(out, line)
	}
	return out, rows.Err()
}

type SapOutboxAckItem struct {
	EntityType  string
	EntityID    int64
	SapDocEntry int
	SapDocNum   int
}

func (uc *SapAgentUseCase) AckSapOutbox(ctx context.Context, tenantSlug string, items []SapOutboxAckItem) (int, error) {
	pool, err := uc.tenantManager.GetPool(ctx, tenantSlug)
	if err != nil {
		return 0, err
	}

	ack := 0
	for _, item := range items {
		metaPatch, _ := json.Marshal(map[string]any{
			"sap_doc_entry":  item.SapDocEntry,
			"sap_doc_num":    item.SapDocNum,
			"sap_posted_at":  time.Now().Format(time.RFC3339),
			"sap_sync_state": "synced",
			"sap_force_post": false,
		})

		var q string
		switch item.EntityType {
		case "sales_return":
			q = `UPDATE sales_returns SET metadata = COALESCE(metadata, '{}'::jsonb) || $2::jsonb WHERE id = $1`
		default:
			q = `UPDATE pos_transactions SET metadata = COALESCE(metadata, '{}'::jsonb) || $2::jsonb WHERE id = $1`
		}
		if _, err := pool.Exec(ctx, q, item.EntityID, metaPatch); err != nil {
			return ack, err
		}
		ack++
	}
	return ack, nil
}

type HoldSapOutboxInput struct {
	EntityType        string
	EntityID          int64
	TransactionNumber string
	StoreID           int32
	WarehouseCode     string
	TotalAmount       float64
	Summary           string
	ShortageItemsJSON []byte
	OrganizationID    int32
}

func (uc *SapAgentUseCase) HoldSapOutbox(ctx context.Context, tenantSlug string, in HoldSapOutboxInput) error {
	pool, err := uc.tenantManager.GetPool(ctx, tenantSlug)
	if err != nil {
		return err
	}

	metaPatch, _ := json.Marshal(map[string]any{
		"sap_sync_state": "held_stock",
		"held_reason":    in.Summary,
		"held_at":        time.Now().Format(time.RFC3339),
	})
	_, err = pool.Exec(ctx, `
		UPDATE pos_transactions
		SET metadata = COALESCE(metadata, '{}'::jsonb) || $2::jsonb, updated_at = NOW()
		WHERE id = $1
	`, in.EntityID, metaPatch)
	if err != nil {
		return err
	}

	var shortage any
	_ = json.Unmarshal(in.ShortageItemsJSON, &shortage)
	payload, _ := json.Marshal(map[string]any{
		"transaction_id":     in.EntityID,
		"transaction_number": in.TransactionNumber,
		"store_id":           in.StoreID,
		"warehouse_code":     in.WarehouseCode,
		"total_amount":       in.TotalAmount,
		"shortage_items":     shortage,
		"reason":             in.Summary,
	})

	entityID := fmt.Sprintf("%d", in.EntityID)
	orgID := in.OrganizationID
	if orgID <= 0 {
		orgID = 1
	}

	upd, err := pool.Exec(ctx, `
		UPDATE approval_requests
		SET payload = $4, summary = $3, updated_at = NOW()
		WHERE organization_id = $1 AND domain = 'sap_invoice' AND action = 'stock_hold'
		  AND entity_type = 'pos_transaction' AND entity_id = $2 AND status = 'pending'
	`, orgID, entityID, in.Summary, payload)
	if err != nil {
		return err
	}
	if upd.RowsAffected() == 0 {
		_, err = pool.Exec(ctx, `
			INSERT INTO approval_requests (
				organization_id, domain, action, entity_type, entity_id,
				status, required_permission, summary, payload, metadata
			) VALUES (
				$1, 'sap_invoice', 'stock_hold', 'pos_transaction', $2,
				'pending', 'sap:invoice:override', $3, $4, '{}'::jsonb
			)`, orgID, entityID, in.Summary, payload)
	}
	return err
}

type MasterDataUpsertResult struct {
	Success    bool
	EntityID   int32
	ResultJSON []byte
	ErrorText  string
}

func (uc *SapAgentUseCase) ApplyMasterDataUpsert(ctx context.Context, tenantSlug string, orgID int32, operation string, payloadJSON []byte) (*MasterDataUpsertResult, error) {
	pool, err := uc.tenantManager.GetPool(ctx, tenantSlug)
	if err != nil {
		return nil, err
	}
	if orgID <= 0 {
		orgID = 1
	}

	res := &MasterDataUpsertResult{Success: true}

	switch operation {
	case "upsert_category":
		var p struct{ Code, Name string }
		if err := json.Unmarshal(payloadJSON, &p); err != nil {
			return nil, err
		}
		err = pool.QueryRow(ctx, `
			INSERT INTO product_categories (code, name, is_active, updated_at)
			VALUES ($1, $2, true, NOW())
			ON CONFLICT (code) DO UPDATE
			SET name = EXCLUDED.name, is_active = true, updated_at = NOW()
			RETURNING id`, p.Code, p.Name).Scan(&res.EntityID)

	case "upsert_uom":
		var p struct{ Code, Name string }
		if err := json.Unmarshal(payloadJSON, &p); err != nil {
			return nil, err
		}
		err = pool.QueryRow(ctx, `
			INSERT INTO units_of_measure (code, name, is_active)
			VALUES ($1, $2, true)
			ON CONFLICT (code) DO UPDATE
			SET name = EXCLUDED.name, is_active = true
			RETURNING id`, p.Code, p.Name).Scan(&res.EntityID)

	case "lookup_product_by_sku":
		var p struct {
			SKU string `json:"sku"`
		}
		if err := json.Unmarshal(payloadJSON, &p); err != nil {
			return nil, err
		}
		var productID int32
		var cost float64
		err = pool.QueryRow(ctx, `
			SELECT id, COALESCE(cost_price, 0) FROM products
			WHERE organization_id = $1 AND sku = $2 LIMIT 1`, orgID, p.SKU).Scan(&productID, &cost)
		if err != nil {
			res.ResultJSON, _ = json.Marshal(map[string]any{"found": false})
			return res, nil
		}
		res.EntityID = productID
		res.ResultJSON, _ = json.Marshal(map[string]any{"found": true, "product_id": productID, "cost": cost})
		return res, nil

	case "upsert_product":
		var p struct {
			SKU, Name, ForeignName string
			CategoryID, UomID, TaxCategoryID *int32
			IsWeighted             bool
			CostPrice              float64
			SkipCostUpdate         bool `json:"skip_cost_update"`
		}
		if err := json.Unmarshal(payloadJSON, &p); err != nil {
			return nil, err
		}
		metaJSON, _ := json.Marshal(map[string]any{
			"foreign_name": p.ForeignName,
			"synced_from":  "SAP_B1",
			"is_weighted":  p.IsWeighted,
		})
		err = pool.QueryRow(ctx, `
			INSERT INTO products (
				organization_id, sku, name, category_id, base_uom_id, tax_category_id,
				allow_decimal_quantity, is_active, is_sellable, is_purchasable,
				cost_price, metadata, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, true, true, true, $9, $8, NOW())
			ON CONFLICT (organization_id, sku) DO UPDATE
			SET name = EXCLUDED.name,
			    category_id = COALESCE(EXCLUDED.category_id, products.category_id),
			    base_uom_id = COALESCE(EXCLUDED.base_uom_id, products.base_uom_id),
			    tax_category_id = COALESCE(EXCLUDED.tax_category_id, products.tax_category_id),
			    allow_decimal_quantity = EXCLUDED.allow_decimal_quantity,
			    cost_price = CASE
			        WHEN $10 THEN products.cost_price
			        WHEN EXCLUDED.cost_price > 0 THEN EXCLUDED.cost_price
			        ELSE products.cost_price
			    END,
			    metadata = products.metadata || EXCLUDED.metadata,
			    updated_at = NOW()
			RETURNING id`, orgID, p.SKU, p.Name, p.CategoryID, p.UomID, p.TaxCategoryID, p.IsWeighted, metaJSON, p.CostPrice, p.SkipCostUpdate).Scan(&res.EntityID)

	case "upsert_barcode":
		var p struct {
			ProductID int32  `json:"product_id"`
			Barcode   string `json:"barcode"`
			IsPrimary bool   `json:"is_primary"`
		}
		if err := json.Unmarshal(payloadJSON, &p); err != nil {
			return nil, err
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO product_barcodes (product_id, barcode, is_primary)
			VALUES ($1, $2, $3)
			ON CONFLICT (barcode) DO UPDATE
			SET product_id = EXCLUDED.product_id,
			    is_primary = EXCLUDED.is_primary`, p.ProductID, p.Barcode, p.IsPrimary)

	case "upsert_product_price":
		var p struct {
			ProductID   int32   `json:"product_id"`
			PriceListID int32   `json:"price_list_id"`
			UomID       *int32  `json:"uom_id"`
			Price       float64 `json:"price"`
		}
		if err := json.Unmarshal(payloadJSON, &p); err != nil {
			return nil, err
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO product_prices (
				product_id, price_list_id, uom_id, price, min_quantity, is_active, updated_at
			)
			VALUES ($1, $2, $3, $4, 1, true, NOW())
			ON CONFLICT (product_id, price_list_id, product_variant_id, uom_id, min_quantity) DO UPDATE
			SET price = EXCLUDED.price,
			    uom_id = COALESCE(EXCLUDED.uom_id, product_prices.uom_id),
			    is_active = true,
			    updated_at = NOW()`, p.ProductID, p.PriceListID, p.UomID, p.Price)
		if err != nil {
			_, _ = pool.Exec(ctx, `
				UPDATE product_prices SET price = $4, updated_at = NOW()
				WHERE product_id = $1 AND price_list_id = $2`, p.ProductID, p.PriceListID, p.UomID, p.Price)
		}

	case "fetch_customer_sap_code":
		var p struct{ CustomerID int32 `json:"customer_id"` }
		if err := json.Unmarshal(payloadJSON, &p); err != nil {
			return nil, err
		}
		var code string
		err = pool.QueryRow(ctx, `
			SELECT COALESCE(NULLIF(customer_code, ''), NULLIF(metadata->>'sap_card_code', ''), '')
			FROM customers WHERE id = $1 LIMIT 1`, p.CustomerID).Scan(&code)
		res.ResultJSON, _ = json.Marshal(map[string]string{"code": code})
		return res, err

	case "fetch_original_sap_doc_entry":
		var p struct{ OriginalTxID int32 `json:"original_tx_id"` }
		if err := json.Unmarshal(payloadJSON, &p); err != nil {
			return nil, err
		}
		var docEntry int32
		err = pool.QueryRow(ctx, `SELECT COALESCE(NULLIF(metadata->>'sap_doc_entry', '')::int, 0) FROM pos_transactions WHERE id = $1`, p.OriginalTxID).Scan(&docEntry)
		res.EntityID = docEntry
		res.ResultJSON, _ = json.Marshal(map[string]int32{"doc_entry": docEntry})
		return res, err

	case "get_active_product_price":
		var p struct {
			ProductID   int32 `json:"product_id"`
			PriceListID int32 `json:"price_list_id"`
		}
		if err := json.Unmarshal(payloadJSON, &p); err != nil {
			return nil, err
		}
		var price float64
		err = pool.QueryRow(ctx, `
			SELECT price FROM product_prices
			WHERE product_id = $1 AND price_list_id = $2 AND is_active = true LIMIT 1`, p.ProductID, p.PriceListID).Scan(&price)
		if err != nil {
			res.Success = false
			return res, nil
		}
		res.ResultJSON, _ = json.Marshal(map[string]float64{"price": price})
		return res, nil

	case "get_product_cost_price":
		var p struct{ ProductID int32 `json:"product_id"` }
		if err := json.Unmarshal(payloadJSON, &p); err != nil {
			return nil, err
		}
		var cost float64
		err = pool.QueryRow(ctx, `SELECT COALESCE(cost_price, 0) FROM products WHERE id = $1`, p.ProductID).Scan(&cost)
		res.ResultJSON, _ = json.Marshal(map[string]float64{"cost": cost})
		return res, err

	case "get_approval_policy_rules":
		var p struct {
			Domain string `json:"domain"`
			Action string `json:"action"`
		}
		if err := json.Unmarshal(payloadJSON, &p); err != nil {
			return nil, err
		}
		if p.Domain == "" {
			p.Domain = "catalog"
		}
		if p.Action == "" {
			p.Action = "price_change"
		}
		var requiresApproval bool
		var rawRules []byte
		err = pool.QueryRow(ctx, `
			SELECT requires_approval, rules FROM approval_policies
			WHERE organization_id = $1 AND domain = $2 AND action = $3 AND is_enabled = true
			LIMIT 1`, orgID, p.Domain, p.Action).Scan(&requiresApproval, &rawRules)
		if err != nil {
			res.ResultJSON, _ = json.Marshal(map[string]any{"requires_approval": false})
			return res, nil
		}
		rules := map[string]any{}
		if len(rawRules) > 0 {
			_ = json.Unmarshal(rawRules, &rules)
		}
		res.ResultJSON, _ = json.Marshal(map[string]any{
			"requires_approval": requiresApproval,
			"rules":             rules,
		})
		return res, nil

	case "submit_price_approval_request":
		var p map[string]any
		if err := json.Unmarshal(payloadJSON, &p); err != nil {
			return nil, err
		}
		payload, _ := json.Marshal(p)
		sku, _ := p["product_sku"].(string)
		reason, _ := p["reason"].(string)
		productID, _ := p["product_id"].(float64)
		summary := fmt.Sprintf("Price/Cost revision for SKU %s: %s", sku, reason)
		err = SubmitPendingCatalogPriceApproval(ctx, pool, orgID, int32(productID), sku, payload, summary)

	default:
		return nil, fmt.Errorf("unsupported master data operation: %s", operation)
	}

	if err != nil {
		res.Success = false
		res.ErrorText = err.Error()
	}
	return res, err
}

type CloudReconciliationMetrics struct {
	Domains []contracts.DomainReconciliation
}

func (uc *SapAgentUseCase) GetReconciliationMetrics(ctx context.Context, tenantSlug string, orgID int) (*CloudReconciliationMetrics, error) {
	pool, err := uc.tenantManager.GetPool(ctx, tenantSlug)
	if err != nil {
		return nil, err
	}
	if orgID <= 0 {
		orgID = 1
	}

	type countSpec struct {
		domain contracts.DomainType
		query  string
	}
	specs := []countSpec{
		{contracts.DomainProducts, "SELECT COUNT(*) FROM products WHERE organization_id = $1"},
		{contracts.DomainPartners, "SELECT COUNT(*) FROM business_partners WHERE organization_id = $1"},
		{contracts.DomainInvoices, "SELECT COUNT(*) FROM invoices WHERE organization_id = $1"},
		{contracts.DomainInventory, "SELECT COUNT(*) FROM inventory_stock WHERE quantity_on_hand > 0"},
	}

	var domains []contracts.DomainReconciliation
	for _, spec := range specs {
		var count int64
		if spec.domain == contracts.DomainInventory {
			err = pool.QueryRow(ctx, spec.query).Scan(&count)
		} else {
			err = pool.QueryRow(ctx, spec.query, orgID).Scan(&count)
		}
		if err != nil {
			continue
		}
		domains = append(domains, contracts.DomainReconciliation{
			Domain:       spec.domain,
			TargetCount:  count,
			Status:       "CLOUD_COUNT",
			Notes:        "Live tenant database count",
		})
	}

	var invoiceSum, invSum sql.NullFloat64
	_ = pool.QueryRow(ctx, "SELECT SUM(total_amount) FROM invoices WHERE organization_id = $1", orgID).Scan(&invoiceSum)
	_ = pool.QueryRow(ctx, "SELECT SUM(quantity_on_hand) FROM inventory_stock WHERE quantity_on_hand > 0").Scan(&invSum)

	if invoiceSum.Valid {
		domains = append(domains, contracts.DomainReconciliation{
			Domain:           contracts.DomainInvoices,
			TargetNumericSum: math.Round(invoiceSum.Float64*100) / 100,
			Status:           "CLOUD_SUM",
			Notes:            "Invoice revenue sum in tenant DB",
		})
	}
	if invSum.Valid {
		domains = append(domains, contracts.DomainReconciliation{
			Domain:           contracts.DomainInventory,
			TargetNumericSum: math.Round(invSum.Float64*100) / 100,
			Status:           "CLOUD_SUM",
			Notes:            "Inventory on-hand sum in tenant DB",
		})
	}

	return &CloudReconciliationMetrics{Domains: domains}, nil
}
