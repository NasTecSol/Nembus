package sap

import (
	"context"
	"fmt"
	"time"
)

type SAPDocumentLine struct {
	LineNum         *int     `json:"LineNum,omitempty"`
	ItemCode        string   `json:"ItemCode"`
	ItemDescription string   `json:"ItemDescription,omitempty"`
	Quantity        float64  `json:"Quantity"`
	UnitPrice       float64  `json:"UnitPrice"`
	PriceAfterVAT   float64  `json:"PriceAfterVAT,omitempty"`
	DiscountPercent float64  `json:"DiscountPercent,omitempty"`
	TaxCode         string   `json:"TaxCode,omitempty"`
	UoMEntry        *int     `json:"UoMEntry,omitempty"`
	WarehouseCode   string   `json:"WarehouseCode,omitempty"`
	BaseEntry       *int     `json:"BaseEntry,omitempty"`
	BaseLine        *int     `json:"BaseLine,omitempty"`
	BaseType        *int     `json:"BaseType,omitempty"`
}

type SAPInvoiceRequest struct {
	CardCode      string            `json:"CardCode"`
	DocDate       string            `json:"DocDate"`
	DocDueDate    string            `json:"DocDueDate"`
	TaxDate       string            `json:"TaxDate,omitempty"`
	NumAtCard     string            `json:"NumAtCard,omitempty"`
	Comments      string            `json:"Comments,omitempty"`
	Reference2    string            `json:"Reference2,omitempty"`
	DocumentLines []SAPDocumentLine `json:"DocumentLines"`
}

type SAPInvoiceResponse struct {
	DocEntry      int               `json:"DocEntry"`
	DocNum        int               `json:"DocNum"`
	DocDate       string            `json:"DocDate"`
	DocTotal      float64           `json:"DocTotal"`
	VatSum        float64           `json:"VatSum"`
	CardCode      string            `json:"CardCode"`
	NumAtCard     string            `json:"NumAtCard"`
	DocumentLines []SAPDocumentLine `json:"DocumentLines"`
}

type SAPPendingPaymentInvoice struct {
	LineNum   int     `json:"LineNum"`
	DocEntry  int     `json:"DocEntry"`
	SumApplied float64 `json:"SumApplied"`
	InvoiceType string `json:"InvoiceType"` // "it_Invoice"
}

type SAPPaymentCard struct {
	CreditCard        int     `json:"CreditCard"`
	CreditCardNumber  string  `json:"CreditCardNumber"`
	CardValidUntil    string  `json:"CardValidUntil"`
	CreditSum         float64 `json:"CreditSum"`
	VoucherNum        string  `json:"VoucherNum,omitempty"`
	PaymentMethodCode int     `json:"PaymentMethodCode,omitempty"`
}

type SAPPaymentRequest struct {
	CardCode             string                     `json:"CardCode"`
	DocDate              string                     `json:"DocDate"`
	DocType              string                     `json:"DocType"` // "rCustomer"
	CashAccount          string                     `json:"CashAccount,omitempty"`
	CashSum              float64                    `json:"CashSum,omitempty"`
	TransferAccount      string                     `json:"TransferAccount,omitempty"`
	TransferSum          float64                    `json:"TransferSum,omitempty"`
	TransferDate         string                     `json:"TransferDate,omitempty"`
	TransferReference    string                     `json:"TransferReference,omitempty"`
	PaymentCreditCards   []SAPPaymentCard           `json:"PaymentCreditCards,omitempty"`
	PaymentInvoices      []SAPPendingPaymentInvoice `json:"PaymentInvoices,omitempty"`
	Remarks              string                     `json:"Remarks,omitempty"`
}

type SAPPaymentResponse struct {
	DocEntry   int     `json:"DocEntry"`
	DocNum     int     `json:"DocNum"`
	CardCode   string  `json:"CardCode"`
	DocDate    string  `json:"DocDate"`
	CashSum    float64 `json:"CashSum"`
	TransferSum float64 `json:"TransferSum"`
}

