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