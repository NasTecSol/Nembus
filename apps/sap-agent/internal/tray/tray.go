package tray

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"

	"github.com/getlantern/systray"
)

// StatusSnapshot holds the live health summary rendered in the tray status bar.
type StatusSnapshot struct {
	Kind    StatusKind
	Label   string
	Tooltip string
}

// Config configures the SAP Agent Windows System Tray lifecycle.
type Config struct {
	BaseURL       string
	Version       string
	GetStatus     func() StatusSnapshot
	OnCheckUpdate func() (statusMsg string, openURL string, hasUpdate bool)
	OnExit        func()
}

// Run starts the system tray event loop and blocks until Quit is selected.
func Run(cfg Config) {
	systray.Run(func() {
		onReady(cfg)
	}, func() {
		if cfg.OnExit != nil {
			cfg.OnExit()
		}
	})
}

// Quit terminates the system tray cleanly (e.g. right before self-update restart).
func Quit() {
	systray.Quit()
}

// OpenBrowser opens url in the user's default browser.
func OpenBrowser(url string) {
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

func onReady(cfg Config) {
	systray.SetIcon(GenerateStatusIcon(StatusStandby))
	systray.SetTitle("Nembus SAP Agent")
	systray.SetTooltip(fmt.Sprintf("Nembus SAP Agent %s", cfg.Version))

	// Live App Status Bar (non-clickable header)
	mStatus := systray.AddMenuItem(fmt.Sprintf("Status: Starting (%s)...", cfg.Version), "Current SAP Agent synchronization and health status")
	mStatus.Disable()
	systray.AddSeparator()

	// 3 Specific Action Buttons
	mControlPanel := systray.AddMenuItem("Control Panel", "Open Nembus SAP Agent Executive Control Panel")
	mConfig := systray.AddMenuItem("Configurations", "Open SAP Agent Configurations & Settings")
	mCheckUpdate := systray.AddMenuItem("Check for Update", "Check Cloud Server for a newer SAP Agent release")

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Stop Nembus SAP Agent")

	refreshStatus := func() {
		if cfg.GetStatus == nil {
			return
		}
		snap := cfg.GetStatus()
		systray.SetIcon(GenerateStatusIcon(snap.Kind))
		if snap.Label != "" {
			mStatus.SetTitle(snap.Label)
		}
		if snap.Tooltip != "" {
			systray.SetTooltip(snap.Tooltip)
		}
	}
	refreshStatus()

	// Live status polling loop
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			refreshStatus()
		}
	}()

	// Menu click handler loop
	go func() {
		for {
			select {
			case <-mControlPanel.ClickedCh:
				OpenBrowser(cfg.BaseURL + "/#tab-dashboard")
			case <-mConfig.ClickedCh:
				OpenBrowser(cfg.BaseURL + "/#tab-config")
			case <-mCheckUpdate.ClickedCh:
				go func() {
					mStatus.SetTitle("Status: Checking for updates...")
					if cfg.OnCheckUpdate != nil {
						msg, openURL, hasUpdate := cfg.OnCheckUpdate()
						if msg != "" {
							mStatus.SetTitle(msg)
						}
						if hasUpdate {
							systray.SetIcon(GenerateStatusIcon(StatusUpdate))
						}
						if openURL != "" {
							OpenBrowser(openURL)
						}
						time.Sleep(8 * time.Second)
						refreshStatus()
					} else {
						OpenBrowser(cfg.BaseURL + "/#check-update")
					}
				}()
			case <-mQuit.ClickedCh:
				log.Println("[SAP-AGENT] Quit requested from System Tray.")
				systray.Quit()
				return
			}
		}
	}()
}
