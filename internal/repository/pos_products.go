package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

// PosGetProductsWithStockParams holds args for fn_pos_get_products_with_stock.
type PosGetProductsWithStockParams struct {
	StoreID           int32
	CategoryID        pgtype.Int4
	SearchTerm        pgtype.Text
	IncludeOutOfStock bool
}

// PosProductWithStockRow maps fn_pos_get_products_with_stock result.
type PosProductWithStockRow struct {
	ProductID         int32          `json:"product_id"`
	Sku               string         `json:"sku"`
	ProductName       string         `json:"product_name"`
	Description       pgtype.Text    `json:"description"`
	CategoryID        pgtype.Int4    `json:"category_id"`
	CategoryName      pgtype.Text    `json:"category_name"`
	BrandName         pgtype.Text    `json:"brand_name"`
	Barcode           pgtype.Text    `json:"barcode"`
	UomCode           pgtype.Text    `json:"uom_code"`
	DecimalPlaces     pgtype.Int4    `json:"decimal_places"`
	RetailPrice       pgtype.Numeric `json:"retail_price"`
	PromoPrice        pgtype.Numeric `json:"promo_price"`
	EffectivePrice    pgtype.Numeric `json:"effective_price"`
	HasPromotion      pgtype.Bool    `json:"has_promotion"`
	PromotionName     pgtype.Text    `json:"promotion_name"`
	DiscountPercent   pgtype.Text    `json:"discount_percent"`
	PromoMinQuantity  pgtype.Numeric `json:"promo_min_quantity"`
	TaxRate           pgtype.Numeric `json:"tax_rate"`
	TaxIsInclusive    pgtype.Bool    `json:"tax_is_inclusive"`
	QuantityAvailable pgtype.Numeric `json:"quantity_available"`
	QuantityOnHand    pgtype.Numeric `json:"quantity_on_hand"`
	QuantityAllocated pgtype.Numeric `json:"quantity_allocated"`
	IsInStock         pgtype.Bool    `json:"is_in_stock"`
	IsLowStock        pgtype.Bool    `json:"is_low_stock"`
	ReorderLevel      pgtype.Numeric `json:"reorder_level"`
	AllowDecimalQty   pgtype.Bool    `json:"allow_decimal_quantity"`
	IsSerialized      pgtype.Bool    `json:"is_serialized"`
	IsBatchManaged    pgtype.Bool    `json:"is_batch_managed"`
	ProductMetadata   []byte         `json:"product_metadata"`
}

