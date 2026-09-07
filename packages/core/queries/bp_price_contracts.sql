-- =====================================================
-- BUSINESS PARTNER PRICE CONTRACTS QUERIES (SQLC)
-- =====================================================

-- name: CreateBPPriceContract :one
INSERT INTO bp_price_contracts (
    organization_id,
    partner_id,
    product_id,
    product_variant_id,
    contract_price,
    discount_percentage,
    min_quantity,
    valid_from,
    valid_to,
    is_active,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetBPPriceContract :one
SELECT 
    bpc.id,
    bpc.organization_id,
    bpc.partner_id,
    bpc.product_id,
    bpc.product_variant_id,
    bpc.contract_price,
    bpc.discount_percentage,
    bpc.min_quantity,
    bpc.valid_from,
    bpc.valid_to,
    bpc.is_active,
    bpc.notes,
    bpc.created_at,
    bpc.updated_at,
    bp.name AS partner_name,
    bp.code AS partner_code,
    p.name AS product_name,
    p.sku AS product_sku,
    pv.variant_name,
    pv.variant_sku
FROM bp_price_contracts bpc
INNER JOIN business_partners bp ON bpc.partner_id = bp.id
INNER JOIN products p ON bpc.product_id = p.id
LEFT JOIN product_variants pv ON bpc.product_variant_id = pv.id
WHERE bpc.id = $1 LIMIT 1;

-- name: GetBPPriceContractRaw :one
SELECT * FROM bp_price_contracts
WHERE id = $1 LIMIT 1;

-- name: GetBPPriceContractByUnique :one
SELECT * FROM bp_price_contracts
WHERE partner_id = $1 
  AND product_id = $2 
  AND (
    (sqlc.narg(product_variant_id)::int IS NULL AND product_variant_id IS NULL)
    OR
    (product_variant_id = sqlc.narg(product_variant_id)::int)
  )
LIMIT 1;

-- name: ListBPPriceContracts :many
SELECT 
    bpc.id,
    bpc.organization_id,
    bpc.partner_id,
    bpc.product_id,
    bpc.product_variant_id,
    bpc.contract_price,
    bpc.discount_percentage,
    bpc.min_quantity,
    bpc.valid_from,
    bpc.valid_to,
    bpc.is_active,
    bpc.notes,
    bpc.created_at,
    bpc.updated_at,
    bp.name AS partner_name,
    bp.code AS partner_code,
    p.name AS product_name,
    p.sku AS product_sku,
    pv.variant_name,
    pv.variant_sku
FROM bp_price_contracts bpc
INNER JOIN business_partners bp ON bpc.partner_id = bp.id
INNER JOIN products p ON bpc.product_id = p.id
LEFT JOIN product_variants pv ON bpc.product_variant_id = pv.id
WHERE bpc.organization_id = $1
  AND (sqlc.narg(partner_id)::int IS NULL OR bpc.partner_id = sqlc.narg(partner_id)::int)
  AND (sqlc.narg(product_id)::int IS NULL OR bpc.product_id = sqlc.narg(product_id)::int)
  AND (sqlc.narg(is_active)::boolean IS NULL OR bpc.is_active = sqlc.narg(is_active)::boolean)
ORDER BY bpc.created_at DESC;

-- name: ListBPPriceContractsByPartner :many
SELECT 
    bpc.id,
    bpc.organization_id,
    bpc.partner_id,
    bpc.product_id,
    bpc.product_variant_id,
    bpc.contract_price,
    bpc.discount_percentage,
    bpc.min_quantity,
    bpc.valid_from,
    bpc.valid_to,
    bpc.is_active,
    bpc.notes,
    bpc.created_at,
    bpc.updated_at,
    bp.name AS partner_name,
    bp.code AS partner_code,
    p.name AS product_name,
    p.sku AS product_sku,
    pv.variant_name,
    pv.variant_sku
FROM bp_price_contracts bpc
INNER JOIN business_partners bp ON bpc.partner_id = bp.id
INNER JOIN products p ON bpc.product_id = p.id
LEFT JOIN product_variants pv ON bpc.product_variant_id = pv.id
WHERE bpc.partner_id = $1
ORDER BY bpc.created_at DESC;

-- name: GetEffectiveBPPriceContract :one
SELECT 
    bpc.id,
    bpc.organization_id,
    bpc.partner_id,
    bpc.product_id,
    bpc.product_variant_id,
    bpc.contract_price,
    bpc.discount_percentage,
    bpc.min_quantity,
    bpc.valid_from,
    bpc.valid_to,
    bpc.is_active,
    bpc.notes,
    bpc.created_at,
    bpc.updated_at,
    bp.name AS partner_name,
    bp.code AS partner_code,
    p.name AS product_name,
    p.sku AS product_sku,
    pv.variant_name,
    pv.variant_sku
FROM bp_price_contracts bpc
INNER JOIN business_partners bp ON bpc.partner_id = bp.id
INNER JOIN products p ON bpc.product_id = p.id
LEFT JOIN product_variants pv ON bpc.product_variant_id = pv.id
WHERE bpc.partner_id = $1
  AND bpc.product_id = $2
  AND (
    (sqlc.narg(product_variant_id)::int IS NULL AND bpc.product_variant_id IS NULL)
    OR
    (bpc.product_variant_id = sqlc.narg(product_variant_id)::int)
  )
  AND bpc.is_active = true
  AND (bpc.valid_from IS NULL OR bpc.valid_from <= CURRENT_DATE)
  AND (bpc.valid_to IS NULL OR bpc.valid_to >= CURRENT_DATE)
  AND bpc.min_quantity <= COALESCE(sqlc.narg(quantity), 1)
ORDER BY bpc.min_quantity DESC
LIMIT 1;

-- name: UpdateBPPriceContract :one
UPDATE bp_price_contracts
SET 
    contract_price = COALESCE(sqlc.narg(contract_price), contract_price),
    discount_percentage = COALESCE(sqlc.narg(discount_percentage), discount_percentage),
    min_quantity = COALESCE(sqlc.narg(min_quantity), min_quantity),
    valid_from = COALESCE(sqlc.narg(valid_from), valid_from),
    valid_to = COALESCE(sqlc.narg(valid_to), valid_to),
    is_active = COALESCE(sqlc.narg(is_active), is_active),
    notes = COALESCE(sqlc.narg(notes), notes),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteBPPriceContract :exec
DELETE FROM bp_price_contracts
WHERE id = $1;

-- name: ToggleBPPriceContractActive :one
UPDATE bp_price_contracts
SET is_active = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
