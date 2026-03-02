package usecase

import (
	"context"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"
)

type ProductVariantUseCase struct {
	repo *repository.Queries
}

func NewProductVariantUseCase() *ProductVariantUseCase {
	return &ProductVariantUseCase{}
}

func (uc *ProductVariantUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *ProductVariantUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

// CreateProductVariant creates a new product variant.
func (uc *ProductVariantUseCase) CreateProductVariant(ctx context.Context, arg repository.CreateProductVariantParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	variant, err := uc.repo.CreateProductVariant(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create product variant", err.Error())
	}

	return utils.NewResponse(utils.CodeCreated, "product variant created successfully", variant)
}

// GetProductVariant retrieves a product variant by its ID.
func (uc *ProductVariantUseCase) GetProductVariant(ctx context.Context, id int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	variant, err := uc.repo.GetProductVariant(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "product variant not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "product variant fetched successfully", variant)
}

// GetProductVariantBySKU retrieves a product variant by SKU.
func (uc *ProductVariantUseCase) GetProductVariantBySKU(ctx context.Context, sku string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	variant, err := uc.repo.GetProductVariantBySKU(ctx, sku)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "product variant not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "product variant fetched successfully", variant)
}

// ListProductVariants lists all product variants.
func (uc *ProductVariantUseCase) ListProductVariants(ctx context.Context) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	variants, err := uc.repo.ListProductVariants(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list product variants", err.Error())
	}

	return utils.NewResponse(utils.CodeOK, "product variants fetched successfully", variants)
}

// ListProductVariantsByProduct lists all variants of a specific product.
func (uc *ProductVariantUseCase) ListProductVariantsByProduct(ctx context.Context, productID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	variants, err := uc.repo.ListProductVariantsByProduct(ctx, productID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list product variants by product", err.Error())
	}

	return utils.NewResponse(utils.CodeOK, "product variants fetched successfully", variants)
}

// ListActiveProductVariantsByProduct lists all active variants of a specific product.
func (uc *ProductVariantUseCase) ListActiveProductVariantsByProduct(ctx context.Context, productID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	variants, err := uc.repo.ListActiveProductVariantsByProduct(ctx, productID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list active product variants", err.Error())
	}

	return utils.NewResponse(utils.CodeOK, "active product variants fetched successfully", variants)
}

// UpdateProductVariant updates an existing product variant.
func (uc *ProductVariantUseCase) UpdateProductVariant(ctx context.Context, arg repository.UpdateProductVariantParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	variant, err := uc.repo.UpdateProductVariant(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update product variant", err.Error())
	}

	return utils.NewResponse(utils.CodeOK, "product variant updated successfully", variant)
}

// DeleteProductVariant deletes a product variant by ID.
func (uc *ProductVariantUseCase) DeleteProductVariant(ctx context.Context, id int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	if err := uc.repo.DeleteProductVariant(ctx, id); err != nil {
		return utils.NewResponse(utils.CodeError, "failed to delete product variant", err.Error())
	}

	return utils.NewResponse(utils.CodeOK, "product variant deleted successfully", nil)
}

// ToggleProductVariantActive sets the active status of a variant.
func (uc *ProductVariantUseCase) ToggleProductVariantActive(ctx context.Context, arg repository.ToggleProductVariantActiveParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	variant, err := uc.repo.ToggleProductVariantActive(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to toggle product variant active status", err.Error())
	}

	return utils.NewResponse(utils.CodeOK, "product variant active status updated successfully", variant)
}

// SearchProductVariants searches product variants by SKU, name, or product name.
func (uc *ProductVariantUseCase) SearchProductVariants(ctx context.Context, arg repository.SearchProductVariantsParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	variants, err := uc.repo.SearchProductVariants(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to search product variants", err.Error())
	}

	return utils.NewResponse(utils.CodeOK, "product variants fetched successfully", variants)
}
