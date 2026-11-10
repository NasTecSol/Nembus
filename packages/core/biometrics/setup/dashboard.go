package setup

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/NasTecSol/nembus-core/biometrics/installer"
	"github.com/NasTecSol/nembus-core/biometrics"
)

//go:embed static/*
var staticFS embed.FS

// AttachHandlers registers the setup dashboard and diagnostic endpoints to the local bridge server.
func AttachHandlers(bridge *biometrics.LocalBridgeServer) {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Printf("[SetupDashboard] Failed to load static dashboard: %v", err)
		return
	}
	fileServer := http.FileServer(http.FS(sub))

	// Serve setup UI at /setup and /setup/
	bridge.HandleFunc("/setup", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/setup/", http.StatusMovedPermanently)
	})
	bridge.Mux().Handle("/setup/", http.StripPrefix("/setup/", fileServer))

	// JSON Diagnostic endpoints
	bridge.HandleFunc("/api/v1/setup/status", handleSetupStatus)
	bridge.HandleFunc("/api/v1/setup/install-drivers", handleInstallDrivers)
}

func handleSetupStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}

	driverStat := installer.GetStatus()
	diag := biometrics.GetCaptureDiagnostics()

	devices, queryErr := biometrics.QueryDevices()
	connected := len(devices) > 0
	var primaryDev map[string]any
	if connected {
		dev := devices[0]
		model := strings.TrimSpace(dev.VendorName + " " + dev.ProductName)
		if model == "" {
			model = "DigitalPersona U.are.U Fingerprint Reader"
		}
		primaryDev = map[string]any{
			"connected":         true,
			"model":             model,
			"serial_number":     dev.SerialNum,
			"vendor_id":         fmt.Sprintf("0x%04X", dev.VendorID),
			"product_id":        fmt.Sprintf("0x%04X", dev.ProductID),
			"device_id":         dev.Name,
			"firmware_version":  dev.FirmwareVersion,
			"hardware_version":  dev.HardwareVersion,
		}
	} else {
		primaryDev = map[string]any{
			"connected": false,
		}
		if queryErr != nil {
			primaryDev["error"] = queryErr.Error()
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":     true,
		"driver":      driverStat,
		"device":      primaryDev,
		"diagnostics": diag,
	})
}

func handleInstallDrivers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}

	go func() {
		if err := installer.InstallDrivers(); err != nil {
			log.Printf("[SetupDashboard] Driver installation returned error: %v", err)
		}
	}()

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Driver installation process launched. If prompted, please approve the Windows User Account Control (UAC) dialog.",
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

