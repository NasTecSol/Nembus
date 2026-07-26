-- name: CreateRole :one
INSERT INTO roles (
    name,
    code,
    description,
    is_system_role,
    is_active,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetRole :one
SELECT * FROM roles
WHERE id = $1;

-- name: GetRoleByCode :one
SELECT * FROM roles
WHERE code = $1;

-- name: ListRoles :many
SELECT * FROM roles
ORDER BY name;

-- name: ListActiveRoles :many
SELECT * FROM roles
WHERE is_active = true
ORDER BY name;

-- name: ListNonSystemRoles :many
SELECT * FROM roles
WHERE is_system_role = false
ORDER BY name;

-- name: UpdateRole :one
UPDATE roles
SET 
    name = $2,
    description = $3,
    is_active = $4,
    metadata = $5
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = $1 AND is_system_role = false;

-- name: ToggleRoleActive :one
UPDATE roles
SET is_active = $2
WHERE id = $1
RETURNING *;

-- name: AssignPermissionToRole :one
INSERT INTO role_permissions (
    role_id,
    permission_id,
    scope,
    metadata
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: RemovePermissionFromRole :exec
DELETE FROM role_permissions
WHERE role_id = $1 AND permission_id = $2;

-- name: GetRolePermissions :many
SELECT rp.*, p.name, p.code, p.description
FROM role_permissions rp
JOIN permissions p ON rp.permission_id = p.id
WHERE rp.role_id = $1;

-- name: CheckRoleHasPermission :one
SELECT EXISTS(
    SELECT 1 FROM role_permissions
    WHERE role_id = $1 AND permission_id = $2
) AS has_permission;
