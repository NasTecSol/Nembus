-- menu_modifier_groups.sql
-- name: GetMenuModifierGroup :one
SELECT * FROM menu_modifier_groups
WHERE id = $1 LIMIT 1;

-- name: ListMenuModifierGroupsByStore :many
SELECT * FROM menu_modifier_groups
WHERE store_id = $1
ORDER BY display_order, name;

-- name: CreateMenuModifierGroup :one
INSERT INTO menu_modifier_groups (
    store_id, name, code, selection_type, min_selections, max_selections, is_active, display_order, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: UpdateMenuModifierGroup :one
UPDATE menu_modifier_groups
SET
    name = $2,
    code = $3,
    selection_type = $4,
    min_selections = $5,
    max_selections = $6,
    is_active = $7,
    display_order = $8,
    metadata = $9
WHERE id = $1
RETURNING *;

-- name: DeleteMenuModifierGroup :exec
DELETE FROM menu_modifier_groups
WHERE id = $1;
