package nembus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	cfg  *config.Config
	pool *pgxpool.Pool
}

type POSTransaction struct {
	ID                int64           `json:"id"`
	StoreID           int32           `json:"store_id"`
	CashierID         int32           `json:"cashier_id"`
	CashierSessionID  int32           `json:"cashier_session_id"`
	CustomerID        *int32          `json:"customer_id"`
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
	ID             int64   `json:"id"`
	TransactionID  int64   `json:"transaction_id"`
	ProductID      int32   `json:"product_id"`
	ProductSKU     string  `json:"product_sku"`
	ProductName    string  `json:"product_name"`
	Quantity       float64 `json:"quantity"`
	UoMID          *int32  `json:"uom_id"`
	UnitPrice      float64 `json:"unit_price"`
	DiscountAmount float64 `json:"discount_amount"`
	TaxAmount      float64 `json:"tax_amount"`
	Subtotal       float64 `json:"subtotal"`
	LineTotal      float64 `json:"line_total"`
	LineNumber     int     `json:"line_number"`
}

type POSPayment struct {
	ID            int64     `json:"id"`
	TransactionID int64     `json:"transaction_id"`
	PaymentMethod string    `json:"payment_method"`
	Amount        float64   `json:"amount"`
	Reference     *string   `json:"reference"`
	PaymentDate   time.Time `json:"payment_date"`
}

type SyncQueueItem struct {
	ID            int64           `json:"id"`
	EntityType    string          `json:"entity_type"`
	EntityID      int64           `json:"entity_id"`
	Action        string          `json:"action"`
	Payload       json.RawMessage `json:"payload"`
	Status        string          `json:"status"`
	CorrelationID *string         `json:"correlation_id"`
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

func (c *Client) UpsertProduct(ctx context.Context, sku, name, foreignName string, categoryID, uomID, taxCategoryID *int32, isWeighted bool) (int32, error) {
	var id int32
	metaJSON, _ := json.Marshal(map[string]any{
		"foreign_name": foreignName,
		"synced_from":  "SAP_B1",
		"is_weighted":  isWeighted,
	})

	query := `
		INSERT INTO products (
			organization_id, sku, name, category_id, base_uom_id, tax_category_id,
			allow_decimal_quantity, is_active, is_sellable, is_purchasable, metadata, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, true, true, true, $8, NOW())
		ON CONFLICT (organization_id, sku) DO UPDATE
		SET name = EXCLUDED.name,
		    category_id = COALESCE(EXCLUDED.category_id, products.category_id),
		    base_uom_id = COALESCE(EXCLUDED.base_uom_id, products.base_uom_id),
		    tax_category_id = COALESCE(EXCLUDED.tax_category_id, products.tax_category_id),
		    allow_decimal_quantity = EXCLUDED.allow_decimal_quantity,
		    metadata = products.metadata || EXCLUDED.metadata,
		    updated_at = NOW()
		RETURNING id
	`
	err := c.pool.QueryRow(ctx, query,
		c.cfg.NembusOrganizationID, sku, name, categoryID, uomID, taxCategoryID, isWeighted, metaJSON,
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
		INSERT INTO product_prices (product_id, price_list_id, uom_id, price, is_active, updated_at)
		VALUES ($1, $2, $3, $4, true, NOW())
		ON CONFLICT (product_id, price_list_id) DO UPDATE
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

// Transaction Query & Posting State

func (c *Client) FetchUnpostedPOSTransactions(ctx context.Context, limit int) ([]POSTransaction, error) {
	query := `
		SELECT id, store_id, cashier_id, cashier_session_id, customer_id,
		       transaction_number, transaction_date, subtotal, discount_amount,
		       tax_amount, total_amount, amount_paid, status, metadata
		FROM pos_transactions
		WHERE status = 'completed'
		  AND (metadata->>'sap_doc_num' IS NULL OR metadata->>'sap_doc_num' = '')
		ORDER BY id ASC
		LIMIT $1
	`
	rows, err := c.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []POSTransaction
	for rows.Next() {
		var t POSTransaction
		if err := rows.Scan(
			&t.ID, &t.StoreID, &t.CashierID, &t.CashierSessionID, &t.CustomerID,
			&t.TransactionNumber, &t.TransactionDate, &t.Subtotal, &t.DiscountAmount,
			&t.TaxAmount, &t.TotalAmount, &t.AmountPaid, &t.Status, &t.Metadata,
		); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (c *Client) FetchTransactionLines(ctx context.Context, transactionID int64) ([]POSTransactionLine, error) {
	query := `
		SELECT l.id, l.transaction_id, l.product_id, p.sku, p.name,
		       l.quantity, l.uom_id, l.unit_price, l.discount_amount,
		       l.tax_amount, l.subtotal, l.line_total, COALESCE(l.line_number, 0)
		FROM pos_transaction_lines l
		JOIN products p ON p.id = l.product_id
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
		if err := rows.Scan(
			&l.ID, &l.TransactionID, &l.ProductID, &l.ProductSKU, &l.ProductName,
			&l.Quantity, &l.UoMID, &l.UnitPrice, &l.DiscountAmount,
			&l.TaxAmount, &l.Subtotal, &l.LineTotal, &l.LineNumber,
		); err != nil {
			return nil, err
		}
		lines = append(lines, l)
	}
	return lines, rows.Err()
}

func (c *Client) FetchTransactionPayments(ctx context.Context, transactionID int64) ([]POSPayment, error) {
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

func (c *Client) MarkTransactionPostedToSAP(ctx context.Context, transactionID int64, sapDocEntry, sapDocNum int) error {
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

func (c *Client) FetchPendingSyncQueue(ctx context.Context, limit int) ([]SyncQueueItem, error) {
	query := `
		SELECT id, entity_type, entity_id, action, payload, status, correlation_id
		FROM sync_queue
		WHERE status = 'pending' AND entity_type IN ('pos_transactions', 'pos_payments')
		ORDER BY priority DESC, created_at ASC
		LIMIT $1
	`
	rows, err := c.pool.Query(ctx, query, limit)
	if err != nil {
		// Table might not exist yet if goose migration hasn't run
		return nil, nil
	}
	defer rows.Close()

	var items []SyncQueueItem
	for rows.Next() {
		var item SyncQueueItem
		if err := rows.Scan(
			&item.ID, &item.EntityType, &item.EntityID, &item.Action,
			&item.Payload, &item.Status, &item.CorrelationID,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (c *Client) UpdateSyncQueueStatus(ctx context.Context, id int64, status string, lastErr string) error {
	var query string
	if status == "synced" {
		query = `UPDATE sync_queue SET status = $2, synced_at = NOW() WHERE id = $1`
		_, err := c.pool.Exec(ctx, query, id, status)
		return err
	}
	query = `UPDATE sync_queue SET status = $2, retry_count = retry_count + 1, last_error = $3 WHERE id = $1`
	_, err := c.pool.Exec(ctx, query, id, status, lastErr)
	return err
}
