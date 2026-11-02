package usecase

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/beevik/etree"
)

const (
	zatcaDS     = "http://www.w3.org/2000/09/xmldsig#"
	zatcaXades  = "http://uri.etsi.org/01903/v1.3.2#"
	zatcaSHA256 = "http://www.w3.org/2001/04/xmlenc#sha256"
)

// The SDK uses Base64 of ASCII hex for certificate and SignedProperties digests;
// the invoice digest itself remains Base64 of the raw 32-byte SHA-256 digest.
func zatcaHexDigest(data []byte) string {
	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString([]byte(hex.EncodeToString(hash[:])))
}

func xmlElement(parent *etree.Element, name, text string) *etree.Element {
	e := parent.CreateElement(name)
	if text != "" {
		e.SetText(text)
	}
	return e
}

func signedProperties(cert *zatcaCertificate, signingTime time.Time) *etree.Element {
	props := etree.NewElement("xades:SignedProperties")
	props.CreateAttr("Id", "xadesSignedProperties")
	p := xmlElement(props, "xades:SignedSignatureProperties", "")
	xmlElement(p, "xades:SigningTime", signingTime.UTC().Format("2006-01-02T15:04:05Z"))
	c := xmlElement(xmlElement(p, "xades:SigningCertificate", ""), "xades:Cert", "")
	digest := xmlElement(c, "xades:CertDigest", "")
	xmlElement(digest, "ds:DigestMethod", "").CreateAttr("Algorithm", zatcaSHA256)
	xmlElement(digest, "ds:DigestValue", zatcaHexDigest([]byte(base64.StdEncoding.EncodeToString(cert.DER))))
	issuer := xmlElement(c, "xades:IssuerSerial", "")
	xmlElement(issuer, "ds:X509IssuerName", cert.Issuer)
	xmlElement(issuer, "ds:X509SerialNumber", cert.Serial.String())
	return props
}

// This hash-only projection follows ZATCA's SignedProperties instructions. It is
// derived from the embedded element so signing time and certificate cannot drift.
func signedPropertiesHash(props *etree.Element) (string, error) {
	copy := props.Copy()
	copy.Attr = append([]etree.Attr{{Space: "xmlns", Key: "xades", Value: zatcaXades}}, copy.Attr...)
	var addDS func(*etree.Element)
	addDS = func(e *etree.Element) {
		if e.Space == "ds" {
			e.Attr = append([]etree.Attr{{Space: "xmlns", Key: "ds", Value: zatcaDS}}, e.Attr...)
		}
		for _, child := range e.ChildElements() {
			addDS(child)
		}
	}
	addDS(copy)
	doc := etree.NewDocument()
	doc.SetRoot(copy)
	data, err := doc.WriteToBytes()
	if err != nil {
		return "", err
	}
	return zatcaHexDigest(data), nil
}

