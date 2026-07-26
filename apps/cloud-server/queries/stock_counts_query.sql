-- name: CreateStockCount :one
INSERT INTO stock_counts (
    count_number,
    store_id,
    count_type,
    status,
    scheduled_date,
    counted_by,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetStockCount :one
SELECT * FROM stock_counts
WHERE id = $1;

-- name: GetStockCountByNumber :one
SELECT * FROM stock_counts
WHERE count_number = $1;

-- name: ListStockCounts :many
SELECT * FROM stock_counts
ORDER BY created_at DESC;

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
    status = $2,
    metadata = $3
WHERE id = $1
RETURNING *;

-- name: StartStockCount :one
UPDATE stock_counts
SET 
    status = 'in_progress',
    started_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: CompleteStockCount :one
UPDATE stock_counts
SET 
    status = 'completed',
    completed_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ApproveStockCount :one
UPDATE stock_counts
SET 
    status = 'approved',
    approved_by = $2
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
    system_quantity,
    counted_quantity,
    variance,
    variance_value,
    batch_number,
    serial_number,
    counted_at,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetStockCountLine :one
SELECT * FROM stock_count_lines
WHERE id = $1;

-- name: ListStockCountLines :many
SELECT * FROM stock_count_lines
WHERE stock_count_id = $1
ORDER BY id;

-- name: UpdateStockCountLine :one
UPDATE stock_count_lines
SET 
    counted_quantity = $2,
    variance = $3,
    variance_value = $4,
    counted_at = $5
WHERE id = $1
RETURNING *;

-- name: DeleteStockCountLine :exec
DELETE FROM stock_count_lines
WHERE id = $1;

-- name: GetStockCountSummary :one
SELECT 
    COUNT(*) AS total_lines,
    SUM(CASE WHEN variance != 0 THEN 1 ELSE 0 END) AS lines_with_variance,
    SUM(variance_value) AS total_variance_value,
    SUM(CASE WHEN variance > 0 THEN variance_value ELSE 0 END) AS positive_variance,
    SUM(CASE WHEN variance < 0 THEN variance_value ELSE 0 END) AS negative_variance
FROM stock_count_lines
WHERE stock_count_id = $1;
