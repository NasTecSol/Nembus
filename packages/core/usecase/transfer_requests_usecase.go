package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type TransferRequestItemInput struct {
	ProductID         int32   `json:"product_id"`
	ProductVariantID  *int32  `json:"product_variant_id,omitempty"`
	FromLocationID    *int32  `json:"from_location_id,omitempty"`
	ToLocationID      *int32  `json:"to_location_id,omitempty"`
	RequestedQuantity float64 `json:"requested_quantity"`
	UomID             *int32  `json:"uom_id,omitempty"`
	BatchNumber       *string `json:"batch_number,omitempty"`
	Notes             *string `json:"notes,omitempty"`
}

type CreateTransferRequestInput struct {
	OrganizationID       int32                      `json:"organization_id"`
	TransferNumber       string                     `json:"transfer_number"`
	FromStoreID          int32                      `json:"from_store_id"`
	ToStoreID            int32                      `json:"to_store_id"`
	RequestedBy          *int32                     `json:"requested_by,omitempty"`
	RequestDate          *time.Time                 `json:"request_date,omitempty"`
	ExpectedDeliveryDate *time.Time                 `json:"expected_delivery_date,omitempty"`
	Notes                *string                    `json:"notes,omitempty"`
	Metadata             map[string]interface{}     `json:"metadata,omitempty"`
	Items                []TransferRequestItemInput `json:"items"`
}

type TransferRequestItemOutput struct {
	ID                int32           `json:"id"`
	TransferRequestID int32           `json:"transfer_request_id"`
	ProductID         int32           `json:"product_id"`
	ProductName       pgtype.Text     `json:"product_name"`
	ProductSKU        pgtype.Text     `json:"product_sku"`
	ProductVariantID  pgtype.Int4     `json:"product_variant_id"`
	FromLocationID    pgtype.Int4     `json:"from_location_id"`
	ToLocationID      pgtype.Int4     `json:"to_location_id"`
	RequestedQuantity pgtype.Numeric  `json:"requested_quantity"`
	ApprovedQuantity  pgtype.Numeric  `json:"approved_quantity"`
	ShippedQuantity   pgtype.Numeric  `json:"shipped_quantity"`
	ReceivedQuantity  pgtype.Numeric  `json:"received_quantity"`
	UomID             pgtype.Int4     `json:"uom_id"`
	UomName           pgtype.Text     `json:"uom_name"`
	BatchNumber       pgtype.Text     `json:"batch_number"`
	Notes             pgtype.Text     `json:"notes"`
	CreatedAt         pgtype.Timestamp `json:"created_at"`
}

type TransferRequestOutput struct {
	ID                   int32                       `json:"id"`
	OrganizationID       int32                       `json:"organization_id"`
	TransferNumber       string                      `json:"transfer_number"`
	FromStoreID          int32                       `json:"from_store_id"`
	FromStoreName        pgtype.Text                 `json:"from_store_name"`
	ToStoreID            int32                       `json:"to_store_id"`
	ToStoreName          pgtype.Text                 `json:"to_store_name"`
	Status               string                      `json:"status"`
	RequestedBy          pgtype.Int4                 `json:"requested_by"`
	RequestedByName      pgtype.Text                 `json:"requested_by_name"`
	ApprovedBy          pgtype.Int4                 `json:"approved_by"`
	ApprovedByName      pgtype.Text                 `json:"approved_by_name"`
	ShippedBy           pgtype.Int4                 `json:"shipped_by"`
	ReceivedBy          pgtype.Int4                 `json:"received_by"`
	RequestDate          pgtype.Timestamp            `json:"request_date"`
	ExpectedDeliveryDate pgtype.Date                 `json:"expected_delivery_date"`
	ShippedAt            pgtype.Timestamp            `json:"shipped_at"`
	ReceivedAt           pgtype.Timestamp            `json:"received_at"`
	Notes                pgtype.Text                 `json:"notes"`
	Metadata             json.RawMessage             `json:"metadata"`
	CreatedAt            pgtype.Timestamp            `json:"created_at"`
	UpdatedAt            pgtype.Timestamp            `json:"updated_at"`
	Items                []TransferRequestItemOutput `json:"items,omitempty"`
}

type TransferRequestsUseCase struct {
	repo *repository.Queries
}

func NewTransferRequestsUseCase() *TransferRequestsUseCase {
	return &TransferRequestsUseCase{}
}

