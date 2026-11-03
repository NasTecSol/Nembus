package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NasTecSol/nembus-biometric-agent/installer"
	"github.com/NasTecSol/nembus-biometric-agent/setup"
	"github.com/NasTecSol/nembus-biometric-agent/tray"
	"github.com/NasTecSol/nembus-core/biometrics"
)

// Version is injected at release build time (-ldflags "-X main.Version=...").
var Version = "dev"

func main() {
	flagSetup := flag.Bool("setup", false, "Start bridge and immediately open Setup & Diagnostics dashboard in browser")
	flagInstall := flag.Bool("install", false, "Install Desktop/Start Menu shortcuts, auto-start, and copy agent to LocalAppData\\Programs\\Nembus")
	flagUninstall := flag.Bool("uninstall", false, "Remove shortcuts, auto-start, install folder, and runtime cache")
	flagUninstallDrivers := flag.Bool("uninstall-remove-drivers", false, "With --uninstall, also remove DigitalPersona driver packages (admin)")
	flagInstallDrivers := flag.Bool("install-drivers", false, "Install DigitalPersona USB drivers (prompts UAC if needed)")
	flagInstallWorker := flag.Bool("install-drivers-worker", false, "Internal elevated worker for driver installation")
	flagWinBioCaptureWorker := flag.Bool("winbio-capture-worker", false, "Internal elevated worker for WBF fingerprint capture")
	flagWinBioUnitID := flag.Uint("unit-id", 0, "WinBio unit id for capture worker")
	flagWinBioTimeoutMs := flag.Int("timeout-ms", 15000, "Capture timeout for capture worker")
	flagWinBioResultFile := flag.String("result-file", "", "JSON result path for capture worker")
	flagStatus := flag.Bool("status", false, "Print system driver and device status as JSON and exit")
	flagHeadless := flag.Bool("headless", false, "Run bridge service without system tray GUI")
	flagPort := flag.String("port", "5050", "Port to bind local hardware bridge (default: 5050)")
	flag.Parse()

	if *flagWinBioCaptureWorker {
		if *flagWinBioResultFile == "" || *flagWinBioUnitID == 0 {
			log.Fatalf("[BiometricAgent] --winbio-capture-worker requires --unit-id and --result-file")
		}
		if err := installer.RunWinBioCaptureWorker(uint32(*flagWinBioUnitID), *flagWinBioTimeoutMs, *flagWinBioResultFile); err != nil {
			log.Fatalf("[BiometricAgent] WinBio capture worker failed: %v", err)
		}
		os.Exit(0)
	}

	if *flagInstall {
		icon := tray.GenerateBiometricIcon(tray.StateReady, 48)
		res, err := installer.InstallApplication(icon)
		if err != nil {
			log.Fatalf("[BiometricAgent] Install failed: %v", err)
		}
		log.Printf("[BiometricAgent] Installed to %s", res.ExePath)
		log.Printf("[BiometricAgent] Desktop shortcut: %s", res.DesktopLNK)
		log.Printf("[BiometricAgent] Start Menu shortcut: %s", res.StartLNK)
		installer.ShowInstallSuccess(res)
		os.Exit(0)
	}

	if *flagUninstall {
		if err := installer.UninstallApplication(*flagUninstallDrivers); err != nil {
			log.Fatalf("[BiometricAgent] Uninstall failed: %v", err)
		}
		log.Printf("[BiometricAgent] Uninstalled successfully.")
		os.Exit(0)
	}

	// 1. Worker Mode: Elevated Driver Installation
	if *flagInstallWorker {
		log.Printf("[BiometricAgent] Running elevated driver installation worker...")
		if err := installer.InstallDrivers(); err != nil {
			log.Fatalf("[BiometricAgent] Elevated driver installation failed: %v", err)
		}
		log.Printf("[BiometricAgent] Drivers installed successfully.")
		os.Exit(0)
	}

	// 2. Install Drivers Flag
	if *flagInstallDrivers {
		log.Printf("[BiometricAgent] Starting driver installation...")
		if err := installer.InstallDrivers(); err != nil {
			log.Fatalf("[BiometricAgent] Driver installation failed: %v", err)
		}
		log.Printf("[BiometricAgent] Driver setup completed.")
		os.Exit(0)
	}

	// 3. Status Dump
	if *flagStatus {
		_ = biometrics.InitRuntime()
		driverStat := installer.GetStatus()
		devices, _ := biometrics.QueryDevices()
		output := map[string]any{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   Version,
			"driver":    driverStat,
			"devices":   devices,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(output)
		os.Exit(0)
	}

	if err := biometrics.InitRuntime(); err != nil {
		log.Fatalf("[BiometricAgent] Failed to initialize DigitalPersona runtime: %v", err)
	}
	biometrics.WinBioElevatedCaptureFunc = installer.RunElevatedWinBioCapture
	log.Printf("[BiometricAgent] Nembus Biometric Agent %s — runtime at %s", Version, biometrics.RuntimeDir())

	// 4. Initialize Local Biometric Bridge Server
	addr := fmt.Sprintf("127.0.0.1:%s", *flagPort)
	bridge := biometrics.NewLocalBridgeServer(addr)
	setup.AttachHandlers(bridge)

	started, err := bridge.StartIfPortFree()
	if err != nil {
		log.Fatalf("[BiometricAgent] Failed to start local bridge on %s: %v", addr, err)
	}
	if !started {
		log.Fatalf("[BiometricAgent] Cannot start: %s is already in use by another process. Stop that process or choose --port.", addr)
	}
	log.Printf("[BiometricAgent] Hardware bridge active on http://%s", addr)
	log.Printf("[BiometricAgent] Device setup dashboard available at http://%s/setup", addr)

	setupURL := fmt.Sprintf("http://%s/setup", addr)

	if *flagSetup {
		go func() {
			time.Sleep(300 * time.Millisecond)
			tray.OpenBrowser(setupURL)
		}()
	}

	shutdownBridge := func() {
		log.Printf("[BiometricAgent] Shutting down bridge...")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = bridge.Stop(ctx)
	}

	// 5. Run Mode: Tray vs Headless
	if *flagHeadless {
		log.Printf("[BiometricAgent] Running in headless daemon mode (Ctrl+C to stop)...")
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		shutdownBridge()
		return
	}

	tray.Run(tray.Config{
		SetupURL:   setupURL,
		BridgeAddr: addr,
		OnExit:     shutdownBridge,
		OnUninstall: func(removeDrivers bool) {
			shutdownBridge()
			if err := installer.UninstallApplication(removeDrivers); err != nil {
				log.Printf("[BiometricAgent] Uninstall error: %v", err)
			}
			os.Exit(0)
		},
	})
}
