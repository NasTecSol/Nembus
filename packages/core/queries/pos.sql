-- pos.sql

-- name: UpdatePosTransactionMetadata :exec
UPDATE pos_transactions SET metadata = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: PatchPosTransactionMetadata :exec
UPDATE pos_transactions
SET metadata = COALESCE(metadata, '{}'::jsonb) || $2::jsonb,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: GetPosTransactionLineByID :one
SELECT * FROM pos_transaction_lines WHERE id = $1;

-- name: LockPosTransactionLineForUpdate :one
SELECT * FROM pos_transaction_lines WHERE id = $1 FOR UPDATE;
