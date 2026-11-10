-- menu_categories.sql

-- name: GetMenuCategory :one
SELECT * FROM menu_categories
WHERE id = $1 LIMIT 1;

-- name: ListMenuCategories :many
SELECT * FROM menu_categories
WHERE store_id = $1
ORDER BY display_order, name;

-- name: CreateMenuCategory :one
INSERT INTO menu_categories (
    store_id, parent_category_id, name, code, description, category_level, display_order, icon, image_url, is_active, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: UpdateMenuCategory :one
UPDATE menu_categories
SET
    parent_category_id = $2,
    name = $3,
    code = $4,
    description = $5,
    category_level = $6,
    display_order = $7,
    icon = $8,
    image_url = $9,
    is_active = $10,
    metadata = $11,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteMenuCategory :exec
DELETE FROM menu_categories
WHERE id = $1;
