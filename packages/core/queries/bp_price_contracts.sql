-- =====================================================
-- B2B PRICE CONTRACT QUERIES
-- Tier-1 pricing: negotiated contract prices per business partner + product SKU
-- =====================================================

-- name: CreateBPPriceContract :one
INSERT INTO bp_price_contracts (
    organization_id,
    business_partner_id,
    product_id,
    product_variant_id,
    contract_price,
    discount_percentage,
    min_quantity,
    valid_from,
    valid_to,
    is_active,
    notes
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING *;

-- name: GetBPPriceContractForProduct :one
-- Used in the 3-tier price resolution: Tier 1 (contract) lookup
SELECT bpc.*, bp.name AS partner_name
FROM bp_price_contracts bpc
JOIN business_partners bp ON bpc.business_partner_id = bp.id
WHERE bpc.business_partner_id = $1
  AND bpc.product_id = $2
  AND (
    sqlc.narg(product_variant_id)::int IS NULL
    OR bpc.product_variant_id = sqlc.narg(product_variant_id)::int
    OR bpc.product_variant_id IS NULL
  )
  AND bpc.is_active = true
  AND (bpc.valid_from IS NULL OR bpc.valid_from <= CURRENT_DATE)
  AND (bpc.valid_to IS NULL OR bpc.valid_to >= CURRENT_DATE)
  AND bpc.min_quantity <= COALESCE(sqlc.narg(quantity)::numeric, 1)
ORDER BY
  bpc.product_variant_id NULLS LAST, -- variant-specific contract preferred over general
  bpc.min_quantity DESC
LIMIT 1;

-- name: ListBPPriceContracts :many
SELECT
    bpc.*,
    bp.name AS partner_name,
    bp.code AS partner_code,
    p.name  AS product_name,
    p.sku   AS product_sku
FROM bp_price_contracts bpc
JOIN business_partners bp ON bpc.business_partner_id = bp.id
JOIN products p ON bpc.product_id = p.id
WHERE bpc.organization_id = $1
  AND (sqlc.narg(business_partner_id)::int IS NULL OR bpc.business_partner_id = sqlc.narg(business_partner_id)::int)
  AND (sqlc.narg(is_active)::boolean IS NULL OR bpc.is_active = sqlc.narg(is_active)::boolean)
ORDER BY bp.name, p.name
LIMIT $2 OFFSET $3;

-- name: UpdateBPPriceContract :one
UPDATE bp_price_contracts SET
    contract_price      = COALESCE(sqlc.narg(contract_price), contract_price),
    discount_percentage = COALESCE(sqlc.narg(discount_percentage), discount_percentage),
    min_quantity        = COALESCE(sqlc.narg(min_quantity), min_quantity),
    valid_from          = COALESCE(sqlc.narg(valid_from), valid_from),
    valid_to            = COALESCE(sqlc.narg(valid_to), valid_to),
    is_active           = COALESCE(sqlc.narg(is_active), is_active),
    notes               = COALESCE(sqlc.narg(notes), notes)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeactivateBPPriceContract :one
UPDATE bp_price_contracts SET is_active = false WHERE id = $1 RETURNING *;

-- name: DeactivateAllContractsForPartner :exec
UPDATE bp_price_contracts SET is_active = false
WHERE business_partner_id = $1 AND is_active = true;
