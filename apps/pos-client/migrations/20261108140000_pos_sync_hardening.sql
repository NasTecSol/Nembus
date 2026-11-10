-- +goose Up
-- Phase 2, 3 & 4 Core Schema Evolution: sync_inbox, change_log, change_log_meta, and pos_terminals sync token auth

ALTER TABLE "public"."pos_terminals"
    ADD COLUMN IF NOT EXISTS "sync_token_hash" character varying(64),
    ADD COLUMN IF NOT EXISTS "token_issued_at" timestamptz;

CREATE TABLE IF NOT EXISTS "public"."sync_inbox" (
    "outbox_uuid" uuid NOT NULL,
    "device_id" character varying(100) NOT NULL,
    "store_id" integer NOT NULL,
    "batch_id" character varying(100) NULL,
    "entity_type" character varying(50) NOT NULL,
    "entity_id" bigint NOT NULL,
    "applied_at" timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY ("outbox_uuid")
);

CREATE INDEX IF NOT EXISTS "idx_sync_inbox_applied_at" ON "public"."sync_inbox" ("applied_at");
CREATE INDEX IF NOT EXISTS "idx_sync_inbox_batch_id" ON "public"."sync_inbox" ("batch_id") WHERE ("batch_id" IS NOT NULL);

CREATE TABLE IF NOT EXISTS "public"."change_log" (
    "seq" bigserial NOT NULL,
    "entity_type" character varying(50) NOT NULL,
    "entity_id" bigint NOT NULL,
    "op" character varying(10) NOT NULL,
    "payload" jsonb NULL,
    "changed_at" timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY ("seq"),
    CONSTRAINT "change_log_op_check" CHECK ((op)::text = ANY ((ARRAY['INSERT'::character varying, 'UPDATE'::character varying, 'DELETE'::character varying])::text[]))
);

CREATE INDEX IF NOT EXISTS "idx_change_log_entity_seq" ON "public"."change_log" ("entity_type", "entity_id", "seq" DESC);

CREATE TABLE IF NOT EXISTS "public"."change_log_meta" (
    "id" smallint NOT NULL DEFAULT 1,
    "compacted_up_to_seq" bigint NOT NULL DEFAULT 0,
    "min_retained_seq" bigint NOT NULL DEFAULT 0,
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY ("id"),
    CONSTRAINT "change_log_meta_id_check" CHECK (id = 1)
);

INSERT INTO "public"."change_log_meta" ("id", "compacted_up_to_seq", "min_retained_seq")
VALUES (1, 0, 0)
ON CONFLICT ("id") DO NOTHING;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION "public"."trg_record_change_log"() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    v_entity_type VARCHAR(50);
    v_entity_id   BIGINT;
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
        v_entity_id := OLD.id;
        v_payload := to_jsonb(OLD);
    ELSE
        v_entity_id := NEW.id;
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
-- +goose StatementEnd

-- Master/transactional pulled entities (AFTER INSERT OR UPDATE)
DROP TRIGGER IF EXISTS trg_change_log_units_of_measure ON units_of_measure;
CREATE TRIGGER trg_change_log_units_of_measure AFTER INSERT OR UPDATE ON units_of_measure FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('units_of_measure');

DROP TRIGGER IF EXISTS trg_change_log_product_categories ON product_categories;
CREATE TRIGGER trg_change_log_product_categories AFTER INSERT OR UPDATE ON product_categories FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('product_categories');

DROP TRIGGER IF EXISTS trg_change_log_payment_terms ON payment_terms;
CREATE TRIGGER trg_change_log_payment_terms AFTER INSERT OR UPDATE ON payment_terms FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('payment_terms');

DROP TRIGGER IF EXISTS trg_change_log_users ON users;
CREATE TRIGGER trg_change_log_users AFTER INSERT OR UPDATE ON users FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('users');

DROP TRIGGER IF EXISTS trg_change_log_roles ON roles;
CREATE TRIGGER trg_change_log_roles AFTER INSERT OR UPDATE ON roles FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('roles');

DROP TRIGGER IF EXISTS trg_change_log_permissions ON permissions;
CREATE TRIGGER trg_change_log_permissions AFTER INSERT OR UPDATE ON permissions FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('permissions');

DROP TRIGGER IF EXISTS trg_change_log_cashiers ON cashiers;
CREATE TRIGGER trg_change_log_cashiers AFTER INSERT OR UPDATE ON cashiers FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('cashiers');

DROP TRIGGER IF EXISTS trg_change_log_business_partners ON business_partners;
CREATE TRIGGER trg_change_log_business_partners AFTER INSERT OR UPDATE ON business_partners FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('business_partners');

