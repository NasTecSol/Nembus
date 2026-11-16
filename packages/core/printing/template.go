package printing

import (
	"bytes"
	"fmt"
	"strings"
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
	Config ReceiptTemplateConfig
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
	Type          string     `json:"type"` // e.g. "Z-REPORT" or standard sales transaction
	ReceiptNumber string     `json:"receipt_number"`
	SalesOrderID  string     `json:"sales_order_id"` // Persisted order used to load the reported ZATCA QR
	StoreID       int32      `json:"store_id"`       // Used for store-scoped template resolution when no sale can be inferred
	Language      string     `json:"language"`       // Language-specific printing template; defaults to English
	Cashier       string     `json:"cashier"`
	Terminal      string     `json:"terminal"`
	Customer      string     `json:"customer"`
	Items         []LineItem `json:"items"`
	Discount      float64    `json:"discount"` // absolute amount
	TaxRate       float64    `json:"tax_rate"` // e.g. 0.05 for 5 %
	Paid          float64    `json:"paid"`     // total cash tendered
	PaymentMethod string     `json:"payment_method"`
	Barcode       string     `json:"barcode"`        // value encoded in barcode (e.g. receipt number)
	QRCodeBase64  string     `json:"qr_code_base64"` // Saved signed payload set by PrintUseCase
	Currency      string     `json:"currency"`       // e.g. "PKR"; defaults to "PKR"

	// Session Summary / Z-Report fields
	OpeningBalance    float64 `json:"opening_balance"`
	TotalSales        float64 `json:"total_sales"`
	ExpectedBalance   float64 `json:"expected_balance"`
	ClosingBalance    float64 `json:"closing_balance"`
	Variance          float64 `json:"variance"`
	TotalTransactions int     `json:"total_transactions"`
	CashSales         float64 `json:"cash_sales"`
	CardSales         float64 `json:"card_sales"`
	OtherSales        float64 `json:"other_sales"`
	OpenedAt          string  `json:"opened_at"`
	ClosedAt          string  `json:"closed_at"`
	ClosingNote       string  `json:"closing_note"`
}

// ─── Receipt builder ──────────────────────────────────────────────────────────

