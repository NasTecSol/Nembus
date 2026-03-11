package usecase

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// StockMovementOutput represents stock movement with JSON-friendly metadata.
type StockMovementOutput struct {
	ID               int32            `json:"id"`
	MovementType     string           `json:"movement_type"`
	ReferenceType    pgtype.Text      `json:"reference_type"`
	ReferenceID      pgtype.Int4      `json:"reference_id"`
	ProductID        int32            `json:"product_id"`
	ProductVariantID pgtype.Int4      `json:"product_variant_id"`
	FromStoreID      pgtype.Int4      `json:"from_store_id"`
	ToStoreID        pgtype.Int4      `json:"to_store_id"`
	FromStoreName    pgtype.Text      `json:"from_store_name"`
	ToStoreName      pgtype.Text      `json:"to_store_name"`
	UomName          pgtype.Text      `json:"uom_name"`
	FromLocationID   pgtype.Int4      `json:"from_location_id"`
	ToLocationID     pgtype.Int4      `json:"to_location_id"`
	Quantity         pgtype.Numeric   `json:"quantity"`
	UomID            pgtype.Int4      `json:"uom_id"`
	BatchNumber      pgtype.Text      `json:"batch_number"`
	SerialNumber     pgtype.Text      `json:"serial_number"`
	MovementDate     pgtype.Timestamp `json:"movement_date"`
	PostedBy         pgtype.Int4      `json:"posted_by"`
	Status           pgtype.Text      `json:"status"`
	CostPerUnit      pgtype.Numeric   `json:"cost_per_unit"`
	TotalValue       pgtype.Numeric   `json:"total_value"`
	Metadata         json.RawMessage  `json:"metadata"`
	CreatedAt        pgtype.Timestamp `json:"created_at"`
}

func stockMovementToOutput(m repository.ListStockMovementsByProductWithDateRangeRow) StockMovementOutput {
	return StockMovementOutput{
		ID:               m.ID,
		MovementType:     m.MovementType,
		ReferenceType:    m.ReferenceType,
		ReferenceID:      m.ReferenceID,
		ProductID:        m.ProductID,
		ProductVariantID: m.ProductVariantID,
		FromStoreID:      m.FromStoreID,
		ToStoreID:        m.ToStoreID,
		FromStoreName:    m.FromStoreName,
		ToStoreName:      m.ToStoreName,
		UomName:          m.UomName,
		FromLocationID:   m.FromLocationID,
		ToLocationID:     m.ToLocationID,
		Quantity:         m.Quantity,
		UomID:            m.UomID,
		BatchNumber:      m.BatchNumber,
		SerialNumber:     m.SerialNumber,
		MovementDate:     m.MovementDate,
		PostedBy:         m.PostedBy,
		Status:           m.Status,
		CostPerUnit:      m.CostPerUnit,
		TotalValue:       m.TotalValue,
		Metadata:         utils.BytesToJSONRawMessage(m.Metadata),
		CreatedAt:        m.CreatedAt,
	}
}

func stockMovementBasicToOutput(m repository.StockMovement) StockMovementOutput {
	return StockMovementOutput{
		ID:               m.ID,
		MovementType:     m.MovementType,
		ReferenceType:    m.ReferenceType,
		ReferenceID:      m.ReferenceID,
		ProductID:        m.ProductID,
		ProductVariantID: m.ProductVariantID,
		FromStoreID:      m.FromStoreID,
		ToStoreID:        m.ToStoreID,
		FromStoreName:    pgtype.Text{},
		ToStoreName:      pgtype.Text{},
		UomName:          pgtype.Text{},
		FromLocationID:   m.FromLocationID,
		ToLocationID:     m.ToLocationID,
		Quantity:         m.Quantity,
		UomID:            m.UomID,
		BatchNumber:      m.BatchNumber,
		SerialNumber:     m.SerialNumber,
		MovementDate:     m.MovementDate,
		PostedBy:         m.PostedBy,
		Status:           m.Status,
		CostPerUnit:      m.CostPerUnit,
		TotalValue:       m.TotalValue,
		Metadata:         utils.BytesToJSONRawMessage(m.Metadata),
		CreatedAt:        m.CreatedAt,
	}
}

