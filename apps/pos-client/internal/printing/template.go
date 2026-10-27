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

	var buf bytes.Buffer
	w := func(b []byte) { buf.Write(b) }
	p := func(s string) { buf.WriteString(s) }
	nl := func() { buf.WriteByte(LF) }

	currency := data.Currency

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

// BuildSummaryReceipt composes the ESC/POS byte stream for a Z-Report / Session Summary report.
func BuildSummaryReceipt(tmpl ReceiptTemplate, data ReceiptData) []byte {
	var buf bytes.Buffer
	w := func(b []byte) { buf.Write(b) }
	p := func(s string) { buf.WriteString(s) }

	currency := data.Currency

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
	w(Divider())

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
	w(Divider())

	// ── FINANCIAL SUMMARY ──────────────────────────────────────────────
	w(CmdAlign("center"))
	w(CmdBold(true))
	p("FINANCIAL SUMMARY\n")
	w(CmdBold(false))
	w(CmdAlign("left"))

	w(Col2("Opening Balance:", fmt.Sprintf("%s %.2f", currency, data.OpeningBalance), COLS))
	w(Col2("Total Sales:", fmt.Sprintf("%s %.2f", currency, data.TotalSales), COLS))
	w(Col2("Expected Balance:", fmt.Sprintf("%s %.2f", currency, data.ExpectedBalance), COLS))
	w(Col2("Actual Closing:", fmt.Sprintf("%s %.2f", currency, data.ClosingBalance), COLS))
	w(CmdBold(true))
	w(Col2("Variance:", fmt.Sprintf("%s %.2f", currency, data.Variance), COLS))
	w(CmdBold(false))
	w(Divider())

	// ── SALES BREAKDOWN ────────────────────────────────────────────────
	w(CmdAlign("center"))
	w(CmdBold(true))
	p("SALES BREAKDOWN\n")
	w(CmdBold(false))
	w(CmdAlign("left"))

	w(Col2("Total Transactions:", fmt.Sprintf("%d", data.TotalTransactions), COLS))
	w(Col2("Cash Sales:", fmt.Sprintf("%s %.2f", currency, data.CashSales), COLS))
	w(Col2("Card Sales:", fmt.Sprintf("%s %.2f", currency, data.CardSales), COLS))
	w(Col2("Other Sales:", fmt.Sprintf("%s %.2f", currency, data.OtherSales), COLS))
	w(Divider())

	if data.ClosingNote != "" {
		w(CmdAlign("left"))
		w(CmdBold(true))
		p("CLOSING NOTE:\n")
		w(CmdBold(false))
		p(data.ClosingNote + "\n")
		w(Divider())
	}

	// ── FOOTER ──────────────────────────────────────────────────────────
	w(CmdAlign("center"))
	p(fmt.Sprintf("Printed on %s\n", now.Format("02-Jan-06 15:04:05")))
	w(Divider())

	// ── Feed & Cut ───────────────────────────────────────────────────────
	w(CmdFeedLines(2))
	w(CmdCut())

	return buf.Bytes()
}

// ─── SAMA mada Receipt Builder ───────────────────────────────────────────────

// MadaReceiptData holds all fields required by SAMA POS v6.0.1 Appendix I
type MadaReceiptData struct {
	IsCustomerCopy bool    `json:"is_customer_copy"`
	MerchantNameAr string  `json:"merchant_name_ar"`
	MerchantNameEn string  `json:"merchant_name_en"`
	AddressAr      string  `json:"address_ar"`
	AddressEn      string  `json:"address_en"`
	Phone          string  `json:"phone"`
	DateTime       string  `json:"date_time"` // DD/MM/YYYY HH:MM:SS
	BankID         string  `json:"bank_id"`
	TerminalID     string  `json:"terminal_id"`
	MerchantID     string  `json:"merchant_id"`
	MCC            string  `json:"mcc"`
	STAN           string  `json:"stan"`
	RRN            string  `json:"rrn"`
	SchemeName     string  `json:"scheme_name"` // mada, Visa, MasterCard, etc.
	TxTypeEn       string  `json:"tx_type_en"`   // PURCHASE, REFUND
	TxTypeAr       string  `json:"tx_type_ar"`   // شراء, استرجاع
	MaskedCard     string  `json:"masked_card"`
	ExpiryDate     string  `json:"expiry_date"` // MM/YY
	Amount         float64 `json:"amount"`
	CashbackAmount float64 `json:"cashback_amount"`
	StatusEn       string  `json:"status_en"`       // APPROVED, DECLINED
	StatusAr       string  `json:"status_ar"`       // مقبولة, مرفوضة
	ApprovalCode   string  `json:"approval_code"`   // Auth code
	VerificationEn string  `json:"verification_en"` // CARDHOLDER PIN VERIFIED, NO VERIFICATION REQUIRED
	VerificationAr string  `json:"verification_ar"` // تم التحقق من الرقم السري للعميل
	CardEntryMode  string  `json:"card_entry_mode"` // DIPPED, CONTACTLESS
	AID            string  `json:"aid"`
	TVR            string  `json:"tvr"`
	TSI            string  `json:"tsi"`
	Cryptogram     string  `json:"cryptogram"`
}

