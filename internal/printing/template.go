package printing

import (
	"bytes"
	"fmt"
	"time"
)

// ─── Organisation branding ────────────────────────────────────────────────────

// OrgHeader holds the organisation details printed at the top of every receipt.
type OrgHeader struct {
	Name    string // e.g. "MY STORE"
	Address string // e.g. "123 Main Street, Shop No. 5, Rawalpindi"
	Phone   string // e.g. "+92-51-1234567"
	TaxID   string // e.g. "PK-123456789"
	Website string // e.g. "www.mystore.com"
}

// OrgFooter holds the sign-off lines printed at the bottom of every receipt.
type OrgFooter struct {
	ThankYou   string // e.g. "Thank you for your visit!"
	ReturnNote string // e.g. "Please keep this receipt for any exchange or return."
	Website    string // optionally repeated at foot
}

// ReceiptTemplate combines header + footer branding for one organisation.
type ReceiptTemplate struct {
	Header OrgHeader
	Footer OrgFooter
}

// ─── Receipt data (provided by caller / frontend) ─────────────────────────────

// LineItem represents one sold item on the receipt.
type LineItem struct {
	Name  string  `json:"name"`
	Qty   float64 `json:"qty"`
	Price float64 `json:"price"`
}

// ReceiptData is the per-transaction payload sent by the frontend.
type ReceiptData struct {
	ReceiptNumber string     `json:"receipt_number"`
	Cashier       string     `json:"cashier"`
	Terminal      string     `json:"terminal"`
	Customer      string     `json:"customer"`
	Items         []LineItem `json:"items"`
	Discount      float64    `json:"discount"` // absolute amount
	TaxRate       float64    `json:"tax_rate"` // e.g. 0.05 for 5 %
	Paid          float64    `json:"paid"`     // total cash tendered
	PaymentMethod string     `json:"payment_method"`
	Barcode       string     `json:"barcode"`  // value encoded in barcode (e.g. receipt number)
	Currency      string     `json:"currency"` // e.g. "PKR"; defaults to "PKR"
}

// ─── Receipt builder ──────────────────────────────────────────────────────────

// BuildReceipt composes the ESC/POS byte stream for a receipt using the
// supplied ReceiptTemplate (organisation branding) and ReceiptData (transaction).
func BuildReceipt(tmpl ReceiptTemplate, data ReceiptData) []byte {
	var buf bytes.Buffer
	w := func(b []byte) { buf.Write(b) }
	p := func(s string) { buf.WriteString(s) }
	nl := func() { buf.WriteByte(LF) }

	currency := data.Currency
	if currency == "" {
		currency = "PKR"
	}

	now := time.Now()

	// ── Initialise printer ──────────────────────────────────────────────
	w(CmdInit())

	// ── HEADER ─────────────────────────────────────────────────────────
	w(CmdAlign("center"))
	w(CmdDoubleSize(true))
	w(CmdBold(true))
	p(tmpl.Header.Name + "\n")
	w(CmdDoubleSize(false))
	w(CmdBold(false))

	if tmpl.Header.Address != "" {
		p(tmpl.Header.Address + "\n")
	}
	if tmpl.Header.Phone != "" {
		p("Tel: " + tmpl.Header.Phone + "\n")
	}
	if tmpl.Header.TaxID != "" {
		p("TIN: " + tmpl.Header.TaxID + "\n")
	}
	w(Divider())

	// ── RECEIPT INFO ────────────────────────────────────────────────────
	w(CmdAlign("left"))
	w(Col2("Receipt #: "+data.ReceiptNumber, now.Format("02-Jan-06 15:04"), COLS))

	if data.Cashier != "" {
		w(Col2("Cashier  : "+data.Cashier, "Terminal: "+data.Terminal, COLS))
	}
	if data.Customer != "" {
		p("Customer : " + data.Customer + "\n")
	}
	w(Divider())

	// ── COLUMN HEADERS ──────────────────────────────────────────────────
	w(CmdBold(true))
	header := PadRight("Item", 24) +
		PadLeft("Qty", 6) +
		PadLeft("Price", 9) +
		PadLeft("Total", 9) + "\n"
	p(header)
	w(CmdBold(false))
	w(Divider())

	// ── LINE ITEMS ──────────────────────────────────────────────────────
	subtotal := 0.0
	for _, it := range data.Items {
		total := it.Qty * it.Price
		subtotal += total
		name := PadRight(it.Name, 24)
		qty := PadLeft(fmt.Sprintf("%.0f", it.Qty), 6)
		price := PadLeft(fmt.Sprintf("%.2f", it.Price), 9)
		tot := PadLeft(fmt.Sprintf("%.2f", total), 9)
		p(name + qty + price + tot + "\n")
	}

	// ── TOTALS ──────────────────────────────────────────────────────────
	w(Divider())
	netAfterDisc := subtotal - data.Discount
	tax := netAfterDisc * data.TaxRate
	grandTotal := netAfterDisc + tax
	change := data.Paid - grandTotal

	w(CmdAlign("right"))
	p(fmt.Sprintf("Subtotal         : %s %9.2f\n", currency, subtotal))
	if data.Discount > 0 {
		discPct := 0.0
		if subtotal > 0 {
			discPct = (data.Discount / subtotal) * 100
		}
		p(fmt.Sprintf("Discount (%.0f%%)    : %s -%8.2f\n", discPct, currency, data.Discount))
	}
	if data.TaxRate > 0 {
		p(fmt.Sprintf("Tax (%.0f%%)          : %s %9.2f\n", data.TaxRate*100, currency, tax))
	}

	w(CmdBold(true))
	w(CmdDoubleHeight(true))
	p(fmt.Sprintf("TOTAL            : %s %9.2f\n", currency, grandTotal))
	w(CmdDoubleHeight(false))
	w(CmdBold(false))

	w(Divider())
	p(fmt.Sprintf("Cash Paid        : %s %9.2f\n", currency, data.Paid))
	p(fmt.Sprintf("Change           : %s %9.2f\n", currency, change))
	w(Divider())

	// ── PAYMENT METHOD ──────────────────────────────────────────────────
	w(CmdAlign("left"))
	if data.PaymentMethod != "" {
		p("Payment : " + data.PaymentMethod + "\n")
	}
	nl()

	// ── FOOTER ──────────────────────────────────────────────────────────
	w(CmdAlign("center"))
	w(Divider())
	if tmpl.Footer.ThankYou != "" {
		w(CmdBold(true))
		p("** " + tmpl.Footer.ThankYou + " **\n")
		w(CmdBold(false))
	}
	if tmpl.Footer.ReturnNote != "" {
		p(tmpl.Footer.ReturnNote + "\n")
	}
	if tmpl.Footer.Website != "" {
		p(tmpl.Footer.Website + "\n")
	}
	w(Divider())

	// ── BARCODE ─────────────────────────────────────────────────────────
	barcodeValue := data.Barcode
	if barcodeValue == "" {
		barcodeValue = data.ReceiptNumber
	}

	if barcodeValue != "" {
		nl()
		nl()
		w(CmdAlign("center"))
		w(CmdBarcodeHeight(50))
		w(CmdBarcodeWidth(2))
		w(CmdBarcodeHRI(1)) // text below
		w(CmdBarcode39(barcodeValue))
		nl()
		nl()
	}

	// ── Feed & Cut ───────────────────────────────────────────────────────
	w(CmdFeedLines(2))
	w(CmdCut())

	return buf.Bytes()
}
