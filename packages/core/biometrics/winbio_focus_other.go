//go:build !windows

package biometrics

func withWinBioCaptureForeground(capture func() (*CaptureResponse, error)) (*CaptureResponse, error) {
	return capture()
}
