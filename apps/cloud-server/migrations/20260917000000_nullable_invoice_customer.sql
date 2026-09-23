-- Migration: Make invoices.customer_id nullable to support SAP imports
-- 
-- Problem: The invoices table required customer_id NOT NULL, but SAP-imported
-- invoices often reference a CardCode (customer) that hasn't been imported yet,
-- causing the entire invoice batch to fail with a NOT NULL constraint violation.
--
-- Solution: Make customer_id nullable. The SAP import now stores sap_customer_code
-- in the invoice metadata so invoices can be reconciled with customers later.
-- The ON CONFLICT update uses COALESCE so customer_id is set if a re-import
-- happens after the customer is created.

ALTER TABLE invoices
    ALTER COLUMN customer_id DROP NOT NULL;

-- Drop RESTRICT FK and re-create as SET NULL
ALTER TABLE invoices
    DROP CONSTRAINT IF EXISTS invoices_customer_id_fkey;

ALTER TABLE invoices
    ADD CONSTRAINT invoices_customer_id_fkey
    FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE SET NULL;

COMMENT ON COLUMN invoices.customer_id IS
    'Nullable: SAP-imported invoices may arrive before their customer record. '
    'Use metadata->sap_customer_code to reconcile after customers are imported.';
