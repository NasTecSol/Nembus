-- restaurant_orders.sql

-- name: GetRestaurantOrder :one
SELECT * FROM restaurant_orders
WHERE id = $1 LIMIT 1;

-- name: GetRestaurantOrderByNumber :one
SELECT * FROM restaurant_orders
WHERE store_id = $1 AND order_number = $2 LIMIT 1;

-- name: ListRestaurantOrders :many
SELECT * FROM restaurant_orders
WHERE store_id = $1
ORDER BY ordered_at DESC;

-- name: CreateRestaurantOrder :one
INSERT INTO restaurant_orders (
    store_id, table_id, cashier_id, cashier_session_id, customer_id, order_number, order_source, status, subtotal, discount_amount, tax_amount, total_amount, amount_paid, change_given, notes, pos_transaction_id, metadata, ordered_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
)
RETURNING *;

-- name: UpdateRestaurantOrder :one
UPDATE restaurant_orders
SET
    table_id = $2,
    cashier_id = $3,
    cashier_session_id = $4,
    customer_id = $5,
    status = $6,
    subtotal = $7,
    discount_amount = $8,
    tax_amount = $9,
    total_amount = $10,
    amount_paid = $11,
    change_given = $12,
    notes = $13,
    pos_transaction_id = $14,
    confirmed_at = $15,
    served_at = $16,
    paid_at = $17,
    metadata = $18,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateRestaurantOrderStatus :one
UPDATE restaurant_orders
SET
    status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteRestaurantOrder :exec
DELETE FROM restaurant_orders
WHERE id = $1;

-- name: GetRestaurantOrderBySourceCartID :one
SELECT * FROM restaurant_orders
WHERE source_cart_id = $1 LIMIT 1;

-- name: CreateRestaurantOrderFull :one
INSERT INTO restaurant_orders (
    organization_id,
    store_id,
    table_id,
    cashier_id,
    cashier_session_id,
    customer_id,
    pos_terminal_id,
    kiosk_session_id,
    created_by_user_id,
    source_cart_id,
    order_number,
    order_source,
    order_type,
    status,
    payment_status,
    customer_name,
    customer_email,
    customer_phone,
    delivery_address,
    subtotal,
    discount_amount,
    tax_amount,
    total_amount,
    amount_paid,
    change_given,
    notes,
    pos_transaction_id,
    metadata,
    ordered_at,
    coupon_code,
    discount_code
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31
)
RETURNING *;

-- name: UpdateRestaurantOrderFull :one
UPDATE restaurant_orders
SET
    table_id = $2,
    cashier_id = $3,
    cashier_session_id = $4,
    customer_id = $5,
    order_type = COALESCE($6, order_type),
    status = $7,
    payment_status = COALESCE($8, payment_status),
    customer_name = $9,
    customer_email = $10,
    customer_phone = $11,
    delivery_address = $12,
    subtotal = $13,
    discount_amount = $14,
    tax_amount = $15,
    total_amount = $16,
    amount_paid = $17,
    change_given = $18,
    notes = $19,
    pos_transaction_id = $20,
    confirmed_at = $21,
    served_at = $22,
    paid_at = $23,
    metadata = $24,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateRestaurantOrderPaymentStatus :one
UPDATE restaurant_orders
SET
    payment_status = $2,
    amount_paid = COALESCE($3, amount_paid),
    change_given = COALESCE($4, change_given),
    paid_at = CASE WHEN $2 = 'paid' THEN COALESCE(paid_at, CURRENT_TIMESTAMP) ELSE paid_at END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: RecalculateRestaurantOrderTotals :one
UPDATE restaurant_orders ro
SET
    subtotal = totals.subtotal,
    discount_amount = totals.total_discount,
    tax_amount = totals.total_tax,
    total_amount = totals.total_amount,
    updated_at = CURRENT_TIMESTAMP
FROM (
    SELECT
        COALESCE(SUM(COALESCE(NULLIF(line_total, 0), subtotal) - COALESCE(tax_amount, 0)), 0)::numeric(15,2) AS subtotal,
        COALESCE(SUM(discount_amount), 0)::numeric(15,2) AS total_discount,
        COALESCE(SUM(tax_amount), 0)::numeric(15,2) AS total_tax,
        COALESCE(SUM(COALESCE(NULLIF(line_total, 0), subtotal)), 0)::numeric(15,2) AS total_amount
    FROM restaurant_order_items
    WHERE order_id = $1
) AS totals
WHERE ro.id = $1
RETURNING ro.*;

