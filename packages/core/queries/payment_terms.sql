-- name: CreatePaymentTerm :one
INSERT INTO payment_terms (
    organization_id, code, name, due_days, discount_days, 
    discount_percentage, late_fee_percentage, is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetPaymentTerm :one
SELECT * FROM payment_terms 
WHERE id = $1 LIMIT 1;

-- name: GetPaymentTermByCode :one
SELECT * FROM payment_terms 
WHERE organization_id = $1 AND code = $2 LIMIT 1;

-- name: ListPaymentTerms :many
SELECT * FROM payment_terms
WHERE organization_id = $1
  AND (sqlc.narg(is_active)::boolean IS NULL OR is_active = sqlc.narg(is_active))
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: UpdatePaymentTerm :one
UPDATE payment_terms
SET 
    code = COALESCE(sqlc.narg(code), code),
    name = COALESCE(sqlc.narg(name), name),
    due_days = COALESCE(sqlc.narg(due_days), due_days),
    discount_days = COALESCE(sqlc.narg(discount_days), discount_days),
    discount_percentage = COALESCE(sqlc.narg(discount_percentage), discount_percentage),
    late_fee_percentage = COALESCE(sqlc.narg(late_fee_percentage), late_fee_percentage),
    is_active = COALESCE(sqlc.narg(is_active), is_active)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeletePaymentTerm :exec
DELETE FROM payment_terms WHERE id = $1;

-- name: CountPaymentTerms :one
SELECT COUNT(*) FROM payment_terms
WHERE organization_id = $1
  AND (sqlc.narg(is_active)::boolean IS NULL OR is_active = sqlc.narg(is_active));
