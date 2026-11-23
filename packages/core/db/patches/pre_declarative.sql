-- Idempotent patches run before Atlas declarative schema apply on existing databases.
-- Fresh CI/test DBs skip these blocks; production DBs upgraded from versioned migrations need them.

-- purchase_orders: rename legacy partners_id → partner_id (views depend on the column;
-- DROP+ADD fails with 2BP01; RENAME updates dependent views automatically).
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'purchase_orders'
          AND column_name = 'partners_id'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'purchase_orders'
          AND column_name = 'partner_id'
    ) THEN
        ALTER TABLE public.purchase_orders RENAME COLUMN partners_id TO partner_id;
    END IF;
END $$;

-- product_prices: drop obsolete unique constraints before Atlas re-applies uq_product_prices_natural_key.
ALTER TABLE public.product_prices
    DROP CONSTRAINT IF EXISTS uq_product_prices_product_price_list;
ALTER TABLE public.product_prices
    DROP CONSTRAINT IF EXISTS uq_product_prices_natural_key;

-- Normalize legacy rows so the natural key matches declarative schema (NULLS NOT DISTINCT).
UPDATE public.product_prices
SET min_quantity = 1
WHERE min_quantity IS NULL;

UPDATE public.product_prices pp
SET uom_id = p.base_uom_id
FROM public.products p
WHERE pp.product_id = p.id
  AND pp.uom_id IS NULL
  AND p.base_uom_id IS NOT NULL;

-- Remove exact duplicates on (product_id, price_list_id, product_variant_id, uom_id, min_quantity).
DELETE FROM public.product_prices a
USING public.product_prices b
WHERE a.id < b.id
  AND a.product_id = b.product_id
  AND a.price_list_id = b.price_list_id
  AND a.product_variant_id IS NOT DISTINCT FROM b.product_variant_id
  AND a.uom_id IS NOT DISTINCT FROM b.uom_id
  AND a.min_quantity IS NOT DISTINCT FROM b.min_quantity;

-- purchase_orders: rename legacy partner index after column rename (idempotent).
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND c.relname = 'idx_purchase_orders_partners_id'
    ) AND NOT EXISTS (
        SELECT 1 FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND c.relname = 'idx_purchase_orders_partner_id'
    ) THEN
        ALTER INDEX public.idx_purchase_orders_partners_id
            RENAME TO idx_purchase_orders_partner_id;
    END IF;
END $$;

-- =====================================================================
-- UUID migration pre-declarative patch for populated Cloud tenant DBs.
-- Idempotent: executes only when pos_transactions.id is still integer/bigint.
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