// BuildReceipt composes the ESC/POS byte stream for a receipt using the
// supplied ReceiptTemplate (organisation branding) and ReceiptData (transaction).
func BuildReceipt(tmpl ReceiptTemplate, data ReceiptData) []byte {
	if strings.EqualFold(data.Type, "Z-REPORT") || strings.Contains(strings.ToUpper(data.Type), "SUMMARY") {
		return BuildSummaryReceipt(tmpl, data)
	}
	config := effectiveReceiptTemplateConfig(tmpl.Config)

	var buf bytes.Buffer
	w := func(b []byte) { buf.Write(b) }
	p := func(s string) { buf.WriteString(s) }
	nl := func() { buf.WriteByte(LF) }

	currency := data.Currency
	if currency == "" {
		currency = "SAR"
	}
	width := config.Columns()

	now := time.Now()

	// ── Initialise printer ──────────────────────────────────────────────
	w(CmdInit())

	// ── HEADER ─────────────────────────────────────────────────────────
	writeReceiptHeader(w, p, tmpl, config)
	w(DividerWidth(width))

	// ── RECEIPT INFO ────────────────────────────────────────────────────
	w(CmdAlign("left"))
	if config.ShowReceiptNumber && config.ShowDate {
		w(Col2("Receipt #: "+data.ReceiptNumber, now.Format("02-Jan-06 15:04"), width))
	} else if config.ShowReceiptNumber && data.ReceiptNumber != "" {
		p("Receipt #: " + data.ReceiptNumber + "\n")
	} else if config.ShowDate {
		p(now.Format("02-Jan-06 15:04") + "\n")
	}
	if config.ShowCashier && data.Cashier != "" && config.ShowTerminal && data.Terminal != "" {
		w(Col2("Cashier: "+data.Cashier, "Terminal: "+data.Terminal, width))
	} else if config.ShowCashier && data.Cashier != "" {
		p("Cashier: " + data.Cashier + "\n")
	} else if config.ShowTerminal && data.Terminal != "" {
		p("Terminal: " + data.Terminal + "\n")
	}
	if config.ShowCustomer && data.Customer != "" {
		p("Customer : " + data.Customer + "\n")
	}
	w(DividerWidth(width))

	// ── COLUMN HEADERS ──────────────────────────────────────────────────
	itemWidth, qtyWidth, priceWidth, totalWidth := receiptColumnWidths(config.PaperWidthMM)
	if config.ShowItems {
		w(CmdBold(true))
		p(PadRight("Item", itemWidth) + PadLeft("Qty", qtyWidth) + PadLeft("Price", priceWidth) + PadLeft("Total", totalWidth) + "\n")
		w(CmdBold(false))
		w(DividerWidth(width))
	}

	// ── LINE ITEMS ──────────────────────────────────────────────────────
	subtotal := 0.0
	for _, it := range data.Items {
		if !config.ShowItems {
			break
		}
		total := it.Qty * it.Price
		subtotal += total
		name := PadRight(it.Name, itemWidth)
		qty := PadLeft(fmt.Sprintf("%.0f", it.Qty), qtyWidth)
		price := PadLeft(fmt.Sprintf("%.2f", it.Price), priceWidth)
		tot := PadLeft(fmt.Sprintf("%.2f", total), totalWidth)
		p(name + qty + price + tot + "\n")
	}
	if !config.ShowItems {
		for _, it := range data.Items {
			subtotal += it.Qty * it.Price
		}
	}

	// ── TOTALS ──────────────────────────────────────────────────────────
	w(DividerWidth(width))
	netAfterDisc := subtotal - data.Discount
	tax := netAfterDisc * data.TaxRate
	grandTotal := netAfterDisc + tax
	change := data.Paid - grandTotal

	w(CmdAlign("right"))
	if config.ShowSubtotal {
		p(fmt.Sprintf("Subtotal         : %s %9.2f\n", currency, subtotal))
	}
	if config.ShowDiscount && data.Discount > 0 {
		discPct := 0.0
		if subtotal > 0 {
			discPct = (data.Discount / subtotal) * 100
		}
		p(fmt.Sprintf("Discount (%.0f%%)    : %s -%8.2f\n", discPct, currency, data.Discount))
	}
	if config.ShowTax && data.TaxRate > 0 {
		p(fmt.Sprintf("Tax (%.0f%%)          : %s %9.2f\n", data.TaxRate*100, currency, tax))
	}

	if config.ShowTotal {
		w(CmdBold(true))
		w(CmdDoubleHeight(true))
		p(fmt.Sprintf("TOTAL            : %s %9.2f\n", currency, grandTotal))
		w(CmdDoubleHeight(false))
		w(CmdBold(false))
	}

	if config.ShowPaidChange {
		w(DividerWidth(width))
		p(fmt.Sprintf("Cash Paid        : %s %9.2f\n", currency, data.Paid))
		p(fmt.Sprintf("Change           : %s %9.2f\n", currency, change))
	}
	w(DividerWidth(width))

	// ── PAYMENT METHOD ──────────────────────────────────────────────────
	w(CmdAlign("left"))
	if config.ShowPaymentMethod && data.PaymentMethod != "" {
		p("Payment : " + data.PaymentMethod + "\n")
	}
	nl()

	// ── FOOTER ──────────────────────────────────────────────────────────
	writeReceiptFooter(w, p, tmpl, config, width)

	if data.QRCodeBase64 != "" {
		nl()
		w(CmdAlign("center"))
		w(CmdQRCode(data.QRCodeBase64))
		nl()
		nl()
	}

	// ── BARCODE ─────────────────────────────────────────────────────────
	barcodeValue := data.Barcode
	if barcodeValue == "" {
		barcodeValue = data.ReceiptNumber
	}

	if config.ShowBarcode && barcodeValue != "" {
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

func effectiveReceiptTemplateConfig(config ReceiptTemplateConfig) ReceiptTemplateConfig {
	if config.Version == 0 {
		return DefaultReceiptTemplateConfig()
	}
	return config
}

func receiptColumnWidths(paperWidthMM int) (item, qty, price, total int) {
	if paperWidthMM == 58 {
		return 14, 4, 7, 7
	}
	return 24, 6, 9, 9
}

func writeReceiptHeader(w func([]byte), p func(string), tmpl ReceiptTemplate, config ReceiptTemplateConfig) {
	w(CmdAlign("center"))
	if config.ShowOrganizationName && tmpl.Header.Name != "" {
		w(CmdDoubleSize(true))
		w(CmdBold(true))
		p(tmpl.Header.Name + "\n")
		w(CmdDoubleSize(false))
		w(CmdBold(false))
	}
	if config.ShowLegalName && tmpl.Header.Address != "" {
		p(tmpl.Header.Address + "\n")
	}
	for _, line := range config.HeaderLines {
		p(line + "\n")
	}
	if config.ShowTaxID && tmpl.Header.TaxID != "" {
		p("TIN: " + tmpl.Header.TaxID + "\n")
	}
	if tmpl.Header.Phone != "" {
		p("Tel: " + tmpl.Header.Phone + "\n")
	}
	if tmpl.Header.Website != "" {
		p(tmpl.Header.Website + "\n")
	}
}

func writeReceiptFooter(w func([]byte), p func(string), tmpl ReceiptTemplate, config ReceiptTemplateConfig, width int) {
	w(CmdAlign("center"))
	w(DividerWidth(width))
	if len(config.FooterLines) > 0 {
		for i, line := range config.FooterLines {
			if i == 0 {
				w(CmdBold(true))
			}
			p(line + "\n")
			if i == 0 {
				w(CmdBold(false))
			}
		}
	}
	w(DividerWidth(width))
}

// BuildSummaryReceipt composes the ESC/POS byte stream for a Z-Report / Session Summary report.
func BuildSummaryReceipt(tmpl ReceiptTemplate, data ReceiptData) []byte {
	config := effectiveReceiptTemplateConfig(tmpl.Config)
	width := config.Columns()
	var buf bytes.Buffer
	w := func(b []byte) { buf.Write(b) }
	p := func(s string) { buf.WriteString(s) }

	currency := data.Currency
	if currency == "" {
		currency = "SAR"
	}

	now := time.Now()

	// ── Initialise printer ──────────────────────────────────────────────
	w(CmdInit())

	// ── HEADER ─────────────────────────────────────────────────────────
	writeReceiptHeader(w, p, tmpl, config)
	w(DividerWidth(width))

	// ── REPORT TITLE ───────────────────────────────────────────────────
	w(CmdAlign("center"))
	w(CmdBold(true))
	w(CmdDoubleHeight(true))
	p("Z-REPORT SUMMARY\n")
	w(CmdDoubleHeight(false))
	if data.ReceiptNumber != "" {
		p(data.ReceiptNumber + "\n")
	}
	w(CmdBold(false))
	w(DividerWidth(width))

	// ── SESSION METADATA ───────────────────────────────────────────────
	w(CmdAlign("left"))
	if data.Cashier != "" {
		p("Cashier  : " + data.Cashier + "\n")
	}
	if data.Terminal != "" {
		p("Terminal : " + data.Terminal + "\n")
	}
	if data.OpenedAt != "" {
		p("Opened   : " + data.OpenedAt + "\n")
	}
	if data.ClosedAt != "" {
		p("Closed   : " + data.ClosedAt + "\n")
	}
	w(DividerWidth(width))

	// ── FINANCIAL SUMMARY ──────────────────────────────────────────────
	w(CmdAlign("center"))
	w(CmdBold(true))
	p("FINANCIAL SUMMARY\n")
	w(CmdBold(false))
	w(CmdAlign("left"))

	w(Col2("Opening Balance:", fmt.Sprintf("%s %.2f", currency, data.OpeningBalance), width))
	w(Col2("Total Sales:", fmt.Sprintf("%s %.2f", currency, data.TotalSales), width))
	w(Col2("Expected Balance:", fmt.Sprintf("%s %.2f", currency, data.ExpectedBalance), width))
	w(Col2("Actual Closing:", fmt.Sprintf("%s %.2f", currency, data.ClosingBalance), width))
	w(CmdBold(true))
	w(Col2("Variance:", fmt.Sprintf("%s %.2f", currency, data.Variance), width))
	w(CmdBold(false))
	w(DividerWidth(width))

	// ── SALES BREAKDOWN ────────────────────────────────────────────────
	w(CmdAlign("center"))
	w(CmdBold(true))
	p("SALES BREAKDOWN\n")
	w(CmdBold(false))
	w(CmdAlign("left"))

	w(Col2("Total Transactions:", fmt.Sprintf("%d", data.TotalTransactions), width))
	w(Col2("Cash Sales:", fmt.Sprintf("%s %.2f", currency, data.CashSales), width))
	w(Col2("Card Sales:", fmt.Sprintf("%s %.2f", currency, data.CardSales), width))
	w(Col2("Other Sales:", fmt.Sprintf("%s %.2f", currency, data.OtherSales), width))
	w(DividerWidth(width))

	if data.ClosingNote != "" {
		w(CmdAlign("left"))
		w(CmdBold(true))
		p("CLOSING NOTE:\n")
		w(CmdBold(false))
		p(data.ClosingNote + "\n")
		w(DividerWidth(width))
	}

	// ── FOOTER ──────────────────────────────────────────────────────────
	writeReceiptFooter(w, p, tmpl, config, width)
	w(CmdAlign("center"))
	p(fmt.Sprintf("Printed on %s\n", now.Format("02-Jan-06 15:04:05")))
	w(DividerWidth(width))

	// ── Feed & Cut ───────────────────────────────────────────────────────
	w(CmdFeedLines(2))
	w(CmdCut())

	return buf.Bytes()
}