type CreateStockMovementInput struct {
	MovementType     string                 `json:"movement_type"`
	ReferenceType    *string                `json:"reference_type"`
	ReferenceID      *int32                 `json:"reference_id"`
	ProductID        int32                  `json:"product_id"`
	ProductVariantID *int32                 `json:"product_variant_id"`
	FromStoreID      *int32                 `json:"from_store_id"`
	ToStoreID        *int32                 `json:"to_store_id"`
	FromLocationID   *int32                 `json:"from_location_id"`
	ToLocationID     *int32                 `json:"to_location_id"`
	Quantity         string                 `json:"quantity"`
	UomID            *int32                 `json:"uom_id"`
	BatchNumber      *string                `json:"batch_number"`
	SerialNumber     *string                `json:"serial_number"`
	MovementDate     string                 `json:"movement_date"`
	PostedBy         *int32                 `json:"posted_by"`
	Status           *string                `json:"status"`
	CostPerUnit      *string                `json:"cost_per_unit"`
	TotalValue       *string                `json:"total_value"`
	Metadata         map[string]interface{} `json:"metadata"`
}

type CreateStockMovementFromSalesOrderInput struct {
	ReferenceID      int32                  `json:"reference_id"`
	ProductID        int32                  `json:"product_id"`
	ProductVariantID *int32                 `json:"product_variant_id"`
	FromStoreID      *int32                 `json:"from_store_id"`
	FromLocationID   *int32                 `json:"from_location_id"`
	Quantity         string                 `json:"quantity"`
	UomID            *int32                 `json:"uom_id"`
	BatchNumber      *string                `json:"batch_number"`
	SerialNumber     *string                `json:"serial_number"`
	MovementDate     *string                `json:"movement_date"`
	PostedBy         *int32                 `json:"posted_by"`
	Status           *string                `json:"status"`
	CostPerUnit      *string                `json:"cost_per_unit"`
	TotalValue       *string                `json:"total_value"`
	Metadata         map[string]interface{} `json:"metadata"`
}

type CreateStockMovementFromPurchaseOrderInput struct {
	ReferenceID      int32                  `json:"reference_id"`
	ProductID        int32                  `json:"product_id"`
	ProductVariantID *int32                 `json:"product_variant_id"`
	ToStoreID        *int32                 `json:"to_store_id"`
	ToLocationID     *int32                 `json:"to_location_id"`
	Quantity         string                 `json:"quantity"`
	UomID            *int32                 `json:"uom_id"`
	BatchNumber      *string                `json:"batch_number"`
	SerialNumber     *string                `json:"serial_number"`
	MovementDate     *string                `json:"movement_date"`
	PostedBy         *int32                 `json:"posted_by"`
	Status           *string                `json:"status"`
	CostPerUnit      *string                `json:"cost_per_unit"`
	TotalValue       *string                `json:"total_value"`
	Metadata         map[string]interface{} `json:"metadata"`
}

type StockMovementsUseCase struct {
	repo *repository.Queries
}

func NewStockMovementsUseCase() *StockMovementsUseCase {
	return &StockMovementsUseCase{}
}

