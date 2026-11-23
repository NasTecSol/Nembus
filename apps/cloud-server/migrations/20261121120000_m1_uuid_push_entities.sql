-- Versioned UUID upgrade for existing databases. Keep in sync with the core and POS migration copies.
-- @nembus:m1_uuid_push_entities_begin
-- M1 UUID push entities upgrade for existing databases.
-- =====================================================================

CREATE OR REPLACE FUNCTION pg_temp.det_uuid(tbl text, old_id bigint)
RETURNS uuid
LANGUAGE plpgsql
IMMUTABLE
STRICT
AS $$
DECLARE
    h text;
BEGIN
    h := md5('nembus:v1:' || tbl || ':' || old_id::text);
    RETURN (
        substr(h, 1, 8) || '-' ||
        substr(h, 9, 4) || '-' ||
        '4' || substr(h, 14, 3) || '-' ||
        substr('89ab', (get_byte(decode(substr(h, 17, 2), 'hex'), 0) & 3) + 1, 1) || substr(h, 18, 3) || '-' ||
        substr(h, 21, 12)
    )::uuid;
END;
$$;

-- Some restored tenant schemas already have UUID columns while others still
-- hold legacy integer IDs. Preserve UUID values while the migration converts
-- the remaining integer columns.
CREATE OR REPLACE FUNCTION pg_temp.det_uuid(tbl text, existing_id uuid)
RETURNS uuid
LANGUAGE sql
IMMUTABLE
STRICT
AS $$
    SELECT existing_id;
$$;

CREATE OR REPLACE FUNCTION pg_temp.det_uuid(tbl text, legacy_id text)
RETURNS uuid
LANGUAGE plpgsql
IMMUTABLE
STRICT
AS $$
BEGIN
    IF legacy_id ~ '^[0-9]+$' THEN
        RETURN pg_temp.det_uuid(tbl, legacy_id::bigint);
    END IF;
    RETURN legacy_id::uuid;
END;
$$;

-- 1. Drop dependent views before altering column types (SQLSTATE 2BP01)
DROP VIEW IF EXISTS
    public.vw_customer_aging_report,
    public.vw_profit_margin_analysis,
    public.vw_active_restaurant_orders,
    public.vw_waste_daily_summary,
    public.vw_realtime_pnl,
    public.vw_cashier_shift_reconciliation,
    public.vw_stock_movement_ledger,
    public.vw_tax_vat_summary
CASCADE;

-- 2. Drop functions whose parameter or RETURNS TABLE types change
DROP FUNCTION IF EXISTS public.fn_process_stock_transfer(integer, integer, integer, integer, numeric, integer, integer, character varying, integer, text);
DROP FUNCTION IF EXISTS public.fn_reconcile_stock_count(integer);
DROP FUNCTION IF EXISTS public.fn_calculate_loyalty_earned(integer);
DROP FUNCTION IF EXISTS public.fn_get_kds_orders(integer, character varying[]);
DROP FUNCTION IF EXISTS public.fn_get_returnable_quantity(integer);

-- 3. Dynamically drop ALL foreign keys referencing the 15 migrated tables
--    (handles legacy constraint name divergences in 20260918111505_initial_schema.sql)
DO $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN
        SELECT c.conrelid::regclass AS tbl_name, c.conname
        FROM pg_constraint c
        WHERE c.contype = 'f'
          AND c.connamespace = 'public'::regnamespace
          AND c.confrelid = ANY(ARRAY[
              'public.user_biometric_templates'::regclass,
              'public.cashier_sessions'::regclass,
              'public.stock_movements'::regclass,
              'public.stock_counts'::regclass,
              'public.stock_count_lines'::regclass,
              'public.customers'::regclass,
              'public.pos_transactions'::regclass,
              'public.pos_transaction_lines'::regclass,
              'public.pos_payments'::regclass,
              'public.restaurant_orders'::regclass,
              'public.restaurant_order_items'::regclass,
              'public.waste_logs'::regclass,
              'public.kiosk_sessions'::regclass,
              'public.sales_returns'::regclass,
              'public.sales_return_lines'::regclass
          ])
    LOOP
        EXECUTE format('ALTER TABLE %s DROP CONSTRAINT IF EXISTS %I', r.tbl_name, r.conname);
    END LOOP;
END $$;

-- 4. Convert all 15 PKs in-place (preserves indexes & avoids firing UPDATE triggers)
ALTER TABLE public.user_biometric_templates
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN id TYPE uuid USING pg_temp.det_uuid('user_biometric_templates', id),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS public.user_biometric_templates_id_seq;

ALTER TABLE public.cashier_sessions
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN id TYPE uuid USING pg_temp.det_uuid('cashier_sessions', id),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS public.cashier_sessions_id_seq;

ALTER TABLE public.stock_movements
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN id TYPE uuid USING pg_temp.det_uuid('stock_movements', id),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS public.stock_movements_id_seq;

ALTER TABLE public.stock_counts
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN id TYPE uuid USING pg_temp.det_uuid('stock_counts', id),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS public.stock_counts_id_seq;

ALTER TABLE public.stock_count_lines
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN id TYPE uuid USING pg_temp.det_uuid('stock_count_lines', id),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS public.stock_count_lines_id_seq;

ALTER TABLE public.customers
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN id TYPE uuid USING pg_temp.det_uuid('customers', id),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS public.customers_id_seq;

ALTER TABLE public.pos_transactions
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN id TYPE uuid USING pg_temp.det_uuid('pos_transactions', id),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS public.pos_transactions_id_seq;

ALTER TABLE public.pos_transaction_lines
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN id TYPE uuid USING pg_temp.det_uuid('pos_transaction_lines', id),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS public.pos_transaction_lines_id_seq;

ALTER TABLE public.pos_payments
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN id TYPE uuid USING pg_temp.det_uuid('pos_payments', id),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS public.pos_payments_id_seq;

ALTER TABLE public.restaurant_orders
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN id TYPE uuid USING pg_temp.det_uuid('restaurant_orders', id),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS public.restaurant_orders_id_seq;

ALTER TABLE public.restaurant_order_items
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN id TYPE uuid USING pg_temp.det_uuid('restaurant_order_items', id),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS public.restaurant_order_items_id_seq;

ALTER TABLE public.waste_logs
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN id TYPE uuid USING pg_temp.det_uuid('waste_logs', id),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS public.waste_logs_id_seq;

ALTER TABLE public.kiosk_sessions
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN id TYPE uuid USING pg_temp.det_uuid('kiosk_sessions', id),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS public.kiosk_sessions_id_seq;

ALTER TABLE public.sales_returns
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN id TYPE uuid USING pg_temp.det_uuid('sales_returns', id),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS public.sales_returns_id_seq;

ALTER TABLE public.sales_return_lines
    ALTER COLUMN id DROP DEFAULT,
    ALTER COLUMN id TYPE uuid USING pg_temp.det_uuid('sales_return_lines', id),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS public.sales_return_lines_id_seq;

-- 5. Convert all 29 referencing FK columns in-place
ALTER TABLE public.stock_count_lines
    ALTER COLUMN stock_count_id TYPE uuid USING pg_temp.det_uuid('stock_counts', stock_count_id);

ALTER TABLE public.sales_orders
    ALTER COLUMN customer_id TYPE uuid USING pg_temp.det_uuid('customers', customer_id);

ALTER TABLE public.carts
    ALTER COLUMN customer_id TYPE uuid USING pg_temp.det_uuid('customers', customer_id);

ALTER TABLE public.draft_cart_templates
    ALTER COLUMN customer_id TYPE uuid USING pg_temp.det_uuid('customers', customer_id);

ALTER TABLE public.sales_orders_v2
    ALTER COLUMN customer_id TYPE uuid USING pg_temp.det_uuid('customers', customer_id);

ALTER TABLE public.pos_transactions
    ALTER COLUMN cashier_session_id TYPE uuid USING pg_temp.det_uuid('cashier_sessions', cashier_session_id),
    ALTER COLUMN customer_id TYPE uuid USING pg_temp.det_uuid('customers', customer_id);

ALTER TABLE public.pos_transaction_lines
    ALTER COLUMN transaction_id TYPE uuid USING pg_temp.det_uuid('pos_transactions', transaction_id);

ALTER TABLE public.pos_payments
    ALTER COLUMN transaction_id TYPE uuid USING pg_temp.det_uuid('pos_transactions', transaction_id);

ALTER TABLE public.restaurant_carts
    ALTER COLUMN customer_id TYPE uuid USING pg_temp.det_uuid('customers', customer_id),
    ALTER COLUMN kiosk_session_id TYPE uuid USING pg_temp.det_uuid('kiosk_sessions', kiosk_session_id),
    ALTER COLUMN converted_to_order_id TYPE uuid USING pg_temp.det_uuid('restaurant_orders', converted_to_order_id);

ALTER TABLE public.restaurant_orders
    ALTER COLUMN cashier_session_id TYPE uuid USING pg_temp.det_uuid('cashier_sessions', cashier_session_id),
    ALTER COLUMN customer_id TYPE uuid USING pg_temp.det_uuid('customers', customer_id),
    ALTER COLUMN kiosk_session_id TYPE uuid USING pg_temp.det_uuid('kiosk_sessions', kiosk_session_id),
    ALTER COLUMN pos_transaction_id TYPE uuid USING pg_temp.det_uuid('pos_transactions', pos_transaction_id);

ALTER TABLE public.restaurant_order_items
    ALTER COLUMN order_id TYPE uuid USING pg_temp.det_uuid('restaurant_orders', order_id);

ALTER TABLE public.restaurant_order_status_history
    ALTER COLUMN restaurant_order_id TYPE uuid USING pg_temp.det_uuid('restaurant_orders', restaurant_order_id);

ALTER TABLE public.waste_logs
    ALTER COLUMN order_id TYPE uuid USING pg_temp.det_uuid('restaurant_orders', order_id);

ALTER TABLE public.sales_analytics
    ALTER COLUMN customer_id TYPE uuid USING pg_temp.det_uuid('customers', customer_id);

ALTER TABLE public.invoices
    ALTER COLUMN customer_id TYPE uuid USING pg_temp.det_uuid('customers', customer_id);

ALTER TABLE public.sales_returns
    ALTER COLUMN cashier_session_id TYPE uuid USING pg_temp.det_uuid('cashier_sessions', cashier_session_id),
    ALTER COLUMN customer_id TYPE uuid USING pg_temp.det_uuid('customers', customer_id),
    ALTER COLUMN original_transaction_id TYPE uuid USING pg_temp.det_uuid('pos_transactions', original_transaction_id);

