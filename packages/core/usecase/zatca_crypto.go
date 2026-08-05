package usecase

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

// ─── ZATCA Configuration ────────────────────────────────────────────────────

// ZatcaConfig holds ZATCA-specific configuration extracted from the app config.
type ZatcaConfig struct {
	Enabled  bool   // Master toggle — false bypasses all ZATCA logic
	Env      string // "sandbox" or "production"
	BaseURL  string // ZATCA API base URL (derived from Env)
	OrgVATID string // 15-digit VAT registration number
}

// GenesisKPIH is the exact Base64 PIH used for the very first invoice in a chain.
// This value is mandated by ZATCA and must be used verbatim.
const GenesisKPIH = "NWZlY2ViNjZmZmM4NmYzOGQ5NTI3ODZjNmQ2OTZjNzljMmRiYzIzOWRkNGU5MWI0NjcyOWQ3M2EyN2ZiNTdlOQ=="

// ─── Key Generation ─────────────────────────────────────────────────────────

// GenerateECDSAKeyPair generates a new secp256k1 ECDSA key pair.
// ZATCA mandates secp256k1 (not P-256). Go's crypto/elliptic does not expose
// secp256k1 directly, so we use the curve parameters.
func GenerateECDSAKeyPair() (*ecdsa.PrivateKey, error) {
	curve := Secp256k1()
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secp256k1 key: %w", err)
	}
	return key, nil
}

// EncodePrivateKeyPEM serializes an ECDSA private key to PEM format.
func EncodePrivateKeyPEM(key *ecdsa.PrivateKey) (string, error) {
	derBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("failed to marshal private key: %w", err)
	}
	block := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: derBytes,
	}
	return string(pem.EncodeToMemory(block)), nil
}

// DecodePrivateKeyPEM parses a PEM-encoded ECDSA private key.
func DecodePrivateKeyPEM(pemStr string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("invalid PEM block type")
	}
	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	return key, nil
}

// ─── CSR Generation ─────────────────────────────────────────────────────────

// ZatcaCSRParams holds the parameters needed to generate a ZATCA-compliant CSR.
type ZatcaCSRParams struct {
	CommonName    string // e.g. "TST-886431145-399999999900003"
	SerialNumber  string // Device serial number
	OrgName       string // Organization legal name
	OrgUnit       string // e.g. "Riyadh Branch"
	CountryCode   string // "SA"
	InvoiceType   string // "1100" for Standard+Simplified
	Location      string // e.g. "Riyadh"
	Industry      string // e.g. "IT"
	IsSandbox     bool   // true = PREZATCA-Code-Signing, false = ZATCA-Code-Signing
}

// GenerateCSR creates a PKCS#10 Certificate Signing Request for ZATCA device onboarding.
// The CSR includes ZATCA-specific OID extensions as required by the FATOORA platform.
func GenerateCSR(key *ecdsa.PrivateKey, params ZatcaCSRParams) (string, error) {
	subject := pkix.Name{
		CommonName:         params.CommonName,
		SerialNumber:       params.SerialNumber,
		Organization:       []string{params.OrgName},
		OrganizationalUnit: []string{params.OrgUnit},
		Country:            []string{params.CountryCode},
	}

	// ZATCA OID extensions
	// OID 2.16.840.1.114171.500.2 — Certificate Template Name
	templateName := "ZATCA-Code-Signing"
	if params.IsSandbox {
		templateName = "PREZATCA-Code-Signing"
	}

	// Build custom extensions
	extensions := []pkix.Extension{
		// Certificate template name
		{
			Id:    asn1.ObjectIdentifier{2, 16, 840, 1, 114171, 500, 2},
			Value: []byte(templateName),
		},
	}

	// SAN (Subject Alternative Name) with ZATCA-specific directory names
	// These encode InvoiceType, Location, Industry as directoryName entries
	sanExt, err := buildZatcaSAN(params)
	if err != nil {
		return "", fmt.Errorf("failed to build SAN extension: %w", err)
	}
	extensions = append(extensions, sanExt)

	template := &x509.CertificateRequest{
		Subject:            subject,
		SignatureAlgorithm: x509.ECDSAWithSHA256,
		ExtraExtensions:    extensions,
	}

	csrDER, err := x509.CreateCertificateRequest(rand.Reader, template, key)
	if err != nil {
		return "", fmt.Errorf("failed to create CSR: %w", err)
	}

	block := &pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrDER,
	}
	return string(pem.EncodeToMemory(block)), nil
}

