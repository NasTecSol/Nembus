//go:build !windows

package biometrics

import "fmt"

type WinBioDevice struct {
	UnitID       uint32
	Model        string
	Manufacturer string
	SerialNumber string
	Description  string
	InstanceID   string
}

func IsWinBioAvailable() bool {
	return false
}

func QueryWinBioDevices() ([]WinBioDevice, error) {
	return nil, nil
}

func CancelWinBioCapture() error {
	return nil
}

func CaptureWinBioSample(unitID uint32, timeoutMs int) (*CaptureResponse, error) {
	return nil, fmt.Errorf("Windows Biometric Framework is not supported on this platform")
}
