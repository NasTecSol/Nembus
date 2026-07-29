package usecase

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// ─── ZATCA UBL 2.1 XML Builder ──────────────────────────────────────────────
//
// Maps Nembus DTOs (invoices, POS transactions) to ZATCA-compliant UBL 2.1 XML.
// Supports:
//   - Standard Tax Invoice (InvoiceTypeCode 388) for B2B clearance
//   - Simplified Tax Invoice (InvoiceTypeCode 388 subtype 0200000) for B2C reporting
//   - Credit Note (InvoiceTypeCode 381)
//   - Debit Note (InvoiceTypeCode 383)
//
// The generated XML is then canonicalized (C14N) and hashed for signing.

// ─── Invoice Type Mapping ───────────────────────────────────────────────────

// ZatcaInvoiceTypeCode maps Nembus invoice types to ZATCA UBL type codes.
type ZatcaInvoiceTypeCode string

const (
	ZatcaTypeStandard   ZatcaInvoiceTypeCode = "388" // Standard Tax Invoice
	ZatcaTypeCreditNote ZatcaInvoiceTypeCode = "381" // Credit Note
	ZatcaTypeDebitNote  ZatcaInvoiceTypeCode = "383" // Debit Note
)

// ZatcaInvoiceSubType defines the transaction nature (B2B or B2C).
type ZatcaInvoiceSubType string

const (
	ZatcaSubTypeB2BStandard  ZatcaInvoiceSubType = "0100000" // Standard (B2B Clearance)
	ZatcaSubTypeB2CSimplified ZatcaInvoiceSubType = "0200000" // Simplified (B2C Reporting)
)

// MapNembusInvoiceType maps a Nembus invoice_type enum value to ZATCA type code.
func MapNembusInvoiceType(nembusType string) ZatcaInvoiceTypeCode {
	switch nembusType {
	case "credit_note":
		return ZatcaTypeCreditNote
	case "debit_note":
		return ZatcaTypeDebitNote
	default:
		return ZatcaTypeStandard
	}
}

// ─── UBL XML Structures ─────────────────────────────────────────────────────

// UBLInvoice represents the top-level UBL 2.1 Invoice document.
type UBLInvoice struct {
	XMLName             xml.Name              `xml:"Invoice"`
	XmlnsUBL            string                `xml:"xmlns,attr"`
	XmlnsCAC            string                `xml:"xmlns:cac,attr"`
	XmlnsCBC            string                `xml:"xmlns:cbc,attr"`
	XmlnsExt            string                `xml:"xmlns:ext,attr"`
	ProfileID           string                `xml:"cbc:ProfileID"`
	ID                  string                `xml:"cbc:ID"`
	UUID                string                `xml:"cbc:UUID"`
	IssueDate           string                `xml:"cbc:IssueDate"`
	IssueTime           string                `xml:"cbc:IssueTime"`
	InvoiceTypeCode     UBLInvoiceTypeCode    `xml:"cbc:InvoiceTypeCode"`
	DocumentCurrencyCode string               `xml:"cbc:DocumentCurrencyCode"`
	TaxCurrencyCode     string                `xml:"cbc:TaxCurrencyCode,omitempty"`
	LineCountNumeric    int                   `xml:"cbc:LineCountNumeric"`
	AdditionalDocRefs   []UBLAdditionalDocRef `xml:"cac:AdditionalDocumentReference,omitempty"`
	Supplier            UBLParty              `xml:"cac:AccountingSupplierParty>cac:Party"`
	Customer            UBLParty              `xml:"cac:AccountingCustomerParty>cac:Party,omitempty"`
	Delivery            *UBLDelivery          `xml:"cac:Delivery,omitempty"`
	PaymentMeans        *UBLPaymentMeans      `xml:"cac:PaymentMeans,omitempty"`
	TaxTotal            []UBLTaxTotal         `xml:"cac:TaxTotal"`
	LegalMonetaryTotal  UBLLegalMonetaryTotal `xml:"cac:LegalMonetaryTotal"`
	InvoiceLines        []UBLInvoiceLine      `xml:"cac:InvoiceLine"`
}

