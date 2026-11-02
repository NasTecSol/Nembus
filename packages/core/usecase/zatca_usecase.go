package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ZatcaUseCase struct {
	repo *repository.Queries
	cfg  *ZatcaConfig
}

func NewZatcaUseCase(cfg *ZatcaConfig) *ZatcaUseCase {
	return &ZatcaUseCase{
		cfg: cfg,
	}
}

func (uc *ZatcaUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *ZatcaUseCase) GetConfig() *ZatcaConfig {
	return uc.cfg
}

func numericToString(n pgtype.Numeric) string {
	if !n.Valid {
		return "0.00"
	}
	val, err := n.Value()
	if err != nil || val == nil {
		return "0.00"
	}
	return fmt.Sprintf("%v", val)
}

// GetNextICVAndPIH returns the next Invoice Counter Value (ICV) and the Previous Invoice Hash (PIH) for a given device.
func (uc *ZatcaUseCase) GetNextICVAndPIH(ctx context.Context, deviceConfigID int32) (int64, string, error) {
	if uc.repo == nil {
		return 0, "", fmt.Errorf("repository not initialized")
	}

	nextICV, err := uc.repo.GetNextICV(ctx, deviceConfigID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get next ICV: %w", err)
	}

	latestEntry, err := uc.repo.GetLatestChainEntry(ctx, deviceConfigID)
	if errors.Is(err, pgx.ErrNoRows) {
		// If no entry exists yet, this is the very first document in the chain
		return int64(nextICV), GenesisKPIH, nil
	}
	if err != nil {
		return 0, "", fmt.Errorf("failed to read previous invoice hash: %w", err)
	}

	return int64(nextICV), latestEntry.XmlHash, nil
}

// SignInvoice processes a B2B Standard Invoice, generates UBL 2.1 XML, signs it, and records the chain entry.
func (uc *ZatcaUseCase) SignInvoice(ctx context.Context, invoice *repository.Invoice, lines []repository.InvoiceLine, org *repository.Organization, store *repository.Store, device *repository.ZatcaDeviceConfig) (*SignedDocument, *repository.ZatcaDocumentChain, error) {
	if err := RequireZatcaSigningOwner(device, uc.cfg); err != nil {
		return nil, nil, err
	}
	enabled, err := OrganizationZatcaEnabled(org)
	if err != nil {
		return nil, nil, err
	}
	if !enabled {
		return nil, nil, ErrZatcaDisabled
	}

	privKey, err := DecodePrivateKeyPEM(device.PrivateKeyPem)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode device private key: %w", err)
	}

	icv, pih, err := uc.GetNextICVAndPIH(ctx, device.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get ICV/PIH: %w", err)
	}

	zatcaUUID := uuid.New().String()

	// Map invoice lines
	xmlLines := make([]InvoiceLineInput, 0, len(lines))
	for idx, line := range lines {
		lineNum := idx + 1
		if line.LineNumber > 0 {
			lineNum = int(line.LineNumber)
		}
		xmlLines = append(xmlLines, InvoiceLineInput{
			LineNumber:  lineNum,
			ProductName: line.Description,
			Quantity:    numericToString(line.Quantity),
			UnitCode:    "PCE",
			UnitPrice:   numericToString(line.UnitPrice),
			DiscountAmt: numericToString(line.DiscountAmount),
			TaxAmount:   numericToString(line.TaxAmount),
			TaxPercent:  "15.00",
			TaxCategory: "S",
			LineTotal:   numericToString(line.LineTotal),
		})
	}

	sellerName := org.LegalName.String
	if sellerName == "" {
		sellerName = org.Name
	}
	vatID := org.TaxID.String
	if vatID == "" {
		vatID = uc.cfg.OrgVATID
	}

	xmlInput := InvoiceXMLInput{
		InvoiceNumber:   invoice.InvoiceNumber,
		ZatcaUUID:       zatcaUUID,
		IssueDate:       invoice.InvoiceDate.Time,
		InvoiceType:     ZatcaTypeStandard,
		SubType:         ZatcaSubTypeB2BStandard,
		CurrencyCode:    invoice.CurrencyCode.String,
		ICV:             icv,
		PIH:             pih,
		SellerName:      sellerName,
		SellerVATNumber: vatID,
		BuyerName:       invoice.CustomerName,
		BuyerVATNumber:  invoice.CustomerTaxID.String,
		BuyerCountry:    "SA",
		Subtotal:        numericToString(invoice.Subtotal),
		DiscountAmount:  numericToString(invoice.DiscountAmount),
		TaxAmount:       numericToString(invoice.TaxAmount),
		TotalAmount:     numericToString(invoice.TotalAmount),
		TaxCategoryID:   "S",
		TaxPercent:      "15.00",
		Lines:           xmlLines,
	}

	if err := applySeller(&xmlInput, org, store, uc.cfg.OrgVATID); err != nil {
		return nil, nil, err
	}
	signedDoc, err := signZatcaDocument(xmlInput, privKey, device)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to stamp ZATCA document: %w", err)
	}
	xmlHash := signedDoc.XMLHash
	qrCodeBase64 := signedDoc.QRCode
	rawXML := signedDoc.RawXML
	// Insert record into zatca_document_chain
	parsedUUID, _ := uuid.Parse(zatcaUUID)
	chainParams := repository.InsertChainEntryParams{
		EntityType:     "invoice",
		EntityID:       invoice.ID.String(),
		DeviceConfigID: device.ID,
		OrganizationID: org.ID,
		ZatcaUuid:      parsedUUID,
		Icv:            icv,
		Pih:            pih,
		XmlHash:        xmlHash,
		ZatcaStatus:    repository.NullZatcaDocStatus{ZatcaDocStatus: repository.ZatcaDocStatusPending, Valid: true},
		QrCodeBase64:   pgtype.Text{String: qrCodeBase64, Valid: true},
		SignedXml:      pgtype.Text{String: string(rawXML), Valid: true},
	}

	chainEntry, err := uc.repo.InsertChainEntry(ctx, chainParams)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to insert document chain entry: %w", err)
	}

	return signedDoc, &chainEntry, nil
}