DO $$
DECLARE
    r RECORD;
    pk_tbls text[] := ARRAY[
        'user_biometric_templates',
        'cashier_sessions',
        'stock_movements',
        'stock_counts',
        'stock_count_lines',
        'customers',
        'pos_transactions',
        'pos_transaction_lines',
        'pos_payments',
        'restaurant_orders',
        'restaurant_order_items',
        'waste_logs',
        'kiosk_sessions',
        'sales_returns',
        'sales_return_lines'
    ];
    t text;
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'pos_transactions'
          AND column_name = 'id'
          AND data_type IN ('integer', 'bigint')
    ) THEN
        -- 1. Drop dependent views before altering columns
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

        -- 2. Drop integer-signature functions before 90_views_functions.sql runs
        DROP FUNCTION IF EXISTS public.fn_process_stock_transfer(integer, integer, integer, integer, numeric, integer, integer, character varying, integer, text);
        DROP FUNCTION IF EXISTS public.fn_reconcile_stock_count(integer);
        DROP FUNCTION IF EXISTS public.fn_calculate_loyalty_earned(integer);
        DROP FUNCTION IF EXISTS public.fn_get_kds_orders(integer, character varying[]);
        DROP FUNCTION IF EXISTS public.fn_get_returnable_quantity(integer);

        -- 3. Dynamically drop ALL FKs referencing any existing migrated table
        FOR r IN
            SELECT c.conrelid::regclass AS tbl_name, c.conname
            FROM pg_constraint c
            JOIN pg_class rel ON rel.oid = c.confrelid
            JOIN pg_namespace n ON n.oid = rel.relnamespace
            WHERE c.contype = 'f'
              AND c.connamespace = 'public'::regnamespace
              AND n.nspname = 'public'
              AND rel.relname = ANY(pk_tbls)
        LOOP
            EXECUTE format('ALTER TABLE %s DROP CONSTRAINT IF EXISTS %I', r.tbl_name, r.conname);
        END LOOP;

        -- 4. Convert all existing migrated PKs in-place
        FOREACH t IN ARRAY pk_tbls
        LOOP
            IF EXISTS (
                SELECT 1 FROM information_schema.columns
                WHERE table_schema = 'public' AND table_name = t
                  AND column_name = 'id' AND data_type IN ('integer', 'bigint')
            ) THEN
                EXECUTE format(
                    'ALTER TABLE public.%I ALTER COLUMN id DROP DEFAULT, ALTER COLUMN id TYPE uuid USING pg_temp.det_uuid(%L, id), ALTER COLUMN id SET DEFAULT gen_random_uuid()',
                    t, t
                );
                EXECUTE format('DROP SEQUENCE IF EXISTS public.%I', t || '_id_seq');
            END IF;
        END LOOP;

        -- 5. Convert all existing referencing FK columns in-place
        FOR r IN
            SELECT * FROM (VALUES
                ('stock_count_lines', 'stock_count_id', 'stock_counts'),
                ('sales_orders', 'customer_id', 'customers'),
                ('carts', 'customer_id', 'customers'),
                ('draft_cart_templates', 'customer_id', 'customers'),
                ('sales_orders_v2', 'customer_id', 'customers'),
                ('pos_transactions', 'cashier_session_id', 'cashier_sessions'),
                ('pos_transactions', 'customer_id', 'customers'),
                ('pos_transaction_lines', 'transaction_id', 'pos_transactions'),
                ('pos_payments', 'transaction_id', 'pos_transactions'),
                ('restaurant_carts', 'customer_id', 'customers'),
                ('restaurant_carts', 'kiosk_session_id', 'kiosk_sessions'),
                ('restaurant_carts', 'converted_to_order_id', 'restaurant_orders'),
                ('restaurant_orders', 'cashier_session_id', 'cashier_sessions'),
                ('restaurant_orders', 'customer_id', 'customers'),
                ('restaurant_orders', 'kiosk_session_id', 'kiosk_sessions'),
                ('restaurant_orders', 'pos_transaction_id', 'pos_transactions'),
                ('restaurant_order_items', 'order_id', 'restaurant_orders'),
                ('restaurant_order_status_history', 'restaurant_order_id', 'restaurant_orders'),
                ('waste_logs', 'order_id', 'restaurant_orders'),
                ('sales_analytics', 'customer_id', 'customers'),
                ('invoices', 'customer_id', 'customers'),
                ('sales_returns', 'cashier_session_id', 'cashier_sessions'),
                ('sales_returns', 'customer_id', 'customers'),
                ('sales_returns', 'original_transaction_id', 'pos_transactions'),
                ('quotes', 'customer_id', 'customers'),
                ('sales_return_lines', 'return_id', 'sales_returns'),
                ('sales_return_lines', 'original_line_id', 'pos_transaction_lines'),
                ('customer_addresses', 'customer_id', 'customers'),
                ('customer_payments', 'customer_id', 'customers')
            ) AS fks(child_tbl, col_name, parent_tbl)
        LOOP
            IF EXISTS (
                SELECT 1 FROM information_schema.columns
                WHERE table_schema = 'public' AND table_name = r.child_tbl
                  AND column_name = r.col_name AND data_type IN ('integer', 'bigint')
            ) THEN
                EXECUTE format(
                    'ALTER TABLE public.%I ALTER COLUMN %I TYPE uuid USING pg_temp.det_uuid(%L, %I)',
                    r.child_tbl, r.col_name, r.parent_tbl, r.col_name
                );
            END IF;
        END LOOP;

        -- 6. Polymorphic references
        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'stock_movements'
              AND column_name = 'reference_id' AND data_type IN ('integer', 'bigint')
        ) THEN
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
        END IF;

        IF to_regclass('public.zatca_document_chain') IS NOT NULL THEN
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
        END IF;

        IF to_regclass('public.stock_reservations') IS NOT NULL THEN
            UPDATE public.stock_reservations
            SET reference_id = pg_temp.det_uuid('pos_transactions', reference_id::bigint)::text
            WHERE reference_type IN ('pos_transaction', 'pos_transactions')
              AND reference_id ~ '^[0-9]+$';
        END IF;

        -- 7. Sync metadata tables (sync_inbox & change_log)
        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'sync_inbox'
              AND column_name = 'entity_id' AND data_type IN ('integer', 'bigint')
        ) THEN
            ALTER TABLE public.sync_inbox
                ALTER COLUMN entity_id TYPE text
                USING CASE
                    WHEN entity_type = ANY(pk_tbls) THEN pg_temp.det_uuid(entity_type, entity_id)::text
                    WHEN entity_type = 'pos_transaction' THEN pg_temp.det_uuid('pos_transactions', entity_id)::text
                    WHEN entity_type = 'sales_return' THEN pg_temp.det_uuid('sales_returns', entity_id)::text
                    ELSE entity_id::text
                END;
        END IF;

        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'change_log'
              AND column_name = 'entity_id' AND data_type IN ('integer', 'bigint')
        ) THEN
            ALTER TABLE public.change_log
                ALTER COLUMN entity_id TYPE text
                USING CASE
                    WHEN entity_type = ANY(pk_tbls) THEN pg_temp.det_uuid(entity_type, entity_id)::text
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
        END IF;

        -- 8. Re-add canonical FKs for existing tables so atlas schema apply sees zero drift
        FOR r IN
            SELECT * FROM (VALUES
                ('stock_count_lines', 'stock_count_lines_stock_count_id_fkey', 'stock_count_id', 'stock_counts', 'ON DELETE CASCADE'),
                ('sales_orders', 'sales_orders_customer_id_fkey', 'customer_id', 'customers', 'ON DELETE SET NULL'),
                ('carts', 'carts_customer_id_fkey', 'customer_id', 'customers', 'ON DELETE SET NULL'),
                ('draft_cart_templates', 'draft_cart_templates_customer_id_fkey', 'customer_id', 'customers', 'ON DELETE CASCADE'),
                ('sales_orders_v2', 'sales_orders_v2_customer_id_fkey', 'customer_id', 'customers', 'ON DELETE SET NULL'),
                ('pos_transactions', 'pos_transactions_cashier_session_id_fkey', 'cashier_session_id', 'cashier_sessions', 'ON DELETE CASCADE'),
                ('pos_transactions', 'pos_transactions_customer_id_fkey', 'customer_id', 'customers', 'ON DELETE SET NULL'),
                ('pos_transaction_lines', 'pos_transaction_lines_transaction_id_fkey', 'transaction_id', 'pos_transactions', 'ON DELETE CASCADE'),
                ('pos_payments', 'pos_payments_transaction_id_fkey', 'transaction_id', 'pos_transactions', 'ON DELETE CASCADE'),
                ('restaurant_carts', 'restaurant_carts_customer_id_fkey', 'customer_id', 'customers', 'ON DELETE SET NULL'),
                ('restaurant_carts', 'fk_restaurant_carts_kiosk_session', 'kiosk_session_id', 'kiosk_sessions', 'ON DELETE SET NULL'),
                ('restaurant_carts', 'fk_restaurant_carts_converted_to_order', 'converted_to_order_id', 'restaurant_orders', 'ON DELETE SET NULL'),
                ('restaurant_orders', 'restaurant_orders_cashier_session_id_fkey', 'cashier_session_id', 'cashier_sessions', 'ON DELETE SET NULL'),
                ('restaurant_orders', 'restaurant_orders_customer_id_fkey', 'customer_id', 'customers', 'ON DELETE SET NULL'),
                ('restaurant_orders', 'fk_restaurant_orders_kiosk_session', 'kiosk_session_id', 'kiosk_sessions', 'ON DELETE SET NULL'),
                ('restaurant_orders', 'restaurant_orders_pos_transaction_id_fkey', 'pos_transaction_id', 'pos_transactions', 'ON DELETE SET NULL'),
                ('restaurant_order_items', 'restaurant_order_items_order_id_fkey', 'order_id', 'restaurant_orders', 'ON DELETE CASCADE'),
                ('restaurant_order_status_history', 'restaurant_order_status_history_restaurant_order_id_fkey', 'restaurant_order_id', 'restaurant_orders', 'ON DELETE CASCADE'),
                ('waste_logs', 'waste_logs_order_id_fkey', 'order_id', 'restaurant_orders', 'ON DELETE SET NULL'),
                ('invoices', 'invoices_customer_id_fkey', 'customer_id', 'customers', 'ON DELETE SET NULL'),
                ('sales_returns', 'sales_returns_cashier_session_id_fkey', 'cashier_session_id', 'cashier_sessions', 'ON DELETE SET NULL'),
                ('sales_returns', 'sales_returns_customer_id_fkey', 'customer_id', 'customers', 'ON DELETE SET NULL'),
                ('sales_returns', 'sales_returns_original_transaction_id_fkey', 'original_transaction_id', 'pos_transactions', 'ON DELETE SET NULL'),
                ('quotes', 'quotes_customer_id_fkey', 'customer_id', 'customers', 'ON DELETE SET NULL'),
                ('sales_return_lines', 'sales_return_lines_return_id_fkey', 'return_id', 'sales_returns', 'ON DELETE CASCADE'),
                ('sales_return_lines', 'sales_return_lines_original_line_id_fkey', 'original_line_id', 'pos_transaction_lines', 'ON DELETE SET NULL'),
                ('customer_addresses', 'customer_addresses_customer_id_fkey', 'customer_id', 'customers', 'ON DELETE CASCADE'),
                ('customer_payments', 'customer_payments_customer_id_fkey', 'customer_id', 'customers', 'ON DELETE SET NULL')
            ) AS cfks(child_tbl, cname, col_name, parent_tbl, on_del)
        LOOP
            IF to_regclass('public.' || r.child_tbl) IS NOT NULL
               AND to_regclass('public.' || r.parent_tbl) IS NOT NULL THEN
                EXECUTE format(
                    'ALTER TABLE public.%I ADD CONSTRAINT %I FOREIGN KEY (%I) REFERENCES public.%I(id) %s',
                    r.child_tbl, r.cname, r.col_name, r.parent_tbl, r.on_del
                );
            END IF;
        END LOOP;
    END IF;
END $$;

DROP FUNCTION IF EXISTS pg_temp.det_uuid(text, bigint);
