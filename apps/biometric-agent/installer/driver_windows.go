//go:build windows

package installer

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/NasTecSol/nembus-core/biometrics"
)

// DriverStatus describes the current OS driver and SDK state.
type DriverStatus struct {
	Installed         bool           `json:"installed"`
	DriverName        string         `json:"driver_name"`
	InfPath           string         `json:"inf_path"`
	IsAdmin           bool           `json:"is_admin"`
	RuntimeDir        string         `json:"runtime_dir"`
	DLLsAvailable     bool           `json:"dlls_available"`
	Message             string         `json:"message"`
	NeedsDriverRepair   bool           `json:"needs_driver_repair"`
	WBFReaderDetected   bool           `json:"wbf_reader_detected"`
	PrimaryReaderStack  string         `json:"primary_reader_stack,omitempty"`
	UnsupportedWBF      bool           `json:"unsupported_wbf"`
	USBReaders          []USBReaderPnP `json:"usb_readers,omitempty"`
	SDKDeviceCount      int            `json:"sdk_device_count"`
	ReaderReady         bool           `json:"reader_ready"`
	WinBioNativeMode    bool           `json:"winbio_native_mode"`
}

var (
	shell32           = syscall.NewLazyDLL("shell32.dll")
	procShellExecute    = shell32.NewProc("ShellExecuteW")
	procIsUserAnAdmin   = shell32.NewProc("IsUserAnAdmin")
)

// IsRunningAsAdmin checks if the current process holds elevated administrator rights.
func IsRunningAsAdmin() bool {
	if procIsUserAnAdmin.Find() != nil {
		return false
	}
	r1, _, _ := procIsUserAnAdmin.Call()
	return r1 != 0
}

// GetStatus checks driver and DLL availability on the current machine.
func GetStatus() DriverStatus {
	_ = biometrics.InitRuntime()
	dir := biometrics.RuntimeDir()
	isAdmin := IsRunningAsAdmin()

	infPath := filepath.Join(dir, "dPersona_x64.inf")
	dllsOK := biometrics.IsCaptureAvailable() && biometrics.IsEngineAvailable()
	diag := biometrics.GetCaptureDiagnostics()

	usbReaders := ListUSBReaders()
	driverInStore := checkPnputilDriverInstalled()
	wbfReader := hasWBFReader(usbReaders)
	stack := primaryReaderStack(usbReaders)
	pnpMismatch := len(usbReaders) > 0 && needsDriverRebind(usbReaders)
	needsRepair := len(usbReaders) > 0 && diag.DeviceCount == 0 && !wbfReader

	installed := diag.DeviceCount > 0
	unsupportedWBF := wbfReader && diag.DeviceCount == 0
	winBioUnits, _ := biometrics.QueryWinBioDevices()
	winBioNative := wbfReader && len(winBioUnits) > 0 && diag.DeviceCount > 0
	readerReady := diag.DeviceCount > 0

	msg := "DigitalPersona drivers and runtime are operational"
	switch {
	case len(usbReaders) == 0 && diag.DeviceCount == 0:
		msg = "No DigitalPersona USB reader detected. Plug in the scanner and refresh."
	case wbfReader && diag.DeviceCount > 0:
		msg = "DigitalPersona reader is active via Windows Biometric Framework (WinBio native mode). Hardware capture and enrollment are fully operational."
	case wbfReader && diag.DeviceCount == 0:
		msg = "This reader is in WBF mode but Windows Biometric Service did not enumerate it. Restart the \"Windows Biometric Service\" (WbioSrvc), then refresh. If count stays 0, run Install / Repair Drivers as administrator and replug USB."
	case needsRepair && !wbfReader:
		msg = "Reader is connected but bound to the wrong Windows driver (often Windows Hello). Click Install / Repair Drivers and approve UAC, then unplug and replug the USB cable."
	case diag.DeviceCount > 0 && pnpMismatch:
		msg = "SDK sees the reader and capture should work. Windows may still list a secondary Hello/WBF USB entry; ignore it if test capture succeeds."
	case !driverInStore && diag.DeviceCount == 0:
		msg = "DigitalPersona USB driver package is not installed. Click Install / Repair Drivers (administrator required)."
	case !dllsOK:
		msg = "DigitalPersona runtime DLLs are missing from the local cache. Reinstall the agent or copy SDK redistributables."
	case diag.DeviceCount == 0:
		msg = "Driver package is present but the SDK cannot see the reader. Run driver repair as administrator and replug the device."
	}

	return DriverStatus{
		Installed:           installed,
		DriverName:          "DigitalPersona U.are.U Fingerprint Reader Driver (UsbdpFP.sys)",
		InfPath:             infPath,
		IsAdmin:             isAdmin,
		RuntimeDir:          dir,
		DLLsAvailable:       dllsOK,
		Message:             msg,
		NeedsDriverRepair:   needsRepair,
		WBFReaderDetected:   wbfReader,
		PrimaryReaderStack:  stack,
		UnsupportedWBF:      unsupportedWBF,
		USBReaders:          usbReaders,
		SDKDeviceCount:      diag.DeviceCount,
		ReaderReady:         readerReady,
		WinBioNativeMode:    winBioNative,
	}
}

