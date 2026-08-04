package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// AccountingUseCase provides the automated double-entry GL posting engine.
// It is injected into PosUseCase and GoodsReceiptNotesUseCase as a non-blocking
// side-effect after each transaction commit.
//
// Design principle: GL posting is NON-FATAL. If no GL mapping is configured,
// the function logs a warning and returns nil — never blocking a POS sale or GRN.
type AccountingUseCase struct {
	repo *repository.Queries
}

func NewAccountingUseCase() *AccountingUseCase {
	return &AccountingUseCase{}
}

func (uc *AccountingUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// resolveGLAccount resolves the CoA account for a given mapping type + store.
// Returns nil, nil if no mapping is configured (non-fatal).
func (uc *AccountingUseCase) resolveGLAccount(
	ctx context.Context,
	orgID int32,
	mappingType string,
	storeID *int32,
) (*repository.GetGLAccountMappingByTypeRow, error) {
	row, err := uc.repo.GetGLAccountMappingByType(ctx, repository.GetGLAccountMappingByTypeParams{
		OrganizationID: orgID,
		MappingType:    mappingType,
		StoreID:        utils.Int32ToPgInt4(storeID),
	})
	if err != nil {
		// No mapping configured — non-fatal
		return nil, nil
	}
	return &row, nil
}

// generateEntryNumber creates a unique journal entry number.
func generateEntryNumber(prefix string, refID string) string {
	return fmt.Sprintf("%s-%s-%d", prefix, refID, time.Now().UnixMicro()%9999999)
}

// PostSaleJournalEntry posts double-entry GL for a completed POS transaction.
//
// Routing logic:
//   - B2C customer (no business_partner_id)  → debit 'ar_retail'  (cash-clearing asset)
//   - B2B customer (business_partner_id set)  → debit 'ar_corporate' (open-account receivable)
//   - Always credit: 'revenue' + 'tax_payable'
//
// Called inside pos_usecase.CreateTransaction() as a non-blocking goroutine.
func (uc *AccountingUseCase) PostSaleJournalEntry(
	ctx context.Context,
	orgID int32,
	storeID int32,
	customerID *int32,
	txnNumber string,
	totalAmount pgtype.Numeric,
	taxAmount pgtype.Numeric,
) error {
	if uc.repo == nil {
		return nil // silently skip if accounting not wired
	}

	// Determine AR account type based on customer profile
	arMappingType := "ar_retail"
	if customerID != nil && *customerID > 0 {
		customer, err := uc.repo.GetCustomerWithPartner(ctx, *customerID)
		if err == nil && customer.BpID.Valid {
			// Customer is linked to a B2B business partner → corporate AR
			arMappingType = "ar_corporate"
		}
	}

	storeIDPtr := &storeID

	// Resolve all required GL accounts
	arAccount, err := uc.resolveGLAccount(ctx, orgID, arMappingType, storeIDPtr)
	if err != nil || arAccount == nil {
		fmt.Printf("[accounting] warn: no GL mapping for '%s' org=%d store=%d — skipping journal entry\n",
			arMappingType, orgID, storeID)
		return nil
	}
	revenueAccount, _ := uc.resolveGLAccount(ctx, orgID, "revenue", storeIDPtr)
	taxAccount, _ := uc.resolveGLAccount(ctx, orgID, "tax_payable", storeIDPtr)

	if revenueAccount == nil {
		fmt.Printf("[accounting] warn: no GL mapping for 'revenue' org=%d store=%d — skipping journal entry\n",
			orgID, storeID)
		return nil
	}

	// Calculate net revenue (total - tax)
	totalF, _ := totalAmount.Float64Value()
	taxF, _ := taxAmount.Float64Value()
	netRevenue := totalF.Float64 - taxF.Float64
	if netRevenue < 0 {
		netRevenue = 0
	}

	// Create journal entry header
	entry, err := uc.repo.CreateJournalEntry(ctx, repository.CreateJournalEntryParams{
		OrganizationID: orgID,
		EntryNumber:    generateEntryNumber("POS", txnNumber),
		PostingDate:    pgtype.Date{Time: time.Now(), Valid: true},
		ReferenceType:  "pos_transaction",
		ReferenceID:    txnNumber,
		Memo:           pgtype.Text{String: fmt.Sprintf("POS sale: %s", txnNumber), Valid: true},
	})
	if err != nil {
		fmt.Printf("[accounting] error creating journal entry for txn %s: %v\n", txnNumber, err)
		return nil // non-fatal
	}

	// DEBIT: Accounts Receivable (ar_retail or ar_corporate) = total amount
	if err := uc.createLine(ctx, entry.ID, arAccount.GlAccountID, 0, totalF.Float64, 0,
		fmt.Sprintf("AR: %s", txnNumber)); err != nil {
		fmt.Printf("[accounting] warn: failed to create AR debit line: %v\n", err)
	}

	// CREDIT: Revenue = net amount (ex-tax)
	if err := uc.createLine(ctx, entry.ID, revenueAccount.GlAccountID, 0, 0, netRevenue,
		fmt.Sprintf("Revenue: %s", txnNumber)); err != nil {
		fmt.Printf("[accounting] warn: failed to create revenue credit line: %v\n", err)
	}

	// CREDIT: Tax Payable = tax amount (if configured and > 0)
	if taxAccount != nil && taxF.Float64 > 0 {
		if err := uc.createLine(ctx, entry.ID, taxAccount.GlAccountID, 0, 0, taxF.Float64,
			fmt.Sprintf("Tax payable: %s", txnNumber)); err != nil {
			fmt.Printf("[accounting] warn: failed to create tax payable credit line: %v\n", err)
		}
	}

	return nil
}

// PostReceivingJournalEntry posts double-entry GL for a completed GRN.
//
// Entry pattern:
//   DEBIT:  inventory_asset (goods received into stock)
//   CREDIT: ap             (accounts payable — amount owed to supplier/vendor)
//
// Called inside GoodsReceiptNotesUseCase.CreateGoodsReceiptNote() as a non-blocking goroutine.
func (uc *AccountingUseCase) PostReceivingJournalEntry(
	ctx context.Context,
	orgID int32,
	storeID int32,
	grnNumber string,
	totalValue pgtype.Numeric,
) error {
	if uc.repo == nil {
		return nil
	}

	storeIDPtr := &storeID
	inventoryAccount, _ := uc.resolveGLAccount(ctx, orgID, "inventory_asset", storeIDPtr)
	apAccount, _ := uc.resolveGLAccount(ctx, orgID, "ap", storeIDPtr)

	if inventoryAccount == nil || apAccount == nil {
		fmt.Printf("[accounting] warn: GL mappings for 'inventory_asset'/'ap' not configured org=%d store=%d — skipping GRN journal\n",
			orgID, storeID)
		return nil
	}

	totalF, _ := totalValue.Float64Value()
	amount := totalF.Float64

	entry, err := uc.repo.CreateJournalEntry(ctx, repository.CreateJournalEntryParams{
		OrganizationID: orgID,
		EntryNumber:    generateEntryNumber("GRN", grnNumber),
		PostingDate:    pgtype.Date{Time: time.Now(), Valid: true},
		ReferenceType:  "grn",
		ReferenceID:    grnNumber,
		Memo:           pgtype.Text{String: fmt.Sprintf("Goods receipt: %s", grnNumber), Valid: true},
	})
	if err != nil {
		fmt.Printf("[accounting] error creating journal entry for GRN %s: %v\n", grnNumber, err)
		return nil
	}

	// DEBIT: Inventory Asset = total received value
	if err := uc.createLine(ctx, entry.ID, inventoryAccount.GlAccountID, 0, amount, 0,
		fmt.Sprintf("Inventory received: %s", grnNumber)); err != nil {
		fmt.Printf("[accounting] warn: failed to create inventory debit line: %v\n", err)
	}

	// CREDIT: Accounts Payable = total received value
	if err := uc.createLine(ctx, entry.ID, apAccount.GlAccountID, 0, 0, amount,
		fmt.Sprintf("AP: %s", grnNumber)); err != nil {
		fmt.Printf("[accounting] warn: failed to create AP credit line: %v\n", err)
	}

	return nil
}

// createLine is a helper that inserts a single journal line.
func (uc *AccountingUseCase) createLine(
	ctx context.Context,
	journalID int64,
	accountID int32,
	costCenterID int32,
	debit, credit float64,
	memo string,
) error {
	var debitNum, creditNum pgtype.Numeric
	_ = debitNum.Scan(fmt.Sprintf("%.2f", debit))
	_ = creditNum.Scan(fmt.Sprintf("%.2f", credit))

	var ccID pgtype.Int4
	if costCenterID > 0 {
		ccID = pgtype.Int4{Int32: costCenterID, Valid: true}
	}

	_, err := uc.repo.CreateJournalLine(ctx, repository.CreateJournalLineParams{
		JournalID:    journalID,
		AccountID:    accountID,
		CostCenterID: ccID,
		Debit:        debitNum,
		Credit:       creditNum,
		Memo:         pgtype.Text{String: memo, Valid: true},
	})
	return err
}
