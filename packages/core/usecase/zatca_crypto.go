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
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// ─── ZATCA Configuration ────────────────────────────────────────────────────

// ZatcaConfig holds ZATCA-specific configuration extracted from the app config.
type ZatcaConfig struct {
	Enabled  bool   // Legacy configuration; enablement uses organization metadata.
	Env      string // Legacy configuration; routing uses the saved device environment.
	BaseURL  string // Legacy configuration; device gateways use the fixed environment map.
	OrgVATID string // 15-digit VAT registration number
}

// GenesisKPIH is the exact Base64 PIH used for the very first invoice in a chain.
// This value is mandated by ZATCA and must be used verbatim.
const GenesisKPIH = "NWZlY2ViNjZmZmM4NmYzOGQ5NTI3ODZjNmQ2OTZjNzljMmRiYzIzOWRkNGU5MWI0NjcyOWQ3M2EyN2ZiNTdlOQ=="

// ─── Key Generation ─────────────────────────────────────────────────────────

// GenerateECDSAKeyPair generates a new secp256k1 ECDSA key pair.
// ZATCA mandates secp256k1 (not P-256).
func GenerateECDSAKeyPair() (*ecdsa.PrivateKey, error) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secp256k1 key: %w", err)
	}
	return privKey.ToECDSA(), nil
}

var oidSecp256k1 = asn1.ObjectIdentifier{1, 3, 132, 0, 10}

type sec1ECPrivateKey struct {
	Version       int
	PrivateKey    []byte
	NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
	PublicKey     asn1.BitString        `asn1:"optional,explicit,tag:1"`
}