ALTER TABLE public.quotes
    ALTER COLUMN customer_id TYPE uuid USING pg_temp.det_uuid('customers', customer_id);

ALTER TABLE public.sales_return_lines
    ALTER COLUMN return_id TYPE uuid USING pg_temp.det_uuid('sales_returns', return_id),
    ALTER COLUMN original_line_id TYPE uuid USING pg_temp.det_uuid('pos_transaction_lines', original_line_id);

ALTER TABLE public.customer_addresses
    ALTER COLUMN customer_id TYPE uuid USING pg_temp.det_uuid('customers', customer_id);

ALTER TABLE public.customer_payments
    ALTER COLUMN customer_id TYPE uuid USING pg_temp.det_uuid('customers', customer_id);

-- 6. Polymorphic references: stock_movements.reference_id, zatca_document_chain.entity_id, stock_reservations.reference_id
ALTER TABLE public.stock_movements
    ALTER COLUMN reference_id TYPE text
    USING CASE
        WHEN reference_id IS NULL THEN NULL
        WHEN reference_type IN ('stock_count', 'stock_counts') THEN pg_temp.det_uuid('stock_counts', reference_id)::text
        WHEN reference_type IN ('pos_transaction', 'pos_transactions') THEN pg_temp.det_uuid('pos_transactions', reference_id)::text
        WHEN reference_type IN ('sales_return', 'sales_returns') THEN pg_temp.det_uuid('sales_returns', reference_id)::text
        WHEN reference_type IN ('restaurant_order', 'restaurant_orders') THEN pg_temp.det_uuid('restaurant_orders', reference_id)::text
        ELSE reference_id::text
    END;

UPDATE public.zatca_document_chain
SET entity_id = pg_temp.det_uuid(
    CASE
        WHEN entity_type IN ('pos_transaction', 'pos_transactions') THEN 'pos_transactions'
        WHEN entity_type IN ('sales_return', 'sales_returns') THEN 'sales_returns'
    END,
    entity_id::bigint
)::text
WHERE entity_type IN ('pos_transaction', 'pos_transactions', 'sales_return', 'sales_returns')
  AND entity_id ~ '^[0-9]+$';

UPDATE public.stock_reservations
SET reference_id = pg_temp.det_uuid('pos_transactions', reference_id::bigint)::text
WHERE reference_type IN ('pos_transaction', 'pos_transactions')
  AND reference_id ~ '^[0-9]+$';

-- 7. Sync metadata tables: sync_inbox & change_log (+ change_log.payload for 5 pull-synced entities)
ALTER TABLE public.sync_inbox
    ALTER COLUMN entity_id TYPE text
    USING CASE
        WHEN entity_type IN (
            'user_biometric_templates', 'cashier_sessions', 'stock_movements',
            'stock_counts', 'stock_count_lines', 'customers',
            'pos_transactions', 'pos_transaction_lines', 'pos_payments',
            'restaurant_orders', 'restaurant_order_items', 'waste_logs',
            'kiosk_sessions', 'sales_returns', 'sales_return_lines', 'carts'
        ) THEN pg_temp.det_uuid(entity_type, entity_id)::text
        WHEN entity_type = 'pos_transaction' THEN pg_temp.det_uuid('pos_transactions', entity_id)::text
        WHEN entity_type = 'sales_return' THEN pg_temp.det_uuid('sales_returns', entity_id)::text
        ELSE entity_id::text
    END;

ALTER TABLE public.change_log
    ALTER COLUMN entity_id TYPE text
    USING CASE
        WHEN entity_type IN (
            'user_biometric_templates', 'cashier_sessions', 'stock_movements',
            'stock_counts', 'stock_count_lines', 'customers',
            'pos_transactions', 'pos_transaction_lines', 'pos_payments',
            'restaurant_orders', 'restaurant_order_items', 'waste_logs',
            'kiosk_sessions', 'sales_returns', 'sales_return_lines', 'carts'
        ) THEN pg_temp.det_uuid(entity_type, entity_id)::text
        WHEN entity_type = 'pos_transaction' THEN pg_temp.det_uuid('pos_transactions', entity_id)::text
        WHEN entity_type = 'sales_return' THEN pg_temp.det_uuid('sales_returns', entity_id)::text
        ELSE entity_id::text
    END;

UPDATE public.change_log
SET payload = payload
    || CASE
        WHEN (payload->>'id') ~ '^[0-9]+$'
            THEN jsonb_build_object('id', pg_temp.det_uuid(entity_type, (payload->>'id')::bigint)::text)
        ELSE '{}'::jsonb
       END
    || CASE
        WHEN entity_type = 'stock_count_lines' AND (payload->>'stock_count_id') ~ '^[0-9]+$'
            THEN jsonb_build_object('stock_count_id', pg_temp.det_uuid('stock_counts', (payload->>'stock_count_id')::bigint)::text)
        ELSE '{}'::jsonb
       END
    || CASE
        WHEN entity_type = 'stock_movements' AND (payload->>'reference_id') ~ '^[0-9]+$' THEN
            jsonb_build_object('reference_id',
                CASE
                    WHEN (payload->>'reference_type') IN ('stock_count', 'stock_counts')
                        THEN pg_temp.det_uuid('stock_counts', (payload->>'reference_id')::bigint)::text
                    WHEN (payload->>'reference_type') IN ('pos_transaction', 'pos_transactions')
                        THEN pg_temp.det_uuid('pos_transactions', (payload->>'reference_id')::bigint)::text
                    WHEN (payload->>'reference_type') IN ('sales_return', 'sales_returns')
                        THEN pg_temp.det_uuid('sales_returns', (payload->>'reference_id')::bigint)::text
                    WHEN (payload->>'reference_type') IN ('restaurant_order', 'restaurant_orders')
                        THEN pg_temp.det_uuid('restaurant_orders', (payload->>'reference_id')::bigint)::text
                    ELSE (payload->>'reference_id')
                END
            )
        ELSE '{}'::jsonb
       END
WHERE payload IS NOT NULL
  AND entity_type IN ('customers', 'stock_counts', 'stock_count_lines', 'stock_movements', 'user_biometric_templates');

-- 8. POS-only outbox (sync_queue & trg_enqueue_sync_outbox) when upgrading an existing POS DB
DO $$
BEGIN
    IF to_regclass('public.sync_queue') IS NOT NULL THEN
        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'sync_queue'
              AND column_name = 'entity_id' AND data_type IN ('integer', 'bigint')
        ) THEN
            EXECUTE $trg$
                ALTER TABLE public.sync_queue
                    ALTER COLUMN entity_id TYPE text
                    USING CASE
                        WHEN entity_type IN (
                            'user_biometric_templates', 'cashier_sessions', 'stock_movements',
                            'stock_counts', 'stock_count_lines', 'customers',
                            'pos_transactions', 'pos_transaction_lines', 'pos_payments',
                            'restaurant_orders', 'restaurant_order_items', 'waste_logs',
                            'kiosk_sessions', 'sales_returns', 'sales_return_lines', 'carts'
                        ) THEN pg_temp.det_uuid(entity_type, entity_id)::text
                        ELSE entity_id::text
                    END
            $trg$;
        END IF;

        UPDATE public.sync_queue
        SET payload = payload
            || CASE WHEN (payload->>'id') ~ '^[0-9]+$'
                    THEN jsonb_build_object('id', pg_temp.det_uuid(entity_type, (payload->>'id')::bigint)::text)
                    ELSE '{}'::jsonb END
            || CASE WHEN (payload->>'cashier_session_id') ~ '^[0-9]+$'
                    THEN jsonb_build_object('cashier_session_id', pg_temp.det_uuid('cashier_sessions', (payload->>'cashier_session_id')::bigint)::text)
                    ELSE '{}'::jsonb END
            || CASE WHEN (payload->>'customer_id') ~ '^[0-9]+$'
                    THEN jsonb_build_object('customer_id', pg_temp.det_uuid('customers', (payload->>'customer_id')::bigint)::text)
                    ELSE '{}'::jsonb END
            || CASE WHEN (payload->>'transaction_id') ~ '^[0-9]+$'
                    THEN jsonb_build_object('transaction_id', pg_temp.det_uuid('pos_transactions', (payload->>'transaction_id')::bigint)::text)
                    ELSE '{}'::jsonb END
            || CASE WHEN (payload->>'stock_count_id') ~ '^[0-9]+$'
                    THEN jsonb_build_object('stock_count_id', pg_temp.det_uuid('stock_counts', (payload->>'stock_count_id')::bigint)::text)
                    ELSE '{}'::jsonb END
            || CASE WHEN (payload->>'order_id') ~ '^[0-9]+$'
                    THEN jsonb_build_object('order_id', pg_temp.det_uuid('restaurant_orders', (payload->>'order_id')::bigint)::text)
                    ELSE '{}'::jsonb END
            || CASE WHEN (payload->>'kiosk_session_id') ~ '^[0-9]+$'
                    THEN jsonb_build_object('kiosk_session_id', pg_temp.det_uuid('kiosk_sessions', (payload->>'kiosk_session_id')::bigint)::text)
                    ELSE '{}'::jsonb END
            || CASE WHEN (payload->>'pos_transaction_id') ~ '^[0-9]+$'
                    THEN jsonb_build_object('pos_transaction_id', pg_temp.det_uuid('pos_transactions', (payload->>'pos_transaction_id')::bigint)::text)
                    ELSE '{}'::jsonb END
            || CASE WHEN (payload->>'return_id') ~ '^[0-9]+$'
                    THEN jsonb_build_object('return_id', pg_temp.det_uuid('sales_returns', (payload->>'return_id')::bigint)::text)
                    ELSE '{}'::jsonb END
            || CASE WHEN (payload->>'original_transaction_id') ~ '^[0-9]+$'
                    THEN jsonb_build_object('original_transaction_id', pg_temp.det_uuid('pos_transactions', (payload->>'original_transaction_id')::bigint)::text)
                    ELSE '{}'::jsonb END
            || CASE WHEN (payload->>'original_line_id') ~ '^[0-9]+$'
                    THEN jsonb_build_object('original_line_id', pg_temp.det_uuid('pos_transaction_lines', (payload->>'original_line_id')::bigint)::text)
                    ELSE '{}'::jsonb END
            || CASE WHEN entity_type = 'stock_movements' AND (payload->>'reference_id') ~ '^[0-9]+$' THEN
                    jsonb_build_object('reference_id',
                        CASE
                            WHEN (payload->>'reference_type') IN ('stock_count', 'stock_counts')
                                THEN pg_temp.det_uuid('stock_counts', (payload->>'reference_id')::bigint)::text
                            WHEN (payload->>'reference_type') IN ('pos_transaction', 'pos_transactions')
                                THEN pg_temp.det_uuid('pos_transactions', (payload->>'reference_id')::bigint)::text
                            WHEN (payload->>'reference_type') IN ('sales_return', 'sales_returns')
                                THEN pg_temp.det_uuid('sales_returns', (payload->>'reference_id')::bigint)::text
                            WHEN (payload->>'reference_type') IN ('restaurant_order', 'restaurant_orders')
                                THEN pg_temp.det_uuid('restaurant_orders', (payload->>'reference_id')::bigint)::text
                            ELSE (payload->>'reference_id')
                        END
                    )
                    ELSE '{}'::jsonb END
        WHERE status IN ('pending', 'failed', 'quarantined')
          AND payload IS NOT NULL
          AND entity_type IN (
              'user_biometric_templates', 'cashier_sessions', 'stock_movements',
              'stock_counts', 'stock_count_lines', 'customers', 'carts',
              'pos_transactions', 'pos_transaction_lines', 'pos_payments',
              'restaurant_orders', 'restaurant_order_items', 'waste_logs',
              'kiosk_sessions', 'sales_returns', 'sales_return_lines'
          );

        EXECUTE $trg$
            CREATE OR REPLACE FUNCTION public.trg_enqueue_sync_outbox()
            RETURNS TRIGGER
            LANGUAGE plpgsql
            AS $fn$
            DECLARE
                entity_name VARCHAR(50);
                prio INTEGER := 5;
                corr_id VARCHAR(100);
                existing_id BIGINT;
            BEGIN
                IF current_setting('nembus.sync_applier', true) = 'true' THEN
                    RETURN NEW;
                END IF;

                entity_name := TG_ARGV[0];
                IF entity_name = 'pos_transactions' OR entity_name = 'pos_payments' OR entity_name = 'sales_returns' THEN
                    prio := 10;
                END IF;

                corr_id := COALESCE(NULLIF(current_setting('nembus.correlation_id', true), ''), pg_current_xact_id()::text);

                SELECT id INTO existing_id
                FROM sync_queue
                WHERE correlation_id = corr_id
                  AND entity_type = entity_name
                  AND entity_id = NEW.id::text
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
                        NEW.id::text,
                        TG_OP,
                        row_to_json(NEW)::jsonb,
                        'pending',
                        prio,
                        corr_id
                    );
                END IF;
                RETURN NEW;
            END;
            $fn$;
        $trg$;

        EXECUTE $trg$
            DROP TRIGGER IF EXISTS trg_sync_stock_movements ON public.stock_movements;
            CREATE TRIGGER trg_sync_stock_movements
                AFTER INSERT ON public.stock_movements
                FOR EACH ROW
                WHEN (NEW.movement_type IN ('stock_adjustment', 'adjustment', 'waste', 'damage', 'physical_count', 'recipe_deduction', 'recipe_reversal'))
                EXECUTE FUNCTION public.trg_enqueue_sync_outbox('stock_movements');
        $trg$;
    END IF;
