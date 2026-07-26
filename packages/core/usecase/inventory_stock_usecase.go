package usecase

import (
	"context"
	"encoding/json"
	"math/big"
	"strconv"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// InventoryStockOutput represents inventory stock with JSON-friendly metadata.
type InventoryStockOutput struct {
	ID                 int32            `json:"id"`
	ProductID          int32            `json:"product_id"`
	ProductVariantID   pgtype.Int4      `json:"product_variant_id"`
	StoreID            int32            `json:"store_id"`
	StorageLocationID  pgtype.Int4      `json:"storage_location_id"`
	QuantityOnHand     pgtype.Numeric   `json:"quantity_on_hand"`
	QuantityAllocated  pgtype.Numeric   `json:"quantity_allocated"`
	QuantityAvailable  pgtype.Numeric   `json:"quantity_available"`
	QuantityOnOrder    pgtype.Numeric   `json:"quantity_on_order"`
	QuantityInTransit  pgtype.Numeric   `json:"quantity_in_transit"`
	ReorderLevel       pgtype.Numeric   `json:"reorder_level"`
	ReorderQuantity    pgtype.Numeric   `json:"reorder_quantity"`
	MaxStockLevel      pgtype.Numeric   `json:"max_stock_level"`
	LastCountedAt      pgtype.Timestamp `json:"last_counted_at"`
	Metadata           json.RawMessage  `json:"metadata"`
	CreatedAt          pgtype.Timestamp `json:"created_at"`
	UpdatedAt          pgtype.Timestamp `json:"updated_at"`
}

func inventoryStockToOutput(s repository.InventoryStock) InventoryStockOutput {
	return InventoryStockOutput{
		ID:                s.ID,
		ProductID:         s.ProductID,
		ProductVariantID:  s.ProductVariantID,
		StoreID:           s.StoreID,
		StorageLocationID: s.StorageLocationID,
		QuantityOnHand:    s.QuantityOnHand,
		QuantityAllocated: s.QuantityAllocated,
		QuantityAvailable: s.QuantityAvailable,
		QuantityOnOrder:   s.QuantityOnOrder,
		QuantityInTransit: s.QuantityInTransit,
		ReorderLevel:      s.ReorderLevel,
		ReorderQuantity:   s.ReorderQuantity,
		MaxStockLevel:     s.MaxStockLevel,
		LastCountedAt:     s.LastCountedAt,
		Metadata:          utils.BytesToJSONRawMessage(s.Metadata),
		CreatedAt:         s.CreatedAt,
		UpdatedAt:         s.UpdatedAt,
	}
}

type InventoryStockUseCase struct {
	repo *repository.Queries
}

func NewInventoryStockUseCase() *InventoryStockUseCase {
	return &InventoryStockUseCase{}
}

func (uc *InventoryStockUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *InventoryStockUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

// CreateInventoryStock creates a new inventory stock record.
func (uc *InventoryStockUseCase) CreateInventoryStock(
	ctx context.Context,
	productID int32,
	productVariantID *int32,
	storeID int32,
	storageLocationID *int32,
	quantityOnHand *string,
	quantityAllocated *string,
	quantityAvailable *string,
	quantityOnOrder *string,
	quantityInTransit *string,
	reorderLevel *string,
	reorderQuantity *string,
	maxStockLevel *string,
	metadata map[string]interface{},
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	if productID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "product_id is required", nil)
	}
	if storeID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "store_id is required", nil)
	}

	var pvID pgtype.Int4
	if productVariantID != nil {
		pvID = pgtype.Int4{Int32: *productVariantID, Valid: true}
	}

	var slID pgtype.Int4
	if storageLocationID != nil {
		slID = pgtype.Int4{Int32: *storageLocationID, Valid: true}
	}

	qtyOnHand := pgtype.Numeric{Int: big.NewInt(0), Exp: 0, Valid: true}
	if quantityOnHand != nil && *quantityOnHand != "" {
		if parsed, err := parseNumeric(*quantityOnHand); err == nil {
			qtyOnHand = parsed
		}
	}

	qtyAllocated := pgtype.Numeric{Int: big.NewInt(0), Exp: 0, Valid: true}
	if quantityAllocated != nil && *quantityAllocated != "" {
		if parsed, err := parseNumeric(*quantityAllocated); err == nil {
			qtyAllocated = parsed
		}
	}

	qtyAvailable := pgtype.Numeric{Int: big.NewInt(0), Exp: 0, Valid: true}
	if quantityAvailable != nil && *quantityAvailable != "" {
		if parsed, err := parseNumeric(*quantityAvailable); err == nil {
			qtyAvailable = parsed
		}
	}

	qtyOnOrder := pgtype.Numeric{Int: big.NewInt(0), Exp: 0, Valid: true}
	if quantityOnOrder != nil && *quantityOnOrder != "" {
		if parsed, err := parseNumeric(*quantityOnOrder); err == nil {
			qtyOnOrder = parsed
		}
	}

	qtyInTransit := pgtype.Numeric{Int: big.NewInt(0), Exp: 0, Valid: true}
	if quantityInTransit != nil && *quantityInTransit != "" {
		if parsed, err := parseNumeric(*quantityInTransit); err == nil {
			qtyInTransit = parsed
		}
	}

	var reorderLvl pgtype.Numeric
	if reorderLevel != nil && *reorderLevel != "" {
		if parsed, err := parseNumeric(*reorderLevel); err == nil {
			reorderLvl = parsed
		}
	}

	var reorderQty pgtype.Numeric
	if reorderQuantity != nil && *reorderQuantity != "" {
		if parsed, err := parseNumeric(*reorderQuantity); err == nil {
			reorderQty = parsed
		}
	}

	var maxStock pgtype.Numeric
	if maxStockLevel != nil && *maxStockLevel != "" {
		if parsed, err := parseNumeric(*maxStockLevel); err == nil {
			maxStock = parsed
		}
	}

	metaBytes := []byte("{}")
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			metaBytes = b
		}
	}

	row, err := uc.repo.CreateInventoryStock(ctx, repository.CreateInventoryStockParams{
		ProductID:         productID,
		ProductVariantID:  pvID,
		StoreID:           storeID,
		StorageLocationID: slID,
		QuantityOnHand:    qtyOnHand,
		QuantityAllocated:  qtyAllocated,
		QuantityAvailable: qtyAvailable,
		QuantityOnOrder:   qtyOnOrder,
		QuantityInTransit: qtyInTransit,
		ReorderLevel:      reorderLvl,
		ReorderQuantity:   reorderQty,
		MaxStockLevel:     maxStock,
		Metadata:           metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "inventory stock created", inventoryStockToOutput(row))
}

