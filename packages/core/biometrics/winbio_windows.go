//go:build windows

package biometrics

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
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
	procWinBioLocateSensor       = dllWinBio.NewProc("WinBioLocateSensor")
	procWinBioCancel             = dllWinBio.NewProc("WinBioCancel")
	procWinBioFree               = dllWinBio.NewProc("WinBioFree")
)

const (
	winbioTypeFingerprint uint32 = 0x00000008
	winbioPoolSystem      uint32 = 1

	// winbio_types.h — configuration + access flags for WinBioOpenSession
	winbioFlagDefault     uint32 = 0x00000000
	winbioFlagRawAccess   uint32 = 0x00000008 // WINBIO_FLAG_RAW

	winbioPurposeVerify uint32 = 0x00000001
	winbioPurposeEnroll uint32 = 0x00000002

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

// winbioGUID matches Windows GUID layout for WinBioOpenSession DatabaseId.
type winbioGUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// WINBIO_DB_DEFAULT from winbio_types.h — required for WINBIO_POOL_SYSTEM (NULL is E_INVALIDARG on many builds).
var winbioDBDefault = winbioGUID{
	Data1: 0x08C81374,
	Data2: 0x941A,
	Data3: 0x4E1D,
	Data4: [8]byte{0x9F, 0x19, 0x9F, 0x81, 0xBD, 0x0A, 0x0F, 0xE2},
}

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
	handle    uintptr
	rawAccess bool
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

func openWinBioCaptureSession(unitID uint32) (winBioSession, error) {
	_ = unitID // unit targeting happens in WinBioLocateSensor / WinBioCaptureSample

	dbDefault := uintptr(unsafe.Pointer(&winbioDBDefault))
	type attempt struct {
		flags     uint32
		rawAccess bool
		database  uintptr
	}
	attempts := []attempt{
		{winbioFlagDefault, false, dbDefault},
		{winbioFlagDefault | winbioFlagRawAccess, true, dbDefault},
		{winbioFlagDefault, false, 0},
		{winbioFlagDefault | winbioFlagRawAccess, true, 0},
	}

	var lastHR uint32
	for _, att := range attempts {
		var sessionHandle uintptr
		hr := winBioOpenSessionCall(att.flags, att.database, &sessionHandle)
		lastHR = hr
		if hr == winbioSOk && sessionHandle != 0 {
			return winBioSession{handle: sessionHandle, rawAccess: att.rawAccess}, nil
		}
		if hr == winbioEAccessDenied {
			continue
		}
	}

	if lastHR == winbioEAccessDenied {
		return winBioSession{}, fmt.Errorf("WinBioOpenSession access denied (0x80070005): run Nembus Biometric Agent as Administrator for WBF capture, or stop other apps using the sensor")
	}
	if lastHR == winbioEInvalidArg {
		return winBioSession{}, fmt.Errorf("WinBioOpenSession invalid parameters (0x80070057): Windows Biometric Service may be disabled or misconfigured — restart WbioSrvc and replug the reader")
	}
	return winBioSession{}, fmt.Errorf("WinBioOpenSession failed: 0x%08X", lastHR)
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
	sess, err := openWinBioCaptureSession(unitID)
	if err != nil {
		return nil, err
	}
	defer func() {
		winbioMu.Lock()
		activeWinBioSession = 0
		winbioMu.Unlock()
		procWinBioCloseSession.Call(sess.handle)
	}()

	winbioMu.Lock()
	activeWinBioSession = sess.handle
	winbioMu.Unlock()

	if procWinBioLocateSensor.Find() == nil {
		_, _, _ = procWinBioLocateSensor.Call(sess.handle, uintptr(unitID))
	}

	hardwareCaptureActive.Store(true)
	defer hardwareCaptureActive.Store(false)

	type captureOutcome struct {
		resp *CaptureResponse
		err  error
	}
	done := make(chan captureOutcome, 1)
	go func() {
		resp, err := winBioCaptureSampleBlocking(sess, unitID)
		done <- captureOutcome{resp: resp, err: err}
	}()

	timer := time.NewTimer(time.Duration(timeoutMs) * time.Millisecond)
	defer timer.Stop()

	select {
	case out := <-done:
		return out.resp, out.err
	case <-timer.C:
		_ = CancelWinBioCapture()
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
	type attempt struct {
		purpose   uint32
		dataFlags uint32
	}
	attempts := []attempt{
		{winbioPurposeVerify, winbioDataFlagRaw},
		{winbioPurposeEnroll, winbioDataFlagRaw},
	}
	if !sess.rawAccess {
		attempts = []attempt{
			{winbioPurposeVerify, winbioDataFlagProcessed},
			{winbioPurposeEnroll, winbioDataFlagProcessed},
		}
	}

	var lastErr error
	for _, att := range attempts {
		resp, err := winBioCaptureSampleOnce(sess.handle, unitID, att.purpose, att.dataFlags)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if errors.Is(err, ErrCaptureTimeout) || errors.Is(err, ErrCaptureCanceled) {
			return nil, err
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("WinBioCaptureSample failed for unit %d", unitID)
}

func winBioCaptureSampleOnce(sessionHandle uintptr, unitID uint32, purpose, dataFlags uint32) (*CaptureResponse, error) {
	captureUnit := unitID
	var samplePtr uintptr
	var sampleSize uintptr
	var rejectDetail uint32

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

	bir := (*winbioBir)(unsafe.Pointer(samplePtr))
	birBytes := unsafe.Slice((*byte)(unsafe.Pointer(samplePtr)), sampleSize)

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
				ReaderSerial: fmt.Sprintf("WBF-UNIT-%d", unitID),
				Width:        width,
				Height:       height,
			}, nil
		}
		lastExtractErr = err
	}
	if lastExtractErr != nil {
		return nil, fmt.Errorf("process WinBio sample: %w", lastExtractErr)
	}
	return nil, fmt.Errorf("process WinBio sample: empty BIR blocks")
}

func winBioBIRPayloads(bir *winbioBir, birBytes []byte) [][]byte {
	sliceBlock := func(b winbioBirData) []byte {
		if b.Size == 0 {
			return nil
		}
		end := int(b.Offset + b.Size)
		if end > len(birBytes) || int(b.Offset) < 0 {
			return nil
		}
		out := make([]byte, b.Size)
		copy(out, birBytes[b.Offset:end])
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

	engineMu.Lock()
	defer engineMu.Unlock()

	fmdBytes, rc, err := rawCreateFMDFromFID(FIDFormatISO19794_4_2005, payload, FMDFormatISO19794_2_2005)
	if err == nil && rc == DPFJSuccess && len(fmdBytes) > 0 {
		fmdB64 := base64.StdEncoding.EncodeToString(fmdBytes)
		bmpB64, _ := EncodeFIDToBMPBase64(payload, 500, 500)
		return fmdB64, bmpB64, 500, 500, nil
	}

	fmdBytes, rc, err = rawCreateFMDFromFID(FIDFormatANSI381_2004, payload, FMDFormatISO19794_2_2005)
	if err == nil && rc == DPFJSuccess && len(fmdBytes) > 0 {
		fmdB64 := base64.StdEncoding.EncodeToString(fmdBytes)
		bmpB64, _ := EncodeFIDToBMPBase64(payload, 500, 500)
		return fmdB64, bmpB64, 500, 500, nil
	}

	for _, dim := range [][2]int{{500, 500}, {350, 390}, {256, 288}, {320, 480}} {
		w, h := dim[0], dim[1]
		need := w * h
		if len(payload) < need {
			continue
		}
		rawPixels := payload[len(payload)-need:]
		fmdBytes, rc, err = rawCreateFMDFromRaw(rawPixels, uint32(w), uint32(h), 500, FMDFormatISO19794_2_2005)
		if err == nil && rc == DPFJSuccess && len(fmdBytes) > 0 {
			fmdB64 := base64.StdEncoding.EncodeToString(fmdBytes)
			bmpB64, _ := EncodeFIDToBMPBase64(rawPixels, w, h)
			return fmdB64, bmpB64, w, h, nil
		}
	}

	return "", "", 0, 0, fmt.Errorf("unable to decode WinBio sample format (size=%d)", len(payload))
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
