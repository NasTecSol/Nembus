//go:build windows

package geidea

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

// SerialTransport implements Transport over Windows COM port
type SerialTransport struct {
	handle syscall.Handle
	port   string
}

// COMMTIMEOUTS structure for Windows Win32
type COMMTIMEOUTS struct {
	ReadIntervalTimeout         uint32
	ReadTotalTimeoutMultiplier  uint32
	ReadTotalTimeoutConstant    uint32
	WriteTotalTimeoutMultiplier uint32
	WriteTotalTimeoutConstant   uint32
}

// DCB structure for Windows Win32
type DCB struct {
	DCBlength  uint32
	BaudRate   uint32
	Flags      uint32
	wReserved  uint16
	XonLim     uint16
	XoffLim    uint16
	ByteSize   byte
	Parity     byte
	StopBits   byte
	XonChar    byte
	XoffChar   byte
	ErrorChar  byte
	EofChar    byte
	EvtChar    byte
	wReserved1 uint16
}

// NewSerialTransport opens a Windows COM port for Geidea terminal communication
func NewSerialTransport(port string, baudRate int) (*SerialTransport, error) {
	if port == "" {
		return nil, fmt.Errorf("serial port name is required")
	}
	if baudRate <= 0 {
		baudRate = 38400
	}

	portPath := `\\.\` + port
	ptr, err := syscall.UTF16PtrFromString(portPath)
	if err != nil {
		return nil, fmt.Errorf("invalid port name %s: %w", port, err)
	}

	h, err := syscall.CreateFile(
		ptr,
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		0,
		nil,
		syscall.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil || h == syscall.InvalidHandle {
		return nil, fmt.Errorf("cannot open COM port %s: %w", port, err)
	}

	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	setCommState := kernel32.MustFindProc("SetCommState")
	setCommTimeouts := kernel32.MustFindProc("SetCommTimeouts")

	// Set DCB (8N1, specified baud)
	dcb := DCB{
		DCBlength: 28,
		BaudRate:  uint32(baudRate),
		ByteSize:  8,
		Parity:    0, // NOPARITY
		StopBits:  0, // ONESTOPBIT
		Flags:     1, // fBinary
	}
	r, _, err := setCommState.Call(uintptr(h), uintptr(unsafe.Pointer(&dcb)))
	if r == 0 {
		syscall.CloseHandle(h)
		return nil, fmt.Errorf("SetCommState failed on %s: %w", port, err)
	}

	// Set Comm Timeouts (non-blocking chunk reads)
	timeouts := COMMTIMEOUTS{
		ReadIntervalTimeout:         50,
		ReadTotalTimeoutMultiplier:  10,
		ReadTotalTimeoutConstant:    500,
		WriteTotalTimeoutMultiplier: 10,
		WriteTotalTimeoutConstant:   1000,
	}
	r, _, err = setCommTimeouts.Call(uintptr(h), uintptr(unsafe.Pointer(&timeouts)))
	if r == 0 {
		syscall.CloseHandle(h)
		return nil, fmt.Errorf("SetCommTimeouts failed on %s: %w", port, err)
	}

	return &SerialTransport{handle: h, port: port}, nil
}

func (s *SerialTransport) Write(data []byte) error {
	var written uint32
	if err := syscall.WriteFile(s.handle, data, &written, nil); err != nil {
		return fmt.Errorf("write to serial %s failed: %w", s.port, err)
	}
	return nil
}

func (s *SerialTransport) Read(buf []byte) (int, error) {
	var bytesRead uint32
	if err := syscall.ReadFile(s.handle, buf, &bytesRead, nil); err != nil {
		return 0, fmt.Errorf("read from serial %s failed: %w", s.port, err)
	}
	return int(bytesRead), nil
}

func (s *SerialTransport) SetReadDeadline(deadline time.Time) error {
	// Handled via COMMTIMEOUTS in Win32
	return nil
}

func (s *SerialTransport) Close() error {
	if s.handle != syscall.InvalidHandle {
		return syscall.CloseHandle(s.handle)
	}
	return nil
}