// BuildMadaReceipt formats an ESC/POS byte stream for a SAMA mada compliant transaction receipt
func BuildMadaReceipt(d MadaReceiptData) []byte {
	var buf bytes.Buffer
	w := func(b []byte) { buf.Write(b) }
	p := func(s string) { buf.WriteString(s) }
	nl := func() { buf.WriteByte(LF) }

	w(CmdInit())

	// 1. Merchant Header
	w(CmdAlign("center"))
	if d.MerchantNameAr != "" {
		w(CmdBold(true))
		p(d.MerchantNameAr + "\n")
		w(CmdBold(false))
	}
	if d.MerchantNameEn != "" {
		w(CmdBold(true))
		p(d.MerchantNameEn + "\n")
		w(CmdBold(false))
	}
	if d.AddressAr != "" {
		p(d.AddressAr + "\n")
	}
	if d.AddressEn != "" {
		p(d.AddressEn + "\n")
	}
	if d.Phone != "" {
		p("TEL: " + d.Phone + "\n")
	}
	w(Divider())

	// 2. Terminal & Audit Metadata
	w(CmdAlign("left"))
	if d.DateTime != "" {
		p(d.DateTime + "\n")
	}
	w(Col2("TID: "+d.TerminalID, "MID: "+d.MerchantID, COLS))
	w(Col2("MCC: "+d.MCC, "STAN: "+d.STAN, COLS))
	if d.BankID != "" || d.RRN != "" {
		w(Col2("BANK: "+d.BankID, "RRN: "+d.RRN, COLS))
	}
	w(Divider())

	// 3. Scheme & Transaction Type
	w(CmdAlign("center"))
	w(CmdBold(true))
	schemeDisplay := d.SchemeName
	if schemeDisplay == "" {
		schemeDisplay = "mada / SPAN"
	}
	p("الشبكة السعودية للمدفوعات\n")
	p(schemeDisplay + "\n")
	nl()

	txTypeAr := d.TxTypeAr
	if txTypeAr == "" {
		txTypeAr = "شراء"
	}
	txTypeEn := d.TxTypeEn
	if txTypeEn == "" {
		txTypeEn = "PURCHASE"
	}
	p(txTypeAr + " / " + txTypeEn + "\n")
	w(CmdBold(false))
	w(Divider())

	// 4. Card Details
	w(CmdAlign("left"))
	if d.MaskedCard != "" {
		w(Col2(d.MaskedCard, d.ExpiryDate, COLS))
	}
	w(Divider())

	// 5. Amount
	w(CmdAlign("center"))
	p("مبلغ الشراء / PURCHASE AMOUNT\n")
	w(CmdDoubleSize(true))
	w(CmdBold(true))
	p(fmt.Sprintf("SAR %.2f\n", d.Amount))
	w(CmdDoubleSize(false))
	w(CmdBold(false))

	if d.CashbackAmount > 0 {
		p(fmt.Sprintf("مبلغ نقد / NAQD: SAR %.2f\n", d.CashbackAmount))
		p(fmt.Sprintf("المجموع / TOTAL: SAR %.2f\n", d.Amount+d.CashbackAmount))
	}
	w(Divider())

	// 6. Approval Status & Verification
	w(CmdAlign("center"))
	w(CmdBold(true))
	statusAr := d.StatusAr
	if statusAr == "" {
		statusAr = "مقبولة"
	}
	statusEn := d.StatusEn
	if statusEn == "" {
		statusEn = "APPROVED"
	}
	p(statusAr + " / " + statusEn + "\n")
	w(CmdBold(false))

	if d.VerificationAr != "" || d.VerificationEn != "" {
		if d.VerificationAr != "" {
			p(d.VerificationAr + "\n")
		}
		if d.VerificationEn != "" {
			p(d.VerificationEn + "\n")
		}
	}
	if d.ApprovalCode != "" {
		p("رمز الموافقة / APPROVAL CODE: " + d.ApprovalCode + "\n")
	}
	w(Divider())

	// 7. SAMA Retain Notice & Copy indicator
	w(CmdAlign("center"))
	p("شكرًا لاستخدامكم مدى\n")
	p("THANK YOU FOR USING mada\n")
	p("يرجى الاحتفاظ بالإيصال\n")
	p("PLEASE RETAIN RECEIPT\n")
	nl()

	w(CmdBold(true))
	if d.IsCustomerCopy {
		p("** نسخة العميل / CUSTOMER COPY **\n")
	} else {
		p("** نسخة التاجر / RETAILER COPY **\n")
	}
	w(CmdBold(false))

	// 8. EMV Security Line
	if d.CardEntryMode != "" || d.AID != "" {
		w(Divider())
		w(CmdAlign("left"))
		emvLine1 := fmt.Sprintf("%s %s %s", d.CardEntryMode, d.AID, d.TVR)
		p(strings.TrimSpace(emvLine1) + "\n")
		if d.TSI != "" || d.Cryptogram != "" {
			emvLine2 := fmt.Sprintf("TSI: %s AC: %s", d.TSI, d.Cryptogram)
			p(strings.TrimSpace(emvLine2) + "\n")
		}
	}

	w(CmdFeedLines(3))
	w(CmdCut())
	return buf.Bytes()
}

