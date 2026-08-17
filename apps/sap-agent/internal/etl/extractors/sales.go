package extractors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/NasTecSol/nembus-sap/mappings"
	"github.com/NasTecSol/nembus-sap/schema"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
)

type SalesExtractor struct {
	mssql *db.MSSQLClient
}

func NewSalesExtractor(mssql *db.MSSQLClient) *SalesExtractor {
	return &SalesExtractor{mssql: mssql}
}

func (e *SalesExtractor) ExtractSalesOrders(ctx context.Context, fromDate, toDate time.Time) ([]mappings.CanonicalSalesOrder, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	if fromDate.IsZero() {
		fromDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if toDate.IsZero() {
		toDate = time.Now().AddDate(1, 0, 0)
	}

	// 1. Extract Headers (ORDR)
	rows, err := e.mssql.DB.QueryContext(ctx, schema.QuerySalesOrdersHeader,
		sql.Named("FromDate", fromDate),
		sql.Named("ToDate", toDate),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query ORDR: %w", err)
	}
	defer rows.Close()

	var orders []mappings.SAPSalesOrder
	docEntries := make([]int64, 0)
	for rows.Next() {
		var o mappings.SAPSalesOrder
		var comments sql.NullString
		if err := rows.Scan(
			&o.DocEntry,
			&o.DocNum,
			&o.DocDate,
			&o.DocDueDate,
			&o.CardCode,
			&o.CardName,
			&o.DocTotal,
			&o.VatSum,
			&o.DiscSum,
			&o.DocStatus,
			&o.SlpCode,
			&comments,
		); err != nil {
			return nil, fmt.Errorf("failed to scan ORDR row: %w", err)
		}
		o.Comments = comments.String
		orders = append(orders, o)
		docEntries = append(docEntries, o.DocEntry)
	}

	if len(orders) == 0 {
		return []mappings.CanonicalSalesOrder{}, nil
	}

	// 2. Extract Lines (RDR1) in a single fast JOIN query
	linesMap := make(map[int64][]mappings.SAPSalesOrderLine)
	lineRows, err := e.mssql.DB.QueryContext(ctx, schema.QuerySalesOrderLinesByDate,
		sql.Named("FromDate", fromDate),
		sql.Named("ToDate", toDate),
	)
	if err == nil {
		for lineRows.Next() {
			var l mappings.SAPSalesOrderLine
			var whsCode, unitMsr sql.NullString
			if err := lineRows.Scan(
				&l.DocEntry,
				&l.LineNum,
				&l.ItemCode,
				&l.Dscription,
				&l.Quantity,
				&l.Price,
				&l.LineTotal,
				&l.VatSum,
				&whsCode,
				&unitMsr,
			); err == nil {
				l.WhsCode = whsCode.String
				l.UnitMsr = unitMsr.String
				linesMap[l.DocEntry] = append(linesMap[l.DocEntry], l)
			}
		}
		lineRows.Close()
	}

	// 3. Transform to Canonical
	result := make([]mappings.CanonicalSalesOrder, len(orders))
	for idx, o := range orders {
		o.Lines = linesMap[o.DocEntry]
		result[idx] = o.ToCanonical()
	}

	return result, nil
}

func (e *SalesExtractor) ExtractInvoices(ctx context.Context, fromDate, toDate time.Time) ([]mappings.CanonicalInvoice, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	if fromDate.IsZero() {
		fromDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if toDate.IsZero() {
		toDate = time.Now().AddDate(1, 0, 0)
	}

	// 1. Extract Headers (OINV)
	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryInvoicesHeader,
		sql.Named("FromDate", fromDate),
		sql.Named("ToDate", toDate),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query OINV: %w", err)
	}
	defer rows.Close()

	var invoices []mappings.SAPInvoice
	for rows.Next() {
		var inv mappings.SAPInvoice
		var comments sql.NullString
		if err := rows.Scan(
			&inv.DocEntry,
			&inv.DocNum,
			&inv.DocDate,
			&inv.DocDueDate,
			&inv.CardCode,
			&inv.CardName,
			&inv.DocTotal,
			&inv.PaidToDate,
			&inv.VatSum,
			&inv.DiscSum,
			&inv.DocStatus,
			&inv.SlpCode,
			&comments,
		); err != nil {
			return nil, fmt.Errorf("failed to scan OINV row: %w", err)
		}
		inv.Comments = comments.String
		invoices = append(invoices, inv)
	}

	if len(invoices) == 0 {
		return []mappings.CanonicalInvoice{}, nil
	}

	// 2. Extract Lines (INV1) in a single fast JOIN query
	linesMap := make(map[int64][]mappings.SAPInvoiceLine)
	lineRows, err := e.mssql.DB.QueryContext(ctx, schema.QueryInvoiceLinesByDate,
		sql.Named("FromDate", fromDate),
		sql.Named("ToDate", toDate),
	)
	if err == nil {
		for lineRows.Next() {
			var l mappings.SAPInvoiceLine
			var whsCode, unitMsr sql.NullString
			if err := lineRows.Scan(
				&l.DocEntry,
				&l.LineNum,
				&l.ItemCode,
				&l.Dscription,
				&l.Quantity,
				&l.Price,
				&l.LineTotal,
				&l.VatSum,
				&whsCode,
				&unitMsr,
			); err == nil {
				l.WhsCode = whsCode.String
				l.UnitMsr = unitMsr.String
				linesMap[l.DocEntry] = append(linesMap[l.DocEntry], l)
			}
		}
		lineRows.Close()
	}

	// 3. Transform to Canonical
	result := make([]mappings.CanonicalInvoice, len(invoices))
	for idx, inv := range invoices {
		inv.Lines = linesMap[inv.DocEntry]
		result[idx] = inv.ToCanonical()
	}

	return result, nil
}
