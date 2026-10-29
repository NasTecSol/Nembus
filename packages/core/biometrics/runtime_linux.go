//go:build linux

package biometrics

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>
#include <stdint.h>

typedef int32_t (*fn_dpfj_compare)(int32_t, const uint8_t*, uint32_t, uint32_t, int32_t, const uint8_t*, uint32_t, uint32_t, uint32_t*);
typedef int32_t (*fn_dpfj_create_fmd_from_fid)(int32_t, const uint8_t*, uint32_t, int32_t, uint8_t*, uint32_t*);
typedef int32_t (*fn_dpfj_start_enrollment)(int32_t);
typedef int32_t (*fn_dpfj_add_to_enrollment)(int32_t, const uint8_t*, uint32_t, uint32_t);
typedef int32_t (*fn_dpfj_create_enrollment_fmd)(uint8_t*, uint32_t*);
typedef int32_t (*fn_dpfj_finish_enrollment)(void);

typedef int32_t (*fn_dpfpdd_init)(void);
typedef int32_t (*fn_dpfpdd_exit)(void);
typedef int32_t (*fn_dpfpdd_query_devices)(uint32_t*, void*);
typedef int32_t (*fn_dpfpdd_open)(const char*, uintptr_t*);
typedef int32_t (*fn_dpfpdd_close)(uintptr_t);
typedef int32_t (*fn_dpfpdd_capture)(uintptr_t, void*, uint32_t, void*, uint32_t*, uint8_t*);
typedef int32_t (*fn_dpfpdd_cancel)(uintptr_t);

static int32_t call_dpfj_compare(void* fn, int32_t fmt1, const uint8_t* fmd1, uint32_t len1, uint32_t v1, int32_t fmt2, const uint8_t* fmd2, uint32_t len2, uint32_t v2, uint32_t* score) {
    return ((fn_dpfj_compare)fn)(fmt1, fmd1, len1, v1, fmt2, fmd2, len2, v2, score);
}

static int32_t call_dpfj_create_fmd_from_fid(void* fn, int32_t fid_fmt, const uint8_t* fid, uint32_t fid_len, int32_t fmd_fmt, uint8_t* fmd, uint32_t* fmd_len) {
    return ((fn_dpfj_create_fmd_from_fid)fn)(fid_fmt, fid, fid_len, fmd_fmt, fmd, fmd_len);
}

static int32_t call_dpfj_start_enrollment(void* fn, int32_t fmd_fmt) {
    return ((fn_dpfj_start_enrollment)fn)(fmd_fmt);
}

static int32_t call_dpfj_add_to_enrollment(void* fn, int32_t fmd_fmt, const uint8_t* fmd, uint32_t fmd_len, uint32_t view_idx) {
    return ((fn_dpfj_add_to_enrollment)fn)(fmd_fmt, fmd, fmd_len, view_idx);
}

static int32_t call_dpfj_create_enrollment_fmd(void* fn, uint8_t* fmd, uint32_t* fmd_len) {
    return ((fn_dpfj_create_enrollment_fmd)fn)(fmd, fmd_len);
}

static int32_t call_dpfj_finish_enrollment(void* fn) {
    return ((fn_dpfj_finish_enrollment)fn)();
}

static int32_t call_dpfpdd_init(void* fn) {
    return ((fn_dpfpdd_init)fn)();
}

static int32_t call_dpfpdd_exit(void* fn) {
    return ((fn_dpfpdd_exit)fn)();
}

static int32_t call_dpfpdd_query_devices(void* fn, uint32_t* cnt, void* infos) {
    return ((fn_dpfpdd_query_devices)fn)(cnt, infos);
}

static int32_t call_dpfpdd_open(void* fn, const char* name, uintptr_t* hdev) {
    return ((fn_dpfpdd_open)fn)(name, hdev);
}

static int32_t call_dpfpdd_close(void* fn, uintptr_t hdev) {
    return ((fn_dpfpdd_close)fn)(hdev);
}

static int32_t call_dpfpdd_capture(void* fn, uintptr_t hdev, void* cparam, uint32_t timeout, void* cresult, uint32_t* size, uint8_t* data) {
    return ((fn_dpfpdd_capture)fn)(hdev, cparam, timeout, cresult, size, data);
}

