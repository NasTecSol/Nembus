//go:build !windows

package installer

import (
	"github.com/NasTecSol/nembus-core/biometrics"
)

type DriverStatus struct {
	Installed         bool   `json:"installed"`
	DriverName        string `json:"driver_name"`
	InfPath           string `json:"inf_path"`
	IsAdmin           bool   `json:"is_admin"`
	RuntimeDir        string `json:"runtime_dir"`
	DLLsAvailable     bool   `json:"dlls_available"`
	Message            string `json:"message"`
	NeedsDriverRepair  bool   `json:"needs_driver_repair"`
	WBFReaderDetected  bool   `json:"wbf_reader_detected"`
	PrimaryReaderStack string `json:"primary_reader_stack,omitempty"`
	UnsupportedWBF     bool   `json:"unsupported_wbf"`
	SDKDeviceCount     int    `json:"sdk_device_count"`
	ReaderReady        bool   `json:"reader_ready"`
	WinBioNativeMode   bool   `json:"winbio_native_mode"`
}

func IsRunningAsAdmin() bool {
	return false
}

func GetStatus() DriverStatus {
	_ = biometrics.InitRuntime()
	dir := biometrics.RuntimeDir()
	dllsOK := biometrics.IsCaptureAvailable() && biometrics.IsEngineAvailable()
	diag := biometrics.GetCaptureDiagnostics()
	return DriverStatus{
		Installed:     dllsOK && diag.DeviceCount > 0,
		DriverName:    "Linux libdpfpdd / udev rules",
		IsAdmin:       false,
		RuntimeDir:    dir,
		DLLsAvailable: dllsOK,
		Message:       "Linux/Unix platform detected",
		SDKDeviceCount: diag.DeviceCount,
		ReaderReady:    diag.DeviceCount > 0,
	}
}

func InstallDrivers() error {
	return nil
}

