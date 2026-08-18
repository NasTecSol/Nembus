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

