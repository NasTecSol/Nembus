package extractors

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/NasTecSol/nembus-sap-agent/internal/db"
	"github.com/NasTecSol/nembus-sap/mappings"
	"github.com/NasTecSol/nembus-sap/schema"
)

type PartnersExtractor struct {
	mssql *db.MSSQLClient
}

func NewPartnersExtractor(mssql *db.MSSQLClient) *PartnersExtractor {
	return &PartnersExtractor{mssql: mssql}
}

func (e *PartnersExtractor) ExtractPartners(ctx context.Context) ([]mappings.CanonicalPartner, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryBusinessPartners)
	if err != nil {
		return nil, fmt.Errorf("failed to query OCRD: %w", err)
	}
	defer rows.Close()

	var partners []mappings.CanonicalPartner
	for rows.Next() {
		var bp mappings.SAPBusinessPartner
		var taxNum, phone, email, currency sql.NullString
		if err := rows.Scan(
			&bp.CardCode,
			&bp.CardName,
			&bp.CardType,
			&taxNum,
			&phone,
			&email,
			&currency,
			&bp.ValidFor,
			&bp.FrozenFor,
			&bp.Balance,
		); err != nil {
			return nil, fmt.Errorf("failed to scan OCRD row: %w", err)
		}
		bp.LicTradNum = taxNum.String
		bp.Phone1 = phone.String
		bp.EMail = email.String
		bp.Currency = currency.String

		partners = append(partners, bp.ToCanonical())
	}

	return partners, nil
}

// ExtractBPAddresses extracts billing and shipping addresses for all business partners from CRD1.
func (e *PartnersExtractor) ExtractBPAddresses(ctx context.Context) ([]mappings.CanonicalBPAddress, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryBPAddresses)
	if err != nil {
		return []mappings.CanonicalBPAddress{}, nil // CRD1 may be empty
	}
	defer rows.Close()

	var addresses []mappings.CanonicalBPAddress
	for rows.Next() {
		var addr mappings.SAPBPAddress
		var address, street, city, country, zipCode, state, phone1, phone2 sql.NullString
		if err := rows.Scan(
			&addr.CardCode,
			&addr.AdresType,
			&address,
			&street,
			&city,
			&country,
			&zipCode,
			&state,
			&phone1,
			&phone2,
		); err != nil {
			continue // Skip malformed rows
		}
		addr.Address = address.String
		addr.Street = street.String
		addr.City = city.String
		addr.Country = country.String
		addr.ZipCode = zipCode.String
		addr.State = state.String
		addr.Phone1 = phone1.String
		addr.Phone2 = phone2.String
		addresses = append(addresses, addr.ToCanonical())
	}
	return addresses, nil
}
