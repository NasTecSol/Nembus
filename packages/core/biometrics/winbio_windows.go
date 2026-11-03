//go:build windows

package biometrics

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"
)

var (
	dllWinBio = syscall.NewLazyDLL("winbio.dll")

	procWinBioEnumBiometricUnits = dllWinBio.NewProc("WinBioEnumBiometricUnits")
	procWinBioOpenSession        = dllWinBio.NewProc("WinBioOpenSession")
	procWinBioCloseSession       = dllWinBio.NewProc("WinBioCloseSession")
	procWinBioCaptureSample      = dllWinBio.NewProc("WinBioCaptureSample")
	procWinBioCancel             = dllWinBio.NewProc("WinBioCancel")
	procWinBioFree               = dllWinBio.NewProc("WinBioFree")
)

const (
	winbioTypeFingerprint uint32 = 0x00000008
	winbioPoolSystem      uint32 = 1

	// winbio_types.h — configuration + access flags for WinBioOpenSession
	winbioFlagDefault   uint32 = 0x00000000
	winbioFlagRawAccess uint32 = 0x00000001 // WINBIO_FLAG_RAW = BIO_UNIT_RAW

	winbioPurposeNone uint32 = 0x00000000 // WINBIO_NO_PURPOSE_AVAILABLE

	// WINBIO_BIR_DATA_FLAGS from winbio_types.h. These are bit masks, not
	// HRESULT-style values; passing the wrong bits makes WinBioCaptureSample
	// reject the request with WINBIO_E_INVALID_PARAMETER.
	winbioDataFlagRaw       uint32 = 0x00000020
	winbioDataFlagProcessed uint32 = 0x00000080

	winbioSOk            uint32 = 0
	winbioEInvalidArg    uint32 = 0x80070057
	winbioEAccessDenied  uint32 = 0x80070005
	winbioETimeout       uint32 = 0x80098048
	winbioECanceled      uint32 = 0x80098049
	winbioELockViolation uint32 = 0x80098045
)

// WINBIO_DB_DEFAULT is a well-known pointer sentinel in winbio_types.h,
// not the address of a GUID object.
const winbioDBDefault uintptr = 1

type winbioVersion struct {
	Major uint32
	Minor uint32
}

type winbioUnitSchema struct {
	UnitID           uint32
	PoolType         uint32
	BiometricFactor  uint32
	SensorSubType    uint32
	Capabilities     uint32
	DeviceInstanceID [256]uint16
	Description      [256]uint16
	Manufacturer     [256]uint16
	Model            [256]uint16
	SerialNumber     [256]uint16
	FirmwareVersion  winbioVersion
}

type winbioBirData struct {
	Size   uint32
	Offset uint32
}

type winbioBir struct {
	HeaderBlock       winbioBirData
	StandardDataBlock winbioBirData
	VendorDataBlock   winbioBirData
	SignatureBlock    winbioBirData
}

// WinBioDevice represents a reader managed by Windows Biometric Framework (WBF).
type WinBioDevice struct {
	UnitID       uint32
	Model        string
	Manufacturer string
	SerialNumber string
	Description  string
	InstanceID   string
}

// IsWinBioAvailable checks if the Windows Biometric Framework library is loaded.
func IsWinBioAvailable() bool {
	return procWinBioEnumBiometricUnits.Find() == nil
}

// QueryWinBioDevices enumerates biometric units connected to Windows Biometric Framework.
func QueryWinBioDevices() ([]WinBioDevice, error) {
	if !IsWinBioAvailable() {
		return nil, nil
	}

	var schemaArray uintptr
	var unitCount uintptr

	r1, _, _ := procWinBioEnumBiometricUnits.Call(
		uintptr(winbioTypeFingerprint),
		uintptr(unsafe.Pointer(&schemaArray)),
		uintptr(unsafe.Pointer(&unitCount)),
	)

	if uint32(r1) != winbioSOk {
		return nil, fmt.Errorf("WinBioEnumBiometricUnits failed: 0x%08X", uint32(r1))
	}
	if schemaArray == 0 || unitCount == 0 {
		return nil, nil
	}
	defer procWinBioFree.Call(schemaArray)

	schemaSize := unsafe.Sizeof(winbioUnitSchema{})
	devices := make([]WinBioDevice, 0, unitCount)

	for i := uintptr(0); i < unitCount; i++ {
		ptr := unsafe.Pointer(schemaArray + (i * schemaSize))
		s := (*winbioUnitSchema)(ptr)

		dev := WinBioDevice{
			UnitID:       s.UnitID,
			Model:        utf16ToString(s.Model[:]),
			Manufacturer: utf16ToString(s.Manufacturer[:]),
			SerialNumber: utf16ToString(s.SerialNumber[:]),
			Description:  utf16ToString(s.Description[:]),
			InstanceID:   utf16ToString(s.DeviceInstanceID[:]),
		}
		devices = append(devices, dev)
	}

	return devices, nil
}

