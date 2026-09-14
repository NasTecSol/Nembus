-- Migration: 20260913014500_combo_bundle_promotions.sql
-- Upgrade combo_bundles to support Organization-wide, Retail, and Wholesale scopes

ALTER TABLE combo_bundles
    ADD COLUMN IF NOT EXISTS organization_id INTEGER REFERENCES organizations(id) ON DELETE CASCADE,
    ALTER COLUMN store_id DROP NOT NULL,
    ADD COLUMN IF NOT EXISTS applicable_customer_types TEXT[] DEFAULT '{"retail","wholesale"}',
    ADD COLUMN IF NOT EXISTS price_list_id INTEGER REFERENCES price_lists(id) ON DELETE SET NULL;

-- Backfill organization_id from stores if store_id is present and organization_id is NULL
UPDATE combo_bundles cb
SET organization_id = s.organization_id
FROM stores s
WHERE cb.store_id = s.id AND cb.organization_id IS NULL;

-- Fallback for any orphaned rows
UPDATE combo_bundles SET organization_id = 1 WHERE organization_id IS NULL;

ALTER TABLE combo_bundles
    ALTER COLUMN organization_id SET NOT NULL;

-- Recreate unique constraint for organization-wide uniqueness
ALTER TABLE combo_bundles DROP CONSTRAINT IF EXISTS combo_bundles_store_id_code_key;
ALTER TABLE combo_bundles DROP CONSTRAINT IF EXISTS combo_bundles_organization_id_code_key;
ALTER TABLE combo_bundles ADD CONSTRAINT combo_bundles_organization_id_code_key UNIQUE(organization_id, code);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_combo_bundles_org_id ON combo_bundles(organization_id);
CREATE INDEX IF NOT EXISTS idx_combo_bundles_price_list_id ON combo_bundles(price_list_id);
