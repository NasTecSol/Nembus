-- =====================================================
-- POS TRANSACTIONS
-- =====================================================

CREATE TABLE pos_transactions (
    id SERIAL PRIMARY KEY,
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
    sales_order_id UUID REFERENCES sales_orders_v2(id) ON DELETE SET NULL,
    source_cart_id UUID REFERENCES carts(id) ON DELETE SET NULL,
    voided_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    voided_at TIMESTAMP,
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
    line_total DECIMAL(15,2) DEFAULT 0,
    cost_price DECIMAL(15,2) DEFAULT 0,
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
    payment_gateway VARCHAR(50),
    amount DECIMAL(15,2) NOT NULL,
    payment_reference VARCHAR(100),
    reference_number VARCHAR(100),
    payment_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- RESTAURANT MODULE TABLES
-- =====================================================

CREATE TABLE restaurant_tables (
    id                  SERIAL PRIMARY KEY,
    store_id            INTEGER     NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    table_number        VARCHAR(20) NOT NULL,
    table_name          VARCHAR(100),
    section             VARCHAR(50),
    capacity            INTEGER     DEFAULT 4,
    is_active           BOOLEAN     DEFAULT true,
    metadata            JSONB       DEFAULT '{}',
    created_at          TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, table_number)
);

CREATE TABLE menu_categories (
    id                  SERIAL PRIMARY KEY,
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
    metadata            JSONB        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, code)
);

CREATE TABLE menu_items (
    id                  SERIAL PRIMARY KEY,
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
    metadata            JSONB        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE menu_item_modifiers (
    id                  SERIAL PRIMARY KEY,
    menu_item_id        INTEGER     NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    modifier_name       VARCHAR(100) NOT NULL,
    modifier_type       VARCHAR(30)  NOT NULL DEFAULT 'addon',
    price_adjustment    DECIMAL(15,2) DEFAULT 0,
    is_active           BOOLEAN     DEFAULT true,
    display_order       INTEGER     DEFAULT 0,
    metadata            JSONB       DEFAULT '{}',
    created_at          TIMESTAMP   DEFAULT CURRENT_TIMESTAMP
);
-- FIX #10 (P1): New menu_modifier_groups table to enforce min/max modifier selections
CREATE TABLE menu_modifier_groups (
    id                 SERIAL PRIMARY KEY,
    store_id           INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    name               VARCHAR(100) NOT NULL,
    code               VARCHAR(50)  NOT NULL,
    selection_type     VARCHAR(20) DEFAULT 'optional' CHECK (selection_type IN ('required','optional','multiple')),
    min_selections     INTEGER DEFAULT 0,
    max_selections     INTEGER,
    is_active          BOOLEAN   DEFAULT true,
    display_order      INTEGER   DEFAULT 0,
    metadata           JSONB     DEFAULT '{}',
    created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, code),
    CONSTRAINT chk_modifier_group_selections CHECK (
        min_selections >= 0
        AND (max_selections IS NULL OR max_selections >= min_selections)
    )
);

-- FIX #14 (P1): Central promotion/coupon definition table
CREATE TABLE promotions (
    id                    SERIAL PRIMARY KEY,
    organization_id       INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    code                  VARCHAR(50) NOT NULL,
    name                  VARCHAR(255) NOT NULL,
    description           TEXT,
    promotion_type        VARCHAR(50) NOT NULL CHECK (promotion_type IN ('percentage_discount','fixed_discount','bogo','buy_x_get_y','free_item','bundle_price','points_multiplier','happy_hour')),
    -- FIX #16: Action metadata for complex rules
    action_metadata       JSONB DEFAULT '{}',
    -- FIX #15: Scheduling
    valid_from            TIMESTAMP,
    valid_to              TIMESTAMP,
    schedule_json         JSONB DEFAULT '{}', -- e.g. {"days":[1,2,3],"start_time":"12:00","end_time":"14:00"}
    -- Applicability
    applies_to            VARCHAR(50) DEFAULT 'all' CHECK (applies_to IN ('all','category','product','customer_type','price_list')),
    target_product_ids    INTEGER[] DEFAULT '{}',
    target_category_ids   INTEGER[] DEFAULT '{}',
    -- FIX #17: customer segmentation
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
    metadata              JSONB     DEFAULT '{}',
    created_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, code)
);




