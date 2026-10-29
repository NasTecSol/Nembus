package biometrics

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
)

// DigitalPersona FingerJet Engine & Capture Constants (from dpfj.h & dpfpdd.h)
const (
	// FMD Formats
	FMDFormatANSI378_2004      int32 = 0x001B0001
	FMDFormatISO19794_2_2005   int32 = 0x00010001
	FMDFormatDPPreRegFeatures  int32 = 0
	FMDFormatDPRegFeatures     int32 = 1
	FMDFormatDPVerFeatures     int32 = 2

	// FID Formats
	FIDFormatANSI381_2004    int32 = 0x001B0401
	FIDFormatISO19794_4_2005 int32 = 0x00010401

	// Probability & Dissimilarity Thresholds
	// DPFJ_PROBABILITY_ONE = 0x7fffffff (2,147,483,647)
	ProbabilityOne uint32 = 0x7fffffff

	// DefaultMaxFalseMatchScore corresponds to 0.001% FPIR (1 in 100,000 false accept rate)
	// Per U.are.U Developer Guide: PROBABILITY_ONE / 100,000 = 21,474
	DefaultMaxFalseMatchScore uint32 = 21474

	// MaxFMDSize is the maximum buffer size required by FingerJet for an FMD template
	MaxFMDSize = 2048

	// ISO19794_4HeaderSize is the 46-byte header (32-byte record header + 14-byte view header)
	// preceding raw 8-bit grayscale pixel bytes in an ISO/IEC 19794-4:2005 FID image.
	ISO19794_4HeaderSize = 46

	// FingerJet Return Codes
	DPFJSuccess        int32 = 0
	DPFJMoreData       int32 = 0x05BA000D // DPFJ_E_MORE_DATA
	DPFJInvalidFMD     int32 = 0x05BA0002
	DPFJEnrollmentInPr int32 = 0x05BA000C
)

var (
	ErrEngineUnavailable            = errors.New("biometric matching engine library is unavailable")
	ErrCaptureUnavailable           = errors.New("biometric capture hardware library is unavailable")
	ErrNoReaderConnected            = errors.New("no fingerprint reader connected")
	ErrInvalidFMD                   = errors.New("invalid or corrupted biometric FMD template")
	ErrPoorQualityScan              = errors.New("fingerprint scan quality is too low; please press finger cleanly on sensor")
	ErrFakeFingerDetected           = errors.New("spoof or fake fingerprint detected by presentation attack detection")
	ErrDuplicateBiometricOtherUser  = errors.New("this fingerprint is already enrolled to another user in your organization")
	ErrBiometricNotMatched          = errors.New("fingerprint did not match any enrolled user")
	ErrCaptureCanceled              = errors.New("fingerprint capture was canceled")
	ErrCaptureTimeout               = errors.New("fingerprint capture timed out waiting for finger")
)

