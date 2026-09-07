-- =====================================================
-- PHASE 1 — DOUBLE-ENTRY ACCOUNTING ENGINE
-- File: 55_accounting.sql
-- Depends on: 10_identity_rbac.sql (organizations, users)
--             50_purchasing_suppliers.sql (chart_of_accounts, cost_centers,
--                                          journal_entries, journal_lines,
--                                          business_partners, currencies)
-- =====================================================

-- =====================================================
-- FISCAL CALENDAR
-- =====================================================

CREATE TABLE fiscal_years (
    id               SERIAL PRIMARY KEY,
    organization_id  INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    year             INTEGER NOT NULL,                        -- e.g. 2026
    name             VARCHAR(100) NOT NULL,                   -- e.g. "FY 2026"
    start_date       DATE NOT NULL,
    end_date         DATE NOT NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'open'
                         CHECK (status IN ('open','closing','closed')),
    closed_by        INTEGER REFERENCES users(id) ON DELETE SET NULL,
    closed_at        TIMESTAMP,
    metadata         JSONB DEFAULT '{}',
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (organization_id, year)
);

-- Posting periods — one per month within a fiscal year
CREATE TABLE posting_periods (
    id                   SERIAL PRIMARY KEY,
    organization_id      INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    fiscal_year_id       INTEGER NOT NULL REFERENCES fiscal_years(id) ON DELETE CASCADE,
    period_no            INTEGER NOT NULL CHECK (period_no BETWEEN 1 AND 13), -- 13 = adjustment period
    period_name          VARCHAR(50) NOT NULL,                -- e.g. "Sep 2026"
    start_date           DATE NOT NULL,
    end_date             DATE NOT NULL,
    status               VARCHAR(20) NOT NULL DEFAULT 'open'
                             CHECK (status IN ('draft','open','closed','locked')),
    allow_retro_posting  BOOLEAN DEFAULT false,               -- allow posting to closed period
    closed_by            INTEGER REFERENCES users(id) ON DELETE SET NULL,
    closed_at            TIMESTAMP,
    metadata             JSONB DEFAULT '{}',
    created_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (organization_id, fiscal_year_id, period_no)
);

-- =====================================================
-- DOCUMENT NUMBERING SERIES
-- Replaces free-text VARCHAR(50) UNIQUE *_number columns
-- =====================================================

CREATE TABLE document_series (
    id               SERIAL PRIMARY KEY,
    organization_id  INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    -- doc_type values: pos_transaction | grn | sales_invoice | ap_invoice |
    --   supplier_payment | customer_payment | journal_entry | purchase_order |
    --   purchase_requisition | purchase_return | delivery_note | credit_note |
    --   sales_order | quote | fixed_asset | stock_count | transfer_request
    doc_type         VARCHAR(80) NOT NULL,
    series_code      VARCHAR(20) NOT NULL,                    -- e.g. "SI-2026"
    name             VARCHAR(100) NOT NULL,
    prefix           VARCHAR(20) DEFAULT '',
    suffix           VARCHAR(20) DEFAULT '',
    start_no         INTEGER NOT NULL DEFAULT 1,
    next_no          INTEGER NOT NULL DEFAULT 1,
    increment        INTEGER NOT NULL DEFAULT 1,
    zero_pad_length  INTEGER DEFAULT 6,                       -- e.g. 6 → "000001"
    is_default       BOOLEAN DEFAULT true,
    per_fiscal_year  BOOLEAN DEFAULT false,                   -- reset next_no each fiscal year
    fiscal_year_id   INTEGER REFERENCES fiscal_years(id) ON DELETE SET NULL,
    is_active        BOOLEAN DEFAULT true,
    metadata         JSONB DEFAULT '{}',
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (organization_id, doc_type, series_code)
);

-- =====================================================
-- PROFIT CENTERS (dimension — alongside cost_centers)
-- =====================================================

CREATE TABLE profit_centers (
    id               SERIAL PRIMARY KEY,
    organization_id  INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    code             VARCHAR(50) NOT NULL,
    name             VARCHAR(100) NOT NULL,
    parent_id        INTEGER REFERENCES profit_centers(id) ON DELETE SET NULL,
    is_active        BOOLEAN DEFAULT true,
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (organization_id, code)
);

-- =====================================================
-- EXTEND EXISTING TABLES (ALTER TABLE — idempotent)
-- =====================================================

