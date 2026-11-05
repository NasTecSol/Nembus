//go:build !windows

package installer

import "github.com/NasTecSol/nembus-core/biometrics"

func RunElevatedWinBioCapture(unitID uint32, timeoutMs int) (*biometrics.CaptureResponse, error) {
	return nil, nil
}

func RunWinBioCaptureWorker(unitID uint32, timeoutMs int, resultFile string) error {
	return nil
}
