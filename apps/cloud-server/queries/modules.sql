-- =====================================================
-- MODULES
-- =====================================================

-- name: CreateModule :one
INSERT INTO modules (
    name, code, description, icon, is_active, display_order, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetModule :one
SELECT * FROM modules WHERE id = $1 LIMIT 1;

-- name: GetModuleByCode :one
SELECT * FROM modules WHERE code = $1 LIMIT 1;

-- name: ListModules :many
SELECT * FROM modules
WHERE is_active = COALESCE(sqlc.narg(is_active), is_active)
ORDER BY display_order, name;

-- name: UpdateModule :one
UPDATE modules
SET 
    name = COALESCE(sqlc.narg(name), name),
    description = COALESCE(sqlc.narg(description), description),
    icon = COALESCE(sqlc.narg(icon), icon),
    is_active = COALESCE(sqlc.narg(is_active), is_active),
    display_order = COALESCE(sqlc.narg(display_order), display_order),
    metadata = COALESCE(sqlc.narg(metadata), metadata)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteModule :exec
DELETE FROM modules WHERE id = $1;

-- =====================================================
-- MENUS
-- Note: Menu queries are in menus_query.sql
-- =====================================================

-- =====================================================
-- SUBMENUS
-- Note: Submenu queries are in submenus_query.sql
-- =====================================================

-- =====================================================
-- NAVIGATION HIERARCHY QUERIES
-- =====================================================

-- name: GetFullNavigationHierarchy :many
SELECT 
    m.id as module_id,
    m.name as module_name,
    m.code as module_code,
    m.icon as module_icon,
    m.display_order as module_order,
    mn.id as menu_id,
    mn.name as menu_name,
    mn.code as menu_code,
    mn.route_path as menu_route,
    mn.icon as menu_icon,
    mn.display_order as menu_order,
    sm.id as submenu_id,
    sm.name as submenu_name,
    sm.code as submenu_code,
    sm.route_path as submenu_route,
    sm.icon as submenu_icon,
    sm.display_order as submenu_order
FROM modules m
LEFT JOIN menus mn ON m.id = mn.module_id AND mn.is_active = true
LEFT JOIN submenus sm ON mn.id = sm.menu_id AND sm.is_active = true
WHERE m.is_active = true
ORDER BY m.display_order, mn.display_order, sm.display_order;