func signZatcaDocument(input InvoiceXMLInput, key *ecdsa.PrivateKey, device *repository.ZatcaDeviceConfig) (*SignedDocument, error) {
	token := device.ProductionCsid.String
	if token == "" {
		token = device.ComplianceCsid.String
	}
	cert, err := parseZatcaCertificate(token)
	if err != nil {
		return nil, err
	}
	if cert.PublicKey.X.Cmp(key.X) != 0 || cert.PublicKey.Y.Cmp(key.Y) != 0 {
		return nil, fmt.Errorf("device certificate does not match the signing private key")
	}
	if time.Now().Before(cert.NotBefore) || !time.Now().Before(cert.NotAfter) {
		return nil, fmt.Errorf("device signing certificate is not currently valid")
	}
	input.QRCodeBase64 = ""
	raw, err := BuildInvoiceXML(input)
	if err != nil {
		return nil, err
	}
	canonical, err := CanonicalizeXML(raw)
	if err != nil {
		return nil, err
	}
	hash, signature, err := SignXMLHash(key, canonical)
	if err != nil {
		return nil, err
	}
	qrInput := ZatcaQRData{SellerName: input.SellerName, VATNumber: input.SellerVATNumber, Timestamp: input.IssueDate.UTC(), InvoiceTotal: input.TotalAmount,
		VATTotal: input.TaxAmount, XMLHash: hash, ECDSASignature: signature, ECDSAPublicKey: base64.StdEncoding.EncodeToString(cert.SPKI)}
	if input.SubType == ZatcaSubTypeB2CSimplified {
		qrInput.ZATCASignature = base64.StdEncoding.EncodeToString(cert.Signature)
	}
	qr, err := GenerateTLVQRCode(qrInput)
	if err != nil {
		return nil, err
	}
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(raw); err != nil {
		return nil, err
	}
	root := doc.Root()
	extensions := etree.NewElement("ext:UBLExtensions")
	extension := xmlElement(extensions, "ext:UBLExtension", "")
	xmlElement(extension, "ext:ExtensionURI", "urn:oasis:names:specification:ubl:dsig:enveloped:xades")
	content := xmlElement(extension, "ext:ExtensionContent", "")
	ublSignatures := xmlElement(content, "sig:UBLDocumentSignatures", "")
	ublSignatures.CreateAttr("xmlns:sig", "urn:oasis:names:specification:ubl:schema:xsd:CommonSignatureComponents-2")
	ublSignatures.CreateAttr("xmlns:sac", "urn:oasis:names:specification:ubl:schema:xsd:SignatureAggregateComponents-2")
	ublSignatures.CreateAttr("xmlns:sbc", "urn:oasis:names:specification:ubl:schema:xsd:SignatureBasicComponents-2")
	info := xmlElement(ublSignatures, "sac:SignatureInformation", "")
	xmlElement(info, "cbc:ID", "urn:oasis:names:specification:ubl:signature:1")
	xmlElement(info, "sbc:ReferencedSignatureID", "urn:oasis:names:specification:ubl:signature:Invoice")
	ds := xmlElement(info, "ds:Signature", "")
	ds.CreateAttr("xmlns:ds", zatcaDS)
	ds.CreateAttr("Id", "signature")
	si := xmlElement(ds, "ds:SignedInfo", "")
	xmlElement(si, "ds:CanonicalizationMethod", "").CreateAttr("Algorithm", "http://www.w3.org/2006/12/xml-c14n11")
	xmlElement(si, "ds:SignatureMethod", "").CreateAttr("Algorithm", "http://www.w3.org/2001/04/xmldsig-more#ecdsa-sha256")
	ref := xmlElement(si, "ds:Reference", "")
	ref.CreateAttr("Id", "invoiceSignedData")
	ref.CreateAttr("URI", "")
	transforms := xmlElement(ref, "ds:Transforms", "")
	for _, xpath := range []string{"not(//ancestor-or-self::ext:UBLExtensions)", "not(//ancestor-or-self::cac:Signature)", "not(//ancestor-or-self::cac:AdditionalDocumentReference[cbc:ID='QR'])"} {
		t := xmlElement(transforms, "ds:Transform", "")
		t.CreateAttr("Algorithm", "http://www.w3.org/TR/1999/REC-xpath-19991116")
		xmlElement(t, "ds:XPath", xpath)
	}
	xmlElement(transforms, "ds:Transform", "").CreateAttr("Algorithm", "http://www.w3.org/2006/12/xml-c14n11")
	xmlElement(ref, "ds:DigestMethod", "").CreateAttr("Algorithm", zatcaSHA256)
	xmlElement(ref, "ds:DigestValue", hash)
	props := signedProperties(cert, time.Now())
	propsHash, err := signedPropertiesHash(props)
	if err != nil {
		return nil, err
	}
	propsRef := xmlElement(si, "ds:Reference", "")
	propsRef.CreateAttr("Type", "http://www.w3.org/2000/09/xmldsig#SignatureProperties")
	propsRef.CreateAttr("URI", "#xadesSignedProperties")
	xmlElement(propsRef, "ds:DigestMethod", "").CreateAttr("Algorithm", zatcaSHA256)
	xmlElement(propsRef, "ds:DigestValue", propsHash)
	xmlElement(ds, "ds:SignatureValue", signature)
	xmlElement(xmlElement(xmlElement(ds, "ds:KeyInfo", ""), "ds:X509Data", ""), "ds:X509Certificate", base64.StdEncoding.EncodeToString(cert.DER))
	object := xmlElement(ds, "ds:Object", "")
	qualifying := xmlElement(object, "xades:QualifyingProperties", "")
	qualifying.CreateAttr("xmlns:xades", zatcaXades)
	qualifying.CreateAttr("Target", "signature")
	qualifying.AddChild(props)
	root.InsertChildAt(0, extensions)
	// Insert excluded nodes without adding whitespace to the invoice's hashed
	// node set. Reformatting after stamping would invalidate its digest.
	qrRef := etree.NewElement("cac:AdditionalDocumentReference")
	xmlElement(qrRef, "cbc:ID", "QR")
	qrValue := xmlElement(xmlElement(qrRef, "cac:Attachment", ""), "cbc:EmbeddedDocumentBinaryObject", qr)
	qrValue.CreateAttr("mimeCode", "text/plain")
	refSignature := etree.NewElement("cac:Signature")
	xmlElement(refSignature, "cbc:ID", "urn:oasis:names:specification:ubl:signature:Invoice")
	xmlElement(refSignature, "cbc:SignatureMethod", "urn:oasis:names:specification:ubl:dsig:enveloped:xades")
	for _, child := range root.ChildElements() {
		if child.Tag == "AccountingSupplierParty" && child.NamespaceURI() == zatcaCAC {
			root.InsertChild(child, qrRef)
			root.InsertChild(child, refSignature)
			break
		}
	}
	finalXML, err := doc.WriteToBytes()
	if err != nil {
		return nil, err
	}
	finalCanonical, err := CanonicalizeXML(finalXML)
	if err != nil || !bytes.Equal(finalCanonical, canonical) {
		return nil, fmt.Errorf("internal error: stamping changed the invoice digest")
	}
	return &SignedDocument{RawXML: finalXML, CanonicalXML: canonical, XMLHash: hash, Signature: signature, QRCode: qr, ZatcaUUID: input.ZatcaUUID, ICV: input.ICV, PIH: input.PIH}, nil
}
