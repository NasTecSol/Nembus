-- name: CreateTransactionFromOrder :one
INSERT INTO pos_transactions (
    store_id, cashier_id, cashier_session_id, customer_id, 
    pos_terminal_id, sales_order_id, source_cart_id,
    transaction_number, transaction_type,
    subtotal, discount_amount, tax_amount, total_amount, amount_paid,
    status, metadata
)
SELECT 
    o.store_id, o.cashier_id, COALESCE(
        (SELECT cs.id FROM cashier_sessions cs JOIN pos_terminals pt ON pt.id = cs.pos_terminal_id WHERE cs.status = 'open' AND cs.cashier_id = o.cashier_id AND pt.store_id = o.store_id ORDER BY cs.opening_time DESC LIMIT 1),
        (SELECT cs.id FROM cashier_sessions cs JOIN pos_terminals pt ON pt.id = cs.pos_terminal_id WHERE cs.status = 'open' AND pt.store_id = o.store_id ORDER BY cs.opening_time DESC LIMIT 1)
    ),
    o.customer_id, o.pos_terminal_id, o.id, o.source_cart_id,
    'TXN-' || o.order_number, 'sale',
    COALESCE(o.subtotal, 0), COALESCE(o.discount_amount, 0), COALESCE(o.tax_amount, 0), o.total_amount, COALESCE(NULLIF(o.paid_amount, 0), o.total_amount, 0),
    'completed',
    COALESCE(o.metadata, '{}'::jsonb)
FROM sales_orders_v2 o
WHERE o.id = $1
RETURNING *;

-- name: SyncTransactionLinesFromOrder :exec
INSERT INTO pos_transaction_lines (
    transaction_id, product_id, product_variant_id, 
    quantity, uom_id, unit_price, discount_amount, tax_amount, subtotal, line_total, cost_price, line_number
)
SELECT 
    $1, product_id, product_variant_id, 
    quantity_ordered, uom_id, unit_price, COALESCE(discount_amount, 0), COALESCE(tax_amount, 0),
    GREATEST(0, (quantity_ordered * unit_price) - COALESCE(discount_amount, 0)), line_total, COALESCE(unit_cost, 0), line_number
FROM sales_order_lines_v2
WHERE sales_order_id = $2;