DROP TRIGGER IF EXISTS trg_change_log_price_lists ON price_lists;
CREATE TRIGGER trg_change_log_price_lists AFTER INSERT OR UPDATE ON price_lists FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('price_lists');

DROP TRIGGER IF EXISTS trg_change_log_products ON products;
CREATE TRIGGER trg_change_log_products AFTER INSERT OR UPDATE ON products FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('products');

DROP TRIGGER IF EXISTS trg_change_log_product_variants ON product_variants;
CREATE TRIGGER trg_change_log_product_variants AFTER INSERT OR UPDATE ON product_variants FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('product_variants');

DROP TRIGGER IF EXISTS trg_change_log_customers ON customers;
CREATE TRIGGER trg_change_log_customers AFTER INSERT OR UPDATE ON customers FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('customers');

DROP TRIGGER IF EXISTS trg_change_log_recipes ON recipes;
CREATE TRIGGER trg_change_log_recipes AFTER INSERT OR UPDATE ON recipes FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('recipes');

DROP TRIGGER IF EXISTS trg_change_log_menu_items ON menu_items;
CREATE TRIGGER trg_change_log_menu_items AFTER INSERT OR UPDATE ON menu_items FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('menu_items');

DROP TRIGGER IF EXISTS trg_change_log_menu_modifier_groups ON menu_modifier_groups;
CREATE TRIGGER trg_change_log_menu_modifier_groups AFTER INSERT OR UPDATE ON menu_modifier_groups FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('menu_modifier_groups');

DROP TRIGGER IF EXISTS trg_change_log_combo_bundles ON combo_bundles;
CREATE TRIGGER trg_change_log_combo_bundles AFTER INSERT OR UPDATE ON combo_bundles FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('combo_bundles');

DROP TRIGGER IF EXISTS trg_change_log_promotions ON promotions;
CREATE TRIGGER trg_change_log_promotions AFTER INSERT OR UPDATE ON promotions FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('promotions');

DROP TRIGGER IF EXISTS trg_change_log_restaurant_promotions ON restaurant_promotions;
CREATE TRIGGER trg_change_log_restaurant_promotions AFTER INSERT OR UPDATE ON restaurant_promotions FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('restaurant_promotions');

DROP TRIGGER IF EXISTS trg_change_log_stock_counts ON stock_counts;
CREATE TRIGGER trg_change_log_stock_counts AFTER INSERT OR UPDATE ON stock_counts FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('stock_counts');

DROP TRIGGER IF EXISTS trg_change_log_stock_count_lines ON stock_count_lines;
CREATE TRIGGER trg_change_log_stock_count_lines AFTER INSERT OR UPDATE ON stock_count_lines FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('stock_count_lines');

DROP TRIGGER IF EXISTS trg_change_log_inventory_stock ON inventory_stock;
CREATE TRIGGER trg_change_log_inventory_stock AFTER INSERT OR UPDATE ON inventory_stock FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('inventory_stock');

DROP TRIGGER IF EXISTS trg_change_log_stock_movements ON stock_movements;
CREATE TRIGGER trg_change_log_stock_movements AFTER INSERT OR UPDATE ON stock_movements FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('stock_movements');

DROP TRIGGER IF EXISTS trg_change_log_organizations ON organizations;
CREATE TRIGGER trg_change_log_organizations AFTER INSERT OR UPDATE ON organizations FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('organizations');

-- Join/leaf pulled entities (AFTER INSERT OR UPDATE OR DELETE)
DROP TRIGGER IF EXISTS trg_change_log_role_permissions ON role_permissions;
CREATE TRIGGER trg_change_log_role_permissions AFTER INSERT OR UPDATE OR DELETE ON role_permissions FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('role_permissions');

DROP TRIGGER IF EXISTS trg_change_log_user_roles ON user_roles;
CREATE TRIGGER trg_change_log_user_roles AFTER INSERT OR UPDATE OR DELETE ON user_roles FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('user_roles');

DROP TRIGGER IF EXISTS trg_change_log_partner_addresses ON partner_addresses;
CREATE TRIGGER trg_change_log_partner_addresses AFTER INSERT OR UPDATE OR DELETE ON partner_addresses FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('partner_addresses');

DROP TRIGGER IF EXISTS trg_change_log_user_biometric_templates ON user_biometric_templates;
CREATE TRIGGER trg_change_log_user_biometric_templates AFTER INSERT OR UPDATE OR DELETE ON user_biometric_templates FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('user_biometric_templates');

DROP TRIGGER IF EXISTS trg_change_log_product_barcodes ON product_barcodes;
CREATE TRIGGER trg_change_log_product_barcodes AFTER INSERT OR UPDATE OR DELETE ON product_barcodes FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('product_barcodes');

