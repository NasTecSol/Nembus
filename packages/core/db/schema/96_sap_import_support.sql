-- =====================================================================
-- Migration: SAP Import Support Columns
-- Adds must_reset_password to users table and customer_addresses table
-- for BP address import from SAP CRD1
-- =====================================================================

-- Add must_reset_password to users table.
-- When the SAP agent imports users with the sentinel password hash
-- ({SAP_IMPORT_MUST_RESET}), this flag is set to true.
-- The authentication layer must check this flag and redirect the user
-- to the password reset flow before allowing any other action.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS must_reset_password BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS sap_imported BOOLEAN NOT NULL DEFAULT false;

-- Index for fast lookup of users who still need to reset their password
CREATE INDEX IF NOT EXISTS idx_users_must_reset ON users(must_reset_password) WHERE must_reset_password = true;

-- Customer Addresses table (sourced from SAP CRD1)
CREATE TABLE IF NOT EXISTS customer_addresses (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    address_type VARCHAR(20) NOT NULL DEFAULT 'shipping', -- 'billing' or 'shipping'
    address_line VARCHAR(255),
    street VARCHAR(255),
    city VARCHAR(100),
    country VARCHAR(100),
    postal_code VARCHAR(30),
    state VARCHAR(100),
    phone VARCHAR(50),
    is_default BOOLEAN DEFAULT false,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(customer_id, address_type, address_line)
);

CREATE INDEX IF NOT EXISTS idx_customer_addresses_customer_id ON customer_addresses(customer_id);
CREATE INDEX IF NOT EXISTS idx_customer_addresses_type ON customer_addresses(address_type);