// buildZatcaSAN constructs the Subject Alternative Name extension with ZATCA directory entries.
func buildZatcaSAN(params ZatcaCSRParams) (pkix.Extension, error) {
	// ZATCA expects SAN with directoryName containing:
	// - UID (OID 0.9.2342.19200300.100.1.1) = InvoiceType (e.g., "1100")
	// - title (OID 2.5.4.12) = Location
	// - registeredAddress (OID 2.5.4.26) = Industry
	// - businessCategory (OID 2.5.4.15) = Industry

	type directoryName struct {
		UID      string `asn1:"utf8,tag:0,optional"`
		Title    string `asn1:"utf8,optional"`
		Category string `asn1:"utf8,optional"`
	}

	// Encode as raw ASN1 for SAN extension
	sanValue := fmt.Sprintf("UID=%s, title=%s, registeredAddress=%s, businessCategory=%s",
		params.InvoiceType, params.Location, params.Location, params.Industry)

	return pkix.Extension{
		Id:    asn1.ObjectIdentifier{2, 5, 29, 17}, // SAN OID
		Value: []byte(sanValue),
	}, nil
}

// ─── Hashing & Signing ──────────────────────────────────────────────────────

// HashSHA256Base64 computes SHA-256 hash of the input and returns Base64-encoded result.
func HashSHA256Base64(data []byte) string {
	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}

// SignHash signs a SHA-256 hash using ECDSA with the given private key.
// Returns the DER-encoded ECDSA signature as Base64.
func SignHash(key *ecdsa.PrivateKey, hash []byte) (string, error) {
	r, s, err := ecdsa.Sign(rand.Reader, key, hash)
	if err != nil {
		return "", fmt.Errorf("failed to sign hash: %w", err)
	}

	// DER encode the signature (r, s)
	type ecdsaSignature struct {
		R, S *big.Int
	}
	sig, err := asn1.Marshal(ecdsaSignature{R: r, S: s})
	if err != nil {
		return "", fmt.Errorf("failed to marshal signature: %w", err)
	}

	return base64.StdEncoding.EncodeToString(sig), nil
}

// SignXMLHash computes SHA-256 of the XML data, signs it, and returns both the hash and signature.
func SignXMLHash(key *ecdsa.PrivateKey, xmlData []byte) (hash string, signature string, err error) {
	hashBytes := sha256.Sum256(xmlData)
	hash = base64.StdEncoding.EncodeToString(hashBytes[:])

	signature, err = SignHash(key, hashBytes[:])
	if err != nil {
		return "", "", err
	}

	return hash, signature, nil
}

// ─── TLV QR Code Generator ─────────────────────────────────────────────────

// TLVTag represents a single Tag-Length-Value entry for the ZATCA QR code.
type TLVTag struct {
	Tag   uint8
	Value []byte
}

// ZatcaQRData holds the 9 mandatory fields for ZATCA QR code generation.
type ZatcaQRData struct {
	SellerName     string    // Tag 1: Seller's name (UTF-8)
	VATNumber      string    // Tag 2: VAT registration number
	Timestamp      time.Time // Tag 3: Invoice timestamp (ISO 8601)
	InvoiceTotal   string    // Tag 4: Invoice total with VAT
	VATTotal       string    // Tag 5: VAT total
	XMLHash        string    // Tag 6: SHA-256 hash of XML (Base64)
	ECDSASignature string    // Tag 7: ECDSA digital signature (Base64)
	ECDSAPublicKey string    // Tag 8: ECDSA public key (Base64)
	ZATCASignature string    // Tag 9: ZATCA's cryptographic stamp (from clearance response, if available)
}

