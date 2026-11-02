//go:build (!windows && !cgo) || (!windows && !linux)

package biometrics

import "unsafe"

// InitRuntime is a stub when CGO is disabled or on unsupported platforms.
func InitRuntime() error {
	return ErrEngineUnavailable
}

// RuntimeDir returns empty string on unsupported / no-cgo platforms.
func RuntimeDir() string {
	return ""
}

func platformEngineAvailable() bool {
	return false
}

func platformCaptureAvailable() bool {
	return false
}

func rawCompare(fmt1 int32, fmd1 []byte, view1 uint32, fmt2 int32, fmd2 []byte, view2 uint32) (uint32, int32, error) {
	return ProbabilityOne, -1, ErrEngineUnavailable
}

func rawCreateFMDFromFID(fidFormat int32, fid []byte, fmdFormat int32) ([]byte, int32, error) {
	return nil, -1, ErrEngineUnavailable
}

func rawCreateFMDFromRaw(imageData []byte, width, height, dpi uint32, fmdFormat int32) ([]byte, int32, error) {
	return nil, -1, ErrEngineUnavailable
}

func rawStartEnrollment(fmdFormat int32) (int32, error) {
	return -1, ErrEngineUnavailable
}

func rawAddToEnrollment(fmdFormat int32, fmd []byte, viewIdx uint32) (int32, error) {
	return -1, ErrEngineUnavailable
}

func rawCreateEnrollmentFMD() ([]byte, int32, error) {
	return nil, -1, ErrEngineUnavailable
}

func rawFinishEnrollment() int32 {
	return -1
}

func rawDPFPDDInit() int32 {
	return -1
}

func rawDPFPDDExit() int32 {
	return -1
}

func rawDPFPDDQueryDevices(devCnt *uint32, devInfos unsafe.Pointer) int32 {
	return -1
}

func rawDPFPDDOpen(devName *byte, phDev *uintptr) int32 {
	return -1
}

func rawDPFPDDClose(hDev uintptr) int32 {
	return -1
}

func rawDPFPDDCapture(hDev uintptr, cparam unsafe.Pointer, timeoutMs int32, cresult unsafe.Pointer, imgSize *uint32, imgData *byte) int32 {
	return -1
}

func rawDPFPDDOpenReader(devName *byte, phDev *uintptr) int32 {
	return -1
}

func rawDPFPDDReaderImageDPI(hDev uintptr) uint32 {
	return 500
}

func rawDPFPDDGetDeviceCapabilities(hDev uintptr, caps unsafe.Pointer) int32 {
	return -1
}

func rawDPFPDDGetDeviceStatus(hDev uintptr, status unsafe.Pointer) int32 {
	return -1
}

func rawDPFPDDCalibrate(hDev uintptr) int32 {
	return -1
}

func rawDPFPDDPrepareReaderForCapture(hDev uintptr) error {
	return ErrEngineUnavailable
}

func rawDPFPDDCancel(hDev uintptr) int32 {
	return -1
}
