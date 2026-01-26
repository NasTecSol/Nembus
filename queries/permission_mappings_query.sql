-- name: CreateModulePermission :one
INSERT INTO module_permissions (
    module_id,
    permission_id,
    metadata
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetModulePermission :one
SELECT * FROM module_permissions
WHERE id = $1;

-- name: ListModulePermissions :many
SELECT mp.*, m.name AS module_name, p.name AS permission_name
FROM module_permissions mp
JOIN modules m ON mp.module_id = m.id
JOIN permissions p ON mp.permission_id = p.id
ORDER BY m.name, p.name;

-- name: ListModulePermissionsByModule :many
SELECT mp.*, p.name AS permission_name, p.code AS permission_code
FROM module_permissions mp
JOIN permissions p ON mp.permission_id = p.id
WHERE mp.module_id = $1
ORDER BY p.name;

-- name: DeleteModulePermission :exec
DELETE FROM module_permissions
WHERE module_id = $1 AND permission_id = $2;

-- name: CreateMenuPermission :one
INSERT INTO menu_permissions (
    menu_id,
    permission_id,
    metadata
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetMenuPermission :one
SELECT * FROM menu_permissions
WHERE id = $1;

-- name: ListMenuPermissions :many
SELECT mp.*, m.name AS menu_name, p.name AS permission_name
FROM menu_permissions mp
JOIN menus m ON mp.menu_id = m.id
JOIN permissions p ON mp.permission_id = p.id
ORDER BY m.name, p.name;

-- name: ListMenuPermissionsByMenu :many
SELECT mp.*, p.name AS permission_name, p.code AS permission_code
FROM menu_permissions mp
JOIN permissions p ON mp.permission_id = p.id
WHERE mp.menu_id = $1
ORDER BY p.name;

-- name: DeleteMenuPermission :exec
DELETE FROM menu_permissions
WHERE menu_id = $1 AND permission_id = $2;

-- name: CreateSubmenuPermission :one
INSERT INTO submenu_permissions (
    submenu_id,
    permission_id,
    metadata
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetSubmenuPermission :one
SELECT * FROM submenu_permissions
WHERE id = $1;

-- name: ListSubmenuPermissions :many
SELECT sp.*, s.name AS submenu_name, p.name AS permission_name
FROM submenu_permissions sp
JOIN submenus s ON sp.submenu_id = s.id
JOIN permissions p ON sp.permission_id = p.id
ORDER BY s.name, p.name;

-- name: ListSubmenuPermissionsBySubmenu :many
SELECT sp.*, p.name AS permission_name, p.code AS permission_code
FROM submenu_permissions sp
JOIN permissions p ON sp.permission_id = p.id
WHERE sp.submenu_id = $1
ORDER BY p.name;

-- name: DeleteSubmenuPermission :exec
DELETE FROM submenu_permissions
WHERE submenu_id = $1 AND permission_id = $2;

-- name: CheckUserHasModuleAccess :one
SELECT EXISTS(
    SELECT 1 
    FROM user_roles ur
    JOIN role_permissions rp ON ur.role_id = rp.role_id
    JOIN module_permissions mp ON rp.permission_id = mp.permission_id
    WHERE ur.user_id = $1 AND mp.module_id = $2
) AS has_access;

-- name: CheckUserHasMenuAccess :one
SELECT EXISTS(
    SELECT 1 
    FROM user_roles ur
    JOIN role_permissions rp ON ur.role_id = rp.role_id
    JOIN menu_permissions mp ON rp.permission_id = mp.permission_id
    WHERE ur.user_id = $1 AND mp.menu_id = $2
) AS has_access;

-- name: CheckUserHasSubmenuAccess :one
SELECT EXISTS(
    SELECT 1 
    FROM user_roles ur
    JOIN role_permissions rp ON ur.role_id = rp.role_id
    JOIN submenu_permissions sp ON rp.permission_id = sp.permission_id
    WHERE ur.user_id = $1 AND sp.submenu_id = $2
) AS has_access;
