package extractors

import (
	"context"
	"fmt"

	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap/mappings"
	"github.com/NasTecSol/nembus-sap/schema"
)

type InventoryExtractor struct {
	mssql *db.MSSQLClient
}

func NewInventoryExtractor(mssql *db.MSSQLClient) *InventoryExtractor {
	return &InventoryExtractor{mssql: mssql}
}

func (e *InventoryExtractor) ExtractInventory(ctx context.Context) ([]mappings.CanonicalInventoryStock, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryInventoryStock)
	if err != nil {
		return nil, fmt.Errorf("failed to query OITW: %w", err)
	}
	defer rows.Close()

	var stockList []mappings.CanonicalInventoryStock
	for rows.Next() {
		var inv mappings.SAPInventoryStock
		if err := rows.Scan(
			&inv.ItemCode,
			&inv.WhsCode,
			&inv.OnHand,
			&inv.IsCommited,
			&inv.OnOrder,
			&inv.MinStock,
			&inv.MaxStock,
		); err != nil {
			return nil, fmt.Errorf("failed to scan OITW row: %w", err)
		}

		stockList = append(stockList, inv.ToCanonical())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate OITW rows: %w", err)
	}

	return stockList, nil
}
