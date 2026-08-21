-- Migration: provision the narrow tenant-local Stage 2D review permission.
-- This creates capability metadata only; it deliberately assigns the
-- permission to no role. Reviewer role/user assignment is an explicit
-- deployment decision.

INSERT INTO permissions (name, code, description, metadata)
VALUES (
    'Product Enrichment Review',
    'product_enrichment:review',
    'May review product enrichment suggestions by listing, inspecting, approving, or rejecting them.',
    '{}'::jsonb
)
ON CONFLICT (code) DO NOTHING;
