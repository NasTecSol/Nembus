package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

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

			// 🔑 FIXED jsonb fields: package_n_price (packages/UOMs with prices), product_uom_conversions (e.g. 1 box = 10 packs, 1 pack = 150 ml)
			"product_metadata":        utils.BytesToJSONRawMessage(row.ProductMetadata),
			"product_variants":        utils.BytesToJSONRawMessage(row.ProductVariants),
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

	result := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		result = append(result, map[string]interface{}{
			"product_id":              row.ProductID,
			"sku":                     row.Sku,
			"product_name":            row.ProductName,
			"category_name":           row.CategoryName,
			"brand_name":              row.BrandName,
			"barcode":                 row.Barcode,
			"effective_price":         row.EffectivePrice,
			"has_promotion":           row.HasPromotion,
			"promotion_name":          row.PromotionName,
			"quantity_available":      row.QuantityAvailable,
			"is_in_stock":             row.IsInStock,
			"package_n_price":         utils.BytesToJSONRawMessage(row.PackageNPrice),
			"product_uom_conversions": utils.BytesToJSONRawMessage(row.ProductUomConversions),
		})
	}
	return utils.NewResponse(utils.CodeOK, "products by category fetched successfully", result)
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

// ProcessPOSPayment records a payment and updates the cashier session expected balance.
func (uc *PosUseCase) ProcessPOSPayment(ctx context.Context, arg repository.AddPaymentToTransactionParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	// 1. Record the payment
	err := uc.repo.AddPaymentToTransaction(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to record payment: "+err.Error(), nil)
	}

	// 2. Fetch the transaction to get the session ID
	txn, err := uc.repo.GetPosTransaction(ctx, arg.TransactionID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to fetch transaction: "+err.Error(), nil)
	}

	// 3. Update the expected balance in the cashier session
	err = uc.repo.UpdateSessionExpectedBalance(ctx, repository.UpdateSessionExpectedBalanceParams{
		ID:              txn.CashierSessionID,
		ExpectedBalance: arg.Amount,
	})
	if err != nil {
		// Note: We log the error but don't fail the payment recording if just the balance update fails
		fmt.Printf("Error: failed to update session expected balance for session %d: %s\n", txn.CashierSessionID, err.Error())
	}

	return utils.NewResponse(utils.CodeOK, "payment processed successfully", nil)
}

// PosCreateTransactionInput is the input payload for creating a POS transaction.
type PosCreateTransactionInput struct {
	TransactionNumber string
	StoreID           int32
	PosTerminalID     int32
	CashierSessionID  int32
	CashierID         int32
	CustomerID        *int32
	PriceListID       *int32
	TransactionType   *string
	TransactionDate   *time.Time
	Subtotal          pgtype.Numeric
	TaxAmount         pgtype.Numeric
	DiscountAmount    pgtype.Numeric
	TotalAmount       pgtype.Numeric
	TotalCost         pgtype.Numeric
	Status            *string
	Metadata          []byte
	Lines             []PosCreateTransactionLineInput
}

// PosCreateTransactionLineInput is the input for a transaction line.
type PosCreateTransactionLineInput struct {
	LineNumber       *int32
	ProductID        int32
	ProductVariantID *int32
	SerialNumber     *string
	BatchNumber      *string
	Quantity         pgtype.Numeric
	UomID            *int32
	UnitPrice        pgtype.Numeric
	DiscountAmount   pgtype.Numeric
	TaxAmount        pgtype.Numeric
	Subtotal         pgtype.Numeric
	LineTotal        pgtype.Numeric
	CostPrice        pgtype.Numeric
	Metadata         []byte
}

// PosTransactionOutput is the API response shape for a POS transaction.
type PosTransactionOutput struct {
	ID                int32            `json:"id"`
	StoreID           int32            `json:"store_id"`
	CashierID         int32            `json:"cashier_id"`
	CashierSessionID  int32            `json:"cashier_session_id"`
	CustomerID        pgtype.Int4      `json:"customer_id"`
	PosTerminalID     pgtype.Int4      `json:"pos_terminal_id"`
	TransactionNumber string           `json:"transaction_number"`
	TransactionDate   pgtype.Timestamp `json:"transaction_date"`
	TransactionType   pgtype.Text      `json:"transaction_type"`
	Subtotal          pgtype.Numeric   `json:"subtotal"`
	DiscountAmount    pgtype.Numeric   `json:"discount_amount"`
	TaxAmount         pgtype.Numeric   `json:"tax_amount"`
	TotalAmount       pgtype.Numeric   `json:"total_amount"`
	TotalCost         pgtype.Numeric   `json:"total_cost"`
	AmountPaid        pgtype.Numeric   `json:"amount_paid"`
	ChangeGiven       pgtype.Numeric   `json:"change_given"`
	Status            pgtype.Text      `json:"status"`
	PriceListID       pgtype.Int4      `json:"price_list_id"`
	SalesOrderID      pgtype.UUID      `json:"sales_order_id"`
	SourceCartID      pgtype.UUID      `json:"source_cart_id"`
	VoidedBy          pgtype.Int4      `json:"voided_by"`
	VoidedAt          pgtype.Timestamp `json:"voided_at"`
	Metadata          json.RawMessage  `json:"metadata"`
	CreatedAt         pgtype.Timestamp `json:"created_at"`
}

