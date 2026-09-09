package usecase

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type GoodsReceiptNoteItemInput struct {
	PurchaseOrderLineID *int32   `json:"purchase_order_line_id,omitempty"`
	ProductID           int32    `json:"product_id"`
	ProductVariantID    *int32   `json:"product_variant_id,omitempty"`
	StorageLocationID   *int32   `json:"storage_location_id,omitempty"`
	QuantityReceived    float64  `json:"quantity_received"`
	QuantityRejected    *float64 `json:"quantity_rejected,omitempty"`
	UomID               *int32   `json:"uom_id,omitempty"`
	UnitCost            *float64 `json:"unit_cost,omitempty"`
	BatchNumber         *string  `json:"batch_number,omitempty"`
	ExpiryDate          *string  `json:"expiry_date,omitempty"`
	RejectionReason     *string  `json:"rejection_reason,omitempty"`
	Notes               *string  `json:"notes,omitempty"`
}

type CreateGoodsReceiptNoteInput struct {
	OrganizationID     int32                       `json:"organization_id"`
	GRNNumber          string                      `json:"grn_number"`
	PurchaseOrderID    *int32                      `json:"purchase_order_id,omitempty"`
	SupplierID         int32                       `json:"supplier_id"`
	StoreID            int32                       `json:"store_id"`
	ReceivedBy         *int32                      `json:"received_by,omitempty"`
	ReceiptDate        *string                     `json:"receipt_date,omitempty"`
	DeliveryNoteNumber *string                     `json:"delivery_note_number,omitempty"`
	Notes              *string                     `json:"notes,omitempty"`
	Metadata           map[string]interface{}      `json:"metadata,omitempty"`
	Items              []GoodsReceiptNoteItemInput `json:"items"`
}

type UpdateGoodsReceiptNoteInput struct {
	PurchaseOrderID    *int32                      `json:"purchase_order_id,omitempty"`
	SupplierID         *int32                      `json:"supplier_id,omitempty"`
	StoreID            *int32                      `json:"store_id,omitempty"`
	ReceivedBy         *int32                      `json:"received_by,omitempty"`
	ReceiptDate        *string                     `json:"receipt_date,omitempty"`
	DeliveryNoteNumber *string                     `json:"delivery_note_number,omitempty"`
	Notes              *string                     `json:"notes,omitempty"`
	Metadata           map[string]interface{}      `json:"metadata,omitempty"`
	Items              []GoodsReceiptNoteItemInput `json:"items,omitempty"`
}

type GoodsReceiptNoteFilter struct {
	OrganizationID  *int32  `json:"organization_id,omitempty"`
	StoreID         *int32  `json:"store_id,omitempty"`
	SupplierID      *int32  `json:"supplier_id,omitempty"`
	PurchaseOrderID *int32  `json:"purchase_order_id,omitempty"`
	Status          *string `json:"status,omitempty"`
	FromDate        *string `json:"from_date,omitempty"`
	ToDate          *string `json:"to_date,omitempty"`
	Search          *string `json:"search,omitempty"`
	Page            int32   `json:"page"`
	PageSize        int32   `json:"page_size"`
}

type GoodsReceiptNoteItemOutput struct {
	ID                  int32            `json:"id"`
	GRNID               int32            `json:"grn_id"`
	PurchaseOrderLineID pgtype.Int4      `json:"purchase_order_line_id"`
	ProductID           int32            `json:"product_id"`
	ProductName         pgtype.Text      `json:"product_name"`
	ProductSKU          pgtype.Text      `json:"product_sku"`
	ProductVariantID    pgtype.Int4      `json:"product_variant_id"`
	StorageLocationID   pgtype.Int4      `json:"storage_location_id"`
	QuantityReceived    pgtype.Numeric   `json:"quantity_received"`
	QuantityRejected    pgtype.Numeric   `json:"quantity_rejected"`
	UomID               pgtype.Int4      `json:"uom_id"`
	UomName             pgtype.Text      `json:"uom_name"`
	UnitCost            pgtype.Numeric   `json:"unit_cost"`
	BatchNumber         pgtype.Text      `json:"batch_number"`
	ExpiryDate          pgtype.Date      `json:"expiry_date"`
	RejectionReason     pgtype.Text      `json:"rejection_reason"`
	Notes               pgtype.Text      `json:"notes"`
	CreatedAt           pgtype.Timestamp `json:"created_at"`
}

