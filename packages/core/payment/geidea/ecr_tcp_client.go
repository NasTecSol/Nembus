package geidea

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var (
	ErrConnectionClosed = errors.New("terminal tcp connection closed")
	ErrFramingLength    = errors.New("invalid message length frame header")
	ErrTerminalTimeout  = errors.New("terminal communication timeout")
)

// FrameMessage packs a JSON payload with a 4-byte big-endian length prefix.
func FrameMessage(payload []byte) []byte {
	length := uint32(len(payload))
	framed := make([]byte, 4+length)
	binary.BigEndian.PutUint32(framed[0:4], length)
	copy(framed[4:], payload)
	return framed
}

// ReadFramedMessage reads exactly one 4-byte big-endian length-prefixed packet.
func ReadFramedMessage(r io.Reader) ([]byte, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(header)
	if length == 0 || length > 10*1024*1024 { // 10MB safety bound
		return nil, ErrFramingLength
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}
	return body, nil
}

type EcrTcpClient struct {
	addr        string
	readTimeout time.Duration
	mu          sync.Mutex
	conn        net.Conn
	statusHook  func(status string)
}

func NewEcrTcpClient(addr string, readTimeout time.Duration, statusHook func(status string)) *EcrTcpClient {
	if readTimeout <= 0 {
		readTimeout = 90 * time.Second
	}
	return &EcrTcpClient{
		addr:        addr,
		readTimeout: readTimeout,
		statusHook:  statusHook,
	}
}

func (c *EcrTcpClient) connect(ctx context.Context) (net.Conn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return c.conn, nil
	}
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial terminal %s: %w", c.addr, err)
	}
	c.conn = conn
	return c.conn, nil
}

func (c *EcrTcpClient) closeConn() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}

func (c *EcrTcpClient) notify(status string) {
	if c.statusHook != nil {
		c.statusHook(status)
	}
}

func (c *EcrTcpClient) roundTrip(ctx context.Context, cmd string, payload map[string]interface{}) (*TransactionResponse, error) {
	conn, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}

	payload["command"] = cmd
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	framed := FrameMessage(data)
	deadline := time.Now().Add(c.readTimeout)
	_ = conn.SetDeadline(deadline)

	c.notify("WAITING_FOR_CARD")
	if _, err := conn.Write(framed); err != nil {
		c.closeConn()
		return nil, fmt.Errorf("tcp write failed: %w", err)
	}

	c.notify("PROCESSING")
	respBody, err := ReadFramedMessage(conn)
	if err != nil {
		c.closeConn()
		return nil, fmt.Errorf("tcp read failed: %w", err)
	}

	var resp TransactionResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse terminal response: %w", err)
	}
	resp.RawResponse = respBody
	return &resp, nil
}

func (c *EcrTcpClient) Purchase(ctx context.Context, req *PurchaseRequest) (*TransactionResponse, error) {
	return c.roundTrip(ctx, "PURCHASE", map[string]interface{}{
		"ecr_ref": req.ECRReferenceID,
		"amount":  req.Amount,
		"cur":     req.Currency,
	})
}

func (c *EcrTcpClient) Refund(ctx context.Context, req *RefundRequest) (*TransactionResponse, error) {
	return c.roundTrip(ctx, "REFUND", map[string]interface{}{
		"ecr_ref": req.ECRReferenceID,
		"rrn":     req.OriginalRRN,
		"amount":  req.Amount,
	})
}

func (c *EcrTcpClient) Void(ctx context.Context, req *VoidRequest) (*TransactionResponse, error) {
	return c.roundTrip(ctx, "VOID", map[string]interface{}{
		"ecr_ref": req.ECRReferenceID,
		"rrn":     req.OriginalRRN,
	})
}

func (c *EcrTcpClient) QueryStatus(ctx context.Context, ecrRef string) (*TransactionResponse, error) {
	return c.roundTrip(ctx, "QUERY_STATUS", map[string]interface{}{
		"ecr_ref": ecrRef,
	})
}

func (c *EcrTcpClient) ReconcileBatch(ctx context.Context) (*ReconciliationResponse, error) {
	conn, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(map[string]interface{}{"command": "RECONCILE"})
	if _, err := conn.Write(FrameMessage(data)); err != nil {
		c.closeConn()
		return nil, err
	}
	respBody, err := ReadFramedMessage(conn)
	if err != nil {
		c.closeConn()
		return nil, err
	}
	var recon ReconciliationResponse
	if err := json.Unmarshal(respBody, &recon); err != nil {
		return nil, err
	}
	return &recon, nil
}

func (c *EcrTcpClient) CancelInFlight(ctx context.Context, ecrRef string) (*TransactionResponse, error) {
	return c.roundTrip(ctx, "CANCEL", map[string]interface{}{
		"ecr_ref": ecrRef,
	})
}

func (c *EcrTcpClient) Ping(ctx context.Context) error {
	conn, err := c.connect(ctx)
	if err != nil {
		return err
	}
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	data, _ := json.Marshal(map[string]interface{}{"command": "PING"})
	if _, err := conn.Write(FrameMessage(data)); err != nil {
		c.closeConn()
		return err
	}
	_, err = ReadFramedMessage(conn)
	if err != nil {
		c.closeConn()
		return err
	}
	return nil
}