var (
	winbioMu            sync.Mutex
	activeWinBioSession uintptr
)

// CancelWinBioCapture cancels an in-progress WinBio capture.
func CancelWinBioCapture() error {
	winbioMu.Lock()
	sess := activeWinBioSession
	winbioMu.Unlock()
	if sess != 0 && procWinBioCancel != nil && procWinBioCancel.Find() == nil {
		r1, _, _ := procWinBioCancel.Call(sess)
		if uint32(r1) != winbioSOk {
			return fmt.Errorf("WinBioCancel failed: 0x%08X", uint32(r1))
		}
	}
	return nil
}

type winBioSession struct {
	handle uintptr
}

func winBioOpenSessionCall(flags uint32, databaseID uintptr, sessionHandle *uintptr) uint32 {
	r1, _, _ := procWinBioOpenSession.Call(
		uintptr(winbioTypeFingerprint),
		uintptr(winbioPoolSystem),
		uintptr(flags),
		0, // UnitArray — system pool: all units (never pass a single unit here; E_INVALIDARG)
		0, // UnitCount
		databaseID,
		uintptr(unsafe.Pointer(sessionHandle)),
	)
	return uint32(r1)
}

func openWinBioCaptureSession() (winBioSession, error) {
	var sessionHandle uintptr
	hr := winBioOpenSessionCall(winbioFlagDefault|winbioFlagRawAccess, winbioDBDefault, &sessionHandle)
	if hr == winbioSOk && sessionHandle != 0 {
		return winBioSession{handle: sessionHandle}, nil
	}
	if hr == winbioEAccessDenied {
		return winBioSession{}, fmt.Errorf("WinBioOpenSession access denied (0x80070005): run Nembus Biometric Agent as Administrator for WBF capture, or stop other apps using the sensor")
	}
	if hr == winbioEInvalidArg {
		return winBioSession{}, fmt.Errorf("WinBioOpenSession invalid parameters (0x80070057): Windows Biometric Service may be disabled or misconfigured — restart WbioSrvc and replug the reader")
	}
	return winBioSession{}, fmt.Errorf("WinBioOpenSession failed: 0x%08X", hr)
}

// CaptureWinBioSample captures a fingerprint sample via WBF and returns an ISO 19794-2 FMD.
func CaptureWinBioSample(unitID uint32, timeoutMs int) (*CaptureResponse, error) {
	if !IsWinBioAvailable() {
		return nil, fmt.Errorf("Windows Biometric Framework (winbio.dll) is not available")
	}
	if timeoutMs <= 0 {
		timeoutMs = 15000
	}

	run := func() (*CaptureResponse, error) {
		return captureWinBioSampleTimed(unitID, timeoutMs)
	}

	resp, err := withWinBioCaptureForeground(run)
	if err == nil {
		return resp, nil
	}
	if WinBioElevatedCaptureFunc != nil && isWinBioAccessDenied(err) {
		return WinBioElevatedCaptureFunc(unitID, timeoutMs)
	}
	return nil, err
}

// CaptureWinBioSampleDirect runs WBF capture without the foreground dialog (for elevated workers).
func CaptureWinBioSampleDirect(unitID uint32, timeoutMs int) (*CaptureResponse, error) {
	if !IsWinBioAvailable() {
		return nil, fmt.Errorf("Windows Biometric Framework (winbio.dll) is not available")
	}
	if timeoutMs <= 0 {
		timeoutMs = 15000
	}
	return captureWinBioSampleTimed(unitID, timeoutMs)
}

func isWinBioAccessDenied(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "0x80070005") || strings.Contains(strings.ToLower(msg), "access denied")
}