type GoodsReceiptNoteSummaryOutput struct {
	ID                    int32            `json:"id"`
	OrganizationID        int32            `json:"organization_id"`
	GRNNumber             string           `json:"grn_number"`
	PurchaseOrderID       pgtype.Int4      `json:"purchase_order_id"`
	PONumber              pgtype.Text      `json:"po_number"`
	SupplierID            int32            `json:"supplier_id"`
	SupplierName          string           `json:"supplier_name"`
	SupplierCode          string           `json:"supplier_code"`
	StoreID               int32            `json:"store_id"`
	StoreName             string           `json:"store_name"`
	ReceivedBy            pgtype.Int4      `json:"received_by"`
	ReceivedByName        pgtype.Text      `json:"received_by_name"`
	ReceiptDate           pgtype.Timestamp `json:"receipt_date"`
	DeliveryNoteNumber    pgtype.Text      `json:"delivery_note_number"`
	Status                pgtype.Text      `json:"status"`
	Notes                 pgtype.Text      `json:"notes"`
	ItemCount             int64            `json:"item_count"`
	TotalReceivedQuantity pgtype.Numeric   `json:"total_received_quantity"`
	CreatedAt             pgtype.Timestamp `json:"created_at"`
	UpdatedAt             pgtype.Timestamp `json:"updated_at"`
}

type GoodsReceiptNoteListOutput struct {
	Data       []GoodsReceiptNoteSummaryOutput `json:"data"`
	TotalCount int64                           `json:"total_count"`
	Page       int32                           `json:"page"`
	Limit      int32                           `json:"limit"`
	TotalPages int32                           `json:"total_pages"`
}

type GoodsReceiptNoteOutput struct {
	ID                 int32                        `json:"id"`
	OrganizationID     int32                        `json:"organization_id"`
	GRNNumber          string                       `json:"grn_number"`
	PurchaseOrderID    pgtype.Int4                  `json:"purchase_order_id"`
	PONumber           pgtype.Text                  `json:"po_number"`
	SupplierID         int32                        `json:"supplier_id"`
	SupplierName       pgtype.Text                  `json:"supplier_name"`
	StoreID            int32                        `json:"store_id"`
	StoreName          pgtype.Text                  `json:"store_name"`
	ReceivedBy         pgtype.Int4                  `json:"received_by"`
	ReceivedByName     pgtype.Text                  `json:"received_by_name"`
	ReceiptDate        pgtype.Timestamp             `json:"receipt_date"`
	DeliveryNoteNumber pgtype.Text                  `json:"delivery_note_number"`
	Status             string                       `json:"status"`
	Notes              pgtype.Text                  `json:"notes"`
	Metadata           json.RawMessage              `json:"metadata"`
	CreatedAt          pgtype.Timestamp             `json:"created_at"`
	UpdatedAt          pgtype.Timestamp             `json:"updated_at"`
	Items              []GoodsReceiptNoteItemOutput `json:"items,omitempty"`
}

type GoodsReceiptNotesUseCase struct {
	repo *repository.Queries
}

func NewGoodsReceiptNotesUseCase() *GoodsReceiptNotesUseCase {
	return &GoodsReceiptNotesUseCase{}
}

