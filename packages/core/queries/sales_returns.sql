-- sales_returns.sql

-- name: CreateSalesReturn :one
INSERT INTO sales_returns (
    return_number,
    store_id,
    cashier_id,
    cashier_session_id,
    customer_id,
    original_transaction_id,
    return_date,
    return_reason,
    status,
    subtotal,
    tax_amount,
    total_refund_amount,
    refund_method,
    refund_reference,
    approved_by,
    notes,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
)
RETURNING *;

-- name: CreateSalesReturnLine :one
INSERT INTO sales_return_lines (
    return_id,
    product_id,
    product_variant_id,
    original_line_id,
    quantity,
    unit_price,
    refund_amount,
    return_to_stock,
    serial_number,
    batch_number,
    condition,
    line_number,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: GetSalesReturnByID :one
SELECT * FROM sales_returns
WHERE id = $1;

-- name: ListSalesReturnsByTransaction :many
SELECT * FROM sales_returns
WHERE original_transaction_id = $1;

-- name: ListSalesReturnLinesByReturnID :many
SELECT * FROM sales_return_lines
WHERE return_id = $1;

-- name: GetReturnedQuantityByLineID :one
SELECT COALESCE(SUM(quantity), 0)::DECIMAL(15,3) AS total_returned_qty
FROM sales_return_lines
WHERE original_line_id = $1;

-- name: UpdatePOSTransactionStatus :exec
UPDATE pos_transactions
SET status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateSalesReturnMetadata :exec
UPDATE sales_returns SET metadata = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: ListUnpostedSalesReturnsForSAP :many
SELECT sr.*,
    COALESCE(NULLIF(c.customer_code, ''), NULLIF(c.metadata->>'sap_card_code', ''), '')::text AS customer_sap_code,
    COALESCE(NULLIF(pt.metadata->>'sap_doc_entry', ''), '0')::int AS original_sap_doc_entry
FROM sales_returns sr
LEFT JOIN customers c ON c.id = sr.customer_id
LEFT JOIN pos_transactions pt ON pt.id = sr.original_transaction_id
WHERE sr.store_id = $1
  AND sr.status = 'completed'
  AND (sr.metadata->>'sap_doc_num' IS NULL OR sr.metadata->>'sap_doc_num' = '')
ORDER BY sr.id ASC
LIMIT $2;

-- name: ListSalesReturnLinesDetailedForSAP :many
SELECT l.*,
    p.sku AS product_sku,
    p.name AS product_name,
    COALESCE(NULLIF(u.metadata->>'sap_uom_entry', ''), NULLIF(bu.metadata->>'sap_uom_entry', ''), NULLIF(p.metadata->>'sap_i_uom_entry', ''), '')::text AS sap_uom_entry,
    COALESCE(NULLIF(p.metadata->>'sap_ugp_entry', ''), '')::text AS sap_ugp_entry,
    COALESCE(puc.conversion_factor, 1)::numeric AS uom_conversion_factor,
    (ptl.uom_id IS NULL OR ptl.uom_id = p.base_uom_id)::boolean AS is_base_uom,
    COALESCE(NULLIF(p.metadata->>'sap_vat_group', ''), tc.code, '')::text AS sap_tax_code
FROM sales_return_lines l
JOIN products p ON p.id = l.product_id
LEFT JOIN pos_transaction_lines ptl ON ptl.id = l.original_line_id
LEFT JOIN units_of_measure u ON u.id = ptl.uom_id
LEFT JOIN units_of_measure bu ON bu.id = p.base_uom_id
LEFT JOIN product_uom_conversions puc ON puc.product_id = l.product_id AND puc.from_uom_id = ptl.uom_id AND puc.to_uom_id = p.base_uom_id
LEFT JOIN tax_categories tc ON tc.id = p.tax_category_id
WHERE l.return_id = $1
ORDER BY l.id ASC;

-- name: MarkSalesReturnSAPSynced :one
UPDATE sales_returns
SET metadata = COALESCE(metadata, '{}'::jsonb) || $2::jsonb,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, return_number, metadata;


