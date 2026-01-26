-- name: CreateProfitLossAnalytics :one
INSERT INTO profit_loss_analytics (
    organization_id,
    store_id,
    date,
    period_type,
    month,
    quarter,
    year,
    gross_revenue,
    sales_discounts,
    sales_returns,
    net_revenue,
    opening_inventory_value,
    purchases,
    closing_inventory_value,
    cogs,
    gross_profit,
    gross_profit_margin,
    total_expenses,
    net_profit,
    net_profit_margin,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
) RETURNING *;

-- name: GetProfitLossAnalytics :one
SELECT * FROM profit_loss_analytics
WHERE id = $1;

-- name: ListProfitLossAnalytics :many
SELECT * FROM profit_loss_analytics
WHERE organization_id = $1
ORDER BY date DESC
LIMIT $2 OFFSET $3;

-- name: GetProfitLossAnalyticsByDateRange :many
SELECT * FROM profit_loss_analytics
WHERE organization_id = $1
  AND date >= $2 AND date <= $3
ORDER BY date DESC;

-- name: GetProfitLossAnalyticsByStore :many
SELECT * FROM profit_loss_analytics
WHERE organization_id = $1 AND store_id = $2
  AND date >= $3 AND date <= $4
ORDER BY date DESC;

-- name: GetProfitLossAnalyticsByPeriod :many
SELECT * FROM profit_loss_analytics
WHERE organization_id = $1 AND period_type = $2
  AND date >= $3 AND date <= $4
ORDER BY date DESC;

-- name: GetProfitLossSummaryByMonth :many
SELECT 
    year,
    month,
    SUM(gross_revenue) AS total_gross_revenue,
    SUM(net_revenue) AS total_net_revenue,
    SUM(cogs) AS total_cogs,
    SUM(gross_profit) AS total_gross_profit,
    AVG(gross_profit_margin) AS avg_gross_profit_margin,
    SUM(total_expenses) AS total_expenses,
    SUM(net_profit) AS total_net_profit,
    AVG(net_profit_margin) AS avg_net_profit_margin
FROM profit_loss_analytics
WHERE organization_id = $1
  AND date >= $2 AND date <= $3
GROUP BY year, month
ORDER BY year DESC, month DESC;

-- name: GetProfitLossSummaryByQuarter :many
SELECT 
    year,
    quarter,
    SUM(gross_revenue) AS total_gross_revenue,
    SUM(net_revenue) AS total_net_revenue,
    SUM(cogs) AS total_cogs,
    SUM(gross_profit) AS total_gross_profit,
    SUM(net_profit) AS total_net_profit
FROM profit_loss_analytics
WHERE organization_id = $1
  AND date >= $2 AND date <= $3
GROUP BY year, quarter
ORDER BY year DESC, quarter DESC;

-- name: GetProfitLossSummaryByYear :many
SELECT 
    year,
    SUM(gross_revenue) AS total_gross_revenue,
    SUM(net_revenue) AS total_net_revenue,
    SUM(cogs) AS total_cogs,
    SUM(gross_profit) AS total_gross_profit,
    SUM(net_profit) AS total_net_profit
FROM profit_loss_analytics
WHERE organization_id = $1
  AND date >= $2 AND date <= $3
GROUP BY year
ORDER BY year DESC;

-- name: UpdateProfitLossAnalytics :one
UPDATE profit_loss_analytics
SET 
    gross_revenue = $2,
    sales_discounts = $3,
    sales_returns = $4,
    net_revenue = $5,
    opening_inventory_value = $6,
    purchases = $7,
    closing_inventory_value = $8,
    cogs = $9,
    gross_profit = $10,
    gross_profit_margin = $11,
    total_expenses = $12,
    net_profit = $13,
    net_profit_margin = $14,
    metadata = $15
WHERE id = $1
RETURNING *;

-- name: DeleteProfitLossAnalytics :exec
DELETE FROM profit_loss_analytics
WHERE id = $1;