func (uc *GoodsReceiptNotesUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *GoodsReceiptNotesUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

func (uc *GoodsReceiptNotesUseCase) CreateGoodsReceiptNote(ctx context.Context, input CreateGoodsReceiptNoteInput) *repository.Response {
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
		return utils.NewResponse(utils.CodeBadReq, "at least one GRN item is required", nil)
	}

	receiptDate := pgtype.Timestamp{Time: time.Now(), Valid: true}
	if parsed := parseDateString(input.ReceiptDate); parsed != nil {
		receiptDate = utils.TimeToPgTimestamp(*parsed)
	}

	metaBytes, _ := json.Marshal(input.Metadata)

	grn, err := uc.repo.CreateGoodsReceiptNote(ctx, repository.CreateGoodsReceiptNoteParams{
		OrganizationID:     input.OrganizationID,
		GrnNumber:          input.GRNNumber,
		PurchaseOrderID:    utils.Int32ToPgInt4(input.PurchaseOrderID),
		PartnersID:         input.SupplierID,
		StoreID:            input.StoreID,
		ReceivedBy:         utils.Int32ToPgInt4(input.ReceivedBy),
		ReceiptDate:        receiptDate,
		DeliveryNoteNumber: utils.StringToPgText(input.DeliveryNoteNumber),
		Status:             pgtype.Text{String: "draft", Valid: true},
		Notes:              utils.StringToPgText(input.Notes),
		Metadata:           metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	itemsOutput := make([]GoodsReceiptNoteItemOutput, 0, len(input.Items))
	for _, item := range input.Items {
		var rejectedQty float64
		if item.QuantityRejected != nil {
			rejectedQty = *item.QuantityRejected
		}

		var expiryDate pgtype.Date
		if parsed := parseDateString(item.ExpiryDate); parsed != nil {
			expiryDate = utils.TimeToPgDate(parsed)
		}

		itemRow, err := uc.repo.CreateGoodsReceiptNoteItem(ctx, repository.CreateGoodsReceiptNoteItemParams{
			GrnID:               grn.ID,
			PurchaseOrderLineID: utils.Int32ToPgInt4(item.PurchaseOrderLineID),
			ProductID:           item.ProductID,
			ProductVariantID:    utils.Int32ToPgInt4(item.ProductVariantID),
			StorageLocationID:   utils.Int32ToPgInt4(item.StorageLocationID),
			QuantityReceived:    utils.Float64ToPgNumeric(item.QuantityReceived),
			QuantityRejected:    utils.Float64ToPgNumeric(rejectedQty),
			UomID:               utils.Int32ToPgInt4(item.UomID),
			UnitCost:            utils.Float64PointerToPgNumeric(item.UnitCost),
			BatchNumber:         utils.StringToPgText(item.BatchNumber),
			ExpiryDate:          expiryDate,
			RejectionReason:     utils.StringToPgText(item.RejectionReason),
			Notes:               utils.StringToPgText(item.Notes),
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, err.Error(), nil)
		}

		itemsOutput = append(itemsOutput, GoodsReceiptNoteItemOutput{
			ID:                  itemRow.ID,
			GRNID:               itemRow.GrnID,
			PurchaseOrderLineID: itemRow.PurchaseOrderLineID,
			ProductID:           itemRow.ProductID,
			ProductVariantID:    itemRow.ProductVariantID,
			StorageLocationID:   itemRow.StorageLocationID,
			QuantityReceived:    itemRow.QuantityReceived,
			QuantityRejected:    itemRow.QuantityRejected,
			UomID:               itemRow.UomID,
			UnitCost:            itemRow.UnitCost,
			BatchNumber:         itemRow.BatchNumber,
			ExpiryDate:          itemRow.ExpiryDate,
			RejectionReason:     itemRow.RejectionReason,
			Notes:               itemRow.Notes,
			CreatedAt:           itemRow.CreatedAt,
		})
	}

	out := GoodsReceiptNoteOutput{
		ID:                 grn.ID,
		OrganizationID:     grn.OrganizationID,
		GRNNumber:          grn.GrnNumber,
		PurchaseOrderID:    grn.PurchaseOrderID,
		SupplierID:         grn.PartnersID,
		StoreID:            grn.StoreID,
		ReceivedBy:         grn.ReceivedBy,
		ReceiptDate:        grn.ReceiptDate,
		DeliveryNoteNumber: grn.DeliveryNoteNumber,
		Status:             grn.Status.String,
		Notes:              grn.Notes,
		Metadata:           utils.BytesToJSONRawMessage(grn.Metadata),
		CreatedAt:          grn.CreatedAt,
		UpdatedAt:          grn.UpdatedAt,
		Items:              itemsOutput,
	}

	return utils.NewResponse(utils.CodeOK, "goods receipt note created successfully", out)
}

func (uc *GoodsReceiptNotesUseCase) ListGoodsReceiptNotes(ctx context.Context, filter GoodsReceiptNoteFilter) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var fromDate, toDate pgtype.Date
	if parsed := parseDateString(filter.FromDate); parsed != nil {
		fromDate = utils.TimeToPgDate(parsed)
	}
	if parsed := parseDateString(filter.ToDate); parsed != nil {
		toDate = utils.TimeToPgDate(parsed)
	}

	params := repository.ListGoodsReceiptNotesParams{
		OrganizationID:  utils.Int32ToPgInt4(filter.OrganizationID),
		StoreID:         utils.Int32ToPgInt4(filter.StoreID),
		SupplierID:      utils.Int32ToPgInt4(filter.SupplierID),
		PurchaseOrderID: utils.Int32ToPgInt4(filter.PurchaseOrderID),
		Status:          utils.StringToPgText(filter.Status),
		FromDate:        fromDate,
		ToDate:          toDate,
		Search:          utils.StringToPgText(filter.Search),
		LimitCount:      pageSize,
		OffsetCount:     offset,
	}

	rows, err := uc.repo.ListGoodsReceiptNotes(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	countParams := repository.CountGoodsReceiptNotesParams{
		OrganizationID:  utils.Int32ToPgInt4(filter.OrganizationID),
		StoreID:         utils.Int32ToPgInt4(filter.StoreID),
		SupplierID:      utils.Int32ToPgInt4(filter.SupplierID),
		PurchaseOrderID: utils.Int32ToPgInt4(filter.PurchaseOrderID),
		Status:          utils.StringToPgText(filter.Status),
		FromDate:        fromDate,
		ToDate:          toDate,
		Search:          utils.StringToPgText(filter.Search),
	}

	totalCount, err := uc.repo.CountGoodsReceiptNotes(ctx, countParams)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	items := make([]GoodsReceiptNoteSummaryOutput, len(rows))
	for i, r := range rows {
		items[i] = GoodsReceiptNoteSummaryOutput{
			ID:                    r.ID,
			OrganizationID:        r.OrganizationID,
			GRNNumber:             r.GrnNumber,
			PurchaseOrderID:       r.PurchaseOrderID,
			PONumber:              r.PoNumber,
			SupplierID:            r.PartnersID,
			SupplierName:          r.SupplierName,
			SupplierCode:          r.SupplierCode,
			StoreID:               r.StoreID,
			StoreName:             r.StoreName,
			ReceivedBy:            r.ReceivedBy,
			ReceivedByName:        r.ReceivedByName,
			ReceiptDate:           r.ReceiptDate,
			DeliveryNoteNumber:    r.DeliveryNoteNumber,
			Status:                r.Status,
			Notes:                 r.Notes,
			ItemCount:             r.ItemCount,
			TotalReceivedQuantity: r.TotalReceivedQuantity,
			CreatedAt:             r.CreatedAt,
			UpdatedAt:             r.UpdatedAt,
		}
	}

	totalPages := int32(math.Ceil(float64(totalCount) / float64(pageSize)))

	return utils.NewResponse(utils.CodeOK, "goods receipt notes retrieved successfully", GoodsReceiptNoteListOutput{
		Data:       items,
		TotalCount: totalCount,
		Page:       page,
		Limit:      pageSize,
		TotalPages: totalPages,
	})
}

func (uc *GoodsReceiptNotesUseCase) GetGoodsReceiptNote(ctx context.Context, id int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	grn, err := uc.repo.GetGoodsReceiptNoteWithDetails(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "goods receipt note not found", nil)
	}

	itemRows, err := uc.repo.ListGoodsReceiptNoteItems(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	items := make([]GoodsReceiptNoteItemOutput, len(itemRows))
	for i, r := range itemRows {
		items[i] = GoodsReceiptNoteItemOutput{
			ID:                  r.ID,
			GRNID:               r.GrnID,
			PurchaseOrderLineID: r.PurchaseOrderLineID,
			ProductID:           r.ProductID,
			ProductName:         pgtype.Text{String: r.ProductName, Valid: true},
			ProductSKU:          pgtype.Text{String: r.ProductSku, Valid: true},
			ProductVariantID:    r.ProductVariantID,
			StorageLocationID:   r.StorageLocationID,
			QuantityReceived:    r.QuantityReceived,
			QuantityRejected:    r.QuantityRejected,
			UomID:               r.UomID,
			UomName:             r.UomName,
			UnitCost:            r.UnitCost,
			BatchNumber:         r.BatchNumber,
			ExpiryDate:          r.ExpiryDate,
			RejectionReason:     r.RejectionReason,
			Notes:               r.Notes,
			CreatedAt:           r.CreatedAt,
		}
	}

	out := GoodsReceiptNoteOutput{
		ID:                 grn.ID,
		OrganizationID:     grn.OrganizationID,
		GRNNumber:          grn.GrnNumber,
		PurchaseOrderID:    grn.PurchaseOrderID,
		PONumber:           grn.PoNumber,
		SupplierID:         grn.PartnersID,
		SupplierName:       pgtype.Text{String: grn.SupplierName, Valid: true},
		StoreID:            grn.StoreID,
		StoreName:          pgtype.Text{String: grn.StoreName, Valid: true},
		ReceivedBy:         grn.ReceivedBy,
		ReceivedByName:     grn.ReceivedByName,
		ReceiptDate:        grn.ReceiptDate,
		DeliveryNoteNumber: grn.DeliveryNoteNumber,
		Status:             grn.Status.String,
		Notes:              grn.Notes,
		Metadata:           utils.BytesToJSONRawMessage(grn.Metadata),
		CreatedAt:          grn.CreatedAt,
		UpdatedAt:          grn.UpdatedAt,
		Items:              items,
	}

	return utils.NewResponse(utils.CodeOK, "goods receipt note retrieved successfully", out)
}

func (uc *GoodsReceiptNotesUseCase) PostGoodsReceiptNote(ctx context.Context, id int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	res, err := uc.repo.CallProcessGoodsReceipt(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	if !res.Success {
		return utils.NewResponse(utils.CodeBadReq, res.Message, nil)
	}

	return utils.NewResponse(utils.CodeOK, res.Message, nil)
}

func (uc *GoodsReceiptNotesUseCase) UpdateGoodsReceiptNote(ctx context.Context, id int32, input UpdateGoodsReceiptNoteInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	existing, err := uc.repo.GetGoodsReceiptNote(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "goods receipt note not found", nil)
	}

	if existing.Status.Valid && existing.Status.String != "draft" {
		return utils.NewResponse(utils.CodeBadReq, "only draft goods receipt notes can be updated", nil)
	}

	supplierID := existing.PartnersID
	if input.SupplierID != nil && *input.SupplierID > 0 {
		supplierID = *input.SupplierID
	}

	storeID := existing.StoreID
	if input.StoreID != nil && *input.StoreID > 0 {
		storeID = *input.StoreID
	}

	purchaseOrderID := existing.PurchaseOrderID
	if input.PurchaseOrderID != nil {
		purchaseOrderID = utils.Int32ToPgInt4(input.PurchaseOrderID)
	}

	receivedBy := existing.ReceivedBy
	if input.ReceivedBy != nil {
		receivedBy = utils.Int32ToPgInt4(input.ReceivedBy)
	}

	receiptDate := existing.ReceiptDate
	if parsed := parseDateString(input.ReceiptDate); parsed != nil {
		receiptDate = utils.TimeToPgTimestamp(*parsed)
	}

	deliveryNoteNumber := existing.DeliveryNoteNumber
	if input.DeliveryNoteNumber != nil {
		deliveryNoteNumber = utils.StringToPgText(input.DeliveryNoteNumber)
	}

	notes := existing.Notes
	if input.Notes != nil {
		notes = utils.StringToPgText(input.Notes)
	}

	metaBytes := existing.Metadata
	if input.Metadata != nil {
		metaBytes, _ = json.Marshal(input.Metadata)
	}

	_, err = uc.repo.UpdateGoodsReceiptNoteHeader(ctx, repository.UpdateGoodsReceiptNoteHeaderParams{
		ID:                 id,
		PartnersID:         supplierID,
		StoreID:            storeID,
		PurchaseOrderID:    purchaseOrderID,
		ReceivedBy:         receivedBy,
		ReceiptDate:        receiptDate,
		DeliveryNoteNumber: deliveryNoteNumber,
		Notes:              notes,
		Metadata:           metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	if len(input.Items) > 0 {
		if err := uc.repo.DeleteGoodsReceiptNoteItems(ctx, id); err != nil {
			return utils.NewResponse(utils.CodeError, err.Error(), nil)
		}

		for _, item := range input.Items {
			var rejectedQty float64
			if item.QuantityRejected != nil {
				rejectedQty = *item.QuantityRejected
			}

			var expiryDate pgtype.Date
			if parsed := parseDateString(item.ExpiryDate); parsed != nil {
				expiryDate = utils.TimeToPgDate(parsed)
			}

			_, err := uc.repo.CreateGoodsReceiptNoteItem(ctx, repository.CreateGoodsReceiptNoteItemParams{
				GrnID:               id,
				PurchaseOrderLineID: utils.Int32ToPgInt4(item.PurchaseOrderLineID),
				ProductID:           item.ProductID,
				ProductVariantID:    utils.Int32ToPgInt4(item.ProductVariantID),
				StorageLocationID:   utils.Int32ToPgInt4(item.StorageLocationID),
				QuantityReceived:    utils.Float64ToPgNumeric(item.QuantityReceived),
				QuantityRejected:    utils.Float64ToPgNumeric(rejectedQty),
				UomID:               utils.Int32ToPgInt4(item.UomID),
				UnitCost:            utils.Float64PointerToPgNumeric(item.UnitCost),
				BatchNumber:         utils.StringToPgText(item.BatchNumber),
				ExpiryDate:          expiryDate,
				RejectionReason:     utils.StringToPgText(item.RejectionReason),
				Notes:               utils.StringToPgText(item.Notes),
			})
			if err != nil {
				return utils.NewResponse(utils.CodeError, err.Error(), nil)
			}
		}
	}

	return uc.GetGoodsReceiptNote(ctx, id)
}

func (uc *GoodsReceiptNotesUseCase) DeleteGoodsReceiptNote(ctx context.Context, id int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	grn, err := uc.repo.GetGoodsReceiptNote(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "goods receipt note not found", nil)
	}

	if grn.Status.Valid && grn.Status.String != "draft" {
		return utils.NewResponse(utils.CodeBadReq, "only draft goods receipt notes can be deleted", nil)
	}

	if err := uc.repo.DeleteGoodsReceiptNote(ctx, id); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "goods receipt note deleted successfully", nil)
}

