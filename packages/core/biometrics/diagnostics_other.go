//go:build !windows

package biometrics

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
	devices, err := QueryDevices()
	if err != nil {
		d.QueryError = err.Error()
		return d
	}
	d.DeviceCount = len(devices)
	d.DPFPDDInitialized = true
	return d
}