-- name: FindOpenCashierSessionForRestaurantStore :one
SELECT
    cs.id,
    cs.cashier_id,
    cs.pos_terminal_id
FROM cashier_sessions cs
JOIN pos_terminals pt ON pt.id = cs.pos_terminal_id
WHERE cs.status = 'open'
  AND cs.closing_time IS NULL
  AND pt.store_id = sqlc.arg(store_id)
  AND (sqlc.narg(cashier_id)::int IS NULL OR cs.cashier_id = sqlc.narg(cashier_id)::int)
ORDER BY cs.opening_time DESC
LIMIT 1;

-- name: ValidateRestaurantCashierSession :one
SELECT
    cs.id,
    cs.cashier_id,
    cs.pos_terminal_id,
    cs.status,
    cs.closing_time,
    pt.store_id
FROM cashier_sessions cs
JOIN pos_terminals pt ON pt.id = cs.pos_terminal_id
WHERE cs.id = $1;

-- name: CreatePosTransactionForRestaurantOrder :one
INSERT INTO pos_transactions (
    store_id,
    cashier_id,
    cashier_session_id,
    customer_id,
    pos_terminal_id,
    transaction_number,
    transaction_type,
    subtotal,
    discount_amount,
    tax_amount,
    total_amount,
    amount_paid,
    change_given,
    status,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
)
RETURNING *;

-- name: LinkRestaurantOrderToPosTransaction :one
UPDATE restaurant_orders
SET
    pos_transaction_id = $2,
    status = COALESCE(NULLIF(sqlc.arg(status)::text, ''), status),
    payment_status = COALESCE(NULLIF(sqlc.arg(payment_status)::text, ''), payment_status),
    amount_paid = COALESCE(sqlc.narg(amount_paid), amount_paid),
    change_given = COALESCE(sqlc.narg(change_given), change_given),
    paid_at = CASE WHEN sqlc.arg(payment_status)::text = 'paid' THEN COALESCE(paid_at, CURRENT_TIMESTAMP) ELSE paid_at END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateRestaurantOrderStatusWithTimestamps :one
UPDATE restaurant_orders
SET
    status = sqlc.arg(status)::text,
    confirmed_at = CASE WHEN sqlc.arg(status)::text IN ('confirmed', 'preparing') THEN COALESCE(confirmed_at, CURRENT_TIMESTAMP) ELSE confirmed_at END,
    served_at = CASE WHEN sqlc.arg(status)::text = 'served' THEN COALESCE(served_at, CURRENT_TIMESTAMP) ELSE served_at END,
    paid_at = CASE WHEN sqlc.arg(status)::text = 'paid' THEN COALESCE(paid_at, CURRENT_TIMESTAMP) ELSE paid_at END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: CreateRestaurantOrderStatusHistory :one
INSERT INTO restaurant_order_status_history (
    restaurant_order_id,
    store_id,
    from_status,
    to_status,
    reason,
    notes,
    changed_by_user_id,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, COALESCE(sqlc.arg(metadata)::jsonb, '{}'::jsonb)
) RETURNING *;

-- name: ListRestaurantOrderStatusHistory :many
SELECT
    h.id,
    h.restaurant_order_id,
    h.store_id,
    h.from_status,
    h.to_status,
    h.reason,
    h.notes,
    h.changed_by_user_id,
    u.username AS changed_by_username,
    COALESCE(NULLIF(TRIM(COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '')), ''), '')::text AS changed_by_name,
    h.metadata,
    h.changed_at
FROM restaurant_order_status_history h
LEFT JOIN users u ON u.id = h.changed_by_user_id
WHERE h.restaurant_order_id = $1
ORDER BY h.changed_at ASC, h.id ASC;

