//go:build windows

package biometrics

import (
	"encoding/binary"
	"testing"
)

// buildANSI381WBFStandardBlock constructs a minimal ANSI 381:2004 record in the
// WBF standard-data-block layout: an 8-byte LE ULONG64 record length followed
// by the ANSI 381 body. "FIR\x00" magic is therefore at byte 8.
//
// Record layout (WBF, little-endian throughout):
//
//	[0:8]   record length (ULONG64 LE)
//	[8:12]  "FIR\x00" format identifier
//	[12:16] "0100" format version
//	[16:18] element count (1)
//	[18:20] scale units (1=PPI)
//	[20:22] X scan resolution
//	[22:24] Y scan resolution
//	[24:26] X image size (width)  ← record-level; view header is authoritative
//	[26:28] Y image size (height) ← record-level; view header is authoritative
//	[28]    pixel depth (8)
//	[29]    compression (0=uncompressed)
//	[30:38] reserved (8 bytes)
//
// View header at byte 38 (14 bytes, LE per WBF):
//
//	[0:4]   block length (view header + pixels)
//	[4]     finger position
//	[5]     view count
//	[6]     quality
//	[7]     impression type
//	[8:10]  horizontal line length (width in pixels)
//	[10:12] vertical line count (height in pixels)
//	[12]    view count in block
//	[13]    reserved
//
// Pixel data immediately follows (width*height bytes, ramp pattern).
func buildANSI381WBFStandardBlock(width, height, dpiX, dpiY int, compression byte) []byte {
	const (
		recHeaderSize  = 38
		viewHeaderSize = 14
	)
	pixelCount := width * height
	totalLen := recHeaderSize + viewHeaderSize + pixelCount

	buf := make([]byte, totalLen)

	// Record length (8-byte LE ULONG64)
	binary.LittleEndian.PutUint64(buf[0:8], uint64(totalLen))
	// "FIR\x00"
	copy(buf[8:12], []byte{0x46, 0x49, 0x52, 0x00})
	// "0100"
	copy(buf[12:16], []byte{0x30, 0x31, 0x30, 0x30})
	// element count = 1
	binary.LittleEndian.PutUint16(buf[16:18], 1)
	// scale units = 1 (PPI)
	binary.LittleEndian.PutUint16(buf[18:20], 1)
	// X/Y scan resolution
	binary.LittleEndian.PutUint16(buf[20:22], uint16(dpiX))
	binary.LittleEndian.PutUint16(buf[22:24], uint16(dpiY))
	// X/Y image size (record-level, informational only)
	binary.LittleEndian.PutUint16(buf[24:26], uint16(width))
	binary.LittleEndian.PutUint16(buf[26:28], uint16(height))
	// pixel depth
	buf[28] = 8
	// compression
	buf[29] = compression
	// reserved [30:38] = zero

	// View header at byte 38
	vh := buf[recHeaderSize:]
	viewBlockLen := uint32(viewHeaderSize + pixelCount)
	binary.LittleEndian.PutUint32(vh[0:4], viewBlockLen)
	vh[4] = 0  // finger position
	vh[5] = 1  // view count
	vh[6] = 60 // quality
	vh[7] = 0  // impression type
	binary.LittleEndian.PutUint16(vh[8:10], uint16(width))
	binary.LittleEndian.PutUint16(vh[10:12], uint16(height))
	vh[12] = 1 // view count in block
	vh[13] = 0 // reserved

	// Pixel ramp
	pixels := buf[recHeaderSize+viewHeaderSize:]
	for i := range pixels {
		pixels[i] = byte(i % 256)
	}
	return buf
}

// ---- Tests ----

func TestMatchesANSI381Format_WBFStandardBlock(t *testing.T) {
	payload := buildANSI381WBFStandardBlock(32, 32, 500, 500, 0)
	// "FIR\x00" + "0100" — both bytes interpreted as big-endian uint32
	if !matchesANSI381Format(payload, 8, 0x46495200, 0x30313030) {
		t.Fatalf("expected magic to match at offset 8 in WBF standard-block\nbytes[8:16] = %02x", payload[8:16])
	}
}

func TestMatchesANSI381Format_ShortPayload(t *testing.T) {
	if matchesANSI381Format([]byte{1, 2, 3}, 8, 0x46495200, 0x30313030) {
		t.Fatalf("should not match on too-short payload")
	}
}

func TestWinBioANSI381RawImageAt_Offset8_Uncompressed(t *testing.T) {
	width, height, dpiX, dpiY := 32, 32, 500, 500
	payload := buildANSI381WBFStandardBlock(width, height, dpiX, dpiY, 0)

	pixels, w, h, dpi, ok := winBioANSI381RawImageAt(payload, 8)
	if !ok {
		t.Fatalf("winBioANSI381RawImageAt(offset=8) returned ok=false for uncompressed 8-bit image")
	}
	if w != width || h != height {
		t.Errorf("dimensions: got %dx%d, want %dx%d", w, h, width, height)
	}
	if dpi != dpiX {
		t.Errorf("DPI: got %d, want %d", dpi, dpiX)
	}
	if len(pixels) != width*height {
		t.Errorf("pixel count: got %d, want %d", len(pixels), width*height)
	}
	for i, p := range pixels {
		if p != byte(i%256) {
			t.Errorf("pixel[%d] = %d, want %d", i, p, byte(i%256))
			break
		}
	}
}

