-- name: CreateProductSerialNumber :one
INSERT INTO product_serial_numbers (
    product_id,
    product_variant_id,
    serial_number,
    status,
    current_store_id,
    manufacturing_date,
    expiry_date,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetProductSerialNumber :one
SELECT * FROM product_serial_numbers
WHERE id = $1;

-- name: GetProductSerialNumberBySerial :one
SELECT * FROM product_serial_numbers
WHERE serial_number = $1;

-- name: ListProductSerialNumbers :many
SELECT * FROM product_serial_numbers
ORDER BY created_at DESC;

-- name: ListProductSerialNumbersByProduct :many
SELECT * FROM product_serial_numbers
WHERE product_id = $1
ORDER BY serial_number;

-- name: ListProductSerialNumbersByStore :many
SELECT * FROM product_serial_numbers
WHERE current_store_id = $1
ORDER BY serial_number;

-- name: ListProductSerialNumbersByStatus :many
SELECT * FROM product_serial_numbers
WHERE status = $1
ORDER BY serial_number;

-- name: ListAvailableSerialNumbers :many
SELECT * FROM product_serial_numbers
WHERE product_id = $1 
  AND status = 'in_stock'
  AND current_store_id = $2
ORDER BY serial_number;

-- name: UpdateProductSerialNumber :one
UPDATE product_serial_numbers
SET 
    status = $2,
    current_store_id = $3,
    metadata = $4
WHERE id = $1
RETURNING *;

-- name: UpdateSerialNumberStatus :one
UPDATE product_serial_numbers
SET status = $2
WHERE serial_number = $1
RETURNING *;

-- name: TransferSerialNumber :one
UPDATE product_serial_numbers
SET 
    current_store_id = $2,
    status = $3
WHERE serial_number = $1
RETURNING *;

-- name: DeleteProductSerialNumber :exec
DELETE FROM product_serial_numbers
WHERE id = $1;

-- name: GetExpiredSerialNumbers :many
SELECT * FROM product_serial_numbers
WHERE expiry_date < CURRENT_DATE
  AND status = 'in_stock'
ORDER BY expiry_date;

-- name: GetExpiringSerialNumbers :many
SELECT * FROM product_serial_numbers
WHERE expiry_date BETWEEN CURRENT_DATE AND $1
  AND status = 'in_stock'
ORDER BY expiry_date;