// GenerateTLVQRCode encodes the 9 ZATCA mandatory tags into a TLV structure
// and returns the result as a Base64-encoded string suitable for QR code printing.
func GenerateTLVQRCode(data ZatcaQRData) (string, error) {
	tags := []TLVTag{
		{Tag: 1, Value: []byte(data.SellerName)},
		{Tag: 2, Value: []byte(data.VATNumber)},
		{Tag: 3, Value: []byte(data.Timestamp.Format("2006-01-02T15:04:05Z"))},
		{Tag: 4, Value: []byte(data.InvoiceTotal)},
		{Tag: 5, Value: []byte(data.VATTotal)},
	}

	// Tags 6-9 are binary (Base64-decoded before TLV encoding)
	xmlHashBytes, err := base64.StdEncoding.DecodeString(data.XMLHash)
	if err != nil {
		return "", fmt.Errorf("failed to decode XML hash: %w", err)
	}
	tags = append(tags, TLVTag{Tag: 6, Value: xmlHashBytes})

	sigBytes, err := base64.StdEncoding.DecodeString(data.ECDSASignature)
	if err != nil {
		return "", fmt.Errorf("failed to decode ECDSA signature: %w", err)
	}
	tags = append(tags, TLVTag{Tag: 7, Value: sigBytes})

	pubKeyBytes, err := base64.StdEncoding.DecodeString(data.ECDSAPublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode ECDSA public key: %w", err)
	}
	tags = append(tags, TLVTag{Tag: 8, Value: pubKeyBytes})

	// Tag 9 (ZATCA's stamp) may be empty for B2C simplified invoices
	if data.ZATCASignature != "" {
		zatcaSigBytes, err := base64.StdEncoding.DecodeString(data.ZATCASignature)
		if err != nil {
			return "", fmt.Errorf("failed to decode ZATCA signature: %w", err)
		}
		tags = append(tags, TLVTag{Tag: 9, Value: zatcaSigBytes})
	}

	// Encode all tags into TLV byte stream
	var tlvBytes []byte
	for _, tag := range tags {
		tlvBytes = append(tlvBytes, tag.Tag)
		// Length encoding: use variable-length encoding for lengths > 127
		length := len(tag.Value)
		if length <= 127 {
			tlvBytes = append(tlvBytes, byte(length))
		} else if length <= 255 {
			tlvBytes = append(tlvBytes, 0x81, byte(length))
		} else {
			lenBytes := make([]byte, 2)
			binary.BigEndian.PutUint16(lenBytes, uint16(length))
			tlvBytes = append(tlvBytes, 0x82)
			tlvBytes = append(tlvBytes, lenBytes...)
		}
		tlvBytes = append(tlvBytes, tag.Value...)
	}

	return base64.StdEncoding.EncodeToString(tlvBytes), nil
}

// EncodePublicKeyBase64 encodes an ECDSA public key as uncompressed point in Base64.
func EncodePublicKeyBase64(key *ecdsa.PublicKey) string {
	pubKeyBytes := elliptic.Marshal(key.Curve, key.X, key.Y)
	return base64.StdEncoding.EncodeToString(pubKeyBytes)
}

// ─── secp256k1 Curve Definition ─────────────────────────────────────────────

// Secp256k1 returns the secp256k1 elliptic curve used by ZATCA for ECDSA signing.
// Go's standard library does not include secp256k1, so we define the parameters here.
func Secp256k1() elliptic.Curve {
	return secp256k1Curve
}

var secp256k1Curve = &elliptic.CurveParams{
	P:       fromHex("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F"),
	N:       fromHex("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141"),
	B:       big.NewInt(7),
	Gx:      fromHex("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798"),
	Gy:      fromHex("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8"),
	BitSize: 256,
	Name:    "secp256k1",
}

func fromHex(hex string) *big.Int {
	n := new(big.Int)
	n.SetString(hex, 16)
	return n
}
