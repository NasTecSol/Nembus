-- =====================================================
-- BUSINESS PARTNERS QUERIES (SQLC)
-- =====================================================

-- name: CreateBusinessPartner :one
INSERT INTO business_partners (
    organization_id,
    code,
    name,
    partner_role,
    tax_id,
    currency_code,
    credit_limit,
    outstanding_balance,
    payment_terms_id,
    sales_rep_user_id,
    is_active,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetBusinessPartner :one
SELECT * FROM business_partners
WHERE id = $1;

-- name: GetBusinessPartnerByCode :one
SELECT * FROM business_partners
WHERE organization_id = $1 AND code = $2;

-- name: ListBusinessPartners :many
SELECT * FROM business_partners
WHERE organization_id = $1
ORDER BY name;

-- name: ListBusinessPartnersByRole :many
SELECT * FROM business_partners
WHERE organization_id = $1 AND partner_role = $2
ORDER BY name;

-- name: SearchBusinessPartners :many
SELECT * FROM business_partners
WHERE organization_id = $1 
  AND (name ILIKE $2 OR code ILIKE $2)
ORDER BY name
LIMIT $3;

-- name: UpdateBusinessPartner :one
UPDATE business_partners
SET 
    code = $2,
    name = $3,
    partner_role = $4,
    tax_id = $5,
    currency_code = $6,
    credit_limit = $7,
    payment_terms_id = $8,
    sales_rep_user_id = $9,
    is_active = $10,
    metadata = $11
WHERE id = $1
RETURNING *;

-- name: UpdateBusinessPartnerBalance :one
UPDATE business_partners
SET outstanding_balance = outstanding_balance + $2
WHERE id = $1
RETURNING *;

-- name: DeleteBusinessPartner :exec
DELETE FROM business_partners
WHERE id = $1;

-- name: ToggleBusinessPartnerActive :one
UPDATE business_partners
SET is_active = $2
WHERE id = $1
RETURNING *;

-- =====================================================
-- PARTNER ADDRESSES QUERIES
-- =====================================================

-- name: CreatePartnerAddress :one
INSERT INTO partner_addresses (
    partner_id,
    address_name,
    address_type,
    street,
    city,
    state,
    zip_code,
    country_code,
    is_default
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetPartnerAddress :one
SELECT * FROM partner_addresses
WHERE id = $1;

-- name: ListPartnerAddresses :many
SELECT * FROM partner_addresses
WHERE partner_id = $1
ORDER BY is_default DESC, created_at DESC;

-- name: UpdatePartnerAddress :one
UPDATE partner_addresses
SET
    address_name = $2,
    address_type = $3,
    street = $4,
    city = $5,
    state = $6,
    zip_code = $7,
    country_code = $8,
    is_default = $9
WHERE id = $1
RETURNING *;

-- name: DeletePartnerAddress :exec
DELETE FROM partner_addresses
WHERE id = $1;

-- name: ClearDefaultPartnerAddresses :exec
UPDATE partner_addresses
SET is_default = false
WHERE partner_id = $1 AND id != $2;

-- =====================================================
-- PARTNER CONTACTS QUERIES
-- =====================================================

-- name: CreatePartnerContact :one
INSERT INTO partner_contacts (
    partner_id,
    first_name,
    last_name,
    email,
    phone,
    position,
    is_primary
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetPartnerContact :one
SELECT * FROM partner_contacts
WHERE id = $1;

-- name: ListPartnerContacts :many
SELECT * FROM partner_contacts
WHERE partner_id = $1
ORDER BY is_primary DESC, created_at DESC;

-- name: UpdatePartnerContact :one
UPDATE partner_contacts
SET
    first_name = $2,
    last_name = $3,
    email = $4,
    phone = $5,
    position = $6,
    is_primary = $7
WHERE id = $1
RETURNING *;

-- name: DeletePartnerContact :exec
DELETE FROM partner_contacts
WHERE id = $1;

-- name: ClearPrimaryPartnerContacts :exec
UPDATE partner_contacts
SET is_primary = false
WHERE partner_id = $1 AND id != $2;
