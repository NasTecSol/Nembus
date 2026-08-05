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
    INSERT INTO uom_packaging_templates (organization_id, name, code, is_active)
    SELECT 
        ($1::jsonb->>'organization_id')::INTEGER,
        t->>'template_name',
        t->>'template_code',
        COALESCE((t->>'template_active')::BOOLEAN, true)
    FROM jsonb_array_elements($1::jsonb->'templates') AS t
    ON CONFLICT (code) DO UPDATE 
    SET name = EXCLUDED.name, 
        is_active = EXCLUDED.is_active
    RETURNING id, code
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
JOIN inserted_templates it ON it.code = t->>'template_code'
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
WHERE t.id IN (
    -- Subquery: Finds any template ID where the input UOM ID is used
    SELECT DISTINCT uom_packaging_template_levels.template_id 
    FROM uom_packaging_template_levels 
    WHERE uom_packaging_template_levels.uom_id = $1
)
GROUP BY t.id, t.name, t.code, t.is_active;

