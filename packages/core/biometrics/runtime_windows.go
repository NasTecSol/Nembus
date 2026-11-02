//go:build windows

package biometrics

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"
)

//go:embed vendor/windows_amd64/*
var embeddedWindowsAssets embed.FS

var (
	runtimeOnce sync.Once
	runtimeDir  string
	runtimeErr  error

	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procSetDllDirW     = kernel32.NewProc("SetDllDirectoryW")
	procAddDllDir      = kernel32.NewProc("AddDllDirectory")

	dllDPFJ   *syscall.LazyDLL
	dllDPFPDD *syscall.LazyDLL

	procCompare             *syscall.LazyProc
	procIdentify            *syscall.LazyProc
	procCreateFMDFromFID    *syscall.LazyProc
	procCreateFMDFromRaw    *syscall.LazyProc
	procStartEnrollment     *syscall.LazyProc
	procAddToEnrollment     *syscall.LazyProc
	procCreateEnrollmentFMD *syscall.LazyProc
	procFinishEnrollment    *syscall.LazyProc

	procDPFPDDInit         *syscall.LazyProc
	procDPFPDDExit         *syscall.LazyProc
	procDPFPDDQueryDevices *syscall.LazyProc
	procDPFPDDOpen         *syscall.LazyProc
	procDPFPDDOpenExt      *syscall.LazyProc
	procDPFPDDClose        *syscall.LazyProc
	procDPFPDDCapture      *syscall.LazyProc
	procDPFPDDCancel       *syscall.LazyProc
	procDPFPDDGetCaps      *syscall.LazyProc
	procDPFPDDGetStatus    *syscall.LazyProc
	procDPFPDDCalibrate    *syscall.LazyProc
)

// InitRuntime extracts embedded DigitalPersona Windows DLLs and signed USB drivers
// to a local runtime directory, configures the Windows DLL search path, and binds
// the FingerJet (dpfj.dll) and Capture (dpfpdd.dll) dynamic libraries.
func InitRuntime() error {
	runtimeOnce.Do(func() {
		cacheBase, err := os.UserCacheDir()
		if err != nil || cacheBase == "" {
			cacheBase = os.TempDir()
		}
		runtimeDir = filepath.Join(cacheBase, "Nembus", "biometrics", "v3.4.0", "windows_amd64")
		if err := os.MkdirAll(runtimeDir, 0755); err != nil {
			runtimeErr = fmt.Errorf("create biometric runtime dir: %w", err)
			return
		}

		entries, err := fs.ReadDir(embeddedWindowsAssets, "vendor/windows_amd64")
		if err != nil {
			runtimeErr = fmt.Errorf("read embedded windows biometric assets: %w", err)
			return
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			data, err := embeddedWindowsAssets.ReadFile("vendor/windows_amd64/" + entry.Name())
			if err != nil {
				continue
			}
			destPath := filepath.Join(runtimeDir, entry.Name())
			if info, statErr := os.Stat(destPath); statErr != nil || info.Size() != int64(len(data)) {
				_ = os.WriteFile(destPath, data, 0755)
			}
		}

		// Register runtimeDir in the Windows DLL search path so dependent DLLs
		// (dpfr6.dll, dpfr7.dll, dpfpdd5000.dll, dpfpdd_4k.dll, dpDevCtlx64.dll) resolve automatically.
		if dirPtr, err := syscall.UTF16PtrFromString(runtimeDir); err == nil {
			_, _, _ = procSetDllDirW.Call(uintptr(unsafe.Pointer(dirPtr)))
			_, _, _ = procAddDllDir.Call(uintptr(unsafe.Pointer(dirPtr)))
		}
		_ = os.Setenv("PATH", runtimeDir+string(os.PathListSeparator)+os.Getenv("PATH"))

		// Bind FingerJet Matching & Enrollment DLL
		dpfjPath := filepath.Join(runtimeDir, "dpfj.dll")
		dllDPFJ = syscall.NewLazyDLL(dpfjPath)
		if err := dllDPFJ.Load(); err != nil {
			// Fallback to system path if needed
			dllDPFJ = syscall.NewLazyDLL("dpfj.dll")
		}
		procCompare = dllDPFJ.NewProc("dpfj_compare")
		procIdentify = dllDPFJ.NewProc("dpfj_identify")
		procCreateFMDFromFID = dllDPFJ.NewProc("dpfj_create_fmd_from_fid")
		procCreateFMDFromRaw = dllDPFJ.NewProc("dpfj_create_fmd_from_raw")
		procStartEnrollment = dllDPFJ.NewProc("dpfj_start_enrollment")
		procAddToEnrollment = dllDPFJ.NewProc("dpfj_add_to_enrollment")
		procCreateEnrollmentFMD = dllDPFJ.NewProc("dpfj_create_enrollment_fmd")
		procFinishEnrollment = dllDPFJ.NewProc("dpfj_finish_enrollment")

		// Bind U.are.U Capture DLL
		dpfpddPath := filepath.Join(runtimeDir, "dpfpdd.dll")
		dllDPFPDD = syscall.NewLazyDLL(dpfpddPath)
		if err := dllDPFPDD.Load(); err != nil {
			dllDPFPDD = syscall.NewLazyDLL("dpfpdd.dll")
		}
		procDPFPDDInit = dllDPFPDD.NewProc("dpfpdd_init")
		procDPFPDDExit = dllDPFPDD.NewProc("dpfpdd_exit")
		procDPFPDDQueryDevices = dllDPFPDD.NewProc("dpfpdd_query_devices")
		procDPFPDDOpen = dllDPFPDD.NewProc("dpfpdd_open")
		procDPFPDDOpenExt = dllDPFPDD.NewProc("dpfpdd_open_ext")
		procDPFPDDClose = dllDPFPDD.NewProc("dpfpdd_close")
		procDPFPDDCapture = dllDPFPDD.NewProc("dpfpdd_capture")
		procDPFPDDCancel = dllDPFPDD.NewProc("dpfpdd_cancel")
		procDPFPDDGetCaps = dllDPFPDD.NewProc("dpfpdd_get_device_capabilities")
		procDPFPDDGetStatus = dllDPFPDD.NewProc("dpfpdd_get_device_status")
		procDPFPDDCalibrate = dllDPFPDD.NewProc("dpfpdd_calibrate")

		// Stage WHQL-signed USB driver silently in background for plug-and-play
		go stageWindowsUSBDriverOnce(runtimeDir)
	})
	return runtimeErr
}

