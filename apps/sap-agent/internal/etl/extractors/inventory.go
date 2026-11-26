package extractors

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-sap/mappings"
	"github.com/NasTecSol/nembus-sap/schema"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
)

type InventoryExtractor struct {
	mssql *db.MSSQLClient
}

func NewInventoryExtractor(mssql *db.MSSQLClient) *InventoryExtractor {
	return &InventoryExtractor{mssql: mssql}
}

func scanInventoryStockRows(rows *sql.Rows, binMap map[string][]mappings.SAPBinBalance) ([]mappings.CanonicalInventoryStock, error) {
	var stockList []mappings.CanonicalInventoryStock
	for rows.Next() {
		var inv mappings.SAPInventoryStock
		var binAbs sql.NullInt64
		var binCode sql.NullString
		if err := rows.Scan(
			&inv.ItemCode,
			&inv.WhsCode,
			&inv.OnHand,
			&inv.IsCommited,
			&inv.OnOrder,
			&inv.MinStock,
			&inv.MaxStock,
			&binAbs,
			&binCode,
		); err != nil {
			return nil, fmt.Errorf("failed to scan OITW row: %w", err)
		}
		if binAbs.Valid {
			inv.BinAbsEntry = binAbs.Int64
		}
		if binCode.Valid {
			inv.BinCode = binCode.String
		}
		canon := inv.ToCanonical()
		if len(binMap) > 0 {
			key := canon.ProductSKU + "|" + canon.StoreCode
			if bins, ok := binMap[key]; ok && len(bins) > 0 {
				if canon.Metadata == nil {
					canon.Metadata = make(map[string]interface{})
				}
				canon.Metadata["sap_bins"] = bins
			}
		}
		stockList = append(stockList, canon)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating OITW rows: %w", err)
	}
	return stockList, nil
}

func (e *InventoryExtractor) fetchBinBalanceMap(ctx context.Context) map[string][]mappings.SAPBinBalance {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil
	}
	qCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	rows, err := e.mssql.DB.QueryContext(qCtx, schema.QueryInventoryBinBalances)
	if err != nil {
		return nil
	}
	defer rows.Close()

	binMap := make(map[string][]mappings.SAPBinBalance)
	for rows.Next() {
		var itemCode, whsCode, binCode string
		var binAbs int64
		var qty float64
		if err := rows.Scan(&itemCode, &whsCode, &binAbs, &binCode, &qty); err != nil {
			continue
		}
		key := strings.TrimSpace(itemCode) + "|" + strings.TrimSpace(whsCode)
		binMap[key] = append(binMap[key], mappings.SAPBinBalance{
			BinAbs:    binAbs,
			BinCode:   strings.TrimSpace(binCode),
			OnHandQty: qty,
		})
	}
	return binMap
}

func (e *InventoryExtractor) ExtractInventory(ctx context.Context) ([]mappings.CanonicalInventoryStock, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryInventoryStock)
	if err != nil {
		rows, err = e.mssql.DB.QueryContext(ctx, schema.QueryInventoryStockFallback)
		if err != nil {
			return nil, fmt.Errorf("failed to query OITW: %w", err)
		}
	}
	defer rows.Close()

	binMap := e.fetchBinBalanceMap(ctx)
	return scanInventoryStockRows(rows, binMap)
}

// ExtractInventoryDelta extracts warehouse stock balances touched in SAP B1 since the given watermark.
func (e *InventoryExtractor) ExtractInventoryDelta(ctx context.Context, since time.Time) ([]mappings.CanonicalInventoryStock, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}
	sinceDate, sinceTime, sinceDateTime := sapDeltaParams(since)

	queryCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	rows, err := e.mssql.DB.QueryContext(queryCtx, schema.QueryInventoryStockDelta,
		sql.Named("SinceDate", sinceDate),
		sql.Named("SinceTime", sinceTime),
		sql.Named("SinceDateTime", sinceDateTime),
	)
	if err != nil {
		if ctx.Err() != nil || errors.Is(err, context.DeadlineExceeded) || strings.Contains(strings.ToLower(err.Error()), "timeout") {
			return nil, fmt.Errorf("failed to query OITW delta: %w", err)
		}
		fbCtx, fbCancel := context.WithTimeout(ctx, 25*time.Second)
		defer fbCancel()
		rows, err = e.mssql.DB.QueryContext(fbCtx, schema.QueryInventoryStockDeltaFallback,
			sql.Named("SinceDate", sinceDate),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query OITW delta: %w", err)
		}
	}
	defer rows.Close()

	binMap := e.fetchBinBalanceMap(ctx)
	return scanInventoryStockRows(rows, binMap)
}