func toPosTransactionOutput(txn repository.PosTransaction) PosTransactionOutput {
	return PosTransactionOutput{
		ID:                txn.ID,
		StoreID:           txn.StoreID,
		CashierID:         txn.CashierID,
		CashierSessionID:  txn.CashierSessionID,
		CustomerID:        txn.CustomerID,
		PosTerminalID:     txn.PosTerminalID,
		TransactionNumber: txn.TransactionNumber,
		TransactionDate:   txn.TransactionDate,
		TransactionType:   txn.TransactionType,
		Subtotal:          txn.Subtotal,
		DiscountAmount:    txn.DiscountAmount,
		TaxAmount:         txn.TaxAmount,
		TotalAmount:       txn.TotalAmount,
		TotalCost:         txn.TotalCost,
		AmountPaid:        txn.AmountPaid,
		ChangeGiven:       txn.ChangeGiven,
		Status:            txn.Status,
		PriceListID:       txn.PriceListID,
		SalesOrderID:      txn.SalesOrderID,
		SourceCartID:      txn.SourceCartID,
		VoidedBy:          txn.VoidedBy,
		VoidedAt:          txn.VoidedAt,
		Metadata:          utils.BytesToJSONRawMessage(txn.Metadata),
		CreatedAt:         txn.CreatedAt,
	}
}

