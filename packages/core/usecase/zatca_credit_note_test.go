package usecase

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// mockZatcaDBTX implements repository.DBTX to verify queries and return mock rows.
type mockZatcaDBTX struct {
	insertedChainParams repository.InsertChainEntryParams
	latestICV           int64
	latestPIH           string
}

func (m *mockZatcaDBTX) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func (m *mockZatcaDBTX) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (m *mockZatcaDBTX) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return &mockZatcaRow{mock: m, sql: sql, args: args}
}

func (m *mockZatcaDBTX) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

type mockZatcaRow struct {
	mock *mockZatcaDBTX
	sql  string
	args []interface{}
}

func (r *mockZatcaRow) Scan(dest ...interface{}) error {
	s := r.sql
	switch {
	case strings.Contains(s, "GetNextICV") || strings.Contains(s, "COALESCE(MAX(icv), 0) + 1"):
		if len(dest) > 0 {
			if ptr, ok := dest[0].(*int32); ok {
				*ptr = int32(r.mock.latestICV + 1)
			} else if ptr, ok := dest[0].(*int64); ok {
				*ptr = r.mock.latestICV + 1
			}
		}
		return nil

	case strings.Contains(s, "GetLatestChainEntry") || strings.Contains(s, "ORDER BY icv DESC"):
		// simulate empty chain initially (triggering GenesisKPIH fallback)
		return pgx.ErrNoRows

	case strings.Contains(s, "InsertChainEntry") || strings.Contains(s, "INSERT INTO zatca_document_chain"):
		if len(r.args) >= 9 {
			r.mock.insertedChainParams = repository.InsertChainEntryParams{
				EntityType:     r.args[0].(string),
				EntityID:       r.args[1].(string),
				DeviceConfigID: r.args[2].(int32),
				OrganizationID: r.args[3].(int32),
				ZatcaUuid:      r.args[4].(uuid.UUID),
				Icv:            r.args[5].(int64),
				Pih:            r.args[6].(string),
				XmlHash:        r.args[7].(string),
			}
		}
		// Return inserted row
		if len(dest) >= 8 {
			if idPtr, ok := dest[0].(*int32); ok {
				*idPtr = 100
			}
		}
		return nil
	}
	return nil
}

// TestZatcaCreditNoteXMLGeneration verifies that BuildInvoiceXML generates
// compliant UBL 2.1 XML for Credit Notes (381) including cac:BillingReference.
func TestZatcaCreditNoteXMLGeneration(t *testing.T) {
	input := InvoiceXMLInput{
		InvoiceNumber:         "RET-2026-001",
		OriginalInvoiceNumber: "POS-TXN-9876",
		ZatcaUUID:             uuid.New().String(),
		IssueDate:             time.Date(2026, 10, 22, 14, 30, 0, 0, time.UTC),
		InvoiceType:           ZatcaTypeCreditNote,
		SubType:               ZatcaSubTypeB2CSimplified,
		CurrencyCode:          "SAR",
		ICV:                   1,
		PIH:                   GenesisKPIH,
		SellerName:            "Nembus Retail",
		SellerVATNumber:       "310123456700003",
		SellerStreet:          "King Fahd Rd",
		SellerCity:            "Riyadh",
		SellerPostalCode:      "11564",
		SellerCountry:         "SA",
		SellerCRNumber:        "1010123456",
		Subtotal:              "100.00",
		DiscountAmount:        "0.00",
		TaxAmount:             "15.00",
		TotalAmount:           "115.00",
		TaxCategoryID:         "S",
		TaxPercent:            "15.00",
		Lines: []InvoiceLineInput{
			{
				LineNumber:  1,
				ProductName: "Returned Wireless Mouse",
				Quantity:    "1.000",
				UnitCode:    "PCE",
				UnitPrice:   "100.00",
				TaxAmount:   "15.00",
				TaxPercent:  "15.00",
				TaxCategory: "S",
				LineTotal:   "100.00",
			},
		},
	}

	xmlBytes, err := BuildInvoiceXML(input)
	if err != nil {
		t.Fatalf("BuildInvoiceXML returned unexpected error: %v", err)
	}

	xmlStr := string(xmlBytes)

	// 1. Verify Credit Note code 381 and subtype
	if !strings.Contains(xmlStr, `<cbc:InvoiceTypeCode name="0200000">381</cbc:InvoiceTypeCode>`) {
		t.Errorf("Expected InvoiceTypeCode 381 with subtype 0200000, got XML:\n%s", xmlStr)
	}

	// 2. Verify cac:BillingReference presence
	if !strings.Contains(xmlStr, "<cac:BillingReference>") {
		t.Errorf("Expected BillingReference tag in Credit Note XML, got:\n%s", xmlStr)
	}

	// 3. Verify cac:InvoiceDocumentReference with original invoice ID
	if !strings.Contains(xmlStr, "<cbc:ID>POS-TXN-9876</cbc:ID>") {
		t.Errorf("Expected original invoice POS-TXN-9876 in BillingReference, got:\n%s", xmlStr)
	}

	// 4. Verify ICV and PIH
	if !strings.Contains(xmlStr, "<cbc:UUID>1</cbc:UUID>") {
		t.Errorf("Expected ICV 1 in AdditionalDocumentReference, got:\n%s", xmlStr)
	}
	if !strings.Contains(xmlStr, GenesisKPIH) {
		t.Errorf("Expected GenesisKPIH in PIH reference, got:\n%s", xmlStr)
	}
}

