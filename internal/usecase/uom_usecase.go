package usecase

import (
	"context"
	"encoding/json"
	"strconv"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// UnitOfMeasureOutput represents a unit of measure with JSON-friendly metadata.
type UnitOfMeasureOutput struct {
	ID            int32            `json:"id"`
	Code          string           `json:"code"`
	Name          string           `json:"name"`
	UomType       pgtype.Text      `json:"uom_type"`
	DecimalPlaces pgtype.Int4      `json:"decimal_places"`
	IsActive      pgtype.Bool      `json:"is_active"`
	Metadata      json.RawMessage  `json:"metadata"`
}

// ProductUOMConversionOutput represents a product UOM conversion with JSON-friendly metadata.
type ProductUOMConversionOutput struct {
	ID               int32            `json:"id"`
	ProductID        int32            `json:"product_id"`
	FromUomID        int32            `json:"from_uom_id"`
	ToUomID          int32            `json:"to_uom_id"`
	ConversionFactor pgtype.Numeric   `json:"conversion_factor"`
	IsDefault        pgtype.Bool      `json:"is_default"`
	Metadata         json.RawMessage  `json:"metadata"`
	CreatedAt        pgtype.Timestamp `json:"created_at"`
}

func unitToOutput(u repository.UnitsOfMeasure) UnitOfMeasureOutput {
	return UnitOfMeasureOutput{
		ID:            u.ID,
		Code:          u.Code,
		Name:          u.Name,
		UomType:       u.UomType,
		DecimalPlaces: u.DecimalPlaces,
		IsActive:      u.IsActive,
		Metadata:      utils.BytesToJSONRawMessage(u.Metadata),
	}
}

func conversionToOutput(c repository.ProductUomConversion) ProductUOMConversionOutput {
	return ProductUOMConversionOutput{
		ID:               c.ID,
		ProductID:        c.ProductID,
		FromUomID:        c.FromUomID,
		ToUomID:          c.ToUomID,
		ConversionFactor: c.ConversionFactor,
		IsDefault:        c.IsDefault,
		Metadata:         utils.BytesToJSONRawMessage(c.Metadata),
		CreatedAt:        c.CreatedAt,
	}
}

type UOMUseCase struct {
	repo *repository.Queries
}

func NewUOMUseCase() *UOMUseCase {
	return &UOMUseCase{}
}

func (uc *UOMUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *UOMUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

// CreateUnitOfMeasure creates a new UOM.
func (uc *UOMUseCase) CreateUnitOfMeasure(
	ctx context.Context,
	code string,
	name string,
	uomType string,
	decimalPlaces int32,
	isActive bool,
	metadata any,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "code is required", nil)
	}
	if name == "" {
		return utils.NewResponse(utils.CodeBadReq, "name is required", nil)
	}

	var uomTypeText pgtype.Text
	if uomType != "" {
		uomTypeText = pgtype.Text{String: uomType, Valid: true}
	}

	var decimal pgtype.Int4
	decimal = pgtype.Int4{Int32: decimalPlaces, Valid: true}

	active := pgtype.Bool{Bool: isActive, Valid: true}

	metaBytes := []byte("{}")
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			metaBytes = b
		}
	}

	row, err := uc.repo.CreateUnitOfMeasure(ctx, repository.CreateUnitOfMeasureParams{
		Code:          code,
		Name:          name,
		UomType:       uomTypeText,
		DecimalPlaces: decimal,
		IsActive:      active,
		Metadata:      metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "unit of measure created", unitToOutput(row))
}

// GetUnitOfMeasure gets a UOM by ID.
func (uc *UOMUseCase) GetUnitOfMeasure(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	row, err := uc.repo.GetUnitOfMeasure(ctx, int32(parsed))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "unit of measure not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "unit of measure fetched", unitToOutput(row))
}

// GetUnitOfMeasureByCode gets a UOM by code.
func (uc *UOMUseCase) GetUnitOfMeasureByCode(ctx context.Context, code string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "code is required", nil)
	}

	row, err := uc.repo.GetUnitOfMeasureByCode(ctx, code)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "unit of measure not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "unit of measure fetched", unitToOutput(row))
}

// ListUnitsOfMeasure lists all UOMs.
func (uc *UOMUseCase) ListUnitsOfMeasure(ctx context.Context) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	rows, err := uc.repo.ListUnitsOfMeasure(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]UnitOfMeasureOutput, len(rows))
	for i := range rows {
		out[i] = unitToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "units of measure listed", out)
}

// ListActiveUnitsOfMeasure lists only active UOMs.
func (uc *UOMUseCase) ListActiveUnitsOfMeasure(ctx context.Context) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	rows, err := uc.repo.ListActiveUnitsOfMeasure(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]UnitOfMeasureOutput, len(rows))
	for i := range rows {
		out[i] = unitToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "active units of measure listed", out)
}

