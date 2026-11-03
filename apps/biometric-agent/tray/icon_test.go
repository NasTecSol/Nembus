package tray

import (
	"bytes"
	"image/png"
	"testing"
)

func TestGenerateBiometricIcon(t *testing.T) {
	states := []AgentState{StateReady, StateCapturing, StateDisconnected, StateError}

	for _, state := range states {
		icoData := GenerateBiometricIcon(state, 32)
		if len(icoData) < 22 {
			t.Fatalf("state %d: icon data too short (%d bytes)", state, len(icoData))
		}

		// Verify ICO header (Type = 1 for icon)
		if icoData[2] != 1 || icoData[3] != 0 {
			t.Errorf("state %d: invalid ICO header type", state)
		}

		// Verify embedded PNG data can be decoded
		pngBytes := icoData[22:]
		img, err := png.Decode(bytes.NewReader(pngBytes))
		if err != nil {
			t.Fatalf("state %d: failed to decode embedded PNG: %v", state, err)
		}

		if img.Bounds().Dx() != 32 || img.Bounds().Dy() != 32 {
			t.Errorf("state %d: unexpected dimensions %dx%d, expected 32x32", state, img.Bounds().Dx(), img.Bounds().Dy())
		}
	}
}

func TestGenerateBiometricPNG(t *testing.T) {
	pngData := GenerateBiometricPNG(StateReady, 64)
	if len(pngData) == 0 {
		t.Fatal("expected non-empty PNG data")
	}
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		t.Fatalf("failed to decode PNG: %v", err)
	}
	if img.Bounds().Dx() != 64 || img.Bounds().Dy() != 64 {
		t.Errorf("unexpected dimensions %dx%d, expected 64x64", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

