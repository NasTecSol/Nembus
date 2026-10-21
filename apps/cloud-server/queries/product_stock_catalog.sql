-- name: GetMasterProductStockCatalog :many
WITH scoped_stock AS (
    SELECT s.*
    FROM inventory_stock s
    JOIN stores st ON st.id = s.store_id
    WHERE st.organization_id = sqlc.arg('organization_id')::int
      AND (sqlc.arg('store_id')::int = 0 OR s.store_id = sqlc.arg('store_id'))
      AND (sqlc.arg('storage_location_id')::int = 0 OR s.storage_location_id = sqlc.arg('storage_location_id'))
), stock_totals AS (
    SELECT product_id, SUM(COALESCE(quantity_available, 0)) AS available_stock
    FROM scoped_stock
    GROUP BY product_id
), matching_products AS (
    SELECT p.id AS product_id, COALESCE(t.available_stock, 0)::numeric AS available_stock
    FROM products p
    LEFT JOIN stock_totals t ON t.product_id = p.id
    WHERE p.organization_id = sqlc.arg('organization_id')
      AND p.track_inventory = true
      AND (sqlc.arg('category_id')::int = 0 OR p.category_id = sqlc.arg('category_id'))
      AND ((sqlc.arg('store_id') = 0 AND sqlc.arg('storage_location_id') = 0) OR t.product_id IS NOT NULL)
      AND (
          sqlc.arg('stock_status')::text = 'all'
          OR (sqlc.arg('stock_status') = 'low' AND COALESCE(t.available_stock, 0) < sqlc.arg('low_stock_threshold')::numeric)
          OR (sqlc.arg('stock_status') = 'non_low' AND COALESCE(t.available_stock, 0) >= sqlc.arg('low_stock_threshold')::numeric)
      )
)
 , page_products AS (
    SELECT * FROM matching_products
    ORDER BY product_id
    LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset')
)
SELECT
    v.product_id,
    v.sku,
    v.name,
    v.description,
    v.product_type,
    v.is_serialized,
    v.is_batch_managed,
    v.is_active,
    v.is_sellable,
    v.is_purchasable,
    v.allow_decimal_quantity,
    v.track_inventory,
    v.metadata,
    v.created_at,
    v.updated_at,
    v.category_id,
    v.category_name,
    v.category_code,
    v.brand_id,
    v.brand_name,
    v.brand_code,
    v.tax_category_id,
    v.tax_category_name,
    v.tax_rate,
    v.tax_inclusive,
    v.base_uom_id,
    v.base_uom_code,
    v.base_uom_name,
    v.uom_conversions,
    v.prices,
    v.variants,
    v.barcodes,
    COALESCE((
        SELECT jsonb_agg(jsonb_build_object(
            'stock_id', s.id,
            'store_id', s.store_id,
            'storage_location_id', s.storage_location_id,
            'storage_location_name', sl.name,
            'storage_location_code', sl.code,
            'product_variant_id', s.product_variant_id,
            'quantity_on_hand', s.quantity_on_hand,
            'quantity_available', s.quantity_available,
            'quantity_allocated', s.quantity_allocated,
            'quantity_on_order', s.quantity_on_order,
            'reorder_level', s.reorder_level,
            'reorder_quantity', s.reorder_quantity,
            'max_stock_level', s.max_stock_level
        ) ORDER BY s.store_id, s.storage_location_id NULLS FIRST, s.id)
        FROM scoped_stock s
        LEFT JOIN storage_locations sl ON sl.id = s.storage_location_id
        WHERE s.product_id = v.product_id
    ), '[]'::jsonb) AS inventory,
    mp.available_stock,
    (mp.available_stock < sqlc.arg('low_stock_threshold')::numeric)::boolean AS is_low_stock
FROM page_products mp
JOIN v_master_product_catalog v ON v.product_id = mp.product_id
ORDER BY v.product_id;

-- name: GetMasterProductStockCatalogCount :one
WITH scoped_stock AS (
    SELECT s.*
    FROM inventory_stock s
    JOIN stores st ON st.id = s.store_id
    WHERE st.organization_id = sqlc.arg('organization_id')::int
      AND (sqlc.arg('store_id')::int = 0 OR s.store_id = sqlc.arg('store_id'))
      AND (sqlc.arg('storage_location_id')::int = 0 OR s.storage_location_id = sqlc.arg('storage_location_id'))
), stock_totals AS (
    SELECT product_id, SUM(COALESCE(quantity_available, 0)) AS available_stock
    FROM scoped_stock
    GROUP BY product_id
), matching_products AS (
    SELECT p.id AS product_id, COALESCE(t.available_stock, 0)::numeric AS available_stock
    FROM products p
    LEFT JOIN stock_totals t ON t.product_id = p.id
    WHERE p.organization_id = sqlc.arg('organization_id')
      AND p.track_inventory = true
      AND (sqlc.arg('category_id')::int = 0 OR p.category_id = sqlc.arg('category_id'))
      AND ((sqlc.arg('store_id') = 0 AND sqlc.arg('storage_location_id') = 0) OR t.product_id IS NOT NULL)
      AND (
          sqlc.arg('stock_status')::text = 'all'
          OR (sqlc.arg('stock_status') = 'low' AND COALESCE(t.available_stock, 0) < sqlc.arg('low_stock_threshold')::numeric)
          OR (sqlc.arg('stock_status') = 'non_low' AND COALESCE(t.available_stock, 0) >= sqlc.arg('low_stock_threshold')::numeric)
      )
)
SELECT COUNT(*) FROM matching_products;
