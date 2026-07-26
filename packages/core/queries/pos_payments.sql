-- AddPaymentToTransaction stores a payment for a POS transaction.
-- payment_gateway: provider identifier (e.g. stripe, square). metadata: JSONB for transient provider data
-- (gateway_txn_id, masked_card, auth_code, etc.) for auditing and reconciliation.
-- name: AddPaymentToTransaction :exec
INSERT INTO pos_payments (
    transaction_id,
    payment_method,
    payment_gateway,
    amount,
    reference_number,
    metadata
) VALUES ($1, $2, $3, $4, $5, $6);

-- name: CreatePosPayment :one
INSERT INTO pos_payments (
    transaction_id,
    payment_method,
    payment_gateway,
    amount,
    payment_reference,
    reference_number,
    metadata
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetPosPayment :one
SELECT * FROM pos_payments
WHERE id = $1;

-- name: GetPaymentsForTransaction :many
SELECT 
    payment_method,
    payment_gateway,
    amount,
    reference_number,
    created_at
FROM pos_payments
WHERE transaction_id = $1
ORDER BY created_at;

-- name: GetPaymentsForTransactionFull :many
SELECT * FROM pos_payments
WHERE transaction_id = $1
ORDER BY created_at;

-- name: UpdatePosPayment :one
UPDATE pos_payments
SET
    payment_method    = $2,
    payment_gateway   = $3,
    amount            = $4,
    payment_reference = $5,
    reference_number  = $6,
    metadata          = $7
WHERE id = $1
RETURNING *;

-- name: DeletePosPayment :exec
DELETE FROM pos_payments
WHERE id = $1;

-- name: GetTransactionPaymentSummary :one
SELECT 
    COALESCE(SUM(amount), 0) AS total_paid,
    json_agg(
        json_build_object(
            'method', payment_method,
            'amount', amount,
            'ref', reference_number
        )
    ) AS payment_details
FROM pos_payments
WHERE transaction_id = $1;