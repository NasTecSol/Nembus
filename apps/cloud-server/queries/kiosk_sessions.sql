-- kiosk_sessions.sql

-- name: GetKioskSession :one
SELECT * FROM kiosk_sessions
WHERE id = $1 LIMIT 1;

-- name: GetKioskSessionByToken :one
SELECT * FROM kiosk_sessions
WHERE session_token = $1 LIMIT 1;

-- name: ListKioskSessions :many
SELECT * FROM kiosk_sessions
WHERE store_id = $1
ORDER BY opened_at DESC;

-- name: CreateKioskSession :one
INSERT INTO kiosk_sessions (
    pos_terminal_id, store_id, session_token, status, opened_at, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateKioskSession :one
UPDATE kiosk_sessions
SET
    status = $2,
    closed_at = $3,
    metadata = $4
WHERE id = $1
RETURNING *;

-- name: DeleteKioskSession :exec
DELETE FROM kiosk_sessions
WHERE id = $1;
