package usecase

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"time"
)

type zatcaCertificate struct {
	DER, SPKI, Signature []byte
	Issuer               string
	Serial               *big.Int
	PublicKey            *ecdsa.PublicKey
	NotBefore, NotAfter  time.Time
}

type zatcaCertificateASN struct {
	TBS struct {
		Raw       asn1.RawContent
		Version   int `asn1:"optional,explicit,tag:0,default:0"`
		Serial    *big.Int
		Signature pkix.AlgorithmIdentifier
		Issuer    asn1.RawValue
		Validity  struct{ NotBefore, NotAfter time.Time }
		Subject   asn1.RawValue
		PublicKey struct {
			Raw       asn1.RawContent
			Algorithm pkix.AlgorithmIdentifier
			Key       asn1.BitString
		}
		IssuerUniqueID  asn1.BitString   `asn1:"optional,tag:1"`
		SubjectUniqueID asn1.BitString   `asn1:"optional,tag:2"`
		Extensions      []pkix.Extension `asn1:"optional,explicit,tag:3"`
	}
	Algorithm pkix.AlgorithmIdentifier
	Signature asn1.BitString
}

// ZATCA's secp256k1 certificates cannot be decoded by crypto/x509. Decode the
// standard certificate envelope while retaining the original DER and SPKI.
func parseZatcaCertificate(token string) (*zatcaCertificate, error) {
	data := []byte(strings.TrimSpace(token))
	for i := 0; i < 4; i++ {
		if block, _ := pem.Decode(data); block != nil {
			if block.Type != "CERTIFICATE" {
				return nil, fmt.Errorf("signing token is not a certificate")
			}
			data = block.Bytes
		}
		var raw zatcaCertificateASN
		if rest, err := asn1.Unmarshal(data, &raw); err == nil && len(rest) == 0 && raw.TBS.Serial != nil {
			var curve asn1.ObjectIdentifier
			if !raw.TBS.PublicKey.Algorithm.Algorithm.Equal(asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}) {
				return nil, fmt.Errorf("certificate must contain an EC public key")
			}
			if _, err := asn1.Unmarshal(raw.TBS.PublicKey.Algorithm.Parameters.FullBytes, &curve); err != nil || !curve.Equal(oidSecp256k1) {
				return nil, fmt.Errorf("certificate must use secp256k1")
			}
			x, y := elliptic.Unmarshal(Secp256k1(), raw.TBS.PublicKey.Key.Bytes)
			if x == nil || y == nil {
				return nil, fmt.Errorf("certificate has an invalid public key")
			}
			var issuerRDN pkix.RDNSequence
			if _, err := asn1.Unmarshal(raw.TBS.Issuer.FullBytes, &issuerRDN); err != nil {
				return nil, fmt.Errorf("invalid certificate issuer")
			}
			// Preserve RFC 4514 RDN ordering and escaping, using the SDK separator.
			return &zatcaCertificate{DER: data, SPKI: raw.TBS.PublicKey.Raw, Signature: raw.Signature.Bytes,
				Issuer: zatcaIssuerName(issuerRDN), Serial: raw.TBS.Serial,
				PublicKey: &ecdsa.PublicKey{Curve: Secp256k1(), X: x, Y: y}, NotBefore: raw.TBS.Validity.NotBefore, NotAfter: raw.TBS.Validity.NotAfter}, nil
		}
		decoded, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(string(data)), ""))
		if err != nil {
			return nil, fmt.Errorf("device signing certificate is not valid PEM, DER or Base64")
		}
		data = decoded
	}
	return nil, fmt.Errorf("device signing certificate could not be decoded")
}

func zatcaIssuerName(rdns pkix.RDNSequence) string {
	parts := make([]string, 0, len(rdns))
	for i := len(rdns) - 1; i >= 0; i-- {
		values := make([]string, 0, len(rdns[i]))
		for _, attribute := range rdns[i] {
			// Go's PKIX formatter does not recognize domainComponent, which
			// ZATCA CA issuer names use. Reuse its escaping with a known OID.
			if attribute.Type.Equal(asn1.ObjectIdentifier{0, 9, 2342, 19200300, 100, 1, 25}) {
				attribute.Type = asn1.ObjectIdentifier{2, 5, 4, 3}
				values = append(values, "DC="+strings.TrimPrefix(pkix.RDNSequence{{attribute}}.String(), "CN="))
			} else {
				values = append(values, pkix.RDNSequence{{attribute}}.String())
			}
		}
		parts = append(parts, strings.Join(values, "+"))
	}
	return strings.Join(parts, ", ")
}
