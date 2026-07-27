-- ================================
-- UOM PACKAGING TEMPLATES
-- ================================

-- name: GetUomPackagingTemplate :one
SELECT *
FROM uom_packaging_templates
WHERE id = $1
LIMIT 1;


-- name: ListUomPackagingTemplates :many
SELECT *
FROM uom_packaging_templates
WHERE organization_id = $1
ORDER BY created_at DESC;


-- name: CreateUomPackagingTemplate :one
INSERT INTO uom_packaging_templates (
    organization_id,
    name,
    code,
    is_active
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;


-- name: UpdateUomPackagingTemplate :one
UPDATE uom_packaging_templates
SET
    name = $2,
    code = $3,
    is_active = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;


-- name: DeleteUomPackagingTemplate :exec
DELETE FROM uom_packaging_templates
WHERE id = $1;



-- ================================
-- UOM PACKAGING TEMPLATE LEVELS
-- ================================

-- name: GetUomPackagingTemplateLevel :one
SELECT *
FROM uom_packaging_template_levels
WHERE id = $1
LIMIT 1;


-- name: ListUomPackagingTemplateLevels :many
SELECT *
FROM uom_packaging_template_levels
WHERE template_id = $1
ORDER BY level_order;


-- name: CreateUomPackagingTemplateLevel :one
INSERT INTO uom_packaging_template_levels (
    template_id,
    level_order,
    uom_id,
    multiplier
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;


-- name: UpdateUomPackagingTemplateLevel :one
UPDATE uom_packaging_template_levels
SET
    level_order = $2,
    uom_id = $3,
    multiplier = $4
WHERE id = $1
RETURNING *;


-- name: DeleteUomPackagingTemplateLevel :exec
DELETE FROM uom_packaging_template_levels
WHERE id = $1;



-- ================================
-- HELPER QUERY (TEMPLATE + LEVELS)
-- ================================

-- name: GetPackagingTemplateWithLevels :many
SELECT 
    t.id AS template_id,
    t.name,
    t.code,
    t.organization_id,
    t.is_active,
    l.id AS level_id,
    l.level_order,
    l.uom_id,
    l.multiplier
FROM uom_packaging_templates t
LEFT JOIN uom_packaging_template_levels l
    ON t.id = l.template_id
WHERE t.id = $1
ORDER BY l.level_order;