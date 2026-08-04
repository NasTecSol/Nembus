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

// PriceResolutionOutput holds the resolved price and source tier details.
type PriceResolutionOutput struct {
	ProductID   int32          `json:"product_id"`
	Price       pgtype.Numeric `json:"price"`
	PricingTier string         `json:"pricing_tier"` // "contract" | "price_list" | "standard"
	ContractID  *int32         `json:"contract_id,omitempty"`
	PriceListID *int32         `json:"price_list_id,omitempty"`
}

// ResolvePrice implements 3-tier pricing resolution (SAP B1 pattern):
//   Tier 1: BP Negotiated Contract Price (if customer is linked to business_partner)
//   Tier 2: Price List Price (customer's price_list_id)
//   Tier 3: Standard Product Price
func (uc *PriceListsUseCase) ResolvePrice(
	ctx context.Context,
	customerID *int32,
	productID int32,
	variantID *int32,
	priceListID *int32,
	quantity float64,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	var qtyNum pgtype.Numeric
	_ = qtyNum.Scan(strconv.FormatFloat(quantity, 'f', 2, 64))

	// Tier 1: Check B2B Contract Price if customer exists and has linked business partner
	if customerID != nil && *customerID > 0 {
		customer, err := uc.repo.GetCustomerWithPartner(ctx, *customerID)
		if err == nil && customer.BpID.Valid {
			bpID := customer.BpID.Int32
			contract, err := uc.repo.GetBPPriceContractForProduct(ctx, repository.GetBPPriceContractForProductParams{
				BusinessPartnerID: bpID,
				ProductID:         productID,
				ProductVariantID:  utils.Int32ToPgInt4(variantID),
				Quantity:          qtyNum,
			})
			if err == nil {
				contractID := contract.ID
				return utils.NewResponse(utils.CodeOK, "contract price resolved (Tier 1)", PriceResolutionOutput{
					ProductID:   productID,
					Price:       contract.ContractPrice,
					PricingTier: "contract",
					ContractID:  &contractID,
				})
			}
		}

		// Use customer's assigned price_list_id if none passed explicitly
		if (priceListID == nil || *priceListID <= 0) && customer.PriceListID.Valid {
			plID := customer.PriceListID.Int32
			priceListID = &plID
		}
	}

	// Tier 2: Check Price List Price if priceListID is available
	if priceListID != nil && *priceListID > 0 {
		priceRow, err := uc.repo.GetEffectivePrice(ctx, repository.GetEffectivePriceParams{
			ProductID:        productID,
			PriceListID:      *priceListID,
			MinQuantity:      qtyNum,
			ProductVariantID: utils.Int32ToPgInt4(variantID),
		})
		if err == nil {
			return utils.NewResponse(utils.CodeOK, "price list price resolved (Tier 2)", PriceResolutionOutput{
				ProductID:   productID,
				Price:       priceRow.Price,
				PricingTier: "price_list",
				PriceListID: priceListID,
			})
		}
	}

	// Tier 3: Default Price List or Standard Price fallback
	defaultPL, err := uc.repo.GetDefaultPriceList(ctx)
	if err == nil {
		priceRow, err := uc.repo.GetEffectivePrice(ctx, repository.GetEffectivePriceParams{
			ProductID:        productID,
			PriceListID:      defaultPL.ID,
			MinQuantity:      qtyNum,
			ProductVariantID: utils.Int32ToPgInt4(variantID),
		})
		if err == nil {
			defaultPLID := defaultPL.ID
			return utils.NewResponse(utils.CodeOK, "default price list price resolved (Tier 3)", PriceResolutionOutput{
				ProductID:   productID,
				Price:       priceRow.Price,
				PricingTier: "standard",
				PriceListID: &defaultPLID,
			})
		}
	}

	zeroPrice := pgtype.Numeric{}
	_ = zeroPrice.Scan("0.00")
	return utils.NewResponse(utils.CodeOK, "no effective price found", PriceResolutionOutput{
		ProductID:   productID,
		Price:       zeroPrice,
		PricingTier: "none",
	})
}