func TestWinBioANSI381RawImageAt_LargeSensorSize(t *testing.T) {
	// Simulate a sensor producing 396×371 at 508 DPI (common U.are.U 4500 output).
	width, height, dpiX, dpiY := 396, 371, 508, 508
	payload := buildANSI381WBFStandardBlock(width, height, dpiX, dpiY, 0)

	pixels, w, h, dpi, ok := winBioANSI381RawImageAt(payload, 8)
	if !ok {
		t.Fatalf("winBioANSI381RawImageAt failed for large sensor (%dx%d)", width, height)
	}
	if w != width || h != height {
		t.Errorf("dimensions: got %dx%d, want %dx%d", w, h, width, height)
	}
	if dpi != dpiX {
		t.Errorf("DPI: got %d, want %d", dpi, dpiX)
	}
	if len(pixels) != width*height {
		t.Errorf("pixel count: got %d, want %d", len(pixels), width*height)
	}
}

func TestWinBioANSI381RawImageAt_SkipsCompressed(t *testing.T) {
	// compression=2 (WSQ) must be rejected
	payload := buildANSI381WBFStandardBlock(32, 32, 500, 500, 2)
	_, _, _, _, ok := winBioANSI381RawImageAt(payload, 8)
	if ok {
		t.Fatalf("expected ok=false for compressed image (compression byte=2)")
	}
}

func TestWinBioANSI381RawImageAt_EmptyPayload(t *testing.T) {
	_, _, _, _, ok := winBioANSI381RawImageAt(nil, 8)
	if ok {
		t.Fatalf("expected ok=false for nil payload")
	}
	_, _, _, _, ok = winBioANSI381RawImageAt([]byte{}, 0)
	if ok {
		t.Fatalf("expected ok=false for empty payload")
	}
}

func TestLogWinBioANSI381Metadata_DoesNotPanic(t *testing.T) {
	payload8 := buildANSI381WBFStandardBlock(32, 32, 500, 500, 0)
	logWinBioANSI381Metadata(payload8, 8)

	// Too-short payload should not panic.
	logWinBioANSI381Metadata([]byte{1, 2, 3}, 8)
	logWinBioANSI381Metadata(nil, 8)
}

func TestNormalizeWinBioANSI381_WBFLayout(t *testing.T) {
	payload := buildANSI381WBFStandardBlock(32, 32, 500, 500, 0)
	groups := normalizeWinBioANSI381(payload)
	if len(groups) == 0 {
		t.Fatalf("normalizeWinBioANSI381 returned no groups for WBF layout")
	}
	for gi, group := range groups {
		if len(group) < 2 {
			t.Errorf("group[%d] has fewer than 2 variants (LE+BE)", gi)
			continue
		}
		for vi, buf := range group {
			// After stripping the 2-byte WBF ULONG64 padding (bytes 6–7), the
			// "FIR\x00" magic shifts from offset 8 to offset 6 in the output
			// buffer. Verify the magic is present somewhere in bytes 0–10.
			if len(buf) < 12 {
				t.Errorf("group[%d] variant[%d]: buffer too short (%d bytes)", gi, vi, len(buf))
				continue
			}
			firMagic := []byte{0x46, 0x49, 0x52, 0x00}
			found := false
			for off := 0; off+4 <= len(buf) && off <= 10; off++ {
				if buf[off] == firMagic[0] && buf[off+1] == firMagic[1] &&
					buf[off+2] == firMagic[2] && buf[off+3] == firMagic[3] {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("group[%d] variant[%d]: 'FIR\\x00' not in bytes[0:10]: %02x",
					gi, vi, buf[:12])
			}
		}
	}
}

func TestWinBioANSI381Preview_WBFBlock(t *testing.T) {
	width, height := 24, 24
	payload := buildANSI381WBFStandardBlock(width, height, 500, 500, 0)
	bmpB64, w, h := winBioANSI381Preview(payload)
	if bmpB64 == "" {
		t.Fatalf("expected BMP preview, got empty string")
	}
	if w != width || h != height {
		t.Errorf("preview dimensions: got %dx%d, want %dx%d", w, h, width, height)
	}
}

func TestWinBioANSI381Preview_Compressed_NoBMP(t *testing.T) {
	payload := buildANSI381WBFStandardBlock(32, 32, 500, 500, 2)
	bmpB64, _, _ := winBioANSI381Preview(payload)
	if bmpB64 != "" {
		t.Fatalf("expected no BMP preview for compressed image, got non-empty base64")
	}
}

func TestWinBioANSI381RawImageAt_ExactUareU4500Payload(t *testing.T) {
	// Construct a synthetic 115,896-byte payload matching the exact U.are.U 4500 BIR layout
	payload := make([]byte, 115896)
	// ULONG64 LE length = 115896
	binary.LittleEndian.PutUint64(payload[0:8], 115896)
	// "FIR\x00" format identifier at byte 8
	copy(payload[8:12], []byte{0x46, 0x49, 0x52, 0x00})
	// "0100" version number at byte 12
	copy(payload[12:16], []byte{0x30, 0x31, 0x30, 0x30})
	// pixel_depth = 8 at byte 34, compression = 0 at byte 35
	payload[34] = 8
	payload[35] = 0

	pixels, w, h, dpi, ok := winBioANSI381RawImageAt(payload, 8)
	if !ok {
		t.Fatalf("expected raw image extraction to succeed for 115896 U.are.U 4500 payload")
	}
	if w != 320 || h != 362 {
		t.Errorf("expected 320x362, got %dx%d", w, h)
	}
	if len(pixels) != 320*362 {
		t.Errorf("expected %d pixels, got %d", 320*362, len(pixels))
	}
	if dpi <= 0 {
		t.Errorf("expected positive DPI, got %d", dpi)
	}
}
