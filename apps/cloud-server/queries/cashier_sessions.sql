-- name: OpenCashierSession :one
INSERT INTO cashier_sessions (
    cashier_id,
    pos_terminal_id,
    session_number,
    opening_time,
    opening_balance,
    expected_balance,
    status
) VALUES (
    $1, $2, $3, CURRENT_TIMESTAMP, $4, $4, 'open'
) RETURNING id, session_number, opening_time, expected_balance, status;

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
    COALESCE(cs.expected_balance, cs.opening_balance, 0) AS expected_balance,
    cs.variance,
    COUNT(t.id) AS transaction_count,
    COALESCE(SUM(t.total_amount), 0) AS total_sales,
    COALESCE(SUM(t.total_amount), 0) AS cash_sales,
    COALESCE(SUM(t.discount_amount), 0) AS total_discounts_given
FROM cashier_sessions cs
LEFT JOIN pos_transactions t 
    ON t.cashier_session_id = cs.id 
   AND t.status = 'completed'
WHERE cs.id = $1
GROUP BY cs.id;

-- name: UpdateSessionExpectedBalance :exec
UPDATE cashier_sessions
SET expected_balance = COALESCE(expected_balance, opening_balance) + $2
WHERE id = $1;

-- name: CloseCashierSessionReconcile :one
-- Close session and set variance = physical closing_balance - expected_balance (reconciliation at shift end)
UPDATE cashier_sessions
SET 
    closing_time     = CURRENT_TIMESTAMP,
    closing_balance  = $2,
    variance         = $2 - COALESCE(expected_balance, opening_balance),
    expected_balance = COALESCE(expected_balance, opening_balance),
    status           = 'closed',
    metadata         = jsonb_set(
        jsonb_set(COALESCE(metadata, '{}'), '{closing_note}', to_jsonb($3::text)),
        '{closed_by}', to_jsonb($4::bigint)
    )
WHERE id = $1
  AND status = 'open'
RETURNING id, session_number, opening_time, closing_time, expected_balance, variance, status;
-- name: GetClosedCashierSessionsByDateRange :many
SELECT 
    cs.*,
    c.cashier_code,
    u.first_name || ' ' || u.last_name AS cashier_name,
    t.terminal_name,
    t.terminal_code
FROM cashier_sessions cs
LEFT JOIN cashiers      c ON cs.cashier_id      = c.id
LEFT JOIN users         u ON c.user_id          = u.id
LEFT JOIN pos_terminals t ON cs.pos_terminal_id = t.id
WHERE cs.cashier_id = $1
  AND (LOWER(cs.status) = 'closed' OR cs.status IS NULL)
  AND cs.closing_time >= $2
  AND cs.closing_time <= $3
ORDER BY cs.closing_time DESC;