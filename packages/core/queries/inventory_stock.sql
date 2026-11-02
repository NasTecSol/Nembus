-- name: GetLowStockProducts :many
SELECT 
    s.product_id,
    p.sku,
    p.name,
    s.store_id,
    st.name AS store_name,
    s.quantity_available,
    s.reorder_level,
    s.reorder_quantity,
    s.max_stock_level
FROM inventory_stock s
JOIN products p ON s.product_id = p.id
JOIN stores st ON s.store_id = st.id
WHERE s.quantity_available <= s.reorder_level
  AND s.quantity_available > 0
  AND p.is_active = true
  AND p.track_inventory = true
ORDER BY s.quantity_available ASC, p.name
LIMIT 50;

-- name: GetStockValuationByStore :many
SELECT 
    s.store_id,
    st.name AS store_name,
    COUNT(DISTINCT s.product_id) AS unique_products,
    SUM(s.quantity_on_hand * COALESCE((
        SELECT AVG(pp.price)
        FROM product_prices pp
        WHERE pp.product_id = s.product_id
          AND pp.is_active = true
    ), 0)) AS total_stock_value
FROM inventory_stock s
JOIN stores st ON s.store_id = st.id
WHERE st.organization_id = sqlc.arg('org_id')
GROUP BY s.store_id, st.name
ORDER BY total_stock_value DESC;

-- =====================================================
-- INVENTORY STOCK CRUD OPERATIONS
-- =====================================================

