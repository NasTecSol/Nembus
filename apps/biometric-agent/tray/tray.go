package tray

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/NasTecSol/nembus-biometric-agent/installer"
	"github.com/NasTecSol/nembus-core/biometrics"
	"github.com/getlantern/systray"
)

var (
	currentState AgentState = StateDisconnected
	stateMu      sync.Mutex
)

// Config configures the system tray lifecycle.
type Config struct {
	SetupURL   string
	BridgeAddr string
	OnExit     func()
	OnUninstall func(removeDrivers bool)
}

// Run starts the system tray event loop. This blocks until systray.Quit() is called.
func Run(cfg Config) {
	setupURL := cfg.SetupURL
	bridgeAddr := cfg.BridgeAddr
	if bridgeAddr == "" {
		bridgeAddr = "127.0.0.1:5050"
	}
	onExit := cfg.OnExit
	onUninstall := cfg.OnUninstall

	systray.Run(func() {
		onReady(setupURL, bridgeAddr, onUninstall)
	}, func() {
		if onExit != nil {
			onExit()
		}
		onTrayExit()
	})
}

func onReady(setupURL, bridgeAddr string, onUninstall func(removeDrivers bool)) {
	SetState(StateDisconnected)

	mSetup := systray.AddMenuItem("🛠️ Device Setup & Diagnostics...", "Open interactive hardware manager and setup")
	mInstall := systray.AddMenuItem("⚡ Install Drivers (Admin)...", "Stage WHQL DigitalPersona kernel drivers via UAC")
	mTest := systray.AddMenuItem("🧪 Test Fingerprint Scanner...", "Calibrate and verify biometric capture")
	systray.AddSeparator()
	mShortcuts := systray.AddMenuItem("📌 Install Desktop & Start Menu Icons...", "Create shortcuts with Nembus icon and enable sign-in auto-start")
	systray.AddSeparator()
	mCloud := systray.AddMenuItem("🌐 Open Cloud Backoffice...", "Launch Nembus Cloud ERP Backoffice")
	systray.AddSeparator()
	mUninstall := systray.AddMenuItem("🗑️ Uninstall Nembus Biometric Agent...", "Remove shortcuts, auto-start, and local agent files")
	mQuit := systray.AddMenuItem("❌ Exit", "Stop local biometric bridge")

	go func() {
		for {
			select {
			case <-mSetup.ClickedCh:
				OpenBrowser(setupURL)
			case <-mInstall.ClickedCh:
				go func() {
					if err := installer.InstallDrivers(); err != nil {
						log.Printf("[Tray] Driver installation failed: %v", err)
					}
				}()
			case <-mTest.ClickedCh:
				OpenBrowser(setupURL)
			case <-mShortcuts.ClickedCh:
				go func() {
					icon := GenerateBiometricIcon(StateReady, 48)
					res, err := installer.InstallApplication(icon)
					if err != nil {
						log.Printf("[Tray] Shortcut install failed: %v", err)
						return
					}
					installer.ShowInstallSuccess(res)
				}()
			case <-mCloud.ClickedCh:
				OpenBrowser("https://backoffice.nembus.com")
			case <-mUninstall.ClickedCh:
				go func() {
					if !installer.ConfirmUninstall() {
						return
					}
					removeDrivers := installer.ConfirmRemoveDrivers()
					if onUninstall != nil {
						onUninstall(removeDrivers)
						return
					}
					_ = installer.UninstallApplication(removeDrivers)
					systray.Quit()
				}()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if biometrics.HardwareCaptureActive() {
				SetState(StateCapturing)
				systray.SetTooltip("Nembus Biometrics: Capturing fingerprint...")
				continue
			}

			status := installer.GetStatus()
			if status.NeedsDriverRepair {
				SetState(StateError)
				systray.SetTooltip(status.Message)
				continue
			}
			if !status.DLLsAvailable {
				SetState(StateError)
				systray.SetTooltip("Nembus Biometrics: Runtime DLLs not loaded")
				continue
			}
			if len(status.USBReaders) == 0 && status.SDKDeviceCount == 0 {
				SetState(StateDisconnected)
				systray.SetTooltip("Nembus Biometrics: No USB reader detected")
				continue
			}

			devices, err := biometrics.QueryDevices()
			if err != nil || len(devices) == 0 {
				SetState(StateDisconnected)
				systray.SetTooltip("Nembus Biometrics: No reader connected")
			} else {
				SetState(StateReady)
				systray.SetTooltip(fmt.Sprintf("Nembus Biometrics: Ready (%s)", bridgeAddr))
			}
		}
	}()
}

func onTrayExit() {
	log.Printf("[Tray] Exiting biometric agent...")
}

// SetState updates the tray icon and state dynamically using the custom biometric renderer.
func SetState(state AgentState) {
	stateMu.Lock()
	defer stateMu.Unlock()
	currentState = state

	iconBytes := GenerateBiometricIcon(state, 32)
	if len(iconBytes) > 0 {
		systray.SetIcon(iconBytes)
	}
}

// OpenBrowser launches the given URL in the user's default web browser.
func OpenBrowser(url string) {
	if url == "" {
		return
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
