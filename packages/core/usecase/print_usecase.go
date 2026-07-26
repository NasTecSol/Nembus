package usecase

import (
	"context"
	"fmt"

	"github.com/NasTecSol/nembus-core/printing"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-client/utils"
)

// PrintUseCase orchestrates receipt printing by fetching organisation branding
// from the database, building an ESC/POS receipt, and dispatching it to the printer.
type PrintUseCase struct {
	repo *repository.Queries
}

// NewPrintUseCase creates a new PrintUseCase (repo injected per-request).
func NewPrintUseCase() *PrintUseCase {
	return &PrintUseCase{}
}

// SetRepository injects the per-request tenant repository.
func (uc *PrintUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// PrintReceiptInput is the combined input for printing a receipt.
type PrintReceiptInput struct {
	// OrgID is the organisation whose branding (header/footer) will be applied.
	OrgID int32 `json:"org_id"`

	// Printer describes how / where to send the ESC/POS data.
	Printer printing.PrinterConfig `json:"printer"`

	// Receipt holds the transaction data (items, totals, cashier info, etc.).
	Receipt printing.ReceiptData `json:"receipt"`
}

// PrintReceipt fetches org branding, builds the receipt, and prints it.
func (uc *PrintUseCase) PrintReceipt(tx context.Context, input PrintReceiptInput) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	// 1. Fetch organisation for branding
	org, err := uc.repo.GetOrganization(tx, input.OrgID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound,
			fmt.Sprintf("organisation %d not found: %s", input.OrgID, err.Error()), nil)
	}

	// 2. Build ReceiptTemplate from org record
	tmpl := printing.ReceiptTemplate{
		Header: printing.OrgHeader{
			Name:  org.Name,
			TaxID: org.TaxID.String,
		},
		Footer: printing.OrgFooter{
			ThankYou:   "Thank you for your visit!",
			ReturnNote: "Please keep this receipt for any exchange or return.",
		},
	}

	// Populate optional header fields from LegalName / metadata
	if org.LegalName.Valid && org.LegalName.String != "" {
		// Show legal name as the sub-heading if different from name
		if org.LegalName.String != org.Name {
			tmpl.Header.Address = org.LegalName.String
		}
	}

	// 3. Build ESC/POS bytes
	receiptBytes := printing.BuildReceipt(tmpl, input.Receipt)

	// 4. Send to printer
	if err := printing.Print(input.Printer, receiptBytes); err != nil {
		return utils.NewResponse(utils.CodeError, "print failed: "+err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "receipt printed successfully", map[string]interface{}{
		"bytes_sent": len(receiptBytes),
		"printer":    input.Printer.Mode,
	})
}
