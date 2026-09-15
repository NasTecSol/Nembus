-- =====================================================
-- PURCHASE ORDERS QUERIES
-- =====================================================

-- name: CreatePurchaseOrder :one
INSERT INTO purchase_orders (
    organization_id,
    po_number,
    partners_id,
    store_id,
    po_date,
    expected_delivery_date,
    status,
    subtotal,
    discount_amount,
    tax_amount,
    total_amount,
    price_list_id,
    created_by,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: CreatePurchaseOrderLine :one
INSERT INTO purchase_order_lines (
    purchase_order_id,
    product_id,
    product_variant_id,
    quantity,
    uom_id,
    unit_price,
    discount_amount,
    tax_amount,
    subtotal,
    line_total,
    received_quantity,
    line_number,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetPurchaseOrderByID :one
SELECT * FROM purchase_orders WHERE id = $1;

-- name: GetPurchaseOrderByNumber :one
SELECT * FROM purchase_orders 
WHERE organization_id = $1 AND po_number = $2;

-- name: GetPurchaseOrderWithDetails :one
SELECT 
    po.*,
    bp.name AS partner_name,
    bp.code AS partner_code,
    s.name AS store_name,
    u_created.username AS created_by_name,
    u_approved.username AS approved_by_name
FROM purchase_orders po
LEFT JOIN business_partners bp ON po.partners_id = bp.id
LEFT JOIN stores s ON po.store_id = s.id
LEFT JOIN users u_created ON po.created_by = u_created.id
LEFT JOIN users u_approved ON po.approved_by = u_approved.id
WHERE po.id = $1;

-- name: ListPurchaseOrderLines :many
SELECT 
    pol.*,
    p.name AS product_name,
    p.sku AS product_sku,
    pv.variant_name,
    pv.variant_sku,
    uom.name AS uom_name,
    COALESCE(pb.barcode, '') AS barcode
FROM purchase_order_lines pol
JOIN products p ON pol.product_id = p.id
LEFT JOIN product_variants pv ON pol.product_variant_id = pv.id
LEFT JOIN units_of_measure uom ON pol.uom_id = uom.id
LEFT JOIN product_barcodes pb ON pb.product_id = p.id AND pb.is_primary = true
WHERE pol.purchase_order_id = $1
ORDER BY pol.line_number ASC, pol.id ASC;

-- name: ListPurchaseOrders :many
SELECT 
    po.*,
    bp.name AS partner_name,
    bp.code AS partner_code,
    s.name AS store_name,
    COUNT(pol.id) AS item_count,
    COALESCE(SUM(pol.quantity), 0)::numeric AS total_quantity
FROM purchase_orders po
LEFT JOIN business_partners bp ON po.partners_id = bp.id
LEFT JOIN stores s ON po.store_id = s.id
LEFT JOIN purchase_order_lines pol ON pol.purchase_order_id = po.id
WHERE po.organization_id = sqlc.arg('organization_id')
  AND (sqlc.narg('store_id')::int IS NULL OR po.store_id = sqlc.narg('store_id'))
  AND (sqlc.narg('partner_id')::int IS NULL OR po.partners_id = sqlc.narg('partner_id'))
  AND (sqlc.narg('status')::text IS NULL OR po.status = sqlc.narg('status'))
  AND (sqlc.narg('from_date')::date IS NULL OR po.po_date >= sqlc.narg('from_date'))
  AND (sqlc.narg('to_date')::date IS NULL OR po.po_date <= sqlc.narg('to_date'))
  AND (
      sqlc.narg('search')::text IS NULL 
      OR po.po_number ILIKE '%' || sqlc.narg('search')::text || '%'
      OR bp.name ILIKE '%' || sqlc.narg('search')::text || '%'
  )
GROUP BY po.id, bp.name, bp.code, s.name
ORDER BY po.created_at DESC
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: CountPurchaseOrders :one
SELECT COUNT(DISTINCT po.id)
FROM purchase_orders po
LEFT JOIN business_partners bp ON po.partners_id = bp.id
WHERE po.organization_id = sqlc.arg('organization_id')
  AND (sqlc.narg('store_id')::int IS NULL OR po.store_id = sqlc.narg('store_id'))
  AND (sqlc.narg('partner_id')::int IS NULL OR po.partners_id = sqlc.narg('partner_id'))
  AND (sqlc.narg('status')::text IS NULL OR po.status = sqlc.narg('status'))
  AND (sqlc.narg('from_date')::date IS NULL OR po.po_date >= sqlc.narg('from_date'))
  AND (sqlc.narg('to_date')::date IS NULL OR po.po_date <= sqlc.narg('to_date'))
  AND (
      sqlc.narg('search')::text IS NULL 
      OR po.po_number ILIKE '%' || sqlc.narg('search')::text || '%'
      OR bp.name ILIKE '%' || sqlc.narg('search')::text || '%'
  );

-- name: UpdatePurchaseOrderHeader :one
UPDATE purchase_orders
SET partners_id = $2,
    store_id = $3,
    po_date = $4,
    expected_delivery_date = $5,
    subtotal = $6,
    discount_amount = $7,
    tax_amount = $8,
    total_amount = $9,
    price_list_id = $10,
    metadata = $11,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdatePurchaseOrderStatus :one
UPDATE purchase_orders
SET status = $2,
    approved_by = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeletePurchaseOrderLines :exec
DELETE FROM purchase_order_lines
WHERE purchase_order_id = $1;

-- name: DeletePurchaseOrder :exec
DELETE FROM purchase_orders
WHERE id = $1;

-- name: GetPurchaseOrderWithReceivedQty :many
SELECT 
    po.id,
    po.po_number,
    po.status,
    po.total_amount,
    po.expected_delivery_date,
    bp.name AS supplier_name,
    pol.line_number,
    pol.product_id,
    pol.quantity           AS ordered_qty,
    pol.received_quantity  AS received_qty,
    pol.quantity - pol.received_quantity AS pending_qty,
    p.sku,
    p.name AS product_name
FROM purchase_orders po
JOIN business_partners bp ON po.partners_id = bp.id
JOIN purchase_order_lines pol ON pol.purchase_order_id = po.id
JOIN products p ON pol.product_id = p.id
WHERE po.id = $1
ORDER BY pol.line_number;

-- name: GetPurchaseOrderLineByID :one
SELECT 
    pol.*,
    p.name AS product_name,
    p.sku AS product_sku,
    pv.variant_name,
    pv.variant_sku,
    uom.name AS uom_name,
    COALESCE(pb.barcode, '') AS barcode
FROM purchase_order_lines pol
JOIN products p ON pol.product_id = p.id
LEFT JOIN product_variants pv ON pol.product_variant_id = pv.id
LEFT JOIN units_of_measure uom ON pol.uom_id = uom.id
LEFT JOIN product_barcodes pb ON pb.product_id = p.id AND pb.is_primary = true
WHERE pol.id = $1;

-- name: UpdatePurchaseOrderLine :one
UPDATE purchase_order_lines
SET product_id = $2,
    product_variant_id = $3,
    quantity = $4,
    uom_id = $5,
    unit_price = $6,
    discount_amount = $7,
    tax_amount = $8,
    subtotal = $9,
    line_total = $10,
    line_number = $11,
    metadata = $12
WHERE id = $1
RETURNING *;

-- name: DeletePurchaseOrderLine :exec
DELETE FROM purchase_order_lines
WHERE id = $1;