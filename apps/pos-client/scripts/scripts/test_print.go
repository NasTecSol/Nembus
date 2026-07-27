// test_print.go
// Demo receipt test for P-25 Thermal Printer (ESC/POS)
// Supports: USB, Serial (COM), Ethernet/IP auto-detection on Windows
//
// Run:
//   go run test_print.go                        → auto-detect
//   go run test_print.go -mode=usb              → force USB
//   go run test_print.go -mode=serial -port=COM3
//   go run test_print.go -mode=network -ip=192.168.1.100

//go:build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

// ─── ESC/POS Commands ────────────────────────────────────────────────────────

var (
	ESC = byte(0x1B)
	GS  = byte(0x1D)
	LF  = byte(0x0A)
	CR  = byte(0x0D)
	HT  = byte(0x09)
)

func cmdInit() []byte         { return []byte{ESC, '@'} }
func cmdCut() []byte          { return []byte{GS, 'V', 66, 0} } // partial cut
func cmdFeedLines(n int) []byte { return []byte{ESC, 'd', byte(n)} }

func cmdAlign(a string) []byte {
	switch a {
	case "center":
		return []byte{ESC, 'a', 1}
	case "right":
		return []byte{ESC, 'a', 2}
	default:
		return []byte{ESC, 'a', 0}
	}
}

func cmdBold(on bool) []byte {
	if on {
		return []byte{ESC, 'E', 1}
	}
	return []byte{ESC, 'E', 0}
}

func cmdDoubleHeight(on bool) []byte {
	if on {
		return []byte{GS, '!', 0x01} // double height only
	}
	return []byte{GS, '!', 0x00}
}

func cmdDoubleSize(on bool) []byte {
	if on {
		return []byte{GS, '!', 0x11} // double width + height
	}
	return []byte{GS, '!', 0x00}
}

func cmdUnderline(on bool) []byte {
	if on {
		return []byte{ESC, '-', 1}
	}
	return []byte{ESC, '-', 0}
}

func cmdBarcode(data string) []byte {
	// CODE128 barcode
	// GS k m d1...dk NUL  (m=73 for CODE128 auto)
	b := []byte{GS, 'k', 73, byte(len(data))}
	b = append(b, []byte(data)...)
	return b
}

func cmdBarcodeHeight(h byte) []byte { return []byte{GS, 'h', h} }
func cmdBarcodeWidth(w byte) []byte  { return []byte{GS, 'w', w} }
func cmdBarcodeHRI(pos byte) []byte  { return []byte{GS, 'H', pos} } // 2=below

// ─── Receipt Builder ─────────────────────────────────────────────────────────

const COLS = 48 // 80mm paper → 48 columns at standard font

func divider() []byte {
	return []byte(strings.Repeat("-", COLS) + "\n")
}

func padRight(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	return s + strings.Repeat(" ", w-len(s))
}

func padLeft(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	return strings.Repeat(" ", w-len(s)) + s
}

func col2(left, right string, width int) []byte {
	// Two-column line: left-aligned + right-aligned
	space := width - len(left) - len(right)
	if space < 1 {
		space = 1
	}
	line := left + strings.Repeat(" ", space) + right + "\n"
	return []byte(line)
}

