package printing_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/printing"
)

func Test_BankSlipFormatting(t *testing.T) {
	data := &printing.BankSlipData{
		MerchantName: "Nembus Retail Store",
		StoreID:      "ST-101",
		TID:          "TID88801",
		MID:          "MID99901",
		CardScheme:   "mada",
		MaskedPAN:    "588845******1234",
		AID:          "A0000002281010",
		RRN:          "999000123456",
		AuthCode:     "OK7890",
		Amount:       250.75,
		Currency:     "SAR",
		Status:       "APPROVED",
		Timestamp:    time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC),
	}

	slipBytes := printing.FormatBankSlip(data)
	if len(slipBytes) == 0 {
		t.Fatal("expected non-empty bank slip bytes")
	}

	// Verify cut command GS V 66 0 is included at the end
	cutCmd := []byte{0x1D, 'V', 66, 0}
	if !bytes.Contains(slipBytes, cutCmd) {
		t.Errorf("expected bank slip to contain cut command GS V 66 0")
	}

	// Test Dual Cut Receipt
	dummyInvoice := []byte("=== TAX INVOICE RECEIPT ===\nTotal: 250.75 SAR\n")
	dualReceipt := printing.BuildDualCutReceipt(dummyInvoice, data)

	// Should contain two cut commands (Cut 1 for invoice, Cut 2 for card slip)
	cutCount := bytes.Count(dualReceipt, cutCmd)
	if cutCount != 2 {
		t.Errorf("expected 2 cut commands in dual-cut receipt, got %d", cutCount)
	}
}
