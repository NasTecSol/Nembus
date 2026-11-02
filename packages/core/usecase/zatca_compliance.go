package usecase

import (
	"fmt"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/google/uuid"
)

// ZatcaComplianceSample is an onboarding-only document. It never enters the
// business invoice chain, stock, payments, or POS transaction tables.
type ZatcaComplianceSample struct {
	Name     string
	Document *SignedDocument
}

// ZatcaCertificateExpiry reads the actual certificate validity, including
// secp256k1 certificates unsupported by crypto/x509.
func ZatcaCertificateExpiry(token string) (time.Time, error) {
	cert, err := parseZatcaCertificate(token)
	if err != nil {
		return time.Time{}, err
	}
	return cert.NotAfter, nil
}

func ValidateZatcaDeviceCertificate(token, privateKeyPEM string) error {
	cert, err := parseZatcaCertificate(token)
	if err != nil {
		return err
	}
	key, err := DecodePrivateKeyPEM(privateKeyPEM)
	if err != nil {
		return err
	}
	if cert.PublicKey.X.Cmp(key.X) != 0 || cert.PublicKey.Y.Cmp(key.Y) != 0 {
		return fmt.Errorf("issued certificate does not match the device private key")
	}
	if time.Now().Before(cert.NotBefore) || !time.Now().Before(cert.NotAfter) {
		return fmt.Errorf("issued certificate is not currently valid")
	}
	return nil
}

// BuildZatcaComplianceSamples covers the 1100 capability used by onboarding:
// standard and simplified invoices, credit notes, and debit notes.
func BuildZatcaComplianceSamples(org *repository.Organization, store *repository.Store, device *repository.ZatcaDeviceConfig, defaultVAT string) ([]ZatcaComplianceSample, error) {
	key, err := DecodePrivateKeyPEM(device.PrivateKeyPem)
	if err != nil {
		return nil, err
	}
	complianceDevice := *device
	complianceDevice.ProductionCsid.String = ""
	complianceDevice.ProductionCsid.Valid = false
	now := time.Now().UTC()
	pih := GenesisKPIH
	var samples []ZatcaComplianceSample
	for _, kind := range []struct {
		name    string
		subtype ZatcaInvoiceSubType
	}{
		{"standard", ZatcaSubTypeB2BStandard}, {"simplified", ZatcaSubTypeB2CSimplified},
	} {
		original := ""
		for _, docType := range []struct {
			name string
			code ZatcaInvoiceTypeCode
		}{
			{"invoice", ZatcaTypeStandard}, {"credit-note", ZatcaTypeCreditNote}, {"debit-note", ZatcaTypeDebitNote},
		} {
			number := "ONBOARD-" + uuid.NewString()
			input := InvoiceXMLInput{
				InvoiceNumber: number, ZatcaUUID: uuid.NewString(), IssueDate: now,
				InvoiceType: docType.code, SubType: kind.subtype, CurrencyCode: "SAR",
				ICV: int64(len(samples) + 1), PIH: pih, Subtotal: "100.00", DiscountAmount: "0.00",
				TaxAmount: "15.00", TotalAmount: "115.00", TaxCategoryID: "S", TaxPercent: "15.00",
				PaymentCode: "10",
				Lines:       []InvoiceLineInput{{LineNumber: 1, ProductName: "Onboarding compliance sample", Quantity: "1", UnitCode: "PCE", UnitPrice: "100.00", LineTotal: "100.00", TaxAmount: "15.00", TaxPercent: "15.00", TaxCategory: "S"}},
			}
			if err := applySeller(&input, org, store, defaultVAT); err != nil {
				return nil, err
			}
			if docType.code == ZatcaTypeStandard {
				original = number
			} else {
				input.OriginalInvoiceNumber = original
				input.NoteReason = "Onboarding sample adjustment"
			}
			if kind.subtype == ZatcaSubTypeB2BStandard {
				// Explicit sample buyer, used solely by the compliance API.
				input.BuyerName = "Onboarding Sample Buyer"
				input.BuyerVATNumber = "399999999900003"
				input.BuyerStreet = "Sample Street"
				input.BuyerBuildingNumber = "1234"
				input.BuyerDistrict = "Sample District"
				input.BuyerCity = "Riyadh"
				input.BuyerPostalCode = "12345"
				input.BuyerCountry = "SA"
				input.DeliveryDate = now.Format("2006-01-02")
			}
			signed, err := signZatcaDocument(input, key, &complianceDevice)
			if err != nil {
				return nil, fmt.Errorf("build %s %s: %w", kind.name, docType.name, err)
			}
			samples = append(samples, ZatcaComplianceSample{Name: kind.name + "-" + docType.name, Document: signed})
			pih = signed.XMLHash
		}
	}
	return samples, nil
}
