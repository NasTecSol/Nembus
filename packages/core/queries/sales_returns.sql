-- sales_returns.sql

-- name: CreateSalesReturn :one
INSERT INTO sales_returns (
    return_number,
    store_id,
    cashier_id,
    cashier_session_id,
    customer_id,
    original_transaction_id,
    return_date,
    return_reason,
    status,
    subtotal,
    tax_amount,
    total_refund_amount,
    refund_method,
    refund_reference,
    approved_by,
    notes,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
)
RETURNING *;

-- name: CreateSalesReturnLine :one
INSERT INTO sales_return_lines (
    return_id,
    product_id,
    product_variant_id,
    original_line_id,
    quantity,
    unit_price,
    refund_amount,
    return_to_stock,
    serial_number,
    batch_number,
    condition,
    line_number,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: GetSalesReturnByID :one
SELECT * FROM sales_returns
WHERE id = $1;

-- name: ListSalesReturnsByTransaction :many
SELECT * FROM sales_returns
WHERE original_transaction_id = $1;

-- name: ListSalesReturnLinesByReturnID :many
SELECT * FROM sales_return_lines
WHERE return_id = $1;

-- name: GetReturnedQuantityByLineID :one
SELECT COALESCE(SUM(quantity), 0)::DECIMAL(15,3) AS total_returned_qty
FROM sales_return_lines
WHERE original_line_id = $1;

-- name: UpdatePOSTransactionStatus :exec
UPDATE pos_transactions
SET status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;


