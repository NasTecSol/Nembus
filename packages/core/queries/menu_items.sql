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
    store_id, menu_category_id, product_id, product_variant_id, recipe_id, name, short_name, description, image_url, base_price, cost_price, preparation_time_min, tax_category_id, is_available, is_active, display_order, item_type, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
)
RETURNING *;

-- name: UpdateMenuItem :one
UPDATE menu_items
SET
    menu_category_id = $2,
    product_id = $3,
    product_variant_id = $4,
    recipe_id = $5,
    name = $6,
    short_name = $7,
    description = $8,
    image_url = $9,
    base_price = $10,
    cost_price = $11,
    preparation_time_min = $12,
    tax_category_id = $13,
    is_available = $14,
    is_active = $15,
    display_order = $16,
    item_type = $17,
    metadata = $18,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteMenuItem :exec
DELETE FROM menu_items
WHERE id = $1;
