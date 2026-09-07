-- =====================================================
-- ACCOUNTING ENGINE QUERIES
-- File: queries/accounting.sql
-- =====================================================

-- =====================================================
-- FISCAL YEARS
-- =====================================================

-- name: CreateFiscalYear :one
INSERT INTO fiscal_years (
    organization_id, year, name, start_date, end_date, status, metadata
) VALUES (
    $1, $2, $3, $4, $5, 'open', $6
)
RETURNING *;

-- name: GetFiscalYear :one
SELECT * FROM fiscal_years
WHERE id = $1 AND organization_id = $2;

-- name: GetFiscalYearByYear :one
SELECT * FROM fiscal_years
WHERE organization_id = $1 AND year = $2;

-- name: ListFiscalYears :many
SELECT * FROM fiscal_years
WHERE organization_id = $1
ORDER BY year DESC;

-- name: CloseFiscalYear :one
UPDATE fiscal_years
SET status = 'closed', closed_by = $3, closed_at = NOW()
WHERE id = $1 AND organization_id = $2 AND status = 'open'
RETURNING *;

-- =====================================================
-- POSTING PERIODS
-- =====================================================

-- name: CreatePostingPeriod :one
INSERT INTO posting_periods (
    organization_id, fiscal_year_id, period_no, period_name,
    start_date, end_date, status, allow_retro_posting
) VALUES ($1, $2, $3, $4, $5, $6, 'open', $7)
RETURNING *;

-- name: GetPostingPeriodByID :one
SELECT pp.*, fy.year AS fiscal_year
FROM posting_periods pp
JOIN fiscal_years fy ON fy.id = pp.fiscal_year_id
WHERE pp.id = $1 AND pp.organization_id = $2;

-- name: GetOpenPostingPeriodForDate :one
-- Returns the open posting period that covers the given date.
-- Used by PostJournalEntry to resolve the period.
SELECT pp.*
FROM posting_periods pp
JOIN fiscal_years fy ON fy.id = pp.fiscal_year_id
WHERE pp.organization_id = $1
  AND $2 BETWEEN pp.start_date AND pp.end_date
  AND pp.status IN ('open')
  AND fy.status = 'open'
ORDER BY pp.period_no ASC
LIMIT 1;

-- name: ListPostingPeriods :many
SELECT pp.*, fy.year AS fiscal_year
FROM posting_periods pp
JOIN fiscal_years fy ON fy.id = pp.fiscal_year_id
WHERE pp.organization_id = $1
  AND ($2::INTEGER IS NULL OR pp.fiscal_year_id = $2)
ORDER BY fy.year DESC, pp.period_no ASC;

-- name: ClosePostingPeriod :one
UPDATE posting_periods
SET status = 'closed', closed_by = $3, closed_at = NOW()
WHERE id = $1 AND organization_id = $2 AND status = 'open'
RETURNING *;

-- name: LockPostingPeriod :one
UPDATE posting_periods
SET status = 'locked'
WHERE id = $1 AND organization_id = $2 AND status = 'closed'
RETURNING *;

-- name: GetPreviousPostingPeriod :one
-- Returns the period immediately before the given period_no in the same fiscal year.
SELECT * FROM posting_periods
WHERE organization_id = $1
  AND fiscal_year_id = $2
  AND period_no = $3 - 1
LIMIT 1;

-- =====================================================
-- DOCUMENT SERIES
-- =====================================================

-- name: CreateDocumentSeries :one
INSERT INTO document_series (
    organization_id, doc_type, series_code, name, prefix, suffix,
    start_no, next_no, increment, zero_pad_length, is_default,
    per_fiscal_year, fiscal_year_id, is_active, metadata
) VALUES ($1, $2, $3, $4, $5, $6, $7, $7, $8, $9, $10, $11, $12, true, $13)
RETURNING *;

-- name: GetDefaultSeriesForDocType :one
SELECT * FROM document_series
WHERE organization_id = $1 AND doc_type = $2 AND is_default = true AND is_active = true
LIMIT 1;

-- name: IncrementSeriesNextNo :one
-- Atomically increments next_no and returns the OLD value (the number to use).
-- Must run inside a transaction with SELECT FOR UPDATE.
UPDATE document_series
SET next_no    = next_no + increment,
    updated_at = NOW()
WHERE id = $1 AND organization_id = $2
RETURNING (next_no - increment) AS allocated_no, prefix, suffix, zero_pad_length;

-- name: ListDocumentSeries :many
SELECT * FROM document_series
WHERE organization_id = $1
  AND ($2::VARCHAR IS NULL OR doc_type = $2)
