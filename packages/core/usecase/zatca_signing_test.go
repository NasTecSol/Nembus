package usecase

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/beevik/etree"
	"github.com/jackc/pgx/v5/pgtype"
)

const testZatcaSellerMetadata = `{"zatca":{"enabled":true,"seller":{"commercial_registration_number":"1010123456","street":"Test Street","building_number":"1234","district":"Test District","city":"Riyadh","postal_code":"12345","country_code":"SA"}}}`

// Self-signed fixture only: never used for gateway authentication or production.
func testZatcaCertificate(t *testing.T, key *ecdsa.PrivateKey) string {
	t.Helper()
	issuer, err := asn1.Marshal(pkix.Name{CommonName: "Test CA", Country: []string{"SA"}}.ToRDNSequence())
	if err != nil {
		t.Fatal(err)
	}
	curve, _ := asn1.Marshal(oidSecp256k1)
	var certificate zatcaCertificateASN
	certificate.TBS.Version = 2
	certificate.TBS.Serial = big.NewInt(123456789)
	certificate.TBS.Signature.Algorithm = asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 2}
	certificate.TBS.Issuer = asn1.RawValue{FullBytes: issuer}
	certificate.TBS.Subject = asn1.RawValue{FullBytes: issuer}
	certificate.TBS.Validity.NotBefore = time.Now().UTC().Add(-24 * time.Hour)
	certificate.TBS.Validity.NotAfter = time.Now().UTC().Add(365 * 24 * time.Hour)
	certificate.TBS.PublicKey.Algorithm = pkix.AlgorithmIdentifier{Algorithm: asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}, Parameters: asn1.RawValue{FullBytes: curve}}
	pub := elliptic.Marshal(key.Curve, key.X, key.Y)
	certificate.TBS.PublicKey.Key = asn1.BitString{Bytes: pub, BitLength: len(pub) * 8}
	tbs, err := asn1.Marshal(certificate.TBS)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(tbs)
	signature, err := ecdsa.SignASN1(rand.Reader, key, hash[:])
	if err != nil {
		t.Fatal(err)
	}
	certificate.Algorithm = certificate.TBS.Signature
	certificate.Signature = asn1.BitString{Bytes: signature, BitLength: len(signature) * 8}
	der, err := asn1.Marshal(certificate)
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(der)
}

func testInvoiceInput() InvoiceXMLInput {
	return InvoiceXMLInput{InvoiceNumber: "TEST-1", ZatcaUUID: "00000000-0000-4000-8000-000000000001", IssueDate: time.Date(2026, 10, 26, 15, 30, 0, 0, time.FixedZone("PKT", 5*3600)),
		InvoiceType: ZatcaTypeStandard, SubType: ZatcaSubTypeB2CSimplified, CurrencyCode: "SAR", ICV: 1, PIH: GenesisKPIH,
		SellerName: "Test Seller", SellerVATNumber: "310123456700003", SellerStreet: "Test Street", SellerBuildingNumber: "1234", SellerDistrict: "Test District", SellerCity: "Riyadh", SellerPostalCode: "12345", SellerCountry: "SA", SellerCRNumber: "1010123456",
		Subtotal: "100.00", DiscountAmount: "0.00", TaxAmount: "15.00", TotalAmount: "115.00", TaxCategoryID: "S", TaxPercent: "15.00",
		Lines: []InvoiceLineInput{{LineNumber: 1, ProductName: "Test product", Quantity: "1.000", UnitCode: "PCE", UnitPrice: "100.00", LineTotal: "100.00", TaxAmount: "15.00", TaxPercent: "15.00", TaxCategory: "S"}}}
}

