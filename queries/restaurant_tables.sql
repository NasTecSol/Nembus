-- restaurant_tables.sql

-- name: GetRestaurantTable :one
SELECT * FROM restaurant_tables
WHERE id = $1 LIMIT 1;

-- name: ListRestaurantTables :many
SELECT * FROM restaurant_tables
WHERE store_id = $1
ORDER BY table_number;

-- name: CreateRestaurantTable :one
INSERT INTO restaurant_tables (
    store_id, table_number, table_name, section, capacity, is_active, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateRestaurantTable :one
UPDATE restaurant_tables
SET
    table_number = $2,
    table_name = $3,
    section = $4,
    capacity = $5,
    is_active = $6,
    metadata = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteRestaurantTable :exec
DELETE FROM restaurant_tables
WHERE id = $1;
