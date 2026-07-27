-- name: CreateProductBarcode :one
INSERT INTO product_barcodes (
    product_id,
    product_variant_id,
    barcode,
    barcode_type,
    is_primary,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetProductBarcode :one
SELECT * FROM product_barcodes
WHERE id = $1;

-- name: GetProductByBarcode :one
SELECT pb.*, p.name AS product_name, p.sku
FROM product_barcodes pb
JOIN products p ON pb.product_id = p.id
WHERE pb.barcode = $1;

-- name: ListProductBarcodes :many
SELECT * FROM product_barcodes
ORDER BY barcode;

-- name: ListProductBarcodesByProduct :many
SELECT * FROM product_barcodes
WHERE product_id = $1
ORDER BY is_primary DESC, barcode;

-- name: ListProductBarcodesByVariant :many
SELECT * FROM product_barcodes
WHERE product_variant_id = $1
ORDER BY is_primary DESC, barcode;

-- name: GetPrimaryBarcode :one
SELECT * FROM product_barcodes
WHERE product_id = $1 AND is_primary = true
LIMIT 1;

-- name: UpdateProductBarcode :one
UPDATE product_barcodes
SET 
    barcode_type = $2,
    is_primary = $3,
    metadata = $4
WHERE id = $1
RETURNING *;

-- name: SetPrimaryBarcode :exec
UPDATE product_barcodes
SET is_primary = (id = $2)
WHERE product_id = $1;

-- name: DeleteProductBarcode :exec
DELETE FROM product_barcodes
WHERE id = $1;

-- name: CheckBarcodeExists :one
SELECT EXISTS(
    SELECT 1 FROM product_barcodes
    WHERE barcode = $1
) AS exists;
