-- name: GetCategoryWithPath :one
WITH RECURSIVE cat_path AS (
    SELECT pc.id, pc.parent_category_id, pc.name, pc.code, ARRAY[pc.id] AS path
    FROM product_categories pc
    WHERE pc.id = $1
    
    UNION ALL
    
    SELECT c.id, c.parent_category_id, c.name, c.code, cp.path || c.id
    FROM product_categories c
    INNER JOIN cat_path cp ON c.id = cp.parent_category_id
),
filtered_path AS (
    SELECT id, name, code, parent_category_id, path
    FROM cat_path
    WHERE id = $1
)
SELECT 
    fp.id, fp.name, fp.code, fp.parent_category_id,
    array_length(fp.path, 1) - 1 AS depth,
    parent_cat.name AS parent_name
FROM filtered_path fp
LEFT JOIN product_categories parent_cat ON parent_cat.id = (fp.path[array_length(fp.path,1)-2]);

-- name: ListCategoriesWithUsageCount :many
SELECT 
    c.id, c.name, c.code, c.parent_category_id, c.is_active,
    COUNT(p.id) AS product_count
FROM product_categories c
LEFT JOIN products p ON p.category_id = c.id AND p.is_active = true
WHERE c.is_active = sqlc.arg('active_only')
GROUP BY c.id
ORDER BY product_count DESC, c.name
LIMIT sqlc.arg('limit');

-- name: GetLeafCategories :many
SELECT c.*
FROM product_categories c
WHERE c.is_active = true
  AND NOT EXISTS (
      SELECT 1 FROM product_categories child 
      WHERE child.parent_category_id = c.id AND child.is_active = true
  )
ORDER BY c.name;