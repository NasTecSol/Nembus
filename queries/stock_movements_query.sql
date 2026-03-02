-- =====================================================
-- STOCK MOVEMENTS
-- =====================================================

-- name: CreateStockMovement :one
INSERT INTO stock_movements (
    movement_type,
    reference_type,
    reference_id,
    product_id,
    product_variant_id,
    from_store_id,
    to_store_id,
    from_location_id,
    to_location_id,
    quantity,
    uom_id,
    batch_number,
    serial_number,
    movement_date,
    posted_by,
    status,
    cost_per_unit,
    total_value,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19
) RETURNING stock_movements.*;

-- name: GetStockMovement :one
SELECT stock_movements.* 
FROM stock_movements
WHERE stock_movements.id = $1;

-- name: ListStockMovements :many
SELECT stock_movements.*
FROM stock_movements
ORDER BY stock_movements.movement_date DESC
LIMIT $1 OFFSET $2;

-- name: ListStockMovementsByProduct :many
SELECT stock_movements.*
FROM stock_movements
WHERE stock_movements.product_id = $1
ORDER BY stock_movements.movement_date DESC
LIMIT $2 OFFSET $3;

-- name: ListStockMovementsByStore :many
SELECT stock_movements.*
FROM stock_movements
WHERE stock_movements.from_store_id = $1 OR stock_movements.to_store_id = $1
ORDER BY stock_movements.movement_date DESC
LIMIT $2 OFFSET $3;

-- name: ListStockMovementsByType :many
SELECT stock_movements.*
FROM stock_movements
WHERE stock_movements.movement_type = $1
ORDER BY stock_movements.movement_date DESC
LIMIT $2 OFFSET $3;

-- name: ListStockMovementsByReference :many
SELECT stock_movements.*
FROM stock_movements
WHERE stock_movements.reference_type = $1 AND stock_movements.reference_id = $2
ORDER BY stock_movements.movement_date DESC;

-- name: ListStockMovementsByDateRange :many
SELECT stock_movements.*
FROM stock_movements
WHERE stock_movements.movement_date >= $1 AND stock_movements.movement_date <= $2
ORDER BY stock_movements.movement_date DESC;

-- name: GetStockMovementsByProductAndStore :many
SELECT stock_movements.*
FROM stock_movements
WHERE stock_movements.product_id = $1 
  AND (stock_movements.from_store_id = $2 OR stock_movements.to_store_id = $2)
  AND stock_movements.movement_date >= $3
ORDER BY stock_movements.movement_date DESC;

-- name: UpdateStockMovementStatus :one
UPDATE stock_movements
SET status = $2
WHERE stock_movements.id = $1
RETURNING stock_movements.*;

-- name: GetStockMovementSummaryByProduct :one
SELECT 
    stock_movements.product_id,
    SUM(CASE WHEN stock_movements.to_store_id = $2 THEN stock_movements.quantity ELSE 0 END) AS total_in,
    SUM(CASE WHEN stock_movements.from_store_id = $2 THEN stock_movements.quantity ELSE 0 END) AS total_out,
    SUM(CASE WHEN stock_movements.to_store_id = $2 THEN stock_movements.total_value ELSE 0 END) AS value_in,
    SUM(CASE WHEN stock_movements.from_store_id = $2 THEN stock_movements.total_value ELSE 0 END) AS value_out
FROM stock_movements
WHERE stock_movements.product_id = $1
  AND stock_movements.movement_date >= $3 AND stock_movements.movement_date <= $4
GROUP BY stock_movements.product_id;

-- =====================================================
-- STOCK MOVEMENTS FROM ORDERS
-- =====================================================

-- name: CreateStockMovementFromSalesOrder :one
INSERT INTO stock_movements (
    movement_type,
    reference_type,
    reference_id,
    product_id,
    product_variant_id,
    from_store_id,
    from_location_id,
    quantity,
    uom_id,
    batch_number,
    serial_number,
    movement_date,
    posted_by,
    status,
    cost_per_unit,
    total_value,
    metadata
) VALUES (
    'sale',
    'sales_order',
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    COALESCE($10, CURRENT_TIMESTAMP),
    $11,
    COALESCE($12, 'completed'),
    $13,
    $14,
    COALESCE($15, '{}'::jsonb)
) RETURNING stock_movements.*;

-- name: CreateStockMovementFromPurchaseOrder :one
INSERT INTO stock_movements (
    movement_type,
    reference_type,
    reference_id,
    product_id,
    product_variant_id,
    to_store_id,
    to_location_id,
    quantity,
    uom_id,
    batch_number,
    serial_number,
    movement_date,
    posted_by,
    status,
    cost_per_unit,
    total_value,
    metadata
) VALUES (
    'receipt',
    'purchase_order',
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    COALESCE($10, CURRENT_TIMESTAMP),
    $11,
    COALESCE($12, 'completed'),
    $13,
    $14,
    COALESCE($15, '{}'::jsonb)
) RETURNING stock_movements.*;

