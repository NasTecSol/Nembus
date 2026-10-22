-- =====================================================
-- SAP BUSINESS ONE MIGRATION STAGING SCHEMA & TABLES
-- =====================================================

CREATE SCHEMA IF NOT EXISTS staging;

-- Staging Batches Audit Log
CREATE TABLE IF NOT EXISTS staging.sap_migration_batches (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) UNIQUE NOT NULL,
    run_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    domain VARCHAR(50) NOT NULL,
    record_count INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(30) DEFAULT 'staged', -- staged, merged, failed
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staging Stores
CREATE TABLE IF NOT EXISTS staging.sap_stores (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    store_type VARCHAR(50),
    is_warehouse BOOLEAN DEFAULT true,
    is_pos_enabled BOOLEAN DEFAULT true,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staging Products
CREATE TABLE IF NOT EXISTS staging.sap_products (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    sku VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_code VARCHAR(50),
    brand_code VARCHAR(50),
    uom_code VARCHAR(20),
    product_type VARCHAR(50) DEFAULT 'standard',
    is_serialized BOOLEAN DEFAULT false,
    is_batch_managed BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    is_sellable BOOLEAN DEFAULT true,
    is_purchasable BOOLEAN DEFAULT true,
    track_inventory BOOLEAN DEFAULT true,
    primary_barcode VARCHAR(100),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staging Inventory Stock
CREATE TABLE IF NOT EXISTS staging.sap_inventory (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    product_sku VARCHAR(100) NOT NULL,
    store_code VARCHAR(50) NOT NULL,
    quantity_on_hand DECIMAL(15,3) DEFAULT 0,
    quantity_allocated DECIMAL(15,3) DEFAULT 0,
    quantity_available DECIMAL(15,3) DEFAULT 0,
    quantity_on_order DECIMAL(15,3) DEFAULT 0,
    reorder_level DECIMAL(15,3),
    max_stock_level DECIMAL(15,3),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staging Business Partners (Domain 5, OCRD)
CREATE TABLE IF NOT EXISTS staging.sap_business_partners (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    partner_type VARCHAR(50), -- customer, supplier, lead
    tax_id VARCHAR(50),
    currency_code VARCHAR(3) DEFAULT 'SAR',
    email VARCHAR(255),
    phone VARCHAR(50),
    credit_limit DECIMAL(15,2) DEFAULT 0,
    outstanding_balance DECIMAL(15,2) DEFAULT 0,
    payment_term_code VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staging BP Addresses (Domain 8, CRD1)
CREATE TABLE IF NOT EXISTS staging.sap_bp_addresses (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    partner_code VARCHAR(50) NOT NULL,
    address_name VARCHAR(100),
    address_type VARCHAR(20) NOT NULL, -- billing, shipping
    street TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    zip_code VARCHAR(20),
    country_code VARCHAR(3) DEFAULT 'SA',
    phone VARCHAR(50),
    is_default BOOLEAN DEFAULT false,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staging Price Lists (Domain 7, OPLN)
CREATE TABLE IF NOT EXISTS staging.sap_price_lists (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    currency_code VARCHAR(3) DEFAULT 'SAR',
    factor DECIMAL(10,4) DEFAULT 1.0,
    based_on_code VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staging Price List Items (Domain 7, ITM1)
CREATE TABLE IF NOT EXISTS staging.sap_price_list_items (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    price_list_code VARCHAR(50) NOT NULL,
    product_sku VARCHAR(100) NOT NULL,
    uom_code VARCHAR(20),
    price DECIMAL(15,2) NOT NULL,
    currency_code VARCHAR(3) DEFAULT 'SAR',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staging Purchase Orders (Domain 8, OPOR)
CREATE TABLE IF NOT EXISTS staging.sap_purchase_orders (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    po_number VARCHAR(50) NOT NULL,
    supplier_code VARCHAR(50) NOT NULL,
    supplier_name VARCHAR(255),
    store_code VARCHAR(50),
    po_date TIMESTAMP NOT NULL,
    expected_delivery_date TIMESTAMP,
    status VARCHAR(50) DEFAULT 'sent',
    subtotal DECIMAL(15,2) DEFAULT 0,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) DEFAULT 0,
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staging Purchase Order Lines (Domain 8, POR1)
CREATE TABLE IF NOT EXISTS staging.sap_purchase_order_lines (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    po_number VARCHAR(50) NOT NULL,
    line_number INTEGER NOT NULL,
    product_sku VARCHAR(100) NOT NULL,
    product_name VARCHAR(255),
    store_code VARCHAR(50),
    quantity DECIMAL(15,3) NOT NULL,
    received_quantity DECIMAL(15,3) DEFAULT 0,
    uom_code VARCHAR(20),
    unit_price DECIMAL(15,4) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    subtotal DECIMAL(15,2) DEFAULT 0,
    line_total DECIMAL(15,2) NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staging Goods Receipt Notes (Domain 9, OPDN)
CREATE TABLE IF NOT EXISTS staging.sap_goods_receipt_notes (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    grn_number VARCHAR(50) NOT NULL,
    po_number VARCHAR(50),
    supplier_code VARCHAR(50) NOT NULL,
    supplier_name VARCHAR(255),
    store_code VARCHAR(50),
    receipt_date TIMESTAMP NOT NULL,
    delivery_note_number VARCHAR(100),
    status VARCHAR(50) DEFAULT 'posted',
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staging Goods Receipt Items (Domain 9, PDN1)
CREATE TABLE IF NOT EXISTS staging.sap_goods_receipt_items (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    grn_number VARCHAR(50) NOT NULL,
    line_number INTEGER NOT NULL,
    product_sku VARCHAR(100) NOT NULL,
    product_name VARCHAR(255),
    store_code VARCHAR(50),
    quantity_received DECIMAL(15,3) NOT NULL,
    quantity_rejected DECIMAL(15,3) DEFAULT 0,
    uom_code VARCHAR(20),
    unit_cost DECIMAL(15,4) NOT NULL,
    source_po_line_num INTEGER,
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staging Stock Movements (Domain 10, OINM)
CREATE TABLE IF NOT EXISTS staging.sap_stock_movements (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    movement_type VARCHAR(50) NOT NULL,
    reference_type VARCHAR(50),
    reference_number VARCHAR(100),
    product_sku VARCHAR(100) NOT NULL,
    from_store_code VARCHAR(50),
    to_store_code VARCHAR(50),
    quantity DECIMAL(15,3) NOT NULL,
    uom_code VARCHAR(20),
    movement_date TIMESTAMP NOT NULL,
    cost_per_unit DECIMAL(15,4),
    total_value DECIMAL(15,2),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staging Incoming Payments (Domain 11, ORCT / RCT2)
CREATE TABLE IF NOT EXISTS staging.sap_incoming_payments (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    payment_number VARCHAR(50) NOT NULL,
    invoice_number VARCHAR(50),
    customer_code VARCHAR(50) NOT NULL,
    customer_name VARCHAR(255),
    payment_date DATE NOT NULL,
    payment_amount DECIMAL(15,2) NOT NULL,
    payment_method VARCHAR(50),
    payment_reference VARCHAR(255),
    currency_code VARCHAR(3) DEFAULT 'SAR',
    is_on_account BOOLEAN DEFAULT false,
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staging Sales Returns (Domain 12, ORIN / ORDN)
CREATE TABLE IF NOT EXISTS staging.sap_sales_returns (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    return_number VARCHAR(50) NOT NULL,
    customer_code VARCHAR(50) NOT NULL,
    customer_name VARCHAR(255),
    store_code VARCHAR(50),
    return_date TIMESTAMP NOT NULL,
    reason TEXT,
    status VARCHAR(50) DEFAULT 'completed',
    subtotal DECIMAL(15,2) DEFAULT 0,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    total_refund_amount DECIMAL(15,2) DEFAULT 0,
    doc_source VARCHAR(50), -- credit_memo, return
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staging Sales Return Lines (Domain 12, RIN1 / RDN1)
CREATE TABLE IF NOT EXISTS staging.sap_sales_return_lines (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    return_number VARCHAR(50) NOT NULL,
    line_number INTEGER NOT NULL,
    product_sku VARCHAR(100) NOT NULL,
    product_name VARCHAR(255),
    store_code VARCHAR(50),
    uom_code VARCHAR(20),
    quantity DECIMAL(15,3) NOT NULL,
    unit_price DECIMAL(15,4) NOT NULL,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    refund_amount DECIMAL(15,2) NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staging Payment Terms (Domain 14, OCTG)
CREATE TABLE IF NOT EXISTS staging.sap_payment_terms (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    due_days INTEGER NOT NULL DEFAULT 0,
    discount_days INTEGER DEFAULT 0,
    discount_percentage DECIMAL(5,2) DEFAULT 0.00,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Staging Outgoing Payments (Domain 15, OVPM / VPM2)
CREATE TABLE IF NOT EXISTS staging.sap_outgoing_payments (
    id SERIAL PRIMARY KEY,
    batch_id VARCHAR(100) NOT NULL,
    organization_id INTEGER NOT NULL,
    payment_number VARCHAR(50) NOT NULL,
    supplier_code VARCHAR(50) NOT NULL,
    supplier_name VARCHAR(255),
    po_doc_entry BIGINT,
    payment_date DATE NOT NULL,
    payment_amount DECIMAL(15,2) NOT NULL,
    payment_method VARCHAR(50),
    payment_reference VARCHAR(255),
    currency_code VARCHAR(3) DEFAULT 'SAR',
    notes TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