func captureWinBioSampleTimed(unitID uint32, timeoutMs int) (*CaptureResponse, error) {
	log.Printf("[WinBio] opening raw system-pool session for unit %d", unitID)
	sess, err := openWinBioCaptureSession()
	if err != nil {
		log.Printf("[WinBio] session open failed: %v", err)
		return nil, err
	}
	log.Printf("[WinBio] raw session opened for unit %d; awaiting capture", unitID)
	defer func() {
		winbioMu.Lock()
		activeWinBioSession = 0
		winbioMu.Unlock()
		procWinBioCloseSession.Call(sess.handle)
	}()

	winbioMu.Lock()
	activeWinBioSession = sess.handle
	winbioMu.Unlock()

	hardwareCaptureActive.Store(true)
	defer hardwareCaptureActive.Store(false)

	type captureOutcome struct {
		resp *CaptureResponse
		err  error
	}
	done := make(chan captureOutcome, 1)
	go func() {
		var out captureOutcome
		defer func() {
			if recovered := recover(); recovered != nil {
				out = captureOutcome{err: fmt.Errorf("WinBio sample processing panicked: %v", recovered)}
			}
			done <- out
		}()
		out.resp, out.err = winBioCaptureSampleBlocking(sess, unitID)
	}()

	timer := time.NewTimer(time.Duration(timeoutMs) * time.Millisecond)
	defer timer.Stop()

	select {
	case out := <-done:
		return out.resp, out.err
	case <-timer.C:
		log.Printf("[WinBio] %dms timer expired for unit %d; canceling WinBio request", timeoutMs, unitID)
		if cancelErr := CancelWinBioCapture(); cancelErr != nil {
			log.Printf("[WinBio] cancel returned an error: %v", cancelErr)
		}
		out := <-done
		if out.resp != nil {
			return out.resp, out.err
		}
		if out.err != nil && errors.Is(out.err, ErrCaptureCanceled) {
			return nil, ErrCaptureTimeout
		}
		return nil, ErrCaptureTimeout
	}
}

func winBioCaptureSampleBlocking(sess winBioSession, unitID uint32) (*CaptureResponse, error) {
	return winBioCaptureSampleOnce(sess.handle, unitID, winbioPurposeNone, winbioDataFlagRaw)
}

func winBioCaptureSampleOnce(sessionHandle uintptr, unitID uint32, purpose, dataFlags uint32) (*CaptureResponse, error) {
	var captureUnit uint32
	var samplePtr uintptr
	var sampleSize uintptr
	var rejectDetail uint32

	startedAt := time.Now()
	r1, _, _ := procWinBioCaptureSample.Call(
		sessionHandle,
		uintptr(purpose),
		uintptr(dataFlags),
		uintptr(unsafe.Pointer(&captureUnit)),
		uintptr(unsafe.Pointer(&samplePtr)),
		uintptr(unsafe.Pointer(&sampleSize)),
		uintptr(unsafe.Pointer(&rejectDetail)),
	)

	hresult := uint32(r1)
	log.Printf("[WinBio] WinBioCaptureSample unit=%d purpose=0x%x flags=0x%x hr=0x%08X reject=0x%x elapsed=%s", unitID, purpose, dataFlags, hresult, rejectDetail, time.Since(startedAt).Round(time.Millisecond))
	if hresult == winbioETimeout {
		return nil, ErrCaptureTimeout
	}
	if hresult == winbioECanceled {
		return nil, ErrCaptureCanceled
	}
	if hresult == winbioEAccessDenied {
		return nil, fmt.Errorf("WinBioCaptureSample access denied (0x80070005): administrator privileges required for raw WBF capture")
	}
	if hresult == winbioELockViolation {
		return nil, ErrReaderInUse
	}
	if hresult != winbioSOk {
		return nil, fmt.Errorf("WinBioCaptureSample failed: 0x%08X (reject detail: 0x%x)", hresult, rejectDetail)
	}
	if samplePtr == 0 || sampleSize == 0 {
		return nil, fmt.Errorf("WinBioCaptureSample returned empty BIR payload")
	}
	defer procWinBioFree.Call(samplePtr)
	readerUnitID := captureUnit
	if readerUnitID == 0 {
		readerUnitID = unitID
	}

	bir := (*winbioBir)(unsafe.Pointer(samplePtr))
	birBytes := unsafe.Slice((*byte)(unsafe.Pointer(samplePtr)), sampleSize)
	logWinBioBIRSummary(bir, birBytes)

	payloads := winBioBIRPayloads(bir, birBytes)
	var lastExtractErr error
	for _, payload := range payloads {
		fmdBytes, bmpBase64, width, height, err := extractFMDFromWinBioPayload(payload)
		if err == nil {
			return &CaptureResponse{
				Success:      true,
				FMDBase64:    fmdBytes,
				BMPBase64:    bmpBase64,
				QualityScore: 100,
				NFIQScore:    1,
				ReaderSerial: fmt.Sprintf("WBF-UNIT-%d", readerUnitID),
				Width:        width,
				Height:       height,
			}, nil
		}
		log.Printf("[WinBio] FingerJet conversion failed for payload size=%d: %v", len(payload), err)
		lastExtractErr = err
	}
	if lastExtractErr != nil {
		return nil, fmt.Errorf("process WinBio sample: %w", lastExtractErr)
	}
	return nil, fmt.Errorf("process WinBio sample: empty BIR blocks")
}

