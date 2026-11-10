-- name: CreatePOSTerminal :one
INSERT INTO pos_terminals (
    store_id,
    terminal_code,
    terminal_name,
    device_id,
    is_active,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetPOSTerminal :one
SELECT * FROM pos_terminals
WHERE id = $1;

-- name: GetPOSTerminalByCode :one
SELECT * FROM pos_terminals
WHERE store_id = $1 AND terminal_code = $2;

-- name: ListPOSTerminals :many
SELECT * FROM pos_terminals
ORDER BY terminal_code;

-- name: ListPOSTerminalsByStore :many
SELECT * FROM pos_terminals
WHERE store_id = $1
ORDER BY terminal_code;

-- name: ListActivePOSTerminalsByStore :many
SELECT * FROM pos_terminals
WHERE store_id = $1 AND is_active = true
ORDER BY terminal_code;

-- name: UpdatePOSTerminal :one
UPDATE pos_terminals
SET 
    terminal_name = $2,
    device_id = $3,
    is_active = $4,
    metadata = $5
WHERE id = $1
RETURNING *;

-- name: DeletePOSTerminal :exec
DELETE FROM pos_terminals
WHERE id = $1;

-- name: TogglePOSTerminalActive :one
UPDATE pos_terminals
SET is_active = $2
WHERE id = $1
RETURNING *;