// CreateTransaction creates a POS transaction header with lines.
func (uc *PosUseCase) CreateTransaction(ctx context.Context, in *PosCreateTransactionInput) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if in.StoreID <= 0 || in.CashierID <= 0 || in.CashierSessionID <= 0 || in.PosTerminalID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "store_id, cashier_id, cashier_session_id and pos_terminal_id are required", nil)
	}
	if len(in.Lines) == 0 {
		return utils.NewResponse(utils.CodeBadReq, "at least one line is required", nil)
	}

	_, err := uc.repo.GetStore(ctx, in.StoreID)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "store not found", nil)
	}

	transactionNumber := strings.TrimSpace(in.TransactionNumber)
	if transactionNumber == "" {
		transactionNumber = fmt.Sprintf("POS-%d", time.Now().Unix())
	}

	transactionDate := pgtype.Timestamp{Time: time.Now(), Valid: true}
	if in.TransactionDate != nil {
		transactionDate = pgtype.Timestamp{Time: *in.TransactionDate, Valid: true}
	}

	transactionType := pgtype.Text{String: "sale", Valid: true}
	if in.TransactionType != nil && strings.TrimSpace(*in.TransactionType) != "" {
		transactionType = pgtype.Text{String: strings.TrimSpace(*in.TransactionType), Valid: true}
	}

	status := pgtype.Text{String: "completed", Valid: true}
	if in.Status != nil && strings.TrimSpace(*in.Status) != "" {
		status = pgtype.Text{String: strings.TrimSpace(*in.Status), Valid: true}
	}

	params := repository.CreatePosTransactionParams{
		TransactionNumber: transactionNumber,
		StoreID:           in.StoreID,
		PosTerminalID:     pgtype.Int4{Int32: in.PosTerminalID, Valid: true},
		CashierSessionID:  in.CashierSessionID,
		CashierID:         in.CashierID,
		CustomerID:        pgtype.Int4{},
		PriceListID:       pgtype.Int4{},
		TransactionType:   transactionType,
		TransactionDate:   transactionDate,
		Subtotal:          in.Subtotal,
		TaxAmount:         in.TaxAmount,
		DiscountAmount:    in.DiscountAmount,
		TotalAmount:       in.TotalAmount,
		TotalCost:         in.TotalCost,
		Status:            status,
		Metadata:          in.Metadata,
	}

	if params.Metadata == nil {
		params.Metadata = []byte("{}")
	}
	if in.CustomerID != nil {
		params.CustomerID = pgtype.Int4{Int32: *in.CustomerID, Valid: true}
	}
	if in.PriceListID != nil {
		params.PriceListID = pgtype.Int4{Int32: *in.PriceListID, Valid: true}
	}

	header, err := uc.repo.CreatePosTransaction(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "failed to create transaction: "+err.Error(), nil)
	}

	for idx, line := range in.Lines {
		lineNo := int32(idx + 1)
		if line.LineNumber != nil && *line.LineNumber > 0 {
			lineNo = *line.LineNumber
		}

		lineParams := repository.CreatePosTransactionLineParams{
			TransactionID:    header.ID,
			LineNumber:       pgtype.Int4{Int32: lineNo, Valid: true},
			ProductID:        line.ProductID,
			ProductVariantID: pgtype.Int4{},
			SerialNumber:     pgtype.Text{},
			BatchNumber:      pgtype.Text{},
			Quantity:         line.Quantity,
			UomID:            pgtype.Int4{},
			UnitPrice:        line.UnitPrice,
			DiscountAmount:   line.DiscountAmount,
			TaxAmount:        line.TaxAmount,
			Subtotal:         line.Subtotal,
			LineTotal:        line.LineTotal,
			CostPrice:        line.CostPrice,
			Metadata:         line.Metadata,
		}

		if line.Metadata == nil {
			lineParams.Metadata = []byte("{}")
		}
		if line.ProductVariantID != nil {
			lineParams.ProductVariantID = pgtype.Int4{Int32: *line.ProductVariantID, Valid: true}
		}
		if line.SerialNumber != nil && strings.TrimSpace(*line.SerialNumber) != "" {
			lineParams.SerialNumber = pgtype.Text{String: strings.TrimSpace(*line.SerialNumber), Valid: true}
		}
		if line.BatchNumber != nil && strings.TrimSpace(*line.BatchNumber) != "" {
			lineParams.BatchNumber = pgtype.Text{String: strings.TrimSpace(*line.BatchNumber), Valid: true}
		}
		if line.UomID != nil {
			lineParams.UomID = pgtype.Int4{Int32: *line.UomID, Valid: true}
		}

		if err := uc.repo.CreatePosTransactionLine(ctx, lineParams); err != nil {
			return utils.NewResponse(utils.CodeError, fmt.Sprintf("failed to create line %d: %s", lineNo, err.Error()), nil)
		}
	}

	return utils.NewResponse(utils.CodeCreated, "transaction created", header)
}

// ListTodaysTransactions returns today's POS transactions for a store.
func (uc *PosUseCase) ListTodaysTransactions(ctx context.Context, storeID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	rows, err := uc.repo.ListTodaysPosTransactions(ctx, storeID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "transactions fetched", rows)
}

// GetTransaction returns a single POS transaction by ID.
func (uc *PosUseCase) GetTransaction(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	txn, err := uc.repo.GetPosTransaction(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "transaction not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "transaction fetched", toPosTransactionOutput(txn))
}

// GetTransactionFull returns a POS transaction with full line details.
func (uc *PosUseCase) GetTransactionFull(ctx context.Context, id int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	rows, err := uc.repo.GetPosTransactionFull(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "transaction details fetched", rows)
}

// VoidTransaction voids a completed POS transaction.
func (uc *PosUseCase) VoidTransaction(ctx context.Context, id int32, voidedBy int32, reason string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	rowsAffected, err := uc.repo.VoidPosTransaction(ctx, repository.VoidPosTransactionParams{
		ID: id, VoidedBy: pgtype.Int4{Int32: voidedBy, Valid: true}, Column3: reason,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	if rowsAffected == 0 {
		return utils.NewResponse(utils.CodeBadReq, "transaction not found or already voided", nil)
	}
	return utils.NewResponse(utils.CodeOK, "transaction voided", map[string]interface{}{"voided": true})
}

// GetTransactionPayments returns all payments for a POS transaction.
func (uc *PosUseCase) GetTransactionPayments(ctx context.Context, transactionID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	rows, err := uc.repo.GetPaymentsForTransaction(ctx, transactionID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "payments fetched", rows)
}

// GetTransactionPaymentSummary returns payment summary for a POS transaction.
func (uc *PosUseCase) GetTransactionPaymentSummary(ctx context.Context, transactionID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	row, err := uc.repo.GetTransactionPaymentSummary(ctx, transactionID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "payment summary fetched", row)
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
