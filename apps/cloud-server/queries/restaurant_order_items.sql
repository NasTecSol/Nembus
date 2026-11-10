-- restaurant_order_items.sql

-- name: GetRestaurantOrderItem :one
SELECT * FROM restaurant_order_items
WHERE id = $1 LIMIT 1;

-- name: ListRestaurantOrderItems :many
SELECT * FROM restaurant_order_items
WHERE order_id = $1
ORDER BY line_number;

-- name: CreateRestaurantOrderItem :one
INSERT INTO restaurant_order_items (
    order_id, menu_item_id, quantity, unit_price, modifiers_snapshot, modifiers_total, discount_amount, tax_amount, subtotal, line_number, notes, status, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: UpdateRestaurantOrderItem :one
UPDATE restaurant_order_items
SET
    menu_item_id = $2,
    quantity = $3,
    unit_price = $4,
    modifiers_snapshot = $5,
    modifiers_total = $6,
    discount_amount = $7,
    tax_amount = $8,
    subtotal = $9,
    line_number = $10,
    notes = $11,
    status = $12,
    metadata = $13,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateRestaurantOrderItemStatus :one
UPDATE restaurant_order_items
SET
    status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteRestaurantOrderItem :exec
DELETE FROM restaurant_order_items
WHERE id = $1;
