-- =====================================================
-- PERMISSIONS
-- =====================================================

-- name: CreatePermission :one
INSERT INTO permissions (
    name, code, description, metadata
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetPermission :one
SELECT * FROM permissions WHERE id = $1 LIMIT 1;

-- name: GetPermissionByCode :one
SELECT * FROM permissions WHERE code = $1 LIMIT 1;

-- name: ListPermissions :many
SELECT * FROM permissions
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: UpdatePermission :one
UPDATE permissions
SET 
    name = COALESCE(sqlc.narg(name), name),
    description = COALESCE(sqlc.narg(description), description),
    metadata = COALESCE(sqlc.narg(metadata), metadata)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeletePermission :exec
DELETE FROM permissions WHERE id = $1;

-- =====================================================
-- ROLES
-- Note: Role queries are in roles_query.sql
-- =====================================================

-- =====================================================
-- ROLE PERMISSIONS
-- Note: AssignPermissionToRole and GetRolePermissions are in roles_query.sql
-- =====================================================

-- name: GetRolePermissionsWithScope :many
SELECT p.*, rp.scope FROM permissions p
INNER JOIN role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = $1;

-- name: RevokePermissionFromRole :exec
DELETE FROM role_permissions 
WHERE role_id = $1 AND permission_id = $2;

-- Note: CheckRoleHasPermission is in roles_query.sql

-- name: UpdateRolePermissionScope :one
UPDATE role_permissions
SET scope = $3
WHERE role_id = $1 AND permission_id = $2
RETURNING *;

-- =====================================================
-- MODULE PERMISSIONS
-- =====================================================

-- name: AssignPermissionToModule :one
INSERT INTO module_permissions (
    module_id, permission_id, metadata
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetModulePermissions :many
SELECT p.* FROM permissions p
INNER JOIN module_permissions mp ON p.id = mp.permission_id
WHERE mp.module_id = $1;

-- name: RevokePermissionFromModule :exec
DELETE FROM module_permissions 
WHERE module_id = $1 AND permission_id = $2;

-- =====================================================
-- MENU PERMISSIONS
-- =====================================================

-- name: AssignPermissionToMenu :one
INSERT INTO menu_permissions (
    menu_id, permission_id, metadata
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetMenuPermissions :many
SELECT p.* FROM permissions p
INNER JOIN menu_permissions mp ON p.id = mp.permission_id
WHERE mp.menu_id = $1;

-- name: RevokePermissionFromMenu :exec
DELETE FROM menu_permissions 
WHERE menu_id = $1 AND permission_id = $2;

-- =====================================================
-- SUBMENU PERMISSIONS
-- =====================================================

-- name: AssignPermissionToSubmenu :one
INSERT INTO submenu_permissions (
    submenu_id, permission_id, metadata
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetSubmenuPermissions :many
SELECT p.* FROM permissions p
INNER JOIN submenu_permissions sp ON p.id = sp.permission_id
WHERE sp.submenu_id = $1;

-- name: RevokePermissionFromSubmenu :exec
DELETE FROM submenu_permissions 
WHERE submenu_id = $1 AND permission_id = $2;

-- =====================================================
-- USER PERMISSIONS (via roles)
-- =====================================================

-- name: GetUserPermissions :many
SELECT DISTINCT p.* FROM permissions p
INNER JOIN role_permissions rp ON p.id = rp.permission_id
INNER JOIN user_roles ur ON rp.role_id = ur.role_id
WHERE ur.user_id = $1;

-- name: GetUserPermissionsWithScope :many
SELECT DISTINCT p.*, rp.scope FROM permissions p
INNER JOIN role_permissions rp ON p.id = rp.permission_id
INNER JOIN user_roles ur ON rp.role_id = ur.role_id
WHERE ur.user_id = $1;

-- name: CheckUserHasPermission :one
SELECT EXISTS(
    SELECT 1 FROM permissions p
    INNER JOIN role_permissions rp ON p.id = rp.permission_id
    INNER JOIN user_roles ur ON rp.role_id = ur.role_id
    WHERE ur.user_id = $1 AND p.code = $2
) as has_permission;

-- name: GetUserAccessibleModules :many
SELECT DISTINCT m.* FROM modules m
INNER JOIN module_permissions mp ON m.id = mp.module_id
INNER JOIN role_permissions rp ON mp.permission_id = rp.permission_id
INNER JOIN user_roles ur ON rp.role_id = ur.role_id
WHERE ur.user_id = $1 AND m.is_active = true
ORDER BY m.display_order;

-- name: GetUserAccessibleMenus :many
SELECT DISTINCT mn.* FROM menus mn
INNER JOIN menu_permissions mnp ON mn.id = mnp.menu_id
INNER JOIN role_permissions rp ON mnp.permission_id = rp.permission_id
INNER JOIN user_roles ur ON rp.role_id = ur.role_id
WHERE ur.user_id = $1 AND mn.is_active = true
ORDER BY mn.display_order;

-- name: GetUserAccessibleSubmenus :many
SELECT DISTINCT sm.* FROM submenus sm
INNER JOIN submenu_permissions sp ON sm.id = sp.submenu_id
INNER JOIN role_permissions rp ON sp.permission_id = rp.permission_id
INNER JOIN user_roles ur ON rp.role_id = ur.role_id
WHERE ur.user_id = $1 AND sm.is_active = true
ORDER BY sm.display_order;