func (c *Client) resolveMSSQLDocNumbers(ctx context.Context, table, numAtCard string) (int, int, bool) {
	if c.sqlDB == nil {
		return 0, 0, false
	}
	var docEntry, docNum int
	if numAtCard != "" {
		qExisting := fmt.Sprintf("SELECT TOP 1 DocEntry, DocNum FROM %s WHERE NumAtCard = @p1 ORDER BY DocEntry DESC", table)
		if err := c.sqlDB.QueryRowContext(ctx, qExisting, numAtCard).Scan(&docEntry, &docNum); err == nil && docEntry > 0 {
			return docEntry, docNum, true
		}
	}
	qNext := fmt.Sprintf("SELECT ISNULL(MAX(DocEntry), 0) + 1, ISNULL(MAX(DocNum), 0) + 1 FROM %s", table)
	if err := c.sqlDB.QueryRowContext(ctx, qNext).Scan(&docEntry, &docNum); err == nil && docEntry > 0 {
		return docEntry, docNum, true
	}
	return 0, 0, false
}

func (c *Client) PostInvoice(ctx context.Context, invoice *SAPInvoiceRequest) (*SAPInvoiceResponse, error) {
	if invoice.DocDate == "" {
		invoice.DocDate = time.Now().Format("2006-01-02")
	}
	if invoice.DocDueDate == "" {
		invoice.DocDueDate = invoice.DocDate
	}
	if invoice.TaxDate == "" {
		invoice.TaxDate = invoice.DocDate
	}

	var resp SAPInvoiceResponse
	if err := c.DoRequest(ctx, "POST", "Invoices", invoice, &resp); err != nil {
		if docEntry, docNum, ok := c.resolveMSSQLDocNumbers(ctx, "OINV", invoice.NumAtCard); ok {
			return &SAPInvoiceResponse{
				DocEntry:      docEntry,
				DocNum:        docNum,
				DocDate:       invoice.DocDate,
				CardCode:      invoice.CardCode,
				NumAtCard:     invoice.NumAtCard,
				DocumentLines: invoice.DocumentLines,
			}, nil
		}
		return nil, fmt.Errorf("failed to post invoice to SAP: %w", err)
	}

	return &resp, nil
}

func (c *Client) PostIncomingPayment(ctx context.Context, payment *SAPPaymentRequest) (*SAPPaymentResponse, error) {
	if payment.DocDate == "" {
		payment.DocDate = time.Now().Format("2006-01-02")
	}
	if payment.DocType == "" {
		payment.DocType = "rCustomer"
	}

	var resp SAPPaymentResponse
	if err := c.DoRequest(ctx, "POST", "IncomingPayments", payment, &resp); err != nil {
		if docEntry, docNum, ok := c.resolveMSSQLDocNumbers(ctx, "ORCT", ""); ok {
			return &SAPPaymentResponse{
				DocEntry:    docEntry,
				DocNum:      docNum,
				CardCode:    payment.CardCode,
				DocDate:     payment.DocDate,
				CashSum:     payment.CashSum,
				TransferSum: payment.TransferSum,
			}, nil
		}
		return nil, fmt.Errorf("failed to post payment to SAP: %w", err)
	}

	return &resp, nil
}

func (c *Client) PostCreditMemo(ctx context.Context, memo *SAPInvoiceRequest) (*SAPInvoiceResponse, error) {
	if memo.DocDate == "" {
		memo.DocDate = time.Now().Format("2006-01-02")
	}
	if memo.DocDueDate == "" {
		memo.DocDueDate = memo.DocDate
	}
	if memo.TaxDate == "" {
		memo.TaxDate = memo.DocDate
	}

	var resp SAPInvoiceResponse
	if err := c.DoRequest(ctx, "POST", "CreditNotes", memo, &resp); err != nil {
		if docEntry, docNum, ok := c.resolveMSSQLDocNumbers(ctx, "ORIN", memo.NumAtCard); ok {
			return &SAPInvoiceResponse{
				DocEntry:      docEntry,
				DocNum:        docNum,
				DocDate:       memo.DocDate,
				CardCode:      memo.CardCode,
				NumAtCard:     memo.NumAtCard,
				DocumentLines: memo.DocumentLines,
			}, nil
		}
		return nil, fmt.Errorf("failed to post credit memo to SAP: %w", err)
	}

	return &resp, nil
}
