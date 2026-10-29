CREATE TABLE organizations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    legal_name VARCHAR(255),
    tax_id VARCHAR(50),
    currency_code VARCHAR(3) DEFAULT 'USD',
    fiscal_year_variant VARCHAR(10),
    is_active BOOLEAN DEFAULT true,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tenants (
    id TEXT PRIMARY KEY ,
    tenant_name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    db_conn_str TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    settings TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE profit_loss_analytics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE discount_analytics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE modules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    icon VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    display_order INTEGER DEFAULT 0,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE menus (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    module_id INTEGER NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    parent_menu_id INTEGER REFERENCES menus(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    route_path VARCHAR(255),
    icon VARCHAR(100),
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(module_id, code)
);

CREATE TABLE submenus (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    menu_id INTEGER NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    parent_submenu_id INTEGER REFERENCES submenus(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    route_path VARCHAR(255),
    icon VARCHAR(100),
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(menu_id, code)
);

CREATE TABLE permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE module_permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    module_id INTEGER NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    metadata TEXT DEFAULT '{}',
    UNIQUE(module_id, permission_id)
);

CREATE TABLE menu_permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    menu_id INTEGER NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    metadata TEXT DEFAULT '{}',
    UNIQUE(menu_id, permission_id)
);

CREATE TABLE submenu_permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    submenu_id INTEGER NOT NULL REFERENCES submenus(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    metadata TEXT DEFAULT '{}',
    UNIQUE(submenu_id, permission_id)
);

CREATE TABLE roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) UNIQUE NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    is_system_role BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE role_permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    scope VARCHAR(50) DEFAULT 'all',
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, permission_id)
);

CREATE TABLE ui_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    submenu_id INTEGER REFERENCES submenus(id) ON DELETE CASCADE,
    setting_key VARCHAR(100) NOT NULL,
    setting_value TEXT NOT NULL,
    description TEXT,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(submenu_id, setting_key)
);

CREATE TABLE role_ui_customizations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    submenu_id INTEGER NOT NULL REFERENCES submenus(id) ON DELETE CASCADE,
    customization_data TEXT,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, submenu_id)
);

CREATE TABLE stores (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    parent_store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    store_type VARCHAR(50),
    is_warehouse BOOLEAN DEFAULT false,
    is_pos_enabled BOOLEAN DEFAULT false,
    timezone VARCHAR(50) DEFAULT 'UTC',
    is_active BOOLEAN DEFAULT true,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, code)
);

CREATE TABLE storage_locations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    location_type VARCHAR(50),
    parent_location_id INTEGER REFERENCES storage_locations(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT true,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, code)
);

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    employee_code VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    metadata TEXT DEFAULT '{}',
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id)
);

CREATE TABLE user_store_access (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT false,
    metadata TEXT DEFAULT '{}',
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, store_id)
);

CREATE TABLE cashiers (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id       INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    store_id      INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    cashier_code  VARCHAR(50) NOT NULL,
    drawer_limit  DECIMAL(15,2),
    discount_limit DECIMAL(5,2) CHECK (discount_limit BETWEEN 0 AND 100),
    is_active     BOOLEAN   DEFAULT true,
    metadata      TEXT     DEFAULT '{}',
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, cashier_code)
);

CREATE TABLE pos_terminals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    terminal_code VARCHAR(50) NOT NULL,
    terminal_name VARCHAR(100),
    device_id VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, terminal_code)
);

CREATE TABLE cashier_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP  
);

CREATE TABLE product_categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_category_id INTEGER REFERENCES product_categories(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    category_level INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT true,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE brands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE units_of_measure (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(50) NOT NULL,
    uom_type VARCHAR(20),
    decimal_places INTEGER DEFAULT 2,
    is_active BOOLEAN DEFAULT true,
    metadata TEXT DEFAULT '{}'
);

CREATE TABLE uom_packaging_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    uom_id INTEGER NOT NULL REFERENCES units_of_measure(id),
    name VARCHAR(255) NOT NULL, 
    code VARCHAR(50) NOT NULL, 
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (uom_id, name)
);

CREATE TABLE uom_packaging_template_levels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INTEGER NOT NULL REFERENCES uom_packaging_templates(id) ON DELETE CASCADE,
    level_order INTEGER NOT NULL, 
    uom_id INTEGER NOT NULL REFERENCES units_of_measure(id),
    multiplier DECIMAL(15,6) NOT NULL DEFAULT 1, 
    UNIQUE(template_id, level_order)
);

