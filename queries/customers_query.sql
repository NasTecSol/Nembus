-- name: CreateCustomer :one
INSERT INTO customers (
    organization_id,
    customer_code,
    name,
    email,
    phone,
    address,
    customer_type,
    price_list_id,
    credit_limit,
    outstanding_balance,
    is_active,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetCustomer :one
SELECT * FROM customers
WHERE id = $1;

-- name: GetCustomerByCode :one
SELECT * FROM customers
WHERE organization_id = $1 AND customer_code = $2;

-- name: ListCustomers :many
SELECT * FROM customers
WHERE organization_id = $1
ORDER BY name;

-- name: ListActiveCustomers :many
SELECT * FROM customers
WHERE organization_id = $1 AND is_active = true
ORDER BY name;

-- name: ListCustomersByType :many
SELECT * FROM customers
WHERE organization_id = $1 AND customer_type = $2
ORDER BY name;

-- name: SearchCustomers :many
SELECT * FROM customers
WHERE organization_id = $1 
  AND (name ILIKE $2 OR customer_code ILIKE $2)
ORDER BY name
LIMIT $3;

-- name: UpdateCustomer :one
UPDATE customers
SET 
    name = $2,
    email = $3,
    phone = $4,
    address = $5,
    customer_type = $6,
    price_list_id = $7,
    credit_limit = $8,
    is_active = $9,
    metadata = $10
WHERE id = $1
RETURNING *;

-- name: UpdateCustomerBalance :one
UPDATE customers
SET outstanding_balance = outstanding_balance + $2
WHERE id = $1
RETURNING *;

-- name: DeleteCustomer :exec
DELETE FROM customers
WHERE id = $1;

-- name: ToggleCustomerActive :one
UPDATE customers
SET is_active = $2
WHERE id = $1
RETURNING *;

-- name: GetCustomersWithOutstandingBalance :many
SELECT * FROM customers
WHERE organization_id = $1 
  AND outstanding_balance > 0
  AND is_active = true
ORDER BY outstanding_balance DESC;

-- name: GetCustomerCreditStatus :one
SELECT 
    id,
    name,
    credit_limit,
    outstanding_balance,
    (credit_limit - outstanding_balance) AS available_credit
FROM customers
WHERE id = $1;

-- name: AdjustCustomerLoyaltyPoints :one
-- Pass a negative value for $2 to redeem/deduct points; positive to add
UPDATE customers
SET loyalty_points = loyalty_points + $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetCustomerLoyaltyBalance :one
-- Lightweight fetch for POS validation before redemption
SELECT id, name, loyalty_points
FROM customers
WHERE id = $1;
