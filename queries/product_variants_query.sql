-- name: CreateProductVariant :one
INSERT INTO product_variants (
    product_id,
    variant_sku,
    variant_name,
    variant_attributes,
    is_active,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetProductVariant :one
SELECT * FROM product_variants
WHERE id = $1;

-- name: GetProductVariantBySKU :one
SELECT * FROM product_variants
WHERE variant_sku = $1;

-- name: ListProductVariants :many
SELECT * FROM product_variants
ORDER BY variant_sku;

-- name: ListProductVariantsByProduct :many
SELECT * FROM product_variants
WHERE product_id = $1
ORDER BY variant_sku;

-- name: ListActiveProductVariantsByProduct :many
SELECT * FROM product_variants
WHERE product_id = $1 AND is_active = true
ORDER BY variant_sku;

-- name: SearchProductVariants :many
SELECT pv.*, p.name AS product_name, p.sku
FROM product_variants pv
JOIN products p ON pv.product_id = p.id
WHERE pv.variant_sku ILIKE $1 OR pv.variant_name ILIKE $1 OR p.name ILIKE $1
ORDER BY pv.variant_sku
LIMIT $2;

-- name: UpdateProductVariant :one
UPDATE product_variants
SET 
    variant_name = $2,
    variant_attributes = $3,
    is_active = $4,
    metadata = $5
WHERE id = $1
RETURNING *;

-- name: DeleteProductVariant :exec
DELETE FROM product_variants
WHERE id = $1;

-- name: ToggleProductVariantActive :one
UPDATE product_variants
SET is_active = $2
WHERE id = $1
RETURNING *;
