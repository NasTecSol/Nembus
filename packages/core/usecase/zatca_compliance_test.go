package usecase

import (
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/beevik/etree"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestZatcaComplianceSamplesCoverSixTypesWithoutReportingCertificate(t *testing.T) {
	key, _ := GenerateECDSAKeyPair()
	pem, _ := EncodePrivateKeyPEM(key)
	token := testZatcaCertificate(t, key)
	org := &repository.Organization{Name: "Seller", TaxID: pgtype.Text{String: "310123456700003", Valid: true}, Metadata: []byte(testZatcaSellerMetadata)}
	device := &repository.ZatcaDeviceConfig{PrivateKeyPem: pem, ComplianceCsid: pgtype.Text{String: token}, ProductionCsid: pgtype.Text{String: "must-not-be-used"}}
	samples, err := BuildZatcaComplianceSamples(org, &repository.Store{}, device, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 6 {
		t.Fatalf("expected six sample types, got %d", len(samples))
	}
	ids := map[string]bool{}
	previous := GenesisKPIH
	for i, sample := range samples {
		xml := etree.NewDocument()
		if err := xml.ReadFromBytes(sample.Document.RawXML); err != nil {
			t.Fatal(err)
		}
		root := xml.Root()
		code := root.FindElement("cbc:InvoiceTypeCode")
		expected := []string{"388", "381", "383"}[i%3]
		subtype := string(ZatcaSubTypeB2BStandard)
		if i >= 3 {
			subtype = string(ZatcaSubTypeB2CSimplified)
		}
		if code.Text() != expected || code.SelectAttrValue("name", "") != subtype {
			t.Fatal("wrong document type", sample.Name)
		}
		if ids[sample.Document.ZatcaUUID] {
			t.Fatal("duplicate UUID")
		}
		ids[sample.Document.ZatcaUUID] = true
		canonical, err := CanonicalizeXML(sample.Document.RawXML)
		if err != nil || HashSHA256Base64(canonical) != sample.Document.XMLHash {
			t.Fatal("sample hash mismatch")
		}
		if sample.Document.PIH != previous || sample.Document.ICV != int64(i+1) {
			t.Fatal("sample chain mismatch")
		}
		previous = sample.Document.XMLHash
		if i%3 != 0 {
			if root.FindElement("cac:BillingReference/cac:InvoiceDocumentReference/cbc:ID") == nil || root.FindElement("cac:PaymentMeans/cbc:InstructionNote") == nil {
				t.Fatal("note reference/reason missing")
			}
		}
		if i < 3 && (root.FindElement("cac:Delivery/cbc:ActualDeliveryDate") == nil || !strings.Contains(string(sample.Document.RawXML), "Onboarding Sample Buyer")) {
			t.Fatal("standard invoice buyer/delivery missing")
		}
	}
	if err := ValidateZatcaDeviceCertificate(token, pem); err != nil {
		t.Fatal(err)
	}
	otherKey, _ := GenerateECDSAKeyPair()
	otherPEM, _ := EncodePrivateKeyPEM(otherKey)
	if ValidateZatcaDeviceCertificate(token, otherPEM) == nil {
		t.Fatal("accepted mismatched certificate")
	}
}