// GetInventoryStock gets inventory stock by ID.
func (uc *InventoryStockUseCase) GetInventoryStock(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	row, err := uc.repo.GetInventoryStock(ctx, int32(parsed))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "inventory stock not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "inventory stock fetched", inventoryStockToOutput(row))
}

// GetInventoryStockByProductAndStore gets inventory stock by product, variant, store, and location.
func (uc *InventoryStockUseCase) GetInventoryStockByProductAndStore(
	ctx context.Context,
	productID string,
	productVariantID *string,
	storeID string,
	storageLocationID *string,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	prodID, err := strconv.ParseInt(productID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil)
	}

	storeIDInt, err := strconv.ParseInt(storeID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil)
	}

	var pvID pgtype.Int4
	if productVariantID != nil && *productVariantID != "" {
		pvIDInt, err := strconv.ParseInt(*productVariantID, 10, 32)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid product_variant_id", nil)
		}
		pvID = pgtype.Int4{Int32: int32(pvIDInt), Valid: true}
	}

	var slID pgtype.Int4
	if storageLocationID != nil && *storageLocationID != "" {
		slIDInt, err := strconv.ParseInt(*storageLocationID, 10, 32)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid storage_location_id", nil)
		}
		slID = pgtype.Int4{Int32: int32(slIDInt), Valid: true}
	}

	row, err := uc.repo.GetInventoryStockByProductAndStore(ctx, repository.GetInventoryStockByProductAndStoreParams{
		ProductID:         int32(prodID),
		ProductVariantID:  pvID,
		StoreID:           int32(storeIDInt),
		StorageLocationID: slID,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "inventory stock not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "inventory stock fetched", inventoryStockToOutput(row))
}

// UpdateInventoryStock updates an existing inventory stock record.
func (uc *InventoryStockUseCase) UpdateInventoryStock(
	ctx context.Context,
	id string,
	quantityOnHand *string,
	quantityAllocated *string,
	quantityAvailable *string,
	quantityOnOrder *string,
	quantityInTransit *string,
	reorderLevel *string,
	reorderQuantity *string,
	maxStockLevel *string,
	lastCountedAt *string,
	metadata map[string]interface{},
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	params := repository.UpdateInventoryStockParams{
		ID: int32(parsed),
	}

	if quantityOnHand != nil && *quantityOnHand != "" {
		if parsed, err := parseNumeric(*quantityOnHand); err == nil {
			params.QuantityOnHand = parsed
		}
	}

	if quantityAllocated != nil && *quantityAllocated != "" {
		if parsed, err := parseNumeric(*quantityAllocated); err == nil {
			params.QuantityAllocated = parsed
		}
	}

	if quantityAvailable != nil && *quantityAvailable != "" {
		if parsed, err := parseNumeric(*quantityAvailable); err == nil {
			params.QuantityAvailable = parsed
		}
	}

	if quantityOnOrder != nil && *quantityOnOrder != "" {
		if parsed, err := parseNumeric(*quantityOnOrder); err == nil {
			params.QuantityOnOrder = parsed
		}
	}

	if quantityInTransit != nil && *quantityInTransit != "" {
		if parsed, err := parseNumeric(*quantityInTransit); err == nil {
			params.QuantityInTransit = parsed
		}
	}

	if reorderLevel != nil && *reorderLevel != "" {
		if parsed, err := parseNumeric(*reorderLevel); err == nil {
			params.ReorderLevel = parsed
		}
	}

	if reorderQuantity != nil && *reorderQuantity != "" {
		if parsed, err := parseNumeric(*reorderQuantity); err == nil {
			params.ReorderQuantity = parsed
		}
	}

	if maxStockLevel != nil && *maxStockLevel != "" {
		if parsed, err := parseNumeric(*maxStockLevel); err == nil {
			params.MaxStockLevel = parsed
		}
	}

	if lastCountedAt != nil && *lastCountedAt != "" {
		parsed, err := time.Parse("2006-01-02T15:04:05Z07:00", *lastCountedAt)
		if err != nil {
			parsed, err = time.Parse("2006-01-02", *lastCountedAt)
		}
		if err == nil {
			params.LastCountedAt = pgtype.Timestamp{Time: parsed, Valid: true}
		}
	}

	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			params.Metadata = b
		}
	}

	row, err := uc.repo.UpdateInventoryStock(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "inventory stock updated", inventoryStockToOutput(row))
}

