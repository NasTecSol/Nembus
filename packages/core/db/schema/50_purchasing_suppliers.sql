-- =====================================================
-- SUPPLIERS & CUSTOMERS
-- =====================================================

-- =====================================================
-- CURRENCIES, PAYMENT TERMS & BUSINESS PARTNERS
-- (must be defined before purchase_orders and goods_receipt_notes)
-- =====================================================

CREATE TABLE IF NOT EXISTS currencies (
    code VARCHAR(3) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    decimal_places INTEGER DEFAULT 2,
    is_active BOOLEAN DEFAULT true
);

CREATE TABLE IF NOT EXISTS exchange_rates (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    from_currency VARCHAR(3) NOT NULL REFERENCES currencies(code),
    to_currency VARCHAR(3) NOT NULL REFERENCES currencies(code),
    rate_date DATE NOT NULL,
    rate DECIMAL(18,6) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, from_currency, to_currency, rate_date)
);

CREATE TABLE IF NOT EXISTS payment_terms (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    due_days INTEGER NOT NULL DEFAULT 0,
    discount_days INTEGER DEFAULT 0,
    discount_percentage DECIMAL(5,2) DEFAULT 0.00,
    late_fee_percentage DECIMAL(5,2) DEFAULT 0.00,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS cost_centers (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    dimension VARCHAR(50) DEFAULT 'general',
    is_active BOOLEAN DEFAULT true
);

CREATE TABLE IF NOT EXISTS chart_of_accounts (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    account_code VARCHAR(50) UNIQUE NOT NULL,
    account_name VARCHAR(255) NOT NULL,
    account_type VARCHAR(30) NOT NULL CHECK (account_type IN ('asset','liability','equity','revenue','expense')),
    parent_account_id INTEGER REFERENCES chart_of_accounts(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS gl_account_mappings (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    mapping_type VARCHAR(50) NOT NULL,
    store_id INTEGER REFERENCES stores(id) ON DELETE CASCADE,
    gl_account_id INTEGER NOT NULL REFERENCES chart_of_accounts(id) ON DELETE CASCADE,
    UNIQUE(organization_id, mapping_type, store_id)
);

CREATE TABLE IF NOT EXISTS business_partners (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    partner_role VARCHAR(20) NOT NULL CHECK (partner_role IN ('supplier','vendor','special_customer','corporate_group')),
    tax_id VARCHAR(50),
    currency_code VARCHAR(3) REFERENCES currencies(code) DEFAULT 'SAR',
    credit_limit DECIMAL(15,2) DEFAULT 0.00,
    outstanding_balance DECIMAL(15,2) DEFAULT 0.00,
    payment_terms_id INTEGER REFERENCES payment_terms(id) ON DELETE SET NULL,
    sales_rep_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS partner_addresses (
    id SERIAL PRIMARY KEY,
    partner_id INTEGER NOT NULL REFERENCES business_partners(id) ON DELETE CASCADE,
    address_name VARCHAR(100) NOT NULL,
    address_type VARCHAR(20) NOT NULL CHECK (address_type IN ('bill_to','ship_to','both')),
    street TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    zip_code VARCHAR(20),
    country_code VARCHAR(3) DEFAULT 'SA',
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS partner_contacts (
    id SERIAL PRIMARY KEY,
    partner_id INTEGER NOT NULL REFERENCES business_partners(id) ON DELETE CASCADE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    email VARCHAR(255),
    phone VARCHAR(50),
    position VARCHAR(100),
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- CUSTOMERS
-- =====================================================

CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    customer_code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    address TEXT,
    customer_type VARCHAR(50),
    price_list_id INTEGER REFERENCES price_lists(id) ON DELETE SET NULL,
    credit_limit DECIMAL(15,2) DEFAULT 0,
    outstanding_balance DECIMAL(15,2) DEFAULT 0,
    loyalty_points DECIMAL(15,2) DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, customer_code)
);

-- =====================================================
-- PURCHASE & SALES ORDERS
-- =====================================================

CREATE TABLE purchase_orders (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    po_number VARCHAR(50) UNIQUE NOT NULL,
    partners_id INTEGER NOT NULL REFERENCES business_partners(id) ON DELETE CASCADE,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    po_date DATE NOT NULL,
    expected_delivery_date DATE,
    status VARCHAR(50) DEFAULT 'draft',
    subtotal DECIMAL(15,2) DEFAULT 0,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) DEFAULT 0,
    price_list_id INTEGER REFERENCES price_lists(id) ON DELETE SET NULL,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    approved_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE purchase_order_lines (
    id SERIAL PRIMARY KEY,
    purchase_order_id INTEGER NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
    quantity DECIMAL(15,3) NOT NULL,
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    unit_price DECIMAL(15,4) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    subtotal DECIMAL(15,2) NOT NULL,
    line_total DECIMAL(15,2) DEFAULT 0,
    received_quantity DECIMAL(15,3) DEFAULT 0,
    line_number INTEGER,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transfer_requests (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    transfer_number VARCHAR(50) UNIQUE NOT NULL,
    from_store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    to_store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    requested_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    approved_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    shipped_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    received_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    request_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expected_delivery_date DATE,
    shipped_at TIMESTAMP,
    received_at TIMESTAMP,
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transfer_request_items (
    id SERIAL PRIMARY KEY,
    transfer_request_id INTEGER NOT NULL REFERENCES transfer_requests(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
    from_location_id INTEGER REFERENCES storage_locations(id) ON DELETE SET NULL,
    to_location_id INTEGER REFERENCES storage_locations(id) ON DELETE SET NULL,
    requested_quantity DECIMAL(15,3) NOT NULL,
    approved_quantity DECIMAL(15,3) DEFAULT 0,
    shipped_quantity DECIMAL(15,3) DEFAULT 0,
    received_quantity DECIMAL(15,3) DEFAULT 0,
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    batch_number VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE goods_receipt_notes (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    grn_number VARCHAR(50) UNIQUE NOT NULL,
    purchase_order_id INTEGER REFERENCES purchase_orders(id) ON DELETE SET NULL,
    partners_id INTEGER NOT NULL REFERENCES business_partners(id) ON DELETE CASCADE,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    received_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    receipt_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    delivery_note_number VARCHAR(100),
    status VARCHAR(50) DEFAULT 'posted',
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE goods_receipt_note_items (
    id SERIAL PRIMARY KEY,
    grn_id INTEGER NOT NULL REFERENCES goods_receipt_notes(id) ON DELETE CASCADE,
    purchase_order_line_id INTEGER REFERENCES purchase_order_lines(id) ON DELETE SET NULL,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
    storage_location_id INTEGER REFERENCES storage_locations(id) ON DELETE SET NULL,
    quantity_received DECIMAL(15,3) NOT NULL,
    quantity_rejected DECIMAL(15,3) DEFAULT 0,
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    unit_cost DECIMAL(15,4),
    batch_number VARCHAR(100),
    expiry_date DATE,
    rejection_reason TEXT,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sales_orders (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    order_date DATE NOT NULL,
    delivery_date DATE,
    status VARCHAR(50) DEFAULT 'draft',
    subtotal DECIMAL(15,2) DEFAULT 0,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) DEFAULT 0,
    price_list_id INTEGER REFERENCES price_lists(id) ON DELETE SET NULL,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sales_order_lines (
    id SERIAL PRIMARY KEY,
    sales_order_id INTEGER NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
    quantity DECIMAL(15,3) NOT NULL,
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    unit_price DECIMAL(15,4) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    subtotal DECIMAL(15,2) NOT NULL,
    line_total DECIMAL(15,2) DEFAULT 0,
    shipped_quantity DECIMAL(15,3) DEFAULT 0,
    line_number INTEGER,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TYPE order_type AS ENUM ('standard', 'quote', 'subscription', 'return', 'exchange');
CREATE TYPE order_status_v2 AS ENUM (
    'draft', 'pending', 'confirmed', 'processing', 
    'partially_fulfilled', 'fulfilled', 'partially_shipped', 'shipped', 
    'delivered', 'cancelled', 'refunded', 'on_hold'
);
CREATE TYPE payment_status AS ENUM ('unpaid', 'partially_paid', 'paid', 'refunded', 'partially_refunded', 'overdue');
CREATE TYPE fulfillment_status AS ENUM ('unfulfilled', 'partially_fulfilled', 'fulfilled', 'restocked');

CREATE TYPE cart_status AS ENUM ('draft', 'active', 'abandoned', 'converted', 'expired');
CREATE TYPE cart_type AS ENUM ('standard', 'quote', 'saved', 'wishlist', 'retail', 'wholesale');

-- Main carts table - supports both guest and registered customers
CREATE TABLE carts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cart_number VARCHAR(50) UNIQUE NOT NULL, -- Human-readable cart reference
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    
    -- Customer information
    customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL, -- NULL for guest carts
    guest_identifier VARCHAR(255), -- Session ID or device ID for guest users
    guest_email VARCHAR(255),
    guest_phone VARCHAR(50),
    
    -- Cart classification
    cart_status cart_status DEFAULT 'draft' NOT NULL,
    cart_type cart_type DEFAULT 'standard' NOT NULL,
    
    -- Sales channel tracking
    channel VARCHAR(50) DEFAULT 'online', -- online, pos, mobile_app, kiosk
    payment_method VARCHAR(100),
    payment_gateway VARCHAR(100),
    device_info JSONB DEFAULT '{}', -- Browser, device type, etc.
    
    -- User/Session tracking
    created_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    cashier_id INTEGER REFERENCES cashiers(id) ON DELETE SET NULL,
    pos_terminal_id INTEGER REFERENCES pos_terminals(id) ON DELETE SET NULL,
    
    -- Pricing and totals
    subtotal DECIMAL(15,2) DEFAULT 0.00,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    shipping_amount DECIMAL(15,2) DEFAULT 0.00,
    total_amount DECIMAL(15,2) DEFAULT 0.00,
    
    -- Applied discounts and promotions
    coupon_code VARCHAR(100),
    discount_code VARCHAR(100),
    promotional_credits DECIMAL(15,2) DEFAULT 0.00,
    
    -- Shipping information
    shipping_address JSONB,
    billing_address JSONB,
    shipping_method VARCHAR(100),
    
    -- Conversion tracking
    converted_to_order_id UUID, -- Reference to sales_orders_v2.id
    converted_at TIMESTAMP,
    
    -- Lifecycle timestamps
    last_activity_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP, -- For abandoned cart cleanup
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Metadata for extensibility
    metadata JSONB DEFAULT '{}', -- Custom fields, analytics tags, etc.
    notes TEXT,
    
    -- Constraints
    CONSTRAINT chk_cart_customer CHECK (
        customer_id IS NOT NULL OR guest_identifier IS NOT NULL
    )
);

-- Cart line items
CREATE TABLE cart_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cart_id UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Product information
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
    
    -- Quantity and UOM
    quantity DECIMAL(15,3) NOT NULL CHECK (quantity > 0),
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    
    -- Pricing
    unit_price DECIMAL(15,2) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    line_total DECIMAL(15,2) NOT NULL,
    
    -- Pricing details
    price_list_id INTEGER REFERENCES price_lists(id) ON DELETE SET NULL,
    tax_category_id INTEGER REFERENCES tax_categories(id) ON DELETE SET NULL,
    
    -- Inventory tracking
    batch_number VARCHAR(100),
    serial_number VARCHAR(100),
    
    -- Customization and options
    customization_details JSONB DEFAULT '{}', -- Product customizations, engravings, etc.
    notes TEXT,
    
    -- Line item metadata
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Prevent duplicate products (unless variants differ)
    UNIQUE(cart_id, product_id, product_variant_id, batch_number, serial_number)
);

-- Cart activity log for tracking changes
CREATE TABLE cart_activity_log (
    id BIGSERIAL PRIMARY KEY,
    cart_id UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    activity_type VARCHAR(50) NOT NULL, -- created, item_added, item_removed, item_updated, status_changed, converted, abandoned
    description TEXT,
    
    -- User tracking
    performed_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    ip_address INET,
    user_agent TEXT,
    
    -- Change details
    old_value JSONB,
    new_value JSONB,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- BP PRICE CONTRACTS, JOURNAL ENTRIES & LINES
-- =====================================================

CREATE TABLE IF NOT EXISTS bp_price_contracts (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    partner_id INTEGER NOT NULL REFERENCES business_partners(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
    contract_price DECIMAL(15,4) NOT NULL,
    discount_percentage DECIMAL(5,2) DEFAULT 0.00,
    min_quantity DECIMAL(15,3) DEFAULT 1,
    valid_from DATE,
    valid_to DATE,
    is_active BOOLEAN DEFAULT true,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(partner_id, product_id, product_variant_id)
);

CREATE TABLE IF NOT EXISTS journal_entries (
    id BIGSERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    entry_number VARCHAR(50) UNIQUE NOT NULL,
    posting_date DATE NOT NULL DEFAULT CURRENT_DATE,
    reference_type VARCHAR(50) NOT NULL, -- 'pos_transaction' | 'grn' | 'sales_order'
    reference_id VARCHAR(100) NOT NULL,
    memo TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS journal_lines (
    id BIGSERIAL PRIMARY KEY,
    journal_id BIGINT NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    account_id INTEGER NOT NULL REFERENCES chart_of_accounts(id) ON DELETE RESTRICT,
    cost_center_id INTEGER REFERENCES cost_centers(id) ON DELETE SET NULL,
    debit DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    credit DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    memo TEXT
);

-- =====================================================
-- INDEXES
-- =====================================================

CREATE INDEX IF NOT EXISTS idx_business_partners_organization_id ON business_partners(organization_id);
CREATE INDEX IF NOT EXISTS idx_business_partners_code ON business_partners(code);
CREATE INDEX IF NOT EXISTS idx_business_partners_partner_role ON business_partners(partner_role);
CREATE INDEX IF NOT EXISTS idx_business_partners_is_active ON business_partners(is_active);
CREATE INDEX IF NOT EXISTS idx_partner_addresses_partner_id ON partner_addresses(partner_id);
CREATE INDEX IF NOT EXISTS idx_partner_contacts_partner_id ON partner_contacts(partner_id);
CREATE INDEX IF NOT EXISTS idx_chart_of_accounts_organization_id ON chart_of_accounts(organization_id);
CREATE INDEX IF NOT EXISTS idx_chart_of_accounts_type ON chart_of_accounts(account_type);
CREATE INDEX IF NOT EXISTS idx_gl_account_mappings_org_type ON gl_account_mappings(organization_id, mapping_type);
CREATE INDEX IF NOT EXISTS idx_journal_entries_organization_id ON journal_entries(organization_id);
CREATE INDEX IF NOT EXISTS idx_journal_entries_posting_date ON journal_entries(posting_date);
CREATE INDEX IF NOT EXISTS idx_journal_entries_reference ON journal_entries(reference_type, reference_id);
CREATE INDEX IF NOT EXISTS idx_journal_lines_journal_id ON journal_lines(journal_id);
CREATE INDEX IF NOT EXISTS idx_journal_lines_account_id ON journal_lines(account_id);
CREATE INDEX IF NOT EXISTS idx_bp_price_contracts_bp_product ON bp_price_contracts(partner_id, product_id);
CREATE INDEX IF NOT EXISTS idx_bp_price_contracts_is_active ON bp_price_contracts(is_active);

-- End of 50_purchasing_suppliers.sql
