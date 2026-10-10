package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type InvoiceOutput struct {
	ID                 uuid.UUID                  `json:"id"`
	InvoiceNumber      string                     `json:"invoice_number"`
	OrganizationID     int32                      `json:"organization_id"`
	StoreID            pgtype.Int4                `json:"store_id"`
	CustomerID         int32                      `json:"customer_id"`
	CustomerName       string                     `json:"customer_name"`
	CustomerEmail      pgtype.Text                `json:"customer_email"`
	CustomerPhone      pgtype.Text                `json:"customer_phone"`
	CustomerTaxID      pgtype.Text                `json:"customer_tax_id"`
	InvoiceType        repository.InvoiceType     `json:"invoice_type"`
	InvoiceStatus      repository.InvoiceStatus   `json:"invoice_status"`
	SalesOrderID       pgtype.UUID                `json:"sales_order_id"`
	RelatedInvoiceID   pgtype.UUID                `json:"related_invoice_id"`
	InvoiceDate        pgtype.Date                `json:"invoice_date"`
	DueDate            pgtype.Date                `json:"due_date"`
	SentDate           pgtype.Date                `json:"sent_date"`
	PaidDate           pgtype.Date                `json:"paid_date"`
	Subtotal           pgtype.Numeric             `json:"subtotal"`
	DiscountAmount     pgtype.Numeric             `json:"discount_amount"`
	TaxAmount          pgtype.Numeric             `json:"tax_amount"`
	ShippingAmount     pgtype.Numeric             `json:"shipping_amount"`
	AdjustmentAmount   pgtype.Numeric             `json:"adjustment_amount"`
	TotalAmount        pgtype.Numeric             `json:"total_amount"`
	PaidAmount         pgtype.Numeric             `json:"paid_amount"`
	CreditApplied      pgtype.Numeric             `json:"credit_applied"`
	BalanceDue         pgtype.Numeric             `json:"balance_due"`
	PaymentTerms       pgtype.Text                `json:"payment_terms"`
	CurrencyCode       pgtype.Text                `json:"currency_code"`
	ExchangeRate       pgtype.Numeric             `json:"exchange_rate"`
	BillingAddress     json.RawMessage            `json:"billing_address"`
	ShippingAddress    json.RawMessage            `json:"shipping_address"`
	IsRecurring        pgtype.Bool                `json:"is_recurring"`
	RecurrencePattern  pgtype.Text                `json:"recurrence_pattern"`
	NextInvoiceDate    pgtype.Date                `json:"next_invoice_date"`
	PdfUrl             pgtype.Text                `json:"pdf_url"`
	DocumentHash       pgtype.Text                `json:"document_hash"`
	ReminderSentCount  pgtype.Int4                `json:"reminder_sent_count"`
	LastReminderSentAt pgtype.Timestamp           `json:"last_reminder_sent_at"`
	Notes              pgtype.Text                `json:"notes"`
	InternalNotes      pgtype.Text                `json:"internal_notes"`
	ReferenceNumber    pgtype.Text                `json:"reference_number"`
	CreatedByUserID    pgtype.Int4                `json:"created_by_user_id"`
	Tags               []string                   `json:"tags"`
	Metadata           json.RawMessage            `json:"metadata"`
	CreatedAt          pgtype.Timestamp                `json:"created_at"`
	UpdatedAt          pgtype.Timestamp                `json:"updated_at"`
	Lines              []repository.ListInvoiceLinesRow `json:"lines,omitempty"`
	Payments           []repository.ListInvoicePaymentsRow `json:"payments,omitempty"`
}

type InvoiceListResponse struct {
	Data       []InvoiceOutput `json:"data"`
	TotalCount int64           `json:"total_count"`
	Page       int32           `json:"page"`
	Limit      int32           `json:"limit"`
	TotalPages int32           `json:"total_pages"`
}

type CreateInvoiceLineInput struct {
	LineNumber       int32                  `json:"line_number"`
	Description      string                 `json:"description"`
	ItemType         *string                `json:"item_type"` // product, service, discount, shipping, fee
	ProductID        *int32                 `json:"product_id"`
	ProductVariantID *int32                 `json:"product_variant_id"`
	ProductSku       *string                `json:"product_sku"`
	OrderLineID      *string                `json:"order_line_id"`
	Quantity         string                 `json:"quantity"`
	UnitPrice        string                 `json:"unit_price"`
	DiscountAmount   *string                `json:"discount_amount"`
	TaxAmount        *string                `json:"tax_amount"`
	LineTotal        string                 `json:"line_total"`
	TaxCategoryID    *int32                 `json:"tax_category_id"`
	TaxRate          *string                `json:"tax_rate"`
	UomID            *int32                 `json:"uom_id"`
	Metadata         map[string]interface{} `json:"metadata"`
}