func (uc *StockMovementsUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *StockMovementsUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

// CreateStockMovement creates a stock movement record.
func (uc *StockMovementsUseCase) CreateStockMovement(ctx context.Context, in *CreateStockMovementInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if in == nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid payload", nil)
	}
	if strings.TrimSpace(in.MovementType) == "" {
		return utils.NewResponse(utils.CodeBadReq, "movement_type is required", nil)
	}
	if in.ProductID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "product_id is required", nil)
	}
	if strings.TrimSpace(in.Quantity) == "" {
		return utils.NewResponse(utils.CodeBadReq, "quantity is required", nil)
	}
	if strings.TrimSpace(in.MovementDate) == "" {
		return utils.NewResponse(utils.CodeBadReq, "movement_date is required", nil)
	}

	qty, err := parseNumeric(in.Quantity)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid quantity", nil)
	}
	parsedDate, err := time.Parse(time.RFC3339, in.MovementDate)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "movement_date must be RFC3339", nil)
	}

	var costPerUnit pgtype.Numeric
	if in.CostPerUnit != nil && strings.TrimSpace(*in.CostPerUnit) != "" {
		if parsed, err := parseNumeric(*in.CostPerUnit); err == nil {
			costPerUnit = parsed
		} else {
			return utils.NewResponse(utils.CodeBadReq, "invalid cost_per_unit", nil)
		}
	}
	var totalValue pgtype.Numeric
	if in.TotalValue != nil && strings.TrimSpace(*in.TotalValue) != "" {
		if parsed, err := parseNumeric(*in.TotalValue); err == nil {
			totalValue = parsed
		} else {
			return utils.NewResponse(utils.CodeBadReq, "invalid total_value", nil)
		}
	}

	metaBytes := []byte("{}")
	if in.Metadata != nil {
		if b, err := json.Marshal(in.Metadata); err == nil {
			metaBytes = b
		}
	}

	row, err := uc.repo.CreateStockMovement(ctx, repository.CreateStockMovementParams{
		MovementType:     strings.TrimSpace(in.MovementType),
		ReferenceType:    pgText(in.ReferenceType),
		ReferenceID:      pgInt4(in.ReferenceID),
		ProductID:        in.ProductID,
		ProductVariantID: pgInt4(in.ProductVariantID),
		FromStoreID:      pgInt4(in.FromStoreID),
		ToStoreID:        pgInt4(in.ToStoreID),
		FromLocationID:   pgInt4(in.FromLocationID),
		ToLocationID:     pgInt4(in.ToLocationID),
		Quantity:         qty,
		UomID:            pgInt4(in.UomID),
		BatchNumber:      pgText(in.BatchNumber),
		SerialNumber:     pgText(in.SerialNumber),
		MovementDate:     pgtype.Timestamp{Time: parsedDate, Valid: true},
		PostedBy:         pgInt4(in.PostedBy),
		Status:           pgText(in.Status),
		CostPerUnit:      costPerUnit,
		TotalValue:       totalValue,
		Metadata:         metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "stock movement created", stockMovementBasicToOutput(row))
}

// CreateStockMovementFromSalesOrder creates a stock movement from sales order.
func (uc *StockMovementsUseCase) CreateStockMovementFromSalesOrder(ctx context.Context, in *CreateStockMovementFromSalesOrderInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if in == nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid payload", nil)
	}
	if in.ReferenceID <= 0 || in.ProductID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "reference_id and product_id are required", nil)
	}
	if strings.TrimSpace(in.Quantity) == "" {
		return utils.NewResponse(utils.CodeBadReq, "quantity is required", nil)
	}

	qty, err := parseNumeric(in.Quantity)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid quantity", nil)
	}
	var costPerUnit pgtype.Numeric
	if in.CostPerUnit != nil && strings.TrimSpace(*in.CostPerUnit) != "" {
		if parsed, err := parseNumeric(*in.CostPerUnit); err == nil {
			costPerUnit = parsed
		} else {
			return utils.NewResponse(utils.CodeBadReq, "invalid cost_per_unit", nil)
		}
	}
	var totalValue pgtype.Numeric
	if in.TotalValue != nil && strings.TrimSpace(*in.TotalValue) != "" {
		if parsed, err := parseNumeric(*in.TotalValue); err == nil {
			totalValue = parsed
		} else {
			return utils.NewResponse(utils.CodeBadReq, "invalid total_value", nil)
		}
	}

	var movementDate interface{}
	if in.MovementDate != nil && strings.TrimSpace(*in.MovementDate) != "" {
		parsed, err := time.Parse(time.RFC3339, *in.MovementDate)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "movement_date must be RFC3339", nil)
		}
		movementDate = parsed
	}

	var status interface{}
	if in.Status != nil && strings.TrimSpace(*in.Status) != "" {
		status = strings.TrimSpace(*in.Status)
	}

	var meta interface{}
	if in.Metadata != nil {
		if b, err := json.Marshal(in.Metadata); err == nil {
			meta = b
		}
	}

	row, err := uc.repo.CreateStockMovementFromSalesOrder(ctx, repository.CreateStockMovementFromSalesOrderParams{
		ReferenceID:      pgtype.Int4{Int32: in.ReferenceID, Valid: true},
		ProductID:        in.ProductID,
		ProductVariantID: pgInt4(in.ProductVariantID),
		FromStoreID:      pgInt4(in.FromStoreID),
		FromLocationID:   pgInt4(in.FromLocationID),
		Quantity:         qty,
		UomID:            pgInt4(in.UomID),
		BatchNumber:      pgText(in.BatchNumber),
		SerialNumber:     pgText(in.SerialNumber),
		Column10:         movementDate,
		PostedBy:         pgInt4(in.PostedBy),
		Column12:         status,
		CostPerUnit:      costPerUnit,
		TotalValue:       totalValue,
		Column15:         meta,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "stock movement created", stockMovementBasicToOutput(row))
}