func buildDemoReceipt() []byte {
	var buf bytes.Buffer
	w := func(b []byte) { buf.Write(b) }
	p := func(s string) { buf.WriteString(s) }
	nl := func() { buf.WriteByte(LF) }

	now := time.Now()

	// ── Initialize printer ──────────────────────────────────────────────
	w(cmdInit())

	// ── HEADER ──────────────────────────────────────────────────────────
	w(cmdAlign("center"))
	w(cmdDoubleSize(true))
	w(cmdBold(true))
	p("MY STORE\n")
	w(cmdDoubleSize(false))
	w(cmdBold(false))

	p("123 Main Street, Shop No. 5\n")
	p("Rawalpindi, Punjab, Pakistan\n")
	p("Tel: +92-51-1234567\n")
	p("TIN: PK-123456789\n")
	w(divider())

	// ── RECEIPT INFO ────────────────────────────────────────────────────
	w(cmdAlign("left"))
	w(col2("Receipt #: RCP-20260310-001", now.Format("02-Jan-06 15:04"), COLS))
	w(col2("Cashier  : Ali Hassan (C001)", "Terminal: T-01", COLS))
	p("Customer : Walk-in Customer\n")
	w(divider())

	// ── COLUMN HEADERS ──────────────────────────────────────────────────
	w(cmdBold(true))
	line := padRight("Item", 24) +
		padLeft("Qty", 6) +
		padLeft("Price", 9) +
		padLeft("Total", 9) + "\n"
	p(line)
	w(cmdBold(false))
	w(divider())

	// ── LINE ITEMS ───────────────────────────────────────────────────────
	type item struct {
		name  string
		qty   float64
		price float64
	}
	items := []item{
		{"Chicken Burger", 2, 450.00},
		{"French Fries (L)", 1, 250.00},
		{"Soft Drink 500ml", 3, 120.00},
		{"Garlic Bread", 2, 180.00},
		{"Chocolate Sundae", 1, 320.00},
	}

	subtotal := 0.0
	for _, it := range items {
		total := it.qty * it.price
		subtotal += total
		name := padRight(it.name, 24)
		qty := padLeft(fmt.Sprintf("%.0f", it.qty), 6)
		price := padLeft(fmt.Sprintf("%.2f", it.price), 9)
		tot := padLeft(fmt.Sprintf("%.2f", total), 9)
		p(name + qty + price + tot + "\n")
	}

	// ── TOTALS ──────────────────────────────────────────────────────────
	w(divider())
	discount := 135.00
	taxRate := 0.05
	netAfterDisc := subtotal - discount
	tax := netAfterDisc * taxRate
	total := netAfterDisc + tax
	paid := 2000.00
	change := paid - total

	w(cmdAlign("right"))
	p(fmt.Sprintf("Subtotal         : PKR %9.2f\n", subtotal))
	p(fmt.Sprintf("Discount (5%%)    : PKR -%8.2f\n", discount))
	p(fmt.Sprintf("Tax (GST 5%%)     : PKR %9.2f\n", tax))

	w(cmdBold(true))
	w(cmdDoubleHeight(true))
	p(fmt.Sprintf("TOTAL            : PKR %9.2f\n", total))
	w(cmdDoubleHeight(false))
	w(cmdBold(false))

	w(divider())
	p(fmt.Sprintf("Cash Paid        : PKR %9.2f\n", paid))
	p(fmt.Sprintf("Change           : PKR %9.2f\n", change))
	w(divider())

	// ── PAYMENT METHOD ──────────────────────────────────────────────────
	w(cmdAlign("left"))
	p("Payment : CASH\n")
	nl()

	// ── BARCODE ─────────────────────────────────────────────────────────
	w(cmdAlign("center"))
	w(cmdBarcodeHeight(60))
	w(cmdBarcodeWidth(2))
	w(cmdBarcodeHRI(2)) // text below
	w(cmdBarcode("RCP-20260310-001"))
	nl()

	// ── FOOTER ──────────────────────────────────────────────────────────
	w(cmdAlign("center"))
	w(divider())
	w(cmdBold(true))
	p("** Thank you for your visit! **\n")
	w(cmdBold(false))
	p("Please keep this receipt\n")
	p("for any exchange or return.\n")
	p("www.mystore.com\n")
	w(divider())

	// ── Feed & Cut ──────────────────────────────────────────────────────
	w(cmdFeedLines(4))
	w(cmdCut())

	return buf.Bytes()
}

// ─── Printer Detection & Connection ──────────────────────────────────────────

type PrintMethod string

const (
	MethodUSB     PrintMethod = "usb"
	MethodSerial  PrintMethod = "serial"
	MethodNetwork PrintMethod = "network"
)

