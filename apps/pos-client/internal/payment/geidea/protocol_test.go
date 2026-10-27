package geidea

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestLRC_Calculation(t *testing.T) {
	// From SPAN.java for amount 500.00:
	// 20 FF 01 00 DF 04 06 00 00 00 05 00 00 DF 81 10 30 30 30 30 30 30 30 30 30 30 30 30 30 30 30 31 11
	raw500, err := hex.DecodeString("20FF0100DF0406000000050000DF81103030303030303030303030303030303111")
	if err != nil {
		t.Fatalf("hex decode failed: %v", err)
	}
	lrc500 := CalculateLRC(raw500)
	if lrc500 != 0x58 {
		t.Errorf("expected LRC 0x58, got 0x%02X", lrc500)
	}

	// For amount 1.00:
	raw1, err := hex.DecodeString("20FF0100DF0406000000000100DF81103030303030303030303030303030303111")
	if err != nil {
		t.Fatalf("hex decode failed: %v", err)
	}
	lrc1 := CalculateLRC(raw1)
	if lrc1 != 0x5C {
		t.Errorf("expected LRC 0x5C, got 0x%02X", lrc1)
	}

	// Reconciliation:
	rawRecon := []byte{0x03, 0xFF, 0x01, 0x09}
	lrcRecon := CalculateLRC(rawRecon)
	if lrcRecon != 0xF4 {
		t.Errorf("expected LRC 0xF4, got 0x%02X", lrcRecon)
	}

	// Check status:
	rawStatus := []byte{0x03, 0xFF, 0x01, 0x04}
	lrcStatus := CalculateLRC(rawStatus)
	if lrcStatus != 0xF9 {
		t.Errorf("expected LRC 0xF9, got 0x%02X", lrcStatus)
	}
}

func TestBuildPurchaseBuffer(t *testing.T) {
	buf, err := BuildPurchaseBuffer(500.00, "1", 0)
	if err != nil {
		t.Fatalf("BuildPurchaseBuffer failed: %v", err)
	}

	expectedHex := "010020ff0100df0406000000050000df8110303030303030303030303030303030311158"
	gotHex := hex.EncodeToString(buf)
	if gotHex != expectedHex {
		t.Errorf("expected %s, got %s", expectedHex, gotHex)
	}
}

func TestBuildPurchaseWithCashbackBuffer(t *testing.T) {
	// 52.00 SAR with 5.00 SAR cashback
	buf, err := BuildPurchaseBuffer(52.00, "1", 5.00)
	if err != nil {
		t.Fatalf("BuildPurchaseBuffer with cashback failed: %v", err)
	}

	expectedHex := "010029ff0100df0406000000005200df3606000000000500df81103030303030303030303030303030303111ec"
	gotHex := hex.EncodeToString(buf)
	if gotHex != expectedHex {
		t.Errorf("expected %s, got %s", expectedHex, gotHex)
	}
}

func TestBuildReconciliationBuffer(t *testing.T) {
	buf := BuildReconciliationBuffer()
	expectedHex := "010003ff0109f4"
	gotHex := hex.EncodeToString(buf)
	if gotHex != expectedHex {
		t.Errorf("expected %s, got %s", expectedHex, gotHex)
	}
}

func TestBuildCancelBuffer(t *testing.T) {
	buf := BuildCancelBuffer()
	expectedHex := "0237320306"
	gotHex := hex.EncodeToString(buf)
	if gotHex != expectedHex {
		t.Errorf("expected %s, got %s", expectedHex, gotHex)
	}
}

func TestBuildCheckStatusBuffer(t *testing.T) {
	buf := BuildCheckStatusBuffer()
	expectedHex := "010003ff0104f9"
	gotHex := hex.EncodeToString(buf)
	if gotHex != expectedHex {
		t.Errorf("expected %s, got %s", expectedHex, gotHex)
	}
}

func TestParseIntermediateStatus(t *testing.T) {
	// Status code 0x41 (Swipe or insert)
	// Frame: 01 00 02 FF 41 <LRC = 02 ^ FF ^ 41 = BE>
	frame := []byte{0x01, 0x00, 0x02, 0xFF, 0x41, 0x02 ^ 0xFF ^ 0x41}
	code, ok := ParseIntermediateStatus(frame)
	if !ok {
		t.Fatalf("expected ParseIntermediateStatus to match")
	}
	if code != 0x41 {
		t.Errorf("expected code 0x41, got 0x%02X", code)
	}

	event := GetStatusEvent(code)
	if event.MessageEn != "Please Swipe, Tap, or Insert Card" {
		t.Errorf("unexpected event message: %s", event.MessageEn)
	}
}

func TestParseTransactionResponse(t *testing.T) {
	// Clean hex response string from ECR_BUFFER_CREATIONpdf.pdf Page 4
	rawHex := strings.ReplaceAll("DF0212353838382A2A2A2A2A2A2A2A2A2A33323235DF0406000000000003DF0B06303030303330DF0C0C313430313233313835393532DF0E021012DF250C313233313835393530303031DF2606373533323235DF2703303030DF291031323334353637383132333435363738DF82025031DF2A0F363530393939303030393131202020DF3606000000000000DF81103030303030303030303030303030303108", " ", "")

	buf, err := hex.DecodeString(rawHex)
	if err != nil {
		t.Fatalf("hex decode failed: %v", err)
	}

	res, err := ParseTransactionResponse(buf)
	if err != nil {
		t.Fatalf("ParseTransactionResponse failed: %v", err)
	}

	if !res.Success {
		t.Errorf("expected approved transaction")
	}
	if res.ResponseCode != "000" {
		t.Errorf("expected response code 000, got %s", res.ResponseCode)
	}
	if res.Amount != 0.03 {
		t.Errorf("expected amount 0.03, got %f", res.Amount)
	}
	if res.CardNumber != "5888**********3225" {
		t.Errorf("expected masked PAN 5888**********3225, got %s", res.CardNumber)
	}
	if res.STAN != "000030" {
		t.Errorf("expected STAN 000030, got %s", res.STAN)
	}
	if res.RRN != "123185950001" {
		t.Errorf("expected RRN 123185950001, got %s", res.RRN)
	}
	if res.AuthCode != "753225" {
		t.Errorf("expected AuthCode 753225, got %s", res.AuthCode)
	}
	if res.CardScheme != "mada" {
		t.Errorf("expected Scheme mada, got %s", res.CardScheme)
	}
	if res.CardSchemeCode != "P1" {
		t.Errorf("expected Scheme code P1, got %s", res.CardSchemeCode)
	}
}
