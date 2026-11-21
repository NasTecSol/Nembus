package nembus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/config"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	cfg  *config.Config
	pool *pgxpool.Pool
}

type POSTransaction struct {
	ID                uuid.UUID       `json:"id"`
	StoreID           int32           `json:"store_id"`
	CashierID         int32           `json:"cashier_id"`
	CashierSessionID  uuid.UUID       `json:"cashier_session_id"`
	CustomerID        *uuid.UUID      `json:"customer_id"`
	CustomerSAPCode   string          `json:"customer_sap_code,omitempty"`
	TransactionNumber string          `json:"transaction_number"`
	TransactionDate   time.Time       `json:"transaction_date"`
	Subtotal          float64         `json:"subtotal"`
	DiscountAmount    float64         `json:"discount_amount"`
	TaxAmount         float64         `json:"tax_amount"`
	TotalAmount       float64         `json:"total_amount"`
	AmountPaid        float64         `json:"amount_paid"`
	Status            string          `json:"status"`
	Metadata          json.RawMessage `json:"metadata"`
}

type POSTransactionLine struct {
	ID                  uuid.UUID `json:"id"`
	TransactionID       uuid.UUID `json:"transaction_id"`
	ProductID           int32   `json:"product_id"`
	ProductSKU          string  `json:"product_sku"`
	ProductName         string  `json:"product_name"`
	Quantity            float64 `json:"quantity"`
	UoMID               *int32  `json:"uom_id"`
	UnitPrice           float64 `json:"unit_price"`
	DiscountAmount      float64 `json:"discount_amount"`
	TaxAmount           float64 `json:"tax_amount"`
	Subtotal            float64 `json:"subtotal"`
	LineTotal           float64 `json:"line_total"`
	LineNumber          int     `json:"line_number"`
	SAPUoMEntry         *int    `json:"sap_uom_entry,omitempty"`
	SAPUgpEntry         *int    `json:"sap_ugp_entry,omitempty"`
	UoMConversionFactor float64 `json:"uom_conversion_factor,omitempty"`
	IsBaseUoM           bool    `json:"is_base_uom,omitempty"`
	SAPTaxCode          string  `json:"sap_tax_code,omitempty"`
}

