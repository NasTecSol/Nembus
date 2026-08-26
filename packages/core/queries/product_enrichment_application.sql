-- Stage 2E E1 tenant-local deterministic application operations.

-- name: LockProductEnrichmentSuggestionForApplication :one
SELECT * FROM product_enrichment_suggestions
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
FOR UPDATE;

-- name: LockProductForEnrichmentApplication :one
SELECT * FROM products
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
FOR UPDATE;

-- name: LockBrandForEnrichmentApplicationByID :one
SELECT * FROM brands
WHERE id = sqlc.arg(id)
FOR UPDATE;

-- name: LockBrandForEnrichmentApplicationByCode :one
SELECT * FROM brands
WHERE code = sqlc.arg(code)
FOR UPDATE;

-- name: LockCategoryForEnrichmentApplicationByID :one
SELECT * FROM product_categories
WHERE id = sqlc.arg(id)
FOR UPDATE;

-- name: LockCategoryForEnrichmentApplicationByCode :one
SELECT * FROM product_categories
WHERE code = sqlc.arg(code)
FOR UPDATE;

-- name: ApplyProductEnrichmentFields :execrows
UPDATE products
SET brand_id = COALESCE(sqlc.narg(brand_id)::INTEGER, brand_id),
    category_id = COALESCE(sqlc.narg(category_id)::INTEGER, category_id),
    description = COALESCE(sqlc.narg(description)::TEXT, description),
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND (sqlc.narg(brand_id)::INTEGER IS NULL OR brand_id IS NULL)
  AND (sqlc.narg(category_id)::INTEGER IS NULL OR category_id IS NULL)
  AND (sqlc.narg(description)::TEXT IS NULL OR description IS NULL OR btrim(description) = '');

-- name: ApproveAndApplyProductEnrichmentSuggestion :one
UPDATE product_enrichment_suggestions
SET status = 'applied',
    reviewer_id = sqlc.narg(reviewer_id)::INTEGER,
    reviewed_at = CURRENT_TIMESTAMP,
    applied_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND status = 'in_review'
RETURNING *;