-- name: ListKdsRestaurantOrdersByStore :many
SELECT
    ro.id AS order_id,
    ro.order_number,
    ro.store_id,
    ro.order_source,
    ro.order_type,
    ro.status AS order_status,
    ro.payment_status,
    ro.subtotal,
    ro.discount_amount,
    ro.tax_amount,
    ro.total_amount,
    ro.notes,
    ro.ordered_at,
    ro.confirmed_at,
    ro.served_at,
    rt.id AS table_id,
    rt.table_number,
    rt.table_name,
    rt.section AS table_section,
    c.id AS cashier_id,
    COALESCE(NULLIF(TRIM(COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '')), ''), '')::text AS waiter_name,
    ro.customer_id,
    COALESCE(NULLIF(ro.customer_name, ''), cust.name, '')::text AS customer_name,
    (COALESCE(ROUND(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - ro.ordered_at)) / 60.0), 0))::int4 AS minutes_since_ordered
FROM restaurant_orders ro
LEFT JOIN restaurant_tables rt ON ro.table_id = rt.id
LEFT JOIN cashiers c           ON ro.cashier_id = c.id
LEFT JOIN users u              ON c.user_id = u.id
LEFT JOIN customers cust       ON ro.customer_id = cust.id
WHERE ro.store_id = $1
  AND ro.status <> 'paid'
  AND ro.status <> 'voided'
ORDER BY ro.ordered_at ASC, ro.id ASC;

-- name: DeductStoreInventoryStockForIngredient :many
UPDATE inventory_stock
SET
    quantity_on_hand = quantity_on_hand - sqlc.arg(deduct_qty)::numeric,
    quantity_available = GREATEST(0, quantity_available - sqlc.arg(deduct_qty)::numeric),
    updated_at = CURRENT_TIMESTAMP
WHERE id = (
    SELECT ist.id
    FROM inventory_stock ist
    WHERE ist.store_id = sqlc.arg(store_id)
      AND ist.product_id = sqlc.arg(product_id)
      AND ist.product_variant_id IS NOT DISTINCT FROM sqlc.narg(product_variant_id)::int
    ORDER BY ist.storage_location_id NULLS FIRST, ist.id ASC
    LIMIT 1
)
RETURNING *;

-- name: RestoreStoreInventoryStockForIngredient :many
UPDATE inventory_stock
SET
    quantity_on_hand = quantity_on_hand + sqlc.arg(restore_qty)::numeric,
    quantity_available = GREATEST(0, (quantity_on_hand + sqlc.arg(restore_qty)::numeric) - COALESCE(quantity_allocated, 0)),
    updated_at = CURRENT_TIMESTAMP
WHERE id = (
    SELECT ist.id
    FROM inventory_stock ist
    WHERE ist.store_id = sqlc.arg(store_id)
      AND ist.product_id = sqlc.arg(product_id)
      AND ist.product_variant_id IS NOT DISTINCT FROM sqlc.narg(product_variant_id)::int
    ORDER BY ist.storage_location_id NULLS FIRST, ist.id ASC
    LIMIT 1
)
RETURNING *;

-- name: CreateRestaurantIngredientStockMovement :one
INSERT INTO stock_movements (
    movement_type,
    reference_type,
    reference_id,
    product_id,
    product_variant_id,
    from_store_id,
    quantity,
    uom_id,
    posted_by,
    status,
    metadata
) VALUES (
    $1, 'restaurant_order', $2, $3, $4, $5, $6, $7, $8, 'completed', COALESCE(sqlc.arg(metadata)::jsonb, '{}'::jsonb)
) RETURNING *;

-- name: ListRestaurantOrderRecipeMovements :many
SELECT * FROM stock_movements
WHERE reference_type = 'restaurant_order'
  AND reference_id = $1
  AND movement_type IN ('recipe_deduction', 'recipe_reversal')
ORDER BY id ASC;

-- name: ResolveIngredientBaseQuantity :one
SELECT COALESCE(
    CASE
        WHEN sqlc.narg(uom_id)::int IS NOT NULL THEN
            fn_convert_uom_quantity(
                sqlc.arg(product_id)::int,
                (SELECT code FROM units_of_measure WHERE id = sqlc.narg(uom_id)::int LIMIT 1),
                sqlc.arg(quantity)::numeric
            )
        ELSE sqlc.arg(quantity)::numeric
    END,
    sqlc.arg(quantity)::numeric
)::numeric(15,3) AS base_quantity;
