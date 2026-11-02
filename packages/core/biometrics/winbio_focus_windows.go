//go:build windows

package biometrics

import (
	"log"
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
	kernel32Focus                = syscall.NewLazyDLL("kernel32.dll")
	procMessageBoxW              = user32Focus.NewProc("MessageBoxW")
	procFindWindowW              = user32Focus.NewProc("FindWindowW")
	procGetForegroundWindow      = user32Focus.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessID = user32Focus.NewProc("GetWindowThreadProcessId")
	procSetForegroundWindow      = user32Focus.NewProc("SetForegroundWindow")
	procSetActiveWindow          = user32Focus.NewProc("SetActiveWindow")
	procSetFocus                 = user32Focus.NewProc("SetFocus")
	procBringWindowToTop         = user32Focus.NewProc("BringWindowToTop")
	procAttachThreadInput        = user32Focus.NewProc("AttachThreadInput")
	procPostMessageW             = user32Focus.NewProc("PostMessageW")
	procAllowSetForegroundWindow = user32Focus.NewProc("AllowSetForegroundWindow")
	procGetCurrentThreadID       = kernel32Focus.NewProc("GetCurrentThreadId")
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
		dialogHWND, foregroundHWND, activated := activateWinBioCaptureDialog()
		log.Printf("[WinBio] focus dialog hwnd=0x%x foreground hwnd=0x%x foreground_match=%t activation_succeeded=%t", dialogHWND, foregroundHWND, dialogHWND != 0 && dialogHWND == foregroundHWND, activated)
		resp, err := capture()
		endWinBioCaptureFocusDialog()
		<-dialogClosed
		ch <- captureResult{resp: resp, err: err}
	}()
	r := <-ch
	return r.resp, r.err
}

// activateWinBioCaptureDialog makes the prompt the foreground window before
// WinBioCaptureSample starts. WBF system-pool requests are routed to the
// foreground window, so a visible but inactive prompt is not sufficient.
func activateWinBioCaptureDialog() (dialogHWND, foregroundHWND uintptr, activated bool) {
	title, err := syscall.UTF16PtrFromString(winBioCaptureDialogTitle)
	if err != nil {
		return 0, 0, false
	}
	dialogHWND, _, _ = procFindWindowW.Call(0, uintptr(unsafe.Pointer(title)))
	if dialogHWND == 0 {
		return 0, procGetForegroundWindowResult(), false
	}

	_, _, _ = procAllowSetForegroundWindow.Call(0xFFFFFFFF) // ASFW_ANY
	_, _, _ = procBringWindowToTop.Call(dialogHWND)
	_, _, _ = procSetForegroundWindow.Call(dialogHWND)

	foregroundHWND = procGetForegroundWindowResult()
	if foregroundHWND != dialogHWND {
		foregroundThread, _, _ := procGetWindowThreadProcessID.Call(foregroundHWND, 0)
		currentThread, _, _ := procGetCurrentThreadID.Call()
		if foregroundThread != 0 && currentThread != 0 && foregroundThread != currentThread {
			attached, _, _ := procAttachThreadInput.Call(currentThread, foregroundThread, 1)
			if attached != 0 {
				_, _, _ = procBringWindowToTop.Call(dialogHWND)
				_, _, _ = procSetActiveWindow.Call(dialogHWND)
				_, _, _ = procSetFocus.Call(dialogHWND)
				_, _, _ = procSetForegroundWindow.Call(dialogHWND)
				_, _, _ = procAttachThreadInput.Call(currentThread, foregroundThread, 0)
			}
			foregroundHWND = procGetForegroundWindowResult()
		}
	}
	return dialogHWND, foregroundHWND, foregroundHWND == dialogHWND
}

func procGetForegroundWindowResult() uintptr {
	hwnd, _, _ := procGetForegroundWindow.Call()
	return hwnd
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