END $$;

-- 9. Re-add all 29 canonical FK constraints with exact schema names and ON DELETE actions
ALTER TABLE public.stock_count_lines
    ADD CONSTRAINT stock_count_lines_stock_count_id_fkey
    FOREIGN KEY (stock_count_id) REFERENCES public.stock_counts(id) ON DELETE CASCADE;

ALTER TABLE public.sales_orders
    ADD CONSTRAINT sales_orders_customer_id_fkey
    FOREIGN KEY (customer_id) REFERENCES public.customers(id) ON DELETE SET NULL;

ALTER TABLE public.carts
    ADD CONSTRAINT carts_customer_id_fkey
    FOREIGN KEY (customer_id) REFERENCES public.customers(id) ON DELETE SET NULL;

ALTER TABLE public.draft_cart_templates
    ADD CONSTRAINT draft_cart_templates_customer_id_fkey
    FOREIGN KEY (customer_id) REFERENCES public.customers(id) ON DELETE CASCADE;

ALTER TABLE public.sales_orders_v2
    ADD CONSTRAINT sales_orders_v2_customer_id_fkey
    FOREIGN KEY (customer_id) REFERENCES public.customers(id) ON DELETE SET NULL;

ALTER TABLE public.pos_transactions
    ADD CONSTRAINT pos_transactions_cashier_session_id_fkey
    FOREIGN KEY (cashier_session_id) REFERENCES public.cashier_sessions(id) ON DELETE CASCADE,
    ADD CONSTRAINT pos_transactions_customer_id_fkey
    FOREIGN KEY (customer_id) REFERENCES public.customers(id) ON DELETE SET NULL;

ALTER TABLE public.pos_transaction_lines
    ADD CONSTRAINT pos_transaction_lines_transaction_id_fkey
    FOREIGN KEY (transaction_id) REFERENCES public.pos_transactions(id) ON DELETE CASCADE;

ALTER TABLE public.pos_payments
    ADD CONSTRAINT pos_payments_transaction_id_fkey
    FOREIGN KEY (transaction_id) REFERENCES public.pos_transactions(id) ON DELETE CASCADE;

ALTER TABLE public.restaurant_carts
    ADD CONSTRAINT restaurant_carts_customer_id_fkey
    FOREIGN KEY (customer_id) REFERENCES public.customers(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_restaurant_carts_kiosk_session
    FOREIGN KEY (kiosk_session_id) REFERENCES public.kiosk_sessions(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_restaurant_carts_converted_to_order
    FOREIGN KEY (converted_to_order_id) REFERENCES public.restaurant_orders(id) ON DELETE SET NULL;

ALTER TABLE public.restaurant_orders
    ADD CONSTRAINT restaurant_orders_cashier_session_id_fkey
    FOREIGN KEY (cashier_session_id) REFERENCES public.cashier_sessions(id) ON DELETE SET NULL,
    ADD CONSTRAINT restaurant_orders_customer_id_fkey
    FOREIGN KEY (customer_id) REFERENCES public.customers(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_restaurant_orders_kiosk_session
    FOREIGN KEY (kiosk_session_id) REFERENCES public.kiosk_sessions(id) ON DELETE SET NULL,
    ADD CONSTRAINT restaurant_orders_pos_transaction_id_fkey
    FOREIGN KEY (pos_transaction_id) REFERENCES public.pos_transactions(id) ON DELETE SET NULL;

ALTER TABLE public.restaurant_order_items
    ADD CONSTRAINT restaurant_order_items_order_id_fkey
    FOREIGN KEY (order_id) REFERENCES public.restaurant_orders(id) ON DELETE CASCADE;

ALTER TABLE public.restaurant_order_status_history
    ADD CONSTRAINT restaurant_order_status_history_restaurant_order_id_fkey
    FOREIGN KEY (restaurant_order_id) REFERENCES public.restaurant_orders(id) ON DELETE CASCADE;

ALTER TABLE public.waste_logs
    ADD CONSTRAINT waste_logs_order_id_fkey
    FOREIGN KEY (order_id) REFERENCES public.restaurant_orders(id) ON DELETE SET NULL;

ALTER TABLE public.sales_analytics
    ADD CONSTRAINT fk_sales_analytics_customer
    FOREIGN KEY (customer_id) REFERENCES public.customers(id) ON DELETE SET NULL;

ALTER TABLE public.invoices
    ADD CONSTRAINT invoices_customer_id_fkey
    FOREIGN KEY (customer_id) REFERENCES public.customers(id) ON DELETE SET NULL;

ALTER TABLE public.sales_returns
    ADD CONSTRAINT sales_returns_cashier_session_id_fkey
    FOREIGN KEY (cashier_session_id) REFERENCES public.cashier_sessions(id) ON DELETE SET NULL,
    ADD CONSTRAINT sales_returns_customer_id_fkey
    FOREIGN KEY (customer_id) REFERENCES public.customers(id) ON DELETE SET NULL,
    ADD CONSTRAINT sales_returns_original_transaction_id_fkey
    FOREIGN KEY (original_transaction_id) REFERENCES public.pos_transactions(id) ON DELETE SET NULL;

ALTER TABLE public.quotes
    ADD CONSTRAINT quotes_customer_id_fkey
    FOREIGN KEY (customer_id) REFERENCES public.customers(id) ON DELETE SET NULL;

ALTER TABLE public.sales_return_lines
    ADD CONSTRAINT sales_return_lines_return_id_fkey
    FOREIGN KEY (return_id) REFERENCES public.sales_returns(id) ON DELETE CASCADE,
    ADD CONSTRAINT sales_return_lines_original_line_id_fkey
    FOREIGN KEY (original_line_id) REFERENCES public.pos_transaction_lines(id) ON DELETE SET NULL;

ALTER TABLE public.customer_addresses
    ADD CONSTRAINT customer_addresses_customer_id_fkey
    FOREIGN KEY (customer_id) REFERENCES public.customers(id) ON DELETE CASCADE;

ALTER TABLE public.customer_payments
    ADD CONSTRAINT customer_payments_customer_id_fkey
    FOREIGN KEY (customer_id) REFERENCES public.customers(id) ON DELETE SET NULL;

-- 10. Recreate the 8 dependent views
CREATE OR REPLACE VIEW public.vw_customer_aging_report AS
SELECT
    c.id                    AS customer_id,
    c.customer_code,
    c.name                  AS customer_name,
    c.email,
    c.phone,
    c.customer_type,
    c.credit_limit,
    c.outstanding_balance,
    i.organization_id,
    COALESCE(SUM(CASE WHEN i.due_date >= CURRENT_DATE THEN i.balance_due ELSE 0 END), 0) AS current_amount,
    COALESCE(SUM(CASE WHEN i.due_date < CURRENT_DATE AND CURRENT_DATE - i.due_date <= 30 THEN i.balance_due ELSE 0 END), 0) AS overdue_1_30,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - i.due_date BETWEEN 31 AND 60 THEN i.balance_due ELSE 0 END), 0) AS overdue_31_60,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - i.due_date BETWEEN 61 AND 90 THEN i.balance_due ELSE 0 END), 0) AS overdue_61_90,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - i.due_date > 90 THEN i.balance_due ELSE 0 END), 0) AS overdue_over_90,
    COALESCE(SUM(CASE WHEN i.balance_due > 0 THEN i.balance_due ELSE 0 END), 0) AS total_outstanding,
    COUNT(CASE WHEN i.invoice_status = 'overdue' THEN 1 END)::INTEGER AS overdue_invoice_count,
    MAX(i.due_date)         AS latest_due_date,
    c.loyalty_points
