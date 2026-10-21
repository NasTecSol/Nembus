-- =====================================================================
-- Migration: SAP Migration Support (A20, A25, A27, A31)
-- Idempotency arbiter for stock movement replay, payment audit tables,
-- and metadata lookup indexes used by sap_migration.go
-- =====================================================================

-- ---------------------------------------------------------------------
-- A25: Arbiter index required by the stock_movements replay-safe upsert.
-- sap_migration.go uses:
--   ON CONFLICT ((metadata->>'sap_trans_num')) WHERE metadata ? 'sap_trans_num'
-- PostgreSQL ON CONFLICT inference requires an exact-match unique index.
-- ---------------------------------------------------------------------
CREATE UNIQUE INDEX IF NOT EXISTS stock_movements_sap_trans_num_uk
    ON stock_movements ((metadata->>'sap_trans_num'))
    WHERE metadata ? 'sap_trans_num';

-- ---------------------------------------------------------------------
-- A20: On-account / pre-cutoff incoming payments.
-- invoice_payments.invoice_id is NOT NULL, so payments that are not tied
-- to a migrated invoice (on-account payments, payments against invoices
-- older than the configured InvoiceStartDate) are stored here instead of
-- being silently dropped. During the historical migration the customer
-- balance authority is the OCRD.Balance snapshot (already reflects these
-- payments), so no balance mutation is performed on ingest.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS customer_payments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id   INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    payment_number    VARCHAR(50) UNIQUE NOT NULL,
    customer_id       INTEGER REFERENCES customers(id) ON DELETE SET NULL,
    customer_code     VARCHAR(50),
    payment_date      DATE NOT NULL,
    payment_amount    DECIMAL(15,2) NOT NULL CHECK (payment_amount > 0),
    payment_method    VARCHAR(100),
    payment_reference VARCHAR(255),
    currency_code     VARCHAR(3) DEFAULT 'SAR',
    reason            VARCHAR(50) DEFAULT 'on_account', -- 'on_account' | 'unmatched_invoice'
    notes             TEXT,
    metadata          JSONB DEFAULT '{}',
    created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_customer_payments_customer_code ON customer_payments(organization_id, customer_code);
CREATE INDEX IF NOT EXISTS idx_customer_payments_payment_date ON customer_payments(payment_date);

-- ---------------------------------------------------------------------
-- A31: Outgoing (vendor) payments from OVPM/VPM2.
-- Audit trail for supplier payments; also the source used to populate
-- purchase_orders.metadata->>'amount_paid' consumed by
-- vw_accounts_payable and vw_supplier_aging_report.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS vendor_payments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id   INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    payment_number    VARCHAR(50) UNIQUE NOT NULL,
    partner_id        INTEGER REFERENCES business_partners(id) ON DELETE SET NULL,
    supplier_code     VARCHAR(50),
    purchase_order_id INTEGER REFERENCES purchase_orders(id) ON DELETE SET NULL,
    po_doc_entry      BIGINT,
    payment_date      DATE NOT NULL,
    payment_amount    DECIMAL(15,2) NOT NULL CHECK (payment_amount > 0),
    payment_method    VARCHAR(100),
    payment_reference VARCHAR(255),
    currency_code     VARCHAR(3) DEFAULT 'SAR',
    notes             TEXT,
    metadata          JSONB DEFAULT '{}',
    created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vendor_payments_supplier_code ON vendor_payments(organization_id, supplier_code);
CREATE INDEX IF NOT EXISTS idx_vendor_payments_po ON vendor_payments(purchase_order_id);

-- ---------------------------------------------------------------------
-- A27: Expression indexes for the per-row document resolution lookups
-- used by sap_migration.go (reference_id on stock movements, GRN->PO,
-- payment->invoice). Non-unique, metadata JSONB expression indexes.
-- ---------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx_purchase_orders_sap_doc_entry ON purchase_orders ((metadata->>'sap_doc_entry'));
CREATE INDEX IF NOT EXISTS idx_goods_receipt_notes_sap_doc_entry ON goods_receipt_notes ((metadata->>'sap_doc_entry'));
CREATE INDEX IF NOT EXISTS idx_invoices_sap_doc_entry ON invoices ((metadata->>'sap_doc_entry'));
CREATE INDEX IF NOT EXISTS idx_invoices_sap_doc_num ON invoices ((metadata->>'sap_doc_num'));
CREATE INDEX IF NOT EXISTS idx_sales_returns_sap_doc_num ON sales_returns ((metadata->>'sap_doc_num'));
CREATE INDEX IF NOT EXISTS idx_transfer_requests_sap_doc_num ON transfer_requests ((metadata->>'sap_doc_num'));
CREATE INDEX IF NOT EXISTS idx_sales_orders_v2_sap_doc_num ON sales_orders_v2 ((metadata->>'sap_doc_num'));

-- ---------------------------------------------------------------------
-- Ensure partner_addresses unique constraint for idempotent address upserts
-- ---------------------------------------------------------------------
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uq_partner_addresses_type_name'
    ) THEN
        ALTER TABLE partner_addresses ADD CONSTRAINT uq_partner_addresses_type_name 
            UNIQUE (partner_id, address_type, address_name);
    END IF;
END $$;
