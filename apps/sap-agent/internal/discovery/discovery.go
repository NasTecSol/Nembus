package discovery

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/NasTecSol/nembus-sap/contracts"
	"github.com/NasTecSol/nembus-sap/schema"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
)

type DiscoveryEngine struct {
	mssql *db.MSSQLClient
}

func NewDiscoveryEngine(mssql *db.MSSQLClient) *DiscoveryEngine {
	return &DiscoveryEngine{mssql: mssql}
}

func (d *DiscoveryEngine) Discover(ctx context.Context) (*contracts.DiscoveryResult, error) {
	if d.mssql == nil || d.mssql.DB == nil {
		return nil, fmt.Errorf("mssql client is not connected")
	}

	result := &contracts.DiscoveryResult{
		ConnectedAt:  time.Now(),
		DatabaseName: d.mssql.Config.Database,
		TableCounts:  make(map[string]int64),
		Warnings:     make([]string, 0),
		IsCompatible: true,
	}

	// 1. Query SAP B1 Company & Version Information
	var compName, version, revNum, compAddr sql.NullString
	row := d.mssql.DB.QueryRowContext(ctx, schema.QuerySAPVersion)
	if err := row.Scan(&compName, &version, &revNum, &compAddr); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Could not read OADM table: %v. Database may not be a standard SAP B1 schema.", err))
		result.CompanyName = "Unknown SAP B1 Company"
		result.SAPVersion = "Unknown"
	} else {
		result.CompanyName = compName.String
		result.SAPVersion = fmt.Sprintf("%s (Rev %s)", version.String, revNum.String)
		result.Address = compAddr.String
		result.PatchLevel = revNum.String
	}

	// 2. Query Row Counts Across Core Tables
	rows, err := d.mssql.DB.QueryContext(ctx, schema.QueryTableCounts)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to query table counts: %v", err))
	} else {
		defer rows.Close()
		for rows.Next() {
			var tblName string
			var count int64
			if err := rows.Scan(&tblName, &count); err == nil {
				result.TableCounts[tblName] = count
			}
		}
	}

	// 3. Validation and sanity checks
	if result.TableCounts[schema.TableOITM] == 0 {
		result.Warnings = append(result.Warnings, "No item master records (OITM) found.")
	}
	if result.TableCounts[schema.TableOWHS] == 0 {
		result.Warnings = append(result.Warnings, "No warehouse records (OWHS) found.")
	}

	return result, nil
}
