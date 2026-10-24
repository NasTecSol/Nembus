package printing

import (
	"bytes"
	"fmt"
	"time"
)

type BankSlipData struct {
	MerchantName string
	StoreID      string
	TID          string
	MID          string
	CardScheme   string
	MaskedPAN    string
	AID          string
	RRN          string
	AuthCode     string
	Amount       float64
	Currency     string
	Status       string
	Timestamp    time.Time
}

// FormatBankSlip formats a standalone Mada/EMV card transaction slip.
func FormatBankSlip(d *BankSlipData) []byte {
	var buf bytes.Buffer

	buf.Write(CmdInit())
	buf.Write(CmdAlign("center"))
	buf.Write(CmdBold(true))
	buf.Write([]byte(fmt.Sprintf("%s\n", d.MerchantName)))
	buf.Write([]byte("CARD PAYMENT RECEIPT / إشعار عملية الدفع\n"))
	buf.Write(CmdBold(false))
	buf.Write(Divider())

	buf.Write(CmdAlign("left"))
	buf.Write(Col2("CARD SCHEME / البطاقة:", d.CardScheme, COLS))
	buf.Write(Col2("CARD NUMBER / رقم البطاقة:", d.MaskedPAN, COLS))
	buf.Write(Col2("AID:", d.AID, COLS))
	buf.Write(Col2("TERMINAL ID / المعرف:", d.TID, COLS))
	buf.Write(Col2("MERCHANT ID / التاجر:", d.MID, COLS))
	buf.Write(Col2("RRN / الرقم المرجعي:", d.RRN, COLS))
	buf.Write(Col2("AUTH CODE / رمز الموافقة:", d.AuthCode, COLS))
	buf.Write(Divider())

	buf.Write(CmdBold(true))
	cur := d.Currency
	if cur == "" {
		cur = "SAR"
	}
	buf.Write(Col2("TOTAL AMOUNT / المبلغ:", fmt.Sprintf("%.2f %s", d.Amount, cur), COLS))
	buf.Write(CmdAlign("center"))
	buf.Write([]byte(fmt.Sprintf("\n*** %s ***\n", d.Status)))
	buf.Write(CmdBold(false))

	ts := d.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}
	buf.Write([]byte(fmt.Sprintf("%s\n\n", ts.Format("2006-01-02 15:04:05"))))
	buf.Write(CmdFeedLines(3))
	buf.Write(CmdCut())

	return buf.Bytes()
}

// BuildDualCutReceipt combines a ZATCA invoice receipt and standalone bank card slip with GS V 66 0 separator.
func BuildDualCutReceipt(invoiceReceiptBytes []byte, slipData *BankSlipData) []byte {
	var combined bytes.Buffer
	// Cut 1: Invoice Receipt followed by partial cut
	combined.Write(invoiceReceiptBytes)
	combined.Write(CmdFeedLines(3))
	combined.Write(CmdCut()) // GS V 66 0

	// Cut 2: Bank Slip followed by partial cut
	if slipData != nil {
		bankSlipBytes := FormatBankSlip(slipData)
		combined.Write(bankSlipBytes)
	}
	return combined.Bytes()
}
