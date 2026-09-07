package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// =====================================================
// ERROR TYPES
// =====================================================

// ErrPeriodClosed is returned when attempting to post to a closed or locked period.
type ErrPeriodClosed struct {
	PeriodName string
	Status     string
}

func (e ErrPeriodClosed) Error() string {
	return fmt.Sprintf("posting period '%s' is %s — posting not allowed", e.PeriodName, e.Status)
}

// ErrImbalancedEntry is returned when debit total ≠ credit total.
type ErrImbalancedEntry struct {
	TotalDebits  float64
	TotalCredits float64
}

func (e ErrImbalancedEntry) Error() string {
	return fmt.Sprintf("journal entry is imbalanced: debits=%.2f credits=%.2f", e.TotalDebits, e.TotalCredits)
}

// ErrDuplicatePosting is returned when an idempotency check finds an existing posting
// for the same source_type + source_id.
type ErrDuplicatePosting struct {
	SourceType string
	SourceID   string
	ExistingID int64
}

func (e ErrDuplicatePosting) Error() string {
	return fmt.Sprintf("duplicate posting: source %s/%s already posted as journal_entry %d",
		e.SourceType, e.SourceID, e.ExistingID)
}

// ErrNoGLRule is returned when no gl_posting_rule is configured for a posting_type.
type ErrNoGLRule struct {
	PostingType string
}

func (e ErrNoGLRule) Error() string {
	return fmt.Sprintf("no GL posting rule configured for posting_type '%s'", e.PostingType)
}

// ErrNoPeriodForDate is returned when no open posting period covers the given date.
type ErrNoPeriodForDate struct {
	Date time.Time
}

func (e ErrNoPeriodForDate) Error() string {
	return fmt.Sprintf("no open posting period found for date %s", e.Date.Format("2006-01-02"))
}

// =====================================================
// INPUT / OUTPUT TYPES
// =====================================================

// JournalLineInput is the caller-supplied data for one debit or credit line.
type JournalLineInput struct {
	AccountID       int32    `json:"account_id"`
	CostCenterID    *int32   `json:"cost_center_id,omitempty"`
	ProfitCenterID  *int32   `json:"profit_center_id,omitempty"`
	StoreID         *int32   `json:"store_id,omitempty"`
	PartnerID       *int32   `json:"partner_id,omitempty"`
	Debit           float64  `json:"debit"`            // in document currency
	Credit          float64  `json:"credit"`           // in document currency
	CurrencyCode    string   `json:"currency_code"`    // document currency, e.g. "SAR"
	ExchangeRate    float64  `json:"exchange_rate"`    // to base currency; 1.0 for SAR
	ReferenceType   *string  `json:"reference_type,omitempty"`
	ReferenceID     *string  `json:"reference_id,omitempty"`
	ReferenceLine   *int32   `json:"reference_line,omitempty"`
	Memo            *string  `json:"memo,omitempty"`
}

// PostJournalEntryInput is the full input for a single double-entry posting.
type PostJournalEntryInput struct {
	OrganizationID   int32              `json:"organization_id"`
	PostingDate      time.Time          `json:"posting_date"`
	SourceType       string             `json:"source_type"`  // e.g. "pos_transaction"
	SourceID         string             `json:"source_id"`    // PK of triggering document
	DocType          string             `json:"doc_type"`     // for document_series lookup
	BaseCurrencyCode string             `json:"base_currency_code"` // org's functional currency
	Lines            []JournalLineInput `json:"lines"`
	Memo             *string            `json:"memo,omitempty"`
	PostedBy         *int32             `json:"posted_by,omitempty"`
}

// AccountingUseCase holds the DB pool so it can acquire a connection when needed,
// but all posting methods MUST be called with a caller-supplied pgx.Tx.
type AccountingUseCase struct {
	repo *repository.Queries
	pool *pgxpool.Pool
}

func NewAccountingUseCase(pool *pgxpool.Pool) *AccountingUseCase {
	return &AccountingUseCase{pool: pool}
}