type POSPayment struct {
	ID            uuid.UUID `json:"id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	PaymentMethod string    `json:"payment_method"`
	Amount        float64   `json:"amount"`
	Reference     *string   `json:"reference"`
	PaymentDate   time.Time `json:"payment_date"`
}

type SalesReturn struct {
	ID                    uuid.UUID       `json:"id"`
	ReturnNumber          string          `json:"return_number"`
	StoreID               int32           `json:"store_id"`
	CashierID             *int32          `json:"cashier_id"`
	CustomerID            *uuid.UUID      `json:"customer_id"`
	CustomerSAPCode       string          `json:"customer_sap_code,omitempty"`
	OriginalTransactionID *uuid.UUID      `json:"original_transaction_id"`
	OriginalSAPDocEntry   int             `json:"original_sap_doc_entry,omitempty"`
	ReturnDate            time.Time       `json:"return_date"`
	ReturnReason          *string         `json:"return_reason"`
	Status                string          `json:"status"`
	Subtotal              float64         `json:"subtotal"`
	TaxAmount             float64         `json:"tax_amount"`
	TotalRefundAmount     float64         `json:"total_refund_amount"`
	Metadata              json.RawMessage `json:"metadata"`
}

type SalesReturnLine struct {
	ID                  uuid.UUID  `json:"id"`
	ReturnID            uuid.UUID  `json:"return_id"`
	ProductID           int32      `json:"product_id"`
	ProductSKU          string     `json:"product_sku"`
	ProductName         string     `json:"product_name"`
	OriginalLineID      *uuid.UUID `json:"original_line_id"`
	Quantity            float64 `json:"quantity"`
	UnitPrice           float64 `json:"unit_price"`
	RefundAmount        float64 `json:"refund_amount"`
	ReturnToStock       bool    `json:"return_to_stock"`
	LineNumber          int     `json:"line_number"`
	SAPUoMEntry         *int    `json:"sap_uom_entry,omitempty"`
	SAPUgpEntry         *int    `json:"sap_ugp_entry,omitempty"`
	UoMConversionFactor float64 `json:"uom_conversion_factor,omitempty"`
	IsBaseUoM           bool    `json:"is_base_uom,omitempty"`
	SAPTaxCode          string  `json:"sap_tax_code,omitempty"`
}

type ShortageItem struct {
	ProductID   int32   `json:"product_id"`
	ProductSKU  string  `json:"product_sku"`
	ProductName string  `json:"product_name"`
	Quantity    float64 `json:"quantity"`
	SAPOnHand   float64 `json:"sap_on_hand"`
}

type SalesOrderV2 struct {
	ID              uuid.UUID
	OrderNumber     string
	StoreID         int32
	CustomerID      *uuid.UUID
	CustomerSAPCode string
	OrderDate       time.Time
	TotalAmount     float64
	PaymentStatus   string
	Metadata        json.RawMessage
}

type SalesOrderLineV2 struct {
	ProductID           int32
	ProductSKU          string
	ProductName         string
	Quantity            float64
	UoMID               *int32
	UnitPrice           float64
	TaxAmount           float64
	LineTotal           float64
	LineNumber          int
	SAPUoMEntry         *int
	SAPUgpEntry         *int
	UoMConversionFactor float64
	IsBaseUoM           bool
	SAPTaxCode          string
}

func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.NembusDBURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Nembus DB connection string: %w", err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Nembus DB pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping Nembus DB: %w", err)
	}

	log.Printf(" Connected to Nembus Database (%s)", poolConfig.ConnConfig.Database)
	return &Client{
		cfg:  cfg,
		pool: pool,
	}, nil
}

func NewClientWithPool(cfg *config.Config, pool *pgxpool.Pool) *Client {
	return &Client{
		cfg:  cfg,
		pool: pool,
	}
}

func (c *Client) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}

func (c *Client) TestConnection(ctx context.Context) error {
	return c.pool.Ping(ctx)
}

// Master Data Upsert Helpers

func (c *Client) UpsertCategory(ctx context.Context, code, name string) (int32, error) {
	var id int32
	query := `
		INSERT INTO product_categories (code, name, is_active, updated_at)
		VALUES ($1, $2, true, NOW())
		ON CONFLICT (code) DO UPDATE
		SET name = EXCLUDED.name, is_active = true, updated_at = NOW()
		RETURNING id
	`
	err := c.pool.QueryRow(ctx, query, code, name).Scan(&id)
	return id, err
}

func (c *Client) UpsertUoM(ctx context.Context, code, name string) (int32, error) {
	var id int32
	query := `
		INSERT INTO units_of_measure (code, name, is_active)
		VALUES ($1, $2, true)
		ON CONFLICT (code) DO UPDATE
		SET name = EXCLUDED.name, is_active = true
		RETURNING id
	`
	err := c.pool.QueryRow(ctx, query, code, name).Scan(&id)
	return id, err
}

func (c *Client) UpsertProduct(ctx context.Context, sku, name, foreignName string, categoryID, uomID, taxCategoryID *int32, isWeighted bool, costPrice float64) (int32, error) {
	var id int32
	metaJSON, _ := json.Marshal(map[string]any{
		"foreign_name": foreignName,
		"synced_from":  "SAP_B1",
		"is_weighted":  isWeighted,
	})

	query := `
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
		    cost_price = CASE WHEN EXCLUDED.cost_price > 0 THEN EXCLUDED.cost_price ELSE products.cost_price END,
		    metadata = products.metadata || EXCLUDED.metadata,
		    updated_at = NOW()
		RETURNING id
	`
	err := c.pool.QueryRow(ctx, query,
		c.cfg.NembusOrganizationID, sku, name, categoryID, uomID, taxCategoryID, isWeighted, metaJSON, costPrice,
	).Scan(&id)
	return id, err
}

func (c *Client) UpsertBarcode(ctx context.Context, productID int32, barcode string, isPrimary bool) error {
	query := `
		INSERT INTO product_barcodes (product_id, barcode, is_primary)
		VALUES ($1, $2, $3)
		ON CONFLICT (barcode) DO UPDATE
		SET product_id = EXCLUDED.product_id,
		    is_primary = EXCLUDED.is_primary
	`
	_, err := c.pool.Exec(ctx, query, productID, barcode, isPrimary)
	return err
}

func (c *Client) UpsertProductPrice(ctx context.Context, productID, priceListID int32, uomID *int32, price float64) error {
	query := `
		INSERT INTO product_prices (
			product_id, price_list_id, uom_id, price, min_quantity, is_active, updated_at
		)
		VALUES ($1, $2, $3, $4, 1, true, NOW())
		ON CONFLICT (product_id, price_list_id, product_variant_id, uom_id, min_quantity) DO UPDATE
		SET price = EXCLUDED.price,
		    uom_id = COALESCE(EXCLUDED.uom_id, product_prices.uom_id),
		    is_active = true,
		    updated_at = NOW()
	`
	_, err := c.pool.Exec(ctx, query, productID, priceListID, uomID, price)
	if err != nil {
		// If conflict constraint is different, fallback to standard update / insert
		queryFallback := `
			UPDATE product_prices SET price = $4, updated_at = NOW()
			WHERE product_id = $1 AND price_list_id = $2;
		`
		if tag, _ := c.pool.Exec(ctx, queryFallback, productID, priceListID, uomID, price); tag.RowsAffected() == 0 {
			_, err = c.pool.Exec(ctx, `
				INSERT INTO product_prices (product_id, price_list_id, uom_id, price, is_active, updated_at)
				VALUES ($1, $2, $3, $4, true, NOW())
			`, productID, priceListID, uomID, price)
		}
	}
	return err
}

func parseOptionalIntString(raw string) *int {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &v
}

// Transaction Query & Posting State

func (c *Client) FetchUnpostedPOSTransactions(ctx context.Context, limit int) ([]POSTransaction, error) {
	query := `
		SELECT t.id, t.store_id, t.cashier_id, t.cashier_session_id, t.customer_id,
		       COALESCE(NULLIF(cust.customer_code, ''), NULLIF(cust.metadata->>'sap_card_code', ''), '')::text AS customer_sap_code,
		       t.transaction_number, t.transaction_date, t.subtotal, t.discount_amount,
		       t.tax_amount, t.total_amount, t.amount_paid, t.status, t.metadata
		FROM pos_transactions t
		LEFT JOIN customers cust ON cust.id = t.customer_id
		WHERE t.status = 'completed'
		  AND t.transaction_number NOT LIKE 'SIM-POS-%'
		  AND (t.metadata->>'sap_doc_num' IS NULL OR t.metadata->>'sap_doc_num' = '')
		  AND (t.metadata->>'sap_sync_state' IS NULL OR t.metadata->>'sap_sync_state' NOT IN ('held_stock', 'adjustment_created_pending_force', 'approved_pending_upload', 'synced'))
		ORDER BY t.id ASC
		LIMIT $1
	`
	return c.scanPOSTransactions(ctx, query, limit)
}

func (c *Client) FetchTransactionLines(ctx context.Context, transactionID uuid.UUID) ([]POSTransactionLine, error) {
	query := `
		SELECT l.id, l.transaction_id, l.product_id, p.sku, p.name,
		       l.quantity, l.uom_id, l.unit_price, l.discount_amount,
		       l.tax_amount, l.subtotal, l.line_total, COALESCE(l.line_number, 0),
		       COALESCE(NULLIF(u.metadata->>'sap_uom_entry', ''), NULLIF(bu.metadata->>'sap_uom_entry', ''), NULLIF(p.metadata->>'sap_i_uom_entry', ''), '')::text AS sap_uom_entry,
		       COALESCE(NULLIF(p.metadata->>'sap_ugp_entry', ''), '')::text AS sap_ugp_entry,
		       COALESCE(puc.conversion_factor, 1)::float8 AS uom_conversion_factor,
		       (l.uom_id IS NULL OR l.uom_id = p.base_uom_id)::boolean AS is_base_uom,
		       COALESCE(NULLIF(p.metadata->>'sap_vat_group', ''), tc.code, '')::text AS sap_tax_code
		FROM pos_transaction_lines l
		JOIN products p ON p.id = l.product_id
		LEFT JOIN units_of_measure u ON u.id = l.uom_id
		LEFT JOIN units_of_measure bu ON bu.id = p.base_uom_id
		LEFT JOIN product_uom_conversions puc ON puc.product_id = l.product_id AND puc.from_uom_id = l.uom_id AND puc.to_uom_id = p.base_uom_id
		LEFT JOIN tax_categories tc ON tc.id = p.tax_category_id
		WHERE l.transaction_id = $1
		ORDER BY l.id ASC
	`
	rows, err := c.pool.Query(ctx, query, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []POSTransactionLine
	for rows.Next() {
		var l POSTransactionLine
		var sapUoMEntryStr, sapUgpEntryStr string
		if err := rows.Scan(
			&l.ID, &l.TransactionID, &l.ProductID, &l.ProductSKU, &l.ProductName,
			&l.Quantity, &l.UoMID, &l.UnitPrice, &l.DiscountAmount,
			&l.TaxAmount, &l.Subtotal, &l.LineTotal, &l.LineNumber,
			&sapUoMEntryStr, &sapUgpEntryStr, &l.UoMConversionFactor, &l.IsBaseUoM, &l.SAPTaxCode,
		); err != nil {
			return nil, err
		}
		l.SAPUoMEntry = parseOptionalIntString(sapUoMEntryStr)
		l.SAPUgpEntry = parseOptionalIntString(sapUgpEntryStr)
		lines = append(lines, l)
	}
	return lines, rows.Err()
}

func (c *Client) FetchTransactionPayments(ctx context.Context, transactionID uuid.UUID) ([]POSPayment, error) {
	query := `
		SELECT id, transaction_id, payment_method, amount, reference_number, payment_date
		FROM pos_payments
		WHERE transaction_id = $1
	`
	rows, err := c.pool.Query(ctx, query, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []POSPayment
	for rows.Next() {
		var p POSPayment
		if err := rows.Scan(&p.ID, &p.TransactionID, &p.PaymentMethod, &p.Amount, &p.Reference, &p.PaymentDate); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}

func (c *Client) MarkTransactionPostedToSAP(ctx context.Context, transactionID uuid.UUID, sapDocEntry, sapDocNum int) error {
	metaPatch, _ := json.Marshal(map[string]any{
		"sap_doc_entry":  sapDocEntry,
		"sap_doc_num":    sapDocNum,
		"sap_posted_at":  time.Now().Format(time.RFC3339),
		"sap_sync_state": "synced",
	})

	query := `
		UPDATE pos_transactions
		SET metadata = COALESCE(metadata, '{}'::jsonb) || $2::jsonb
		WHERE id = $1
	`
	_, err := c.pool.Exec(ctx, query, transactionID, metaPatch)
	return err
}

func (c *Client) FetchUnpostedSalesReturns(ctx context.Context, limit int) ([]SalesReturn, error) {
	query := `
		SELECT sr.id, sr.return_number, sr.store_id, sr.cashier_id, sr.customer_id,
		       COALESCE(NULLIF(cust.customer_code, ''), NULLIF(cust.metadata->>'sap_card_code', ''), '')::text AS customer_sap_code,
		       sr.original_transaction_id, COALESCE(NULLIF(pt.metadata->>'sap_doc_entry', ''), '0')::int AS original_sap_doc_entry,
		       sr.return_date, sr.return_reason, sr.status,
		       sr.subtotal, sr.tax_amount, sr.total_refund_amount, sr.metadata
		FROM sales_returns sr
		LEFT JOIN customers cust ON cust.id = sr.customer_id
		LEFT JOIN pos_transactions pt ON pt.id = sr.original_transaction_id
		WHERE sr.status = 'completed'
		  AND (sr.metadata->>'sap_doc_num' IS NULL OR sr.metadata->>'sap_doc_num' = '')
		ORDER BY sr.id ASC
		LIMIT $1
	`
	rows, err := c.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []SalesReturn
	for rows.Next() {
		var r SalesReturn
		if err := rows.Scan(
			&r.ID, &r.ReturnNumber, &r.StoreID, &r.CashierID, &r.CustomerID,
			&r.CustomerSAPCode, &r.OriginalTransactionID, &r.OriginalSAPDocEntry,
			&r.ReturnDate, &r.ReturnReason, &r.Status,
			&r.Subtotal, &r.TaxAmount, &r.TotalRefundAmount, &r.Metadata,
		); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func (c *Client) FetchSalesReturnLines(ctx context.Context, returnID uuid.UUID) ([]SalesReturnLine, error) {
	query := `
		SELECT l.id, l.return_id, l.product_id, p.sku, p.name,
		       l.original_line_id, l.quantity, l.unit_price, l.refund_amount,
		       l.return_to_stock, COALESCE(l.line_number, 0),
		       COALESCE(NULLIF(u.metadata->>'sap_uom_entry', ''), NULLIF(bu.metadata->>'sap_uom_entry', ''), NULLIF(p.metadata->>'sap_i_uom_entry', ''), '')::text AS sap_uom_entry,
		       COALESCE(NULLIF(p.metadata->>'sap_ugp_entry', ''), '')::text AS sap_ugp_entry,
		       COALESCE(puc.conversion_factor, 1)::float8 AS uom_conversion_factor,
		       (ptl.uom_id IS NULL OR ptl.uom_id = p.base_uom_id)::boolean AS is_base_uom,
		       COALESCE(NULLIF(p.metadata->>'sap_vat_group', ''), tc.code, '')::text AS sap_tax_code
		FROM sales_return_lines l
		JOIN products p ON p.id = l.product_id
		LEFT JOIN pos_transaction_lines ptl ON ptl.id = l.original_line_id
		LEFT JOIN units_of_measure u ON u.id = ptl.uom_id
		LEFT JOIN units_of_measure bu ON bu.id = p.base_uom_id
		LEFT JOIN product_uom_conversions puc ON puc.product_id = l.product_id AND puc.from_uom_id = ptl.uom_id AND puc.to_uom_id = p.base_uom_id
		LEFT JOIN tax_categories tc ON tc.id = p.tax_category_id
		WHERE l.return_id = $1
		ORDER BY l.id ASC
	`
	rows, err := c.pool.Query(ctx, query, returnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []SalesReturnLine
	for rows.Next() {
		var l SalesReturnLine
		var sapUoMEntryStr, sapUgpEntryStr string
		if err := rows.Scan(
			&l.ID, &l.ReturnID, &l.ProductID, &l.ProductSKU, &l.ProductName,
			&l.OriginalLineID, &l.Quantity, &l.UnitPrice, &l.RefundAmount,
			&l.ReturnToStock, &l.LineNumber,
			&sapUoMEntryStr, &sapUgpEntryStr, &l.UoMConversionFactor, &l.IsBaseUoM, &l.SAPTaxCode,
		); err != nil {
			return nil, err
		}
		l.SAPUoMEntry = parseOptionalIntString(sapUoMEntryStr)
		l.SAPUgpEntry = parseOptionalIntString(sapUgpEntryStr)
		lines = append(lines, l)
	}
	return lines, rows.Err()
}

func (c *Client) MarkSalesReturnPostedToSAP(ctx context.Context, returnID uuid.UUID, sapDocEntry, sapDocNum int) error {
	metaPatch, _ := json.Marshal(map[string]any{
		"sap_doc_entry":  sapDocEntry,
		"sap_doc_num":    sapDocNum,
		"sap_posted_at":  time.Now().Format(time.RFC3339),
		"sap_sync_state": "synced",
	})

	query := `
		UPDATE sales_returns
		SET metadata = COALESCE(metadata, '{}'::jsonb) || $2::jsonb
		WHERE id = $1
	`
	_, err := c.pool.Exec(ctx, query, returnID, metaPatch)
	return err
}

func (c *Client) FetchOriginalSAPDocEntry(ctx context.Context, originalTxID uuid.UUID) (int, error) {
	var docEntry int
	query := `SELECT COALESCE(NULLIF(metadata->>'sap_doc_entry', '')::int, 0) FROM pos_transactions WHERE id = $1`
	err := c.pool.QueryRow(ctx, query, originalTxID).Scan(&docEntry)
	return docEntry, err
}

// FetchCustomerSAPCode looks up the SAP CardCode for a given Nembus customer ID.
// Reads from customers.customer_code or metadata->>'sap_card_code'.
func (c *Client) FetchCustomerSAPCode(ctx context.Context, customerID uuid.UUID) (string, error) {
	var code string
	query := `
		SELECT COALESCE(NULLIF(customer_code, ''), NULLIF(metadata->>'sap_card_code', ''), '')
		FROM customers
		WHERE id = $1
		LIMIT 1
	`
	err := c.pool.QueryRow(ctx, query, customerID).Scan(&code)
	return code, err
}

// HoldTransactionForStockReview marks a transaction as held in Nembus and submits an approval request.
func (c *Client) HoldTransactionForStockReview(ctx context.Context, txID uuid.UUID, txNum string, storeID int32, whCode string, totalAmount float64, shortageItems []ShortageItem, summary string) error {
	metaPatch, _ := json.Marshal(map[string]any{
		"sap_sync_state": "held_stock",
		"held_reason":    summary,
		"held_at":        time.Now().Format(time.RFC3339),
	})
	_, err := c.pool.Exec(ctx, `
		UPDATE pos_transactions
		SET metadata = COALESCE(metadata, '{}'::jsonb) || $2::jsonb, updated_at = NOW()
		WHERE id = $1
	`, txID, metaPatch)
	if err != nil {
		return fmt.Errorf("failed updating transaction metadata: %w", err)
	}

	payload, _ := json.Marshal(map[string]any{
		"transaction_id":     txID.String(),
		"transaction_number": txNum,
		"store_id":           storeID,
		"warehouse_code":     whCode,
		"total_amount":       totalAmount,
		"shortage_items":     shortageItems,
		"reason":             summary,
	})

	perm, err := c.getApprovalPolicyPermission(ctx, "sap_invoice", "stock_hold")
	if err != nil {
		return err
	}

	entityID := txID.String()
	query := `
		INSERT INTO approval_requests (
			organization_id, domain, action, entity_type, entity_id,
			status, required_permission, summary, payload, metadata
		) VALUES (
			$1, 'sap_invoice', 'stock_hold', 'pos_transaction', $2,
			'pending', $3, $4, $5, '{}'::jsonb
		)
		ON CONFLICT DO NOTHING
	`
	_, err = c.pool.Exec(ctx, query, c.cfg.NembusOrganizationID, entityID, perm, summary, payload)
	return err
}

// GetActiveProductPrice looks up the currently active price for a product on a price list and optional UOM.
func (c *Client) GetActiveProductPrice(ctx context.Context, productID, priceListID int32, uomID *int32) (float64, bool, error) {
	var price float64
	var err error
	if uomID != nil {
		err = c.pool.QueryRow(ctx, `
			SELECT price FROM product_prices
			WHERE product_id = $1 AND price_list_id = $2 AND uom_id = $3 AND is_active = true
			LIMIT 1
		`, productID, priceListID, *uomID).Scan(&price)
	} else {
		err = c.pool.QueryRow(ctx, `
			SELECT price FROM product_prices
			WHERE product_id = $1 AND price_list_id = $2 AND uom_id IS NULL AND is_active = true
			LIMIT 1
		`, productID, priceListID).Scan(&price)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return price, true, nil
}

// GetProductCostPrice returns the current cost price stored on the product header.
func (c *Client) GetProductCostPrice(ctx context.Context, productID int32) (float64, error) {
	var cost float64
	err := c.pool.QueryRow(ctx, `SELECT COALESCE(cost_price, 0) FROM products WHERE id = $1`, productID).Scan(&cost)
	return cost, err
}

func (c *Client) GetProductCostPriceBySKU(ctx context.Context, sku string) (float64, error) {
	var cost float64
	err := c.pool.QueryRow(ctx, `
		SELECT COALESCE(cost_price, 0) FROM products
		WHERE organization_id = $1 AND sku = $2
	`, c.cfg.NembusOrganizationID, sku).Scan(&cost)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return cost, err
}

// getApprovalPolicyPermission reads required_permission for a configured policy.
func (c *Client) getApprovalPolicyPermission(ctx context.Context, domain, action string) (string, error) {
	var perm string
	err := c.pool.QueryRow(ctx, `
		SELECT required_permission FROM approval_policies
		WHERE organization_id = $1 AND domain = $2 AND action = $3 AND is_enabled = true
		LIMIT 1
	`, c.cfg.NembusOrganizationID, domain, action).Scan(&perm)
	if err != nil {
		return "", fmt.Errorf("approval policy %s.%s not configured or disabled: %w", domain, action, err)
	}
	return perm, nil
}

// GetApprovalPolicyRules retrieves the JSON rules for a specific domain & action policy.
func (c *Client) GetApprovalPolicyRules(ctx context.Context, domain, action string) (map[string]any, bool, error) {
	var rawRules []byte
	var requiresApproval bool
	err := c.pool.QueryRow(ctx, `
		SELECT requires_approval, rules FROM approval_policies
		WHERE organization_id = $1 AND domain = $2 AND action = $3 AND is_enabled = true
		LIMIT 1
	`, c.cfg.NembusOrganizationID, domain, action).Scan(&requiresApproval, &rawRules)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("load approval policy: %w", err)
	}
	if !requiresApproval {
		return nil, false, nil
	}
	rules := make(map[string]any)
	if len(rawRules) > 0 {
		_ = json.Unmarshal(rawRules, &rules)
	}
	return rules, true, nil
}

// SubmitPriceApprovalRequest creates or updates a pending price/cost review request.
func (c *Client) SubmitPriceApprovalRequest(ctx context.Context, productID int32, sku string, priceListID int32, uomID *int32, oldPrice, newPrice, oldCost, newCost *float64, reason string) error {
	payload, _ := json.Marshal(map[string]any{
		"product_id":     productID,
		"product_sku":    sku,
		"price_list_id":  priceListID,
		"uom_id":         uomID,
		"old_price":      oldPrice,
		"new_price":      newPrice,
		"old_cost_price": oldCost,
		"new_cost_price": newCost,
		"reason":         reason,
	})
	entityID := fmt.Sprintf("%d", productID)
	summary := fmt.Sprintf("Price/Cost revision for SKU %s: %s", sku, reason)

	var existingID int64
	err := c.pool.QueryRow(ctx, `
		SELECT id FROM approval_requests
		WHERE organization_id = $1 AND domain = 'catalog' AND action = 'price_change'
		  AND entity_type = 'product' AND entity_id = $2 AND status = 'pending'
		ORDER BY id DESC LIMIT 1
	`, c.cfg.NembusOrganizationID, entityID).Scan(&existingID)
	if err == nil && existingID > 0 {
		_, err = c.pool.Exec(ctx, `
			UPDATE approval_requests
			SET payload = $3, summary = $4, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND organization_id = $2 AND status = 'pending'
		`, existingID, c.cfg.NembusOrganizationID, payload, summary)
		return err
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	perm, err := c.getApprovalPolicyPermission(ctx, "catalog", "price_change")
	if err != nil {
		return err
	}

	_, err = c.pool.Exec(ctx, `
		INSERT INTO approval_requests (
			organization_id, domain, action, entity_type, entity_id,
			status, required_permission, summary, payload, metadata
		) VALUES (
			$1, 'catalog', 'price_change', 'product', $2,
			'pending', $3, $4, $5, '{}'::jsonb
		)
	`, c.cfg.NembusOrganizationID, entityID, perm, summary, payload)
	return err
}

func (c *Client) FetchApprovedPendingPOSTransactions(ctx context.Context, limit int) ([]POSTransaction, error) {
	query := `
		SELECT t.id, t.store_id, t.cashier_id, t.cashier_session_id, t.customer_id,
		       COALESCE(NULLIF(cust.customer_code, ''), NULLIF(cust.metadata->>'sap_card_code', ''), '')::text AS customer_sap_code,
		       t.transaction_number, t.transaction_date, t.subtotal, t.discount_amount,
		       t.tax_amount, t.total_amount, t.amount_paid, t.status, t.metadata
		FROM pos_transactions t
		LEFT JOIN customers cust ON cust.id = t.customer_id
		WHERE t.status = 'completed'
		  AND t.metadata->>'sap_sync_state' = 'approved_pending_upload'
		  AND (t.metadata->>'sap_doc_num' IS NULL OR t.metadata->>'sap_doc_num' = '')
		ORDER BY t.id ASC
		LIMIT $1
	`
	return c.scanPOSTransactions(ctx, query, limit)
}

func (c *Client) scanPOSTransactions(ctx context.Context, query string, limit int) ([]POSTransaction, error) {
	rows, err := c.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []POSTransaction
	for rows.Next() {
		var t POSTransaction
		if err := rows.Scan(
			&t.ID, &t.StoreID, &t.CashierID, &t.CashierSessionID, &t.CustomerID, &t.CustomerSAPCode,
			&t.TransactionNumber, &t.TransactionDate, &t.Subtotal, &t.DiscountAmount,
			&t.TaxAmount, &t.TotalAmount, &t.AmountPaid, &t.Status, &t.Metadata,
		); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (c *Client) FetchUnpostedSalesOrdersV2(ctx context.Context, limit int) ([]SalesOrderV2, error) {
	query := `
		SELECT o.id, o.order_number, COALESCE(o.store_id, 0), o.customer_id,
		       COALESCE(NULLIF(cust.customer_code, ''), NULLIF(cust.metadata->>'sap_card_code', ''), '')::text AS customer_sap_code,
		       o.order_date, COALESCE(o.total_amount, 0), COALESCE(o.payment_status::text, ''),
		       COALESCE(o.metadata, '{}'::jsonb)
		FROM sales_orders_v2 o
		LEFT JOIN customers cust ON cust.id = o.customer_id
		WHERE o.organization_id = $1
		  AND o.order_number NOT LIKE 'SO-SAP-%'
		  AND LOWER(COALESCE(o.sales_channel, '')) <> 'pos'
		  AND (o.metadata->>'sap_doc_num' IS NULL OR o.metadata->>'sap_doc_num' = '')
		  AND (o.metadata->>'sap_sync_state' IS NULL OR o.metadata->>'sap_sync_state' NOT IN ('held_stock', 'adjustment_created_pending_force', 'approved_pending_upload', 'synced'))
		  AND (
		    o.fulfillment_status = 'fulfilled'
		    OR o.order_status IN ('confirmed', 'fulfilled')
		  )
		  AND NOT EXISTS (
		    SELECT 1 FROM pos_transactions pt WHERE pt.sales_order_id = o.id
		  )
		ORDER BY o.order_date ASC
		LIMIT $2
	`
	rows, err := c.pool.Query(ctx, query, c.cfg.NembusOrganizationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []SalesOrderV2
	for rows.Next() {
		var o SalesOrderV2
		var customerID pgtype.UUID
		if err := rows.Scan(&o.ID, &o.OrderNumber, &o.StoreID, &customerID, &o.CustomerSAPCode, &o.OrderDate, &o.TotalAmount, &o.PaymentStatus, &o.Metadata); err != nil {
			return nil, err
		}
		if customerID.Valid {
			u := uuid.UUID(customerID.Bytes)
			o.CustomerID = &u
		}
		list = append(list, o)
	}
	return list, rows.Err()
}

func (c *Client) FetchApprovedPendingSalesOrdersV2(ctx context.Context, limit int) ([]SalesOrderV2, error) {
	query := `
		SELECT o.id, o.order_number, COALESCE(o.store_id, 0), o.customer_id,
		       COALESCE(NULLIF(cust.customer_code, ''), NULLIF(cust.metadata->>'sap_card_code', ''), '')::text AS customer_sap_code,
		       o.order_date, COALESCE(o.total_amount, 0), COALESCE(o.payment_status::text, ''),
		       COALESCE(o.metadata, '{}'::jsonb)
		FROM sales_orders_v2 o
		LEFT JOIN customers cust ON cust.id = o.customer_id
		WHERE o.organization_id = $1
		  AND o.metadata->>'sap_sync_state' = 'approved_pending_upload'
		  AND (o.metadata->>'sap_doc_num' IS NULL OR o.metadata->>'sap_doc_num' = '')
		  AND NOT EXISTS (
		    SELECT 1 FROM pos_transactions pt WHERE pt.sales_order_id = o.id
		  )
		ORDER BY o.order_date ASC
		LIMIT $2
	`
	rows, err := c.pool.Query(ctx, query, c.cfg.NembusOrganizationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []SalesOrderV2
	for rows.Next() {
		var o SalesOrderV2
		var customerID pgtype.UUID
		if err := rows.Scan(&o.ID, &o.OrderNumber, &o.StoreID, &customerID, &o.CustomerSAPCode, &o.OrderDate, &o.TotalAmount, &o.PaymentStatus, &o.Metadata); err != nil {
			return nil, err
		}
		if customerID.Valid {
			u := uuid.UUID(customerID.Bytes)
			o.CustomerID = &u
		}
		list = append(list, o)
	}
	return list, rows.Err()
}

func (c *Client) FetchSalesOrderLinesV2(ctx context.Context, orderID uuid.UUID) ([]SalesOrderLineV2, error) {
	query := `
		SELECT l.product_id, COALESCE(p.sku, l.product_sku, ''), COALESCE(p.name, l.product_name, ''),
		       l.quantity_ordered, l.uom_id, l.unit_price, COALESCE(l.tax_amount, 0), COALESCE(l.line_total, 0),
		       COALESCE(l.line_number, 0),
		       COALESCE(NULLIF(u.metadata->>'sap_uom_entry', ''), NULLIF(bu.metadata->>'sap_uom_entry', ''), NULLIF(p.metadata->>'sap_i_uom_entry', ''), '')::text AS sap_uom_entry,
		       COALESCE(NULLIF(p.metadata->>'sap_ugp_entry', ''), '')::text AS sap_ugp_entry,
		       COALESCE(puc.conversion_factor, 1)::float8 AS uom_conversion_factor,
		       (l.uom_id IS NULL OR l.uom_id = p.base_uom_id)::boolean AS is_base_uom,
		       COALESCE(NULLIF(p.metadata->>'sap_vat_group', ''), tc.code, '')::text AS sap_tax_code
		FROM sales_order_lines_v2 l
		LEFT JOIN products p ON p.id = l.product_id
		LEFT JOIN units_of_measure u ON u.id = l.uom_id
		LEFT JOIN units_of_measure bu ON bu.id = p.base_uom_id
		LEFT JOIN product_uom_conversions puc ON puc.product_id = l.product_id AND puc.from_uom_id = l.uom_id AND puc.to_uom_id = p.base_uom_id
		LEFT JOIN tax_categories tc ON tc.id = COALESCE(l.tax_category_id, p.tax_category_id)
		WHERE l.sales_order_id = $1
		ORDER BY l.line_number, l.id
	`
	rows, err := c.pool.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []SalesOrderLineV2
	for rows.Next() {
		var l SalesOrderLineV2
		var sapUoMEntryStr, sapUgpEntryStr string
		if err := rows.Scan(
			&l.ProductID, &l.ProductSKU, &l.ProductName,
			&l.Quantity, &l.UoMID, &l.UnitPrice, &l.TaxAmount, &l.LineTotal, &l.LineNumber,
			&sapUoMEntryStr, &sapUgpEntryStr, &l.UoMConversionFactor, &l.IsBaseUoM, &l.SAPTaxCode,
		); err != nil {
			return nil, err
		}
		l.SAPUoMEntry = parseOptionalIntString(sapUoMEntryStr)
		l.SAPUgpEntry = parseOptionalIntString(sapUgpEntryStr)
		lines = append(lines, l)
	}
	return lines, rows.Err()
}

func (c *Client) MarkSalesOrderPostedToSAP(ctx context.Context, orderID uuid.UUID, sapDocEntry, sapDocNum int) error {
	metaPatch, _ := json.Marshal(map[string]any{
		"sap_doc_entry":  sapDocEntry,
		"sap_doc_num":    sapDocNum,
		"sap_posted_at":  time.Now().Format(time.RFC3339),
		"sap_sync_state": "synced",
	})

	query := `
		UPDATE sales_orders_v2
		SET metadata = COALESCE(metadata, '{}'::jsonb) || $2::jsonb, updated_at = NOW()
		WHERE id = $1
	`
	_, err := c.pool.Exec(ctx, query, orderID, metaPatch)
	return err
}

func (c *Client) HoldSalesOrderForStockReview(ctx context.Context, orderID uuid.UUID, orderNum string, storeID int32, whCode string, totalAmount float64, shortageItems []ShortageItem, summary string) error {
	metaPatch, _ := json.Marshal(map[string]any{
		"sap_sync_state": "held_stock",
		"held_reason":    summary,
		"held_at":        time.Now().Format(time.RFC3339),
	})
	_, err := c.pool.Exec(ctx, `
		UPDATE sales_orders_v2
		SET metadata = COALESCE(metadata, '{}'::jsonb) || $2::jsonb, updated_at = NOW()
		WHERE id = $1
	`, orderID, metaPatch)
	if err != nil {
		return fmt.Errorf("failed updating sales order metadata: %w", err)
	}

	payload, _ := json.Marshal(map[string]any{
		"sales_order_id": orderID.String(),
		"order_number":   orderNum,
		"store_id":       storeID,
		"warehouse_code": whCode,
		"total_amount":   totalAmount,
		"shortage_items": shortageItems,
		"reason":         summary,
	})

	perm, err := c.getApprovalPolicyPermission(ctx, "sap_invoice", "stock_hold")
	if err != nil {
		return err
	}

	entityID := orderID.String()
	query := `
		INSERT INTO approval_requests (
			organization_id, domain, action, entity_type, entity_id,
			status, required_permission, summary, payload, metadata
		) VALUES (
			$1, 'sap_invoice', 'stock_hold', 'sales_order_v2', $2,
			'pending', $3, $4, $5, '{}'::jsonb
		)
	`
	_, err = c.pool.Exec(ctx, query, c.cfg.NembusOrganizationID, entityID, perm, summary, payload)
	return err
}

