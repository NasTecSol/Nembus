-- =====================================================
-- STOCK COUNTS QUERIES
-- =====================================================

-- name: CreateStockCount :one
INSERT INTO stock_counts (
    count_number,
    store_id,
    storage_location_id,
    count_type,
    status,
    scheduled_date,
    counted_by,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetStockCount :one
SELECT * FROM stock_counts
WHERE id = $1;

-- name: GetStockCountWithDetails :one
SELECT 
    sc.*,
    st.name AS store_name,
    sl.name AS storage_location_name,
    u_counted.username AS counted_by_name,
    u_approved.username AS approved_by_name
FROM stock_counts sc
JOIN stores st ON sc.store_id = st.id
LEFT JOIN storage_locations sl ON sc.storage_location_id = sl.id
LEFT JOIN users u_counted ON sc.counted_by = u_counted.id
LEFT JOIN users u_approved ON sc.approved_by = u_approved.id
WHERE sc.id = $1;

-- name: GetStockCountByNumber :one
SELECT * FROM stock_counts
WHERE count_number = $1;

-- name: ListStockCounts :many
SELECT * FROM stock_counts
ORDER BY created_at DESC;

-- name: ListStockCountsWithDetails :many
SELECT 
    sc.id,
    sc.count_number,
    sc.store_id,
    sc.storage_location_id,
    sc.count_type,
    sc.status,
    sc.scheduled_date,
    sc.started_at,
    sc.completed_at,
    sc.counted_by,
    sc.approved_by,
    sc.metadata,
    sc.created_at,
    sc.updated_at,
    st.name AS store_name,
    sl.name AS storage_location_name,
    u_counted.username AS counted_by_name,
    u_approved.username AS approved_by_name,
    COUNT(scl.id)::bigint AS total_lines,
    COALESCE(SUM(CASE WHEN scl.variance != 0 THEN 1 ELSE 0 END), 0)::bigint AS lines_with_variance,
    COALESCE(SUM(scl.variance_value), 0)::numeric AS total_variance_value
FROM stock_counts sc
JOIN stores st ON sc.store_id = st.id
LEFT JOIN storage_locations sl ON sc.storage_location_id = sl.id
LEFT JOIN users u_counted ON sc.counted_by = u_counted.id
LEFT JOIN users u_approved ON sc.approved_by = u_approved.id
LEFT JOIN stock_count_lines scl ON sc.id = scl.stock_count_id
GROUP BY sc.id, st.name, sl.name, u_counted.username, u_approved.username
ORDER BY sc.created_at DESC;

-- name: ListStockCountsByStore :many
SELECT * FROM stock_counts
WHERE store_id = $1
ORDER BY created_at DESC;

-- name: ListStockCountsByStatus :many
SELECT * FROM stock_counts
WHERE status = $1
ORDER BY created_at DESC;

-- name: UpdateStockCount :one
UPDATE stock_counts
SET 
    storage_location_id = COALESCE($2, storage_location_id),
    count_type = COALESCE($3, count_type),
    status = COALESCE($4, status),
    scheduled_date = COALESCE($5, scheduled_date),
    counted_by = COALESCE($6, counted_by),
    metadata = COALESCE($7, metadata),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: StartStockCount :one
UPDATE stock_counts
SET 
    status = 'in_progress',
    started_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: CompleteStockCount :one
UPDATE stock_counts
SET 
    status = 'completed',
    completed_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ApproveStockCount :one
UPDATE stock_counts
SET 
    status = 'approved',
    approved_by = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ReconcileStockCountStatus :one
UPDATE stock_counts
SET 
    status = 'reconciled',
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteStockCount :exec
DELETE FROM stock_counts
WHERE id = $1;

-- name: CreateStockCountLine :one
INSERT INTO stock_count_lines (
    stock_count_id,
    product_id,
    product_variant_id,
    storage_location_id,
    expected_quantity,
    system_quantity,
    counted_quantity,
    variance,
    variance_value,
    counted_at,
    uom_id,
    batch_number,
    serial_number,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: GetStockCountLine :one
SELECT * FROM stock_count_lines
WHERE id = $1;

-- name: ListStockCountLines :many
SELECT * FROM stock_count_lines
WHERE stock_count_id = $1
ORDER BY id;

-- name: ListStockCountLinesWithDetails :many
SELECT 
    scl.*,
    p.name AS product_name,
    p.sku AS product_sku,
    pv.variant_name AS variant_name,
    pv.variant_sku AS variant_sku,
    sl.name AS storage_location_name,
    uom.name AS uom_name
FROM stock_count_lines scl
JOIN products p ON scl.product_id = p.id
LEFT JOIN product_variants pv ON scl.product_variant_id = pv.id
LEFT JOIN storage_locations sl ON scl.storage_location_id = sl.id
LEFT JOIN units_of_measure uom ON scl.uom_id = uom.id
WHERE scl.stock_count_id = $1
ORDER BY scl.id;

-- name: UpdateStockCountLine :one
UPDATE stock_count_lines
SET 
    counted_quantity = $2,
    variance = $3,
    variance_value = $4,
    counted_at = $5,
    batch_number = $6,
    serial_number = $7,
    metadata = COALESCE($8, metadata),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteStockCountLine :exec
DELETE FROM stock_count_lines
WHERE id = $1;

-- name: DeleteStockCountLinesByCountID :exec
DELETE FROM stock_count_lines
WHERE stock_count_id = $1;

-- name: GetStockCountSummary :one
SELECT 
    COUNT(*)::bigint AS total_lines,
    COALESCE(SUM(CASE WHEN variance != 0 THEN 1 ELSE 0 END), 0)::bigint AS lines_with_variance,
    COALESCE(SUM(variance_value), 0)::numeric AS total_variance_value,
    COALESCE(SUM(CASE WHEN variance > 0 THEN variance_value ELSE 0 END), 0)::numeric AS positive_variance,
    COALESCE(SUM(CASE WHEN variance < 0 THEN variance_value ELSE 0 END), 0)::numeric AS negative_variance
FROM stock_count_lines
WHERE stock_count_id = $1;