// ListUnitsByType lists UOMs filtered by type.
func (uc *UOMUseCase) ListUnitsByType(ctx context.Context, uomType string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	if uomType == "" {
		return utils.NewResponse(utils.CodeBadReq, "uom_type is required", nil)
	}

	rows, err := uc.repo.ListUnitsByType(ctx, pgtype.Text{String: uomType, Valid: true})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]UnitOfMeasureOutput, len(rows))
	for i := range rows {
		out[i] = unitToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "units of measure listed by type", out)
}

// UpdateUnitOfMeasure updates a UOM by ID.
func (uc *UOMUseCase) UpdateUnitOfMeasure(
	ctx context.Context,
	id string,
	name string,
	uomType string,
	decimalPlaces int32,
	isActive bool,
	metadata any,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	if name == "" {
		return utils.NewResponse(utils.CodeBadReq, "name is required", nil)
	}

	var uomTypeText pgtype.Text
	if uomType != "" {
		uomTypeText = pgtype.Text{String: uomType, Valid: true}
	}

	decimal := pgtype.Int4{Int32: decimalPlaces, Valid: true}
	active := pgtype.Bool{Bool: isActive, Valid: true}

	metaBytes := []byte("{}")
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			metaBytes = b
		}
	}

	row, err := uc.repo.UpdateUnitOfMeasure(ctx, repository.UpdateUnitOfMeasureParams{
		ID:            int32(parsed),
		Name:          name,
		UomType:       uomTypeText,
		DecimalPlaces: decimal,
		IsActive:      active,
		Metadata:      metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "unit of measure updated", unitToOutput(row))
}

// DeleteUnitOfMeasure deletes a UOM by ID.
func (uc *UOMUseCase) DeleteUnitOfMeasure(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	if err := uc.repo.DeleteUnitOfMeasure(ctx, int32(parsed)); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "unit of measure deleted", nil)
}

// CreateProductUOMConversion creates a product UOM conversion.
func (uc *UOMUseCase) CreateProductUOMConversion(
	ctx context.Context,
	productID int32,
	fromUomID int32,
	toUomID int32,
	conversionFactor string,
	isDefault bool,
	metadata any,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	if productID == 0 || fromUomID == 0 || toUomID == 0 {
		return utils.NewResponse(utils.CodeBadReq, "product_id, from_uom_id and to_uom_id are required", nil)
	}

	var factor pgtype.Numeric
	if err := factor.Scan(conversionFactor); err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid conversion_factor", nil)
	}

	active := pgtype.Bool{Bool: isDefault, Valid: true}

	metaBytes := []byte("{}")
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			metaBytes = b
		}
	}

	row, err := uc.repo.CreateProductUOMConversion(ctx, repository.CreateProductUOMConversionParams{
		ProductID:        productID,
		FromUomID:        fromUomID,
		ToUomID:          toUomID,
		ConversionFactor: factor,
		IsDefault:        active,
		Metadata:         metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "product UOM conversion created", conversionToOutput(row))
}

// GetProductUOMConversion retrieves a specific product UOM conversion.
func (uc *UOMUseCase) GetProductUOMConversion(
	ctx context.Context,
	productID int32,
	fromUomID int32,
	toUomID int32,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	row, err := uc.repo.GetProductUOMConversion(ctx, repository.GetProductUOMConversionParams{
		ProductID: productID,
		FromUomID: fromUomID,
		ToUomID:   toUomID,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "product UOM conversion not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "product UOM conversion fetched", conversionToOutput(row))
}

// ListProductUOMConversions lists conversions for a product.
func (uc *UOMUseCase) ListProductUOMConversions(ctx context.Context, productID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	if productID == 0 {
		return utils.NewResponse(utils.CodeBadReq, "product_id is required", nil)
	}

	rows, err := uc.repo.ListProductUOMConversions(ctx, productID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]ProductUOMConversionOutput, len(rows))
	for i := range rows {
		out[i] = conversionToOutput(rows[i])
	}

	return utils.NewResponse(utils.CodeOK, "product UOM conversions listed", out)
}

// UpdateProductUOMConversion updates a product UOM conversion.
func (uc *UOMUseCase) UpdateProductUOMConversion(
	ctx context.Context,
	id string,
	conversionFactor string,
	isDefault bool,
	metadata any,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	var factor pgtype.Numeric
	if err := factor.Scan(conversionFactor); err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid conversion_factor", nil)
	}

	active := pgtype.Bool{Bool: isDefault, Valid: true}

	metaBytes := []byte("{}")
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			metaBytes = b
		}
	}

	row, err := uc.repo.UpdateProductUOMConversion(ctx, repository.UpdateProductUOMConversionParams{
		ID:               int32(parsed),
		ConversionFactor: factor,
		IsDefault:        active,
		Metadata:         metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "product UOM conversion updated", conversionToOutput(row))
}

// DeleteProductUOMConversion deletes a product UOM conversion.
func (uc *UOMUseCase) DeleteProductUOMConversion(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	if err := uc.repo.DeleteProductUOMConversion(ctx, int32(parsed)); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "product UOM conversion deleted", nil)
}