-- name: CreateInventoryStock :one
INSERT INTO inventory_stock (
    product_id,
    product_variant_id,
    store_id,
    storage_location_id,
    quantity_on_hand,
    quantity_allocated,
    quantity_available,
    quantity_on_order,
    quantity_in_transit,
    reorder_level,
    reorder_quantity,
    max_stock_level,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetInventoryStock :one
SELECT * FROM inventory_stock
WHERE id = $1;

-- name: GetInventoryStockByProductAndStore :one
SELECT * FROM inventory_stock
WHERE product_id = $1
  AND product_variant_id IS NOT DISTINCT FROM $2
  AND store_id = $3
  AND storage_location_id IS NOT DISTINCT FROM $4
LIMIT 1;

-- name: ListInventoryStock :many
SELECT * FROM inventory_stock
ORDER BY created_at DESC;

-- name: ListInventoryStockByStore :many
SELECT * FROM inventory_stock
WHERE store_id = $1
ORDER BY product_id;

-- name: ListInventoryStockByProduct :many
SELECT * FROM inventory_stock
WHERE product_id = $1
ORDER BY store_id;

-- name: ListInventoryStockByStorageLocation :many
SELECT * FROM inventory_stock
WHERE storage_location_id = $1
ORDER BY product_id;

-- name: ListInventoryStockByStoreAndLocation :many
SELECT * FROM inventory_stock
WHERE store_id = $1
  AND storage_location_id IS NOT DISTINCT FROM $2
ORDER BY product_id;

-- name: UpdateInventoryStock :one
UPDATE inventory_stock
SET 
    quantity_on_hand = COALESCE($2, quantity_on_hand),
    quantity_allocated = COALESCE($3, quantity_allocated),
    quantity_available = COALESCE($4, quantity_available),
    quantity_on_order = COALESCE($5, quantity_on_order),
    quantity_in_transit = COALESCE($6, quantity_in_transit),
    reorder_level = COALESCE($7, reorder_level),
    reorder_quantity = COALESCE($8, reorder_quantity),
    max_stock_level = COALESCE($9, max_stock_level),
    last_counted_at = COALESCE($10, last_counted_at),
    metadata = COALESCE($11, metadata)
WHERE id = $1
RETURNING *;

-- Note: UpsertInventoryStock should be handled in usecase layer by checking GetInventoryStockByProductAndStore first

-- name: AdjustInventoryStock :one
UPDATE inventory_stock
SET 
    quantity_on_hand = quantity_on_hand + COALESCE($2, 0),
    quantity_available = quantity_available + COALESCE($3, 0),
    quantity_allocated = quantity_allocated + COALESCE($4, 0),
    quantity_on_order = quantity_on_order + COALESCE($5, 0),
    quantity_in_transit = quantity_in_transit + COALESCE($6, 0)
WHERE id = $1
RETURNING *;

-- name: AdjustInventoryStockByProductAndStore :one
UPDATE inventory_stock
SET 
    quantity_on_hand = quantity_on_hand + COALESCE($4, 0),
    quantity_available = quantity_available + COALESCE($5, 0),
    quantity_allocated = quantity_allocated + COALESCE($6, 0),
    quantity_on_order = quantity_on_order + COALESCE($7, 0),
    quantity_in_transit = quantity_in_transit + COALESCE($8, 0)
WHERE product_id = $1
  AND product_variant_id IS NOT DISTINCT FROM $2
  AND store_id = $3
RETURNING *;

-- name: DeleteInventoryStock :exec
DELETE FROM inventory_stock
WHERE id = $1;

-- name: GetInventoryStockSummary :one
SELECT 
    COUNT(*) AS total_records,
    COUNT(DISTINCT product_id) AS unique_products,
    COUNT(DISTINCT store_id) AS unique_stores,
    SUM(quantity_on_hand) AS total_on_hand,
    SUM(quantity_available) AS total_available,
    SUM(quantity_allocated) AS total_allocated
FROM inventory_stock
WHERE store_id = $1;

-- name: GetInventoryTotals :one
WITH scope AS (
    SELECT sqlc.arg('organization_id')::int AS organization_id,
           sqlc.arg('store_id')::int AS store_id,
           sqlc.arg('storage_location_id')::int AS storage_location_id,
           sqlc.arg('price_list_id')::int AS price_list_id
), scoped_stores AS (
    SELECT st.id, st.name
    FROM stores st, scope sc
    WHERE st.organization_id = sc.organization_id
      AND (sc.store_id = 0 OR st.id = sc.store_id)
), stock AS (
    SELECT s.store_id, s.product_id, s.product_variant_id,
           SUM(COALESCE(s.quantity_on_hand, 0)) AS units,
           SUM(COALESCE(s.quantity_available, 0)) AS available_units
    FROM inventory_stock s
    JOIN scoped_stores st ON st.id = s.store_id
    JOIN products p ON p.id = s.product_id
    CROSS JOIN scope sc
    WHERE p.organization_id = sc.organization_id AND p.track_inventory = true
      AND (sc.storage_location_id = 0 OR s.storage_location_id = sc.storage_location_id)
    GROUP BY s.store_id, s.product_id, s.product_variant_id
), valued_stock AS (
    SELECT s.*, NULLIF(to_jsonb(p)->>'cost_price', '')::numeric AS unit_cost,
           sale.price AS unit_sale_price
    FROM stock s
    JOIN products p ON p.id = s.product_id
    CROSS JOIN scope sc
    LEFT JOIN LATERAL (
        SELECT pp.price
        FROM product_prices pp
        WHERE pp.product_id = p.id AND pp.price_list_id = sc.price_list_id
          AND (pp.product_variant_id = s.product_variant_id OR pp.product_variant_id IS NULL)
          AND (pp.uom_id = p.base_uom_id OR pp.uom_id IS NULL)
          AND pp.is_active = true
          AND (pp.valid_from IS NULL OR pp.valid_from <= CURRENT_DATE)
          AND (pp.valid_to IS NULL OR pp.valid_to >= CURRENT_DATE)
          AND COALESCE(pp.min_quantity, 1) <= 1
          AND (pp.max_quantity IS NULL OR pp.max_quantity >= 1)
        ORDER BY (pp.product_variant_id IS NOT NULL) DESC,
                 (pp.uom_id IS NOT NULL) DESC, pp.min_quantity DESC NULLS LAST,
                 pp.valid_from DESC NULLS LAST, pp.updated_at DESC NULLS LAST, pp.id DESC
        LIMIT 1
    ) sale ON true
), store_totals AS (
    SELECT st.id AS store_id, st.name AS store_name,
           COUNT(DISTINCT v.product_id) AS product_count,
           COALESCE(SUM(v.units), 0) AS total_units,
           COALESCE(SUM(v.available_units), 0) AS total_available_units,
           COALESCE(SUM(v.units * v.unit_cost), 0) AS total_cost_value,
           COALESCE(SUM(v.units * v.unit_sale_price), 0) AS total_sale_value,
           COALESCE(SUM(ABS(v.units)) FILTER (WHERE v.unit_cost IS NULL), 0) AS units_missing_cost_price,
           COALESCE(SUM(ABS(v.units)) FILTER (WHERE v.unit_sale_price IS NULL), 0) AS units_missing_sale_price
    FROM scoped_stores st
    LEFT JOIN valued_stock v ON v.store_id = st.id
    GROUP BY st.id, st.name
)
SELECT
    EXISTS(SELECT 1 FROM organizations o WHERE o.id = sc.organization_id)::boolean AS organization_exists,
    (sc.store_id = 0 OR EXISTS(SELECT 1 FROM scoped_stores))::boolean AS store_exists,
    (sc.storage_location_id = 0 OR EXISTS(
        SELECT 1 FROM storage_locations sl
        JOIN scoped_stores st ON st.id = sl.store_id
        WHERE sl.id = sc.storage_location_id AND sl.store_id = sc.store_id
    ))::boolean AS location_exists,
    EXISTS(
        SELECT 1 FROM price_lists pl JOIN organizations o ON o.id = sc.organization_id
        WHERE pl.id = sc.price_list_id AND pl.is_active = true
          AND pl.currency_code = o.currency_code
          AND (pl.valid_from IS NULL OR pl.valid_from <= CURRENT_DATE)
          AND (pl.valid_to IS NULL OR pl.valid_to >= CURRENT_DATE)
    )::boolean AS price_list_valid,
    jsonb_build_object(
        'organization_id', sc.organization_id,
        'store_id', NULLIF(sc.store_id, 0),
        'storage_location_id', NULLIF(sc.storage_location_id, 0),
        'price_list_id', sc.price_list_id,
        'currency_code', (SELECT o.currency_code FROM organizations o WHERE o.id = sc.organization_id),
        'quantity_basis', 'quantity_on_hand',
        'product_count', (SELECT COUNT(DISTINCT product_id) FROM stock),
        'store_count', (SELECT COUNT(*) FROM scoped_stores),
        'total_units', (SELECT COALESCE(SUM(total_units), 0) FROM store_totals),
        'total_available_units', (SELECT COALESCE(SUM(total_available_units), 0) FROM store_totals),
        'total_cost_value', (SELECT COALESCE(SUM(total_cost_value), 0) FROM store_totals),
        'total_sale_value', (SELECT COALESCE(SUM(total_sale_value), 0) FROM store_totals),
        'units_missing_cost_price', (SELECT COALESCE(SUM(units_missing_cost_price), 0) FROM store_totals),
        'units_missing_sale_price', (SELECT COALESCE(SUM(units_missing_sale_price), 0) FROM store_totals),
        'stores', (SELECT COALESCE(jsonb_agg(to_jsonb(t) ORDER BY t.store_id), '[]'::jsonb) FROM store_totals t)
    )::jsonb AS data
FROM scope sc;
