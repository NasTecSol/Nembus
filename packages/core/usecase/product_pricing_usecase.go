package usecase

import (
	"context"
	"encoding/json"
	"math/big"
	"strconv"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// ProductPriceOutput represents a product price with JSON-friendly metadata.
type ProductPriceOutput struct {
	ID               int32            `json:"id"`
	ProductID        int32            `json:"product_id"`
	ProductVariantID pgtype.Int4      `json:"product_variant_id"`
	PriceListID      int32            `json:"price_list_id"`
	UomID            pgtype.Int4      `json:"uom_id"`
	Price            pgtype.Numeric   `json:"price"`
	MinQuantity      pgtype.Numeric   `json:"min_quantity"`
	MaxQuantity      pgtype.Numeric   `json:"max_quantity"`
	ValidFrom        pgtype.Date      `json:"valid_from"`
	ValidTo          pgtype.Date      `json:"valid_to"`
	IsActive         pgtype.Bool      `json:"is_active"`
	Metadata         json.RawMessage  `json:"metadata"`
	CreatedAt        pgtype.Timestamp `json:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at"`
}

func productPriceToOutput(p repository.ProductPrice) ProductPriceOutput {
	return ProductPriceOutput{
		ID:               p.ID,
		ProductID:        p.ProductID,
		ProductVariantID: p.ProductVariantID,
		PriceListID:      p.PriceListID,
		UomID:            p.UomID,
		Price:            p.Price,
		MinQuantity:      p.MinQuantity,
		MaxQuantity:      p.MaxQuantity,
		ValidFrom:        p.ValidFrom,
		ValidTo:          p.ValidTo,
		IsActive:         p.IsActive,
		Metadata:         utils.BytesToJSONRawMessage(p.Metadata),
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}

type ProductPricingUseCase struct {
	repo *repository.Queries
}

func NewProductPricingUseCase() *ProductPricingUseCase {
	return &ProductPricingUseCase{}
}

func (uc *ProductPricingUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *ProductPricingUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

// CreateProductPrice creates a new product price.
// CreateProductPrice creates a new product price.
func (uc *ProductPricingUseCase) CreateProductPrice(
	ctx context.Context,
	productID int32,
	productVariantID *int32,
	priceListID int32,
	uomID *int32,
	price float64,
	minQuantity *float64,
	maxQuantity *float64,
	validFrom *time.Time,
	validTo *time.Time,
	isActive *bool,
	metadata map[string]interface{},
) *repository.Response {

	// ---- Repo check ----
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	// ---- Basic validation ----
	if productID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "product_id is required", nil)
	}
	if priceListID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "price_list_id is required", nil)
	}
	if price <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "price must be greater than 0", nil)
	}

	// ---- Int4 conversions ----
	var pvID pgtype.Int4
	if productVariantID != nil {
		pvID = pgtype.Int4{Int32: *productVariantID, Valid: true}
	}

	var uID pgtype.Int4
	if uomID != nil {
		uID = pgtype.Int4{Int32: *uomID, Valid: true}
	}

	// ---- Numeric conversions ----
	priceNum := pgtype.Numeric{Valid: true}
	if err := priceNum.Scan(strconv.FormatFloat(price, 'f', -1, 64)); err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid price format", nil)
	}

	var minQty pgtype.Numeric
	if minQuantity != nil {
		if err := minQty.Scan(strconv.FormatFloat(*minQuantity, 'f', -1, 64)); err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid min_quantity format", nil)
		}
		minQty.Valid = true
	}

	var maxQty pgtype.Numeric
	if maxQuantity != nil {
		if err := maxQty.Scan(strconv.FormatFloat(*maxQuantity, 'f', -1, 64)); err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid max_quantity format", nil)
		}
		maxQty.Valid = true
	}

	// ---- Date conversions ----
	var fromDate pgtype.Date
	if validFrom != nil {
		fromDate = pgtype.Date{Time: *validFrom, Valid: true}
	}

	var toDate pgtype.Date
	if validTo != nil {
		toDate = pgtype.Date{Time: *validTo, Valid: true}
	}

	// ---- Bool conversion ----
	active := pgtype.Bool{Bool: true, Valid: true} // default true
	if isActive != nil {
		active.Bool = *isActive
	}

	// ---- Metadata ----
	metaBytes := []byte("{}")
	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			metaBytes = b
		}
	}

	// ---- DB call ----
	row, err := uc.repo.CreateProductPrice(ctx, repository.CreateProductPriceParams{
		ProductID:        productID,
		ProductVariantID: pvID,
		PriceListID:      priceListID,
		UomID:            uID,
		Price:            priceNum,
		MinQuantity:      minQty,
		MaxQuantity:      maxQty,
		ValidFrom:        fromDate,
		ValidTo:          toDate,
		IsActive:         active,
		Metadata:         metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "product price created", productPriceToOutput(row))
}

