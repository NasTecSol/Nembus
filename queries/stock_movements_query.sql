-- name: CreateStockMovement :one
INSERT INTO stock_movements (
    movement_type,
    reference_type,
    reference_id,
    product_id,
    product_variant_id,
    from_store_id,
    to_store_id,
    from_location_id,
    to_location_id,
    quantity,
    uom_id,
    batch_number,
    serial_number,
    movement_date,
    posted_by,
    status,
    cost_per_unit,
    total_value,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19
) RETURNING *;

-- name: GetStockMovement :one
SELECT * FROM stock_movements
WHERE id = $1;

-- name: ListStockMovements :many
SELECT * FROM stock_movements
ORDER BY movement_date DESC
LIMIT $1 OFFSET $2;

-- name: ListStockMovementsByProduct :many
SELECT * FROM stock_movements
WHERE product_id = $1
ORDER BY movement_date DESC
LIMIT $2 OFFSET $3;

-- name: ListStockMovementsByStore :many
SELECT * FROM stock_movements
WHERE from_store_id = $1 OR to_store_id = $1
ORDER BY movement_date DESC
LIMIT $2 OFFSET $3;

-- name: ListStockMovementsByType :many
SELECT * FROM stock_movements
WHERE movement_type = $1
ORDER BY movement_date DESC
LIMIT $2 OFFSET $3;

-- name: ListStockMovementsByReference :many
SELECT * FROM stock_movements
WHERE reference_type = $1 AND reference_id = $2
ORDER BY movement_date DESC;

-- name: ListStockMovementsByDateRange :many
SELECT * FROM stock_movements
WHERE movement_date >= $1 AND movement_date <= $2
ORDER BY movement_date DESC;

-- name: GetStockMovementsByProductAndStore :many
SELECT * FROM stock_movements
WHERE product_id = $1 
  AND (from_store_id = $2 OR to_store_id = $2)
  AND movement_date >= $3
ORDER BY movement_date DESC;

-- name: UpdateStockMovementStatus :one
UPDATE stock_movements
SET status = $2
WHERE id = $1
RETURNING *;

-- name: GetStockMovementSummaryByProduct :one
SELECT 
    product_id,
    SUM(CASE WHEN to_store_id = $2 THEN quantity ELSE 0 END) AS total_in,
    SUM(CASE WHEN from_store_id = $2 THEN quantity ELSE 0 END) AS total_out,
    SUM(CASE WHEN to_store_id = $2 THEN total_value ELSE 0 END) AS value_in,
    SUM(CASE WHEN from_store_id = $2 THEN total_value ELSE 0 END) AS value_out
FROM stock_movements
WHERE product_id = $1
  AND movement_date >= $3 AND movement_date <= $4
GROUP BY product_id;
