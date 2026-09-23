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

type PaymentExtractor struct {
	mssql *db.MSSQLClient
}

func NewPaymentExtractor(mssql *db.MSSQLClient) *PaymentExtractor {
	return &PaymentExtractor{mssql: mssql}
}

// ExtractIncomingPayments extracts incoming payments from ORCT and linked invoices from RCT2 within the given date range.
func (e *PaymentExtractor) ExtractIncomingPayments(ctx context.Context, fromDate, toDate time.Time) ([]mappings.CanonicalIncomingPayment, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	if fromDate.IsZero() {
		fromDate = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if toDate.IsZero() {
		toDate = time.Now()
	}

	// 1. Extract Headers (ORCT)
	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryIncomingPaymentsHeader,
		sql.Named("FromDate", fromDate),
		sql.Named("ToDate", toDate),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query ORCT: %w", err)
	}
	defer rows.Close()

	var payments []mappings.SAPIncomingPayment
	for rows.Next() {
		var p mappings.SAPIncomingPayment
		var trsfrRef, comments, jrnlMemo sql.NullString
		if err := rows.Scan(
			&p.DocEntry,
			&p.DocNum,
			&p.DocDate,
			&p.CardCode,
			&p.CardName,
			&p.DocCurr,
			&p.DocTotal,
			&p.CashSum,
			&p.TrsfrSum,
			&trsfrRef,
			&p.CheckSum,
			&p.CreditSum,
			&comments,
			&jrnlMemo,
		); err != nil {
			return nil, fmt.Errorf("failed to scan ORCT row: %w", err)
		}
		p.TrsfrRef = trsfrRef.String
		p.Comments = comments.String
		p.JrnlMemo = jrnlMemo.String
		payments = append(payments, p)
	}

	if len(payments) == 0 {
		return []mappings.CanonicalIncomingPayment{}, nil
	}

	// 2. Extract Linked Invoices (RCT2) in chunked batches by DocEntry
	invoicesMap := make(map[int64][]mappings.SAPIncomingPaymentInvoice, len(payments))
	const docEntryBatchSize = 500
	for i := 0; i < len(payments); i += docEntryBatchSize {
		end := i + docEntryBatchSize
		if end > len(payments) {
			end = len(payments)
		}
		chunk := payments[i:end]
		var idStrs []string
		for _, p := range chunk {
			idStrs = append(idStrs, fmt.Sprintf("%d", p.DocEntry))
		}
		if len(idStrs) == 0 {
			continue
		}
		query := fmt.Sprintf(schema.QueryIncomingPaymentInvoices, strings.Join(idStrs, ","))
		invRows, err := e.mssql.DB.QueryContext(ctx, query)
		if err == nil {
			for invRows.Next() {
				var inv mappings.SAPIncomingPaymentInvoice
				if err := invRows.Scan(
					&inv.DocNum,
					&inv.LineID,
					&inv.InvoiceID,
					&inv.InvType,
					&inv.SumApplied,
					&inv.InvoiceDocNum,
				); err == nil {
					invoicesMap[inv.DocNum] = append(invoicesMap[inv.DocNum], inv)
				}
			}
			invRows.Close()
		}
	}

	// 3. Map to Canonical List
	var result []mappings.CanonicalIncomingPayment
	for i := range payments {
		payments[i].Invoices = invoicesMap[payments[i].DocEntry]
		canonList := payments[i].ToCanonicalList()
		result = append(result, canonList...)
	}

	return result, nil
}
