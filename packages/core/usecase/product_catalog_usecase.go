package usecase

import (
	"context"
	"strconv"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
)

// ProductCatalogUseCase handles the admin product catalog (products + embedded variants).
type ProductCatalogUseCase struct {
	repo *repository.Queries
}

// NewProductCatalogUseCase creates a new ProductCatalogUseCase.
func NewProductCatalogUseCase() *ProductCatalogUseCase {
	return &ProductCatalogUseCase{}
}

// SetRepository sets the tenant-scoped repository for this request.
func (uc *ProductCatalogUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *ProductCatalogUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

// ListProductsWithVariants returns all master products with their variants embedded as a JSON array.
// orgIDStr is required. categoryIDStr is optional — pass "" or "0" to fetch all categories.
func (uc *ProductCatalogUseCase) ListProductsWithVariants(
	ctx context.Context,
	orgIDStr string,
	categoryIDStr string,
	limit int32,
	offset int32,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil || orgID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid or missing organization_id", nil)
	}

	// Use 0 as the sentinel for "no category filter" (SQL: $2 = 0 OR p.category_id = $2).
	// Category IDs are always >= 1 so 0 is safe.
	var categoryID int32
	if categoryIDStr != "" && categoryIDStr != "0" {
		catID, err := strconv.ParseInt(categoryIDStr, 10, 32)
		if err != nil || catID < 0 {
			return utils.NewResponse(utils.CodeBadReq, "invalid category_id", nil)
		}
		categoryID = int32(catID)
	}

	if limit <= 0 {
		limit = 20
	}

	params := repository.ListProductsWithVariantsParams{
		OrganizationID: int32(orgID),
		Column2:        categoryID, // 0 = no filter, >0 = filter by that category
		Limit:          limit,
		Offset:         offset,
	}

	products, err := uc.repo.ListProductsWithVariants(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch product catalog", err.Error())
	}

	return utils.NewResponse(utils.CodeOK, "products fetched successfully", products)
}

// GetMasterProductCatalog returns the detailed master catalog for an organization with pagination.
func (uc *ProductCatalogUseCase) GetMasterProductCatalog(
	ctx context.Context,
	orgIDStr string,
	limit int32,
	offset int32,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil || orgID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid or missing organization_id", nil)
	}

	if limit <= 0 {
		limit = 100 // default 100 per page
	}
	if offset < 0 {
		offset = 0
	}

	// 1. Get total count
	totalCount, err := uc.repo.GetMasterProductCatalogCount(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch master product catalog count", err.Error())
	}

	// 2. Get catalog products
	params := repository.GetMasterProductCatalogParams{
		OrganizationID: int32(orgID),
		Limit:          limit,
		Offset:         offset,
	}
	catalog, err := uc.repo.GetMasterProductCatalog(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch master product catalog", err.Error())
	}

	// 3. Compute total pages
	totalPages := int32(0)
	if limit > 0 {
		totalPages = int32((totalCount + int64(limit) - 1) / int64(limit))
	}

	// 4. Return paginated response
	type PaginatedCatalog struct {
		TotalCount int64                                   `json:"total_count"`
		TotalPages int32                                   `json:"total_pages"`
		Page       int32                                   `json:"page"`
		Limit      int32                                   `json:"limit"`
		Data       []repository.GetMasterProductCatalogRow `json:"data"`
	}

	currentPage := (offset / limit) + 1

	respData := PaginatedCatalog{
		TotalCount: totalCount,
		TotalPages: totalPages,
		Page:       currentPage,
		Limit:      limit,
		Data:       catalog,
	}

	return utils.NewResponse(utils.CodeOK, "master product catalog fetched successfully", respData)
}