func TestZatcaStampedInvoiceHashSignatureAndQR(t *testing.T) {
	key, err := GenerateECDSAKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	token := testZatcaCertificate(t, key)
	input := testInvoiceInput()
	signed, err := signZatcaDocument(input, key, &repository.ZatcaDeviceConfig{ComplianceCsid: pgtype.Text{String: token, Valid: true}})
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := CanonicalizeXML(signed.RawXML)
	if err != nil {
		t.Fatal(err)
	}
	if HashSHA256Base64(canonical) != signed.XMLHash {
		t.Fatal("submitted XML hash differs from the API hash")
	}
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(signed.RawXML); err != nil {
		t.Fatal(err)
	}
	root := doc.Root()
	if root.ChildElements()[0].Tag != "UBLExtensions" {
		t.Fatal("signature extension must precede invoice fields")
	}
	var invoiceDigest, propertiesDigest, signatureValue, embeddedQR string
	var props *etree.Element
	var walk func(*etree.Element)
	walk = func(e *etree.Element) {
		if e.Tag == "Reference" {
			for _, c := range e.ChildElements() {
				if c.Tag == "DigestValue" {
					if e.SelectAttrValue("Id", "") == "invoiceSignedData" {
						invoiceDigest = c.Text()
					} else if e.SelectAttrValue("URI", "") == "#xadesSignedProperties" {
						propertiesDigest = c.Text()
					}
				}
			}
		}
		if e.Tag == "SignedProperties" {
			props = e
		}
		if e.Tag == "SignatureValue" {
			signatureValue = e.Text()
		}
		if e.Tag == "AdditionalDocumentReference" && e.FindElement("cbc:ID") != nil && e.FindElement("cbc:ID").Text() == "QR" {
			embeddedQR = e.FindElement("cac:Attachment/cbc:EmbeddedDocumentBinaryObject").Text()
		}
		for _, c := range e.ChildElements() {
			walk(c)
		}
	}
	walk(root)
	if invoiceDigest != signed.XMLHash || embeddedQR != signed.QRCode || signatureValue != signed.Signature {
		t.Fatal("XML, QR and API values disagree")
	}
	if props == nil {
		t.Fatal("SignedProperties is missing")
	}
	propHash, err := signedPropertiesHash(props)
	if err != nil || propHash != propertiesDigest {
		t.Fatal("properties hash does not match embedded properties")
	}
	qrBytes, err := base64.StdEncoding.DecodeString(embeddedQR)
	if err != nil {
		t.Fatal(err)
	}
	tags := map[byte][]byte{}
	for pos := 0; pos < len(qrBytes); {
		if pos+2 > len(qrBytes) {
			t.Fatal("truncated TLV")
		}
		tag, length := qrBytes[pos], int(qrBytes[pos+1])
		pos += 2
		if pos+length > len(qrBytes) {
			t.Fatal("invalid TLV length")
		}
		tags[tag] = qrBytes[pos : pos+length]
		pos += length
	}
	if len(tags) != 9 || string(tags[3]) != "2026-10-26T10:30:00Z" || string(tags[6]) != signed.XMLHash || string(tags[7]) != signed.Signature {
		t.Fatal("QR tags differ from the invoice")
	}
	issueDate := root.FindElement("cbc:IssueDate")
	issueTime := root.FindElement("cbc:IssueTime")
	if issueDate == nil || issueTime == nil {
		t.Fatal("invoice issue date/time is missing")
	}
	if issueTime.Text() != "10:30:00Z" || issueDate.Text()+"T"+issueTime.Text() != string(tags[3]) {
		t.Fatalf("XML issue timestamp %sT%s must match QR tag 3 %s", issueDate.Text(), issueTime.Text(), tags[3])
	}
	cert, err := parseZatcaCertificate(token)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(tags[8], cert.SPKI) || !bytes.Equal(tags[9], cert.Signature) {
		t.Fatal("QR certificate public key or CA signature is wrong")
	}
	hash, _ := base64.StdEncoding.DecodeString(signed.XMLHash)
	sig, _ := base64.StdEncoding.DecodeString(signatureValue)
	if !ecdsa.VerifyASN1(cert.PublicKey, hash, sig) {
		t.Fatal("invoice signature does not verify against certificate")
	}
	changed := bytes.Replace(signed.RawXML, []byte("Test product"), []byte("Changed product"), 1)
	changedCanonical, _ := CanonicalizeXML(changed)
	if HashSHA256Base64(changedCanonical) == signed.XMLHash {
		t.Fatal("tampering was not detected")
	}
	if strings.Contains(string(signed.RawXML), "<cbc:IdentificationCode></cbc:IdentificationCode>") {
		t.Fatal("anonymous buyer emitted an empty country code")
	}
	if !strings.Contains(string(signed.RawXML), `<cbc:RoundingAmount currencyID="SAR">115.00</cbc:RoundingAmount>`) {
		t.Fatal("line VAT-inclusive amount missing")
	}
}

func TestZatcaCanonicalizationGoldenAndFilters(t *testing.T) {
	input := []byte(`<?xml version="1.0"?><Invoice xmlns="urn:test" xmlns:cac="` + zatcaCAC + `" xmlns:cbc="` + zatcaCBC + `" xmlns:ext="` + zatcaEXT + `" z="2" a="1"><ext:UBLExtensions/><cbc:ID>1 &amp; 2</cbc:ID><cac:Signature/><cac:AdditionalDocumentReference><cbc:ID>QR</cbc:ID></cac:AdditionalDocumentReference><!--comment--><Empty/></Invoice>`)
	want := `<Invoice xmlns="urn:test" xmlns:cac="` + zatcaCAC + `" xmlns:cbc="` + zatcaCBC + `" xmlns:ext="` + zatcaEXT + `" a="1" z="2"><cbc:ID>1 &amp; 2</cbc:ID><Empty></Empty></Invoice>`
	got, err := CanonicalizeXML(input)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("canonical mismatch\ngot %s\nwant %s", got, want)
	}
	if _, err := CanonicalizeXML([]byte(`<Invoice><Broken>`)); err == nil {
		t.Fatal("malformed XML must fail")
	}
}

