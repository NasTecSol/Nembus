-- name: CreateInventoryAnalytics :one
INSERT INTO inventory_analytics (
    organization_id,
    store_id,
    product_id,
    category_id,
    date,
    month,
    quarter,
    year,
    opening_stock,
    closing_stock,
    average_stock,
    stock_value,
    receipts,
    issues,
    adjustments,
    stock_turnover_ratio,
    days_of_inventory,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18
) RETURNING *;

-- name: GetInventoryAnalytics :one
SELECT * FROM inventory_analytics
WHERE id = $1;

-- name: ListInventoryAnalytics :many
SELECT * FROM inventory_analytics
WHERE organization_id = $1
ORDER BY date DESC
LIMIT $2 OFFSET $3;

-- name: GetInventoryAnalyticsByDateRange :many
SELECT * FROM inventory_analytics
WHERE organization_id = $1
  AND date >= $2 AND date <= $3
ORDER BY date DESC;

-- name: GetInventoryAnalyticsByProduct :many
SELECT * FROM inventory_analytics
WHERE organization_id = $1 AND product_id = $2
  AND date >= $3 AND date <= $4
ORDER BY date DESC;

-- name: GetInventoryAnalyticsByStore :many
SELECT * FROM inventory_analytics
WHERE organization_id = $1 AND store_id = $2
  AND date >= $3 AND date <= $4
ORDER BY date DESC;

-- name: GetInventorySummaryByMonth :many
SELECT 
    year,
    month,
    SUM(closing_stock) AS total_closing_stock,
    SUM(stock_value) AS total_stock_value,
    SUM(receipts) AS total_receipts,
    SUM(issues) AS total_issues
FROM inventory_analytics
WHERE organization_id = $1
  AND date >= $2 AND date <= $3
GROUP BY year, month
ORDER BY year DESC, month DESC;

-- name: GetSlowMovingProducts :many
SELECT 
    product_id,
    AVG(stock_turnover_ratio) AS avg_turnover_ratio,
    AVG(days_of_inventory) AS avg_days_of_inventory,
    SUM(closing_stock) AS total_stock
FROM inventory_analytics
WHERE organization_id = $1
  AND date >= $2 AND date <= $3
  AND product_id IS NOT NULL
GROUP BY product_id
HAVING AVG(stock_turnover_ratio) < $4
ORDER BY avg_days_of_inventory DESC
LIMIT $5;

-- name: GetFastMovingProducts :many
SELECT 
    product_id,
    AVG(stock_turnover_ratio) AS avg_turnover_ratio,
    AVG(days_of_inventory) AS avg_days_of_inventory,
    SUM(issues) AS total_issues
FROM inventory_analytics
WHERE organization_id = $1
  AND date >= $2 AND date <= $3
  AND product_id IS NOT NULL
GROUP BY product_id
HAVING AVG(stock_turnover_ratio) > $4
ORDER BY avg_turnover_ratio DESC
LIMIT $5;

-- name: UpdateInventoryAnalytics :one
UPDATE inventory_analytics
SET 
    opening_stock = $2,
    closing_stock = $3,
    average_stock = $4,
    stock_value = $5,
    receipts = $6,
    issues = $7,
    adjustments = $8,
    stock_turnover_ratio = $9,
    days_of_inventory = $10,
    metadata = $11
WHERE id = $1
RETURNING *;

-- name: DeleteInventoryAnalytics :exec
DELETE FROM inventory_analytics
WHERE id = $1;