// ValidateAndDecodeFMD decodes a Base64-encoded FMD and validates its header and size.
func ValidateAndDecodeFMD(fmdBase64 string) ([]byte, error) {
	if len(fmdBase64) == 0 {
		return nil, fmt.Errorf("%w: empty fmd_base64", ErrInvalidFMD)
	}
	raw, err := base64.StdEncoding.DecodeString(fmdBase64)
	if err != nil {
		// Also try RawStdEncoding or URLEncoding for WebAPI compatibility
		raw, err = base64.RawStdEncoding.DecodeString(fmdBase64)
		if err != nil {
			raw, err = base64.URLEncoding.DecodeString(fmdBase64)
			if err != nil {
				return nil, fmt.Errorf("%w: invalid base64 encoding", ErrInvalidFMD)
			}
		}
	}
	if err := ValidateFMDBytes(raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// ValidateFMDBytes verifies that raw FMD bytes have a valid length and header.
// ISO/IEC 19794-2:2005 FMDs begin with ASCII "FMR\x00" ("20" version) and are at least 24 bytes.
func ValidateFMDBytes(fmd []byte) error {
	if len(fmd) < 24 || len(fmd) > MaxFMDSize*2 {
		return fmt.Errorf("%w: unexpected FMD byte length %d", ErrInvalidFMD, len(fmd))
	}
	// ISO 19794-2:2005 FMR header magic is "FMR\x00"
	if bytes.HasPrefix(fmd, []byte{'F', 'M', 'R', 0x00}) {
		return nil
	}
	// Also allow ANSI INCITS 378-2004 ("FMR\x00" or CBEFF) or DP proprietary pre-reg/reg templates (>24 bytes)
	return nil
}

// EncodeFIDToBMPBase64 converts an ISO/IEC 19794-4:2005 FID buffer into an 8-bit grayscale BMP image
// completely in-memory (fixing the -270 out-of-bounds offset and disk race in legacy fingerprint_capture.c).
func EncodeFIDToBMPBase64(fid []byte, width, height int) (string, error) {
	if width <= 0 || height <= 0 {
		return "", errors.New("invalid image dimensions")
	}
	expectedPixels := width * height
	var pixels []byte
	if len(fid) >= ISO19794_4HeaderSize+expectedPixels {
		pixels = fid[ISO19794_4HeaderSize : ISO19794_4HeaderSize+expectedPixels]
	} else if len(fid) >= expectedPixels {
		pixels = fid[len(fid)-expectedPixels:]
	} else {
		return "", fmt.Errorf("FID buffer too short (%d bytes) for %dx%d image", len(fid), width, height)
	}

	// BMP rows must be padded to a multiple of 4 bytes
	rowStride := (width + 3) & ^3
	imageSize := rowStride * height
	const headerSize = 54
	const paletteSize = 256 * 4 // 1024 bytes for 8-bit grayscale palette
	fileSize := headerSize + paletteSize + imageSize

	buf := bytes.NewBuffer(make([]byte, 0, fileSize))

	// 14-byte BITMAPFILEHEADER
	buf.Write([]byte{'B', 'M'})
	_ = binary.Write(buf, binary.LittleEndian, uint32(fileSize))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0)) // Reserved
	_ = binary.Write(buf, binary.LittleEndian, uint32(headerSize+paletteSize))

	// 40-byte BITMAPINFOHEADER
	_ = binary.Write(buf, binary.LittleEndian, uint32(40))
	_ = binary.Write(buf, binary.LittleEndian, int32(width))
	_ = binary.Write(buf, binary.LittleEndian, int32(-height)) // Top-down DIB
	_ = binary.Write(buf, binary.LittleEndian, uint16(1))      // Planes
	_ = binary.Write(buf, binary.LittleEndian, uint16(8))      // 8-bit grayscale
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))      // BI_RGB
	_ = binary.Write(buf, binary.LittleEndian, uint32(imageSize))
	_ = binary.Write(buf, binary.LittleEndian, int32(19685)) // ~500 DPI (19685 pixels/meter)
	_ = binary.Write(buf, binary.LittleEndian, int32(19685))
	_ = binary.Write(buf, binary.LittleEndian, uint32(256))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))

	// 256-entry grayscale color table
	for i := 0; i < 256; i++ {
		c := byte(i)
		buf.Write([]byte{c, c, c, 0})
	}

	// Pixel rows with 4-byte stride padding
	padding := make([]byte, rowStride-width)
	for y := 0; y < height; y++ {
		rowStart := y * width
		buf.Write(pixels[rowStart : rowStart+width])
		if len(padding) > 0 {
			buf.Write(padding)
		}
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// FingerName returns the human-readable NIST finger name for finger_index (1..10).
func FingerName(fingerIndex int16) string {
	switch fingerIndex {
	case 1:
		return "Right Thumb"
	case 2:
		return "Right Index"
	case 3:
		return "Right Middle"
	case 4:
		return "Right Ring"
	case 5:
		return "Right Little"
	case 6:
		return "Left Thumb"
	case 7:
		return "Left Index"
	case 8:
		return "Left Middle"
	case 9:
		return "Left Ring"
	case 10:
		return "Left Little"
	default:
		return fmt.Sprintf("Finger %d", fingerIndex)
	}
}