// UpsertInventoryStock creates or updates inventory stock.
func (uc *InventoryStockUseCase) UpsertInventoryStock(
	ctx context.Context,
	productID int32,
	productVariantID *int32,
	storeID int32,
	storageLocationID *int32,
	quantityOnHand *string,
	quantityAllocated *string,
	quantityAvailable *string,
	quantityOnOrder *string,
	quantityInTransit *string,
	reorderLevel *string,
	reorderQuantity *string,
	maxStockLevel *string,
	metadata map[string]interface{},
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	var pvID pgtype.Int4
	if productVariantID != nil {
		pvID = pgtype.Int4{Int32: *productVariantID, Valid: true}
	}

	var slID pgtype.Int4
	if storageLocationID != nil {
		slID = pgtype.Int4{Int32: *storageLocationID, Valid: true}
	}

	// Try to get existing record
	existing, err := uc.repo.GetInventoryStockByProductAndStore(ctx, repository.GetInventoryStockByProductAndStoreParams{
		ProductID:         productID,
		ProductVariantID:  pvID,
		StoreID:           storeID,
		StorageLocationID: slID,
	})

	if err != nil {
		// Not found, create new
		return uc.CreateInventoryStock(ctx, productID, productVariantID, storeID, storageLocationID,
			quantityOnHand, quantityAllocated, quantityAvailable, quantityOnOrder, quantityInTransit,
			reorderLevel, reorderQuantity, maxStockLevel, metadata)
	}

	// Update existing
	return uc.UpdateInventoryStock(ctx, strconv.Itoa(int(existing.ID)),
		quantityOnHand, quantityAllocated, quantityAvailable, quantityOnOrder, quantityInTransit,
		reorderLevel, reorderQuantity, maxStockLevel, nil, metadata)
}