// EncodePrivateKeyPEM serializes an ECDSA private key to PEM format (RFC 5915 SEC 1).
func EncodePrivateKeyPEM(key *ecdsa.PrivateKey) (string, error) {
	dBytes := key.D.Bytes()
	padded := make([]byte, 32)
	copy(padded[32-len(dBytes):], dBytes)

	pubBytes := elliptic.Marshal(key.Curve, key.X, key.Y)

	ecPriv := sec1ECPrivateKey{
		Version:       1,
		PrivateKey:    padded,
		NamedCurveOID: oidSecp256k1,
		PublicKey: asn1.BitString{
			Bytes:     pubBytes,
			BitLength: len(pubBytes) * 8,
		},
	}

	derBytes, err := asn1.Marshal(ecPriv)
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

	var ecPriv sec1ECPrivateKey
	if _, err := asn1.Unmarshal(block.Bytes, &ecPriv); err == nil && len(ecPriv.PrivateKey) > 0 {
		curve := Secp256k1()
		d := new(big.Int).SetBytes(ecPriv.PrivateKey)
		x, y := curve.ScalarBaseMult(ecPriv.PrivateKey)

		return &ecdsa.PrivateKey{
			PublicKey: ecdsa.PublicKey{
				Curve: curve,
				X:     x,
				Y:     y,
			},
			D: d,
		}, nil
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
	CommonName             string // e.g. "TST-886431145-399999999900003"
	SerialNumber           string // Device serial number
	OrganizationIdentifier string // Organization VAT registration number (SAN UID)
	OrgName                string // Organization legal name
	OrgUnit                string // e.g. "Riyadh Branch"
	CountryCode            string // "SA"
	InvoiceType            string // "1100" for Standard+Simplified
	Location               string // e.g. "Riyadh"
	Industry               string // e.g. "IT"
	IsSandbox              bool   // true = PREZATCA-Code-Signing, false = ZATCA-Code-Signing
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

	// Certificate template name encoded as an ASN.1 PrintableString.
	templateName := "ZATCA-Code-Signing"
	if params.IsSandbox {
		templateName = "PREZATCA-Code-Signing"
	}

	templateDER, err := asn1.MarshalWithParams(templateName, "printable")
	if err != nil {
		return "", fmt.Errorf("failed to encode certificate template: %w", err)
	}
	// Build custom extensions
	extensions := []pkix.Extension{
		// Certificate template name
		{
			Id:    asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 20, 2},
			Value: templateDER,
		},
	}

	// SAN (Subject Alternative Name) with ZATCA-specific directory names
	// These encode InvoiceType, Location, Industry as directoryName entries
	sanExt, err := buildZatcaSAN(params)
	if err != nil {
		return "", fmt.Errorf("failed to build SAN extension: %w", err)
	}
	extensions = append(extensions, sanExt)

	csrDER, err := createSecp256k1CSR(key, subject, extensions)
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
	// GeneralNames contains an explicitly tagged directoryName [4], whose
	// Name is an RDNSequence. Extension values must contain DER, not text.
	values := []pkix.AttributeTypeAndValue{
		{Type: asn1.ObjectIdentifier{2, 5, 4, 4}, Value: params.SerialNumber},
		{Type: asn1.ObjectIdentifier{0, 9, 2342, 19200300, 100, 1, 1}, Value: params.OrganizationIdentifier},
		{Type: asn1.ObjectIdentifier{2, 5, 4, 12}, Value: params.InvoiceType},
		{Type: asn1.ObjectIdentifier{2, 5, 4, 26}, Value: params.Location},
		{Type: asn1.ObjectIdentifier{2, 5, 4, 15}, Value: params.Industry},
	}
	rdns := make(pkix.RDNSequence, 0, len(values))
	for _, value := range values {
		rdns = append(rdns, pkix.RelativeDistinguishedNameSET{value})
	}
	nameDER, err := asn1.Marshal(rdns)
	if err != nil {
		return pkix.Extension{}, err
	}
	sanDER, err := asn1.Marshal([]asn1.RawValue{{Class: 2, Tag: 4, IsCompound: true, Bytes: nameDER}})
	if err != nil {
		return pkix.Extension{}, err
	}

	return pkix.Extension{
		Id:    asn1.ObjectIdentifier{2, 5, 29, 17}, // SAN OID
		Value: sanDER,
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
	ECDSAPublicKey string    // Tag 8: DER SubjectPublicKeyInfo (Base64)
	ZATCASignature string    // Tag 9: CA signature on the signing certificate (Base64)
}

// GenerateTLVQRCode encodes the 9 ZATCA mandatory tags into a TLV structure
// and returns the result as a Base64-encoded string suitable for QR code printing.
func GenerateTLVQRCode(data ZatcaQRData) (string, error) {
	tags := []TLVTag{
		{Tag: 1, Value: []byte(data.SellerName)},
		{Tag: 2, Value: []byte(data.VATNumber)},
		{Tag: 3, Value: []byte(data.Timestamp.UTC().Format("2006-01-02T15:04:05Z"))},
		{Tag: 4, Value: []byte(data.InvoiceTotal)},
		{Tag: 5, Value: []byte(data.VATTotal)},
	}

	// The gateway/SDK QR representation carries the Base64 text for tags 6 and
	// 7; tags 8 and 9 carry DER bytes. Validate the text before embedding it.
	xmlHashBytes, err := base64.StdEncoding.DecodeString(data.XMLHash)
	if err != nil {
		return "", fmt.Errorf("failed to decode XML hash: %w", err)
	}
	if len(xmlHashBytes) != 32 {
		return "", fmt.Errorf("invoice hash must contain 32 SHA-256 bytes")
	}
	tags = append(tags, TLVTag{Tag: 6, Value: []byte(data.XMLHash)})

	sigBytes, err := base64.StdEncoding.DecodeString(data.ECDSASignature)
	if err != nil {
		return "", fmt.Errorf("failed to decode ECDSA signature: %w", err)
	}
	if len(sigBytes) == 0 {
		return "", fmt.Errorf("invoice signature is required")
	}
	tags = append(tags, TLVTag{Tag: 7, Value: []byte(data.ECDSASignature)})

	pubKeyBytes, err := base64.StdEncoding.DecodeString(data.ECDSAPublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode ECDSA public key: %w", err)
	}
	tags = append(tags, TLVTag{Tag: 8, Value: pubKeyBytes})

	// Tag 9 is the CA's signature on the signing certificate, required for B2C.
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
		// This TLV uses one unsigned length byte, not ASN.1 BER lengths.
		length := len(tag.Value)
		if length > 255 {
			return "", fmt.Errorf("QR tag %d exceeds the 255-byte TLV limit", tag.Tag)
		}
		tlvBytes = append(tlvBytes, byte(length))
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
func Secp256k1() elliptic.Curve {
	return secp256k1.S256()
}
