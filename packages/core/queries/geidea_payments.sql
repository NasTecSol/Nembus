-- name: CreateGeideaPayment :one
INSERT INTO pos_payments (
    transaction_id,
    payment_method,
    payment_gateway,
    amount,
    status,
    ecr_reference_id,
    metadata,
    raw_response
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpdateGeideaPaymentApproved :one
UPDATE pos_payments
SET 
    status = 'approved',
    rrn = $2,
    auth_code = $3,
    tid = $4,
    mid = $5,
    card_scheme = $6,
    masked_pan = $7,
    aid = $8,
    raw_response = $9,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateGeideaPaymentStatus :one
UPDATE pos_payments
SET 
    status = $2,
    raw_response = COALESCE($3, raw_response),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetGeideaPaymentByEcrRef :one
SELECT * FROM pos_payments
WHERE ecr_reference_id = $1
LIMIT 1;
