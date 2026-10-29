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

	// Primary query includes OCRD.CreditLine (the SAP B1 credit-limit column)
	// and OCRD.GroupNum (payment terms). Some SAP B1 versions/localizations
	// lack CreditLine — fall back to a query without it instead of failing the
	// whole partners domain.
	partners, pErr := e.scanPartners(ctx, schema.QueryBusinessPartners, true)
	if pErr != nil {
		fallback, fbErr := e.scanPartners(ctx, schema.QueryBusinessPartnersFallback, false)
		if fbErr != nil {
			return nil, fmt.Errorf("failed to query OCRD: %w (fallback also failed: %w)", pErr, fbErr)
		}
		partners = fallback
	}

	return partners, nil
}

func (e *PartnersExtractor) scanPartners(ctx context.Context, query string, withCreditLine bool) ([]mappings.CanonicalPartner, error) {
	rows, err := e.mssql.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var partners []mappings.CanonicalPartner
	for rows.Next() {
		var bp mappings.SAPBusinessPartner
		var taxNum, phone, email, currency sql.NullString
		if withCreditLine {
			// Scan order must match QueryBusinessPartners (11 columns, including
			// A19 CreditLimit and A23 GroupNum).
			if err := rows.Scan(
				&bp.CardCode,
				&bp.CardName,
				&bp.CardType,
				&taxNum,
				&phone,
				&email,
				&currency,
				&bp.ValidFor,
				&bp.Balance,
				&bp.CreditLimit,
				&bp.GroupNum,
			); err != nil {
				return nil, fmt.Errorf("failed to scan OCRD row: %w", err)
			}
		} else {
			// Scan order must match QueryBusinessPartnersFallback (10 columns,
			// credit limit omitted).
			if err := rows.Scan(
				&bp.CardCode,
				&bp.CardName,
				&bp.CardType,
				&taxNum,
				&phone,
				&email,
				&currency,
				&bp.ValidFor,
				&bp.Balance,
				&bp.GroupNum,
			); err != nil {
				return nil, fmt.Errorf("failed to scan OCRD row: %w", err)
			}
		}
		bp.LicTradNum = taxNum.String
		bp.Phone1 = phone.String
		bp.EMail = email.String
		bp.Currency = currency.String

		partners = append(partners, bp.ToCanonical())
	}

	return partners, rows.Err()
}

// ExtractPaymentTerms extracts the payment terms master from OCTG — A23.
// Must be ingested before partners so OCRD.GroupNum can resolve to
// business_partners.payment_terms_id.
func (e *PartnersExtractor) ExtractPaymentTerms(ctx context.Context) ([]mappings.CanonicalPaymentTerm, error) {
	if e.mssql == nil || e.mssql.DB == nil {
		return nil, fmt.Errorf("mssql database is not connected")
	}

	rows, err := e.mssql.DB.QueryContext(ctx, schema.QueryPaymentTerms)
	if err != nil {
		// OCTG always exists in SAP B1, but keep the tolerant behaviour
		return []mappings.CanonicalPaymentTerm{}, nil
	}
	defer rows.Close()

	var paymentTerms []mappings.CanonicalPaymentTerm
	for rows.Next() {
		var pt mappings.SAPPaymentTerm
		var pymGroup sql.NullString
		if err := rows.Scan(
			&pt.GroupNum,
			&pymGroup,
			&pt.InstMonths,
			&pt.InstDays,
			&pt.ExtraMonths,
			&pt.ExtraDays,
			&pt.DiscPrcnt,
			&pt.DiscDays,
		); err != nil {
			return nil, fmt.Errorf("failed to scan OCTG row: %w", err)
		}
		pt.PymGroup = pymGroup.String
		paymentTerms = append(paymentTerms, pt.ToCanonical())
	}

	return paymentTerms, nil
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
