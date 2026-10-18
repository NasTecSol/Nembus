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