func (uc *TransferRequestsUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *TransferRequestsUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

func (uc *TransferRequestsUseCase) CreateTransferRequest(ctx context.Context, input CreateTransferRequestInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	if input.OrganizationID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}
	if input.FromStoreID <= 0 || input.ToStoreID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "from_store_id and to_store_id are required", nil)
	}
	if input.FromStoreID == input.ToStoreID {
		return utils.NewResponse(utils.CodeBadReq, "from_store_id and to_store_id must be different", nil)
	}
	if len(input.Items) == 0 {
		return utils.NewResponse(utils.CodeBadReq, "at least one transfer item is required", nil)
	}

	reqDate := pgtype.Timestamp{Time: time.Now(), Valid: true}
	if input.RequestDate != nil {
		reqDate = utils.TimeToPgTimestamp(*input.RequestDate)
	}

	metaBytes, _ := json.Marshal(input.Metadata)

	req, err := uc.repo.CreateTransferRequest(ctx, repository.CreateTransferRequestParams{
		OrganizationID: input.OrganizationID,
		TransferNumber: input.TransferNumber,
		FromStoreID:    input.FromStoreID,
		ToStoreID:      input.ToStoreID,
		Status:         "draft",
		RequestedBy:    utils.Int32ToPgInt4(input.RequestedBy),
		RequestDate:    reqDate,
		ExpectedDeliveryDate: utils.TimeToPgDate(input.ExpectedDeliveryDate),
		Notes:          utils.StringToPgText(input.Notes),
		Metadata:       metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	itemsOutput := make([]TransferRequestItemOutput, 0, len(input.Items))
	for _, item := range input.Items {
		itemRow, err := uc.repo.CreateTransferRequestItem(ctx, repository.CreateTransferRequestItemParams{
			TransferRequestID: req.ID,
			ProductID:         item.ProductID,
			ProductVariantID:  utils.Int32ToPgInt4(item.ProductVariantID),
			FromLocationID:    utils.Int32ToPgInt4(item.FromLocationID),
			ToLocationID:      utils.Int32ToPgInt4(item.ToLocationID),
			RequestedQuantity: utils.Float64ToPgNumeric(item.RequestedQuantity),
			ApprovedQuantity:  utils.Float64ToPgNumeric(0),
			ShippedQuantity:   utils.Float64ToPgNumeric(0),
			ReceivedQuantity:  utils.Float64ToPgNumeric(0),
			UomID:             utils.Int32ToPgInt4(item.UomID),
			BatchNumber:       utils.StringToPgText(item.BatchNumber),
			Notes:             utils.StringToPgText(item.Notes),
		})
		if err != nil {
			return utils.NewResponse(utils.CodeError, err.Error(), nil)
		}
		itemsOutput = append(itemsOutput, TransferRequestItemOutput{
			ID:                itemRow.ID,
			TransferRequestID: itemRow.TransferRequestID,
			ProductID:         itemRow.ProductID,
			ProductVariantID:  itemRow.ProductVariantID,
			FromLocationID:    itemRow.FromLocationID,
			ToLocationID:      itemRow.ToLocationID,
			RequestedQuantity: itemRow.RequestedQuantity,
			ApprovedQuantity:  itemRow.ApprovedQuantity,
			ShippedQuantity:   itemRow.ShippedQuantity,
			ReceivedQuantity:  itemRow.ReceivedQuantity,
			UomID:             itemRow.UomID,
			BatchNumber:       itemRow.BatchNumber,
			Notes:             itemRow.Notes,
			CreatedAt:         itemRow.CreatedAt,
		})
	}

	out := TransferRequestOutput{
		ID:                   req.ID,
		OrganizationID:       req.OrganizationID,
		TransferNumber:       req.TransferNumber,
		FromStoreID:          req.FromStoreID,
		ToStoreID:            req.ToStoreID,
		Status:               req.Status,
		RequestedBy:          req.RequestedBy,
		ApprovedBy:          req.ApprovedBy,
		ShippedBy:           req.ShippedBy,
		ReceivedBy:          req.ReceivedBy,
		RequestDate:          req.RequestDate,
		ExpectedDeliveryDate: req.ExpectedDeliveryDate,
		ShippedAt:            req.ShippedAt,
		ReceivedAt:           req.ReceivedAt,
		Notes:                req.Notes,
		Metadata:             utils.BytesToJSONRawMessage(req.Metadata),
		CreatedAt:            req.CreatedAt,
		UpdatedAt:            req.UpdatedAt,
		Items:                itemsOutput,
	}

	return utils.NewResponse(utils.CodeOK, "transfer request created successfully", out)
}

func (uc *TransferRequestsUseCase) GetTransferRequest(ctx context.Context, id int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	tr, err := uc.repo.GetTransferRequestWithDetails(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "transfer request not found", nil)
	}

	itemRows, err := uc.repo.ListTransferRequestItems(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	items := make([]TransferRequestItemOutput, len(itemRows))
	for i, r := range itemRows {
		items[i] = TransferRequestItemOutput{
			ID:                r.ID,
			TransferRequestID: r.TransferRequestID,
			ProductID:         r.ProductID,
			ProductName:       pgtype.Text{String: r.ProductName, Valid: true},
			ProductSKU:        pgtype.Text{String: r.ProductSku, Valid: true},
			ProductVariantID:  r.ProductVariantID,
			FromLocationID:    r.FromLocationID,
			ToLocationID:      r.ToLocationID,
			RequestedQuantity: r.RequestedQuantity,
			ApprovedQuantity:  r.ApprovedQuantity,
			ShippedQuantity:   r.ShippedQuantity,
			ReceivedQuantity:  r.ReceivedQuantity,
			UomID:             r.UomID,
			UomName:           r.UomName,
			BatchNumber:       r.BatchNumber,
			Notes:             r.Notes,
			CreatedAt:         r.CreatedAt,
		}
	}

	out := TransferRequestOutput{
		ID:                   tr.ID,
		OrganizationID:       tr.OrganizationID,
		TransferNumber:       tr.TransferNumber,
		FromStoreID:          tr.FromStoreID,
		FromStoreName:        tr.FromStoreName,
		ToStoreID:            tr.ToStoreID,
		ToStoreName:          tr.ToStoreName,
		Status:               tr.Status,
		RequestedBy:          tr.RequestedBy,
		RequestedByName:      tr.RequestedByName,
		ApprovedBy:          tr.ApprovedBy,
		ApprovedByName:      tr.ApprovedByName,
		ShippedBy:           tr.ShippedBy,
		ReceivedBy:          tr.ReceivedBy,
		RequestDate:          tr.RequestDate,
		ExpectedDeliveryDate: tr.ExpectedDeliveryDate,
		ShippedAt:            tr.ShippedAt,
		ReceivedAt:           tr.ReceivedAt,
		Notes:                tr.Notes,
		Metadata:             utils.BytesToJSONRawMessage(tr.Metadata),
		CreatedAt:            tr.CreatedAt,
		UpdatedAt:            tr.UpdatedAt,
		Items:                items,
	}

	return utils.NewResponse(utils.CodeOK, "transfer request retrieved successfully", out)
}

func (uc *TransferRequestsUseCase) ListTransferRequestsByOrganization(ctx context.Context, orgID int32, limit, offset int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	rows, err := uc.repo.ListTransferRequestsByOrganization(ctx, repository.ListTransferRequestsByOrganizationParams{
		OrganizationID: orgID,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]TransferRequestOutput, len(rows))
	for i, r := range rows {
		out[i] = TransferRequestOutput{
			ID:                   r.ID,
			OrganizationID:       r.OrganizationID,
			TransferNumber:       r.TransferNumber,
			FromStoreID:          r.FromStoreID,
			FromStoreName:        r.FromStoreName,
			ToStoreID:            r.ToStoreID,
			ToStoreName:          r.ToStoreName,
			Status:               r.Status,
			RequestedBy:          r.RequestedBy,
			ApprovedBy:          r.ApprovedBy,
			ShippedBy:           r.ShippedBy,
			ReceivedBy:          r.ReceivedBy,
			RequestDate:          r.RequestDate,
			ExpectedDeliveryDate: r.ExpectedDeliveryDate,
			ShippedAt:            r.ShippedAt,
			ReceivedAt:           r.ReceivedAt,
			Notes:                r.Notes,
			Metadata:             utils.BytesToJSONRawMessage(r.Metadata),
			CreatedAt:            r.CreatedAt,
			UpdatedAt:            r.UpdatedAt,
		}
	}

	return utils.NewResponse(utils.CodeOK, "transfer requests fetched successfully", out)
}

func (uc *TransferRequestsUseCase) ApproveTransferRequest(ctx context.Context, id int32, approvedBy int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	res, err := uc.repo.CallApproveTransferRequest(ctx, repository.CallApproveTransferRequestParams{
		PTransferRequestID: id,
		PApprovedBy:        approvedBy,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	if !res.Success {
		return utils.NewResponse(utils.CodeBadReq, res.Message, nil)
	}

	return utils.NewResponse(utils.CodeOK, res.Message, nil)
}

func (uc *TransferRequestsUseCase) ShipTransferRequest(ctx context.Context, id int32, shippedBy int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	res, err := uc.repo.CallShipTransferRequest(ctx, repository.CallShipTransferRequestParams{
		PTransferRequestID: id,
		PShippedBy:         shippedBy,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	if !res.Success {
		return utils.NewResponse(utils.CodeBadReq, res.Message, nil)
	}

	return utils.NewResponse(utils.CodeOK, res.Message, nil)
}

func (uc *TransferRequestsUseCase) ReceiveTransferRequest(ctx context.Context, id int32, receivedBy int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	res, err := uc.repo.CallReceiveTransferRequest(ctx, repository.CallReceiveTransferRequestParams{
		PTransferRequestID: id,
		PReceivedBy:        receivedBy,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	if !res.Success {
		return utils.NewResponse(utils.CodeBadReq, res.Message, nil)
	}

	return utils.NewResponse(utils.CodeOK, res.Message, nil)
}