ORDER BY doc_type, is_default DESC, series_code;

-- name: UpdateDocumentSeries :one
UPDATE document_series
SET name = $3, prefix = $4, suffix = $5, increment = $6,
    zero_pad_length = $7, is_default = $8, is_active = $9, updated_at = NOW()
WHERE id = $1 AND organization_id = $2
RETURNING *;

-- =====================================================
-- PROFIT CENTERS
-- =====================================================

-- name: CreateProfitCenter :one
INSERT INTO profit_centers (organization_id, code, name, parent_id, is_active)
VALUES ($1, $2, $3, $4, true)
RETURNING *;

-- name: GetProfitCenter :one
SELECT * FROM profit_centers WHERE id = $1 AND organization_id = $2;

-- name: ListProfitCenters :many
SELECT * FROM profit_centers
WHERE organization_id = $1 AND is_active = true
ORDER BY code;

-- =====================================================
-- JOURNAL ENTRIES
-- =====================================================

-- name: CreateJournalEntry :one
INSERT INTO journal_entries (
    organization_id, entry_number, posting_date,
    reference_type, reference_id,
    fiscal_year_id, posting_period_id,
    series_id, document_number, status,
    source_type, source_id,
    base_currency_code, total_debit_base, total_credit_base,
    posted_by, posted_at, memo
) VALUES (
    $1, $2, $3,
    $4, $5,
    $6, $7,
    $8, $9, $10,
    $11, $12,
    $13, $14, $15,
    $16, NOW(), $17
)
RETURNING *;

-- name: GetJournalEntry :one
SELECT je.*,
       pp.period_name,
       fy.year AS fiscal_year_no
FROM journal_entries je
LEFT JOIN posting_periods pp ON pp.id = je.posting_period_id
LEFT JOIN fiscal_years    fy ON fy.id = je.fiscal_year_id
WHERE je.id = $1 AND je.organization_id = $2;

-- name: GetJournalEntryBySourceIDempotency :one
-- Used by the posting engine to detect duplicate/replayed postings.
SELECT id, status
FROM journal_entries
WHERE source_type = $1 AND source_id = $2 AND status <> 'void'
LIMIT 1;

-- name: ListJournalEntries :many
SELECT je.*,
       pp.period_name
FROM journal_entries je
LEFT JOIN posting_periods pp ON pp.id = je.posting_period_id
WHERE je.organization_id = $1
  AND ($2::INTEGER IS NULL OR je.posting_period_id = $2)
  AND ($3::DATE IS NULL OR je.posting_date >= $3)
  AND ($4::DATE IS NULL OR je.posting_date <= $4)
  AND ($5::VARCHAR IS NULL OR je.status = $5)
  AND ($6::VARCHAR IS NULL OR je.source_type = $6)
ORDER BY je.posting_date DESC, je.id DESC
LIMIT $7 OFFSET $8;

-- name: VoidJournalEntry :one
UPDATE journal_entries
SET status = 'void', updated_at = NOW()
WHERE id = $1 AND organization_id = $2 AND status = 'posted'
RETURNING *;

-- name: SetJournalEntryReversed :one
UPDATE journal_entries
SET status = 'reversed', updated_at = NOW()
WHERE id = $1 AND organization_id = $2
RETURNING *;

-- =====================================================
-- JOURNAL LINES
-- =====================================================

-- name: CreateJournalLine :one
INSERT INTO journal_lines (
    journal_id, account_id, cost_center_id, profit_center_id, store_id, partner_id,
    line_no, debit, credit,
    currency_code, exchange_rate, amount_doc_currency, amount_base, debit_base, credit_base,
    reference_type, reference_id, reference_line, memo
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9,
    $10, $11, $12, $13, $14, $15,
    $16, $17, $18, $19
)
RETURNING *;

-- name: GetJournalLines :many
SELECT jl.*,
       coa.account_code,
       coa.account_name,
       coa.account_type,
       cc.name  AS cost_center_name,
       pc.name  AS profit_center_name,
       bp.name  AS partner_name,
       s.name   AS store_name
FROM journal_lines jl
JOIN chart_of_accounts coa ON coa.id = jl.account_id
LEFT JOIN cost_centers  cc  ON cc.id  = jl.cost_center_id
LEFT JOIN profit_centers pc  ON pc.id  = jl.profit_center_id
LEFT JOIN business_partners bp ON bp.id = jl.partner_id
LEFT JOIN stores            s  ON s.id  = jl.store_id
WHERE jl.journal_id = $1
ORDER BY jl.line_no;