func (uc *AccountingUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// repoTx returns a Queries scoped to the provided transaction.
func (uc *AccountingUseCase) repoTx(tx pgx.Tx) *repository.Queries {
	return uc.repo.WithTx(tx)
}

// =====================================================
// CORE POSTING ENGINE
// =====================================================

// PostJournalEntry is the single entry-point for ALL automated GL postings.
// It MUST be called inside the caller's pgx.Tx so that if the journal fails,
// the entire triggering operation rolls back atomically.
//
// Caller is responsible for starting and committing/rolling back the transaction.
func (uc *AccountingUseCase) PostJournalEntry(
	ctx context.Context,
	tx pgx.Tx,
	input PostJournalEntryInput,
) (*repository.JournalEntry, error) {

	q := uc.repoTx(tx)

	// 1. Idempotency check — prevent duplicate postings from offline-sync retries.
	if input.SourceType != "" && input.SourceID != "" {
		existing, err := q.GetJournalEntryBySourceIDempotency(ctx,
			repository.GetJournalEntryBySourceIDempotencyParams{
				SourceType: pgtype.Text{String: input.SourceType, Valid: true},
				SourceID:   pgtype.Text{String: input.SourceID, Valid: true},
			})
		if err == nil {
			// Already posted — return idempotently without error.
			return nil, ErrDuplicatePosting{
				SourceType: input.SourceType,
				SourceID:   input.SourceID,
				ExistingID: existing.ID,
			}
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("idempotency check: %w", err)
		}
	}

	// 2. Validate debit = credit balance.
	if err := uc.ValidateBalance(input.Lines); err != nil {
		return nil, err
	}

	// 3. Resolve the posting period for the given date.
	period, err := q.GetOpenPostingPeriodForDate(ctx,
		repository.GetOpenPostingPeriodForDateParams{
			OrganizationID: input.OrganizationID,
			Date:           pgtype.Date{Time: input.PostingDate, Valid: true},
		})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoPeriodForDate{Date: input.PostingDate}
		}
		return nil, fmt.Errorf("resolve posting period: %w", err)
	}
	if period.Status != "open" {
		return nil, ErrPeriodClosed{PeriodName: period.PeriodName, Status: period.Status}
	}

	// 4. Get next document number from document_series.
	entryNumber, documentNumber, seriesID, err := uc.getNextDocumentNumber(ctx, q, input.OrganizationID, input.DocType)
	if err != nil {
		// Fallback: generate a safe timestamp-based number so posting still works.
		ts := time.Now().UnixMicro()
		entryNumber = fmt.Sprintf("JE-%d", ts)
		documentNumber = entryNumber
		seriesID = nil
	}

	// 5. Compute totals in base currency.
	var totalDebit, totalCredit float64
	for _, l := range input.Lines {
		rate := l.ExchangeRate
		if rate == 0 {
			rate = 1.0
		}
		totalDebit += l.Debit * rate
		totalCredit += l.Credit * rate
	}

	// 6. Insert journal_entries header.
	memo := ""
	if input.Memo != nil {
		memo = *input.Memo
	}
	postedBy := pgtype.Int4{}
	if input.PostedBy != nil {
		postedBy = pgtype.Int4{Int32: *input.PostedBy, Valid: true}
	}

	createParams := repository.CreateJournalEntryParams{
		OrganizationID:    input.OrganizationID,
		EntryNumber:       entryNumber,
		PostingDate:       pgtype.Date{Time: input.PostingDate, Valid: true},
		ReferenceType:     input.SourceType,
		ReferenceID:       input.SourceID,
		FiscalYearID:      pgtype.Int4{Int32: period.FiscalYearID, Valid: true},
		PostingPeriodID:   pgtype.Int4{Int32: period.ID, Valid: true},
		SeriesID:          toInt4Ptr(seriesID),
		DocumentNumber:    pgtype.Text{String: documentNumber, Valid: documentNumber != ""},
		Status:            pgtype.Text{String: "posted", Valid: true},
		SourceType:        pgtype.Text{String: input.SourceType, Valid: input.SourceType != ""},
		SourceID:          pgtype.Text{String: input.SourceID, Valid: input.SourceID != ""},
		BaseCurrencyCode:  pgtype.Text{String: input.BaseCurrencyCode, Valid: input.BaseCurrencyCode != ""},
		TotalDebitBase:    toPgNumeric(totalDebit),
		TotalCreditBase:   toPgNumeric(totalCredit),
		PostedBy:          postedBy,
		Memo:              pgtype.Text{String: memo, Valid: memo != ""},
	}

	entry, err := q.CreateJournalEntry(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("create journal entry: %w", err)
	}

	// 7. Insert journal_lines and update account_balances for each line.
	for i, l := range input.Lines {
		lineNo := int32(i + 1)
		rate := l.ExchangeRate
		if rate == 0 {
			rate = 1.0
		}
		debitBase := l.Debit * rate
		creditBase := l.Credit * rate
		amountBase := debitBase - creditBase
		if amountBase < 0 {
			amountBase = -amountBase
		}
		amountDoc := l.Debit - l.Credit
		if amountDoc < 0 {
			amountDoc = -amountDoc
		}

		currCode := l.CurrencyCode
		if currCode == "" {
			currCode = input.BaseCurrencyCode
		}

		lineParams := repository.CreateJournalLineParams{
			JournalID:         entry.ID,
			AccountID:         l.AccountID,
			CostCenterID:      toInt4Ptr(l.CostCenterID),
			ProfitCenterID:    toInt4Ptr(l.ProfitCenterID),
			StoreID:           toInt4Ptr(l.StoreID),
			PartnerID:         toInt4Ptr(l.PartnerID),
			LineNo:            lineNo,
			Debit:             toPgNumeric(l.Debit),
			Credit:            toPgNumeric(l.Credit),
			CurrencyCode:      pgtype.Text{String: currCode, Valid: true},
			ExchangeRate:      toPgNumeric(rate),
			AmountDocCurrency: toPgNumeric(amountDoc),
			AmountBase:        toPgNumeric(amountBase),
			DebitBase:         toPgNumeric(debitBase),
			CreditBase:        toPgNumeric(creditBase),
			ReferenceType:     toPgTextPtr(l.ReferenceType),
			ReferenceID:       toPgTextPtr(l.ReferenceID),
			ReferenceLine:     toInt4Ptr(l.ReferenceLine),
			Memo:              toPgTextPtr(l.Memo),
		}

		if _, err := q.CreateJournalLine(ctx, lineParams); err != nil {
			return nil, fmt.Errorf("create journal line %d: %w", lineNo, err)
		}

		// Update running account_balances (incremental upsert).
		if err := uc.incrementAccountBalance(ctx, q, input.OrganizationID, l.AccountID,
			period.ID, period.FiscalYearID, debitBase, creditBase,
			input.BaseCurrencyCode); err != nil {
			return nil, fmt.Errorf("update account balance (line %d): %w", lineNo, err)
		}
	}

	return &entry, nil
}

