package usecase

import (
	"context"
	"strings"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// StorageLocationsUseCase handles storage location business logic.
type StorageLocationsUseCase struct {
	repo *repository.Queries
}

// NewStorageLocationsUseCase creates a new storage locations use case.
func NewStorageLocationsUseCase() *StorageLocationsUseCase {
	return &StorageLocationsUseCase{}
}

// SetRepository sets the repository (called per-request from handler with tenant repo).
func (uc *StorageLocationsUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// CreateStorageLocationInput is the input for CreateStorageLocation.
type CreateStorageLocationInput struct {
	StoreID          int32
	Code             string
	Name             string
	LocationType     *string
	ParentLocationID *int32
	IsActive         *bool
	Metadata         []byte
}

// CreateStorageLocation creates a storage location for a store.
func (uc *StorageLocationsUseCase) CreateStorageLocation(ctx context.Context, in *CreateStorageLocationInput) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetStore(ctx, in.StoreID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "store not found", nil)
	}
	code := strings.TrimSpace(in.Code)
	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "code is required", nil)
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return utils.NewResponse(utils.CodeBadReq, "name is required", nil)
	}
	_, err = uc.repo.GetStorageLocationByCode(ctx, repository.GetStorageLocationByCodeParams{
		StoreID: in.StoreID,
		Code:    code,
	})
	if err == nil {
		return utils.NewResponse(utils.CodeBadReq, "storage location with this code already exists for store", nil)
	}
	params := repository.CreateStorageLocationParams{
		StoreID:          in.StoreID,
		Code:             code,
		Name:             name,
		LocationType:     pgtype.Text{},
		ParentLocationID: pgtype.Int4{},
		IsActive:         pgtype.Bool{Bool: true, Valid: true},
		Metadata:         in.Metadata,
	}
	if in.LocationType != nil && strings.TrimSpace(*in.LocationType) != "" {
		params.LocationType = pgtype.Text{String: strings.TrimSpace(*in.LocationType), Valid: true}
	}
	if in.ParentLocationID != nil {
		params.ParentLocationID = pgtype.Int4{Int32: *in.ParentLocationID, Valid: true}
	}
	if in.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *in.IsActive, Valid: true}
	}
	loc, err := uc.repo.CreateStorageLocation(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "storage location created", loc)
}

// GetStorageLocation returns a single storage location by ID.
func (uc *StorageLocationsUseCase) GetStorageLocation(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	loc, err := uc.repo.GetStorageLocation(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "storage location not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "storage location fetched successfully", loc)
}

// GetStorageLocationByCode returns a storage location by store ID and code.
func (uc *StorageLocationsUseCase) GetStorageLocationByCode(ctx context.Context, storeID int32, code string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "store not found", nil)
	}
	loc, err := uc.repo.GetStorageLocationByCode(ctx, repository.GetStorageLocationByCodeParams{
		StoreID: storeID,
		Code:    strings.TrimSpace(code),
	})
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "storage location not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "storage location fetched successfully", loc)
}

// ListStorageLocations returns all storage locations.
func (uc *StorageLocationsUseCase) ListStorageLocations(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	locs, err := uc.repo.ListStorageLocations(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "storage locations fetched successfully", locs)
}

// ListStorageLocationsByStore returns all storage locations for a store.
func (uc *StorageLocationsUseCase) ListStorageLocationsByStore(ctx context.Context, storeID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "store not found", nil)
	}
	locs, err := uc.repo.ListStorageLocationsByStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "storage locations fetched successfully", locs)
}

// ListActiveStorageLocationsByStore returns active storage locations for a store.
func (uc *StorageLocationsUseCase) ListActiveStorageLocationsByStore(ctx context.Context, storeID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "store not found", nil)
	}
	locs, err := uc.repo.ListActiveStorageLocationsByStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "storage locations fetched successfully", locs)
}

// ListStorageLocationsByParent returns storage locations by parent ID.
func (uc *StorageLocationsUseCase) ListStorageLocationsByParent(ctx context.Context, parentLocationID *int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	arg := pgtype.Int4{}
	if parentLocationID != nil {
		arg = pgtype.Int4{Int32: *parentLocationID, Valid: true}
	}
	locs, err := uc.repo.ListStorageLocationsByParent(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "storage locations fetched successfully", locs)
}

// ListStorageLocationsByType returns storage locations for a store by type.
func (uc *StorageLocationsUseCase) ListStorageLocationsByType(ctx context.Context, storeID int32, locationType string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "store not found", nil)
	}
	arg := repository.ListStorageLocationsByTypeParams{
		StoreID:      storeID,
		LocationType: pgtype.Text{String: strings.TrimSpace(locationType), Valid: true},
	}
	locs, err := uc.repo.ListStorageLocationsByType(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "storage locations fetched successfully", locs)
}

// UpdateStorageLocationInput is the input for UpdateStorageLocation.
type UpdateStorageLocationInput struct {
	ID               int32
	Name             *string
	LocationType     *string
	ParentLocationID *int32
	IsActive         *bool
	Metadata         []byte
}

// UpdateStorageLocation updates a storage location.
func (uc *StorageLocationsUseCase) UpdateStorageLocation(ctx context.Context, in *UpdateStorageLocationInput) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	existing, err := uc.repo.GetStorageLocation(ctx, in.ID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "storage location not found", nil)
	}
	name := existing.Name
	if in.Name != nil && strings.TrimSpace(*in.Name) != "" {
		name = strings.TrimSpace(*in.Name)
	}
	params := repository.UpdateStorageLocationParams{
		ID:               in.ID,
		Name:             name,
		LocationType:     pgtype.Text{},
		ParentLocationID: pgtype.Int4{},
		IsActive:         pgtype.Bool{},
		Metadata:         in.Metadata,
	}
	if in.LocationType != nil {
		params.LocationType = pgtype.Text{String: *in.LocationType, Valid: true}
	}
	if in.ParentLocationID != nil {
		params.ParentLocationID = pgtype.Int4{Int32: *in.ParentLocationID, Valid: true}
	}
	if in.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *in.IsActive, Valid: true}
	}
	loc, err := uc.repo.UpdateStorageLocation(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "storage location updated", loc)
}

// DeleteStorageLocation deletes a storage location.
func (uc *StorageLocationsUseCase) DeleteStorageLocation(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetStorageLocation(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "storage location not found", nil)
	}
	if err := uc.repo.DeleteStorageLocation(ctx, id); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "storage location deleted", nil)
}

// ToggleStorageLocationActive sets the active state of a storage location.
func (uc *StorageLocationsUseCase) ToggleStorageLocationActive(ctx context.Context, id int32, isActive bool) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetStorageLocation(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "storage location not found", nil)
	}
	loc, err := uc.repo.ToggleStorageLocationActive(ctx, repository.ToggleStorageLocationActiveParams{
		ID:       id,
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "storage location updated", loc)
}
