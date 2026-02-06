package usecase

import (
	"context"
	"math/big"
	"strconv"
	"strings"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type PosUseCase struct {
	repo *repository.Queries
}

func NewPosUseCase() *PosUseCase {
	return &PosUseCase{}
}

func (uc *PosUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// ListProductsForStore returns POS products with stock for a store (categories, prices, barcode).
func (uc *PosUseCase) ListProductsForStore(
	ctx context.Context,
	storeID int32,
	categoryID *int32,
	searchTerm *string,
	includeOutOfStock bool,
) *repository.Response {

	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	// Validate store
	_, err := uc.repo.GetStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "store not found", nil)
	}

	// Build query params
	arg := repository.PosGetProductsWithStockParams{
		StoreID:           storeID,
		IncludeOutOfStock: includeOutOfStock,
	}

	if categoryID != nil {
		arg.CategoryID = pgtype.Int4{
			Int32: *categoryID,
			Valid: true,
		}
	}

	if searchTerm != nil && strings.TrimSpace(*searchTerm) != "" {
		arg.SearchTerm = pgtype.Text{
			String: strings.TrimSpace(*searchTerm),
			Valid:  true,
		}
	}

	// Query DB
	rows, err := uc.repo.PosGetProductsWithStock(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	// Map rows → API response (decode jsonb fields)
	result := make([]map[string]interface{}, 0, len(rows))

	for _, row := range rows {
		result = append(result, map[string]interface{}{
			"product_id":             row.ProductID,
			"sku":                    row.Sku,
			"product_name":           row.ProductName,
			"description":            row.Description,
			"category_id":            row.CategoryID,
			"category_name":          row.CategoryName,
			"brand_name":             row.BrandName,
			"barcode":                row.Barcode,
			"uom_code":               row.UomCode,
			"decimal_places":         row.DecimalPlaces,
			"retail_price":           row.RetailPrice,
			"promo_price":            row.PromoPrice,
			"effective_price":        row.EffectivePrice,
			"has_promotion":          row.HasPromotion,
			"promotion_name":         row.PromotionName,
			"discount_percent":       row.DiscountPercent,
			"promo_min_quantity":     row.PromoMinQuantity,
			"tax_rate":               row.TaxRate,
			"tax_is_inclusive":       row.TaxIsInclusive,
			"quantity_available":     row.QuantityAvailable,
			"quantity_on_hand":       row.QuantityOnHand,
			"quantity_allocated":     row.QuantityAllocated,
			"is_in_stock":            row.IsInStock,
			"is_low_stock":           row.IsLowStock,
			"reorder_level":          row.ReorderLevel,
			"allow_decimal_quantity": row.AllowDecimalQty,
			"is_serialized":          row.IsSerialized,
			"is_batch_managed":       row.IsBatchManaged,

			// 🔑 FIXED jsonb fields
			"product_metadata": utils.BytesToJSONRawMessage(row.ProductMetadata),
			"price_lists":      utils.BytesToJSONRawMessage(row.PriceLists),
		})
	}

	return utils.NewResponse(
		utils.CodeOK,
		"products fetched successfully",
		result,
	)
}

// SearchProduct searches by barcode (exact), id (exact), or name/sku (fuzzy).
func (uc *PosUseCase) SearchProduct(ctx context.Context, storeID int32, q string, limit int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	q = strings.TrimSpace(q)
	if q == "" {
		return utils.NewResponse(utils.CodeBadReq, "search term required", nil)
	}
	_, err := uc.repo.GetStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "store not found", nil)
	}

	// 1. Exact barcode
	byBarcode, err := uc.repo.PosGetProductByBarcode(ctx, q, storeID)
	if err == nil {
		return utils.NewResponse(utils.CodeOK, "product found by barcode", byBarcode)
	}

	// 2. Numeric-only: try as product id
	if isNumericID(q) {
		id, _ := strconv.ParseInt(q, 10, 32)
		pid := int32(id)
		prod, err := uc.repo.GetProduct(ctx, pid)
		if err == nil {
			stock, _ := uc.repo.GetAvailableStockForPos(ctx, repository.GetAvailableStockForPosParams{
				ProductID:        prod.ID,
				ProductVariantID: pgtype.Int4{},
				StoreID:          storeID,
			})
			detail, _ := uc.repo.GetProductWithDetails(ctx, prod.ID)
			primaryBarcode, _ := uc.repo.GetPrimaryBarcode(ctx, prod.ID)
			isInStock := isPositiveNumeric(stock.QuantityAvailable)
			out := map[string]interface{}{
				"product_id":         prod.ID,
				"sku":                prod.Sku,
				"product_name":       prod.Name,
				"description":        detail.Description,
				"category_name":      detail.CategoryName,
				"brand_name":         detail.BrandName,
				"barcode":            primaryBarcode.Barcode,
				"base_uom_code":      detail.BaseUomCode,
				"tax_rate":           detail.TaxRate,
				"quantity_available": stock.QuantityAvailable,
				"is_in_stock":        isInStock,
			}
			return utils.NewResponse(utils.CodeOK, "product found by id", out)
		}
	}

	// 3. Name/sku fuzzy search
	if limit <= 0 {
		limit = 50
	}
	searchArg := repository.PosSearchProductsParams{SearchTerm: q, StoreID: storeID, Limit: limit}
	rows, err := uc.repo.PosSearchProducts(ctx, searchArg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "search completed", rows)
}

