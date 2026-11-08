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
    outbox_uuid     UUID         NOT NULL DEFAULT gen_random_uuid(),
    entity_type     VARCHAR(50)  NOT NULL,
    entity_id       BIGINT       NOT NULL,
    action          VARCHAR(20)  NOT NULL, -- 'INSERT', 'UPDATE', 'DELETE', 'POS_CHECKOUT'
    payload         JSONB        NOT NULL,
    status          VARCHAR(20)  DEFAULT 'pending',
    priority        INTEGER      DEFAULT 0,         -- Higher = process first
    correlation_id  VARCHAR(100),                    -- Group related sync items
    max_retries     INTEGER      DEFAULT 5,
    retry_count     INTEGER      DEFAULT 0,
    last_error      TEXT,
    error_class     VARCHAR(32),
    next_attempt_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    quarantined_at  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  DEFAULT NOW(),
    synced_at       TIMESTAMPTZ
);

-- Index for efficient outbox draining (pending items ready for attempt, highest priority first)
CREATE INDEX IF NOT EXISTS idx_sync_queue_pending
    ON sync_queue(status, next_attempt_at ASC, priority DESC, id ASC)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_sync_queue_correlation
    ON sync_queue(correlation_id, id)
    WHERE status = 'pending';

-- Local POS terminal hardware identity with ZATCA sync tracking
CREATE TABLE IF NOT EXISTS local_device_config (
    id                  SERIAL       PRIMARY KEY,
    device_name         VARCHAR(100) NOT NULL,
    store_id            INTEGER      REFERENCES stores(id),
    pos_terminal_id     INTEGER      REFERENCES pos_terminals(id),
    sync_token          TEXT,                                 -- Per-terminal gRPC sync auth token (Phase 4)
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
    corr_id VARCHAR(100);
    existing_id BIGINT;
BEGIN
    -- Prevent echo-loop re-queuing when rows are applied from Cloud StreamPull/StreamChanges
    IF current_setting('nembus.sync_applier', true) = 'true' THEN
        RETURN NEW;
    END IF;

    entity_name := TG_ARGV[0];
    IF entity_name = 'pos_transactions' OR entity_name = 'pos_payments' OR entity_name = 'sales_returns' THEN
        prio := 10;
    END IF;

    corr_id := COALESCE(NULLIF(current_setting('nembus.correlation_id', true), ''), pg_current_xact_id()::text);

    -- Coalesce same-(correlation_id, entity_type, entity_id) pending rows within the same transaction
    SELECT id INTO existing_id
    FROM sync_queue
    WHERE correlation_id = corr_id
      AND entity_type = entity_name
      AND entity_id = NEW.id
      AND status = 'pending'
    ORDER BY id DESC
    LIMIT 1;

    IF existing_id IS NOT NULL THEN
        UPDATE sync_queue
        SET payload = row_to_json(NEW)::jsonb
        WHERE id = existing_id;
    ELSE
        INSERT INTO sync_queue (entity_type, entity_id, action, payload, status, priority, correlation_id)
        VALUES (
            entity_name,
            NEW.id,
            TG_OP,
            row_to_json(NEW)::jsonb,
            'pending',
            prio,
            corr_id
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Automatically enqueue user biometric templates into sync_queue
DROP TRIGGER IF EXISTS trg_sync_user_biometric_templates ON user_biometric_templates;
CREATE TRIGGER trg_sync_user_biometric_templates
    AFTER INSERT OR UPDATE ON user_biometric_templates
    FOR EACH ROW
    EXECUTE FUNCTION trg_enqueue_sync_outbox('user_biometric_templates');

-- Automatically enqueue POS transactions into sync_queue
DROP TRIGGER IF EXISTS trg_sync_pos_transactions ON pos_transactions;
CREATE TRIGGER trg_sync_pos_transactions
    AFTER INSERT OR UPDATE ON pos_transactions
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

-- Automatically enqueue sales returns into sync_queue
DROP TRIGGER IF EXISTS trg_sync_sales_returns ON sales_returns;
CREATE TRIGGER trg_sync_sales_returns
    AFTER INSERT OR UPDATE ON sales_returns
    FOR EACH ROW
    EXECUTE FUNCTION trg_enqueue_sync_outbox('sales_returns');

-- Automatically enqueue sales return lines into sync_queue
DROP TRIGGER IF EXISTS trg_sync_sales_return_lines ON sales_return_lines;
CREATE TRIGGER trg_sync_sales_return_lines
    AFTER INSERT ON sales_return_lines
    FOR EACH ROW
    EXECUTE FUNCTION trg_enqueue_sync_outbox('sales_return_lines');

-- Automatically enqueue stock movements with loop-filter into sync_queue
DROP TRIGGER IF EXISTS trg_sync_stock_movements ON stock_movements;
CREATE TRIGGER trg_sync_stock_movements
    AFTER INSERT ON stock_movements
    FOR EACH ROW
    WHEN (NEW.movement_type IN ('stock_adjustment', 'adjustment', 'waste', 'damage', 'physical_count'))
    EXECUTE FUNCTION trg_enqueue_sync_outbox('stock_movements');

-- +goose Down
DROP TRIGGER IF EXISTS trg_sync_user_biometric_templates ON user_biometric_templates;
DROP TRIGGER IF EXISTS trg_sync_stock_movements ON stock_movements;
DROP TRIGGER IF EXISTS trg_sync_sales_return_lines ON sales_return_lines;
DROP TRIGGER IF EXISTS trg_sync_sales_returns ON sales_returns;
DROP TRIGGER IF EXISTS trg_sync_cashier_sessions ON cashier_sessions;
DROP TRIGGER IF EXISTS trg_sync_pos_payments ON pos_payments;
DROP TRIGGER IF EXISTS trg_sync_pos_transaction_lines ON pos_transaction_lines;
DROP TRIGGER IF EXISTS trg_sync_pos_transactions ON pos_transactions;
DROP FUNCTION IF EXISTS trg_enqueue_sync_outbox();
DROP TABLE IF EXISTS local_device_config;
DROP TABLE IF EXISTS sync_queue;
DROP TABLE IF EXISTS local_printer_configs;
