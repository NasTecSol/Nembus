-- =====================================================================
-- Migration: Unify Business Partners (SAP-aligned)
-- Enhances business_partners to handle all partner roles (supplier, customer, lead)
-- =====================================================================

-- 1. Relax/update partner_role check constraint to include 'customer' and 'lead'
ALTER TABLE business_partners DROP CONSTRAINT IF EXISTS business_partners_partner_role_check;
ALTER TABLE business_partners ADD CONSTRAINT business_partners_partner_role_check 
    CHECK (partner_role::text = ANY (ARRAY[
        'supplier'::text, 
        'vendor'::text, 
        'customer'::text, 
        'lead'::text, 
        'special_customer'::text, 
        'corporate_group'::text
    ]));

-- 2. Ensure partner_addresses has a unique constraint for idempotent upserts
ALTER TABLE partner_addresses DROP CONSTRAINT IF EXISTS uq_partner_addresses_type_name;
ALTER TABLE partner_addresses ADD CONSTRAINT uq_partner_addresses_type_name 
    UNIQUE(partner_id, address_type, address_name);
