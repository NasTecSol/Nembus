-- name: CreateDiscountAnalytics :one
INSERT INTO discount_analytics (
    organization_id,
    store_id,
    cashier_id,
    product_id,
    discount_type,
    date,
    month,
    quarter,
    year,
    total_discounts_given,
    transactions_with_discount,
    total_transactions,
    discount_percentage,
    revenue_impact,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15
) RETURNING *;

-- name: GetDiscountAnalytics :one
SELECT * FROM discount_analytics
WHERE id = $1;

-- name: ListDiscountAnalytics :many
SELECT * FROM discount_analytics
WHERE organization_id = $1
ORDER BY date DESC
LIMIT $2 OFFSET $3;

-- name: GetDiscountAnalyticsByDateRange :many
SELECT * FROM discount_analytics
WHERE organization_id = $1
  AND date >= $2 AND date <= $3
ORDER BY date DESC;

-- name: GetDiscountAnalyticsByStore :many
SELECT * FROM discount_analytics
WHERE organization_id = $1 AND store_id = $2
  AND date >= $3 AND date <= $4
ORDER BY date DESC;

-- name: GetDiscountAnalyticsByCashier :many
SELECT * FROM discount_analytics
WHERE organization_id = $1 AND cashier_id = $2
  AND date >= $3 AND date <= $4
ORDER BY date DESC;

-- name: GetDiscountAnalyticsByProduct :many
SELECT * FROM discount_analytics
WHERE organization_id = $1 AND product_id = $2
  AND date >= $3 AND date <= $4
ORDER BY date DESC;

-- name: GetDiscountAnalyticsByType :many
SELECT * FROM discount_analytics
WHERE organization_id = $1 AND discount_type = $2
  AND date >= $3 AND date <= $4
ORDER BY date DESC;

-- name: GetDiscountSummaryByMonth :many
SELECT 
    year,
    month,
    SUM(total_discounts_given) AS total_discounts,
    SUM(transactions_with_discount) AS total_transactions_with_discount,
    SUM(total_transactions) AS total_transactions,
    AVG(discount_percentage) AS avg_discount_percentage,
    SUM(revenue_impact) AS total_revenue_impact
FROM discount_analytics
WHERE organization_id = $1
  AND date >= $2 AND date <= $3
GROUP BY year, month
ORDER BY year DESC, month DESC;

-- name: GetTopDiscountingCashiers :many
SELECT 
    cashier_id,
    SUM(total_discounts_given) AS total_discounts,
    SUM(transactions_with_discount) AS total_transactions,
    AVG(discount_percentage) AS avg_discount_percentage
FROM discount_analytics
WHERE organization_id = $1
  AND date >= $2 AND date <= $3
  AND cashier_id IS NOT NULL
GROUP BY cashier_id
ORDER BY total_discounts DESC
LIMIT $4;

-- name: GetDiscountSummaryByType :many
SELECT 
    discount_type,
    SUM(total_discounts_given) AS total_discounts,
    SUM(transactions_with_discount) AS total_transactions,
    SUM(revenue_impact) AS total_revenue_impact
FROM discount_analytics
WHERE organization_id = $1
  AND date >= $2 AND date <= $3
GROUP BY discount_type
ORDER BY total_discounts DESC;

-- name: UpdateDiscountAnalytics :one
UPDATE discount_analytics
SET 
    total_discounts_given = $2,
    transactions_with_discount = $3,
    total_transactions = $4,
    discount_percentage = $5,
    revenue_impact = $6,
    metadata = $7
WHERE id = $1
RETURNING *;

-- name: DeleteDiscountAnalytics :exec
DELETE FROM discount_analytics
WHERE id = $1;
