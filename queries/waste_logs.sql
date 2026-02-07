-- waste_logs.sql

-- name: GetWasteLog :one
SELECT * FROM waste_logs
WHERE id = $1 LIMIT 1;

-- name: ListWasteLogs :many
SELECT * FROM waste_logs
WHERE store_id = $1
ORDER BY wasted_at DESC;

-- name: CreateWasteLog :one
INSERT INTO waste_logs (
    store_id, product_id, menu_item_id, recipe_id, waste_source, quantity, uom_id, unit_cost, total_cost, reason, logged_by, order_id, wasted_at, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
RETURNING *;

-- name: UpdateWasteLog :one
UPDATE waste_logs
SET
    product_id = $2,
    menu_item_id = $3,
    recipe_id = $4,
    waste_source = $5,
    quantity = $6,
    uom_id = $7,
    unit_cost = $8,
    total_cost = $9,
    reason = $10,
    logged_by = $11,
    order_id = $12,
    wasted_at = $13,
    metadata = $14
WHERE id = $1
RETURNING *;

-- name: DeleteWasteLog :exec
DELETE FROM waste_logs
WHERE id = $1;
