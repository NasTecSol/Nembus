-- Migration: Unify Business Partners (SAP-aligned)
-- Enhances business_partners to handle all partner roles (supplier, customer, lead)

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

ALTER TABLE partner_addresses DROP CONSTRAINT IF EXISTS uq_partner_addresses_type_name;
ALTER TABLE partner_addresses ADD CONSTRAINT uq_partner_addresses_type_name 
    UNIQUE(partner_id, address_type, address_name);