CREATE TABLE price_lists (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    price_list_type VARCHAR(50),
    currency_code VARCHAR(3) DEFAULT 'USD',
    valid_from DATE,
    valid_to DATE,
    is_default BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tax_categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    tax_rate DECIMAL(5,2) NOT NULL,
    is_inclusive BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, sku)
);

CREATE TABLE product_variants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    variant_sku VARCHAR(100) UNIQUE NOT NULL,
    variant_name VARCHAR(255),
    variant_attributes TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_barcodes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
    barcode VARCHAR(100) UNIQUE NOT NULL,
    barcode_type VARCHAR(50),
    is_primary BOOLEAN DEFAULT false,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_prices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_uom_conversions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    from_uom_id INTEGER NOT NULL REFERENCES units_of_measure(id) ON DELETE CASCADE,
    to_uom_id INTEGER NOT NULL REFERENCES units_of_measure(id) ON DELETE CASCADE,
    conversion_factor DECIMAL(15,6) NOT NULL,
    is_default BOOLEAN DEFAULT false,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, from_uom_id, to_uom_id)
);

CREATE TABLE product_serial_numbers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
    serial_number VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'in_stock',
    current_store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    manufacturing_date DATE,
    expiry_date DATE,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE stock_reservations (
    id                   INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata             TEXT     DEFAULT '{}',
    created_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_batches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE CASCADE,
    batch_number VARCHAR(100) NOT NULL,
    manufacturing_date DATE,
    expiry_date DATE,
    store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    quantity_available DECIMAL(15,3) DEFAULT 0,
    status VARCHAR(50) DEFAULT 'active',
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, batch_number, store_id)
);

CREATE TABLE inventory_stock (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE stock_movements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE stock_counts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE stock_count_lines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE suppliers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, code)
);

CREATE TABLE customers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, customer_code)
);

CREATE TABLE purchase_orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE purchase_order_lines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transfer_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transfer_request_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE goods_receipt_note_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sales_order_lines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE carts (
    id TEXT PRIMARY KEY ,
    cart_number VARCHAR(50) UNIQUE NOT NULL, 
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL, 
    guest_identifier VARCHAR(255), 
    guest_email VARCHAR(255),
    guest_phone VARCHAR(50),
    cart_status VARCHAR(50) DEFAULT 'draft' NOT NULL,
    cart_type VARCHAR(50) DEFAULT 'standard' NOT NULL,
    channel VARCHAR(50) DEFAULT 'online', 
    payment_method VARCHAR(100),
    payment_gateway VARCHAR(100),
    device_info TEXT DEFAULT '{}', 
    created_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    cashier_id INTEGER REFERENCES cashiers(id) ON DELETE SET NULL,
    pos_terminal_id INTEGER REFERENCES pos_terminals(id) ON DELETE SET NULL,
    subtotal DECIMAL(15,2) DEFAULT 0.00,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    shipping_amount DECIMAL(15,2) DEFAULT 0.00,
    total_amount DECIMAL(15,2) DEFAULT 0.00,
    coupon_code VARCHAR(100),
    discount_code VARCHAR(100),
    promotional_credits DECIMAL(15,2) DEFAULT 0.00,
    shipping_address TEXT,
    billing_address TEXT,
    shipping_method VARCHAR(100),
    converted_to_order_id TEXT, 
    converted_at TIMESTAMP,
    last_activity_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP, 
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata TEXT DEFAULT '{}', 
    notes TEXT,
    CONSTRAINT chk_cart_customer CHECK (
        customer_id IS NOT NULL OR guest_identifier IS NOT NULL
    )
);

CREATE TABLE cart_items (
    id TEXT PRIMARY KEY ,
    cart_id TEXT NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
    quantity DECIMAL(15,3) NOT NULL CHECK (quantity > 0),
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    unit_price DECIMAL(15,2) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    line_total DECIMAL(15,2) NOT NULL,
    price_list_id INTEGER REFERENCES price_lists(id) ON DELETE SET NULL,
    tax_category_id INTEGER REFERENCES tax_categories(id) ON DELETE SET NULL,
    batch_number VARCHAR(100),
    serial_number VARCHAR(100),
    customization_details TEXT DEFAULT '{}', 
    notes TEXT,
    metadata TEXT DEFAULT '{}',
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(cart_id, product_id, product_variant_id, batch_number, serial_number)
);

