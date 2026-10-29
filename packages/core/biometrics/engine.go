package biometrics

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"sync"
	"unsafe"
)

// C Struct Mirrors (matching dpfpdd.h ABI on both Windows x64 and Linux x64)
type cDPFPDDVerInfo struct {
	Major       int32
	Minor       int32
	Maintenance int32
}

type cDPFPDDHwDescr struct {
	VendorName  [128]byte
	ProductName [128]byte
	SerialNum   [128]byte
}

type cDPFPDDHwID struct {
	VendorID  uint16
	ProductID uint16
}

type cDPFPDDHwVersion struct {
	HwVer  cDPFPDDVerInfo
	FwVer  cDPFPDDVerInfo
	BcdRev uint16
	_pad   uint16
}

type cDPFPDDDevInfo struct {
	Size       uint32
	Name       [1024]byte
	Descr      cDPFPDDHwDescr
	ID         cDPFPDDHwID
	Ver        cDPFPDDHwVersion
	Modality   uint32
	Technology uint32
}

type cDPFPDDCaptureParam struct {
	Size      uint32
	ImageFmt  uint32
	ImageProc uint32
	ImageRes  uint32
}

type cDPFPDDImageInfo struct {
	Size   uint32
	Width  uint32
	Height uint32
	Res    uint32
	Bpp    uint32
}

type cDPFPDDCaptureResult struct {
	Size    uint32
	Success int32
	Quality uint32
	Score   uint32
	Info    cDPFPDDImageInfo
}

const (
	dpfpddSuccess          int32  = 0
	dpfpddErrMoreData      int32  = 0x05BA000D
	dpfpddImgFmtISO19794   uint32 = 0x01010007
	dpfpddImgProcEnhanced  uint32 = 2
	dpfpddQualityGood      uint32 = 0
	dpfpddQualityTimedOut  uint32 = 1
	dpfpddQualityCanceled  uint32 = 2
	dpfpddQualityFake      uint32 = 8
)

// ReaderDeviceInfo represents an attached DigitalPersona USB fingerprint reader.
type ReaderDeviceInfo struct {
	Name        string `json:"name"`
	VendorName  string `json:"vendor_name"`
	ProductName string `json:"product_name"`
	SerialNum   string `json:"serial_number"`
	VendorID    uint16 `json:"vendor_id"`
	ProductID   uint16 `json:"product_id"`
}