static int32_t call_dpfpdd_cancel(void* fn, uintptr_t hdev) {
    return ((fn_dpfpdd_cancel)fn)(hdev);
}
*/
import "C"

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"unsafe"
)

//go:embed vendor/linux_amd64/*
var embeddedLinuxAssets embed.FS

var (
	runtimeOnce sync.Once
	runtimeDir  string
	runtimeErr  error

	handleDPFJ   unsafe.Pointer
	handleDPFPDD unsafe.Pointer

	fnCompare             unsafe.Pointer
	fnCreateFMDFromFID    unsafe.Pointer
	fnStartEnrollment     unsafe.Pointer
	fnAddToEnrollment     unsafe.Pointer
	fnCreateEnrollmentFMD unsafe.Pointer
	fnFinishEnrollment    unsafe.Pointer

	fnDPFPDDInit         unsafe.Pointer
	fnDPFPDDExit         unsafe.Pointer
	fnDPFPDDQueryDevices unsafe.Pointer
	fnDPFPDDOpen         unsafe.Pointer
	fnDPFPDDClose        unsafe.Pointer
	fnDPFPDDCapture      unsafe.Pointer
	fnDPFPDDCancel       unsafe.Pointer
)

func InitRuntime() error {
	runtimeOnce.Do(func() {
		cacheBase, err := os.UserCacheDir()
		if err != nil || cacheBase == "" {
			cacheBase = os.TempDir()
		}
		runtimeDir = filepath.Join(cacheBase, "Nembus", "biometrics", "v3.4.0", "linux_amd64")
		if err := os.MkdirAll(runtimeDir, 0755); err != nil {
			runtimeErr = fmt.Errorf("create linux biometric runtime dir: %w", err)
			return
		}

		entries, err := fs.ReadDir(embeddedLinuxAssets, "vendor/linux_amd64")
		if err != nil {
			runtimeErr = fmt.Errorf("read embedded linux biometric assets: %w", err)
			return
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			data, err := embeddedLinuxAssets.ReadFile("vendor/linux_amd64/" + entry.Name())
			if err != nil {
				continue
			}
			destPath := filepath.Join(runtimeDir, entry.Name())
			if info, statErr := os.Stat(destPath); statErr != nil || info.Size() != int64(len(data)) {
				_ = os.WriteFile(destPath, data, 0755)
			}
		}

		// Create standard SONAME symlinks required by dynamic dependencies
		symlinks := map[string][]string{
			"libdpfj.so.3.0.0":        {"libdpfj.so.3", "libdpfj.so"},
			"libdpfr6.so.3.0.0":       {"libdpfr6.so.3", "libdpfr6.so"},
			"libdpfr7.so.3.0.0":       {"libdpfr7.so.3", "libdpfr7.so"},
			"libdpfpdd.so.3.0.0":      {"libdpfpdd.so.3", "libdpfpdd.so"},
			"libdpfpdd_4k.so.5.0.0":   {"libdpfpdd_4k.so.5", "libdpfpdd_4k.so"},
			"libdpfpdd5000.so.3.1.1":  {"libdpfpdd5000.so.3", "libdpfpdd5000.so"},
			"libdpfpdd_ptapi.so.1.0.0": {"libdpfpdd_ptapi.so.1", "libdpfpdd_ptapi.so"},
			"libdpuvc.so.3.1.1":       {"libdpuvc.so.3", "libdpuvc.so"},
		}
		for target, links := range symlinks {
			for _, link := range links {
				linkPath := filepath.Join(runtimeDir, link)
				_ = os.Remove(linkPath)
				_ = os.Symlink(target, linkPath)
			}
		}

		// Preload dependent recognition engines into global symbol space
		for _, dep := range []string{"libdpfr6.so.3.0.0", "libdpfr7.so.3.0.0", "libtfm.so", "libdpuvc.so.3.1.1"} {
			cPath := C.CString(filepath.Join(runtimeDir, dep))
			C.dlopen(cPath, C.RTLD_NOW|C.RTLD_GLOBAL)
			C.free(unsafe.Pointer(cPath))
		}

		// Load FingerJet Engine
		dpfjPath := C.CString(filepath.Join(runtimeDir, "libdpfj.so.3.0.0"))
		handleDPFJ = C.dlopen(dpfjPath, C.RTLD_NOW|C.RTLD_GLOBAL)
		C.free(unsafe.Pointer(dpfjPath))
		if handleDPFJ != nil {
			fnCompare = dlsymLookup(handleDPFJ, "dpfj_compare")
			fnCreateFMDFromFID = dlsymLookup(handleDPFJ, "dpfj_create_fmd_from_fid")
			fnStartEnrollment = dlsymLookup(handleDPFJ, "dpfj_start_enrollment")
			fnAddToEnrollment = dlsymLookup(handleDPFJ, "dpfj_add_to_enrollment")
			fnCreateEnrollmentFMD = dlsymLookup(handleDPFJ, "dpfj_create_enrollment_fmd")
			fnFinishEnrollment = dlsymLookup(handleDPFJ, "dpfj_finish_enrollment")
		}

		// Load Capture Driver
		dpfpddPath := C.CString(filepath.Join(runtimeDir, "libdpfpdd.so.3.0.0"))
		handleDPFPDD = C.dlopen(dpfpddPath, C.RTLD_NOW|C.RTLD_GLOBAL)
		C.free(unsafe.Pointer(dpfpddPath))
		if handleDPFPDD != nil {
			fnDPFPDDInit = dlsymLookup(handleDPFPDD, "dpfpdd_init")
			fnDPFPDDExit = dlsymLookup(handleDPFPDD, "dpfpdd_exit")
			fnDPFPDDQueryDevices = dlsymLookup(handleDPFPDD, "dpfpdd_query_devices")
			fnDPFPDDOpen = dlsymLookup(handleDPFPDD, "dpfpdd_open")
			fnDPFPDDClose = dlsymLookup(handleDPFPDD, "dpfpdd_close")
			fnDPFPDDCapture = dlsymLookup(handleDPFPDD, "dpfpdd_capture")
			fnDPFPDDCancel = dlsymLookup(handleDPFPDD, "dpfpdd_cancel")
		}
	})
	return runtimeErr
}