type CreateInvoiceInput struct {
	InvoiceNumber     string                   `json:"invoice_number"`
	OrganizationID    int32                    `json:"organization_id"`
	StoreID           *int32                   `json:"store_id"`
	CustomerID        int32                    `json:"customer_id"`
	CustomerName      string                   `json:"customer_name"`
	CustomerEmail     *string                  `json:"customer_email"`
	CustomerPhone     *string                  `json:"customer_phone"`
	CustomerTaxID     *string                  `json:"customer_tax_id"`
	InvoiceType       string                   `json:"invoice_type"`   // standard, proforma, credit_note, debit_note, recurring
	InvoiceStatus     string                   `json:"invoice_status"` // draft, sent, viewed, partially_paid, paid, overdue, cancelled, refunded
	SalesOrderID      *string                  `json:"sales_order_id"`
	RelatedInvoiceID  *string                  `json:"related_invoice_id"`
	InvoiceDate       string                   `json:"invoice_date"` // YYYY-MM-DD
	DueDate           string                   `json:"due_date"`     // YYYY-MM-DD
	PaymentTerms      *string                  `json:"payment_terms"`
	CurrencyCode      *string                  `json:"currency_code"`
	ExchangeRate      *string                  `json:"exchange_rate"`
	BillingAddress    map[string]interface{}   `json:"billing_address"`
	ShippingAddress   map[string]interface{}   `json:"shipping_address"`
	IsRecurring       *bool                    `json:"is_recurring"`
	RecurrencePattern *string                  `json:"recurrence_pattern"`
	NextInvoiceDate   *string                  `json:"next_invoice_date"`
	Notes             *string                  `json:"notes"`
	InternalNotes     *string                  `json:"internal_notes"`
	ReferenceNumber   *string                  `json:"reference_number"`
	CreatedByUserID   *int32                   `json:"created_by_user_id"`
	Tags              []string                 `json:"tags"`
	Metadata          map[string]interface{}   `json:"metadata"`
	Lines             []CreateInvoiceLineInput `json:"lines"`
}

type UpdateInvoiceInput struct {
	CustomerName      string                 `json:"customer_name"`
	CustomerEmail     *string                `json:"customer_email"`
	CustomerPhone     *string                `json:"customer_phone"`
	DueDate           string                 `json:"due_date"`
	PaymentTerms      *string                `json:"payment_terms"`
	BillingAddress    map[string]interface{} `json:"billing_address"`
	ShippingAddress   map[string]interface{} `json:"shipping_address"`
	Notes             *string                `json:"notes"`
	InternalNotes     *string                `json:"internal_notes"`
	ReferenceNumber   *string                `json:"reference_number"`
	Tags              []string               `json:"tags"`
	Metadata          map[string]interface{} `json:"metadata"`
	IsRecurring       *bool                  `json:"is_recurring"`
	RecurrencePattern *string                `json:"recurrence_pattern"`
	NextInvoiceDate   *string                `json:"next_invoice_date"`
}

type RecordInvoicePaymentInput struct {
	PaymentNumber    string                 `json:"payment_number"`
	PaymentDate      string                 `json:"payment_date"` // YYYY-MM-DD
	PaymentAmount    string                 `json:"payment_amount"`
	PaymentMethod    string                 `json:"payment_method"` // cash, card, bank_transfer, etc.
	PaymentGateway   *string                `json:"payment_gateway"`
	PaymentReference *string                `json:"payment_reference"`
	CurrencyCode     *string                `json:"currency_code"`
	ExchangeRate     *string                `json:"exchange_rate"`
	BankAccountID    *int32                 `json:"bank_account_id"`
	Notes            *string                `json:"notes"`
	ReceivedByUserID *int32                 `json:"received_by_user_id"`
	Metadata         map[string]interface{} `json:"metadata"`
}

type InvoiceFilter struct {
	OrganizationID int32   `json:"organization_id"`
	StoreID        *int32  `json:"store_id"`
	CustomerID     *int32  `json:"customer_id"`
	Status         *string `json:"status"`
	FromDate       *string `json:"from_date"`
	ToDate         *string `json:"to_date"`
	Page           int32   `json:"page"`
	Limit          int32   `json:"limit"`
}

type InvoiceUseCase struct {
	repo *repository.Queries
}

func NewInvoiceUseCase() *InvoiceUseCase {
	return &InvoiceUseCase{}
}

func (uc *InvoiceUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *InvoiceUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

func calcInvoicePagination(page, limit int32) (int32, int32, int32) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit
	return page, limit, offset
}

func calcInvoiceTotalPages(totalCount int64, limit int32) int32 {
	if limit <= 0 || totalCount <= 0 {
		return 0
	}
	return int32((totalCount + int64(limit) - 1) / int64(limit))
}

