package geidea

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Client handles high-level communication with a Geidea payment terminal
type Client struct {
	config TerminalConfig
	mu     sync.Mutex
}

// NewClient creates a new Geidea payment client with the specified configuration
func NewClient(config TerminalConfig) *Client {
	if config.Mode == "" {
		config.Mode = "network"
	}
	if config.Port == "" {
		config.Port = "80"
	}
	if config.BaudRate <= 0 {
		config.BaudRate = 38400
	}
	if config.TimeoutSeconds <= 0 {
		config.TimeoutSeconds = 120 // 2 minutes default timeout for customer card interaction
	}
	return &Client{
		config: config,
	}
}

// UpdateConfig updates the client's connection configuration
func (c *Client) UpdateConfig(cfg TerminalConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = cfg
}

// openTransport opens the appropriate transport (TCP or Serial) based on configuration
func (c *Client) openTransport(ctx context.Context) (Transport, error) {
	c.mu.Lock()
	cfg := c.config
	c.mu.Unlock()

	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second

	switch strings.ToLower(cfg.Mode) {
	case "serial":
		return NewSerialTransport(cfg.SerialPort, cfg.BaudRate)
	case "network", "tcp":
		if cfg.IP == "" {
			return nil, fmt.Errorf("terminal IP address is required for network mode")
		}
		return NewTCPTransport(ctx, cfg.IP, cfg.Port, timeout)
	default:
		return nil, fmt.Errorf("unsupported transport mode %q; use 'network' or 'serial'", cfg.Mode)
	}
}

// executeExchange manages the complete send-wait-read lifecycle of a terminal command
func (c *Client) executeExchange(ctx context.Context, reqBytes []byte, onStatus StatusCallback) ([]byte, error) {
	c.mu.Lock()
	timeoutSec := c.config.TimeoutSeconds
	c.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	transport, err := c.openTransport(ctx)
	if err != nil {
		return nil, err
	}
	defer transport.Close()

	// 1. Send the request buffer
	if err := transport.Write(reqBytes); err != nil {
		return nil, fmt.Errorf("failed to send request to terminal: %w", err)
	}

	readChunk := make([]byte, 1024)
	var accumulated []byte
	ackReceived := false
	responseStarted := false
	var lastStatusByte byte

	// Read loop with timeout and progress detection
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("transaction timed out or cancelled: %w", ctx.Err())
		default:
		}

		_ = transport.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, err := transport.Read(readChunk)
		if n > 0 {
			chunk := readChunk[:n]

			// Check ACK / NAK
			if !ackReceived {
				for i, b := range chunk {
					if b == ByteACK {
						ackReceived = true
						chunk = chunk[i+1:]
						break
					} else if b == ByteNAK {
						return nil, fmt.Errorf("terminal rejected message with NAK (0x15)")
					}
				}
			}

			if len(chunk) == 0 {
				continue
			}

			// Check for intermediate status updates (6-byte packets: 01 00 02 FF <Code> <LRC>)
			if statusCode, ok := ParseIntermediateStatus(chunk); ok {
				if statusCode != lastStatusByte {
					lastStatusByte = statusCode
					if onStatus != nil {
						onStatus(GetStatusEvent(statusCode))
					}
				}
			}

			// Check for Terminal Busy
			if IsTerminalBusyResponse(chunk) {
				return nil, fmt.Errorf("terminal is busy with another transaction")
			}

			// Check for response start (contains tag 'DF')
			if bytes.Contains(chunk, []byte{0xDF}) {
				responseStarted = true
			}

			if responseStarted {
				accumulated = append(accumulated, chunk...)

				// If we have received a substantial TLV response and either see Idle Screen (0x39)
				// or the tags contain response code DF27, we can check completion
				if len(accumulated) >= 30 && bytes.Contains(accumulated, []byte{0xDF, 0x27}) {
					// Check if tags can be extracted successfully
					tags := ExtractTLVs(accumulated)
					if _, hasRespCode := tags["DF27"]; hasRespCode {
						// Give a tiny window (50ms) to ensure remaining buffer bytes arrive
						time.Sleep(50 * time.Millisecond)
						return accumulated, nil
					}
				}
			}
		}

		if err != nil {
			if err == io.EOF {
				if responseStarted && len(accumulated) > 0 {
					return accumulated, nil
				}
				break
			}
			// On read timeout, continue checking context deadline
			continue
		}
	}

	if len(accumulated) > 0 {
		return accumulated, nil
	}
	return nil, fmt.Errorf("connection closed by terminal without complete response")
}