// GetProductPrice gets a product price by ID.
func (uc *ProductPricingUseCase) GetProductPrice(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	row, err := uc.repo.GetProductPrice(ctx, int32(parsed))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "product price not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "product price fetched", productPriceToOutput(row))
}

// UpdateProductPrice updates an existing product price.
func (uc *ProductPricingUseCase) UpdateProductPrice(
	ctx context.Context,
	id string,
	price *string,
	minQuantity *string,
	maxQuantity *string,
	validFrom *string,
	validTo *string,
	isActive *bool,
	metadata map[string]interface{},
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	params := repository.UpdateProductPriceParams{
		ID: int32(parsed),
	}

	if price != nil && *price != "" {
		priceNum, err := parseNumeric(*price)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid price format", nil)
		}
		params.Price = priceNum
	}

	if minQuantity != nil && *minQuantity != "" {
		minQtyNum, err := parseNumeric(*minQuantity)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid min_quantity format", nil)
		}
		params.MinQuantity = minQtyNum
	}

	if maxQuantity != nil && *maxQuantity != "" {
		maxQtyNum, err := parseNumeric(*maxQuantity)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid max_quantity format", nil)
		}
		params.MaxQuantity = maxQtyNum
	}

	if validFrom != nil && *validFrom != "" {
		parsed, err := time.Parse("2006-01-02", *validFrom)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid valid_from format", nil)
		}
		params.ValidFrom = pgtype.Date{Time: parsed, Valid: true}
	}

	if validTo != nil && *validTo != "" {
		parsed, err := time.Parse("2006-01-02", *validTo)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid valid_to format", nil)
		}
		params.ValidTo = pgtype.Date{Time: parsed, Valid: true}
	}

	if isActive != nil {
		params.IsActive = pgtype.Bool{Bool: *isActive, Valid: true}
	}

	if metadata != nil {
		if b, err := json.Marshal(metadata); err == nil {
			params.Metadata = b
		}
	}

	row, err := uc.repo.UpdateProductPrice(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "product price updated", productPriceToOutput(row))
}

// DeleteProductPrice deletes a product price by ID.
func (uc *ProductPricingUseCase) DeleteProductPrice(ctx context.Context, id string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	parsed, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid id", nil)
	}

	if err := uc.repo.DeleteProductPrice(ctx, int32(parsed)); err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "product price deleted", nil)
}

// ListProductPrices lists product prices for a product.
func (uc *ProductPricingUseCase) ListProductPrices(
	ctx context.Context,
	productID string,
	productVariantID *string,
	isActive *bool,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	prodID, err := strconv.ParseInt(productID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil)
	}

	params := repository.ListProductPricesParams{
		ProductID: int32(prodID),
	}

	if productVariantID != nil && *productVariantID != "" {
		pvID, err := strconv.ParseInt(*productVariantID, 10, 32)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid product_variant_id", nil)
		}
		params.ProductVariantID = pgtype.Int4{Int32: int32(pvID), Valid: true}
	}

	if isActive != nil {
		params.IsActive = pgtype.Bool{Bool: *isActive, Valid: true}
	}

	rows, err := uc.repo.ListProductPrices(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]repository.ListProductPricesRow, len(rows))
	copy(out, rows)
	return utils.NewResponse(utils.CodeOK, "product prices listed", out)
}

// GetEffectivePrice gets the effective price for a product.
func (uc *ProductPricingUseCase) GetEffectivePrice(
	ctx context.Context,
	productID string,
	priceListID string,
	quantity string,
	productVariantID *string,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	prodID, err := strconv.ParseInt(productID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil)
	}

	plID, err := strconv.ParseInt(priceListID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid price_list_id", nil)
	}

	qtyNum, err := parseNumeric(quantity)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid quantity format", nil)
	}

	params := repository.GetEffectivePriceParams{
		ProductID:   int32(prodID),
		PriceListID: int32(plID),
		MinQuantity: qtyNum,
	}

	if productVariantID != nil && *productVariantID != "" {
		pvID, err := strconv.ParseInt(*productVariantID, 10, 32)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid product_variant_id", nil)
		}
		params.ProductVariantID = pgtype.Int4{Int32: int32(pvID), Valid: true}
	}

	row, err := uc.repo.GetEffectivePrice(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "effective price not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "effective price fetched", productPriceToOutput(row))
}