func invoiceToOutput(inv repository.Invoice, lines []repository.ListInvoiceLinesRow, payments []repository.ListInvoicePaymentsRow) InvoiceOutput {
	return InvoiceOutput{
		ID:                 inv.ID,
		InvoiceNumber:      inv.InvoiceNumber,
		OrganizationID:     inv.OrganizationID,
		StoreID:            inv.StoreID,
		CustomerID:         inv.CustomerID,
		CustomerName:       inv.CustomerName,
		CustomerEmail:      inv.CustomerEmail,
		CustomerPhone:      inv.CustomerPhone,
		CustomerTaxID:      inv.CustomerTaxID,
		InvoiceType:        inv.InvoiceType,
		InvoiceStatus:      inv.InvoiceStatus,
		SalesOrderID:       inv.SalesOrderID,
		RelatedInvoiceID:   inv.RelatedInvoiceID,
		InvoiceDate:        inv.InvoiceDate,
		DueDate:            inv.DueDate,
		SentDate:           inv.SentDate,
		PaidDate:           inv.PaidDate,
		Subtotal:           inv.Subtotal,
		DiscountAmount:     inv.DiscountAmount,
		TaxAmount:          inv.TaxAmount,
		ShippingAmount:     inv.ShippingAmount,
		AdjustmentAmount:   inv.AdjustmentAmount,
		TotalAmount:        inv.TotalAmount,
		PaidAmount:         inv.PaidAmount,
		CreditApplied:      inv.CreditApplied,
		BalanceDue:         inv.BalanceDue,
		PaymentTerms:       inv.PaymentTerms,
		CurrencyCode:       inv.CurrencyCode,
		ExchangeRate:       inv.ExchangeRate,
		BillingAddress:     utils.BytesToJSONRawMessage(inv.BillingAddress),
		ShippingAddress:    utils.BytesToJSONRawMessage(inv.ShippingAddress),
		IsRecurring:        inv.IsRecurring,
		RecurrencePattern:  inv.RecurrencePattern,
		NextInvoiceDate:    inv.NextInvoiceDate,
		PdfUrl:             inv.PdfUrl,
		DocumentHash:       inv.DocumentHash,
		ReminderSentCount:  inv.ReminderSentCount,
		LastReminderSentAt: inv.LastReminderSentAt,
		Notes:              inv.Notes,
		InternalNotes:      inv.InternalNotes,
		ReferenceNumber:    inv.ReferenceNumber,
		CreatedByUserID:    inv.CreatedByUserID,
		Tags:               inv.Tags,
		Metadata:           utils.BytesToJSONRawMessage(inv.Metadata),
		CreatedAt:          inv.CreatedAt,
		UpdatedAt:          inv.UpdatedAt,
		Lines:              lines,
		Payments:           payments,
	}
}

