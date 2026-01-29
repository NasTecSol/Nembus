-- name: GetPOSProductsWithStock :many
SELECT
    cat.*,
    COALESCE(inv.quantity_available, 0.0)::numeric as quantity_available,
    COALESCE(inv.quantity_on_hand, 0.0)::numeric as quantity_on_hand,
    COALESCE(inv.quantity_allocated, 0.0)::numeric as quantity_allocated,
    CASE WHEN COALESCE(inv.quantity_available, 0) > 0 THEN true ELSE false END as is_in_stock,
    CASE
        WHEN COALESCE(inv.quantity_available, 0) <= COALESCE(inv.reorder_level, 0)
             AND COALESCE(inv.quantity_available, 0) > 0
        THEN true
        ELSE false
    END as is_low_stock,
    COALESCE(inv.reorder_level, 0.0)::numeric as reorder_level
FROM vw_pos_product_catalog cat
LEFT JOIN inventory_stock inv
    ON cat.product_id = inv.product_id
    AND inv.store_id = sqlc.arg('store_id')
WHERE
    (sqlc.narg('category_id')::int IS NULL OR cat.category_id = sqlc.narg('category_id'))
    AND (sqlc.narg('search_term')::text IS NULL
         OR cat.product_name ILIKE '%' || sqlc.narg('search_term') || '%'
         OR cat.sku ILIKE '%' || sqlc.narg('search_term') || '%'
         OR cat.barcode ILIKE '%' || sqlc.narg('search_term') || '%')
    AND (sqlc.arg('include_out_of_stock')::boolean = true OR COALESCE(inv.quantity_available, 0) > 0)
ORDER BY cat.category_name, cat.product_name;

-- name: GetPOSProductByBarcode :one
SELECT
    cat.*,
    COALESCE(inv.quantity_available, 0.0)::numeric as quantity_available,
    CASE WHEN COALESCE(inv.quantity_available, 0) > 0 THEN true ELSE false END as is_in_stock
FROM vw_pos_product_catalog cat
LEFT JOIN inventory_stock inv
    ON cat.product_id = inv.product_id
    AND inv.store_id = sqlc.arg('store_id')
WHERE cat.barcode = sqlc.arg('barcode')
LIMIT 1;

-- name: GetPOSProductByID :one
SELECT
    cat.*,
    COALESCE(inv.quantity_available, 0.0)::numeric as quantity_available,
    CASE WHEN COALESCE(inv.quantity_available, 0) > 0 THEN true ELSE false END as is_in_stock
FROM vw_pos_product_catalog cat
LEFT JOIN inventory_stock inv
    ON cat.product_id = inv.product_id
    AND inv.store_id = sqlc.arg('store_id')
WHERE cat.product_id = sqlc.arg('product_id')
LIMIT 1;

-- name: SearchPOSProducts :many
SELECT
    cat.*,
    COALESCE(inv.quantity_available, 0.0)::numeric as quantity_available,
    CASE WHEN COALESCE(inv.quantity_available, 0) > 0 THEN true ELSE false END as is_in_stock,
    CASE
        WHEN cat.sku ILIKE sqlc.arg('search_term') THEN 100
        WHEN cat.product_name ILIKE sqlc.arg('search_term') THEN 90
        WHEN cat.barcode = sqlc.arg('search_term') THEN 95
        WHEN cat.sku ILIKE sqlc.arg('search_term') || '%' THEN 80
        WHEN cat.product_name ILIKE sqlc.arg('search_term') || '%' THEN 70
        WHEN cat.sku ILIKE '%' || sqlc.arg('search_term') || '%' THEN 60
        WHEN cat.product_name ILIKE '%' || sqlc.arg('search_term') || '%' THEN 50
        ELSE 40
    END as relevance_score
FROM vw_pos_product_catalog cat
LEFT JOIN inventory_stock inv
    ON cat.product_id = inv.product_id
    AND inv.store_id = sqlc.arg('store_id')
WHERE
    (cat.product_name ILIKE '%' || sqlc.arg('search_term') || '%'
     OR cat.sku ILIKE '%' || sqlc.arg('search_term') || '%'
     OR cat.barcode ILIKE '%' || sqlc.arg('search_term') || '%'
     OR (CASE WHEN sqlc.arg('search_term')::text ~ '^\d+$' THEN cat.product_id = sqlc.arg('search_term')::int ELSE false END))
ORDER BY relevance_score DESC, cat.product_name
LIMIT sqlc.arg('limit_count');

-- name: GetPOSCategories :many
SELECT * FROM vw_pos_categories;

-- name: GetPOSPromotedProducts :many
SELECT
    cat.*,
    (cat.retail_price - cat.promo_price)::numeric as discount_amount,
    COALESCE(inv.quantity_available, 0.0)::numeric as quantity_available,
    CASE WHEN COALESCE(inv.quantity_available, 0) > 0 THEN true ELSE false END as is_in_stock
FROM vw_pos_product_catalog cat
LEFT JOIN inventory_stock inv
    ON cat.product_id = inv.product_id
    AND (sqlc.narg('store_id')::int IS NULL OR inv.store_id = sqlc.narg('store_id'))
WHERE cat.has_active_promotion = true
  AND (sqlc.narg('store_id')::int IS NULL OR COALESCE(inv.quantity_available, 0) > 0)
ORDER BY discount_amount DESC;