-- =====================================================
-- GL POSTING RULES
-- =====================================================

-- name: UpsertGLPostingRule :one
INSERT INTO gl_posting_rules (
    organization_id, posting_type, debit_account_id, credit_account_id,
    tax_account_id, cost_center_id, profit_center_id, store_id, payment_method,
    description, is_active
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, true)
ON CONFLICT ON CONSTRAINT gl_posting_rules_pkey
DO UPDATE SET
    debit_account_id  = EXCLUDED.debit_account_id,
    credit_account_id = EXCLUDED.credit_account_id,
    tax_account_id    = EXCLUDED.tax_account_id,
    cost_center_id    = EXCLUDED.cost_center_id,
    profit_center_id  = EXCLUDED.profit_center_id,
    description       = EXCLUDED.description,
    is_active         = EXCLUDED.is_active,
    updated_at        = NOW()
RETURNING *;

-- name: ResolveGLPostingRule :one
-- Returns the most specific rule for a posting_type.
-- Prefers store+payment_method match over generic rule.
SELECT * FROM gl_posting_rules
WHERE organization_id = $1
  AND posting_type = $2
  AND is_active = true
  AND (store_id IS NULL OR store_id = $3)
  AND (payment_method IS NULL OR payment_method = $4)
ORDER BY
    (store_id IS NOT NULL)::INT       DESC,
    (payment_method IS NOT NULL)::INT DESC
LIMIT 1;

-- name: ListGLPostingRules :many
SELECT gpr.*,
       da.account_code AS debit_code,  da.account_name AS debit_name,
       ca.account_code AS credit_code, ca.account_name AS credit_name
FROM gl_posting_rules gpr
LEFT JOIN chart_of_accounts da ON da.id = gpr.debit_account_id
LEFT JOIN chart_of_accounts ca ON ca.id = gpr.credit_account_id
WHERE gpr.organization_id = $1 AND gpr.is_active = true
ORDER BY gpr.posting_type, gpr.store_id NULLS LAST;

-- =====================================================
-- ACCOUNT BALANCES
-- =====================================================

