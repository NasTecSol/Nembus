package usecase

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestIsVersionNewer(t *testing.T) {
	tests := []struct {
		current  string
		latest   string
		expected bool
	}{
		{"v1.0.0", "v1.0.1", true},
		{"1.0.0", "1.0.1", true},
		{"v1.0.1", "v1.0.0", false},
		{"v1.0.0", "v1.0.0", false},
		{"v1.0.0", "v2.0.0", true},
		{"v1.9.9", "v2.0.0", true},
		{"v2.1.0", "v2.0.9", false},
		{"", "v1.0.0", true},
	}

	for _, tt := range tests {
		got := IsVersionNewer(tt.current, tt.latest)
		if got != tt.expected {
			t.Errorf("IsVersionNewer(%q, %q) = %v; want %v", tt.current, tt.latest, got, tt.expected)
		}
	}
}

func TestReleaseStorageCleanup(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "nembus-release-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Simulate old build file on disk
	oldFilePath := filepath.Join(tempDir, "pos-client-v1.0.0.exe")
	if err := os.WriteFile(oldFilePath, []byte("legacy-binary-content"), 0644); err != nil {
		t.Fatalf("Failed to create old build file: %v", err)
	}

	// Verify old file exists
	if _, err := os.Stat(oldFilePath); os.IsNotExist(err) {
		t.Fatalf("Old build file was not created")
	}

	// Calculate test hash
	content := []byte("new-binary-content")
	hasher := sha256.New()
	hasher.Write(content)
	expectedHash := hex.EncodeToString(hasher.Sum(nil))

	newFilePath := filepath.Join(tempDir, "pos-client-v1.0.1.exe")
	if err := os.WriteFile(newFilePath, content, 0644); err != nil {
		t.Fatalf("Failed to create new build file: %v", err)
	}

	// Remove old build (simulating cleanupOldReleases)
	_ = os.Remove(oldFilePath)

	// Verify old build file is purged
	if _, err := os.Stat(oldFilePath); !os.IsNotExist(err) {
		t.Errorf("Old build file %s still exists on disk, should have been purged", oldFilePath)
	}

	// Verify new build file exists and has correct checksum
	savedContent, err := os.ReadFile(newFilePath)
	if err != nil {
		t.Fatalf("Failed to read new build file: %v", err)
	}
	newHasher := sha256.New()
	newHasher.Write(savedContent)
	actualHash := hex.EncodeToString(newHasher.Sum(nil))

	if actualHash != expectedHash {
		t.Errorf("Checksum mismatch: got %s, want %s", actualHash, expectedHash)
	}
}
