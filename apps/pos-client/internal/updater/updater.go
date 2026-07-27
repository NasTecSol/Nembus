package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// UpdateInfo contains update status and release details
type UpdateInfo struct {
	HasUpdate      bool   `json:"hasUpdate"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseNotes   string `json:"releaseNotes"`
	DownloadURL    string `json:"downloadUrl"`
	AssetID        int64  `json:"assetId"`
}

type githubAsset struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	URL                string `json:"url"`
}

type githubRelease struct {
	TagName    string        `json:"tag_name"`
	Name       string        `json:"name"`
	Body       string        `json:"body"`
	Draft      bool          `json:"draft"`
	Prerelease bool          `json:"prerelease"`
	Assets     []githubAsset `json:"assets"`
}

// CheckForUpdate checks GitHub Releases API for a newer version than currentVersion
func CheckForUpdate(currentVersion, ghRepo, ghToken string) (*UpdateInfo, error) {
	if ghRepo == "" {
		ghRepo = "NasTecSol/Nembus"
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", ghRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create update request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Nembus-POS-Client-Updater")
	if ghToken != "" {
		req.Header.Set("Authorization", "Bearer "+ghToken)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api returned status %d: %s", resp.StatusCode, string(body))
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse github release response: %w", err)
	}

	latestVer := strings.TrimPrefix(release.TagName, "v")
	currVer := strings.TrimPrefix(currentVersion, "v")

	hasUpdate := isVersionNewer(currVer, latestVer)

	var downloadURL string
	var assetID int64
	expectedExt := ".exe"
	if runtime.GOOS != "windows" {
		expectedExt = ""
	}

	for _, asset := range release.Assets {
		if (expectedExt != "" && strings.HasSuffix(asset.Name, expectedExt)) || strings.Contains(asset.Name, "nembus-client") {
			downloadURL = asset.BrowserDownloadURL
			assetID = asset.ID
			break
		}
	}

	return &UpdateInfo{
		HasUpdate:      hasUpdate,
		CurrentVersion: currentVersion,
		LatestVersion:  release.TagName,
		ReleaseNotes:   release.Body,
		DownloadURL:    downloadURL,
		AssetID:        assetID,
	}, nil
}

// ApplyUpdate downloads the new release asset and replaces the running executable
func ApplyUpdate(downloadURL string, assetID int64, ghRepo string, ghToken string) error {
	var downloadReqURL string
	if ghToken != "" && assetID > 0 {
		if ghRepo == "" {
			ghRepo = "NasTecSol/Nembus"
		}
		downloadReqURL = fmt.Sprintf("https://api.github.com/repos/%s/releases/assets/%d", ghRepo, assetID)
	} else {
		downloadReqURL = downloadURL
	}

	if downloadReqURL == "" {
		return fmt.Errorf("no valid download URL or asset ID found")
	}

	req, err := http.NewRequest("GET", downloadReqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	if ghToken != "" && assetID > 0 {
		req.Header.Set("Accept", "application/octet-stream")
		req.Header.Set("Authorization", "Bearer "+ghToken)
	}
	req.Header.Set("User-Agent", "Nembus-POS-Client-Updater")

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download update asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to download asset, status %d: %s", resp.StatusCode, string(body))
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks for executable: %w", err)
	}

	dir := filepath.Dir(exePath)
	newExePath := filepath.Join(dir, "nembus-client-update.tmp")
	oldExePath := exePath + ".old"

	out, err := os.Create(newExePath)
	if err != nil {
		return fmt.Errorf("failed to create temporary update file: %w", err)
	}

	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		os.Remove(newExePath)
		return fmt.Errorf("failed to write update content: %w", err)
	}

	_ = os.Chmod(newExePath, 0755)
	_ = os.Remove(oldExePath)

	if err := os.Rename(exePath, oldExePath); err != nil {
		os.Remove(newExePath)
		return fmt.Errorf("failed to backup current executable: %w", err)
	}

	if err := os.Rename(newExePath, exePath); err != nil {
		_ = os.Rename(oldExePath, exePath)
		return fmt.Errorf("failed to replace executable: %w", err)
	}

	cmd := exec.Command(exePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start updated application: %w", err)
	}

	log.Println("[UPDATER] Update applied successfully! Restarting process...")
	os.Exit(0)
	return nil
}

// CleanupOldExecutables removes legacy .old executable files from previous updates
func CleanupOldExecutables() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	oldExePath := exePath + ".old"
	if _, err := os.Stat(oldExePath); err == nil {
		_ = os.Remove(oldExePath)
		log.Println("[UPDATER] Cleaned up legacy executable:", oldExePath)
	}
}

func isVersionNewer(curr, latest string) bool {
	currParts := parseVersion(curr)
	latestParts := parseVersion(latest)

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

func parseVersion(v string) []int {
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
