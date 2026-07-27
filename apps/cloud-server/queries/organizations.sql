-- name: CreateOrganization :one
INSERT INTO organizations (
    name, code, legal_name, tax_id, currency_code, 
    fiscal_year_variant, is_active, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetOrganization :one
SELECT * FROM organizations WHERE id = $1 LIMIT 1;

-- name: GetOrganizationByCode :one
SELECT * FROM organizations WHERE code = $1 LIMIT 1;

-- name: ListOrganizations :many
SELECT * FROM organizations
WHERE is_active = COALESCE(sqlc.narg(is_active), is_active)
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: UpdateOrganization :one
UPDATE organizations
SET 
    name = COALESCE(sqlc.narg(name), name),
    legal_name = COALESCE(sqlc.narg(legal_name), legal_name),
    tax_id = COALESCE(sqlc.narg(tax_id), tax_id),
    currency_code = COALESCE(sqlc.narg(currency_code), currency_code),
    fiscal_year_variant = COALESCE(sqlc.narg(fiscal_year_variant), fiscal_year_variant),
    is_active = COALESCE(sqlc.narg(is_active), is_active),
    metadata = COALESCE(sqlc.narg(metadata), metadata)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteOrganization :exec
DELETE FROM organizations WHERE id = $1;

-- name: CountOrganizations :one
SELECT COUNT(*) FROM organizations
WHERE is_active = COALESCE(sqlc.narg(is_active), is_active);