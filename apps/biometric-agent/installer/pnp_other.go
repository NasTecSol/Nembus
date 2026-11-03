//go:build !windows

package installer

const (
	ReaderStackLegacy  = "legacy"
	ReaderStackWBF     = "wbf"
	ReaderStackHello   = "hello"
	ReaderStackUnknown = "unknown"
)

// USBReaderPnP describes a DigitalPersona USB reader visible to Windows Plug-and-Play.
type USBReaderPnP struct {
	Present           bool   `json:"present"`
	FriendlyName      string `json:"friendly_name,omitempty"`
	InstanceID        string `json:"instance_id,omitempty"`
	Class             string `json:"class,omitempty"`
	Status            string `json:"status,omitempty"`
	DriverDesc        string `json:"driver_desc,omitempty"`
	ReaderStack       string `json:"reader_stack,omitempty"`
	UsingWindowsHello bool   `json:"using_windows_hello_driver"`
}

func ListUSBReaders() []USBReaderPnP {
	return nil
}
