-- +goose Up

-- Local printer hardware configuration
CREATE TABLE IF NOT EXISTS local_printer_configs (
    id              SERIAL PRIMARY KEY,
    printer_name    VARCHAR(100) NOT NULL,
    connection_type VARCHAR(50)  NOT NULL, -- 'USB', 'Network', 'Serial'
    ip_address      VARCHAR(45),
    port            INTEGER,
    is_default      BOOLEAN      DEFAULT false,
    created_at      TIMESTAMPTZ  DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  DEFAULT NOW()
);

-- Offline operation queue for sync tracking
CREATE TABLE IF NOT EXISTS sync_queue (
    id          BIGSERIAL    PRIMARY KEY,
    entity_type VARCHAR(50)  NOT NULL,
    entity_id   BIGINT       NOT NULL,
    action      VARCHAR(20)  NOT NULL, -- 'INSERT', 'UPDATE', 'DELETE'
    payload     JSONB        NOT NULL,
    status      VARCHAR(20)  DEFAULT 'pending',
    retry_count INTEGER      DEFAULT 0,
    last_error  TEXT,
    created_at  TIMESTAMPTZ  DEFAULT NOW(),
    synced_at   TIMESTAMPTZ
);

-- Local POS terminal hardware identity
CREATE TABLE IF NOT EXISTS local_device_config (
    id              SERIAL       PRIMARY KEY,
    device_name     VARCHAR(100) NOT NULL,
    store_id        INTEGER      REFERENCES stores(id),
    pos_terminal_id INTEGER      REFERENCES pos_terminals(id),
    created_at      TIMESTAMPTZ  DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS local_device_config;
DROP TABLE IF EXISTS sync_queue;
DROP TABLE IF EXISTS local_printer_configs;
