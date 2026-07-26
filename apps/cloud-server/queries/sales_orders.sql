-- name: CreateSalesOrderHeader :one
INSERT INTO sales_orders (
    order_number, organization_id, customer_id, store_id,
    order_date, delivery_date, price_list_id, status,
    subtotal, tax_amount, discount_amount, total_amount,
    created_by, metadata
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
RETURNING id, order_number, status, total_amount;

-- name: GetSalesOrderFull :many
SELECT 
    so.*,
    c.name AS customer_name,
    c.customer_code,
    st.name AS store_name,
    u.first_name || ' ' || u.last_name AS created_by_name,
    sol.id AS line_id,
    sol.line_number,
    sol.product_id,
    sol.quantity,
    sol.unit_price,
    sol.discount_amount,
    sol.tax_amount,
    sol.line_total,
    p.sku,
    p.name AS product_name,
    COALESCE(pb.barcode, '') AS primary_barcode
FROM sales_orders so
LEFT JOIN customers     c  ON so.customer_id = c.id
LEFT JOIN stores        st ON so.store_id    = st.id
LEFT JOIN users         u  ON so.created_by  = u.id
LEFT JOIN sales_order_lines sol ON sol.sales_order_id = so.id
LEFT JOIN products      p  ON sol.product_id = p.id
LEFT JOIN product_barcodes pb 
    ON pb.product_id = p.id AND pb.is_primary = true
WHERE so.id = $1
ORDER BY sol.line_number NULLS LAST;

-- name: ListSalesOrdersDashboard :many
SELECT 
    so.id,
    so.order_number,
    so.order_date,
    so.status,
    so.total_amount,
    c.name AS customer_name,
    st.name AS store_name,
    COUNT(sol.id) AS line_count,
    SUM(sol.quantity) AS total_items
FROM sales_orders so
LEFT JOIN customers c ON so.customer_id = c.id
LEFT JOIN stores st ON so.store_id = st.id
LEFT JOIN sales_order_lines sol ON sol.sales_order_id = so.id
WHERE so.organization_id = sqlc.arg('org_id')
  AND so.order_date >= sqlc.arg('from_date')::date
  AND so.order_date <= sqlc.arg('to_date')::date
  AND (sqlc.arg('status')::text IS NULL OR so.status = sqlc.arg('status'))
GROUP BY so.id, c.name, st.name
ORDER BY so.order_date DESC, so.id DESC
LIMIT 100;

-- name: GetSalesOrderTotalsByStatus :many
SELECT 
    status,
    COUNT(*) AS order_count,
    COALESCE(SUM(total_amount), 0) AS total_revenue,
    COALESCE(AVG(total_amount), 0) AS avg_order_value
FROM sales_orders
WHERE organization_id = $1
  AND order_date >= $2::date
  AND order_date <= $3::date
GROUP BY status
ORDER BY status;