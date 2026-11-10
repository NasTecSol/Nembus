package usecase

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// BrandOutput is the response shape for brand APIs
type BrandOutput struct {
	ID        int32            `json:"id"`
	Name      string           `json:"name"`
	Code      string           `json:"code"`
	IsActive  pgtype.Bool      `json:"is_active"`
	Metadata  json.RawMessage  `json:"metadata"`
	CreatedAt pgtype.Timestamp `json:"created_at"`
	UpdatedAt pgtype.Timestamp `json:"updated_at"`
}

func brandToOutput(b repository.Brand) BrandOutput {
	return BrandOutput{
		ID:        b.ID,
		Name:      b.Name,
		Code:      b.Code,
		IsActive:  b.IsActive,
		Metadata:  utils.BytesToJSONRawMessage(b.Metadata),
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}

type BrandUseCase struct {
	repo *repository.Queries
}

// NewBrandUseCase creates a new brand use case without repository
func NewBrandUseCase() *BrandUseCase {
	return &BrandUseCase{}
}

// SetRepository sets the repository for this request
func (uc *BrandUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// CreateBrand creates a new brand
func (uc *BrandUseCase) CreateBrand(
	ctx context.Context,
	name string,
	code string,
	isActive bool,
	metadata *json.RawMessage,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if name == "" {
		return utils.NewResponse(utils.CodeBadReq, "brand name cannot be empty", nil)
	}

	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "brand code cannot be empty", nil)
	}

	// Check if code already exists
	exists, err := uc.repo.BrandCodeExists(ctx, code)
	if err == nil && exists {
		return utils.NewResponse(utils.CodeBadReq, "brand code already exists", nil)
	}

	var metaBytes []byte
	if metadata != nil {
		metaBytes = *metadata
	} else {
		metaBytes = []byte("{}")
	}

	brand, err := uc.repo.CreateBrand(ctx, repository.CreateBrandParams{
		Name:     name,
		Code:     code,
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
		Metadata: metaBytes,
	})

	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "brand created successfully", brandToOutput(brand))
}

// CreateBrandWithDefaults creates a new brand with default active status
func (uc *BrandUseCase) CreateBrandWithDefaults(
	ctx context.Context,
	name string,
	code string,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if name == "" {
		return utils.NewResponse(utils.CodeBadReq, "brand name cannot be empty", nil)
	}

	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "brand code cannot be empty", nil)
	}

	// Check if code already exists
	exists, err := uc.repo.BrandCodeExists(ctx, code)
	if err == nil && exists {
		return utils.NewResponse(utils.CodeBadReq, "brand code already exists", nil)
	}

	brand, err := uc.repo.CreateBrandWithDefaults(ctx, repository.CreateBrandWithDefaultsParams{
		Name: name,
		Code: code,
	})

	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "brand created successfully", brandToOutput(brand))
}

// GetBrandByID retrieves a brand by ID
func (uc *BrandUseCase) GetBrandByID(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brandID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid brand id", nil)
	}

	brand, err := uc.repo.GetBrandByID(ctx, int32(brandID))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "brand not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "brand fetched successfully", brandToOutput(brand))
}

// GetBrandByCode retrieves a brand by code
func (uc *BrandUseCase) GetBrandByCode(ctx context.Context, code string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "brand code is required", nil)
	}

	brand, err := uc.repo.GetBrandByCode(ctx, code)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "brand not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "brand fetched successfully", brandToOutput(brand))
}

// ListAllBrands lists all brands without pagination
func (uc *BrandUseCase) ListAllBrands(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brands, err := uc.repo.ListAllBrands(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]BrandOutput, len(brands))
	for i := range brands {
		out[i] = brandToOutput(brands[i])
	}

	return utils.NewResponse(utils.CodeOK, "brands fetched successfully", out)
}

// ListActiveBrands lists only active brands
func (uc *BrandUseCase) ListActiveBrands(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brands, err := uc.repo.ListActiveBrands(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]BrandOutput, len(brands))
	for i := range brands {
		out[i] = brandToOutput(brands[i])
	}

	return utils.NewResponse(utils.CodeOK, "active brands fetched successfully", out)
}

// ListBrands lists brands with pagination
func (uc *BrandUseCase) ListBrands(ctx context.Context, limit, offset int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brands, err := uc.repo.ListBrands(ctx, repository.ListBrandsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]BrandOutput, len(brands))
	for i := range brands {
		out[i] = brandToOutput(brands[i])
	}

	return utils.NewResponse(utils.CodeOK, "brands fetched successfully", out)
}

