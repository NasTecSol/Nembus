//go:build windows

package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/NasTecSol/nembus-core/biometrics"
)

// RunElevatedWinBioCapture runs a one-shot elevated worker that performs WBF capture.
func RunElevatedWinBioCapture(unitID uint32, timeoutMs int) (*biometrics.CaptureResponse, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve executable: %w", err)
	}
	resultFile := filepath.Join(os.TempDir(), fmt.Sprintf("nembus-winbio-cap-%d.json", os.Getpid()))
	_ = os.Remove(resultFile)

	args := fmt.Sprintf(
		`--winbio-capture-worker --unit-id=%d --timeout-ms=%d --result-file="%s"`,
		unitID,
		timeoutMs,
		resultFile,
	)
	if err := runElevated(exePath, args); err != nil {
		return nil, fmt.Errorf("elevated WBF capture: %w", err)
	}

	deadline := time.Now().Add(time.Duration(timeoutMs+20000) * time.Millisecond)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(resultFile); err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	raw, err := os.ReadFile(resultFile)
	if err != nil {
		return nil, fmt.Errorf("elevated capture did not produce a result (approve UAC if prompted): %w", err)
	}
	var payload struct {
		Success bool                        `json:"success"`
		Message string                      `json:"message"`
		Data    *biometrics.CaptureResponse   `json:"data"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("parse elevated capture result: %w", err)
	}
	if !payload.Success || payload.Data == nil {
		if payload.Message != "" {
			return nil, fmt.Errorf("%s", payload.Message)
		}
		return nil, fmt.Errorf("elevated WBF capture failed")
	}
	return payload.Data, nil
}

// RunWinBioCaptureWorker is invoked by the elevated --winbio-capture-worker process.
func RunWinBioCaptureWorker(unitID uint32, timeoutMs int, resultFile string) error {
	_ = biometrics.InitRuntime()
	// System-pool WinBio capture requires the calling process to own foreground
	// focus. The elevated worker has its own process, so it must show the same
	// capture dialog instead of calling the direct, no-dialog variant.
	res, err := biometrics.CaptureWinBioSample(unitID, timeoutMs)
	payload := map[string]any{"success": err == nil, "message": ""}
	if err != nil {
		payload["message"] = err.Error()
	} else {
		payload["data"] = res
	}
	raw, _ := json.Marshal(payload)
	return os.WriteFile(resultFile, raw, 0600)
}
