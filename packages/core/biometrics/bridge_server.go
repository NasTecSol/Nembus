package biometrics

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// LocalBridgeServer provides the local hardware HTTP API on http://127.0.0.1:5050/api/v1
// directly from the embedded DigitalPersona runtime libraries (dpfpdd + dpfj).
type LocalBridgeServer struct {
	addr       string
	mux        *http.ServeMux
	httpServer *http.Server

	captureHTTPMu sync.Mutex // one blocking hardware capture at a time (POST capture / enrollment step)

	enrollMu     sync.Mutex
	enrollViews  [][]byte
	enrollSerial string
}

// NewLocalBridgeServer creates a new embedded biometric hardware bridge server.
func NewLocalBridgeServer(addr string) *LocalBridgeServer {
	if addr == "" {
		addr = "127.0.0.1:5050"
	}
	mux := http.NewServeMux()
	b := &LocalBridgeServer{addr: addr, mux: mux}
	mux.HandleFunc("/api/v1/device", b.withCORS(b.handleDevice))
	mux.HandleFunc("/api/v1/device/info", b.withCORS(b.handleDeviceInfo))
	mux.HandleFunc("/api/v1/capture", b.withCORS(b.handleCapture))
	mux.HandleFunc("/api/v1/enrollment", b.withCORS(b.handleEnrollment))
	mux.HandleFunc("/api/v1/enrollment/start", b.withCORS(b.handleEnrollmentStart))
	mux.HandleFunc("/api/v1/enrollment/capture-step", b.withCORS(b.handleEnrollmentStep))

	b.httpServer = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return b
}

// Mux returns the underlying HTTP ServeMux to register custom or extended routes.
func (b *LocalBridgeServer) Mux() *http.ServeMux {
	return b.mux
}

// HandleFunc registers a CORS-wrapped HTTP handler on the bridge mux.
func (b *LocalBridgeServer) HandleFunc(pattern string, handler http.HandlerFunc) {
	b.mux.HandleFunc(pattern, b.withCORS(handler))
}

// StartIfPortFree starts the embedded hardware bridge when the listen address is free.
// It returns started=true when this instance is listening, or started=false when another
// process already holds the port (err is nil in that case).
func (b *LocalBridgeServer) StartIfPortFree() (started bool, err error) {
	ln, err := net.Listen("tcp", b.addr)
	if err != nil {
		log.Printf("[BiometricBridge] Port %s already in use: %v", b.addr, err)
		return false, nil
	}
	log.Printf("[BiometricBridge] Embedded plug-and-play hardware bridge listening on http://%s/api/v1", b.addr)
	go func() {
		if serveErr := b.httpServer.Serve(ln); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			log.Printf("[BiometricBridge] Server stopped: %v", serveErr)
		}
	}()
	return true, nil
}

// Addr returns the configured listen address (host:port).
func (b *LocalBridgeServer) Addr() string {
	return b.addr
}

// Stop gracefully shuts down the embedded hardware bridge.
func (b *LocalBridgeServer) Stop(ctx context.Context) error {
	_ = CancelActiveCapture()
	if b.httpServer != nil {
		return b.httpServer.Shutdown(ctx)
	}
	return nil
}

func (b *LocalBridgeServer) withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Authorization-Token")
		if r.Header.Get("Access-Control-Request-Private-Network") == "true" {
			w.Header().Set("Access-Control-Allow-Private-Network", "true")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func (b *LocalBridgeServer) handleDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeBridgeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "message": "method not allowed"})
		return
	}
	devices, err := QueryDevices()
	if err != nil || len(devices) == 0 {
		writeBridgeJSON(w, http.StatusOK, map[string]any{
			"success":   false,
			"connected": false,
			"count":     0,
			"message":   "No fingerprint reader detected",
		})
		return
	}
	writeBridgeJSON(w, http.StatusOK, map[string]any{
		"success":   true,
		"connected": true,
		"count":     len(devices),
		"device":    devices[0].Name,
		"devices":   devices,
	})
}