// ListActiveBrandsWithPagination lists active brands with pagination
func (uc *BrandUseCase) ListActiveBrandsWithPagination(ctx context.Context, limit, offset int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brands, err := uc.repo.ListActiveBrandsWithPagination(ctx, repository.ListActiveBrandsWithPaginationParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]BrandOutput, len(brands))
	for i := range brands {
		out[i] = brandToOutput(brands[i])
	}

	return utils.NewResponse(utils.CodeOK, "active brands fetched successfully", out)
}

// CountBrands counts total number of brands
func (uc *BrandUseCase) CountBrands(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	count, err := uc.repo.CountBrands(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brand count fetched successfully", count)
}

// CountActiveBrands counts total number of active brands
func (uc *BrandUseCase) CountActiveBrands(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	count, err := uc.repo.CountActiveBrands(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "active brand count fetched successfully", count)
}

// SearchBrands searches brands by name or code
func (uc *BrandUseCase) SearchBrands(ctx context.Context, searchTerm string, limit, offset int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if searchTerm == "" {
		return utils.NewResponse(utils.CodeBadReq, "search term is required", nil)
	}

	brands, err := uc.repo.SearchBrands(ctx, repository.SearchBrandsParams{
		Lower:  "%" + searchTerm + "%",
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]BrandOutput, len(brands))
	for i := range brands {
		out[i] = brandToOutput(brands[i])
	}

	return utils.NewResponse(utils.CodeOK, "brands searched successfully", out)
}

// SearchActiveBrands searches active brands by name or code
func (uc *BrandUseCase) SearchActiveBrands(ctx context.Context, searchTerm string, limit, offset int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if searchTerm == "" {
		return utils.NewResponse(utils.CodeBadReq, "search term is required", nil)
	}

	brands, err := uc.repo.SearchActiveBrands(ctx, repository.SearchActiveBrandsParams{
		Lower:  "%" + searchTerm + "%",
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]BrandOutput, len(brands))
	for i := range brands {
		out[i] = brandToOutput(brands[i])
	}

	return utils.NewResponse(utils.CodeOK, "active brands searched successfully", out)
}

// UpdateBrand updates brand information
func (uc *BrandUseCase) UpdateBrand(
	ctx context.Context,
	id string,
	name *string,
	code *string,
	isActive *bool,
	metadata *json.RawMessage,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brandID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid brand id", nil)
	}

	// Check if brand exists
	exists, err := uc.repo.BrandExists(ctx, int32(brandID))
	if err != nil || !exists {
		return utils.NewResponse(utils.CodeNotFound, "brand not found", nil)
	}

	// If code is being updated, check if it already exists
	if code != nil {
		exists, err := uc.repo.BrandCodeExistsExcludingID(ctx, repository.BrandCodeExistsExcludingIDParams{
			Code: *code,
			ID:    int32(brandID),
		})
		if err == nil && exists {
			return utils.NewResponse(utils.CodeBadReq, "brand code already exists", nil)
		}
	}

	var metaBytes []byte
	if metadata != nil {
		metaBytes = *metadata
	} else {
		// Get current brand to preserve metadata if not provided
		currentBrand, err := uc.repo.GetBrandByID(ctx, int32(brandID))
		if err != nil {
			return utils.NewResponse(utils.CodeError, err.Error(), nil)
		}
		metaBytes = currentBrand.Metadata
	}

	// Build update params
	params := repository.UpdateBrandParams{
		ID: int32(brandID),
	}

	if name != nil {
		params.Name = *name
	} else {
		currentBrand, _ := uc.repo.GetBrandByID(ctx, int32(brandID))
		params.Name = currentBrand.Name
	}

	if code != nil {
		params.Code = *code
	} else {
		currentBrand, _ := uc.repo.GetBrandByID(ctx, int32(brandID))
		params.Code = currentBrand.Code
	}

	if isActive != nil {
		params.IsActive = pgtype.Bool{Bool: *isActive, Valid: true}
	} else {
		currentBrand, _ := uc.repo.GetBrandByID(ctx, int32(brandID))
		params.IsActive = currentBrand.IsActive
	}

	params.Metadata = metaBytes

	brand, err := uc.repo.UpdateBrand(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brand updated successfully", brandToOutput(brand))
}

// UpdateBrandName updates only the brand name
func (uc *BrandUseCase) UpdateBrandName(ctx context.Context, id string, name string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brandID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid brand id", nil)
	}

	if name == "" {
		return utils.NewResponse(utils.CodeBadReq, "brand name cannot be empty", nil)
	}

	brand, err := uc.repo.UpdateBrandName(ctx, repository.UpdateBrandNameParams{
		ID:   int32(brandID),
		Name: name,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brand name updated successfully", brandToOutput(brand))
}

// UpdateBrandCode updates only the brand code
func (uc *BrandUseCase) UpdateBrandCode(ctx context.Context, id string, code string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brandID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid brand id", nil)
	}

	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "brand code cannot be empty", nil)
	}

	// Check if code already exists
	exists, err := uc.repo.BrandCodeExistsExcludingID(ctx, repository.BrandCodeExistsExcludingIDParams{
		Code: code,
		ID:   int32(brandID),
	})
	if err == nil && exists {
		return utils.NewResponse(utils.CodeBadReq, "brand code already exists", nil)
	}

	brand, err := uc.repo.UpdateBrandCode(ctx, repository.UpdateBrandCodeParams{
		ID:   int32(brandID),
		Code: code,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brand code updated successfully", brandToOutput(brand))
}

// UpdateBrandMetadata updates only the brand metadata
func (uc *BrandUseCase) UpdateBrandMetadata(ctx context.Context, id string, metadata json.RawMessage) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brandID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid brand id", nil)
	}

	brand, err := uc.repo.UpdateBrandMetadata(ctx, repository.UpdateBrandMetadataParams{
		ID:       int32(brandID),
		Metadata: metadata,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brand metadata updated successfully", brandToOutput(brand))
}

// ActivateBrand activates a brand
func (uc *BrandUseCase) ActivateBrand(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brandID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid brand id", nil)
	}

	brand, err := uc.repo.ActivateBrand(ctx, int32(brandID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brand activated successfully", brandToOutput(brand))
}

// DeactivateBrand deactivates a brand
func (uc *BrandUseCase) DeactivateBrand(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brandID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid brand id", nil)
	}

	brand, err := uc.repo.DeactivateBrand(ctx, int32(brandID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brand deactivated successfully", brandToOutput(brand))
}

// ToggleBrandStatus toggles brand active status
func (uc *BrandUseCase) ToggleBrandStatus(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brandID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid brand id", nil)
	}

	brand, err := uc.repo.ToggleBrandStatus(ctx, int32(brandID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brand status toggled successfully", brandToOutput(brand))
}

// DeleteBrand hard deletes a brand
func (uc *BrandUseCase) DeleteBrand(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brandID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid brand id", nil)
	}

	err = uc.repo.DeleteBrand(ctx, int32(brandID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brand deleted successfully", nil)
}

// DeleteBrandByCode hard deletes a brand by code
func (uc *BrandUseCase) DeleteBrandByCode(ctx context.Context, code string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "brand code is required", nil)
	}

	err := uc.repo.DeleteBrandByCode(ctx, code)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brand deleted successfully", nil)
}

// SoftDeleteBrand soft deletes a brand by deactivating it
func (uc *BrandUseCase) SoftDeleteBrand(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brandID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid brand id", nil)
	}

	brand, err := uc.repo.SoftDeleteBrand(ctx, int32(brandID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brand soft deleted successfully", brandToOutput(brand))
}

// BrandExists checks if a brand exists by ID
func (uc *BrandUseCase) BrandExists(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brandID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid brand id", nil)
	}

	exists, err := uc.repo.BrandExists(ctx, int32(brandID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brand existence checked", exists)
}

// BrandCodeExists checks if a brand code already exists
func (uc *BrandUseCase) BrandCodeExists(ctx context.Context, code string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "brand code is required", nil)
	}

	exists, err := uc.repo.BrandCodeExists(ctx, code)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brand code existence checked", exists)
}

// GetBrandWithProductCount gets brand with count of associated products
func (uc *BrandUseCase) GetBrandWithProductCount(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brandID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid brand id", nil)
	}

	result, err := uc.repo.GetBrandWithProductCount(ctx, int32(brandID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brand with product count fetched successfully", result)
}

// ListBrandsWithProductCounts lists all brands with their product counts
func (uc *BrandUseCase) ListBrandsWithProductCounts(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brands, err := uc.repo.ListBrandsWithProductCounts(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brands with product counts fetched successfully", brands)
}

// ListActiveBrandsWithProductCounts lists active brands with their product counts
func (uc *BrandUseCase) ListActiveBrandsWithProductCounts(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brands, err := uc.repo.ListActiveBrandsWithProductCounts(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "active brands with product counts fetched successfully", brands)
}

// GetTopBrandsByProductCount gets top N brands by number of products
func (uc *BrandUseCase) GetTopBrandsByProductCount(ctx context.Context, limit int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brands, err := uc.repo.GetTopBrandsByProductCount(ctx, limit)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "top brands fetched successfully", brands)
}

// GetBrandsWithNoProducts gets brands that have no associated products
func (uc *BrandUseCase) GetBrandsWithNoProducts(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brands, err := uc.repo.GetBrandsWithNoProducts(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]BrandOutput, len(brands))
	for i := range brands {
		out[i] = brandToOutput(brands[i])
	}

	return utils.NewResponse(utils.CodeOK, "brands with no products fetched successfully", out)
}

// GetInactiveBrandsWithActiveProducts gets inactive brands that still have active products
func (uc *BrandUseCase) GetInactiveBrandsWithActiveProducts(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brands, err := uc.repo.GetInactiveBrandsWithActiveProducts(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "inactive brands with active products fetched successfully", brands)
}

// BulkActivateBrands activates multiple brands by IDs
func (uc *BrandUseCase) BulkActivateBrands(ctx context.Context, ids []int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if len(ids) == 0 {
		return utils.NewResponse(utils.CodeBadReq, "at least one brand id is required", nil)
	}

	err := uc.repo.BulkActivateBrands(ctx, ids)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brands activated successfully", nil)
}

// BulkDeactivateBrands deactivates multiple brands by IDs
func (uc *BrandUseCase) BulkDeactivateBrands(ctx context.Context, ids []int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if len(ids) == 0 {
		return utils.NewResponse(utils.CodeBadReq, "at least one brand id is required", nil)
	}

	err := uc.repo.BulkDeactivateBrands(ctx, ids)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brands deactivated successfully", nil)
}

// BulkDeleteBrands deletes multiple brands by IDs
func (uc *BrandUseCase) BulkDeleteBrands(ctx context.Context, ids []int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	if len(ids) == 0 {
		return utils.NewResponse(utils.CodeBadReq, "at least one brand id is required", nil)
	}

	err := uc.repo.BulkDeleteBrands(ctx, ids)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brands deleted successfully", nil)
}

// GetRecentlyCreatedBrands gets brands created in the last N days
func (uc *BrandUseCase) GetRecentlyCreatedBrands(ctx context.Context, days int) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	interval := pgtype.Interval{
		Microseconds: int64(days) * 24 * 60 * 60 * 1000000,
		Valid:        true,
	}

	brands, err := uc.repo.GetRecentlyCreatedBrands(ctx, interval)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]BrandOutput, len(brands))
	for i := range brands {
		out[i] = brandToOutput(brands[i])
	}

	return utils.NewResponse(utils.CodeOK, "recently created brands fetched successfully", out)
}

