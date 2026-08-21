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