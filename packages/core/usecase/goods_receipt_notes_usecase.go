package usecase

import (
	"context"
	"encoding/json"
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
	ExpiryDate          *time.Time `json:"expiry_date,omitempty"`
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
	ReceiptDate        *time.Time                  `json:"receipt_date,omitempty"`
	DeliveryNoteNumber *string                     `json:"delivery_note_number,omitempty"`
	Notes              *string                     `json:"notes,omitempty"`
	Metadata           map[string]interface{}      `json:"metadata,omitempty"`
	Items              []GoodsReceiptNoteItemInput `json:"items"`
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
	if input.ReceiptDate != nil {
		receiptDate = utils.TimeToPgTimestamp(*input.ReceiptDate)
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
			ExpiryDate:          utils.TimeToPgDate(item.ExpiryDate),
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