// RuntimeDir returns the local directory where embedded DigitalPersona runtime files were extracted.
func RuntimeDir() string {
	return runtimeDir
}

func dlsymLookup(handle unsafe.Pointer, symbol string) unsafe.Pointer {
	cSym := C.CString(symbol)
	defer C.free(unsafe.Pointer(cSym))
	return C.dlsym(handle, cSym)
}

func platformEngineAvailable() bool {
	if err := InitRuntime(); err != nil {
		return false
	}
	return fnCompare != nil
}

func platformCaptureAvailable() bool {
	if err := InitRuntime(); err != nil {
		return false
	}
	return fnDPFPDDInit != nil
}

func rawCompare(fmt1 int32, fmd1 []byte, view1 uint32, fmt2 int32, fmd2 []byte, view2 uint32) (uint32, int32, error) {
	if !platformEngineAvailable() {
		return ProbabilityOne, -1, ErrEngineUnavailable
	}
	var score C.uint32_t = C.uint32_t(ProbabilityOne)
	res := C.call_dpfj_compare(
		fnCompare,
		C.int32_t(fmt1),
		(*C.uint8_t)(unsafe.Pointer(&fmd1[0])),
		C.uint32_t(len(fmd1)),
		C.uint32_t(view1),
		C.int32_t(fmt2),
		(*C.uint8_t)(unsafe.Pointer(&fmd2[0])),
		C.uint32_t(len(fmd2)),
		C.uint32_t(view2),
		&score,
	)
	return uint32(score), int32(res), nil
}

func rawCreateFMDFromFID(fidFormat int32, fid []byte, fmdFormat int32) ([]byte, int32, error) {
	if !platformEngineAvailable() || fnCreateFMDFromFID == nil {
		return nil, -1, ErrEngineUnavailable
	}
	outBuf := make([]byte, MaxFMDSize)
	var outSize C.uint32_t = C.uint32_t(MaxFMDSize)
	res := C.call_dpfj_create_fmd_from_fid(
		fnCreateFMDFromFID,
		C.int32_t(fidFormat),
		(*C.uint8_t)(unsafe.Pointer(&fid[0])),
		C.uint32_t(len(fid)),
		C.int32_t(fmdFormat),
		(*C.uint8_t)(unsafe.Pointer(&outBuf[0])),
		&outSize,
	)
	code := int32(res)
	if code != DPFJSuccess {
		return nil, code, fmt.Errorf("dpfj_create_fmd_from_fid returned 0x%x", uint32(code))
	}
	return outBuf[:uint32(outSize)], DPFJSuccess, nil
}

