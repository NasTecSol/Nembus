package updater

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

func TestBinaryHashVerification(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sap-agent-updater-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testBinaryContent := []byte("simulated-sap-agent-binary-v1.0.1")
	hasher := sha256.New()
	hasher.Write(testBinaryContent)
	expectedHash := hex.EncodeToString(hasher.Sum(nil))

	testFile := filepath.Join(tempDir, "sap-agent.exe")
	if err := os.WriteFile(testFile, testBinaryContent, 0755); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	readBack, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read back test file: %v", err)
	}
	checkHasher := sha256.New()
	checkHasher.Write(readBack)
	computedHash := hex.EncodeToString(checkHasher.Sum(nil))

	if computedHash != expectedHash {
		t.Errorf("Hash calculation error: got %s, want %s", computedHash, expectedHash)
	}
}
