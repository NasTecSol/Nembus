package extractors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/NasTecSol/nembus-sap/mappings"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
)

type StockCountExtractor struct {
	mssql *db.MSSQLClient
}

func NewStockCountExtractor(mssql *db.MSSQLClient) *StockCountExtractor {
	return &StockCountExtractor{mssql: mssql}
}

func (e *StockCountExtractor) ExtractStockCounts(ctx context.Context, fromDate, toDate time.Time) ([]mappings.CanonicalStockCount, error) {
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

	q := `
	SELECT 
		TransNum, TransType, DocDate, Warehouse, ItemCode, InQty, OutQty, TransValue, CreatedBy, BASE_REF
	FROM OINM 
	WHERE TransType IN (59, 60, 162) 
	  AND DocDate >= @FromDate AND DocDate <= @ToDate
	ORDER BY TransType, TransNum, Warehouse
	`

	rows, err := e.mssql.DB.QueryContext(queryCtx, q,
		sql.Named("FromDate", fromDate),
		sql.Named("ToDate", toDate),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query OINM for stock counts: %w", err)
	}
	defer rows.Close()

	type groupKey struct {
		TransType int
		TransNum  int
		Warehouse string
	}

	groups := make(map[groupKey]*mappings.CanonicalStockCount)

	for rows.Next() {
		var transNum, transType, createdBy int
		var docDate time.Time
		var warehouse, itemCode, baseRef sql.NullString
		var inQty, outQty, transValue float64

		if err := rows.Scan(
			&transNum, &transType, &docDate, &warehouse, &itemCode, &inQty, &outQty, &transValue, &createdBy, &baseRef,
		); err != nil {
			return nil, fmt.Errorf("failed to scan stock count row: %w", err)
		}

		whs := warehouse.String
		key := groupKey{TransType: transType, TransNum: transNum, Warehouse: whs}
		
		refNum := baseRef.String
		if refNum == "" {
			refNum = fmt.Sprintf("%d", createdBy)
		}

		sc, exists := groups[key]
		if !exists {
			countType := "adjustment"
			if transType == 162 {
				countType = "recount"
			}
			sc = &mappings.CanonicalStockCount{
				CountNumber: fmt.Sprintf("SC-SAP-%d-%s", transType, refNum),
				StoreCode:   whs,
				CountType:   countType,
				Status:      "completed",
				CompletedAt: docDate,
				Metadata: map[string]interface{}{
					"sap_trans_num": transNum,
					"sap_trans_type": transType,
					"sap_doc_num":   refNum,
				},
				Lines: []mappings.CanonicalStockCountLine{},
			}
			groups[key] = sc
		}

		variance := inQty - outQty
		varianceVal := transValue
		if outQty > 0 {
			varianceVal = -transValue
		}

		sc.Lines = append(sc.Lines, mappings.CanonicalStockCountLine{
			ProductSKU:      itemCode.String,
			SystemQuantity:  0, // Since it's just an adjustment, we don't know the exact system quantity at that point from OINM
			CountedQuantity: variance,
			Variance:        variance,
			VarianceValue:   varianceVal,
			Metadata: map[string]interface{}{},
		})
	}

	var result []mappings.CanonicalStockCount
	for _, sc := range groups {
		result = append(result, *sc)
	}

	return result, nil
}