// RuntimeDir returns the local directory where embedded DigitalPersona runtime files were extracted.
func RuntimeDir() string {
	return runtimeDir
}

func stageWindowsUSBDriverOnce(dir string) {
	marker := filepath.Join(dir, ".driver_staged")
	if _, err := os.Stat(marker); err == nil {
		return
	}
	infPath := filepath.Join(dir, "dPersona_x64.inf")
	if _, err := os.Stat(infPath); err != nil {
		return
	}
	cmd := exec.Command("pnputil.exe", "/add-driver", infPath, "/install")
	if err := cmd.Run(); err == nil {
		_ = os.WriteFile(marker, []byte("ok"), 0644)
	}
}

func platformEngineAvailable() bool {
	if err := InitRuntime(); err != nil || procCompare == nil {
		return false
	}
	return procCompare.Find() == nil
}

func platformCaptureAvailable() bool {
	if err := InitRuntime(); err != nil || procDPFPDDInit == nil {
		return false
	}
	return procDPFPDDInit.Find() == nil
}

func rawCompare(fmt1 int32, fmd1 []byte, view1 uint32, fmt2 int32, fmd2 []byte, view2 uint32) (uint32, int32, error) {
	if !platformEngineAvailable() {
		return ProbabilityOne, -1, ErrEngineUnavailable
	}
	var score uint32 = ProbabilityOne
	r1, _, _ := procCompare.Call(
		uintptr(fmt1),
		uintptr(unsafe.Pointer(&fmd1[0])),
		uintptr(len(fmd1)),
		uintptr(view1),
		uintptr(fmt2),
		uintptr(unsafe.Pointer(&fmd2[0])),
		uintptr(len(fmd2)),
		uintptr(view2),
		uintptr(unsafe.Pointer(&score)),
	)
	return score, int32(r1), nil
}

func rawCreateFMDFromFID(fidFormat int32, fid []byte, fmdFormat int32) ([]byte, int32, error) {
	if !platformEngineAvailable() {
		return nil, -1, ErrEngineUnavailable
	}
	outBuf := make([]byte, MaxFMDSize)
	var outSize uint32 = MaxFMDSize
	r1, _, _ := procCreateFMDFromFID.Call(
		uintptr(fidFormat),
		uintptr(unsafe.Pointer(&fid[0])),
		uintptr(len(fid)),
		uintptr(fmdFormat),
		uintptr(unsafe.Pointer(&outBuf[0])),
		uintptr(unsafe.Pointer(&outSize)),
	)
	code := int32(r1)
	if code != DPFJSuccess {
		return nil, code, fmt.Errorf("dpfj_create_fmd_from_fid returned 0x%x", uint32(code))
	}
	return outBuf[:outSize], DPFJSuccess, nil
}

