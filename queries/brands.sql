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