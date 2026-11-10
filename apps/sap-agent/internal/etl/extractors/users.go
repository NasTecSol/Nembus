package extractors

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/NasTecSol/nembus-sap/mappings"
	"github.com/NasTecSol/nembus-sap/schema"
	"github.com/NasTecSol/nembus-sap-agent/internal/db"
)

type UsersExtractor struct {
	mssql *db.MSSQLClient
}

func NewUsersExtractor(mssql *db.MSSQLClient) *UsersExtractor {
	return &UsersExtractor{mssql: mssql}
}

func (e *UsersExtractor) ExtractUsers(ctx context.Context, defaults mappings.CashierDefaults) ([]mappings.CanonicalUser, []mappings.CanonicalCashier, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, nil, fmt.Errorf("mssql database is not connected")
	}

	// 1. Extract Users (OUSR)
	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryUsers)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query OUSR: %w", err)
	}
	defer rows.Close()

	var users []mappings.CanonicalUser
	for rows.Next() {
		var u mappings.SAPUser
		var userCode, uName, email, locked sql.NullString
		if err := rows.Scan(&u.UserID, &userCode, &uName, &email, &locked); err != nil {
			return nil, nil, fmt.Errorf("failed to scan OUSR row: %w", err)
		}
		u.UserCode = userCode.String
		u.UName = uName.String
		u.EMail = email.String
		u.Locked = locked.String
		if u.Locked == "" {
			u.Locked = "N"
		}
		users = append(users, u.ToCanonical())
	}

	// 2. Extract Cashiers / Sales Persons (OSLP)
	cshRows, err := e.mssql.DB.QueryContext(ctx, schema.QueryCashiers)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query OSLP: %w", err)
	}
	defer cshRows.Close()

	var cashiers []mappings.CanonicalCashier
	for cshRows.Next() {
		var c mappings.SAPCashier
		var slpName, memo, active, email, telephone sql.NullString
		if err := cshRows.Scan(&c.SlpCode, &slpName, &memo, &active, &email, &telephone); err != nil {
			return nil, nil, fmt.Errorf("failed to scan OSLP row: %w", err)
		}
		c.SlpName = slpName.String
		c.Memo = memo.String
		c.Active = active.String
		if c.Active == "" {
			c.Active = "Y"
		}
		c.Email = email.String
		c.Telephone = telephone.String
		cashiers = append(cashiers, c.ToCanonical(defaults))
	}

	return users, cashiers, nil
}
