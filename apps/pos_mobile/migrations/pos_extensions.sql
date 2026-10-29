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
