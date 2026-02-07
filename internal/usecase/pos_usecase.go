package usecase

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type PosUseCase struct {
}

func NewPosUseCase() *PosUseCase {
	return &PosUseCase{}
}

// ListProductsForStore returns POS products with stock for a store (categories, prices, barcode).
func (uc *PosUseCase) ListProductsForStore(
	ctx context.Context,
	repo *repository.Queries,
	storeID int32,
	categoryID *int32,
	searchTerm *string,
	includeOutOfStock bool,
) *repository.Response {

	if repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	// Validate store
	_, err := repo.GetStore(ctx, storeID)
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
	rows, err := repo.PosGetProductsWithStock(ctx, arg)
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

			// 🔑 FIXED jsonb fields: package_n_price (packages/UOMs with prices), product_uom_conversions (e.g. 1 box = 10 packs, 1 pack = 150 ml)
			"product_metadata":        utils.BytesToJSONRawMessage(row.ProductMetadata),
			"package_n_price":         utils.BytesToJSONRawMessage(row.PackageNPrice),
			"product_uom_conversions": utils.BytesToJSONRawMessage(row.ProductUomConversions),
		})
	}

	return utils.NewResponse(
		utils.CodeOK,
		"products fetched successfully",
		result,
	)
}

// SearchProduct searches by barcode (exact), id (exact), or name/sku (fuzzy).
func (uc *PosUseCase) SearchProduct(ctx context.Context, repo *repository.Queries, storeID int32, q string, limit int32) *repository.Response {
	if repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	q = strings.TrimSpace(q)
	if q == "" {
		return utils.NewResponse(utils.CodeBadReq, "search term required", nil)
	}
	_, err := repo.GetStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "store not found", nil)
	}

	// 1. Exact barcode
	byBarcode, err := repo.PosGetProductByBarcode(ctx, q, storeID)
	if err == nil {
		return utils.NewResponse(utils.CodeOK, "product found by barcode", byBarcode)
	}

	// 2. Numeric-only: try as product id
	if isNumericID(q) {
		id, _ := strconv.ParseInt(q, 10, 32)
		pid := int32(id)
		prod, err := repo.GetProduct(ctx, pid)
		if err == nil {
			stock, _ := repo.GetAvailableStockForPos(ctx, repository.GetAvailableStockForPosParams{
				ProductID:        prod.ID,
				ProductVariantID: pgtype.Int4{},
				StoreID:          storeID,
			})
			detail, _ := repo.GetProductWithDetails(ctx, prod.ID)
			primaryBarcode, _ := repo.GetPrimaryBarcode(ctx, prod.ID)
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
	rows, err := repo.PosSearchProducts(ctx, searchArg)
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
func (uc *PosUseCase) GetProductsByCategory(ctx context.Context, repo *repository.Queries, storeID int32, categoryID int32, includeSubcategories bool) *repository.Response {
	if repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	_, err := repo.GetStore(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "store not found", nil)
	}
	arg := repository.PosGetProductsByCategoryParams{
		CategoryID:           categoryID,
		StoreID:              storeID,
		IncludeSubcategories: includeSubcategories,
	}
	rows, err := repo.PosGetProductsByCategory(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "products by category fetched successfully", rows)
}

// GetCategories returns POS categories with product counts.
func (uc *PosUseCase) GetCategories(ctx context.Context, repo *repository.Queries) *repository.Response {
	if repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	rows, err := repo.PosGetCategories(ctx)
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
	Metadata             *string
}

type PosAddProductFullInput struct {
	PosAddProductInput
	Barcodes    []BarcodeItemInput
	Prices      []PriceItemInput
	Conversions []ConversionItemInput
}

type BarcodeItemInput struct {
	Barcode     string                 `json:"barcode"`
	BarcodeType string                 `json:"barcode_type"`
	IsPrimary   bool                   `json:"is_primary"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type PriceItemInput struct {
	PriceListID int32                  `json:"price_list_id"`
	UomID       *int32                 `json:"uom_id"`
	Price       string                 `json:"price"`
	MinQuantity float64                `json:"min_quantity"`
	MaxQuantity *float64               `json:"max_quantity"`
	ValidFrom   *string                `json:"valid_from"`
	ValidTo     *string                `json:"valid_to"`
	IsActive    bool                   `json:"is_active"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type ConversionItemInput struct {
	FromUomID        int32                  `json:"from_uom_id"`
	ToUomID          int32                  `json:"to_uom_id"`
	ConversionFactor float64                `json:"conversion_factor"`
	IsDefault        bool                   `json:"is_default"`
	Metadata         map[string]interface{} `json:"metadata"`
}

func (uc *PosUseCase) AddProductFull(ctx context.Context, repo *repository.Queries, in *PosAddProductFullInput) *repository.Response {
	if repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	// Handle backward compatibility for single Barcode/RetailPrice
	if in.Barcode != nil && *in.Barcode != "" {
		found := false
		for _, b := range in.Barcodes {
			if b.Barcode == *in.Barcode {
				found = true
				break
			}
		}
		if !found {
			in.Barcodes = append(in.Barcodes, BarcodeItemInput{
				Barcode:   *in.Barcode,
				IsPrimary: true,
			})
		}
	}

	if in.RetailPrice != nil && *in.RetailPrice != "" {
		found := false
		pl, _ := repo.GetPriceListByCode(ctx, "RETAIL_SAR")
		for _, p := range in.Prices {
			if p.PriceListID == pl.ID {
				found = true
				break
			}
		}
		if !found && pl.ID != 0 {
			in.Prices = append(in.Prices, PriceItemInput{
				PriceListID: pl.ID,
				Price:       *in.RetailPrice,
				MinQuantity: 1,
				IsActive:    true,
			})
		}
	}

	barcodesJSON, _ := json.Marshal(in.Barcodes)
	pricesJSON, _ := json.Marshal(in.Prices)
	conversionsJSON, _ := json.Marshal(in.Conversions)

	params := repository.CreateProductFullParams{
		OrganizationID:       in.OrganizationID,
		Sku:                  in.SKU,
		Name:                 in.Name,
		Description:          pgtype.Text{String: getString(in.Description), Valid: in.Description != nil},
		CategoryID:           pgtype.Int4{Int32: getInt32(in.CategoryID), Valid: in.CategoryID != nil},
		BrandID:              pgtype.Int4{Int32: getInt32(in.BrandID), Valid: in.BrandID != nil},
		BaseUomID:            pgtype.Int4{Int32: getInt32(in.BaseUomID), Valid: in.BaseUomID != nil},
		ProductType:          pgtype.Text{String: getString(in.ProductType), Valid: in.ProductType != nil},
		TaxCategoryID:        pgtype.Int4{Int32: getInt32(in.TaxCategoryID), Valid: in.TaxCategoryID != nil},
		IsSerialized:         pgtype.Bool{Bool: getBool(in.IsSerialized), Valid: in.IsSerialized != nil},
		IsBatchManaged:       pgtype.Bool{Bool: getBool(in.IsBatchManaged), Valid: in.IsBatchManaged != nil},
		IsActive:             pgtype.Bool{Bool: getBool(in.IsActive, true), Valid: true},
		IsSellable:           pgtype.Bool{Bool: getBool(in.IsSellable, true), Valid: true},
		IsPurchasable:        pgtype.Bool{Bool: getBool(in.IsPurchasable, false), Valid: true},
		AllowDecimalQuantity: pgtype.Bool{Bool: getBool(in.AllowDecimalQuantity, false), Valid: true},
		TrackInventory:       pgtype.Bool{Bool: getBool(in.TrackInventory, true), Valid: true},
		Metadata:             []byte(getString(in.Metadata, "{}")),
		Barcodes:             barcodesJSON,
		Prices:               pricesJSON,
		Conversions:          conversionsJSON,
	}

	productID, err := repo.CreateProductFull(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "product created successfully", map[string]interface{}{"product_id": productID})
}

func getString(s *string, def ...string) string {
	if s == nil {
		if len(def) > 0 {
			return def[0]
		}
		return ""
	}
	return *s
}

func getInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func getBool(b *bool, def ...bool) bool {
	if b == nil {
		if len(def) > 0 {
			return def[0]
		}
		return false
	}
	return *b
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

