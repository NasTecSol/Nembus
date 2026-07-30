-- name: CreateTransactionFromOrder :one
INSERT INTO pos_transactions (
    store_id, cashier_id, cashier_session_id, customer_id, 
    pos_terminal_id, sales_order_id, source_cart_id,
    transaction_number, total_amount, status
)
SELECT 
    store_id, cashier_id, COALESCE(
        (SELECT id FROM cashier_sessions WHERE status = 'open' AND cashier_id = o.cashier_id LIMIT 1),
        (SELECT id FROM cashier_sessions WHERE status = 'open' AND store_id = o.store_id LIMIT 1),
        (SELECT id FROM cashier_sessions WHERE status = 'open' ORDER BY opening_time DESC LIMIT 1)
    ),
    customer_id, pos_terminal_id, id, source_cart_id,
    'TXN-' || order_number, total_amount, 'completed'
FROM sales_orders_v2 o
WHERE o.id = $1
RETURNING *;

-- name: SyncTransactionLinesFromOrder :exec
INSERT INTO pos_transaction_lines (
    transaction_id, product_id, product_variant_id, 
    quantity, unit_price, subtotal, line_total
)
SELECT 
    $1, product_id, product_variant_id, 
    quantity_ordered, unit_price, (quantity_ordered * unit_price), line_total
FROM sales_order_lines_v2
WHERE sales_order_id = $2;
