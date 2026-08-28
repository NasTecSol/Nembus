package extractors

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap/mappings"
	"github.com/NasTecSol/nembus-sap/schema"
)

type UOMExtractor struct {
	mssql *db.MSSQLClient
}

func NewUOMExtractor(mssql *db.MSSQLClient) *UOMExtractor {
	return &UOMExtractor{mssql: mssql}
}

// ExtractUOMs extracts all units of measure from OUOM and item master fields.
func (e *UOMExtractor) ExtractUOMs(ctx context.Context) ([]mappings.CanonicalUOM, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryUOM)
	if err != nil {
		// Fallback to item master UOMs if OUOM query fails
		var fallbackErr error
		rows, fallbackErr = e.mssql.DB.QueryContext(ctx, schema.QueryUOMFallback)
		if fallbackErr != nil {
			return []mappings.CanonicalUOM{}, nil
		}
	}
	defer rows.Close()

	var uoms []mappings.CanonicalUOM
	for rows.Next() {
		var u mappings.SAPUOM
		var name sql.NullString
		if err := rows.Scan(&u.UomEntry, &u.UomCode, &name, &u.Locked); err != nil {
			continue
		}
		u.UomName = name.String
		uoms = append(uoms, u.ToCanonical())
	}
	return uoms, nil
}

// ExtractUOMGroups extracts UoM groups and their hierarchical conversion definitions from OUGP & UGP1.
func (e *UOMExtractor) ExtractUOMGroups(ctx context.Context) ([]mappings.CanonicalUOMGroup, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryUOMGroups)
	if err != nil {
		// OUGP / UGP1 tables might not exist in older SAP installs
		return []mappings.CanonicalUOMGroup{}, nil
	}
	defer rows.Close()

	// Group rows by UgpEntry
	groupMap := make(map[int64]*mappings.CanonicalUOMGroup)
	groupOrder := make([]int64, 0)

	for rows.Next() {
		var detail mappings.SAPUOMGroupDetail
		var name, baseCode, altCode sql.NullString
		if err := rows.Scan(
			&detail.UgpEntry,
			&detail.UgpCode,
			&name,
			&detail.BaseUomEntry,
			&baseCode,
			&detail.AltUomEntry,
			&altCode,
			&detail.AltQty,
			&detail.BaseQty,
		); err != nil {
			continue // Skip malformed rows
		}

		detail.UgpName = name.String
		detail.BaseUomCode = baseCode.String
		detail.AltUomCode = altCode.String

		grp, exists := groupMap[detail.UgpEntry]
		if !exists {
			groupOrder = append(groupOrder, detail.UgpEntry)
			grp = &mappings.CanonicalUOMGroup{
				Code:        fmt.Sprintf("UGP-%d", detail.UgpEntry),
				Name:        detail.UgpName,
				BaseUOMCode: detail.BaseUomCode,
				IsActive:    true,
				Levels:      make([]mappings.CanonicalUOMGroupLevel, 0),
				Metadata: map[string]interface{}{
					"sap_ugp_entry":      detail.UgpEntry,
					"sap_ugp_code":       detail.UgpCode,
					"sap_base_uom_entry": detail.BaseUomEntry,
				},
			}
			groupMap[detail.UgpEntry] = grp
		}

		// Keep the one authoritative SAP direction used by inventory
		// normalization: BaseQty / AltQty (alternate UoM -> group base UoM).
		conversionFactor, err := mappings.SAPUOMConversionFactor(detail.BaseQty, detail.AltQty)
		if err != nil {
			return nil, fmt.Errorf("invalid UGP1 conversion for group %d and UoM %q: %w", detail.UgpEntry, detail.AltUomCode, err)
		}

		level := mappings.CanonicalUOMGroupLevel{
			LevelOrder:       len(grp.Levels) + 1,
			UOMCode:          detail.AltUomCode,
			Multiplier:       detail.BaseQty,
			ConversionFactor: conversionFactor,
		}
		grp.Levels = append(grp.Levels, level)
	}

	result := make([]mappings.CanonicalUOMGroup, 0, len(groupOrder))
	for _, entry := range groupOrder {
		if grp, ok := groupMap[entry]; ok {
			result = append(result, *grp)
		}
	}
	return result, nil
}
