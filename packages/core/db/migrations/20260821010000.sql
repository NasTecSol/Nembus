-- Migration: provision the narrow tenant-local Stage 2E apply permission.
-- This creates capability metadata only; it deliberately assigns the
-- permission to no role. Application role/user assignment is an explicit
-- deployment decision and is separate from product enrichment review access.

INSERT INTO permissions (name, code, description, metadata)
VALUES (
    'Product Enrichment Apply',
    'product_enrichment:apply',
    'May apply an already-approved product enrichment suggestion using the deterministic Stage 2E rules.',
    '{}'::jsonb
)
ON CONFLICT (code) DO NOTHING;