func (b *LocalBridgeServer) handleDeviceInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeBridgeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "message": "method not allowed"})
		return
	}
	devices, err := QueryDevices()
	if err != nil || len(devices) == 0 {
		writeBridgeJSON(w, http.StatusNotFound, map[string]any{
			"success": false,
			"message": "No fingerprint reader connected",
		})
		return
	}
	d := devices[0]
	writeBridgeJSON(w, http.StatusOK, map[string]any{
		"success":          true,
		"name":             d.Name,
		"vendor_name":      d.VendorName,
		"product_name":     d.ProductName,
		"serial_number":    d.SerialNum,
		"vendor_id":        d.VendorID,
		"product_id":       d.ProductID,
		"firmware_version": d.FirmwareVersion,
		"hardware_version": d.HardwareVersion,
	})
}

func (b *LocalBridgeServer) handleCapture(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		_ = CancelActiveCapture()
		writeBridgeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"message": "Capture canceled",
		})
		return
	case http.MethodPost:
		var req struct {
			TimeoutMs int `json:"timeout"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.TimeoutMs <= 0 {
			req.TimeoutMs = 15000
		}

		type captureOut struct {
			res *CaptureResponse
			err error
		}
		ch := make(chan captureOut, 1)
		go func() {
			b.captureHTTPMu.Lock()
			defer b.captureHTTPMu.Unlock()
			res, err := CaptureSingleFMD(req.TimeoutMs)
			ch <- captureOut{res: res, err: err}
		}()

		watchdogMs := req.TimeoutMs + 8000
		if watchdogMs < 12000 {
			watchdogMs = 12000
		}
		// WBF capture may prompt for UAC elevation on access-denied fallback.
		if devices, qErr := QueryDevices(); qErr == nil && len(devices) > 0 && strings.HasPrefix(devices[0].Name, "WBF:") {
			if watchdogMs < req.TimeoutMs+90000 {
				watchdogMs = req.TimeoutMs + 90000
			}
		}
		timer := time.NewTimer(time.Duration(watchdogMs) * time.Millisecond)
		defer timer.Stop()

		select {
		case out := <-ch:
			if out.err != nil {
				writeBridgeJSON(w, bridgeCaptureHTTPStatus(out.err), map[string]any{
					"success": false,
					"message": CaptureUserMessage(out.err),
				})
				return
			}
			res := out.res
			writeBridgeJSON(w, http.StatusOK, map[string]any{
				"success":       true,
				"fmd":           res.FMDBase64,
				"fmd_base64":    res.FMDBase64,
				"bmp":           res.BMPBase64,
				"image_base64":  res.BMPBase64,
				"quality":       res.NFIQScore,
				"quality_score": res.NFIQScore,
				"nfiq_score":    res.NFIQScore,
				"reader_serial": res.ReaderSerial,
				"width":         res.Width,
				"height":        res.Height,
			})
		case <-timer.C:
			log.Printf("[BiometricBridge] Capture watchdog fired after %dms — canceling hardware capture", watchdogMs)
			_ = CancelActiveCapture()
			writeBridgeJSON(w, http.StatusRequestTimeout, map[string]any{
				"success": false,
				"message": "Capture timed out waiting for a good fingerprint scan. Press the sensor firmly until the light stays on, then retry. If this repeats, run Install / Repair Drivers (Biometric class) and replug USB.",
			})
		case <-r.Context().Done():
			_ = CancelActiveCapture()
			writeBridgeJSON(w, http.StatusRequestTimeout, map[string]any{
				"success": false,
				"message": "Capture request canceled",
			})
		}
	default:
		writeBridgeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "message": "method not allowed"})
	}
}

func (b *LocalBridgeServer) handleEnrollmentStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeBridgeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "message": "method not allowed"})
		return
	}
	b.enrollMu.Lock()
	b.enrollViews = nil
	b.enrollSerial = ""
	b.enrollMu.Unlock()

	writeBridgeJSON(w, http.StatusOK, map[string]any{
		"success":        true,
		"step":           0,
		"required_steps": 4,
		"message":        "Enrollment started. Please scan the same finger 4 times.",
	})
}

func (b *LocalBridgeServer) handleEnrollmentStep(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeBridgeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "message": "method not allowed"})
		return
	}
	b.captureHTTPMu.Lock()
	res, err := CaptureSingleFMD(15000)
	b.captureHTTPMu.Unlock()
	if err != nil {
		writeBridgeJSON(w, bridgeCaptureHTTPStatus(err), map[string]any{
			"success": false,
			"message": CaptureUserMessage(err),
		})
		return
	}
	if res.NFIQScore >= 4 {
		writeBridgeJSON(w, http.StatusBadRequest, map[string]any{
			"success":    false,
			"nfiq_score": res.NFIQScore,
			"bmp":        res.BMPBase64,
			"message":    ErrPoorQualityScan.Error(),
		})
		return
	}
	viewBytes, err := ValidateAndDecodeFMD(res.FMDBase64)
	if err != nil {
		writeBridgeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	b.enrollMu.Lock()
	b.enrollViews = append(b.enrollViews, viewBytes)
	b.enrollSerial = res.ReaderSerial
	step := len(b.enrollViews)
	var viewsCopy [][]byte
	if step >= 4 {
		viewsCopy = append([][]byte(nil), b.enrollViews...)
		b.enrollViews = nil
	}
	b.enrollMu.Unlock()

	if step < 4 {
		writeBridgeJSON(w, http.StatusOK, map[string]any{
			"success":          true,
			"step":             step,
			"required_steps":   4,
			"needs_more_data":  true,
			"nfiq_score":       res.NFIQScore,
			"quality":          res.QualityScore,
			"bmp":              res.BMPBase64,
			"reader_serial":    res.ReaderSerial,
			"consolidated_fmd": nil,
		})
		return
	}

	consolidated, err := ConsolidateEnrollmentFMD(viewsCopy)
	if err != nil {
		writeBridgeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	fmdB64 := base64.StdEncoding.EncodeToString(consolidated)
	writeBridgeJSON(w, http.StatusOK, map[string]any{
		"success":          true,
		"step":             4,
		"required_steps":   4,
		"needs_more_data":  false,
		"nfiq_score":       res.NFIQScore,
		"quality":          res.QualityScore,
		"bmp":              res.BMPBase64,
		"reader_serial":    res.ReaderSerial,
		"consolidated_fmd": fmdB64,
		"fmd":              fmdB64,
	})
}

func (b *LocalBridgeServer) handleEnrollment(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		_ = CancelActiveCapture()
		b.enrollMu.Lock()
		b.enrollViews = nil
		b.enrollMu.Unlock()
		writeBridgeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Enrollment canceled"})
		return
	}
	if r.Method != http.MethodPost {
		writeBridgeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "message": "method not allowed"})
		return
	}

	var req struct {
		Count int `json:"count"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Count <= 0 || req.Count > 10 {
		req.Count = 4
	}

	views := make([][]byte, 0, req.Count)
	var lastCapture *CaptureResponse
	for len(views) < req.Count {
		b.captureHTTPMu.Lock()
		capRes, err := CaptureSingleFMD(15000)
		b.captureHTTPMu.Unlock()
		if err != nil {
			writeBridgeJSON(w, bridgeCaptureHTTPStatus(err), map[string]any{
				"success": false,
				"message": CaptureUserMessage(err),
			})
			return
		}
		if capRes.NFIQScore >= 4 {
			continue
		}
		vBytes, err := ValidateAndDecodeFMD(capRes.FMDBase64)
		if err != nil {
			continue
		}
		views = append(views, vBytes)
		lastCapture = capRes
	}

	consolidated, err := ConsolidateEnrollmentFMD(views)
	if err != nil {
		writeBridgeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	fmdB64 := base64.StdEncoding.EncodeToString(consolidated)
	resp := map[string]any{
		"success":          true,
		"count":            req.Count,
		"consolidated_fmd": fmdB64,
		"fmd":              fmdB64,
	}
	if lastCapture != nil {
		resp["bmp"] = lastCapture.BMPBase64
		resp["quality"] = lastCapture.QualityScore
		resp["nfiq_score"] = lastCapture.NFIQScore
		resp["reader_serial"] = lastCapture.ReaderSerial
	}
	writeBridgeJSON(w, http.StatusOK, resp)
}

func writeBridgeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func bridgeCaptureHTTPStatus(err error) int {
	if errors.Is(err, ErrNoReaderConnected) {
		return http.StatusServiceUnavailable
	}
	if errors.Is(err, ErrReaderInUse) {
		return http.StatusConflict
	}
	if errors.Is(err, ErrCaptureTimeout) {
		return http.StatusRequestTimeout
	}
	if errors.Is(err, ErrCaptureCanceled) {
		return http.StatusConflict
	}
	return http.StatusBadRequest
}