// logWinBioBIRSummary records only BIR layout metadata, never biometric payload bytes.
func logWinBioBIRSummary(bir *winbioBir, birBytes []byte) {
	log.Printf("[WinBio] BIR total=%d header=%d@%d standard=%d@%d vendor=%d@%d signature=%d@%d",
		len(birBytes),
		bir.HeaderBlock.Size, bir.HeaderBlock.Offset,
		bir.StandardDataBlock.Size, bir.StandardDataBlock.Offset,
		bir.VendorDataBlock.Size, bir.VendorDataBlock.Offset,
		bir.SignatureBlock.Size, bir.SignatureBlock.Offset,
	)
	if bir.HeaderBlock.Size < 48 {
		return
	}
	start := uint64(bir.HeaderBlock.Offset)
	end := start + uint64(bir.HeaderBlock.Size)
	if end < start || end > uint64(len(birBytes)) {
		return
	}
	header := birBytes[int(start):int(end)]
	if len(header) < 48 {
		return
	}
	owner := binary.LittleEndian.Uint16(header[40:42])
	formatType := binary.LittleEndian.Uint16(header[42:44])
	dataFlags := header[4]
	purpose := header[13]
	log.Printf("[WinBio] BIR format owner=0x%04x type=0x%04x data_flags=0x%02x purpose=0x%02x", owner, formatType, dataFlags, purpose)
}

func winBioBIRPayloads(bir *winbioBir, birBytes []byte) [][]byte {
	sliceBlock := func(b winbioBirData) []byte {
		if b.Size == 0 {
			return nil
		}
		start := uint64(b.Offset)
		end := start + uint64(b.Size)
		if end < start || end > uint64(len(birBytes)) {
			return nil
		}
		out := make([]byte, int(b.Size))
		copy(out, birBytes[int(start):int(end)])
		return out
	}
	var out [][]byte
	seen := map[int]struct{}{}
	add := func(b []byte) {
		if len(b) == 0 {
			return
		}
		key := len(b)<<16 | int(b[0])
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, b)
	}
	add(sliceBlock(bir.HeaderBlock))
	add(sliceBlock(bir.StandardDataBlock))
	add(sliceBlock(bir.VendorDataBlock))
	add(birBytes)
	return out
}

