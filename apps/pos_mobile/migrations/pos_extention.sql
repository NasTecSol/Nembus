-- Local printer hardware configuration
CREATE TABLE IF NOT EXISTS local_printer_configs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    printer_name    VARCHAR(100) NOT NULL,
    connection_type VARCHAR(50)  NOT NULL, -- 'USB', 'Network', 'Serial'
    ip_address      VARCHAR(45),
    port            INTEGER,
    is_default      BOOLEAN      DEFAULT 0,
    created_at      TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

-- Offline operation queue for sync tracking (Transactional Outbox)
CREATE TABLE IF NOT EXISTS sync_queue (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_type     VARCHAR(50)  NOT NULL,
    entity_id       TEXT         NOT NULL,
    action          VARCHAR(20)  NOT NULL, -- 'INSERT', 'UPDATE', 'DELETE'
    payload         TEXT         NOT NULL, -- JSON payload
    status          VARCHAR(20)  DEFAULT 'pending', -- 'pending', 'synced', 'failed'
    priority        INTEGER      DEFAULT 0,         -- Higher = process first
    correlation_id  VARCHAR(100),                   -- Group related sync items
    max_retries     INTEGER      DEFAULT 5,
    retry_count     INTEGER      DEFAULT 0,
    last_error      TEXT,
    created_at      TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    synced_at       TIMESTAMP
);

-- Index for efficient outbox draining (pending items, highest priority first)
CREATE INDEX IF NOT EXISTS idx_sync_queue_pending
    ON sync_queue(status, priority DESC, created_at ASC);

-- Local POS terminal hardware identity with ZATCA sync tracking
CREATE TABLE IF NOT EXISTS local_device_config (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    device_name         VARCHAR(100) NOT NULL DEFAULT 'POS Terminal',
    store_id            INTEGER      REFERENCES stores(id),
    pos_terminal_id     INTEGER      REFERENCES pos_terminals(id),
    last_zatca_sync_at  TIMESTAMP    DEFAULT '1970-01-01 00:00:00',
    zatca_enabled       BOOLEAN      DEFAULT 0,
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

-- Default local device config row if none exists
INSERT INTO local_device_config (id, device_name, store_id, pos_terminal_id, last_zatca_sync_at, zatca_enabled)
SELECT 1, 'Main POS Terminal', 1, 1, '1970-01-01 00:00:00', 0
WHERE NOT EXISTS (SELECT 1 FROM local_device_config WHERE id = 1);

-- Restaurant menu view for POS & Mobile
CREATE VIEW IF NOT EXISTS vw_restaurant_menu AS
SELECT
    mi.id                       AS menu_item_id,
    mi.store_id,
    mi.name                     AS item_name,
    mi.short_name,
    mi.description,
    mi.image_url,
    mi.base_price,
    mi.cost_price,
    mi.preparation_time_min,
    mi.is_available,
    mi.is_active,
    mi.display_order,
    'standard'                  AS item_type,
    COALESCE(mi.metadata, '{}') AS item_metadata,
    mc.id                       AS category_id,
    mc.name                     AS category_name,
    mc.code                     AS category_code,
    mc.parent_category_id,
    mc.display_order            AS category_display_order,
    mc.image_url                AS category_image_url,
    mc_parent.name              AS parent_category_name,
    tc.id                       AS tax_category_id,
    tc.tax_rate,
    tc.is_inclusive             AS tax_is_inclusive,
    mi.recipe_id,
    r.recipe_name,
    r.yield_quantity            AS recipe_yield,
    mi.product_id,
    NULL                        AS product_variant_id,
    p.sku                       AS product_sku,
    (SELECT COUNT(*) FROM menu_item_modifiers m WHERE m.menu_item_id = mi.id AND m.is_active = 1)
                                AS active_modifier_count,
    CASE
        WHEN mi.base_price > 0 AND mi.cost_price > 0
        THEN ROUND(((mi.base_price - mi.cost_price) / mi.base_price) * 100, 2)
        ELSE NULL
    END                         AS margin_percent
FROM menu_items mi
JOIN menu_categories mc         ON mi.menu_category_id = mc.id
LEFT JOIN menu_categories mc_parent ON mc.parent_category_id = mc_parent.id
LEFT JOIN tax_categories tc     ON mi.tax_category_id = tc.id
LEFT JOIN recipes r             ON mi.recipe_id = r.id
LEFT JOIN products p            ON mi.product_id = p.id
WHERE mi.is_active = 1;

-- Active restaurant orders view for KDS & Floor Management
CREATE VIEW IF NOT EXISTS vw_active_restaurant_orders AS
SELECT
    ro.id                       AS order_id,
    ro.order_number,
    ro.store_id,
    ro.order_source,
    ro.status                   AS order_status,
    ro.subtotal,
    ro.tax_amount,
    ro.total_amount,
    ro.notes,
    ro.ordered_at,
    ro.confirmed_at,
    rt.id                       AS table_id,
    rt.table_number,
    rt.table_name,
    rt.section                  AS table_section,
    c.id                        AS cashier_id,
    COALESCE(u.first_name || ' ' || u.last_name, 'Cashier') AS waiter_name,
    ro.customer_id,
    cust.name                   AS customer_name,
    0.0                         AS minutes_since_ordered
FROM restaurant_orders ro
LEFT JOIN restaurant_tables rt  ON ro.table_id = rt.id
LEFT JOIN cashiers c            ON ro.cashier_id = c.id
LEFT JOIN users u               ON c.user_id = u.id
LEFT JOIN customers cust        ON ro.customer_id = cust.id
WHERE ro.status <> 'paid' AND ro.status <> 'voided' AND ro.status <> 'completed';

