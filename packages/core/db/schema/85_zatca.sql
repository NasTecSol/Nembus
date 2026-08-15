-- =====================================================
-- ENHANCED: CART MANAGEMENT SYSTEM
-- =====================================================

-- Cart status enum for better type safety


-- =====================================================
-- ENHANCED: UNIFIED ORDER MANAGEMENT (V2)
-- =====================================================


-- =====================================================
-- ENHANCED: INVOICE MANAGEMENT
-- =====================================================

CREATE TYPE invoice_type AS ENUM ('standard', 'proforma', 'credit_note', 'debit_note', 'recurring');
CREATE TYPE invoice_status AS ENUM ('draft', 'sent', 'viewed', 'partially_paid', 'paid', 'overdue', 'cancelled', 'refunded');

-- Invoices table
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    
    -- Customer information
    customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    customer_name VARCHAR(255) NOT NULL,
    customer_email VARCHAR(255),
    customer_phone VARCHAR(50),
    customer_tax_id VARCHAR(50),
    
    -- Invoice classification
    invoice_type invoice_type DEFAULT 'standard' NOT NULL,
    invoice_status invoice_status DEFAULT 'draft' NOT NULL,
    
    -- Source references
    sales_order_id UUID REFERENCES sales_orders_v2(id) ON DELETE SET NULL,
    related_invoice_id UUID REFERENCES invoices(id) ON DELETE SET NULL, -- For credit notes, etc.
    
    -- Dates
    invoice_date DATE NOT NULL DEFAULT CURRENT_DATE,
    due_date DATE NOT NULL,
    sent_date DATE,
    paid_date DATE,
    
    -- Financial details
    subtotal DECIMAL(15,2) DEFAULT 0.00,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    shipping_amount DECIMAL(15,2) DEFAULT 0.00,
    adjustment_amount DECIMAL(15,2) DEFAULT 0.00,
    total_amount DECIMAL(15,2) NOT NULL,
    
    -- Payment tracking
    paid_amount DECIMAL(15,2) DEFAULT 0.00,
    credit_applied DECIMAL(15,2) DEFAULT 0.00,
    balance_due DECIMAL(15,2) NOT NULL,
    
    -- Payment terms
    payment_terms VARCHAR(100), -- Net 30, Due on Receipt, etc.
    currency_code VARCHAR(3) DEFAULT 'USD',
    exchange_rate DECIMAL(15,6) DEFAULT 1.000000,
    
    -- Addresses
    billing_address JSONB NOT NULL,
    shipping_address JSONB,
    
    -- Recurring invoice settings
    is_recurring BOOLEAN DEFAULT false,
    recurrence_pattern VARCHAR(50), -- monthly, quarterly, annually
    next_invoice_date DATE,
    
    -- Document management
    pdf_url TEXT,
    document_hash VARCHAR(255), -- For integrity verification
    
    -- Communication tracking
    reminder_sent_count INTEGER DEFAULT 0,
    last_reminder_sent_at TIMESTAMP,
    
    -- Notes and references
    notes TEXT,
    internal_notes TEXT,
    reference_number VARCHAR(100), -- PO number, etc.
    
    -- User tracking
    created_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    tags TEXT[],
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- FIX #7 (P1): New sales_returns and sales_return_lines tables
CREATE TABLE sales_returns (
    id                  SERIAL PRIMARY KEY,
    return_number       VARCHAR(50) UNIQUE NOT NULL,
    store_id            INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    cashier_id          INTEGER REFERENCES cashiers(id) ON DELETE SET NULL,
    cashier_session_id  INTEGER REFERENCES cashier_sessions(id) ON DELETE SET NULL,
    customer_id         INTEGER REFERENCES customers(id) ON DELETE SET NULL,
    original_transaction_id INTEGER REFERENCES pos_transactions(id) ON DELETE SET NULL,
    return_date         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    return_reason       VARCHAR(255),
    status              VARCHAR(30) DEFAULT 'pending' CHECK (status IN ('pending','approved','completed','cancelled')),
    subtotal            DECIMAL(15,2) DEFAULT 0,
    tax_amount          DECIMAL(15,2) DEFAULT 0,
    total_refund_amount DECIMAL(15,2) DEFAULT 0,
    refund_method       VARCHAR(50),
    refund_reference    VARCHAR(100),
    approved_by         INTEGER REFERENCES users(id) ON DELETE SET NULL,
    notes               TEXT,
    metadata            JSONB     DEFAULT '{}',
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Invoice line items
CREATE TABLE invoice_lines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    line_number INTEGER NOT NULL,
    
    -- Item description
    description TEXT NOT NULL,
    item_type VARCHAR(50) DEFAULT 'product', -- product, service, discount, shipping, fee
    
    -- Product reference (optional)
    product_id INTEGER REFERENCES products(id) ON DELETE SET NULL,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
    product_sku VARCHAR(100),
    
    -- Order line reference
    order_line_id UUID REFERENCES sales_order_lines_v2(id) ON DELETE SET NULL,
    
    -- Quantity and pricing
    quantity DECIMAL(15,3) DEFAULT 1.000,
    unit_price DECIMAL(15,2) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    line_total DECIMAL(15,2) NOT NULL,
    
    -- Tax details
    tax_category_id INTEGER REFERENCES tax_categories(id) ON DELETE SET NULL,
    tax_rate DECIMAL(5,2),
    
    -- UOM
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(invoice_id, line_number)
);