// CreateStockMovementFromPurchaseOrder creates a stock movement from purchase order.
func (uc *StockMovementsUseCase) CreateStockMovementFromPurchaseOrder(ctx context.Context, in *CreateStockMovementFromPurchaseOrderInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if in == nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid payload", nil)
	}
	if in.ReferenceID <= 0 || in.ProductID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "reference_id and product_id are required", nil)
	}
	if strings.TrimSpace(in.Quantity) == "" {
		return utils.NewResponse(utils.CodeBadReq, "quantity is required", nil)
	}

	qty, err := parseNumeric(in.Quantity)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid quantity", nil)
	}
	var costPerUnit pgtype.Numeric
	if in.CostPerUnit != nil && strings.TrimSpace(*in.CostPerUnit) != "" {
		if parsed, err := parseNumeric(*in.CostPerUnit); err == nil {
			costPerUnit = parsed
		} else {
			return utils.NewResponse(utils.CodeBadReq, "invalid cost_per_unit", nil)
		}
	}
	var totalValue pgtype.Numeric
	if in.TotalValue != nil && strings.TrimSpace(*in.TotalValue) != "" {
		if parsed, err := parseNumeric(*in.TotalValue); err == nil {
			totalValue = parsed
		} else {
			return utils.NewResponse(utils.CodeBadReq, "invalid total_value", nil)
		}
	}

	var movementDate interface{}
	if in.MovementDate != nil && strings.TrimSpace(*in.MovementDate) != "" {
		parsed, err := time.Parse(time.RFC3339, *in.MovementDate)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "movement_date must be RFC3339", nil)
		}
		movementDate = parsed
	}

	var status interface{}
	if in.Status != nil && strings.TrimSpace(*in.Status) != "" {
		status = strings.TrimSpace(*in.Status)
	}

	var meta interface{}
	if in.Metadata != nil {
		if b, err := json.Marshal(in.Metadata); err == nil {
			meta = b
		}
	}

	row, err := uc.repo.CreateStockMovementFromPurchaseOrder(ctx, repository.CreateStockMovementFromPurchaseOrderParams{
		ReferenceID:      pgtype.Int4{Int32: in.ReferenceID, Valid: true},
		ProductID:        in.ProductID,
		ProductVariantID: pgInt4(in.ProductVariantID),
		ToStoreID:        pgInt4(in.ToStoreID),
		ToLocationID:     pgInt4(in.ToLocationID),
		Quantity:         qty,
		UomID:            pgInt4(in.UomID),
		BatchNumber:      pgText(in.BatchNumber),
		SerialNumber:     pgText(in.SerialNumber),
		Column10:         movementDate,
		PostedBy:         pgInt4(in.PostedBy),
		Column12:         status,
		CostPerUnit:      costPerUnit,
		TotalValue:       totalValue,
		Column15:         meta,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "stock movement created", stockMovementBasicToOutput(row))
}

// CreateStockMovementsFromOrderFulfillment creates movements from sales order fulfillment.
func (uc *StockMovementsUseCase) CreateStockMovementsFromOrderFulfillment(ctx context.Context, salesOrderID string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	id, err := parseUUID(salesOrderID)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid sales_order_id", nil)
	}
	rows, err := uc.repo.CreateStockMovementsFromOrderFulfillment(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]StockMovementOutput, len(rows))
	for i := range rows {
		out[i] = stockMovementBasicToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeCreated, "stock movements created", out)
}

// GetStockMovement gets a stock movement by ID.
func (uc *StockMovementsUseCase) GetStockMovement(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}
	row, err := uc.repo.GetStockMovement(ctx, int32(parsed))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "stock movement not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "stock movement fetched", stockMovementBasicToOutput(row))
}

// ListStockMovements lists stock movements with pagination.
func (uc *StockMovementsUseCase) ListStockMovements(ctx context.Context, limit, offset int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	rows, err := uc.repo.ListStockMovements(ctx, repository.ListStockMovementsParams{Limit: limit, Offset: offset})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]StockMovementOutput, len(rows))
	for i := range rows {
		out[i] = stockMovementBasicToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "stock movements fetched", out)
}