-- Extend chart_of_accounts with ERP-grade fields
ALTER TABLE chart_of_accounts
    ADD COLUMN IF NOT EXISTS account_group        VARCHAR(50),
    -- account_group examples: 'current_assets','fixed_assets','current_liabilities',
    --   'long_term_liabilities','equity','revenue','cogs','operating_expense','tax'
    ADD COLUMN IF NOT EXISTS is_control_account   BOOLEAN DEFAULT false,
    -- true for AR control, AP control, VAT control — postings only through subledger
    ADD COLUMN IF NOT EXISTS control_type         VARCHAR(30),
    -- control_type values: 'ar' | 'ap' | 'vat_output' | 'vat_input' | 'bank'
    ADD COLUMN IF NOT EXISTS is_bank_account      BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS external_code        VARCHAR(50),   -- statutory / SOCPA code
    ADD COLUMN IF NOT EXISTS currency_code        VARCHAR(3) REFERENCES currencies(code),
    -- set for hard-currency accounts; NULL means use org base currency
    ADD COLUMN IF NOT EXISTS allow_direct_posting BOOLEAN DEFAULT true,
    -- false for header/group accounts (no journal lines allowed directly)
    ADD COLUMN IF NOT EXISTS updated_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Extend journal_entries to become the ledger header
