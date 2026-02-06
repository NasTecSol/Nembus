-- +goose Up
-- Combined Initial Schema Migration: Base Tables + POS Views/Functions (with Type Fixes) + Indexes

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- CORE MASTER DATA
-- =====================================================

CREATE TABLE organizations (
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
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    cashier_code VARCHAR(50) NOT NULL,
    drawer_limit DECIMAL(15,2),
    discount_limit DECIMAL(5,2),
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
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
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
    expected_quantity DECIMAL(15,3) DEFAULT 0,
    counted_quantity DECIMAL(15,3) DEFAULT 0,
    variance DECIMAL(15,3) DEFAULT 0,
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
    received_quantity DECIMAL(15,3) DEFAULT 0,
    line_number INTEGER,
    metadata JSONB DEFAULT '{}',
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
    shipped_quantity DECIMAL(15,3) DEFAULT 0,
    line_number INTEGER,
    metadata JSONB DEFAULT '{}',
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
    transaction_number VARCHAR(50) UNIQUE NOT NULL,
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    subtotal DECIMAL(15,2) DEFAULT 0,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) DEFAULT 0,
    amount_paid DECIMAL(15,2) DEFAULT 0,
    change_given DECIMAL(15,2) DEFAULT 0,
    status VARCHAR(50) DEFAULT 'completed',
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
    amount DECIMAL(15,2) NOT NULL,
    payment_reference VARCHAR(100),
    payment_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
    date DATE NOT NULL,
    month INTEGER,
    quarter INTEGER,
    year INTEGER,
    opening_stock DECIMAL(15,3) DEFAULT 0,
    stock_in DECIMAL(15,3) DEFAULT 0,
    stock_out DECIMAL(15,3) DEFAULT 0,
    closing_stock DECIMAL(15,3) DEFAULT 0,
    stock_value DECIMAL(15,2) DEFAULT 0,
    turnover_rate DECIMAL(5,2),
    days_in_stock DECIMAL(5,2),
    low_stock_alerts INTEGER DEFAULT 0,
    out_of_stock_days INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

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

-- Inventory Analytics
ALTER TABLE inventory_analytics 
    ADD CONSTRAINT fk_inventory_analytics_store 
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE SET NULL;

ALTER TABLE inventory_analytics 
    ADD CONSTRAINT fk_inventory_analytics_product 
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL;

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
CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_modules_updated_at BEFORE UPDATE ON modules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_menus_updated_at BEFORE UPDATE ON menus FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_submenus_updated_at BEFORE UPDATE ON submenus FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ui_settings_updated_at BEFORE UPDATE ON ui_settings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_role_ui_customizations_updated_at BEFORE UPDATE ON role_ui_customizations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_stores_updated_at BEFORE UPDATE ON stores FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pos_terminals_updated_at BEFORE UPDATE ON pos_terminals FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_product_categories_updated_at BEFORE UPDATE ON product_categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_brands_updated_at BEFORE UPDATE ON brands FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_price_lists_updated_at BEFORE UPDATE ON price_lists FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_product_variants_updated_at BEFORE UPDATE ON product_variants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_product_prices_updated_at BEFORE UPDATE ON product_prices FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_product_serial_numbers_updated_at BEFORE UPDATE ON product_serial_numbers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_product_batches_updated_at BEFORE UPDATE ON product_batches FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_inventory_stock_updated_at BEFORE UPDATE ON inventory_stock FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_suppliers_updated_at BEFORE UPDATE ON suppliers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_purchase_orders_updated_at BEFORE UPDATE ON purchase_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_sales_orders_updated_at BEFORE UPDATE ON sales_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_sales_analytics_updated_at BEFORE UPDATE ON sales_analytics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_purchase_analytics_updated_at BEFORE UPDATE ON purchase_analytics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_inventory_analytics_updated_at BEFORE UPDATE ON inventory_analytics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_profit_loss_analytics_updated_at BEFORE UPDATE ON profit_loss_analytics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_discount_analytics_updated_at BEFORE UPDATE ON discount_analytics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

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
    p.id AS product_id,
    p.sku,
    p.name AS product_name,
    p.description,
    p.product_type,
    pc.id AS category_id,
    pc.name AS category_name,
    pc.code AS category_code,
    pc_parent.id AS parent_category_id,
    pc_parent.name AS parent_category_name,
    b.id AS brand_id,
    b.name AS brand_name,
    uom.id AS uom_id,
    uom.code AS uom_code,
    uom.name AS uom_name,
    uom.decimal_places,
    pb.barcode,
    pb.barcode_type,
    tc.id AS tax_category_id,
    tc.name AS tax_category_name,
    tc.tax_rate,
    tc.is_inclusive AS tax_is_inclusive,
    pp_retail.price AS retail_price,
    pp_retail.id AS retail_price_id,
    pp_promo.price AS promo_price,
    pp_promo.id AS promo_price_id,
    pp_promo.min_quantity AS promo_min_quantity,
    pp_promo.valid_from AS promo_valid_from,
    pp_promo.valid_to AS promo_valid_to,
    pp_promo.metadata->>'promotion_name' AS promotion_name,
    pp_promo.metadata->>'discount_percent' AS discount_percent,
    CASE 
        WHEN pp_promo.id IS NOT NULL 
             AND pp_promo.is_active = true
             AND pp_promo.valid_from <= CURRENT_DATE 
             AND (pp_promo.valid_to IS NULL OR pp_promo.valid_to >= CURRENT_DATE)
        THEN pp_promo.price
        ELSE pp_retail.price
    END AS effective_price,
    CASE 
        WHEN pp_promo.id IS NOT NULL 
             AND pp_promo.is_active = true
             AND pp_promo.valid_from <= CURRENT_DATE 
             AND (pp_promo.valid_to IS NULL OR pp_promo.valid_to >= CURRENT_DATE)
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
LEFT JOIN product_categories pc ON p.category_id = pc.id
LEFT JOIN product_categories pc_parent ON pc.parent_category_id = pc_parent.id
LEFT JOIN brands b ON p.brand_id = b.id
LEFT JOIN units_of_measure uom ON p.base_uom_id = uom.id
LEFT JOIN product_barcodes pb ON p.id = pb.product_id AND pb.is_primary = true
LEFT JOIN tax_categories tc ON p.tax_category_id = tc.id
LEFT JOIN product_prices pp_retail 
    ON p.id = pp_retail.product_id 
    AND pp_retail.price_list_id = (SELECT id FROM price_lists WHERE code = 'RETAIL_SAR' AND is_active = true)
    AND pp_retail.is_active = true
