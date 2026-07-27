package printing

// PrinterConfig describes how to reach the thermal printer.
type PrinterConfig struct {
	// Mode selects the connection method: "usb", "serial", or "network".
	Mode string `json:"mode"`

	// PrinterName is the Windows spooler printer name (USB / Generic Text Only).
	PrinterName string `json:"printer_name"`

	// SerialPort is the COM port for serial connections (e.g. "COM3").
	SerialPort string `json:"serial_port"`

	// NetworkIP is the printer's IP address for Ethernet/Wi-Fi printers.
	NetworkIP string `json:"network_ip"`

	// NetworkPort is the TCP port (defaults to "9100" if empty).
	NetworkPort string `json:"network_port"`
}
