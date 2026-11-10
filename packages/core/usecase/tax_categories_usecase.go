package usecase

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// TaxCategoryOutput represents a tax category with JSON-friendly metadata.
type TaxCategoryOutput struct {
	ID          int32            `json:"id"`
	Name        string           `json:"name"`
	Code        string           `json:"code"`
	TaxRate     pgtype.Numeric   `json:"tax_rate"`
	IsInclusive pgtype.Bool      `json:"is_inclusive"`
	IsActive    pgtype.Bool      `json:"is_active"`
	Metadata    json.RawMessage  `json:"metadata"`
	CreatedAt   pgtype.Timestamp `json:"created_at"`
}

func taxCategoryToOutput(t repository.TaxCategory) TaxCategoryOutput {
	return TaxCategoryOutput{
		ID:          t.ID,
		Name:        t.Name,
		Code:        t.Code,
		TaxRate:     t.TaxRate,
		IsInclusive: t.IsInclusive,
		IsActive:    t.IsActive,
		Metadata:    utils.BytesToJSONRawMessage(t.Metadata),
		CreatedAt:   t.CreatedAt,
	}
}

type TaxCategoriesUseCase struct {
	repo *repository.Queries
}

func NewTaxCategoriesUseCase() *TaxCategoriesUseCase {
	return &TaxCategoriesUseCase{}
}

func (uc *TaxCategoriesUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *TaxCategoriesUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

// CreateTaxCategory creates a new tax category.
func (uc *TaxCategoriesUseCase) CreateTaxCategory(
	ctx context.Context,
	name string,
	code string,
	taxRate string,
	isInclusive bool,
	isActive bool,
	metadata any,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	if name == "" {
		return utils.NewResponse(utils.CodeBadReq, "name is required", nil)
	}
	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "code is required", nil)
	}
	if taxRate == "" {
		return utils.NewResponse(utils.CodeBadReq, "tax_rate is required", nil)
	}

	var numeric pgtype.Numeric
	if err := numeric.Scan(taxRate); err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid tax_rate", nil)
	}

	inclusive := pgtype.Bool{Bool: isInclusive, Valid: true}
	active := pgtype.Bool{Bool: isActive, Valid: true}

	metaBytes := []byte("{}")
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			metaBytes = b
		}
	}

	row, err := uc.repo.CreateTaxCategory(ctx, repository.CreateTaxCategoryParams{
		Name:        name,
		Code:        code,
		TaxRate:     numeric,
		IsInclusive: inclusive,
		IsActive:    active,
		Metadata:    metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "tax category created", taxCategoryToOutput(row))
}

// GetTaxCategory retrieves a tax category by ID.
func (uc *TaxCategoriesUseCase) GetTaxCategory(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	row, err := uc.repo.GetTaxCategory(ctx, int32(parsed))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "tax category not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "tax category fetched", taxCategoryToOutput(row))
}

// GetTaxCategoryByCode retrieves a tax category by code.
func (uc *TaxCategoriesUseCase) GetTaxCategoryByCode(ctx context.Context, code string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "code is required", nil)
	}

	row, err := uc.repo.GetTaxCategoryByCode(ctx, code)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "tax category not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "tax category fetched", taxCategoryToOutput(row))
}

// ListTaxCategories lists all tax categories.
func (uc *TaxCategoriesUseCase) ListTaxCategories(ctx context.Context) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	rows, err := uc.repo.ListTaxCategories(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]TaxCategoryOutput, len(rows))
	for i := range rows {
		out[i] = taxCategoryToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "tax categories listed", out)
}

// ListActiveTaxCategories lists only active tax categories.
func (uc *TaxCategoriesUseCase) ListActiveTaxCategories(ctx context.Context) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	rows, err := uc.repo.ListActiveTaxCategories(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]TaxCategoryOutput, len(rows))
	for i := range rows {
		out[i] = taxCategoryToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "active tax categories listed", out)
}

// UpdateTaxCategory updates an existing tax category.
func (uc *TaxCategoriesUseCase) UpdateTaxCategory(
	ctx context.Context,
	id string,
	name string,
	taxRate string,
	isInclusive bool,
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
	if taxRate == "" {
		return utils.NewResponse(utils.CodeBadReq, "tax_rate is required", nil)
	}

	var numeric pgtype.Numeric
	if err := numeric.Scan(taxRate); err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid tax_rate", nil)
	}

	inclusive := pgtype.Bool{Bool: isInclusive, Valid: true}
	active := pgtype.Bool{Bool: isActive, Valid: true}

	metaBytes := []byte("{}")
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			metaBytes = b
		}
	}

	row, err := uc.repo.UpdateTaxCategory(ctx, repository.UpdateTaxCategoryParams{
		ID:          int32(parsed),
		Name:        name,
		TaxRate:     numeric,
		IsInclusive: inclusive,
		IsActive:    active,
		Metadata:    metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "tax category updated", taxCategoryToOutput(row))
}

// DeleteTaxCategory deletes a tax category by ID.
func (uc *TaxCategoriesUseCase) DeleteTaxCategory(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	if err := uc.repo.DeleteTaxCategory(ctx, int32(parsed)); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "tax category deleted", nil)
}

// ToggleTaxCategoryActive toggles the active flag for a tax category.
func (uc *TaxCategoriesUseCase) ToggleTaxCategoryActive(ctx context.Context, id string, isActive bool) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	row, err := uc.repo.ToggleTaxCategoryActive(ctx, repository.ToggleTaxCategoryActiveParams{
		ID:       int32(parsed),
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "tax category active flag updated", taxCategoryToOutput(row))
}