// SignPosTransaction processes a B2C Simplified Tax Invoice for a POS transaction, signs it, and generates the TLV QR code.
func (uc *ZatcaUseCase) SignPosTransaction(ctx context.Context, txn *repository.PosTransaction, lines []repository.PosTransactionLine, org *repository.Organization, device *repository.ZatcaDeviceConfig) (*SignedDocument, *repository.ZatcaDocumentChain, error) {
	if err := RequireZatcaSigningOwner(device, uc.cfg); err != nil {
		return nil, nil, err
	}
	enabled, err := OrganizationZatcaEnabled(org)
	if err != nil {
		return nil, nil, err
	}
	if !enabled {
		return nil, nil, ErrZatcaDisabled
	}

	privKey, err := DecodePrivateKeyPEM(device.PrivateKeyPem)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode device private key: %w", err)
	}

	icv, pih, err := uc.GetNextICVAndPIH(ctx, device.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get ICV/PIH: %w", err)
	}

	zatcaUUID := uuid.New().String()

	xmlLines := make([]InvoiceLineInput, 0, len(lines))
	for idx, line := range lines {
		lineNum := idx + 1
		if line.LineNumber.Valid {
			lineNum = int(line.LineNumber.Int32)
		}
		xmlLines = append(xmlLines, InvoiceLineInput{
			LineNumber:  lineNum,
			ProductName: fmt.Sprintf("Product #%d", line.ProductID),
			Quantity:    numericToString(line.Quantity),
			UnitCode:    "PCE",
			UnitPrice:   numericToString(line.UnitPrice),
			DiscountAmt: numericToString(line.DiscountAmount),
			TaxAmount:   numericToString(line.TaxAmount),
			TaxPercent:  "15.00",
			TaxCategory: "S",
			LineTotal:   numericToString(line.Subtotal),
		})
	}

	sellerName := org.LegalName.String
	if sellerName == "" {
		sellerName = org.Name
	}
	vatID := org.TaxID.String
	if vatID == "" {
		vatID = uc.cfg.OrgVATID
	}

	txnDate := txn.TransactionDate.Time
	if txnDate.IsZero() {
		txnDate = time.Now()
	}

	xmlInput := InvoiceXMLInput{
		InvoiceNumber:   txn.TransactionNumber,
		ZatcaUUID:       zatcaUUID,
		IssueDate:       txnDate,
		InvoiceType:     ZatcaTypeStandard,
		SubType:         ZatcaSubTypeB2CSimplified,
		CurrencyCode:    "SAR",
		ICV:             icv,
		PIH:             pih,
		SellerName:      sellerName,
		SellerVATNumber: vatID,
		Subtotal:        numericToString(txn.Subtotal),
		DiscountAmount:  numericToString(txn.DiscountAmount),
		TaxAmount:       numericToString(txn.TaxAmount),
		TotalAmount:     numericToString(txn.TotalAmount),
		TaxCategoryID:   "S",
		TaxPercent:      "15.00",
		Lines:           xmlLines,
	}

	store, err := uc.repo.GetStore(ctx, txn.StoreID)
	if err != nil {
		return nil, nil, fmt.Errorf("ZATCA store not found: %w", err)
	}
	if store.OrganizationID != org.ID {
		return nil, nil, fmt.Errorf("ZATCA store belongs to a different organization")
	}
	if err := applySeller(&xmlInput, org, &store, uc.cfg.OrgVATID); err != nil {
		return nil, nil, err
	}
	signedDoc, err := signZatcaDocument(xmlInput, privKey, device)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to stamp ZATCA document: %w", err)
	}
	xmlHash := signedDoc.XMLHash
	qrCodeBase64 := signedDoc.QRCode
	rawXML := signedDoc.RawXML
	parsedUUID, _ := uuid.Parse(zatcaUUID)
	chainParams := repository.InsertChainEntryParams{
		EntityType:     "pos_transaction",
		EntityID:       strconv.Itoa(int(txn.ID)),
		DeviceConfigID: device.ID,
		OrganizationID: org.ID,
		ZatcaUuid:      parsedUUID,
		Icv:            icv,
		Pih:            pih,
		XmlHash:        xmlHash,
		ZatcaStatus:    repository.NullZatcaDocStatus{ZatcaDocStatus: repository.ZatcaDocStatusPending, Valid: true},
		QrCodeBase64:   pgtype.Text{String: qrCodeBase64, Valid: true},
		SignedXml:      pgtype.Text{String: string(rawXML), Valid: true},
	}

	chainEntry, err := uc.repo.InsertChainEntry(ctx, chainParams)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to insert document chain entry: %w", err)
	}

	return signedDoc, &chainEntry, nil
}