-- Invoice payments
CREATE TABLE invoice_payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    payment_number VARCHAR(50) UNIQUE NOT NULL,
    
    -- Payment details
    payment_date DATE NOT NULL DEFAULT CURRENT_DATE,
    payment_amount DECIMAL(15,2) NOT NULL CHECK (payment_amount > 0),
    
    payment_method VARCHAR(100) NOT NULL, -- cash, card, bank_transfer, check, etc.
    payment_gateway VARCHAR(100),
    payment_reference VARCHAR(255), -- Transaction ID, check number, etc.
    
    -- Currency handling
    currency_code VARCHAR(3) DEFAULT 'USD',
    exchange_rate DECIMAL(15,6) DEFAULT 1.000000,
    
    -- Bank reconciliation
    bank_account_id INTEGER, -- Reference to bank accounts if you have that table
    reconciled BOOLEAN DEFAULT false,
    reconciled_date DATE,
    
    notes TEXT,
    
    -- User tracking
    received_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Invoice status history
CREATE TABLE invoice_status_history (
    id BIGSERIAL PRIMARY KEY,
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    from_status invoice_status,
    to_status invoice_status NOT NULL,
    
    reason VARCHAR(255),
    notes TEXT,
    
    changed_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- ENHANCED: QUOTE MANAGEMENT
-- =====================================================

CREATE TYPE quote_status AS ENUM ('draft', 'sent', 'viewed', 'accepted', 'declined', 'expired', 'converted');

CREATE TABLE quotes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quote_number VARCHAR(50) UNIQUE NOT NULL,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    
    -- Customer information
    customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
    customer_name VARCHAR(255) NOT NULL,
    customer_email VARCHAR(255),
    customer_phone VARCHAR(50),
    
    -- Quote status and validity
    quote_status quote_status DEFAULT 'draft' NOT NULL,
    
    -- Dates
    quote_date DATE NOT NULL DEFAULT CURRENT_DATE,
    valid_until DATE NOT NULL,
    sent_date DATE,
    accepted_date DATE,
    converted_date DATE,
    
    -- Financial totals
    subtotal DECIMAL(15,2) DEFAULT 0.00,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    total_amount DECIMAL(15,2) NOT NULL,
    
    -- Conversion
    converted_to_order_id UUID REFERENCES sales_orders_v2(id) ON DELETE SET NULL,
    
    -- Terms and conditions
    payment_terms VARCHAR(100),
    delivery_terms TEXT,
    terms_and_conditions TEXT,
    
    -- Notes
    notes TEXT,
    internal_notes TEXT,
    
    -- User tracking
    created_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE combo_bundle_items (
    id               SERIAL PRIMARY KEY,
    combo_bundle_id  INTEGER NOT NULL REFERENCES combo_bundles(id) ON DELETE CASCADE,
    menu_item_id     INTEGER REFERENCES menu_items(id) ON DELETE CASCADE,
    product_id       INTEGER REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
    item_type        VARCHAR(20) DEFAULT 'menu_item' CHECK (item_type IN ('menu_item','product')),
    quantity         DECIMAL(15,3) DEFAULT 1,
    is_required      BOOLEAN   DEFAULT true,
    group_tag        VARCHAR(50),
    price_override   DECIMAL(15,2),
    display_order    INTEGER   DEFAULT 0,
    metadata         JSONB     DEFAULT '{}',
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Quote line items
CREATE TABLE quote_lines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quote_id UUID NOT NULL REFERENCES quotes(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    line_number INTEGER NOT NULL,
    
    -- Product information
    product_id INTEGER REFERENCES products(id) ON DELETE SET NULL,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
    description TEXT NOT NULL,
    
    -- Quantity and pricing
    quantity DECIMAL(15,3) NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(15,2) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    line_total DECIMAL(15,2) NOT NULL,
    
    -- UOM
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(quote_id, line_number)
);


CREATE TABLE sales_return_lines (
    id                 SERIAL PRIMARY KEY,
    return_id          INTEGER NOT NULL REFERENCES sales_returns(id) ON DELETE CASCADE,
    product_id         INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
    original_line_id   INTEGER REFERENCES pos_transaction_lines(id) ON DELETE SET NULL,
    quantity           DECIMAL(15,3) NOT NULL,
    unit_price         DECIMAL(15,4) NOT NULL,
    refund_amount      DECIMAL(15,2) NOT NULL,
    return_to_stock    BOOLEAN   DEFAULT true,
    serial_number      VARCHAR(100),
    batch_number       VARCHAR(100),
    condition          VARCHAR(50) DEFAULT 'good' CHECK (condition IN ('good','damaged','defective','opened')),
    line_number        INTEGER,
    metadata           JSONB     DEFAULT '{}',
    created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- =====================================================
-- INDEXES FOR CART SYSTEM
-- =====================================================

CREATE INDEX idx_carts_organization_id ON carts(organization_id);
CREATE INDEX idx_carts_store_id ON carts(store_id);
CREATE INDEX idx_carts_customer_id ON carts(customer_id);
CREATE INDEX idx_carts_cart_status ON carts(cart_status);
CREATE INDEX idx_carts_cart_type ON carts(cart_type);
CREATE INDEX idx_carts_cart_number ON carts(cart_number);
CREATE INDEX idx_carts_guest_identifier ON carts(guest_identifier);
CREATE INDEX idx_carts_created_at ON carts(created_at);
CREATE INDEX idx_carts_last_activity_at ON carts(last_activity_at);
CREATE INDEX idx_carts_expires_at ON carts(expires_at);
CREATE INDEX idx_carts_channel ON carts(channel);

CREATE INDEX idx_cart_items_cart_id ON cart_items(cart_id);
CREATE INDEX idx_cart_items_product_id ON cart_items(product_id);
CREATE INDEX idx_cart_items_product_variant_id ON cart_items(product_variant_id);
CREATE INDEX idx_cart_items_added_at ON cart_items(added_at);

CREATE INDEX idx_cart_activity_log_cart_id ON cart_activity_log(cart_id);
CREATE INDEX idx_cart_activity_log_activity_type ON cart_activity_log(activity_type);
CREATE INDEX idx_cart_activity_log_created_at ON cart_activity_log(created_at);

-- =====================================================
-- INDEXES FOR DRAFT CART TEMPLATES
-- =====================================================

CREATE INDEX idx_draft_cart_templates_organization_id ON draft_cart_templates(organization_id);
CREATE INDEX idx_draft_cart_templates_customer_id ON draft_cart_templates(customer_id);
CREATE INDEX idx_draft_cart_templates_template_type ON draft_cart_templates(template_type);
CREATE INDEX idx_draft_cart_templates_is_favorite ON draft_cart_templates(is_favorite);
CREATE INDEX idx_draft_cart_templates_auto_reorder ON draft_cart_templates(auto_reorder_enabled);
CREATE INDEX idx_draft_cart_templates_next_reorder_date ON draft_cart_templates(next_reorder_date);

CREATE INDEX idx_draft_cart_template_items_template_id ON draft_cart_template_items(template_id);
CREATE INDEX idx_draft_cart_template_items_product_id ON draft_cart_template_items(product_id);

-- =====================================================
-- INDEXES FOR ENHANCED ORDERS
-- =====================================================

CREATE INDEX idx_sales_orders_v2_organization_id ON sales_orders_v2(organization_id);
CREATE INDEX idx_sales_orders_v2_store_id ON sales_orders_v2(store_id);
CREATE INDEX idx_sales_orders_v2_customer_id ON sales_orders_v2(customer_id);
CREATE INDEX idx_sales_orders_v2_order_number ON sales_orders_v2(order_number);
CREATE INDEX idx_sales_orders_v2_order_status ON sales_orders_v2(order_status);
CREATE INDEX idx_sales_orders_v2_payment_status ON sales_orders_v2(payment_status);
CREATE INDEX idx_sales_orders_v2_fulfillment_status ON sales_orders_v2(fulfillment_status);
CREATE INDEX idx_sales_orders_v2_order_date ON sales_orders_v2(order_date);
CREATE INDEX idx_sales_orders_v2_order_type ON sales_orders_v2(order_type);
CREATE INDEX idx_sales_orders_v2_sales_channel ON sales_orders_v2(sales_channel);
CREATE INDEX idx_sales_orders_v2_source_cart_id ON sales_orders_v2(source_cart_id);
CREATE INDEX idx_sales_orders_v2_created_at ON sales_orders_v2(created_at);

CREATE INDEX idx_sales_order_lines_v2_sales_order_id ON sales_order_lines_v2(sales_order_id);
CREATE INDEX idx_sales_order_lines_v2_product_id ON sales_order_lines_v2(product_id);
CREATE INDEX idx_sales_order_lines_v2_product_variant_id ON sales_order_lines_v2(product_variant_id);
CREATE INDEX idx_sales_order_lines_v2_line_status ON sales_order_lines_v2(line_status);

CREATE INDEX idx_order_status_history_sales_order_id ON order_status_history(sales_order_id);
CREATE INDEX idx_order_status_history_changed_at ON order_status_history(changed_at);

CREATE INDEX idx_order_fulfillments_sales_order_id ON order_fulfillments(sales_order_id);
CREATE INDEX idx_order_fulfillments_fulfillment_number ON order_fulfillments(fulfillment_number);
CREATE INDEX idx_order_fulfillments_fulfillment_status ON order_fulfillments(fulfillment_status);
CREATE INDEX idx_order_fulfillments_shipment_status ON order_fulfillments(shipment_status);

CREATE INDEX idx_order_fulfillment_items_fulfillment_id ON order_fulfillment_items(fulfillment_id);
CREATE INDEX idx_order_fulfillment_items_order_line_id ON order_fulfillment_items(order_line_id);

-- =====================================================
-- INDEXES FOR INVOICES
-- =====================================================

CREATE INDEX idx_invoices_organization_id ON invoices(organization_id);
CREATE INDEX idx_invoices_store_id ON invoices(store_id);
CREATE INDEX idx_invoices_customer_id ON invoices(customer_id);
CREATE INDEX idx_invoices_invoice_number ON invoices(invoice_number);
CREATE INDEX idx_invoices_invoice_status ON invoices(invoice_status);
CREATE INDEX idx_invoices_invoice_type ON invoices(invoice_type);
CREATE INDEX idx_invoices_invoice_date ON invoices(invoice_date);
CREATE INDEX idx_invoices_due_date ON invoices(due_date);
CREATE INDEX idx_invoices_sales_order_id ON invoices(sales_order_id);
CREATE INDEX idx_invoices_is_recurring ON invoices(is_recurring);
CREATE INDEX idx_invoices_next_invoice_date ON invoices(next_invoice_date);

CREATE INDEX idx_invoice_lines_invoice_id ON invoice_lines(invoice_id);
CREATE INDEX idx_invoice_lines_product_id ON invoice_lines(product_id);
CREATE INDEX idx_invoice_lines_order_line_id ON invoice_lines(order_line_id);

CREATE INDEX idx_invoice_payments_invoice_id ON invoice_payments(invoice_id);
CREATE INDEX idx_invoice_payments_payment_date ON invoice_payments(payment_date);
CREATE INDEX idx_invoice_payments_payment_number ON invoice_payments(payment_number);
CREATE INDEX idx_invoice_payments_reconciled ON invoice_payments(reconciled);

CREATE INDEX idx_invoice_status_history_invoice_id ON invoice_status_history(invoice_id);
CREATE INDEX idx_invoice_status_history_changed_at ON invoice_status_history(changed_at);

-- =====================================================
-- INDEXES FOR QUOTES
-- =====================================================

CREATE INDEX idx_quotes_organization_id ON quotes(organization_id);
CREATE INDEX idx_quotes_customer_id ON quotes(customer_id);
CREATE INDEX idx_quotes_quote_number ON quotes(quote_number);
CREATE INDEX idx_quotes_quote_status ON quotes(quote_status);
CREATE INDEX idx_quotes_quote_date ON quotes(quote_date);
CREATE INDEX idx_quotes_valid_until ON quotes(valid_until);
CREATE INDEX idx_quotes_converted_to_order_id ON quotes(converted_to_order_id);

CREATE INDEX idx_quote_lines_quote_id ON quote_lines(quote_id);
CREATE INDEX idx_quote_lines_product_id ON quote_lines(product_id);

-- =====================================================
-- TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES
-- =====================================================

-- Function to update updated_at timestamp
-- Generic updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $func$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$func$ LANGUAGE plpgsql;

-- Apply updated_at triggers
DO $$
DECLARE
    tbl TEXT;
    tbls TEXT[] := ARRAY[
        'organizations','tenants','modules','menus','submenus','roles',
        'ui_settings','role_ui_customizations','stores','users',
        'cashiers','pos_terminals','product_categories','brands',
        'price_lists','products','product_variants','product_prices',
        'product_serial_numbers','product_batches','inventory_stock',
        'stock_reservations','suppliers','customers','purchase_orders',
        'pos_transactions','sales_returns','restaurant_tables',
        'menu_categories','recipes','menu_modifier_groups','menu_items',
        'combo_bundles','sales_analytics','purchase_analytics',
        'inventory_analytics','profit_loss_analytics','discount_analytics',
        'carts','cart_items','draft_cart_templates','draft_cart_template_items',
        'sales_orders_v2','sales_order_lines_v2','order_fulfillments',
        'invoices','invoice_lines','invoice_payments','quotes','quote_lines',
        'promotions','loyalty_redemption_rules','restaurant_orders',
        'restaurant_order_items', 'cashier_sessions'
    ];
BEGIN
    FOREACH tbl IN ARRAY tbls LOOP
        EXECUTE format(
            'DROP TRIGGER IF EXISTS trg_%s_updated_at ON %I;
             CREATE TRIGGER trg_%s_updated_at BEFORE UPDATE ON %I
             FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();',
            tbl, tbl, tbl, tbl
        );
    END LOOP;
END;
$$;

-- =====================================================
-- TRIGGERS FOR CART ACTIVITY TRACKING
-- =====================================================

-- =====================================================
-- TRIGGERS FOR CASHIER SESSION BALANCE TRACKING
-- =====================================================

-- DB trigger trg_update_cashier_session_balance dropped in favor of explicit app-level balance updates.
DROP TRIGGER IF EXISTS trg_update_cashier_session_balance ON pos_transactions;
DROP FUNCTION IF EXISTS update_cashier_session_balance();


-- Log cart status changes
CREATE OR REPLACE FUNCTION log_cart_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.cart_status IS DISTINCT FROM NEW.cart_status THEN
        INSERT INTO cart_activity_log (
            cart_id, 
            organization_id, 
            activity_type, 
            description, 
            old_value, 
            new_value
        )
        VALUES (
            NEW.id,
            NEW.organization_id,
            'status_changed',
            'Cart status changed from ' || OLD.cart_status || ' to ' || NEW.cart_status,
            jsonb_build_object('status', OLD.cart_status),
            jsonb_build_object('status', NEW.cart_status)
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cart_status_change_trigger
    AFTER UPDATE ON carts
    FOR EACH ROW
    WHEN (OLD.cart_status IS DISTINCT FROM NEW.cart_status)
    EXECUTE FUNCTION log_cart_status_change();

-- Update cart last_activity_at when items change
CREATE OR REPLACE FUNCTION update_cart_activity()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE carts 
    SET last_activity_at = CURRENT_TIMESTAMP
    WHERE id = COALESCE(NEW.cart_id, OLD.cart_id);
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cart_items_activity_trigger
    AFTER INSERT OR UPDATE OR DELETE ON cart_items
    FOR EACH ROW
    EXECUTE FUNCTION update_cart_activity();


-- ZATCA PHASE 2 COMPLIANCE
-- =====================================================

-- ZATCA document submission status
CREATE TYPE zatca_doc_status AS ENUM (
    'pending',      -- Awaiting submission
    'cleared',      -- ZATCA cleared (B2B Standard)
    'reported',     -- ZATCA reported (B2C Simplified)
    'warning',      -- Cleared/reported with warnings
    'rejected',     -- ZATCA rejected
    'failed'        -- Network/system failure
);

-- Per-device (EGS unit) cryptographic configuration
-- Stores CSIDs for both Cloud server (B2B) and POS terminals (B2C)
CREATE TABLE zatca_device_configs (
    id              SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    store_id        INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    pos_terminal_id INTEGER REFERENCES pos_terminals(id) ON DELETE SET NULL,

    -- Device identity
    device_serial   VARCHAR(255) NOT NULL,        -- EGS serial number
    device_type     VARCHAR(20) NOT NULL,         -- 'cloud' or 'pos'

    -- Cryptographic material (encrypted at rest)
    csr_pem         TEXT,                          -- Certificate Signing Request
    private_key_pem TEXT NOT NULL,                 -- ECDSA secp256k1 private key (PEM)
    compliance_csid TEXT,                          -- Compliance CSID (temporary, used during onboarding)
    production_csid TEXT,                          -- Production CSID (active signing certificate)
    csid_expiry     TIMESTAMPTZ,                  -- Certificate expiry date

    -- ZATCA environment
    zatca_env       VARCHAR(20) DEFAULT 'sandbox', -- 'sandbox' or 'production'

    -- Status
    is_active       BOOLEAN DEFAULT true,
    is_revoked      BOOLEAN DEFAULT false,
    revoked_at      TIMESTAMPTZ,
    revoked_reason  TEXT,

    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(organization_id, device_serial)
);

-- Lightweight sequential chaining ledger
-- Tracks the cryptographic hash chain per device for ZATCA compliance
CREATE TABLE zatca_document_chain (
    id               BIGSERIAL PRIMARY KEY,

    -- Link to source document (invoice or POS transaction)
    entity_type      VARCHAR(20) NOT NULL,          -- 'invoice' or 'pos_transaction'
    entity_id        TEXT NOT NULL,                  -- UUID for invoices, integer cast to text for pos_txn

    -- Device that signed this document
    device_config_id INTEGER NOT NULL REFERENCES zatca_device_configs(id),
    organization_id  INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- ZATCA sequential fields
    zatca_uuid       UUID NOT NULL DEFAULT uuid_generate_v4(),
    icv              BIGINT NOT NULL,                -- Invoice Counter Value (sequential per device)
    pih              TEXT NOT NULL,                   -- Previous Invoice Hash (Base64 SHA-256)
    xml_hash         TEXT NOT NULL,                   -- This document's XML hash (Base64 SHA-256)

    -- ZATCA API response
    zatca_status     zatca_doc_status DEFAULT 'pending',
    zatca_response   JSONB DEFAULT '{}',             -- Full API response payload
    qr_code_base64   TEXT,                           -- TLV QR code (Base64 encoded)
    signed_xml       TEXT,                           -- Full signed UBL 2.1 XML document

    -- Submission tracking
    submitted_at     TIMESTAMPTZ,
    cleared_at       TIMESTAMPTZ,

    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Ensure sequential ICV per device (no gaps allowed by ZATCA)
    UNIQUE(device_config_id, icv)
);

-- Index for fast chain lookups (latest entry per device)
CREATE INDEX idx_zatca_chain_device_icv ON zatca_document_chain(device_config_id, icv DESC);
-- Index for entity lookups (find chain entry for a specific invoice/transaction)
CREATE INDEX idx_zatca_chain_entity ON zatca_document_chain(entity_type, entity_id);
-- Index for pending/failed entries (reporting worker picks these up)
CREATE INDEX idx_zatca_chain_status ON zatca_document_chain(zatca_status) WHERE zatca_status IN ('pending', 'failed');

-- Sync watermarks for delta-fetch mechanism (Cloud ↔ POS)
CREATE TABLE sync_watermarks (
    id              SERIAL PRIMARY KEY,
    entity_type     VARCHAR(50) NOT NULL,           -- 'zatca_config', 'orders', 'inventory', etc.
    store_id        INTEGER REFERENCES stores(id) ON DELETE CASCADE,
    last_sync_at    TIMESTAMPTZ NOT NULL DEFAULT '1970-01-01',
    metadata        JSONB DEFAULT '{}',

    UNIQUE(entity_type, store_id)
);