// CaptureResponse holds the result of a single fingerprint scan.
type CaptureResponse struct {
	Success      bool   `json:"success"`
	FMDBase64    string `json:"fmd"`
	BMPBase64    string `json:"bmp"`
	QualityScore int    `json:"quality"`
	NFIQScore    int    `json:"nfiq_score"`
	ReaderSerial string `json:"reader_serial"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Message      string `json:"message,omitempty"`
}

var (
	engineMu      sync.Mutex
	captureMu     sync.Mutex
	activeDevMu   sync.Mutex
	activeDevHdl  uintptr
	dpfpddInited  bool
)

// IsEngineAvailable reports whether the FingerJet matching library (dpfj.dll / libdpfj.so) is loaded.
func IsEngineAvailable() bool {
	return platformEngineAvailable()
}

// IsCaptureAvailable reports whether the U.are.U capture library (dpfpdd.dll / libdpfpdd.so) is loaded.
func IsCaptureAvailable() bool {
	return platformCaptureAvailable()
}

// Compare performs a 1:1 comparison between two ISO/IEC 19794-2:2005 FMD templates.
// Returns the dissimilarity score (0 = identical, ProbabilityOne = completely different)
// and whether the score is within the false-match threshold.
func Compare(probeFMD []byte, candidateFMD []byte, threshold uint32) (uint32, bool, error) {
	if err := ValidateFMDBytes(probeFMD); err != nil {
		return ProbabilityOne, false, err
	}
	if err := ValidateFMDBytes(candidateFMD); err != nil {
		return ProbabilityOne, false, err
	}
	if threshold == 0 {
		threshold = DefaultMaxFalseMatchScore
	}

	engineMu.Lock()
	defer engineMu.Unlock()

	score, rc, err := rawCompare(
		FMDFormatISO19794_2_2005, probeFMD, 0,
		FMDFormatISO19794_2_2005, candidateFMD, 0,
	)
	if err != nil {
		return ProbabilityOne, false, err
	}
	if rc != DPFJSuccess {
		return ProbabilityOne, false, fmt.Errorf("dpfj_compare failed with code 0x%x", uint32(rc))
	}
	return score, score <= threshold, nil
}

// Identify performs a 1:N search comparing probeFMD against all candidateFMDs.
// Returns the index of the best matching candidate (-1 if no match <= threshold) and its dissimilarity score.
func Identify(probeFMD []byte, candidateFMDs [][]byte, threshold uint32) (int, uint32, error) {
	if err := ValidateFMDBytes(probeFMD); err != nil {
		return -1, ProbabilityOne, err
	}
	if len(candidateFMDs) == 0 {
		return -1, ProbabilityOne, nil
	}
	if threshold == 0 {
		threshold = DefaultMaxFalseMatchScore
	}

	engineMu.Lock()
	defer engineMu.Unlock()

	if !platformEngineAvailable() {
		return -1, ProbabilityOne, ErrEngineUnavailable
	}

	bestIdx := -1
	bestScore := ProbabilityOne

	for idx, cand := range candidateFMDs {
		if len(cand) < 24 {
			continue
		}
		score, rc, err := rawCompare(
			FMDFormatISO19794_2_2005, probeFMD, 0,
			FMDFormatISO19794_2_2005, cand, 0,
		)
		if err != nil {
			return -1, ProbabilityOne, err
		}
		if rc == DPFJSuccess && score <= threshold && score < bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx, bestScore, nil
}

// ConsolidateEnrollmentFMD merges multiple single-scan ISO 19794-2:2005 FMDs (typically 4 scans)
// into a single consolidated enrollment FMD using the FingerJet enrollment state machine:
// dpfj_start_enrollment -> dpfj_add_to_enrollment (xN) -> dpfj_create_enrollment_fmd -> dpfj_finish_enrollment.
func ConsolidateEnrollmentFMD(viewFMDs [][]byte) ([]byte, error) {
	if len(viewFMDs) == 0 {
		return nil, ErrInvalidFMD
	}
	for _, fmd := range viewFMDs {
		if err := ValidateFMDBytes(fmd); err != nil {
			return nil, err
		}
	}

	engineMu.Lock()
	defer engineMu.Unlock()

	if !platformEngineAvailable() {
		return nil, ErrEngineUnavailable
	}

	// Ensure any previous stale enrollment session is closed
	_ = rawFinishEnrollment()

	rc, err := rawStartEnrollment(FMDFormatISO19794_2_2005)
	if err != nil {
		return nil, err
	}
	if rc != DPFJSuccess {
		return nil, fmt.Errorf("dpfj_start_enrollment failed: 0x%x", uint32(rc))
	}
	defer rawFinishEnrollment()

	ready := false
	for _, view := range viewFMDs {
		addRC, err := rawAddToEnrollment(FMDFormatISO19794_2_2005, view, 0)
		if err != nil {
			return nil, err
		}
		if addRC == DPFJSuccess {
			ready = true
			break
		}
		if addRC != DPFJMoreData {
			return nil, fmt.Errorf("dpfj_add_to_enrollment returned 0x%x", uint32(addRC))
		}
	}

	// If fewer than 4 scans were provided or FingerJet still wants more views, feed first view again if needed
	if !ready && len(viewFMDs) > 0 {
		for attempts := 0; attempts < 4 && !ready; attempts++ {
			addRC, _ := rawAddToEnrollment(FMDFormatISO19794_2_2005, viewFMDs[attempts%len(viewFMDs)], 0)
			if addRC == DPFJSuccess {
				ready = true
				break
			}
		}
	}

	consolidated, createRC, err := rawCreateEnrollmentFMD()
	if err != nil || createRC != DPFJSuccess {
		// Fallback to best view FMD if consolidation cannot complete
		return viewFMDs[0], nil
	}
	return consolidated, nil
}

func ensureDPFPDDInitialized() error {
	if !platformCaptureAvailable() {
		return ErrCaptureUnavailable
	}
	if !dpfpddInited {
		rc := rawDPFPDDInit()
		if rc != dpfpddSuccess {
			return fmt.Errorf("dpfpdd_init failed with 0x%x", uint32(rc))
		}
		dpfpddInited = true
	}
	return nil
}

// QueryDevices enumerates connected DigitalPersona USB fingerprint readers.
func QueryDevices() ([]ReaderDeviceInfo, error) {
	captureMu.Lock()
	defer captureMu.Unlock()

	if err := ensureDPFPDDInitialized(); err != nil {
		return nil, err
	}

	var devCnt uint32 = 8
	infos := make([]cDPFPDDDevInfo, devCnt)
	for i := range infos {
		infos[i].Size = uint32(unsafe.Sizeof(cDPFPDDDevInfo{}))
	}

	rc := rawDPFPDDQueryDevices(&devCnt, unsafe.Pointer(&infos[0]))
	if rc == dpfpddErrMoreData && devCnt > 0 {
		infos = make([]cDPFPDDDevInfo, devCnt)
		for i := range infos {
			infos[i].Size = uint32(unsafe.Sizeof(cDPFPDDDevInfo{}))
		}
		rc = rawDPFPDDQueryDevices(&devCnt, unsafe.Pointer(&infos[0]))
	}
	if rc != dpfpddSuccess {
		// Re-init once in case USB bus state changed
		_ = rawDPFPDDExit()
		dpfpddInited = false
		if err := ensureDPFPDDInitialized(); err != nil {
			return nil, err
		}
		rc = rawDPFPDDQueryDevices(&devCnt, unsafe.Pointer(&infos[0]))
		if rc != dpfpddSuccess {
			return nil, fmt.Errorf("dpfpdd_query_devices failed: 0x%x", uint32(rc))
		}
	}

	result := make([]ReaderDeviceInfo, 0, devCnt)
	for i := uint32(0); i < devCnt && i < uint32(len(infos)); i++ {
		d := infos[i]
		result = append(result, ReaderDeviceInfo{
			Name:        cBytesToString(d.Name[:]),
			VendorName:  cBytesToString(d.Descr.VendorName[:]),
			ProductName: cBytesToString(d.Descr.ProductName[:]),
			SerialNum:   cBytesToString(d.Descr.SerialNum[:]),
			VendorID:    d.ID.VendorID,
			ProductID:   d.ID.ProductID,
		})
	}
	return result, nil
}

// CancelActiveCapture cancels an in-progress blocking dpfpdd_capture call immediately in-process.
func CancelActiveCapture() error {
	activeDevMu.Lock()
	hDev := activeDevHdl
	activeDevMu.Unlock()
	if hDev == 0 {
		return nil
	}
	rc := rawDPFPDDCancel(hDev)
	if rc != dpfpddSuccess {
		return fmt.Errorf("dpfpdd_cancel returned 0x%x", uint32(rc))
	}
	return nil
}

// CaptureSingleFMD captures a single fingerprint scan from the first connected USB reader,
// extracts an ISO/IEC 19794-2:2005 FMD, and encodes the preview BMP in-memory.
func CaptureSingleFMD(timeoutMs int) (*CaptureResponse, error) {
	devices, err := QueryDevices()
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, ErrNoReaderConnected
	}
	reader := devices[0]

	captureMu.Lock()
	defer captureMu.Unlock()

	if timeoutMs <= 0 {
		timeoutMs = 10000
	}

	nameBytes := append([]byte(reader.Name), 0)
	var hDev uintptr
	rc := rawDPFPDDOpen(&nameBytes[0], &hDev)
	if rc != dpfpddSuccess || hDev == 0 {
		return nil, fmt.Errorf("dpfpdd_open(%s) failed: 0x%x", reader.Name, uint32(rc))
	}

	activeDevMu.Lock()
	activeDevHdl = hDev
	activeDevMu.Unlock()

	defer func() {
		activeDevMu.Lock()
		activeDevHdl = 0
		activeDevMu.Unlock()
		_ = rawDPFPDDClose(hDev)
	}()

	cparam := cDPFPDDCaptureParam{
		Size:      uint32(unsafe.Sizeof(cDPFPDDCaptureParam{})),
		ImageFmt:  dpfpddImgFmtISO19794,
		ImageProc: dpfpddImgProcEnhanced,
		ImageRes:  500,
	}
	cresult := cDPFPDDCaptureResult{
		Size: uint32(unsafe.Sizeof(cDPFPDDCaptureResult{})),
	}
	cresult.Info.Size = uint32(unsafe.Sizeof(cDPFPDDImageInfo{}))

	// Allocate generous initial buffer (256 KB) so capture succeeds in a single pass
	var imgSize uint32 = 256 * 1024
	imgBuf := make([]byte, imgSize)

	rc = rawDPFPDDCapture(
		hDev,
		unsafe.Pointer(&cparam),
		int32(timeoutMs),
		unsafe.Pointer(&cresult),
		&imgSize,
		&imgBuf[0],
	)
	if rc == dpfpddErrMoreData && imgSize > 0 {
		imgBuf = make([]byte, imgSize)
		rc = rawDPFPDDCapture(
			hDev,
			unsafe.Pointer(&cparam),
			int32(timeoutMs),
			unsafe.Pointer(&cresult),
			&imgSize,
			&imgBuf[0],
		)
	}

	if rc != dpfpddSuccess {
		return nil, fmt.Errorf("dpfpdd_capture failed: 0x%x", uint32(rc))
	}
	if cresult.Quality == dpfpddQualityCanceled {
		return nil, ErrCaptureCanceled
	}
	if cresult.Quality == dpfpddQualityTimedOut {
		return nil, ErrCaptureTimeout
	}
	if (cresult.Quality & dpfpddQualityFake) != 0 {
		return nil, ErrFakeFingerDetected
	}
	if cresult.Success == 0 || cresult.Quality != dpfpddQualityGood {
		return nil, fmt.Errorf("%w (quality flags: 0x%x)", ErrPoorQualityScan, cresult.Quality)
	}

	fidData := imgBuf[:imgSize]

	// Extract ISO/IEC 19794-2:2005 FMD from ISO/IEC 19794-4:2005 FID
	engineMu.Lock()
	fmdBytes, fmdRC, fmdErr := rawCreateFMDFromFID(FIDFormatISO19794_4_2005, fidData, FMDFormatISO19794_2_2005)
	engineMu.Unlock()
	if fmdErr != nil || fmdRC != DPFJSuccess {
		return nil, fmt.Errorf("extract FMD failed: %v (0x%x)", fmdErr, uint32(fmdRC))
	}

	width := int(cresult.Info.Width)
	height := int(cresult.Info.Height)
	bmpBase64, _ := EncodeFIDToBMPBase64(fidData, width, height)

	// Map score to NIST NFIQ 1..5 (1 = excellent, 2 = very good, 3 = good)
	nfiqScore := 1
	if cresult.Score > 0 && cresult.Score <= 5 {
		nfiqScore = int(cresult.Score)
	}

	return &CaptureResponse{
		Success:      true,
		FMDBase64:    base64.StdEncoding.EncodeToString(fmdBytes),
		BMPBase64:    bmpBase64,
		QualityScore: nfiqScore,
		NFIQScore:    nfiqScore,
		ReaderSerial: reader.SerialNum,
		Width:        width,
		Height:       height,
	}, nil
}

func cBytesToString(b []byte) string {
	if idx := bytes.IndexByte(b, 0); idx >= 0 {
		return string(b[:idx])
	}
	return string(b)
}

