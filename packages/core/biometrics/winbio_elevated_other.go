//go:build !windows

package biometrics

var WinBioElevatedCaptureFunc func(unitID uint32, timeoutMs int) (*CaptureResponse, error)

func CaptureWinBioSampleDirect(unitID uint32, timeoutMs int) (*CaptureResponse, error) {
	return CaptureWinBioSample(unitID, timeoutMs)
}