func rawCreateFMDFromRaw(imageData []byte, width, height, dpi uint32, fmdFormat int32) ([]byte, int32, error) {
	if !platformEngineAvailable() || procCreateFMDFromRaw == nil {
		return nil, -1, ErrEngineUnavailable
	}
	outBuf := make([]byte, MaxFMDSize)
	var outSize uint32 = MaxFMDSize
	r1, _, _ := procCreateFMDFromRaw.Call(
		uintptr(unsafe.Pointer(&imageData[0])),
		uintptr(len(imageData)),
		uintptr(width),
		uintptr(height),
		uintptr(dpi),
		uintptr(0), // finger_pos: DPFJ_POSITION_UNKNOWN
		uintptr(0), // cbeff_id: 0
		uintptr(fmdFormat),
		uintptr(unsafe.Pointer(&outBuf[0])),
		uintptr(unsafe.Pointer(&outSize)),
	)
	code := int32(r1)
	if code != DPFJSuccess {
		return nil, code, fmt.Errorf("dpfj_create_fmd_from_raw returned 0x%x", uint32(code))
	}
	return outBuf[:outSize], DPFJSuccess, nil
}

func rawStartEnrollment(fmdFormat int32) (int32, error) {
	if !platformEngineAvailable() {
		return -1, ErrEngineUnavailable
	}
	r1, _, _ := procStartEnrollment.Call(uintptr(fmdFormat))
	return int32(r1), nil
}

func rawAddToEnrollment(fmdFormat int32, fmd []byte, viewIdx uint32) (int32, error) {
	if !platformEngineAvailable() {
		return -1, ErrEngineUnavailable
	}
	r1, _, _ := procAddToEnrollment.Call(
		uintptr(fmdFormat),
		uintptr(unsafe.Pointer(&fmd[0])),
		uintptr(len(fmd)),
		uintptr(viewIdx),
	)
	return int32(r1), nil
}

func rawCreateEnrollmentFMD() ([]byte, int32, error) {
	if !platformEngineAvailable() {
		return nil, -1, ErrEngineUnavailable
	}
	outBuf := make([]byte, MaxFMDSize)
	var outSize uint32 = MaxFMDSize
	r1, _, _ := procCreateEnrollmentFMD.Call(
		uintptr(unsafe.Pointer(&outBuf[0])),
		uintptr(unsafe.Pointer(&outSize)),
	)
	code := int32(r1)
	if code != DPFJSuccess {
		return nil, code, fmt.Errorf("dpfj_create_enrollment_fmd returned 0x%x", uint32(code))
	}
	return outBuf[:outSize], DPFJSuccess, nil
}

func rawFinishEnrollment() int32 {
	if !platformEngineAvailable() {
		return -1
	}
	r1, _, _ := procFinishEnrollment.Call()
	return int32(r1)
}

func rawDPFPDDInit() int32 {
	if !platformCaptureAvailable() {
		return -1
	}
	r1, _, _ := procDPFPDDInit.Call()
	return int32(r1)
}

func rawDPFPDDExit() int32 {
	if !platformCaptureAvailable() {
		return -1
	}
	r1, _, _ := procDPFPDDExit.Call()
	return int32(r1)
}

func rawDPFPDDQueryDevices(devCnt *uint32, devInfos unsafe.Pointer) int32 {
	if !platformCaptureAvailable() {
		return -1
	}
	r1, _, _ := procDPFPDDQueryDevices.Call(
		uintptr(unsafe.Pointer(devCnt)),
		uintptr(devInfos),
	)
	return int32(r1)
}

func rawDPFPDDOpen(devName *byte, phDev *uintptr) int32 {
	if !platformCaptureAvailable() {
		return -1
	}
	r1, _, _ := procDPFPDDOpen.Call(
		uintptr(unsafe.Pointer(devName)),
		uintptr(unsafe.Pointer(phDev)),
	)
	return int32(r1)
}

func rawDPFPDDClose(hDev uintptr) int32 {
	if !platformCaptureAvailable() {
		return -1
	}
	r1, _, _ := procDPFPDDClose.Call(hDev)
	return int32(r1)
}