// UBLInvoiceTypeCode carries the type code plus ZATCA subtype attribute.
type UBLInvoiceTypeCode struct {
	Value   string `xml:",chardata"`
	Name    string `xml:"name,attr,omitempty"` // ZATCA subtype: "0100000" or "0200000"
}

// UBLAdditionalDocRef holds additional document references (ICV, PIH, QR).
type UBLAdditionalDocRef struct {
	ID         string              `xml:"cbc:ID"`
	UUID       string              `xml:"cbc:UUID,omitempty"`
	Attachment *UBLAttachment      `xml:"cac:Attachment,omitempty"`
}

// UBLAttachment wraps embedded document content.
type UBLAttachment struct {
	EmbeddedBinaryObject UBLBinaryObject `xml:"cbc:EmbeddedDocumentBinaryObject"`
}

// UBLBinaryObject is a base64-encoded binary object with MIME type.
type UBLBinaryObject struct {
	MimeCode string `xml:"mimeCode,attr"`
	Value    string `xml:",chardata"`
}

// UBLParty represents a business party (supplier or customer).
type UBLParty struct {
	PartyIdentification *UBLPartyID    `xml:"cac:PartyIdentification,omitempty"`
	PostalAddress       UBLPostalAddr  `xml:"cac:PostalAddress"`
	PartyTaxScheme      UBLTaxScheme   `xml:"cac:PartyTaxScheme"`
	PartyLegalEntity    UBLLegalEntity `xml:"cac:PartyLegalEntity"`
}

// UBLPartyID holds a party identification (e.g., CR number, VAT ID).
type UBLPartyID struct {
	ID UBLSchemeID `xml:"cbc:ID"`
}

// UBLSchemeID is an ID with a scheme identifier attribute.
type UBLSchemeID struct {
	SchemeID string `xml:"schemeID,attr,omitempty"`
	Value    string `xml:",chardata"`
}

// UBLPostalAddr represents a postal address.
type UBLPostalAddr struct {
	StreetName           string          `xml:"cbc:StreetName,omitempty"`
	BuildingNumber       string          `xml:"cbc:BuildingNumber,omitempty"`
	PlotIdentification   string          `xml:"cbc:PlotIdentification,omitempty"`
	CitySubdivisionName  string          `xml:"cbc:CitySubdivisionName,omitempty"`
	CityName             string          `xml:"cbc:CityName,omitempty"`
	PostalZone           string          `xml:"cbc:PostalZone,omitempty"`
	CountrySubentity     string          `xml:"cbc:CountrySubentity,omitempty"`
	Country              UBLCountry      `xml:"cac:Country"`
}

// UBLCountry holds a country identification code.
type UBLCountry struct {
	IdentificationCode string `xml:"cbc:IdentificationCode"`
}

// UBLTaxScheme holds the party's tax registration info.
type UBLTaxScheme struct {
	CompanyID string `xml:"cbc:CompanyID"`
	TaxScheme struct {
		ID string `xml:"cbc:ID"`
	} `xml:"cac:TaxScheme"`
}

// UBLLegalEntity represents the legal registration of a party.
type UBLLegalEntity struct {
	RegistrationName string `xml:"cbc:RegistrationName"`
}

// UBLDelivery represents delivery details.
type UBLDelivery struct {
	ActualDeliveryDate string `xml:"cbc:ActualDeliveryDate,omitempty"`
	LatestDeliveryDate string `xml:"cbc:LatestDeliveryDate,omitempty"`
}

// UBLPaymentMeans describes how the invoice is paid.
type UBLPaymentMeans struct {
	PaymentMeansCode string `xml:"cbc:PaymentMeansCode"`
}

// UBLTaxTotal holds tax totals.
type UBLTaxTotal struct {
	TaxAmount    UBLAmount       `xml:"cbc:TaxAmount"`
	TaxSubtotals []UBLTaxSubtotal `xml:"cac:TaxSubtotal,omitempty"`
}

// UBLTaxSubtotal breaks down tax by category.
type UBLTaxSubtotal struct {
	TaxableAmount UBLAmount       `xml:"cbc:TaxableAmount"`
	TaxAmount     UBLAmount       `xml:"cbc:TaxAmount"`
	TaxCategory   UBLTaxCategory  `xml:"cac:TaxCategory"`
}

