-- Organization-owned print layouts. An organization selects one default per
-- template type and language; stores can override each type/language pair.
CREATE TABLE printing_templates (
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

CREATE INDEX idx_printing_templates_org_active
    ON printing_templates (organization_id, type, language, is_active, name);
CREATE UNIQUE INDEX idx_printing_templates_org_default
    ON printing_templates (organization_id, type, language) WHERE is_default;

CREATE TABLE store_receipt_settings (
    id BIGSERIAL PRIMARY KEY,
    store_id INTEGER NOT NULL,
    organization_id INTEGER NOT NULL,
    type VARCHAR(32) NOT NULL DEFAULT 'receipt',
    language VARCHAR(16) NOT NULL DEFAULT 'en',
    default_template_id BIGINT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (store_id, type, language),
    FOREIGN KEY (organization_id, default_template_id)
        REFERENCES printing_templates (organization_id, id) ON DELETE RESTRICT
);
