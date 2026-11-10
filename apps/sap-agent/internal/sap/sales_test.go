package sap_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NasTecSol/nembus-sap-agent/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/sap"
)

func setupMockSAPServer(t *testing.T, creditNotesHandler func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, *sap.Client) {
	mux := http.NewServeMux()

	// Mock SAP Service Layer Login
	mux.HandleFunc("/b1s/v1/Login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"SessionId":      "mock-session-id-12345",
			"Version":        "10.0",
			"SessionTimeout": 30,
		})
	})

	if creditNotesHandler != nil {
		mux.HandleFunc("/b1s/v1/CreditNotes", creditNotesHandler)
	}

	server := httptest.NewServer(mux)

	cfg := &config.Config{
		SAPServiceLayerURL: server.URL,
		SAPCompanyDB:       "TEST_DB",
		SAPUsername:        "manager",
		SAPPassword:        "password",
		SAPInsecureSkipSSL: true,
	}

	client, err := sap.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create sap client: %v", err)
	}

	return server, client
}

func TestPostCreditMemo_Linked(t *testing.T) {
	var receivedPayload map[string]any

	server, client := setupMockSAPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		if err := json.Unmarshal(body, &receivedPayload); err != nil {
			t.Fatalf("failed to unmarshal request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sap.SAPInvoiceResponse{
			DocEntry: 9876,
			DocNum:   5432,
			CardCode: "C1001",
		})
	})
	defer server.Close()

	ctx := context.Background()
	baseEntry := 1234
	baseType := 13 // 13 = A/R Invoice
	lineNum := 0

	req := &sap.SAPInvoiceRequest{
		CardCode:   "C1001",
		DocDate:    "2026-10-22",
		DocDueDate: "2026-10-22",
		TaxDate:    "2026-10-22",
		NumAtCard:  "RET-001",
		Comments:   "Return for original invoice 1234",
		DocumentLines: []sap.SAPDocumentLine{
			{
				LineNum:         &lineNum,
				ItemCode:        "ITEM-ABC",
				ItemDescription: "Test Item",
				Quantity:        2.0,
				UnitPrice:       50.0,
				PriceAfterVAT:   57.5,
				WarehouseCode:   "01",
				BaseEntry:       &baseEntry,
				BaseType:        &baseType,
			},
		},
	}

	resp, err := client.PostCreditMemo(ctx, req)
	if err != nil {
		t.Fatalf("PostCreditMemo failed: %v", err)
	}

	if resp.DocEntry != 9876 || resp.DocNum != 5432 {
		t.Errorf("unexpected response: docEntry=%d, docNum=%d", resp.DocEntry, resp.DocNum)
	}

	// Verify payload received by mock SAP Service Layer
	if receivedPayload["CardCode"] != "C1001" {
		t.Errorf("expected CardCode C1001, got %v", receivedPayload["CardCode"])
	}
	lines, ok := receivedPayload["DocumentLines"].([]any)
	if !ok || len(lines) != 1 {
		t.Fatalf("expected 1 document line, got %v", receivedPayload["DocumentLines"])
	}

	lineMap, ok := lines[0].(map[string]any)
	if !ok {
		t.Fatalf("invalid line format")
	}

	if lineMap["ItemCode"] != "ITEM-ABC" {
		t.Errorf("expected ItemCode ITEM-ABC, got %v", lineMap["ItemCode"])
	}

	baseEntryReceived, ok := lineMap["BaseEntry"].(float64)
	if !ok || int(baseEntryReceived) != 1234 {
		t.Errorf("expected BaseEntry 1234, got %v", lineMap["BaseEntry"])
	}

	baseTypeReceived, ok := lineMap["BaseType"].(float64)
	if !ok || int(baseTypeReceived) != 13 {
		t.Errorf("expected BaseType 13, got %v", lineMap["BaseType"])
	}
}

func TestPostCreditMemo_Standalone(t *testing.T) {
	var receivedPayload map[string]any

	server, client := setupMockSAPServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		if err := json.Unmarshal(body, &receivedPayload); err != nil {
			t.Fatalf("failed to unmarshal request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sap.SAPInvoiceResponse{
			DocEntry: 9877,
			DocNum:   5433,
			CardCode: "C000001",
		})
	})
	defer server.Close()

	ctx := context.Background()
	req := &sap.SAPInvoiceRequest{
		CardCode:  "C000001",
		NumAtCard: "RET-STANDALONE",
		DocumentLines: []sap.SAPDocumentLine{
			{
				ItemCode:      "ITEM-XYZ",
				Quantity:      1.0,
				UnitPrice:     100.0,
				WarehouseCode: "01",
			},
		},
	}

	resp, err := client.PostCreditMemo(ctx, req)
	if err != nil {
		t.Fatalf("PostCreditMemo failed: %v", err)
	}

	if resp.DocEntry != 9877 || resp.DocNum != 5433 {
		t.Errorf("unexpected response: docEntry=%d, docNum=%d", resp.DocEntry, resp.DocNum)
	}

	// Verify BaseEntry and BaseType are omitted for standalone memo
	lines := receivedPayload["DocumentLines"].([]any)
	lineMap := lines[0].(map[string]any)
	if _, exists := lineMap["BaseEntry"]; exists {
		t.Errorf("BaseEntry should be omitted for standalone credit memo")
	}
	if _, exists := lineMap["BaseType"]; exists {
		t.Errorf("BaseType should be omitted for standalone credit memo")
	}

	// Verify dates were defaulted
	if receivedPayload["DocDate"] == "" || receivedPayload["DocDueDate"] == "" || receivedPayload["TaxDate"] == "" {
		t.Errorf("expected default dates to be populated, got DocDate=%v, DocDueDate=%v, TaxDate=%v",
			receivedPayload["DocDate"], receivedPayload["DocDueDate"], receivedPayload["TaxDate"])
	}
}

func TestPostCreditMemo_ErrorHandling(t *testing.T) {
	server, client := setupMockSAPServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code": -5002,
				"message": map[string]any{
					"lang":  "en-us",
					"value": "Item quantity cannot be negative",
				},
			},
		})
	})
	defer server.Close()

	ctx := context.Background()
	req := &sap.SAPInvoiceRequest{
		CardCode: "C1001",
		DocumentLines: []sap.SAPDocumentLine{
			{
				ItemCode: "ITEM-ERR",
				Quantity: -1.0,
			},
		},
	}

	_, err := client.PostCreditMemo(ctx, req)
	if err == nil {
		t.Fatal("expected error from PostCreditMemo, got nil")
	}
}
