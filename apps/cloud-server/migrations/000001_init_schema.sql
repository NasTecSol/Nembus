

-- +goose Up
-- Combined Initial Schema Migration: Base Tables + POS Views/Functions (with Type Fixes) + Indexes

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- CORE MASTER DATA
-- =====================================================

CREATE TABLE IF NOT EXISTS organizations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    legal_name VARCHAR(255),
    tax_id VARCHAR(50),
    currency_code VARCHAR(3) DEFAULT 'USD',
    fiscal_year_variant VARCHAR(10),
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    db_conn_str TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE profit_loss_analytics (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    store_id INTEGER,
    date DATE NOT NULL,
    period_type VARCHAR(20),
    month INTEGER,
    quarter INTEGER,
    year INTEGER,
    gross_revenue DECIMAL(15,2) DEFAULT 0,
    sales_discounts DECIMAL(15,2) DEFAULT 0,
    sales_returns DECIMAL(15,2) DEFAULT 0,
    net_revenue DECIMAL(15,2) DEFAULT 0,
    opening_inventory_value DECIMAL(15,2) DEFAULT 0,
    purchases DECIMAL(15,2) DEFAULT 0,
    closing_inventory_value DECIMAL(15,2) DEFAULT 0,
    cogs DECIMAL(15,2) DEFAULT 0,
    gross_profit DECIMAL(15,2) DEFAULT 0,
    gross_profit_margin DECIMAL(5,2),
    total_expenses DECIMAL(15,2) DEFAULT 0,
    net_profit DECIMAL(15,2) DEFAULT 0,
    net_profit_margin DECIMAL(5,2),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE discount_analytics (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    store_id INTEGER,
    cashier_id INTEGER,
    product_id INTEGER,
    discount_type VARCHAR(50),
    date DATE NOT NULL,
    month INTEGER,
    quarter INTEGER,
    year INTEGER,
    total_discounts_given DECIMAL(15,2) DEFAULT 0,
    transactions_with_discount INTEGER DEFAULT 0,
    total_transactions INTEGER DEFAULT 0,
    discount_percentage DECIMAL(5,2),
    revenue_impact DECIMAL(15,2) DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);





CREATE TABLE modules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    icon VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    display_order INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE menus (
    id SERIAL PRIMARY KEY,
    module_id INTEGER NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    parent_menu_id INTEGER REFERENCES menus(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    route_path VARCHAR(255),
    icon VARCHAR(100),
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(module_id, code)
);

CREATE TABLE submenus (
    id SERIAL PRIMARY KEY,
    menu_id INTEGER NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    parent_submenu_id INTEGER REFERENCES submenus(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    route_path VARCHAR(255),
    icon VARCHAR(100),
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(menu_id, code)
);

-- =====================================================
-- PERMISSION & ACCESS CONTROL
-- =====================================================

CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE module_permissions (
    id SERIAL PRIMARY KEY,
    module_id INTEGER NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    metadata JSONB DEFAULT '{}',
    UNIQUE(module_id, permission_id)
);

CREATE TABLE menu_permissions (
    id SERIAL PRIMARY KEY,
    menu_id INTEGER NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    metadata JSONB DEFAULT '{}',
    UNIQUE(menu_id, permission_id)
);

CREATE TABLE submenu_permissions (
    id SERIAL PRIMARY KEY,
    submenu_id INTEGER NOT NULL REFERENCES submenus(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    metadata JSONB DEFAULT '{}',
    UNIQUE(submenu_id, permission_id)
);

CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    is_system_role BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE role_permissions (
    id SERIAL PRIMARY KEY,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    scope VARCHAR(50) DEFAULT 'all',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, permission_id)
);

CREATE TABLE ui_settings (
    id SERIAL PRIMARY KEY,
    submenu_id INTEGER REFERENCES submenus(id) ON DELETE CASCADE,
    setting_key VARCHAR(100) NOT NULL,
    setting_value JSONB NOT NULL,
    description TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(submenu_id, setting_key)
);

CREATE TABLE role_ui_customizations (
    id SERIAL PRIMARY KEY,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    submenu_id INTEGER NOT NULL REFERENCES submenus(id) ON DELETE CASCADE,
    customization_data JSONB,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, submenu_id)
);

-- =====================================================
-- STORES & LOCATIONS
-- =====================================================

CREATE TABLE stores (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    parent_store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    store_type VARCHAR(50),
    is_warehouse BOOLEAN DEFAULT false,
    is_pos_enabled BOOLEAN DEFAULT false,
    timezone VARCHAR(50) DEFAULT 'UTC',
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, code)
);

CREATE TABLE storage_locations (
    id SERIAL PRIMARY KEY,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    location_type VARCHAR(50),
    parent_location_id INTEGER REFERENCES storage_locations(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, code)
);

-- =====================================================
-- USER MANAGEMENT
-- =====================================================

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    employee_code VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_roles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    metadata JSONB DEFAULT '{}',
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id)
);

CREATE TABLE user_store_access (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT false,
    metadata JSONB DEFAULT '{}',
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, store_id)
);

CREATE TABLE cashiers (
    id            SERIAL PRIMARY KEY,
    user_id       INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    store_id      INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    cashier_code  VARCHAR(50) NOT NULL,
    drawer_limit  DECIMAL(15,2),
    discount_limit DECIMAL(5,2) CHECK (discount_limit BETWEEN 0 AND 100),
    is_active     BOOLEAN   DEFAULT true,
    metadata      JSONB     DEFAULT '{}',
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, cashier_code)
);
CREATE TABLE pos_terminals (
    id SERIAL PRIMARY KEY,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    terminal_code VARCHAR(50) NOT NULL,
    terminal_name VARCHAR(100),
    device_id VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, terminal_code)
);

CREATE TABLE cashier_sessions (
    id SERIAL PRIMARY KEY,
    cashier_id INTEGER NOT NULL REFERENCES cashiers(id) ON DELETE CASCADE,
    pos_terminal_id INTEGER NOT NULL REFERENCES pos_terminals(id) ON DELETE CASCADE,
    session_number VARCHAR(50) NOT NULL,
    opening_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closing_time TIMESTAMP,
    opening_balance DECIMAL(15,2) DEFAULT 0,
    closing_balance DECIMAL(15,2),
    expected_balance DECIMAL(15,2),
    variance DECIMAL(15,2),
    status VARCHAR(20) DEFAULT 'open',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP  -- ✅ new column for triggers
);

-- =====================================================
-- PRODUCT MASTER DATA
-- =====================================================