// TestZatcaSignCreditNote verifies full end-to-end signing of a Credit Note.
func TestZatcaSignCreditNote(t *testing.T) {
	key, err := GenerateECDSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateECDSAKeyPair failed: %v", err)
	}
	pemStr, err := EncodePrivateKeyPEM(key)
	if err != nil {
		t.Fatalf("EncodePrivateKeyPEM failed: %v", err)
	}

	mockDB := &mockZatcaDBTX{latestICV: 0}
	repo := repository.New(mockDB)

	cfg := &ZatcaConfig{
		Enabled:  true,
		Env:      "sandbox",
		OrgVATID: "310123456700003",
	}

	zatcaUC := NewZatcaUseCase(cfg)
	zatcaUC.SetRepository(repo)

	salesReturn := repository.SalesReturn{
		ID:                42,
		ReturnNumber:      "RET-2026-0042",
		ReturnDate:        pgtype.Timestamp{Time: time.Now(), Valid: true},
		Subtotal:          pgtype.Numeric{Valid: true},
		TaxAmount:         pgtype.Numeric{Valid: true},
		TotalRefundAmount: pgtype.Numeric{Valid: true},
	}
	_ = salesReturn.Subtotal.Scan("200.00")
	_ = salesReturn.TaxAmount.Scan("30.00")
	_ = salesReturn.TotalRefundAmount.Scan("230.00")

	returnLines := []repository.SalesReturnLine{
		{
			ID:           1,
			ReturnID:     42,
			ProductID:    101,
			LineNumber:   pgtype.Int4{Int32: 1, Valid: true},
			Quantity:     pgtype.Numeric{Valid: true},
			UnitPrice:    pgtype.Numeric{Valid: true},
			RefundAmount: pgtype.Numeric{Valid: true},
		},
	}
	_ = returnLines[0].Quantity.Scan("2.000")
	_ = returnLines[0].UnitPrice.Scan("100.00")
	_ = returnLines[0].RefundAmount.Scan("200.00")

	org := repository.Organization{
		ID:        1,
		Name:      "Nembus Supermarkets",
		Code:      "ORG-100",
		TaxID:     pgtype.Text{String: "310123456700003", Valid: true},
		LegalName: pgtype.Text{String: "Nembus Supermarkets Ltd", Valid: true},
	}

	device := repository.ZatcaDeviceConfig{
		ID:            5,
		PrivateKeyPem: pemStr,
	}

	signedDoc, chainEntry, err := zatcaUC.SignCreditNote(
		context.Background(),
		&salesReturn,
		returnLines,
		"POS-INV-1000",
		&org,
		&device,
	)
	if err != nil {
		t.Fatalf("SignCreditNote returned error: %v", err)
	}

	if signedDoc == nil {
		t.Fatal("Expected signedDoc to be non-nil")
	}

	// Verify cryptographic artifacts
	if signedDoc.QRCode == "" {
		t.Error("Expected non-empty TLV QR code")
	}
	if signedDoc.XMLHash == "" {
		t.Error("Expected non-empty XML hash")
	}
	if signedDoc.Signature == "" {
		t.Error("Expected non-empty ECDSA signature")
	}
	if signedDoc.ICV != 1 {
		t.Errorf("Expected ICV 1, got %d", signedDoc.ICV)
	}
	if signedDoc.PIH != GenesisKPIH {
		t.Errorf("Expected PIH %s, got %s", GenesisKPIH, signedDoc.PIH)
	}

	// Verify QR code is valid Base64
	qrBytes, err := base64.StdEncoding.DecodeString(signedDoc.QRCode)
	if err != nil {
		t.Fatalf("QRCode is not valid base64: %v", err)
	}
	if len(qrBytes) < 20 {
		t.Errorf("QRCode payload suspiciously small: %d bytes", len(qrBytes))
	}

	// Verify chain entry parameters passed to repository
	if mockDB.insertedChainParams.EntityType != "sales_return" {
		t.Errorf("Expected chain EntityType 'sales_return', got '%s'", mockDB.insertedChainParams.EntityType)
	}
	if mockDB.insertedChainParams.EntityID != "42" {
		t.Errorf("Expected chain EntityID '42', got '%s'", mockDB.insertedChainParams.EntityID)
	}
	if chainEntry == nil {
		t.Error("Expected non-nil chainEntry")
	}

	// Verify line-level tax calculation matching document tax (BR-CO-16)
	if !strings.Contains(string(signedDoc.RawXML), "<cbc:TaxAmount currencyID=\"SAR\">30.00</cbc:TaxAmount>") {
		t.Errorf("Expected line-level TaxAmount of 30.00 in Credit Note XML, got:\n%s", string(signedDoc.RawXML))
	}
	if strings.Contains(string(signedDoc.RawXML), "<cbc:TaxAmount currencyID=\"SAR\">0.00</cbc:TaxAmount>") {
		t.Error("Credit Note line still contains hardcoded 0.00 TaxAmount")
	}
}