// GetProductPriceForList gets product price for a specific price list.
func (uc *ProductPricingUseCase) GetProductPriceForList(
	ctx context.Context,
	productID string,
	priceListID string,
	productVariantID *string,
	uomID *string,
	quantity *string,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	prodID, err := strconv.ParseInt(productID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil)
	}

	plID, err := strconv.ParseInt(priceListID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid price_list_id", nil)
	}

	params := repository.GetProductPriceForListParams{
		ProductID:   int32(prodID),
		PriceListID: int32(plID),
	}

	if productVariantID != nil && *productVariantID != "" {
		pvID, err := strconv.ParseInt(*productVariantID, 10, 32)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid product_variant_id", nil)
		}
		params.ProductVariantID = pgtype.Int4{Int32: int32(pvID), Valid: true}
	}

	if uomID != nil && *uomID != "" {
		uID, err := strconv.ParseInt(*uomID, 10, 32)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid uom_id", nil)
		}
		params.UomID = pgtype.Int4{Int32: int32(uID), Valid: true}
	}

	if quantity != nil && *quantity != "" {
		qtyNum, err := parseNumeric(*quantity)
		if err != nil {
			return utils.NewResponse(utils.CodeBadReq, "invalid quantity format", nil)
		}
		params.Quantity = qtyNum
	} else {
		params.Quantity = pgtype.Numeric{Int: big.NewInt(1), Exp: 0, Valid: true}
	}

	row, err := uc.repo.GetProductPriceForList(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "product price not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "product price fetched", productPriceToOutput(row))
}

// GetPriceComparison gets price comparison across all price lists for a product.
func (uc *ProductPricingUseCase) GetPriceComparison(ctx context.Context, productID string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	prodID, err := strconv.ParseInt(productID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil)
	}

	rows, err := uc.repo.GetPriceComparison(ctx, int32(prodID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]repository.GetPriceComparisonRow, len(rows))
	copy(out, rows)
	return utils.NewResponse(utils.CodeOK, "price comparison fetched", out)
}

// ListPricesByPriceList lists all prices for a specific price list.
func (uc *ProductPricingUseCase) ListPricesByPriceList(ctx context.Context, priceListID string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	plID, err := strconv.ParseInt(priceListID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid price_list_id", nil)
	}

	rows, err := uc.repo.ListPricesByPriceList(ctx, int32(plID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]repository.ListPricesByPriceListRow, len(rows))
	copy(out, rows)
	return utils.NewResponse(utils.CodeOK, "prices listed", out)
}

// GetProductWithPricing gets a product with all its pricing information.
func (uc *ProductPricingUseCase) GetProductWithPricing(ctx context.Context, productID string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	prodID, err := strconv.ParseInt(productID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid product_id", nil)
	}

	row, err := uc.repo.GetProductWithPricing(ctx, int32(prodID))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "product not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "product with pricing fetched", row)
}

// SearchProductsWithPrices searches products with prices for a specific price list.
func (uc *ProductPricingUseCase) SearchProductsWithPrices(
	ctx context.Context,
	organizationID string,
	priceListID string,
	searchTerm string,
	limit int32,
	offset int32,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	orgID, err := strconv.ParseInt(organizationID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid organization_id", nil)
	}

	plID, err := strconv.ParseInt(priceListID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid price_list_id", nil)
	}

	searchText := pgtype.Text{String: searchTerm, Valid: true}

	rows, err := uc.repo.SearchProductsWithPrices(ctx, repository.SearchProductsWithPricesParams{
		OrganizationID: int32(orgID),
		PriceListID:    int32(plID),
		Column3:        searchText,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]repository.SearchProductsWithPricesRow, len(rows))
	copy(out, rows)
	return utils.NewResponse(utils.CodeOK, "products searched", out)
}

// BulkUpdatePrices updates prices in bulk for a price list.
func (uc *ProductPricingUseCase) BulkUpdatePrices(
	ctx context.Context,
	priceListID string,
	percentageChange string,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	plID, err := strconv.ParseInt(priceListID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid price_list_id", nil)
	}

	percentage, err := parseNumeric(percentageChange)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid percentage_change format", nil)
	}

	err = uc.repo.BulkUpdatePrices(ctx, repository.BulkUpdatePricesParams{
		PriceListID: int32(plID),
		Column2:     percentage,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "prices updated in bulk", nil)
}

// ExpirePrices expires all active prices for a price list.
func (uc *ProductPricingUseCase) ExpirePrices(ctx context.Context, priceListID string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	plID, err := strconv.ParseInt(priceListID, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid price_list_id", nil)
	}

	err = uc.repo.ExpirePrices(ctx, int32(plID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "prices expired", nil)
}

// Helper function to parse numeric strings to pgtype.Numeric
func parseNumeric(s string) (pgtype.Numeric, error) {
	var num pgtype.Numeric
	if err := num.Scan(s); err != nil {
		return num, err
	}
	return num, nil
}