CREATE TABLE product_categories (
    id SERIAL PRIMARY KEY,
    parent_category_id INTEGER REFERENCES product_categories(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    category_level INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE brands (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE units_of_measure (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(50) NOT NULL,
    uom_type VARCHAR(20),
    decimal_places INTEGER DEFAULT 2,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}'
);

CREATE TABLE uom_packaging_templates (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    uom_id INTEGER NOT NULL REFERENCES units_of_measure(id),
    name VARCHAR(255) NOT NULL, -- e.g., 'Beverage Standard Pattern'
    code VARCHAR(50) NOT NULL, -- e.g., '1-24-12'
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (uom_id, name)
);

-- Pattern Levels Table
CREATE TABLE uom_packaging_template_levels (
    id SERIAL PRIMARY KEY,
    template_id INTEGER NOT NULL REFERENCES uom_packaging_templates(id) ON DELETE CASCADE,
    level_order INTEGER NOT NULL, -- 1=Base, 2=Middle, 3=Top
    uom_id INTEGER NOT NULL REFERENCES units_of_measure(id),
    multiplier DECIMAL(15,6) NOT NULL DEFAULT 1, -- Multiplier relative to the level below it
    UNIQUE(template_id, level_order)
);

CREATE TABLE price_lists (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    price_list_type VARCHAR(50),
    currency_code VARCHAR(3) DEFAULT 'USD',
    valid_from DATE,
    valid_to DATE,
    is_default BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tax_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    tax_rate DECIMAL(5,2) NOT NULL,
    is_inclusive BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    sku VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id INTEGER REFERENCES product_categories(id) ON DELETE SET NULL,
    brand_id INTEGER REFERENCES brands(id) ON DELETE SET NULL,
    base_uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    product_type VARCHAR(50),
    tax_category_id INTEGER REFERENCES tax_categories(id) ON DELETE SET NULL,
    is_serialized BOOLEAN DEFAULT false,
    is_batch_managed BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    is_sellable BOOLEAN DEFAULT true,
    is_purchasable BOOLEAN DEFAULT true,
    allow_decimal_quantity BOOLEAN DEFAULT false,
    track_inventory BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, sku)
);

CREATE TABLE product_variants (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    variant_sku VARCHAR(100) UNIQUE NOT NULL,
    variant_name VARCHAR(255),
    variant_attributes JSONB NOT NULL,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_barcodes (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
    barcode VARCHAR(100) UNIQUE NOT NULL,
    barcode_type VARCHAR(50),
    is_primary BOOLEAN DEFAULT false,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_prices (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
    price_list_id INTEGER NOT NULL REFERENCES price_lists(id) ON DELETE CASCADE,
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    price DECIMAL(15,2) NOT NULL,
    min_quantity DECIMAL(15,3) DEFAULT 1,
    max_quantity DECIMAL(15,3),
    valid_from DATE,
    valid_to DATE,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_uom_conversions (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    from_uom_id INTEGER NOT NULL REFERENCES units_of_measure(id) ON DELETE CASCADE,
    to_uom_id INTEGER NOT NULL REFERENCES units_of_measure(id) ON DELETE CASCADE,
    conversion_factor DECIMAL(15,6) NOT NULL,
    is_default BOOLEAN DEFAULT false,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, from_uom_id, to_uom_id)
);

CREATE TABLE product_serial_numbers (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
    serial_number VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'in_stock',
    current_store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    manufacturing_date DATE,
    expiry_date DATE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- FIX #1 (P0): New stock_reservations table to link pending orders to allocated inventory
CREATE TABLE stock_reservations (
    id                   SERIAL PRIMARY KEY,
    reservation_number   VARCHAR(50) UNIQUE NOT NULL,
    product_id           INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id   INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
    store_id             INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    reference_type       VARCHAR(50) NOT NULL CHECK (reference_type IN ('sales_order','pos_transaction','cart','transfer','manual')),
    reference_id         VARCHAR(100) NOT NULL,
    quantity_reserved    DECIMAL(15,3) NOT NULL CHECK (quantity_reserved > 0),
    reserved_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at           TIMESTAMP,
    status               VARCHAR(30) DEFAULT 'active' CHECK (status IN ('active','fulfilled','cancelled','expired')),
    reserved_by          INTEGER REFERENCES users(id) ON DELETE SET NULL,
    notes                TEXT,
    metadata             JSONB     DEFAULT '{}',
    created_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_batches (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
    batch_number VARCHAR(100) NOT NULL,
    manufacturing_date DATE,
    expiry_date DATE,
    store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    quantity_available DECIMAL(15,3) DEFAULT 0,
    status VARCHAR(50) DEFAULT 'active',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, batch_number, store_id)
);

-- =====================================================
-- INVENTORY MANAGEMENT
-- =====================================================

CREATE TABLE inventory_stock (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    storage_location_id INTEGER REFERENCES storage_locations(id) ON DELETE SET NULL,
    quantity_on_hand DECIMAL(15,3) DEFAULT 0,
    quantity_allocated DECIMAL(15,3) DEFAULT 0,
    quantity_available DECIMAL(15,3) DEFAULT 0,
    quantity_on_order DECIMAL(15,3) DEFAULT 0,
    quantity_in_transit DECIMAL(15,3) DEFAULT 0,
    reorder_level DECIMAL(15,3),
    reorder_quantity DECIMAL(15,3),
    max_stock_level DECIMAL(15,3),
    last_counted_at TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE stock_movements (
    id SERIAL PRIMARY KEY,
    movement_type VARCHAR(50) NOT NULL,
    reference_type VARCHAR(50),
    reference_id INTEGER,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
    from_store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    to_store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    from_location_id INTEGER REFERENCES storage_locations(id) ON DELETE SET NULL,
    to_location_id INTEGER REFERENCES storage_locations(id) ON DELETE SET NULL,
    quantity DECIMAL(15,3) NOT NULL,
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    batch_number VARCHAR(100),
    serial_number VARCHAR(100),
    movement_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    posted_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(50) DEFAULT 'completed',
    cost_per_unit DECIMAL(15,4),
    total_value DECIMAL(15,2),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE stock_counts (
    id SERIAL PRIMARY KEY,
    count_number VARCHAR(50) UNIQUE NOT NULL,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    storage_location_id INTEGER REFERENCES storage_locations(id) ON DELETE SET NULL,
    count_type VARCHAR(50),
    status VARCHAR(50) DEFAULT 'planned',
    scheduled_date DATE,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    counted_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    approved_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE stock_count_lines (
    id SERIAL PRIMARY KEY,
    stock_count_id INTEGER NOT NULL REFERENCES stock_counts(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
    storage_location_id INTEGER REFERENCES storage_locations(id) ON DELETE SET NULL,
    expected_quantity DECIMAL(15,3) DEFAULT 0,
    system_quantity DECIMAL(15,3) DEFAULT 0,
    counted_quantity DECIMAL(15,3) DEFAULT 0,
    variance DECIMAL(15,3) DEFAULT 0,
    variance_value DECIMAL(15,2) DEFAULT 0,
    counted_at TIMESTAMP,
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    batch_number VARCHAR(100),
    serial_number VARCHAR(100),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- SUPPLIERS & CUSTOMERS
-- =====================================================

CREATE TABLE suppliers (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    supplier_type VARCHAR(50),
    credit_limit DECIMAL(15,2) DEFAULT 0,
    contact_person VARCHAR(100),
    email VARCHAR(255),
    phone VARCHAR(50),
    address TEXT,
    currency_code VARCHAR(3) DEFAULT 'USD',
    payment_terms VARCHAR(100),
    tax_id VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, code)
);

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
    supplier_id INTEGER NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
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
    supplier_id INTEGER NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
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
-- ENHANCED: DRAFT CARTS (Saved for Later)
-- =====================================================

-- Draft cart templates for quick reordering
CREATE TABLE draft_cart_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    
    template_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Classification
    template_type VARCHAR(50) DEFAULT 'saved_cart', -- saved_cart, wishlist, reorder_list
    is_favorite BOOLEAN DEFAULT false,
    
    -- Auto-reorder settings
    auto_reorder_enabled BOOLEAN DEFAULT false,
    reorder_frequency_days INTEGER,
    next_reorder_date DATE,
    
    -- Template metadata
    total_items INTEGER DEFAULT 0,
    estimated_total DECIMAL(15,2) DEFAULT 0.00,
    
    metadata JSONB DEFAULT '{}',
    notes TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(organization_id, customer_id, template_name)
);

-- Items in draft cart templates
CREATE TABLE draft_cart_template_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_id UUID NOT NULL REFERENCES draft_cart_templates(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
    
    quantity DECIMAL(15,3) NOT NULL CHECK (quantity > 0),
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    
    -- Price reference (for comparison)
    last_known_price DECIMAL(15,2),
    
    -- Priority for auto-reorder
    priority INTEGER DEFAULT 0,
    
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Enhanced sales orders table
CREATE TABLE sales_orders_v2 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_number VARCHAR(50) UNIQUE NOT NULL,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    
    -- Customer information
    customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
    customer_name VARCHAR(255),
    customer_email VARCHAR(255),
    customer_phone VARCHAR(50),
    
    -- Order classification
    order_type order_type DEFAULT 'standard' NOT NULL,
    order_status order_status_v2 DEFAULT 'draft' NOT NULL,
    payment_status payment_status DEFAULT 'unpaid' NOT NULL,
    fulfillment_status fulfillment_status DEFAULT 'unfulfilled' NOT NULL,
    
    -- Channel and source
    sales_channel VARCHAR(50) DEFAULT 'online', -- online, pos, phone, mobile_app
    order_source VARCHAR(100), -- website, mobile_app, marketplace, etc.
    referral_source VARCHAR(255),
    
    -- Created from cart
    source_cart_id UUID REFERENCES carts(id) ON DELETE SET NULL,
    
    -- User tracking
    created_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    assigned_to_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    
    -- Dates
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    confirmed_date TIMESTAMP,
    expected_delivery_date DATE,
    actual_delivery_date DATE,
    cancelled_date TIMESTAMP,
    
    -- Financial totals
    subtotal DECIMAL(15,2) DEFAULT 0.00,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    shipping_amount DECIMAL(15,2) DEFAULT 0.00,
    adjustment_amount DECIMAL(15,2) DEFAULT 0.00,
    total_amount DECIMAL(15,2) DEFAULT 0.00,
    
    -- Payment tracking
    paid_amount DECIMAL(15,2) DEFAULT 0.00,
    refunded_amount DECIMAL(15,2) DEFAULT 0.00,
    balance_due DECIMAL(15,2) DEFAULT 0.00,
    
    -- Discounts and promotions
    coupon_code VARCHAR(100),
    discount_codes TEXT[], -- Array for multiple discount codes
    promotional_credits DECIMAL(15,2) DEFAULT 0.00,
    
    -- Addresses
    shipping_address JSONB NOT NULL,
    billing_address JSONB NOT NULL,
    
    -- Shipping details
    shipping_method VARCHAR(100),
    shipping_carrier VARCHAR(100),
    tracking_number VARCHAR(255),
    tracking_url TEXT,
    
    -- Payment method
    payment_method VARCHAR(100),
    payment_gateway VARCHAR(100),
    payment_terms VARCHAR(100),
    payment_due_date DATE,
    
    -- POS specific
    pos_terminal_id INTEGER REFERENCES pos_terminals(id) ON DELETE SET NULL,
    cashier_id INTEGER REFERENCES cashiers(id) ON DELETE SET NULL,
    
    -- Special handling
    is_gift BOOLEAN DEFAULT false,
    gift_message TEXT,
    special_instructions TEXT,
    internal_notes TEXT,
    
    -- Tags and categorization
    tags TEXT[],
    priority VARCHAR(20) DEFAULT 'normal', -- low, normal, high, urgent
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Enhanced order line items
CREATE TABLE sales_order_lines_v2 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sales_order_id UUID NOT NULL REFERENCES sales_orders_v2(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    line_number INTEGER NOT NULL,
    
    -- Product information
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
    product_name VARCHAR(255) NOT NULL, -- Snapshot for historical accuracy
    product_sku VARCHAR(100),
    
    -- Quantity and UOM
    quantity_ordered DECIMAL(15,3) NOT NULL CHECK (quantity_ordered > 0),
    quantity_fulfilled DECIMAL(15,3) DEFAULT 0.00,
    quantity_cancelled DECIMAL(15,3) DEFAULT 0.00,
    quantity_returned DECIMAL(15,3) DEFAULT 0.00,
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    
    -- Pricing
    unit_price DECIMAL(15,2) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    discount_percentage DECIMAL(5,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    line_total DECIMAL(15,2) NOT NULL,
    
    -- Tax details
    tax_category_id INTEGER REFERENCES tax_categories(id) ON DELETE SET NULL,
    tax_rate DECIMAL(5,2),
    
    -- Inventory tracking
    batch_number VARCHAR(100),
    serial_numbers TEXT[], -- Array of serial numbers for this line
    expiry_date DATE,
    
    -- Fulfillment status
    line_status VARCHAR(50) DEFAULT 'pending', -- pending, fulfilled, cancelled, returned
    
    -- Product configuration
    customization_details JSONB DEFAULT '{}',
    
    -- Cost tracking (for profitability)
    unit_cost DECIMAL(15,2),
    
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(sales_order_id, line_number)
);

-- Order status history
CREATE TABLE order_status_history (
    id BIGSERIAL PRIMARY KEY,
    sales_order_id UUID NOT NULL REFERENCES sales_orders_v2(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    from_status order_status_v2,
    to_status order_status_v2 NOT NULL,
    
    reason VARCHAR(255),
    notes TEXT,
    
    changed_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Order fulfillment tracking
CREATE TABLE order_fulfillments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sales_order_id UUID NOT NULL REFERENCES sales_orders_v2(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    fulfillment_number VARCHAR(50) UNIQUE NOT NULL,
    
    -- Fulfillment details
    fulfillment_status VARCHAR(50) DEFAULT 'pending',
    shipment_status VARCHAR(50) DEFAULT 'pending', -- pending, picked, packed, shipped, delivered
    
    -- Warehouse/Store
    fulfillment_store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    
    -- Shipping
    shipping_carrier VARCHAR(100),
    shipping_method VARCHAR(100),
    tracking_number VARCHAR(255),
    tracking_url TEXT,
    
    -- Dates
    picked_at TIMESTAMP,
    packed_at TIMESTAMP,
    shipped_at TIMESTAMP,
    estimated_delivery_date DATE,
    actual_delivery_date DATE,
    
    -- Personnel
    picked_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    packed_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Items in each fulfillment
CREATE TABLE order_fulfillment_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fulfillment_id UUID NOT NULL REFERENCES order_fulfillments(id) ON DELETE CASCADE,
    order_line_id UUID NOT NULL REFERENCES sales_order_lines_v2(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    quantity_fulfilled DECIMAL(15,3) NOT NULL CHECK (quantity_fulfilled > 0),
    
    batch_number VARCHAR(100),
    serial_numbers TEXT[],
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- =====================================================
-- POS TRANSACTIONS
-- =====================================================

CREATE TABLE pos_transactions (
    id SERIAL PRIMARY KEY,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    cashier_id INTEGER NOT NULL REFERENCES cashiers(id) ON DELETE CASCADE,
    cashier_session_id INTEGER NOT NULL REFERENCES cashier_sessions(id) ON DELETE CASCADE,
    customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
    pos_terminal_id INTEGER REFERENCES pos_terminals(id) ON DELETE SET NULL,
    transaction_number VARCHAR(50) UNIQUE NOT NULL,
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    transaction_type VARCHAR(50),
    subtotal DECIMAL(15,2) DEFAULT 0,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) DEFAULT 0,
    total_cost DECIMAL(15,2) DEFAULT 0,
    amount_paid DECIMAL(15,2) DEFAULT 0,
    change_given DECIMAL(15,2) DEFAULT 0,
    status VARCHAR(50) DEFAULT 'completed',
    price_list_id INTEGER REFERENCES price_lists(id) ON DELETE SET NULL,
    sales_order_id UUID REFERENCES sales_orders_v2(id) ON DELETE SET NULL,
    source_cart_id UUID REFERENCES carts(id) ON DELETE SET NULL,
    voided_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    voided_at TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE pos_transaction_lines (
    id SERIAL PRIMARY KEY,
    transaction_id INTEGER NOT NULL REFERENCES pos_transactions(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
    quantity DECIMAL(15,3) NOT NULL,
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    unit_price DECIMAL(15,4) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    subtotal DECIMAL(15,2) NOT NULL,
    line_total DECIMAL(15,2) DEFAULT 0,
    cost_price DECIMAL(15,2) DEFAULT 0,
    line_number INTEGER,
    serial_number VARCHAR(100),
    batch_number VARCHAR(100),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE pos_payments (
    id SERIAL PRIMARY KEY,
    transaction_id INTEGER NOT NULL REFERENCES pos_transactions(id) ON DELETE CASCADE,
    payment_method VARCHAR(50) NOT NULL,
    payment_gateway VARCHAR(50),
    amount DECIMAL(15,2) NOT NULL,
    payment_reference VARCHAR(100),
    reference_number VARCHAR(100),
    payment_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- RESTAURANT MODULE TABLES
-- =====================================================

CREATE TABLE restaurant_tables (
    id                  SERIAL PRIMARY KEY,
    store_id            INTEGER     NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    table_number        VARCHAR(20) NOT NULL,
    table_name          VARCHAR(100),
    section             VARCHAR(50),
    capacity            INTEGER     DEFAULT 4,
    is_active           BOOLEAN     DEFAULT true,
    metadata            JSONB       DEFAULT '{}',
    created_at          TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, table_number)
);

CREATE TABLE menu_categories (
    id                  SERIAL PRIMARY KEY,
    store_id            INTEGER     NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    parent_category_id  INTEGER     REFERENCES menu_categories(id) ON DELETE SET NULL,
    name                VARCHAR(255) NOT NULL,
    code                VARCHAR(50)  NOT NULL,
    description         TEXT,
    category_level      INTEGER      DEFAULT 1,
    display_order       INTEGER      DEFAULT 0,
    icon                VARCHAR(100),
    image_url           TEXT,
    is_active           BOOLEAN      DEFAULT true,
    metadata            JSONB        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, code)
);

CREATE TABLE menu_items (
    id                  SERIAL PRIMARY KEY,
    store_id            INTEGER      NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    menu_category_id    INTEGER      NOT NULL REFERENCES menu_categories(id) ON DELETE CASCADE,
    product_id          INTEGER      REFERENCES products(id) ON DELETE SET NULL,
    recipe_id           INTEGER,
    name                VARCHAR(255) NOT NULL,
    short_name          VARCHAR(50),
    description         TEXT,
    image_url           TEXT,
    base_price          DECIMAL(15,2) NOT NULL,
    cost_price          DECIMAL(15,2) DEFAULT 0,
    preparation_time_min INTEGER     DEFAULT 0,
    tax_category_id     INTEGER      REFERENCES tax_categories(id) ON DELETE SET NULL,
    is_available        BOOLEAN      DEFAULT true,
    is_active           BOOLEAN      DEFAULT true,
    display_order       INTEGER      DEFAULT 0,
    metadata            JSONB        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE menu_item_modifiers (
    id                  SERIAL PRIMARY KEY,
    menu_item_id        INTEGER     NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    modifier_name       VARCHAR(100) NOT NULL,
    modifier_type       VARCHAR(30)  NOT NULL DEFAULT 'addon',
    price_adjustment    DECIMAL(15,2) DEFAULT 0,
    is_active           BOOLEAN     DEFAULT true,
    display_order       INTEGER     DEFAULT 0,
    metadata            JSONB       DEFAULT '{}',
    created_at          TIMESTAMP   DEFAULT CURRENT_TIMESTAMP
);
-- FIX #10 (P1): New menu_modifier_groups table to enforce min/max modifier selections
CREATE TABLE menu_modifier_groups (
    id                 SERIAL PRIMARY KEY,
    store_id           INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    name               VARCHAR(100) NOT NULL,
    code               VARCHAR(50)  NOT NULL,
    selection_type     VARCHAR(20) DEFAULT 'optional' CHECK (selection_type IN ('required','optional','multiple')),
    min_selections     INTEGER DEFAULT 0,
    max_selections     INTEGER,
    is_active          BOOLEAN   DEFAULT true,
    display_order      INTEGER   DEFAULT 0,
    metadata           JSONB     DEFAULT '{}',
    created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, code),
    CONSTRAINT chk_modifier_group_selections CHECK (
        min_selections >= 0
        AND (max_selections IS NULL OR max_selections >= min_selections)
    )
);

-- FIX #14 (P1): Central promotion/coupon definition table
CREATE TABLE promotions (
    id                    SERIAL PRIMARY KEY,
    organization_id       INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    code                  VARCHAR(50) NOT NULL,
    name                  VARCHAR(255) NOT NULL,
    description           TEXT,
    promotion_type        VARCHAR(50) NOT NULL CHECK (promotion_type IN ('percentage_discount','fixed_discount','bogo','buy_x_get_y','free_item','bundle_price','points_multiplier','happy_hour')),
    -- FIX #16: Action metadata for complex rules
    action_metadata       JSONB DEFAULT '{}',
    -- FIX #15: Scheduling
    valid_from            TIMESTAMP,
    valid_to              TIMESTAMP,
    schedule_json         JSONB DEFAULT '{}', -- e.g. {"days":[1,2,3],"start_time":"12:00","end_time":"14:00"}
    -- Applicability
    applies_to            VARCHAR(50) DEFAULT 'all' CHECK (applies_to IN ('all','category','product','customer_type','price_list')),
    target_product_ids    INTEGER[] DEFAULT '{}',
    target_category_ids   INTEGER[] DEFAULT '{}',
    -- FIX #17: customer segmentation
    target_customer_types TEXT[]    DEFAULT '{}',
    min_order_amount      DECIMAL(15,2),
    min_quantity          DECIMAL(15,3),
    coupon_code           VARCHAR(50),
    usage_limit           INTEGER,
    usage_count           INTEGER DEFAULT 0,
    usage_per_customer    INTEGER,
    discount_value        DECIMAL(15,4),
    is_stackable          BOOLEAN DEFAULT false,
    is_active             BOOLEAN DEFAULT true,
    store_ids             INTEGER[] DEFAULT '{}',
    created_by            INTEGER REFERENCES users(id) ON DELETE SET NULL,
    metadata              JSONB     DEFAULT '{}',
    created_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, code)
);




CREATE TABLE recipes (
    id                    SERIAL PRIMARY KEY,
    organization_id       INTEGER     NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    recipe_code           VARCHAR(50) NOT NULL,
    recipe_name           VARCHAR(255) NOT NULL,
    description           TEXT,
    finished_product_id   INTEGER     REFERENCES products(id) ON DELETE SET NULL,
    yield_quantity        DECIMAL(15,3) DEFAULT 1,
    yield_uom_id          INTEGER     REFERENCES units_of_measure(id) ON DELETE SET NULL,
    preparation_steps     TEXT,
    preparation_time_min  INTEGER     DEFAULT 0,
    cooking_time_min      INTEGER     DEFAULT 0,
    is_active             BOOLEAN     DEFAULT true,
    metadata              JSONB       DEFAULT '{}',
    created_at            TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, recipe_code)
);

CREATE TABLE recipe_ingredients (
    id                  SERIAL PRIMARY KEY,
    recipe_id           INTEGER      NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    product_id          INTEGER      NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id  INTEGER      REFERENCES product_variants(id) ON DELETE SET NULL,
    quantity            DECIMAL(15,3) NOT NULL,
    uom_id              INTEGER      REFERENCES units_of_measure(id) ON DELETE SET NULL,
    is_optional         BOOLEAN      DEFAULT false,
    is_byproduct        BOOLEAN      DEFAULT false,
    line_number         INTEGER,
    metadata            JSONB        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(recipe_id, product_id, product_variant_id)
);
-- FIX #13 (P1): Combo / meal deal / bundle support
CREATE TABLE combo_bundles (
    id              SERIAL PRIMARY KEY,
    store_id        INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    code            VARCHAR(50) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    bundle_price    DECIMAL(15,2) NOT NULL,
    bundle_type     VARCHAR(30) DEFAULT 'fixed' CHECK (bundle_type IN ('fixed','build_your_own','meal_deal','bogo')),
    is_active       BOOLEAN   DEFAULT true,
    valid_from      DATE,
    valid_to        DATE,
    display_order   INTEGER   DEFAULT 0,
    metadata        JSONB     DEFAULT '{}',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, code)
);

ALTER TABLE menu_items
    ADD CONSTRAINT fk_menu_items_recipe
    FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE SET NULL;

-- FIX #11 (P1): New time-based / day-of-week menu availability schedules
CREATE TABLE menu_item_availability_schedules (
    id              SERIAL PRIMARY KEY,
    menu_item_id    INTEGER NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    day_of_week     INTEGER CHECK (day_of_week BETWEEN 0 AND 6), -- 0=Sunday, 6=Saturday; NULL = every day
    start_time      TIME    NOT NULL,
    end_time        TIME    NOT NULL,
    is_active       BOOLEAN   DEFAULT true,
    valid_from      DATE,
    valid_to        DATE,
    metadata        JSONB     DEFAULT '{}',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_schedule_times CHECK (end_time > start_time)
);    

CREATE TABLE restaurant_orders (
    id                    SERIAL PRIMARY KEY,
    store_id              INTEGER      NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    table_id              INTEGER      REFERENCES restaurant_tables(id) ON DELETE SET NULL,
    cashier_id            INTEGER      REFERENCES cashiers(id) ON DELETE SET NULL,
    cashier_session_id    INTEGER      REFERENCES cashier_sessions(id) ON DELETE SET NULL,
    customer_id           INTEGER      REFERENCES customers(id) ON DELETE SET NULL,
    order_number          VARCHAR(50)  NOT NULL,
    order_source          VARCHAR(30)  NOT NULL DEFAULT 'counter',
    status                VARCHAR(30)  NOT NULL DEFAULT 'pending',
    subtotal              DECIMAL(15,2) DEFAULT 0,
    discount_amount       DECIMAL(15,2) DEFAULT 0,
    tax_amount            DECIMAL(15,2) DEFAULT 0,
    total_amount          DECIMAL(15,2) DEFAULT 0,
    amount_paid           DECIMAL(15,2) DEFAULT 0,
    change_given          DECIMAL(15,2) DEFAULT 0,
    notes                 TEXT,
    pos_transaction_id    INTEGER      REFERENCES pos_transactions(id) ON DELETE SET NULL,
    ordered_at            TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    confirmed_at          TIMESTAMP,
    served_at             TIMESTAMP,
    paid_at               TIMESTAMP,
    metadata              JSONB        DEFAULT '{}',
    created_at            TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, order_number)
);

CREATE TABLE restaurant_order_items (
    id                  SERIAL PRIMARY KEY,
    order_id            INTEGER      NOT NULL REFERENCES restaurant_orders(id) ON DELETE CASCADE,
    menu_item_id        INTEGER      NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    quantity            DECIMAL(15,3) NOT NULL DEFAULT 1,
    unit_price          DECIMAL(15,4) NOT NULL,
    modifiers_snapshot  JSONB        DEFAULT '[]',
    modifiers_total     DECIMAL(15,2) DEFAULT 0,
    discount_amount     DECIMAL(15,2) DEFAULT 0,
    tax_amount          DECIMAL(15,2) DEFAULT 0,
    subtotal            DECIMAL(15,2) NOT NULL,
    line_number         INTEGER,
    notes               TEXT,
    status              VARCHAR(30)  DEFAULT 'pending',
    metadata            JSONB        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE waste_logs (
    id                  SERIAL PRIMARY KEY,
    store_id            INTEGER      NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    product_id          INTEGER      REFERENCES products(id) ON DELETE SET NULL,
    menu_item_id        INTEGER      REFERENCES menu_items(id) ON DELETE SET NULL,
    recipe_id           INTEGER      REFERENCES recipes(id) ON DELETE SET NULL,
    waste_source        VARCHAR(30)  NOT NULL DEFAULT 'kitchen',
    quantity            DECIMAL(15,3) NOT NULL,
    uom_id              INTEGER      REFERENCES units_of_measure(id) ON DELETE SET NULL,
    unit_cost           DECIMAL(15,4) DEFAULT 0,
    total_cost          DECIMAL(15,2) DEFAULT 0,
    reason              TEXT,
    logged_by           INTEGER      REFERENCES users(id) ON DELETE SET NULL,
    order_id            INTEGER      REFERENCES restaurant_orders(id) ON DELETE SET NULL,
    wasted_at           TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    metadata            JSONB        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE kiosk_sessions (
    id                  SERIAL PRIMARY KEY,
    pos_terminal_id     INTEGER      NOT NULL REFERENCES pos_terminals(id) ON DELETE CASCADE,
    store_id            INTEGER      NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    session_token       VARCHAR(255) NOT NULL UNIQUE,
    status              VARCHAR(20)  DEFAULT 'active',
    opened_at           TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    closed_at           TIMESTAMP,
    metadata            JSONB        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- ANALYTICS TABLES
-- =====================================================

CREATE TABLE sales_analytics (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    store_id INTEGER,
    product_id INTEGER,
    category_id INTEGER,
    customer_id INTEGER,
    date DATE NOT NULL,
    hour INTEGER,
    day_of_week INTEGER,
    month INTEGER,
    quarter INTEGER,
    year INTEGER,
    units_sold DECIMAL(15,3) DEFAULT 0,
    revenue DECIMAL(15,2) DEFAULT 0,
    discounts DECIMAL(15,2) DEFAULT 0,
    taxes DECIMAL(15,2) DEFAULT 0,
    net_revenue DECIMAL(15,2) DEFAULT 0,
    transactions INTEGER DEFAULT 0,
    payment_method VARCHAR(50),
    payment_gateway VARCHAR(50),
    average_order_value DECIMAL(15,2),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE purchase_analytics (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    store_id INTEGER,
    supplier_id INTEGER,
    product_id INTEGER,
    category_id INTEGER,
    date DATE NOT NULL,
    month INTEGER,
    quarter INTEGER,
    year INTEGER,
    units_purchased DECIMAL(15,3) DEFAULT 0,
    total_cost DECIMAL(15,2) DEFAULT 0,
    discounts DECIMAL(15,2) DEFAULT 0,
    taxes DECIMAL(15,2) DEFAULT 0,
    net_cost DECIMAL(15,2) DEFAULT 0,
    orders INTEGER DEFAULT 0,
    total_orders INTEGER DEFAULT 0,
    total_quantity DECIMAL(15,3) DEFAULT 0,
    total_amount DECIMAL(15,2) DEFAULT 0,
    discounts_received DECIMAL(15,2) DEFAULT 0,
    taxes_paid DECIMAL(15,2) DEFAULT 0,
    net_amount DECIMAL(15,2) DEFAULT 0,
    average_order_value DECIMAL(15,2),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE inventory_analytics (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    store_id INTEGER,
    product_id INTEGER,
    category_id INTEGER,
    date DATE NOT NULL,
    month INTEGER,
    quarter INTEGER,
    year INTEGER,
    opening_stock DECIMAL(15,3) DEFAULT 0,
    stock_in DECIMAL(15,3) DEFAULT 0,
    stock_out DECIMAL(15,3) DEFAULT 0,
    receipts DECIMAL(15,3) DEFAULT 0,
    issues DECIMAL(15,3) DEFAULT 0,
    adjustments DECIMAL(15,3) DEFAULT 0,
    closing_stock DECIMAL(15,3) DEFAULT 0,
    average_stock DECIMAL(15,3) DEFAULT 0,
    stock_value DECIMAL(15,2) DEFAULT 0,
    turnover_rate DECIMAL(5,2),
    stock_turnover_ratio DECIMAL(5,2),
    days_of_inventory DECIMAL(15,3) DEFAULT 0,
    days_in_stock DECIMAL(5,2),
    low_stock_alerts INTEGER DEFAULT 0,
    out_of_stock_days INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- FIX #18 (P1): Loyalty points redemption rules
CREATE TABLE loyalty_redemption_rules (
    id                      SERIAL PRIMARY KEY,
    organization_id         INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    rule_name               VARCHAR(255) NOT NULL,
    points_earning_rate     DECIMAL(10,4) DEFAULT 1,  -- points per currency unit spent
    points_redemption_rate  DECIMAL(10,4) DEFAULT 1,  -- currency value per point
    min_points_to_redeem    DECIMAL(15,2) DEFAULT 0,
    max_points_per_txn      DECIMAL(15,2),
    max_redemption_percent  DECIMAL(5,2) CHECK (max_redemption_percent BETWEEN 0 AND 100),
    eligible_product_types  TEXT[]   DEFAULT '{}',
    expiry_days             INTEGER,
    is_active               BOOLEAN  DEFAULT true,
    valid_from              DATE,
    valid_to                DATE,
    metadata                JSONB    DEFAULT '{}',
    created_at              TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- FIX #28 (P2): Audit trail table
CREATE TABLE audit_logs (
    id            BIGSERIAL PRIMARY KEY,
    organization_id INTEGER REFERENCES organizations(id) ON DELETE CASCADE,
    table_name    VARCHAR(100) NOT NULL,
    record_id     VARCHAR(100) NOT NULL,
    action        VARCHAR(20)  NOT NULL CHECK (action IN ('INSERT','UPDATE','DELETE','SELECT')),
    old_values    JSONB,
    new_values    JSONB,
    changed_fields TEXT[],
    performed_by  INTEGER REFERENCES users(id) ON DELETE SET NULL,
    ip_address    INET,
    user_agent    TEXT,
    session_id    VARCHAR(255),
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

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
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Apply updated_at triggers
-- +goose StatementBegin
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
-- +goose StatementEnd

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
-- +goose StatementBegin
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
-- +goose StatementEnd

CREATE TRIGGER cart_status_change_trigger
    AFTER UPDATE ON carts
    FOR EACH ROW
    WHEN (OLD.cart_status IS DISTINCT FROM NEW.cart_status)
    EXECUTE FUNCTION log_cart_status_change();

-- Update cart last_activity_at when items change
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_cart_activity()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE carts 
    SET last_activity_at = CURRENT_TIMESTAMP
    WHERE id = COALESCE(NEW.cart_id, OLD.cart_id);
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER cart_items_activity_trigger
    AFTER INSERT OR UPDATE OR DELETE ON cart_items
    FOR EACH ROW
    EXECUTE FUNCTION update_cart_activity();

-- =====================================================
-- TRIGGERS FOR ORDER FINANCIAL CALCULATIONS
-- =====================================================

-- Calculate order totals
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION calculate_order_totals()
RETURNS TRIGGER AS $$
DECLARE
    v_subtotal DECIMAL(15,2);
    v_tax_amount DECIMAL(15,2);
BEGIN
    -- Calculate subtotal and tax from order lines
    SELECT 
        COALESCE(SUM(line_total - tax_amount), 0),
        COALESCE(SUM(tax_amount), 0)
    INTO v_subtotal, v_tax_amount
    FROM sales_order_lines_v2
    WHERE sales_order_id = COALESCE(NEW.sales_order_id, OLD.sales_order_id);
    
    -- Update the order
    UPDATE sales_orders_v2
    SET 
        subtotal = v_subtotal,
        tax_amount = v_tax_amount,
        total_amount = v_subtotal + v_tax_amount + COALESCE(shipping_amount, 0) + COALESCE(adjustment_amount, 0) - COALESCE(discount_amount, 0),
        balance_due = (v_subtotal + v_tax_amount + COALESCE(shipping_amount, 0) + COALESCE(adjustment_amount, 0) - COALESCE(discount_amount, 0)) - COALESCE(paid_amount, 0)
    WHERE id = COALESCE(NEW.sales_order_id, OLD.sales_order_id);
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER calculate_order_totals_trigger
    AFTER INSERT OR UPDATE OR DELETE ON sales_order_lines_v2
    FOR EACH ROW
    EXECUTE FUNCTION calculate_order_totals();

-- =====================================================
-- TRIGGERS FOR INVOICE CALCULATIONS
-- =====================================================

-- Calculate invoice totals
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION calculate_invoice_totals()
RETURNS TRIGGER AS $$
DECLARE
    v_subtotal DECIMAL(15,2);
    v_tax_amount DECIMAL(15,2);
BEGIN
    -- Calculate from invoice lines
    SELECT 
        COALESCE(SUM(line_total - tax_amount), 0),
        COALESCE(SUM(tax_amount), 0)
    INTO v_subtotal, v_tax_amount
    FROM invoice_lines
    WHERE invoice_id = COALESCE(NEW.invoice_id, OLD.invoice_id);
    
    -- Update the invoice
    UPDATE invoices
    SET 
        subtotal = v_subtotal,
        tax_amount = v_tax_amount,
        total_amount = v_subtotal + v_tax_amount + COALESCE(shipping_amount, 0) + COALESCE(adjustment_amount, 0) - COALESCE(discount_amount, 0),
        balance_due = (v_subtotal + v_tax_amount + COALESCE(shipping_amount, 0) + COALESCE(adjustment_amount, 0) - COALESCE(discount_amount, 0)) - COALESCE(paid_amount, 0) - COALESCE(credit_applied, 0)
    WHERE id = COALESCE(NEW.invoice_id, OLD.invoice_id);
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER calculate_invoice_totals_trigger
    AFTER INSERT OR UPDATE OR DELETE ON invoice_lines
    FOR EACH ROW
    EXECUTE FUNCTION calculate_invoice_totals();

-- Update invoice paid amount when payment received
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_invoice_payment()
RETURNS TRIGGER AS $$
DECLARE
    v_total_paid DECIMAL(15,2);
BEGIN
    -- Calculate total paid
    SELECT COALESCE(SUM(payment_amount), 0)
    INTO v_total_paid
    FROM invoice_payments
    WHERE invoice_id = COALESCE(NEW.invoice_id, OLD.invoice_id);
    
    -- Update invoice
    UPDATE invoices
    SET 
        paid_amount = v_total_paid,
        balance_due = total_amount - v_total_paid - COALESCE(credit_applied, 0),
        invoice_status = CASE
            WHEN v_total_paid = 0 THEN 'sent'::invoice_status
            WHEN v_total_paid >= total_amount - COALESCE(credit_applied, 0) THEN 'paid'::invoice_status
            ELSE 'partially_paid'::invoice_status
        END,
        paid_date = CASE 
            WHEN v_total_paid >= total_amount - COALESCE(credit_applied, 0) THEN CURRENT_DATE
            ELSE NULL
        END
    WHERE id = COALESCE(NEW.invoice_id, OLD.invoice_id);
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER update_invoice_payment_trigger
    AFTER INSERT OR UPDATE OR DELETE ON invoice_payments
    FOR EACH ROW
    EXECUTE FUNCTION update_invoice_payment();

-- =====================================================
-- COMMENTS FOR DOCUMENTATION
-- =====================================================

COMMENT ON TABLE carts IS 'Shopping carts for online and POS channels, supporting both registered customers and guests';
COMMENT ON TABLE cart_items IS 'Line items in shopping carts with pricing and customization details';
COMMENT ON TABLE cart_activity_log IS 'Audit trail of all cart activities and changes';
COMMENT ON TABLE draft_cart_templates IS 'Saved carts and wishlists for quick reordering';
COMMENT ON TABLE sales_orders_v2 IS 'Enhanced order management with comprehensive tracking across all sales channels';
COMMENT ON TABLE sales_order_lines_v2 IS 'Order line items with fulfillment tracking';
COMMENT ON TABLE order_fulfillments IS 'Shipment and fulfillment tracking for orders';
COMMENT ON TABLE invoices IS 'Customer invoices with payment tracking and recurring billing support';
COMMENT ON TABLE invoice_payments IS 'Payment records against invoices';
COMMENT ON TABLE quotes IS 'Sales quotations with approval workflow';

-- =====================================================
-- FOREIGN KEY CONSTRAINTS (DEFERRED)
-- =====================================================

-- Sales Analytics
ALTER TABLE sales_analytics 
    ADD CONSTRAINT fk_sales_analytics_store 
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE SET NULL;

ALTER TABLE sales_analytics 
    ADD CONSTRAINT fk_sales_analytics_product 
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL;

ALTER TABLE sales_analytics 
    ADD CONSTRAINT fk_sales_analytics_category 
    FOREIGN KEY (category_id) REFERENCES product_categories(id) ON DELETE SET NULL;

ALTER TABLE sales_analytics 
    ADD CONSTRAINT fk_sales_analytics_customer 
    FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE SET NULL;

-- Purchase Analytics
ALTER TABLE purchase_analytics 
    ADD CONSTRAINT fk_purchase_analytics_store 
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE SET NULL;

ALTER TABLE purchase_analytics 
    ADD CONSTRAINT fk_purchase_analytics_supplier 
    FOREIGN KEY (supplier_id) REFERENCES suppliers(id) ON DELETE SET NULL;

ALTER TABLE purchase_analytics 
    ADD CONSTRAINT fk_purchase_analytics_product 
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL;

ALTER TABLE purchase_analytics
    ADD CONSTRAINT fk_purchase_analytics_category
    FOREIGN KEY (category_id) REFERENCES product_categories(id) ON DELETE SET NULL;

-- Inventory Analytics
ALTER TABLE inventory_analytics 
    ADD CONSTRAINT fk_inventory_analytics_store 
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE SET NULL;

ALTER TABLE inventory_analytics 
    ADD CONSTRAINT fk_inventory_analytics_product 
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL;

ALTER TABLE inventory_analytics
    ADD CONSTRAINT fk_inventory_analytics_category
    FOREIGN KEY (category_id) REFERENCES product_categories(id) ON DELETE SET NULL;

-- Profit Loss Analytics
ALTER TABLE profit_loss_analytics 
    ADD CONSTRAINT fk_profit_loss_analytics_store 
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE SET NULL;

-- Discount Analytics
ALTER TABLE discount_analytics 
    ADD CONSTRAINT fk_discount_analytics_store 
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE SET NULL;

ALTER TABLE discount_analytics 
    ADD CONSTRAINT fk_discount_analytics_cashier 
    FOREIGN KEY (cashier_id) REFERENCES cashiers(id) ON DELETE SET NULL;

ALTER TABLE discount_analytics 
    ADD CONSTRAINT fk_discount_analytics_product 
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL;

-- =====================================================
-- TRIGGERS FOR UPDATED_AT
-- =====================================================

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $func$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$func$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Apply triggers to all tables with updated_at
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;
CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_modules_updated_at ON modules;
CREATE TRIGGER update_modules_updated_at BEFORE UPDATE ON modules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_menus_updated_at ON menus;
CREATE TRIGGER update_menus_updated_at BEFORE UPDATE ON menus FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_submenus_updated_at ON submenus;
CREATE TRIGGER update_submenus_updated_at BEFORE UPDATE ON submenus FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_ui_settings_updated_at ON ui_settings;
CREATE TRIGGER update_ui_settings_updated_at BEFORE UPDATE ON ui_settings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_role_ui_customizations_updated_at ON role_ui_customizations;
CREATE TRIGGER update_role_ui_customizations_updated_at BEFORE UPDATE ON role_ui_customizations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_stores_updated_at ON stores;
CREATE TRIGGER update_stores_updated_at BEFORE UPDATE ON stores FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_pos_terminals_updated_at ON pos_terminals;
CREATE TRIGGER update_pos_terminals_updated_at BEFORE UPDATE ON pos_terminals FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_product_categories_updated_at ON product_categories;
CREATE TRIGGER update_product_categories_updated_at BEFORE UPDATE ON product_categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_brands_updated_at ON brands;
CREATE TRIGGER update_brands_updated_at BEFORE UPDATE ON brands FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_price_lists_updated_at ON price_lists;
CREATE TRIGGER update_price_lists_updated_at BEFORE UPDATE ON price_lists FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_product_variants_updated_at ON product_variants;
CREATE TRIGGER update_product_variants_updated_at BEFORE UPDATE ON product_variants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_product_prices_updated_at ON product_prices;
CREATE TRIGGER update_product_prices_updated_at BEFORE UPDATE ON product_prices FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_product_serial_numbers_updated_at ON product_serial_numbers;
CREATE TRIGGER update_product_serial_numbers_updated_at BEFORE UPDATE ON product_serial_numbers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_product_batches_updated_at ON product_batches;
CREATE TRIGGER update_product_batches_updated_at BEFORE UPDATE ON product_batches FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_inventory_stock_updated_at ON inventory_stock;
CREATE TRIGGER update_inventory_stock_updated_at BEFORE UPDATE ON inventory_stock FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_suppliers_updated_at ON suppliers;
CREATE TRIGGER update_suppliers_updated_at BEFORE UPDATE ON suppliers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_customers_updated_at ON customers;
CREATE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_purchase_orders_updated_at ON purchase_orders;
CREATE TRIGGER update_purchase_orders_updated_at BEFORE UPDATE ON purchase_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_transfer_requests_updated_at ON transfer_requests;
CREATE TRIGGER update_transfer_requests_updated_at BEFORE UPDATE ON transfer_requests FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_goods_receipt_notes_updated_at ON goods_receipt_notes;
CREATE TRIGGER update_goods_receipt_notes_updated_at BEFORE UPDATE ON goods_receipt_notes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_sales_orders_updated_at ON sales_orders;
CREATE TRIGGER update_sales_orders_updated_at BEFORE UPDATE ON sales_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_sales_analytics_updated_at ON sales_analytics;
CREATE TRIGGER update_sales_analytics_updated_at BEFORE UPDATE ON sales_analytics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_purchase_analytics_updated_at ON purchase_analytics;
CREATE TRIGGER update_purchase_analytics_updated_at BEFORE UPDATE ON purchase_analytics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_inventory_analytics_updated_at ON inventory_analytics;
CREATE TRIGGER update_inventory_analytics_updated_at BEFORE UPDATE ON inventory_analytics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_profit_loss_analytics_updated_at ON profit_loss_analytics;
CREATE TRIGGER update_profit_loss_analytics_updated_at BEFORE UPDATE ON profit_loss_analytics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS update_discount_analytics_updated_at ON discount_analytics;
CREATE TRIGGER update_discount_analytics_updated_at BEFORE UPDATE ON discount_analytics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Restaurant triggers
DROP TRIGGER IF EXISTS trg_restaurant_tables_updated_at ON restaurant_tables;
CREATE TRIGGER trg_restaurant_tables_updated_at BEFORE UPDATE ON restaurant_tables FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS trg_menu_categories_updated_at ON menu_categories;
CREATE TRIGGER trg_menu_categories_updated_at BEFORE UPDATE ON menu_categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS trg_menu_items_updated_at ON menu_items;
CREATE TRIGGER trg_menu_items_updated_at BEFORE UPDATE ON menu_items FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS trg_recipes_updated_at ON recipes;
CREATE TRIGGER trg_recipes_updated_at BEFORE UPDATE ON recipes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS trg_restaurant_orders_updated_at ON restaurant_orders;
CREATE TRIGGER trg_restaurant_orders_updated_at BEFORE UPDATE ON restaurant_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
DROP TRIGGER IF EXISTS trg_restaurant_order_items_updated_at ON restaurant_order_items;
CREATE TRIGGER trg_restaurant_order_items_updated_at BEFORE UPDATE ON restaurant_order_items FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- INDEXES FOR PERFORMANCE
-- =====================================================

-- Organizations
CREATE INDEX idx_organizations_code ON organizations(code);
CREATE INDEX idx_organizations_is_active ON organizations(is_active);

-- Tenants
CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_is_active ON tenants(is_active);

-- Modules
CREATE INDEX idx_modules_code ON modules(code);
CREATE INDEX idx_modules_is_active ON modules(is_active);
CREATE INDEX idx_modules_display_order ON modules(display_order);

-- Menus
CREATE INDEX idx_menus_module_id ON menus(module_id);
CREATE INDEX idx_menus_parent_menu_id ON menus(parent_menu_id);
CREATE INDEX idx_menus_is_active ON menus(is_active);
CREATE INDEX idx_menus_display_order ON menus(display_order);

-- Submenus
CREATE INDEX idx_submenus_menu_id ON submenus(menu_id);
CREATE INDEX idx_submenus_parent_submenu_id ON submenus(parent_submenu_id);
CREATE INDEX idx_submenus_is_active ON submenus(is_active);
CREATE INDEX idx_submenus_display_order ON submenus(display_order);

-- Permissions
CREATE INDEX idx_permissions_code ON permissions(code);

-- Roles
CREATE INDEX idx_roles_code ON roles(code);
CREATE INDEX idx_roles_is_active ON roles(is_active);

-- Role Permissions
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

-- Stores
CREATE INDEX idx_stores_organization_id ON stores(organization_id);
CREATE INDEX idx_stores_parent_store_id ON stores(parent_store_id);
CREATE INDEX idx_stores_code ON stores(code);
CREATE INDEX idx_stores_is_active ON stores(is_active);
CREATE INDEX idx_stores_store_type ON stores(store_type);

-- Storage Locations
CREATE INDEX idx_storage_locations_store_id ON storage_locations(store_id);
CREATE INDEX idx_storage_locations_parent_location_id ON storage_locations(parent_location_id);
CREATE INDEX idx_storage_locations_code ON storage_locations(code);

-- Users
CREATE INDEX idx_users_organization_id ON users(organization_id);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_employee_code ON users(employee_code);
CREATE INDEX idx_users_is_active ON users(is_active);

-- User Roles
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

-- User Store Access
CREATE INDEX idx_user_store_access_user_id ON user_store_access(user_id);
CREATE INDEX idx_user_store_access_store_id ON user_store_access(store_id);

-- Cashiers
CREATE INDEX idx_cashiers_user_id ON cashiers(user_id);
CREATE INDEX idx_cashiers_store_id ON cashiers(store_id);
CREATE INDEX idx_cashiers_is_active ON cashiers(is_active);

-- POS Terminals
CREATE INDEX idx_pos_terminals_store_id ON pos_terminals(store_id);
CREATE INDEX idx_pos_terminals_is_active ON pos_terminals(is_active);

-- Cashier Sessions
CREATE INDEX idx_cashier_sessions_cashier_id ON cashier_sessions(cashier_id);
CREATE INDEX idx_cashier_sessions_pos_terminal_id ON cashier_sessions(pos_terminal_id);
CREATE INDEX idx_cashier_sessions_status ON cashier_sessions(status);
CREATE INDEX idx_cashier_sessions_opening_time ON cashier_sessions(opening_time);

-- Product Categories
CREATE INDEX idx_product_categories_parent_category_id ON product_categories(parent_category_id);
CREATE INDEX idx_product_categories_code ON product_categories(code);
CREATE INDEX idx_product_categories_is_active ON product_categories(is_active);

-- Brands
CREATE INDEX idx_brands_code ON brands(code);
CREATE INDEX idx_brands_is_active ON brands(is_active);

-- Units of Measure
CREATE INDEX idx_units_of_measure_code ON units_of_measure(code);
CREATE INDEX idx_units_of_measure_uom_type ON units_of_measure(uom_type);

-- UOM Packaging Templates
CREATE INDEX idx_uom_packaging_templates_organization_id ON uom_packaging_templates(organization_id);
CREATE INDEX idx_uom_packaging_templates_code ON uom_packaging_templates(code);
CREATE INDEX idx_uom_pkg_template_levels_template_id ON uom_packaging_template_levels(template_id);
CREATE INDEX idx_uom_pkg_template_levels_uom_id ON uom_packaging_template_levels(uom_id);

-- Price Lists
CREATE INDEX idx_price_lists_code ON price_lists(code);
CREATE INDEX idx_price_lists_is_active ON price_lists(is_active);
CREATE INDEX idx_price_lists_valid_from ON price_lists(valid_from);
CREATE INDEX idx_price_lists_valid_to ON price_lists(valid_to);

-- Tax Categories
CREATE INDEX idx_tax_categories_code ON tax_categories(code);
CREATE INDEX idx_tax_categories_is_active ON tax_categories(is_active);

-- Products
CREATE INDEX idx_products_organization_id ON products(organization_id);
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_brand_id ON products(brand_id);
CREATE INDEX idx_products_is_active ON products(is_active);
CREATE INDEX idx_products_is_sellable ON products(is_sellable);
CREATE INDEX idx_products_is_purchasable ON products(is_purchasable);
CREATE INDEX idx_products_product_type ON products(product_type);

-- Product Variants
CREATE INDEX idx_product_variants_product_id ON product_variants(product_id);
CREATE INDEX idx_product_variants_variant_sku ON product_variants(variant_sku);
CREATE INDEX idx_product_variants_is_active ON product_variants(is_active);

-- Product Barcodes
CREATE INDEX idx_product_barcodes_product_id ON product_barcodes(product_id);
CREATE INDEX idx_product_barcodes_product_variant_id ON product_barcodes(product_variant_id);
CREATE INDEX idx_product_barcodes_barcode ON product_barcodes(barcode);

-- Product Prices
CREATE INDEX idx_product_prices_product_id ON product_prices(product_id);
CREATE INDEX idx_product_prices_product_variant_id ON product_prices(product_variant_id);
CREATE INDEX idx_product_prices_price_list_id ON product_prices(price_list_id);
CREATE INDEX idx_product_prices_is_active ON product_prices(is_active);

-- Product Serial Numbers
CREATE INDEX idx_product_serial_numbers_product_id ON product_serial_numbers(product_id);
CREATE INDEX idx_product_serial_numbers_serial_number ON product_serial_numbers(serial_number);
CREATE INDEX idx_product_serial_numbers_status ON product_serial_numbers(status);
CREATE INDEX idx_product_serial_numbers_current_store_id ON product_serial_numbers(current_store_id);

-- Product Batches
CREATE INDEX idx_product_batches_product_id ON product_batches(product_id);
CREATE INDEX idx_product_batches_batch_number ON product_batches(batch_number);
CREATE INDEX idx_product_batches_store_id ON product_batches(store_id);
CREATE INDEX idx_product_batches_status ON product_batches(status);
CREATE INDEX idx_product_batches_expiry_date ON product_batches(expiry_date);

-- Inventory Stock
CREATE INDEX idx_inventory_stock_product_id ON inventory_stock(product_id);
CREATE INDEX idx_inventory_stock_product_variant_id ON inventory_stock(product_variant_id);
CREATE INDEX idx_inventory_stock_store_id ON inventory_stock(store_id);
CREATE INDEX idx_inventory_stock_storage_location_id ON inventory_stock(storage_location_id);

-- Stock Movements
CREATE INDEX idx_stock_movements_product_id ON stock_movements(product_id);
CREATE INDEX idx_stock_movements_from_store_id ON stock_movements(from_store_id);
CREATE INDEX idx_stock_movements_to_store_id ON stock_movements(to_store_id);
CREATE INDEX idx_stock_movements_movement_type ON stock_movements(movement_type);
CREATE INDEX idx_stock_movements_movement_date ON stock_movements(movement_date);
CREATE INDEX idx_stock_movements_reference_type_id ON stock_movements(reference_type, reference_id);

-- Stock Counts
CREATE INDEX idx_stock_counts_store_id ON stock_counts(store_id);
CREATE INDEX idx_stock_counts_status ON stock_counts(status);
CREATE INDEX idx_stock_counts_count_number ON stock_counts(count_number);

-- Stock Count Lines
CREATE INDEX idx_stock_count_lines_stock_count_id ON stock_count_lines(stock_count_id);
CREATE INDEX idx_stock_count_lines_product_id ON stock_count_lines(product_id);

-- Suppliers
CREATE INDEX idx_suppliers_organization_id ON suppliers(organization_id);
CREATE INDEX idx_suppliers_code ON suppliers(code);
CREATE INDEX idx_suppliers_is_active ON suppliers(is_active);

-- Customers
CREATE INDEX idx_customers_organization_id ON customers(organization_id);
CREATE INDEX idx_customers_customer_code ON customers(customer_code);
CREATE INDEX idx_customers_is_active ON customers(is_active);
CREATE INDEX idx_customers_customer_type ON customers(customer_type);

-- Purchase Orders
CREATE INDEX idx_purchase_orders_organization_id ON purchase_orders(organization_id);
CREATE INDEX idx_purchase_orders_supplier_id ON purchase_orders(supplier_id);
CREATE INDEX idx_purchase_orders_store_id ON purchase_orders(store_id);
CREATE INDEX idx_purchase_orders_po_number ON purchase_orders(po_number);
CREATE INDEX idx_purchase_orders_status ON purchase_orders(status);
CREATE INDEX idx_purchase_orders_po_date ON purchase_orders(po_date);

-- Purchase Order Lines
CREATE INDEX idx_purchase_order_lines_purchase_order_id ON purchase_order_lines(purchase_order_id);
CREATE INDEX idx_purchase_order_lines_product_id ON purchase_order_lines(product_id);

-- Sales Orders
CREATE INDEX idx_sales_orders_organization_id ON sales_orders(organization_id);
CREATE INDEX idx_sales_orders_customer_id ON sales_orders(customer_id);
CREATE INDEX idx_sales_orders_store_id ON sales_orders(store_id);
CREATE INDEX idx_sales_orders_order_number ON sales_orders(order_number);
CREATE INDEX idx_sales_orders_status ON sales_orders(status);
CREATE INDEX idx_sales_orders_order_date ON sales_orders(order_date);

-- Sales Order Lines
CREATE INDEX idx_sales_order_lines_sales_order_id ON sales_order_lines(sales_order_id);
CREATE INDEX idx_sales_order_lines_product_id ON sales_order_lines(product_id);

-- POS Transactions
CREATE INDEX idx_pos_transactions_store_id ON pos_transactions(store_id);
CREATE INDEX idx_pos_transactions_cashier_id ON pos_transactions(cashier_id);
CREATE INDEX idx_pos_transactions_cashier_session_id ON pos_transactions(cashier_session_id);
CREATE INDEX idx_pos_transactions_customer_id ON pos_transactions(customer_id);
CREATE INDEX idx_pos_transactions_transaction_number ON pos_transactions(transaction_number);
CREATE INDEX idx_pos_transactions_transaction_date ON pos_transactions(transaction_date);
CREATE INDEX idx_pos_transactions_status ON pos_transactions(status);

-- POS Transaction Lines
CREATE INDEX idx_pos_transaction_lines_transaction_id ON pos_transaction_lines(transaction_id);
CREATE INDEX idx_pos_transaction_lines_product_id ON pos_transaction_lines(product_id);

-- FIX #7: returns indexes
CREATE INDEX idx_sales_returns_store_id              ON sales_returns(store_id);
CREATE INDEX idx_sales_returns_original_transaction  ON sales_returns(original_transaction_id);
CREATE INDEX idx_sales_returns_status                ON sales_returns(status);
CREATE INDEX idx_sales_return_lines_return_id        ON sales_return_lines(return_id);

-- POS Payments
CREATE INDEX idx_pos_payments_transaction_id ON pos_payments(transaction_id);
CREATE INDEX idx_pos_payments_payment_method ON pos_payments(payment_method);

-- Sales Analytics
CREATE INDEX idx_sales_analytics_organization_id ON sales_analytics(organization_id);
CREATE INDEX idx_sales_analytics_store_id ON sales_analytics(store_id);
CREATE INDEX idx_sales_analytics_product_id ON sales_analytics(product_id);
CREATE INDEX idx_sales_analytics_category_id ON sales_analytics(category_id);
CREATE INDEX idx_sales_analytics_customer_id ON sales_analytics(customer_id);
CREATE INDEX idx_sales_analytics_date ON sales_analytics(date);
CREATE INDEX idx_sales_analytics_year_month ON sales_analytics(year, month);

-- Purchase Analytics
CREATE INDEX idx_purchase_analytics_organization_id ON purchase_analytics(organization_id);
CREATE INDEX idx_purchase_analytics_store_id ON purchase_analytics(store_id);
CREATE INDEX idx_purchase_analytics_supplier_id ON purchase_analytics(supplier_id);
CREATE INDEX idx_purchase_analytics_product_id ON purchase_analytics(product_id);
CREATE INDEX idx_purchase_analytics_date ON purchase_analytics(date);

-- Inventory Analytics
CREATE INDEX idx_inventory_analytics_organization_id ON inventory_analytics(organization_id);
CREATE INDEX idx_inventory_analytics_store_id ON inventory_analytics(store_id);
CREATE INDEX idx_inventory_analytics_product_id ON inventory_analytics(product_id);
CREATE INDEX idx_inventory_analytics_date ON inventory_analytics(date);

-- FIX #1: stock reservations indexes
CREATE INDEX idx_stock_reservations_product_id         ON stock_reservations(product_id);
CREATE INDEX idx_stock_reservations_product_variant_id ON stock_reservations(product_variant_id);
CREATE INDEX idx_stock_reservations_store_id           ON stock_reservations(store_id);
CREATE INDEX idx_stock_reservations_reference          ON stock_reservations(reference_type, reference_id);
CREATE INDEX idx_stock_reservations_status             ON stock_reservations(status);
CREATE INDEX idx_stock_reservations_expires_at         ON stock_reservations(expires_at);

-- Profit Loss Analytics
CREATE INDEX idx_profit_loss_analytics_organization_id ON profit_loss_analytics(organization_id);
CREATE INDEX idx_profit_loss_analytics_store_id ON profit_loss_analytics(store_id);
CREATE INDEX idx_profit_loss_analytics_date ON profit_loss_analytics(date);
CREATE INDEX idx_profit_loss_analytics_period_type ON profit_loss_analytics(period_type);

-- Discount Analytics
CREATE INDEX idx_discount_analytics_organization_id ON discount_analytics(organization_id);
CREATE INDEX idx_discount_analytics_store_id ON discount_analytics(store_id);
CREATE INDEX idx_discount_analytics_cashier_id ON discount_analytics(cashier_id);
CREATE INDEX idx_discount_analytics_date ON discount_analytics(date);

-- Restaurant Module Indexes
CREATE INDEX idx_restaurant_tables_store_id         ON restaurant_tables(store_id);
CREATE INDEX idx_restaurant_tables_is_active        ON restaurant_tables(is_active);
CREATE INDEX idx_restaurant_tables_section          ON restaurant_tables(section);

CREATE INDEX idx_menu_categories_store_id           ON menu_categories(store_id);
CREATE INDEX idx_menu_categories_parent_id          ON menu_categories(parent_category_id);
CREATE INDEX idx_menu_categories_is_active          ON menu_categories(is_active);
CREATE INDEX idx_menu_categories_display_order      ON menu_categories(display_order);

CREATE INDEX idx_menu_items_store_id                ON menu_items(store_id);
CREATE INDEX idx_menu_items_category_id             ON menu_items(menu_category_id);
CREATE INDEX idx_menu_items_product_id              ON menu_items(product_id);
CREATE INDEX idx_menu_items_recipe_id               ON menu_items(recipe_id);
CREATE INDEX idx_menu_items_is_active               ON menu_items(is_active);
CREATE INDEX idx_menu_items_is_available            ON menu_items(is_available);
CREATE INDEX idx_menu_items_display_order           ON menu_items(display_order);

CREATE INDEX idx_menu_item_modifiers_item_id        ON menu_item_modifiers(menu_item_id);
CREATE INDEX idx_menu_item_modifiers_is_active      ON menu_item_modifiers(is_active);

CREATE INDEX idx_recipes_organization_id            ON recipes(organization_id);
CREATE INDEX idx_recipes_finished_product_id        ON recipes(finished_product_id);
CREATE INDEX idx_recipes_is_active                  ON recipes(is_active);
CREATE INDEX idx_recipes_code                       ON recipes(recipe_code);

CREATE INDEX idx_recipe_ingredients_recipe_id       ON recipe_ingredients(recipe_id);
CREATE INDEX idx_recipe_ingredients_product_id      ON recipe_ingredients(product_id);
CREATE INDEX idx_recipe_ingredients_variant_id      ON recipe_ingredients(product_variant_id);

CREATE INDEX idx_restaurant_orders_store_id         ON restaurant_orders(store_id);
CREATE INDEX idx_restaurant_orders_table_id         ON restaurant_orders(table_id);
CREATE INDEX idx_restaurant_orders_cashier_id       ON restaurant_orders(cashier_id);
CREATE INDEX idx_restaurant_orders_session_id       ON restaurant_orders(cashier_session_id);
CREATE INDEX idx_restaurant_orders_customer_id      ON restaurant_orders(customer_id);
CREATE INDEX idx_restaurant_orders_status           ON restaurant_orders(status);
CREATE INDEX idx_restaurant_orders_source           ON restaurant_orders(order_source);
CREATE INDEX idx_restaurant_orders_ordered_at       ON restaurant_orders(ordered_at);
CREATE INDEX idx_restaurant_orders_pos_txn_id       ON restaurant_orders(pos_transaction_id);
CREATE INDEX idx_restaurant_orders_store_status_time ON restaurant_orders(store_id, status, ordered_at);

CREATE INDEX idx_restaurant_order_items_order_id    ON restaurant_order_items(order_id);
CREATE INDEX idx_restaurant_order_items_menu_item   ON restaurant_order_items(menu_item_id);
CREATE INDEX idx_restaurant_order_items_status      ON restaurant_order_items(status);

CREATE INDEX idx_waste_logs_store_id                ON waste_logs(store_id);
CREATE INDEX idx_waste_logs_product_id              ON waste_logs(product_id);
CREATE INDEX idx_waste_logs_menu_item_id            ON waste_logs(menu_item_id);
CREATE INDEX idx_waste_logs_recipe_id               ON waste_logs(recipe_id);
CREATE INDEX idx_waste_logs_waste_source            ON waste_logs(waste_source);
CREATE INDEX idx_waste_logs_wasted_at               ON waste_logs(wasted_at);
CREATE INDEX idx_waste_logs_order_id                ON waste_logs(order_id);
CREATE INDEX idx_waste_logs_store_source_date       ON waste_logs(store_id, waste_source, wasted_at);

CREATE INDEX idx_kiosk_sessions_terminal_id         ON kiosk_sessions(pos_terminal_id);
CREATE INDEX idx_kiosk_sessions_store_id            ON kiosk_sessions(store_id);
CREATE INDEX idx_kiosk_sessions_status              ON kiosk_sessions(status);
CREATE INDEX idx_kiosk_sessions_token               ON kiosk_sessions(session_token);


-- FIX #11 index
-- CREATE INDEX idx_menu_availability_menu_item_id ON menu_item_availability_schedules(menu_item_id);
-- CREATE INDEX idx_menu_availability_day          ON menu_item_availability_schedules(day_of_week);
-- CREATE INDEX idx_recipes_organization_id        ON recipes(organization_id);
-- CREATE INDEX idx_recipes_finished_product_id    ON recipes(finished_product_id);
-- CREATE INDEX idx_recipes_is_active              ON recipes(is_active);
-- CREATE INDEX idx_recipes_code                   ON recipes(recipe_code);
-- CREATE INDEX idx_recipe_ingredients_recipe_id   ON recipe_ingredients(recipe_id);
-- CREATE INDEX idx_recipe_ingredients_product_id  ON recipe_ingredients(product_id);
-- CREATE INDEX idx_restaurant_orders_store_id     ON restaurant_orders(store_id);
-- CREATE INDEX idx_restaurant_orders_table_id     ON restaurant_orders(table_id);
-- CREATE INDEX idx_restaurant_orders_cashier_id   ON restaurant_orders(cashier_id);
-- CREATE INDEX idx_restaurant_orders_status       ON restaurant_orders(status);
-- CREATE INDEX idx_restaurant_orders_ordered_at   ON restaurant_orders(ordered_at);
-- CREATE INDEX idx_restaurant_order_items_order_id ON restaurant_order_items(order_id);
-- CREATE INDEX idx_restaurant_order_items_menu_item ON restaurant_order_items(menu_item_id);
-- CREATE INDEX idx_restaurant_order_items_status   ON restaurant_order_items(status);
-- FIX #13 indexes
CREATE INDEX idx_combo_bundles_store_id  ON combo_bundles(store_id);
CREATE INDEX idx_combo_bundles_is_active ON combo_bundles(is_active);
CREATE INDEX idx_combo_bundle_items_bundle_id ON combo_bundle_items(combo_bundle_id);
-- CREATE INDEX idx_waste_logs_store_id     ON waste_logs(store_id);
-- CREATE INDEX idx_waste_logs_wasted_at    ON waste_logs(wasted_at);
-- CREATE INDEX idx_kiosk_sessions_token    ON kiosk_sessions(session_token);
-- Additional POS Indexes
CREATE INDEX IF NOT EXISTS idx_product_barcodes_barcode_lookup 
ON product_barcodes(barcode) WHERE is_primary = true;

CREATE INDEX IF NOT EXISTS idx_products_sku_varchar_pattern 
ON products(sku varchar_pattern_ops);

CREATE INDEX IF NOT EXISTS idx_inventory_stock_store_product_qty 
ON inventory_stock(store_id, product_id, quantity_available);

CREATE INDEX IF NOT EXISTS idx_products_active_sellable 
ON products(is_active, is_sellable) WHERE is_active = true AND is_sellable = true;

-- =====================================================
-- POS VIEWS AND FUNCTIONS (with Type Fixes)
-- =====================================================

CREATE OR REPLACE VIEW vw_pos_product_catalog AS
SELECT
    -- Identity
    p.id                            AS product_id,
    pv.id                           AS product_variant_id,
    p.sku                           AS base_sku,
    COALESCE(pv.variant_sku, p.sku) AS sku,
    p.name                          AS base_product_name,
    COALESCE(pv.variant_name, p.name) AS product_name,
    pv.variant_attributes,
    p.description,
    p.product_type,
    -- Category
    pc.id                           AS category_id,
    pc.name                         AS category_name,
    pc.code                         AS category_code,
    pc_parent.id                    AS parent_category_id,
    pc_parent.name                  AS parent_category_name,
    -- Brand
    b.id                            AS brand_id,
    b.name                          AS brand_name,
    -- UOM
    uom.id                          AS uom_id,
    uom.code                        AS uom_code,
    uom.name                        AS uom_name,
    uom.decimal_places,
    -- Barcode: prefer variant barcode, fall back to base product barcode
    COALESCE(pb_variant.barcode, pb_base.barcode) AS barcode,
    COALESCE(pb_variant.barcode_type, pb_base.barcode_type) AS barcode_type,
    -- Tax
    tc.id                           AS tax_category_id,
    tc.name                         AS tax_category_name,
    tc.tax_rate,
    tc.is_inclusive                 AS tax_is_inclusive,
    -- Prices: prefer variant-level prices, fall back to product-level
    COALESCE(pp_retail_v.price, pp_retail.price)   AS retail_price,
    COALESCE(pp_retail_v.id,    pp_retail.id)       AS retail_price_id,
    COALESCE(pp_promo_v.price,  pp_promo.price)     AS promo_price,
    COALESCE(pp_promo_v.id,     pp_promo.id)        AS promo_price_id,
    COALESCE(pp_promo_v.min_quantity, pp_promo.min_quantity) AS promo_min_quantity,
    COALESCE(pp_promo_v.valid_from,   pp_promo.valid_from)   AS promo_valid_from,
    COALESCE(pp_promo_v.valid_to,     pp_promo.valid_to)     AS promo_valid_to,
    COALESCE(pp_promo_v.metadata->>'promotion_name', pp_promo.metadata->>'promotion_name') AS promotion_name,
    COALESCE(pp_promo_v.metadata->>'discount_percent', pp_promo.metadata->>'discount_percent') AS discount_percent,
    -- Effective price
    CASE
        WHEN COALESCE(pp_promo_v.id, pp_promo.id) IS NOT NULL
             AND COALESCE(pp_promo_v.is_active, pp_promo.is_active) = true
             AND COALESCE(pp_promo_v.valid_from, pp_promo.valid_from) <= CURRENT_DATE
             AND (COALESCE(pp_promo_v.valid_to, pp_promo.valid_to) IS NULL
                  OR COALESCE(pp_promo_v.valid_to, pp_promo.valid_to) >= CURRENT_DATE)
        THEN COALESCE(pp_promo_v.price, pp_promo.price)
        ELSE COALESCE(pp_retail_v.price, pp_retail.price)
    END AS effective_price,
    -- Active promotion flag
    CASE
        WHEN COALESCE(pp_promo_v.id, pp_promo.id) IS NOT NULL
             AND COALESCE(pp_promo_v.is_active, pp_promo.is_active) = true
             AND COALESCE(pp_promo_v.valid_from, pp_promo.valid_from) <= CURRENT_DATE
             AND (COALESCE(pp_promo_v.valid_to, pp_promo.valid_to) IS NULL
                  OR COALESCE(pp_promo_v.valid_to, pp_promo.valid_to) >= CURRENT_DATE)
        THEN true
        ELSE false
    END AS has_active_promotion,
    p.is_active,
    p.is_sellable,
    p.is_serialized,
    p.is_batch_managed,
    p.allow_decimal_quantity,
    p.track_inventory,
    p.metadata AS product_metadata
FROM products p
LEFT JOIN product_variants      pv           ON pv.product_id = p.id AND pv.is_active = true
LEFT JOIN product_categories    pc           ON p.category_id = pc.id
LEFT JOIN product_categories    pc_parent    ON pc.parent_category_id = pc_parent.id
LEFT JOIN brands                b            ON p.brand_id = b.id
LEFT JOIN units_of_measure      uom          ON p.base_uom_id = uom.id
-- Base product barcodes
LEFT JOIN product_barcodes      pb_base      ON p.id = pb_base.product_id AND pb_base.product_variant_id IS NULL AND pb_base.is_primary = true
-- Variant barcodes (FIX 9.1 / 9.4)
LEFT JOIN product_barcodes      pb_variant   ON pv.id = pb_variant.product_variant_id AND pb_variant.is_primary = true
LEFT JOIN tax_categories        tc           ON p.tax_category_id = tc.id
-- Retail price: base product
LEFT JOIN product_prices pp_retail
    ON p.id = pp_retail.product_id
    AND pp_retail.product_variant_id IS NULL
    AND pp_retail.price_list_id = (SELECT id FROM price_lists WHERE code = 'RETAIL' AND is_active = true LIMIT 1)
    AND pp_retail.is_active = true
-- Retail price: variant (FIX 9.2 / 9.3)
LEFT JOIN product_prices pp_retail_v
    ON p.id = pp_retail_v.product_id
    AND pp_retail_v.product_variant_id = pv.id
    AND pp_retail_v.price_list_id = (SELECT id FROM price_lists WHERE code = 'RETAIL' AND is_active = true LIMIT 1)
    AND pp_retail_v.is_active = true
-- Promo price: base product
LEFT JOIN product_prices pp_promo
    ON p.id = pp_promo.product_id
    AND pp_promo.product_variant_id IS NULL
    AND pp_promo.price_list_id = (SELECT id FROM price_lists WHERE code = 'PROMO' AND is_active = true LIMIT 1)
    AND pp_promo.is_active = true
-- Promo price: variant
LEFT JOIN product_prices pp_promo_v
    ON p.id = pp_promo_v.product_id
    AND pp_promo_v.product_variant_id = pv.id
    AND pp_promo_v.price_list_id = (SELECT id FROM price_lists WHERE code = 'PROMO' AND is_active = true LIMIT 1)
    AND pp_promo_v.is_active = true
WHERE p.is_active = true
  AND p.is_sellable = true
ORDER BY pc.name, p.name, pv.variant_name;
-- =====================================================
-- FIX #9 (continued): Variant-aware POS functions
-- =====================================================

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_pos_get_products_with_stock(
    p_store_id              INTEGER,
    p_category_id           INTEGER  DEFAULT NULL,
    p_search_term           VARCHAR  DEFAULT NULL,
    p_include_out_of_stock  BOOLEAN  DEFAULT false
)
RETURNS TABLE (
    product_id          INTEGER,
    product_variant_id  INTEGER,
    sku                 VARCHAR,
    product_name        VARCHAR,
    variant_attributes  JSONB,
    description         TEXT,
    category_id         INTEGER,
    category_name       VARCHAR,
    brand_name          VARCHAR,
    barcode             VARCHAR,
    uom_code            VARCHAR,
    decimal_places      INTEGER,
    retail_price        NUMERIC,
    promo_price         NUMERIC,
    effective_price     NUMERIC,
    has_promotion       BOOLEAN,
    promotion_name      VARCHAR,
    discount_percent    VARCHAR,
    promo_min_quantity  NUMERIC,
    tax_rate            NUMERIC,
    tax_is_inclusive    BOOLEAN,
    quantity_available  NUMERIC,
    quantity_on_hand    NUMERIC,
    quantity_allocated  NUMERIC,
    is_in_stock         BOOLEAN,
    is_low_stock        BOOLEAN,
    reorder_level       NUMERIC,
    allow_decimal_quantity BOOLEAN,
    is_serialized       BOOLEAN,
    is_batch_managed    BOOLEAN,
    product_metadata    JSONB,
    product_variants    JSONB,
    package_n_price     JSONB,
    product_uom_conversions JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        cat.product_id,
        cat.product_variant_id,
        cat.sku::VARCHAR,
        cat.product_name::VARCHAR,
        cat.variant_attributes,
        cat.description,
        cat.category_id,
        cat.category_name::VARCHAR,
        cat.brand_name::VARCHAR,
        cat.barcode::VARCHAR,
        cat.uom_code::VARCHAR,
        cat.decimal_places::INTEGER,
        cat.retail_price,
        COALESCE(cat.promo_price, promo_rule.calculated_promo_price) AS promo_price,
        COALESCE(cat.effective_price, promo_rule.calculated_promo_price, cat.retail_price) AS effective_price,
        COALESCE(cat.has_active_promotion, (promo_rule.promo_name IS NOT NULL)) AS has_promotion,
        COALESCE(cat.promotion_name, promo_rule.promo_name)::VARCHAR AS promotion_name,
        COALESCE(cat.discount_percent, promo_rule.calc_discount_percent)::VARCHAR AS discount_percent,
        COALESCE(cat.promo_min_quantity, promo_rule.promo_min_qty) AS promo_min_quantity,
        cat.tax_rate,
        cat.tax_is_inclusive,
        -- FIX 9.3: Join on both product_id AND product_variant_id
        COALESCE(inv.quantity_available, 0)::NUMERIC,
        COALESCE(inv.quantity_on_hand,   0)::NUMERIC,
        COALESCE(inv.quantity_allocated, 0)::NUMERIC,
        (COALESCE(inv.quantity_available, 0) > 0),
        (COALESCE(inv.quantity_available, 0) <= COALESCE(inv.reorder_level, 0) AND COALESCE(inv.quantity_available, 0) > 0),
        COALESCE(inv.reorder_level, 0)::NUMERIC,
        cat.allow_decimal_quantity,
        cat.is_serialized,
        cat.is_batch_managed,
        cat.product_metadata,
        (SELECT COALESCE(jsonb_agg(
            jsonb_build_object(
                'id', pv.id,
                'product_id', pv.product_id,
                'variant_sku', pv.variant_sku,
                'variant_name', pv.variant_name,
                'variant_attributes', pv.variant_attributes,
                'price', (
                    SELECT ppv.price
                    FROM product_prices ppv
                    INNER JOIN price_lists plv ON ppv.price_list_id = plv.id AND plv.is_active = true
                    WHERE ppv.product_id = pv.product_id
                      AND ppv.product_variant_id = pv.id
                      AND ppv.is_active = true
                      AND (ppv.valid_from IS NULL OR ppv.valid_from <= CURRENT_DATE)
                      AND (ppv.valid_to   IS NULL OR ppv.valid_to   >= CURRENT_DATE)
                    ORDER BY ppv.valid_from DESC NULLS LAST, ppv.id DESC
                    LIMIT 1
                ),
                'is_active', pv.is_active,
                'metadata', COALESCE(pv.metadata, '{}'::jsonb),
                'created_at', pv.created_at,
                'updated_at', pv.updated_at
            )
            ORDER BY pv.id
        ), '[]'::jsonb)
         FROM product_variants pv
         WHERE pv.product_id = cat.product_id),
        -- FIX 9.2: Include variant_id and variant_attributes in package_n_price JSON
        (SELECT COALESCE(jsonb_agg(s.rec ORDER BY s.pl_code, s.uom_code), '[]'::jsonb)
         FROM (
             SELECT pl.code AS pl_code, uom.code AS uom_code,
                    jsonb_build_object(
                        'product_name',        cat.product_name,
                        'price_list_id',       pl.id,
                        'price_list_code',     pl.code,
                        'price_list_name',     pl.name,
                        'price_list',          pl.name,
                        'price_list_type',     pl.price_list_type,
                        'currency_code',       pl.currency_code,
                        'uom_id',              uom.id,
                        'uom_code',            uom.code,
                        'uom',                 uom.code,
                        'uom_name',            uom.name,
                        'decimal_places',      uom.decimal_places,
                        'price',               pp.price,
                        'min_quantity',        pp.min_quantity,
                        'max_quantity',        pp.max_quantity,
                        'valid_from',          pp.valid_from,
                        'valid_to',            pp.valid_to,
                        'metadata',            COALESCE(pp.metadata, '{}'::jsonb),
                        'barcodes',            (SELECT COALESCE(jsonb_agg(pb.barcode), '[]'::jsonb)
                                                FROM product_barcodes pb
                                                WHERE pb.product_id = pp.product_id
                                                  AND (pb.product_variant_id = cat.product_variant_id
                                                       OR (cat.product_variant_id IS NULL AND pb.product_variant_id IS NULL)))
                    ) AS rec
             FROM product_prices pp
             INNER JOIN price_lists pl ON pp.price_list_id = pl.id AND pl.is_active = true
             LEFT JOIN units_of_measure uom ON pp.uom_id = uom.id
             WHERE pp.product_id = cat.product_id
               AND (pp.product_variant_id = cat.product_variant_id
                    OR (cat.product_variant_id IS NULL AND pp.product_variant_id IS NULL))
               AND pp.is_active = true
               AND (pp.valid_from IS NULL OR pp.valid_from <= CURRENT_DATE)
               AND (pp.valid_to   IS NULL OR pp.valid_to   >= CURRENT_DATE)
         ) AS s),
        (SELECT COALESCE(jsonb_agg(conv.cv ORDER BY conv.fu_code, conv.tu_code), '[]'::jsonb)
         FROM (
             SELECT fu.code AS fu_code, tu.code AS tu_code,
                    jsonb_build_object(
                        'from_uom_id', fu.id, 'from_uom_code', fu.code, 'from_uom_name', fu.name,
                        'to_uom_id',   tu.id, 'to_uom_code',   tu.code, 'to_uom_name',   tu.name,
                        'conversion_factor', puc.conversion_factor
                    ) AS cv
             FROM product_uom_conversions puc
             JOIN units_of_measure fu ON puc.from_uom_id = fu.id
             JOIN units_of_measure tu ON puc.to_uom_id   = tu.id
             WHERE puc.product_id = cat.product_id
         ) AS conv)
    FROM vw_pos_product_catalog cat
    LEFT JOIN LATERAL (
        SELECT 
            pr.name AS promo_name,
            pr.min_quantity AS promo_min_qty,
            pr.discount_value,
            pr.promotion_type,
            CASE 
                WHEN pr.promotion_type = 'percentage_discount' AND cat.retail_price IS NOT NULL THEN
                    ROUND(cat.retail_price * (1.0 - (pr.discount_value / 100.0)), 2)
                WHEN pr.promotion_type = 'fixed_discount' AND cat.retail_price IS NOT NULL THEN
                    GREATEST(0.00, cat.retail_price - pr.discount_value)
                ELSE cat.retail_price
            END AS calculated_promo_price,
            CASE
                WHEN pr.promotion_type = 'percentage_discount' AND pr.discount_value IS NOT NULL THEN
                    CONCAT(TRIM(TRAILING '.' FROM TRIM(TRAILING '0' FROM pr.discount_value::text)), '%')
                ELSE NULL
            END AS calc_discount_percent
        FROM promotions pr
        WHERE pr.is_active = true
          AND (pr.valid_from IS NULL OR pr.valid_from <= CURRENT_TIMESTAMP)
          AND (pr.valid_to IS NULL OR pr.valid_to >= CURRENT_TIMESTAMP)
          AND (cardinality(pr.store_ids) = 0 OR p_store_id = ANY(pr.store_ids))
          AND (
              pr.applies_to = 'all'
              OR (pr.applies_to = 'product' AND cat.product_id = ANY(pr.target_product_ids))
              OR (pr.applies_to = 'category' AND cat.category_id = ANY(pr.target_category_ids))
          )
        ORDER BY pr.created_at DESC
        LIMIT 1
    ) promo_rule ON true
    -- FIX 9.3: correct variant-aware inventory join
    LEFT JOIN inventory_stock inv
        ON inv.product_id = cat.product_id
        AND inv.store_id = p_store_id
        AND (inv.product_variant_id = cat.product_variant_id
             OR (cat.product_variant_id IS NULL AND inv.product_variant_id IS NULL))
    WHERE
        (p_category_id IS NULL OR cat.category_id = p_category_id)
        AND (p_search_term IS NULL
             OR cat.product_name ILIKE '%' || p_search_term || '%'
             OR cat.sku         ILIKE '%' || p_search_term || '%'
             OR cat.barcode     ILIKE '%' || p_search_term || '%')
        AND (p_include_out_of_stock = true OR COALESCE(inv.quantity_available, 0) > 0)
    ORDER BY cat.category_name, cat.product_name;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- =====================================================
-- FIX #19 (P1): Low stock alert view
-- =====================================================

CREATE OR REPLACE VIEW vw_low_stock_alerts AS
SELECT
    ist.id              AS inventory_stock_id,
    s.id                AS store_id,
    s.name              AS store_name,
    s.code              AS store_code,
    p.id                AS product_id,
    p.sku,
    COALESCE(pv.variant_sku, p.sku) AS effective_sku,
    p.name              AS product_name,
    COALESCE(pv.variant_name, '')   AS variant_name,
    pv.id               AS product_variant_id,
    pc.name             AS category_name,
    ist.quantity_on_hand,
    ist.quantity_available,
    ist.quantity_allocated,
    ist.reorder_level,
    ist.reorder_quantity,
    CASE
        WHEN ist.quantity_available <= 0                                              THEN 'out_of_stock'
        WHEN ist.quantity_available <= COALESCE(ist.reorder_level, 0)               THEN 'low_stock'
        WHEN ist.quantity_available <= COALESCE(ist.reorder_level, 0) * 1.5         THEN 'near_reorder'
        ELSE 'adequate'
    END AS stock_status,
    (ist.quantity_available <= 0)                                                    AS is_out_of_stock,
    (ist.quantity_available > 0 AND ist.quantity_available <= COALESCE(ist.reorder_level, 0)) AS is_low_stock,
    ist.last_counted_at
FROM inventory_stock ist
JOIN stores s   ON s.id = ist.store_id AND s.is_active = true
JOIN products p ON p.id = ist.product_id AND p.is_active = true AND p.track_inventory = true
LEFT JOIN product_variants pv ON pv.id = ist.product_variant_id
LEFT JOIN product_categories pc ON pc.id = p.category_id
WHERE ist.quantity_available <= COALESCE(ist.reorder_level, 0)
   OR ist.quantity_available <= 0
ORDER BY s.name, CASE WHEN ist.quantity_available <= 0 THEN 0 ELSE 1 END, p.name;

-- =====================================================
-- FIX #20 (P1): Pending / overdue purchase orders view
-- =====================================================

CREATE OR REPLACE VIEW vw_pending_purchase_orders AS
SELECT
    po.id                   AS po_id,
    po.po_number,
    po.po_date,
    po.expected_delivery_date,
    po.status,
    CURRENT_DATE - po.expected_delivery_date AS days_overdue,
    (po.expected_delivery_date < CURRENT_DATE AND po.status NOT IN ('received','cancelled','closed')) AS is_overdue,
    s.id    AS store_id,
    s.name  AS store_name,
    sup.id  AS supplier_id,
    sup.name AS supplier_name,
    sup.contact_person,
    sup.email AS supplier_email,
    po.subtotal,
    po.discount_amount,
    po.tax_amount,
    po.total_amount,
    (SELECT COALESCE(SUM(pol.quantity - pol.received_quantity), 0)
     FROM purchase_order_lines pol WHERE pol.purchase_order_id = po.id) AS outstanding_quantity,
    u_created.username AS created_by_username,
    u_approved.username AS approved_by_username,
    po.created_at
FROM purchase_orders po
JOIN stores s    ON s.id = po.store_id
JOIN suppliers sup ON sup.id = po.supplier_id
LEFT JOIN users u_created  ON u_created.id  = po.created_by
LEFT JOIN users u_approved ON u_approved.id = po.approved_by
WHERE po.status NOT IN ('received','cancelled','closed')
ORDER BY is_overdue DESC, days_overdue DESC NULLS LAST, po.expected_delivery_date;

-- =====================================================
-- FIX #21 (P1): Customer aging report
-- =====================================================

CREATE OR REPLACE VIEW vw_customer_aging_report AS
SELECT
    c.id                    AS customer_id,
    c.customer_code,
    c.name                  AS customer_name,
    c.email,
    c.phone,
    c.customer_type,
    c.credit_limit,
    c.outstanding_balance,
    i.organization_id,
    COALESCE(SUM(CASE WHEN i.due_date >= CURRENT_DATE THEN i.balance_due ELSE 0 END), 0) AS current_amount,
    COALESCE(SUM(CASE WHEN i.due_date < CURRENT_DATE AND CURRENT_DATE - i.due_date <= 30 THEN i.balance_due ELSE 0 END), 0) AS overdue_1_30,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - i.due_date BETWEEN 31 AND 60 THEN i.balance_due ELSE 0 END), 0) AS overdue_31_60,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - i.due_date BETWEEN 61 AND 90 THEN i.balance_due ELSE 0 END), 0) AS overdue_61_90,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - i.due_date > 90 THEN i.balance_due ELSE 0 END), 0) AS overdue_over_90,
    COALESCE(SUM(CASE WHEN i.balance_due > 0 THEN i.balance_due ELSE 0 END), 0) AS total_outstanding,
    COUNT(CASE WHEN i.invoice_status = 'overdue' THEN 1 END)::INTEGER AS overdue_invoice_count,
    MAX(i.due_date)         AS latest_due_date,
    c.loyalty_points
FROM customers c
LEFT JOIN invoices i
    ON i.customer_id = c.id
    AND i.invoice_status NOT IN ('cancelled','draft')
    AND i.balance_due > 0
LEFT JOIN organizations o 
    ON o.id = i.organization_id
WHERE c.is_active = true
GROUP BY 
    c.id, c.customer_code, c.name, c.email, c.phone, 
    c.customer_type, c.credit_limit, c.outstanding_balance, 
    c.loyalty_points, i.organization_id
ORDER BY total_outstanding DESC;

-- =====================================================
-- FIX #22 (P1): Accounts payable overview
-- =====================================================

CREATE OR REPLACE VIEW vw_accounts_payable AS
SELECT
    po.id                       AS po_id,
    po.po_number,
    po.organization_id,
    org.name                    AS organization_name,
    sup.id                      AS supplier_id,
    sup.name                    AS supplier_name,
    sup.contact_person,
    sup.email,
    sup.payment_terms           AS supplier_payment_terms,
    s.name                      AS store_name,
    po.po_date,
    po.expected_delivery_date,
    po.status,
    po.total_amount             AS po_total,
    po.discount_amount,
    po.tax_amount,
    -- Amount outstanding: items received but not yet paid
    (SELECT COALESCE(SUM(pol.received_quantity * pol.unit_price), 0)
     FROM purchase_order_lines pol WHERE pol.purchase_order_id = po.id) AS received_amount,
    po.metadata->>'amount_paid' AS amount_paid_str,
    po.created_at
FROM purchase_orders po
JOIN organizations org ON org.id = po.organization_id
JOIN suppliers     sup ON sup.id = po.supplier_id
JOIN stores        s   ON s.id   = po.store_id
WHERE po.status IN ('partially_received','received','approved')
ORDER BY po.po_date;

-- =====================================================
-- FIX #24 (P2): Profit margin analysis view
-- =====================================================

CREATE OR REPLACE VIEW vw_profit_margin_analysis AS
SELECT
    ptl.product_id,
    ptl.product_variant_id,
    p.sku,
    p.name AS product_name,
    pc.name AS category_name,
    pt.store_id,
    s.name AS store_name,
    DATE_TRUNC('month', pt.transaction_date)::DATE AS period_month,
    SUM(ptl.quantity) AS units_sold,
    SUM(ptl.line_total) AS total_revenue,
    SUM(ptl.cost_price * ptl.quantity) AS total_cost,
    SUM(ptl.line_total) - SUM(ptl.cost_price * ptl.quantity) AS gross_profit,
    CASE
        WHEN SUM(ptl.line_total) > 0
        THEN ROUND(
            (SUM(ptl.line_total) - SUM(ptl.cost_price * ptl.quantity))
            / SUM(ptl.line_total) * 100, 2
        )
        ELSE 0
    END AS gross_margin_pct,
    SUM(ptl.discount_amount) AS total_discounts,
    SUM(ptl.tax_amount) AS total_taxes
FROM pos_transaction_lines ptl
JOIN pos_transactions pt 
    ON pt.id = ptl.transaction_id 
    AND pt.status = 'completed'
JOIN products p 
    ON p.id = ptl.product_id
LEFT JOIN product_variants pv_id 
    ON pv_id.id = ptl.product_variant_id
LEFT JOIN product_categories pc 
    ON pc.id = p.category_id
JOIN stores s 
    ON s.id = pt.store_id
GROUP BY 
    ptl.product_id, 
    ptl.product_variant_id,
    p.sku, 
    p.name, 
    pc.name,
    pt.store_id, 
    s.name, 
    DATE_TRUNC('month', pt.transaction_date)
ORDER BY 
    period_month DESC, 
    gross_profit DESC;

-- =====================================================
-- FIX #25 (P1): Flattened user effective permissions view
-- =====================================================

CREATE OR REPLACE VIEW vw_user_effective_permissions AS
SELECT DISTINCT
    u.id            AS user_id,
    u.username,
    u.email,
    u.organization_id,
    r.id            AS role_id,
    r.name          AS role_name,
    r.code          AS role_code,
    per.id          AS permission_id,
    per.name        AS permission_name,
    per.code        AS permission_code,
    rp.scope,
    usa.store_id    AS accessible_store_id
FROM users u
JOIN user_roles    ur  ON ur.user_id = u.id
JOIN roles         r   ON r.id = ur.role_id AND r.is_active = true
JOIN role_permissions rp ON rp.role_id = r.id
JOIN permissions   per ON per.id = rp.permission_id
LEFT JOIN user_store_access usa ON usa.user_id = u.id
WHERE u.is_active = true
ORDER BY u.username, r.name, per.code;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_pos_get_product_by_barcode(p_barcode VARCHAR, p_store_id INTEGER)
RETURNS TABLE (
    product_id INTEGER,
    sku VARCHAR,
    product_name VARCHAR,
    description TEXT,
    category_name VARCHAR,
    brand_name VARCHAR,
    barcode VARCHAR,
    uom_code VARCHAR,
    decimal_places INTEGER,
    retail_price NUMERIC,
    promo_price NUMERIC,
    effective_price NUMERIC,
    has_promotion BOOLEAN,
    promotion_name VARCHAR,
    promo_min_quantity NUMERIC,
    tax_rate NUMERIC,
    tax_is_inclusive BOOLEAN,
    quantity_available NUMERIC,
    is_in_stock BOOLEAN,
    allow_decimal_quantity BOOLEAN,
    is_serialized BOOLEAN,
    is_batch_managed BOOLEAN,
    product_metadata JSONB,
    package_n_price JSONB,
    product_uom_conversions JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cat.product_id,
        cat.sku::VARCHAR,
        cat.product_name::VARCHAR,
        cat.description,
        cat.category_name::VARCHAR,
        cat.brand_name::VARCHAR,
        cat.barcode::VARCHAR,
        cat.uom_code::VARCHAR,
        (cat.decimal_places)::INTEGER,
        cat.retail_price,
        COALESCE(cat.promo_price, promo_rule.calculated_promo_price) AS promo_price,
        COALESCE(cat.effective_price, promo_rule.calculated_promo_price, cat.retail_price) AS effective_price,
        COALESCE(cat.has_active_promotion, (promo_rule.promo_name IS NOT NULL)) AS has_promotion,
        COALESCE(cat.promotion_name, promo_rule.promo_name)::VARCHAR AS promotion_name,
        COALESCE(cat.promo_min_quantity, promo_rule.promo_min_qty) AS promo_min_quantity,
        cat.tax_rate,
        cat.tax_is_inclusive,
        COALESCE(inv.quantity_available, 0)::NUMERIC,
        (COALESCE(inv.quantity_available, 0) > 0),
        cat.allow_decimal_quantity,
        cat.is_serialized,
        cat.is_batch_managed,
        cat.product_metadata,
        (SELECT COALESCE(jsonb_agg(s.rec ORDER BY s.pl_code, s.uom_code), '[]'::jsonb)
         FROM (
             SELECT 
                 pl.code AS pl_code,
                 uom.code AS uom_code,
                 jsonb_build_object(
                     'product_name', p.name,
                     'price_list_id', pl.id,
                     'price_list_code', pl.code,
                     'price_list_name', pl.name,
                     'price_list', pl.name,
                     'price_list_type', pl.price_list_type,
                     'currency_code', pl.currency_code,
                     'uom_id', uom.id,
                     'uom_code', uom.code,
                     'uom', uom.code,
                     'uom_name', uom.name,
                     'decimal_places', uom.decimal_places,
                     'price', pp.price,
                     'min_quantity', pp.min_quantity,
                     'max_quantity', pp.max_quantity,
                     'valid_from', pp.valid_from,
                     'valid_to', pp.valid_to,
                     'metadata', COALESCE(pp.metadata, '{}'::jsonb),
                     'barcodes', (SELECT COALESCE(jsonb_agg(pb.barcode), '[]'::jsonb) FROM product_barcodes pb WHERE pb.product_id = pp.product_id)
                 ) AS rec
             FROM product_prices pp
             INNER JOIN products p ON pp.product_id = p.id
             INNER JOIN price_lists pl ON pp.price_list_id = pl.id AND pl.is_active = true
             LEFT JOIN units_of_measure uom ON pp.uom_id = uom.id
             WHERE pp.product_id = cat.product_id
               AND pp.is_active = true
               AND (pp.valid_from IS NULL OR pp.valid_from <= CURRENT_DATE)
               AND (pp.valid_to IS NULL OR pp.valid_to >= CURRENT_DATE)
         ) AS s),
        (SELECT COALESCE(jsonb_agg(conv.cv ORDER BY conv.fu_code, conv.tu_code), '[]'::jsonb)
         FROM (
             SELECT fu.code AS fu_code, tu.code AS tu_code,
                    jsonb_build_object(
                        'from_uom_id', fu.id, 'from_uom_code', fu.code, 'from_uom_name', fu.name,
                        'to_uom_id', tu.id, 'to_uom_code', tu.code, 'to_uom_name', tu.name,
                        'conversion_factor', puc.conversion_factor
                    ) AS cv
             FROM product_uom_conversions puc
             JOIN units_of_measure fu ON puc.from_uom_id = fu.id
             JOIN units_of_measure tu ON puc.to_uom_id = tu.id
             WHERE puc.product_id = cat.product_id
         ) AS conv)
    FROM vw_pos_product_catalog cat
    LEFT JOIN LATERAL (
        SELECT 
            pr.name AS promo_name,
            pr.min_quantity AS promo_min_qty,
            pr.discount_value,
            pr.promotion_type,
            CASE 
                WHEN pr.promotion_type = 'percentage_discount' AND cat.retail_price IS NOT NULL THEN
                    ROUND(cat.retail_price * (1.0 - (pr.discount_value / 100.0)), 2)
                WHEN pr.promotion_type = 'fixed_discount' AND cat.retail_price IS NOT NULL THEN
                    GREATEST(0.00, cat.retail_price - pr.discount_value)
                ELSE cat.retail_price
            END AS calculated_promo_price
        FROM promotions pr
        WHERE pr.is_active = true
          AND (pr.valid_from IS NULL OR pr.valid_from <= CURRENT_TIMESTAMP)
          AND (pr.valid_to IS NULL OR pr.valid_to >= CURRENT_TIMESTAMP)
          AND (cardinality(pr.store_ids) = 0 OR p_store_id = ANY(pr.store_ids))
          AND (
              pr.applies_to = 'all'
              OR (pr.applies_to = 'product' AND cat.product_id = ANY(pr.target_product_ids))
              OR (pr.applies_to = 'category' AND cat.category_id = ANY(pr.target_category_ids))
          )
        ORDER BY pr.created_at DESC
        LIMIT 1
    ) promo_rule ON true
    LEFT JOIN inventory_stock inv ON cat.product_id = inv.product_id AND inv.store_id = p_store_id
    WHERE cat.barcode = p_barcode
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_pos_get_products_by_category(
    p_category_id INTEGER,
    p_store_id INTEGER,
    p_include_subcategories BOOLEAN DEFAULT true
)
RETURNS TABLE (
    product_id INTEGER,
    sku VARCHAR,
    product_name VARCHAR,
    category_name VARCHAR,
    brand_name VARCHAR,
    barcode VARCHAR,
    effective_price NUMERIC,
    has_promotion BOOLEAN,
    promotion_name VARCHAR,
    quantity_available NUMERIC,
    is_in_stock BOOLEAN,
    package_n_price JSONB,
    product_uom_conversions JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cat.product_id,
        cat.sku::VARCHAR,
        cat.product_name::VARCHAR,
        cat.category_name::VARCHAR,
        cat.brand_name::VARCHAR,
        cat.barcode::VARCHAR,
        cat.effective_price,
        cat.has_active_promotion,
        cat.promotion_name::VARCHAR,
        COALESCE(inv.quantity_available, 0)::NUMERIC,
        (COALESCE(inv.quantity_available, 0) > 0),
        (SELECT COALESCE(jsonb_agg(s.rec ORDER BY s.pl_code, s.uom_code), '[]'::jsonb)
         FROM (
             SELECT 
                 pl.code AS pl_code,
                 uom.code AS uom_code,
                 jsonb_build_object(
                     'product_name', p.name,
                     'price_list_id', pl.id,
                     'price_list_code', pl.code,
                     'price_list_name', pl.name,
                     'price_list', pl.name,
                     'price_list_type', pl.price_list_type,
                     'currency_code', pl.currency_code,
                     'uom_id', uom.id,
                     'uom_code', uom.code,
                     'uom', uom.code,
                     'uom_name', uom.name,
                     'decimal_places', uom.decimal_places,
                     'price', pp.price,
                     'min_quantity', pp.min_quantity,
                     'max_quantity', pp.max_quantity,
                     'valid_from', pp.valid_from,
                     'valid_to', pp.valid_to,
                     'metadata', COALESCE(pp.metadata, '{}'::jsonb),
                     'barcodes', (SELECT COALESCE(jsonb_agg(pb.barcode), '[]'::jsonb) FROM product_barcodes pb WHERE pb.product_id = pp.product_id)
                 ) AS rec
             FROM product_prices pp
             INNER JOIN products p ON pp.product_id = p.id
             INNER JOIN price_lists pl ON pp.price_list_id = pl.id AND pl.is_active = true
             LEFT JOIN units_of_measure uom ON pp.uom_id = uom.id
             WHERE pp.product_id = cat.product_id
               AND pp.is_active = true
               AND (pp.valid_from IS NULL OR pp.valid_from <= CURRENT_DATE)
               AND (pp.valid_to IS NULL OR pp.valid_to >= CURRENT_DATE)
         ) AS s),
        (SELECT COALESCE(jsonb_agg(conv.cv ORDER BY conv.fu_code, conv.tu_code), '[]'::jsonb)
         FROM (
             SELECT fu.code AS fu_code, tu.code AS tu_code,
                    jsonb_build_object(
                        'from_uom_id', fu.id, 'from_uom_code', fu.code, 'from_uom_name', fu.name,
                        'to_uom_id', tu.id, 'to_uom_code', tu.code, 'to_uom_name', tu.name,
                        'conversion_factor', puc.conversion_factor
                    ) AS cv
             FROM product_uom_conversions puc
             JOIN units_of_measure fu ON puc.from_uom_id = fu.id
             JOIN units_of_measure tu ON puc.to_uom_id = tu.id
             WHERE puc.product_id = cat.product_id
         ) AS conv)
    FROM vw_pos_product_catalog cat
    LEFT JOIN inventory_stock inv ON cat.product_id = inv.product_id AND inv.store_id = p_store_id
    WHERE 
        (cat.category_id = p_category_id 
         OR (p_include_subcategories = true AND cat.parent_category_id = p_category_id))
        AND COALESCE(inv.quantity_available, 0) > 0
    ORDER BY cat.product_name;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_pos_search_products(
    p_search_term VARCHAR,
    p_store_id INTEGER,
    p_limit INTEGER DEFAULT 50
)
RETURNS TABLE (
    product_id INTEGER,
    sku VARCHAR,
    product_name VARCHAR,
    category_name VARCHAR,
    brand_name VARCHAR,
    barcode VARCHAR,
    effective_price NUMERIC,
    has_promotion BOOLEAN,
    quantity_available NUMERIC,
    is_in_stock BOOLEAN,
    relevance_score INTEGER,
    package_n_price JSONB,
    product_uom_conversions JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cat.product_id,
        cat.sku::VARCHAR,
        cat.product_name::VARCHAR,
        cat.category_name::VARCHAR,
        cat.brand_name::VARCHAR,
        cat.barcode::VARCHAR,
        cat.effective_price,
        cat.has_active_promotion,
        COALESCE(inv.quantity_available, 0)::NUMERIC,
        (COALESCE(inv.quantity_available, 0) > 0),
        (CASE 
            WHEN cat.sku ILIKE p_search_term THEN 100
            WHEN cat.product_name ILIKE p_search_term THEN 90
            WHEN cat.barcode = p_search_term THEN 95
            WHEN cat.sku ILIKE p_search_term || '%' THEN 80
            WHEN cat.product_name ILIKE p_search_term || '%' THEN 70
            WHEN cat.sku ILIKE '%' || p_search_term || '%' THEN 60
            WHEN cat.product_name ILIKE '%' || p_search_term || '%' THEN 50
            ELSE 40
        END)::INTEGER,
        (SELECT COALESCE(jsonb_agg(s.rec ORDER BY s.pl_code, s.uom_code), '[]'::jsonb)
         FROM (
             SELECT 
                 pl.code AS pl_code,
                 uom.code AS uom_code,
                 jsonb_build_object(
                     'product_name', p.name,
                     'price_list_id', pl.id,
                     'price_list_code', pl.code,
                     'price_list_name', pl.name,
                     'price_list', pl.name,
                     'price_list_type', pl.price_list_type,
                     'currency_code', pl.currency_code,
                     'uom_id', uom.id,
                     'uom_code', uom.code,
                     'uom', uom.code,
                     'uom_name', uom.name,
                     'decimal_places', uom.decimal_places,
                     'price', pp.price,
                     'min_quantity', pp.min_quantity,
                     'max_quantity', pp.max_quantity,
                     'valid_from', pp.valid_from,
                     'valid_to', pp.valid_to,
                     'metadata', COALESCE(pp.metadata, '{}'::jsonb),
                     'barcodes', (SELECT COALESCE(jsonb_agg(pb.barcode), '[]'::jsonb) FROM product_barcodes pb WHERE pb.product_id = pp.product_id)
                 ) AS rec
             FROM product_prices pp
             INNER JOIN products p ON pp.product_id = p.id
             INNER JOIN price_lists pl ON pp.price_list_id = pl.id AND pl.is_active = true
             LEFT JOIN units_of_measure uom ON pp.uom_id = uom.id
             WHERE pp.product_id = cat.product_id
               AND pp.is_active = true
               AND (pp.valid_from IS NULL OR pp.valid_from <= CURRENT_DATE)
               AND (pp.valid_to IS NULL OR pp.valid_to >= CURRENT_DATE)
         ) AS s),
        (SELECT COALESCE(jsonb_agg(conv.cv ORDER BY conv.fu_code, conv.tu_code), '[]'::jsonb)
         FROM (
             SELECT fu.code AS fu_code, tu.code AS tu_code,
                    jsonb_build_object(
                        'from_uom_id', fu.id, 'from_uom_code', fu.code, 'from_uom_name', fu.name,
                        'to_uom_id', tu.id, 'to_uom_code', tu.code, 'to_uom_name', tu.name,
                        'conversion_factor', puc.conversion_factor
                    ) AS cv
             FROM product_uom_conversions puc
             JOIN units_of_measure fu ON puc.from_uom_id = fu.id
             JOIN units_of_measure tu ON puc.to_uom_id = tu.id
             WHERE puc.product_id = cat.product_id
         ) AS conv)
    FROM vw_pos_product_catalog cat
    LEFT JOIN inventory_stock inv ON cat.product_id = inv.product_id AND inv.store_id = p_store_id
    WHERE 
        cat.product_name ILIKE '%' || p_search_term || '%'
        OR cat.sku ILIKE '%' || p_search_term || '%'
        OR cat.barcode ILIKE '%' || p_search_term || '%'
    ORDER BY 11 DESC, cat.product_name
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE OR REPLACE VIEW vw_pos_categories AS
SELECT
    pc.id               AS category_id,
    pc.code             AS category_code,
    pc.name             AS category_name,
    pc.parent_category_id,
    pc_parent.name      AS parent_category_name,
    COUNT(DISTINCT p.id)::INTEGER AS product_count,
    COUNT(DISTINCT CASE 
        WHEN inv.quantity_available > 0 THEN p.id 
    END)::INTEGER AS in_stock_count,
    pc.metadata         AS category_metadata
FROM product_categories pc
LEFT JOIN product_categories pc_parent 
    ON pc.parent_category_id = pc_parent.id
LEFT JOIN products p 
    ON pc.id = p.category_id
LEFT JOIN inventory_stock inv 
    ON p.id = inv.product_id
GROUP BY 
    pc.id, 
    pc.code, 
    pc.name, 
    pc.parent_category_id, 
    pc_parent.name, 
    pc.metadata
ORDER BY 
    pc_parent.name NULLS FIRST, 
    pc.name;

-- =====================================================
-- RESTAURANT MODULE VIEWS
-- =====================================================

CREATE OR REPLACE VIEW vw_restaurant_menu AS
SELECT
    mi.id                       AS menu_item_id,
    mi.store_id,
    mi.name                     AS item_name,
    mi.short_name,
    mi.description,
    mi.image_url,
    mi.base_price,
    mi.cost_price,
    mi.preparation_time_min,
    mi.is_available,
    mi.is_active,
    mi.display_order,
    mi.metadata                 AS item_metadata,
    mc.id                       AS category_id,
    mc.name                     AS category_name,
    mc.code                     AS category_code,
    mc.parent_category_id,
    mc.display_order            AS category_display_order,
    mc.image_url                AS category_image_url,
    mc_parent.name              AS parent_category_name,
    tc.id                       AS tax_category_id,
    tc.tax_rate,
    tc.is_inclusive             AS tax_is_inclusive,
    mi.recipe_id,
    r.recipe_name,
    r.yield_quantity            AS recipe_yield,
    mi.product_id,
    p.sku                       AS product_sku,
    (SELECT COUNT(*) FROM menu_item_modifiers m WHERE m.menu_item_id = mi.id AND m.is_active = true)::INTEGER
                                AS active_modifier_count,
    CASE
        WHEN mi.base_price > 0 AND mi.cost_price > 0
        THEN ROUND(((mi.base_price - mi.cost_price) / mi.base_price) * 100, 2)
        ELSE NULL
    END                         AS margin_percent
FROM menu_items mi
JOIN menu_categories mc         ON mi.menu_category_id = mc.id
LEFT JOIN menu_categories mc_parent ON mc.parent_category_id = mc_parent.id
LEFT JOIN tax_categories tc     ON mi.tax_category_id = tc.id
LEFT JOIN recipes r             ON mi.recipe_id = r.id
LEFT JOIN products p            ON mi.product_id = p.id
WHERE mi.is_active = true;

CREATE OR REPLACE VIEW vw_recipe_bom AS
SELECT
    r.id                        AS recipe_id,
    r.recipe_code,
    r.recipe_name,
    r.yield_quantity,
    r.organization_id,
    ri.id                       AS ingredient_line_id,
    ri.line_number,
    ri.quantity                 AS ingredient_qty,
    ri.is_optional,
    ri.is_byproduct,
    p.id                        AS product_id,
    p.sku,
    p.name                      AS product_name,
    pv.id                       AS variant_id,
    pv.variant_name,
    uom.id                      AS uom_id,
    uom.code                    AS uom_code,
    uom.name                    AS uom_name,
    pp.price                    AS unit_cost_estimate,
    ROUND(ri.quantity * COALESCE(pp.price, 0), 4) AS line_cost_estimate
FROM recipes r
JOIN recipe_ingredients ri      ON r.id = ri.recipe_id
JOIN products p                 ON ri.product_id = p.id
LEFT JOIN product_variants pv   ON ri.product_variant_id = pv.id
LEFT JOIN units_of_measure uom  ON ri.uom_id = uom.id
LEFT JOIN product_prices pp     ON p.id = pp.product_id
    AND pp.price_list_id        = (SELECT id FROM price_lists WHERE code = 'RETAIL' AND is_active = true LIMIT 1)
    AND pp.is_active            = true
WHERE r.is_active = true;

CREATE OR REPLACE VIEW vw_active_restaurant_orders AS
SELECT
    ro.id                       AS order_id,
    ro.order_number,
    ro.store_id,
    ro.order_source,
    ro.status                   AS order_status,
    ro.subtotal,
    ro.tax_amount,
    ro.total_amount,
    ro.notes,
    ro.ordered_at,
    ro.confirmed_at,
    rt.id                       AS table_id,
    rt.table_number,
    rt.table_name,
    rt.section                  AS table_section,
    c.id                        AS cashier_id,
    u.first_name || ' ' || u.last_name AS waiter_name,
    ro.customer_id,
    cust.name                   AS customer_name,
    EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - ro.ordered_at)) / 60.0 AS minutes_since_ordered
FROM restaurant_orders ro
LEFT JOIN restaurant_tables rt  ON ro.table_id = rt.id
LEFT JOIN cashiers c            ON ro.cashier_id = c.id
LEFT JOIN users u               ON c.user_id = u.id
LEFT JOIN customers cust        ON ro.customer_id = cust.id
WHERE ro.status NOT IN ('paid', 'voided');

CREATE OR REPLACE VIEW vw_waste_daily_summary AS
SELECT
    wl.store_id,
    DATE(wl.wasted_at)          AS waste_date,
    wl.waste_source,
    COUNT(*)                    AS waste_entries,
    SUM(wl.quantity)            AS total_quantity_wasted,
    SUM(wl.total_cost)          AS total_cost_wasted,
    AVG(wl.total_cost)          AS avg_cost_per_entry
FROM waste_logs wl
GROUP BY wl.store_id, DATE(wl.wasted_at), wl.waste_source;

-- =====================================================
-- RESTAURANT MODULE FUNCTIONS
-- =====================================================

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_get_restaurant_menu(
    p_store_id          INTEGER,
    p_category_id       INTEGER  DEFAULT NULL,
    p_include_unavail   BOOLEAN  DEFAULT false
)
RETURNS TABLE (
    menu_item_id            INTEGER,
    item_name               VARCHAR,
    short_name              VARCHAR,
    description             TEXT,
    image_url               TEXT,
    base_price              NUMERIC,
    preparation_time_min    INTEGER,
    is_available            BOOLEAN,
    category_id             INTEGER,
    category_name           VARCHAR,
    parent_category_name    VARCHAR,
    tax_rate                NUMERIC,
    tax_is_inclusive        BOOLEAN,
    recipe_id               INTEGER,
    product_id              INTEGER,
    active_modifier_count   INTEGER,
    margin_percent          NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        vm.menu_item_id,
        vm.item_name::VARCHAR,
        vm.short_name::VARCHAR,
        vm.description,
        vm.image_url,
        vm.base_price,
        vm.preparation_time_min,
        vm.is_available,
        vm.category_id,
        vm.category_name::VARCHAR,
        vm.parent_category_name::VARCHAR,
        vm.tax_rate,
        vm.tax_is_inclusive,
        vm.recipe_id,
        vm.product_id,
        vm.active_modifier_count,
        vm.margin_percent
    FROM vw_restaurant_menu vm
    WHERE vm.store_id = p_store_id
      AND (p_category_id IS NULL OR vm.category_id = p_category_id)
      AND (p_include_unavail = true OR vm.is_available = true)
    ORDER BY vm.category_display_order, vm.display_order;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_get_item_modifiers(
    p_menu_item_id INTEGER
)
RETURNS TABLE (
    modifier_id         INTEGER,
    modifier_name       VARCHAR,
    modifier_type       VARCHAR,
    price_adjustment    NUMERIC,
    display_order       INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        m.id,
        m.modifier_name::VARCHAR,
        m.modifier_type::VARCHAR,
        m.price_adjustment,
        m.display_order
    FROM menu_item_modifiers m
    WHERE m.menu_item_id = p_menu_item_id
      AND m.is_active    = true
    ORDER BY m.display_order;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- =====================================================
-- FIX #2 (P0): Atomic inter-store / warehouse stock transfer function
-- =====================================================

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_process_stock_transfer(
    p_from_store_id     INTEGER,
    p_to_store_id       INTEGER,
    p_product_id        INTEGER,
    p_product_variant_id INTEGER,
    p_quantity          DECIMAL(15,3),
    p_from_location_id  INTEGER DEFAULT NULL,
    p_to_location_id    INTEGER DEFAULT NULL,
    p_batch_number      VARCHAR DEFAULT NULL,
    p_performed_by      INTEGER DEFAULT NULL,
    p_notes             TEXT    DEFAULT NULL
)
RETURNS TABLE (
    success          BOOLEAN,
    message          TEXT,
    movement_id      INTEGER
) AS $$
DECLARE
    v_available  DECIMAL(15,3);
    v_movement_id INTEGER;
    v_ref_num    VARCHAR(50);
BEGIN
    -- Validate same store check
    IF p_from_store_id = p_to_store_id THEN
        RETURN QUERY SELECT false, 'Source and destination stores must differ.', NULL::INTEGER;
        RETURN;
    END IF;

    -- Lock and read available stock at source
    SELECT quantity_available INTO v_available
    FROM inventory_stock
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_from_store_id
    FOR UPDATE;

    IF v_available IS NULL OR v_available < p_quantity THEN
        RETURN QUERY SELECT false,
            format('Insufficient stock. Available: %s, Requested: %s', COALESCE(v_available, 0), p_quantity),
            NULL::INTEGER;
        RETURN;
    END IF;

    v_ref_num := 'TRF-' || to_char(CURRENT_TIMESTAMP, 'YYYYMMDDHH24MISS') || '-' || p_from_store_id || '-' || p_to_store_id;

    -- Deduct from source
    UPDATE inventory_stock
    SET quantity_on_hand   = quantity_on_hand   - p_quantity,
        quantity_available = quantity_available - p_quantity,
        quantity_in_transit = quantity_in_transit + p_quantity,
        updated_at = CURRENT_TIMESTAMP
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_from_store_id;

    -- Add to destination
    INSERT INTO inventory_stock (product_id, product_variant_id, store_id, storage_location_id,
        quantity_on_hand, quantity_available, quantity_in_transit)
    VALUES (p_product_id, p_product_variant_id, p_to_store_id, p_to_location_id,
            p_quantity, p_quantity, 0)
    ON CONFLICT (product_id, COALESCE(product_variant_id, -1), store_id)
    DO UPDATE SET
        quantity_on_hand   = inventory_stock.quantity_on_hand   + EXCLUDED.quantity_on_hand,
        quantity_available = inventory_stock.quantity_available + EXCLUDED.quantity_available,
        updated_at = CURRENT_TIMESTAMP;

    -- Clear in-transit at source
    UPDATE inventory_stock
    SET quantity_in_transit = GREATEST(0, quantity_in_transit - p_quantity),
        updated_at = CURRENT_TIMESTAMP
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_from_store_id;

    -- Record movement
    INSERT INTO stock_movements (movement_type, reference_type, product_id, product_variant_id,
        from_store_id, to_store_id, from_location_id, to_location_id,
        quantity, batch_number, posted_by, status,
        metadata)
    VALUES ('transfer', 'manual', p_product_id, p_product_variant_id,
            p_from_store_id, p_to_store_id, p_from_location_id, p_to_location_id,
            p_quantity, p_batch_number, p_performed_by, 'completed',
            jsonb_build_object('reference_number', v_ref_num, 'notes', p_notes))
    RETURNING id INTO v_movement_id;

    RETURN QUERY SELECT true, 'Transfer completed successfully. Ref: ' || v_ref_num, v_movement_id;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- =====================================================
-- LOGISTICS & IN-TRANSIT WORKFLOW FUNCTIONS
-- =====================================================

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_approve_transfer_request(
    p_transfer_request_id INTEGER,
    p_approved_by INTEGER
)
RETURNS TABLE (
    success BOOLEAN,
    message TEXT
) AS $$
DECLARE
    v_req RECORD;
BEGIN
    SELECT * INTO v_req FROM transfer_requests WHERE id = p_transfer_request_id FOR UPDATE;
    IF v_req IS NULL THEN
        RETURN QUERY SELECT false, 'Transfer request not found.';
        RETURN;
    END IF;

    IF v_req.status NOT IN ('draft', 'pending_approval') THEN
        RETURN QUERY SELECT false, 'Transfer request can only be approved from draft or pending_approval state.';
        RETURN;
    END IF;

    UPDATE transfer_requests
    SET status = 'approved',
        approved_by = p_approved_by,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_transfer_request_id;

    RETURN QUERY SELECT true, 'Transfer request approved successfully.';
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_ship_transfer_request(
    p_transfer_request_id INTEGER,
    p_shipped_by INTEGER
)
RETURNS TABLE (
    success BOOLEAN,
    message TEXT
) AS $$
DECLARE
    v_req RECORD;
    v_item RECORD;
    v_available DECIMAL(15,3);
    v_qty DECIMAL(15,3);
BEGIN
    SELECT * INTO v_req FROM transfer_requests WHERE id = p_transfer_request_id FOR UPDATE;
    IF v_req IS NULL THEN
        RETURN QUERY SELECT false, 'Transfer request not found.';
        RETURN;
    END IF;

    IF v_req.status NOT IN ('approved', 'pending_approval', 'draft') THEN
        RETURN QUERY SELECT false, 'Transfer request must be approved to ship.';
        RETURN;
    END IF;

    FOR v_item IN SELECT * FROM transfer_request_items WHERE transfer_request_id = p_transfer_request_id FOR UPDATE LOOP
        v_qty := CASE WHEN v_item.approved_quantity > 0 THEN v_item.approved_quantity ELSE v_item.requested_quantity END;
        IF v_qty <= 0 THEN
            CONTINUE;
        END IF;

        -- Check available stock at source
        SELECT quantity_available INTO v_available
        FROM inventory_stock
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_req.from_store_id
        FOR UPDATE;

        IF v_available IS NULL OR v_available < v_qty THEN
            RETURN QUERY SELECT false, format('Insufficient stock for product ID %s at source store.', v_item.product_id);
            RETURN;
        END IF;

        -- Deduct from source store
        UPDATE inventory_stock
        SET quantity_on_hand = quantity_on_hand - v_qty,
            quantity_available = quantity_available - v_qty,
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_req.from_store_id;

        -- Increment quantity_in_transit at destination store
        INSERT INTO inventory_stock (product_id, product_variant_id, store_id, storage_location_id,
            quantity_on_hand, quantity_available, quantity_in_transit)
        VALUES (v_item.product_id, v_item.product_variant_id, v_req.to_store_id, v_item.to_location_id,
                0, 0, v_qty)
        ON CONFLICT (product_id, COALESCE(product_variant_id, -1), store_id)
        DO UPDATE SET
            quantity_in_transit = inventory_stock.quantity_in_transit + EXCLUDED.quantity_in_transit,
            updated_at = CURRENT_TIMESTAMP;

        -- Update item shipped_quantity
        UPDATE transfer_request_items
        SET shipped_quantity = v_qty,
            approved_quantity = v_qty
        WHERE id = v_item.id;

        -- Record stock movement (transfer_out / shipped)
        INSERT INTO stock_movements (
            movement_type, reference_type, reference_id, product_id, product_variant_id,
            from_store_id, to_store_id, from_location_id, to_location_id,
            quantity, uom_id, batch_number, posted_by, status, metadata
        ) VALUES (
            'transfer_out', 'transfer_request', p_transfer_request_id, v_item.product_id, v_item.product_variant_id,
            v_req.from_store_id, v_req.to_store_id, v_item.from_location_id, v_item.to_location_id,
            v_qty, v_item.uom_id, v_item.batch_number, p_shipped_by, 'shipped',
            jsonb_build_object('transfer_number', v_req.transfer_number)
        );
    END LOOP;

    -- Update header
    UPDATE transfer_requests
    SET status = 'shipped',
        shipped_by = p_shipped_by,
        shipped_at = CURRENT_TIMESTAMP,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_transfer_request_id;

    RETURN QUERY SELECT true, 'Transfer request shipped successfully.';
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_receive_transfer_request(
    p_transfer_request_id INTEGER,
    p_received_by INTEGER
)
RETURNS TABLE (
    success BOOLEAN,
    message TEXT
) AS $$
DECLARE
    v_req RECORD;
    v_item RECORD;
    v_qty DECIMAL(15,3);
BEGIN
    SELECT * INTO v_req FROM transfer_requests WHERE id = p_transfer_request_id FOR UPDATE;
    IF v_req IS NULL THEN
        RETURN QUERY SELECT false, 'Transfer request not found.';
        RETURN;
    END IF;

    IF v_req.status NOT IN ('shipped', 'partially_received') THEN
        RETURN QUERY SELECT false, 'Transfer request must be shipped or partially_received to receive.';
        RETURN;
    END IF;

    FOR v_item IN SELECT * FROM transfer_request_items WHERE transfer_request_id = p_transfer_request_id FOR UPDATE LOOP
        v_qty := CASE WHEN v_item.shipped_quantity > 0 THEN v_item.shipped_quantity ELSE v_item.requested_quantity END;
        IF v_qty <= 0 THEN
            CONTINUE;
        END IF;

        -- Decrement in-transit and increment on_hand & available at destination store
        UPDATE inventory_stock
        SET quantity_in_transit = GREATEST(0, quantity_in_transit - v_qty),
            quantity_on_hand = quantity_on_hand + v_qty,
            quantity_available = quantity_available + v_qty,
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_req.to_store_id;

        -- Update item received_quantity
        UPDATE transfer_request_items
        SET received_quantity = v_qty
        WHERE id = v_item.id;

        -- Record stock movement (transfer_in / completed)
        INSERT INTO stock_movements (
            movement_type, reference_type, reference_id, product_id, product_variant_id,
            from_store_id, to_store_id, from_location_id, to_location_id,
            quantity, uom_id, batch_number, posted_by, status, metadata
        ) VALUES (
            'transfer_in', 'transfer_request', p_transfer_request_id, v_item.product_id, v_item.product_variant_id,
            v_req.from_store_id, v_req.to_store_id, v_item.from_location_id, v_item.to_location_id,
            v_qty, v_item.uom_id, v_item.batch_number, p_received_by, 'completed',
            jsonb_build_object('transfer_number', v_req.transfer_number)
        );
    END LOOP;

    -- Update header
    UPDATE transfer_requests
    SET status = 'received',
        received_by = p_received_by,
        received_at = CURRENT_TIMESTAMP,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_transfer_request_id;

    RETURN QUERY SELECT true, 'Transfer request received successfully.';
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_process_goods_receipt(
    p_grn_id INTEGER
)
RETURNS TABLE (
    success BOOLEAN,
    message TEXT
) AS $$
DECLARE
    v_grn RECORD;
    v_item RECORD;
    v_all_po_received BOOLEAN := true;
    v_any_po_received BOOLEAN := false;
    v_po_line RECORD;
BEGIN
    SELECT * INTO v_grn FROM goods_receipt_notes WHERE id = p_grn_id FOR UPDATE;
    IF v_grn IS NULL THEN
        RETURN QUERY SELECT false, 'Goods receipt note not found.';
        RETURN;
    END IF;

    IF v_grn.status = 'completed' THEN
        RETURN QUERY SELECT false, 'Goods receipt note is already completed.';
        RETURN;
    END IF;

    FOR v_item IN SELECT * FROM goods_receipt_note_items WHERE grn_id = p_grn_id FOR UPDATE LOOP
        IF v_item.quantity_received <= 0 THEN
            CONTINUE;
        END IF;

        -- Update PO Line if associated
        IF v_item.purchase_order_line_id IS NOT NULL THEN
            UPDATE purchase_order_lines
            SET received_quantity = received_quantity + v_item.quantity_received
            WHERE id = v_item.purchase_order_line_id;
        END IF;

        -- Increment inventory stock at store
        INSERT INTO inventory_stock (
            product_id, product_variant_id, store_id, storage_location_id,
            quantity_on_hand, quantity_available
        ) VALUES (
            v_item.product_id, v_item.product_variant_id, v_grn.store_id, v_item.storage_location_id,
            v_item.quantity_received, v_item.quantity_received
        )
        ON CONFLICT (product_id, COALESCE(product_variant_id, -1), store_id)
        DO UPDATE SET
            quantity_on_hand = inventory_stock.quantity_on_hand + EXCLUDED.quantity_on_hand,
            quantity_available = inventory_stock.quantity_available + EXCLUDED.quantity_available,
            updated_at = CURRENT_TIMESTAMP;

        -- Insert stock movement (purchase_receipt)
        INSERT INTO stock_movements (
            movement_type, reference_type, reference_id, product_id, product_variant_id,
            to_store_id, to_location_id, quantity, uom_id, batch_number,
            posted_by, status, cost_per_unit, total_value, metadata
        ) VALUES (
            'purchase_receipt', 'goods_receipt_note', p_grn_id, v_item.product_id, v_item.product_variant_id,
            v_grn.store_id, v_item.storage_location_id, v_item.quantity_received, v_item.uom_id, COALESCE(v_item.batch_number, ''),
            v_grn.received_by, 'completed', COALESCE(v_item.unit_cost, 0), (COALESCE(v_item.unit_cost, 0) * v_item.quantity_received),
            jsonb_build_object('grn_number', v_grn.grn_number, 'delivery_note_number', COALESCE(v_grn.delivery_note_number, ''))
        );
    END LOOP;

    -- Update GRN status
    UPDATE goods_receipt_notes
    SET status = 'completed',
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_grn_id;

    -- Update Purchase Order status if PO ID present
    IF v_grn.purchase_order_id IS NOT NULL THEN
        FOR v_po_line IN SELECT quantity, received_quantity FROM purchase_order_lines WHERE purchase_order_id = v_grn.purchase_order_id LOOP
            IF v_po_line.received_quantity < v_po_line.quantity THEN
                v_all_po_received := false;
            END IF;
            IF v_po_line.received_quantity > 0 THEN
                v_any_po_received := true;
            END IF;
        END LOOP;

        IF v_all_po_received THEN
            UPDATE purchase_orders SET status = 'received', updated_at = CURRENT_TIMESTAMP WHERE id = v_grn.purchase_order_id;
        ELSIF v_any_po_received THEN
            UPDATE purchase_orders SET status = 'partially_received', updated_at = CURRENT_TIMESTAMP WHERE id = v_grn.purchase_order_id;
        END IF;
    END IF;

    RETURN QUERY SELECT true, 'Goods receipt processed successfully.';
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- =====================================================
-- FIX #3 (P0): Auto stock adjustment after physical count
-- =====================================================

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_reconcile_stock_count(p_count_id INTEGER)
RETURNS TABLE (
    success       BOOLEAN,
    message       TEXT,
    lines_updated INTEGER
) AS $$
DECLARE
    v_count        RECORD;
    v_line         RECORD;
    v_lines_updated INTEGER := 0;
BEGIN
    SELECT * INTO v_count FROM stock_counts WHERE id = p_count_id;

    IF NOT FOUND THEN
        RETURN QUERY SELECT false, 'Stock count not found.', 0;
        RETURN;
    END IF;
    IF v_count.status <> 'approved' THEN
        RETURN QUERY SELECT false, 'Stock count must be in ''approved'' status to reconcile.', 0;
        RETURN;
    END IF;

    FOR v_line IN
        SELECT * FROM stock_count_lines WHERE stock_count_id = p_count_id
    LOOP
        -- Update inventory_stock to match counted quantity
        UPDATE inventory_stock
        SET quantity_on_hand   = v_line.counted_quantity,
            quantity_available = GREATEST(0, v_line.counted_quantity - COALESCE(quantity_allocated, 0)),
            last_counted_at    = CURRENT_TIMESTAMP,
            updated_at         = CURRENT_TIMESTAMP
        WHERE product_id = v_line.product_id
          AND (product_variant_id = v_line.product_variant_id
               OR (product_variant_id IS NULL AND v_line.product_variant_id IS NULL))
          AND store_id = v_count.store_id;

        -- Record the adjustment movement
        INSERT INTO stock_movements (movement_type, reference_type, reference_id, product_id,
            product_variant_id, to_store_id, quantity, batch_number, serial_number,
            posted_by, status, metadata)
        VALUES ('count_adjustment', 'stock_count', p_count_id, v_line.product_id,
                v_line.product_variant_id, v_count.store_id,
                v_line.counted_quantity - v_line.system_quantity,
                v_line.batch_number, v_line.serial_number,
                v_count.approved_by, 'completed',
                jsonb_build_object('count_id', p_count_id, 'variance', v_line.variance));

        v_lines_updated := v_lines_updated + 1;
    END LOOP;

    -- Mark count as reconciled
    UPDATE stock_counts SET status = 'approved', updated_at = CURRENT_TIMESTAMP
    WHERE id = p_count_id;

    RETURN QUERY SELECT true, format('Reconciliation complete. %s lines adjusted.', v_lines_updated), v_lines_updated;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- =====================================================
-- FIX #27 (P0): Auto stock deduction on order fulfillment
-- =====================================================

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_trigger_deduct_inventory_on_fulfillment()
RETURNS TRIGGER AS $$
DECLARE
    v_order_line RECORD;
    v_fulfilled_qty DECIMAL(15,3);
    v_reservation RECORD;
BEGIN
    -- Only process when order status changes to 'fulfilled'
    IF NOT (OLD.order_status IS DISTINCT FROM NEW.order_status AND NEW.order_status = 'fulfilled') THEN
        RETURN NEW;
    END IF;
        
    -- Ensure store_id is present
    IF NEW.store_id IS NULL THEN
        RAISE WARNING 'Order % has no store_id, skipping inventory deduction', NEW.id;
        RETURN NEW;
    END IF;

    -- Loop through all order lines
    FOR v_order_line IN
        SELECT 
            id,
            product_id,
            product_variant_id,
            quantity_ordered,
            quantity_fulfilled,
            uom_id
        FROM sales_order_lines_v2
        WHERE sales_order_id = NEW.id
    LOOP
        -- Determine the fulfilled quantity
        IF v_order_line.quantity_fulfilled IS NOT NULL AND v_order_line.quantity_fulfilled > 0 THEN
            v_fulfilled_qty := v_order_line.quantity_fulfilled;
        ELSE
            v_fulfilled_qty := v_order_line.quantity_ordered;
        END IF;
        
        IF v_fulfilled_qty <= 0 THEN
            CONTINUE;
        END IF;

        -- FULFILLMENT: Deduct from on-hand and reduce allocated
        UPDATE inventory_stock
        SET 
            quantity_on_hand = quantity_on_hand - v_fulfilled_qty,
            quantity_allocated = GREATEST(0, quantity_allocated - v_fulfilled_qty),
            quantity_available = GREATEST(0, 
                (quantity_on_hand - v_fulfilled_qty) - 
                GREATEST(0, quantity_allocated - v_fulfilled_qty)
            ),
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_order_line.product_id
          AND (product_variant_id = v_order_line.product_variant_id 
               OR (product_variant_id IS NULL AND v_order_line.product_variant_id IS NULL))
          AND store_id = NEW.store_id;

        IF NOT FOUND THEN
            RAISE WARNING 'No inventory_stock record found for product_id=%, product_variant_id=%, store_id=%. Movement recorded but stock not updated.',
                v_order_line.product_id, v_order_line.product_variant_id, NEW.store_id;
        END IF;

        -- Record the stock movement for auditing
        INSERT INTO stock_movements (
            movement_type,
            reference_type,
            reference_id,
            product_id,
            product_variant_id,
            from_store_id,
            quantity,
            uom_id,
            status,
            metadata
        )
        VALUES (
            'sale',
            'sales_order',
            NULL,
            v_order_line.product_id,
            v_order_line.product_variant_id,
            NEW.store_id,
            v_fulfilled_qty,
            v_order_line.uom_id,
            'completed',
            jsonb_build_object(
                'sales_order_id', NEW.id::TEXT,
                'sales_order_number', NEW.order_number,
                'order_line_id', v_order_line.id::TEXT,
                'order_status', NEW.order_status
            )
        );

        -- Mark active reservations as 'fulfilled' when order is fulfilled
        FOR v_reservation IN
            SELECT id, quantity_reserved
            FROM stock_reservations
            WHERE reference_type = 'sales_order'
              AND reference_id = NEW.id::TEXT
              AND product_id = v_order_line.product_id
              AND (product_variant_id = v_order_line.product_variant_id 
                   OR (product_variant_id IS NULL AND v_order_line.product_variant_id IS NULL))
              AND store_id = NEW.store_id
              AND status = 'active'
        LOOP
            UPDATE stock_reservations
            SET 
                status = 'fulfilled',
                updated_at = CURRENT_TIMESTAMP
            WHERE id = v_reservation.id;
        END LOOP;

    END LOOP;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
-- Trigger for UPDATE: fires when order status changes to 'fulfilled'
CREATE TRIGGER trg_deduct_inventory_on_fulfillment
    AFTER UPDATE ON sales_orders_v2
    FOR EACH ROW
    WHEN (OLD.order_status IS DISTINCT FROM NEW.order_status AND NEW.order_status = 'fulfilled')
    EXECUTE FUNCTION fn_trigger_deduct_inventory_on_fulfillment();
-- +goose StatementEnd

-- =====================================================
-- Trigger for order line insertion: allocate stock when order is pending/confirmed
-- =====================================================

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_trigger_allocate_inventory_on_order_line()
RETURNS TRIGGER AS $$
DECLARE
    v_order RECORD;
    v_quantity DECIMAL(15,3);
BEGIN
    -- Get the parent order
    SELECT order_status, store_id INTO v_order
    FROM sales_orders_v2
    WHERE id = NEW.sales_order_id;

    -- Only process if order status is 'pending' or 'confirmed'
    IF v_order.order_status IN ('pending', 'confirmed') AND v_order.store_id IS NOT NULL THEN
        v_quantity := NEW.quantity_ordered;
        
        IF v_quantity > 0 THEN
            -- Allocate stock (increase allocated, decrease available)
            UPDATE inventory_stock
            SET 
                quantity_allocated = quantity_allocated + v_quantity,
                quantity_available = GREATEST(0, quantity_on_hand - (quantity_allocated + v_quantity)),
                updated_at = CURRENT_TIMESTAMP
            WHERE product_id = NEW.product_id
              AND (product_variant_id = NEW.product_variant_id 
                   OR (product_variant_id IS NULL AND NEW.product_variant_id IS NULL))
              AND store_id = v_order.store_id;

            -- Record the allocation movement
            INSERT INTO stock_movements (
                movement_type,
                reference_type,
                reference_id,
                product_id,
                product_variant_id,
                from_store_id,
                quantity,
                uom_id,
                status,
                metadata
            )
            VALUES (
                'allocation',
                'sales_order',
                NULL,
                NEW.product_id,
                NEW.product_variant_id,
                v_order.store_id,
                v_quantity,
                NEW.uom_id,
                'completed',
                jsonb_build_object(
                    'sales_order_id', NEW.sales_order_id::TEXT,
                    'order_line_id', NEW.id::TEXT,
                    'order_status', v_order.order_status
                )
            );
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_allocate_inventory_on_order_line_insert
    AFTER INSERT ON sales_order_lines_v2
    FOR EACH ROW
    EXECUTE FUNCTION fn_trigger_allocate_inventory_on_order_line();
-- +goose StatementEnd

-- =====================================================
-- FIX #26 (P1): Loyalty points earning calculation
-- =====================================================

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_calculate_loyalty_earned(p_transaction_id INTEGER)
RETURNS TABLE (
    points_earned  DECIMAL(15,2),
    rule_applied   VARCHAR(255),
    customer_id    INTEGER
) AS $$
DECLARE
    v_txn        RECORD;
    v_rule       RECORD;
    v_points     DECIMAL(15,2) := 0;
BEGIN
    SELECT pt.*, pt.total_amount, pt.customer_id AS cust_id
    INTO v_txn
    FROM pos_transactions pt
    WHERE pt.id = p_transaction_id;

    IF NOT FOUND OR v_txn.cust_id IS NULL THEN
        RETURN QUERY SELECT 0::DECIMAL(15,2), 'No customer on transaction'::VARCHAR(255), NULL::INTEGER;
        RETURN;
    END IF;

    SELECT * INTO v_rule
    FROM loyalty_redemption_rules
    WHERE is_active = true
      AND (valid_from IS NULL OR valid_from <= CURRENT_DATE)
      AND (valid_to   IS NULL OR valid_to   >= CURRENT_DATE)
    ORDER BY id DESC LIMIT 1;

    IF NOT FOUND THEN
        RETURN QUERY SELECT 0::DECIMAL(15,2), 'No active loyalty rule found'::VARCHAR(255), v_txn.cust_id;
        RETURN;
    END IF;

    v_points := FLOOR(v_txn.total_amount * v_rule.points_earning_rate);

    -- Update customer loyalty_points balance
    UPDATE customers
    SET loyalty_points = loyalty_points + v_points,
        updated_at     = CURRENT_TIMESTAMP
    WHERE id = v_txn.cust_id;

    RETURN QUERY SELECT v_points, v_rule.rule_name::VARCHAR(255), v_txn.cust_id;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- =====================================================
-- FIX #23 (P1): Daily analytics refresh function
-- =====================================================

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_refresh_daily_analytics(p_date DATE DEFAULT CURRENT_DATE)
RETURNS VOID AS $$
BEGIN
    -- Refresh sales_analytics from pos_transactions and sales_orders_v2
    INSERT INTO sales_analytics (
        organization_id, store_id, product_id, category_id,
        date, hour, day_of_week, month, quarter, year,
        units_sold, revenue, discounts, taxes, net_revenue, transactions,
        payment_method, payment_gateway
    )
    SELECT
        organization_id, store_id, product_id, category_id,
        date, hour, day_of_week, month, quarter, year,
        SUM(units_sold), SUM(revenue), SUM(discounts), SUM(taxes), SUM(net_revenue), SUM(transactions),
        payment_method, payment_gateway
    FROM (
        -- Data from POS transactions
        SELECT
            s.organization_id,
            pt.store_id,
            ptl.product_id,
            p.category_id,
            p_date AS date,
            EXTRACT(HOUR FROM pt.transaction_date)::INTEGER AS hour,
            EXTRACT(DOW  FROM pt.transaction_date)::INTEGER AS day_of_week,
            EXTRACT(MONTH FROM p_date)::INTEGER AS month,
            EXTRACT(QUARTER FROM p_date)::INTEGER AS quarter,
            EXTRACT(YEAR FROM p_date)::INTEGER AS year,
            ptl.quantity AS units_sold,
            ptl.line_total AS revenue,
            ptl.discount_amount AS discounts,
            ptl.tax_amount AS taxes,
            (ptl.line_total - ptl.tax_amount) AS net_revenue,
            1 AS transactions,
            pp.payment_method,
            pp.payment_gateway
        FROM pos_transactions pt
        JOIN pos_transaction_lines ptl ON ptl.transaction_id = pt.id
        LEFT JOIN pos_payments pp ON pp.transaction_id = pt.id
        JOIN products p ON p.id = ptl.product_id
        JOIN stores s ON s.id = pt.store_id
        WHERE DATE(pt.transaction_date) = p_date
          AND pt.status = 'completed'

        UNION ALL

        -- Data from sales_orders_v2 (excluding those that were synced to POS to avoid double counting)
        SELECT
            o.organization_id,
            o.store_id,
            ol.product_id,
            p.category_id,
            p_date AS date,
            EXTRACT(HOUR FROM o.order_date)::INTEGER AS hour,
            EXTRACT(DOW  FROM o.order_date)::INTEGER AS day_of_week,
            EXTRACT(MONTH FROM p_date)::INTEGER AS month,
            EXTRACT(QUARTER FROM p_date)::INTEGER AS quarter,
            EXTRACT(YEAR FROM p_date)::INTEGER AS year,
            ol.quantity_ordered::DECIMAL(15,3) AS units_sold,
            ol.line_total AS revenue,
            COALESCE(ol.discount_amount, 0) AS discounts,
            COALESCE(ol.tax_amount, 0) AS taxes,
            (ol.line_total - COALESCE(ol.tax_amount, 0)) AS net_revenue,
            1 AS transactions,
            o.payment_method,
            o.payment_gateway
        FROM sales_orders_v2 o
        JOIN sales_order_lines_v2 ol ON ol.sales_order_id = o.id
        JOIN products p ON p.id = ol.product_id
        WHERE DATE(o.order_date) = p_date
          AND o.order_status IN ('confirmed', 'processing', 'partially_fulfilled', 'fulfilled', 'shipped', 'delivered')
          AND NOT EXISTS (SELECT 1 FROM pos_transactions WHERE sales_order_id = o.id)
    ) aggregated_sales
    GROUP BY organization_id, store_id, product_id, category_id,
             date, hour, day_of_week, month, quarter, year,
             payment_method, payment_gateway
    ON CONFLICT (organization_id, store_id, product_id, date, hour, payment_method, payment_gateway) 
    DO UPDATE SET
        units_sold = EXCLUDED.units_sold,
        revenue = EXCLUDED.revenue,
        discounts = EXCLUDED.discounts,
        taxes = EXCLUDED.taxes,
        net_revenue = EXCLUDED.net_revenue,
        transactions = EXCLUDED.transactions,
        updated_at = CURRENT_TIMESTAMP;

    -- Refresh profit_loss_analytics
    INSERT INTO profit_loss_analytics (
        organization_id, store_id, date, period_type, month, quarter, year,
        gross_revenue, sales_discounts, net_revenue, cogs
    )
    SELECT
        s.organization_id,
        pt.store_id,
        p_date,
        'daily',
        EXTRACT(MONTH FROM p_date)::INTEGER,
        EXTRACT(QUARTER FROM p_date)::INTEGER,
        EXTRACT(YEAR FROM p_date)::INTEGER,
        SUM(pt.total_amount),
        SUM(pt.discount_amount),
        SUM(pt.total_amount - pt.discount_amount),
        SUM(pt.total_cost)
    FROM pos_transactions pt
    JOIN stores s ON s.id = pt.store_id
    WHERE DATE(pt.transaction_date) = p_date
      AND pt.status = 'completed'
    GROUP BY s.organization_id, pt.store_id
    ON CONFLICT DO NOTHING;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_calculate_recipe_cost(
    p_recipe_id INTEGER
)
RETURNS NUMERIC AS $$
DECLARE
    v_total_cost NUMERIC := 0;
BEGIN
    SELECT COALESCE(SUM(vb.line_cost_estimate), 0)
      INTO v_total_cost
    FROM vw_recipe_bom vb
    WHERE vb.recipe_id    = p_recipe_id
      AND vb.is_byproduct = false
      AND vb.is_optional  = false;

    RETURN v_total_cost;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_get_waste_report(
    p_store_id      INTEGER,
    p_from_date     DATE,
    p_to_date       DATE,
    p_waste_source  VARCHAR DEFAULT NULL
)
RETURNS TABLE (
    waste_date          DATE,
    waste_source        VARCHAR,
    product_id          INTEGER,
    product_name        VARCHAR,
    menu_item_id        INTEGER,
    menu_item_name      VARCHAR,
    quantity            NUMERIC,
    uom_code            VARCHAR,
    total_cost          NUMERIC,
    reason              TEXT,
    logged_by_name      VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        DATE(wl.wasted_at),
        wl.waste_source::VARCHAR,
        wl.product_id,
        p.name::VARCHAR,
        wl.menu_item_id,
        mi.name::VARCHAR,
        wl.quantity,
        uom.code::VARCHAR,
        wl.total_cost,
        wl.reason,
        (u.first_name || ' ' || u.last_name)::VARCHAR
    FROM waste_logs wl
    LEFT JOIN products p            ON wl.product_id   = p.id
    LEFT JOIN menu_items mi         ON wl.menu_item_id = mi.id
    LEFT JOIN units_of_measure uom  ON wl.uom_id       = uom.id
    LEFT JOIN users u               ON wl.logged_by    = u.id
    WHERE wl.store_id           = p_store_id
      AND DATE(wl.wasted_at)    BETWEEN p_from_date AND p_to_date
      AND (p_waste_source IS NULL OR wl.waste_source = p_waste_source)
    ORDER BY wl.wasted_at DESC;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_get_kds_orders(
    p_store_id      INTEGER,
    p_statuses      VARCHAR[] DEFAULT ARRAY['pending','confirmed','preparing']
)
RETURNS TABLE (
    order_id            INTEGER,
    order_number        VARCHAR,
    table_number        VARCHAR,
    waiter_name         VARCHAR,
    order_status        VARCHAR,
    ordered_at          TIMESTAMP,
    minutes_elapsed     NUMERIC,
    item_id             INTEGER,
    item_name           VARCHAR,
    item_short_name     VARCHAR,
    item_qty            NUMERIC,
    item_notes          TEXT,
    item_modifiers      JSONB,
    item_status         VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        ro.id,
        ro.order_number::VARCHAR,
        rt.table_number::VARCHAR,
        (u.first_name || ' ' || u.last_name)::VARCHAR,
        ro.status::VARCHAR,
        ro.ordered_at,
        ROUND(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - ro.ordered_at)) / 60.0, 1),
        roi.id,
        mi.name::VARCHAR,
        mi.short_name::VARCHAR,
        roi.quantity,
        roi.notes,
        roi.modifiers_snapshot,
        roi.status::VARCHAR
    FROM restaurant_orders ro
    LEFT JOIN restaurant_tables rt      ON ro.table_id = rt.id
    LEFT JOIN cashiers c                ON ro.cashier_id = c.id
    LEFT JOIN users u                   ON c.user_id = u.id
    JOIN  restaurant_order_items roi    ON ro.id = roi.order_id
    JOIN  menu_items mi                 ON roi.menu_item_id = mi.id
    WHERE ro.store_id = p_store_id
      AND ro.status = ANY(p_statuses)
    ORDER BY ro.ordered_at, roi.line_number;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


-- =====================================================
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


-- +goose Down

-- ZATCA Phase 2 drops
DROP TABLE IF EXISTS sync_watermarks CASCADE;
DROP TABLE IF EXISTS zatca_document_chain CASCADE;
DROP TABLE IF EXISTS zatca_device_configs CASCADE;
DROP TYPE IF EXISTS zatca_doc_status;

DROP VIEW IF EXISTS vw_pos_categories CASCADE;
DROP FUNCTION IF EXISTS fn_get_kds_orders CASCADE;
DROP FUNCTION IF EXISTS fn_get_waste_report CASCADE;
DROP FUNCTION IF EXISTS fn_calculate_recipe_cost CASCADE;
DROP FUNCTION IF EXISTS fn_get_item_modifiers CASCADE;
DROP FUNCTION IF EXISTS fn_get_restaurant_menu CASCADE;
DROP VIEW IF EXISTS vw_waste_daily_summary CASCADE;
DROP VIEW IF EXISTS vw_active_restaurant_orders CASCADE;
DROP VIEW IF EXISTS vw_recipe_bom CASCADE;
DROP VIEW IF EXISTS vw_restaurant_menu CASCADE;
DROP FUNCTION IF EXISTS fn_pos_search_products CASCADE;
DROP FUNCTION IF EXISTS fn_pos_get_products_by_category CASCADE;
DROP FUNCTION IF EXISTS fn_pos_get_product_by_barcode CASCADE;
DROP FUNCTION IF EXISTS fn_pos_get_products_with_stock CASCADE;
DROP VIEW IF EXISTS vw_pos_product_catalog CASCADE;

DROP INDEX IF EXISTS idx_products_active_sellable;
DROP INDEX IF EXISTS idx_inventory_stock_store_product_qty;
DROP INDEX IF EXISTS idx_products_sku_varchar_pattern;
DROP INDEX IF EXISTS idx_product_barcodes_barcode_lookup;

DROP TRIGGER IF EXISTS update_discount_analytics_updated_at ON discount_analytics;
DROP TRIGGER IF EXISTS update_profit_loss_analytics_updated_at ON profit_loss_analytics;

DROP TRIGGER IF EXISTS trg_restaurant_order_items_updated_at ON restaurant_order_items;
DROP TRIGGER IF EXISTS trg_restaurant_orders_updated_at ON restaurant_orders;
DROP TRIGGER IF EXISTS trg_recipes_updated_at ON recipes;
DROP TRIGGER IF EXISTS trg_menu_items_updated_at ON menu_items;
DROP TRIGGER IF EXISTS trg_menu_categories_updated_at ON menu_categories;
DROP TRIGGER IF EXISTS trg_restaurant_tables_updated_at ON restaurant_tables;
DROP TRIGGER IF EXISTS update_inventory_analytics_updated_at ON inventory_analytics;
DROP TRIGGER IF EXISTS update_purchase_analytics_updated_at ON purchase_analytics;
DROP TRIGGER IF EXISTS update_sales_analytics_updated_at ON sales_analytics;
DROP TRIGGER IF EXISTS update_sales_orders_updated_at ON sales_orders;
DROP TRIGGER IF EXISTS update_purchase_orders_updated_at ON purchase_orders;
DROP TRIGGER IF EXISTS update_customers_updated_at ON customers;
DROP TRIGGER IF EXISTS update_suppliers_updated_at ON suppliers;
DROP TRIGGER IF EXISTS update_inventory_stock_updated_at ON inventory_stock;
DROP TRIGGER IF EXISTS update_product_batches_updated_at ON product_batches;
DROP TRIGGER IF EXISTS update_product_serial_numbers_updated_at ON product_serial_numbers;
DROP TRIGGER IF EXISTS update_product_prices_updated_at ON product_prices;
DROP TRIGGER IF EXISTS update_product_variants_updated_at ON product_variants;
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP TRIGGER IF EXISTS update_price_lists_updated_at ON price_lists;
DROP TRIGGER IF EXISTS update_brands_updated_at ON brands;
DROP TRIGGER IF EXISTS update_product_categories_updated_at ON product_categories;
DROP TRIGGER IF EXISTS update_pos_terminals_updated_at ON pos_terminals;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_stores_updated_at ON stores;
DROP TRIGGER IF EXISTS update_role_ui_customizations_updated_at ON role_ui_customizations;
DROP TRIGGER IF EXISTS update_ui_settings_updated_at ON ui_settings;
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS update_submenus_updated_at ON submenus;
DROP TRIGGER IF EXISTS update_menus_updated_at ON menus;
DROP TRIGGER IF EXISTS update_modules_updated_at ON modules;
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;

DROP FUNCTION IF EXISTS update_updated_at_column CASCADE;

DROP INDEX IF EXISTS idx_discount_analytics_date;

DROP INDEX IF EXISTS idx_kiosk_sessions_token;
DROP INDEX IF EXISTS idx_kiosk_sessions_status;
DROP INDEX IF EXISTS idx_kiosk_sessions_store_id;
DROP INDEX IF EXISTS idx_kiosk_sessions_terminal_id;
DROP INDEX IF EXISTS idx_waste_logs_store_source_date;
DROP INDEX IF EXISTS idx_waste_logs_order_id;
DROP INDEX IF EXISTS idx_waste_logs_wasted_at;
DROP INDEX IF EXISTS idx_waste_logs_waste_source;
DROP INDEX IF EXISTS idx_waste_logs_recipe_id;
DROP INDEX IF EXISTS idx_waste_logs_menu_item_id;
DROP INDEX IF EXISTS idx_waste_logs_product_id;
DROP INDEX IF EXISTS idx_waste_logs_store_id;
DROP INDEX IF EXISTS idx_restaurant_order_items_status;
DROP INDEX IF EXISTS idx_restaurant_order_items_menu_item;
DROP INDEX IF EXISTS idx_restaurant_order_items_order_id;
DROP INDEX IF EXISTS idx_restaurant_orders_store_status_time;
DROP INDEX IF EXISTS idx_restaurant_orders_pos_txn_id;
DROP INDEX IF EXISTS idx_restaurant_orders_ordered_at;
DROP INDEX IF EXISTS idx_restaurant_orders_source;
DROP INDEX IF EXISTS idx_restaurant_orders_status;
DROP INDEX IF EXISTS idx_restaurant_orders_customer_id;
DROP INDEX IF EXISTS idx_restaurant_orders_session_id;
DROP INDEX IF EXISTS idx_restaurant_orders_cashier_id;
DROP INDEX IF EXISTS idx_restaurant_orders_table_id;
DROP INDEX IF EXISTS idx_restaurant_orders_store_id;
DROP INDEX IF EXISTS idx_recipe_ingredients_variant_id;
DROP INDEX IF EXISTS idx_recipe_ingredients_product_id;
DROP INDEX IF EXISTS idx_recipe_ingredients_recipe_id;
DROP INDEX IF EXISTS idx_recipes_code;
DROP INDEX IF EXISTS idx_recipes_is_active;
DROP INDEX IF EXISTS idx_recipes_finished_product_id;
DROP INDEX IF EXISTS idx_recipes_organization_id;
DROP INDEX IF EXISTS idx_menu_item_modifiers_is_active;
DROP INDEX IF EXISTS idx_menu_item_modifiers_item_id;
DROP INDEX IF EXISTS idx_menu_items_display_order;
DROP INDEX IF EXISTS idx_menu_items_is_available;
DROP INDEX IF EXISTS idx_menu_items_is_active;
DROP INDEX IF EXISTS idx_menu_items_recipe_id;
DROP INDEX IF EXISTS idx_menu_items_product_id;
DROP INDEX IF EXISTS idx_menu_items_category_id;
DROP INDEX IF EXISTS idx_menu_items_store_id;
DROP INDEX IF EXISTS idx_menu_categories_display_order;
DROP INDEX IF EXISTS idx_menu_categories_is_active;
DROP INDEX IF EXISTS idx_menu_categories_parent_id;
DROP INDEX IF EXISTS idx_menu_categories_store_id;
DROP INDEX IF EXISTS idx_restaurant_tables_section;
DROP INDEX IF EXISTS idx_restaurant_tables_is_active;
DROP INDEX IF EXISTS idx_restaurant_tables_store_id;

DROP INDEX IF EXISTS idx_discount_analytics_cashier_id;
DROP INDEX IF EXISTS idx_discount_analytics_store_id;
DROP INDEX IF EXISTS idx_discount_analytics_organization_id;
DROP INDEX IF EXISTS idx_profit_loss_analytics_period_type;
DROP INDEX IF EXISTS idx_profit_loss_analytics_date;
DROP INDEX IF EXISTS idx_profit_loss_analytics_store_id;
DROP INDEX IF EXISTS idx_profit_loss_analytics_organization_id;
DROP INDEX IF EXISTS idx_inventory_analytics_date;
DROP INDEX IF EXISTS idx_inventory_analytics_product_id;
DROP INDEX IF EXISTS idx_inventory_analytics_store_id;
DROP INDEX IF EXISTS idx_inventory_analytics_organization_id;
DROP INDEX IF EXISTS idx_purchase_analytics_date;
DROP INDEX IF EXISTS idx_purchase_analytics_product_id;
DROP INDEX IF EXISTS idx_purchase_analytics_supplier_id;
DROP INDEX IF EXISTS idx_purchase_analytics_store_id;
DROP INDEX IF EXISTS idx_purchase_analytics_organization_id;
DROP INDEX IF EXISTS idx_sales_analytics_year_month;
DROP INDEX IF EXISTS idx_sales_analytics_date;
DROP INDEX IF EXISTS idx_sales_analytics_customer_id;
DROP INDEX IF EXISTS idx_sales_analytics_category_id;
DROP INDEX IF EXISTS idx_sales_analytics_product_id;
DROP INDEX IF EXISTS idx_sales_analytics_store_id;
DROP INDEX IF EXISTS idx_sales_analytics_organization_id;
DROP INDEX IF EXISTS idx_pos_payments_payment_method;
DROP INDEX IF EXISTS idx_pos_payments_transaction_id;
DROP INDEX IF EXISTS idx_pos_transaction_lines_product_id;
DROP INDEX IF EXISTS idx_pos_transaction_lines_transaction_id;
DROP INDEX IF EXISTS idx_pos_transactions_status;
DROP INDEX IF EXISTS idx_pos_transactions_transaction_date;
DROP INDEX IF EXISTS idx_pos_transactions_transaction_number;
DROP INDEX IF EXISTS idx_pos_transactions_customer_id;
DROP INDEX IF EXISTS idx_pos_transactions_cashier_session_id;
DROP INDEX IF EXISTS idx_pos_transactions_cashier_id;
DROP INDEX IF EXISTS idx_pos_transactions_store_id;
-- DROP INDEX IF EXISTS idx_sales_order_lines_product_id;
-- DROP INDEX IF EXISTS idx_sales_order_lines_sales_order_id;
DROP INDEX IF EXISTS idx_sales_orders_order_date;
DROP INDEX IF EXISTS idx_sales_orders_status;
DROP INDEX IF EXISTS idx_sales_orders_order_number;
DROP INDEX IF EXISTS idx_sales_orders_store_id;
DROP INDEX IF EXISTS idx_sales_orders_customer_id;
DROP INDEX IF EXISTS idx_sales_orders_organization_id;
DROP INDEX IF EXISTS idx_purchase_order_lines_product_id;
DROP INDEX IF EXISTS idx_purchase_order_lines_purchase_order_id;
DROP INDEX IF EXISTS idx_purchase_orders_po_date;
DROP INDEX IF EXISTS idx_purchase_orders_status;
DROP INDEX IF EXISTS idx_purchase_orders_po_number;
DROP INDEX IF EXISTS idx_purchase_orders_store_id;
DROP INDEX IF EXISTS idx_purchase_orders_supplier_id;
DROP INDEX IF EXISTS idx_purchase_orders_organization_id;
DROP INDEX IF EXISTS idx_customers_customer_type;
DROP INDEX IF EXISTS idx_customers_is_active;
DROP INDEX IF EXISTS idx_customers_customer_code;
DROP INDEX IF EXISTS idx_customers_organization_id;
DROP INDEX IF EXISTS idx_suppliers_is_active;
DROP INDEX IF EXISTS idx_suppliers_code;
DROP INDEX IF EXISTS idx_suppliers_organization_id;
DROP INDEX IF EXISTS idx_stock_count_lines_product_id;
DROP INDEX IF EXISTS idx_stock_count_lines_stock_count_id;
DROP INDEX IF EXISTS idx_stock_counts_count_number;
DROP INDEX IF EXISTS idx_stock_counts_status;
DROP INDEX IF EXISTS idx_stock_counts_store_id;
DROP INDEX IF EXISTS idx_stock_movements_reference_type_id;
DROP INDEX IF EXISTS idx_stock_movements_movement_date;
DROP INDEX IF EXISTS idx_stock_movements_movement_type;
DROP INDEX IF EXISTS idx_stock_movements_to_store_id;
DROP INDEX IF EXISTS idx_stock_movements_from_store_id;
DROP INDEX IF EXISTS idx_stock_movements_product_id;
DROP INDEX IF EXISTS idx_inventory_stock_storage_location_id;
DROP INDEX IF EXISTS idx_inventory_stock_store_id;
DROP INDEX IF EXISTS idx_inventory_stock_product_variant_id;
DROP INDEX IF EXISTS idx_inventory_stock_product_id;
DROP INDEX IF EXISTS idx_product_batches_expiry_date;
DROP INDEX IF EXISTS idx_product_batches_status;
DROP INDEX IF EXISTS idx_product_batches_store_id;
DROP INDEX IF EXISTS idx_product_batches_batch_number;
DROP INDEX IF EXISTS idx_product_batches_product_id;
DROP INDEX IF EXISTS idx_product_serial_numbers_current_store_id;
DROP INDEX IF EXISTS idx_product_serial_numbers_status;
DROP INDEX IF EXISTS idx_product_serial_numbers_serial_number;
DROP INDEX IF EXISTS idx_product_serial_numbers_product_id;
DROP INDEX IF EXISTS idx_product_prices_is_active;
DROP INDEX IF EXISTS idx_product_prices_price_list_id;
DROP INDEX IF EXISTS idx_product_prices_product_variant_id;
DROP INDEX IF EXISTS idx_product_prices_product_id;
DROP INDEX IF EXISTS idx_product_barcodes_barcode;
DROP INDEX IF EXISTS idx_product_barcodes_product_variant_id;
DROP INDEX IF EXISTS idx_product_barcodes_product_id;
DROP INDEX IF EXISTS idx_product_variants_is_active;
DROP INDEX IF EXISTS idx_product_variants_variant_sku;
DROP INDEX IF EXISTS idx_product_variants_product_id;
DROP INDEX IF EXISTS idx_products_product_type;
DROP INDEX IF EXISTS idx_products_is_purchasable;
DROP INDEX IF EXISTS idx_products_is_sellable;
DROP INDEX IF EXISTS idx_products_is_active;
DROP INDEX IF EXISTS idx_products_brand_id;
DROP INDEX IF EXISTS idx_products_category_id;
DROP INDEX IF EXISTS idx_products_sku;
DROP INDEX IF EXISTS idx_products_organization_id;
DROP INDEX IF EXISTS idx_tax_categories_is_active;
DROP INDEX IF EXISTS idx_tax_categories_code;
DROP INDEX IF EXISTS idx_price_lists_valid_to;
DROP INDEX IF EXISTS idx_price_lists_valid_from;
DROP INDEX IF EXISTS idx_price_lists_is_active;
DROP INDEX IF EXISTS idx_price_lists_code;
DROP INDEX IF EXISTS idx_uom_pkg_template_levels_uom_id;
DROP INDEX IF EXISTS idx_uom_pkg_template_levels_template_id;
DROP INDEX IF EXISTS idx_uom_packaging_templates_code;
DROP INDEX IF EXISTS idx_uom_packaging_templates_organization_id;
DROP INDEX IF EXISTS idx_units_of_measure_uom_type;
DROP INDEX IF EXISTS idx_units_of_measure_code;
DROP INDEX IF EXISTS idx_brands_is_active;
DROP INDEX IF EXISTS idx_brands_code;
DROP INDEX IF EXISTS idx_product_categories_is_active;
DROP INDEX IF EXISTS idx_product_categories_code;
DROP INDEX IF EXISTS idx_product_categories_parent_category_id;
DROP INDEX IF EXISTS idx_cashier_sessions_opening_time;
DROP INDEX IF EXISTS idx_cashier_sessions_status;
DROP INDEX IF EXISTS idx_cashier_sessions_pos_terminal_id;
DROP INDEX IF EXISTS idx_cashier_sessions_cashier_id;
DROP INDEX IF EXISTS idx_pos_terminals_is_active;
DROP INDEX IF EXISTS idx_pos_terminals_store_id;
DROP INDEX IF EXISTS idx_cashiers_is_active;
DROP INDEX IF EXISTS idx_cashiers_store_id;
DROP INDEX IF EXISTS idx_cashiers_user_id;
DROP INDEX IF EXISTS idx_user_store_access_store_id;
DROP INDEX IF EXISTS idx_user_store_access_user_id;
DROP INDEX IF EXISTS idx_user_roles_role_id;
DROP INDEX IF EXISTS idx_user_roles_user_id;
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_employee_code;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_organization_id;
DROP INDEX IF EXISTS idx_storage_locations_code;
DROP INDEX IF EXISTS idx_storage_locations_parent_location_id;
DROP INDEX IF EXISTS idx_storage_locations_store_id;
DROP INDEX IF EXISTS idx_stores_store_type;
DROP INDEX IF EXISTS idx_stores_is_active;
DROP INDEX IF EXISTS idx_stores_code;
DROP INDEX IF EXISTS idx_stores_parent_store_id;
DROP INDEX IF EXISTS idx_stores_organization_id;
DROP INDEX IF EXISTS idx_role_permissions_permission_id;
DROP INDEX IF EXISTS idx_role_permissions_role_id;
DROP INDEX IF EXISTS idx_roles_is_active;
DROP INDEX IF EXISTS idx_roles_code;
DROP INDEX IF EXISTS idx_permissions_code;
DROP INDEX IF EXISTS idx_submenus_display_order;
DROP INDEX IF EXISTS idx_submenus_is_active;
DROP INDEX IF EXISTS idx_submenus_parent_submenu_id;
DROP INDEX IF EXISTS idx_submenus_menu_id;
DROP INDEX IF EXISTS idx_menus_display_order;
DROP INDEX IF EXISTS idx_menus_is_active;
DROP INDEX IF EXISTS idx_menus_parent_menu_id;
DROP INDEX IF EXISTS idx_menus_module_id;
DROP INDEX IF EXISTS idx_modules_display_order;
DROP INDEX IF EXISTS idx_modules_is_active;
DROP INDEX IF EXISTS idx_modules_code;
DROP INDEX IF EXISTS idx_tenants_is_active;
DROP INDEX IF EXISTS idx_tenants_slug;
DROP INDEX IF EXISTS idx_organizations_is_active;
DROP INDEX IF EXISTS idx_organizations_code;

DROP TABLE IF EXISTS discount_analytics CASCADE;
DROP TABLE IF EXISTS profit_loss_analytics CASCADE;
DROP TABLE IF EXISTS inventory_analytics CASCADE;
DROP TABLE IF EXISTS purchase_analytics CASCADE;
DROP TABLE IF EXISTS sales_analytics CASCADE;
DROP TABLE IF EXISTS kiosk_sessions CASCADE;
DROP TABLE IF EXISTS waste_logs CASCADE;
DROP TABLE IF EXISTS restaurant_order_items CASCADE;
DROP TABLE IF EXISTS restaurant_orders CASCADE;
DROP TABLE IF EXISTS recipe_ingredients CASCADE;
DROP TABLE IF EXISTS recipes CASCADE;
DROP TABLE IF EXISTS menu_item_modifiers CASCADE;
DROP TABLE IF EXISTS menu_items CASCADE;
DROP TABLE IF EXISTS menu_categories CASCADE;
DROP TABLE IF EXISTS restaurant_tables CASCADE;

DROP TABLE IF EXISTS pos_payments CASCADE;
DROP TABLE IF EXISTS pos_transaction_lines CASCADE;
DROP TABLE IF EXISTS pos_transactions CASCADE;
DROP TABLE IF EXISTS sales_order_lines CASCADE;
DROP TABLE IF EXISTS sales_orders CASCADE;
DROP TABLE IF EXISTS purchase_order_lines CASCADE;
DROP TABLE IF EXISTS purchase_orders CASCADE;
DROP TABLE IF EXISTS customers CASCADE;
DROP TABLE IF EXISTS suppliers CASCADE;
DROP TABLE IF EXISTS stock_count_lines CASCADE;
DROP TABLE IF EXISTS stock_counts CASCADE;
DROP TABLE IF EXISTS stock_movements CASCADE;
DROP TABLE IF EXISTS inventory_stock CASCADE;
DROP TABLE IF EXISTS product_batches CASCADE;
DROP TABLE IF EXISTS product_serial_numbers CASCADE;
DROP TABLE IF EXISTS product_uom_conversions CASCADE;
DROP TABLE IF EXISTS product_prices CASCADE;
DROP TABLE IF EXISTS product_barcodes CASCADE;
DROP TABLE IF EXISTS product_variants CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS tax_categories CASCADE;
DROP TABLE IF EXISTS price_lists CASCADE;
DROP TABLE IF EXISTS uom_packaging_template_levels CASCADE;
DROP TABLE IF EXISTS uom_packaging_templates CASCADE;
DROP TABLE IF EXISTS units_of_measure CASCADE;
DROP TABLE IF EXISTS brands CASCADE;
DROP TABLE IF EXISTS product_categories CASCADE;
DROP TABLE IF EXISTS cashier_sessions CASCADE;
DROP TABLE IF EXISTS pos_terminals CASCADE;
DROP TABLE IF EXISTS cashiers CASCADE;
DROP TABLE IF EXISTS user_store_access CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS storage_locations CASCADE;
DROP TABLE IF EXISTS stores CASCADE;
DROP TABLE IF EXISTS role_ui_customizations CASCADE;
DROP TABLE IF EXISTS ui_settings CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS submenu_permissions CASCADE;
DROP TABLE IF EXISTS menu_permissions CASCADE;
DROP TABLE IF EXISTS module_permissions CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS submenus CASCADE;
DROP TABLE IF EXISTS menus CASCADE;
DROP TABLE IF EXISTS modules CASCADE;
DROP TABLE IF EXISTS tenants CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;

-- Drop triggers
DROP TRIGGER IF EXISTS update_quote_lines_updated_at ON quote_lines;
DROP TRIGGER IF EXISTS update_quotes_updated_at ON quotes;
DROP TRIGGER IF EXISTS update_invoice_payments_updated_at ON invoice_payments;
DROP TRIGGER IF EXISTS update_invoice_lines_updated_at ON invoice_lines;
DROP TRIGGER IF EXISTS update_invoices_updated_at ON invoices;
DROP TRIGGER IF EXISTS update_order_fulfillments_updated_at ON order_fulfillments;
DROP TRIGGER IF EXISTS update_sales_order_lines_v2_updated_at ON sales_order_lines_v2;
DROP TRIGGER IF EXISTS update_sales_orders_v2_updated_at ON sales_orders_v2;
DROP TRIGGER IF EXISTS update_draft_cart_template_items_updated_at ON draft_cart_template_items;
DROP TRIGGER IF EXISTS update_draft_cart_templates_updated_at ON draft_cart_templates;
DROP TRIGGER IF EXISTS update_cart_items_updated_at ON cart_items;
DROP TRIGGER IF EXISTS update_carts_updated_at ON carts;

DROP TRIGGER IF EXISTS update_invoice_payment_trigger ON invoice_payments;
DROP TRIGGER IF EXISTS calculate_invoice_totals_trigger ON invoice_lines;
DROP TRIGGER IF EXISTS calculate_order_totals_trigger ON sales_order_lines_v2;
DROP TRIGGER IF EXISTS cart_items_activity_trigger ON cart_items;
DROP TRIGGER IF EXISTS cart_status_change_trigger ON carts;

-- Drop functions
DROP FUNCTION IF EXISTS update_invoice_payment();
DROP FUNCTION IF EXISTS calculate_invoice_totals();
DROP FUNCTION IF EXISTS calculate_order_totals();
DROP FUNCTION IF EXISTS update_cart_activity();
DROP FUNCTION IF EXISTS log_cart_status_change();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS quote_lines CASCADE;
DROP TABLE IF EXISTS quotes CASCADE;
DROP TABLE IF EXISTS invoice_status_history CASCADE;
DROP TABLE IF EXISTS invoice_payments CASCADE;
DROP TABLE IF EXISTS invoice_lines CASCADE;
DROP TABLE IF EXISTS invoices CASCADE;
DROP TABLE IF EXISTS order_fulfillment_items CASCADE;
DROP TABLE IF EXISTS order_fulfillments CASCADE;
DROP TABLE IF EXISTS order_status_history CASCADE;
DROP TABLE IF EXISTS sales_order_lines_v2 CASCADE;
DROP TABLE IF EXISTS sales_orders_v2 CASCADE;
DROP TABLE IF EXISTS draft_cart_template_items CASCADE;
DROP TABLE IF EXISTS draft_cart_templates CASCADE;
DROP TABLE IF EXISTS cart_activity_log CASCADE;
DROP TABLE IF EXISTS cart_items CASCADE;
DROP TABLE IF EXISTS carts CASCADE;
DROP TABLE IF EXISTS stock_reservations CASCADE;

-- Drop types
DROP TYPE IF EXISTS quote_status;
DROP TYPE IF EXISTS invoice_status;
DROP TYPE IF EXISTS invoice_type;
DROP TYPE IF EXISTS fulfillment_status;
DROP TYPE IF EXISTS payment_status;
DROP TYPE IF EXISTS order_status_v2;
DROP TYPE IF EXISTS order_type;
DROP TYPE IF EXISTS cart_type;
DROP TYPE IF EXISTS cart_status;
-- Note: Be careful with this in production
-- DROP EXTENSION IF EXISTS "uuid-ossp";