FROM customers c
LEFT JOIN invoices i
    ON i.customer_id = c.id
    AND i.invoice_status <> 'cancelled' AND i.invoice_status <> 'draft'
    AND i.balance_due > 0
LEFT JOIN organizations o
    ON o.id = i.organization_id
WHERE c.is_active = true
GROUP BY
    c.id, c.customer_code, c.name, c.email, c.phone,
    c.customer_type, c.credit_limit, c.outstanding_balance,
    c.loyalty_points, i.organization_id
ORDER BY total_outstanding DESC;

CREATE OR REPLACE VIEW public.vw_profit_margin_analysis AS
SELECT
    ptl.product_id,
    ptl.product_variant_id,
    p.sku,
    p.name AS product_name,
    pc.name AS category_name,
    pt.store_id,
    s.name AS store_name,
    DATE_TRUNC('month', pt.transaction_date)::DATE AS period_month,
    SUM(ptl.quantity) AS units_sold,
    SUM(ptl.line_total) AS total_revenue,
    SUM(ptl.cost_price * ptl.quantity) AS total_cost,
    SUM(ptl.line_total) - SUM(ptl.cost_price * ptl.quantity) AS gross_profit,
    CASE
        WHEN SUM(ptl.line_total) > 0
        THEN ROUND(
            (SUM(ptl.line_total) - SUM(ptl.cost_price * ptl.quantity))
            / SUM(ptl.line_total) * 100, 2
        )
        ELSE 0
    END AS gross_margin_pct,
    SUM(ptl.discount_amount) AS total_discounts,
    SUM(ptl.tax_amount) AS total_taxes
FROM pos_transaction_lines ptl
JOIN pos_transactions pt
    ON pt.id = ptl.transaction_id
    AND pt.status = 'completed'
JOIN products p
    ON p.id = ptl.product_id
LEFT JOIN product_variants pv_id
    ON pv_id.id = ptl.product_variant_id
LEFT JOIN product_categories pc
    ON pc.id = p.category_id
JOIN stores s
    ON s.id = pt.store_id
GROUP BY
    ptl.product_id,
    ptl.product_variant_id,
    p.sku,
    p.name,
    pc.name,
    pt.store_id,
    s.name,
    DATE_TRUNC('month', pt.transaction_date)
ORDER BY
    period_month DESC,
    gross_profit DESC;

CREATE OR REPLACE VIEW public.vw_active_restaurant_orders AS
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
    COALESCE(NULLIF(TRIM(COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '')), ''), '')::text AS waiter_name,
    ro.customer_id,
    cust.name                   AS customer_name,
    COALESCE(ROUND(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - ro.ordered_at)) / 60.0), 0)::INTEGER AS minutes_since_ordered
FROM restaurant_orders ro
LEFT JOIN restaurant_tables rt  ON ro.table_id = rt.id
LEFT JOIN cashiers c            ON ro.cashier_id = c.id
LEFT JOIN users u               ON c.user_id = u.id
LEFT JOIN customers cust        ON ro.customer_id = cust.id
WHERE ro.status <> 'paid' AND ro.status <> 'voided';

CREATE OR REPLACE VIEW public.vw_waste_daily_summary AS
SELECT
    wl.store_id,
    DATE(wl.wasted_at)          AS waste_date,
    wl.waste_source,
    COUNT(*)                    AS waste_entries,
    SUM(wl.quantity)            AS total_quantity_wasted,
    SUM(wl.total_cost)          AS total_cost_wasted,
    AVG(wl.total_cost)          AS avg_cost_per_entry
FROM waste_logs wl
GROUP BY wl.store_id, DATE(wl.wasted_at), wl.waste_source;

CREATE OR REPLACE VIEW public.vw_realtime_pnl AS
WITH daily_pos_sales AS (
    SELECT
        pt.store_id,
        s.organization_id,
        s.name AS store_name,
        pt.transaction_date::DATE AS period_date,
        COUNT(DISTINCT pt.id) AS transaction_count,
        COALESCE(SUM(ptl.line_total), 0) AS gross_sales,
        COALESCE(SUM(ptl.discount_amount), 0) AS total_discounts,
        COALESCE(SUM(ptl.tax_amount), 0) AS total_tax,
        COALESCE(SUM(ptl.cost_price * ptl.quantity), 0) AS total_cogs
    FROM pos_transactions pt
    JOIN stores s ON s.id = pt.store_id
    JOIN pos_transaction_lines ptl ON ptl.transaction_id = pt.id
    WHERE pt.status = 'completed'
    GROUP BY pt.store_id, s.organization_id, s.name, pt.transaction_date::DATE
),
daily_returns AS (
    SELECT
        sr.store_id,
        sr.return_date::DATE AS period_date,
        COUNT(DISTINCT sr.id) AS return_count,
        COALESCE(SUM(sr.total_refund_amount), 0) AS total_returns
    FROM sales_returns sr
    WHERE sr.status = 'completed'
    GROUP BY sr.store_id, sr.return_date::DATE
),
daily_invoices AS (
    SELECT
        inv.store_id,
        inv.organization_id,
        s.name AS store_name,
        inv.invoice_date AS period_date,
        COUNT(DISTINCT inv.id) AS invoice_count,
        COALESCE(SUM(inv.total_amount), 0) AS invoice_gross,
        COALESCE(SUM(inv.tax_amount), 0) AS invoice_tax,
        COALESCE(SUM(il_cost.total_cost), 0) AS invoice_cogs
    FROM invoices inv
    JOIN stores s ON s.id = inv.store_id
    LEFT JOIN (
        SELECT
            il.invoice_id,
            SUM(COALESCE(p.cost_price, 0) * il.quantity) AS total_cost
        FROM invoice_lines il
        LEFT JOIN products p ON p.id = il.product_id
        GROUP BY il.invoice_id
    ) il_cost ON il_cost.invoice_id = inv.id
    WHERE inv.invoice_status NOT IN ('draft', 'cancelled')
    GROUP BY inv.store_id, inv.organization_id, s.name, inv.invoice_date
),
all_periods AS (
    SELECT store_id, organization_id, store_name, period_date FROM daily_pos_sales
    UNION
    SELECT store_id, organization_id, store_name, period_date FROM daily_invoices
)
SELECT
    ap.organization_id,
    ap.store_id,
    ap.store_name,
    ap.period_date,
    DATE_TRUNC('month', ap.period_date)::DATE AS period_month,
    COALESCE(ps.transaction_count, 0) + COALESCE(di.invoice_count, 0) AS total_transactions,
    COALESCE(ps.gross_sales, 0) + COALESCE(di.invoice_gross, 0) AS gross_revenue,
    COALESCE(ps.total_discounts, 0) AS total_discounts,
    COALESCE(ret.total_returns, 0) AS total_returns,
    (COALESCE(ps.gross_sales, 0) + COALESCE(di.invoice_gross, 0) - COALESCE(ps.total_discounts, 0) - COALESCE(ret.total_returns, 0)) AS net_revenue,
    COALESCE(ps.total_cogs, 0) + COALESCE(di.invoice_cogs, 0) AS cogs,
    (COALESCE(ps.gross_sales, 0) + COALESCE(di.invoice_gross, 0) - COALESCE(ps.total_discounts, 0) - COALESCE(ret.total_returns, 0))
        - (COALESCE(ps.total_cogs, 0) + COALESCE(di.invoice_cogs, 0)) AS gross_profit,
    CASE
        WHEN (COALESCE(ps.gross_sales, 0) + COALESCE(di.invoice_gross, 0) - COALESCE(ps.total_discounts, 0) - COALESCE(ret.total_returns, 0)) > 0
        THEN ROUND(
            ((COALESCE(ps.gross_sales, 0) + COALESCE(di.invoice_gross, 0) - COALESCE(ps.total_discounts, 0) - COALESCE(ret.total_returns, 0))
             - (COALESCE(ps.total_cogs, 0) + COALESCE(di.invoice_cogs, 0)))
            / (COALESCE(ps.gross_sales, 0) + COALESCE(di.invoice_gross, 0) - COALESCE(ps.total_discounts, 0) - COALESCE(ret.total_returns, 0)) * 100, 2
        )
        ELSE 0
    END AS gross_margin_pct,
    COALESCE(ps.total_tax, 0) + COALESCE(di.invoice_tax, 0) AS total_tax_collected
FROM all_periods ap
LEFT JOIN daily_pos_sales ps ON ps.store_id = ap.store_id AND ps.period_date = ap.period_date
LEFT JOIN daily_returns ret ON ret.store_id = ap.store_id AND ret.period_date = ap.period_date
LEFT JOIN daily_invoices di ON di.store_id = ap.store_id AND di.period_date = ap.period_date;