func extractFMDFromWinBioPayload(payload []byte) (fmdBase64 string, bmpBase64 string, width, height int, err error) {
	if len(payload) == 0 {
		return "", "", 0, 0, fmt.Errorf("empty payload")
	}

	// Log ANSI 381 metadata when the magic is found at any supported offset.
	// WBF standard-data block slices start at offset 0 of the slice; full BIR
	// buffers carry the record at offset 8 (after a CBEFF header prefix).
	for _, off := range []int{0, 8} {
		if matchesANSI381Format(payload, off, 0x46495200, 0x30313030) {
			logWinBioANSI381Metadata(payload, off)
			break
		}
	}

	engineMu.Lock()
	defer engineMu.Unlock()

	fmdBytes, isoRC, err := rawCreateFMDFromFID(FIDFormatISO19794_4_2005, payload, FMDFormatISO19794_2_2005)
	if err == nil && isoRC == DPFJSuccess && len(fmdBytes) > 0 {
		fmdB64 := base64.StdEncoding.EncodeToString(fmdBytes)
		bmpB64, _ := EncodeFIDToBMPBase64(payload, 500, 500)
		return fmdB64, bmpB64, 500, 500, nil
	}

	fmdBytes, ansiRC, err := rawCreateFMDFromFID(FIDFormatANSI381_2004, payload, FMDFormatISO19794_2_2005)
	if err == nil && ansiRC == DPFJSuccess && len(fmdBytes) > 0 {
		fmdB64 := base64.StdEncoding.EncodeToString(fmdBytes)
		bmpB64, _ := EncodeFIDToBMPBase64(payload, 500, 500)
		return fmdB64, bmpB64, 500, 500, nil
	}

	// --- Normalisation pass: strip WBF's extra 2-byte padding from ULONG64 ---
	// Try both the full-BIR offset-8 layout and the bare standard-block offset-0 layout.
	for _, normalized := range normalizeWinBioANSI381(payload) {
		for variant, buf := range normalized {
			fmdBytes, rc, normalizeErr := rawCreateFMDFromFID(FIDFormatANSI381_2004, buf, FMDFormatISO19794_2_2005)
			if normalizeErr == nil && rc == DPFJSuccess && len(fmdBytes) > 0 {
				fmdB64 := base64.StdEncoding.EncodeToString(fmdBytes)
				bmpB64, w, h := winBioANSI381Preview(payload)
				return fmdB64, bmpB64, w, h, nil
			}
			log.Printf("[WinBio] normalized ANSI381 variant=%d FingerJet rc=0x%x error=%v", variant, uint32(rc), normalizeErr)
		}
	}

	// --- Raw-pixel path: parse image geometry from the ANSI 381 view header ---
	// Tries both BIR-embedded (magic at offset 8) and bare standard-block (magic at offset 0).
	for _, off := range []int{0, 8} {
		if pixels, imageWidth, imageHeight, dpi, ok := winBioANSI381RawImageAt(payload, off); ok {
			fmdBytes, rc, rawErr := rawCreateFMDFromRaw(pixels, uint32(imageWidth), uint32(imageHeight), uint32(dpi), FMDFormatISO19794_2_2005)
			if rawErr == nil && rc == DPFJSuccess && len(fmdBytes) > 0 {
				fmdB64 := base64.StdEncoding.EncodeToString(fmdBytes)
				bmpB64, _ := EncodeFIDToBMPBase64(pixels, imageWidth, imageHeight)
				return fmdB64, bmpB64, imageWidth, imageHeight, nil
			}
			log.Printf("[WinBio] raw pixel FMD failed (magic@%d): width=%d height=%d dpi=%d rc=0x%x error=%v",
				off, imageWidth, imageHeight, dpi, uint32(rc), rawErr)
		}
	}

	// --- Last-resort: treat the tail of the payload as unformatted 8-bit grey pixels ---
	for _, dim := range [][2]int{
		{320, 362}, // DigitalPersona U.are.U 4500 native WBF frame (320x362 = 115,840 bytes, exactly leaves 56-byte header in 115,896 BIR)
		{362, 320}, // U.are.U 4500 transposed
		{300, 384}, // Alternative optical sensor frame (115,200 bytes)
		{384, 300},
		{500, 500},
		{350, 390},
		{256, 288},
		{320, 480},
	} {
		w, h := dim[0], dim[1]
		need := w * h
		if len(payload) < need {
			continue
		}
		rawPixels := payload[len(payload)-need:]
		for _, dpi := range []uint32{508, 500, 512} {
			var rawRC int32
			fmdBytes, rawRC, err = rawCreateFMDFromRaw(rawPixels, uint32(w), uint32(h), dpi, FMDFormatISO19794_2_2005)
			if err == nil && rawRC == DPFJSuccess && len(fmdBytes) > 0 {
				fmdB64 := base64.StdEncoding.EncodeToString(fmdBytes)
				bmpB64, _ := EncodeFIDToBMPBase64(rawPixels, w, h)
				return fmdB64, bmpB64, w, h, nil
			}
		}
	}

	return "", "", 0, 0, fmt.Errorf("unable to decode WinBio sample format (size=%d, ISO rc=0x%x, ANSI381 rc=0x%x)", len(payload), uint32(isoRC), uint32(ansiRC))
}

