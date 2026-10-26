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

-- product_prices: drop obsolete narrow unique if a prior deploy attempted it (declarative schema now uses uq_product_prices_natural_key).
ALTER TABLE public.product_prices
    DROP CONSTRAINT IF EXISTS uq_product_prices_product_price_list;
