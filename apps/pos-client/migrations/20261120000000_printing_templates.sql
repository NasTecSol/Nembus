-- +goose Up
-- Idempotent forward migration for organization printing templates and store receipt settings.
CREATE TABLE IF NOT EXISTS "public"."printing_templates" (
    id BIGSERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL,
    name VARCHAR(120) NOT NULL,
    type VARCHAR(32) NOT NULL DEFAULT 'receipt',
    language VARCHAR(16) NOT NULL DEFAULT 'en',
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_default BOOLEAN NOT NULL DEFAULT false,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, id),
    UNIQUE (organization_id, name, type, language),
    CHECK (jsonb_typeof(config) = 'object')
);
CREATE INDEX IF NOT EXISTS "idx_printing_templates_org_active"
    ON "public"."printing_templates" (organization_id, type, language, is_active, name);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_printing_templates_org_default"
    ON "public"."printing_templates" (organization_id, type, language) WHERE is_default;

CREATE TABLE IF NOT EXISTS "public"."store_receipt_settings" (
    id BIGSERIAL PRIMARY KEY,
    store_id INTEGER NOT NULL,
    organization_id INTEGER NOT NULL,
    type VARCHAR(32) NOT NULL DEFAULT 'receipt',
    language VARCHAR(16) NOT NULL DEFAULT 'en',
    default_template_id BIGINT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (store_id, type, language),
    FOREIGN KEY (organization_id, default_template_id)
        REFERENCES "public"."printing_templates" (organization_id, id) ON DELETE RESTRICT
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION "public"."seed_default_printing_template_for_organization"()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM "public"."printing_templates"
        WHERE organization_id=NEW.id AND type='receipt' AND language='en' AND is_default
    ) THEN
        INSERT INTO "public"."printing_templates" (organization_id, name, type, language, config, is_default)
        VALUES (
            NEW.id,
            'Standard receipt',
            'receipt',
            'en',
            '{"version":1,"paper_width_mm":80,"header_lines":[],"show_organization_name":true,"show_legal_name":true,"show_tax_id":true,"show_receipt_number":true,"show_date":true,"show_cashier":true,"show_terminal":true,"show_customer":true,"show_items":true,"show_subtotal":true,"show_discount":true,"show_tax":true,"show_total":true,"show_payment_method":true,"show_paid_change":true,"show_barcode":true,"footer_lines":["Thank you for your visit!","Please keep this receipt for any exchange or return."]}'::jsonb,
            true
        )
        ON CONFLICT DO NOTHING;
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS "trg_organizations_seed_printing_template" ON "public"."organizations";
CREATE TRIGGER "trg_organizations_seed_printing_template"
AFTER INSERT ON "public"."organizations"
FOR EACH ROW EXECUTE FUNCTION "public"."seed_default_printing_template_for_organization"();

INSERT INTO "public"."printing_templates" (organization_id, name, type, language, config, is_default)
SELECT o.id, 'Standard receipt', 'receipt', 'en',
  '{"version":1,"paper_width_mm":80,"header_lines":[],"show_organization_name":true,"show_legal_name":true,"show_tax_id":true,"show_receipt_number":true,"show_date":true,"show_cashier":true,"show_terminal":true,"show_customer":true,"show_items":true,"show_subtotal":true,"show_discount":true,"show_tax":true,"show_total":true,"show_payment_method":true,"show_paid_change":true,"show_barcode":true,"footer_lines":["Thank you for your visit!","Please keep this receipt for any exchange or return."]}'::jsonb,
  true
FROM "public"."organizations" o
WHERE NOT EXISTS (
    SELECT 1 FROM "public"."printing_templates" t
    WHERE t.organization_id=o.id AND t.type='receipt' AND t.language='en' AND t.is_default
)
AND NOT EXISTS (
    SELECT 1 FROM "public"."printing_templates" t
    WHERE t.organization_id=o.id AND t.name='Standard receipt' AND t.type='receipt' AND t.language='en'
);

UPDATE "public"."printing_templates" t SET is_default=true
WHERE t.is_active AND t.type='receipt' AND t.language='en'
  AND NOT EXISTS (
      SELECT 1 FROM "public"."printing_templates" d
      WHERE d.organization_id=t.organization_id AND d.type=t.type AND d.language=t.language AND d.is_default
  )
  AND t.id=(
      SELECT min(x.id) FROM "public"."printing_templates" x
      WHERE x.organization_id=t.organization_id AND x.type=t.type AND x.language=t.language AND x.is_active
  );

DROP TRIGGER IF EXISTS "trg_change_log_printing_templates" ON "public"."printing_templates";
CREATE TRIGGER "trg_change_log_printing_templates" AFTER INSERT OR UPDATE ON "public"."printing_templates"
FOR EACH ROW EXECUTE FUNCTION "public"."trg_record_change_log"('printing_templates');

DROP TRIGGER IF EXISTS "trg_change_log_store_receipt_settings" ON "public"."store_receipt_settings";
CREATE TRIGGER "trg_change_log_store_receipt_settings" AFTER INSERT OR UPDATE ON "public"."store_receipt_settings"
FOR EACH ROW EXECUTE FUNCTION "public"."trg_record_change_log"('store_receipt_settings');
