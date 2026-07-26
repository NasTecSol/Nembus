-- menu_item_modifiers.sql

-- name: GetMenuItemModifier :one
SELECT * FROM menu_item_modifiers
WHERE id = $1 LIMIT 1;

-- name: ListMenuItemModifiers :many
SELECT * FROM menu_item_modifiers
WHERE menu_item_id = $1
ORDER BY display_order;

-- name: CreateMenuItemModifier :one
INSERT INTO menu_item_modifiers (
    menu_item_id, modifier_name, modifier_type, price_adjustment, is_active, display_order, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateMenuItemModifier :one
UPDATE menu_item_modifiers
SET
    modifier_name = $2,
    modifier_type = $3,
    price_adjustment = $4,
    is_active = $5,
    display_order = $6,
    metadata = $7
WHERE id = $1
RETURNING *;

-- name: DeleteMenuItemModifier :exec
DELETE FROM menu_item_modifiers
WHERE id = $1;