CREATE OR REPLACE VIEW public.vw_cashier_shift_reconciliation AS
WITH payments_summary AS (
    SELECT
        pt.cashier_session_id,
        COUNT(DISTINCT pt.id) AS transaction_count,
        COALESCE(SUM(pt.total_amount), 0) AS total_sales,
        COALESCE(SUM(CASE WHEN LOWER(p.payment_method) = 'cash' THEN p.amount ELSE 0 END), 0) AS cash_collected,
        COALESCE(SUM(CASE WHEN LOWER(p.payment_method) IN ('card', 'credit_card', 'debit_card') THEN p.amount ELSE 0 END), 0) AS card_collected,
        COALESCE(SUM(CASE WHEN LOWER(p.payment_method) NOT IN ('cash', 'card', 'credit_card', 'debit_card') THEN p.amount ELSE 0 END), 0) AS other_tender_collected
    FROM pos_transactions pt
    LEFT JOIN pos_payments p ON p.transaction_id = pt.id
    WHERE pt.status = 'completed'
      AND pt.cashier_session_id IS NOT NULL
    GROUP BY pt.cashier_session_id
),
refunds_summary AS (
    SELECT
        sr.cashier_session_id,
        COUNT(DISTINCT sr.id) AS return_count,
        COALESCE(SUM(CASE WHEN LOWER(COALESCE(sr.refund_method, 'cash')) = 'cash' THEN sr.total_refund_amount ELSE 0 END), 0) AS cash_refunds,
        COALESCE(SUM(sr.total_refund_amount), 0) AS total_refunds
    FROM sales_returns sr
    WHERE sr.status = 'completed'
      AND sr.cashier_session_id IS NOT NULL
    GROUP BY sr.cashier_session_id
)
SELECT
    cs.id AS session_id,
    cs.session_number,
    t.store_id,
    s.name AS store_name,
    cs.pos_terminal_id,
    t.terminal_code,
    t.terminal_name,
    cs.cashier_id,
    c.cashier_code,
    u.first_name || ' ' || u.last_name AS cashier_name,
    cs.opening_time,
    cs.closing_time,
    cs.status,
    COALESCE(cs.opening_balance, 0) AS opening_balance,
    COALESCE(ps.cash_collected, 0) AS cash_collected,
    COALESCE(ps.card_collected, 0) AS card_collected,
    COALESCE(ps.other_tender_collected, 0) AS other_tender_collected,
    COALESCE(rs.cash_refunds, 0) AS cash_refunds,
    (COALESCE(cs.opening_balance, 0) + COALESCE(ps.cash_collected, 0) - COALESCE(rs.cash_refunds, 0)) AS expected_closing_cash,
    cs.closing_balance AS actual_closing_cash,
    COALESCE(
        cs.variance,
        cs.closing_balance - (COALESCE(cs.opening_balance, 0) + COALESCE(ps.cash_collected, 0) - COALESCE(rs.cash_refunds, 0))
    ) AS over_short_variance,
    COALESCE(ps.total_sales, 0) AS total_sales,
    COALESCE(ps.transaction_count, 0) AS transaction_count,
    COALESCE(rs.total_refunds, 0) AS total_refunds,
    COALESCE(rs.return_count, 0) AS return_count
FROM cashier_sessions cs
JOIN pos_terminals t ON t.id = cs.pos_terminal_id
JOIN stores s ON s.id = t.store_id
JOIN cashiers c ON c.id = cs.cashier_id
JOIN users u ON u.id = c.user_id
LEFT JOIN payments_summary ps ON ps.cashier_session_id = cs.id
LEFT JOIN refunds_summary rs ON rs.cashier_session_id = cs.id;

CREATE OR REPLACE VIEW public.vw_stock_movement_ledger AS
SELECT
    sm.id AS movement_id,
    sm.movement_date,
    COALESCE(sm.to_store_id, sm.from_store_id) AS store_id,
    s.name AS store_name,
    s.organization_id,
    sm.product_id,
    p.sku,
    p.name AS product_name,
    pb.barcode,
    pc.name AS category_name,
    sm.movement_type,
    sm.reference_type,
    sm.reference_id,
    sm.batch_number,
    sm.serial_number,
    CASE WHEN sm.quantity > 0 THEN sm.quantity ELSE 0 END AS quantity_in,
    CASE WHEN sm.quantity < 0 THEN ABS(sm.quantity) ELSE 0 END AS quantity_out,
    sm.quantity AS net_quantity,
    COALESCE(sm.cost_per_unit, p.cost_price, 0) AS unit_cost,
    COALESCE(sm.total_value, sm.quantity * COALESCE(sm.cost_per_unit, p.cost_price, 0)) AS movement_value,
    SUM(sm.quantity) OVER (
        PARTITION BY COALESCE(sm.to_store_id, sm.from_store_id), sm.product_id
        ORDER BY sm.movement_date, sm.id
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) AS running_quantity,
    SUM(sm.quantity) OVER (
        PARTITION BY COALESCE(sm.to_store_id, sm.from_store_id), sm.product_id
        ORDER BY sm.movement_date, sm.id
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) * COALESCE(sm.cost_per_unit, p.cost_price, 0) AS running_valuation,
    sm.status,
    sm.metadata
FROM stock_movements sm
JOIN products p ON p.id = sm.product_id
LEFT JOIN product_barcodes pb ON pb.product_id = p.id AND pb.is_primary = true
LEFT JOIN product_categories pc ON pc.id = p.category_id
LEFT JOIN stores s ON s.id = COALESCE(sm.to_store_id, sm.from_store_id);

CREATE OR REPLACE VIEW public.vw_tax_vat_summary AS
WITH output_vat AS (
    SELECT
        s.organization_id,
        pt.store_id,
        DATE_TRUNC('month', pt.transaction_date)::DATE AS period_month,
        COALESCE(SUM(pt.subtotal), 0) AS taxable_sales,
        COALESCE(SUM(pt.tax_amount), 0) AS output_vat
    FROM pos_transactions pt
    JOIN stores s ON s.id = pt.store_id
    WHERE pt.status = 'completed'
    GROUP BY s.organization_id, pt.store_id, DATE_TRUNC('month', pt.transaction_date)::DATE

    UNION ALL

    SELECT
        inv.organization_id,
        inv.store_id,
        DATE_TRUNC('month', inv.invoice_date::timestamp)::DATE AS period_month,
        COALESCE(SUM(inv.subtotal), 0) AS taxable_sales,
        COALESCE(SUM(inv.tax_amount), 0) AS output_vat
    FROM invoices inv
    WHERE inv.invoice_status NOT IN ('draft', 'cancelled')
    GROUP BY inv.organization_id, inv.store_id, DATE_TRUNC('month', inv.invoice_date::timestamp)::DATE
),
input_vat AS (
    SELECT
        po.organization_id,
        po.store_id,
        DATE_TRUNC('month', po.po_date::timestamp)::DATE AS period_month,
        COALESCE(SUM(po.subtotal), 0) AS taxable_purchases,
        COALESCE(SUM(po.tax_amount), 0) AS input_vat
    FROM purchase_orders po
    WHERE po.status NOT IN ('draft', 'cancelled')
    GROUP BY po.organization_id, po.store_id, DATE_TRUNC('month', po.po_date::timestamp)::DATE
),
all_periods AS (
    SELECT organization_id, store_id, period_month FROM output_vat
    UNION
    SELECT organization_id, store_id, period_month FROM input_vat
)
SELECT
    ap.organization_id,
    ap.store_id,
    s.name AS store_name,
    ap.period_month,
    COALESCE(SUM(ov.taxable_sales), 0) AS total_taxable_sales,
    COALESCE(SUM(ov.output_vat), 0) AS total_output_vat,
    COALESCE(SUM(iv.taxable_purchases), 0) AS total_taxable_purchases,
    COALESCE(SUM(iv.input_vat), 0) AS total_input_vat,
    (COALESCE(SUM(ov.output_vat), 0) - COALESCE(SUM(iv.input_vat), 0)) AS net_vat_payable
FROM all_periods ap
JOIN stores s ON s.id = ap.store_id
LEFT JOIN output_vat ov ON ov.organization_id = ap.organization_id AND ov.store_id = ap.store_id AND ov.period_month = ap.period_month
LEFT JOIN input_vat iv ON iv.organization_id = ap.organization_id AND iv.store_id = ap.store_id AND iv.period_month = ap.period_month
GROUP BY ap.organization_id, ap.store_id, s.name, ap.period_month
ORDER BY ap.period_month DESC, s.name;