-- name: CreateStockMovementsFromOrderFulfillment :many
INSERT INTO stock_movements (
    movement_type,
    reference_type,
    reference_id,
    product_id,
    product_variant_id,
    from_store_id,
    from_location_id,
    quantity,
    uom_id,
    batch_number,
    serial_number,
    movement_date,
    posted_by,
    status,
    cost_per_unit,
    total_value,
    metadata
)
SELECT 
    'sale' AS movement_type,
    'sales_order' AS reference_type,
    sol.sales_order_id AS reference_id,
    sol.product_id,
    sol.product_variant_id,
    so.store_id AS from_store_id,
    NULL::integer AS from_location_id,
    sol.quantity_fulfilled AS quantity,
    sol.uom_id,
    NULL::varchar AS batch_number,
    NULL::varchar AS serial_number,
    CURRENT_TIMESTAMP AS movement_date,
    so.created_by_user_id AS posted_by,
    'completed' AS status,
    sol.unit_cost AS cost_per_unit,
    (sol.unit_cost * sol.quantity_fulfilled) AS total_value,
    jsonb_build_object('order_line_id', sol.id, 'order_number', so.order_number) AS metadata
FROM sales_order_lines_v2 sol
JOIN sales_orders_v2 so ON sol.sales_order_id = so.id
WHERE sol.sales_order_id = $1
  AND sol.quantity_fulfilled > 0
RETURNING stock_movements.*;

-- name: GetStockMovementsBySalesOrder :many
SELECT stock_movements.*
FROM stock_movements
WHERE stock_movements.reference_type = 'sales_order'
  AND stock_movements.reference_id = $1
ORDER BY stock_movements.movement_date DESC;

-- name: GetStockMovementsByPurchaseOrder :many
SELECT stock_movements.*
FROM stock_movements
WHERE stock_movements.reference_type = 'purchase_order'
  AND stock_movements.reference_id = $1
ORDER BY stock_movements.movement_date DESC;

-- =====================================================
-- INVENTORY STOCK UPDATES FROM STOCK MOVEMENTS
-- =====================================================

-- name: UpdateInventoryStockFromMovement :one
WITH movement_data AS (
    SELECT 
        CASE 
            WHEN sm.to_store_id IS NOT NULL THEN sm.to_store_id
            WHEN sm.from_store_id IS NOT NULL THEN sm.from_store_id
        END AS store_id,
        CASE 
            WHEN sm.to_store_id IS NOT NULL THEN sm.to_location_id
            WHEN sm.from_store_id IS NOT NULL THEN sm.from_location_id
        END AS location_id,
        sm.product_id,
        sm.product_variant_id,
        CASE 
            WHEN sm.to_store_id IS NOT NULL THEN sm.quantity
            WHEN sm.from_store_id IS NOT NULL THEN -sm.quantity
        END AS quantity_delta
    FROM stock_movements sm
    WHERE sm.id = $1
)
UPDATE inventory_stock
SET 
    quantity_on_hand = inventory_stock.quantity_on_hand + md.quantity_delta,
    quantity_available = GREATEST(0, inventory_stock.quantity_available + md.quantity_delta),
    updated_at = CURRENT_TIMESTAMP
FROM movement_data md
WHERE inventory_stock.product_id = md.product_id
  AND inventory_stock.product_variant_id IS NOT DISTINCT FROM md.product_variant_id
  AND inventory_stock.store_id = md.store_id
  AND inventory_stock.storage_location_id IS NOT DISTINCT FROM md.location_id
RETURNING inventory_stock.*;

-- name: UpsertInventoryStockFromMovement :one
WITH movement_data AS (
    SELECT 
        CASE 
            WHEN sm.to_store_id IS NOT NULL THEN sm.to_store_id
            WHEN sm.from_store_id IS NOT NULL THEN sm.from_store_id
        END AS store_id,
        CASE 
            WHEN sm.to_store_id IS NOT NULL THEN sm.to_location_id
            WHEN sm.from_store_id IS NOT NULL THEN sm.from_location_id
        END AS location_id,
        sm.product_id,
        sm.product_variant_id,
        CASE 
            WHEN sm.to_store_id IS NOT NULL THEN sm.quantity
            WHEN sm.from_store_id IS NOT NULL THEN -sm.quantity
        END AS quantity_delta
    FROM stock_movements sm
    WHERE sm.id = $1
),
existing_stock AS (
    SELECT inventory_stock.id 
    FROM inventory_stock
    WHERE inventory_stock.product_id = (SELECT product_id FROM movement_data)
      AND inventory_stock.product_variant_id IS NOT DISTINCT FROM (SELECT product_variant_id FROM movement_data)
      AND inventory_stock.store_id = (SELECT store_id FROM movement_data)
      AND inventory_stock.storage_location_id IS NOT DISTINCT FROM (SELECT location_id FROM movement_data)
    LIMIT 1
),
updated AS (
    UPDATE inventory_stock
    SET 
        quantity_on_hand = inventory_stock.quantity_on_hand + md.quantity_delta,
        quantity_available = GREATEST(0, inventory_stock.quantity_available + md.quantity_delta),
        updated_at = CURRENT_TIMESTAMP
    FROM movement_data md, existing_stock es
    WHERE inventory_stock.id = es.id
      AND inventory_stock.product_id = md.product_id
      AND inventory_stock.product_variant_id IS NOT DISTINCT FROM md.product_variant_id
      AND inventory_stock.store_id = md.store_id
      AND inventory_stock.storage_location_id IS NOT DISTINCT FROM md.location_id
    RETURNING inventory_stock.*
)
INSERT INTO inventory_stock (
    product_id,
    product_variant_id,
    store_id,
    storage_location_id,
    quantity_on_hand,
    quantity_available
)
SELECT 
    md.product_id,
    md.product_variant_id,
    md.store_id,
    md.location_id,
    GREATEST(0, md.quantity_delta) AS quantity_on_hand,
    GREATEST(0, md.quantity_delta) AS quantity_available
FROM movement_data md
WHERE NOT EXISTS (SELECT 1 FROM existing_stock)
RETURNING inventory_stock.*;
