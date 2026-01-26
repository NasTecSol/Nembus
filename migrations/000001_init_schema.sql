-- +goose Up
-- Migration: 001_complete_schema.up.sql
-- ERP System - Complete Database Schema
-- All Tables and Relations

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
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
    storage_location_id INTEGER REFERENCES storage_locations(id) ON DELETE SET NULL,
    system_quantity DECIMAL(15,3),
    counted_quantity DECIMAL(15,3),
    variance DECIMAL(15,3),
    variance_value DECIMAL(15,2),
    batch_number VARCHAR(100),
    serial_number VARCHAR(100),
    counted_at TIMESTAMP,
    metadata JSONB DEFAULT '{}'
);

-- =====================================================
-- SUPPLIERS & CUSTOMERS
-- =====================================================

CREATE TABLE suppliers (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    supplier_type VARCHAR(50),
    payment_terms VARCHAR(50),
    credit_limit DECIMAL(15,2),
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, code)
);

CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    customer_code VARCHAR(50),
    name VARCHAR(255) NOT NULL,
    customer_type VARCHAR(50),
    price_list_id INTEGER REFERENCES price_lists(id) ON DELETE SET NULL,
    credit_limit DECIMAL(15,2),
    outstanding_balance DECIMAL(15,2) DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, customer_code)
);

-- =====================================================
-- PURCHASE ORDERS
-- =====================================================

CREATE TABLE purchase_orders (
    id SERIAL PRIMARY KEY,
    po_number VARCHAR(50) UNIQUE NOT NULL,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    supplier_id INTEGER NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    po_date DATE NOT NULL,
    expected_delivery_date DATE,
    status VARCHAR(50) DEFAULT 'draft',
    subtotal DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) NOT NULL,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    approved_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE purchase_order_lines (
    id SERIAL PRIMARY KEY,
    purchase_order_id INTEGER NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    line_number INTEGER NOT NULL,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
    quantity DECIMAL(15,3) NOT NULL,
    received_quantity DECIMAL(15,3) DEFAULT 0,
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    unit_price DECIMAL(15,2) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    line_total DECIMAL(15,2) NOT NULL,
    metadata JSONB DEFAULT '{}',
    UNIQUE(purchase_order_id, line_number)
);

-- =====================================================
-- SALES ORDERS
-- =====================================================

CREATE TABLE sales_orders (
    id SERIAL PRIMARY KEY,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    order_date DATE NOT NULL,
    delivery_date DATE,
    price_list_id INTEGER REFERENCES price_lists(id) ON DELETE SET NULL,
    status VARCHAR(50) DEFAULT 'draft',
    subtotal DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) NOT NULL,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    approved_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sales_order_lines (
    id SERIAL PRIMARY KEY,
    sales_order_id INTEGER NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    line_number INTEGER NOT NULL,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
    quantity DECIMAL(15,3) NOT NULL,
    shipped_quantity DECIMAL(15,3) DEFAULT 0,
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    unit_price DECIMAL(15,2) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    line_total DECIMAL(15,2) NOT NULL,
    cost_price DECIMAL(15,2),
    metadata JSONB DEFAULT '{}',
    UNIQUE(sales_order_id, line_number)
);

-- =====================================================
-- POS TRANSACTIONS
-- =====================================================

CREATE TABLE pos_transactions (
    id SERIAL PRIMARY KEY,
    transaction_number VARCHAR(50) UNIQUE NOT NULL,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    pos_terminal_id INTEGER NOT NULL REFERENCES pos_terminals(id) ON DELETE CASCADE,
    cashier_session_id INTEGER NOT NULL REFERENCES cashier_sessions(id) ON DELETE CASCADE,
    cashier_id INTEGER NOT NULL REFERENCES cashiers(id) ON DELETE CASCADE,
    customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
    price_list_id INTEGER REFERENCES price_lists(id) ON DELETE SET NULL,
    transaction_type VARCHAR(50) DEFAULT 'sale',
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    subtotal DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) NOT NULL,
    total_cost DECIMAL(15,2),
    status VARCHAR(50) DEFAULT 'completed',
    voided_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    voided_at TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE pos_transaction_lines (
    id SERIAL PRIMARY KEY,
    transaction_id INTEGER NOT NULL REFERENCES pos_transactions(id) ON DELETE CASCADE,
    line_number INTEGER NOT NULL,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
    serial_number VARCHAR(100),
    batch_number VARCHAR(100),
    quantity DECIMAL(15,3) NOT NULL,
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    unit_price DECIMAL(15,2) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    line_total DECIMAL(15,2) NOT NULL,
    cost_price DECIMAL(15,2),
    metadata JSONB DEFAULT '{}',
    UNIQUE(transaction_id, line_number)
);

