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
        name::text as full_path
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
    WHEN sqlc.narg(filter_is_active)::boolean IS NULL THEN true
    ELSE ct.is_active = sqlc.narg(filter_is_active)::boolean
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
-- Note: Brand queries are in brands.sql
-- =====================================================

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

-- =====================================================
-- PRODUCT CATALOG (Admin: products + embedded variants)
-- =====================================================

-- name: ListProductsWithVariants :many
SELECT 
    p.id, 
    p.sku, 
    p.name, 
    p.description, 
    p.is_active,
    pc.name AS category_name,
    b.name AS brand_name,
    COALESCE(
        json_agg(
            DISTINCT jsonb_build_object(
                'id', pv.id,
                'variant_sku', pv.variant_sku,
                'variant_name', pv.variant_name,
                'variant_attributes', pv.variant_attributes,
                'is_active', pv.is_active
            )
        ) FILTER (WHERE pv.id IS NOT NULL), 
        '[]'
    )::jsonb AS variants,
    COALESCE(
        json_agg(
            DISTINCT jsonb_build_object(
                'stock_id', ist.id,
                'store_id', ist.store_id,
                'storage_location_id', ist.storage_location_id,
                'storage_location_name', sl.name,
                'storage_location_code', sl.code,
                'product_variant_id', ist.product_variant_id,
                'quantity_on_hand', ist.quantity_on_hand,
                'quantity_available', ist.quantity_available,
                'quantity_allocated', ist.quantity_allocated,
                'quantity_on_order', ist.quantity_on_order,
                'reorder_level', ist.reorder_level,
                'reorder_quantity', ist.reorder_quantity,
                'max_stock_level', ist.max_stock_level
            )
        ) FILTER (WHERE ist.id IS NOT NULL),
        '[]'
    )::jsonb AS inventory
FROM products p
LEFT JOIN product_categories pc ON p.category_id = pc.id
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN product_variants pv ON p.id = pv.product_id
LEFT JOIN inventory_stock ist ON p.id = ist.product_id
LEFT JOIN storage_locations sl ON ist.storage_location_id = sl.id
WHERE p.organization_id = $1
  AND ($2 = 0 OR p.category_id = $2)
GROUP BY p.id, pc.name, b.name
ORDER BY p.name
LIMIT $3 OFFSET $4;
