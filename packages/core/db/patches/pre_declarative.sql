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
