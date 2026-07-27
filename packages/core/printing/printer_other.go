//go:build !windows

package printing

import (
	"fmt"
	"net"
	"time"
)

// ─── Unified dispatcher for non-Windows platforms ───────────────────────────

// Print sends raw ESC/POS data to network printers or returns error for Windows-only modes.
func Print(cfg PrinterConfig, data []byte) error {
	netPort := cfg.NetworkPort
	if netPort == "" {
		netPort = "9100"
	}

	switch cfg.Mode {
	case "network":
		if cfg.NetworkIP == "" {
			return fmt.Errorf("network_ip is required for network mode")
		}
		return printViaNetwork(cfg.NetworkIP, netPort, data)
	case "usb", "serial":
		return fmt.Errorf("printer mode %q is supported on Windows only", cfg.Mode)
	default:
		return fmt.Errorf("unsupported printer mode %q; use network (usb and serial are Windows only)", cfg.Mode)
	}
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