// AdjustInventoryStock adjusts inventory stock quantities.
func (uc *InventoryStockUseCase) AdjustInventoryStock(
	ctx context.Context,
	id string,
	quantityOnHandDelta *string,
	quantityAvailableDelta *string,
	quantityAllocatedDelta *string,
	quantityOnOrderDelta *string,
	quantityInTransitDelta *string,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	params := repository.AdjustInventoryStockParams{
		ID: int32(parsed),
	}

	if quantityOnHandDelta != nil && *quantityOnHandDelta != "" {
		if parsed, err := parseNumeric(*quantityOnHandDelta); err == nil {
			params.QuantityOnHand = parsed
		}
	}

	if quantityAvailableDelta != nil && *quantityAvailableDelta != "" {
		if parsed, err := parseNumeric(*quantityAvailableDelta); err == nil {
			params.QuantityAvailable = parsed
		}
	}

	if quantityAllocatedDelta != nil && *quantityAllocatedDelta != "" {
		if parsed, err := parseNumeric(*quantityAllocatedDelta); err == nil {
			params.QuantityAllocated = parsed
		}
	}

	if quantityOnOrderDelta != nil && *quantityOnOrderDelta != "" {
		if parsed, err := parseNumeric(*quantityOnOrderDelta); err == nil {
			params.QuantityOnOrder = parsed
		}
	}

	if quantityInTransitDelta != nil && *quantityInTransitDelta != "" {
		if parsed, err := parseNumeric(*quantityInTransitDelta); err == nil {
			params.QuantityInTransit = parsed
		}
	}

	row, err := uc.repo.AdjustInventoryStock(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "inventory stock adjusted", inventoryStockToOutput(row))
}

// AdjustInventoryStockByProductAndStore adjusts inventory stock by product and store.
func (uc *InventoryStockUseCase) AdjustInventoryStockByProductAndStore(
	ctx context.Context,
	productID string,
	productVariantID *string,
	storeID string,
	quantityOnHandDelta *string,
	quantityAvailableDelta *string,
	quantityAllocatedDelta *string,
	quantityOnOrderDelta *string,
	quantityInTransitDelta *string,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	prodID, err := strconv.ParseInt(productID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil)
	}

	storeIDInt, err := strconv.ParseInt(storeID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil)
	}

	var pvID pgtype.Int4
	if productVariantID != nil && *productVariantID != "" {
		pvIDInt, err := strconv.ParseInt(*productVariantID, 10, 32)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid product_variant_id", nil)
		}
		pvID = pgtype.Int4{Int32: int32(pvIDInt), Valid: true}
	}

	params := repository.AdjustInventoryStockByProductAndStoreParams{
		ProductID:        int32(prodID),
		ProductVariantID: pvID,
		StoreID:          int32(storeIDInt),
	}

	if quantityOnHandDelta != nil && *quantityOnHandDelta != "" {
		if parsed, err := parseNumeric(*quantityOnHandDelta); err == nil {
			params.QuantityOnHand = parsed
		}
	}

	if quantityAvailableDelta != nil && *quantityAvailableDelta != "" {
		if parsed, err := parseNumeric(*quantityAvailableDelta); err == nil {
			params.QuantityAvailable = parsed
		}
	}

	if quantityAllocatedDelta != nil && *quantityAllocatedDelta != "" {
		if parsed, err := parseNumeric(*quantityAllocatedDelta); err == nil {
			params.QuantityAllocated = parsed
		}
	}

	if quantityOnOrderDelta != nil && *quantityOnOrderDelta != "" {
		if parsed, err := parseNumeric(*quantityOnOrderDelta); err == nil {
			params.QuantityOnOrder = parsed
		}
	}

	if quantityInTransitDelta != nil && *quantityInTransitDelta != "" {
		if parsed, err := parseNumeric(*quantityInTransitDelta); err == nil {
			params.QuantityInTransit = parsed
		}
	}

	row, err := uc.repo.AdjustInventoryStockByProductAndStore(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "inventory stock adjusted", inventoryStockToOutput(row))
}

