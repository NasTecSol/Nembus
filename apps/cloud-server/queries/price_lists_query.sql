-- name: CreatePriceList :one
INSERT INTO price_lists (
    name,
    code,
    price_list_type,
    currency_code,
    valid_from,
    valid_to,
    is_default,
    is_active,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetPriceList :one
SELECT * FROM price_lists
WHERE id = $1;

-- name: GetPriceListByCode :one
SELECT * FROM price_lists
WHERE code = $1;

-- name: ListPriceLists :many
SELECT * FROM price_lists
ORDER BY name;

-- name: ListActivePriceLists :many
SELECT * FROM price_lists
WHERE is_active = true
ORDER BY name;

-- name: ListValidPriceLists :many
SELECT * FROM price_lists
WHERE is_active = true
  AND (valid_from IS NULL OR valid_from <= CURRENT_DATE)
  AND (valid_to IS NULL OR valid_to >= CURRENT_DATE)
ORDER BY name;

-- name: GetDefaultPriceList :one
SELECT * FROM price_lists
WHERE is_default = true AND is_active = true
LIMIT 1;

-- name: UpdatePriceList :one
UPDATE price_lists
SET 
    name = $2,
    price_list_type = $3,
    currency_code = $4,
    valid_from = $5,
    valid_to = $6,
    is_default = $7,
    is_active = $8,
    metadata = $9
WHERE id = $1
RETURNING *;

-- name: SetDefaultPriceList :exec
UPDATE price_lists
SET is_default = (id = $1);

-- name: DeletePriceList :exec
DELETE FROM price_lists
WHERE id = $1;

-- name: TogglePriceListActive :one
UPDATE price_lists
SET is_active = $2
WHERE id = $1
RETURNING *;