func TestZatcaAmountValidationRejectsInconsistentSale(t *testing.T) {
	input := testInvoiceInput()
	input.Lines[0].Quantity = "2.000"
	if _, err := BuildInvoiceXML(input); err == nil || !strings.Contains(err.Error(), "quantity") {
		t.Fatalf("expected quantity/price validation, got %v", err)
	}
	input.Lines[0].UnitPrice = "50.00"
	if _, err := BuildInvoiceXML(input); err != nil {
		t.Fatal(err)
	}
}

func TestZatcaSellerProfileRequiresRealData(t *testing.T) {
	org := repository.Organization{Name: "Test", TaxID: pgtype.Text{String: "310123456700003", Valid: true}}
	store := repository.Store{}
	if err := applySeller(&InvoiceXMLInput{}, &org, &store, ""); err == nil {
		t.Fatal("missing seller data must fail")
	}
	org.Metadata = json.RawMessage(testZatcaSellerMetadata)
	input := InvoiceXMLInput{}
	if err := applySeller(&input, &org, &store, ""); err != nil {
		t.Fatal(err)
	}
	if input.SellerBuildingNumber != "1234" || input.SellerCRNumber != "1010123456" {
		t.Fatal("seller profile was not applied")
	}
}

func TestZatcaCheckoutRejectsBadAmountsBeforeWritingSale(t *testing.T) {
	db := &workflowMockDBTX{}
	repo := repository.New(db)
	zatca := NewZatcaUseCase(&ZatcaConfig{Enabled: true})
	zatca.SetRepository(repo)
	pos := NewPosUseCase()
	pos.SetRepository(repo)
	pos.SetZatcaUseCase(zatca)
	var quantity, price, tax, total pgtype.Numeric
	_ = quantity.Scan("2.000")
	_ = price.Scan("100.00")
	_ = tax.Scan("15.00")
	_ = total.Scan("115.00")
	response := pos.CreateTransaction(context.Background(), &PosCreateTransactionInput{StoreID: 1, CashierID: 1, CashierSessionID: 1, PosTerminalID: 1, Subtotal: price, TaxAmount: tax, TotalAmount: total,
		Lines: []PosCreateTransactionLineInput{{ProductID: 1, Quantity: quantity, UnitPrice: price, Subtotal: price, TaxAmount: tax, LineTotal: total}}})
	if response.StatusCode != 400 || db.createdTransactions != 0 {
		t.Fatalf("invalid sale must fail before inserts: %+v, writes=%d", response, db.createdTransactions)
	}
}

func TestZatcaSigningRejectsWrongCertificate(t *testing.T) {
	key, _ := GenerateECDSAKeyPair()
	other, _ := GenerateECDSAKeyPair()
	_, err := signZatcaDocument(testInvoiceInput(), key, &repository.ZatcaDeviceConfig{ComplianceCsid: pgtype.Text{String: testZatcaCertificate(t, other), Valid: true}})
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("expected key mismatch, got %v", err)
	}
}

func TestZatcaQRUsesSingleByteLengthForLongUTF8Name(t *testing.T) {
	key, _ := GenerateECDSAKeyPair()
	input := testInvoiceInput()
	input.SellerName = strings.Repeat("س", 70)
	signed, err := signZatcaDocument(input, key, &repository.ZatcaDeviceConfig{ComplianceCsid: pgtype.Text{String: testZatcaCertificate(t, key), Valid: true}})
	if err != nil {
		t.Fatal(err)
	}
	data, _ := base64.StdEncoding.DecodeString(signed.QRCode)
	if data[1] != 140 || string(data[2:142]) != input.SellerName || data[142] != 2 {
		t.Fatal("QR length was encoded as ASN.1 BER rather than a single byte")
	}
}

func TestZatcaCertificateIssuerDomainComponents(t *testing.T) {
	rdns := pkix.RDNSequence{
		{{Type: asn1.ObjectIdentifier{0, 9, 2342, 19200300, 100, 1, 25}, Value: "local"}},
		{{Type: asn1.ObjectIdentifier{0, 9, 2342, 19200300, 100, 1, 25}, Value: "gov"}},
		{{Type: asn1.ObjectIdentifier{2, 5, 4, 3}, Value: "Test, CA"}},
	}
	if got := zatcaIssuerName(rdns); got != `CN=Test\, CA, DC=gov, DC=local` {
		t.Fatalf("wrong issuer name: %s", got)
	}
}
