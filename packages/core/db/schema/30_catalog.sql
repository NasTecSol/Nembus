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