ALTER TABLE journal_entries
    ADD COLUMN IF NOT EXISTS fiscal_year_id       INTEGER REFERENCES fiscal_years(id),
    ADD COLUMN IF NOT EXISTS posting_period_id    INTEGER REFERENCES posting_periods(id),
    ADD COLUMN IF NOT EXISTS series_id            INTEGER REFERENCES document_series(id),
    ADD COLUMN IF NOT EXISTS document_number      VARCHAR(80),   -- rendered from series
    ADD COLUMN IF NOT EXISTS status               VARCHAR(20) DEFAULT 'posted'
                                 CHECK (status IN ('draft','posted','void','reversed')),
    ADD COLUMN IF NOT EXISTS source_type          VARCHAR(80),
    -- source_type: pos_transaction | grn | sales_invoice | ap_invoice |
    --   supplier_payment | customer_payment | stock_adjustment | transfer |
    --   stock_count | depreciation | manual
    ADD COLUMN IF NOT EXISTS source_id            VARCHAR(100),  -- PK of triggering document
    ADD COLUMN IF NOT EXISTS reversal_of_id       BIGINT REFERENCES journal_entries(id),
    ADD COLUMN IF NOT EXISTS base_currency_code   VARCHAR(3) REFERENCES currencies(code) DEFAULT 'SAR',
    ADD COLUMN IF NOT EXISTS total_debit_base     DECIMAL(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_credit_base    DECIMAL(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS posted_by            INTEGER REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS posted_at            TIMESTAMP,
    ADD COLUMN IF NOT EXISTS updated_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Unique idempotency constraint: prevent duplicate postings from offline sync retries
-- Only enforce when source_type and source_id are set (manual JEs excluded)
CREATE UNIQUE INDEX IF NOT EXISTS idx_journal_entries_source_idempotency
    ON journal_entries (source_type, source_id)
    WHERE source_type IS NOT NULL AND source_id IS NOT NULL AND status <> 'void';

-- Extend journal_lines into a full dimension-bearing line store
ALTER TABLE journal_lines
    ADD COLUMN IF NOT EXISTS line_no              INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS partner_id           INTEGER REFERENCES business_partners(id),
    -- partner_id set for AR/AP lines: identifies the customer/supplier
    ADD COLUMN IF NOT EXISTS profit_center_id     INTEGER REFERENCES profit_centers(id),
    ADD COLUMN IF NOT EXISTS store_id             INTEGER REFERENCES stores(id),
    -- store_id for store-level P&L analysis
    ADD COLUMN IF NOT EXISTS currency_code        VARCHAR(3) REFERENCES currencies(code) DEFAULT 'SAR',
    ADD COLUMN IF NOT EXISTS exchange_rate        DECIMAL(15,6) DEFAULT 1.000000,
    ADD COLUMN IF NOT EXISTS amount_doc_currency  DECIMAL(15,2) DEFAULT 0,  -- amount in document currency
    ADD COLUMN IF NOT EXISTS amount_base          DECIMAL(15,2) DEFAULT 0,  -- amount in org base currency
    ADD COLUMN IF NOT EXISTS debit_base           DECIMAL(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS credit_base          DECIMAL(15,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reference_type       VARCHAR(80),   -- line-level source doc type
    ADD COLUMN IF NOT EXISTS reference_id         VARCHAR(100),  -- line-level source doc PK
    ADD COLUMN IF NOT EXISTS reference_line       INTEGER,       -- source doc line number
    ADD COLUMN IF NOT EXISTS updated_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- =====================================================
-- GL POSTING RULES ENGINE
-- Replaces simple gl_account_mappings with a configurable rule table.
-- Finance team maintains this; use-cases call ResolveGLAccounts(posting_type).
-- =====================================================

CREATE TABLE gl_posting_rules (
    id                  SERIAL PRIMARY KEY,
    organization_id     INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    -- posting_type key uniquely identifies the business event:
    -- pos_cash_sale | pos_card_sale | pos_other_sale | pos_cogs |
    -- pos_return_cash | pos_return_stock |
    -- grn_inventory | grn_grir | grn_cogs |
    -- ap_invoice_payable | ap_invoice_vat_input | ap_invoice_grir |
    -- supplier_payment_cash | supplier_payment_bank |
    -- ar_invoice_receivable | ar_invoice_revenue | ar_invoice_vat_output |
    -- customer_payment_cash | customer_payment_bank |
    -- stock_adjustment_gain | stock_adjustment_loss |
    -- transfer_out | transfer_in |
    -- stock_count_variance_gain | stock_count_variance_loss |
    -- depreciation_expense | depreciation_accumulated |
    -- wht_receivable | wht_payable
    posting_type        VARCHAR(80) NOT NULL,
    debit_account_id    INTEGER REFERENCES chart_of_accounts(id),
    credit_account_id   INTEGER REFERENCES chart_of_accounts(id),
    tax_account_id      INTEGER REFERENCES chart_of_accounts(id),
    cost_center_id      INTEGER REFERENCES cost_centers(id),
    profit_center_id    INTEGER REFERENCES profit_centers(id),
    -- Optional store-level override (NULL = applies to all stores)
    store_id            INTEGER REFERENCES stores(id),
    -- Optional payment-method override (NULL = applies to all methods)
    payment_method      VARCHAR(50),
    description         TEXT,
    is_active           BOOLEAN DEFAULT true,
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    -- Uniqueness enforced via partial indexes below (store_id/payment_method can be NULL)
);

-- Partial unique indexes for gl_posting_rules specificity levels:
-- Level 0: generic rule (no store, no payment method override)
CREATE UNIQUE INDEX IF NOT EXISTS idx_gl_posting_rules_generic
    ON gl_posting_rules (organization_id, posting_type)
    WHERE store_id IS NULL AND payment_method IS NULL;

-- Level 1: store-specific, any payment method
CREATE UNIQUE INDEX IF NOT EXISTS idx_gl_posting_rules_store
    ON gl_posting_rules (organization_id, posting_type, store_id)
    WHERE store_id IS NOT NULL AND payment_method IS NULL;

-- Level 2: payment-method-specific, any store
CREATE UNIQUE INDEX IF NOT EXISTS idx_gl_posting_rules_payment
    ON gl_posting_rules (organization_id, posting_type, payment_method)
    WHERE store_id IS NULL AND payment_method IS NOT NULL;

-- Level 3: store + payment method specific
CREATE UNIQUE INDEX IF NOT EXISTS idx_gl_posting_rules_store_payment
    ON gl_posting_rules (organization_id, posting_type, store_id, payment_method)
    WHERE store_id IS NOT NULL AND payment_method IS NOT NULL;

-- =====================================================
-- ACCOUNT BALANCES
-- Materialized per-period summary — updated by PostJournalEntry + RunPeriodClose.
-- Enables fast trial balance and balance sheet queries without scanning journal_lines.
-- =====================================================

CREATE TABLE account_balances (
    id               BIGSERIAL PRIMARY KEY,
    organization_id  INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    account_id       INTEGER NOT NULL REFERENCES chart_of_accounts(id) ON DELETE CASCADE,
    posting_period_id INTEGER NOT NULL REFERENCES posting_periods(id) ON DELETE CASCADE,
    fiscal_year_id   INTEGER NOT NULL REFERENCES fiscal_years(id) ON DELETE CASCADE,
    opening_balance  DECIMAL(15,2) DEFAULT 0,  -- carried from prior period closing
    period_debits    DECIMAL(15,2) DEFAULT 0,
    period_credits   DECIMAL(15,2) DEFAULT 0,
    closing_balance  DECIMAL(15,2) DEFAULT 0,  -- recomputed on every post: opening + debits - credits
    currency_code    VARCHAR(3) REFERENCES currencies(code) DEFAULT 'SAR',
    last_updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (organization_id, account_id, posting_period_id)
);

-- =====================================================
-- WITHHOLDING TAX ENTRIES
-- KSA WHT on service payments to non-residents.
-- Created alongside supplier_payment postings.
-- =====================================================

CREATE TABLE withholding_tax_entries (
    id                  BIGSERIAL PRIMARY KEY,
    organization_id     INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    business_partner_id INTEGER REFERENCES business_partners(id) ON DELETE SET NULL,
    reference_type      VARCHAR(80) NOT NULL,    -- 'supplier_invoice' | 'supplier_payment'
    reference_id        VARCHAR(100) NOT NULL,   -- FK of the triggering document
    gross_amount        DECIMAL(15,2) NOT NULL,
    wht_rate            DECIMAL(5,2) NOT NULL,   -- e.g. 5.00 = 5%
    wht_amount          DECIMAL(15,2) NOT NULL,
    net_amount          DECIMAL(15,2) NOT NULL,  -- gross - wht
    posting_date        DATE NOT NULL DEFAULT CURRENT_DATE,
    gl_account_id       INTEGER REFERENCES chart_of_accounts(id),
    journal_entry_id    BIGINT REFERENCES journal_entries(id),
    is_remitted         BOOLEAN DEFAULT false,
    remitted_at         TIMESTAMP,
    notes               TEXT,
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (reference_type, reference_id)        -- one WHT entry per source document
);

-- =====================================================
-- INDEXES — ACCOUNTING ENGINE
-- =====================================================

-- fiscal_years
CREATE INDEX IF NOT EXISTS idx_fiscal_years_org_year
    ON fiscal_years (organization_id, year);
CREATE INDEX IF NOT EXISTS idx_fiscal_years_status
    ON fiscal_years (organization_id, status);

-- posting_periods
CREATE INDEX IF NOT EXISTS idx_posting_periods_org_fy
    ON posting_periods (organization_id, fiscal_year_id);
CREATE INDEX IF NOT EXISTS idx_posting_periods_dates
    ON posting_periods (organization_id, start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_posting_periods_status
    ON posting_periods (organization_id, status);

-- document_series
CREATE INDEX IF NOT EXISTS idx_document_series_org_type
    ON document_series (organization_id, doc_type, is_default);

-- profit_centers
CREATE INDEX IF NOT EXISTS idx_profit_centers_org
    ON profit_centers (organization_id, is_active);

-- extended journal_entries
CREATE INDEX IF NOT EXISTS idx_journal_entries_period
    ON journal_entries (organization_id, posting_period_id);
CREATE INDEX IF NOT EXISTS idx_journal_entries_status
    ON journal_entries (organization_id, status);
CREATE INDEX IF NOT EXISTS idx_journal_entries_source
    ON journal_entries (source_type, source_id)
    WHERE source_type IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_journal_entries_document_number
    ON journal_entries (organization_id, document_number)
    WHERE document_number IS NOT NULL;

-- extended journal_lines
CREATE INDEX IF NOT EXISTS idx_journal_lines_period_account
    ON journal_lines (account_id, journal_id);
CREATE INDEX IF NOT EXISTS idx_journal_lines_partner
    ON journal_lines (partner_id)
    WHERE partner_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_journal_lines_store
    ON journal_lines (store_id)
    WHERE store_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_journal_lines_reference
    ON journal_lines (reference_type, reference_id)
    WHERE reference_type IS NOT NULL;

-- gl_posting_rules
CREATE INDEX IF NOT EXISTS idx_gl_posting_rules_org_type
    ON gl_posting_rules (organization_id, posting_type, is_active);

-- account_balances
CREATE INDEX IF NOT EXISTS idx_account_balances_org_period
    ON account_balances (organization_id, posting_period_id);
CREATE INDEX IF NOT EXISTS idx_account_balances_account_period
    ON account_balances (account_id, posting_period_id);
CREATE INDEX IF NOT EXISTS idx_account_balances_fy
    ON account_balances (organization_id, fiscal_year_id);

-- withholding_tax_entries
CREATE INDEX IF NOT EXISTS idx_wht_entries_org
    ON withholding_tax_entries (organization_id, posting_date);
CREATE INDEX IF NOT EXISTS idx_wht_entries_partner
    ON withholding_tax_entries (business_partner_id)
    WHERE business_partner_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_wht_entries_remitted
    ON withholding_tax_entries (organization_id, is_remitted);

-- End of 55_accounting.sql