// ListStockMovementsByProduct lists stock movements by product with pagination.
func (uc *StockMovementsUseCase) ListStockMovementsByProduct(ctx context.Context, productID string, limit, offset int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	prodID, err := strconv.ParseInt(productID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil)
	}
	rows, err := uc.repo.ListStockMovementsByProduct(ctx, repository.ListStockMovementsByProductParams{
		ProductID: int32(prodID),
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]StockMovementOutput, len(rows))
	for i := range rows {
		out[i] = stockMovementBasicToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "stock movements fetched", out)
}

// ListStockMovementsByStore lists stock movements by store with pagination.
func (uc *StockMovementsUseCase) ListStockMovementsByStore(ctx context.Context, storeID string, limit, offset int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	sID, err := strconv.ParseInt(storeID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil)
	}
	rows, err := uc.repo.ListStockMovementsByStore(ctx, repository.ListStockMovementsByStoreParams{
		FromStoreID: pgtype.Int4{Int32: int32(sID), Valid: true},
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]StockMovementOutput, len(rows))
	for i := range rows {
		out[i] = stockMovementBasicToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "stock movements fetched", out)
}

// ListStockMovementsByType lists stock movements by type with pagination.
func (uc *StockMovementsUseCase) ListStockMovementsByType(ctx context.Context, movementType string, limit, offset int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if strings.TrimSpace(movementType) == "" {
		return utils.NewResponse(utils.CodeBadReq, "movement_type is required", nil)
	}
	rows, err := uc.repo.ListStockMovementsByType(ctx, repository.ListStockMovementsByTypeParams{
		MovementType: strings.TrimSpace(movementType),
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]StockMovementOutput, len(rows))
	for i := range rows {
		out[i] = stockMovementBasicToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "stock movements fetched", out)
}

// ListStockMovementsByReference lists stock movements by reference.
func (uc *StockMovementsUseCase) ListStockMovementsByReference(ctx context.Context, referenceType string, referenceID string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if strings.TrimSpace(referenceType) == "" {
		return utils.NewResponse(utils.CodeBadReq, "reference_type is required", nil)
	}
	refID, err := strconv.ParseInt(referenceID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid reference_id", nil)
	}
	rows, err := uc.repo.ListStockMovementsByReference(ctx, repository.ListStockMovementsByReferenceParams{
		ReferenceType: pgtype.Text{String: strings.TrimSpace(referenceType), Valid: true},
		ReferenceID:   pgtype.Int4{Int32: int32(refID), Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]StockMovementOutput, len(rows))
	for i := range rows {
		out[i] = stockMovementBasicToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "stock movements fetched", out)
}

// ListStockMovementsByDateRange lists stock movements by movement date range.
func (uc *StockMovementsUseCase) ListStockMovementsByDateRange(ctx context.Context, startDate, endDate *time.Time) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if startDate == nil || endDate == nil {
		return utils.NewResponse(utils.CodeBadReq, "start_date and end_date are required", nil)
	}
	rows, err := uc.repo.ListStockMovementsByDateRange(ctx, repository.ListStockMovementsByDateRangeParams{
		MovementDate:   pgtype.Timestamp{Time: *startDate, Valid: true},
		MovementDate_2: pgtype.Timestamp{Time: *endDate, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]StockMovementOutput, len(rows))
	for i := range rows {
		out[i] = stockMovementBasicToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "stock movements fetched", out)
}

// GetStockMovementsByProductAndStore lists movements by product and store from a start date.
func (uc *StockMovementsUseCase) GetStockMovementsByProductAndStore(ctx context.Context, productID, storeID string, startDate *time.Time) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	prodID, err := strconv.ParseInt(productID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil)
	}
	sID, err := strconv.ParseInt(storeID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil)
	}
	if startDate == nil {
		return utils.NewResponse(utils.CodeBadReq, "start_date is required", nil)
	}
	rows, err := uc.repo.GetStockMovementsByProductAndStore(ctx, repository.GetStockMovementsByProductAndStoreParams{
		ProductID:    int32(prodID),
		FromStoreID:  pgtype.Int4{Int32: int32(sID), Valid: true},
		MovementDate: pgtype.Timestamp{Time: *startDate, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]StockMovementOutput, len(rows))
	for i := range rows {
		out[i] = stockMovementBasicToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "stock movements fetched", out)
}

