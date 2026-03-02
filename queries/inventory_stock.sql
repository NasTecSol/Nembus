-- name: GetLowStockProducts :many
SELECT 
    s.product_id,
    p.sku,
    p.name,
    s.store_id,
    st.name AS store_name,
    s.quantity_available,
    s.reorder_level,
    s.reorder_quantity,
    s.max_stock_level
FROM inventory_stock s
JOIN products p ON s.product_id = p.id
JOIN stores st ON s.store_id = st.id
WHERE s.quantity_available <= s.reorder_level
  AND s.quantity_available > 0
  AND p.is_active = true
  AND p.track_inventory = true
ORDER BY s.quantity_available ASC, p.name
LIMIT 50;

-- name: GetStockValuationByStore :many
SELECT 
    s.store_id,
    st.name AS store_name,
    COUNT(DISTINCT s.product_id) AS unique_products,
    SUM(s.quantity_on_hand * COALESCE((
        SELECT AVG(pp.price)
        FROM product_prices pp
        WHERE pp.product_id = s.product_id
          AND pp.is_active = true
    ), 0)) AS total_stock_value
FROM inventory_stock s
JOIN stores st ON s.store_id = st.id
WHERE st.organization_id = sqlc.arg('org_id')
GROUP BY s.store_id, st.name
ORDER BY total_stock_value DESC;

-- =====================================================
-- INVENTORY STOCK CRUD OPERATIONS
-- =====================================================

-- name: CreateInventoryStock :one
INSERT INTO inventory_stock (
    product_id,
    product_variant_id,
    store_id,
    storage_location_id,
    quantity_on_hand,
    quantity_allocated,
    quantity_available,
    quantity_on_order,
    quantity_in_transit,
    reorder_level,
    reorder_quantity,
    max_stock_level,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetInventoryStock :one
SELECT * FROM inventory_stock
WHERE id = $1;

-- name: GetInventoryStockByProductAndStore :one
SELECT * FROM inventory_stock
WHERE product_id = $1
  AND product_variant_id IS NOT DISTINCT FROM $2
  AND store_id = $3
  AND storage_location_id IS NOT DISTINCT FROM $4
LIMIT 1;

-- name: ListInventoryStock :many
SELECT * FROM inventory_stock
ORDER BY created_at DESC;

-- name: ListInventoryStockByStore :many
SELECT * FROM inventory_stock
WHERE store_id = $1
ORDER BY product_id;

-- name: ListInventoryStockByProduct :many
SELECT * FROM inventory_stock
WHERE product_id = $1
ORDER BY store_id;

-- name: ListInventoryStockByStorageLocation :many
SELECT * FROM inventory_stock
WHERE storage_location_id = $1
ORDER BY product_id;

-- name: ListInventoryStockByStoreAndLocation :many
SELECT * FROM inventory_stock
WHERE store_id = $1
  AND storage_location_id IS NOT DISTINCT FROM $2
ORDER BY product_id;

-- name: UpdateInventoryStock :one
UPDATE inventory_stock
SET 
    quantity_on_hand = COALESCE($2, quantity_on_hand),
    quantity_allocated = COALESCE($3, quantity_allocated),
    quantity_available = COALESCE($4, quantity_available),
    quantity_on_order = COALESCE($5, quantity_on_order),
    quantity_in_transit = COALESCE($6, quantity_in_transit),
    reorder_level = COALESCE($7, reorder_level),
    reorder_quantity = COALESCE($8, reorder_quantity),
    max_stock_level = COALESCE($9, max_stock_level),
    last_counted_at = COALESCE($10, last_counted_at),
    metadata = COALESCE($11, metadata)
WHERE id = $1
RETURNING *;

-- Note: UpsertInventoryStock should be handled in usecase layer by checking GetInventoryStockByProductAndStore first

-- name: AdjustInventoryStock :one
UPDATE inventory_stock
SET 
    quantity_on_hand = quantity_on_hand + COALESCE($2, 0),
    quantity_available = quantity_available + COALESCE($3, 0),
    quantity_allocated = quantity_allocated + COALESCE($4, 0),
    quantity_on_order = quantity_on_order + COALESCE($5, 0),
    quantity_in_transit = quantity_in_transit + COALESCE($6, 0)
WHERE id = $1
RETURNING *;

-- name: AdjustInventoryStockByProductAndStore :one
UPDATE inventory_stock
SET 
    quantity_on_hand = quantity_on_hand + COALESCE($4, 0),
    quantity_available = quantity_available + COALESCE($5, 0),
    quantity_allocated = quantity_allocated + COALESCE($6, 0),
    quantity_on_order = quantity_on_order + COALESCE($7, 0),
    quantity_in_transit = quantity_in_transit + COALESCE($8, 0)
WHERE product_id = $1
  AND product_variant_id IS NOT DISTINCT FROM $2
  AND store_id = $3
RETURNING *;

-- name: DeleteInventoryStock :exec
DELETE FROM inventory_stock
WHERE id = $1;

-- name: GetInventoryStockSummary :one
SELECT 
    COUNT(*) AS total_records,
    COUNT(DISTINCT product_id) AS unique_products,
    COUNT(DISTINCT store_id) AS unique_stores,
    SUM(quantity_on_hand) AS total_on_hand,
    SUM(quantity_available) AS total_available,
    SUM(quantity_allocated) AS total_allocated
FROM inventory_stock
WHERE store_id = $1;