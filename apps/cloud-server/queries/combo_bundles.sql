-- =====================================================
-- COMBO BUNDLES & PROMOTION DEALS (Retail, Wholesale & Restaurant)
-- =====================================================

-- name: CreateComboBundle :one
INSERT INTO combo_bundles (
    organization_id,
    store_id,
    price_list_id,
    applicable_customer_types,
    code,
    name,
    description,
    bundle_price,
    bundle_type,
    is_active,
    valid_from,
    valid_to,
    display_order,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: GetComboBundle :one
SELECT * FROM combo_bundles
WHERE id = $1;

-- name: GetComboBundleByCode :one
SELECT * FROM combo_bundles
WHERE organization_id = $1 AND code = $2;

-- name: ListComboBundlesByOrganization :many
SELECT * FROM combo_bundles
WHERE organization_id = $1
  AND (sqlc.narg('store_id')::int IS NULL OR store_id IS NULL OR store_id = sqlc.narg('store_id'))
  AND (sqlc.narg('is_active')::boolean IS NULL OR is_active = sqlc.narg('is_active'))
  AND (sqlc.narg('price_list_id')::int IS NULL OR price_list_id IS NULL OR price_list_id = sqlc.narg('price_list_id'))
ORDER BY display_order ASC, name ASC;

-- name: ListComboBundlesByStore :many
SELECT * FROM combo_bundles
WHERE (store_id = $1 OR (organization_id = (SELECT organization_id FROM stores WHERE id = $1) AND store_id IS NULL))
  AND is_active = true
  AND (valid_from IS NULL OR valid_from <= CURRENT_DATE)
  AND (valid_to IS NULL OR valid_to >= CURRENT_DATE)
ORDER BY display_order ASC, name ASC;

-- name: UpdateComboBundle :one
UPDATE combo_bundles
SET
    store_id                  = $2,
    price_list_id             = $3,
    applicable_customer_types = $4,
    code                      = $5,
    name                      = $6,
    description               = $7,
    bundle_price              = $8,
    bundle_type               = $9,
    is_active                 = $10,
    valid_from                = $11,
    valid_to                  = $12,
    display_order             = $13,
    metadata                  = $14,
    updated_at                = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ToggleComboBundleActive :one
UPDATE combo_bundles
SET is_active = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteComboBundle :exec
DELETE FROM combo_bundles
WHERE id = $1;

-- =====================================================
-- COMBO BUNDLE ITEMS
-- =====================================================

-- name: CreateComboBundleItem :one
INSERT INTO combo_bundle_items (
    combo_bundle_id,
    menu_item_id,
    product_id,
    product_variant_id,
    item_type,
    quantity,
    is_required,
    group_tag,
    price_override,
    display_order,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetComboBundleItem :one
SELECT * FROM combo_bundle_items
WHERE id = $1;

-- name: ListComboBundleItems :many
SELECT * FROM combo_bundle_items
WHERE combo_bundle_id = $1
ORDER BY display_order ASC, id ASC;

-- name: UpdateComboBundleItem :one
UPDATE combo_bundle_items
SET
    menu_item_id       = $2,
    product_id         = $3,
    product_variant_id = $4,
    item_type          = $5,
    quantity           = $6,
    is_required        = $7,
    group_tag          = $8,
    price_override     = $9,
    display_order      = $10,
    metadata           = $11
WHERE id = $1
RETURNING *;

-- name: DeleteComboBundleItem :exec
DELETE FROM combo_bundle_items
WHERE id = $1;

-- name: DeleteComboBundleItemsByBundle :exec
DELETE FROM combo_bundle_items
WHERE combo_bundle_id = $1;

-- name: GetComboBundleWithItems :many
SELECT
    cbi.id AS item_id,
    cbi.combo_bundle_id,
    cbi.item_type,
    cbi.quantity,
    cbi.is_required,
    cbi.group_tag,
    cbi.price_override,
    cbi.display_order,
    cbi.menu_item_id,
    mi.name AS menu_item_name,
    mi.base_price AS menu_item_price,
    cbi.product_id,
    p.name AS product_name,
    p.sku AS product_sku,
    cbi.product_variant_id,
    pv.variant_name,
    pv.variant_sku
FROM combo_bundle_items cbi
LEFT JOIN menu_items mi ON cbi.menu_item_id = mi.id
LEFT JOIN products p ON cbi.product_id = p.id
LEFT JOIN product_variants pv ON cbi.product_variant_id = pv.id
WHERE cbi.combo_bundle_id = $1
ORDER BY cbi.display_order ASC, cbi.id ASC;