func checkPnputilDriverInstalled() bool {
	cmd := exec.Command("pnputil.exe", "/enum-drivers")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return false
	}
	output := strings.ToLower(out.String())
	return strings.Contains(output, "dpersona") || strings.Contains(output, "digitalpersona")
}

// InstallDrivers runs driver installation.
// If the caller is not an administrator, it automatically requests UAC elevation via Windows ShellExecute.
func InstallDrivers() error {
	if err := biometrics.InitRuntime(); err != nil {
		return fmt.Errorf("initialize runtime: %w", err)
	}
	dir := biometrics.RuntimeDir()

	if !IsRunningAsAdmin() {
		log.Printf("[DriverInstaller] Requesting administrative elevation via UAC...")
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("resolve executable path: %w", err)
		}
		if err := runElevated(exePath, "--install-drivers-worker"); err != nil {
			return fmt.Errorf("elevation failed: %w", err)
		}
		return nil
	}

	log.Printf("[DriverInstaller] Running elevated driver staging from: %s", dir)
	return executeDriverSetup(dir)
}

func executeDriverSetup(dir string) error {
	infPath := filepath.Join(dir, "dPersona_x64.inf")
	dpinstPath := filepath.Join(dir, "DPInst64.exe")

	var lastErr error
	installedViaPnP := false

	if _, err := os.Stat(infPath); err == nil {
		log.Printf("[DriverInstaller] Executing pnputil.exe /add-driver \"%s\" /install", infPath)
		cmd := exec.Command("pnputil.exe", "/add-driver", infPath, "/install")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		out, err := cmd.CombinedOutput()
		log.Printf("[DriverInstaller] pnputil output: %s", strings.TrimSpace(string(out)))
		if err != nil {
			lastErr = fmt.Errorf("pnputil install: %w (%s)", err, strings.TrimSpace(string(out)))
		} else {
			installedViaPnP = true
		}
	}

	if _, err := os.Stat(dpinstPath); err == nil {
		log.Printf("[DriverInstaller] Running DPInst64.exe /S /SA")
		cmd := exec.Command(dpinstPath, "/S", "/SA", "/PATH", dir)
		cmd.Dir = dir
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if err := cmd.Run(); err == nil {
			installedViaPnP = true
		} else if lastErr == nil {
			lastErr = fmt.Errorf("DPInst64 install failed: %w", err)
		}
	}

	if !installedViaPnP && lastErr != nil {
		return lastErr
	}
	if !installedViaPnP {
		return fmt.Errorf("no valid driver installer or INF found at %s", dir)
	}

	if err := rebindDigitalPersonaReaders(infPath); err != nil {
		log.Printf("[DriverInstaller] Rebind pass returned: %v", err)
	}

	diag := biometrics.GetCaptureDiagnostics()
	if diag.DeviceCount == 0 {
		if hasWBFReader(ListUSBReaders()) {
			wbfDevs, _ := biometrics.QueryWinBioDevices()
			if len(wbfDevs) > 0 {
				log.Printf("[DriverInstaller] WBF reader active via WinBio (%d unit(s)); legacy dpfpdd count may stay 0", len(wbfDevs))
				_ = os.WriteFile(filepath.Join(dir, ".driver_staged"), []byte("ok"), 0644)
				return nil
			}
			return fmt.Errorf("WBF reader detected but Windows Biometric Service did not enumerate it. Restart WbioSrvc, replug USB, then refresh status")
		}
		return fmt.Errorf("driver package installed but SDK still sees 0 readers (device_count=0). Unplug the USB reader, wait 5 seconds, plug it back in, then refresh status")
	}

	_ = os.WriteFile(filepath.Join(dir, ".driver_staged"), []byte("ok"), 0644)
	log.Printf("[DriverInstaller] SDK reports %d reader(s) ready", diag.DeviceCount)
	return nil
}

func runElevated(exePath, args string) error {
	verbPtr, _ := syscall.UTF16PtrFromString("runas")
	filePtr, _ := syscall.UTF16PtrFromString(exePath)
	argPtr, _ := syscall.UTF16PtrFromString(args)
	cwdPtr, _ := syscall.UTF16PtrFromString(filepath.Dir(exePath))

	ret, _, _ := procShellExecute.Call(
		0,
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(filePtr)),
		uintptr(unsafe.Pointer(argPtr)),
		uintptr(unsafe.Pointer(cwdPtr)),
		1,
	)
	if ret <= 32 {
		return fmt.Errorf("ShellExecute failed with code %d (User rejected UAC prompt or access denied)", ret)
	}
	return nil
}
