package extractors

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap/mappings"
	"github.com/NasTecSol/nembus-sap/schema"
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
	}

	if len(orders) == 0 {
		return []mappings.CanonicalSalesOrder{}, nil
	}

	// 2. Extract Lines (RDR1) in chunked batches by DocEntry for fast indexed lookups
	linesMap := make(map[int64][]mappings.SAPSalesOrderLine, len(orders))
	const docEntryBatchSize = 500
	for i := 0; i < len(orders); i += docEntryBatchSize {
		end := i + docEntryBatchSize
		if end > len(orders) {
			end = len(orders)
		}
		chunk := orders[i:end]
		var idStrs []string
		for _, o := range chunk {
			idStrs = append(idStrs, fmt.Sprintf("%d", o.DocEntry))
		}
		if len(idStrs) == 0 {
			continue
		}
		query := fmt.Sprintf(schema.QuerySalesOrderLines, strings.Join(idStrs, ","))
		lineRows, err := e.mssql.DB.QueryContext(ctx, query)
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
		// Scan order must match QueryInvoicesHeader exactly (17 columns):
		// A02 (DocCur/DocRate), A03 (CANCELED), A04 (DocType) were added to the
		// query but the scan was never extended — a mismatch makes every
		// ExtractInvoices call fail with a Scan argument-count error.
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
			&inv.DocCur,
			&inv.DocRate,
			&inv.Canceled,
			&inv.DocType,
		); err != nil {
			return nil, fmt.Errorf("failed to scan OINV row: %w", err)
		}
		inv.Comments = comments.String
		invoices = append(invoices, inv)
	}

	if len(invoices) == 0 {
		return []mappings.CanonicalInvoice{}, nil
	}

	// 2. Extract Lines (INV1) in chunked batches by DocEntry (clustered index seek)
	linesMap := make(map[int64][]mappings.SAPInvoiceLine, len(invoices))
	const docEntryBatchSize = 500
	for i := 0; i < len(invoices); i += docEntryBatchSize {
		end := i + docEntryBatchSize
		if end > len(invoices) {
			end = len(invoices)
		}
		chunk := invoices[i:end]
		var idStrs []string
		for _, inv := range chunk {
			idStrs = append(idStrs, fmt.Sprintf("%d", inv.DocEntry))
		}
		if len(idStrs) == 0 {
			continue
		}
		query := fmt.Sprintf(schema.QueryInvoiceLines, strings.Join(idStrs, ","))
		lineRows, err := e.mssql.DB.QueryContext(ctx, query)
		if err == nil {
			for lineRows.Next() {
				var l mappings.SAPInvoiceLine
				var whsCode, unitMsr sql.NullString
				// Scan order must match QueryInvoiceLines exactly (12 columns),
				// including the A01 discount evidence (DiscPrcnt, PriceBefDi).
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
					&l.DiscPrcnt,
					&l.PriceBefDi,
				); err == nil {
					l.WhsCode = whsCode.String
					l.UnitMsr = unitMsr.String
					linesMap[l.DocEntry] = append(linesMap[l.DocEntry], l)
				}
			}
			lineRows.Close()
		}
	}

	// 3. Transform to Canonical
	result := make([]mappings.CanonicalInvoice, len(invoices))
	for idx, inv := range invoices {
		inv.Lines = linesMap[inv.DocEntry]
		result[idx] = inv.ToCanonical()
	}

	return result, nil
}

// ExtractSalesReturns extracts A/R Credit Memos (ORIN/RIN1) and A/R Returns
// (ORDN/RDN1) within the given date range — A21. docSource selects the SAP
// document family: "credit_memo" (ORIN) or "return" (ORDN).
func (e *SalesExtractor) ExtractSalesReturns(ctx context.Context, fromDate, toDate time.Time, docSource string) ([]mappings.CanonicalSalesReturn, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	if fromDate.IsZero() {
		fromDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if toDate.IsZero() {
		toDate = time.Now()
	}

	// A21: select the document family — credit memos (ORIN) or returns (ORDN)
	var headerQuery, lineQuery, headerName string
	switch docSource {
	case "return":
		headerQuery = schema.QueryReturnOrdersHeader
		lineQuery = schema.QueryReturnOrderLines
		headerName = "ORDN"
	default:
		docSource = "credit_memo"
		headerQuery = schema.QueryCreditMemosHeader
		lineQuery = schema.QueryCreditMemoLines
		headerName = "ORIN"
	}

	// 1. Extract Headers
	rows, err := e.mssql.DB.QueryContext(ctx, headerQuery,
		sql.Named("FromDate", fromDate),
		sql.Named("ToDate", toDate),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query %s: %w", headerName, err)
	}
	defer rows.Close()

	var returns []mappings.SAPSalesReturn
	for rows.Next() {
		var sr mappings.SAPSalesReturn
		// ISNULL(Comments,'') in the query allows scanning directly into the string field
		if err := rows.Scan(
			&sr.DocEntry,
			&sr.DocNum,
			&sr.DocDate,
			&sr.CardCode,
			&sr.CardName,
			&sr.DocTotal,
			&sr.VatSum,
			&sr.DiscSum,
			&sr.Canceled,
			&sr.Comments,
		); err != nil {
			return nil, fmt.Errorf("failed to scan %s row: %w", headerName, err)
		}
		returns = append(returns, sr)
	}

	if len(returns) == 0 {
		return []mappings.CanonicalSalesReturn{}, nil
	}

	// 2. Extract Lines in chunked batches by DocEntry
	linesMap := make(map[int64][]mappings.SAPSalesReturnLine, len(returns))
	const docEntryBatchSize = 500
	for i := 0; i < len(returns); i += docEntryBatchSize {
		end := i + docEntryBatchSize
		if end > len(returns) {
			end = len(returns)
		}
		chunk := returns[i:end]
		var idStrs []string
		for _, sr := range chunk {
			idStrs = append(idStrs, fmt.Sprintf("%d", sr.DocEntry))
		}
		if len(idStrs) == 0 {
			continue
		}
		query := fmt.Sprintf(lineQuery, strings.Join(idStrs, ","))
		lineRows, err := e.mssql.DB.QueryContext(ctx, query)
		if err == nil {
			for lineRows.Next() {
				var l mappings.SAPSalesReturnLine
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
					&l.BaseEntry,
					&l.BaseType,
				); err == nil {
					l.WhsCode = whsCode.String
					l.UnitMsr = unitMsr.String
					linesMap[l.DocEntry] = append(linesMap[l.DocEntry], l)
				}
			}
			lineRows.Close()
		}
	}

	// 3. Transform to Canonical
	result := make([]mappings.CanonicalSalesReturn, len(returns))
	for idx := range returns {
		returns[idx].DocSource = docSource
		returns[idx].Lines = linesMap[returns[idx].DocEntry]
		result[idx] = returns[idx].ToCanonical()
	}

	return result, nil
}