LEFT JOIN product_prices pp_promo 
    ON p.id = pp_promo.product_id 
    AND pp_promo.price_list_id = (SELECT id FROM price_lists WHERE code = 'PROMO_SAR' AND is_active = true)
    AND pp_promo.is_active = true
WHERE p.is_active = true AND p.is_sellable = true
ORDER BY pc.name, p.name;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_pos_get_products_with_stock(
    p_store_id INTEGER,
    p_category_id INTEGER DEFAULT NULL,
    p_search_term VARCHAR DEFAULT NULL,
    p_include_out_of_stock BOOLEAN DEFAULT false
)
RETURNS TABLE (
    product_id INTEGER,
    sku VARCHAR,
    product_name VARCHAR,
    description TEXT,
    category_id INTEGER,
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
    discount_percent VARCHAR,
    promo_min_quantity NUMERIC,
    tax_rate NUMERIC,
    tax_is_inclusive BOOLEAN,
    quantity_available NUMERIC,
    quantity_on_hand NUMERIC,
    quantity_allocated NUMERIC,
    is_in_stock BOOLEAN,
    is_low_stock BOOLEAN,
    reorder_level NUMERIC,
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
        cat.category_id,
        cat.category_name::VARCHAR,
        cat.brand_name::VARCHAR,
        cat.barcode::VARCHAR,
        cat.uom_code::VARCHAR,
        (cat.decimal_places)::INTEGER,
        cat.retail_price,
        cat.promo_price,
        cat.effective_price,
        cat.has_active_promotion,
        cat.promotion_name::VARCHAR,
        cat.discount_percent::VARCHAR,
        cat.promo_min_quantity,
        cat.tax_rate,
        cat.tax_is_inclusive,
        COALESCE(inv.quantity_available, 0)::NUMERIC,
        COALESCE(inv.quantity_on_hand, 0)::NUMERIC,
        COALESCE(inv.quantity_allocated, 0)::NUMERIC,
        (COALESCE(inv.quantity_available, 0) > 0),
        (COALESCE(inv.quantity_available, 0) <= COALESCE(inv.reorder_level, 0) AND COALESCE(inv.quantity_available, 0) > 0),
        COALESCE(inv.reorder_level, 0)::NUMERIC,
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
    LEFT JOIN inventory_stock inv ON cat.product_id = inv.product_id AND inv.store_id = p_store_id
    WHERE 
        (p_category_id IS NULL OR cat.category_id = p_category_id)
        AND (p_search_term IS NULL 
             OR cat.product_name ILIKE '%' || p_search_term || '%'
             OR cat.sku ILIKE '%' || p_search_term || '%'
             OR cat.barcode ILIKE '%' || p_search_term || '%')
        AND (p_include_out_of_stock = true OR COALESCE(inv.quantity_available, 0) > 0)
    ORDER BY cat.category_name, cat.product_name;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

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
        cat.promo_price,
        cat.effective_price,
        cat.has_active_promotion,
        cat.promotion_name::VARCHAR,
        cat.promo_min_quantity,
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
    pc.id AS category_id,
    pc.code AS category_code,
    pc.name AS category_name,
    pc.parent_category_id,
    pc_parent.name AS parent_category_name,
    COUNT(DISTINCT p.id)::INTEGER AS product_count,
    COUNT(DISTINCT CASE WHEN inv.quantity_available > 0 THEN p.id END)::INTEGER AS in_stock_count,
    pc.metadata AS category_metadata
FROM product_categories pc
LEFT JOIN product_categories pc_parent ON pc.parent_category_id = pc_parent.id
LEFT JOIN products p ON pc.id = p.category_id AND p.is_active = true AND p.is_sellable = true
LEFT JOIN inventory_stock inv ON p.id = inv.product_id
WHERE pc.is_active = true
GROUP BY pc.id, pc.code, pc.name, pc.parent_category_id, pc_parent.name, pc.metadata
HAVING COUNT(DISTINCT p.id) > 0
ORDER BY pc_parent.name NULLS FIRST, pc.name;

-- +goose Down

DROP VIEW IF EXISTS vw_pos_categories CASCADE;
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
DROP INDEX IF EXISTS idx_sales_order_lines_product_id;
DROP INDEX IF EXISTS idx_sales_order_lines_sales_order_id;
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

-- Note: Be careful with this in production
-- DROP EXTENSION IF EXISTS "uuid-ossp";