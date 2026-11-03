//go:build !windows

package installer

// AppInstallResult summarizes a non-Windows install attempt.
type AppInstallResult struct {
	InstallDir string `json:"install_dir"`
	ExePath    string `json:"exe_path"`
	DesktopLNK string `json:"desktop_shortcut,omitempty"`
	StartLNK   string `json:"start_menu_shortcut,omitempty"`
}

// InstallApplication is a no-op on non-Windows platforms.
func InstallApplication(iconICO []byte) (AppInstallResult, error) {
	_ = iconICO
	return AppInstallResult{}, nil
}

// UninstallApplication is a no-op on non-Windows platforms.
func UninstallApplication(removeDrivers bool) error {
	_ = removeDrivers
	return nil
}

// ShowInstallSuccess is a no-op off Windows.
func ShowInstallSuccess(_ AppInstallResult) {}

// ConfirmUninstall always returns false on non-Windows.
func ConfirmUninstall() bool {
	return false
}

func ConfirmRemoveDrivers() bool {
	return false
}
