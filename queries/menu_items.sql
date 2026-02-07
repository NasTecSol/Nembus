-- menu_items.sql

-- name: GetMenuItem :one
SELECT * FROM menu_items
WHERE id = $1 LIMIT 1;

-- name: ListMenuItems :many
SELECT * FROM menu_items
WHERE menu_category_id = $1
ORDER BY display_order, name;

-- name: ListMenuItemsByStore :many
SELECT * FROM menu_items
WHERE store_id = $1
ORDER BY display_order, name;

-- name: CreateMenuItem :one
INSERT INTO menu_items (
    store_id, menu_category_id, product_id, recipe_id, name, short_name, description, image_url, base_price, cost_price, preparation_time_min, tax_category_id, is_available, is_active, display_order, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
)
RETURNING *;

-- name: UpdateMenuItem :one
UPDATE menu_items
SET
    menu_category_id = $2,
    product_id = $3,
    recipe_id = $4,
    name = $5,
    short_name = $6,
    description = $7,
    image_url = $8,
    base_price = $9,
    cost_price = $10,
    preparation_time_min = $11,
    tax_category_id = $12,
    is_available = $13,
    is_active = $14,
    display_order = $15,
    metadata = $16,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteMenuItem :exec
DELETE FROM menu_items
WHERE id = $1;