func rawStartEnrollment(fmdFormat int32) (int32, error) {
	if !platformEngineAvailable() || fnStartEnrollment == nil {
		return -1, ErrEngineUnavailable
	}
	res := C.call_dpfj_start_enrollment(fnStartEnrollment, C.int32_t(fmdFormat))
	return int32(res), nil
}

func rawAddToEnrollment(fmdFormat int32, fmd []byte, viewIdx uint32) (int32, error) {
	if !platformEngineAvailable() || fnAddToEnrollment == nil {
		return -1, ErrEngineUnavailable
	}
	res := C.call_dpfj_add_to_enrollment(
		fnAddToEnrollment,
		C.int32_t(fmdFormat),
		(*C.uint8_t)(unsafe.Pointer(&fmd[0])),
		C.uint32_t(len(fmd)),
		C.uint32_t(viewIdx),
	)
	return int32(res), nil
}

func rawCreateEnrollmentFMD() ([]byte, int32, error) {
	if !platformEngineAvailable() || fnCreateEnrollmentFMD == nil {
		return nil, -1, ErrEngineUnavailable
	}
	outBuf := make([]byte, MaxFMDSize)
	var outSize C.uint32_t = C.uint32_t(MaxFMDSize)
	res := C.call_dpfj_create_enrollment_fmd(
		fnCreateEnrollmentFMD,
		(*C.uint8_t)(unsafe.Pointer(&outBuf[0])),
		&outSize,
	)
	code := int32(res)
	if code != DPFJSuccess {
		return nil, code, fmt.Errorf("dpfj_create_enrollment_fmd returned 0x%x", uint32(code))
	}
	return outBuf[:uint32(outSize)], DPFJSuccess, nil
}

func rawFinishEnrollment() int32 {
	if !platformEngineAvailable() || fnFinishEnrollment == nil {
		return -1
	}
	return int32(C.call_dpfj_finish_enrollment(fnFinishEnrollment))
}

func rawDPFPDDInit() int32 {
	if !platformCaptureAvailable() {
		return -1
	}
	return int32(C.call_dpfpdd_init(fnDPFPDDInit))
}

func rawDPFPDDExit() int32 {
	if !platformCaptureAvailable() {
		return -1
	}
	return int32(C.call_dpfpdd_exit(fnDPFPDDExit))
}

func rawDPFPDDQueryDevices(devCnt *uint32, devInfos unsafe.Pointer) int32 {
	if !platformCaptureAvailable() {
		return -1
	}
	return int32(C.call_dpfpdd_query_devices(fnDPFPDDQueryDevices, (*C.uint32_t)(unsafe.Pointer(devCnt)), devInfos))
}

func rawDPFPDDOpen(devName *byte, phDev *uintptr) int32 {
	if !platformCaptureAvailable() {
		return -1
	}
	return int32(C.call_dpfpdd_open(fnDPFPDDOpen, (*C.char)(unsafe.Pointer(devName)), (*C.uintptr_t)(unsafe.Pointer(phDev))))
}

func rawDPFPDDClose(hDev uintptr) int32 {
	if !platformCaptureAvailable() {
		return -1
	}
	return int32(C.call_dpfpdd_close(fnDPFPDDClose, C.uintptr_t(hDev)))
}

func rawDPFPDDCapture(hDev uintptr, cparam unsafe.Pointer, timeoutMs int32, cresult unsafe.Pointer, imgSize *uint32, imgData *byte) int32 {
	if !platformCaptureAvailable() {
		return -1
	}
	return int32(C.call_dpfpdd_capture(
		fnDPFPDDCapture,
		C.uintptr_t(hDev),
		cparam,
		C.uint32_t(timeoutMs),
		cresult,
		(*C.uint32_t)(unsafe.Pointer(imgSize)),
		(*C.uint8_t)(unsafe.Pointer(imgData)),
	))
}

func rawDPFPDDCancel(hDev uintptr) int32 {
	if !platformCaptureAvailable() {
		return -1
	}
	return int32(C.call_dpfpdd_cancel(fnDPFPDDCancel, C.uintptr_t(hDev)))
}
