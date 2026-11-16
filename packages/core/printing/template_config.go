package printing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// ReceiptTemplateType selects a template family independently from transaction data.
func ReceiptTemplateType(receiptType string) string {
	t := strings.ToLower(strings.TrimSpace(receiptType))
	if strings.Contains(t, "z-report") || strings.Contains(t, "summary") {
		return "z_report"
	}
	if strings.Contains(t, "return") || strings.Contains(t, "refund") {
		return "return"
	}
	return "receipt"
}

// ReceiptTemplateConfig is the safe, versioned set of layout options stored
// per organization. Rendering remains in Go; templates cannot execute code.
type ReceiptTemplateConfig struct {
	Version              int      `json:"version"`
	PaperWidthMM         int      `json:"paper_width_mm"`
	HeaderLines          []string `json:"header_lines"`
	ShowOrganizationName bool     `json:"show_organization_name"`
	ShowLegalName        bool     `json:"show_legal_name"`
	ShowTaxID            bool     `json:"show_tax_id"`
	ShowReceiptNumber    bool     `json:"show_receipt_number"`
	ShowDate             bool     `json:"show_date"`
	ShowCashier          bool     `json:"show_cashier"`
	ShowTerminal         bool     `json:"show_terminal"`
	ShowCustomer         bool     `json:"show_customer"`
	ShowItems            bool     `json:"show_items"`
	ShowSubtotal         bool     `json:"show_subtotal"`
	ShowDiscount         bool     `json:"show_discount"`
	ShowTax              bool     `json:"show_tax"`
	ShowTotal            bool     `json:"show_total"`
	ShowPaymentMethod    bool     `json:"show_payment_method"`
	ShowPaidChange       bool     `json:"show_paid_change"`
	ShowBarcode          bool     `json:"show_barcode"`
	FooterLines          []string `json:"footer_lines"`
}

func DefaultReceiptTemplateConfig() ReceiptTemplateConfig {
	return ReceiptTemplateConfig{
		Version: 1, PaperWidthMM: 80,
		ShowOrganizationName: true, ShowLegalName: true, ShowTaxID: true,
		ShowReceiptNumber: true, ShowDate: true, ShowCashier: true,
		ShowTerminal: true, ShowCustomer: true, ShowItems: true,
		ShowSubtotal: true, ShowDiscount: true, ShowTax: true,
		ShowTotal: true, ShowPaymentMethod: true, ShowPaidChange: true,
		ShowBarcode: true,
		FooterLines: []string{"Thank you for your visit!", "Please keep this receipt for any exchange or return."},
	}
}

func ParseReceiptTemplateConfig(raw json.RawMessage) (ReceiptTemplateConfig, error) {
	if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return DefaultReceiptTemplateConfig(), nil
	}
	var config ReceiptTemplateConfig
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return ReceiptTemplateConfig{}, fmt.Errorf("invalid receipt template config: %w", err)
	}
	if err := config.Validate(); err != nil {
		return ReceiptTemplateConfig{}, err
	}
	return config, nil
}

func (c ReceiptTemplateConfig) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("unsupported receipt template version %d", c.Version)
	}
	if c.PaperWidthMM != 58 && c.PaperWidthMM != 80 {
		return fmt.Errorf("paper_width_mm must be 58 or 80")
	}
	if len(c.HeaderLines) > 4 || len(c.FooterLines) > 6 {
		return fmt.Errorf("receipt templates support up to 4 header lines and 6 footer lines")
	}
	for _, line := range append(append([]string{}, c.HeaderLines...), c.FooterLines...) {
		line = strings.TrimSpace(line)
		if len(line) > 96 {
			return fmt.Errorf("receipt template lines must be 96 bytes or fewer")
		}
		for _, r := range line {
			if r < 0x20 || r == 0x7f {
				return fmt.Errorf("receipt template lines cannot contain control characters")
			}
		}
	}
	return nil
}

func (c ReceiptTemplateConfig) Columns() int {
	if c.PaperWidthMM == 58 {
		return 32
	}
	return 48
}
