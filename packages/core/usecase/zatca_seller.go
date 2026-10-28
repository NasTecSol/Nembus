package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

// Stored under metadata.zatca.seller; store fields override organization fields.
type ZatcaSellerProfile struct {
	CommercialRegistrationNumber string `json:"commercial_registration_number"`
	Street                       string `json:"street"`
	BuildingNumber               string `json:"building_number"`
	District                     string `json:"district"`
	City                         string `json:"city"`
	PostalCode                   string `json:"postal_code"`
	CountryCode                  string `json:"country_code"`
}

func applySeller(input *InvoiceXMLInput, org *repository.Organization, store *repository.Store, defaultVAT string) error {
	profile := ZatcaSellerProfile{CountryCode: "SA"}
	for _, metadata := range []json.RawMessage{org.Metadata, store.Metadata} {
		if len(metadata) == 0 {
			continue
		}
		var settings struct {
			Zatca struct {
				Seller ZatcaSellerProfile `json:"seller"`
			} `json:"zatca"`
		}
		if err := json.Unmarshal(metadata, &settings); err != nil {
			return fmt.Errorf("invalid ZATCA seller metadata: %w", err)
		}
		p := settings.Zatca.Seller
		for _, item := range []struct {
			target *string
			value  string
		}{
			{&profile.CommercialRegistrationNumber, p.CommercialRegistrationNumber}, {&profile.Street, p.Street}, {&profile.BuildingNumber, p.BuildingNumber},
			{&profile.District, p.District}, {&profile.City, p.City}, {&profile.PostalCode, p.PostalCode}, {&profile.CountryCode, p.CountryCode},
		} {
			if strings.TrimSpace(item.value) != "" {
				*item.target = strings.TrimSpace(item.value)
			}
		}
	}
	var missing []string
	for _, field := range []struct{ name, value string }{
		{"commercial_registration_number", profile.CommercialRegistrationNumber}, {"street", profile.Street}, {"building_number", profile.BuildingNumber},
		{"district", profile.District}, {"city", profile.City}, {"postal_code", profile.PostalCode},
	} {
		if field.value == "" {
			missing = append(missing, field.name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("configure organization/store metadata.zatca.seller: missing %s", strings.Join(missing, ", "))
	}
	for _, check := range []struct{ value, pattern, name string }{
		{profile.CommercialRegistrationNumber, `^[0-9]{10}$`, "commercial_registration_number (10 digits)"},
		{profile.BuildingNumber, `^[0-9]{4}$`, "building_number (4 digits)"}, {profile.PostalCode, `^[0-9]{5}$`, "postal_code (5 digits)"},
	} {
		if !regexp.MustCompile(check.pattern).MatchString(check.value) {
			return fmt.Errorf("invalid ZATCA seller %s", check.name)
		}
	}
	if profile.CountryCode != "SA" {
		return fmt.Errorf("ZATCA seller country_code must be SA")
	}
	input.SellerName = org.LegalName.String
	if input.SellerName == "" {
		input.SellerName = org.Name
	}
	input.SellerVATNumber = org.TaxID.String
	if input.SellerVATNumber == "" {
		input.SellerVATNumber = defaultVAT
	}
	if !regexp.MustCompile(`^3[0-9]{13}3$`).MatchString(input.SellerVATNumber) {
		return fmt.Errorf("ZATCA seller VAT number must be 15 digits starting and ending with 3")
	}
	input.SellerStreet = profile.Street
	input.SellerBuildingNumber = profile.BuildingNumber
	input.SellerDistrict = profile.District
	input.SellerCity = profile.City
	input.SellerPostalCode = profile.PostalCode
	input.SellerCountry = profile.CountryCode
	input.SellerCRNumber = profile.CommercialRegistrationNumber
	return nil
}

func (uc *ZatcaUseCase) validatePOSCheckout(ctx context.Context, input *PosCreateTransactionInput, store *repository.Store) error {
	if err := validatePOSAmounts(input); err != nil {
		return err
	}
	org, err := uc.repo.GetOrganization(ctx, store.OrganizationID)
	if err != nil {
		return fmt.Errorf("ZATCA organization not found: %w", err)
	}
	if err := applySeller(&InvoiceXMLInput{}, &org, store, uc.cfg.OrgVATID); err != nil {
		return err
	}
	device, err := uc.SelectZatcaDevice(ctx, org.ID, pgtype.Int4{Int32: input.PosTerminalID, Valid: true}, input.Metadata)
	if err != nil {
		return err
	}

	key, err := DecodePrivateKeyPEM(device.PrivateKeyPem)
	if err != nil {
		return err
	}
	token := device.ProductionCsid.String
	if token == "" {
		token = device.ComplianceCsid.String
	}
	cert, err := parseZatcaCertificate(token)
	if err != nil {
		return err
	}
	if key.X.Cmp(cert.PublicKey.X) != 0 || key.Y.Cmp(cert.PublicKey.Y) != 0 {
		return fmt.Errorf("ZATCA certificate and private key do not match")
	}
	if time.Now().Before(cert.NotBefore) || !time.Now().Before(cert.NotAfter) {
		return fmt.Errorf("ZATCA signing certificate is not currently valid")
	}
	return nil
}
