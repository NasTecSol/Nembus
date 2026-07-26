package printing

import (
	"strings"
	"testing"
)

func TestBuildReceipt_StandardSales(t *testing.T) {
	tmpl := ReceiptTemplate{
		Header: OrgHeader{Name: "Test Store"},
	}
	data := ReceiptData{
		ReceiptNumber: "INV-1001",
		Cashier:       "John",
		Items: []LineItem{
			{Name: "Item 1", Qty: 2, Price: 10.0},
		},
		Paid: 20.0,
	}

	out := BuildReceipt(tmpl, data)
	if len(out) == 0 {
		t.Fatal("expected non-empty byte slice for standard receipt")
	}

	strOut := string(out)
	if !strings.Contains(strOut, "Item") || !strings.Contains(strOut, "Qty") {
		t.Errorf("expected standard receipt header columns 'Item' and 'Qty'")
	}
}

func TestBuildReceipt_ZReportSummary(t *testing.T) {
	tmpl := ReceiptTemplate{
		Header: OrgHeader{Name: "Test Store"},
	}
	data := ReceiptData{
		Type:              "Z-REPORT",
		ReceiptNumber:     "SESS-99",
		Cashier:           "Alice",
		OpeningBalance:    100.0,
		TotalSales:        550.0,
		ExpectedBalance:   650.0,
		ClosingBalance:    650.0,
		Variance:          0.0,
		TotalTransactions: 12,
		CashSales:         300.0,
		CardSales:         250.0,
	}

	out := BuildReceipt(tmpl, data)
	if len(out) == 0 {
		t.Fatal("expected non-empty byte slice for Z-Report")
	}

	strOut := string(out)
	if !strings.Contains(strOut, "Z-REPORT SUMMARY") {
		t.Errorf("expected 'Z-REPORT SUMMARY' in output")
	}
	if !strings.Contains(strOut, "FINANCIAL SUMMARY") {
		t.Errorf("expected 'FINANCIAL SUMMARY' section")
	}
	if !strings.Contains(strOut, "SALES BREAKDOWN") {
		t.Errorf("expected 'SALES BREAKDOWN' section")
	}
}
