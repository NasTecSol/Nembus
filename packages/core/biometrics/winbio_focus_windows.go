//go:build windows

package biometrics

import (
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

const (
	mbOK            = 0x00000000
	mbTopmost       = 0x00040000
	mbSetForeground = 0x00010000
	wmClose         = 0x0010
)

const winBioCaptureDialogTitle = "Nembus Biometric Capture"

var (
	user32Focus                  = syscall.NewLazyDLL("user32.dll")
	procMessageBoxW              = user32Focus.NewProc("MessageBoxW")
	procFindWindowW              = user32Focus.NewProc("FindWindowW")
	procPostMessageW             = user32Focus.NewProc("PostMessageW")
	procAllowSetForegroundWindow = user32Focus.NewProc("AllowSetForegroundWindow")
)

// WinBioElevatedCaptureFunc is optionally set by the biometric agent to run WBF capture in an
// elevated child process when raw WinBio access is denied for the current user token.
var WinBioElevatedCaptureFunc func(unitID uint32, timeoutMs int) (*CaptureResponse, error)

type captureResult struct {
	resp *CaptureResponse
	err  error
}

// withWinBioCaptureForeground runs capture on a dedicated OS thread with a visible dialog so
// WinBio system-pool capture receives foreground focus in this process.
func withWinBioCaptureForeground(capture func() (*CaptureResponse, error)) (*CaptureResponse, error) {
	ch := make(chan captureResult, 1)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		_, _, _ = procAllowSetForegroundWindow.Call(0xFFFFFFFF) // ASFW_ANY

		dialogClosed := make(chan struct{})
		go func() {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			defer close(dialogClosed)

			title, err := syscall.UTF16PtrFromString(winBioCaptureDialogTitle)
			if err != nil {
				return
			}
			text, err := syscall.UTF16PtrFromString(
				"Place your finger flat on the reader now.\n" +
					"Leave this window open until capture completes (it closes automatically).",
			)
			if err != nil {
				return
			}
			_, _, _ = procMessageBoxW.Call(
				0,
				uintptr(unsafe.Pointer(text)),
				uintptr(unsafe.Pointer(title)),
				uintptr(mbOK|mbTopmost|mbSetForeground),
			)
		}()

		time.Sleep(600 * time.Millisecond)
		resp, err := capture()
		endWinBioCaptureFocusDialog()
		<-dialogClosed
		ch <- captureResult{resp: resp, err: err}
	}()
	r := <-ch
	return r.resp, r.err
}

func endWinBioCaptureFocusDialog() {
	title, err := syscall.UTF16PtrFromString(winBioCaptureDialogTitle)
	if err != nil {
		return
	}
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(title)))
	if hwnd != 0 {
		_, _, _ = procPostMessageW.Call(hwnd, wmClose, 0, 0)
	}
}