// PosGetProductsWithStock calls fn_pos_get_products_with_stock.
func (q *Queries) PosGetProductsWithStock(ctx context.Context, arg PosGetProductsWithStockParams) ([]PosProductWithStockRow, error) {
	rows, err := q.db.Query(ctx, posGetProductsWithStockSQL,
		arg.StoreID, arg.CategoryID, arg.SearchTerm, arg.IncludeOutOfStock)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PosProductWithStockRow
	for rows.Next() {
		var i PosProductWithStockRow
		err := rows.Scan(
			&i.ProductID, &i.Sku, &i.ProductName, &i.Description, &i.CategoryID, &i.CategoryName,
			&i.BrandName, &i.Barcode, &i.UomCode, &i.DecimalPlaces, &i.RetailPrice, &i.PromoPrice,
			&i.EffectivePrice, &i.HasPromotion, &i.PromotionName, &i.DiscountPercent, &i.PromoMinQuantity,
			&i.TaxRate, &i.TaxIsInclusive, &i.QuantityAvailable, &i.QuantityOnHand, &i.QuantityAllocated,
			&i.IsInStock, &i.IsLowStock, &i.ReorderLevel, &i.AllowDecimalQty, &i.IsSerialized, &i.IsBatchManaged,
			&i.ProductMetadata,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const posGetProductsWithStockSQL = `SELECT product_id, sku, product_name, description, category_id, category_name, brand_name,
    barcode, uom_code, decimal_places, retail_price, promo_price, effective_price,
    has_promotion, promotion_name, discount_percent, promo_min_quantity, tax_rate, tax_is_inclusive,
    quantity_available, quantity_on_hand, quantity_allocated, is_in_stock, is_low_stock,
    reorder_level, allow_decimal_quantity, is_serialized, is_batch_managed, product_metadata
FROM fn_pos_get_products_with_stock($1, $2, $3, $4)`

// PosGetProductByBarcode calls fn_pos_get_product_by_barcode.
func (q *Queries) PosGetProductByBarcode(ctx context.Context, barcode string, storeID int32) (PosProductByBarcodeRow, error) {
	var i PosProductByBarcodeRow
	row := q.db.QueryRow(ctx, posGetProductByBarcodeSQL, barcode, storeID)
	err := row.Scan(
		&i.ProductID, &i.Sku, &i.ProductName, &i.Description, &i.CategoryName, &i.BrandName,
		&i.Barcode, &i.UomCode, &i.DecimalPlaces, &i.RetailPrice, &i.PromoPrice, &i.EffectivePrice,
		&i.HasPromotion, &i.PromotionName, &i.PromoMinQuantity, &i.TaxRate, &i.TaxIsInclusive,
		&i.QuantityAvailable, &i.IsInStock, &i.AllowDecimalQty, &i.IsSerialized, &i.IsBatchManaged,
		&i.ProductMetadata,
	)
	return i, err
}

// PosProductByBarcodeRow maps fn_pos_get_product_by_barcode result.
type PosProductByBarcodeRow struct {
	ProductID         int32          `json:"product_id"`
	Sku               string         `json:"sku"`
	ProductName       string         `json:"product_name"`
	Description       pgtype.Text    `json:"description"`
	CategoryName      pgtype.Text    `json:"category_name"`
	BrandName         pgtype.Text    `json:"brand_name"`
	Barcode           pgtype.Text    `json:"barcode"`
	UomCode           pgtype.Text    `json:"uom_code"`
	DecimalPlaces     pgtype.Int4    `json:"decimal_places"`
	RetailPrice       pgtype.Numeric `json:"retail_price"`
	PromoPrice        pgtype.Numeric `json:"promo_price"`
	EffectivePrice    pgtype.Numeric `json:"effective_price"`
	HasPromotion      pgtype.Bool    `json:"has_promotion"`
	PromotionName     pgtype.Text    `json:"promotion_name"`
	PromoMinQuantity  pgtype.Numeric `json:"promo_min_quantity"`
	TaxRate           pgtype.Numeric `json:"tax_rate"`
	TaxIsInclusive    pgtype.Bool    `json:"tax_is_inclusive"`
	QuantityAvailable pgtype.Numeric `json:"quantity_available"`
	IsInStock         pgtype.Bool    `json:"is_in_stock"`
	AllowDecimalQty   pgtype.Bool    `json:"allow_decimal_quantity"`
	IsSerialized      pgtype.Bool    `json:"is_serialized"`
	IsBatchManaged    pgtype.Bool    `json:"is_batch_managed"`
	ProductMetadata   []byte         `json:"product_metadata"`
}

const posGetProductByBarcodeSQL = `SELECT product_id, sku, product_name, description, category_name, brand_name,
    barcode, uom_code, decimal_places, retail_price, promo_price, effective_price,
    has_promotion, promotion_name, promo_min_quantity, tax_rate, tax_is_inclusive,
    quantity_available, is_in_stock, allow_decimal_quantity, is_serialized, is_batch_managed, product_metadata
FROM fn_pos_get_product_by_barcode($1, $2) LIMIT 1`

// PosSearchProductsParams holds args for fn_pos_search_products.
type PosSearchProductsParams struct {
	SearchTerm string
	StoreID    int32
	Limit      int32
}

// PosSearchProductRow maps fn_pos_search_products result.
type PosSearchProductRow struct {
	ProductID         int32          `json:"product_id"`
	Sku               string         `json:"sku"`
	ProductName       string         `json:"product_name"`
	CategoryName      pgtype.Text    `json:"category_name"`
	BrandName         pgtype.Text    `json:"brand_name"`
	Barcode           pgtype.Text    `json:"barcode"`
	EffectivePrice    pgtype.Numeric `json:"effective_price"`
	HasPromotion      pgtype.Bool    `json:"has_promotion"`
	QuantityAvailable pgtype.Numeric `json:"quantity_available"`
	IsInStock         pgtype.Bool    `json:"is_in_stock"`
	RelevanceScore    pgtype.Int4    `json:"relevance_score"`
}

// PosSearchProducts calls fn_pos_search_products.
func (q *Queries) PosSearchProducts(ctx context.Context, arg PosSearchProductsParams) ([]PosSearchProductRow, error) {
	limit := arg.Limit
	if limit <= 0 {
		limit = 50
	}
	rows, err := q.db.Query(ctx, posSearchProductsSQL, arg.SearchTerm, arg.StoreID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PosSearchProductRow
	for rows.Next() {
		var i PosSearchProductRow
		err := rows.Scan(
			&i.ProductID, &i.Sku, &i.ProductName, &i.CategoryName, &i.BrandName, &i.Barcode,
			&i.EffectivePrice, &i.HasPromotion, &i.QuantityAvailable, &i.IsInStock, &i.RelevanceScore,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const posSearchProductsSQL = `SELECT product_id, sku, product_name, category_name, brand_name, barcode,
    effective_price, has_promotion, quantity_available, is_in_stock, relevance_score
FROM fn_pos_search_products($1, $2, $3)`

// PosGetProductsByCategoryParams holds args for fn_pos_get_products_by_category.
type PosGetProductsByCategoryParams struct {
	CategoryID           int32
	StoreID              int32
	IncludeSubcategories bool
}

// PosProductByCategoryRow maps fn_pos_get_products_by_category result.
type PosProductByCategoryRow struct {
	ProductID         int32          `json:"product_id"`
	Sku               string         `json:"sku"`
	ProductName       string         `json:"product_name"`
	CategoryName      pgtype.Text    `json:"category_name"`
	BrandName         pgtype.Text    `json:"brand_name"`
	Barcode           pgtype.Text    `json:"barcode"`
	EffectivePrice    pgtype.Numeric `json:"effective_price"`
	HasPromotion      pgtype.Bool    `json:"has_promotion"`
	PromotionName     pgtype.Text    `json:"promotion_name"`
	QuantityAvailable pgtype.Numeric `json:"quantity_available"`
	IsInStock         pgtype.Bool    `json:"is_in_stock"`
}

// PosGetProductsByCategory calls fn_pos_get_products_by_category.
func (q *Queries) PosGetProductsByCategory(ctx context.Context, arg PosGetProductsByCategoryParams) ([]PosProductByCategoryRow, error) {
	rows, err := q.db.Query(ctx, posGetProductsByCategorySQL, arg.CategoryID, arg.StoreID, arg.IncludeSubcategories)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PosProductByCategoryRow
	for rows.Next() {
		var i PosProductByCategoryRow
		err := rows.Scan(
			&i.ProductID, &i.Sku, &i.ProductName, &i.CategoryName, &i.BrandName, &i.Barcode,
			&i.EffectivePrice, &i.HasPromotion, &i.PromotionName, &i.QuantityAvailable, &i.IsInStock,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const posGetProductsByCategorySQL = `SELECT product_id, sku, product_name, category_name, brand_name, barcode,
    effective_price, has_promotion, promotion_name, quantity_available, is_in_stock
FROM fn_pos_get_products_by_category($1, $2, $3)`

// PosCategoryRow maps vw_pos_categories result.
type PosCategoryRow struct {
	CategoryID         int32       `json:"category_id"`
	CategoryCode       pgtype.Text `json:"category_code"`
	CategoryName       pgtype.Text `json:"category_name"`
	ParentCategoryID   pgtype.Int4 `json:"parent_category_id"`
	ParentCategoryName pgtype.Text `json:"parent_category_name"`
	ProductCount       int32       `json:"product_count"`
	InStockCount       int32       `json:"in_stock_count"`
	CategoryMetadata   []byte      `json:"category_metadata"`
}

// PosGetCategories returns rows from vw_pos_categories.
func (q *Queries) PosGetCategories(ctx context.Context) ([]PosCategoryRow, error) {
	rows, err := q.db.Query(ctx, posGetCategoriesSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PosCategoryRow
	for rows.Next() {
		var i PosCategoryRow
		err := rows.Scan(
			&i.CategoryID, &i.CategoryCode, &i.CategoryName, &i.ParentCategoryID, &i.ParentCategoryName,
			&i.ProductCount, &i.InStockCount, &i.CategoryMetadata,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const posGetCategoriesSQL = `SELECT category_id, category_code, category_name, parent_category_id, parent_category_name,
    product_count, in_stock_count, category_metadata
FROM vw_pos_categories`

// ParseNumeric parses a decimal string into pgtype.Numeric via SELECT $1::numeric.
func (q *Queries) ParseNumeric(ctx context.Context, s string) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	row := q.db.QueryRow(ctx, `SELECT $1::numeric`, s)
	err := row.Scan(&n)
	return n, err
}