func rawDPFPDDCapture(hDev uintptr, cparam unsafe.Pointer, timeoutMs int32, cresult unsafe.Pointer, imgSize *uint32, imgData *byte) int32 {
	if !platformCaptureAvailable() {
		return -1
	}
	var imgPtr uintptr
	if imgData != nil {
		imgPtr = uintptr(unsafe.Pointer(imgData))
	}
	r1, _, _ := procDPFPDDCapture.Call(
		hDev,
		uintptr(cparam),
		uintptr(uint32(timeoutMs)),
		uintptr(cresult),
		uintptr(unsafe.Pointer(imgSize)),
		imgPtr,
	)
	return int32(r1)
}

func rawDPFPDDOpenReader(devName *byte, phDev *uintptr) int32 {
	if !platformCaptureAvailable() {
		return -1
	}
	if procDPFPDDOpenExt != nil && procDPFPDDOpenExt.Find() == nil {
		r1, _, _ := procDPFPDDOpenExt.Call(
			uintptr(unsafe.Pointer(devName)),
			uintptr(dpfpddPriorityExclusive),
			uintptr(unsafe.Pointer(phDev)),
		)
		return int32(r1)
	}
	return rawDPFPDDOpen(devName, phDev)
}

func rawDPFPDDReaderImageDPI(hDev uintptr) uint32 {
	const defaultDPI uint32 = 500
	if procDPFPDDGetCaps == nil || procDPFPDDGetCaps.Find() != nil {
		return defaultDPI
	}
	caps := cDPFPDDDevCaps{Size: uint32(unsafe.Sizeof(cDPFPDDDevCaps{}))}
	rc := rawDPFPDDGetDeviceCapabilities(hDev, unsafe.Pointer(&caps))
	if rc == dpfpddErrMoreData && caps.Size > uint32(unsafe.Sizeof(cDPFPDDDevCaps{})) {
		buf := make([]byte, caps.Size)
		*(*uint32)(unsafe.Pointer(&buf[0])) = caps.Size
		rc = rawDPFPDDGetDeviceCapabilities(hDev, unsafe.Pointer(&buf[0]))
		if rc == dpfpddSuccess && len(buf) >= 52 {
			cnt := *(*uint32)(unsafe.Pointer(&buf[44]))
			if cnt > 0 {
				return *(*uint32)(unsafe.Pointer(&buf[48]))
			}
		}
		return defaultDPI
	}
	if rc == dpfpddSuccess && caps.ResolutionCnt > 0 {
		return caps.Resolutions[0]
	}
	return defaultDPI
}

func rawDPFPDDGetDeviceCapabilities(hDev uintptr, caps unsafe.Pointer) int32 {
	if !platformCaptureAvailable() || procDPFPDDGetCaps == nil {
		return -1
	}
	r1, _, _ := procDPFPDDGetCaps.Call(hDev, uintptr(caps))
	return int32(r1)
}

func rawDPFPDDGetDeviceStatus(hDev uintptr, status unsafe.Pointer) int32 {
	if !platformCaptureAvailable() || procDPFPDDGetStatus == nil {
		return -1
	}
	r1, _, _ := procDPFPDDGetStatus.Call(hDev, uintptr(status))
	return int32(r1)
}

func rawDPFPDDCalibrate(hDev uintptr) int32 {
	if !platformCaptureAvailable() || procDPFPDDCalibrate == nil {
		return -1
	}
	r1, _, _ := procDPFPDDCalibrate.Call(hDev)
	return int32(r1)
}

func rawDPFPDDPrepareReaderForCapture(hDev uintptr) error {
	st := cDPFPDDDevStatus{Size: uint32(unsafe.Sizeof(cDPFPDDDevStatus{}))}
	rc := rawDPFPDDGetDeviceStatus(hDev, unsafe.Pointer(&st))
	if rc == dpfpddErrMoreData {
		buf := make([]byte, st.Size)
		if st.Size < 12 {
			buf = make([]byte, 64)
		}
		*(*uint32)(unsafe.Pointer(&buf[0])) = uint32(len(buf))
		rc = rawDPFPDDGetDeviceStatus(hDev, unsafe.Pointer(&buf[0]))
		if rc == dpfpddSuccess {
			st.Status = *(*uint32)(unsafe.Pointer(&buf[4]))
		}
	} else if rc != dpfpddSuccess {
		return nil
	}
	switch st.Status {
	case dpfpddStatusNeedCalib, dpfpddStatusFailure:
		if calRC := rawDPFPDDCalibrate(hDev); calRC != dpfpddSuccess {
			return fmt.Errorf("reader calibration failed: 0x%x", uint32(calRC))
		}
	}
	return nil
}

func rawDPFPDDCancel(hDev uintptr) int32 {
	if !platformCaptureAvailable() {
		return -1
	}
	r1, _, _ := procDPFPDDCancel.Call(hDev)
	return int32(r1)
}
