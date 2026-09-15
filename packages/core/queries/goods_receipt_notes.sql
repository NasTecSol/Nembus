-- =====================================================
-- GOODS RECEIPT NOTES (GRN) QUERIES
-- =====================================================

-- name: CreateGoodsReceiptNote :one
INSERT INTO goods_receipt_notes (
    organization_id,
    grn_number,
    purchase_order_id,
    partners_id,
    store_id,
    received_by,
    receipt_date,
    delivery_note_number,
    status,
    notes,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: CreateGoodsReceiptNoteItem :one
INSERT INTO goods_receipt_note_items (
    grn_id,
    purchase_order_line_id,
    product_id,
    product_variant_id,
    storage_location_id,
    quantity_received,
    quantity_rejected,
    uom_id,
    unit_cost,
    batch_number,
    expiry_date,
    rejection_reason,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetGoodsReceiptNote :one
SELECT * FROM goods_receipt_notes WHERE id = $1;

-- name: GetGoodsReceiptNoteWithDetails :one
SELECT 
    grn.*,
    bp.name AS supplier_name,
    st.name AS store_name,
    u.username AS received_by_name,
    po.po_number
FROM goods_receipt_notes grn
JOIN business_partners bp ON grn.partners_id = bp.id
JOIN stores st ON grn.store_id = st.id
LEFT JOIN users u ON grn.received_by = u.id
LEFT JOIN purchase_orders po ON grn.purchase_order_id = po.id
WHERE grn.id = $1;

-- name: ListGoodsReceiptNoteItems :many
SELECT 
    grni.*,
    p.name AS product_name,
    p.sku AS product_sku,
    uom.name AS uom_name
FROM goods_receipt_note_items grni
JOIN products p ON grni.product_id = p.id
LEFT JOIN units_of_measure uom ON grni.uom_id = uom.id
WHERE grni.grn_id = $1;

-- name: ListGoodsReceiptNotesByOrganization :many
SELECT 
    grn.*,
    bp.name AS supplier_name,
    st.name AS store_name
FROM goods_receipt_notes grn
JOIN business_partners bp ON grn.partners_id = bp.id
JOIN stores st ON grn.store_id = st.id
WHERE grn.organization_id = $1
ORDER BY grn.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListGoodsReceiptNotesByPurchaseOrder :many
SELECT 
    grn.*,
    bp.name AS supplier_name,
    st.name AS store_name
FROM goods_receipt_notes grn
JOIN business_partners bp ON grn.partners_id = bp.id
JOIN stores st ON grn.store_id = st.id
WHERE grn.purchase_order_id = $1
ORDER BY grn.created_at DESC;

-- name: ListGoodsReceiptNotes :many
SELECT 
    grn.id,
    grn.organization_id,
    grn.grn_number,
    grn.purchase_order_id,
    grn.partners_id,
    grn.store_id,
    grn.received_by,
    grn.receipt_date,
    grn.delivery_note_number,
    grn.status,
    grn.notes,
    grn.metadata,
    grn.created_at,
    grn.updated_at,
    bp.name AS supplier_name,
    bp.code AS supplier_code,
    st.name AS store_name,
    po.po_number,
    u.username AS received_by_name,
    COUNT(grni.id) AS item_count,
    COALESCE(SUM(grni.quantity_received), 0)::numeric AS total_received_quantity
FROM goods_receipt_notes grn
JOIN business_partners bp ON grn.partners_id = bp.id
JOIN stores st ON grn.store_id = st.id
LEFT JOIN users u ON grn.received_by = u.id
LEFT JOIN purchase_orders po ON grn.purchase_order_id = po.id
LEFT JOIN goods_receipt_note_items grni ON grni.grn_id = grn.id
WHERE (sqlc.narg('organization_id')::int IS NULL OR grn.organization_id = sqlc.narg('organization_id'))
  AND (sqlc.narg('store_id')::int IS NULL OR grn.store_id = sqlc.narg('store_id'))
  AND (sqlc.narg('supplier_id')::int IS NULL OR grn.partners_id = sqlc.narg('supplier_id'))
  AND (sqlc.narg('purchase_order_id')::int IS NULL OR grn.purchase_order_id = sqlc.narg('purchase_order_id'))
  AND (sqlc.narg('status')::text IS NULL OR grn.status = sqlc.narg('status'))
  AND (sqlc.narg('from_date')::date IS NULL OR grn.receipt_date::date >= sqlc.narg('from_date'))
  AND (sqlc.narg('to_date')::date IS NULL OR grn.receipt_date::date <= sqlc.narg('to_date'))
  AND (
      sqlc.narg('search')::text IS NULL 
      OR grn.grn_number ILIKE '%' || sqlc.narg('search')::text || '%'
      OR grn.delivery_note_number ILIKE '%' || sqlc.narg('search')::text || '%'
      OR po.po_number ILIKE '%' || sqlc.narg('search')::text || '%'
      OR bp.name ILIKE '%' || sqlc.narg('search')::text || '%'
  )
GROUP BY grn.id, bp.name, bp.code, st.name, po.po_number, u.username
ORDER BY grn.created_at DESC
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: CountGoodsReceiptNotes :one
SELECT COUNT(DISTINCT grn.id)
FROM goods_receipt_notes grn
JOIN business_partners bp ON grn.partners_id = bp.id
LEFT JOIN purchase_orders po ON grn.purchase_order_id = po.id
WHERE (sqlc.narg('organization_id')::int IS NULL OR grn.organization_id = sqlc.narg('organization_id'))
  AND (sqlc.narg('store_id')::int IS NULL OR grn.store_id = sqlc.narg('store_id'))
  AND (sqlc.narg('supplier_id')::int IS NULL OR grn.partners_id = sqlc.narg('supplier_id'))
  AND (sqlc.narg('purchase_order_id')::int IS NULL OR grn.purchase_order_id = sqlc.narg('purchase_order_id'))
  AND (sqlc.narg('status')::text IS NULL OR grn.status = sqlc.narg('status'))
  AND (sqlc.narg('from_date')::date IS NULL OR grn.receipt_date::date >= sqlc.narg('from_date'))
  AND (sqlc.narg('to_date')::date IS NULL OR grn.receipt_date::date <= sqlc.narg('to_date'))
  AND (
      sqlc.narg('search')::text IS NULL 
      OR grn.grn_number ILIKE '%' || sqlc.narg('search')::text || '%'
      OR grn.delivery_note_number ILIKE '%' || sqlc.narg('search')::text || '%'
      OR po.po_number ILIKE '%' || sqlc.narg('search')::text || '%'
      OR bp.name ILIKE '%' || sqlc.narg('search')::text || '%'
  );

-- name: UpdateGoodsReceiptNoteHeader :one
UPDATE goods_receipt_notes
SET partners_id = $2,
    store_id = $3,
    purchase_order_id = $4,
    received_by = $5,
    receipt_date = $6,
    delivery_note_number = $7,
    notes = $8,
    metadata = $9,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteGoodsReceiptNoteItems :exec
DELETE FROM goods_receipt_note_items WHERE grn_id = $1;

-- name: DeleteGoodsReceiptNote :exec
DELETE FROM goods_receipt_notes WHERE id = $1;

-- name: CallProcessGoodsReceipt :one
SELECT success::boolean AS success, message::text AS message FROM fn_process_goods_receipt($1);