// CreateInvoice creates a new invoice with optional item lines.
func (uc *InvoiceUseCase) CreateInvoice(ctx context.Context, input CreateInvoiceInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if strings.TrimSpace(input.InvoiceNumber) == "" {
		return utils.NewResponse(utils.CodeBadReq, "invoice_number is required", nil)
	}
	if input.OrganizationID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}
	if input.CustomerID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "customer_id is required", nil)
	}
	if strings.TrimSpace(input.CustomerName) == "" {
		return utils.NewResponse(utils.CodeBadReq, "customer_name is required", nil)
	}

	invDateParsed, err := time.Parse("2006-01-02", input.InvoiceDate)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid invoice_date format (YYYY-MM-DD)", nil)
	}
	dueDateParsed, err := time.Parse("2006-01-02", input.DueDate)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid due_date format (YYYY-MM-DD)", nil)
	}

	var soUUID pgtype.UUID
	if input.SalesOrderID != nil && *input.SalesOrderID != "" {
		if u, err := uuid.Parse(*input.SalesOrderID); err == nil {
			soUUID = pgtype.UUID{Bytes: u, Valid: true}
		}
	}
	var relInvUUID pgtype.UUID
	if input.RelatedInvoiceID != nil && *input.RelatedInvoiceID != "" {
		if u, err := uuid.Parse(*input.RelatedInvoiceID); err == nil {
			relInvUUID = pgtype.UUID{Bytes: u, Valid: true}
		}
	}

	var nextInvDate pgtype.Date
	if input.NextInvoiceDate != nil && *input.NextInvoiceDate != "" {
		if nid, err := time.Parse("2006-01-02", *input.NextInvoiceDate); err == nil {
			nextInvDate = pgtype.Date{Time: nid, Valid: true}
		}
	}

	invType := repository.InvoiceTypeStandard
	if input.InvoiceType != "" {
		invType = repository.InvoiceType(input.InvoiceType)
	}

	invStatus := repository.InvoiceStatusDraft
	if input.InvoiceStatus != "" {
		invStatus = repository.InvoiceStatus(input.InvoiceStatus)
	}

	billingBytes, _ := json.Marshal(input.BillingAddress)
	shippingBytes, _ := json.Marshal(input.ShippingAddress)
	metaBytes, _ := json.Marshal(input.Metadata)

	var exRate pgtype.Numeric
	if input.ExchangeRate != nil && *input.ExchangeRate != "" {
		_ = exRate.Scan(*input.ExchangeRate)
	} else {
		_ = exRate.Scan("1.000000")
	}

	created, err := uc.repo.CreateInvoice(ctx, repository.CreateInvoiceParams{
		InvoiceNumber:     strings.TrimSpace(input.InvoiceNumber),
		OrganizationID:    input.OrganizationID,
		StoreID:           pgInt4(input.StoreID),
		CustomerID:        input.CustomerID,
		CustomerName:      strings.TrimSpace(input.CustomerName),
		CustomerEmail:     pgText(input.CustomerEmail),
		CustomerPhone:     pgText(input.CustomerPhone),
		CustomerTaxID:     pgText(input.CustomerTaxID),
		InvoiceType:       invType,
		InvoiceStatus:     invStatus,
		SalesOrderID:      soUUID,
		RelatedInvoiceID:  relInvUUID,
		InvoiceDate:       pgtype.Date{Time: invDateParsed, Valid: true},
		DueDate:           pgtype.Date{Time: dueDateParsed, Valid: true},
		PaymentTerms:      pgText(input.PaymentTerms),
		CurrencyCode:      pgText(input.CurrencyCode),
		ExchangeRate:      exRate,
		BillingAddress:    billingBytes,
		ShippingAddress:   shippingBytes,
		IsRecurring:       pgtype.Bool{Bool: input.IsRecurring != nil && *input.IsRecurring, Valid: true},
		RecurrencePattern: pgText(input.RecurrencePattern),
		NextInvoiceDate:   nextInvDate,
		Notes:             pgText(input.Notes),
		InternalNotes:     pgText(input.InternalNotes),
		ReferenceNumber:   pgText(input.ReferenceNumber),
		CreatedByUserID:   pgInt4(input.CreatedByUserID),
		Tags:              input.Tags,
		Metadata:          metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to create invoice: %v", err), nil)
	}

	// Insert invoice lines if provided
	for _, l := range input.Lines {
		var qNum, upNum, ltNum, discNum, taxNum, trNum pgtype.Numeric
		_ = qNum.Scan(l.Quantity)
		_ = upNum.Scan(l.UnitPrice)
		_ = ltNum.Scan(l.LineTotal)
		if l.DiscountAmount != nil {
			_ = discNum.Scan(*l.DiscountAmount)
		}
		if l.TaxAmount != nil {
			_ = taxNum.Scan(*l.TaxAmount)
		}
		if l.TaxRate != nil {
			_ = trNum.Scan(*l.TaxRate)
		}
		var oLineID pgtype.UUID
		if l.OrderLineID != nil && *l.OrderLineID != "" {
			if u, err := uuid.Parse(*l.OrderLineID); err == nil {
				oLineID = pgtype.UUID{Bytes: u, Valid: true}
			}
		}
		lineMeta, _ := json.Marshal(l.Metadata)

		_, _ = uc.repo.CreateInvoiceLine(ctx, repository.CreateInvoiceLineParams{
			InvoiceID:        created.ID,
			OrganizationID:   input.OrganizationID,
			LineNumber:       l.LineNumber,
			Description:      l.Description,
			ItemType:         pgText(l.ItemType),
			ProductID:        pgInt4(l.ProductID),
			ProductVariantID: pgInt4(l.ProductVariantID),
			ProductSku:       pgText(l.ProductSku),
			OrderLineID:      oLineID,
			Quantity:         qNum,
			UnitPrice:        upNum,
			DiscountAmount:   discNum,
			TaxAmount:        taxNum,
			LineTotal:        ltNum,
			TaxCategoryID:    pgInt4(l.TaxCategoryID),
			TaxRate:          trNum,
			UomID:            pgInt4(l.UomID),
			Metadata:         lineMeta,
		})
	}

	lines, _ := uc.repo.ListInvoiceLines(ctx, created.ID)
	return utils.NewResponse(utils.CodeCreated, "invoice created successfully", invoiceToOutput(created, lines, nil))
}

// CreateInvoiceFromOrder generates an invoice directly from a sales order.
func (uc *InvoiceUseCase) CreateInvoiceFromOrder(ctx context.Context, orderIDStr, invoiceNumber string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	orderUUID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid order_id UUID", nil)
	}
	if strings.TrimSpace(invoiceNumber) == "" {
		return utils.NewResponse(utils.CodeBadReq, "invoice_number is required", nil)
	}

	created, err := uc.repo.CreateInvoiceFromOrder(ctx, repository.CreateInvoiceFromOrderParams{
		ID:            orderUUID,
		InvoiceNumber: strings.TrimSpace(invoiceNumber),
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to generate invoice from order: %v", err), nil)
	}

	lines, _ := uc.repo.ListInvoiceLines(ctx, created.ID)
	return utils.NewResponse(utils.CodeCreated, "invoice generated from order", invoiceToOutput(created, lines, nil))
}

