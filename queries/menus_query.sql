-- name: CreateMenu :one
INSERT INTO menus (
    module_id,
    parent_menu_id,
    name,
    code,
    route_path,
    icon,
    display_order,
    is_active,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetMenu :one
SELECT * FROM menus
WHERE id = $1;

-- name: GetMenuByCode :one
SELECT * FROM menus
WHERE module_id = $1 AND code = $2;

-- name: ListMenus :many
SELECT * FROM menus
ORDER BY display_order, id;

-- name: ListMenusByModule :many
SELECT * FROM menus
WHERE module_id = $1
ORDER BY display_order, id;

-- name: ListActiveMenusByModule :many
SELECT * FROM menus
WHERE module_id = $1 AND is_active = true
ORDER BY display_order, id;

-- name: ListMenusByParent :many
SELECT * FROM menus
WHERE parent_menu_id = $1
ORDER BY display_order, id;

-- name: UpdateMenu :one
UPDATE menus
SET 
    parent_menu_id = $2,
    name = $3,
    route_path = $4,
    icon = $5,
    display_order = $6,
    is_active = $7,
    metadata = $8
WHERE id = $1
RETURNING *;

-- name: DeleteMenu :exec
DELETE FROM menus
WHERE id = $1;

-- name: ToggleMenuActive :one
UPDATE menus
SET is_active = $2
WHERE id = $1
RETURNING *;
