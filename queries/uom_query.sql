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
