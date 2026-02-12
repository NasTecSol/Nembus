package usecase

import (
	"context"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type ProductBarcodeUseCase struct {
	repo *repository.Queries
}

func NewProductBarcodeUseCase() *ProductBarcodeUseCase {
	return &ProductBarcodeUseCase{}
}

func (uc *ProductBarcodeUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *ProductBarcodeUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

func (uc *ProductBarcodeUseCase) CreateProductBarcode(ctx context.Context, arg repository.CreateProductBarcodeParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	// Check if barcode already exists
	exists, err := uc.repo.CheckBarcodeExists(ctx, arg.Barcode)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to check barcode existence", err.Error())
	}
	if exists {
		return utils.NewResponse(utils.CodeBadReq, "barcode already exists", nil)
	}

	barcode, err := uc.repo.CreateProductBarcode(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create product barcode", err.Error())
	}
	return utils.NewResponse(utils.CodeCreated, "product barcode created successfully", barcode)
}

func (uc *ProductBarcodeUseCase) GetProductBarcode(ctx context.Context, id int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	barcode, err := uc.repo.GetProductBarcode(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "product barcode not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "product barcode fetched successfully", barcode)
}

func (uc *ProductBarcodeUseCase) GetProductByBarcode(ctx context.Context, barcodeStr string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	product, err := uc.repo.GetProductByBarcode(ctx, barcodeStr)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "product with barcode not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "product fetched successfully", product)
}

func (uc *ProductBarcodeUseCase) ListProductBarcodes(ctx context.Context) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	barcodes, err := uc.repo.ListProductBarcodes(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list product barcodes", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "product barcodes listed successfully", barcodes)
}

func (uc *ProductBarcodeUseCase) ListProductBarcodesByProduct(ctx context.Context, productID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	barcodes, err := uc.repo.ListProductBarcodesByProduct(ctx, productID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list barcodes for product", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "product barcodes listed successfully", barcodes)
}

func (uc *ProductBarcodeUseCase) ListProductBarcodesByVariant(ctx context.Context, productVariantID pgtype.Int4) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	barcodes, err := uc.repo.ListProductBarcodesByVariant(ctx, productVariantID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to list barcodes for variant", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "product barcodes listed successfully", barcodes)
}

func (uc *ProductBarcodeUseCase) UpdateProductBarcode(ctx context.Context, arg repository.UpdateProductBarcodeParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	barcode, err := uc.repo.UpdateProductBarcode(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to update product barcode", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "product barcode updated successfully", barcode)
}

func (uc *ProductBarcodeUseCase) SetPrimaryBarcode(ctx context.Context, arg repository.SetPrimaryBarcodeParams) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	err := uc.repo.SetPrimaryBarcode(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to set primary barcode", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "primary barcode set successfully", nil)
}

func (uc *ProductBarcodeUseCase) GetPrimaryBarcode(ctx context.Context, productID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	barcode, err := uc.repo.GetPrimaryBarcode(ctx, productID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "primary barcode not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "primary barcode fetched successfully", barcode)
}

func (uc *ProductBarcodeUseCase) DeleteProductBarcode(ctx context.Context, id int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	err := uc.repo.DeleteProductBarcode(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to delete product barcode", err.Error())
	}
	return utils.NewResponse(utils.CodeOK, "product barcode deleted successfully", nil)
}
