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

-- +goose StatementBegin
-- Generic outbox trigger function for automatic sync_queue population
CREATE OR REPLACE FUNCTION trg_enqueue_sync_outbox()
RETURNS TRIGGER AS $$
DECLARE
    entity_name VARCHAR(50);
    prio INTEGER := 5;
BEGIN
    entity_name := TG_ARGV[0];
    IF entity_name = 'pos_transactions' OR entity_name = 'pos_payments' THEN
        prio := 10;
    END IF;

    INSERT INTO sync_queue (entity_type, entity_id, action, payload, status, priority)
    VALUES (
        entity_name,
        NEW.id,
        TG_OP,
        row_to_json(NEW)::jsonb,
        'pending',
        prio
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Automatically enqueue POS transactions into sync_queue
DROP TRIGGER IF EXISTS trg_sync_pos_transactions ON pos_transactions;
CREATE TRIGGER trg_sync_pos_transactions
    AFTER INSERT ON pos_transactions
    FOR EACH ROW
    EXECUTE FUNCTION trg_enqueue_sync_outbox('pos_transactions');

-- Automatically enqueue POS transaction lines into sync_queue
DROP TRIGGER IF EXISTS trg_sync_pos_transaction_lines ON pos_transaction_lines;
CREATE TRIGGER trg_sync_pos_transaction_lines
    AFTER INSERT ON pos_transaction_lines
    FOR EACH ROW
    EXECUTE FUNCTION trg_enqueue_sync_outbox('pos_transaction_lines');

-- Automatically enqueue POS payments into sync_queue
DROP TRIGGER IF EXISTS trg_sync_pos_payments ON pos_payments;
CREATE TRIGGER trg_sync_pos_payments
    AFTER INSERT ON pos_payments
    FOR EACH ROW
    EXECUTE FUNCTION trg_enqueue_sync_outbox('pos_payments');

-- Automatically enqueue cashier sessions into sync_queue
DROP TRIGGER IF EXISTS trg_sync_cashier_sessions ON cashier_sessions;
CREATE TRIGGER trg_sync_cashier_sessions
    AFTER INSERT OR UPDATE ON cashier_sessions
    FOR EACH ROW
    EXECUTE FUNCTION trg_enqueue_sync_outbox('cashier_sessions');

-- Automatically enqueue sales orders into sync_queue
DROP TRIGGER IF EXISTS trg_sync_sales_orders_v2 ON sales_orders_v2;
CREATE TRIGGER trg_sync_sales_orders_v2
    AFTER INSERT OR UPDATE ON sales_orders_v2
    FOR EACH ROW
    EXECUTE FUNCTION trg_enqueue_sync_outbox('sales_orders_v2');

-- Automatically enqueue sales order lines into sync_queue
DROP TRIGGER IF EXISTS trg_sync_sales_order_lines_v2 ON sales_order_lines_v2;
CREATE TRIGGER trg_sync_sales_order_lines_v2
    AFTER INSERT OR UPDATE ON sales_order_lines_v2
    FOR EACH ROW
    EXECUTE FUNCTION trg_enqueue_sync_outbox('sales_order_lines_v2');

-- +goose Down
DROP TRIGGER IF EXISTS trg_sync_sales_order_lines_v2 ON sales_order_lines_v2;
DROP TRIGGER IF EXISTS trg_sync_sales_orders_v2 ON sales_orders_v2;
DROP TRIGGER IF EXISTS trg_sync_cashier_sessions ON cashier_sessions;
DROP TRIGGER IF EXISTS trg_sync_pos_payments ON pos_payments;
DROP TRIGGER IF EXISTS trg_sync_pos_transaction_lines ON pos_transaction_lines;
DROP TRIGGER IF EXISTS trg_sync_pos_transactions ON pos_transactions;
DROP FUNCTION IF EXISTS trg_enqueue_sync_outbox();
DROP TABLE IF EXISTS local_device_config;
DROP TABLE IF EXISTS sync_queue;
DROP TABLE IF EXISTS local_printer_configs;
