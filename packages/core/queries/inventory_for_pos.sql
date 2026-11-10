-- name: GetAvailableStockForPos :one
SELECT 
    quantity_available,
    quantity_on_hand,
    quantity_allocated
FROM inventory_stock
WHERE product_id = $1
  AND product_variant_id IS NOT DISTINCT FROM $2
  AND store_id = $3
LIMIT 1;

-- name: GetProductsWithStockForQuickSearch :many
SELECT 
    p.id,
    p.sku,
    p.name,
    COALESCE(s.quantity_available, 0) AS available_qty,
    pb.barcode
FROM products p
LEFT JOIN inventory_stock s 
    ON s.product_id = p.id 
   AND s.store_id = $1
LEFT JOIN product_barcodes pb 
    ON pb.product_id = p.id 
   AND pb.is_primary = true
WHERE p.is_active = true
  AND p.is_sellable = true
  AND p.track_inventory = true
  AND (p.name ILIKE '%' || $2 || '%' OR p.sku ILIKE '%' || $2 || '%' OR pb.barcode = $2)
ORDER BY available_qty DESC, p.name
LIMIT 50;