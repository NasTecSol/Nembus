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

-- Offline operation queue for sync tracking (Transactional Outbox)
CREATE TABLE IF NOT EXISTS sync_queue (
    id              BIGSERIAL    PRIMARY KEY,
    entity_type     VARCHAR(50)  NOT NULL,
    entity_id       BIGINT       NOT NULL,
    action          VARCHAR(20)  NOT NULL, -- 'INSERT', 'UPDATE', 'DELETE'
    payload         JSONB        NOT NULL,
    status          VARCHAR(20)  DEFAULT 'pending',
    priority        INTEGER      DEFAULT 0,         -- Higher = process first
    correlation_id  VARCHAR(100),                    -- Group related sync items
    max_retries     INTEGER      DEFAULT 5,
    retry_count     INTEGER      DEFAULT 0,
    last_error      TEXT,
    created_at      TIMESTAMPTZ  DEFAULT NOW(),
    synced_at       TIMESTAMPTZ
);

-- Index for efficient outbox draining (pending items, highest priority first)
CREATE INDEX IF NOT EXISTS idx_sync_queue_pending
    ON sync_queue(status, priority DESC, created_at ASC)
    WHERE status = 'pending';

-- Local POS terminal hardware identity with ZATCA sync tracking
CREATE TABLE IF NOT EXISTS local_device_config (
    id                  SERIAL       PRIMARY KEY,
    device_name         VARCHAR(100) NOT NULL,
    store_id            INTEGER      REFERENCES stores(id),
    pos_terminal_id     INTEGER      REFERENCES pos_terminals(id),
    last_zatca_sync_at  TIMESTAMPTZ  DEFAULT '1970-01-01',   -- Delta-fetch watermark
    zatca_enabled       BOOLEAN      DEFAULT false,           -- Local ZATCA toggle
    created_at          TIMESTAMPTZ  DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS local_device_config;
DROP TABLE IF EXISTS sync_queue;
DROP TABLE IF EXISTS local_printer_configs;