DROP TRIGGER IF EXISTS trg_change_log_menu_item_availability_schedules ON menu_item_availability_schedules;
CREATE TRIGGER trg_change_log_menu_item_availability_schedules AFTER INSERT OR UPDATE OR DELETE ON menu_item_availability_schedules FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('menu_item_availability_schedules');

DROP TRIGGER IF EXISTS trg_change_log_product_prices ON product_prices;
CREATE TRIGGER trg_change_log_product_prices AFTER INSERT OR UPDATE OR DELETE ON product_prices FOR EACH ROW EXECUTE FUNCTION trg_record_change_log('product_prices');

-- Seed baseline change_log in parent-before-child dependency order on Cloud databases
INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'units_of_measure', t.id, 'INSERT', to_jsonb(t) FROM units_of_measure t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'units_of_measure' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'product_categories', t.id, 'INSERT', to_jsonb(t) FROM product_categories t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'product_categories' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'payment_terms', t.id, 'INSERT', to_jsonb(t) FROM payment_terms t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'payment_terms' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'users', t.id, 'INSERT', to_jsonb(t) FROM users t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'users' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'roles', t.id, 'INSERT', to_jsonb(t) FROM roles t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'roles' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'permissions', t.id, 'INSERT', to_jsonb(t) FROM permissions t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'permissions' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'role_permissions', t.id, 'INSERT', to_jsonb(t) FROM role_permissions t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'role_permissions' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'user_roles', t.id, 'INSERT', to_jsonb(t) FROM user_roles t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'user_roles' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'cashiers', t.id, 'INSERT', to_jsonb(t) FROM cashiers t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'cashiers' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'business_partners', t.id, 'INSERT', to_jsonb(t) FROM business_partners t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'business_partners' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'partner_addresses', t.id, 'INSERT', to_jsonb(t) FROM partner_addresses t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'partner_addresses' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'user_biometric_templates', t.id, 'INSERT', to_jsonb(t) FROM user_biometric_templates t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'user_biometric_templates' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'price_lists', t.id, 'INSERT', to_jsonb(t) FROM price_lists t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'price_lists' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'products', t.id, 'INSERT', to_jsonb(t) FROM products t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'products' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'product_variants', t.id, 'INSERT', to_jsonb(t) FROM product_variants t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'product_variants' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'product_barcodes', t.id, 'INSERT', to_jsonb(t) FROM product_barcodes t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'product_barcodes' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'customers', t.id, 'INSERT', to_jsonb(t) FROM customers t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'customers' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'recipes', t.id, 'INSERT', to_jsonb(t) FROM recipes t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'recipes' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'menu_items', t.id, 'INSERT', to_jsonb(t) FROM menu_items t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'menu_items' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'menu_modifier_groups', t.id, 'INSERT', to_jsonb(t) FROM menu_modifier_groups t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'menu_modifier_groups' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'combo_bundles', t.id, 'INSERT', to_jsonb(t) FROM combo_bundles t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'combo_bundles' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'menu_item_availability_schedules', t.id, 'INSERT', to_jsonb(t) FROM menu_item_availability_schedules t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'menu_item_availability_schedules' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'promotions', t.id, 'INSERT', to_jsonb(t) FROM promotions t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'promotions' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'restaurant_promotions', t.id, 'INSERT', to_jsonb(t) FROM restaurant_promotions t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'restaurant_promotions' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'product_prices', t.id, 'INSERT', to_jsonb(t) FROM product_prices t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'product_prices' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'stock_counts', t.id, 'INSERT', to_jsonb(t) FROM stock_counts t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'stock_counts' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'stock_count_lines', t.id, 'INSERT', to_jsonb(t) FROM stock_count_lines t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'stock_count_lines' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'inventory_stock', t.id, 'INSERT', to_jsonb(t) FROM inventory_stock t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'inventory_stock' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'stock_movements', t.id, 'INSERT', to_jsonb(t) FROM stock_movements t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'stock_movements' AND c.entity_id = t.id)
ORDER BY t.id;

INSERT INTO change_log (entity_type, entity_id, op, payload)
SELECT 'organizations', t.id, 'INSERT', to_jsonb(t) FROM organizations t
WHERE to_regclass('public.local_device_config') IS NULL
  AND NOT EXISTS (SELECT 1 FROM change_log c WHERE c.entity_type = 'organizations' AND c.entity_id = t.id)
ORDER BY t.id;

UPDATE change_log_meta
SET compacted_up_to_seq = COALESCE((SELECT MAX(seq) FROM change_log), 0),
    updated_at = NOW()
WHERE id = 1;