// autoDetect tries USB → Serial → Network in order
func autoDetect() (PrintMethod, string, error) {
	fmt.Println("[AUTO-DETECT] Scanning for P-25 thermal printer...")

	// 1. Try USB via Windows print spooler (Generic/Text Only or ESC/POS driver)
	if name, ok := detectWindowsPrinter(); ok {
		fmt.Printf("[AUTO-DETECT] Found Windows printer: %s\n", name)
		return MethodUSB, name, nil
	}

	// 2. Try common Serial COM ports
	for _, com := range []string{"COM1", "COM2", "COM3", "COM4", "COM5", "COM6"} {
		if testCOMPort(com) {
			fmt.Printf("[AUTO-DETECT] Found printer on %s\n", com)
			return MethodSerial, com, nil
		}
	}

	// 3. Try common network IPs on port 9100
	commonIPs := []string{
		"192.168.1.100", "192.168.1.101", "192.168.1.200",
		"192.168.0.100", "192.168.0.101",
		"10.0.0.100",
	}
	for _, ip := range commonIPs {
		if testNetworkPrinter(ip, "9100") {
			fmt.Printf("[AUTO-DETECT] Found printer at %s:9100\n", ip)
			return MethodNetwork, ip, nil
		}
	}

	return "", "", fmt.Errorf("no printer found – check connections and try manual mode")
}

