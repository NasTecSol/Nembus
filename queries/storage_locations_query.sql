-- name: CreateStorageLocation :one
INSERT INTO storage_locations (
    store_id,
    code,
    name,
    location_type,
    parent_location_id,
    is_active,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetStorageLocation :one
SELECT * FROM storage_locations
WHERE id = $1;

-- name: GetStorageLocationByCode :one
SELECT * FROM storage_locations
WHERE store_id = $1 AND code = $2;

-- name: ListStorageLocations :many
SELECT * FROM storage_locations
ORDER BY code;

-- name: ListStorageLocationsByStore :many
SELECT * FROM storage_locations
WHERE store_id = $1
ORDER BY code;

-- name: ListActiveStorageLocationsByStore :many
SELECT * FROM storage_locations
WHERE store_id = $1 AND is_active = true
ORDER BY code;

-- name: ListStorageLocationsByParent :many
SELECT * FROM storage_locations
WHERE parent_location_id = $1
ORDER BY code;

-- name: ListStorageLocationsByType :many
SELECT * FROM storage_locations
WHERE store_id = $1 AND location_type = $2
ORDER BY code;

-- name: UpdateStorageLocation :one
UPDATE storage_locations
SET 
    name = $2,
    location_type = $3,
    parent_location_id = $4,
    is_active = $5,
    metadata = $6
WHERE id = $1
RETURNING *;

-- name: DeleteStorageLocation :exec
DELETE FROM storage_locations
WHERE id = $1;

-- name: ToggleStorageLocationActive :one
UPDATE storage_locations
SET is_active = $2
WHERE id = $1
RETURNING *;
