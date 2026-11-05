package updater

import "testing"

func TestMinSupportedVersionForcesUpdate(t *testing.T) {
	cleanCurr := "1.0.0"
	minVer := "1.2.0"
	if !isVersionNewer(cleanCurr, minVer) {
		t.Fatal("expected current version to be below min supported")
	}
}

func TestLaunchArgsRoundTrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)

	PersistLaunchArgs([]string{"prod", "extra"})
	args := LoadLaunchArgs()
	if len(args) != 2 || args[0] != "prod" || args[1] != "extra" {
		t.Fatalf("unexpected args: %#v", args)
	}
}
