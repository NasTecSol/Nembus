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
SET brand_id = COALESCE(sqlc.narg(brand_id), brand_id),
    category_id = COALESCE(sqlc.narg(category_id), category_id),
    description = COALESCE(sqlc.narg(description), description),
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND (sqlc.narg(brand_id) IS NULL OR brand_id IS NULL)
  AND (sqlc.narg(category_id) IS NULL OR category_id IS NULL)
  AND (sqlc.narg(description) IS NULL OR description IS NULL OR btrim(description) = '');
