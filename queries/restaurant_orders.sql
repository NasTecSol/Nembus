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