// UBLTaxCategory identifies the tax category and rate.
type UBLTaxCategory struct {
	ID        string `xml:"cbc:ID"`
	Percent   string `xml:"cbc:Percent"`
	TaxScheme struct {
		ID string `xml:"cbc:ID"`
	} `xml:"cac:TaxScheme"`
}

// UBLLegalMonetaryTotal holds the invoice monetary summary.
type UBLLegalMonetaryTotal struct {
	LineExtensionAmount UBLAmount `xml:"cbc:LineExtensionAmount"`
	TaxExclusiveAmount  UBLAmount `xml:"cbc:TaxExclusiveAmount"`
	TaxInclusiveAmount  UBLAmount `xml:"cbc:TaxInclusiveAmount"`
	AllowanceTotalAmount *UBLAmount `xml:"cbc:AllowanceTotalAmount,omitempty"`
	PrepaidAmount       *UBLAmount `xml:"cbc:PrepaidAmount,omitempty"`
	PayableAmount       UBLAmount `xml:"cbc:PayableAmount"`
}

// UBLAmount is a monetary value with currency attribute.
type UBLAmount struct {
	CurrencyID string `xml:"currencyID,attr"`
	Value      string `xml:",chardata"`
}

// UBLInvoiceLine represents a single line item on the invoice.
type UBLInvoiceLine struct {
	ID                  string         `xml:"cbc:ID"`
	InvoicedQuantity    UBLQuantity    `xml:"cbc:InvoicedQuantity"`
	LineExtensionAmount UBLAmount      `xml:"cbc:LineExtensionAmount"`
	TaxTotal            UBLTaxTotal    `xml:"cac:TaxTotal"`
	Item                UBLItem        `xml:"cac:Item"`
	Price               UBLPrice       `xml:"cac:Price"`
}

// UBLQuantity is a quantity with unit code.
type UBLQuantity struct {
	UnitCode string `xml:"unitCode,attr"`
	Value    string `xml:",chardata"`
}

// UBLItem describes the invoiced item.
type UBLItem struct {
	Name                string            `xml:"cbc:Name"`
	ClassifiedTaxCategory UBLTaxCategory  `xml:"cac:ClassifiedTaxCategory"`
}

// UBLPrice holds the unit price of the item.
type UBLPrice struct {
	PriceAmount UBLAmount `xml:"cbc:PriceAmount"`
}

// ─── XML Document Builder ───────────────────────────────────────────────────

// InvoiceXMLInput holds all the data needed to build a ZATCA-compliant UBL 2.1 XML.
type InvoiceXMLInput struct {
	// Invoice identification
	InvoiceNumber string
	ZatcaUUID     string
	IssueDate     time.Time
	InvoiceType   ZatcaInvoiceTypeCode
	SubType       ZatcaInvoiceSubType
	CurrencyCode  string

	// Chain data
	ICV int64  // Invoice Counter Value
	PIH string // Previous Invoice Hash (Base64)

	// Seller
	SellerName       string
	SellerVATNumber  string
	SellerStreet     string
	SellerCity       string
	SellerPostalCode string
	SellerCountry    string
	SellerCRNumber   string

	// Buyer (required for B2B, optional for B2C)
	BuyerName       string
	BuyerVATNumber  string
	BuyerStreet     string
	BuyerCity       string
	BuyerPostalCode string
	BuyerCountry    string

	// Totals
	Subtotal       string // Line extension amount (before tax)
	DiscountAmount string // Total discount
	TaxAmount      string // Total VAT
	TotalAmount    string // Grand total (including tax)

	// Tax breakdown
	TaxCategoryID string // e.g., "S" for standard rate
	TaxPercent    string // e.g., "15.00"

	// Line items
	Lines []InvoiceLineInput

	// QR Code (populated after signing)
	QRCodeBase64 string
}

// InvoiceLineInput holds a single invoice line for XML generation.
type InvoiceLineInput struct {
	LineNumber   int
	ProductName  string
	Quantity     string
	UnitCode     string // e.g., "PCE" (piece), "KGM" (kilogram)
	UnitPrice    string
	DiscountAmt  string
	TaxAmount    string
	TaxPercent   string
	TaxCategory  string // e.g., "S" (standard), "E" (exempt), "Z" (zero-rated)
	LineTotal    string
}

