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
    so.order_status::text AS status,
    so.total_amount,
    so.sales_channel,
    COALESCE(so.customer_name, c.name, '') AS customer_name,
    COALESCE(st.name, '') AS store_name,
    COUNT(sol.id) AS line_count,
    COALESCE(SUM(sol.quantity_ordered), 0)::numeric AS total_items
FROM sales_orders_v2 so
LEFT JOIN customers c ON so.customer_id = c.id
LEFT JOIN stores st ON so.store_id = st.id
LEFT JOIN sales_order_lines_v2 sol ON sol.sales_order_id = so.id
WHERE so.organization_id = sqlc.arg('org_id')
  AND so.order_date >= sqlc.arg('from_date')::timestamp
  AND so.order_date <= sqlc.arg('to_date')::timestamp
  AND (sqlc.arg('status')::text IS NULL OR so.order_status::text = sqlc.arg('status'))
  AND (sqlc.arg('sales_channel')::text IS NULL OR so.sales_channel = sqlc.arg('sales_channel'))
GROUP BY so.id, so.order_number, so.order_date, so.order_status, so.total_amount, so.sales_channel, so.customer_name, c.name, st.name
ORDER BY so.order_date DESC;

-- name: GetSalesOrderTotalsByStatus :many
SELECT 
    order_status::text AS status,
    COUNT(*) AS order_count,
    COALESCE(SUM(total_amount), 0)::numeric AS total_revenue,
    COALESCE(AVG(total_amount), 0)::numeric AS avg_order_value
FROM sales_orders_v2
WHERE organization_id = sqlc.arg('organization_id')
  AND order_date >= sqlc.arg('from_date')::timestamp
  AND order_date <= sqlc.arg('to_date')::timestamp
GROUP BY order_status
ORDER BY order_status;

-- name: LinkPosTransactionToSalesOrder :exec
UPDATE pos_transactions
SET sales_order_id = $2
WHERE id = $1;