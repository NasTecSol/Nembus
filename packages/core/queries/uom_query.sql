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

-- name: CreateUomPackagingTemplatePipeline :one
WITH new_uom AS (
    -- 1. Insert the Base UOM from the Left Panel
    INSERT INTO units_of_measure (code, name, uom_type, decimal_places, is_active)
    VALUES ($1, $2, $3, $4, $5)
    ON CONFLICT (code) DO UPDATE 
    SET name = EXCLUDED.name, uom_type = EXCLUDED.uom_type, decimal_places = EXCLUDED.decimal_places, is_active = EXCLUDED.is_active
    RETURNING id
),
new_template AS (
    -- 2. Insert the Packaging Template Container
    INSERT INTO uom_packaging_templates (organization_id, name, code, is_active)
    VALUES ($6, $7, $8, $9)
    ON CONFLICT (code) DO UPDATE 
    SET name = EXCLUDED.name, is_active = EXCLUDED.is_active
    RETURNING id
),
inserted_levels AS (
    -- 3. Insert the 3 levels of your pipeline
    INSERT INTO uom_packaging_template_levels (template_id, level_order, uom_id, multiplier)
    VALUES 
        -- Tier 1 (Base Level using the new KG ID)
        ((SELECT id FROM new_template), 1, (SELECT id FROM new_uom), $10),
        
        -- Tier 2 (PKT Level - looks up the database ID for PKT)
        ((SELECT id FROM new_template), 2, (SELECT id FROM units_of_measure WHERE units_of_measure.code = $11 LIMIT 1), $12),
        
        -- Tier 3 (CAN Level - looks up the database ID for CAN)
        ((SELECT id FROM new_template), 3, (SELECT id FROM units_of_measure WHERE units_of_measure.code = $13 LIMIT 1), $14)
    RETURNING template_id
)
SELECT 
    (SELECT id FROM new_template) AS template_id,
    (SELECT id FROM new_uom) AS base_uom_id;



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

