//go:build windows

package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	appProductName   = "Nembus Biometric Agent"
	shortcutBaseName = "Nembus Biometric Agent"
	runRegistryName  = "NembusBiometricAgent"
)

// AppInstallResult summarizes what was created during installation.
type AppInstallResult struct {
	InstallDir string `json:"install_dir"`
	ExePath    string `json:"exe_path"`
	IconPath   string `json:"icon_path,omitempty"`
	DesktopLNK string `json:"desktop_shortcut,omitempty"`
	StartLNK   string `json:"start_menu_shortcut,omitempty"`
}

// InstallApplication copies the agent into the user Programs folder, writes shortcuts
// (Desktop + Start Menu) with the supplied ICO, and registers logon auto-start.
func InstallApplication(iconICO []byte) (AppInstallResult, error) {
	srcExe, err := os.Executable()
	if err != nil {
		return AppInstallResult{}, fmt.Errorf("resolve executable: %w", err)
	}
	srcExe, _ = filepath.Abs(srcExe)

	installDir := filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Nembus")
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return AppInstallResult{}, fmt.Errorf("create install dir: %w", err)
	}

	destExe := filepath.Join(installDir, "nembus-biometric-agent.exe")
	if !strings.EqualFold(filepath.Clean(srcExe), filepath.Clean(destExe)) {
		if err := copyFile(srcExe, destExe); err != nil {
			return AppInstallResult{}, fmt.Errorf("copy agent binary: %w", err)
		}
	}

	iconPath := filepath.Join(installDir, "nembus-biometric-agent.ico")
	if len(iconICO) > 0 {
		_ = os.WriteFile(iconPath, iconICO, 0644)
	} else if _, statErr := os.Stat(iconPath); statErr != nil {
		iconPath = destExe
	}

	desktopLNK, err := createShortcut(
		filepath.Join(os.Getenv("USERPROFILE"), "Desktop", shortcutBaseName+".lnk"),
		destExe, "", installDir, iconPath,
	)
	if err != nil {
		return AppInstallResult{}, err
	}

	startDir := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs")
	if err := os.MkdirAll(startDir, 0755); err != nil {
		return AppInstallResult{}, err
	}
	startLNK, err := createShortcut(
		filepath.Join(startDir, shortcutBaseName+".lnk"),
		destExe, "", installDir, iconPath,
	)
	if err != nil {
		return AppInstallResult{}, err
	}

	if err := setAutoStart(destExe); err != nil {
		return AppInstallResult{}, err
	}

	return AppInstallResult{
		InstallDir: installDir,
		ExePath:    destExe,
		IconPath:   iconPath,
		DesktopLNK: desktopLNK,
		StartLNK:   startLNK,
	}, nil
}

// UninstallApplication stops the agent, removes shortcuts, auto-start, runtime cache, and install dir.
func UninstallApplication(removeDrivers bool) error {
	_ = exec.Command("taskkill", "/F", "/IM", "nembus-biometric-agent.exe").Run()

	removeShortcut(filepath.Join(os.Getenv("USERPROFILE"), "Desktop", shortcutBaseName+".lnk"))
	startDir := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs")
	removeShortcut(filepath.Join(startDir, shortcutBaseName+".lnk"))

	_ = removeAutoStart()

	runtimeDir := filepath.Join(os.Getenv("LOCALAPPDATA"), "Nembus", "biometrics")
	_ = os.RemoveAll(runtimeDir)

	installDir := filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Nembus")
	_ = os.RemoveAll(installDir)

	if removeDrivers {
		_, _ = runHiddenPowerShell(`
$ErrorActionPreference = 'SilentlyContinue'
$drivers = pnputil.exe /enum-drivers
$oemList = @()
$currentOem = ''
foreach ($line in ($drivers -split [Environment]::NewLine)) {
  if ($line -match 'Published Name:\s+(oem\d+\.inf)') { $currentOem = $matches[1] }
  elseif ($line -match 'dpersona' -and $currentOem) { $oemList += $currentOem; $currentOem = '' }
}
($oemList | Select-Object -Unique) | ForEach-Object { pnputil.exe /delete-driver $_ /uninstall /force | Out-Null }
`)
	}
	return nil
}

