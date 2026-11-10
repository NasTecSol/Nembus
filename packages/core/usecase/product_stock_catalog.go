package usecase

import (
	"context"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
)

// GetMasterProductStockCatalog returns full products and pagination for the same stock filters.
func (uc *ProductCatalogUseCase) GetMasterProductStockCatalog(ctx context.Context, params repository.GetMasterProductStockCatalogParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if params.Limit <= 0 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	total, err := uc.repo.GetMasterProductStockCatalogCount(ctx, repository.GetMasterProductStockCatalogCountParams{
		OrganizationID:    params.OrganizationID,
		CategoryID:        params.CategoryID,
		StoreID:           params.StoreID,
		StorageLocationID: params.StorageLocationID,
		StockStatus:       params.StockStatus,
		LowStockThreshold: params.LowStockThreshold,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch stock catalog count", err.Error())
	}
	products, err := uc.repo.GetMasterProductStockCatalog(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch stock catalog", err.Error())
	}
	if products == nil {
		products = []repository.GetMasterProductStockCatalogRow{}
	}
	data := struct {
		TotalCount int64                                        `json:"total_count"`
		TotalPages int64                                        `json:"total_pages"`
		Page       int64                                        `json:"page"`
		Limit      int32                                        `json:"limit"`
		Data       []repository.GetMasterProductStockCatalogRow `json:"data"`
	}{
		TotalCount: total,
		TotalPages: (total + int64(params.Limit) - 1) / int64(params.Limit),
		Page:       int64(params.Offset)/int64(params.Limit) + 1,
		Limit:      params.Limit,
		Data:       products,
	}
	return utils.NewResponse(utils.CodeOK, "master product stock catalog fetched successfully", data)
}
