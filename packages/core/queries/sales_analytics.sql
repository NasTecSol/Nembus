-- name: GetDailySalesByStore :many
SELECT 
    date_trunc('day', t.transaction_date) AS sale_date,
    t.store_id,
    s.name AS store_name,
    COUNT(DISTINCT t.id) AS transaction_count,
    SUM(t.total_amount) AS gross_sales,
    SUM(t.discount_amount) AS total_discounts,
    SUM(t.tax_amount) AS total_taxes,
    SUM(t.total_amount - t.discount_amount) AS net_sales,
    AVG(t.total_amount) AS avg_transaction_value
FROM pos_transactions t
JOIN stores s ON t.store_id = s.id
WHERE t.status = 'completed'
  AND t.transaction_date >= $1::date
  AND t.transaction_date < $2::date + INTERVAL '1 day'
GROUP BY sale_date, t.store_id, s.name
ORDER BY sale_date DESC, net_sales DESC;

-- name: GetTopSellingProductsToday :many
SELECT 
    tl.product_id,
    p.sku,
    p.name,
    SUM(tl.quantity) AS total_qty,
    SUM(tl.line_total) AS total_amount,
    COUNT(DISTINCT t.id) AS transaction_count
FROM pos_transaction_lines tl
JOIN pos_transactions t ON tl.transaction_id = t.id
JOIN products p ON tl.product_id = p.id
WHERE t.status = 'completed'
  AND t.transaction_date >= CURRENT_DATE
  AND t.transaction_date < CURRENT_DATE + INTERVAL '1 day'
  AND t.store_id = $1
GROUP BY tl.product_id, p.sku, p.name
ORDER BY total_amount DESC
LIMIT 20;