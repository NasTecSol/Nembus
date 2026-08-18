-- =====================================================
-- CORE MASTER DATA
-- =====================================================

CREATE TABLE IF NOT EXISTS organizations(
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    legal_name VARCHAR(255),
    tax_id VARCHAR(50),
    currency_code VARCHAR(3) DEFAULT 'SAR',
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