-- 11. Recreate the 11 updated functions & triggers (LANGUAGE plpgsql AS $$ ... $$; for Goose wrapping)
CREATE OR REPLACE FUNCTION public.fn_process_stock_transfer(
    p_from_store_id     INTEGER,
    p_to_store_id       INTEGER,
    p_product_id        INTEGER,
    p_product_variant_id INTEGER,
    p_quantity          DECIMAL(15,3),
    p_from_location_id  INTEGER DEFAULT NULL,
    p_to_location_id    INTEGER DEFAULT NULL,
    p_batch_number      VARCHAR DEFAULT NULL,
    p_performed_by      INTEGER DEFAULT NULL,
    p_notes             TEXT    DEFAULT NULL
)
RETURNS TABLE (
    success          BOOLEAN,
    message          TEXT,
    movement_id      UUID
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_available  DECIMAL(15,3);
    v_movement_id UUID;
    v_ref_num    VARCHAR(50);
BEGIN
    IF p_from_store_id = p_to_store_id THEN
        RETURN QUERY SELECT false, 'Source and destination stores must differ.', NULL::UUID;
        RETURN;
    END IF;

    SELECT quantity_available INTO v_available
    FROM inventory_stock
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_from_store_id
    FOR UPDATE;

    IF v_available IS NULL OR v_available < p_quantity THEN
        RETURN QUERY SELECT false,
            format('Insufficient stock. Available: %s, Requested: %s', COALESCE(v_available, 0), p_quantity),
            NULL::UUID;
        RETURN;
    END IF;

    v_ref_num := 'TRF-' || to_char(CURRENT_TIMESTAMP, 'YYYYMMDDHH24MISS') || '-' || p_from_store_id || '-' || p_to_store_id;

    UPDATE inventory_stock
    SET quantity_on_hand   = quantity_on_hand   - p_quantity,
        quantity_available = quantity_available - p_quantity,
        quantity_in_transit = quantity_in_transit + p_quantity,
        updated_at = CURRENT_TIMESTAMP
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_from_store_id;

    UPDATE inventory_stock
    SET quantity_on_hand   = quantity_on_hand   + p_quantity,
        quantity_available = quantity_available + p_quantity,
        updated_at         = CURRENT_TIMESTAMP
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_to_store_id;

    IF NOT FOUND THEN
        INSERT INTO inventory_stock (product_id, product_variant_id, store_id, storage_location_id,
            quantity_on_hand, quantity_available, quantity_in_transit)
        VALUES (p_product_id, p_product_variant_id, p_to_store_id, p_to_location_id,
                p_quantity, p_quantity, 0);
    END IF;

    UPDATE inventory_stock
    SET quantity_in_transit = GREATEST(0, quantity_in_transit - p_quantity),
        updated_at = CURRENT_TIMESTAMP
    WHERE product_id = p_product_id
      AND (product_variant_id = p_product_variant_id OR (product_variant_id IS NULL AND p_product_variant_id IS NULL))
      AND store_id = p_from_store_id;

    INSERT INTO stock_movements (movement_type, reference_type, product_id, product_variant_id,
        from_store_id, to_store_id, from_location_id, to_location_id,
        quantity, batch_number, posted_by, status,
        metadata)
    VALUES ('transfer', 'manual', p_product_id, p_product_variant_id,
            p_from_store_id, p_to_store_id, p_from_location_id, p_to_location_id,
            p_quantity, p_batch_number, p_performed_by, 'completed',
            jsonb_build_object('reference_number', v_ref_num, 'notes', p_notes))
    RETURNING id INTO v_movement_id;

    RETURN QUERY SELECT true, 'Transfer completed successfully. Ref: ' || v_ref_num, v_movement_id;
END;
$$;

CREATE OR REPLACE FUNCTION public.fn_ship_transfer_request(
    p_transfer_request_id INTEGER,
    p_shipped_by INTEGER
)
RETURNS TABLE (
    success BOOLEAN,
    message TEXT
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_req RECORD;
    v_item RECORD;
    v_available DECIMAL(15,3);
    v_qty DECIMAL(15,3);
    v_uom_code VARCHAR;
    v_base_qty DECIMAL(15,3);
BEGIN
    SELECT * INTO v_req FROM transfer_requests WHERE id = p_transfer_request_id FOR UPDATE;
    IF v_req IS NULL THEN
        RETURN QUERY SELECT false, 'Transfer request not found.';
        RETURN;
    END IF;

    IF v_req.status NOT IN ('approved', 'pending_approval', 'draft') THEN
        RETURN QUERY SELECT false, 'Transfer request must be approved to ship.';
        RETURN;
    END IF;

    FOR v_item IN SELECT * FROM transfer_request_items WHERE transfer_request_id = p_transfer_request_id FOR UPDATE LOOP
        v_qty := CASE WHEN v_item.approved_quantity > 0 THEN v_item.approved_quantity ELSE v_item.requested_quantity END;
        IF v_qty <= 0 THEN
            CONTINUE;
        END IF;

        SELECT code INTO v_uom_code FROM units_of_measure WHERE id = v_item.uom_id;

        v_base_qty := fn_convert_uom_quantity(v_item.product_id, v_uom_code, v_qty);
        IF v_base_qty IS NULL THEN
            v_base_qty := v_qty;
        END IF;

        IF v_req.is_stock_reserved THEN
            UPDATE inventory_stock
            SET quantity_allocated = GREATEST(0, quantity_allocated - v_base_qty),
                quantity_on_hand = quantity_on_hand - v_base_qty,
                updated_at = CURRENT_TIMESTAMP
            WHERE product_id = v_item.product_id
              AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
              AND store_id = v_req.from_store_id;

            UPDATE stock_reservations
            SET status = 'fulfilled',
                updated_at = CURRENT_TIMESTAMP
            WHERE reference_type = 'transfer'
              AND reference_id = p_transfer_request_id::text
              AND product_id = v_item.product_id
              AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
              AND status = 'active';
        ELSE
            SELECT quantity_available INTO v_available
            FROM inventory_stock
            WHERE product_id = v_item.product_id
              AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
              AND store_id = v_req.from_store_id
            FOR UPDATE;

            IF v_available IS NULL OR v_available < v_base_qty THEN
                RETURN QUERY SELECT false, format('Insufficient stock for product ID %s at source store.', v_item.product_id);
                RETURN;
            END IF;

            UPDATE inventory_stock
            SET quantity_on_hand = quantity_on_hand - v_base_qty,
                quantity_available = quantity_available - v_base_qty,
                updated_at = CURRENT_TIMESTAMP
            WHERE product_id = v_item.product_id
              AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
              AND store_id = v_req.from_store_id;
        END IF;

        UPDATE inventory_stock
        SET quantity_in_transit = quantity_in_transit + v_base_qty,
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_req.to_store_id;

        IF NOT FOUND THEN
            INSERT INTO inventory_stock (product_id, product_variant_id, store_id, storage_location_id,
                quantity_on_hand, quantity_available, quantity_in_transit)
            VALUES (v_item.product_id, v_item.product_variant_id, v_req.to_store_id, v_item.to_location_id,
                    0, 0, v_base_qty);
        END IF;

        UPDATE transfer_request_items
        SET shipped_quantity = v_qty,
            approved_quantity = v_qty
        WHERE id = v_item.id;

        INSERT INTO stock_movements (
            movement_type, reference_type, reference_id, product_id, product_variant_id,
            from_store_id, to_store_id, from_location_id, to_location_id,
            quantity, uom_id, batch_number, posted_by, status, metadata
        ) VALUES (
            'transfer_out', 'transfer_request', p_transfer_request_id::text, v_item.product_id, v_item.product_variant_id,
            v_req.from_store_id, v_req.to_store_id, v_item.from_location_id, v_item.to_location_id,
            v_base_qty, v_item.uom_id, v_item.batch_number, p_shipped_by, 'shipped',
            jsonb_build_object('transfer_number', v_req.transfer_number)
        );
    END LOOP;

    UPDATE transfer_requests
    SET status = 'shipped',
        shipped_by = p_shipped_by,
        shipped_at = CURRENT_TIMESTAMP,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_transfer_request_id;

    RETURN QUERY SELECT true, 'Transfer request shipped successfully.';
END;
$$;

CREATE OR REPLACE FUNCTION public.fn_receive_transfer_request(
    p_transfer_request_id INTEGER,
    p_received_by INTEGER
)
RETURNS TABLE (
    success BOOLEAN,
    message TEXT
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_req RECORD;
    v_item RECORD;
    v_qty DECIMAL(15,3);
    v_uom_code VARCHAR;
    v_base_qty DECIMAL(15,3);
BEGIN
    SELECT * INTO v_req FROM transfer_requests WHERE id = p_transfer_request_id FOR UPDATE;
    IF v_req IS NULL THEN
        RETURN QUERY SELECT false, 'Transfer request not found.';
        RETURN;
    END IF;

    IF v_req.status NOT IN ('shipped', 'partially_received') THEN
        RETURN QUERY SELECT false, 'Transfer request must be shipped or partially_received to receive.';
        RETURN;
    END IF;

    FOR v_item IN SELECT * FROM transfer_request_items WHERE transfer_request_id = p_transfer_request_id FOR UPDATE LOOP
        v_qty := CASE WHEN v_item.shipped_quantity > 0 THEN v_item.shipped_quantity ELSE v_item.requested_quantity END;
        IF v_qty <= 0 THEN
            CONTINUE;
        END IF;

        SELECT code INTO v_uom_code FROM units_of_measure WHERE id = v_item.uom_id;

        v_base_qty := fn_convert_uom_quantity(v_item.product_id, v_uom_code, v_qty);
        IF v_base_qty IS NULL THEN
            v_base_qty := v_qty;
        END IF;

        UPDATE inventory_stock
        SET quantity_in_transit = GREATEST(0, quantity_in_transit - v_base_qty),
            quantity_on_hand = quantity_on_hand + v_base_qty,
            quantity_available = quantity_available + v_base_qty,
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_req.to_store_id;

        IF NOT FOUND THEN
            INSERT INTO inventory_stock (product_id, product_variant_id, store_id, storage_location_id,
                quantity_on_hand, quantity_available, quantity_in_transit)
            VALUES (v_item.product_id, v_item.product_variant_id, v_req.to_store_id, v_item.to_location_id,
                    v_base_qty, v_base_qty, 0);
        END IF;

        UPDATE transfer_request_items
        SET received_quantity = v_qty
        WHERE id = v_item.id;

        INSERT INTO stock_movements (
            movement_type, reference_type, reference_id, product_id, product_variant_id,
            from_store_id, to_store_id, from_location_id, to_location_id,
            quantity, uom_id, batch_number, posted_by, status, metadata
        ) VALUES (
            'transfer_in', 'transfer_request', p_transfer_request_id::text, v_item.product_id, v_item.product_variant_id,
            v_req.from_store_id, v_req.to_store_id, v_item.from_location_id, v_item.to_location_id,
            v_base_qty, v_item.uom_id, v_item.batch_number, p_received_by, 'completed',
            jsonb_build_object('transfer_number', v_req.transfer_number)
        );
    END LOOP;

    UPDATE transfer_requests
    SET status = 'received',
        received_by = p_received_by,
        received_at = CURRENT_TIMESTAMP,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_transfer_request_id;

    RETURN QUERY SELECT true, 'Transfer request received successfully.';
END;
$$;

CREATE OR REPLACE FUNCTION public.fn_process_goods_receipt(
    p_grn_id INTEGER
)
RETURNS TABLE (
    success BOOLEAN,
    message TEXT
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_grn RECORD;
    v_item RECORD;
    v_all_po_received BOOLEAN := true;
    v_any_po_received BOOLEAN := false;
    v_po_line RECORD;
    v_org_metadata JSONB;
    v_strategy VARCHAR(50);
    v_current_price NUMERIC;
    v_old_qty NUMERIC;
    v_new_price NUMERIC;
    v_incoming_price NUMERIC;
    v_total_qty NUMERIC;
BEGIN
    SELECT * INTO v_grn FROM goods_receipt_notes WHERE id = p_grn_id FOR UPDATE;
    IF v_grn IS NULL THEN
        RETURN QUERY SELECT false, 'Goods receipt note not found.';
        RETURN;
    END IF;

    IF v_grn.status = 'completed' THEN
        RETURN QUERY SELECT false, 'Goods receipt note is already completed.';
        RETURN;
    END IF;

    FOR v_item IN SELECT * FROM goods_receipt_note_items WHERE grn_id = p_grn_id FOR UPDATE LOOP
        IF v_item.quantity_received <= 0 THEN
            CONTINUE;
        END IF;

        IF v_item.purchase_order_line_id IS NOT NULL THEN
            UPDATE purchase_order_lines
            SET received_quantity = received_quantity + v_item.quantity_received
            WHERE id = v_item.purchase_order_line_id;
        END IF;

        UPDATE inventory_stock
        SET quantity_on_hand = quantity_on_hand + v_item.quantity_received,
            quantity_available = quantity_available + v_item.quantity_received,
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = v_item.product_id
          AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
          AND store_id = v_grn.store_id;

        IF NOT FOUND THEN
            INSERT INTO inventory_stock (
                product_id, product_variant_id, store_id, storage_location_id,
                quantity_on_hand, quantity_available
            ) VALUES (
                v_item.product_id, v_item.product_variant_id, v_grn.store_id, v_item.storage_location_id,
                v_item.quantity_received, v_item.quantity_received
            );
        END IF;

        IF v_item.batch_number IS NOT NULL AND v_item.batch_number <> '' THEN
            UPDATE product_batches
            SET quantity_available = quantity_available + v_item.quantity_received,
                expiry_date = COALESCE(v_item.expiry_date, expiry_date),
                updated_at = CURRENT_TIMESTAMP
            WHERE product_id = v_item.product_id
              AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
              AND store_id = v_grn.store_id
              AND batch_number = v_item.batch_number;

            IF NOT FOUND THEN
                INSERT INTO product_batches (
                    product_id, product_variant_id, batch_number,
                    expiry_date, store_id, quantity_available, status
                ) VALUES (
                    v_item.product_id, v_item.product_variant_id, v_item.batch_number,
                    v_item.expiry_date, v_grn.store_id, v_item.quantity_received, 'active'
                );
            END IF;
        END IF;

        INSERT INTO stock_movements (
            movement_type, reference_type, reference_id, product_id, product_variant_id,
            to_store_id, to_location_id, quantity, uom_id, batch_number,
            posted_by, status, cost_per_unit, total_value, metadata
        ) VALUES (
            'purchase_receipt', 'goods_receipt_note', p_grn_id::text, v_item.product_id, v_item.product_variant_id,
            v_grn.store_id, v_item.storage_location_id, v_item.quantity_received, v_item.uom_id, COALESCE(v_item.batch_number, ''),
            v_grn.received_by, 'completed', COALESCE(v_item.unit_cost, 0), (COALESCE(v_item.unit_cost, 0) * v_item.quantity_received),
            jsonb_build_object('grn_number', v_grn.grn_number, 'delivery_note_number', COALESCE(v_grn.delivery_note_number, ''))
        );

        SELECT metadata INTO v_org_metadata FROM organizations WHERE id = v_grn.organization_id;

        IF jsonb_typeof(v_org_metadata) = 'array' THEN
            v_strategy := COALESCE((v_org_metadata->0)->>'inventory_strategy', 'WEIGHTED_AVERAGE');
        ELSE
            v_strategy := COALESCE(v_org_metadata->>'inventory_strategy', 'WEIGHTED_AVERAGE');
        END IF;

        IF v_strategy = 'WEIGHTED_AVERAGE' THEN
            SELECT quantity_on_hand INTO v_total_qty
            FROM inventory_stock
            WHERE product_id = v_item.product_id
              AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
              AND store_id = v_grn.store_id;

            v_old_qty := COALESCE(v_total_qty, 0) - v_item.quantity_received;
            IF v_old_qty < 0 THEN
                v_old_qty := 0;
            END IF;

            SELECT price INTO v_current_price
            FROM product_prices
            WHERE product_id = v_item.product_id
              AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
              AND price_list_id IN (SELECT id FROM price_lists WHERE code IN ('RETAIL', 'RETAIL_SAR') AND is_active = true)
              AND is_active = true
            LIMIT 1;

            v_incoming_price := COALESCE(v_item.unit_cost, 0);

            IF v_old_qty = 0 OR v_current_price IS NULL THEN
                v_new_price := v_incoming_price;
            ELSE
                v_new_price := ROUND(((v_old_qty * v_current_price) + (v_item.quantity_received * v_incoming_price)) / (v_old_qty + v_item.quantity_received), 2);
            END IF;

            IF v_new_price > 0 THEN
                UPDATE product_prices
                SET price = v_new_price,
                    updated_at = CURRENT_TIMESTAMP
                WHERE product_id = v_item.product_id
                  AND (product_variant_id = v_item.product_variant_id OR (product_variant_id IS NULL AND v_item.product_variant_id IS NULL))
                  AND price_list_id IN (SELECT id FROM price_lists WHERE code IN ('RETAIL', 'RETAIL_SAR') AND is_active = true);
            END IF;
        END IF;
    END LOOP;

    UPDATE goods_receipt_notes
    SET status = 'completed',
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_grn_id;

    IF v_grn.purchase_order_id IS NOT NULL THEN
        FOR v_po_line IN SELECT quantity, received_quantity FROM purchase_order_lines WHERE purchase_order_id = v_grn.purchase_order_id LOOP
            IF v_po_line.received_quantity < v_po_line.quantity THEN
                v_all_po_received := false;
            END IF;
            IF v_po_line.received_quantity > 0 THEN
                v_any_po_received := true;
            END IF;
        END LOOP;

        IF v_all_po_received THEN
            UPDATE purchase_orders SET status = 'received', updated_at = CURRENT_TIMESTAMP WHERE id = v_grn.purchase_order_id;
        ELSIF v_any_po_received THEN
            UPDATE purchase_orders SET status = 'partially_received', updated_at = CURRENT_TIMESTAMP WHERE id = v_grn.purchase_order_id;
        END IF;
    END IF;

    RETURN QUERY SELECT true, 'Goods receipt processed successfully.';
END;
$$;

CREATE OR REPLACE FUNCTION public.fn_reconcile_stock_count(p_count_id UUID)
RETURNS TABLE (
    success       BOOLEAN,
    message       TEXT,
    lines_updated INTEGER
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_count        RECORD;
    v_line         RECORD;
    v_lines_updated INTEGER := 0;
BEGIN
    SELECT * INTO v_count FROM stock_counts WHERE id = p_count_id;

    IF NOT FOUND THEN
        RETURN QUERY SELECT false, 'Stock count not found.', 0;
        RETURN;
    END IF;
    IF v_count.status <> 'approved' THEN
        RETURN QUERY SELECT false, 'Stock count must be in ''approved'' status to reconcile.', 0;
        RETURN;
    END IF;

    FOR v_line IN
        SELECT * FROM stock_count_lines WHERE stock_count_id = p_count_id
    LOOP
        UPDATE inventory_stock
        SET quantity_on_hand   = v_line.counted_quantity,
            quantity_available = GREATEST(0, v_line.counted_quantity - COALESCE(quantity_allocated, 0)),
            last_counted_at    = CURRENT_TIMESTAMP,
            updated_at         = CURRENT_TIMESTAMP
        WHERE product_id = v_line.product_id
          AND (product_variant_id = v_line.product_variant_id
               OR (product_variant_id IS NULL AND v_line.product_variant_id IS NULL))
          AND store_id = v_count.store_id;

        INSERT INTO stock_movements (movement_type, reference_type, reference_id, product_id,
            product_variant_id, to_store_id, quantity, batch_number, serial_number,
            posted_by, status, metadata)
        VALUES ('count_adjustment', 'stock_count', p_count_id::text, v_line.product_id,
                v_line.product_variant_id, v_count.store_id,
                v_line.counted_quantity - v_line.system_quantity,
                v_line.batch_number, v_line.serial_number,
                v_count.approved_by, 'completed',
                jsonb_build_object('count_id', p_count_id, 'variance', v_line.variance));

        v_lines_updated := v_lines_updated + 1;
    END LOOP;

    UPDATE stock_counts SET status = 'approved', updated_at = CURRENT_TIMESTAMP
    WHERE id = p_count_id;

    RETURN QUERY SELECT true, format('Reconciliation complete. %s lines adjusted.', v_lines_updated), v_lines_updated;
END;
$$;

CREATE OR REPLACE FUNCTION public.fn_trigger_deduct_inventory_on_pos_transaction()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_store_id INTEGER;
    v_status VARCHAR(50);
    v_base_quantity NUMERIC(15,3);
    v_uom_code VARCHAR(50);
BEGIN
    SELECT store_id, status INTO v_store_id, v_status FROM pos_transactions WHERE id = NEW.transaction_id;

    IF v_store_id IS NULL OR v_status = 'voided' THEN
        RETURN NEW;
    END IF;

    IF NEW.uom_id IS NOT NULL THEN
        SELECT code INTO v_uom_code FROM units_of_measure WHERE id = NEW.uom_id;
        IF v_uom_code IS NOT NULL THEN
            v_base_quantity := fn_convert_uom_quantity(NEW.product_id, v_uom_code, NEW.quantity);
        END IF;
    END IF;
    v_base_quantity := COALESCE(v_base_quantity, NEW.quantity);

    UPDATE inventory_stock
    SET quantity_on_hand = quantity_on_hand - v_base_quantity,
        quantity_available = GREATEST(0, quantity_available - v_base_quantity),
        updated_at = CURRENT_TIMESTAMP
    WHERE product_id = NEW.product_id
      AND (product_variant_id = NEW.product_variant_id OR (product_variant_id IS NULL AND NEW.product_variant_id IS NULL))
      AND store_id = v_store_id;

    IF NEW.batch_number IS NOT NULL AND NEW.batch_number <> '' THEN
        UPDATE product_batches
        SET quantity_available = GREATEST(0, quantity_available - v_base_quantity),
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = NEW.product_id
          AND (product_variant_id = NEW.product_variant_id OR (product_variant_id IS NULL AND NEW.product_variant_id IS NULL))
          AND store_id = v_store_id
          AND batch_number = NEW.batch_number;
    END IF;

    INSERT INTO stock_movements (
        movement_type, reference_type, reference_id, product_id, product_variant_id,
        from_store_id, quantity, uom_id, status, metadata
    )
    VALUES (
        'sale', 'pos_transaction', NEW.transaction_id::text, NEW.product_id, NEW.product_variant_id,
        v_store_id, NEW.quantity, NEW.uom_id, 'completed',
        jsonb_build_object('pos_line_id', NEW.id, 'batch_number', COALESCE(NEW.batch_number, ''))
    );

    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION public.fn_calculate_loyalty_earned(p_transaction_id UUID)
RETURNS TABLE (
    points_earned  DECIMAL(15,2),
    rule_applied   VARCHAR(255),
    customer_id    UUID
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_txn        RECORD;
    v_rule       RECORD;
    v_points     DECIMAL(15,2) := 0;
BEGIN
    SELECT pt.*, pt.total_amount, pt.customer_id AS cust_id
    INTO v_txn
    FROM pos_transactions pt
    WHERE pt.id = p_transaction_id;

    IF NOT FOUND OR v_txn.cust_id IS NULL THEN
        RETURN QUERY SELECT 0::DECIMAL(15,2), 'No customer on transaction'::VARCHAR(255), NULL::UUID;
        RETURN;
    END IF;

    SELECT * INTO v_rule
    FROM loyalty_redemption_rules
    WHERE is_active = true
      AND (valid_from IS NULL OR valid_from <= CURRENT_DATE)
      AND (valid_to   IS NULL OR valid_to   >= CURRENT_DATE)
    ORDER BY id DESC LIMIT 1;

    IF NOT FOUND THEN
        RETURN QUERY SELECT 0::DECIMAL(15,2), 'No active loyalty rule found'::VARCHAR(255), v_txn.cust_id;
        RETURN;
    END IF;

    v_points := FLOOR(v_txn.total_amount * v_rule.points_earning_rate);

    UPDATE customers
    SET loyalty_points = loyalty_points + v_points,
        updated_at     = CURRENT_TIMESTAMP
    WHERE id = v_txn.cust_id;

    RETURN QUERY SELECT v_points, v_rule.rule_name::VARCHAR(255), v_txn.cust_id;
END;
$$;

CREATE OR REPLACE FUNCTION public.fn_get_kds_orders(
    p_store_id      INTEGER,
    p_statuses      VARCHAR[] DEFAULT ARRAY['pending','confirmed','preparing']
)
RETURNS TABLE (
    order_id            UUID,
    order_number        VARCHAR,
    table_number        VARCHAR,
    waiter_name         VARCHAR,
    order_status        VARCHAR,
    ordered_at          TIMESTAMP,
    minutes_elapsed     NUMERIC,
    item_id             UUID,
    item_name           VARCHAR,
    item_short_name     VARCHAR,
    item_qty            NUMERIC,
    item_notes          TEXT,
    item_modifiers      JSONB,
    item_status         VARCHAR
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT
        ro.id,
        ro.order_number::VARCHAR,
        rt.table_number::VARCHAR,
        (u.first_name || ' ' || u.last_name)::VARCHAR,
        ro.status::VARCHAR,
        ro.ordered_at,
        ROUND(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - ro.ordered_at)) / 60.0, 1),
        roi.id,
        mi.name::VARCHAR,
        mi.short_name::VARCHAR,
        roi.quantity,
        roi.notes,
        roi.modifiers_snapshot,
        roi.status::VARCHAR
    FROM restaurant_orders ro
    LEFT JOIN restaurant_tables rt      ON ro.table_id = rt.id
    LEFT JOIN cashiers c                ON ro.cashier_id = c.id
    LEFT JOIN users u                   ON c.user_id = u.id
    JOIN  restaurant_order_items roi    ON ro.id = roi.order_id
    JOIN  menu_items mi                 ON roi.menu_item_id = mi.id
    WHERE ro.store_id = p_store_id
      AND ro.status = ANY(p_statuses)
    ORDER BY ro.ordered_at, roi.line_number;
END;
$$;

CREATE OR REPLACE FUNCTION public.fn_trigger_process_sales_return()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_store_id INT;
    v_cost_price NUMERIC(15,4);
    v_uom_id INT;
    v_reason TEXT;
    v_ret_metadata JSONB;
BEGIN
    SELECT store_id, COALESCE(return_reason, 'POS Sales Return'), metadata
    INTO v_store_id, v_reason, v_ret_metadata
    FROM sales_returns
    WHERE id = NEW.return_id;

    IF v_ret_metadata IS NOT NULL AND (v_ret_metadata ? 'sap_doc_entry' OR v_ret_metadata ? 'sap_doc_num') THEN
        RETURN NEW;
    END IF;

    IF NEW.original_line_id IS NOT NULL THEN
        SELECT COALESCE(cost_price, 0), uom_id
        INTO v_cost_price, v_uom_id
        FROM pos_transaction_lines
        WHERE id = NEW.original_line_id;
    ELSE
        v_cost_price := 0;
        v_uom_id := NULL;
    END IF;

    IF NEW.return_to_stock = true AND (NEW.condition IS NULL OR NEW.condition = 'good') THEN
        UPDATE inventory_stock
        SET quantity_on_hand = quantity_on_hand + NEW.quantity,
            quantity_available = quantity_available + NEW.quantity,
            updated_at = CURRENT_TIMESTAMP
        WHERE product_id = NEW.product_id
          AND (product_variant_id = NEW.product_variant_id OR (product_variant_id IS NULL AND NEW.product_variant_id IS NULL))
          AND store_id = v_store_id;

        INSERT INTO stock_movements (
            movement_type, reference_type, reference_id, product_id, product_variant_id,
            to_store_id, quantity, uom_id, batch_number, serial_number, metadata
        ) VALUES (
            'return', 'sales_return', NEW.return_id::text, NEW.product_id, NEW.product_variant_id,
            v_store_id, NEW.quantity, v_uom_id, NEW.batch_number, NEW.serial_number,
            jsonb_build_object('return_line_id', NEW.id, 'condition', COALESCE(NEW.condition, 'good'))
        );
    ELSE
        INSERT INTO waste_logs (
            store_id, product_id, waste_source, quantity, uom_id, unit_cost, total_cost,
            reason, metadata
        ) VALUES (
            v_store_id, NEW.product_id, 'register', NEW.quantity, v_uom_id, v_cost_price,
            (v_cost_price * NEW.quantity),
            CONCAT('Customer Return - Condition: ', COALESCE(NEW.condition, 'damaged'), '. Reason: ', v_reason),
            jsonb_build_object('return_id', NEW.return_id, 'return_line_id', NEW.id, 'condition', NEW.condition)
        );
    END IF;

    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION public.fn_get_returnable_quantity(p_original_line_id UUID)
RETURNS NUMERIC
LANGUAGE plpgsql
AS $$
DECLARE
    v_orig_qty NUMERIC(15,3);
    v_ret_qty NUMERIC(15,3);
BEGIN
    SELECT quantity INTO v_orig_qty FROM pos_transaction_lines WHERE id = p_original_line_id;
    IF v_orig_qty IS NULL THEN
        RETURN 0;
    END IF;

    SELECT COALESCE(SUM(quantity), 0) INTO v_ret_qty FROM sales_return_lines WHERE original_line_id = p_original_line_id;

    RETURN GREATEST(0, v_orig_qty - v_ret_qty);
END;
$$;

CREATE OR REPLACE FUNCTION public.trg_record_change_log()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    v_entity_type VARCHAR(50);
    v_entity_id   TEXT;
    v_payload     JSONB;
    v_seq         BIGINT;
BEGIN
    IF current_setting('nembus.sync_applier', true) = 'true'
       OR current_setting('nembus.suppress_change_log', true) = 'true'
       OR to_regclass('public.local_device_config') IS NOT NULL THEN
        IF TG_OP = 'DELETE' THEN
            RETURN OLD;
        END IF;
        RETURN NEW;
    END IF;

    v_entity_type := TG_ARGV[0];
    IF TG_OP = 'DELETE' THEN
        v_entity_id := OLD.id::text;
        v_payload := to_jsonb(OLD);
    ELSE
        v_entity_id := NEW.id::text;
        v_payload := to_jsonb(NEW);
    END IF;

    PERFORM pg_advisory_xact_lock(hashtext('nembus.change_log_seq'));

    INSERT INTO change_log (entity_type, entity_id, op, payload, changed_at)
    VALUES (v_entity_type, v_entity_id, TG_OP, v_payload, NOW())
    RETURNING seq INTO v_seq;

    PERFORM pg_notify('nembus_changes', v_seq::text);

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$;

-- Pending SAP stock approvals store numeric transaction_id in JSON payload.
UPDATE public.approval_requests
SET payload = jsonb_set(
    payload,
    '{transaction_id}',
    to_jsonb(pg_temp.det_uuid('pos_transactions', (payload->>'transaction_id')::bigint)::text),
    true
)
WHERE payload ? 'transaction_id'
  AND (payload->>'transaction_id') ~ '^[0-9]+$';

-- Keep queued POS receipt references and idempotency keys aligned with the
-- UUID transaction ID. If a UUID-keyed event already exists, archive the
-- numeric-key event as completed so a later sync cannot send a second receipt.
WITH candidates AS (
    SELECT
        e.id,
        pg_temp.det_uuid(
            'pos_transactions',
            CASE
                WHEN (e.payload->>'transaction_id') ~ '^[0-9]+$'
                    THEN (e.payload->>'transaction_id')::bigint
                WHEN e.entity_id ~ '^[0-9]+$'
                    THEN e.entity_id::bigint
                ELSE split_part(e.idempotency_key, ':', 2)::bigint
            END
        )::text AS transaction_uuid,
        COALESCE(NULLIF(e.payload->>'transaction_number', ''), split_part(e.idempotency_key, ':', 3)) AS transaction_number,
        e.idempotency_key,
        e.created_at
    FROM public.notification_events e
    WHERE e.event_type_id = 'pos.receipt'
      AND e.entity_type = 'pos_transaction'
      AND (
          (e.payload->>'transaction_id') ~ '^[0-9]+$'
          OR e.entity_id ~ '^[0-9]+$'
          OR (split_part(e.idempotency_key, ':', 2)) ~ '^[0-9]+$'
      )
), ranked AS (
    SELECT
        c.*,
        'pos.receipt:' || c.transaction_uuid || ':' || c.transaction_number AS canonical_key,
        row_number() OVER (
            PARTITION BY 'pos.receipt:' || c.transaction_uuid || ':' || c.transaction_number
            ORDER BY (c.idempotency_key = 'pos.receipt:' || c.transaction_uuid || ':' || c.transaction_number) DESC,
                     c.created_at ASC,
                     c.id ASC
        ) AS key_rank
    FROM candidates c
)
UPDATE public.notification_events e
SET payload = jsonb_set(e.payload, '{transaction_id}', to_jsonb(r.transaction_uuid), true),
    entity_id = r.transaction_uuid,
    idempotency_key = CASE
        WHEN r.key_rank = 1 AND NOT EXISTS (
            SELECT 1
            FROM public.notification_events existing
            WHERE existing.event_type_id = 'pos.receipt'
              AND existing.idempotency_key = r.canonical_key
              AND existing.id <> e.id
        ) THEN r.canonical_key
        ELSE 'pos.receipt:deduped:' || e.id::text
    END,
    status = CASE
        WHEN r.key_rank = 1 AND NOT EXISTS (
            SELECT 1
            FROM public.notification_events existing
            WHERE existing.event_type_id = 'pos.receipt'
              AND existing.idempotency_key = r.canonical_key
              AND existing.id <> e.id
        ) THEN e.status
        ELSE 'completed'
    END
FROM ranked r
WHERE e.id = r.id;

DROP FUNCTION IF EXISTS pg_temp.det_uuid(text, text);
DROP FUNCTION IF EXISTS pg_temp.det_uuid(text, uuid);
DROP FUNCTION IF EXISTS pg_temp.det_uuid(text, bigint);
