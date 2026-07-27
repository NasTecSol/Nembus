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

// PriceListOutput represents a price list with JSON-friendly metadata.
type PriceListOutput struct {
	ID            int32            `json:"id"`
	Name          string           `json:"name"`
	Code          string           `json:"code"`
	PriceListType pgtype.Text      `json:"price_list_type"`
	CurrencyCode  pgtype.Text      `json:"currency_code"`
	ValidFrom     pgtype.Date      `json:"valid_from"`
	ValidTo       pgtype.Date      `json:"valid_to"`
	IsDefault     pgtype.Bool      `json:"is_default"`
	IsActive      pgtype.Bool      `json:"is_active"`
	Metadata      json.RawMessage  `json:"metadata"`
	CreatedAt     pgtype.Timestamp `json:"created_at"`
	UpdatedAt     pgtype.Timestamp `json:"updated_at"`
}

func priceListToOutput(p repository.PriceList) PriceListOutput {
	return PriceListOutput{
		ID:            p.ID,
		Name:          p.Name,
		Code:          p.Code,
		PriceListType: p.PriceListType,
		CurrencyCode:  p.CurrencyCode,
		ValidFrom:     p.ValidFrom,
		ValidTo:       p.ValidTo,
		IsDefault:     p.IsDefault,
		IsActive:      p.IsActive,
		Metadata:      utils.BytesToJSONRawMessage(p.Metadata),
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

type PriceListsUseCase struct {
	repo *repository.Queries
}

func NewPriceListsUseCase() *PriceListsUseCase {
	return &PriceListsUseCase{}
}

func (uc *PriceListsUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *PriceListsUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

// CreatePriceList creates a new price list.
func (uc *PriceListsUseCase) CreatePriceList(
	ctx context.Context,
	name string,
	code string,
	priceListType *string,
	currencyCode *string,
	validFrom *string,
	validTo *string,
	isDefault bool,
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

	var plt pgtype.Text
	if priceListType != nil {
		plt = pgtype.Text{String: *priceListType, Valid: true}
	}

	var curr pgtype.Text
	if currencyCode != nil {
		curr = pgtype.Text{String: *currencyCode, Valid: true}
	}

	var fromDate pgtype.Date
	if validFrom != nil && *validFrom != "" {
		parsed, err := time.Parse("2006-01-02", *validFrom)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid valid_from", nil)
		}
		fromDate = pgtype.Date{Time: parsed, Valid: true}
	}

	var toDate pgtype.Date
	if validTo != nil && *validTo != "" {
		parsed, err := time.Parse("2006-01-02", *validTo)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid valid_to", nil)
		}
		toDate = pgtype.Date{Time: parsed, Valid: true}
	}

	def := pgtype.Bool{Bool: isDefault, Valid: true}
	active := pgtype.Bool{Bool: isActive, Valid: true}

	metaBytes := []byte("{}")
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			metaBytes = b
		}
	}

	row, err := uc.repo.CreatePriceList(ctx, repository.CreatePriceListParams{
		Name:          name,
		Code:          code,
		PriceListType: plt,
		CurrencyCode:  curr,
		ValidFrom:     fromDate,
		ValidTo:       toDate,
		IsDefault:     def,
		IsActive:      active,
		Metadata:      metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "price list created", priceListToOutput(row))
}

// GetPriceList gets a price list by ID.
func (uc *PriceListsUseCase) GetPriceList(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	row, err := uc.repo.GetPriceList(ctx, int32(parsed))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "price list not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "price list fetched", priceListToOutput(row))
}

// GetPriceListByCode gets a price list by code.
func (uc *PriceListsUseCase) GetPriceListByCode(ctx context.Context, code string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "code is required", nil)
	}

	row, err := uc.repo.GetPriceListByCode(ctx, code)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "price list not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "price list fetched", priceListToOutput(row))
}

// GetDefaultPriceList returns the default active price list.
func (uc *PriceListsUseCase) GetDefaultPriceList(ctx context.Context) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	row, err := uc.repo.GetDefaultPriceList(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "default price list not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "default price list fetched", priceListToOutput(row))
}

// ListPriceLists lists all price lists.
func (uc *PriceListsUseCase) ListPriceLists(ctx context.Context) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	rows, err := uc.repo.ListPriceLists(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]PriceListOutput, len(rows))
	for i := range rows {
		out[i] = priceListToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "price lists listed", out)
}

// ListActivePriceLists lists only active price lists.
func (uc *PriceListsUseCase) ListActivePriceLists(ctx context.Context) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	rows, err := uc.repo.ListActivePriceLists(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]PriceListOutput, len(rows))
	for i := range rows {
		out[i] = priceListToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "active price lists listed", out)
}

// ListValidPriceLists lists active price lists valid for current date.
func (uc *PriceListsUseCase) ListValidPriceLists(ctx context.Context) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	rows, err := uc.repo.ListValidPriceLists(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]PriceListOutput, len(rows))
	for i := range rows {
		out[i] = priceListToOutput(rows[i])
	}
	return utils.NewResponse(utils.CodeOK, "valid price lists listed", out)
}

// UpdatePriceList updates an existing price list.
func (uc *PriceListsUseCase) UpdatePriceList(
	ctx context.Context,
	id string,
	name string,
	priceListType *string,
	currencyCode *string,
	validFrom *string,
	validTo *string,
	isDefault bool,
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

	var plt pgtype.Text
	if priceListType != nil {
		plt = pgtype.Text{String: *priceListType, Valid: true}
	}

	var curr pgtype.Text
	if currencyCode != nil {
		curr = pgtype.Text{String: *currencyCode, Valid: true}
	}

	var fromDate pgtype.Date
	if validFrom != nil && *validFrom != "" {
		parsed, err := time.Parse("2006-01-02", *validFrom)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid valid_from", nil)
		}
		fromDate = pgtype.Date{Time: parsed, Valid: true}
	}

	var toDate pgtype.Date
	if validTo != nil && *validTo != "" {
		parsed, err := time.Parse("2006-01-02", *validTo)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid valid_to", nil)
		}
		toDate = pgtype.Date{Time: parsed, Valid: true}
	}

	def := pgtype.Bool{Bool: isDefault, Valid: true}
	active := pgtype.Bool{Bool: isActive, Valid: true}

	metaBytes := []byte("{}")
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			metaBytes = b
		}
	}

	row, err := uc.repo.UpdatePriceList(ctx, repository.UpdatePriceListParams{
		ID:            int32(parsed),
		Name:          name,
		PriceListType: plt,
		CurrencyCode:  curr,
		ValidFrom:     fromDate,
		ValidTo:       toDate,
		IsDefault:     def,
		IsActive:      active,
		Metadata:      metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "price list updated", priceListToOutput(row))
}

// DeletePriceList deletes a price list by ID.
func (uc *PriceListsUseCase) DeletePriceList(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	if err := uc.repo.DeletePriceList(ctx, int32(parsed)); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "price list deleted", nil)
}

// SetDefaultPriceList marks the given price list as default.
func (uc *PriceListsUseCase) SetDefaultPriceList(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	if err := uc.repo.SetDefaultPriceList(ctx, int32(parsed)); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	// Return the updated default
	row, err := uc.repo.GetPriceList(ctx, int32(parsed))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "default price list set", priceListToOutput(row))
}

// TogglePriceListActive toggles active flag for a price list.
func (uc *PriceListsUseCase) TogglePriceListActive(ctx context.Context, id string, isActive bool) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	row, err := uc.repo.TogglePriceListActive(ctx, repository.TogglePriceListActiveParams{
		ID:       int32(parsed),
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "price list active flag updated", priceListToOutput(row))
}
