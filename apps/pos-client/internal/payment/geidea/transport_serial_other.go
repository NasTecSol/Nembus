//go:build !windows

package geidea

import (
	"fmt"
	"time"
)

type SerialTransport struct{}

func NewSerialTransport(port string, baudRate int) (*SerialTransport, error) {
	return nil, fmt.Errorf("serial COM port communication is only supported on Windows")
}

func (s *SerialTransport) Write(data []byte) error {
	return fmt.Errorf("serial transport not available on this platform")
}

func (s *SerialTransport) Read(buf []byte) (int, error) {
	return 0, fmt.Errorf("serial transport not available on this platform")
}

func (s *SerialTransport) SetReadDeadline(deadline time.Time) error {
	return nil
}

func (s *SerialTransport) Close() error {
	return nil
}