// normalizeWinBioANSI381 returns pairs of candidate ANSI 381 buffers with the
// WBF-specific 2-byte ULONG64 padding removed from the record-length field.
// The ANSI 381:2004 record length is a 6-byte big-endian integer; WBF stores
// it in an 8-byte little-endian ULONG64, so bytes 6–7 are always zero and
// must be stripped before FingerJet can parse the header.
//
// Two layout variants are tried:
//   - offset 8  — full BIR: the ANSI 381 "FIR\x00"+"0100" magic follows an
//     8-byte CBEFF-style prefix (as seen when the whole BIR buffer is passed).
//   - offset 0  — bare standard-block: WBF slices the standard data block
//     directly; the magic starts at byte 0 of the slice.
//
// For each layout, both little-endian and big-endian record-length encodings
// are returned so FingerJet can find the one it accepts.
func normalizeWinBioANSI381(payload []byte) [][][]byte {
	const (
		formatIdentifier uint32 = 0x46495200
		versionNumber    uint32 = 0x30313030
	)

	type layout struct {
		magicOffset int // byte offset of "FIR\x00" within payload
		headerSize  int // total ANSI 381 record-header size in this variant
	}
	// offset 8 → 38-byte WBF header (8-byte prefix + 30-byte ANSI body)
	// offset 0 → 36-byte standard header (6-byte length + 2-byte version + 28-byte ANSI body)
	layouts := []layout{
		{magicOffset: 8, headerSize: 38},
		{magicOffset: 0, headerSize: 36},
	}

	var result [][][]byte
	for _, l := range layouts {
		minLen := l.headerSize + 14 // at least one view-header must follow
		if len(payload) < minLen || !matchesANSI381Format(payload, l.magicOffset, formatIdentifier, versionNumber) {
			continue
		}

		// Strip the 2 padding bytes from the ULONG64 record-length field.
		// For offset-8 layout the length field is bytes 0–7 (ULONG64 LE),
		// of which bytes 6–7 are the extra WBF padding.
		// For offset-0 layout the length is already 6 bytes (bytes 0–5) so
		// only the big-/little-endian re-encoding variants are needed.
		var standard []byte
		if l.magicOffset == 8 {
			// Remove bytes 6–7 to collapse ULONG64 → 6-byte ANSI length.
			standard = make([]byte, len(payload)-2)
			copy(standard[:6], payload[:6])   // bytes 0–5: lower 6 of ULONG64
			copy(standard[6:], payload[8:])   // skip bytes 6–7 (padding)
		} else {
			// Already 36-byte header; just copy as-is for re-encoding.
			standard = make([]byte, len(payload))
			copy(standard, payload)
		}

		length := uint64(len(standard))
		leVariant := append([]byte(nil), standard...)
		for i := 0; i < 6; i++ {
			leVariant[i] = byte(length >> (8 * i))
		}
		beVariant := append([]byte(nil), standard...)
		for i := 0; i < 6; i++ {
			beVariant[i] = byte(length >> (8 * (5 - i)))
		}
		result = append(result, [][]byte{leVariant, beVariant})
	}
	return result
}

func matchesANSI381Format(payload []byte, offset int, formatIdentifier, versionNumber uint32) bool {
	if offset < 0 || len(payload) < offset+8 {
		return false
	}
	formatLE := binary.LittleEndian.Uint32(payload[offset : offset+4])
	versionLE := binary.LittleEndian.Uint32(payload[offset+4 : offset+8])
	if formatLE == formatIdentifier && versionLE == versionNumber {
		return true
	}
	formatBE := binary.BigEndian.Uint32(payload[offset : offset+4])
	versionBE := binary.BigEndian.Uint32(payload[offset+4 : offset+8])
	return formatBE == formatIdentifier && versionBE == versionNumber
}

// logWinBioANSI381Metadata logs key fields from an ANSI 381 record.
// magicOffset is where "FIR\x00" starts within payload (0 or 8).
func logWinBioANSI381Metadata(payload []byte, magicOffset int) {
	// ANSI 381:2004 record header layout (relative to start of record at magicOffset-8 for off=8, or 0 for off=0):
	//   For the WBF full-BIR variant (magicOffset=8):
	//     bytes  0– 7: record length (ULONG64 LE)
	//     bytes  8–11: "FIR\x00" format identifier
	//     bytes 12–15: "0100" version
	//     bytes 16–17: element count
	//     bytes 18–19: scale units (1=PPI, 2=PPCM)
	//     bytes 20–21: X scan resolution
	//     bytes 22–23: Y scan resolution
	//     bytes 24–25: X image size
	//     bytes 26–27: Y image size
	//     bytes 28:    pixel depth
	//     bytes 29:    compression algorithm
	//   For the bare standard-block variant (magicOffset=0):
	//     bytes  0– 5: record length (6-byte BE)
	//     bytes  6– 7: "01" version (big-endian)
	//     remainder follows same layout shifted by -2 bytes.
	base := magicOffset - 8 // byte index of the record-length field
	if base < 0 {
		base = 0
	}
	if len(payload) < base+46 {
		return
	}
	p := payload[base:]
	if len(p) < 46 {
		return
	}

	var recLen uint64
	var elemCount, scaleUnits, xRes, yRes, xSize, ySize uint16
	var pixDepth, compression byte

	if magicOffset == 8 {
		// 8-byte LE record length
		if len(p) < 8 {
			return
		}
		recLen = binary.LittleEndian.Uint64(p[0:8])
		elemCount = binary.LittleEndian.Uint16(p[16:18])
		scaleUnits = binary.LittleEndian.Uint16(p[18:20])
		xRes = binary.LittleEndian.Uint16(p[20:22])
		yRes = binary.LittleEndian.Uint16(p[22:24])
		xSize = binary.LittleEndian.Uint16(p[24:26])
		ySize = binary.LittleEndian.Uint16(p[26:28])
		pixDepth = p[28]
		compression = p[29]
	} else {
		// 6-byte BE record length
		for i := 0; i < 6; i++ {
			recLen = (recLen << 8) | uint64(p[i])
		}
		elemCount = binary.BigEndian.Uint16(p[14:16])
		scaleUnits = binary.BigEndian.Uint16(p[16:18])
		xRes = binary.BigEndian.Uint16(p[18:20])
		yRes = binary.BigEndian.Uint16(p[20:22])
		xSize = binary.BigEndian.Uint16(p[22:24])
		ySize = binary.BigEndian.Uint16(p[24:26])
		pixDepth = p[26]
		compression = p[27]
	}

	prefixLen := 48
	if len(payload) < prefixLen {
		prefixLen = len(payload)
	}
	log.Printf("[WinBio] ANSI381 header magic_off=%d total_bytes=%d rec_len=%d elements=%d scale=%d x_res=%d y_res=%d x_size=%d y_size=%d pixel_depth=%d compression=%d hex=%x",
		magicOffset, len(payload), recLen, elemCount, scaleUnits, xRes, yRes, xSize, ySize, pixDepth, compression, payload[:prefixLen])
}

