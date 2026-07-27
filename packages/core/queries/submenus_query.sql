-- name: CreateSubmenu :one
INSERT INTO submenus (
    menu_id,
    parent_submenu_id,
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

-- name: GetSubmenu :one
SELECT * FROM submenus
WHERE id = $1;

-- name: GetSubmenuByCode :one
SELECT * FROM submenus
WHERE menu_id = $1 AND code = $2;

-- name: ListSubmenus :many
SELECT * FROM submenus
ORDER BY display_order, id;

-- name: ListSubmenusByMenu :many
SELECT * FROM submenus
WHERE menu_id = $1
ORDER BY display_order, id;

-- name: ListActiveSubmenusByMenu :many
SELECT * FROM submenus
WHERE menu_id = $1 AND is_active = true
ORDER BY display_order, id;

-- name: ListSubmenusByParent :many
SELECT * FROM submenus
WHERE parent_submenu_id = $1
ORDER BY display_order, id;

-- name: UpdateSubmenu :one
UPDATE submenus
SET 
    parent_submenu_id = $2,
    name = $3,
    route_path = $4,
    icon = $5,
    display_order = $6,
    is_active = $7,
    metadata = $8
WHERE id = $1
RETURNING *;

-- name: DeleteSubmenu :exec
DELETE FROM submenus
WHERE id = $1;

-- name: ToggleSubmenuActive :one
UPDATE submenus
SET is_active = $2
WHERE id = $1
RETURNING *;
