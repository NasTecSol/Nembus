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

