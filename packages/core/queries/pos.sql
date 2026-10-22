-- pos.sql

-- name: UpdatePosTransactionMetadata :exec
UPDATE pos_transactions SET metadata = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: GetPosTransactionLineByID :one
SELECT * FROM pos_transaction_lines WHERE id = $1;
