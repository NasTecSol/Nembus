-- name: CreateTaxCategory :one
INSERT INTO tax_categories (
    name,
    code,
    tax_rate,
    is_inclusive,
    is_active,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetTaxCategory :one
SELECT * FROM tax_categories
WHERE id = $1;

-- name: GetTaxCategoryByCode :one
SELECT * FROM tax_categories
WHERE code = $1;

-- name: ListTaxCategories :many
SELECT * FROM tax_categories
ORDER BY name;

-- name: ListActiveTaxCategories :many
SELECT * FROM tax_categories
WHERE is_active = true
ORDER BY name;

-- name: UpdateTaxCategory :one
UPDATE tax_categories
SET 
    name = $2,
    tax_rate = $3,
    is_inclusive = $4,
    is_active = $5,
    metadata = $6
WHERE id = $1
RETURNING *;

-- name: DeleteTaxCategory :exec
DELETE FROM tax_categories
WHERE id = $1;

-- name: ToggleTaxCategoryActive :one
UPDATE tax_categories
SET is_active = $2
WHERE id = $1
RETURNING *;
