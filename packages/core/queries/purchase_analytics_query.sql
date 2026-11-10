-- name: CreatePurchaseAnalytics :one
INSERT INTO purchase_analytics (
    organization_id,
    store_id,
    supplier_id,
    product_id,
    category_id,
    date,
    month,
    quarter,
    year,
    total_orders,
    total_quantity,
    total_amount,
    discounts_received,
    taxes_paid,
    net_amount,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16
) RETURNING *;

-- name: GetPurchaseAnalytics :one
SELECT * FROM purchase_analytics
WHERE id = $1;

-- name: ListPurchaseAnalytics :many
SELECT * FROM purchase_analytics
WHERE organization_id = $1
ORDER BY date DESC
LIMIT $2 OFFSET $3;

-- name: GetPurchaseAnalyticsByDateRange :many
SELECT * FROM purchase_analytics
WHERE organization_id = $1
  AND date >= $2 AND date <= $3
ORDER BY date DESC;

-- name: GetPurchaseAnalyticsBySupplier :many
SELECT * FROM purchase_analytics
WHERE organization_id = $1 AND supplier_id = $2
  AND date >= $3 AND date <= $4
ORDER BY date DESC;

-- name: GetPurchaseAnalyticsByProduct :many
SELECT * FROM purchase_analytics
WHERE organization_id = $1 AND product_id = $2
  AND date >= $3 AND date <= $4
ORDER BY date DESC;

-- name: GetPurchaseAnalyticsByStore :many
SELECT * FROM purchase_analytics
WHERE organization_id = $1 AND store_id = $2
  AND date >= $3 AND date <= $4
ORDER BY date DESC;

-- name: GetPurchaseSummaryByMonth :many
SELECT 
    year,
    month,
    SUM(total_orders) AS total_orders,
    SUM(total_quantity) AS total_quantity,
    SUM(total_amount) AS total_amount,
    SUM(net_amount) AS net_amount
FROM purchase_analytics
WHERE organization_id = $1
  AND date >= $2 AND date <= $3
GROUP BY year, month
ORDER BY year DESC, month DESC;

-- name: GetTopSuppliersByPurchaseAmount :many
SELECT 
    supplier_id,
    SUM(net_amount) AS total_purchase_amount,
    SUM(total_orders) AS total_orders
FROM purchase_analytics
WHERE organization_id = $1
  AND date >= $2 AND date <= $3
  AND supplier_id IS NOT NULL
GROUP BY supplier_id
ORDER BY total_purchase_amount DESC
LIMIT $4;

-- name: UpdatePurchaseAnalytics :one
UPDATE purchase_analytics
SET 
    total_orders = $2,
    total_quantity = $3,
    total_amount = $4,
    discounts_received = $5,
    taxes_paid = $6,
    net_amount = $7,
    metadata = $8
WHERE id = $1
RETURNING *;

-- name: DeletePurchaseAnalytics :exec
DELETE FROM purchase_analytics
WHERE id = $1;
