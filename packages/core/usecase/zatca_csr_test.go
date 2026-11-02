package usecase

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"testing"
)

func TestZatcaCSRSignatureAndExtensions(t *testing.T) {
	for _, sandbox := range []bool{true, false} {
		templateName := "ZATCA-Code-Signing"
		if sandbox {
			templateName = "PREZATCA-Code-Signing"
		}
		t.Run(templateName, func(t *testing.T) {
			key, err := GenerateECDSAKeyPair()
			if err != nil {
				t.Fatal(err)
			}
			params := ZatcaCSRParams{
				CommonName: "POS-01", SerialNumber: "1-Nembus|2-POS|3-test-device",
				OrganizationIdentifier: "310123456700003", OrgName: "شركة الاختبار",
				OrgUnit: "Riyadh Branch", CountryCode: "SA", InvoiceType: "1100",
				Location: "Riyadh", Industry: "Retail", IsSandbox: sandbox,
			}
			csrPEM, err := GenerateCSR(key, params)
			if err != nil {
				t.Fatal(err)
			}
			block, rest := pem.Decode([]byte(csrPEM))
			if block == nil || block.Type != "CERTIFICATE REQUEST" || len(rest) != 0 {
				t.Fatal("invalid CSR PEM")
			}
			var csr struct {
				Info      asn1.RawValue
				Algorithm pkix.AlgorithmIdentifier
				Signature asn1.BitString
			}
			if rest, err := asn1.Unmarshal(block.Bytes, &csr); err != nil || len(rest) != 0 {
				t.Fatalf("invalid CSR DER: %v", err)
			}
			if !csr.Algorithm.Algorithm.Equal(asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 2}) || len(csr.Algorithm.Parameters.FullBytes) != 0 {
				t.Fatal("expected ecdsa-with-SHA256 without parameters")
			}
			digest := sha256.Sum256(csr.Info.FullBytes)
			if !ecdsa.VerifyASN1(&key.PublicKey, digest[:], csr.Signature.Bytes) {
				t.Fatal("CSR signature verification failed")
			}
			var info struct {
				Version int
				Subject pkix.RDNSequence
				Key     struct {
					Algorithm pkix.AlgorithmIdentifier
					PublicKey asn1.BitString
				}
				Attributes []struct {
					Type   asn1.ObjectIdentifier
					Values [][]pkix.Extension `asn1:"set"`
				} `asn1:"tag:0"`
			}
			if _, err := asn1.Unmarshal(csr.Info.FullBytes, &info); err != nil {
				t.Fatal(err)
			}
			var subject pkix.Name
			subject.FillFromRDNSequence(&info.Subject)
			if info.Version != 0 || subject.CommonName != params.CommonName || len(subject.Organization) != 1 || subject.Organization[0] != params.OrgName {
				t.Fatal("CSR subject or version mismatch")
			}
			var curve asn1.ObjectIdentifier
			if _, err := asn1.Unmarshal(info.Key.Algorithm.Parameters.FullBytes, &curve); err != nil || !curve.Equal(oidSecp256k1) || !info.Key.Algorithm.Algorithm.Equal(asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}) {
				t.Fatal("CSR does not identify a secp256k1 EC key")
			}
			if !bytes.Equal(info.Key.PublicKey.Bytes, elliptic.Marshal(key.Curve, key.X, key.Y)) {
				t.Fatal("CSR public key mismatch")
			}
			if len(info.Attributes) != 1 || !info.Attributes[0].Type.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 14}) || len(info.Attributes[0].Values) != 1 || len(info.Attributes[0].Values[0]) != 2 {
				t.Fatal("invalid PKCS#9 extension request")
			}
			extensions := info.Attributes[0].Values[0]
			var template asn1.RawValue
			if _, err := asn1.Unmarshal(extensions[0].Value, &template); err != nil || template.Tag != asn1.TagPrintableString || string(template.Bytes) != templateName || !extensions[0].Id.Equal(asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 20, 2}) {
				t.Fatal("invalid certificate template extension")
			}
			var names []asn1.RawValue
			if _, err := asn1.Unmarshal(extensions[1].Value, &names); err != nil || len(names) != 1 || names[0].Class != 2 || names[0].Tag != 4 {
				t.Fatal("SAN must contain a DER directoryName")
			}
			var rdns pkix.RDNSequence
			if _, err := asn1.Unmarshal(names[0].Bytes, &rdns); err != nil {
				t.Fatal(err)
			}
			got := map[string]interface{}{}
			for _, set := range rdns {
				for _, attribute := range set {
					got[attribute.Type.String()] = attribute.Value
				}
			}
			for oid, want := range map[string]string{
				"2.5.4.4": params.SerialNumber, "0.9.2342.19200300.100.1.1": params.OrganizationIdentifier,
				"2.5.4.12": params.InvoiceType, "2.5.4.26": params.Location, "2.5.4.15": params.Industry,
			} {
				if got[oid] != want {
					t.Errorf("SAN %s = %v; want %s", oid, got[oid], want)
				}
			}
		})
	}
}

func TestZatcaCSRRejectsInvalidKeys(t *testing.T) {
	p256, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []*ecdsa.PrivateKey{nil, p256} {
		if _, err := GenerateCSR(key, ZatcaCSRParams{}); err == nil {
			t.Fatal("expected rejection of a missing or non-secp256k1 key")
		}
	}
}