// DeleteInventoryStock deletes inventory stock by ID.
func (uc *InventoryStockUseCase) DeleteInventoryStock(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	if err := uc.repo.DeleteInventoryStock(ctx, int32(parsed)); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "inventory stock deleted", nil)
}

// ListInventoryStock lists all inventory stock records.
func (uc *InventoryStockUseCase) ListInventoryStock(ctx context.Context) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	rows, err := uc.repo.ListInventoryStock(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]InventoryStockOutput, len(rows))
	for i := range rows {
		out[i] = inventoryStockToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "inventory stock listed", out)
}

// ListInventoryStockByStore lists inventory stock by store.
func (uc *InventoryStockUseCase) ListInventoryStockByStore(ctx context.Context, storeID string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	storeIDInt, err := strconv.ParseInt(storeID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil)
	}

	rows, err := uc.repo.ListInventoryStockByStore(ctx, int32(storeIDInt))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]InventoryStockOutput, len(rows))
	for i := range rows {
		out[i] = inventoryStockToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "inventory stock listed", out)
}

// ListInventoryStockByProduct lists inventory stock by product.
func (uc *InventoryStockUseCase) ListInventoryStockByProduct(ctx context.Context, productID string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	prodID, err := strconv.ParseInt(productID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil)
	}

	rows, err := uc.repo.ListInventoryStockByProduct(ctx, int32(prodID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]InventoryStockOutput, len(rows))
	for i := range rows {
		out[i] = inventoryStockToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "inventory stock listed", out)
}

// ListInventoryStockByStorageLocation lists inventory stock by storage location.
func (uc *InventoryStockUseCase) ListInventoryStockByStorageLocation(ctx context.Context, storageLocationID string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	slID, err := strconv.ParseInt(storageLocationID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid storage_location_id", nil)
	}

	rows, err := uc.repo.ListInventoryStockByStorageLocation(ctx, pgtype.Int4{Int32: int32(slID), Valid: true})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]InventoryStockOutput, len(rows))
	for i := range rows {
		out[i] = inventoryStockToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "inventory stock listed", out)
}

// ListInventoryStockByStoreAndLocation lists inventory stock by store and location.
func (uc *InventoryStockUseCase) ListInventoryStockByStoreAndLocation(
	ctx context.Context,
	storeID string,
	storageLocationID *string,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	storeIDInt, err := strconv.ParseInt(storeID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil)
	}

	var slID pgtype.Int4
	if storageLocationID != nil && *storageLocationID != "" {
		slIDInt, err := strconv.ParseInt(*storageLocationID, 10, 32)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid storage_location_id", nil)
		}
		slID = pgtype.Int4{Int32: int32(slIDInt), Valid: true}
	}

	rows, err := uc.repo.ListInventoryStockByStoreAndLocation(ctx, repository.ListInventoryStockByStoreAndLocationParams{
		StoreID:          int32(storeIDInt),
		StorageLocationID: slID,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]InventoryStockOutput, len(rows))
	for i := range rows {
		out[i] = inventoryStockToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "inventory stock listed", out)
}

// GetInventoryStockSummary gets inventory stock summary for a store.
func (uc *InventoryStockUseCase) GetInventoryStockSummary(ctx context.Context, storeID string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	storeIDInt, err := strconv.ParseInt(storeID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid store_id", nil)
	}

	row, err := uc.repo.GetInventoryStockSummary(ctx, int32(storeIDInt))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "inventory stock summary fetched", row)
}
