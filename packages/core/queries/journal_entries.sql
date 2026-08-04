-- =====================================================
-- JOURNAL ENTRIES & GL ACCOUNT MAPPINGS QUERIES
-- =====================================================

-- name: CreateJournalEntry :one
INSERT INTO journal_entries (
    organization_id, entry_number, posting_date, reference_type, reference_id, memo
) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: CreateJournalLine :one
INSERT INTO journal_lines (
    journal_id, account_id, cost_center_id, debit, credit, memo
) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetJournalEntry :one
SELECT * FROM journal_entries WHERE id = $1;

-- name: GetJournalEntryByNumber :one
SELECT * FROM journal_entries WHERE entry_number = $1;

-- name: ListJournalEntriesByReference :many
SELECT je.*, json_agg(
    jsonb_build_object(
        'id', jl.id,
        'account_id', jl.account_id,
        'account_code', coa.account_code,
        'account_name', coa.account_name,
        'cost_center_id', jl.cost_center_id,
        'debit', jl.debit,
        'credit', jl.credit,
        'memo', jl.memo
    ) ORDER BY jl.id
) AS lines
FROM journal_entries je
LEFT JOIN journal_lines jl ON jl.journal_id = je.id
LEFT JOIN chart_of_accounts coa ON jl.account_id = coa.id
WHERE je.organization_id = $1
  AND je.reference_type = $2
  AND je.reference_id = $3
GROUP BY je.id
ORDER BY je.posting_date DESC;

-- name: ListJournalEntriesByOrganization :many
SELECT je.* FROM journal_entries je
WHERE je.organization_id = $1
  AND (sqlc.narg(from_date)::date IS NULL OR je.posting_date >= sqlc.narg(from_date)::date)
  AND (sqlc.narg(to_date)::date IS NULL OR je.posting_date <= sqlc.narg(to_date)::date)
ORDER BY je.posting_date DESC, je.id DESC
LIMIT $2 OFFSET $3;

-- name: ListJournalLinesByEntry :many
SELECT jl.*, coa.account_code, coa.account_name, coa.account_type
FROM journal_lines jl
JOIN chart_of_accounts coa ON jl.account_id = coa.id
WHERE jl.journal_id = $1
ORDER BY jl.id;

-- name: GetGLAccountMappingByType :one
-- Resolve the GL account for a given transaction context.
-- Use store_id = NULL for org-level fallback mappings.
SELECT gm.*, coa.account_code, coa.account_name, coa.account_type
FROM gl_account_mappings gm
JOIN chart_of_accounts coa ON gm.gl_account_id = coa.id
WHERE gm.organization_id = $1
  AND gm.mapping_type = $2
  AND (
    -- Store-specific mapping takes precedence over org-level
    gm.store_id = sqlc.narg(store_id)::int
    OR (gm.store_id IS NULL AND NOT EXISTS (
        SELECT 1 FROM gl_account_mappings
        WHERE organization_id = $1 AND mapping_type = $2 AND store_id = sqlc.narg(store_id)::int
    ))
  )
ORDER BY gm.store_id NULLS LAST
LIMIT 1;

-- name: CreateGLAccountMapping :one
INSERT INTO gl_account_mappings (organization_id, mapping_type, store_id, gl_account_id)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: ListGLAccountMappings :many
SELECT gm.*, coa.account_code, coa.account_name
FROM gl_account_mappings gm
JOIN chart_of_accounts coa ON gm.gl_account_id = coa.id
WHERE gm.organization_id = $1
ORDER BY gm.mapping_type, gm.store_id NULLS LAST;

-- name: DeleteGLAccountMapping :exec
DELETE FROM gl_account_mappings WHERE id = $1;

-- name: SumJournalBalanceByAccount :one
-- AR aging / balance checks per account
SELECT
    coa.id AS account_id,
    coa.account_code,
    coa.account_name,
    COALESCE(SUM(jl.debit), 0)  AS total_debit,
    COALESCE(SUM(jl.credit), 0) AS total_credit,
    COALESCE(SUM(jl.debit), 0) - COALESCE(SUM(jl.credit), 0) AS net_balance
FROM chart_of_accounts coa
LEFT JOIN journal_lines jl ON jl.account_id = coa.id
LEFT JOIN journal_entries je ON jl.journal_id = je.id AND je.organization_id = $1
WHERE coa.id = $2
GROUP BY coa.id, coa.account_code, coa.account_name;