// GetInvoice retrieves an invoice with its line items and payment records.
func (uc *InvoiceUseCase) GetInvoice(ctx context.Context, idStr string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	invUUID, err := uuid.Parse(idStr)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid invoice ID UUID", nil)
	}

	inv, err := uc.repo.GetInvoice(ctx, invUUID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "invoice not found", nil)
	}

	lines, _ := uc.repo.ListInvoiceLines(ctx, inv.ID)
	payments, _ := uc.repo.ListInvoicePayments(ctx, inv.ID)

	return utils.NewResponse(utils.CodeOK, "invoice fetched successfully", invoiceToOutput(inv, lines, payments))
}

// GetInvoiceByNumber retrieves an invoice by its invoice number.
func (uc *InvoiceUseCase) GetInvoiceByNumber(ctx context.Context, invoiceNumber string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if strings.TrimSpace(invoiceNumber) == "" {
		return utils.NewResponse(utils.CodeBadReq, "invoice_number is required", nil)
	}

	inv, err := uc.repo.GetInvoiceByNumber(ctx, strings.TrimSpace(invoiceNumber))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "invoice not found", nil)
	}

	lines, _ := uc.repo.ListInvoiceLines(ctx, inv.ID)
	payments, _ := uc.repo.ListInvoicePayments(ctx, inv.ID)

	return utils.NewResponse(utils.CodeOK, "invoice fetched successfully", invoiceToOutput(inv, lines, payments))
}

