package usecase

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
)

// createSecp256k1CSR builds PKCS#10 directly because crypto/x509 cannot marshal
// secp256k1 public keys. The signed request uses id-ecPublicKey, the named
// secp256k1 curve, and ecdsa-with-SHA256 without algorithm parameters.
func createSecp256k1CSR(key *ecdsa.PrivateKey, subject pkix.Name, extensions []pkix.Extension) ([]byte, error) {
	if key == nil || key.Curve != Secp256k1() || key.D == nil || key.D.Sign() <= 0 || key.D.Cmp(key.Curve.Params().N) >= 0 || key.X == nil || key.Y == nil || !key.Curve.IsOnCurve(key.X, key.Y) {
		return nil, fmt.Errorf("a valid secp256k1 private key is required")
	}
	x, y := key.Curve.ScalarBaseMult(key.D.Bytes())
	if x.Cmp(key.X) != 0 || y.Cmp(key.Y) != 0 {
		return nil, fmt.Errorf("private and public keys do not match")
	}
	subjectDER, err := asn1.Marshal(subject.ToRDNSequence())
	if err != nil {
		return nil, err
	}
	curveDER, err := asn1.Marshal(oidSecp256k1)
	if err != nil {
		return nil, err
	}
	attributeDER, err := asn1.Marshal(struct {
		Type   asn1.ObjectIdentifier
		Values [][]pkix.Extension `asn1:"set"`
	}{asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 14}, [][]pkix.Extension{extensions}})
	if err != nil {
		return nil, err
	}
	publicKey := elliptic.Marshal(key.Curve, key.X, key.Y)
	type publicKeyInfo struct {
		Algorithm pkix.AlgorithmIdentifier
		PublicKey asn1.BitString
	}
	info := struct {
		Version    int
		Subject    asn1.RawValue
		PublicKey  publicKeyInfo
		Attributes []asn1.RawValue `asn1:"tag:0"`
	}{
		Subject: asn1.RawValue{FullBytes: subjectDER},
		PublicKey: publicKeyInfo{
			Algorithm: pkix.AlgorithmIdentifier{
				Algorithm:  asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1},
				Parameters: asn1.RawValue{FullBytes: curveDER},
			},
			PublicKey: asn1.BitString{Bytes: publicKey, BitLength: len(publicKey) * 8},
		},
		Attributes: []asn1.RawValue{{FullBytes: attributeDER}},
	}
	infoDER, err := asn1.Marshal(info)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(infoDER)
	signature, err := ecdsa.SignASN1(rand.Reader, key, digest[:])
	if err != nil {
		return nil, err
	}
	return asn1.Marshal(struct {
		Info      asn1.RawValue
		Algorithm pkix.AlgorithmIdentifier
		Signature asn1.BitString
	}{
		Info:      asn1.RawValue{FullBytes: infoDER},
		Algorithm: pkix.AlgorithmIdentifier{Algorithm: asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 2}},
		Signature: asn1.BitString{Bytes: signature, BitLength: len(signature) * 8},
	})
}
