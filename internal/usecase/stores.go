package usecase

import (
	"context"
	"encoding/json"
	"strconv"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type StoreUseCase struct {
	repo *repository.Queries
}

// NewStoreUseCase creates a new store use case
func NewStoreUseCase() *StoreUseCase {
	return &StoreUseCase{}
}

// SetRepository injects repository per request
func (uc *StoreUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// --------------------------------------------------
// helpers
// --------------------------------------------------

func (uc *StoreUseCase) getOrganizationID(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	orgs, err := uc.repo.ListOrganizations(ctx, repository.ListOrganizationsParams{
		Limit:    1,
		Offset:   0,
		IsActive: pgtype.Bool{Bool: true, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	if len(orgs) == 0 {
		return utils.NewResponse(utils.CodeNotFound, "no active organization found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "organization found", orgs[0].ID)
}

// --------------------------------------------------
// Create Store
// --------------------------------------------------

func (uc *StoreUseCase) CreateStore(
	ctx context.Context,
	name string,
	code string,
	storeType *string,
	parentStoreID *int32,
	isWarehouse bool,
	isPOSEnabled bool,
	timezone *string,
	isActive bool,
	metadata any,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if name == "" {
		return utils.NewResponse(utils.CodeBadReq, "store name is required", nil)
	}
	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "store code is required", nil)
	}

	orgResp := uc.getOrganizationID(ctx)
	if orgResp.StatusCode != utils.CodeOK {
		return orgResp
	}
	orgID := orgResp.Data.(int32)

	// Optional fields
	var parentID pgtype.Int4
	if parentStoreID != nil {
		parentID = pgtype.Int4{Int32: *parentStoreID, Valid: true}
	}

	var storeTypeText pgtype.Text
	if storeType != nil {
		storeTypeText = pgtype.Text{String: *storeType, Valid: true}
	}

	var timezoneText pgtype.Text
	if timezone != nil {
		timezoneText = pgtype.Text{String: *timezone, Valid: true}
	}

	// Metadata → JSON → []byte
	metaBytes := []byte("{}")
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			metaBytes = b
		}
	}

	store, err := uc.repo.CreateStore(ctx, repository.CreateStoreParams{
		OrganizationID: orgID,
		ParentStoreID:  parentID,
		Name:           name,
		Code:           code,
		StoreType:      storeTypeText,
		IsWarehouse:    pgtype.Bool{Bool: isWarehouse, Valid: true},
		IsPosEnabled:   pgtype.Bool{Bool: isPOSEnabled, Valid: true},
		Timezone:       timezoneText,
		IsActive:       pgtype.Bool{Bool: isActive, Valid: true},
		Metadata:       metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "store created successfully", store)
}

// --------------------------------------------------
// Get Store
// --------------------------------------------------

func (uc *StoreUseCase) GetStore(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	storeID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid store id", nil)
	}

	store, err := uc.repo.GetStore(ctx, int32(storeID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "store fetched successfully", store)
}

// --------------------------------------------------
// List Stores
// --------------------------------------------------

func (uc *StoreUseCase) ListStores(
	ctx context.Context,
	limit, offset int32,
	isActive *bool,
	storeType *string,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	orgResp := uc.getOrganizationID(ctx)
	if orgResp.StatusCode != utils.CodeOK {
		return orgResp
	}
	orgID := orgResp.Data.(int32)

	var activeBool pgtype.Bool
	if isActive != nil {
		activeBool = pgtype.Bool{Bool: *isActive, Valid: true}
	}

	var storeTypeText pgtype.Text
	if storeType != nil {
		storeTypeText = pgtype.Text{String: *storeType, Valid: true}
	}

	stores, err := uc.repo.ListStores(ctx, repository.ListStoresParams{
		OrganizationID: orgID,
		Limit:          limit,
		Offset:         offset,
		IsActive:       activeBool,
		StoreType:      storeTypeText,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "stores fetched successfully", stores)
}

// --------------------------------------------------
// Update Store
// --------------------------------------------------

func (uc *StoreUseCase) UpdateStore(
	ctx context.Context,
	id string,
	name *string,
	storeType *string,
	isWarehouse *bool,
	isPOSEnabled *bool,
	timezone *string,
	isActive *bool,
	metadata any,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	storeID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid store id", nil)
	}

	var nameText pgtype.Text
	if name != nil {
		nameText = pgtype.Text{String: *name, Valid: true}
	}

	var storeTypeText pgtype.Text
	if storeType != nil {
		storeTypeText = pgtype.Text{String: *storeType, Valid: true}
	}

	var timezoneText pgtype.Text
	if timezone != nil {
		timezoneText = pgtype.Text{String: *timezone, Valid: true}
	}

	var warehouseBool pgtype.Bool
	if isWarehouse != nil {
		warehouseBool = pgtype.Bool{Bool: *isWarehouse, Valid: true}
	}

	var posBool pgtype.Bool
	if isPOSEnabled != nil {
		posBool = pgtype.Bool{Bool: *isPOSEnabled, Valid: true}
	}

	var activeBool pgtype.Bool
	if isActive != nil {
		activeBool = pgtype.Bool{Bool: *isActive, Valid: true}
	}

	metaBytes := []byte("{}")
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			metaBytes = b
		}
	}

	store, err := uc.repo.UpdateStore(ctx, repository.UpdateStoreParams{
		Name:         nameText,
		StoreType:    storeTypeText,
		IsWarehouse:  warehouseBool,
		IsPosEnabled: posBool,
		Timezone:     timezoneText,
		IsActive:     activeBool,
		Metadata:     metaBytes,
		ID:           int32(storeID),
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "store updated successfully", store)
}

// --------------------------------------------------
// Delete Store
// --------------------------------------------------

func (uc *StoreUseCase) DeleteStore(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	storeID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid store id", nil)
	}

	if err := uc.repo.DeleteStore(ctx, int32(storeID)); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "store deleted successfully", nil)
}

// --------------------------------------------------
// List POS Enabled Stores
// --------------------------------------------------
func (uc *StoreUseCase) ListPOSEnabledStores(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	orgResp := uc.getOrganizationID(ctx)
	if orgResp.StatusCode != utils.CodeOK {
		return orgResp
	}
	orgID := orgResp.Data.(int32)

	stores, err := uc.repo.ListPOSEnabledStores(ctx, orgID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "POS enabled stores fetched successfully", stores)
}

// --------------------------------------------------
// List Warehouse Stores
// --------------------------------------------------
func (uc *StoreUseCase) ListWarehouseStores(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	orgResp := uc.getOrganizationID(ctx)
	if orgResp.StatusCode != utils.CodeOK {
		return orgResp
	}
	orgID := orgResp.Data.(int32)

	stores, err := uc.repo.ListWarehouseStores(ctx, orgID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "warehouse stores fetched successfully", stores)
}

// --------------------------------------------------
// List Stores by Parent
// --------------------------------------------------
func (uc *StoreUseCase) ListStoresByParent(ctx context.Context, parentStoreID int32, isActive *bool) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	var activeBool pgtype.Bool
	if isActive != nil {
		activeBool = pgtype.Bool{Bool: *isActive, Valid: true}
	}

	stores, err := uc.repo.ListStoresByParent(ctx, repository.ListStoresByParentParams{
		ParentStoreID: pgtype.Int4{Int32: parentStoreID, Valid: true},
		IsActive:      activeBool,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "stores by parent fetched successfully", stores)
}

// --------------------------------------------------
// Get Storage Location Hierarchy
// --------------------------------------------------
func (uc *StoreUseCase) GetStorageLocationHierarchy(ctx context.Context, storeID int32, filterIsActive *bool) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	var filter interface{}
	if filterIsActive != nil {
		filter = *filterIsActive
	} else {
		filter = nil
	}

	hierarchy, err := uc.repo.GetStorageLocationHierarchy(ctx, repository.GetStorageLocationHierarchyParams{
		StoreID:        storeID,
		FilterIsActive: filter,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "storage location hierarchy fetched successfully", hierarchy)
}