// detectWindowsPrinter scans the Windows registry for installed printers
// and looks for thermal / ESC/POS / receipt printers by name keywords
func detectWindowsPrinter() (string, bool) {
	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Print\Printers`,
		registry.READ,
	)
	if err != nil {
		return "", false
	}
	defer key.Close()

	names, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return "", false
	}

	keywords := []string{
		"thermal", "receipt", "pos", "escpos", "esc/pos",
		"epson", "star", "bixolon", "sewoo", "p-25", "p25",
		"generic", "text only", "usb",
	}

	for _, name := range names {
		lower := strings.ToLower(name)
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				return name, true
			}
		}
	}

	// If only one printer installed, just use it
	if len(names) == 1 {
		return names[0], true
	}

	return "", false
}

// testCOMPort checks if a COM port is accessible (likely has a device)
func testCOMPort(port string) bool {
	// Try to open the COM port
	portPath := `\\.\` + port
	ptr, err := syscall.UTF16PtrFromString(portPath)
	if err != nil {
		return false
	}
	h, err := syscall.CreateFile(
		ptr,
		syscall.GENERIC_WRITE,
		0, nil,
		syscall.OPEN_EXISTING,
		0, 0,
	)
	if err != nil || h == syscall.InvalidHandle {
		return false
	}
	syscall.CloseHandle(h)
	return true
}

// testNetworkPrinter tries a TCP connection on port 9100
func testNetworkPrinter(ip, port string) bool {
	addr := fmt.Sprintf("%s:%s", ip, port)
	conn, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// ─── Print Drivers ───────────────────────────────────────────────────────────

// printViaWindowsSpooler sends raw bytes using winspool.drv
// Works with USB printers installed as "Generic / Text Only"
func printViaWindowsSpooler(printerName string, data []byte) error {
	winspool := syscall.MustLoadDLL("winspool.drv")
	openPrinter := winspool.MustFindProc("OpenPrinterW")
	startDoc := winspool.MustFindProc("StartDocPrinterW")
	startPage := winspool.MustFindProc("StartPagePrinter")
	writePrinter := winspool.MustFindProc("WritePrinter")
	endPage := winspool.MustFindProc("EndPagePrinter")
	endDoc := winspool.MustFindProc("EndDocPrinter")
	closePrinter := winspool.MustFindProc("ClosePrinter")

	// DOC_INFO_1 structure
	type DocInfo1 struct {
		DocName    *uint16
		OutputFile *uint16
		Datatype   *uint16
	}

	namePtr, _ := syscall.UTF16PtrFromString(printerName)
	datatype, _ := syscall.UTF16PtrFromString("RAW")
	docName, _ := syscall.UTF16PtrFromString("POS Receipt")

	// Open printer
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

	// Start document
	di := DocInfo1{
		DocName:    docName,
		OutputFile: nil,
		Datatype:   datatype,
	}
	r, _, err = startDoc.Call(hPrinter, 1, uintptr(unsafe.Pointer(&di)))
	if r == 0 {
		return fmt.Errorf("StartDocPrinter failed: %w", err)
	}
	defer endDoc.Call(hPrinter)

	// Start page
	r, _, err = startPage.Call(hPrinter)
	if r == 0 {
		return fmt.Errorf("StartPagePrinter failed: %w", err)
	}
	defer endPage.Call(hPrinter)

	// Write raw ESC/POS bytes
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

	fmt.Printf("[USB] Sent %d bytes to '%s'\n", written, printerName)
	return nil
}

// printViaCOMPort sends raw bytes to a serial COM port
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

	// DCB – serial port config: 9600 baud, 8N1
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
	dcb := DCB{
		DCBlength: 28,
		BaudRate:  9600,
		ByteSize:  8,
		Parity:    0,
		StopBits:  0,
	}

	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	setCommState := kernel32.MustFindProc("SetCommState")
	setCommState.Call(uintptr(h), uintptr(unsafe.Pointer(&dcb)))

	var written uint32
	err = syscall.WriteFile(h, data, &written, nil)
	if err != nil {
		return fmt.Errorf("write to %s failed: %w", port, err)
	}
	fmt.Printf("[SERIAL] Sent %d bytes to %s\n", written, port)
	return nil
}

// printViaNetwork sends raw bytes to TCP port 9100
func printViaNetwork(ip, port string, data []byte) error {
	addr := fmt.Sprintf("%s:%s", ip, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("cannot connect to %s: %w", addr, err)
	}
	defer conn.Close()
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	n, err := conn.Write(data)
	if err != nil {
		return fmt.Errorf("network write failed: %w", err)
	}
	fmt.Printf("[NETWORK] Sent %d bytes to %s\n", n, addr)
	return nil
}

// ─── Main ─────────────────────────────────────────────────────────────────────

func main() {
	mode := flag.String("mode", "auto", "Connection mode: auto | usb | serial | network")
	port := flag.String("port", "", "Serial port (e.g. COM3) – used with -mode=serial")
	ip := flag.String("ip", "", "Printer IP address – used with -mode=network")
	netPort := flag.String("netport", "9100", "Network port (default 9100)")
	flag.Parse()

	fmt.Println("==============================================")
	fmt.Println("  P-25 Thermal Printer – Demo Receipt Test   ")
	fmt.Println("  Paper: 80mm | ESC/POS | Auto-detect        ")
	fmt.Println("==============================================")

	// Build receipt bytes
	receipt := buildDemoReceipt()
	fmt.Printf("[INFO] Receipt built: %d bytes\n", len(receipt))

	var printErr error

	switch *mode {
	case "usb":
		// Force USB via Windows spooler
		printerName, ok := detectWindowsPrinter()
		if !ok {
			fmt.Println("[ERROR] No printer found in Windows Printers list.")
			fmt.Println("        Install printer as 'Generic / Text Only' first.")
			os.Exit(1)
		}
		fmt.Printf("[USB] Printing to: %s\n", printerName)
		printErr = printViaWindowsSpooler(printerName, receipt)

	case "serial":
		if *port == "" {
			*port = "COM3"
			fmt.Printf("[SERIAL] No port specified, trying %s\n", *port)
		}
		printErr = printViaCOMPort(*port, receipt)

	case "network":
		if *ip == "" {
			fmt.Println("[ERROR] Provide -ip=<printer-ip> for network mode")
			os.Exit(1)
		}
		printErr = printViaNetwork(*ip, *netPort, receipt)

	default: // auto
		method, target, err := autoDetect()
		if err != nil {
			fmt.Println("[ERROR]", err)
			fmt.Println()
			fmt.Println("Manual options:")
			fmt.Println("  USB    : go run test_print.go -mode=usb")
			fmt.Println("  Serial : go run test_print.go -mode=serial -port=COM3")
			fmt.Println("  Network: go run test_print.go -mode=network -ip=192.168.1.100")
			os.Exit(1)
		}

		switch method {
		case MethodUSB:
			printErr = printViaWindowsSpooler(target, receipt)
		case MethodSerial:
			printErr = printViaCOMPort(target, receipt)
		case MethodNetwork:
			printErr = printViaNetwork(target, *netPort, receipt)
		}
	}

	if printErr != nil {
		fmt.Println("[FAIL]", printErr)
		os.Exit(1)
	}

	fmt.Println("[OK] Demo receipt sent to printer successfully!")
}