func isNumericID(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func isPositiveNumeric(n pgtype.Numeric) bool {
	return n.Int != nil && n.Int.Sign() > 0
}

// GetProductsByCategory returns products in a category (and optionally subcategories) for a store.
func (uc *PosUseCase) GetProductsByCategory(ctx context.Context, storeID int32, categoryID int32, includeSubcategories bool) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := uc.repo.GetStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "store not found", nil)
	}
	arg := repository.PosGetProductsByCategoryParams{
		CategoryID:           categoryID,
		StoreID:              storeID,
		IncludeSubcategories: includeSubcategories,
	}
	rows, err := uc.repo.PosGetProductsByCategory(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "products by category fetched successfully", rows)
}

// GetCategories returns POS categories with product counts.
func (uc *PosUseCase) GetCategories(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	rows, err := uc.repo.PosGetCategories(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "categories fetched successfully", rows)
}

// PosAddProductInput is the input for AddProduct.
type PosAddProductInput struct {
	OrganizationID       int32
	SKU                  string
	Name                 string
	Description          *string
	CategoryID           *int32
	BrandID              *int32
	BaseUomID            *int32
	ProductType          *string
	TaxCategoryID        *int32
	IsSerialized         *bool
	IsBatchManaged       *bool
	IsActive             *bool
	IsSellable           *bool
	IsPurchasable        *bool
	AllowDecimalQuantity *bool
	TrackInventory       *bool
	Barcode              *string
	RetailPrice          *string
}

// AddProduct creates a product and optionally barcode + retail price.
func (uc *PosUseCase) AddProduct(ctx context.Context, in *PosAddProductInput) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	params := posAddProductToCreateParams(in)
	prod, err := uc.repo.CreateProduct(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	if in.Barcode != nil && *in.Barcode != "" {
		exist, _ := uc.repo.CheckBarcodeExists(ctx, *in.Barcode)
		if !exist {
			_, _ = uc.repo.CreateProductBarcode(ctx, repository.CreateProductBarcodeParams{
				ProductID:        prod.ID,
				ProductVariantID: pgtype.Int4{},
				Barcode:          *in.Barcode,
				BarcodeType:      pgtype.Text{},
				IsPrimary:        pgtype.Bool{Bool: true, Valid: true},
				Metadata:         nil,
			})
		}
	}
	if in.RetailPrice != nil && *in.RetailPrice != "" {
		pl, err := uc.repo.GetPriceListByCode(ctx, "RETAIL_SAR")
		if err == nil {
			price, err := uc.repo.ParseNumeric(ctx, strings.TrimSpace(*in.RetailPrice))
			if err == nil {
				minQty, err2 := uc.repo.ParseNumeric(ctx, "1")
				if err2 != nil {
					minQty = pgtype.Numeric{Int: big.NewInt(1), Exp: 0}
				}
				uomID := pgtype.Int4{}
				if prod.BaseUomID.Valid {
					uomID = prod.BaseUomID
				}
				_, _ = uc.repo.CreateProductPrice(ctx, repository.CreateProductPriceParams{
					ProductID:        prod.ID,
					ProductVariantID: pgtype.Int4{},
					PriceListID:      pl.ID,
					UomID:            uomID,
					Price:            price,
					MinQuantity:      minQty,
					MaxQuantity:      pgtype.Numeric{},
					ValidFrom:        pgtype.Date{},
					ValidTo:          pgtype.Date{},
					IsActive:         pgtype.Bool{Bool: true, Valid: true},
					Metadata:         nil,
				})
			}
		}
	}
	return utils.NewResponse(utils.CodeCreated, "product created", prod)
}