// GetStockMovementsBySalesOrder lists movements by sales order.
func (uc *StockMovementsUseCase) GetStockMovementsBySalesOrder(ctx context.Context, referenceID string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	refID, err := strconv.ParseInt(referenceID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid reference_id", nil)
	}
	rows, err := uc.repo.GetStockMovementsBySalesOrder(ctx, pgtype.Int4{Int32: int32(refID), Valid: true})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]StockMovementOutput, len(rows))
	for i := range rows {
		out[i] = stockMovementBasicToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "stock movements fetched", out)
}

// GetStockMovementsByPurchaseOrder lists movements by purchase order.
func (uc *StockMovementsUseCase) GetStockMovementsByPurchaseOrder(ctx context.Context, referenceID string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	refID, err := strconv.ParseInt(referenceID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid reference_id", nil)
	}
	rows, err := uc.repo.GetStockMovementsByPurchaseOrder(ctx, pgtype.Int4{Int32: int32(refID), Valid: true})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]StockMovementOutput, len(rows))
	for i := range rows {
		out[i] = stockMovementBasicToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "stock movements fetched", out)
}

// GetStockMovementSummaryByProduct returns movement summary for a product and store.
func (uc *StockMovementsUseCase) GetStockMovementSummaryByProduct(ctx context.Context, productID, storeID string, startDate, endDate *time.Time) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	prodID, err := strconv.ParseInt(productID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil)
	}
	sID, err := strconv.ParseInt(storeID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil)
	}
	if startDate == nil || endDate == nil {
		return utils.NewResponse(utils.CodeBadReq, "start_date and end_date are required", nil)
	}
	row, err := uc.repo.GetStockMovementSummaryByProduct(ctx, repository.GetStockMovementSummaryByProductParams{
		ProductID:      int32(prodID),
		ToStoreID:      pgtype.Int4{Int32: int32(sID), Valid: true},
		MovementDate:   pgtype.Timestamp{Time: *startDate, Valid: true},
		MovementDate_2: pgtype.Timestamp{Time: *endDate, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "stock movement summary fetched", row)
}

// UpdateStockMovementStatus updates stock movement status.
func (uc *StockMovementsUseCase) UpdateStockMovementStatus(ctx context.Context, id string, status string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}
	if strings.TrimSpace(status) == "" {
		return utils.NewResponse(utils.CodeBadReq, "status is required", nil)
	}
	row, err := uc.repo.UpdateStockMovementStatus(ctx, repository.UpdateStockMovementStatusParams{
		ID:     int32(parsed),
		Status: pgtype.Text{String: strings.TrimSpace(status), Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "stock movement status updated", stockMovementBasicToOutput(row))
}

// UpdateInventoryStockFromMovement updates inventory stock using movement data.
func (uc *StockMovementsUseCase) UpdateInventoryStockFromMovement(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}
	row, err := uc.repo.UpdateInventoryStockFromMovement(ctx, int32(parsed))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "inventory stock updated", row)
}

// UpsertInventoryStockFromMovement upserts inventory stock using movement data.
func (uc *StockMovementsUseCase) UpsertInventoryStockFromMovement(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}
	row, err := uc.repo.UpsertInventoryStockFromMovement(ctx, int32(parsed))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "inventory stock upserted", row)
}

// ListStockMovementsByProductWithDateRange lists stock movements by product with optional movement date range.
func (uc *StockMovementsUseCase) ListStockMovementsByProductWithDateRange(
	ctx context.Context,
	productID string,
	startDate *time.Time,
	endDate *time.Time,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	prodID, err := strconv.ParseInt(productID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil)
	}

	params := repository.ListStockMovementsByProductWithDateRangeParams{
		ProductID: int32(prodID),
		Column2:   pgtype.Timestamp{},
		Column3:   pgtype.Timestamp{},
	}

	if startDate != nil {
		params.Column2 = pgtype.Timestamp{Time: *startDate, Valid: true}
	}
	if endDate != nil {
		params.Column3 = pgtype.Timestamp{Time: *endDate, Valid: true}
	}

	rows, err := uc.repo.ListStockMovementsByProductWithDateRange(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]StockMovementOutput, len(rows))
	for i := range rows {
		out[i] = stockMovementToOutput(rows[i])
	}

	return utils.NewResponse(utils.CodeOK, "stock movements fetched", out)
}

func pgText(s *string) pgtype.Text {
	if s == nil || strings.TrimSpace(*s) == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: strings.TrimSpace(*s), Valid: true}
}

func pgInt4(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
