package transport_test

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/transport"
	"github.com/NasTecSol/nembus-sap/contracts"
	"github.com/NasTecSol/nembus-sap/mappings"
)

func makeTestPayload() *contracts.MigrationBatchPayload {
	return &contracts.MigrationBatchPayload{
		BatchID:        "b-123",
		RunID:          "r-456",
		OrganizationID: 1,
		Domain:         contracts.DomainStores,
		Stores: []mappings.CanonicalStore{
			{Code: "01", Name: "Warehouse 1", IsActive: true},
		},
		IsLastBatch: true,
		Timestamp:   time.Now(),
	}
}

func TestCloudClientSendBatchGzip(t *testing.T) {
	// Create Mock Cloud Server receiving Gzip payload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/migration/batch" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Content-Encoding") != "gzip" {
			t.Errorf("expected gzip Content-Encoding header")
		}
		if r.Header.Get("Authorization") != "Bearer test-m2m-token" {
			t.Errorf("expected configured M2M bearer authorization")
		}
		if r.Header.Get("x-tenant-id") != "tenant-a" {
			t.Errorf("expected tenant slug consistency header")
		}
		if r.Header.Get("x-organization-id") != "1" {
			t.Errorf("expected organization consistency header")
		}

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			t.Fatalf("failed to decode gzip body: %v", err)
		}
		defer gz.Close()

		bodyBytes, err := io.ReadAll(gz)
		if err != nil {
			t.Fatalf("failed to read decompressed body: %v", err)
		}

		var payload contracts.MigrationBatchPayload
		if err := json.Unmarshal(bodyBytes, &payload); err != nil {
			t.Fatalf("failed to parse json payload: %v", err)
		}

		if payload.Domain != contracts.DomainStores {
			t.Errorf("expected domain 'stores', got %s", payload.Domain)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(contracts.MigrationBatchResponse{
			Success:       true,
			BatchID:       payload.BatchID,
			Domain:        payload.Domain,
			RecordsStaged: len(payload.Stores),
		})
	}))
	defer server.Close()

	client := transport.NewCloudClient(config.CloudConfig{
		BaseURL:        server.URL,
		M2MToken:       "test-m2m-token",
		TenantSlug:     "tenant-a",
		OrganizationID: 1,
		TimeoutSeconds: 5,
	})

	resp, err := client.SendBatch(context.Background(), makeTestPayload())

	if err != nil {
		t.Fatalf("SendBatch failed: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success response")
	}
	if resp.RecordsStaged != 1 {
		t.Errorf("expected 1 record staged, got %d", resp.RecordsStaged)
	}
}

func TestCloudClientRejectsAPIKeyOnlyMigration(t *testing.T) {
	client := transport.NewCloudClient(config.CloudConfig{
		BaseURL:        "http://127.0.0.1:1",
		APIKey:         "legacy-api-key",
		TenantSlug:     "tenant-a",
		OrganizationID: 1,
	})
	if _, err := client.SendBatch(context.Background(), makeTestPayload()); err == nil {
		t.Fatal("expected API-key-only migration to be rejected")
	}
}

func TestCloudClientRequestIDHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Request-ID") == "" {
			t.Errorf("expected X-Request-ID header to be set")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(contracts.MigrationBatchResponse{Success: true})
	}))
	defer server.Close()

	client := transport.NewCloudClient(config.CloudConfig{
		BaseURL:        server.URL,
		M2MToken:       "test-m2m-token",
		TenantSlug:     "tenant-a",
		OrganizationID: 1,
		TimeoutSeconds: 5,
	})
	_, err := client.SendBatch(context.Background(), makeTestPayload())
	if err != nil {
		t.Fatalf("SendBatch failed: %v", err)
	}
}

func TestSendBatchWithRetry_RetriesOn503(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		if n < 3 {
			// Fail first two attempts with 503
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("service unavailable"))
			return
		}
		// Third attempt succeeds
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(contracts.MigrationBatchResponse{
			Success:       true,
			RecordsStaged: 1,
		})
	}))
	defer server.Close()

	client := transport.NewCloudClient(config.CloudConfig{
		BaseURL:        server.URL,
		M2MToken:       "test-m2m-token",
		TenantSlug:     "tenant-a",
		OrganizationID: 1,
		TimeoutSeconds: 5,
	})

	resp, err := client.SendBatchWithRetry(context.Background(), makeTestPayload())
	if err != nil {
		t.Fatalf("SendBatchWithRetry failed: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success on third attempt")
	}
	if callCount.Load() != 3 {
		t.Errorf("expected 3 HTTP calls, got %d", callCount.Load())
	}
}

func TestSendBatchWithRetry_NoRetryOn400(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	client := transport.NewCloudClient(config.CloudConfig{
		BaseURL:        server.URL,
		M2MToken:       "test-m2m-token",
		TenantSlug:     "tenant-a",
		OrganizationID: 1,
		TimeoutSeconds: 5,
	})

	_, err := client.SendBatchWithRetry(context.Background(), makeTestPayload())
	if err == nil {
		t.Fatal("expected error on 400 response")
	}
	if callCount.Load() != 1 {
		t.Errorf("expected exactly 1 HTTP call on 400 (no retry), got %d", callCount.Load())
	}
}
