-- name: CreateUnitOfMeasure :one
INSERT INTO units_of_measure (
    code,
    name,
    uom_type,
    decimal_places,
    is_active,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetUnitOfMeasure :one
SELECT * FROM units_of_measure
WHERE id = $1;

-- name: GetUnitOfMeasureByCode :one
SELECT * FROM units_of_measure
WHERE code = $1;

-- name: ListUnitsOfMeasure :many
SELECT * FROM units_of_measure
ORDER BY name;

-- name: ListActiveUnitsOfMeasure :many
SELECT * FROM units_of_measure
WHERE is_active = true
ORDER BY name;

-- name: ListUnitsByType :many
SELECT * FROM units_of_measure
WHERE uom_type = $1
ORDER BY name;

-- name: UpdateUnitOfMeasure :one
UPDATE units_of_measure
SET 
    name = $2,
    uom_type = $3,
    decimal_places = $4,
    is_active = $5,
    metadata = $6
WHERE id = $1
RETURNING *;

-- name: DeleteUnitOfMeasure :exec
DELETE FROM units_of_measure
WHERE id = $1;

-- name: CreateProductUOMConversion :one
INSERT INTO product_uom_conversions (
    product_id,
    from_uom_id,
    to_uom_id,
    conversion_factor,
    is_default,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetProductUOMConversion :one
SELECT * FROM product_uom_conversions
WHERE product_id = $1 AND from_uom_id = $2 AND to_uom_id = $3;

-- name: ListProductUOMConversions :many
SELECT * FROM product_uom_conversions
WHERE product_id = $1
ORDER BY from_uom_id, to_uom_id;

-- name: UpdateProductUOMConversion :one
UPDATE product_uom_conversions
SET 
    conversion_factor = $2,
    is_default = $3,
    metadata = $4
WHERE id = $1
RETURNING *;

-- name: DeleteProductUOMConversion :exec
DELETE FROM product_uom_conversions
WHERE id = $1;


-- ================================
-- PIPELINE & LOOKUP QUERIES
-- ================================

-- name: CreateUomPackagingTemplatesPipeline :exec
WITH inserted_templates AS (
    -- 1. Extract and insert all templates from the templates array
    INSERT INTO uom_packaging_templates (organization_id, uom_id, name, code, is_active)
    SELECT 
        ($1::jsonb->>'organization_id')::INTEGER,
        ($1::jsonb->>'uom_id')::INTEGER,
        t->>'template_name',
        t->>'template_code',
        COALESCE((t->>'template_active')::BOOLEAN, true)
    FROM jsonb_array_elements($1::jsonb->'templates') AS t
    ON CONFLICT (uom_id, name) DO UPDATE 
    SET code = EXCLUDED.code, 
        is_active = EXCLUDED.is_active,
        uom_id = EXCLUDED.uom_id
    RETURNING id, name
)
-- 2. Extract and insert levels for all templates dynamically
INSERT INTO uom_packaging_template_levels (template_id, level_order, uom_id, multiplier)
SELECT 
    it.id AS template_id,
    (lvl->>'level_order')::INTEGER AS level_order,
    -- Resolve UOM ID: lookup existing ID by code
    (SELECT id FROM units_of_measure WHERE code = lvl->>'uom_code' LIMIT 1) AS uom_id,
    (lvl->>'multiplier')::DECIMAL AS multiplier
FROM jsonb_array_elements($1::jsonb->'templates') AS t
JOIN inserted_templates it ON it.name = t->>'template_name'
CROSS JOIN LATERAL jsonb_array_elements(t->'levels') AS lvl
ON CONFLICT (template_id, level_order) DO UPDATE
SET uom_id = EXCLUDED.uom_id,
    multiplier = EXCLUDED.multiplier;



-- name: GetUomPackagingTemplatesByUomID :many
SELECT 
    t.id AS template_id,
    t.name AS template_name,
    t.code AS template_pattern_code,
    t.is_active AS template_is_active,
    -- Aggregates the levels in hierarchical order into a clean JSON array
    jsonb_agg(
        jsonb_build_object(
            'level_order', tl.level_order,
            'multiplier', tl.multiplier,
            'uom_id', u.id,
            'uom_code', u.code,
            'uom_name', u.name,
            'uom_type', u.uom_type,
            'decimal_places', u.decimal_places
        ) ORDER BY tl.level_order
    ) AS levels
FROM uom_packaging_templates t
JOIN uom_packaging_template_levels tl ON t.id = tl.template_id
JOIN units_of_measure u ON tl.uom_id = u.id
WHERE t.uom_id = $1
GROUP BY t.id, t.name, t.code, t.is_active;


-- name: GetProductUOMConversionsDetailed :many
SELECT 
    p.id AS product_id,
    p.sku AS product_sku,
    p.name AS product_name,
    pkg_uom.code AS packaging_uom_code,
    pkg_uom.name AS packaging_uom_name,
    target_uom.code AS target_uom_code,
    target_uom.name AS target_uom_name,
    puc.conversion_factor,
    puc.is_default AS is_default_packaging
FROM product_uom_conversions puc
JOIN products p ON puc.product_id = p.id
JOIN units_of_measure pkg_uom ON puc.from_uom_id = pkg_uom.id
JOIN units_of_measure target_uom ON puc.to_uom_id = target_uom.id
WHERE puc.product_id = $1;


