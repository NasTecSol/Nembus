//go:build windows

package biometrics

import (
	"fmt"
	"unsafe"
)

// CaptureDiagnostics reports DigitalPersona runtime and enumeration state for troubleshooting.
type CaptureDiagnostics struct {
	RuntimeDir        string `json:"runtime_dir"`
	CaptureLibraryOK  bool   `json:"capture_library_ok"`
	EngineLibraryOK   bool   `json:"engine_library_ok"`
	DPFPDDInitialized bool   `json:"dpfpdd_initialized"`
	QueryResultCode   int32  `json:"query_result_code,omitempty"`
	QueryError        string `json:"query_error,omitempty"`
	DeviceCount       int    `json:"device_count"`
}

// GetCaptureDiagnostics probes dpfpdd initialization and device enumeration.
func GetCaptureDiagnostics() CaptureDiagnostics {
	d := CaptureDiagnostics{
		RuntimeDir:       RuntimeDir(),
		CaptureLibraryOK: IsCaptureAvailable(),
		EngineLibraryOK:  IsEngineAvailable(),
	}
	if err := InitRuntime(); err != nil {
		d.QueryError = err.Error()
		return d
	}
	if !IsCaptureAvailable() {
		d.QueryError = ErrCaptureUnavailable.Error()
		return d
	}

	captureMu.Lock()
	defer captureMu.Unlock()

	if err := ensureDPFPDDInitialized(); err != nil {
		d.QueryError = err.Error()
		return d
	}
	d.DPFPDDInitialized = dpfpddInited

	var devCnt uint32 = 1
	infos := make([]cDPFPDDDevInfo, devCnt)
	for i := range infos {
		infos[i].Size = uint32(unsafe.Sizeof(cDPFPDDDevInfo{}))
	}
	rc := rawDPFPDDQueryDevices(&devCnt, unsafe.Pointer(&infos[0]))
	d.QueryResultCode = rc
	if rc == dpfpddErrMoreData && devCnt > 0 {
		infos = make([]cDPFPDDDevInfo, devCnt)
		for i := range infos {
			infos[i].Size = uint32(unsafe.Sizeof(cDPFPDDDevInfo{}))
		}
		rc = rawDPFPDDQueryDevices(&devCnt, unsafe.Pointer(&infos[0]))
		d.QueryResultCode = rc
	}
	if rc != dpfpddSuccess {
		// If dpfpdd failed or saw 0 readers, check Windows Biometric Framework
		wbfDevs, _ := QueryWinBioDevices()
		if len(wbfDevs) > 0 {
			d.DeviceCount = len(wbfDevs)
			d.QueryError = ""
			return d
		}
		d.QueryError = fmt.Sprintf("dpfpdd_query_devices failed: 0x%x", uint32(rc))
		return d
	}
	d.DeviceCount = int(devCnt)
	if d.DeviceCount == 0 {
		wbfDevs, _ := QueryWinBioDevices()
		if len(wbfDevs) > 0 {
			d.DeviceCount = len(wbfDevs)
		}
	}
	return d
}
