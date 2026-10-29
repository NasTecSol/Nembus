package geidea

import (
	"context"
	"encoding/hex"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// mockTerminalServer simulates the physical Geidea EFT terminal over TCP
func startMockTerminal(t *testing.T) (string, func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock terminal server: %v", err)
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleMockTerminalConn(conn)
		}
	}()

	return ln.Addr().String(), func() { _ = ln.Close() }
}

func handleMockTerminalConn(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		return
	}

	req := buf[:n]

	// 1. Send immediate ACK (0x06)
	_, _ = conn.Write([]byte{ByteACK})
	time.Sleep(20 * time.Millisecond)

	// Check if this is a Status Check request
	if len(req) >= 4 && req[2] == 0x03 && req[5] == 0x04 {
		// Terminal Available response: DF 85 02 30 30 AA 01 00 02 FF 39 C4
		resp, _ := hex.DecodeString("DF85023030AA010002FF39C4")
		_, _ = conn.Write(resp)
		return
	}

	// 2. Stream intermediate progress events
	// Swipe/Insert: 01 00 02 FF 41 BE
	_, _ = conn.Write([]byte{0x01, 0x00, 0x02, 0xFF, 0x41, 0x02 ^ 0xFF ^ 0x41})
	time.Sleep(30 * time.Millisecond)

	// PIN Entry: 01 00 02 FF 43 BC
	_, _ = conn.Write([]byte{0x01, 0x00, 0x02, 0xFF, 0x43, 0x02 ^ 0xFF ^ 0x43})
	time.Sleep(30 * time.Millisecond)

	// Connecting to SAMA: 01 00 02 FF 46 BB
	_, _ = conn.Write([]byte{0x01, 0x00, 0x02, 0xFF, 0x46, 0x02 ^ 0xFF ^ 0x46})
	time.Sleep(30 * time.Millisecond)

	// Waiting Response: 01 00 02 FF 79 84
	_, _ = conn.Write([]byte{0x01, 0x00, 0x02, 0xFF, 0x79, 0x02 ^ 0xFF ^ 0x79})
	time.Sleep(30 * time.Millisecond)

	// 3. Send final transaction response (Approved SAR 50.00 on mada card)
	// DF02 (Masked Card), DF04 (50.00 SAR), DF0B (STAN), DF0C (DateTime), DF0E (Expiry),
	// DF25 (RRN), DF26 (Auth), DF27 (000), DF29 (TID), DF82 (P1), DF81 (ECR)
	sampleResp := "DF0210353838382A2A2A2A2A2A2A2A33323235" +
		"DF0406000000005000" +
		"DF0B06303030303438" +
		"DF0C0C323630343138313431393339" +
		"DF0E021012" +
		"DF250C383032313137323730303038" +
		"DF2606313630393339" +
		"DF2703303030" +
		"DF291036333435303030313031343536393638" +
		"DF82025031" +
		"DF870442534642" +
		"DF880435343131" +
		"DF903C5350414E7C4449505045447C41303030303030323238313031307C303834303034313030307C45383030" +
		"DF811030303030303030303030303030303031"

	finalBytes, _ := hex.DecodeString(sampleResp)
	_, _ = conn.Write(finalBytes)
}

func TestClientPurchaseFlow(t *testing.T) {
	addr, cleanup := startMockTerminal(t)
	defer cleanup()

	host, port, _ := net.SplitHostPort(addr)

	client := NewClient(TerminalConfig{
		Mode:           "network",
		IP:             host,
		Port:           port,
		TimeoutSeconds: 10,
	})

	var mu sync.Mutex
	var receivedStatuses []byte

	res, err := client.Purchase(context.Background(), PurchaseRequest{
		Amount:        50.00,
		InvoiceNumber: "1",
	}, func(evt TerminalStatusEvent) {
		mu.Lock()
		defer mu.Unlock()
		receivedStatuses = append(receivedStatuses, evt.Code)
	})

	if err != nil {
		t.Fatalf("Purchase failed: %v", err)
	}

	if !res.Success {
		t.Errorf("expected success=true, got false")
	}
	if res.ResponseCode != "000" {
		t.Errorf("expected code 000, got %s", res.ResponseCode)
	}
	if res.Amount != 50.00 {
		t.Errorf("expected amount 50.00, got %f", res.Amount)
	}
	if res.CardScheme != "mada" {
		t.Errorf("expected scheme mada, got %s", res.CardScheme)
	}
	if res.CardEntryMode != "DIPPED" {
		t.Errorf("expected entry mode DIPPED, got %s", res.CardEntryMode)
	}
	if res.BankID != "BSFB" {
		t.Errorf("expected bank BSFB, got %s", res.BankID)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(receivedStatuses) < 2 {
		t.Errorf("expected at least 2 status callbacks, got %d", len(receivedStatuses))
	}
}

func TestClientCheckStatus(t *testing.T) {
	addr, cleanup := startMockTerminal(t)
	defer cleanup()

	host, port, _ := net.SplitHostPort(addr)

	client := NewClient(TerminalConfig{
		Mode:           "network",
		IP:             host,
		Port:           port,
		TimeoutSeconds: 5,
	})

	ok, err := client.CheckStatus(context.Background())
	if err != nil {
		t.Fatalf("CheckStatus error: %v", err)
	}
	if !ok {
		t.Errorf("expected CheckStatus=true, got false")
	}
}

func TestClientValidation(t *testing.T) {
	client := NewClient(TerminalConfig{
		Mode: "network",
		IP:   "127.0.0.1",
	})

	// Zero or negative amount
	_, err := client.Purchase(context.Background(), PurchaseRequest{Amount: 0}, nil)
	if err == nil || !strings.Contains(err.Error(), "greater than zero") {
		t.Errorf("expected validation error for amount <= 0, got: %v", err)
	}

	// Refund without RRN
	_, err = client.Refund(context.Background(), RefundRequest{Amount: 10, RRN: ""}, nil)
	if err == nil || !strings.Contains(err.Error(), "RRN") {
		t.Errorf("expected validation error for missing RRN, got: %v", err)
	}
}