// winBioANSI381RawImageAt parses raw 8-bit uncompressed pixel data from an
// ANSI 381:2004 record whose "FIR\x00" magic starts at magicOffset within payload.
// It returns false for any compressed or non-8-bit image.
func winBioANSI381RawImageAt(payload []byte, magicOffset int) (pixels []byte, width, height, dpi int, ok bool) {
	const viewHeaderSize = 14 // bytes per finger-view header in ANSI 381

	// The record-length field starts before the magic:
	//   magicOffset==8 → 8-byte LE ULONG64 at [0:8], magic at [8:16]
	//   magicOffset==0 → 6-byte BE length at [0:6], magic at [0:8] (overlap)
	base := magicOffset - 8
	if base < 0 {
		base = 0
	}
	if len(payload) < base+30+viewHeaderSize {
		return nil, 0, 0, 0, false
	}
	if !matchesANSI381Format(payload, magicOffset, 0x46495200, 0x30313030) {
		return nil, 0, 0, 0, false
	}

	p := payload[base:]

	var pixDepth, compression byte
	var scaleUnits uint16
	var xRes, yRes uint16
	var recHeaderEnd int // byte index just after the record-level header

	if magicOffset == 8 {
		if len(p) >= 36 && p[34] == 8 && p[35] == 0 {
			pixDepth = p[34]
			compression = p[35]
		} else if len(p) >= 30 {
			pixDepth = p[28]
			compression = p[29]
		}
		scaleUnits = binary.LittleEndian.Uint16(p[18:20])
		xRes = binary.LittleEndian.Uint16(p[20:22])
		yRes = binary.LittleEndian.Uint16(p[22:24])
		recHeaderEnd = 38
	} else {
		pixDepth = p[26]
		compression = p[27]
		scaleUnits = binary.BigEndian.Uint16(p[16:18])
		xRes = binary.BigEndian.Uint16(p[18:20])
		yRes = binary.BigEndian.Uint16(p[20:22])
		// Record header: 6-byte length + 2 version + 4 magic + 4 version + 2 elem +
		// 2 scale + 2 xRes + 2 yRes + 2 xSize + 2 ySize + 1 pixDepth + 1 compr = 30 bytes
		recHeaderEnd = 30
	}

	if pixDepth != 8 || compression != 0 {
		log.Printf("[WinBio] ANSI381 skipping raw-pixel path: pixel_depth=%d compression=%d (need 8/0)", pixDepth, compression)
		return nil, 0, 0, 0, false
	}

	// Read the first finger-view header that immediately follows the record header.
	vh := p[recHeaderEnd:]
	if len(vh) < viewHeaderSize {
		return nil, 0, 0, 0, false
	}

	// View header layout (14 bytes, per ANSI 381:2004 and WBF):
	//   [0:4]   finger view block length (includes these 14 bytes + pixel data)
	//   [4]     finger position
	//   [5]     view count / view number
	//   [6]     image quality
	//   [7]     impression type
	//   [8:10]  horizontal line length (width in pixels)
	//   [10:12] vertical line count (height in pixels)
	//   [12]    number of finger views in this block
	//   [13]    reserved
	var blockLen uint32
	var viewWidth, viewHeight uint16
	if magicOffset == 8 {
		// WBF stores the view header in little-endian
		blockLen = binary.LittleEndian.Uint32(vh[0:4])
		viewWidth = binary.LittleEndian.Uint16(vh[8:10])
		viewHeight = binary.LittleEndian.Uint16(vh[10:12])
	} else {
		blockLen = binary.BigEndian.Uint32(vh[0:4])
		viewWidth = binary.BigEndian.Uint16(vh[8:10])
		viewHeight = binary.BigEndian.Uint16(vh[10:12])
	}

	width = int(viewWidth)
	height = int(viewHeight)
	pixelCount := width * height

	// If direct view-header parsing didn't yield a valid dimension fitting the payload,
	// scan header bytes [36..64] for any (candW, candH) pair that fits len(payload).
	if width <= 0 || height <= 0 || pixelCount <= 0 || pixelCount > len(payload) || len(payload)-pixelCount > 256 {
		foundFit := false
		for i := 36; i <= 60 && i+4 <= len(payload); i += 2 {
			wLE := int(binary.LittleEndian.Uint16(payload[i : i+2]))
			hLE := int(binary.LittleEndian.Uint16(payload[i+2 : i+4]))
			if wLE >= 160 && wLE <= 800 && hLE >= 160 && hLE <= 800 {
				cnt := wLE * hLE
				if cnt <= len(payload) && len(payload)-cnt <= 256 {
					width, height, pixelCount = wLE, hLE, cnt
					foundFit = true
					break
				}
			}
			wBE := int(binary.BigEndian.Uint16(payload[i : i+2]))
			hBE := int(binary.BigEndian.Uint16(payload[i+2 : i+4]))
			if wBE >= 160 && wBE <= 800 && hBE >= 160 && hBE <= 800 {
				cnt := wBE * hBE
				if cnt <= len(payload) && len(payload)-cnt <= 256 {
					width, height, pixelCount = wBE, hBE, cnt
					foundFit = true
					break
				}
			}
		}
		// Known DigitalPersona U.are.U 4500 optical sensor frame (115,840 bytes)
		if !foundFit && len(payload) >= 115840 && len(payload) <= 116000 {
			width, height = 320, 362
			pixelCount = width * height
			foundFit = true
		}
		if !foundFit {
			return nil, 0, 0, 0, false
		}
	}
	_ = blockLen

	if pixelCount > len(payload) {
		return nil, 0, 0, 0, false
	}
	pixels = payload[len(payload)-pixelCount:]

	// Derive DPI from resolution fields
	dpiX := int(xRes)
	dpiY := int(yRes)
	if scaleUnits == 2 { // pixels per centimetre → convert to PPI
		dpiX = int(float64(dpiX)*2.54 + 0.5)
		dpiY = int(float64(dpiY)*2.54 + 0.5)
	}
	dpi = (dpiX + dpiY) / 2
	if dpi <= 0 {
		dpi = 500
	}

	log.Printf("[WinBio] ANSI381 raw image extracted: magic_off=%d width=%d height=%d dpi=%d pixels=%d", magicOffset, width, height, dpi, pixelCount)
	return pixels, width, height, dpi, true
}

// winBioANSI381RawImage is the legacy entry point that searches only at the
// full-BIR offset (magic at byte 8). Kept for backward compatibility.
func winBioANSI381RawImage(payload []byte) (pixels []byte, width, height, dpi int, ok bool) {
	return winBioANSI381RawImageAt(payload, 8)
}

// winBioANSI381Preview extracts an uncompressed 8-bit image when the WBF
// ANSI381 block contains one. Compressed images still produce an FMD, but no
// BMP preview is returned.
func winBioANSI381Preview(payload []byte) (string, int, int) {
	for _, off := range []int{0, 8} {
		pixels, width, height, _, ok := winBioANSI381RawImageAt(payload, off)
		if ok {
			bmp, _ := EncodeFIDToBMPBase64(pixels, width, height)
			return bmp, width, height
		}
	}
	return "", 0, 0
}

func utf16ToString(s []uint16) string {
	n := 0
	for n < len(s) && s[n] != 0 {
		n++
	}
	runes := utf16.Decode(s[:n])
	var buf bytes.Buffer
	for _, r := range runes {
		buf.WriteRune(r)
	}
	return buf.String()
}
