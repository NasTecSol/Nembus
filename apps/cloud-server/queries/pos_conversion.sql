-- name: CreateTransactionFromOrder :one
INSERT INTO pos_transactions (
    store_id, cashier_id, cashier_session_id, customer_id, 
    pos_terminal_id, sales_order_id, source_cart_id,
    transaction_number, total_amount, status, metadata
)
SELECT 
    o.store_id, o.cashier_id, COALESCE(
        (SELECT cs.id FROM cashier_sessions cs JOIN pos_terminals pt ON pt.id = cs.pos_terminal_id WHERE cs.status = 'open' AND cs.cashier_id = o.cashier_id AND pt.store_id = o.store_id ORDER BY cs.opening_time DESC LIMIT 1),
        (SELECT cs.id FROM cashier_sessions cs JOIN pos_terminals pt ON pt.id = cs.pos_terminal_id WHERE cs.status = 'open' AND pt.store_id = o.store_id ORDER BY cs.opening_time DESC LIMIT 1)
    ),
    o.customer_id, o.pos_terminal_id, o.id, o.source_cart_id,
    'TXN-' || o.order_number, o.total_amount, 'completed',
    COALESCE(o.metadata, '{}'::jsonb)
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
