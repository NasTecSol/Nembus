-- =====================================================
-- CASHIERS QUERIES (SQLC)
-- =====================================================
-- File: queries/cashiers.sql
-- Purpose: CRUD operations and essential POS backend queries for cashiers

-- =====================================================
-- CREATE OPERATIONS
-- =====================================================

-- name: CreateCashier :one
-- Create a new cashier
INSERT INTO cashiers (
    user_id,
    store_id,
    cashier_code,
    drawer_limit,
    discount_limit,
    is_active,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: CreateCashierWithDefaults :one
-- Create a new cashier with default values
INSERT INTO cashiers (
    user_id,
    store_id,
    cashier_code,
    is_active
) VALUES (
    $1, $2, $3, true
)
RETURNING *;

-- =====================================================
-- READ OPERATIONS
-- =====================================================

-- name: GetCashierByID :one
-- Get a single cashier by ID
SELECT * FROM cashiers
WHERE id = $1;

-- name: GetCashierByCode :one
-- Get a cashier by cashier_code and store_id
SELECT * FROM cashiers
WHERE cashier_code = $1 AND store_id = $2;

-- name: GetCashierByUserID :one
-- Get a cashier by user_id (one user can be one cashier per store)
SELECT * FROM cashiers
WHERE user_id = $1 AND store_id = $2;

-- name: ListAllCashiers :many
-- List all cashiers
SELECT * FROM cashiers
ORDER BY cashier_code;

-- name: ListActiveCashiers :many
-- List only active cashiers
SELECT * FROM cashiers
WHERE is_active = true
ORDER BY cashier_code;

-- name: ListCashiersByStore :many
-- List all cashiers for a specific store
SELECT * FROM cashiers
WHERE store_id = $1
ORDER BY cashier_code;

-- name: ListActiveCashiersByStore :many
-- List active cashiers for a specific store
SELECT * FROM cashiers
WHERE store_id = $1 AND is_active = true
ORDER BY cashier_code;

-- name: ListCashiersWithPagination :many
-- List cashiers with pagination
SELECT * FROM cashiers
ORDER BY cashier_code
LIMIT $1 OFFSET $2;

-- name: CountCashiers :one
-- Count total number of cashiers
SELECT COUNT(*) FROM cashiers;

-- name: CountActiveCashiers :one
-- Count active cashiers
SELECT COUNT(*) FROM cashiers
WHERE is_active = true;

-- name: CountCashiersByStore :one
-- Count cashiers in a specific store
SELECT COUNT(*) FROM cashiers
WHERE store_id = $1;

-- =====================================================
-- UPDATE OPERATIONS
-- =====================================================

-- name: UpdateCashier :one
-- Update cashier information
UPDATE cashiers
SET
    user_id = $2,
    store_id = $3,
    cashier_code = $4,
    drawer_limit = $5,
    discount_limit = $6,
    is_active = $7,
    metadata = $8
WHERE id = $1
RETURNING *;

-- name: UpdateCashierLimits :one
-- Update cashier drawer and discount limits
UPDATE cashiers
SET
    drawer_limit = $2,
    discount_limit = $3
WHERE id = $1
RETURNING *;

-- name: UpdateCashierDrawerLimit :one
-- Update only the drawer limit
UPDATE cashiers
SET drawer_limit = $2
WHERE id = $1
RETURNING *;

-- name: UpdateCashierDiscountLimit :one
-- Update only the discount limit
UPDATE cashiers
SET discount_limit = $2
WHERE id = $1
RETURNING *;

-- name: UpdateCashierMetadata :one
-- Update cashier metadata
UPDATE cashiers
SET metadata = $2
WHERE id = $1
RETURNING *;

-- name: ActivateCashier :one
-- Activate a cashier
UPDATE cashiers
SET is_active = true
WHERE id = $1
RETURNING *;

-- name: DeactivateCashier :one
-- Deactivate a cashier
UPDATE cashiers
SET is_active = false
WHERE id = $1
RETURNING *;

-- =====================================================
-- DELETE OPERATIONS
-- =====================================================

-- name: DeleteCashier :exec
-- Hard delete a cashier (use with caution)
DELETE FROM cashiers
WHERE id = $1;

-- name: SoftDeleteCashier :one
-- Soft delete by deactivating
UPDATE cashiers
SET is_active = false
WHERE id = $1
RETURNING *;

-- =====================================================
-- VALIDATION & EXISTENCE CHECKS
-- =====================================================

-- name: CashierExists :one
-- Check if a cashier exists by ID
SELECT EXISTS(
    SELECT 1 FROM cashiers WHERE id = $1
);

-- name: CashierCodeExists :one
-- Check if cashier code exists for a store
SELECT EXISTS(
    SELECT 1 FROM cashiers 
    WHERE cashier_code = $1 AND store_id = $2
);

-- name: CashierCodeExistsExcludingID :one
-- Check if cashier code exists excluding specific cashier ID (for updates)
SELECT EXISTS(
    SELECT 1 FROM cashiers 
    WHERE cashier_code = $1 AND store_id = $2 AND id != $3
);

-- =====================================================
-- SESSION LIFECYCLE QUERIES
-- =====================================================

-- name: GetActiveSessionByCashier :one
-- Get the currently open session for a specific cashier
-- Prevents double-opening sessions on different terminals
SELECT * FROM cashier_sessions
WHERE cashier_id = $1 
  AND status = 'open'
ORDER BY opening_time DESC
LIMIT 1;

-- name: GetActiveSessionByCashierAndTerminal :one
-- Get active session for specific cashier and terminal
SELECT * FROM cashier_sessions
WHERE cashier_id = $1 
  AND pos_terminal_id = $2
  AND status = 'open'
ORDER BY opening_time DESC
LIMIT 1;

-- name: CreateCashierSession :one
-- Create a new cashier session
INSERT INTO cashier_sessions (
    cashier_id,
    pos_terminal_id,
    session_number,
    opening_time,
    opening_balance,
    status,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, 'open', $6
)
RETURNING *;

-- Note: CloseCashierSession query is in cashier_sessions.sql to avoid duplication

-- name: GetSessionByID :one
-- Get a specific cashier session by ID
SELECT * FROM cashier_sessions
WHERE id = $1;

-- name: GetSessionByNumber :one
-- Get a session by session number
SELECT * FROM cashier_sessions
WHERE session_number = $1;

-- name: ListCashierSessions :many
-- List all sessions for a cashier with pagination
SELECT * FROM cashier_sessions
WHERE cashier_id = $1
ORDER BY opening_time DESC
LIMIT $2 OFFSET $3;

-- name: ListSessionsByDateRange :many
-- Get sessions within a date range
SELECT * FROM cashier_sessions
WHERE cashier_id = $1
  AND opening_time::date BETWEEN $2 AND $3
ORDER BY opening_time DESC;

-- =====================================================
-- CALCULATE SESSION EXPECTED BALANCE
-- =====================================================

-- name: CalculateSessionExpectedBalance :one
-- Calculate expected balance for a cashier session
-- Formula: opening_balance + cash_payments - change_given + drawer_actions
SELECT 
    cs.id,
    cs.opening_balance,
    COALESCE(SUM(CASE WHEN pp.payment_method = 'cash' THEN pp.amount ELSE 0 END), 0) as total_cash_payments,
    COALESCE(SUM(pt.change_given), 0) as total_change_given,
    (cs.opening_balance + 
     COALESCE(SUM(CASE WHEN pp.payment_method = 'cash' THEN pp.amount ELSE 0 END), 0) - 
     COALESCE(SUM(pt.change_given), 0)) as expected_balance
FROM cashier_sessions cs
LEFT JOIN pos_transactions pt ON pt.cashier_session_id = cs.id AND pt.status != 'voided'
LEFT JOIN pos_payments pp ON pp.transaction_id = pt.id
WHERE cs.id = $1
GROUP BY cs.id, cs.opening_balance;

-- name: GetSessionFinancialSummary :one
-- Get comprehensive financial summary for a session
SELECT 
    cs.id,
    cs.session_number,
    cs.opening_balance,
    cs.closing_balance,
    cs.expected_balance,
    cs.variance,
    COUNT(DISTINCT pt.id) as transaction_count,
    COALESCE(SUM(pt.subtotal), 0) as total_subtotal,
    COALESCE(SUM(pt.discount_amount), 0) as total_discounts,
    COALESCE(SUM(pt.tax_amount), 0) as total_tax,
    COALESCE(SUM(pt.total_amount), 0) as total_sales,
    COALESCE(SUM(CASE WHEN pp.payment_method = 'cash' THEN pp.amount ELSE 0 END), 0) as cash_payments,
    COALESCE(SUM(CASE WHEN pp.payment_method = 'card' THEN pp.amount ELSE 0 END), 0) as card_payments,
    COALESCE(SUM(CASE WHEN pp.payment_method = 'mobile' THEN pp.amount ELSE 0 END), 0) as mobile_payments,
    COALESCE(SUM(pt.change_given), 0) as total_change_given
FROM cashier_sessions cs
LEFT JOIN pos_transactions pt ON pt.cashier_session_id = cs.id AND pt.status != 'voided'
LEFT JOIN pos_payments pp ON pp.transaction_id = pt.id
WHERE cs.id = $1
GROUP BY cs.id, cs.session_number, cs.opening_balance, cs.closing_balance, cs.expected_balance, cs.variance;

-- =====================================================
-- TRANSACTION & PERFORMANCE QUERIES
-- =====================================================

-- name: GetSessionTransactionSummary :one
-- Get transaction summary for X-Report (mid-shift summary)
SELECT 
    COUNT(pt.id) as total_transactions,
    COALESCE(SUM(pt.subtotal), 0) as gross_sales,
    COALESCE(SUM(pt.discount_amount), 0) as total_discounts,
    COALESCE(SUM(pt.tax_amount), 0) as total_tax,
    COALESCE(SUM(pt.total_amount), 0) as net_sales,
    COALESCE(AVG(pt.total_amount), 0) as average_transaction_value,
    COUNT(CASE WHEN pt.status = 'voided' THEN 1 END) as voided_count,
    COALESCE(SUM(CASE WHEN pt.status = 'voided' THEN pt.total_amount ELSE 0 END), 0) as voided_amount
FROM pos_transactions pt
WHERE pt.cashier_session_id = $1;

-- name: GetSessionPaymentMethodBreakdown :many
-- Get payment method breakdown for a session
SELECT 
    pp.payment_method,
    COUNT(DISTINCT pp.transaction_id) as transaction_count,
    COALESCE(SUM(pp.amount), 0) as total_amount
FROM pos_payments pp
INNER JOIN pos_transactions pt ON pt.id = pp.transaction_id
WHERE pt.cashier_session_id = $1
  AND pt.status != 'voided'
GROUP BY pp.payment_method
ORDER BY total_amount DESC;

-- name: GetCashierDailyPerformance :one
-- Get daily performance metrics for a cashier
SELECT 
    c.id as cashier_id,
    c.cashier_code,
    u.username,
    u.first_name || ' ' || u.last_name as full_name,
    COUNT(DISTINCT cs.id) as sessions_worked,
    COUNT(DISTINCT pt.id) as total_transactions,
    COALESCE(SUM(pt.subtotal), 0) as gross_sales,
    COALESCE(SUM(pt.discount_amount), 0) as total_discounts,
    COALESCE(SUM(pt.tax_amount), 0) as total_tax,
    COALESCE(SUM(pt.total_amount), 0) as net_sales,
    COALESCE(AVG(pt.total_amount), 0) as avg_transaction_value,
    COUNT(CASE WHEN pt.status = 'voided' THEN 1 END) as voids_count,
    COALESCE(SUM(CASE WHEN pt.status = 'voided' THEN pt.total_amount ELSE 0 END), 0) as voids_amount
FROM cashiers c
INNER JOIN users u ON u.id = c.user_id
LEFT JOIN cashier_sessions cs ON cs.cashier_id = c.id 
    AND cs.opening_time::date = $2
LEFT JOIN pos_transactions pt ON pt.cashier_session_id = cs.id
WHERE c.id = $1
GROUP BY c.id, c.cashier_code, u.username, u.first_name, u.last_name;

-- name: GetCashierPerformanceByDateRange :one
-- Get cashier performance for a date range
SELECT 
    c.id as cashier_id,
    c.cashier_code,
    u.username,
    u.first_name || ' ' || u.last_name as full_name,
    COUNT(DISTINCT cs.id) as total_sessions,
    COUNT(DISTINCT pt.id) as total_transactions,
    COALESCE(SUM(pt.subtotal), 0) as gross_sales,
    COALESCE(SUM(pt.discount_amount), 0) as total_discounts,
    COALESCE(SUM(pt.tax_amount), 0) as total_tax,
    COALESCE(SUM(pt.total_amount), 0) as net_sales,
    COALESCE(AVG(pt.total_amount), 0) as avg_transaction_value
FROM cashiers c
INNER JOIN users u ON u.id = c.user_id
LEFT JOIN cashier_sessions cs ON cs.cashier_id = c.id 
    AND cs.opening_time::date BETWEEN $2 AND $3
LEFT JOIN pos_transactions pt ON pt.cashier_session_id = cs.id
    AND pt.status != 'voided'
WHERE c.id = $1
GROUP BY c.id, c.cashier_code, u.username, u.first_name, u.last_name;

-- name: GetTopPerformingCashiers :many
-- Get top N performing cashiers by sales for a date range
SELECT 
    c.id,
    c.cashier_code,
    u.first_name || ' ' || u.last_name as full_name,
    COUNT(DISTINCT pt.id) as transaction_count,
    COALESCE(SUM(pt.total_amount), 0) as total_sales,
    COALESCE(AVG(pt.total_amount), 0) as avg_transaction_value
FROM cashiers c
INNER JOIN users u ON u.id = c.user_id
INNER JOIN cashier_sessions cs ON cs.cashier_id = c.id
INNER JOIN pos_transactions pt ON pt.cashier_session_id = cs.id
WHERE c.store_id = $1
  AND cs.opening_time::date BETWEEN $2 AND $3
  AND pt.status != 'voided'
GROUP BY c.id, c.cashier_code, u.first_name, u.last_name
ORDER BY total_sales DESC
LIMIT $4;

-- =====================================================
-- VOID TRANSACTION WITH AUDIT
-- =====================================================

-- name: VoidTransaction :one
-- Void a transaction and record who voided it
UPDATE pos_transactions
SET
    status = 'voided',
    voided_by = $2,
    voided_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetVoidedTransactionsBySession :many
-- Get all voided transactions for a session
SELECT 
    pt.*,
    u.username as voided_by_username,
    u.first_name || ' ' || u.last_name as voided_by_name
FROM pos_transactions pt
LEFT JOIN users u ON u.id = pt.voided_by
WHERE pt.cashier_session_id = $1
  AND pt.status = 'voided'
ORDER BY pt.voided_at DESC;

-- name: GetVoidedTransactionsByDate :many
-- Get voided transactions for a date range
SELECT 
    pt.*,
    c.cashier_code,
    u.username as voided_by_username
FROM pos_transactions pt
INNER JOIN cashiers c ON c.id = pt.cashier_id
LEFT JOIN users u ON u.id = pt.voided_by
WHERE pt.store_id = $1
  AND pt.transaction_date::date BETWEEN $2 AND $3
  AND pt.status = 'voided'
ORDER BY pt.voided_at DESC;

-- =====================================================
-- CONVERT CART TO POS TRANSACTION
-- =====================================================

-- name: CreatePosTransactionFromCart :one
-- Create a POS transaction record linking to source cart
INSERT INTO pos_transactions (
    store_id,
    cashier_id,
    cashier_session_id,
    customer_id,
    pos_terminal_id,
    transaction_number,
    transaction_date,
    transaction_type,
    subtotal,
    discount_amount,
    tax_amount,
    total_amount,
    total_cost,
    amount_paid,
    change_given,
    status,
    price_list_id,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
    jsonb_build_object('source_cart_id', $18)
)
RETURNING *;

-- name: BulkCreatePosTransactionLines :copyfrom
-- Bulk insert transaction lines from cart items
INSERT INTO pos_transaction_lines (
    transaction_id,
    product_id,
    product_variant_id,
    quantity,
    uom_id,
    unit_price,
    discount_amount,
    tax_amount,
    subtotal,
    line_total,
    cost_price,
    line_number,
    serial_number,
    batch_number,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
);

-- name: UpdateCartToConverted :exec
-- Mark cart as converted after creating POS transaction
UPDATE carts
SET
    cart_status = 'converted',
    converted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- =====================================================
-- ANALYTICS & REPORTING
-- =====================================================

-- name: GetCashierVarianceReport :many
-- Get variance report for all cashiers in a store
SELECT 
    c.id,
    c.cashier_code,
    u.first_name || ' ' || u.last_name as full_name,
    COUNT(cs.id) as session_count,
    COALESCE(SUM(cs.variance), 0) as total_variance,
    COALESCE(AVG(cs.variance), 0) as avg_variance,
    COUNT(CASE WHEN cs.variance > 0 THEN 1 END) as over_count,
    COUNT(CASE WHEN cs.variance < 0 THEN 1 END) as short_count
FROM cashiers c
INNER JOIN users u ON u.id = c.user_id
LEFT JOIN cashier_sessions cs ON cs.cashier_id = c.id
    AND cs.opening_time::date BETWEEN $2 AND $3
    AND cs.status = 'closed'
WHERE c.store_id = $1
GROUP BY c.id, c.cashier_code, u.first_name, u.last_name
ORDER BY total_variance DESC;

-- name: GetSessionsByStatus :many
-- Get sessions by status for a store
SELECT 
    cs.*,
    c.cashier_code,
    u.first_name || ' ' || u.last_name as cashier_name,
    pt.terminal_code
FROM cashier_sessions cs
INNER JOIN cashiers c ON c.id = cs.cashier_id
INNER JOIN users u ON u.id = c.user_id
INNER JOIN pos_terminals pt ON pt.id = cs.pos_terminal_id
WHERE c.store_id = $1
  AND cs.status = $2
ORDER BY cs.opening_time DESC;

-- name: GetOpenSessionsForStore :many
-- Get all currently open sessions for a store
SELECT 
    cs.*,
    c.cashier_code,
    u.first_name || ' ' || u.last_name as cashier_name,
    pt.terminal_code,
    pt.terminal_name
FROM cashier_sessions cs
INNER JOIN cashiers c ON c.id = cs.cashier_id
INNER JOIN users u ON u.id = c.user_id
INNER JOIN pos_terminals pt ON pt.id = cs.pos_terminal_id
WHERE c.store_id = $1
  AND cs.status = 'open'
ORDER BY cs.opening_time DESC;

-- name: GetCashierWithUserDetails :one
-- Get cashier with full user details
SELECT 
    c.*,
    u.username,
    u.first_name || ' ' || u.last_name as full_name,
    u.email,
    u.employee_code,
    u.is_active as user_active
FROM cashiers c
INNER JOIN users u ON u.id = c.user_id
WHERE c.id = $1;

-- name: ListCashiersWithUserDetails :many
-- List all cashiers with user details
SELECT 
    c.*,
    u.username,
    u.first_name || ' ' || u.last_name as full_name,
    u.email,
    u.employee_code
FROM cashiers c
INNER JOIN users u ON u.id = c.user_id
WHERE c.store_id = $1
  AND c.is_active = true
ORDER BY c.cashier_code;

-- name: GetCashierWithLimits :one
-- Get cashier with limits and user details
SELECT 
    c.id,
    c.cashier_code,
    c.drawer_limit,
    c.discount_limit,
    c.is_active,
    u.first_name,
    u.last_name,
    u.email,
    s.name AS store_name
FROM cashiers c
JOIN users u ON c.user_id = u.id
JOIN stores s ON c.store_id = s.id
WHERE c.id = $1;

-- name: ListActiveCashiersInStore :many
-- List active cashiers in a store with session information
SELECT 
    c.id,
    c.cashier_code,
    u.first_name || ' ' || u.last_name AS full_name,
    c.discount_limit,
    c.drawer_limit,
    COUNT(cs.id) FILTER (WHERE cs.status = 'open') AS active_sessions
FROM cashiers c
JOIN users u ON c.user_id = u.id
LEFT JOIN cashier_sessions cs ON cs.cashier_id = c.id AND cs.status = 'open'
WHERE c.store_id = $1
  AND c.is_active = true
GROUP BY c.id, u.first_name, u.last_name
ORDER BY u.first_name;