// Purchase initiates a card payment on the terminal and streams progress updates
func (c *Client) Purchase(ctx context.Context, req PurchaseRequest, onStatus StatusCallback) (*TransactionResult, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("purchase amount must be greater than zero")
	}

	reqBuf, err := BuildPurchaseBuffer(req.Amount, req.InvoiceNumber, req.CashbackAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to build purchase request: %w", err)
	}

	rawResp, err := c.executeExchange(ctx, reqBuf, onStatus)
	if err != nil {
		return nil, err
	}

	return ParseTransactionResponse(rawResp)
}

// Refund performs a refund to the cardholder using the original transaction RRN
func (c *Client) Refund(ctx context.Context, req RefundRequest, onStatus StatusCallback) (*TransactionResult, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("refund amount must be greater than zero")
	}
	if strings.TrimSpace(req.RRN) == "" {
		return nil, fmt.Errorf("RRN (Retrieval Reference Number) is required for refund")
	}

	reqBuf, err := BuildRefundBuffer(req.Amount, req.InvoiceNumber, req.RRN)
	if err != nil {
		return nil, fmt.Errorf("failed to build refund request: %w", err)
	}

	rawResp, err := c.executeExchange(ctx, reqBuf, onStatus)
	if err != nil {
		return nil, err
	}

	return ParseTransactionResponse(rawResp)
}

// Reconcile triggers End of Day batch settlement and reconciliation with SAMA
func (c *Client) Reconcile(ctx context.Context, onStatus StatusCallback) (*ReconciliationResult, error) {
	reqBuf := BuildReconciliationBuffer()

	rawResp, err := c.executeExchange(ctx, reqBuf, onStatus)
	if err != nil {
		return nil, err
	}

	return ParseReconciliationResponse(rawResp)
}

// RepeatLast queries the terminal for the details of the last executed transaction
func (c *Client) RepeatLast(ctx context.Context) (*TransactionResult, error) {
	reqBuf := BuildRepeatLastBuffer()

	rawResp, err := c.executeExchange(ctx, reqBuf, nil)
	if err != nil {
		return nil, err
	}

	return ParseTransactionResponse(rawResp)
}

// Cancel sends the abort command to cancel an in-progress transaction
func (c *Client) Cancel(ctx context.Context) error {
	cancelBuf := BuildCancelBuffer()

	transport, err := c.openTransport(ctx)
	if err != nil {
		return err
	}
	defer transport.Close()

	if err := transport.Write(cancelBuf); err != nil {
		return fmt.Errorf("failed to send cancel request to terminal: %w", err)
	}
	return nil
}

// CheckStatus pings the terminal to verify connectivity
func (c *Client) CheckStatus(ctx context.Context) (bool, error) {
	statusBuf := BuildCheckStatusBuffer()

	transport, err := c.openTransport(ctx)
	if err != nil {
		return false, err
	}
	defer transport.Close()

	if err := transport.Write(statusBuf); err != nil {
		return false, err
	}

	readBuf := make([]byte, 128)
	_ = transport.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, err := transport.Read(readBuf)
	if err != nil {
		return false, fmt.Errorf("no response from terminal: %w", err)
	}

	resp := readBuf[:n]
	// If response contains ACK (0x06) or 0xDF 0x85 (Terminal Available tag)
	if bytes.Contains(resp, []byte{ByteACK}) || bytes.Contains(resp, []byte{0xDF, 0x85}) {
		return true, nil
	}
	return false, nil
}
