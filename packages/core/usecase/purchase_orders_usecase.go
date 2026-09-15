package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type PurchaseOrderLineInput struct {
	ProductID        int32                  `json:"product_id"`
	ProductVariantID *int32                 `json:"product_variant_id,omitempty"`
	Quantity         float64                `json:"quantity"`
	UomID            *int32                 `json:"uom_id,omitempty"`
	UnitPrice        float64                `json:"unit_price"`
	DiscountAmount   *float64               `json:"discount_amount,omitempty"`
	TaxAmount        *float64               `json:"tax_amount,omitempty"`
	LineNumber       *int32                 `json:"line_number,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type CreatePurchaseOrderInput struct {
	OrganizationID       int32                    `json:"organization_id"`
	PoNumber             *string                  `json:"po_number,omitempty"`
	SupplierID           int32                    `json:"supplier_id"` // maps to partners_id
	StoreID              int32                    `json:"store_id"`
	PoDate               *string                  `json:"po_date,omitempty"`
	ExpectedDeliveryDate *string                  `json:"expected_delivery_date,omitempty"`
	Status               *string                  `json:"status,omitempty"` // default 'draft'
	PriceListID          *int32                   `json:"price_list_id,omitempty"`
	CreatedBy            *int32                   `json:"created_by,omitempty"`
	Metadata             map[string]interface{}   `json:"metadata,omitempty"`
	Items                []PurchaseOrderLineInput `json:"items"`
}

type UpdatePurchaseOrderInput struct {
	SupplierID           *int32                   `json:"supplier_id,omitempty"`
	StoreID              *int32                   `json:"store_id,omitempty"`
	PoDate               *string                  `json:"po_date,omitempty"`
	ExpectedDeliveryDate *string                  `json:"expected_delivery_date,omitempty"`
	PriceListID          *int32                   `json:"price_list_id,omitempty"`
	Metadata             map[string]interface{}   `json:"metadata,omitempty"`
	Items                []PurchaseOrderLineInput `json:"items,omitempty"`
}

type UpdatePurchaseOrderStatusInput struct {
	Status     string `json:"status"` // 'draft', 'submitted', 'approved', 'cancelled', 'closed'
	ApprovedBy *int32 `json:"approved_by,omitempty"`
}

type PurchaseOrderFilter struct {
	OrganizationID int32      `json:"organization_id"`
	StoreID        *int32     `json:"store_id,omitempty"`
	PartnerID      *int32     `json:"partner_id,omitempty"`
	Status         *string    `json:"status,omitempty"`
	FromDate       *time.Time `json:"from_date,omitempty"`
	ToDate         *time.Time `json:"to_date,omitempty"`
	Search         *string    `json:"search,omitempty"`
	Page           int32      `json:"page"`
	Limit          int32      `json:"limit"`
}

type PurchaseOrderLineOutput struct {
	ID               int32            `json:"id"`
	PurchaseOrderID  int32            `json:"purchase_order_id"`
	ProductID        int32            `json:"product_id"`
	ProductName      string           `json:"product_name"`
	ProductSKU       string           `json:"product_sku"`
	ProductVariantID pgtype.Int4      `json:"product_variant_id"`
	VariantName      pgtype.Text      `json:"variant_name"`
	VariantSKU       pgtype.Text      `json:"variant_sku"`
	Quantity         pgtype.Numeric   `json:"quantity"`
	UomID            pgtype.Int4      `json:"uom_id"`
	UomName          pgtype.Text      `json:"uom_name"`
	UnitPrice        pgtype.Numeric   `json:"unit_price"`
	DiscountAmount   pgtype.Numeric   `json:"discount_amount"`
	TaxAmount        pgtype.Numeric   `json:"tax_amount"`
	Subtotal         pgtype.Numeric   `json:"subtotal"`
	LineTotal        pgtype.Numeric   `json:"line_total"`
	ReceivedQuantity pgtype.Numeric   `json:"received_quantity"`
	LineNumber       pgtype.Int4      `json:"line_number"`
	Barcode          string           `json:"barcode,omitempty"`
	Metadata         json.RawMessage  `json:"metadata"`
	CreatedAt        pgtype.Timestamp `json:"created_at"`
}

type PurchaseOrderOutput struct {
	ID                   int32                     `json:"id"`
	OrganizationID       int32                     `json:"organization_id"`
	PoNumber             string                    `json:"po_number"`
	SupplierID           int32                     `json:"supplier_id"`
	SupplierName         pgtype.Text               `json:"supplier_name"`
	SupplierCode         pgtype.Text               `json:"supplier_code"`
	StoreID              int32                     `json:"store_id"`
	StoreName            pgtype.Text               `json:"store_name"`
	PoDate               pgtype.Date               `json:"po_date"`
	ExpectedDeliveryDate pgtype.Date               `json:"expected_delivery_date"`
	Status               string                    `json:"status"`
	Subtotal             pgtype.Numeric            `json:"subtotal"`
	DiscountAmount       pgtype.Numeric            `json:"discount_amount"`
	TaxAmount            pgtype.Numeric            `json:"tax_amount"`
	TotalAmount          pgtype.Numeric            `json:"total_amount"`
	PriceListID          pgtype.Int4               `json:"price_list_id"`
	CreatedBy            pgtype.Int4               `json:"created_by"`
	CreatedByName        pgtype.Text               `json:"created_by_name"`
	ApprovedBy           pgtype.Int4               `json:"approved_by"`
	ApprovedByName       pgtype.Text               `json:"approved_by_name"`
	Metadata             json.RawMessage           `json:"metadata"`
	CreatedAt            pgtype.Timestamp          `json:"created_at"`
	UpdatedAt            pgtype.Timestamp          `json:"updated_at"`
	Items                []PurchaseOrderLineOutput `json:"items,omitempty"`
}

type PurchaseOrderSummaryOutput struct {
	ID                   int32            `json:"id"`
	OrganizationID       int32            `json:"organization_id"`
	PoNumber             string           `json:"po_number"`
	SupplierID           int32            `json:"supplier_id"`
	SupplierName         pgtype.Text      `json:"supplier_name"`
	SupplierCode         pgtype.Text      `json:"supplier_code"`
	StoreID              int32            `json:"store_id"`
	StoreName            pgtype.Text      `json:"store_name"`
	PoDate               pgtype.Date      `json:"po_date"`
	ExpectedDeliveryDate pgtype.Date      `json:"expected_delivery_date"`
	Status               string           `json:"status"`
	Subtotal             pgtype.Numeric   `json:"subtotal"`
	DiscountAmount       pgtype.Numeric   `json:"discount_amount"`
	TaxAmount            pgtype.Numeric   `json:"tax_amount"`
	TotalAmount          pgtype.Numeric   `json:"total_amount"`
	PriceListID          pgtype.Int4      `json:"price_list_id"`
	ItemCount            int64            `json:"item_count"`
	TotalQuantity        pgtype.Numeric   `json:"total_quantity"`
	CreatedAt            pgtype.Timestamp `json:"created_at"`
	UpdatedAt            pgtype.Timestamp `json:"updated_at"`
}

type PurchaseOrderListResponse struct {
	Data       []PurchaseOrderSummaryOutput `json:"data"`
	TotalCount int64                        `json:"total_count"`
	Page       int32                        `json:"page"`
	Limit      int32                        `json:"limit"`
	TotalPages int32                        `json:"total_pages"`
}

type PurchaseOrdersUseCase struct {
	repo *repository.Queries
}

func NewPurchaseOrdersUseCase() *PurchaseOrdersUseCase {
	return &PurchaseOrdersUseCase{}
}

func (uc *PurchaseOrdersUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *PurchaseOrdersUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

func generatePONumber() string {
	return fmt.Sprintf("PO-%s-%04d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)
}

func parseDateString(dateStr *string) *time.Time {
	if dateStr == nil || *dateStr == "" {
		return nil
	}
	formats := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, *dateStr); err == nil {
			return &t
		}
	}
	return nil
}

// CreatePurchaseOrder creates a new purchase order with lines.
func (uc *PurchaseOrdersUseCase) CreatePurchaseOrder(ctx context.Context, input CreatePurchaseOrderInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	if input.OrganizationID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}
	if input.SupplierID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "supplier_id is required", nil)
	}
	if input.StoreID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "store_id is required", nil)
	}
	if len(input.Items) == 0 {
		return utils.NewResponse(utils.CodeBadReq, "at least one purchase order item is required", nil)
	}

	poNumber := ""
	if input.PoNumber != nil && *input.PoNumber != "" {
		poNumber = *input.PoNumber
	} else {
		poNumber = generatePONumber()
	}

	poDate := time.Now()
	if parsed := parseDateString(input.PoDate); parsed != nil {
		poDate = *parsed
	}
	expectedDeliveryDate := parseDateString(input.ExpectedDeliveryDate)

	status := "draft"
	if input.Status != nil && *input.Status != "" {
		status = *input.Status
	}

	// Calculate totals from line items
	var calculatedSubtotal float64
	var calculatedDiscount float64
	var calculatedTax float64
	var calculatedTotal float64

	type computedLine struct {
		item           PurchaseOrderLineInput
		subtotal       float64
		discountAmount float64
		taxAmount      float64
		lineTotal      float64
		lineNumber     int32
	}

	var computedLines []computedLine
	for idx, item := range input.Items {
		if item.ProductID <= 0 {
			return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("item %d: product_id is required", idx+1), nil)
		}
		if item.Quantity <= 0 {
			return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("item %d: quantity must be greater than 0", idx+1), nil)
		}
		if item.UnitPrice < 0 {
			return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("item %d: unit_price cannot be negative", idx+1), nil)
		}

		subtotal := item.Quantity * item.UnitPrice
		discount := 0.0
		if item.DiscountAmount != nil && *item.DiscountAmount > 0 {
			discount = *item.DiscountAmount
		}
		tax := 0.0
		if item.TaxAmount != nil && *item.TaxAmount > 0 {
			tax = *item.TaxAmount
		}
		lineTotal := subtotal - discount + tax
		if lineTotal < 0 {
			lineTotal = 0
		}

		lineNum := int32(idx + 1)
		if item.LineNumber != nil && *item.LineNumber > 0 {
			lineNum = *item.LineNumber
		}

		computedLines = append(computedLines, computedLine{
			item:           item,
			subtotal:       subtotal,
			discountAmount: discount,
			taxAmount:      tax,
			lineTotal:      lineTotal,
			lineNumber:     lineNum,
		})

		calculatedSubtotal += subtotal
		calculatedDiscount += discount
		calculatedTax += tax
		calculatedTotal += lineTotal
	}

	metaBytes, _ := json.Marshal(input.Metadata)

	poRow, err := uc.repo.CreatePurchaseOrder(ctx, repository.CreatePurchaseOrderParams{
		OrganizationID:       input.OrganizationID,
		PoNumber:             poNumber,
		PartnersID:           input.SupplierID,
		StoreID:              input.StoreID,
		PoDate:               utils.TimeToPgDate(&poDate),
		ExpectedDeliveryDate: utils.TimeToPgDate(expectedDeliveryDate),
		Status:               pgtype.Text{String: status, Valid: true},
		Subtotal:             utils.Float64ToPgNumeric(calculatedSubtotal),
		DiscountAmount:       utils.Float64ToPgNumeric(calculatedDiscount),
		TaxAmount:            utils.Float64ToPgNumeric(calculatedTax),
		TotalAmount:          utils.Float64ToPgNumeric(calculatedTotal),
		PriceListID:          utils.Int32ToPgInt4(input.PriceListID),
		CreatedBy:            utils.Int32ToPgInt4(input.CreatedBy),
		Metadata:             metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to create purchase order: %v", err), nil)
	}

	for _, cl := range computedLines {
		itemMetaBytes, _ := json.Marshal(cl.item.Metadata)
		_, err := uc.repo.CreatePurchaseOrderLine(ctx, repository.CreatePurchaseOrderLineParams{
			PurchaseOrderID:  poRow.ID,
			ProductID:        cl.item.ProductID,
			ProductVariantID: utils.Int32ToPgInt4(cl.item.ProductVariantID),
			Quantity:         utils.Float64ToPgNumeric(cl.item.Quantity),
			UomID:            utils.Int32ToPgInt4(cl.item.UomID),
			UnitPrice:        utils.Float64ToPgNumeric(cl.item.UnitPrice),
			DiscountAmount:   utils.Float64ToPgNumeric(cl.discountAmount),
			TaxAmount:        utils.Float64ToPgNumeric(cl.taxAmount),
			Subtotal:         utils.Float64ToPgNumeric(cl.subtotal),
			LineTotal:        utils.Float64ToPgNumeric(cl.lineTotal),
			ReceivedQuantity: utils.Float64ToPgNumeric(0),
			LineNumber:       pgtype.Int4{Int32: cl.lineNumber, Valid: true},
			Metadata:         itemMetaBytes,
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to create purchase order line: %v", err), nil)
		}
	}

	return uc.GetPurchaseOrder(ctx, poRow.ID)
}

// GetPurchaseOrder retrieves full purchase order details by ID.
func (uc *PurchaseOrdersUseCase) GetPurchaseOrder(ctx context.Context, id int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	poRow, err := uc.repo.GetPurchaseOrderWithDetails(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "purchase order not found", nil)
	}

	lines, err := uc.repo.ListPurchaseOrderLines(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to load purchase order lines: %v", err), nil)
	}

	items := make([]PurchaseOrderLineOutput, len(lines))
	for i, line := range lines {
		items[i] = PurchaseOrderLineOutput{
			ID:               line.ID,
			PurchaseOrderID:  line.PurchaseOrderID,
			ProductID:        line.ProductID,
			ProductName:      line.ProductName,
			ProductSKU:       line.ProductSku,
			ProductVariantID: line.ProductVariantID,
			VariantName:      line.VariantName,
			VariantSKU:       line.VariantSku,
			Quantity:         line.Quantity,
			UomID:            line.UomID,
			UomName:          line.UomName,
			UnitPrice:        line.UnitPrice,
			DiscountAmount:   line.DiscountAmount,
			TaxAmount:        line.TaxAmount,
			Subtotal:         line.Subtotal,
			LineTotal:        line.LineTotal,
			ReceivedQuantity: line.ReceivedQuantity,
			LineNumber:       line.LineNumber,
			Barcode:          line.Barcode,
			Metadata:         line.Metadata,
			CreatedAt:        line.CreatedAt,
		}
	}

	statusStr := ""
	if poRow.Status.Valid {
		statusStr = poRow.Status.String
	}

	out := PurchaseOrderOutput{
		ID:                   poRow.ID,
		OrganizationID:       poRow.OrganizationID,
		PoNumber:             poRow.PoNumber,
		SupplierID:           poRow.PartnersID,
		SupplierName:         poRow.PartnerName,
		SupplierCode:         poRow.PartnerCode,
		StoreID:              poRow.StoreID,
		StoreName:            poRow.StoreName,
		PoDate:               poRow.PoDate,
		ExpectedDeliveryDate: poRow.ExpectedDeliveryDate,
		Status:               statusStr,
		Subtotal:             poRow.Subtotal,
		DiscountAmount:       poRow.DiscountAmount,
		TaxAmount:            poRow.TaxAmount,
		TotalAmount:          poRow.TotalAmount,
		PriceListID:          poRow.PriceListID,
		CreatedBy:            poRow.CreatedBy,
		CreatedByName:        poRow.CreatedByName,
		ApprovedBy:           poRow.ApprovedBy,
		ApprovedByName:       poRow.ApprovedByName,
		Metadata:             poRow.Metadata,
		CreatedAt:            poRow.CreatedAt,
		UpdatedAt:            poRow.UpdatedAt,
		Items:                items,
	}

	return utils.NewResponse(utils.CodeOK, "purchase order retrieved successfully", out)
}

// ListPurchaseOrders lists purchase orders with filtering and pagination.
func (uc *PurchaseOrdersUseCase) ListPurchaseOrders(ctx context.Context, filter PurchaseOrderFilter) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	if filter.OrganizationID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	var storeID pgtype.Int4
	if filter.StoreID != nil && *filter.StoreID > 0 {
		storeID = pgtype.Int4{Int32: *filter.StoreID, Valid: true}
	}
	var partnerID pgtype.Int4
	if filter.PartnerID != nil && *filter.PartnerID > 0 {
		partnerID = pgtype.Int4{Int32: *filter.PartnerID, Valid: true}
	}
	var status pgtype.Text
	if filter.Status != nil && *filter.Status != "" {
		status = pgtype.Text{String: *filter.Status, Valid: true}
	}
	var fromDate pgtype.Date
	if filter.FromDate != nil {
		fromDate = utils.TimeToPgDate(filter.FromDate)
	}
	var toDate pgtype.Date
	if filter.ToDate != nil {
		toDate = utils.TimeToPgDate(filter.ToDate)
	}
	var search pgtype.Text
	if filter.Search != nil && *filter.Search != "" {
		search = pgtype.Text{String: *filter.Search, Valid: true}
	}

	totalCount, err := uc.repo.CountPurchaseOrders(ctx, repository.CountPurchaseOrdersParams{
		OrganizationID: filter.OrganizationID,
		StoreID:        storeID,
		PartnerID:      partnerID,
		Status:         status,
		FromDate:       fromDate,
		ToDate:         toDate,
		Search:         search,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to count purchase orders: %v", err), nil)
	}

	rows, err := uc.repo.ListPurchaseOrders(ctx, repository.ListPurchaseOrdersParams{
		OrganizationID: filter.OrganizationID,
		StoreID:        storeID,
		PartnerID:      partnerID,
		Status:         status,
		FromDate:       fromDate,
		ToDate:         toDate,
		Search:         search,
		LimitCount:     limit,
		OffsetCount:    offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to list purchase orders: %v", err), nil)
	}

	data := make([]PurchaseOrderSummaryOutput, len(rows))
	for i, r := range rows {
		st := ""
		if r.Status.Valid {
			st = r.Status.String
		}
		data[i] = PurchaseOrderSummaryOutput{
			ID:                   r.ID,
			OrganizationID:       r.OrganizationID,
			PoNumber:             r.PoNumber,
			SupplierID:           r.PartnersID,
			SupplierName:         r.PartnerName,
			SupplierCode:         r.PartnerCode,
			StoreID:              r.StoreID,
			StoreName:            r.StoreName,
			PoDate:               r.PoDate,
			ExpectedDeliveryDate: r.ExpectedDeliveryDate,
			Status:               st,
			Subtotal:             r.Subtotal,
			DiscountAmount:       r.DiscountAmount,
			TaxAmount:            r.TaxAmount,
			TotalAmount:          r.TotalAmount,
			PriceListID:          r.PriceListID,
			ItemCount:            r.ItemCount,
			TotalQuantity:        r.TotalQuantity,
			CreatedAt:            r.CreatedAt,
			UpdatedAt:            r.UpdatedAt,
		}
	}

	totalPages := int32(math.Ceil(float64(totalCount) / float64(limit)))

	return utils.NewResponse(utils.CodeOK, "purchase orders retrieved successfully", PurchaseOrderListResponse{
		Data:       data,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// UpdatePurchaseOrder updates a draft purchase order and its line items.
func (uc *PurchaseOrdersUseCase) UpdatePurchaseOrder(ctx context.Context, id int32, input UpdatePurchaseOrderInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	existing, err := uc.repo.GetPurchaseOrderByID(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "purchase order not found", nil)
	}

	// Only allow editing draft orders
	currentStatus := "draft"
	if existing.Status.Valid && existing.Status.String != "" {
		currentStatus = existing.Status.String
	}
	if currentStatus != "draft" {
		return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("cannot update purchase order with status '%s'; only 'draft' orders can be edited", currentStatus), nil)
	}

	supplierID := existing.PartnersID
	if input.SupplierID != nil && *input.SupplierID > 0 {
		supplierID = *input.SupplierID
	}

	storeID := existing.StoreID
	if input.StoreID != nil && *input.StoreID > 0 {
		storeID = *input.StoreID
	}

	poDate := existing.PoDate
	if input.PoDate != nil {
		if parsed := parseDateString(input.PoDate); parsed != nil {
			poDate = utils.TimeToPgDate(parsed)
		}
	}

	expectedDeliveryDate := existing.ExpectedDeliveryDate
	if input.ExpectedDeliveryDate != nil {
		if parsed := parseDateString(input.ExpectedDeliveryDate); parsed != nil {
			expectedDeliveryDate = utils.TimeToPgDate(parsed)
		}
	}

	priceListID := existing.PriceListID
	if input.PriceListID != nil {
		priceListID = utils.Int32ToPgInt4(input.PriceListID)
	}

	metadataBytes := existing.Metadata
	if input.Metadata != nil {
		metadataBytes, _ = json.Marshal(input.Metadata)
	}

	// If items are provided, replace lines and recalculate totals
	subtotalVal := existing.Subtotal
	discountVal := existing.DiscountAmount
	taxVal := existing.TaxAmount
	totalVal := existing.TotalAmount

	if len(input.Items) > 0 {
		var calculatedSubtotal float64
		var calculatedDiscount float64
		var calculatedTax float64
		var calculatedTotal float64

		type computedLine struct {
			item           PurchaseOrderLineInput
			subtotal       float64
			discountAmount float64
			taxAmount      float64
			lineTotal      float64
			lineNumber     int32
		}

		var computedLines []computedLine
		for idx, item := range input.Items {
			if item.ProductID <= 0 {
				return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("item %d: product_id is required", idx+1), nil)
			}
			if item.Quantity <= 0 {
				return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("item %d: quantity must be greater than 0", idx+1), nil)
			}
			if item.UnitPrice < 0 {
				return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("item %d: unit_price cannot be negative", idx+1), nil)
			}

			sub := item.Quantity * item.UnitPrice
			disc := 0.0
			if item.DiscountAmount != nil && *item.DiscountAmount > 0 {
				disc = *item.DiscountAmount
			}
			tax := 0.0
			if item.TaxAmount != nil && *item.TaxAmount > 0 {
				tax = *item.TaxAmount
			}
			lt := sub - disc + tax
			if lt < 0 {
				lt = 0
			}

			lineNum := int32(idx + 1)
			if item.LineNumber != nil && *item.LineNumber > 0 {
				lineNum = *item.LineNumber
			}

			computedLines = append(computedLines, computedLine{
				item:           item,
				subtotal:       sub,
				discountAmount: disc,
				taxAmount:      tax,
				lineTotal:      lt,
				lineNumber:     lineNum,
			})

			calculatedSubtotal += sub
			calculatedDiscount += disc
			calculatedTax += tax
			calculatedTotal += lt
		}

		// Delete existing lines
		if err := uc.repo.DeletePurchaseOrderLines(ctx, id); err != nil {
			return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to update purchase order lines: %v", err), nil)
		}

		// Recreate lines
		for _, cl := range computedLines {
			itemMetaBytes, _ := json.Marshal(cl.item.Metadata)
			_, err := uc.repo.CreatePurchaseOrderLine(ctx, repository.CreatePurchaseOrderLineParams{
				PurchaseOrderID:  id,
				ProductID:        cl.item.ProductID,
				ProductVariantID: utils.Int32ToPgInt4(cl.item.ProductVariantID),
				Quantity:         utils.Float64ToPgNumeric(cl.item.Quantity),
				UomID:            utils.Int32ToPgInt4(cl.item.UomID),
				UnitPrice:        utils.Float64ToPgNumeric(cl.item.UnitPrice),
				DiscountAmount:   utils.Float64ToPgNumeric(cl.discountAmount),
				TaxAmount:        utils.Float64ToPgNumeric(cl.taxAmount),
				Subtotal:         utils.Float64ToPgNumeric(cl.subtotal),
				LineTotal:        utils.Float64ToPgNumeric(cl.lineTotal),
				ReceivedQuantity: utils.Float64ToPgNumeric(0),
				LineNumber:       pgtype.Int4{Int32: cl.lineNumber, Valid: true},
				Metadata:         itemMetaBytes,
			})
			if err != nil {
				return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to create updated line: %v", err), nil)
			}
		}

		subtotalVal = utils.Float64ToPgNumeric(calculatedSubtotal)
		discountVal = utils.Float64ToPgNumeric(calculatedDiscount)
		taxVal = utils.Float64ToPgNumeric(calculatedTax)
		totalVal = utils.Float64ToPgNumeric(calculatedTotal)
	}

	_, err = uc.repo.UpdatePurchaseOrderHeader(ctx, repository.UpdatePurchaseOrderHeaderParams{
		ID:                   id,
		PartnersID:           supplierID,
		StoreID:              storeID,
		PoDate:               poDate,
		ExpectedDeliveryDate: expectedDeliveryDate,
		Subtotal:             subtotalVal,
		DiscountAmount:       discountVal,
		TaxAmount:            taxVal,
		TotalAmount:          totalVal,
		PriceListID:          priceListID,
		Metadata:             metadataBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to update purchase order header: %v", err), nil)
	}

	return uc.GetPurchaseOrder(ctx, id)
}

// UpdatePurchaseOrderStatus updates status and approved_by.
func (uc *PurchaseOrdersUseCase) UpdatePurchaseOrderStatus(ctx context.Context, id int32, input UpdatePurchaseOrderStatusInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	if input.Status == "" {
		return utils.NewResponse(utils.CodeBadReq, "status is required", nil)
	}

	existing, err := uc.repo.GetPurchaseOrderByID(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "purchase order not found", nil)
	}

	validStatuses := map[string]bool{
		"draft":              true,
		"submitted":          true,
		"approved":           true,
		"partially_received": true,
		"received":           true,
		"cancelled":          true,
		"closed":             true,
	}
	if !validStatuses[input.Status] {
		return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("invalid status '%s'", input.Status), nil)
	}

	approvedBy := existing.ApprovedBy
	if input.ApprovedBy != nil && *input.ApprovedBy > 0 {
		approvedBy = pgtype.Int4{Int32: *input.ApprovedBy, Valid: true}
	} else if input.Status == "approved" && !approvedBy.Valid {
		// If status is approved and no approved_by provided in input, we keep existing or require it
	}

	_, err = uc.repo.UpdatePurchaseOrderStatus(ctx, repository.UpdatePurchaseOrderStatusParams{
		ID:         id,
		Status:     pgtype.Text{String: input.Status, Valid: true},
		ApprovedBy: approvedBy,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to update purchase order status: %v", err), nil)
	}

	return uc.GetPurchaseOrder(ctx, id)
}

// DeletePurchaseOrder deletes a draft purchase order.
func (uc *PurchaseOrdersUseCase) DeletePurchaseOrder(ctx context.Context, id int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	existing, err := uc.repo.GetPurchaseOrderByID(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "purchase order not found", nil)
	}

	currentStatus := "draft"
	if existing.Status.Valid && existing.Status.String != "" {
		currentStatus = existing.Status.String
	}
	if currentStatus != "draft" && currentStatus != "cancelled" {
		return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("cannot delete purchase order with status '%s'; only 'draft' or 'cancelled' orders can be deleted", currentStatus), nil)
	}

	if err := uc.repo.DeletePurchaseOrder(ctx, id); err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to delete purchase order: %v", err), nil)
	}

	return utils.NewResponse(utils.CodeOK, "purchase order deleted successfully", nil)
}

// GetPurchaseOrderLines retrieves all lines for a specific purchase order.
func (uc *PurchaseOrdersUseCase) GetPurchaseOrderLines(ctx context.Context, poID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	_, err := uc.repo.GetPurchaseOrderByID(ctx, poID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "purchase order not found", nil)
	}

	lines, err := uc.repo.ListPurchaseOrderLines(ctx, poID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to load purchase order lines: %v", err), nil)
	}

	items := make([]PurchaseOrderLineOutput, len(lines))
	for i, line := range lines {
		items[i] = formatPOLineOutput(line)
	}

	return utils.NewResponse(utils.CodeOK, "purchase order lines retrieved successfully", items)
}

// GetPurchaseOrderLine retrieves a single purchase order line by its ID.
func (uc *PurchaseOrdersUseCase) GetPurchaseOrderLine(ctx context.Context, lineID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	line, err := uc.repo.GetPurchaseOrderLineByID(ctx, lineID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "purchase order line not found", nil)
	}

	out := formatPOLineByIDOutput(line)
	return utils.NewResponse(utils.CodeOK, "purchase order line retrieved successfully", out)
}

// AddPurchaseOrderLine adds a new line item to a draft purchase order.
func (uc *PurchaseOrdersUseCase) AddPurchaseOrderLine(ctx context.Context, poID int32, item PurchaseOrderLineInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	po, err := uc.repo.GetPurchaseOrderByID(ctx, poID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "purchase order not found", nil)
	}

	status := "draft"
	if po.Status.Valid && po.Status.String != "" {
		status = po.Status.String
	}
	if status != "draft" {
		return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("cannot add items to purchase order with status '%s'; only 'draft' orders can be edited", status), nil)
	}

	if item.ProductID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "product_id is required", nil)
	}
	if item.Quantity <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "quantity must be greater than 0", nil)
	}
	if item.UnitPrice < 0 {
		return utils.NewResponse(utils.CodeBadReq, "unit_price cannot be negative", nil)
	}

	subtotal := item.Quantity * item.UnitPrice
	discount := 0.0
	if item.DiscountAmount != nil && *item.DiscountAmount > 0 {
		discount = *item.DiscountAmount
	}
	tax := 0.0
	if item.TaxAmount != nil && *item.TaxAmount > 0 {
		tax = *item.TaxAmount
	}
	lineTotal := subtotal - discount + tax
	if lineTotal < 0 {
		lineTotal = 0
	}

	lineNum := int32(1)
	if item.LineNumber != nil && *item.LineNumber > 0 {
		lineNum = *item.LineNumber
	} else {
		existingLines, _ := uc.repo.ListPurchaseOrderLines(ctx, poID)
		lineNum = int32(len(existingLines) + 1)
	}

	itemMetaBytes, _ := json.Marshal(item.Metadata)
	newLine, err := uc.repo.CreatePurchaseOrderLine(ctx, repository.CreatePurchaseOrderLineParams{
		PurchaseOrderID:  poID,
		ProductID:        item.ProductID,
		ProductVariantID: utils.Int32ToPgInt4(item.ProductVariantID),
		Quantity:         utils.Float64ToPgNumeric(item.Quantity),
		UomID:            utils.Int32ToPgInt4(item.UomID),
		UnitPrice:        utils.Float64ToPgNumeric(item.UnitPrice),
		DiscountAmount:   utils.Float64ToPgNumeric(discount),
		TaxAmount:        utils.Float64ToPgNumeric(tax),
		Subtotal:         utils.Float64ToPgNumeric(subtotal),
		LineTotal:        utils.Float64ToPgNumeric(lineTotal),
		ReceivedQuantity: utils.Float64ToPgNumeric(0),
		LineNumber:       pgtype.Int4{Int32: lineNum, Valid: true},
		Metadata:         itemMetaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to create purchase order line: %v", err), nil)
	}

	_ = uc.recalculatePOTotals(ctx, poID)
	return uc.GetPurchaseOrderLine(ctx, newLine.ID)
}

// UpdatePurchaseOrderLine updates an existing line in a draft purchase order.
func (uc *PurchaseOrdersUseCase) UpdatePurchaseOrderLine(ctx context.Context, lineID int32, item PurchaseOrderLineInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	existingLine, err := uc.repo.GetPurchaseOrderLineByID(ctx, lineID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "purchase order line not found", nil)
	}

	po, err := uc.repo.GetPurchaseOrderByID(ctx, existingLine.PurchaseOrderID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "parent purchase order not found", nil)
	}

	status := "draft"
	if po.Status.Valid && po.Status.String != "" {
		status = po.Status.String
	}
	if status != "draft" {
		return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("cannot update items on purchase order with status '%s'; only 'draft' orders can be edited", status), nil)
	}

	productID := existingLine.ProductID
	if item.ProductID > 0 {
		productID = item.ProductID
	}

	productVariantID := existingLine.ProductVariantID
	if item.ProductVariantID != nil {
		productVariantID = utils.Int32ToPgInt4(item.ProductVariantID)
	}

	uomID := existingLine.UomID
	if item.UomID != nil {
		uomID = utils.Int32ToPgInt4(item.UomID)
	}

	qty := 0.0
	if qf, err := existingLine.Quantity.Float64Value(); err == nil && qf.Valid {
		qty = qf.Float64
	}
	if item.Quantity > 0 {
		qty = item.Quantity
	}

	unitPrice := 0.0
	if upf, err := existingLine.UnitPrice.Float64Value(); err == nil && upf.Valid {
		unitPrice = upf.Float64
	}
	if item.UnitPrice >= 0 {
		unitPrice = item.UnitPrice
	}

	discount := 0.0
	if df, err := existingLine.DiscountAmount.Float64Value(); err == nil && df.Valid {
		discount = df.Float64
	}
	if item.DiscountAmount != nil {
		discount = *item.DiscountAmount
	}

	tax := 0.0
	if tf, err := existingLine.TaxAmount.Float64Value(); err == nil && tf.Valid {
		tax = tf.Float64
	}
	if item.TaxAmount != nil {
		tax = *item.TaxAmount
	}

	subtotal := qty * unitPrice
	lineTotal := subtotal - discount + tax
	if lineTotal < 0 {
		lineTotal = 0
	}

	lineNum := existingLine.LineNumber
	if item.LineNumber != nil && *item.LineNumber > 0 {
		lineNum = pgtype.Int4{Int32: *item.LineNumber, Valid: true}
	}

	metadataBytes := existingLine.Metadata
	if item.Metadata != nil {
		metadataBytes, _ = json.Marshal(item.Metadata)
	}

	_, err = uc.repo.UpdatePurchaseOrderLine(ctx, repository.UpdatePurchaseOrderLineParams{
		ID:               lineID,
		ProductID:        productID,
		ProductVariantID: productVariantID,
		Quantity:         utils.Float64ToPgNumeric(qty),
		UomID:            uomID,
		UnitPrice:        utils.Float64ToPgNumeric(unitPrice),
		DiscountAmount:   utils.Float64ToPgNumeric(discount),
		TaxAmount:        utils.Float64ToPgNumeric(tax),
		Subtotal:         utils.Float64ToPgNumeric(subtotal),
		LineTotal:        utils.Float64ToPgNumeric(lineTotal),
		LineNumber:       lineNum,
		Metadata:         metadataBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to update purchase order line: %v", err), nil)
	}

	_ = uc.recalculatePOTotals(ctx, existingLine.PurchaseOrderID)
	return uc.GetPurchaseOrderLine(ctx, lineID)
}

// DeletePurchaseOrderLine deletes a line from a draft purchase order.
func (uc *PurchaseOrdersUseCase) DeletePurchaseOrderLine(ctx context.Context, lineID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	existingLine, err := uc.repo.GetPurchaseOrderLineByID(ctx, lineID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "purchase order line not found", nil)
	}

	po, err := uc.repo.GetPurchaseOrderByID(ctx, existingLine.PurchaseOrderID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "parent purchase order not found", nil)
	}

	status := "draft"
	if po.Status.Valid && po.Status.String != "" {
		status = po.Status.String
	}
	if status != "draft" {
		return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("cannot delete items from purchase order with status '%s'; only 'draft' orders can be edited", status), nil)
	}

	if err := uc.repo.DeletePurchaseOrderLine(ctx, lineID); err != nil {
		return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to delete purchase order line: %v", err), nil)
	}

	_ = uc.recalculatePOTotals(ctx, existingLine.PurchaseOrderID)
	return utils.NewResponse(utils.CodeOK, "purchase order line deleted successfully", nil)
}

func (uc *PurchaseOrdersUseCase) recalculatePOTotals(ctx context.Context, poID int32) error {
	lines, err := uc.repo.ListPurchaseOrderLines(ctx, poID)
	if err != nil {
		return err
	}
	po, err := uc.repo.GetPurchaseOrderByID(ctx, poID)
	if err != nil {
		return err
	}

	var subtotal, discount, tax, total float64
	for _, l := range lines {
		if sf, err := l.Subtotal.Float64Value(); err == nil && sf.Valid {
			subtotal += sf.Float64
		}
		if df, err := l.DiscountAmount.Float64Value(); err == nil && df.Valid {
			discount += df.Float64
		}
		if tf, err := l.TaxAmount.Float64Value(); err == nil && tf.Valid {
			tax += tf.Float64
		}
		if totf, err := l.LineTotal.Float64Value(); err == nil && totf.Valid {
			total += totf.Float64
		}
	}

	_, err = uc.repo.UpdatePurchaseOrderHeader(ctx, repository.UpdatePurchaseOrderHeaderParams{
		ID:                   poID,
		PartnersID:           po.PartnersID,
		StoreID:              po.StoreID,
		PoDate:               po.PoDate,
		ExpectedDeliveryDate: po.ExpectedDeliveryDate,
		Subtotal:             utils.Float64ToPgNumeric(subtotal),
		DiscountAmount:       utils.Float64ToPgNumeric(discount),
		TaxAmount:            utils.Float64ToPgNumeric(tax),
		TotalAmount:          utils.Float64ToPgNumeric(total),
		PriceListID:          po.PriceListID,
		Metadata:             po.Metadata,
	})
	return err
}

func formatPOLineOutput(line repository.ListPurchaseOrderLinesRow) PurchaseOrderLineOutput {
	return PurchaseOrderLineOutput{
		ID:               line.ID,
		PurchaseOrderID:  line.PurchaseOrderID,
		ProductID:        line.ProductID,
		ProductName:      line.ProductName,
		ProductSKU:       line.ProductSku,
		ProductVariantID: line.ProductVariantID,
		VariantName:      line.VariantName,
		VariantSKU:       line.VariantSku,
		Quantity:         line.Quantity,
		UomID:            line.UomID,
		UomName:          line.UomName,
		UnitPrice:        line.UnitPrice,
		DiscountAmount:   line.DiscountAmount,
		TaxAmount:        line.TaxAmount,
		Subtotal:         line.Subtotal,
		LineTotal:        line.LineTotal,
		ReceivedQuantity: line.ReceivedQuantity,
		LineNumber:       line.LineNumber,
		Barcode:          line.Barcode,
		Metadata:         line.Metadata,
		CreatedAt:        line.CreatedAt,
	}
}

func formatPOLineByIDOutput(line repository.GetPurchaseOrderLineByIDRow) PurchaseOrderLineOutput {
	return PurchaseOrderLineOutput{
		ID:               line.ID,
		PurchaseOrderID:  line.PurchaseOrderID,
		ProductID:        line.ProductID,
		ProductName:      line.ProductName,
		ProductSKU:       line.ProductSku,
		ProductVariantID: line.ProductVariantID,
		VariantName:      line.VariantName,
		VariantSKU:       line.VariantSku,
		Quantity:         line.Quantity,
		UomID:            line.UomID,
		UomName:          line.UomName,
		UnitPrice:        line.UnitPrice,
		DiscountAmount:   line.DiscountAmount,
		TaxAmount:        line.TaxAmount,
		Subtotal:         line.Subtotal,
		LineTotal:        line.LineTotal,
		ReceivedQuantity: line.ReceivedQuantity,
		LineNumber:       line.LineNumber,
		Barcode:          line.Barcode,
		Metadata:         line.Metadata,
		CreatedAt:        line.CreatedAt,
	}
}