// BuildInvoiceXML generates a complete UBL 2.1 XML document from the given input.
func BuildInvoiceXML(input InvoiceXMLInput) ([]byte, error) {
	inv := UBLInvoice{
		XmlnsUBL: "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2",
		XmlnsCAC: "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2",
		XmlnsCBC: "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2",
		XmlnsExt: "urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2",
		ProfileID:            "reporting:1.0",
		ID:                   input.InvoiceNumber,
		UUID:                 input.ZatcaUUID,
		IssueDate:            input.IssueDate.Format("2006-01-02"),
		IssueTime:            input.IssueDate.Format("15:04:05"),
		InvoiceTypeCode: UBLInvoiceTypeCode{
			Value: string(input.InvoiceType),
			Name:  string(input.SubType),
		},
		DocumentCurrencyCode: input.CurrencyCode,
		TaxCurrencyCode:      "SAR",
		LineCountNumeric:     len(input.Lines),
	}

	// Additional Document References: ICV, PIH
	inv.AdditionalDocRefs = []UBLAdditionalDocRef{
		{
			ID: "ICV",
			UUID: fmt.Sprintf("%d", input.ICV),
		},
		{
			ID: "PIH",
			Attachment: &UBLAttachment{
				EmbeddedBinaryObject: UBLBinaryObject{
					MimeCode: "text/plain",
					Value:    input.PIH,
				},
			},
		},
	}

	// QR Code reference (added after signing)
	if input.QRCodeBase64 != "" {
		inv.AdditionalDocRefs = append(inv.AdditionalDocRefs, UBLAdditionalDocRef{
			ID: "QR",
			Attachment: &UBLAttachment{
				EmbeddedBinaryObject: UBLBinaryObject{
					MimeCode: "text/plain",
					Value:    input.QRCodeBase64,
				},
			},
		})
	}

	// Supplier (Seller)
	inv.Supplier = UBLParty{
		PartyIdentification: &UBLPartyID{
			ID: UBLSchemeID{SchemeID: "CRN", Value: input.SellerCRNumber},
		},
		PostalAddress: UBLPostalAddr{
			StreetName: input.SellerStreet,
			CityName:   input.SellerCity,
			PostalZone: input.SellerPostalCode,
			Country:    UBLCountry{IdentificationCode: input.SellerCountry},
		},
		PartyTaxScheme: UBLTaxScheme{
			CompanyID: input.SellerVATNumber,
		},
		PartyLegalEntity: UBLLegalEntity{
			RegistrationName: input.SellerName,
		},
	}
	inv.Supplier.PartyTaxScheme.TaxScheme.ID = "VAT"

	// Customer (Buyer) — required for B2B
	if input.BuyerName != "" {
		inv.Customer = UBLParty{
			PostalAddress: UBLPostalAddr{
				StreetName: input.BuyerStreet,
				CityName:   input.BuyerCity,
				PostalZone: input.BuyerPostalCode,
				Country:    UBLCountry{IdentificationCode: input.BuyerCountry},
			},
			PartyTaxScheme: UBLTaxScheme{
				CompanyID: input.BuyerVATNumber,
			},
			PartyLegalEntity: UBLLegalEntity{
				RegistrationName: input.BuyerName,
			},
		}
		inv.Customer.PartyTaxScheme.TaxScheme.ID = "VAT"
	}

	// Tax totals
	inv.TaxTotal = []UBLTaxTotal{
		{
			TaxAmount: UBLAmount{CurrencyID: input.CurrencyCode, Value: input.TaxAmount},
			TaxSubtotals: []UBLTaxSubtotal{
				{
					TaxableAmount: UBLAmount{CurrencyID: input.CurrencyCode, Value: input.Subtotal},
					TaxAmount:     UBLAmount{CurrencyID: input.CurrencyCode, Value: input.TaxAmount},
					TaxCategory: UBLTaxCategory{
						ID:      input.TaxCategoryID,
						Percent: input.TaxPercent,
					},
				},
			},
		},
	}
	inv.TaxTotal[0].TaxSubtotals[0].TaxCategory.TaxScheme.ID = "VAT"

	// Legal monetary totals
	inv.LegalMonetaryTotal = UBLLegalMonetaryTotal{
		LineExtensionAmount: UBLAmount{CurrencyID: input.CurrencyCode, Value: input.Subtotal},
		TaxExclusiveAmount:  UBLAmount{CurrencyID: input.CurrencyCode, Value: input.Subtotal},
		TaxInclusiveAmount:  UBLAmount{CurrencyID: input.CurrencyCode, Value: input.TotalAmount},
		PayableAmount:       UBLAmount{CurrencyID: input.CurrencyCode, Value: input.TotalAmount},
	}

	if input.DiscountAmount != "" && input.DiscountAmount != "0" && input.DiscountAmount != "0.00" {
		inv.LegalMonetaryTotal.AllowanceTotalAmount = &UBLAmount{
			CurrencyID: input.CurrencyCode,
			Value:      input.DiscountAmount,
		}
	}

	// Invoice lines
	for _, line := range input.Lines {
		ublLine := UBLInvoiceLine{
			ID:               fmt.Sprintf("%d", line.LineNumber),
			InvoicedQuantity: UBLQuantity{UnitCode: line.UnitCode, Value: line.Quantity},
			LineExtensionAmount: UBLAmount{CurrencyID: input.CurrencyCode, Value: line.LineTotal},
			TaxTotal: UBLTaxTotal{
				TaxAmount: UBLAmount{CurrencyID: input.CurrencyCode, Value: line.TaxAmount},
			},
			Item: UBLItem{
				Name: line.ProductName,
				ClassifiedTaxCategory: UBLTaxCategory{
					ID:      line.TaxCategory,
					Percent: line.TaxPercent,
				},
			},
			Price: UBLPrice{
				PriceAmount: UBLAmount{CurrencyID: input.CurrencyCode, Value: line.UnitPrice},
			},
		}
		ublLine.Item.ClassifiedTaxCategory.TaxScheme.ID = "VAT"
		inv.InvoiceLines = append(inv.InvoiceLines, ublLine)
	}

	// Marshal to XML
	xmlBytes, err := xml.MarshalIndent(inv, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal UBL XML: %w", err)
	}

	// Prepend XML declaration
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	buf.Write(xmlBytes)

	return buf.Bytes(), nil
}

