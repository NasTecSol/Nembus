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
