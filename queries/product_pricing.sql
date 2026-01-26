-- =====================================================
-- PRICE LISTS
-- Note: Price list queries are in price_lists_query.sql
-- =====================================================

-- =====================================================
-- PRODUCT PRICES
-- =====================================================

-- name: CreateProductPrice :one
INSERT INTO product_prices (
    product_id, product_variant_id, price_list_id, uom_id,
    price, min_quantity, max_quantity, valid_from, valid_to,
    is_active, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetProductPrice :one
SELECT * FROM product_prices WHERE id = $1 LIMIT 1;

-- name: GetProductPriceForList :one
SELECT pp.* FROM product_prices pp
WHERE pp.product_id = $1
  AND pp.product_variant_id = COALESCE(sqlc.narg(product_variant_id), pp.product_variant_id)
  AND pp.price_list_id = $2
  AND pp.uom_id = COALESCE(sqlc.narg(uom_id), pp.uom_id)
  AND pp.is_active = true
  AND (pp.valid_from IS NULL OR pp.valid_from <= CURRENT_DATE)
  AND (pp.valid_to IS NULL OR pp.valid_to >= CURRENT_DATE)
  AND pp.min_quantity <= COALESCE(sqlc.narg(quantity), 1)
  AND (pp.max_quantity IS NULL OR pp.max_quantity >= COALESCE(sqlc.narg(quantity), 1))
ORDER BY pp.min_quantity DESC
LIMIT 1;

-- name: ListProductPrices :many
SELECT 
    pp.*,
    pl.name as price_list_name,
    pl.code as price_list_code,
    uom.name as uom_name,
    uom.code as uom_code
FROM product_prices pp
INNER JOIN price_lists pl ON pp.price_list_id = pl.id
LEFT JOIN units_of_measure uom ON pp.uom_id = uom.id
WHERE pp.product_id = $1
  AND pp.product_variant_id = COALESCE(sqlc.narg(product_variant_id), pp.product_variant_id)
  AND pp.is_active = COALESCE(sqlc.narg(is_active), pp.is_active)
ORDER BY pl.name, pp.min_quantity;

-- name: ListPricesByPriceList :many
SELECT 
    pp.*,
    p.name as product_name,
    p.sku as product_sku,
    pv.variant_name,
    pv.variant_sku
FROM product_prices pp
INNER JOIN products p ON pp.product_id = p.id
LEFT JOIN product_variants pv ON pp.product_variant_id = pv.id
WHERE pp.price_list_id = $1
  AND pp.is_active = true
ORDER BY p.name;

-- name: GetEffectivePrice :one
SELECT pp.* FROM product_prices pp
WHERE pp.product_id = $1
  AND pp.product_variant_id = COALESCE(sqlc.narg(product_variant_id), pp.product_variant_id)
  AND pp.price_list_id = $2
  AND pp.is_active = true
  AND (pp.valid_from IS NULL OR pp.valid_from <= CURRENT_DATE)
  AND (pp.valid_to IS NULL OR pp.valid_to >= CURRENT_DATE)
  AND pp.min_quantity <= $3
  AND (pp.max_quantity IS NULL OR pp.max_quantity >= $3)
ORDER BY pp.min_quantity DESC
LIMIT 1;

-- name: UpdateProductPrice :one
UPDATE product_prices
SET 
    price = COALESCE(sqlc.narg(price), price),
    min_quantity = COALESCE(sqlc.narg(min_quantity), min_quantity),
    max_quantity = COALESCE(sqlc.narg(max_quantity), max_quantity),
    valid_from = COALESCE(sqlc.narg(valid_from), valid_from),
    valid_to = COALESCE(sqlc.narg(valid_to), valid_to),
    is_active = COALESCE(sqlc.narg(is_active), is_active),
    metadata = COALESCE(sqlc.narg(metadata), metadata)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteProductPrice :exec
DELETE FROM product_prices WHERE id = $1;

-- name: BulkUpdatePrices :exec
UPDATE product_prices
SET price = price * (1 + $2 / 100.0)
WHERE price_list_id = $1
  AND is_active = true;

-- name: ExpirePrices :exec
UPDATE product_prices
SET valid_to = CURRENT_DATE - INTERVAL '1 day'
WHERE price_list_id = $1
  AND is_active = true
  AND (valid_to IS NULL OR valid_to > CURRENT_DATE);

-- =====================================================
-- PRODUCT PRICING QUERIES
-- =====================================================

-- name: GetProductWithPricing :one
SELECT 
    p.*,
    json_agg(DISTINCT jsonb_build_object(
        'price_list_id', pl.id,
        'price_list_name', pl.name,
        'price_list_code', pl.code,
        'price', pp.price,
        'min_quantity', pp.min_quantity,
        'max_quantity', pp.max_quantity
    )) FILTER (WHERE pp.id IS NOT NULL) as prices
FROM products p
LEFT JOIN product_prices pp ON p.id = pp.product_id AND pp.is_active = true
LEFT JOIN price_lists pl ON pp.price_list_id = pl.id
WHERE p.id = $1
GROUP BY p.id;

-- name: SearchProductsWithPrices :many
SELECT 
    p.id,
    p.sku,
    p.name,
    p.description,
    pc.name as category_name,
    b.name as brand_name,
    pp.price,
    pl.name as price_list_name
FROM products p
LEFT JOIN product_categories pc ON p.category_id = pc.id
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN product_prices pp ON p.id = pp.product_id 
    AND pp.price_list_id = $2
    AND pp.is_active = true
LEFT JOIN price_lists pl ON pp.price_list_id = pl.id
WHERE p.organization_id = $1
  AND p.is_active = true
  AND p.is_sellable = true
  AND (
    p.sku ILIKE '%' || $3 || '%' OR
    p.name ILIKE '%' || $3 || '%'
  )
ORDER BY p.name
LIMIT $4 OFFSET $5;

-- name: GetPriceComparison :many
SELECT 
    pl.id as price_list_id,
    pl.name as price_list_name,
    pl.code as price_list_code,
    pp.price,
    pp.min_quantity,
    pp.valid_from,
    pp.valid_to
FROM price_lists pl
LEFT JOIN product_prices pp ON pl.id = pp.price_list_id 
    AND pp.product_id = $1
    AND pp.is_active = true
WHERE pl.is_active = true
ORDER BY pl.name;