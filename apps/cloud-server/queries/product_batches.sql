-- =====================================================
-- PRODUCT BATCHES
-- =====================================================

-- name: CreateProductBatch :one
INSERT INTO product_batches (
    product_id, product_variant_id, batch_number,
    manufacturing_date, expiry_date, store_id,
    quantity_available, status, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetProductBatch :one
SELECT * FROM product_batches WHERE id = $1 LIMIT 1;

-- name: GetProductBatchByNumber :one
SELECT * FROM product_batches
WHERE product_id = $1
  AND batch_number = $2
  AND store_id = COALESCE(sqlc.narg(store_id), store_id)
LIMIT 1;

-- name: ListProductBatches :many
SELECT 
    pb.*,
    p.name as product_name,
    p.sku as product_sku,
    s.name as store_name,
    s.code as store_code
FROM product_batches pb
INNER JOIN products p ON pb.product_id = p.id
LEFT JOIN stores s ON pb.store_id = s.id
WHERE pb.product_id = $1
  AND pb.product_variant_id = COALESCE(sqlc.narg(product_variant_id), pb.product_variant_id)
  AND pb.store_id = COALESCE(sqlc.narg(store_id), pb.store_id)
  AND pb.status = COALESCE(sqlc.narg(status), pb.status)
ORDER BY pb.expiry_date NULLS LAST, pb.manufacturing_date DESC;

-- name: ListBatchesByStore :many
SELECT 
    pb.*,
    p.name as product_name,
    p.sku as product_sku
FROM product_batches pb
INNER JOIN products p ON pb.product_id = p.id
WHERE pb.store_id = $1
  AND pb.status = 'active'
  AND pb.quantity_available > 0
ORDER BY p.name, pb.expiry_date NULLS LAST;

-- name: GetExpiringSoonBatches :many
SELECT 
    pb.*,
    p.name as product_name,
    p.sku as product_sku,
    s.name as store_name
FROM product_batches pb
INNER JOIN products p ON pb.product_id = p.id
LEFT JOIN stores s ON pb.store_id = s.id
WHERE pb.store_id = COALESCE(sqlc.narg(store_id), pb.store_id)
  AND pb.status = 'active'
  AND pb.quantity_available > 0
  AND pb.expiry_date IS NOT NULL
  AND pb.expiry_date <= CURRENT_DATE + $1::interval
ORDER BY pb.expiry_date, p.name;

-- name: GetExpiredBatches :many
SELECT 
    pb.*,
    p.name as product_name,
    p.sku as product_sku,
    s.name as store_name
FROM product_batches pb
INNER JOIN products p ON pb.product_id = p.id
LEFT JOIN stores s ON pb.store_id = s.id
WHERE pb.store_id = COALESCE(sqlc.narg(store_id), pb.store_id)
  AND pb.status = 'active'
  AND pb.expiry_date IS NOT NULL
  AND pb.expiry_date < CURRENT_DATE
ORDER BY pb.expiry_date, p.name;

-- name: UpdateProductBatch :one
UPDATE product_batches
SET 
    quantity_available = COALESCE(sqlc.narg(quantity_available), quantity_available),
    status = COALESCE(sqlc.narg(status), status),
    metadata = COALESCE(sqlc.narg(metadata), metadata)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: AdjustBatchQuantity :one
UPDATE product_batches
SET quantity_available = quantity_available + $2
WHERE id = $1
RETURNING *;

-- name: DeleteProductBatch :exec
DELETE FROM product_batches WHERE id = $1;

-- name: ExpireBatches :exec
UPDATE product_batches
SET status = 'expired'
WHERE expiry_date < CURRENT_DATE
  AND status = 'active';

-- name: GetAvailableBatches :many
SELECT * FROM product_batches
WHERE product_id = $1
  AND product_variant_id = COALESCE(sqlc.narg(product_variant_id), product_variant_id)
  AND store_id = $2
  AND status = 'active'
  AND quantity_available > 0
ORDER BY expiry_date NULLS LAST, manufacturing_date;

-- =====================================================
-- PRODUCT SERIAL NUMBERS
-- Note: Product serial number queries are in product_serial_numbers_query.sql
-- =====================================================

-- name: GetSerialNumberHistory :many
SELECT 
    sm.*,
    p.name as product_name,
    p.sku as product_sku,
    from_store.name as from_store_name,
    to_store.name as to_store_name
FROM stock_movements sm
INNER JOIN products p ON sm.product_id = p.id
LEFT JOIN stores from_store ON sm.from_store_id = from_store.id
LEFT JOIN stores to_store ON sm.to_store_id = to_store.id
WHERE sm.serial_number = $1
ORDER BY sm.movement_date DESC;

-- =====================================================
-- BATCH AND SERIAL NUMBER REPORTS
-- =====================================================

-- name: GetBatchStockSummary :many
SELECT 
    p.id as product_id,
    p.name as product_name,
    p.sku as product_sku,
    pb.batch_number,
    pb.expiry_date,
    s.name as store_name,
    pb.quantity_available,
    pb.status
FROM product_batches pb
INNER JOIN products p ON pb.product_id = p.id
INNER JOIN stores s ON pb.store_id = s.id
WHERE s.organization_id = $1
  AND pb.quantity_available > 0
ORDER BY pb.expiry_date NULLS LAST, p.name;

-- name: GetSerialNumberStockSummary :many
SELECT 
    p.id as product_id,
    p.name as product_name,
    p.sku as product_sku,
    s.name as store_name,
    psn.status,
    COUNT(*) as count
FROM product_serial_numbers psn
INNER JOIN products p ON psn.product_id = p.id
LEFT JOIN stores s ON psn.current_store_id = s.id
WHERE p.organization_id = $1
GROUP BY p.id, p.name, p.sku, s.name, psn.status
ORDER BY p.name, s.name;

-- name: GetExpiryCalendar :many
SELECT 
    DATE_TRUNC('month', pb.expiry_date)::date as expiry_month,
    COUNT(DISTINCT pb.id) as batch_count,
    COUNT(DISTINCT pb.product_id) as product_count,
    SUM(pb.quantity_available) as total_quantity
FROM product_batches pb
WHERE pb.store_id = COALESCE(sqlc.narg(store_id), pb.store_id)
  AND pb.status = 'active'
  AND pb.expiry_date IS NOT NULL
  AND pb.expiry_date >= CURRENT_DATE
GROUP BY expiry_month
ORDER BY expiry_month;

-- name: GetNearExpiryProducts :many
SELECT 
    p.id,
    p.sku,
    p.name,
    pb.batch_number,
    pb.expiry_date,
    pb.quantity_available,
    s.name as store_name,
    pb.expiry_date - CURRENT_DATE as days_until_expiry
FROM product_batches pb
INNER JOIN products p ON pb.product_id = p.id
INNER JOIN stores s ON pb.store_id = s.id
WHERE pb.status = 'active'
  AND pb.quantity_available > 0
  AND pb.expiry_date IS NOT NULL
  AND pb.expiry_date BETWEEN CURRENT_DATE AND CURRENT_DATE + $1::interval
  AND s.organization_id = $2
ORDER BY pb.expiry_date, p.name;