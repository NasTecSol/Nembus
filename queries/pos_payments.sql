-- name: AddPaymentToTransaction :exec
INSERT INTO pos_payments (
    transaction_id,
    payment_method,
    amount,
    reference_number,
    metadata
) VALUES ($1, $2, $3, $4, $5);

-- name: GetPaymentsForTransaction :many
SELECT 
    payment_method,
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