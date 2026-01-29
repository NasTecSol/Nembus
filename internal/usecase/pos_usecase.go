package usecase

import (
	"context"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type POSUseCase struct {
}

func NewPOSUseCase() *POSUseCase {
	return &POSUseCase{}
}

// GetPOSProducts fetches products with stock for a specific store
func (uc *POSUseCase) GetPOSProducts(ctx context.Context, repo *repository.Queries, storeID int32, categoryID *int32, searchTerm string, includeOutOfStock bool) *repository.Response {
	var pgCategoryID pgtype.Int4
	if categoryID != nil {
		pgCategoryID = pgtype.Int4{Int32: *categoryID, Valid: true}
	}

	products, err := repo.GetPOSProductsWithStock(ctx, repository.GetPOSProductsWithStockParams{
		StoreID:           storeID,
		CategoryID:        pgCategoryID,
		SearchTerm:        pgtype.Text{String: searchTerm, Valid: searchTerm != ""},
		IncludeOutOfStock: includeOutOfStock,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "products fetched successfully", products)
}

// GetPOSProductByBarcode fetches a product by barcode for a specific store
func (uc *POSUseCase) GetPOSProductByBarcode(ctx context.Context, repo *repository.Queries, barcode string, storeID int32) *repository.Response {
	product, err := repo.GetPOSProductByBarcode(ctx, repository.GetPOSProductByBarcodeParams{
		Barcode: pgtype.Text{String: barcode, Valid: true},
		StoreID: storeID,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "product fetched successfully", product)
}

// GetPOSProductByID fetches a product by ID for a specific store
func (uc *POSUseCase) GetPOSProductByID(ctx context.Context, repo *repository.Queries, id int32, storeID int32) *repository.Response {
	product, err := repo.GetPOSProductByID(ctx, repository.GetPOSProductByIDParams{
		ProductID: id,
		StoreID:   storeID,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "product fetched successfully", product)
}

// SearchPOSProducts searches for products by name, SKU, barcode or ID
func (uc *POSUseCase) SearchPOSProducts(ctx context.Context, repo *repository.Queries, searchTerm string, storeID int32, limit int32) *repository.Response {
	if limit <= 0 {
		limit = 50
	}

	products, err := repo.SearchPOSProducts(ctx, repository.SearchPOSProductsParams{
		SearchTerm: searchTerm,
		StoreID:    storeID,
		LimitCount: limit,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "products searched successfully", products)
}

// GetPOSCategories fetches all categories with product counts
func (uc *POSUseCase) GetPOSCategories(ctx context.Context, repo *repository.Queries) *repository.Response {
	categories, err := repo.GetPOSCategories(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "categories fetched successfully", categories)
}

// GetPOSPromotedProducts fetches all products on promotion for a store
func (uc *POSUseCase) GetPOSPromotedProducts(ctx context.Context, repo *repository.Queries, storeID *int32) *repository.Response {
	var pgStoreID pgtype.Int4
	if storeID != nil {
		pgStoreID = pgtype.Int4{Int32: *storeID, Valid: true}
	}

	products, err := repo.GetPOSPromotedProducts(ctx, pgStoreID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "promoted products fetched successfully", products)
}

// CreateProduct creates a new product
func (uc *POSUseCase) CreateProduct(ctx context.Context, repo *repository.Queries, arg repository.CreateProductParams) *repository.Response {
	product, err := repo.CreateProduct(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "product created successfully", product)
}
