-- =====================================================
-- TRANSFER REQUESTS QUERIES
-- =====================================================

-- name: CreateTransferRequest :one
INSERT INTO transfer_requests (
    organization_id,
    transfer_number,
    from_store_id,
    to_store_id,
    status,
    requested_by,
    request_date,
    expected_delivery_date,
    notes,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: CreateTransferRequestItem :one
INSERT INTO transfer_request_items (
    transfer_request_id,
    product_id,
    product_variant_id,
    from_location_id,
    to_location_id,
    requested_quantity,
    approved_quantity,
    shipped_quantity,
    received_quantity,
    uom_id,
    batch_number,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetTransferRequest :one
SELECT * FROM transfer_requests WHERE id = $1;

-- name: GetTransferRequestWithDetails :one
SELECT 
    tr.*,
    fs.name AS from_store_name,
    ts.name AS to_store_name,
    req_u.username AS requested_by_name,
    app_u.username AS approved_by_name
FROM transfer_requests tr
LEFT JOIN stores fs ON tr.from_store_id = fs.id
LEFT JOIN stores ts ON tr.to_store_id = ts.id
LEFT JOIN users req_u ON tr.requested_by = req_u.id
LEFT JOIN users app_u ON tr.approved_by = app_u.id
WHERE tr.id = $1;

-- name: ListTransferRequestItems :many
SELECT 
    tri.*,
    p.name AS product_name,
    p.sku AS product_sku,
    uom.name AS uom_name
FROM transfer_request_items tri
JOIN products p ON tri.product_id = p.id
LEFT JOIN units_of_measure uom ON tri.uom_id = uom.id
WHERE tri.transfer_request_id = $1;

-- name: ListTransferRequestsByOrganization :many
SELECT 
    tr.*,
    fs.name AS from_store_name,
    ts.name AS to_store_name
FROM transfer_requests tr
LEFT JOIN stores fs ON tr.from_store_id = fs.id
LEFT JOIN stores ts ON tr.to_store_id = ts.id
WHERE tr.organization_id = $1
ORDER BY tr.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListTransferRequestsByStore :many
SELECT 
    tr.*,
    fs.name AS from_store_name,
    ts.name AS to_store_name
FROM transfer_requests tr
LEFT JOIN stores fs ON tr.from_store_id = fs.id
LEFT JOIN stores ts ON tr.to_store_id = ts.id
WHERE tr.from_store_id = $1 OR tr.to_store_id = $1
ORDER BY tr.created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateTransferRequestStatus :one
UPDATE transfer_requests
SET status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: CallApproveTransferRequest :one
SELECT success::boolean AS success, message::text AS message FROM fn_approve_transfer_request($1, $2);

-- name: CallShipTransferRequest :one
SELECT success::boolean AS success, message::text AS message FROM fn_ship_transfer_request($1, $2);

-- name: CallReceiveTransferRequest :one
SELECT success::boolean AS success, message::text AS message FROM fn_receive_transfer_request($1, $2);
