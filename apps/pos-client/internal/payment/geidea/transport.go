package geidea

import (
	"context"
	"fmt"
	"net"
	"time"
)

// Transport defines the common interface for communicating with a physical terminal
type Transport interface {
	Write(data []byte) error
	Read(buf []byte) (int, error)
	SetReadDeadline(t time.Time) error
	Close() error
}

// TCPTransport implements Transport over Ethernet / Wi-Fi
type TCPTransport struct {
	conn net.Conn
}

// NewTCPTransport dials the terminal over TCP
func NewTCPTransport(ctx context.Context, ip, port string, timeout time.Duration) (*TCPTransport, error) {
	if port == "" {
		port = "80"
	}
	addr := net.JoinHostPort(ip, port)
	dialer := net.Dialer{Timeout: timeout}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Geidea terminal at %s: %w", addr, err)
	}

	return &TCPTransport{conn: conn}, nil
}

func (t *TCPTransport) Write(data []byte) error {
	_, err := t.conn.Write(data)
	return err
}

func (t *TCPTransport) Read(buf []byte) (int, error) {
	return t.conn.Read(buf)
}

func (t *TCPTransport) SetReadDeadline(deadline time.Time) error {
	return t.conn.SetReadDeadline(deadline)
}

func (t *TCPTransport) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}
