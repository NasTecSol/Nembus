-- =====================================================
-- STORES
-- =====================================================

-- name: CreateStore :one
INSERT INTO stores (
    organization_id, parent_store_id, name, code, store_type,
    is_warehouse, is_pos_enabled, timezone, is_active, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetStore :one
SELECT * FROM stores WHERE id = $1 LIMIT 1;

-- name: GetStoreByCode :one
SELECT * FROM stores 
WHERE organization_id = $1 AND code = $2 
LIMIT 1;

-- name: ListStores :many
SELECT * FROM stores
WHERE organization_id = $1
  AND is_active = COALESCE(sqlc.narg(is_active), is_active)
  AND store_type = COALESCE(sqlc.narg(store_type), store_type)
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: ListStoresByParent :many
SELECT * FROM stores
WHERE parent_store_id = $1 
  AND is_active = COALESCE(sqlc.narg(is_active), is_active)
ORDER BY name;

-- name: ListWarehouseStores :many
SELECT * FROM stores
WHERE organization_id = $1
  AND is_warehouse = true
  AND is_active = true
ORDER BY name;

-- name: ListPOSEnabledStores :many
SELECT * FROM stores
WHERE organization_id = $1
  AND is_pos_enabled = true
  AND is_active = true
ORDER BY name;

-- name: UpdateStore :one
UPDATE stores
SET 
    parent_store_id = COALESCE(sqlc.narg(parent_store_id), parent_store_id),
    name = COALESCE(sqlc.narg(name), name),
    store_type = COALESCE(sqlc.narg(store_type), store_type),
    is_warehouse = COALESCE(sqlc.narg(is_warehouse), is_warehouse),
    is_pos_enabled = COALESCE(sqlc.narg(is_pos_enabled), is_pos_enabled),
    timezone = COALESCE(sqlc.narg(timezone), timezone),
    is_active = COALESCE(sqlc.narg(is_active), is_active),
    metadata = COALESCE(sqlc.narg(metadata), metadata)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteStore :exec
DELETE FROM stores WHERE id = $1;

-- name: CountStores :one
SELECT COUNT(*) FROM stores
WHERE organization_id = $1
  AND is_active = COALESCE(sqlc.narg(is_active), is_active);

-- =====================================================
-- STORAGE LOCATIONS
-- Note: Storage location queries are in storage_locations_query.sql
-- =====================================================

-- name: GetStorageLocationHierarchy :many
WITH RECURSIVE location_tree AS (
    -- Base case: top-level locations
    SELECT 
        id, store_id, code, name, location_type,
        parent_location_id, is_active, metadata,
        1 as level,
        ARRAY[id] as path
    FROM storage_locations sl_base
    WHERE sl_base.store_id = $1 AND sl_base.parent_location_id IS NULL
    
    UNION ALL
    
    -- Recursive case: child locations
    SELECT 
        sl.id, sl.store_id, sl.code, sl.name, sl.location_type,
        sl.parent_location_id, sl.is_active, sl.metadata,
        lt.level + 1,
        lt.path || sl.id
    FROM storage_locations sl
    INNER JOIN location_tree lt ON sl.parent_location_id = lt.id
)
SELECT 
    tree_data.id, tree_data.store_id, tree_data.code, tree_data.name, tree_data.location_type,
    tree_data.parent_location_id, tree_data.is_active, tree_data.metadata,
    tree_data.level, tree_data.path
FROM (
    SELECT * FROM location_tree
) AS tree_data
WHERE CASE 
    WHEN sqlc.narg(filter_is_active) IS NULL THEN true
    ELSE tree_data.is_active = sqlc.narg(filter_is_active)
END
ORDER BY tree_data.path;