CREATE TABLE cart_activity_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cart_id TEXT NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    activity_type VARCHAR(50) NOT NULL, 
    description TEXT,
    performed_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    ip_address INET,
    user_agent TEXT,
    old_value TEXT,
    new_value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE draft_cart_templates (
    id TEXT PRIMARY KEY ,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    template_name VARCHAR(255) NOT NULL,
    description TEXT,
    template_type VARCHAR(50) DEFAULT 'saved_cart', 
    is_favorite BOOLEAN DEFAULT false,
    auto_reorder_enabled BOOLEAN DEFAULT false,
    reorder_frequency_days INTEGER,
    next_reorder_date DATE,
    total_items INTEGER DEFAULT 0,
    estimated_total DECIMAL(15,2) DEFAULT 0.00,
    metadata TEXT DEFAULT '{}',
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, customer_id, template_name)
);

CREATE TABLE draft_cart_template_items (
    id TEXT PRIMARY KEY ,
    template_id TEXT NOT NULL REFERENCES draft_cart_templates(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
    quantity DECIMAL(15,3) NOT NULL CHECK (quantity > 0),
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    last_known_price DECIMAL(15,2),
    priority INTEGER DEFAULT 0,
    notes TEXT,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sales_orders_v2 (
    id TEXT PRIMARY KEY ,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
    customer_name VARCHAR(255),
    customer_email VARCHAR(255),
    customer_phone VARCHAR(50),
    order_type VARCHAR(50) DEFAULT 'standard' NOT NULL,
    order_status VARCHAR(50) DEFAULT 'draft' NOT NULL,
    payment_status VARCHAR(50) DEFAULT 'unpaid' NOT NULL,
    fulfillment_status VARCHAR(50) DEFAULT 'unfulfilled' NOT NULL,
    sales_channel VARCHAR(50) DEFAULT 'online', 
    order_source VARCHAR(100), 
    referral_source VARCHAR(255),
    source_cart_id TEXT REFERENCES carts(id) ON DELETE SET NULL,
    created_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    assigned_to_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    confirmed_date TIMESTAMP,
    expected_delivery_date DATE,
    actual_delivery_date DATE,
    cancelled_date TIMESTAMP,
    subtotal DECIMAL(15,2) DEFAULT 0.00,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    shipping_amount DECIMAL(15,2) DEFAULT 0.00,
    adjustment_amount DECIMAL(15,2) DEFAULT 0.00,
    total_amount DECIMAL(15,2) DEFAULT 0.00,
    paid_amount DECIMAL(15,2) DEFAULT 0.00,
    refunded_amount DECIMAL(15,2) DEFAULT 0.00,
    balance_due DECIMAL(15,2) DEFAULT 0.00,
    coupon_code VARCHAR(100),
    discount_codes TEXT[], 
    promotional_credits DECIMAL(15,2) DEFAULT 0.00,
    shipping_address TEXT NOT NULL,
    billing_address TEXT NOT NULL,
    shipping_method VARCHAR(100),
    shipping_carrier VARCHAR(100),
    tracking_number VARCHAR(255),
    tracking_url TEXT,
    payment_method VARCHAR(100),
    payment_gateway VARCHAR(100),
    payment_terms VARCHAR(100),
    payment_due_date DATE,
    pos_terminal_id INTEGER REFERENCES pos_terminals(id) ON DELETE SET NULL,
    cashier_id INTEGER REFERENCES cashiers(id) ON DELETE SET NULL,
    is_gift BOOLEAN DEFAULT false,
    gift_message TEXT,
    special_instructions TEXT,
    internal_notes TEXT,
    tags TEXT[],
    priority VARCHAR(20) DEFAULT 'normal', 
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sales_order_lines_v2 (
    id TEXT PRIMARY KEY ,
    sales_order_id TEXT NOT NULL REFERENCES sales_orders_v2(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    line_number INTEGER NOT NULL,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
    product_name VARCHAR(255) NOT NULL, 
    product_sku VARCHAR(100),
    quantity_ordered DECIMAL(15,3) NOT NULL CHECK (quantity_ordered > 0),
    quantity_fulfilled DECIMAL(15,3) DEFAULT 0.00,
    quantity_cancelled DECIMAL(15,3) DEFAULT 0.00,
    quantity_returned DECIMAL(15,3) DEFAULT 0.00,
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    unit_price DECIMAL(15,2) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    discount_percentage DECIMAL(5,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    line_total DECIMAL(15,2) NOT NULL,
    tax_category_id INTEGER REFERENCES tax_categories(id) ON DELETE SET NULL,
    tax_rate DECIMAL(5,2),
    batch_number VARCHAR(100),
    serial_numbers TEXT[], 
    expiry_date DATE,
    line_status VARCHAR(50) DEFAULT 'pending', 
    customization_details TEXT DEFAULT '{}',
    unit_cost DECIMAL(15,2),
    notes TEXT,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(sales_order_id, line_number)
);

CREATE TABLE order_status_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sales_order_id TEXT NOT NULL REFERENCES sales_orders_v2(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    reason VARCHAR(255),
    notes TEXT,
    changed_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE order_fulfillments (
    id TEXT PRIMARY KEY ,
    sales_order_id TEXT NOT NULL REFERENCES sales_orders_v2(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    fulfillment_number VARCHAR(50) UNIQUE NOT NULL,
    fulfillment_status VARCHAR(50) DEFAULT 'pending',
    shipment_status VARCHAR(50) DEFAULT 'pending', 
    fulfillment_store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    shipping_carrier VARCHAR(100),
    shipping_method VARCHAR(100),
    tracking_number VARCHAR(255),
    tracking_url TEXT,
    picked_at TIMESTAMP,
    packed_at TIMESTAMP,
    shipped_at TIMESTAMP,
    estimated_delivery_date DATE,
    actual_delivery_date DATE,
    picked_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    packed_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    notes TEXT,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE order_fulfillment_items (
    id TEXT PRIMARY KEY ,
    fulfillment_id TEXT NOT NULL REFERENCES order_fulfillments(id) ON DELETE CASCADE,
    order_line_id TEXT NOT NULL REFERENCES sales_order_lines_v2(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    quantity_fulfilled DECIMAL(15,3) NOT NULL CHECK (quantity_fulfilled > 0),
    batch_number VARCHAR(100),
    serial_numbers TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE pos_transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    sales_order_id TEXT REFERENCES sales_orders_v2(id) ON DELETE SET NULL,
    source_cart_id TEXT REFERENCES carts(id) ON DELETE SET NULL,
    voided_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    voided_at TIMESTAMP,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE pos_transaction_lines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE pos_payments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    transaction_id INTEGER NOT NULL REFERENCES pos_transactions(id) ON DELETE CASCADE,
    payment_method VARCHAR(50) NOT NULL,
    payment_gateway VARCHAR(50),
    amount DECIMAL(15,2) NOT NULL,
    payment_reference VARCHAR(100),
    reference_number VARCHAR(100),
    payment_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE restaurant_tables (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    store_id            INTEGER     NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    table_number        VARCHAR(20) NOT NULL,
    table_name          VARCHAR(100),
    section             VARCHAR(50),
    capacity            INTEGER     DEFAULT 4,
    is_active           BOOLEAN     DEFAULT true,
    metadata            TEXT       DEFAULT '{}',
    created_at          TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, table_number)
);

CREATE TABLE menu_categories (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata            TEXT        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, code)
);

CREATE TABLE menu_items (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata            TEXT        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE menu_item_modifiers (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    menu_item_id        INTEGER     NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    modifier_name       VARCHAR(100) NOT NULL,
    modifier_type       VARCHAR(30)  NOT NULL DEFAULT 'addon',
    price_adjustment    DECIMAL(15,2) DEFAULT 0,
    is_active           BOOLEAN     DEFAULT true,
    display_order       INTEGER     DEFAULT 0,
    metadata            TEXT       DEFAULT '{}',
    created_at          TIMESTAMP   DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE menu_modifier_groups (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    store_id           INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    name               VARCHAR(100) NOT NULL,
    code               VARCHAR(50)  NOT NULL,
    selection_type     VARCHAR(20) DEFAULT 'optional' CHECK (selection_type IN ('required','optional','multiple')),
    min_selections     INTEGER DEFAULT 0,
    max_selections     INTEGER,
    is_active          BOOLEAN   DEFAULT true,
    display_order      INTEGER   DEFAULT 0,
    metadata           TEXT     DEFAULT '{}',
    created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, code),
    CONSTRAINT chk_modifier_group_selections CHECK (
        min_selections >= 0
        AND (max_selections IS NULL OR max_selections >= min_selections)
    )
);

CREATE TABLE promotions (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    organization_id       INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    code                  VARCHAR(50) NOT NULL,
    name                  VARCHAR(255) NOT NULL,
    description           TEXT,
    promotion_type        VARCHAR(50) NOT NULL CHECK (promotion_type IN ('percentage_discount','fixed_discount','bogo','buy_x_get_y','free_item','bundle_price','points_multiplier','happy_hour')),
    action_metadata       TEXT DEFAULT '{}',
    valid_from            TIMESTAMP,
    valid_to              TIMESTAMP,
    schedule_json         TEXT DEFAULT '{}', 
    applies_to            VARCHAR(50) DEFAULT 'all' CHECK (applies_to IN ('all','category','product','customer_type','price_list')),
    target_product_ids    INTEGER[] DEFAULT '{}',
    target_category_ids   INTEGER[] DEFAULT '{}',
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
    metadata              TEXT     DEFAULT '{}',
    created_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, code)
);

CREATE TABLE recipes (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata              TEXT       DEFAULT '{}',
    created_at            TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, recipe_code)
);

CREATE TABLE recipe_ingredients (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    recipe_id           INTEGER      NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    product_id          INTEGER      NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id  INTEGER      REFERENCES product_variants(id) ON DELETE SET NULL,
    quantity            DECIMAL(15,3) NOT NULL,
    uom_id              INTEGER      REFERENCES units_of_measure(id) ON DELETE SET NULL,
    is_optional         BOOLEAN      DEFAULT false,
    is_byproduct        BOOLEAN      DEFAULT false,
    line_number         INTEGER,
    metadata            TEXT        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(recipe_id, product_id, product_variant_id)
);

CREATE TABLE combo_bundles (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata        TEXT     DEFAULT '{}',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, code)
);

CREATE TABLE menu_item_availability_schedules (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    menu_item_id    INTEGER NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    day_of_week     INTEGER CHECK (day_of_week BETWEEN 0 AND 6), 
    start_time      TIME    NOT NULL,
    end_time        TIME    NOT NULL,
    is_active       BOOLEAN   DEFAULT true,
    valid_from      DATE,
    valid_to        DATE,
    metadata        TEXT     DEFAULT '{}',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_schedule_times CHECK (end_time > start_time)
);

CREATE TABLE restaurant_orders (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata              TEXT        DEFAULT '{}',
    created_at            TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, order_number)
);

CREATE TABLE restaurant_order_items (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id            INTEGER      NOT NULL REFERENCES restaurant_orders(id) ON DELETE CASCADE,
    menu_item_id        INTEGER      NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    quantity            DECIMAL(15,3) NOT NULL DEFAULT 1,
    unit_price          DECIMAL(15,4) NOT NULL,
    modifiers_snapshot  TEXT        DEFAULT '[]',
    modifiers_total     DECIMAL(15,2) DEFAULT 0,
    discount_amount     DECIMAL(15,2) DEFAULT 0,
    tax_amount          DECIMAL(15,2) DEFAULT 0,
    subtotal            DECIMAL(15,2) NOT NULL,
    line_number         INTEGER,
    notes               TEXT,
    status              VARCHAR(30)  DEFAULT 'pending',
    metadata            TEXT        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE waste_logs (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata            TEXT        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE kiosk_sessions (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    pos_terminal_id     INTEGER      NOT NULL REFERENCES pos_terminals(id) ON DELETE CASCADE,
    store_id            INTEGER      NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    session_token       VARCHAR(255) NOT NULL UNIQUE,
    status              VARCHAR(20)  DEFAULT 'active',
    opened_at           TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    closed_at           TIMESTAMP,
    metadata            TEXT        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sales_analytics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE purchase_analytics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE inventory_analytics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE loyalty_redemption_rules (
    id                      INTEGER PRIMARY KEY AUTOINCREMENT,
    organization_id         INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    rule_name               VARCHAR(255) NOT NULL,
    points_earning_rate     DECIMAL(10,4) DEFAULT 1,  
    points_redemption_rate  DECIMAL(10,4) DEFAULT 1,  
    min_points_to_redeem    DECIMAL(15,2) DEFAULT 0,
    max_points_per_txn      DECIMAL(15,2),
    max_redemption_percent  DECIMAL(5,2) CHECK (max_redemption_percent BETWEEN 0 AND 100),
    eligible_product_types  TEXT[]   DEFAULT '{}',
    expiry_days             INTEGER,
    is_active               BOOLEAN  DEFAULT true,
    valid_from              DATE,
    valid_to                DATE,
    metadata                TEXT    DEFAULT '{}',
    created_at              TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE audit_logs (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    organization_id INTEGER REFERENCES organizations(id) ON DELETE CASCADE,
    table_name    VARCHAR(100) NOT NULL,
    record_id     VARCHAR(100) NOT NULL,
    action        VARCHAR(20)  NOT NULL CHECK (action IN ('INSERT','UPDATE','DELETE','SELECT')),
    old_values    TEXT,
    new_values    TEXT,
    changed_fields TEXT[],
    performed_by  INTEGER REFERENCES users(id) ON DELETE SET NULL,
    ip_address    INET,
    user_agent    TEXT,
    session_id    VARCHAR(255),
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE invoices (
    id TEXT PRIMARY KEY ,
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    customer_name VARCHAR(255) NOT NULL,
    customer_email VARCHAR(255),
    customer_phone VARCHAR(50),
    customer_tax_id VARCHAR(50),
    invoice_type VARCHAR(50) DEFAULT 'standard' NOT NULL,
    invoice_status VARCHAR(50) DEFAULT 'draft' NOT NULL,
    sales_order_id TEXT REFERENCES sales_orders_v2(id) ON DELETE SET NULL,
    related_invoice_id TEXT REFERENCES invoices(id) ON DELETE SET NULL, 
    invoice_date DATE NOT NULL DEFAULT CURRENT_DATE,
    due_date DATE NOT NULL,
    sent_date DATE,
    paid_date DATE,
    subtotal DECIMAL(15,2) DEFAULT 0.00,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    shipping_amount DECIMAL(15,2) DEFAULT 0.00,
    adjustment_amount DECIMAL(15,2) DEFAULT 0.00,
    total_amount DECIMAL(15,2) NOT NULL,
    paid_amount DECIMAL(15,2) DEFAULT 0.00,
    credit_applied DECIMAL(15,2) DEFAULT 0.00,
    balance_due DECIMAL(15,2) NOT NULL,
    payment_terms VARCHAR(100), 
    currency_code VARCHAR(3) DEFAULT 'USD',
    exchange_rate DECIMAL(15,6) DEFAULT 1.000000,
    billing_address TEXT NOT NULL,
    shipping_address TEXT,
    is_recurring BOOLEAN DEFAULT false,
    recurrence_pattern VARCHAR(50), 
    next_invoice_date DATE,
    pdf_url TEXT,
    document_hash VARCHAR(255), 
    reminder_sent_count INTEGER DEFAULT 0,
    last_reminder_sent_at TIMESTAMP,
    notes TEXT,
    internal_notes TEXT,
    reference_number VARCHAR(100), 
    created_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    metadata TEXT DEFAULT '{}',
    tags TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sales_returns (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata            TEXT     DEFAULT '{}',
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE invoice_lines (
    id TEXT PRIMARY KEY ,
    invoice_id TEXT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    line_number INTEGER NOT NULL,
    description TEXT NOT NULL,
    item_type VARCHAR(50) DEFAULT 'product', 
    product_id INTEGER REFERENCES products(id) ON DELETE SET NULL,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
    product_sku VARCHAR(100),
    order_line_id TEXT REFERENCES sales_order_lines_v2(id) ON DELETE SET NULL,
    quantity DECIMAL(15,3) DEFAULT 1.000,
    unit_price DECIMAL(15,2) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    line_total DECIMAL(15,2) NOT NULL,
    tax_category_id INTEGER REFERENCES tax_categories(id) ON DELETE SET NULL,
    tax_rate DECIMAL(5,2),
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(invoice_id, line_number)
);

CREATE TABLE invoice_payments (
    id TEXT PRIMARY KEY ,
    invoice_id TEXT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    payment_number VARCHAR(50) UNIQUE NOT NULL,
    payment_date DATE NOT NULL DEFAULT CURRENT_DATE,
    payment_amount DECIMAL(15,2) NOT NULL CHECK (payment_amount > 0),
    payment_method VARCHAR(100) NOT NULL, 
    payment_gateway VARCHAR(100),
    payment_reference VARCHAR(255), 
    currency_code VARCHAR(3) DEFAULT 'USD',
    exchange_rate DECIMAL(15,6) DEFAULT 1.000000,
    bank_account_id INTEGER, 
    reconciled BOOLEAN DEFAULT false,
    reconciled_date DATE,
    notes TEXT,
    received_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE invoice_status_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invoice_id TEXT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    reason VARCHAR(255),
    notes TEXT,
    changed_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE quotes (
    id TEXT PRIMARY KEY ,
    quote_number VARCHAR(50) UNIQUE NOT NULL,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    customer_id INTEGER REFERENCES customers(id) ON DELETE SET NULL,
    customer_name VARCHAR(255) NOT NULL,
    customer_email VARCHAR(255),
    customer_phone VARCHAR(50),
    quote_status VARCHAR(50) DEFAULT 'draft' NOT NULL,
    quote_date DATE NOT NULL DEFAULT CURRENT_DATE,
    valid_until DATE NOT NULL,
    sent_date DATE,
    accepted_date DATE,
    converted_date DATE,
    subtotal DECIMAL(15,2) DEFAULT 0.00,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    total_amount DECIMAL(15,2) NOT NULL,
    converted_to_order_id TEXT REFERENCES sales_orders_v2(id) ON DELETE SET NULL,
    payment_terms VARCHAR(100),
    delivery_terms TEXT,
    terms_and_conditions TEXT,
    notes TEXT,
    internal_notes TEXT,
    created_by_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE combo_bundle_items (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata         TEXT     DEFAULT '{}',
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE quote_lines (
    id TEXT PRIMARY KEY ,
    quote_id TEXT NOT NULL REFERENCES quotes(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    line_number INTEGER NOT NULL,
    product_id INTEGER REFERENCES products(id) ON DELETE SET NULL,
    product_variant_id INTEGER REFERENCES product_variants(id) ON DELETE SET NULL,
    description TEXT NOT NULL,
    quantity DECIMAL(15,3) NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(15,2) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    line_total DECIMAL(15,2) NOT NULL,
    uom_id INTEGER REFERENCES units_of_measure(id) ON DELETE SET NULL,
    notes TEXT,
    metadata TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(quote_id, line_number)
);

CREATE TABLE sales_return_lines (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
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
    metadata           TEXT     DEFAULT '{}',
    created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

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

CREATE INDEX idx_draft_cart_templates_organization_id ON draft_cart_templates(organization_id);

CREATE INDEX idx_draft_cart_templates_customer_id ON draft_cart_templates(customer_id);

CREATE INDEX idx_draft_cart_templates_template_type ON draft_cart_templates(template_type);

CREATE INDEX idx_draft_cart_templates_is_favorite ON draft_cart_templates(is_favorite);

CREATE INDEX idx_draft_cart_templates_auto_reorder ON draft_cart_templates(auto_reorder_enabled);

CREATE INDEX idx_draft_cart_templates_next_reorder_date ON draft_cart_templates(next_reorder_date);

CREATE INDEX idx_draft_cart_template_items_template_id ON draft_cart_template_items(template_id);

CREATE INDEX idx_draft_cart_template_items_product_id ON draft_cart_template_items(product_id);

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

CREATE INDEX idx_quotes_organization_id ON quotes(organization_id);

CREATE INDEX idx_quotes_customer_id ON quotes(customer_id);

CREATE INDEX idx_quotes_quote_number ON quotes(quote_number);

CREATE INDEX idx_quotes_quote_status ON quotes(quote_status);

CREATE INDEX idx_quotes_quote_date ON quotes(quote_date);

CREATE INDEX idx_quotes_valid_until ON quotes(valid_until);

CREATE INDEX idx_quotes_converted_to_order_id ON quotes(converted_to_order_id);

CREATE INDEX idx_quote_lines_quote_id ON quote_lines(quote_id);

CREATE INDEX idx_quote_lines_product_id ON quote_lines(product_id);
