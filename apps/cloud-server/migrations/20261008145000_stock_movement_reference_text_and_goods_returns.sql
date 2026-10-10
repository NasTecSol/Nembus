-- Migration: Support polymorphic text reference_id on stock_movements, add goods_returns and purchase_invoices tables, and auto-link trigger

-- 1. Create goods_returns table
CREATE TABLE IF NOT EXISTS "public"."goods_returns" (
    "id" serial NOT NULL,
    "organization_id" integer NULL,
    "return_number" character varying(50) NOT NULL,
    "grn_id" integer NULL,
    "partners_id" integer NULL,
    "store_id" integer NULL,
    "return_date" timestamp without time zone NULL DEFAULT CURRENT_TIMESTAMP,
    "status" character varying(50) NULL DEFAULT 'completed'::character varying,
    "notes" text NULL,
    "metadata" jsonb NULL DEFAULT '{}'::jsonb,
    "created_at" timestamp without time zone NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp without time zone NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id"),
    CONSTRAINT "goods_returns_return_number_key" UNIQUE ("return_number"),
    CONSTRAINT "goods_returns_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT "goods_returns_grn_id_fkey" FOREIGN KEY ("grn_id") REFERENCES "public"."goods_receipt_notes" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
    CONSTRAINT "goods_returns_partners_id_fkey" FOREIGN KEY ("partners_id") REFERENCES "public"."business_partners" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
    CONSTRAINT "goods_returns_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS "idx_goods_returns_return_number" ON "public"."goods_returns" ("return_number");
CREATE INDEX IF NOT EXISTS "idx_goods_returns_store_id" ON "public"."goods_returns" ("store_id");

-- 2. Create purchase_invoices table
CREATE TABLE IF NOT EXISTS "public"."purchase_invoices" (
    "id" serial NOT NULL,
    "organization_id" integer NULL,
    "invoice_number" character varying(50) NOT NULL,
    "partners_id" integer NULL,
    "store_id" integer NULL,
    "invoice_date" timestamp without time zone NULL DEFAULT CURRENT_TIMESTAMP,
    "status" character varying(50) NULL DEFAULT 'posted'::character varying,
    "notes" text NULL,
    "metadata" jsonb NULL DEFAULT '{}'::jsonb,
    "created_at" timestamp without time zone NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp without time zone NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id"),
    CONSTRAINT "purchase_invoices_invoice_number_key" UNIQUE ("invoice_number"),
    CONSTRAINT "purchase_invoices_organization_id_fkey" FOREIGN KEY ("organization_id") REFERENCES "public"."organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT "purchase_invoices_partners_id_fkey" FOREIGN KEY ("partners_id") REFERENCES "public"."business_partners" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
    CONSTRAINT "purchase_invoices_store_id_fkey" FOREIGN KEY ("store_id") REFERENCES "public"."stores" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS "idx_purchase_invoices_invoice_number" ON "public"."purchase_invoices" ("invoice_number");
CREATE INDEX IF NOT EXISTS "idx_purchase_invoices_store_id" ON "public"."purchase_invoices" ("store_id");

-- 3. Alter stock_movements.reference_id from INTEGER to TEXT
DROP VIEW IF EXISTS "public"."vw_stock_movement_ledger";
ALTER TABLE "public"."stock_movements" ALTER COLUMN "reference_id" TYPE text;

-- 4. Recreate vw_stock_movement_ledger
CREATE OR REPLACE VIEW "public"."vw_stock_movement_ledger" AS
 SELECT sm.id AS movement_id,
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
        CASE
            WHEN sm.quantity > 0::numeric THEN sm.quantity
            ELSE 0::numeric
        END AS quantity_in,
        CASE
            WHEN sm.quantity < 0::numeric THEN abs(sm.quantity)
            ELSE 0::numeric
        END AS quantity_out,
    sm.quantity AS net_quantity,
    COALESCE(sm.cost_per_unit, p.cost_price, 0::numeric) AS unit_cost,
    COALESCE(sm.total_value, sm.quantity * COALESCE(sm.cost_per_unit, p.cost_price, 0::numeric)) AS movement_value,
    sum(sm.quantity) OVER (PARTITION BY (COALESCE(sm.to_store_id, sm.from_store_id)), sm.product_id ORDER BY sm.movement_date, sm.id ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS running_quantity,
    sum(sm.quantity) OVER (PARTITION BY (COALESCE(sm.to_store_id, sm.from_store_id)), sm.product_id ORDER BY sm.movement_date, sm.id ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) * COALESCE(sm.cost_per_unit, p.cost_price, 0::numeric) AS running_valuation,
    sm.status,
    sm.metadata
   FROM "public"."stock_movements" sm
     JOIN "public"."products" p ON p.id = sm.product_id
     LEFT JOIN "public"."product_barcodes pb ON pb.product_id = p.id AND pb.is_primary = true
     LEFT JOIN "public"."product_categories" pc ON pc.id = p.category_id
     LEFT JOIN "public"."stores" s ON s.id = COALESCE(sm.to_store_id, sm.from_store_id);

-- 5. Trigger for automatic on-the-fly linking
CREATE OR REPLACE FUNCTION "public"."fn_auto_link_stock_movement_reference"()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.reference_id IS NULL AND NEW.metadata ? 'sap_base_ref' THEN
        CASE NEW.reference_type
            WHEN 'invoice' THEN
                SELECT id::text INTO NEW.reference_id
                FROM public.invoices
                WHERE invoice_number = 'INV-SAP-' || (NEW.metadata->>'sap_base_ref')
                LIMIT 1;

            WHEN 'goods_receipt_note' THEN
                SELECT id::text INTO NEW.reference_id
                FROM public.goods_receipt_notes
                WHERE grn_number = 'GRN-' || (NEW.metadata->>'sap_base_ref')
                LIMIT 1;

            WHEN 'purchase_credit_note' THEN
                SELECT id::text INTO NEW.reference_id
                FROM public.sales_returns
                WHERE return_number = 'CN-SAP-' || (NEW.metadata->>'sap_base_ref')
                LIMIT 1;

            WHEN 'goods_return' THEN
                SELECT id::text INTO NEW.reference_id
                FROM public.goods_returns
                WHERE return_number = 'GRTN-SAP-' || (NEW.metadata->>'sap_base_ref')
                LIMIT 1;

            WHEN 'purchase_invoice' THEN
                SELECT id::text INTO NEW.reference_id
                FROM public.purchase_invoices
                WHERE invoice_number = 'PI-SAP-' || (NEW.metadata->>'sap_base_ref')
                LIMIT 1;

            WHEN 'stock_count' THEN
                SELECT id::text INTO NEW.reference_id
                FROM public.stock_counts
                WHERE count_number = 'SC-SAP-' || (NEW.metadata->>'sap_base_ref')
                LIMIT 1;

            ELSE
                -- no-op
        END CASE;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS "trg_auto_link_stock_movement_reference" ON "public"."stock_movements";
CREATE TRIGGER "trg_auto_link_stock_movement_reference"
BEFORE INSERT OR UPDATE ON "public"."stock_movements"
FOR EACH ROW
EXECUTE FUNCTION "public"."fn_auto_link_stock_movement_reference"();