// TestZatcaCreditNoteMissingBillingReferenceGuard verifies that SalesReturnUseCase
// guards against signing a Credit Note when OriginalTransactionID / origInvoiceNumber is absent.
func TestZatcaCreditNoteMissingBillingReferenceGuard(t *testing.T) {
	mockDB := &mockZatcaDBTX{latestICV: 0}
	repo := repository.New(mockDB)

	cfg := &ZatcaConfig{
		Enabled:  true,
		Env:      "sandbox",
		OrgVATID: "310123456700003",
	}

	zatcaUC := NewZatcaUseCase(cfg)
	zatcaUC.SetRepository(repo)

	salesReturnUC := NewSalesReturnUseCase()
	salesReturnUC.SetRepository(repo)
	salesReturnUC.SetZatcaUseCase(zatcaUC)

	// Return without OriginalTransactionID
	input := ProcessReturnInput{
		StoreID:           1,
		ReturnReason:      "Untracked return",
		Subtotal:          "50.00",
		TaxAmount:         "7.50",
		TotalRefundAmount: "57.50",
		RefundMethod:      "cash",
		Lines: []ProcessReturnLineInput{
			{
				ProductID:     101,
				Quantity:      "1.000",
				UnitPrice:     "50.00",
				RefundAmount:  "50.00",
				ReturnToStock: false,
			},
		},
	}

	resp := salesReturnUC.ProcessSalesReturn(context.Background(), input)
	if resp.StatusCode != 201 {
		t.Fatalf("ProcessSalesReturn failed: %s", resp.Message)
	}

	out, ok := resp.Data.(ProcessSalesReturnOutput)
	if !ok {
		t.Fatalf("Expected ProcessSalesReturnOutput, got %T", resp.Data)
	}

	// Credit Note must NOT be signed without a valid BillingReference to avoid ZATCA rejection
	if out.QRCode != "" {
		t.Error("Expected no ZATCA Credit Note QRCode generated when OriginalTransactionID is missing")
	}
	if len(mockDB.insertedChainParams.EntityID) > 0 {
		t.Error("Expected no ZATCA chain entry created for unreferenced credit note")
	}
}

