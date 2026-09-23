package extractors

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/NasTecSol/nembus-sap/mappings"
	"github.com/NasTecSol/nembus-sap/schema"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
)

type StoresExtractor struct {
	mssql *db.MSSQLClient
}

func NewStoresExtractor(mssql *db.MSSQLClient) *StoresExtractor {
	return &StoresExtractor{mssql: mssql}
}

func (e *StoresExtractor) ExtractStores(ctx context.Context) ([]mappings.CanonicalStore, []mappings.CanonicalStorageLocation, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, nil, fmt.Errorf("mssql database is not connected")
	}

	// 1. Extract Warehouses (OWHS)
	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryStores)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query OWHS: %w", err)
	}
	defer rows.Close()

	var stores []mappings.CanonicalStore
	for rows.Next() {
		var s mappings.SAPStore
		var whsName, locked, street, city, country, zipCode sql.NullString
		if err := rows.Scan(&s.WhsCode, &whsName, &locked, &street, &city, &country, &zipCode); err != nil {
			return nil, nil, fmt.Errorf("failed to scan OWHS row: %w", err)
		}
		s.WhsName = whsName.String
		s.Locked = locked.String
		if s.Locked == "" {
			s.Locked = "N"
		}
		s.Street = street.String
		s.City = city.String
		s.Country = country.String
		s.ZipCode = zipCode.String

		stores = append(stores, s.ToCanonical())
	}

	// 2. Extract Bins (OBIN) if table exists
	var locations []mappings.CanonicalStorageLocation
	binRows, err := e.mssql.DB.QueryContext(ctx, schema.QueryStorageLocations)
	if err == nil {
		defer binRows.Close()
		for binRows.Next() {
			var b mappings.SAPStorageLocation
			var descr sql.NullString
			if err := binRows.Scan(&b.AbsEntry, &b.BinCode, &b.WhsCode, &descr, &b.Disabled); err == nil {
				b.Descr = descr.String
				locations = append(locations, b.ToCanonical())
			}
		}
	}

	return stores, locations, nil
}