CREATE TABLE pos_payments (
    id SERIAL PRIMARY KEY,
    transaction_id INTEGER NOT NULL REFERENCES pos_transactions(id) ON DELETE CASCADE,
    payment_method VARCHAR(50) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    reference_number VARCHAR(100),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- ANALYTICS TABLES
-- =====================================================

CREATE TABLE sales_analytics (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    product_id INTEGER REFERENCES products(id) ON DELETE SET NULL,
    category_id INTEGER REFERENCES product_categories(id) ON DELETE SET NULL,
    customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
    date DATE NOT NULL,
    hour INTEGER,
    day_of_week INTEGER,
    week_number INTEGER,
    month INTEGER,
    quarter INTEGER,
    year INTEGER,
    total_transactions INTEGER DEFAULT 0,
    total_quantity DECIMAL(15,3) DEFAULT 0,
    gross_sales DECIMAL(15,2) DEFAULT 0,
    discounts DECIMAL(15,2) DEFAULT 0,
    taxes DECIMAL(15,2) DEFAULT 0,
    net_sales DECIMAL(15,2) DEFAULT 0,
    total_cost DECIMAL(15,2) DEFAULT 0,
    gross_profit DECIMAL(15,2) DEFAULT 0,
    profit_margin DECIMAL(5,2),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE purchase_analytics (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    supplier_id INTEGER REFERENCES suppliers(id) ON DELETE SET NULL,
    product_id INTEGER REFERENCES products(id) ON DELETE SET NULL,
    category_id INTEGER REFERENCES product_categories(id) ON DELETE SET NULL,
    date DATE NOT NULL,
    month INTEGER,
    quarter INTEGER,
    year INTEGER,
    total_orders INTEGER DEFAULT 0,
    total_quantity DECIMAL(15,3) DEFAULT 0,
    total_amount DECIMAL(15,2) DEFAULT 0,
    discounts_received DECIMAL(15,2) DEFAULT 0,
    taxes_paid DECIMAL(15,2) DEFAULT 0,
    net_amount DECIMAL(15,2) DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE inventory_analytics (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES product_categories(id) ON DELETE SET NULL,
    date DATE NOT NULL,
    month INTEGER,
    quarter INTEGER,
    year INTEGER,
    opening_stock DECIMAL(15,3) DEFAULT 0,
    closing_stock DECIMAL(15,3) DEFAULT 0,
    average_stock DECIMAL(15,3) DEFAULT 0,
    stock_value DECIMAL(15,2) DEFAULT 0,
    receipts DECIMAL(15,3) DEFAULT 0,
    issues DECIMAL(15,3) DEFAULT 0,
    adjustments DECIMAL(15,3) DEFAULT 0,
    stock_turnover_ratio DECIMAL(10,2),
    days_of_inventory DECIMAL(10,2),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- ADD FOREIGN KEY CONSTRAINTS (after all tables created)
-- =====================================================

ALTER TABLE profit_loss_analytics 
    ADD CONSTRAINT fk_profit_loss_analytics_store 
    FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE SET NULL;

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

-- +goose Down

DROP TABLE IF EXISTS discount_analytics CASCADE;
DROP TABLE IF EXISTS profit_loss_analytics CASCADE;
DROP TABLE IF EXISTS inventory_analytics CASCADE;
DROP TABLE IF EXISTS purchase_analytics CASCADE;
DROP TABLE IF EXISTS sales_analytics CASCADE;

-- Drop POS Tables
DROP TABLE IF EXISTS pos_payments CASCADE;
DROP TABLE IF EXISTS pos_transaction_lines CASCADE;
DROP TABLE IF EXISTS pos_transactions CASCADE;

-- Drop Sales Order Tables
DROP TABLE IF EXISTS sales_order_lines CASCADE;
DROP TABLE IF EXISTS sales_orders CASCADE;

-- Drop Purchase Order Tables
DROP TABLE IF EXISTS purchase_order_lines CASCADE;
DROP TABLE IF EXISTS purchase_orders CASCADE;

-- Drop Customers and Suppliers
DROP TABLE IF EXISTS customers CASCADE;
DROP TABLE IF EXISTS suppliers CASCADE;

-- Drop Inventory Tables
DROP TABLE IF EXISTS stock_count_lines CASCADE;
DROP TABLE IF EXISTS stock_counts CASCADE;
DROP TABLE IF EXISTS stock_movements CASCADE;
DROP TABLE IF EXISTS inventory_stock CASCADE;

-- Drop Product Related Tables
DROP TABLE IF EXISTS product_batches CASCADE;
DROP TABLE IF EXISTS product_serial_numbers CASCADE;
DROP TABLE IF EXISTS product_uom_conversions CASCADE;
DROP TABLE IF EXISTS product_prices CASCADE;
DROP TABLE IF EXISTS product_barcodes CASCADE;
DROP TABLE IF EXISTS product_variants CASCADE;
DROP TABLE IF EXISTS products CASCADE;

-- Drop Product Master Data
DROP TABLE IF EXISTS tax_categories CASCADE;
DROP TABLE IF EXISTS price_lists CASCADE;
DROP TABLE IF EXISTS units_of_measure CASCADE;
DROP TABLE IF EXISTS brands CASCADE;
DROP TABLE IF EXISTS product_categories CASCADE;

-- Drop Cashier and POS Terminal Tables
DROP TABLE IF EXISTS cashier_sessions CASCADE;
DROP TABLE IF EXISTS pos_terminals CASCADE;
DROP TABLE IF EXISTS cashiers CASCADE;

-- Drop User Management Tables
DROP TABLE IF EXISTS user_store_access CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Drop Store Tables
DROP TABLE IF EXISTS storage_locations CASCADE;
DROP TABLE IF EXISTS stores CASCADE;

-- Drop Permission and Access Control Tables
DROP TABLE IF EXISTS role_ui_customizations CASCADE;
DROP TABLE IF EXISTS ui_settings CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS submenu_permissions CASCADE;
DROP TABLE IF EXISTS menu_permissions CASCADE;
DROP TABLE IF EXISTS module_permissions CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;

-- Drop Navigation Tables
DROP TABLE IF EXISTS submenus CASCADE;
DROP TABLE IF EXISTS menus CASCADE;
DROP TABLE IF EXISTS modules CASCADE;

-- Drop Core Tables
DROP TABLE IF EXISTS tenants CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;

-- Drop Functions
DROP FUNCTION IF EXISTS update_updated_at_column CASCADE;
-- Drop Extensions
-- Note: Be careful with this in production
-- DROP EXTENSION IF EXISTS "uuid-ossp";