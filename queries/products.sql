-- =====================================================
-- PRODUCT CATEGORIES
-- =====================================================

-- name: CreateProductCategory :one
INSERT INTO product_categories (
    parent_category_id, name, code, description, 
    category_level, is_active, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetProductCategory :one
SELECT * FROM product_categories WHERE id = $1 LIMIT 1;

-- name: GetProductCategoryByCode :one
SELECT * FROM product_categories WHERE code = $1 LIMIT 1;

-- name: ListProductCategories :many
SELECT * FROM product_categories
WHERE is_active = COALESCE(sqlc.narg(is_active), is_active)
  AND parent_category_id IS NULL
ORDER BY name;

-- name: ListCategoryChildren :many
SELECT * FROM product_categories
WHERE parent_category_id = $1
  AND is_active = COALESCE(sqlc.narg(is_active), is_active)
ORDER BY name;

-- name: GetCategoryHierarchy :many
WITH RECURSIVE category_tree AS (
    SELECT 
        id, parent_category_id, name, code, description,
        category_level, is_active, metadata,
        1 as level,
        ARRAY[id] as path,
        name as full_path
    FROM product_categories
    WHERE parent_category_id IS NULL
    
    UNION ALL
    
    SELECT 
        pc.id, pc.parent_category_id, pc.name, pc.code, pc.description,
        pc.category_level, pc.is_active, pc.metadata,
        ct.level + 1,
        ct.path || pc.id,
        ct.full_path || ' > ' || pc.name
    FROM product_categories pc
    INNER JOIN category_tree ct ON pc.parent_category_id = ct.id
)
SELECT * FROM category_tree ct
WHERE CASE 
    WHEN sqlc.narg(filter_is_active) IS NULL THEN true
    ELSE ct.is_active = sqlc.narg(filter_is_active)
END
ORDER BY ct.path;

-- name: UpdateProductCategory :one
UPDATE product_categories
SET 
    parent_category_id = COALESCE(sqlc.narg(parent_category_id), parent_category_id),
    name = COALESCE(sqlc.narg(name), name),
    description = COALESCE(sqlc.narg(description), description),
    category_level = COALESCE(sqlc.narg(category_level), category_level),
    is_active = COALESCE(sqlc.narg(is_active), is_active),
    metadata = COALESCE(sqlc.narg(metadata), metadata)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteProductCategory :exec
DELETE FROM product_categories WHERE id = $1;

-- =====================================================
-- BRANDS
-- =====================================================

-- name: CreateBrand :one
INSERT INTO brands (
    name, code, is_active, metadata
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetBrand :one
SELECT * FROM brands WHERE id = $1 LIMIT 1;

-- name: GetBrandByCode :one
SELECT * FROM brands WHERE code = $1 LIMIT 1;

-- name: ListBrands :many
SELECT * FROM brands
WHERE is_active = COALESCE(sqlc.narg(is_active), is_active)
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: SearchBrands :many
SELECT * FROM brands
WHERE is_active = true
  AND (name ILIKE '%' || $1 || '%' OR code ILIKE '%' || $1 || '%')
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: UpdateBrand :one
UPDATE brands
SET 
    name = COALESCE(sqlc.narg(name), name),
    is_active = COALESCE(sqlc.narg(is_active), is_active),
    metadata = COALESCE(sqlc.narg(metadata), metadata)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteBrand :exec
DELETE FROM brands WHERE id = $1;

-- =====================================================
-- UNITS OF MEASURE
-- Note: UOM queries are in uom_query.sql
-- =====================================================

-- =====================================================
-- TAX CATEGORIES
-- Note: Tax category queries are in tax_categories_query.sql
-- =====================================================

-- =====================================================
-- PRODUCTS
-- =====================================================

-- name: CreateProduct :one
INSERT INTO products (
    organization_id, sku, name, description, category_id,
    brand_id, base_uom_id, product_type, tax_category_id,
    is_serialized, is_batch_managed, is_active, is_sellable,
    is_purchasable, allow_decimal_quantity, track_inventory, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
) RETURNING *;

-- name: GetProduct :one
SELECT * FROM products WHERE id = $1 LIMIT 1;

-- name: GetProductBySKU :one
SELECT * FROM products 
WHERE organization_id = $1 AND sku = $2 
LIMIT 1;

-- name: ListProducts :many
SELECT * FROM products
WHERE organization_id = $1
  AND is_active = COALESCE(sqlc.narg(is_active), is_active)
  AND category_id = COALESCE(sqlc.narg(category_id), category_id)
  AND brand_id = COALESCE(sqlc.narg(brand_id), brand_id)
  AND product_type = COALESCE(sqlc.narg(product_type), product_type)
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: SearchProducts :many
SELECT * FROM products
WHERE organization_id = $1
  AND is_active = true
  AND (
    sku ILIKE '%' || $2 || '%' OR
    name ILIKE '%' || $2 || '%' OR
    description ILIKE '%' || $2 || '%'
  )
ORDER BY name
LIMIT $3 OFFSET $4;

-- name: ListSellableProducts :many
SELECT * FROM products
WHERE organization_id = $1
  AND is_sellable = true
  AND is_active = true
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: ListPurchasableProducts :many
SELECT * FROM products
WHERE organization_id = $1
  AND is_purchasable = true
  AND is_active = true
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: GetProductWithDetails :one
SELECT 
    p.*,
    pc.name as category_name,
    pc.code as category_code,
    b.name as brand_name,
    b.code as brand_code,
    uom.name as base_uom_name,
    uom.code as base_uom_code,
    tc.name as tax_category_name,
    tc.tax_rate as tax_rate
FROM products p
LEFT JOIN product_categories pc ON p.category_id = pc.id
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN units_of_measure uom ON p.base_uom_id = uom.id
LEFT JOIN tax_categories tc ON p.tax_category_id = tc.id
WHERE p.id = $1
LIMIT 1;

-- name: UpdateProduct :one
UPDATE products
SET 
    name = COALESCE(sqlc.narg(name), name),
    description = COALESCE(sqlc.narg(description), description),
    category_id = COALESCE(sqlc.narg(category_id), category_id),
    brand_id = COALESCE(sqlc.narg(brand_id), brand_id),
    base_uom_id = COALESCE(sqlc.narg(base_uom_id), base_uom_id),
    product_type = COALESCE(sqlc.narg(product_type), product_type),
    tax_category_id = COALESCE(sqlc.narg(tax_category_id), tax_category_id),
    is_serialized = COALESCE(sqlc.narg(is_serialized), is_serialized),
    is_batch_managed = COALESCE(sqlc.narg(is_batch_managed), is_batch_managed),
    is_active = COALESCE(sqlc.narg(is_active), is_active),
    is_sellable = COALESCE(sqlc.narg(is_sellable), is_sellable),
    is_purchasable = COALESCE(sqlc.narg(is_purchasable), is_purchasable),
    allow_decimal_quantity = COALESCE(sqlc.narg(allow_decimal_quantity), allow_decimal_quantity),
    track_inventory = COALESCE(sqlc.narg(track_inventory), track_inventory),
    metadata = COALESCE(sqlc.narg(metadata), metadata)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1;

-- name: CountProducts :one
SELECT COUNT(*) FROM products
WHERE organization_id = $1
  AND is_active = COALESCE(sqlc.narg(is_active), is_active);

-- =====================================================
-- PRODUCT VARIANTS
-- Note: Product variant queries are in product_variants_query.sql
-- =====================================================

-- =====================================================
-- PRODUCT BARCODES
-- Note: Product barcode queries are in product_barcodes_query.sql
-- =====================================================

-- =====================================================
-- PRODUCT UOM CONVERSIONS
-- Note: Product UOM conversion queries are in uom_query.sql
-- =====================================================