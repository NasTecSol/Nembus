-- =====================================================
-- BUSINESS PARTNERS QUERIES
-- Supersedes suppliers_query.sql for all B2B partner operations
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
SELECT * FROM business_partners WHERE id = $1;

-- name: GetBusinessPartnerByCode :one
SELECT * FROM business_partners
WHERE organization_id = $1 AND code = $2;

-- name: ListBusinessPartners :many
SELECT * FROM business_partners
WHERE organization_id = $1
  AND (sqlc.narg(partner_role)::varchar IS NULL OR partner_role = sqlc.narg(partner_role)::varchar)
ORDER BY name;

-- name: ListActiveBusinessPartners :many
SELECT * FROM business_partners
WHERE organization_id = $1
  AND is_active = true
  AND (sqlc.narg(partner_role)::varchar IS NULL OR partner_role = sqlc.narg(partner_role)::varchar)
ORDER BY name;

-- name: SearchBusinessPartners :many
SELECT * FROM business_partners
WHERE organization_id = $1
  AND (name ILIKE $2 OR code ILIKE $2)
  AND (sqlc.narg(partner_role)::varchar IS NULL OR partner_role = sqlc.narg(partner_role)::varchar)
ORDER BY name
LIMIT $3;

-- name: UpdateBusinessPartner :one
UPDATE business_partners SET
    name                = COALESCE(sqlc.narg(name), name),
    tax_id              = COALESCE(sqlc.narg(tax_id), tax_id),
    currency_code       = COALESCE(sqlc.narg(currency_code), currency_code),
    credit_limit        = COALESCE(sqlc.narg(credit_limit), credit_limit),
    payment_terms_id    = COALESCE(sqlc.narg(payment_terms_id), payment_terms_id),
    sales_rep_user_id   = COALESCE(sqlc.narg(sales_rep_user_id), sales_rep_user_id),
    is_active           = COALESCE(sqlc.narg(is_active), is_active),
    metadata            = COALESCE(sqlc.narg(metadata), metadata)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ToggleBusinessPartnerActive :one
UPDATE business_partners SET is_active = $2 WHERE id = $1 RETURNING *;

-- name: UpdateBusinessPartnerBalance :one
UPDATE business_partners
SET outstanding_balance = outstanding_balance + $2
WHERE id = $1
RETURNING *;

-- name: GetBusinessPartnerForSupplierLookup :one
-- Returns business partners that act as suppliers or vendors (GRN/PO context)
SELECT * FROM business_partners
WHERE organization_id = $1
  AND id = $2
  AND partner_role IN ('supplier', 'vendor')
  AND is_active = true;

-- name: ListSuppliersAsBusinessPartners :many
-- Replacement for ListSuppliers - returns all active supplier/vendor partners
SELECT * FROM business_partners
WHERE organization_id = $1
  AND partner_role IN ('supplier', 'vendor')
  AND is_active = true
ORDER BY name;

-- =====================================================
-- PARTNER ADDRESSES
-- =====================================================

-- name: CreatePartnerAddress :one
INSERT INTO partner_addresses (
    partner_id, address_name, address_type,
    street, city, state, zip_code, country_code, is_default
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;

-- name: ListPartnerAddresses :many
SELECT * FROM partner_addresses WHERE partner_id = $1 ORDER BY is_default DESC;

-- name: GetDefaultPartnerAddress :one
SELECT * FROM partner_addresses
WHERE partner_id = $1 AND address_type = $2 AND is_default = true
LIMIT 1;

-- name: UpdatePartnerAddress :one
UPDATE partner_addresses SET
    address_name = COALESCE(sqlc.narg(address_name), address_name),
    street       = COALESCE(sqlc.narg(street), street),
    city         = COALESCE(sqlc.narg(city), city),
    state        = COALESCE(sqlc.narg(state), state),
    zip_code     = COALESCE(sqlc.narg(zip_code), zip_code),
    country_code = COALESCE(sqlc.narg(country_code), country_code),
    is_default   = COALESCE(sqlc.narg(is_default), is_default)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeletePartnerAddress :exec
DELETE FROM partner_addresses WHERE id = $1;

-- =====================================================
-- PARTNER CONTACTS
-- =====================================================

-- name: CreatePartnerContact :one
INSERT INTO partner_contacts (
    partner_id, first_name, last_name, email, phone, position, is_primary
) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: ListPartnerContacts :many
SELECT * FROM partner_contacts WHERE partner_id = $1 ORDER BY is_primary DESC;

-- name: GetPrimaryPartnerContact :one
SELECT * FROM partner_contacts WHERE partner_id = $1 AND is_primary = true LIMIT 1;

-- name: DeletePartnerContact :exec
DELETE FROM partner_contacts WHERE id = $1;
