package extractors

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-sap/mappings"
	"github.com/NasTecSol/nembus-sap/schema"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
)

type PurchasingExtractor struct {
	mssql *db.MSSQLClient
}

func NewPurchasingExtractor(mssql *db.MSSQLClient) *PurchasingExtractor {
	return &PurchasingExtractor{mssql: mssql}
}

// ExtractPurchaseOrders extracts Purchase Orders from OPOR and POR1 within the given date range.
func (e *PurchasingExtractor) ExtractPurchaseOrders(ctx context.Context, fromDate, toDate time.Time) ([]mappings.CanonicalPurchaseOrder, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	if fromDate.IsZero() {
		fromDate = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if toDate.IsZero() {
		toDate = time.Now().AddDate(1, 0, 0)
	}

	// 1. Extract Headers (OPOR)
	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryPurchaseOrdersHeader,
		sql.Named("FromDate", fromDate),
		sql.Named("ToDate", toDate),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query OPOR: %w", err)
	}
	defer rows.Close()

	var orders []mappings.SAPPurchaseOrder
	for rows.Next() {
		var o mappings.SAPPurchaseOrder
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
			return nil, fmt.Errorf("failed to scan OPOR row: %w", err)
		}
		o.Comments = comments.String
		orders = append(orders, o)
	}

	if len(orders) == 0 {
		return []mappings.CanonicalPurchaseOrder{}, nil
	}

	// 2. Extract Lines (POR1) in chunked batches by DocEntry for fast indexed lookups
	linesMap := make(map[int64][]mappings.SAPPurchaseOrderLine, len(orders))
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
		query := fmt.Sprintf(schema.QueryPurchaseOrderLines, strings.Join(idStrs, ","))
		lineRows, err := e.mssql.DB.QueryContext(ctx, query)
		if err == nil {
			for lineRows.Next() {
				var l mappings.SAPPurchaseOrderLine
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
					&l.OpenQty,
				); err == nil {
					l.WhsCode = whsCode.String
					l.UnitMsr = unitMsr.String
					linesMap[l.DocEntry] = append(linesMap[l.DocEntry], l)
				}
			}
			lineRows.Close()
		}
	}

	// 3. Map to Canonical
	result := make([]mappings.CanonicalPurchaseOrder, len(orders))
	for i := range orders {
		orders[i].Lines = linesMap[orders[i].DocEntry]
		result[i] = orders[i].ToCanonical()
	}

	return result, nil
}

// ExtractGoodsReceipts extracts Goods Receipt POs from OPDN and PDN1 within the given date range.
func (e *PurchasingExtractor) ExtractGoodsReceipts(ctx context.Context, fromDate, toDate time.Time) ([]mappings.CanonicalGoodsReceiptNote, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	if fromDate.IsZero() {
		fromDate = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if toDate.IsZero() {
		toDate = time.Now().AddDate(1, 0, 0)
	}

	// 1. Extract Headers (OPDN)
	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryGoodsReceiptsHeader,
		sql.Named("FromDate", fromDate),
		sql.Named("ToDate", toDate),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query OPDN: %w", err)
	}
	defer rows.Close()

	var receipts []mappings.SAPGoodsReceipt
	for rows.Next() {
		var gr mappings.SAPGoodsReceipt
		var comments sql.NullString
		if err := rows.Scan(
			&gr.DocEntry,
			&gr.DocNum,
			&gr.DocDate,
			&gr.DocDueDate,
			&gr.CardCode,
			&gr.CardName,
			&gr.DocTotal,
			&gr.VatSum,
			&gr.DocStatus,
			&comments,
		); err != nil {
			return nil, fmt.Errorf("failed to scan OPDN row: %w", err)
		}
		gr.Comments = comments.String
		receipts = append(receipts, gr)
	}

	if len(receipts) == 0 {
		return []mappings.CanonicalGoodsReceiptNote{}, nil
	}

	// 2. Extract Lines (PDN1) in chunked batches by DocEntry
	linesMap := make(map[int64][]mappings.SAPGoodsReceiptLine, len(receipts))
	const docEntryBatchSize = 500
	for i := 0; i < len(receipts); i += docEntryBatchSize {
		end := i + docEntryBatchSize
		if end > len(receipts) {
			end = len(receipts)
		}
		chunk := receipts[i:end]
		var idStrs []string
		for _, r := range chunk {
			idStrs = append(idStrs, fmt.Sprintf("%d", r.DocEntry))
		}
		if len(idStrs) == 0 {
			continue
		}
		query := fmt.Sprintf(schema.QueryGoodsReceiptLines, strings.Join(idStrs, ","))
		lineRows, err := e.mssql.DB.QueryContext(ctx, query)
		if err == nil {
			for lineRows.Next() {
				var l mappings.SAPGoodsReceiptLine
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
					&l.BaseLine,
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

	// 3. Map to Canonical
	result := make([]mappings.CanonicalGoodsReceiptNote, len(receipts))
	for i := range receipts {
		receipts[i].Lines = linesMap[receipts[i].DocEntry]
		result[i] = receipts[i].ToCanonical()
	}

	return result, nil
}
