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

type StockMovementExtractor struct {
	mssql *db.MSSQLClient
}

func NewStockMovementExtractor(mssql *db.MSSQLClient) *StockMovementExtractor {
	return &StockMovementExtractor{mssql: mssql}
}

// ExtractStockMovements extracts stock movements from OINM (Item Ledger) within the given date range.
func (e *StockMovementExtractor) ExtractStockMovements(ctx context.Context, fromDate, toDate time.Time) ([]mappings.CanonicalStockMovement, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	if fromDate.IsZero() {
		fromDate = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if toDate.IsZero() {
		toDate = time.Now()
	}

	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	rows, err := e.mssql.DB.QueryContext(queryCtx, schema.QueryStockMovementsOINM,
		sql.Named("FromDate", fromDate),
		sql.Named("ToDate", toDate),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query OINM: %w", err)
	}
	defer rows.Close()

	var result []mappings.CanonicalStockMovement
	for rows.Next() {
		var sm mappings.SAPStockMovement
		var baseRef, itemCode, warehouse sql.NullString
		if err := rows.Scan(
			&sm.TransNum,
			&sm.TransType,
			&sm.CreatedBy,
			&baseRef,
			&sm.DocDate,
			&itemCode,
			&warehouse,
			&sm.InQty,
			&sm.OutQty,
			&sm.Price,
			&sm.TransValue,
		); err != nil {
			return nil, fmt.Errorf("failed to scan OINM row: %w", err)
		}
		sm.BaseRef = baseRef.String
		sm.ItemCode = itemCode.String
		sm.Warehouse = warehouse.String

		result = append(result, sm.ToCanonical())
	}

	return result, nil
}
