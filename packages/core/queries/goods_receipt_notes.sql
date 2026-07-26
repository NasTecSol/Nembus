-- =====================================================
-- GOODS RECEIPT NOTES (GRN) QUERIES
-- =====================================================

-- name: CreateGoodsReceiptNote :one
INSERT INTO goods_receipt_notes (
    organization_id,
    grn_number,
    purchase_order_id,
    supplier_id,
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
    sup.name AS supplier_name,
    st.name AS store_name,
    u.username AS received_by_name,
    po.po_number
FROM goods_receipt_notes grn
JOIN suppliers sup ON grn.supplier_id = sup.id
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
    sup.name AS supplier_name,
    st.name AS store_name
FROM goods_receipt_notes grn
JOIN suppliers sup ON grn.supplier_id = sup.id
JOIN stores st ON grn.store_id = st.id
WHERE grn.organization_id = $1
ORDER BY grn.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListGoodsReceiptNotesByPurchaseOrder :many
SELECT 
    grn.*,
    sup.name AS supplier_name,
    st.name AS store_name
FROM goods_receipt_notes grn
JOIN suppliers sup ON grn.supplier_id = sup.id
JOIN stores st ON grn.store_id = st.id
WHERE grn.purchase_order_id = $1
ORDER BY grn.created_at DESC;

-- name: CallProcessGoodsReceipt :one
SELECT success::boolean AS success, message::text AS message FROM fn_process_goods_receipt($1);
