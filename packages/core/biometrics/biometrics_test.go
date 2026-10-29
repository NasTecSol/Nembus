package biometrics

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestInitRuntimeAndBMPConversion(t *testing.T) {
	if err := InitRuntime(); err != nil {
		t.Fatalf("InitRuntime failed: %v", err)
	}
	if !IsEngineAvailable() {
		t.Logf("Warning: FingerJet engine not available on this host")
	}

	// Verify in-memory ISO 19794-4 FID -> 8-bit grayscale BMP conversion (46-byte header + pixels)
	width, height := 16, 16
	fid := make([]byte, ISO19794_4HeaderSize+(width*height))
	copy(fid[:4], []byte{'F', 'I', 'R', 0x00})
	for i := 0; i < width*height; i++ {
		fid[ISO19794_4HeaderSize+i] = byte(i % 256)
	}

	bmpB64, err := EncodeFIDToBMPBase64(fid, width, height)
	if err != nil {
		t.Fatalf("EncodeFIDToBMPBase64 failed: %v", err)
	}
	rawBMP, err := base64.StdEncoding.DecodeString(bmpB64)
	if err != nil {
		t.Fatalf("Decode BMP base64 failed: %v", err)
	}
	if len(rawBMP) < 54+1024+(width*height) || string(rawBMP[:2]) != "BM" {
		t.Fatalf("unexpected BMP output header or length: %d", len(rawBMP))
	}
}

func TestValidateAndDecodeFMD(t *testing.T) {
	validISO := append([]byte{'F', 'M', 'R', 0x00, ' ', '2', '0', 0x00}, make([]byte, 40)...)
	b64 := base64.StdEncoding.EncodeToString(validISO)

	decoded, err := ValidateAndDecodeFMD(b64)
	if err != nil {
		t.Fatalf("expected valid FMD, got error: %v", err)
	}
	if len(decoded) != len(validISO) {
		t.Fatalf("expected length %d, got %d", len(validISO), len(decoded))
	}

	if _, err := ValidateAndDecodeFMD(strings.Repeat("A", 8)); err == nil {
		t.Fatalf("expected error for too-short FMD")
	}
}

