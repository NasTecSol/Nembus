-- =====================================================
-- BRANDS QUERIES (SQLC)
-- =====================================================
-- File: queries/brands.sql
-- Purpose: Comprehensive CRUD and query operations for brands table

-- =====================================================
-- CREATE OPERATIONS
-- =====================================================

-- name: CreateBrand :one
-- Create a new brand
INSERT INTO brands (
    name,
    code,
    is_active,
    metadata
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: CreateBrandWithDefaults :one
-- Create a new brand with default active status
INSERT INTO brands (
    name,
    code,
    is_active
) VALUES (
    $1, $2, true
)
RETURNING *;

-- =====================================================
-- READ OPERATIONS
-- =====================================================

-- name: GetBrandByID :one
-- Get a single brand by its ID
SELECT * FROM brands
WHERE id = $1;

-- name: GetBrandByCode :one
-- Get a single brand by its unique code
SELECT * FROM brands
WHERE code = $1;

-- name: ListAllBrands :many
-- List all brands without pagination
SELECT * FROM brands
ORDER BY name;

-- name: ListActiveBrands :many
-- List only active brands
SELECT * FROM brands
WHERE is_active = true
ORDER BY name;

-- name: ListBrands :many
-- List brands with pagination
SELECT * FROM brands
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: ListActiveBrandsWithPagination :many
-- List active brands with pagination
SELECT * FROM brands
WHERE is_active = true
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: CountBrands :one
-- Count total number of brands
SELECT COUNT(*) FROM brands;

-- name: CountActiveBrands :one
-- Count total number of active brands
SELECT COUNT(*) FROM brands
WHERE is_active = true;

-- name: SearchBrands :many
-- Search brands by name or code (case-insensitive)
SELECT * FROM brands
WHERE 
    LOWER(name) LIKE LOWER($1) OR 
    LOWER(code) LIKE LOWER($1)
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: SearchActiveBrands :many
-- Search active brands by name or code (case-insensitive)
SELECT * FROM brands
WHERE 
    is_active = true AND
    (LOWER(name) LIKE LOWER($1) OR LOWER(code) LIKE LOWER($1))
ORDER BY name
LIMIT $2 OFFSET $3;

-- =====================================================
-- UPDATE OPERATIONS
-- =====================================================

-- name: UpdateBrand :one
-- Update brand information
UPDATE brands
SET
    name = $2,
    code = $3,
    is_active = $4,
    metadata = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateBrandName :one
-- Update only the brand name
UPDATE brands
SET
    name = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateBrandCode :one
-- Update only the brand code
UPDATE brands
SET
    code = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateBrandMetadata :one
-- Update only the brand metadata
UPDATE brands
SET
    metadata = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ActivateBrand :one
-- Activate a brand
UPDATE brands
SET
    is_active = true,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeactivateBrand :one
-- Deactivate a brand
UPDATE brands
SET
    is_active = false,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ToggleBrandStatus :one
-- Toggle brand active status
UPDATE brands
SET
    is_active = NOT is_active,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- =====================================================
-- DELETE OPERATIONS
-- =====================================================

-- name: DeleteBrand :exec
-- Hard delete a brand (use with caution - may fail if referenced by products)
DELETE FROM brands
WHERE id = $1;

-- name: DeleteBrandByCode :exec
-- Hard delete a brand by code (use with caution)
DELETE FROM brands
WHERE code = $1;

-- name: SoftDeleteBrand :one
-- Soft delete a brand by deactivating it
UPDATE brands
SET
    is_active = false,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- =====================================================
-- VALIDATION & EXISTENCE CHECKS
-- =====================================================

-- name: BrandExists :one
-- Check if a brand exists by ID
SELECT EXISTS(
    SELECT 1 FROM brands WHERE id = $1
);

-- name: BrandCodeExists :one
-- Check if a brand code already exists (useful for uniqueness validation)
SELECT EXISTS(
    SELECT 1 FROM brands WHERE code = $1
);

-- name: BrandCodeExistsExcludingID :one
-- Check if a brand code exists excluding a specific brand ID (for updates)
SELECT EXISTS(
    SELECT 1 FROM brands 
    WHERE code = $1 AND id != $2
);

-- =====================================================
-- REPORTING & ANALYTICS
-- =====================================================

-- name: GetBrandWithProductCount :one
-- Get brand with count of associated products
SELECT 
    b.*,
    COUNT(p.id) as product_count
FROM brands b
LEFT JOIN products p ON p.brand_id = b.id
WHERE b.id = $1
GROUP BY b.id;

-- name: ListBrandsWithProductCounts :many
-- List all brands with their product counts
SELECT 
    b.*,
    COUNT(p.id) as product_count
FROM brands b
LEFT JOIN products p ON p.brand_id = b.id
GROUP BY b.id
ORDER BY b.name;

-- name: ListActiveBrandsWithProductCounts :many
-- List active brands with their product counts
SELECT 
    b.*,
    COUNT(p.id) as product_count
FROM brands b
LEFT JOIN products p ON p.brand_id = b.id AND p.is_active = true
WHERE b.is_active = true
GROUP BY b.id
ORDER BY b.name;

-- name: GetTopBrandsByProductCount :many
-- Get top N brands by number of products
SELECT 
    b.*,
    COUNT(p.id) as product_count
FROM brands b
LEFT JOIN products p ON p.brand_id = b.id
WHERE b.is_active = true
GROUP BY b.id
ORDER BY product_count DESC
LIMIT $1;

-- name: GetBrandsWithNoProducts :many
-- Get brands that have no associated products
SELECT b.*
FROM brands b
LEFT JOIN products p ON p.brand_id = b.id
WHERE p.id IS NULL
ORDER BY b.name;

-- name: GetInactiveBrandsWithActiveProducts :many
-- Get inactive brands that still have active products (data quality check)
SELECT DISTINCT
    b.*,
    COUNT(p.id) as active_product_count
FROM brands b
INNER JOIN products p ON p.brand_id = b.id
WHERE b.is_active = false AND p.is_active = true
GROUP BY b.id
ORDER BY b.name;

-- =====================================================
-- BULK OPERATIONS
-- =====================================================

-- name: BulkActivateBrands :exec
-- Activate multiple brands by IDs
UPDATE brands
SET
    is_active = true,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ANY($1::int[]);

-- name: BulkDeactivateBrands :exec
-- Deactivate multiple brands by IDs
UPDATE brands
SET
    is_active = false,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ANY($1::int[]);

-- name: BulkDeleteBrands :exec
-- Delete multiple brands by IDs
DELETE FROM brands
WHERE id = ANY($1::int[]);

-- =====================================================
-- AUDIT & METADATA QUERIES
-- =====================================================

-- name: GetRecentlyCreatedBrands :many
-- Get brands created in the last N days
SELECT * FROM brands
WHERE created_at >= CURRENT_DATE - $1::interval
ORDER BY created_at DESC;

-- name: GetRecentlyUpdatedBrands :many
-- Get brands updated in the last N days
SELECT * FROM brands
WHERE updated_at >= CURRENT_DATE - $1::interval
ORDER BY updated_at DESC;

-- name: GetBrandsByCreationDate :many
-- Get brands created between two dates
SELECT * FROM brands
WHERE created_at::date BETWEEN $1 AND $2
ORDER BY created_at DESC;

-- name: GetBrandMetadataByKey :one
-- Get a specific metadata field from a brand
SELECT 
    id,
    name,
    code,
    metadata->$2 as metadata_value
FROM brands
WHERE id = $1;


-- name: ListBrandsWithStats :many
SELECT 
    b.id, b.name, b.code, b.is_active,
    COUNT(p.id)           AS product_count,
    COUNT(DISTINCT p.category_id) AS category_count
FROM brands b
LEFT JOIN products p ON p.brand_id = b.id AND p.is_active = true
WHERE b.is_active = sqlc.arg('active_only')
  AND (sqlc.arg('search')::text IS NULL OR b.name ILIKE '%' || sqlc.arg('search') || '%')
GROUP BY b.id
ORDER BY product_count DESC, b.name
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');