// SignCreditNote processes a B2C Simplified Credit Note (381) for a sales return, signs it, and generates the TLV QR code.
func (uc *ZatcaUseCase) SignCreditNote(
	ctx context.Context,
	ret *repository.SalesReturn,
	lines []repository.SalesReturnLine,
	origInvoiceNumber string,
	org *repository.Organization,
	device *repository.ZatcaDeviceConfig,
) (*SignedDocument, *repository.ZatcaDocumentChain, error) {
	if err := RequireZatcaSigningOwner(device, uc.cfg); err != nil {
		return nil, nil, err
	}
	enabled, err := OrganizationZatcaEnabled(org)
	if err != nil {
		return nil, nil, err
	}
	if !enabled {
		return nil, nil, ErrZatcaDisabled
	}

	privKey, err := DecodePrivateKeyPEM(device.PrivateKeyPem)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode device private key: %w", err)
	}

	icv, pih, err := uc.GetNextICVAndPIH(ctx, device.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get ICV/PIH: %w", err)
	}

	zatcaUUID := uuid.New().String()

	xmlLines := make([]InvoiceLineInput, 0, len(lines))
	for idx, line := range lines {
		lineNum := idx + 1
		if line.LineNumber.Valid {
			lineNum = int(line.LineNumber.Int32)
		}
		refund, err := decimalValue(numericToString(line.RefundAmount))
		if err != nil {
			return nil, nil, err
		}
		lineTaxStr := new(big.Rat).Mul(refund, big.NewRat(15, 100)).FloatString(2)
		xmlLines = append(xmlLines, InvoiceLineInput{
			LineNumber:  lineNum,
			ProductName: fmt.Sprintf("Returned Product #%d", line.ProductID),
			Quantity:    numericToString(line.Quantity),
			UnitCode:    "PCE",
			UnitPrice:   numericToString(line.UnitPrice),
			TaxAmount:   lineTaxStr,
			TaxPercent:  "15.00",
			TaxCategory: "S",
			LineTotal:   numericToString(line.RefundAmount),
		})
	}

	sellerName := org.LegalName.String
	if sellerName == "" {
		sellerName = org.Name
	}
	vatID := org.TaxID.String
	if vatID == "" {
		vatID = uc.cfg.OrgVATID
	}

	retDate := ret.ReturnDate.Time
	if retDate.IsZero() {
		retDate = time.Now()
	}

	xmlInput := InvoiceXMLInput{
		InvoiceNumber:         ret.ReturnNumber,
		OriginalInvoiceNumber: origInvoiceNumber,
		ZatcaUUID:             zatcaUUID,
		IssueDate:             retDate,
		InvoiceType:           ZatcaTypeCreditNote,
		SubType:               ZatcaSubTypeB2CSimplified,
		CurrencyCode:          "SAR",
		ICV:                   icv,
		PIH:                   pih,
		SellerName:            sellerName,
		SellerVATNumber:       vatID,
		Subtotal:              numericToString(ret.Subtotal),
		DiscountAmount:        "0.00",
		TaxAmount:             numericToString(ret.TaxAmount),
		TotalAmount:           numericToString(ret.TotalRefundAmount),
		TaxCategoryID:         "S",
		TaxPercent:            "15.00",
		Lines:                 xmlLines,
	}

	store, err := uc.repo.GetStore(ctx, ret.StoreID)
	if err != nil {
		return nil, nil, fmt.Errorf("ZATCA store not found: %w", err)
	}
	if store.OrganizationID != org.ID {
		return nil, nil, fmt.Errorf("ZATCA store belongs to a different organization")
	}
	if err := applySeller(&xmlInput, org, &store, uc.cfg.OrgVATID); err != nil {
		return nil, nil, err
	}
	signedDoc, err := signZatcaDocument(xmlInput, privKey, device)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to stamp ZATCA document: %w", err)
	}
	xmlHash := signedDoc.XMLHash
	qrCodeBase64 := signedDoc.QRCode
	rawXML := signedDoc.RawXML
	parsedUUID, _ := uuid.Parse(zatcaUUID)
	chainParams := repository.InsertChainEntryParams{
		EntityType:     "sales_return",
		EntityID:       strconv.Itoa(int(ret.ID)),
		DeviceConfigID: device.ID,
		OrganizationID: org.ID,
		ZatcaUuid:      parsedUUID,
		Icv:            icv,
		Pih:            pih,
		XmlHash:        xmlHash,
		ZatcaStatus:    repository.NullZatcaDocStatus{ZatcaDocStatus: repository.ZatcaDocStatusPending, Valid: true},
		QrCodeBase64:   pgtype.Text{String: qrCodeBase64, Valid: true},
		SignedXml:      pgtype.Text{String: string(rawXML), Valid: true},
	}

	chainEntry, err := uc.repo.InsertChainEntry(ctx, chainParams)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to insert document chain entry: %w", err)
	}

	return signedDoc, &chainEntry, nil
}

// GetDeviceConfigsDelta retrieves updated ZATCA device configs for delta sync.
func (uc *ZatcaUseCase) GetDeviceConfigsDelta(ctx context.Context, storeID int32, since time.Time) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not initialized", nil)
	}

	arg := repository.GetZatcaConfigsDeltaParams{
		StoreID:   pgtype.Int4{Int32: storeID, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: since, Valid: true},
	}

	configs, err := uc.repo.GetZatcaConfigsDelta(ctx, arg)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "ZATCA device configs fetched successfully", configs)
}
