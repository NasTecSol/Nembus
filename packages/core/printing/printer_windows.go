//go:build windows

package printing

import (
	"fmt"
	"net"
	"syscall"
	"time"
	"unsafe"
)

// ─── Unified dispatcher ───────────────────────────────────────────────────────

// Print sends raw ESC/POS data to the printer described by cfg.
func Print(cfg PrinterConfig, data []byte) error {
	netPort := cfg.NetworkPort
	if netPort == "" {
		netPort = "9100"
	}

	switch cfg.Mode {
	case "usb":
		if cfg.PrinterName == "" {
			return fmt.Errorf("printer_name is required for USB mode")
		}
		return printViaWindowsSpooler(cfg.PrinterName, data)
	case "serial":
		if cfg.SerialPort == "" {
			return fmt.Errorf("serial_port is required for serial mode")
		}
		return printViaCOMPort(cfg.SerialPort, data)
	case "network":
		if cfg.NetworkIP == "" {
			return fmt.Errorf("network_ip is required for network mode")
		}
		return printViaNetwork(cfg.NetworkIP, netPort, data)
	default:
		return fmt.Errorf("unsupported printer mode %q; use usb, serial, or network", cfg.Mode)
	}
}

// ─── USB – Windows spooler ────────────────────────────────────────────────────

// printViaWindowsSpooler sends raw bytes using winspool.drv.
// Works with USB printers installed as "Generic / Text Only".
func printViaWindowsSpooler(printerName string, data []byte) error {
	winspool := syscall.MustLoadDLL("winspool.drv")
	openPrinter := winspool.MustFindProc("OpenPrinterW")
	startDoc := winspool.MustFindProc("StartDocPrinterW")
	startPage := winspool.MustFindProc("StartPagePrinter")
	writePrinter := winspool.MustFindProc("WritePrinter")
	endPage := winspool.MustFindProc("EndPagePrinter")
	endDoc := winspool.MustFindProc("EndDocPrinter")
	closePrinter := winspool.MustFindProc("ClosePrinter")

	type DocInfo1 struct {
		DocName    *uint16
		OutputFile *uint16
		Datatype   *uint16
	}

	namePtr, _ := syscall.UTF16PtrFromString(printerName)
	datatype, _ := syscall.UTF16PtrFromString("RAW")
	docName, _ := syscall.UTF16PtrFromString("POS Receipt")

	var hPrinter uintptr
	r, _, err := openPrinter.Call(
		uintptr(unsafe.Pointer(namePtr)),
		uintptr(unsafe.Pointer(&hPrinter)),
		0,
	)
	if r == 0 {
		return fmt.Errorf("OpenPrinter failed: %w", err)
	}
	defer closePrinter.Call(hPrinter)

	di := DocInfo1{DocName: docName, OutputFile: nil, Datatype: datatype}
	r, _, err = startDoc.Call(hPrinter, 1, uintptr(unsafe.Pointer(&di)))
	if r == 0 {
		return fmt.Errorf("StartDocPrinter failed: %w", err)
	}
	defer endDoc.Call(hPrinter)

	r, _, err = startPage.Call(hPrinter)
	if r == 0 {
		return fmt.Errorf("StartPagePrinter failed: %w", err)
	}
	defer endPage.Call(hPrinter)

	var written uint32
	r, _, err = writePrinter.Call(
		hPrinter,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
		uintptr(unsafe.Pointer(&written)),
	)
	if r == 0 {
		return fmt.Errorf("WritePrinter failed: %w", err)
	}

	return nil
}

// ─── Serial – COM port ────────────────────────────────────────────────────────

// printViaCOMPort sends raw bytes to a serial COM port (9600 baud, 8N1).
func printViaCOMPort(port string, data []byte) error {
	portPath := `\\.\` + port
	ptr, err := syscall.UTF16PtrFromString(portPath)
	if err != nil {
		return err
	}

	h, err := syscall.CreateFile(
		ptr,
		syscall.GENERIC_WRITE,
		0, nil,
		syscall.OPEN_EXISTING,
		0, 0,
	)
	if err != nil || h == syscall.InvalidHandle {
		return fmt.Errorf("cannot open %s: %w", port, err)
	}
	defer syscall.CloseHandle(h)

	// DCB – serial port config
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
	dcb := DCB{DCBlength: 28, BaudRate: 9600, ByteSize: 8}
	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	setCommState := kernel32.MustFindProc("SetCommState")
	setCommState.Call(uintptr(h), uintptr(unsafe.Pointer(&dcb)))

	var written uint32
	if err = syscall.WriteFile(h, data, &written, nil); err != nil {
		return fmt.Errorf("write to %s failed: %w", port, err)
	}
	return nil
}

// ─── Network – TCP port 9100 ─────────────────────────────────────────────────

// printViaNetwork sends raw bytes over TCP to the printer's network port.
func printViaNetwork(ip, port string, data []byte) error {
	addr := net.JoinHostPort(ip, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("cannot connect to %s: %w", addr, err)
	}
	defer conn.Close()
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	if _, err = conn.Write(data); err != nil {
		return fmt.Errorf("network write failed: %w", err)
	}
	return nil
}
