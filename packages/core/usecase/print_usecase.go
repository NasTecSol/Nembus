package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/NasTecSol/nembus-client/utils"
	"github.com/NasTecSol/nembus-core/printing"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/google/uuid"
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
	// The QR on a tax receipt must be the order's saved signed QR, and its
	// report must be accepted. Never accept an arbitrary QR from the browser.
	input.Receipt.QRCodeBase64 = ""
	qrStatus := "not_applicable"
	if !strings.EqualFold(input.Receipt.Type, "Z-REPORT") && !strings.Contains(strings.ToUpper(input.Receipt.Type), "SUMMARY") {
		enabled, err := OrganizationZatcaEnabled(&org)
		if err != nil {
			qrStatus = "unavailable"
		} else if enabled {
			qrStatus = "not_reported"
			orderID := strings.TrimSpace(input.Receipt.SalesOrderID)
			if orderID == "" {
				orderID = strings.TrimSpace(input.Receipt.ReceiptNumber)
			}
			if _, err := uuid.Parse(orderID); err != nil {
				qrStatus = "unavailable"
			} else {
				qr, err := uc.repo.GetReportedSalesOrderQR(tx, input.OrgID, orderID)
				if err != nil || len(qr) > 2000 {
					qrStatus = "unavailable"
				} else if qr != "" {
					input.Receipt.QRCodeBase64 = qr
					qrStatus = "included"
				}
			}
		}
	}

	// Use organization currency_code from DB if receipt currency is not specified
	if input.Receipt.Currency == "" && org.CurrencyCode.Valid && strings.TrimSpace(org.CurrencyCode.String) != "" {
		input.Receipt.Currency = strings.TrimSpace(org.CurrencyCode.String)
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

	message := "receipt printed successfully"
	if qrStatus == "not_reported" || qrStatus == "unavailable" {
		message = "receipt printed without ZATCA QR; reporting is not complete or QR is unavailable"
	}
	return utils.NewResponse(utils.CodeOK, message, map[string]interface{}{
		"bytes_sent":     len(receiptBytes),
		"printer":        input.Printer.Mode,
		"qr_code_base64": input.Receipt.QRCodeBase64,
		"qr_status":      qrStatus,
	})
}
