package extractors

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap/mappings"
	"github.com/NasTecSol/nembus-sap/schema"
)

type CatalogExtractor struct {
	mssql *db.MSSQLClient
}

func NewCatalogExtractor(mssql *db.MSSQLClient) *CatalogExtractor {
	return &CatalogExtractor{mssql: mssql}
}

func (e *CatalogExtractor) ExtractCategories(ctx context.Context) ([]mappings.CanonicalCategory, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryCategories)
	if err != nil {
		return nil, fmt.Errorf("failed to query OITB: %w", err)
	}
	defer rows.Close()

	var categories []mappings.CanonicalCategory
	for rows.Next() {
		var cat mappings.SAPCategory
		var name sql.NullString
		if err := rows.Scan(&cat.ItmsGrpCod, &name); err != nil {
			return nil, fmt.Errorf("failed to scan OITB row: %w", err)
		}
		cat.ItmsGrpNam = name.String
		if mappings.ClassifySAPProductType(cat.ItmsGrpNam) != "" {
			continue
		}
		categories = append(categories, cat.ToCanonical())
	}
	return categories, nil
}

func (e *CatalogExtractor) ExtractBrands(ctx context.Context) ([]mappings.CanonicalBrand, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryBrands)
	if err != nil {
		return nil, fmt.Errorf("failed to query OMRC: %w", err)
	}
	defer rows.Close()

	var brands []mappings.CanonicalBrand
	for rows.Next() {
		var b mappings.SAPBrand
		var name sql.NullString
		if err := rows.Scan(&b.FirmCode, &name); err != nil {
			return nil, fmt.Errorf("failed to scan OMRC row: %w", err)
		}
		b.FirmName = name.String
		brands = append(brands, b.ToCanonical())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate OMRC rows: %w", err)
	}
	return brands, nil
}

func (e *CatalogExtractor) ExtractProducts(ctx context.Context) ([]mappings.CanonicalProduct, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryProducts)
	if err != nil {
		return nil, fmt.Errorf("failed to query OITM: %w", err)
	}
	defer rows.Close()

	var products []mappings.CanonicalProduct
	for rows.Next() {
		var p mappings.SAPProduct
		var userText, itmGrpNam, codeBars, buyUnitMsr, salUnitMsr, invntryUom, vatGroup sql.NullString
		if err := rows.Scan(
			&p.ItemCode,
			&p.ItemName,
			&userText,
			&p.ItmsGrpCod,
			&itmGrpNam,
			&p.FirmCode,
			&p.InvntItem,
			&p.SellItem,
			&p.PrchseItem,
			&p.ValidFor,
			&codeBars,
			&buyUnitMsr,
			&salUnitMsr,
			&invntryUom,
			&p.NumInSale,
			&p.NumInBuy,
			&p.UgpEntry,
			&p.IUoMEntry,
			&p.SUoMEntry,
			&p.PUoMEntry,
			&p.ManSerNum,
			&p.ManBtchNum,
			&vatGroup,
		); err != nil {
			return nil, fmt.Errorf("failed to scan OITM row: %w", err)
		}
		p.UserText = userText.String
		p.ItmsGrpNam = itmGrpNam.String
		p.CodeBars = codeBars.String
		p.BuyUnitMsr = buyUnitMsr.String
		p.SalUnitMsr = salUnitMsr.String
		p.InvntryUom = invntryUom.String
		p.VatGourpSa = vatGroup.String

		products = append(products, p.ToCanonical())
	}
	return products, nil
}

func (e *CatalogExtractor) ExtractBarcodes(ctx context.Context) ([]mappings.CanonicalBarcode, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryProductBarcodes)
	if err != nil {
		// OBCD might be empty or not used in some SAP installations
		return []mappings.CanonicalBarcode{}, nil
	}
	defer rows.Close()

	var barcodes []mappings.CanonicalBarcode
	for rows.Next() {
		var b mappings.SAPBarcode
		var uomCode sql.NullString
		if err := rows.Scan(&b.BcdEntry, &b.BcdCode, &b.ItemCode, &b.UomEntry, &uomCode); err != nil {
			continue
		}
		b.UomCode = uomCode.String
		barcodes = append(barcodes, b.ToCanonical(false))
	}
	return barcodes, nil
}

// ExtractPriceLists extracts all price list headers from OPLN.
func (e *CatalogExtractor) ExtractPriceLists(ctx context.Context) ([]mappings.CanonicalPriceList, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryPriceLists)
	if err != nil {
		return nil, fmt.Errorf("failed to query OPLN: %w", err)
	}
	defer rows.Close()

	var lists []mappings.CanonicalPriceList
	for rows.Next() {
		var pl mappings.SAPPriceList
		var name, currency sql.NullString
		if err := rows.Scan(&pl.ListNum, &name, &currency, &pl.Factor, &pl.BasedOn, &pl.ValidFor); err != nil {
			return nil, fmt.Errorf("failed to scan OPLN row: %w", err)
		}
		pl.ListName = name.String
		pl.Currency = currency.String
		lists = append(lists, pl.ToCanonical())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate OPLN rows: %w", err)
	}
	return lists, nil
}

// ExtractPriceListItems extracts all item prices from ITM1.
func (e *CatalogExtractor) ExtractPriceListItems(ctx context.Context) ([]mappings.CanonicalPriceListItem, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryPriceListItems)
	if err != nil {
		return []mappings.CanonicalPriceListItem{}, nil
	}
	defer rows.Close()

	var items []mappings.CanonicalPriceListItem
	for rows.Next() {
		var item mappings.SAPPriceListItem
		var currency, uomCode sql.NullString
		if err := rows.Scan(&item.ItemCode, &item.PriceList, &item.Price, &currency, &item.UomEntry, &uomCode); err != nil {
			continue // Skip malformed rows
		}
		item.Currency = currency.String
		item.UomCode = uomCode.String
		items = append(items, item.ToCanonical())
	}
	return items, nil
}