// ListInvoices returns a paginated list of invoices matching filters.
func (uc *InvoiceUseCase) ListInvoices(ctx context.Context, filter InvoiceFilter) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if filter.OrganizationID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}

	p, l, offset := calcInvoicePagination(filter.Page, filter.Limit)

	var sID pgtype.Int4
	if filter.StoreID != nil {
		sID = pgtype.Int4{Int32: *filter.StoreID, Valid: true}
	}

	var invStatus repository.NullInvoiceStatus
	if filter.Status != nil && *filter.Status != "" {
		invStatus = repository.NullInvoiceStatus{
			InvoiceStatus: repository.InvoiceStatus(*filter.Status),
			Valid:         true,
		}
	}

	var fromDate, toDate pgtype.Date
	if filter.FromDate != nil && *filter.FromDate != "" {
		if fd, err := time.Parse("2006-01-02", *filter.FromDate); err == nil {
			fromDate = pgtype.Date{Time: fd, Valid: true}
		}
	}
	if filter.ToDate != nil && *filter.ToDate != "" {
		if td, err := time.Parse("2006-01-02", *filter.ToDate); err == nil {
			toDate = pgtype.Date{Time: td, Valid: true}
		}
	}

	totalCount, err := uc.repo.CountInvoices(ctx, repository.CountInvoicesParams{
		OrganizationID: filter.OrganizationID,
		StoreID:        sID,
		InvoiceStatus:  invStatus,
		FromDate:       fromDate,
		ToDate:         toDate,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to count invoices: %v", err), nil)
	}

	rows, err := uc.repo.ListInvoices(ctx, repository.ListInvoicesParams{
		OrganizationID: filter.OrganizationID,
		StoreID:        sID,
		InvoiceStatus:  invStatus,
		FromDate:       fromDate,
		ToDate:         toDate,
		Offset:         offset,
		Limit:          l,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to list invoices: %v", err), nil)
	}

	outputs := make([]InvoiceOutput, len(rows))
	for i := range rows {
		outputs[i] = invoiceToOutput(rows[i], nil, nil)
	}

	return utils.NewResponse(utils.CodeOK, "invoices fetched successfully", InvoiceListResponse{
		Data:       outputs,
		TotalCount: totalCount,
		Page:       p,
		Limit:      l,
		TotalPages: calcInvoiceTotalPages(totalCount, l),
	})
}

// ListCustomerInvoices returns paginated invoices for a specific customer.
func (uc *InvoiceUseCase) ListCustomerInvoices(ctx context.Context, customerID, orgID, page, limit int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if customerID <= 0 || orgID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "customer_id and organization_id are required", nil)
	}

	p, l, offset := calcInvoicePagination(page, limit)

	totalCount, err := uc.repo.CountCustomerInvoices(ctx, repository.CountCustomerInvoicesParams{
		CustomerID:     customerID,
		OrganizationID: orgID,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to count customer invoices: %v", err), nil)
	}

	rows, err := uc.repo.ListCustomerInvoices(ctx, repository.ListCustomerInvoicesParams{
		CustomerID:     customerID,
		OrganizationID: orgID,
		Limit:          l,
		Offset:         offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to list customer invoices: %v", err), nil)
	}

	outputs := make([]InvoiceOutput, len(rows))
	for i := range rows {
		outputs[i] = invoiceToOutput(rows[i], nil, nil)
	}

	return utils.NewResponse(utils.CodeOK, "customer invoices fetched successfully", InvoiceListResponse{
		Data:       outputs,
		TotalCount: totalCount,
		Page:       p,
		Limit:      l,
		TotalPages: calcInvoiceTotalPages(totalCount, l),
	})
}

// ListOverdueInvoices returns paginated overdue invoices for a store.
func (uc *InvoiceUseCase) ListOverdueInvoices(ctx context.Context, storeID, page, limit int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if storeID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "store_id is required", nil)
	}

	p, l, offset := calcInvoicePagination(page, limit)
	storeIDParam := pgtype.Int4{Int32: storeID, Valid: true}

	totalCount, err := uc.repo.CountOverdueInvoices(ctx, storeIDParam)
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to count overdue invoices: %v", err), nil)
	}

	rows, err := uc.repo.ListOverdueInvoices(ctx, repository.ListOverdueInvoicesParams{
		StoreID: storeIDParam,
		Limit:   l,
		Offset:  offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to list overdue invoices: %v", err), nil)
	}

	type OverdueOutput struct {
		InvoiceOutput
		CustomerNameFull  pgtype.Text `json:"customer_name_full"`
		CustomerEmailFull pgtype.Text `json:"customer_email_full"`
	}

	outputs := make([]OverdueOutput, len(rows))
	for i, r := range rows {
		outputs[i] = OverdueOutput{
			InvoiceOutput: InvoiceOutput{
				ID:                 r.ID,
				InvoiceNumber:      r.InvoiceNumber,
				OrganizationID:     r.OrganizationID,
				StoreID:            r.StoreID,
				CustomerID:         r.CustomerID,
				CustomerName:       r.CustomerName,
				CustomerEmail:      r.CustomerEmail,
				CustomerPhone:      r.CustomerPhone,
				CustomerTaxID:      r.CustomerTaxID,
				InvoiceType:        r.InvoiceType,
				InvoiceStatus:      r.InvoiceStatus,
				SalesOrderID:       r.SalesOrderID,
				RelatedInvoiceID:   r.RelatedInvoiceID,
				InvoiceDate:        r.InvoiceDate,
				DueDate:            r.DueDate,
				SentDate:           r.SentDate,
				PaidDate:           r.PaidDate,
				Subtotal:           r.Subtotal,
				DiscountAmount:     r.DiscountAmount,
				TaxAmount:          r.TaxAmount,
				ShippingAmount:     r.ShippingAmount,
				AdjustmentAmount:   r.AdjustmentAmount,
				TotalAmount:        r.TotalAmount,
				PaidAmount:         r.PaidAmount,
				CreditApplied:      r.CreditApplied,
				BalanceDue:         r.BalanceDue,
				PaymentTerms:       r.PaymentTerms,
				CurrencyCode:       r.CurrencyCode,
				ExchangeRate:       r.ExchangeRate,
				BillingAddress:     utils.BytesToJSONRawMessage(r.BillingAddress),
				ShippingAddress:    utils.BytesToJSONRawMessage(r.ShippingAddress),
				IsRecurring:        r.IsRecurring,
				RecurrencePattern:  r.RecurrencePattern,
				NextInvoiceDate:    r.NextInvoiceDate,
				PdfUrl:             r.PdfUrl,
				DocumentHash:       r.DocumentHash,
				ReminderSentCount:  r.ReminderSentCount,
				LastReminderSentAt: r.LastReminderSentAt,
				Notes:              r.Notes,
				InternalNotes:      r.InternalNotes,
				ReferenceNumber:    r.ReferenceNumber,
				CreatedByUserID:    r.CreatedByUserID,
				Tags:               r.Tags,
				Metadata:           utils.BytesToJSONRawMessage(r.Metadata),
				CreatedAt:          r.CreatedAt,
				UpdatedAt:          r.UpdatedAt,
			},
			CustomerNameFull:  r.CustomerNameFull,
			CustomerEmailFull: r.CustomerEmailFull,
		}
	}

	return utils.NewResponse(utils.CodeOK, "overdue invoices fetched successfully", map[string]interface{}{
		"data":        outputs,
		"total_count": totalCount,
		"page":        p,
		"limit":       l,
		"total_pages": calcInvoiceTotalPages(totalCount, l),
	})
}

// UpdateInvoice updates core invoice fields.
func (uc *InvoiceUseCase) UpdateInvoice(ctx context.Context, idStr string, input UpdateInvoiceInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	invUUID, err := uuid.Parse(idStr)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid invoice ID UUID", nil)
	}

	dueDateParsed, err := time.Parse("2006-01-02", input.DueDate)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid due_date format (YYYY-MM-DD)", nil)
	}

	billingBytes, _ := json.Marshal(input.BillingAddress)
	shippingBytes, _ := json.Marshal(input.ShippingAddress)
	metaBytes, _ := json.Marshal(input.Metadata)

	updated, err := uc.repo.UpdateInvoice(ctx, repository.UpdateInvoiceParams{
		ID:              invUUID,
		CustomerName:    strings.TrimSpace(input.CustomerName),
		CustomerEmail:   pgText(input.CustomerEmail),
		CustomerPhone:   pgText(input.CustomerPhone),
		DueDate:         pgtype.Date{Time: dueDateParsed, Valid: true},
		PaymentTerms:    pgText(input.PaymentTerms),
		BillingAddress:  billingBytes,
		ShippingAddress: shippingBytes,
		Notes:           pgText(input.Notes),
		InternalNotes:   pgText(input.InternalNotes),
		ReferenceNumber: pgText(input.ReferenceNumber),
		Tags:            input.Tags,
		Metadata:        metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to update invoice: %v", err), nil)
	}

	lines, _ := uc.repo.ListInvoiceLines(ctx, updated.ID)
	return utils.NewResponse(utils.CodeOK, "invoice updated successfully", invoiceToOutput(updated, lines, nil))
}

