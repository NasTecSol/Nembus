-- =====================================================
-- CHART OF ACCOUNTS QUERIES
-- =====================================================

-- name: CreateAccount :one
INSERT INTO chart_of_accounts (
    organization_id, account_code, account_name, account_type, parent_account_id, is_active
) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetAccount :one
SELECT * FROM chart_of_accounts WHERE id = $1;

-- name: GetAccountByCode :one
SELECT * FROM chart_of_accounts WHERE account_code = $1;

-- name: ListAccounts :many
SELECT * FROM chart_of_accounts
WHERE organization_id = $1 AND is_active = COALESCE(sqlc.narg(is_active), is_active)
ORDER BY account_code;

-- name: ListAccountsByType :many
SELECT * FROM chart_of_accounts
WHERE organization_id = $1 AND account_type = $2 AND is_active = true
ORDER BY account_code;

-- name: UpdateAccount :one
UPDATE chart_of_accounts SET
    account_name      = COALESCE(sqlc.narg(account_name), account_name),
    account_type      = COALESCE(sqlc.narg(account_type), account_type),
    parent_account_id = COALESCE(sqlc.narg(parent_account_id), parent_account_id),
    is_active         = COALESCE(sqlc.narg(is_active), is_active)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM chart_of_accounts WHERE id = $1;

-- =====================================================
-- COST CENTERS QUERIES
-- =====================================================

-- name: CreateCostCenter :one
INSERT INTO cost_centers (organization_id, code, name, dimension, is_active)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetCostCenter :one
SELECT * FROM cost_centers WHERE id = $1;

-- name: ListCostCenters :many
SELECT * FROM cost_centers
WHERE organization_id = $1 AND is_active = true
ORDER BY code;

-- =====================================================
-- PAYMENT TERMS QUERIES
-- =====================================================

-- name: CreatePaymentTerm :one
INSERT INTO payment_terms (
    organization_id, code, name, due_days, discount_days,
    discount_percentage, late_fee_percentage, is_active
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

-- name: GetPaymentTerm :one
SELECT * FROM payment_terms WHERE id = $1;

-- name: GetPaymentTermByCode :one
SELECT * FROM payment_terms WHERE organization_id = $1 AND code = $2;

-- name: ListPaymentTerms :many
SELECT * FROM payment_terms WHERE organization_id = $1 AND is_active = true ORDER BY code;

-- =====================================================
-- CURRENCIES & EXCHANGE RATES QUERIES
-- =====================================================

-- name: CreateCurrency :one
INSERT INTO currencies (code, name, symbol, decimal_places, is_active)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetCurrency :one
SELECT * FROM currencies WHERE code = $1;

-- name: ListCurrencies :many
SELECT * FROM currencies WHERE is_active = true ORDER BY code;

-- name: CreateExchangeRate :one
INSERT INTO exchange_rates (organization_id, from_currency, to_currency, rate_date, rate)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetLatestExchangeRate :one
SELECT * FROM exchange_rates
WHERE organization_id = $1 AND from_currency = $2 AND to_currency = $3
ORDER BY rate_date DESC
LIMIT 1;

-- name: GetExchangeRateOnDate :one
SELECT * FROM exchange_rates
WHERE organization_id = $1 AND from_currency = $2 AND to_currency = $3 AND rate_date = $4;
