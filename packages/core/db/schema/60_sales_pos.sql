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


