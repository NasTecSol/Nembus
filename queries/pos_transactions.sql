-- name: CreatePosTransaction :one
INSERT INTO pos_transactions (
    transaction_number,
    store_id,
    pos_terminal_id,
    cashier_session_id,
    cashier_id,
    customer_id,
    price_list_id,
    transaction_type,
    transaction_date,
    subtotal,
    tax_amount,
    discount_amount,
    total_amount,
    total_cost,
    status,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12, $13, $14, $15, $16
) RETURNING id, transaction_number, status, total_amount;

-- name: CreatePosTransactionLine :exec
INSERT INTO pos_transaction_lines (
    transaction_id, line_number, product_id, product_variant_id,
    serial_number, batch_number, quantity, uom_id,
    unit_price, discount_amount, tax_amount, line_total, cost_price, metadata
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14);

-- name: GetPosTransactionFull :many
SELECT 
    t.id, t.transaction_number, t.transaction_date, t.status,
    t.subtotal, t.tax_amount, t.discount_amount, t.total_amount,
    t.total_cost,
    cashier.first_name || ' ' || cashier.last_name AS cashier_name,
    term.terminal_name,
    sess.session_number,
    cust.name AS customer_name,
    tl.line_number,
    tl.product_id,
    tl.quantity,
    tl.unit_price,
    tl.discount_amount,
    tl.line_total,
    p.sku,
    p.name AS product_name,
    COALESCE(pb.barcode, '') AS scanned_barcode
FROM pos_transactions t
JOIN cashiers          cshr  ON t.cashier_id         = cshr.id
JOIN users             cashier ON cshr.user_id        = cashier.id
JOIN pos_terminals     term  ON t.pos_terminal_id     = term.id
JOIN cashier_sessions  sess  ON t.cashier_session_id  = sess.id
LEFT JOIN customers    cust  ON t.customer_id         = cust.id
JOIN pos_transaction_lines tl   ON tl.transaction_id = t.id
JOIN products          p     ON tl.product_id         = p.id
LEFT JOIN product_barcodes pb 
    ON pb.product_id = p.id 
   AND pb.is_primary = true
   AND (pb.product_variant_id = tl.product_variant_id OR tl.product_variant_id IS NULL)
WHERE t.id = $1
ORDER BY tl.line_number;

-- name: ListTodaysPosTransactions :many
SELECT 
    t.id,
    t.transaction_number,
    t.transaction_date,
    t.total_amount,
    t.status,
    cashier.first_name || ' ' || cashier.last_name AS cashier_name,
    term.terminal_name,
    COUNT(tl.id) AS items_count,
    SUM(tl.quantity) AS total_quantity
FROM pos_transactions t
JOIN cashiers cshr ON t.cashier_id = cshr.id
JOIN users cashier ON cshr.user_id = cashier.id
JOIN pos_terminals term ON t.pos_terminal_id = term.id
JOIN pos_transaction_lines tl ON tl.transaction_id = t.id
WHERE t.store_id = $1
  AND t.transaction_date >= CURRENT_DATE
  AND t.transaction_date < CURRENT_DATE + INTERVAL '1 day'
GROUP BY t.id, cashier.first_name, cashier.last_name, term.terminal_name
ORDER BY t.transaction_date DESC
LIMIT 200;

-- name: VoidPosTransaction :execrows
UPDATE pos_transactions
SET 
    status     = 'voided',
    voided_by  = $2,
    voided_at  = CURRENT_TIMESTAMP,
    metadata   = jsonb_set(metadata, '{void_reason}', to_jsonb($3::text))
WHERE id = $1
  AND status = 'completed'
  AND voided_at IS NULL;