// UpdateInvoiceStatus updates status and records audit history.
func (uc *InvoiceUseCase) UpdateInvoiceStatus(ctx context.Context, idStr string, status string, changedByUserID *int32, reason, notes *string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	invUUID, err := uuid.Parse(idStr)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid invoice ID UUID", nil)
	}
	if strings.TrimSpace(status) == "" {
		return utils.NewResponse(utils.CodeBadReq, "status is required", nil)
	}

	existing, err := uc.repo.GetInvoice(ctx, invUUID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "invoice not found", nil)
	}

	newStatus := repository.InvoiceStatus(status)
	updated, err := uc.repo.UpdateInvoiceStatus(ctx, repository.UpdateInvoiceStatusParams{
		ID:            invUUID,
		InvoiceStatus: newStatus,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to update invoice status: %v", err), nil)
	}

	// Record audit history
	_, _ = uc.repo.CreateInvoiceStatusHistory(ctx, repository.CreateInvoiceStatusHistoryParams{
		InvoiceID:         invUUID,
		OrganizationID:    existing.OrganizationID,
		FromStatus:        repository.NullInvoiceStatus{InvoiceStatus: existing.InvoiceStatus, Valid: true},
		ToStatus:          newStatus,
		Reason:            pgText(reason),
		Notes:             pgText(notes),
		ChangedByUserID:   pgInt4(changedByUserID),
	})

	return utils.NewResponse(utils.CodeOK, "invoice status updated", invoiceToOutput(updated, nil, nil))
}

// DeleteInvoice removes an invoice.
func (uc *InvoiceUseCase) DeleteInvoice(ctx context.Context, idStr string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	invUUID, err := uuid.Parse(idStr)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid invoice ID UUID", nil)
	}

	if err := uc.repo.DeleteInvoice(ctx, invUUID); err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to delete invoice: %v", err), nil)
	}

	return utils.NewResponse(utils.CodeOK, "invoice deleted successfully", nil)
}

// CreateInvoicePayment records a payment against an invoice.
func (uc *InvoiceUseCase) CreateInvoicePayment(ctx context.Context, idStr string, input RecordInvoicePaymentInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	invUUID, err := uuid.Parse(idStr)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid invoice ID UUID", nil)
	}
	if strings.TrimSpace(input.PaymentNumber) == "" {
		return utils.NewResponse(utils.CodeBadReq, "payment_number is required", nil)
	}
	if strings.TrimSpace(input.PaymentAmount) == "" {
		return utils.NewResponse(utils.CodeBadReq, "payment_amount is required", nil)
	}
	if strings.TrimSpace(input.PaymentMethod) == "" {
		return utils.NewResponse(utils.CodeBadReq, "payment_method is required", nil)
	}

	inv, err := uc.repo.GetInvoice(ctx, invUUID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "invoice not found", nil)
	}

	pDate := time.Now()
	if strings.TrimSpace(input.PaymentDate) != "" {
		if d, err := time.Parse("2006-01-02", input.PaymentDate); err == nil {
			pDate = d
		}
	}

	var pAmount, exRate pgtype.Numeric
	_ = pAmount.Scan(input.PaymentAmount)
	if input.ExchangeRate != nil && *input.ExchangeRate != "" {
		_ = exRate.Scan(*input.ExchangeRate)
	} else {
		_ = exRate.Scan("1.000000")
	}

	metaBytes, _ := json.Marshal(input.Metadata)

	payment, err := uc.repo.CreateInvoicePayment(ctx, repository.CreateInvoicePaymentParams{
		InvoiceID:        invUUID,
		OrganizationID:   inv.OrganizationID,
		PaymentNumber:    strings.TrimSpace(input.PaymentNumber),
		PaymentDate:      pgtype.Date{Time: pDate, Valid: true},
		PaymentAmount:    pAmount,
		PaymentMethod:    strings.TrimSpace(input.PaymentMethod),
		PaymentGateway:   pgText(input.PaymentGateway),
		PaymentReference: pgText(input.PaymentReference),
		CurrencyCode:     pgText(input.CurrencyCode),
		ExchangeRate:     exRate,
		BankAccountID:    pgInt4(input.BankAccountID),
		Notes:            pgText(input.Notes),
		ReceivedByUserID: pgInt4(input.ReceivedByUserID),
		Metadata:         metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to record payment: %v", err), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "payment recorded successfully", payment)
}

// ListInvoicePayments lists all payments for an invoice.
func (uc *InvoiceUseCase) ListInvoicePayments(ctx context.Context, idStr string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	invUUID, err := uuid.Parse(idStr)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid invoice ID UUID", nil)
	}

	payments, err := uc.repo.ListInvoicePayments(ctx, invUUID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to list payments: %v", err), nil)
	}

	return utils.NewResponse(utils.CodeOK, "invoice payments fetched", payments)
}