// ConfirmUninstall shows a Yes/No dialog before removal.
func ConfirmUninstall() bool {
	out, err := runHiddenPowerShell(`
Add-Type -AssemblyName System.Windows.Forms
$r = [System.Windows.Forms.MessageBox]::Show(
  'This removes Nembus Biometric Agent shortcuts, auto-start, the local install folder, and extracted DigitalPersona runtime files. USB drivers are kept unless you uninstall with --remove-drivers.',
  'Uninstall Nembus Biometric Agent?',
  [System.Windows.Forms.MessageBoxButtons]::YesNo,
  [System.Windows.Forms.MessageBoxIcon]::Warning
)
if ($r -eq 'Yes') { 'YES' } else { 'NO' }
`)
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToUpper(string(out)), "YES")
}

// ConfirmRemoveDrivers asks whether to remove DigitalPersona driver packages.
func ConfirmRemoveDrivers() bool {
	out, err := runHiddenPowerShell(`
Add-Type -AssemblyName System.Windows.Forms
$r = [System.Windows.Forms.MessageBox]::Show(
  'Also remove DigitalPersona USB driver packages from Windows? Choose Yes only if no other PCs/apps need the legacy driver on this machine.',
  'Remove USB drivers?',
  [System.Windows.Forms.MessageBoxButtons]::YesNo,
  [System.Windows.Forms.MessageBoxIcon]::Question
)
if ($r -eq 'Yes') { 'YES' } else { 'NO' }
`)
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToUpper(string(out)), "YES")
}

// ConfirmUninstallWithDrivers is deprecated; use ConfirmUninstall + ConfirmRemoveDrivers.
func ConfirmUninstallWithDrivers() (proceed bool, removeDrivers bool) {
	if !ConfirmUninstall() {
		return false, false
	}
	return true, ConfirmRemoveDrivers()
}

func createShortcut(lnkPath, target, args, workDir, iconPath string) (string, error) {
	escaped := func(s string) string { return strings.ReplaceAll(s, `'`, `''`) }
	script := fmt.Sprintf(`
$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut('%s')
$Shortcut.TargetPath = '%s'
$Shortcut.Arguments = '%s'
$Shortcut.WorkingDirectory = '%s'
$Shortcut.IconLocation = '%s,0'
$Shortcut.Description = '%s'
$Shortcut.Save()
`, escaped(lnkPath), escaped(target), escaped(args), escaped(workDir), escaped(iconPath), escaped(appProductName))
	if _, err := runHiddenPowerShell(script); err != nil {
		return "", fmt.Errorf("create shortcut %s: %w", lnkPath, err)
	}
	return lnkPath, nil
}

func removeShortcut(lnkPath string) {
	_ = os.Remove(lnkPath)
	publicDesktop := filepath.Join(os.Getenv("PUBLIC"), "Desktop", shortcutBaseName+".lnk")
	_ = os.Remove(publicDesktop)
}

func setAutoStart(exePath string) error {
	escaped := strings.ReplaceAll(exePath, `'`, `''`)
	script := fmt.Sprintf(`
Set-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Run' -Name '%s' -Value '%s' -Type String -Force
`, runRegistryName, escaped)
	_, err := runHiddenPowerShell(script)
	return err
}

func removeAutoStart() error {
	script := fmt.Sprintf(`
Remove-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Run' -Name '%s' -ErrorAction SilentlyContinue
`, runRegistryName)
	_, err := runHiddenPowerShell(script)
	return err
}

func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, in, 0755)
}

func showInfo(title, message string) {
	script := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
[System.Windows.Forms.MessageBox]::Show('%s', '%s', [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Information) | Out-Null
`, strings.ReplaceAll(message, `'`, `''`), strings.ReplaceAll(title, `'`, `''`))
	_, _ = runHiddenPowerShell(script)
}

// ShowInstallSuccess displays a confirmation dialog after --install.
func ShowInstallSuccess(res AppInstallResult) {
	showInfo(appProductName, fmt.Sprintf(
		"Installed to:\n%s\n\nDesktop and Start Menu shortcuts were created with the Nembus biometric icon.\nThe agent is registered to start automatically when you sign in to Windows.",
		res.ExePath,
	))
}