// ─── XML Canonicalization (C14N) ────────────────────────────────────────────

// CanonicalizeXML performs a simplified XML canonicalization (C14N 1.1).
// For ZATCA compliance, the key requirements are:
//   - Normalize attribute ordering
//   - Normalize namespace declarations
//   - Remove XML declaration
//   - Use UTF-8 encoding
//
// Note: This is a simplified implementation suitable for ZATCA's requirements.
// For production, consider using a full C14N library.
func CanonicalizeXML(xmlData []byte) ([]byte, error) {
	// Remove XML declaration if present
	content := string(xmlData)
	if idx := strings.Index(content, "?>"); idx >= 0 {
		content = strings.TrimSpace(content[idx+2:])
	}

	// Remove whitespace between tags for canonical form
	// Parse and re-serialize to normalize
	decoder := xml.NewDecoder(strings.NewReader(content))
	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		if err := encoder.EncodeToken(token); err != nil {
			return nil, fmt.Errorf("failed to encode token during C14N: %w", err)
		}
	}
	if err := encoder.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush C14N encoder: %w", err)
	}

	return buf.Bytes(), nil
}

// ─── Signed Document ────────────────────────────────────────────────────────

// SignedDocument holds all the outputs from signing an invoice/transaction.
type SignedDocument struct {
	RawXML       []byte // Original UBL 2.1 XML
	CanonicalXML []byte // Canonicalized XML (C14N)
	XMLHash      string // SHA-256 hash of canonical XML (Base64)
	Signature    string // ECDSA signature of hash (Base64)
	QRCode       string // TLV QR code (Base64)
	ZatcaUUID    string // UUID used for this document
	ICV          int64  // Invoice Counter Value
	PIH          string // Previous Invoice Hash used
}