CREATE TABLE recipes (
    id                    SERIAL PRIMARY KEY,
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
    metadata              JSONB       DEFAULT '{}',
    created_at            TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, recipe_code)
);

CREATE TABLE recipe_ingredients (
    id                  SERIAL PRIMARY KEY,
    recipe_id           INTEGER      NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    product_id          INTEGER      NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_variant_id  INTEGER      REFERENCES product_variants(id) ON DELETE SET NULL,
    quantity            DECIMAL(15,3) NOT NULL,
    uom_id              INTEGER      REFERENCES units_of_measure(id) ON DELETE SET NULL,
    is_optional         BOOLEAN      DEFAULT false,
    is_byproduct        BOOLEAN      DEFAULT false,
    line_number         INTEGER,
    metadata            JSONB        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(recipe_id, product_id, product_variant_id)
);
-- FIX #13 (P1): Combo / meal deal / bundle support
CREATE TABLE combo_bundles (
    id              SERIAL PRIMARY KEY,
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
    metadata        JSONB     DEFAULT '{}',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, code)
);

ALTER TABLE menu_items
    ADD CONSTRAINT fk_menu_items_recipe
    FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE SET NULL;

-- FIX #11 (P1): New time-based / day-of-week menu availability schedules
CREATE TABLE menu_item_availability_schedules (
    id              SERIAL PRIMARY KEY,
    menu_item_id    INTEGER NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    day_of_week     INTEGER CHECK (day_of_week BETWEEN 0 AND 6), -- 0=Sunday, 6=Saturday; NULL = every day
    start_time      TIME    NOT NULL,
    end_time        TIME    NOT NULL,
    is_active       BOOLEAN   DEFAULT true,
    valid_from      DATE,
    valid_to        DATE,
    metadata        JSONB     DEFAULT '{}',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_schedule_times CHECK (end_time > start_time)
);    

CREATE TABLE restaurant_orders (
    id                    SERIAL PRIMARY KEY,
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
    metadata              JSONB        DEFAULT '{}',
    created_at            TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, order_number)
);

CREATE TABLE restaurant_order_items (
    id                  SERIAL PRIMARY KEY,
    order_id            INTEGER      NOT NULL REFERENCES restaurant_orders(id) ON DELETE CASCADE,
    menu_item_id        INTEGER      NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    quantity            DECIMAL(15,3) NOT NULL DEFAULT 1,
    unit_price          DECIMAL(15,4) NOT NULL,
    modifiers_snapshot  JSONB        DEFAULT '[]',
    modifiers_total     DECIMAL(15,2) DEFAULT 0,
    discount_amount     DECIMAL(15,2) DEFAULT 0,
    tax_amount          DECIMAL(15,2) DEFAULT 0,
    subtotal            DECIMAL(15,2) NOT NULL,
    line_number         INTEGER,
    notes               TEXT,
    status              VARCHAR(30)  DEFAULT 'pending',
    metadata            JSONB        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE waste_logs (
    id                  SERIAL PRIMARY KEY,
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
    metadata            JSONB        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE kiosk_sessions (
    id                  SERIAL PRIMARY KEY,
    pos_terminal_id     INTEGER      NOT NULL REFERENCES pos_terminals(id) ON DELETE CASCADE,
    store_id            INTEGER      NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    session_token       VARCHAR(255) NOT NULL UNIQUE,
    status              VARCHAR(20)  DEFAULT 'active',
    opened_at           TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    closed_at           TIMESTAMP,
    metadata            JSONB        DEFAULT '{}',
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

