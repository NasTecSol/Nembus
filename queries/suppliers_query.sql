-- name: CreateSupplier :one
INSERT INTO suppliers (
    organization_id,
    name,
    code,
    supplier_type,
    payment_terms,
    credit_limit,
    is_active,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetSupplier :one
SELECT * FROM suppliers
WHERE id = $1;

-- name: GetSupplierByCode :one
SELECT * FROM suppliers
WHERE organization_id = $1 AND code = $2;

-- name: ListSuppliers :many
SELECT * FROM suppliers
WHERE organization_id = $1
ORDER BY name;

-- name: ListActiveSuppliers :many
SELECT * FROM suppliers
WHERE organization_id = $1 AND is_active = true
ORDER BY name;

-- name: ListSuppliersByType :many
SELECT * FROM suppliers
WHERE organization_id = $1 AND supplier_type = $2
ORDER BY name;

-- name: SearchSuppliers :many
SELECT * FROM suppliers
WHERE organization_id = $1 
  AND (name ILIKE $2 OR code ILIKE $2)
ORDER BY name
LIMIT $3;

-- name: UpdateSupplier :one
UPDATE suppliers
SET 
    name = $2,
    supplier_type = $3,
    payment_terms = $4,
    credit_limit = $5,
    is_active = $6,
    metadata = $7
WHERE id = $1
RETURNING *;

-- name: DeleteSupplier :exec
DELETE FROM suppliers
WHERE id = $1;

-- name: ToggleSupplierActive :one
UPDATE suppliers
SET is_active = $2
WHERE id = $1
RETURNING *;