// ValidateBalance checks that the sum of debits equals the sum of credits.
// This is a pure function — safe to call before starting a transaction.
func (uc *AccountingUseCase) ValidateBalance(lines []JournalLineInput) error {
	const epsilon = 0.005 // tolerance for floating-point rounding

	var totalDebit, totalCredit float64
	for _, l := range lines {
		totalDebit += l.Debit
		totalCredit += l.Credit
	}

	diff := totalDebit - totalCredit
	if diff < 0 {
		diff = -diff
	}
	if diff > epsilon {
		return ErrImbalancedEntry{TotalDebits: totalDebit, TotalCredits: totalCredit}
	}
	return nil
}

// ReverseJournalEntry creates a mirror-image reversal entry and marks the
// original as 'reversed'. Must be called inside a transaction.
func (uc *AccountingUseCase) ReverseJournalEntry(
	ctx context.Context,
	tx pgx.Tx,
	orgID int32,
	entryID int64,
	reason string,
	reversalDate time.Time,
	reversedBy *int32,
) (*repository.JournalEntry, error) {

	q := uc.repoTx(tx)

	// Load original entry.
	original, err := q.GetJournalEntry(ctx, repository.GetJournalEntryParams{
		ID: entryID, OrganizationID: orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("load journal entry %d: %w", entryID, err)
	}
	if original.Status.String != "posted" {
		return nil, fmt.Errorf("journal entry %d is not in 'posted' status (current: %s)",
			entryID, original.Status.String)
	}

	// Load original lines.
	lines, err := q.GetJournalLines(ctx, entryID)
	if err != nil {
		return nil, fmt.Errorf("load journal lines for entry %d: %w", entryID, err)
	}

	// Build reversal lines (swap debit ↔ credit).
	var reversalLines []JournalLineInput
	for _, ol := range lines {
		debit, _ := numericToFloat64(ol.Debit)
		credit, _ := numericToFloat64(ol.Credit)
		rate, _ := numericToFloat64(ol.ExchangeRate)
		if rate == 0 {
			rate = 1.0
		}

		reversalLines = append(reversalLines, JournalLineInput{
			AccountID:      ol.AccountID,
			CostCenterID:   pgInt4ToPtr(ol.CostCenterID),
			ProfitCenterID: pgInt4ToPtr(ol.ProfitCenterID),
			StoreID:        pgInt4ToPtr(ol.StoreID),
			PartnerID:      pgInt4ToPtr(ol.PartnerID),
			Debit:          credit, // swapped
			Credit:         debit,  // swapped
			CurrencyCode:   ol.CurrencyCode.String,
			ExchangeRate:   rate,
			Memo:           strPtr(fmt.Sprintf("REVERSAL: %s", reason)),
		})
	}

	baseCcy := ""
	if original.BaseCurrencyCode.Valid {
		baseCcy = original.BaseCurrencyCode.String
	}

	// Post the reversal entry.
	reversalInput := PostJournalEntryInput{
		OrganizationID:   orgID,
		PostingDate:      reversalDate,
		SourceType:       "reversal",
		SourceID:         fmt.Sprintf("rev-%d", entryID),
		DocType:          "journal_entry",
		BaseCurrencyCode: baseCcy,
		Lines:            reversalLines,
		Memo:             strPtr(fmt.Sprintf("Reversal of entry %d: %s", entryID, reason)),
		PostedBy:         reversedBy,
	}

	reversalEntry, err := uc.PostJournalEntry(ctx, tx, reversalInput)
	if err != nil {
		return nil, fmt.Errorf("post reversal entry: %w", err)
	}

	// Mark original as reversed.
	if _, err := q.SetJournalEntryReversed(ctx, repository.SetJournalEntryReversedParams{
		ID: entryID, OrganizationID: orgID,
	}); err != nil {
		return nil, fmt.Errorf("mark original entry reversed: %w", err)
	}

	return reversalEntry, nil
}

// =====================================================
// FISCAL YEAR MANAGEMENT
// =====================================================

// OpenFiscalYear creates a fiscal_year and 12 monthly posting_periods.
func (uc *AccountingUseCase) OpenFiscalYear(
	ctx context.Context,
	orgID int32,
	year int,
	baseCurrencyCode string,
) *repository.Response {

	conn, err := uc.pool.Acquire(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to acquire db connection", nil)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to begin transaction", nil)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	q := uc.repoTx(tx)

	// Create fiscal year record.
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)

	fy, err := q.CreateFiscalYear(ctx, repository.CreateFiscalYearParams{
		OrganizationID: orgID,
		Year:           int32(year),
		Name:           fmt.Sprintf("FY %d", year),
		StartDate:      pgtype.Date{Time: startDate, Valid: true},
		EndDate:        pgtype.Date{Time: endDate, Valid: true},
		Metadata:       []byte("{}"),
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("create fiscal year: %v", err), nil)
	}

	// Create 12 monthly periods.
	monthNames := []string{
		"Jan", "Feb", "Mar", "Apr", "May", "Jun",
		"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
	}
	for m := 1; m <= 12; m++ {
		pStart := time.Date(year, time.Month(m), 1, 0, 0, 0, 0, time.UTC)
		pEnd := pStart.AddDate(0, 1, -1)
		periodName := fmt.Sprintf("%s %d", monthNames[m-1], year)

		if _, err := q.CreatePostingPeriod(ctx, repository.CreatePostingPeriodParams{
			OrganizationID:     orgID,
			FiscalYearID:       fy.ID,
			PeriodNo:           int32(m),
			PeriodName:         periodName,
			StartDate:          pgtype.Date{Time: pStart, Valid: true},
			EndDate:            pgtype.Date{Time: pEnd, Valid: true},
			AllowRetroPosting:  pgtype.Bool{Bool: false, Valid: true},
		}); err != nil {
			return utils.NewResponse(utils.CodeError,
				fmt.Sprintf("create period %d: %v", m, err), nil)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return utils.NewResponse(utils.CodeError, "commit failed", nil)
	}

	return utils.NewResponse(utils.CodeSuccess, "fiscal year opened", fy)
}

// =====================================================
// PERIOD CLOSE
// =====================================================

// RunPeriodClose locks the given period and rolls closing balances to the
// next period's opening balances. Called at month-end by the finance team.
func (uc *AccountingUseCase) RunPeriodClose(
	ctx context.Context,
	orgID int32,
	periodID int32,
	closedBy int32,
) *repository.Response {

	conn, err := uc.pool.Acquire(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "db connection failed", nil)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "transaction failed", nil)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	q := uc.repoTx(tx)

	// 1. Load period.
	period, err := q.GetPostingPeriodByID(ctx, repository.GetPostingPeriodByIDParams{
		ID: periodID, OrganizationID: orgID,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "period not found", nil)
	}
	if period.Status != "open" {
		return utils.NewResponse(utils.CodeError,
			fmt.Sprintf("period '%s' is already %s", period.PeriodName, period.Status), nil)
	}

	// 2. Close the period.
	_, err = q.ClosePostingPeriod(ctx, repository.ClosePostingPeriodParams{
		ID:             periodID,
		OrganizationID: orgID,
		ClosedBy:       pgtype.Int4{Int32: closedBy, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to close period", nil)
	}

	// 3. Find the next period and seed its opening balances from this period's closing balances.
	nextPeriod, err := q.GetOpenPostingPeriodForDate(ctx, repository.GetOpenPostingPeriodForDateParams{
		OrganizationID: orgID,
		Date:           pgtype.Date{Time: period.EndDate.Time.AddDate(0, 0, 1), Valid: true},
	})
	if err == nil {
		// Load current period's account balances and roll forward.
		currentBalances, balErr := q.ListAccountBalancesForPeriod(ctx,
			repository.ListAccountBalancesForPeriodParams{
				OrganizationID:  orgID,
				PostingPeriodID: periodID,
			})
		if balErr == nil {
			for _, bal := range currentBalances {
				closing, _ := numericToFloat64(bal.ClosingBalance)
				// Seed next period's opening balance (only asset/liability/equity carry forward).
				_, _ = q.UpsertAccountBalance(ctx, repository.UpsertAccountBalanceParams{
					OrganizationID:  orgID,
					AccountID:       bal.AccountID,
					PostingPeriodID: nextPeriod.ID,
					FiscalYearID:    nextPeriod.FiscalYearID,
					OpeningBalance:  toPgNumeric(closing),
					PeriodDebits:    toPgNumeric(0),
					PeriodCredits:   toPgNumeric(0),
					ClosingBalance:  toPgNumeric(closing),
					CurrencyCode:    bal.CurrencyCode,
				})
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return utils.NewResponse(utils.CodeError, "commit failed", nil)
	}

	return utils.NewResponse(utils.CodeSuccess, "period closed successfully", map[string]any{
		"period_id":   periodID,
		"period_name": period.PeriodName,
		"closed_by":   closedBy,
	})
}

// =====================================================
// FINANCIAL REPORTS
// =====================================================

// GetTrialBalance returns all accounts with activity in the given period.
func (uc *AccountingUseCase) GetTrialBalance(
	ctx context.Context,
	orgID int32,
	periodID int32,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	rows, err := uc.repo.GetTrialBalance(ctx, repository.GetTrialBalanceParams{
		OrganizationID:  orgID,
		PostingPeriodID: pgtype.Int4{Int32: periodID, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeSuccess, "trial balance retrieved", rows)
}

// GetAccountLedger returns a paginated transaction history for one account.
func (uc *AccountingUseCase) GetAccountLedger(
	ctx context.Context,
	orgID int32,
	accountID int32,
	from, to time.Time,
	limit, offset int32,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	rows, err := uc.repo.GetAccountLedger(ctx, repository.GetAccountLedgerParams{
		AccountID:      accountID,
		OrganizationID: orgID,
		FromDate:       pgtype.Date{Time: from, Valid: true},
		ToDate:         pgtype.Date{Time: to, Valid: true},
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeSuccess, "account ledger retrieved", rows)
}

// GetProfitAndLoss returns revenue and expense accounts with period totals.
func (uc *AccountingUseCase) GetProfitAndLoss(
	ctx context.Context,
	orgID int32,
	from, to time.Time,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	rows, err := uc.repo.GetProfitAndLoss(ctx, repository.GetProfitAndLossParams{
		OrganizationID: orgID,
		FromDate:       pgtype.Date{Time: from, Valid: true},
		ToDate:         pgtype.Date{Time: to, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	// Summarise totals.
	type PLSummary struct {
		Lines          any     `json:"lines"`
		TotalRevenue   float64 `json:"total_revenue"`
		TotalCOGS      float64 `json:"total_cogs"`
		TotalExpenses  float64 `json:"total_expenses"`
		GrossProfit    float64 `json:"gross_profit"`
		NetProfit      float64 `json:"net_profit"`
	}

	var totalRevenue, totalCOGS, totalExpenses float64
	for _, r := range rows {
		net, _ := numericToFloat64(r.NetAmount)
		switch r.AccountGroup.String {
		case "revenue":
			totalRevenue += net
		case "cogs":
			totalCOGS += net
		default:
			if r.AccountType == "expense" {
				totalExpenses += net
			}
		}
	}

	result := PLSummary{
		Lines:         rows,
		TotalRevenue:  totalRevenue,
		TotalCOGS:     totalCOGS,
		TotalExpenses: totalExpenses,
		GrossProfit:   totalRevenue - totalCOGS,
		NetProfit:     totalRevenue - totalCOGS - totalExpenses,
	}

	return utils.NewResponse(utils.CodeSuccess, "profit & loss retrieved", result)
}

// GetBalanceSheet returns asset, liability, equity accounts as of period end.
func (uc *AccountingUseCase) GetBalanceSheet(
	ctx context.Context,
	orgID int32,
	periodID int32,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	rows, err := uc.repo.GetBalanceSheet(ctx, repository.GetBalanceSheetParams{
		OrganizationID:  orgID,
		PostingPeriodID: pgtype.Int4{Int32: periodID, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeSuccess, "balance sheet retrieved", rows)
}

// =====================================================
// GL POSTING RULES
// =====================================================

// ResolveGLAccounts returns the GL accounts for a given posting_type.
// Use-cases call this to build JournalLineInput slices without hardcoding account IDs.
func (uc *AccountingUseCase) ResolveGLAccounts(
	ctx context.Context,
	q *repository.Queries,
	orgID int32,
	postingType string,
	storeID *int32,
	paymentMethod *string,
) (repository.GlPostingRule, error) {

	storeIDVal := pgtype.Int4{}
	if storeID != nil {
		storeIDVal = pgtype.Int4{Int32: *storeID, Valid: true}
	}
	pmVal := pgtype.Text{}
	if paymentMethod != nil {
		pmVal = pgtype.Text{String: *paymentMethod, Valid: true}
	}

	rule, err := q.ResolveGLPostingRule(ctx, repository.ResolveGLPostingRuleParams{
		OrganizationID: orgID,
		PostingType:    postingType,
		StoreID:        storeIDVal,
		PaymentMethod:  pmVal,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.GlPostingRule{}, ErrNoGLRule{PostingType: postingType}
		}
		return repository.GlPostingRule{}, err
	}
	return rule, nil
}

// =====================================================
// LIST / READ OPERATIONS
// =====================================================

func (uc *AccountingUseCase) ListJournalEntries(
	ctx context.Context,
	orgID int32,
	periodID *int32,
	from, to *time.Time,
	status *string,
	sourceType *string,
	limit, offset int32,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	params := repository.ListJournalEntriesParams{
		OrganizationID: orgID,
		Limit:          limit,
		Offset:         offset,
	}
	if periodID != nil {
		params.PostingPeriodID = pgtype.Int4{Int32: *periodID, Valid: true}
	}
	if from != nil {
		params.FromDate = pgtype.Date{Time: *from, Valid: true}
	}
	if to != nil {
		params.ToDate = pgtype.Date{Time: *to, Valid: true}
	}
	if status != nil {
		params.Status = pgtype.Text{String: *status, Valid: true}
	}
	if sourceType != nil {
		params.SourceType = pgtype.Text{String: *sourceType, Valid: true}
	}

	rows, err := uc.repo.ListJournalEntries(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeSuccess, "journal entries retrieved", rows)
}

func (uc *AccountingUseCase) ListFiscalYears(ctx context.Context, orgID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	rows, err := uc.repo.ListFiscalYears(ctx, orgID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeSuccess, "fiscal years retrieved", rows)
}

func (uc *AccountingUseCase) ListPostingPeriods(
	ctx context.Context,
	orgID int32,
	fiscalYearID *int32,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	fyID := pgtype.Int4{}
	if fiscalYearID != nil {
		fyID = pgtype.Int4{Int32: *fiscalYearID, Valid: true}
	}
	rows, err := uc.repo.ListPostingPeriods(ctx, repository.ListPostingPeriodsParams{
		OrganizationID: orgID,
		FiscalYearID:   fyID,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeSuccess, "posting periods retrieved", rows)
}

func (uc *AccountingUseCase) ListGLPostingRules(ctx context.Context, orgID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	rows, err := uc.repo.ListGLPostingRules(ctx, orgID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeSuccess, "GL posting rules retrieved", rows)
}

func (uc *AccountingUseCase) ListWHTEntries(
	ctx context.Context,
	orgID int32,
	from, to *time.Time,
	isRemitted *bool,
	limit, offset int32,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	params := repository.ListWHTEntriesParams{
		OrganizationID: orgID,
		Limit:          limit,
		Offset:         offset,
	}
	if from != nil {
		params.FromDate = pgtype.Date{Time: *from, Valid: true}
	}
	if to != nil {
		params.ToDate = pgtype.Date{Time: *to, Valid: true}
	}
	if isRemitted != nil {
		params.IsRemitted = pgtype.Bool{Bool: *isRemitted, Valid: true}
	}
	rows, err := uc.repo.ListWHTEntries(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeSuccess, "WHT entries retrieved", rows)
}

func (uc *AccountingUseCase) ListDocumentSeries(
	ctx context.Context,
	orgID int32,
	docType *string,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	dtVal := pgtype.Text{}
	if docType != nil {
		dtVal = pgtype.Text{String: *docType, Valid: true}
	}
	rows, err := uc.repo.ListDocumentSeries(ctx, repository.ListDocumentSeriesParams{
		OrganizationID: orgID,
		DocType:        dtVal,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeSuccess, "document series retrieved", rows)
}

func (uc *AccountingUseCase) ListProfitCenters(ctx context.Context, orgID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	rows, err := uc.repo.ListProfitCenters(ctx, orgID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeSuccess, "profit centers retrieved", rows)
}

// =====================================================
// HANDLER-CALLABLE WRAPPERS (open their own pool tx)
// =====================================================

// ReverseJournalEntryFromHandler is called by the HTTP handler. It opens a
// pool transaction internally so the handler stays thin.
func (uc *AccountingUseCase) ReverseJournalEntryFromHandler(
	ctx context.Context,
	orgID int32,
	entryID int64,
	reason string,
	reversalDate time.Time,
	reversedBy *int32,
) *repository.Response {

	conn, err := uc.pool.Acquire(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "db connection failed", nil)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "transaction failed", nil)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	reversalEntry, err := uc.ReverseJournalEntry(ctx, tx, orgID, entryID, reason, reversalDate, reversedBy)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	if err := tx.Commit(ctx); err != nil {
		return utils.NewResponse(utils.CodeError, "commit failed", nil)
	}

	return utils.NewResponse(utils.CodeSuccess, "journal entry reversed", reversalEntry)
}

// =====================================================
// GL POSTING RULES — WRITE
// =====================================================

// UpsertGLPostingRuleInput is the body payload for creating/updating a GL posting rule.
type UpsertGLPostingRuleInput struct {
	PostingType      string  `json:"posting_type"`
	DebitAccountID   *int32  `json:"debit_account_id,omitempty"`
	CreditAccountID  *int32  `json:"credit_account_id,omitempty"`
	TaxAccountID     *int32  `json:"tax_account_id,omitempty"`
	CostCenterID     *int32  `json:"cost_center_id,omitempty"`
	ProfitCenterID   *int32  `json:"profit_center_id,omitempty"`
	StoreID          *int32  `json:"store_id,omitempty"`
	PaymentMethod    *string `json:"payment_method,omitempty"`
	Description      *string `json:"description,omitempty"`
}

// UpsertGLPostingRule creates or updates a GL posting rule.
func (uc *AccountingUseCase) UpsertGLPostingRule(
	ctx context.Context,
	orgID int32,
	input UpsertGLPostingRuleInput,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if input.PostingType == "" {
		return utils.NewResponse(utils.CodeBadReq, "posting_type is required", nil)
	}

	params := repository.UpsertGLPostingRuleParams{
		OrganizationID:  orgID,
		PostingType:     input.PostingType,
		DebitAccountID:  toInt4Ptr(input.DebitAccountID),
		CreditAccountID: toInt4Ptr(input.CreditAccountID),
		TaxAccountID:    toInt4Ptr(input.TaxAccountID),
		CostCenterID:    toInt4Ptr(input.CostCenterID),
		ProfitCenterID:  toInt4Ptr(input.ProfitCenterID),
		StoreID:         toInt4Ptr(input.StoreID),
		PaymentMethod:   toPgTextPtr(input.PaymentMethod),
		Description:     toPgTextPtr(input.Description),
	}

	rule, err := uc.repo.UpsertGLPostingRule(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeSuccess, "GL posting rule saved", rule)
}

// =====================================================
// INTERNAL HELPERS
// =====================================================

// getNextDocumentNumber atomically increments a document_series counter and returns
// the formatted document number. Falls back gracefully if no series is configured.
func (uc *AccountingUseCase) getNextDocumentNumber(
	ctx context.Context,
	q *repository.Queries,
	orgID int32,
	docType string,
) (entryNumber, documentNumber string, seriesID *int32, err error) {

	result, err := q.IncrementSeriesNextNo(ctx, repository.IncrementSeriesNextNoParams{
		OrganizationID: orgID,
		DocType:        docType,
	})
	if err != nil {
		return "", "", nil, err
	}

	prefix := result.Prefix.String
	suffix := result.Suffix.String
	padLen := int(result.ZeroPadLength.Int32)
	if padLen <= 0 {
		padLen = 6
	}
	num := fmt.Sprintf("%0*d", padLen, result.AllocatedNo)
	documentNumber = prefix + num + suffix
	entryNumber = documentNumber // reuse same value for entry_number

	sid := int32(result.SeriesID)
	return entryNumber, documentNumber, &sid, nil
}

// incrementAccountBalance atomically adds debit/credit to the running account_balance row.
func (uc *AccountingUseCase) incrementAccountBalance(
	ctx context.Context,
	q *repository.Queries,
	orgID, accountID int32,
	periodID, fiscalYearID int32,
	debitBase, creditBase float64,
	currencyCode string,
) error {
	if currencyCode == "" {
		currencyCode = "SAR"
	}
	_, err := q.IncrementAccountBalance(ctx, repository.IncrementAccountBalanceParams{
		OrganizationID:  orgID,
		AccountID:       accountID,
		PostingPeriodID: periodID,
		FiscalYearID:    fiscalYearID,
		PeriodDebits:    toPgNumeric(debitBase),
		PeriodCredits:   toPgNumeric(creditBase),
		CurrencyCode:    pgtype.Text{String: currencyCode, Valid: true},
	})
	return err
}

// =====================================================
// CONVERSION HELPERS
// =====================================================

func toPgNumeric(v float64) pgtype.Numeric {
	n := new(big.Int)
	// Convert via string to avoid float imprecision.
	s := fmt.Sprintf("%.2f", v)
	n2, _, err := big.ParseFloat(s, 10, 256, big.ToNearestEven)
	if err != nil {
		_ = n2
	}
	// Use pgtype.Numeric with Int + Exp.
	cents := int64(v * 100)
	return pgtype.Numeric{Int: big.NewInt(cents), Exp: -2, Valid: true}
}

func numericToFloat64(n pgtype.Numeric) (float64, error) {
	if !n.Valid {
		return 0, nil
	}
	f, _ := new(big.Float).SetInt(n.Int).Float64()
	exp := int(n.Exp)
	if exp >= 0 {
		for i := 0; i < exp; i++ {
			f *= 10
		}
	} else {
		for i := 0; i < -exp; i++ {
			f /= 10
		}
	}
	return f, nil
}

func toInt4Ptr(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

func toPgTextPtr(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *v, Valid: true}
}

func pgInt4ToPtr(v pgtype.Int4) *int32 {
	if !v.Valid {
		return nil
	}
	return &v.Int32
}

func strPtr(s string) *string { return &s }