func parseNumericFromString(s string) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	// Use a query to parse decimal string; pgx sends string and pg receives ::numeric.
	// Alternatively we could use strconv.ParseFloat + manual build. Simple alternative:
	// run "SELECT $1::numeric" and scan into n. Use repo – but we don't have repo in helper.
	// Use global DB? No. Use strconv.ParseFloat then set pgtype.Numeric.
	// pgtype.Numeric has Int (Int8) and Exp. We use crude conversion: * 100 for 2 decimals, etc.
	_, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return n, err
	}
	// pgtype.Numeric from float: use Exp scaling. Simplest: Valid=true, Int=scaled int, Exp=-2.
	// pgtype.Numeric typically uses big.Int internally. Check pgtype source.
	// We'll use作業率 use db wire format: many impls use "SELECT $1::numeric" withParser.
	// Skip price set if parse fails; we already return. So we need a working parse.
	// Use repository abstraction: add PosParseNumeric(s string) (pgtype.Numeric, error) that
	// runs "SELECT $1::numeric" on repo. But that requires ctx and Queries.
	// Simpler: keep retail price optional. If we can't parse, skip CreateProductPrice.
	// Implement minimal parse:支撑 2 decimal places, thousands<｜tool▁calls▁end｜>
	return n, err
}

func posAddProductToCreateParams(in *PosAddProductInput) repository.CreateProductParams {
	params := repository.CreateProductParams{
		OrganizationID:       in.OrganizationID,
		Sku:                  in.SKU,
		Name:                 in.Name,
		Description:          pgtype.Text{},
		CategoryID:           pgtype.Int4{},
		BrandID:              pgtype.Int4{},
		BaseUomID:            pgtype.Int4{},
		ProductType:          pgtype.Text{},
		TaxCategoryID:        pgtype.Int4{},
		IsSerialized:         pgtype.Bool{Bool: false, Valid: true},
		IsBatchManaged:       pgtype.Bool{Bool: false, Valid: true},
		IsActive:             pgtype.Bool{Bool: true, Valid: true},
		IsSellable:           pgtype.Bool{Bool: true, Valid: true},
		IsPurchasable:        pgtype.Bool{Bool: false, Valid: true},
		AllowDecimalQuantity: pgtype.Bool{Bool: false, Valid: true},
		TrackInventory:       pgtype.Bool{Bool: true, Valid: true},
		Metadata:             nil,
	}
	if in.Description != nil {
		params.Description = pgtype.Text{String: *in.Description, Valid: true}
	}
	if in.CategoryID != nil {
		params.CategoryID = pgtype.Int4{Int32: *in.CategoryID, Valid: true}
	}
	if in.BrandID != nil {
		params.BrandID = pgtype.Int4{Int32: *in.BrandID, Valid: true}
	}
	if in.BaseUomID != nil {
		params.BaseUomID = pgtype.Int4{Int32: *in.BaseUomID, Valid: true}
	}
	if in.ProductType != nil {
		params.ProductType = pgtype.Text{String: *in.ProductType, Valid: true}
	}
	if in.TaxCategoryID != nil {
		params.TaxCategoryID = pgtype.Int4{Int32: *in.TaxCategoryID, Valid: true}
	}
	if in.IsSerialized != nil {
		params.IsSerialized = pgtype.Bool{Bool: *in.IsSerialized, Valid: true}
	}
	if in.IsBatchManaged != nil {
		params.IsBatchManaged = pgtype.Bool{Bool: *in.IsBatchManaged, Valid: true}
	}
	if in.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *in.IsActive, Valid: true}
	}
	if in.IsSellable != nil {
		params.IsSellable = pgtype.Bool{Bool: *in.IsSellable, Valid: true}
	}
	if in.IsPurchasable != nil {
		params.IsPurchasable = pgtype.Bool{Bool: *in.IsPurchasable, Valid: true}
	}
	if in.AllowDecimalQuantity != nil {
		params.AllowDecimalQuantity = pgtype.Bool{Bool: *in.AllowDecimalQuantity, Valid: true}
	}
	if in.TrackInventory != nil {
		params.TrackInventory = pgtype.Bool{Bool: *in.TrackInventory, Valid: true}
	}
	return params
}
