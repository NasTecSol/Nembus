package printing

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

func TestReceiptQRPreservesSavedBase64AndUsesTwoByteLength(t *testing.T) {
	// Longer than 255 bytes, like the nine-tag Phase 2 payload.
	payload := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("signed-invoice-data", 32)))
	output := BuildReceipt(ReceiptTemplate{}, ReceiptData{ReceiptNumber: "TXN-6", QRCodeBase64: payload})
	start := bytes.Index(output, []byte{GS, '(', 'k', byte((len(payload) + 3) % 256), byte((len(payload) + 3) / 256), 49, 80, 48})
	if start < 0 {
		t.Fatal("QR store command or length is missing")
	}
	dataStart := start + 8
	if !bytes.Equal(output[dataStart:dataStart+len(payload)], []byte(payload)) {
		t.Fatal("QR must encode the exact Base64 text, not decoded TLV or receipt number")
	}
	if !bytes.Contains(output[dataStart+len(payload):], []byte{GS, '(', 'k', 3, 0, 49, 81, 48}) {
		t.Fatal("QR print command missing")
	}
	if bytes.Count(output, []byte(payload)) != 1 {
		t.Fatal("QR printed multiple times")
	}
	if !bytes.Contains(output, CmdCut()) {
		t.Fatal("receipt cut missing")
	}
}

func TestReceiptDoesNotInventQRForLegacySalesOrSessionReports(t *testing.T) {
	for _, data := range []ReceiptData{
		{ReceiptNumber: "legacy"},
		{Type: "Z-REPORT", QRCodeBase64: "must-not-print-on-session-summary"},
	} {
		if bytes.Contains(BuildReceipt(ReceiptTemplate{}, data), []byte{GS, '(', 'k'}) {
			t.Fatal("QR emitted without a signed sale payload")
		}
	}
}
