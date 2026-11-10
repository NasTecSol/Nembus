-- =====================================================
-- USERS
-- =====================================================

-- name: CreateUser :one
INSERT INTO users (
    organization_id, username, email, password_hash,
    first_name, last_name, employee_code, is_active, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByEmployeeCode :one
SELECT * FROM users 
WHERE organization_id = $1 AND employee_code = $2 
LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
WHERE organization_id = $1
  AND is_active = COALESCE(sqlc.narg(is_active), is_active)
ORDER BY first_name, last_name
LIMIT $2 OFFSET $3;

-- name: SearchUsers :many
SELECT * FROM users
WHERE organization_id = $1
  AND is_active = true
  AND (
    username ILIKE '%' || $2 || '%' OR
    email ILIKE '%' || $2 || '%' OR
    first_name ILIKE '%' || $2 || '%' OR
    last_name ILIKE '%' || $2 || '%' OR
    employee_code ILIKE '%' || $2 || '%'
  )
ORDER BY first_name, last_name
LIMIT $3 OFFSET $4;

-- name: UpdateUser :one
UPDATE users
SET 
    email = COALESCE(sqlc.narg(email), email),
    first_name = COALESCE(sqlc.narg(first_name), first_name),
    last_name = COALESCE(sqlc.narg(last_name), last_name),
    employee_code = COALESCE(sqlc.narg(employee_code), employee_code),
    is_active = COALESCE(sqlc.narg(is_active), is_active),
    metadata = COALESCE(sqlc.narg(metadata), metadata)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users
SET password_hash = $2
WHERE id = $1
RETURNING id, username, email;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: CountUsers :one
SELECT COUNT(*) FROM users
WHERE organization_id = $1
  AND is_active = COALESCE(sqlc.narg(is_active), is_active);

-- =====================================================
-- USER ROLES
-- =====================================================

-- name: AssignRoleToUser :one
INSERT INTO user_roles (
    user_id, role_id, metadata
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetUserRoles :many
SELECT r.* FROM roles r
INNER JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1
ORDER BY r.name;

-- name: GetUsersWithRole :many
SELECT u.* FROM users u
INNER JOIN user_roles ur ON u.id = ur.user_id
WHERE ur.role_id = $1
  AND u.is_active = true
ORDER BY u.first_name, u.last_name;

-- name: RevokeRoleFromUser :exec
DELETE FROM user_roles 
WHERE user_id = $1 AND role_id = $2;

-- name: CheckUserHasRole :one
SELECT EXISTS(
    SELECT 1 FROM user_roles 
    WHERE user_id = $1 AND role_id = $2
) as has_role;

-- name: RevokeAllRolesFromUser :exec
DELETE FROM user_roles WHERE user_id = $1;

-- =====================================================
-- USER STORE ACCESS
-- =====================================================

-- name: GrantStoreAccessToUser :one
INSERT INTO user_store_access (
    user_id, store_id, is_primary, metadata
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetUserStores :many
SELECT s.* FROM stores s
INNER JOIN user_store_access usa ON s.id = usa.store_id
WHERE usa.user_id = $1
ORDER BY usa.is_primary DESC, s.name;

-- name: GetUserPrimaryStore :one
SELECT s.* FROM stores s
INNER JOIN user_store_access usa ON s.id = usa.store_id
WHERE usa.user_id = $1 AND usa.is_primary = true
LIMIT 1;

-- name: GetStoreUsers :many
SELECT u.* FROM users u
INNER JOIN user_store_access usa ON u.id = usa.user_id
WHERE usa.store_id = $1
  AND u.is_active = true
ORDER BY u.first_name, u.last_name;

-- name: RevokeStoreAccessFromUser :exec
DELETE FROM user_store_access 
WHERE user_id = $1 AND store_id = $2;

-- name: CheckUserHasStoreAccess :one
SELECT EXISTS(
    SELECT 1 FROM user_store_access 
    WHERE user_id = $1 AND store_id = $2
) as has_access;

-- name: SetUserPrimaryStore :exec
-- First, unset all primary flags for the user
UPDATE user_store_access
SET is_primary = false
WHERE user_id = $1;

-- name: UpdateUserStoreAccess :one
UPDATE user_store_access
SET is_primary = $3, metadata = COALESCE(sqlc.narg(metadata), metadata)
WHERE user_id = $1 AND store_id = $2
RETURNING *;

-- name: RevokeAllStoreAccessFromUser :exec
DELETE FROM user_store_access WHERE user_id = $1;

-- =====================================================
-- USER DETAILS WITH ROLES AND STORES
-- =====================================================

-- name: GetUserWithDetails :one
SELECT 
    u.*,
    json_agg(DISTINCT jsonb_build_object(
        'id', r.id,
        'name', r.name,
        'code', r.code
    )) FILTER (WHERE r.id IS NOT NULL) as roles,
    json_agg(DISTINCT jsonb_build_object(
        'id', s.id,
        'name', s.name,
        'code', s.code,
        'is_primary', usa.is_primary
    )) FILTER (WHERE s.id IS NOT NULL) as stores
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
LEFT JOIN user_store_access usa ON u.id = usa.user_id
LEFT JOIN stores s ON usa.store_id = s.id
WHERE u.id = $1
GROUP BY u.id;

-- name: ListUsersWithDetails :many
SELECT 
    u.*,
    json_agg(DISTINCT jsonb_build_object(
        'id', r.id,
        'name', r.name,
        'code', r.code
    )) FILTER (WHERE r.id IS NOT NULL) as roles,
    json_agg(DISTINCT jsonb_build_object(
        'id', s.id,
        'name', s.name,
        'code', s.code,
        'is_primary', usa.is_primary
    )) FILTER (WHERE s.id IS NOT NULL) as stores
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
LEFT JOIN user_store_access usa ON u.id = usa.user_id
LEFT JOIN stores s ON usa.store_id = s.id
WHERE u.organization_id = $1
  AND u.is_active = COALESCE(sqlc.narg(is_active), u.is_active)
GROUP BY u.id
ORDER BY u.first_name, u.last_name
LIMIT $2 OFFSET $3;