// GetInvoiceStats returns overview statistics for an organization's invoices.
func (uc *InvoiceUseCase) GetInvoiceStats(ctx context.Context, orgID int32, storeID *int32, fromDateStr, toDateStr *string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if orgID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}

	var sID pgtype.Int4
	if storeID != nil {
		sID = pgtype.Int4{Int32: *storeID, Valid: true}
	}
	var fromDate, toDate pgtype.Date
	if fromDateStr != nil && *fromDateStr != "" {
		if fd, err := time.Parse("2006-01-02", *fromDateStr); err == nil {
			fromDate = pgtype.Date{Time: fd, Valid: true}
		}
	}
	if toDateStr != nil && *toDateStr != "" {
		if td, err := time.Parse("2006-01-02", *toDateStr); err == nil {
			toDate = pgtype.Date{Time: td, Valid: true}
		}
	}

	stats, err := uc.repo.GetInvoiceStats(ctx, repository.GetInvoiceStatsParams{
		OrganizationID: orgID,
		StoreID:        sID,
		FromDate:       fromDate,
		ToDate:         toDate,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to get invoice stats: %v", err), nil)
	}

	return utils.NewResponse(utils.CodeOK, "invoice stats fetched", stats)
}

// ListInvoiceLines lists all item lines for an invoice.
func (uc *InvoiceUseCase) ListInvoiceLines(ctx context.Context, idStr string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	invUUID, err := uuid.Parse(idStr)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid invoice ID UUID", nil)
	}

	lines, err := uc.repo.ListInvoiceLines(ctx, invUUID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to list lines: %v", err), nil)
	}

	return utils.NewResponse(utils.CodeOK, "invoice lines fetched", lines)
}

// CreateInvoiceLine adds a line item to an existing invoice.
func (uc *InvoiceUseCase) CreateInvoiceLine(ctx context.Context, idStr string, input CreateInvoiceLineInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	invUUID, err := uuid.Parse(idStr)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid invoice ID UUID", nil)
	}

	inv, err := uc.repo.GetInvoice(ctx, invUUID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "invoice not found", nil)
	}

	var qNum, upNum, ltNum, discNum, taxNum, trNum pgtype.Numeric
	_ = qNum.Scan(input.Quantity)
	_ = upNum.Scan(input.UnitPrice)
	_ = ltNum.Scan(input.LineTotal)
	if input.DiscountAmount != nil {
		_ = discNum.Scan(*input.DiscountAmount)
	}
	if input.TaxAmount != nil {
		_ = taxNum.Scan(*input.TaxAmount)
	}
	if input.TaxRate != nil {
		_ = trNum.Scan(*input.TaxRate)
	}
	var oLineID pgtype.UUID
	if input.OrderLineID != nil && *input.OrderLineID != "" {
		if u, err := uuid.Parse(*input.OrderLineID); err == nil {
			oLineID = pgtype.UUID{Bytes: u, Valid: true}
		}
	}
	lineMeta, _ := json.Marshal(input.Metadata)

	line, err := uc.repo.CreateInvoiceLine(ctx, repository.CreateInvoiceLineParams{
		InvoiceID:        invUUID,
		OrganizationID:   inv.OrganizationID,
		LineNumber:       input.LineNumber,
		Description:      input.Description,
		ItemType:         pgText(input.ItemType),
		ProductID:        pgInt4(input.ProductID),
		ProductVariantID: pgInt4(input.ProductVariantID),
		ProductSku:       pgText(input.ProductSku),
		OrderLineID:      oLineID,
		Quantity:         qNum,
		UnitPrice:        upNum,
		DiscountAmount:   discNum,
		TaxAmount:        taxNum,
		LineTotal:        ltNum,
		TaxCategoryID:    pgInt4(input.TaxCategoryID),
		TaxRate:          trNum,
		UomID:            pgInt4(input.UomID),
		Metadata:         lineMeta,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to add invoice line: %v", err), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "invoice line added", line)
}

// DeleteInvoiceLine deletes a line item from an invoice.
func (uc *InvoiceUseCase) DeleteInvoiceLine(ctx context.Context, lineIDStr string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	lineUUID, err := uuid.Parse(lineIDStr)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid line ID UUID", nil)
	}

	if err := uc.repo.DeleteInvoiceLine(ctx, lineUUID); err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to delete invoice line: %v", err), nil)
	}

	return utils.NewResponse(utils.CodeOK, "invoice line deleted", nil)
}
