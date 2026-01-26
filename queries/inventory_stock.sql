-- name: GetLowStockProducts :many
SELECT 
    s.product_id,
    p.sku,
    p.name,
    s.store_id,
    st.name AS store_name,
    s.quantity_available,
    s.reorder_level,
    s.reorder_quantity,
    s.max_stock_level
FROM inventory_stock s
JOIN products p ON s.product_id = p.id
JOIN stores st ON s.store_id = st.id
WHERE s.quantity_available <= s.reorder_level
  AND s.quantity_available > 0
  AND p.is_active = true
  AND p.track_inventory = true
ORDER BY s.quantity_available ASC, p.name
LIMIT 50;

-- name: GetStockValuationByStore :many
SELECT 
    s.store_id,
    st.name AS store_name,
    COUNT(DISTINCT s.product_id) AS unique_products,
    SUM(s.quantity_on_hand * COALESCE((
        SELECT AVG(pp.price)
        FROM product_prices pp
        WHERE pp.product_id = s.product_id
          AND pp.is_active = true
    ), 0)) AS total_stock_value
FROM inventory_stock s
JOIN stores st ON s.store_id = st.id
WHERE st.organization_id = sqlc.arg('org_id')
GROUP BY s.store_id, st.name
ORDER BY total_stock_value DESC;