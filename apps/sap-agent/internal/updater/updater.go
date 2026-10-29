package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// UpdateInfo contains update status and release details for SAP Agent
type UpdateInfo struct {
	HasUpdate           bool      `json:"hasUpdate"`
	CurrentVersion      string    `json:"currentVersion"`
	LatestVersion       string    `json:"latestVersion"`
	ReleaseNotes        string    `json:"releaseNotes"`
	DownloadURL         string    `json:"downloadUrl"`
	ReleaseID           int32     `json:"releaseId"`
	FileName            string    `json:"fileName"`
	FileSizeBytes       int64     `json:"fileSizeBytes"`
	ChecksumSHA256      string    `json:"checksumSha256"`
	IsMandatory         bool      `json:"isMandatory"`
	MinSupportedVersion string    `json:"minSupportedVersion"`
	CreatedAt           time.Time `json:"createdAt"`
}

// CheckForUpdate queries the internal Nembus Cloud Server for newer versions of SAP Agent.
// When offline or unreachable, it gracefully logs and returns HasUpdate = false.
func CheckForUpdate(currentVersion, cloudURL, clientKey string) (*UpdateInfo, error) {
	if cloudURL == "" {
		cloudURL = "https://nembus.nashrms.com"
	}
	baseURL := strings.TrimRight(cloudURL, "/")
	checkURL := fmt.Sprintf("%s/api/v1/updates/sap-agent/latest?current_version=%s", baseURL, url.QueryEscape(currentVersion))

	req, err := http.NewRequest("GET", checkURL, nil)
	if err != nil {
		return &UpdateInfo{HasUpdate: false, CurrentVersion: currentVersion}, fmt.Errorf("failed to create update request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Nembus-SAP-Agent-Updater")
	if clientKey != "" {
		req.Header.Set("X-Nembus-App-Key", clientKey)
	}

	// 5-second timeout for graceful non-blocking check
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[SAP-UPDATER] Cloud server unreachable, skipping update check: %v", err)
		return &UpdateInfo{HasUpdate: false, CurrentVersion: currentVersion}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &UpdateInfo{HasUpdate: false, CurrentVersion: currentVersion}, fmt.Errorf("update check returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var info UpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return &UpdateInfo{HasUpdate: false, CurrentVersion: currentVersion}, fmt.Errorf("failed to parse update response: %w", err)
	}

	// Resolve relative download URL against cloudURL
	if info.DownloadURL != "" && !strings.HasPrefix(info.DownloadURL, "http://") && !strings.HasPrefix(info.DownloadURL, "https://") {
		info.DownloadURL = baseURL + info.DownloadURL
	}

	return &info, nil
}

// ApplyUpdate downloads the new release binary, validates SHA-256 strictly,
// replaces the executable on disk, and handles process/service restart.
func ApplyUpdate(downloadURL, expectedChecksum, clientKey string, isWindowsService bool, serviceName string) error {
	if downloadURL == "" {
		return fmt.Errorf("no valid download URL provided")
	}

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("User-Agent", "Nembus-SAP-Agent-Updater")
	if clientKey != "" {
		req.Header.Set("X-Nembus-App-Key", clientKey)
	}

	// Allow up to 30 minutes for large downloads
	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download update binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed with HTTP %d: %s", resp.StatusCode, string(body))
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to resolve current executable path: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks for executable: %w", err)
	}

	dir := filepath.Dir(exePath)
	newExePath := filepath.Join(dir, "sap-agent-update.tmp")
	oldExePath := exePath + ".old"

	out, err := os.Create(newExePath)
	if err != nil {
		return fmt.Errorf("failed to create temporary update file: %w", err)
	}

	hasher := sha256.New()
	multiWriter := io.MultiWriter(out, hasher)
	_, err = io.Copy(multiWriter, resp.Body)
	closeErr := out.Close()
	if err != nil {
		_ = os.Remove(newExePath)
		return fmt.Errorf("failed to write update content: %w", err)
	}
	if closeErr != nil {
		_ = os.Remove(newExePath)
		return fmt.Errorf("failed to flush update file to disk: %w", closeErr)
	}

	// STRICT SHA-256 CRYPTOGRAPHIC INTEGRITY CHECK
	computedChecksum := hex.EncodeToString(hasher.Sum(nil))
	if expectedChecksum != "" && !strings.EqualFold(expectedChecksum, computedChecksum) {
		_ = os.Remove(newExePath)
		return fmt.Errorf("binary integrity violation: SHA-256 mismatch (expected: %s, got: %s)", expectedChecksum, computedChecksum)
	}

	_ = os.Chmod(newExePath, 0755)
	_ = os.Remove(oldExePath)

	// Atomic rename of running binary to .old backup
	if err := os.Rename(exePath, oldExePath); err != nil {
		_ = os.Remove(newExePath)
		return fmt.Errorf("failed to backup current executable: %w", err)
	}

	// Place updated binary into production path
	if err := os.Rename(newExePath, exePath); err != nil {
		_ = os.Rename(oldExePath, exePath)
		return fmt.Errorf("failed to replace executable: %w", err)
	}

	log.Println("[SAP-UPDATER] Update applied and checksum verified successfully!")

	// Restart sequence based on execution environment
	if isWindowsService && runtime.GOOS == "windows" {
		if serviceName == "" {
			serviceName = "NembusSAPMigrationAgent"
		}
		log.Printf("[SAP-UPDATER] Restarting Windows Service '%s'...", serviceName)
		// Spawn detached helper to restart the Windows Service
		restartCmd := exec.Command("cmd.exe", "/C", fmt.Sprintf("net stop %s && net start %s", serviceName, serviceName))
		_ = restartCmd.Start()
		time.Sleep(300 * time.Millisecond)
		os.Exit(0)
		return nil
	}

	// Standalone interactive process: spawn and exit
	cmd := exec.Command(exePath, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		// Rollback if new process fails to start
		_ = os.Rename(oldExePath, exePath)
		return fmt.Errorf("failed to launch updated agent application: %w", err)
	}

	log.Println("[SAP-UPDATER] Updated agent process launched. Terminating old instance...")
	time.Sleep(250 * time.Millisecond)
	os.Exit(0)
	return nil
}

// CleanupOldExecutables removes legacy .old executable backups left behind from prior updates
func CleanupOldExecutables() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	oldExePath := exePath + ".old"
	if _, err := os.Stat(oldExePath); err == nil {
		_ = os.Remove(oldExePath)
		log.Println("[SAP-UPDATER] Cleaned up legacy executable backup:", oldExePath)
	}
}

// IsVersionNewer compares semver strings without 'v' prefix
func IsVersionNewer(curr, latest string) bool {
	if curr == "" {
		return true
	}
	currParts := ParseVersion(curr)
	latestParts := ParseVersion(latest)

	for i := 0; i < len(currParts) || i < len(latestParts); i++ {
		var c, l int
		if i < len(currParts) {
			c = currParts[i]
		}
		if i < len(latestParts) {
			l = latestParts[i]
		}
		if l > c {
			return true
		}
		if l < c {
			return false
		}
	}
	return false
}

// ParseVersion extracts integer components of semver string
func ParseVersion(v string) []int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	res := make([]int, 0, len(parts))
	for _, p := range parts {
		var num int
		num, _ = strconv.Atoi(p)
		res = append(res, num)
	}
	return res
}
