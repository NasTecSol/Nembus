package geidea_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/payment/geidea"
)

func Test_TCPFraming(t *testing.T) {
	payload := []byte(`{"command":"PING","timestamp":12345678}`)
	framed := geidea.FrameMessage(payload)

	if len(framed) != 4+len(payload) {
		t.Fatalf("expected framed length %d, got %d", 4+len(payload), len(framed))
	}

	reader := bytes.NewReader(framed)
	readPayload, err := geidea.ReadFramedMessage(reader)
	if err != nil {
		t.Fatalf("unexpected read framing error: %v", err)
	}

	if !bytes.Equal(payload, readPayload) {
		t.Fatalf("framed payload mismatch: got %s, want %s", string(readPayload), string(payload))
	}

	// Partial packet buffering
	half := len(framed) / 2
	chunkedReader := io.MultiReader(bytes.NewReader(framed[:half]), bytes.NewReader(framed[half:]))
	chunkedPayload, err := geidea.ReadFramedMessage(chunkedReader)
	if err != nil {
		t.Fatalf("partial framing read error: %v", err)
	}
	if !bytes.Equal(payload, chunkedPayload) {
		t.Fatalf("chunked payload mismatch: got %s, want %s", string(chunkedPayload), string(payload))
	}

	// Zero length boundary safety
	zeroHeader := make([]byte, 4)
	_, err = geidea.ReadFramedMessage(bytes.NewReader(zeroHeader))
	if err == nil {
		t.Fatal("expected error on zero frame length header")
	}
}

func Test_SimulatorLifecycle(t *testing.T) {
	sim := geidea.NewSimulator(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Normal purchase -> Approved Mada EMV
	resp, err := sim.Purchase(ctx, &geidea.PurchaseRequest{
		ECRReferenceID: "REF-001",
		Amount:         150.00,
		Currency:       "SAR",
	})
	if err != nil {
		t.Fatalf("simulator purchase failed: %v", err)
	}
	if resp.Status != geidea.StatusApproved {
		t.Errorf("expected status %s, got %s", geidea.StatusApproved, resp.Status)
	}
	if resp.CardScheme != "mada" || resp.RRN == "" || resp.AuthCode == "" {
		t.Errorf("missing EMV parameters in response: %+v", resp)
	}

	// 2. Simulated Decline on .99 cents
	respDecline, err := sim.Purchase(ctx, &geidea.PurchaseRequest{
		ECRReferenceID: "REF-002",
		Amount:         99.99,
		Currency:       "SAR",
	})
	if err != nil {
		t.Fatalf("simulator purchase error on decline scenario: %v", err)
	}
	if respDecline.Status != geidea.StatusDeclined {
		t.Errorf("expected status %s, got %s", geidea.StatusDeclined, respDecline.Status)
	}

	// 3. Status Query
	queryResp, err := sim.QueryStatus(ctx, "REF-001")
	if err != nil || queryResp.Status != geidea.StatusApproved {
		t.Errorf("query status failed: resp=%+v err=%v", queryResp, err)
	}

	// 4. Heartbeat / Ping
	if err := sim.Ping(ctx); err != nil {
		t.Errorf("ping failed: %v", err)
	}
}

func Test_UseCase_UncertainFlow(t *testing.T) {
	sim := geidea.NewSimulator(nil)
	sim.SetDropConnection(true)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Trigger purchase while connection is dropped
	_, err := sim.Purchase(ctx, &geidea.PurchaseRequest{
		ECRReferenceID: "DROP-001",
		Amount:         200.00,
		Currency:       "SAR",
	})
	if err == nil {
		t.Fatal("expected error on dropped connection")
	}

	// 2. Query status on dropped transaction returns uncertain / not found
	queryResp, err := sim.QueryStatus(ctx, "DROP-001")
	if err != nil {
		t.Fatalf("query status error: %v", err)
	}
	if queryResp.Status != geidea.StatusUncertain {
		t.Errorf("expected status %s, got %s", geidea.StatusUncertain, queryResp.Status)
	}

	// 3. Recovery via Reversal
	sim.SetDropConnection(false)
	revResp, err := sim.Void(ctx, &geidea.VoidRequest{
		ECRReferenceID: "DROP-001",
		OriginalRRN:    "RRN-9999",
	})
	if err != nil {
		t.Fatalf("void failed: %v", err)
	}
	if revResp.Status != geidea.StatusVoided {
		t.Errorf("expected status %s, got %s", geidea.StatusVoided, revResp.Status)
	}
}
