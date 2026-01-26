-- name: OpenCashierSession :one
INSERT INTO cashier_sessions (
    cashier_id,
    pos_terminal_id,
    session_number,
    opening_time,
    opening_balance,
    status
) VALUES (
    $1, $2, $3, CURRENT_TIMESTAMP, $4, 'open'
) RETURNING id, session_number, opening_time, status;

-- name: GetActiveCashierSession :one
SELECT 
    cs.*,
    c.cashier_code,
    u.first_name || ' ' || u.last_name AS cashier_name,
    t.terminal_name,
    t.terminal_code
FROM cashier_sessions cs
JOIN cashiers      c ON cs.cashier_id      = c.id
JOIN users         u ON c.user_id          = u.id
JOIN pos_terminals t ON cs.pos_terminal_id = t.id
WHERE cs.cashier_id = $1
  AND cs.status = 'open'
  AND cs.closing_time IS NULL
ORDER BY cs.opening_time DESC
LIMIT 1;

-- name: CloseCashierSession :one
UPDATE cashier_sessions
SET 
    closing_time     = CURRENT_TIMESTAMP,
    closing_balance  = $2,
    expected_balance = $3,
    variance         = $4,
    status           = 'closed',
    metadata         = jsonb_set(
        jsonb_set(metadata, '{closing_note}', to_jsonb($5::text)),
        '{closed_by}', to_jsonb($6::bigint)
    )
WHERE id = $1
  AND status = 'open'
RETURNING id, session_number, opening_time, closing_time, variance, status;

-- name: GetSessionSummary :one
SELECT 
    cs.id,
    cs.session_number,
    cs.opening_time,
    cs.closing_time,
    cs.opening_balance,
    cs.closing_balance,
    cs.expected_balance,
    cs.variance,
    COUNT(t.id) AS transaction_count,
    COALESCE(SUM(t.total_amount), 0) AS total_sales,
    COALESCE(SUM(t.discount_amount), 0) AS total_discounts_given
FROM cashier_sessions cs
LEFT JOIN pos_transactions t 
    ON t.cashier_session_id = cs.id 
   AND t.status = 'completed'
WHERE cs.id = $1
GROUP BY cs.id;