// GetRecentlyUpdatedBrands gets brands updated in the last N days
func (uc *BrandUseCase) GetRecentlyUpdatedBrands(ctx context.Context, days int) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	interval := pgtype.Interval{
		Microseconds: int64(days) * 24 * 60 * 60 * 1000000,
		Valid:        true,
	}

	brands, err := uc.repo.GetRecentlyUpdatedBrands(ctx, interval)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]BrandOutput, len(brands))
	for i := range brands {
		out[i] = brandToOutput(brands[i])
	}

	return utils.NewResponse(utils.CodeOK, "recently updated brands fetched successfully", out)
}

// GetBrandsByCreationDate gets brands created between two dates
func (uc *BrandUseCase) GetBrandsByCreationDate(ctx context.Context, startDate, endDate time.Time) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brands, err := uc.repo.GetBrandsByCreationDate(ctx, repository.GetBrandsByCreationDateParams{
		CreatedAt:   pgtype.Timestamp{Time: startDate, Valid: true},
		CreatedAt_2: pgtype.Timestamp{Time: endDate, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]BrandOutput, len(brands))
	for i := range brands {
		out[i] = brandToOutput(brands[i])
	}

	return utils.NewResponse(utils.CodeOK, "brands by creation date fetched successfully", out)
}

// GetBrandMetadataByKey gets a specific metadata field from a brand
func (uc *BrandUseCase) GetBrandMetadataByKey(ctx context.Context, id string, key string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	brandID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid brand id", nil)
	}

	if key == "" {
		return utils.NewResponse(utils.CodeBadReq, "metadata key is required", nil)
	}

	// The Metadata field expects the key as JSONB bytes
	keyBytes := []byte(`"` + key + `"`)

	result, err := uc.repo.GetBrandMetadataByKey(ctx, repository.GetBrandMetadataByKeyParams{
		ID:       int32(brandID),
		Metadata: keyBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brand metadata fetched successfully", result)
}

// ListBrandsWithStats lists brands with statistics
func (uc *BrandUseCase) ListBrandsWithStats(ctx context.Context, activeOnly bool, search *string, limit, offset int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	var searchStr string
	if search != nil && *search != "" {
		searchStr = *search
	}

	var activeOnlyVal pgtype.Bool
	activeOnlyVal = pgtype.Bool{Bool: activeOnly, Valid: true}

	brands, err := uc.repo.ListBrandsWithStats(ctx, repository.ListBrandsWithStatsParams{
		ActiveOnly: activeOnlyVal,
		Search:     searchStr,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "brands with stats fetched successfully", brands)
}