-- name: UpsertAccountBalance :one
INSERT INTO account_balances (
    organization_id, account_id, posting_period_id, fiscal_year_id,
    opening_balance, period_debits, period_credits, closing_balance,
    currency_code, last_updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
ON CONFLICT (organization_id, account_id, posting_period_id)
DO UPDATE SET
    period_debits    = EXCLUDED.period_debits,
    period_credits   = EXCLUDED.period_credits,
    closing_balance  = EXCLUDED.closing_balance,
    last_updated_at  = NOW()
RETURNING *;

-- name: IncrementAccountBalance :one
-- Atomically adds debit/credit amounts to an existing balance row.
-- Called by PostJournalEntry for each journal line.
INSERT INTO account_balances (
    organization_id, account_id, posting_period_id, fiscal_year_id,
    opening_balance, period_debits, period_credits, closing_balance,
    currency_code, last_updated_at
) VALUES ($1, $2, $3, $4, 0, $5, $6, (0 + $5 - $6), $7, NOW())
ON CONFLICT (organization_id, account_id, posting_period_id)
DO UPDATE SET
    period_debits   = account_balances.period_debits  + EXCLUDED.period_debits,
    period_credits  = account_balances.period_credits + EXCLUDED.period_credits,
    closing_balance = account_balances.opening_balance
                      + account_balances.period_debits  + EXCLUDED.period_debits
                      - (account_balances.period_credits + EXCLUDED.period_credits),
    last_updated_at = NOW()
RETURNING *;

-- name: GetAccountBalance :one
SELECT ab.*,
       coa.account_code, coa.account_name, coa.account_type, coa.account_group
FROM account_balances ab
JOIN chart_of_accounts coa ON coa.id = ab.account_id
WHERE ab.organization_id = $1
  AND ab.account_id = $2
  AND ab.posting_period_id = $3;

-- name: GetTrialBalance :many
-- Returns all accounts with non-zero activity in the given period.
SELECT
    coa.id              AS account_id,
    coa.account_code,
    coa.account_name,
    coa.account_type,
    coa.account_group,
    COALESCE(ab.opening_balance, 0) AS opening_balance,
    COALESCE(ab.period_debits,   0) AS period_debits,
    COALESCE(ab.period_credits,  0) AS period_credits,
    COALESCE(ab.closing_balance, 0) AS closing_balance
FROM chart_of_accounts coa
LEFT JOIN account_balances ab
       ON ab.account_id = coa.id
      AND ab.posting_period_id = $2
      AND ab.organization_id   = $1
WHERE coa.organization_id = $1
  AND coa.is_active = true
  AND coa.allow_direct_posting = true
  AND (
      ab.period_debits  <> 0 OR
      ab.period_credits <> 0 OR
      ab.opening_balance <> 0
  )
ORDER BY coa.account_code;

-- name: GetAccountLedger :many
-- Per-account transaction history between two dates.
SELECT
    je.id               AS journal_id,
    je.posting_date,
    je.document_number,
    je.source_type,
    je.source_id,
    je.memo             AS entry_memo,
    jl.line_no,
    jl.debit,
    jl.credit,
    jl.debit_base,
    jl.credit_base,
    jl.currency_code,
    jl.amount_doc_currency,
    jl.memo             AS line_memo,
    jl.reference_type,
    jl.reference_id,
    bp.name             AS partner_name,
    s.name              AS store_name
FROM journal_lines jl
JOIN journal_entries je ON je.id = jl.journal_id
LEFT JOIN business_partners bp ON bp.id = jl.partner_id
LEFT JOIN stores            s  ON s.id  = jl.store_id
WHERE jl.account_id       = $1
  AND je.organization_id  = $2
  AND je.posting_date    >= $3
  AND je.posting_date    <= $4
  AND je.status          <> 'void'
ORDER BY je.posting_date, je.id, jl.line_no
LIMIT $5 OFFSET $6;

-- name: GetProfitAndLoss :many
-- Revenue and expense account balances across multiple periods in a date range.
SELECT
    coa.id              AS account_id,
    coa.account_code,
    coa.account_name,
    coa.account_type,
    coa.account_group,
    SUM(COALESCE(ab.period_debits,  0)) AS total_debits,
    SUM(COALESCE(ab.period_credits, 0)) AS total_credits,
    -- For revenue: credits > debits = net revenue; for expense: debits > credits = net expense
    SUM(COALESCE(ab.period_credits, 0) - COALESCE(ab.period_debits, 0)) AS net_amount
FROM chart_of_accounts coa
JOIN account_balances ab
       ON ab.account_id      = coa.id
      AND ab.organization_id = $1
JOIN posting_periods pp ON pp.id = ab.posting_period_id
WHERE coa.organization_id = $1
  AND coa.account_type IN ('revenue','expense')
  AND pp.start_date >= $2
  AND pp.end_date   <= $3
GROUP BY coa.id, coa.account_code, coa.account_name, coa.account_type, coa.account_group
ORDER BY coa.account_type, coa.account_code;

-- name: GetBalanceSheet :many
-- Asset, liability, equity accounts as of the end of the given period.
SELECT
    coa.id              AS account_id,
    coa.account_code,
    coa.account_name,
    coa.account_type,
    coa.account_group,
    COALESCE(ab.closing_balance, 0) AS closing_balance
FROM chart_of_accounts coa
LEFT JOIN account_balances ab
       ON ab.account_id      = coa.id
      AND ab.posting_period_id = $2
      AND ab.organization_id = $1
WHERE coa.organization_id = $1
  AND coa.account_type IN ('asset','liability','equity')
  AND coa.is_active = true
ORDER BY coa.account_type, coa.account_code;

-- =====================================================
-- WITHHOLDING TAX ENTRIES
-- =====================================================

-- name: CreateWHTEntry :one
INSERT INTO withholding_tax_entries (
    organization_id, business_partner_id,
    reference_type, reference_id,
    gross_amount, wht_rate, wht_amount, net_amount,
    posting_date, gl_account_id, journal_entry_id, notes
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: ListWHTEntries :many
SELECT wt.*,
       bp.name AS partner_name,
       coa.account_code AS gl_account_code
FROM withholding_tax_entries wt
LEFT JOIN business_partners  bp  ON bp.id  = wt.business_partner_id
LEFT JOIN chart_of_accounts  coa ON coa.id = wt.gl_account_id
WHERE wt.organization_id = $1
  AND ($2::DATE IS NULL OR wt.posting_date >= $2)
  AND ($3::DATE IS NULL OR wt.posting_date <= $3)
  AND ($4::BOOLEAN IS NULL OR wt.is_remitted = $4)
ORDER BY wt.posting_date DESC
LIMIT $5 OFFSET $6;

-- name: MarkWHTEntryRemitted :one
UPDATE withholding_tax_entries
SET is_remitted = true, remitted_at = NOW()
WHERE id = $1 AND organization_id = $2 AND is_remitted = false
RETURNING *;

-- End of accounting.sql
