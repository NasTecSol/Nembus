package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NasTecSol/nembus-core/repository"
)

func TestZatcaCredentialsRejectPlaintextRemoteURL(t *testing.T) {
	for _, endpoint := range []string{"http://example.com", "http://127.0.0.1.example.com", "http://localhost.example.com", "grpc://example.com:50051", "https://user:password@example.com", "https://example.com?secret=value"} {
		if _, err := NewZatcaCloud(context.Background(), endpoint, "test-tenant"); err == nil {
			t.Errorf("unsafe cloud endpoint accepted: %s", endpoint)
		}
	}
}

func TestZatcaPushRequiresDurableMatchingAcknowledgment(t *testing.T) {
	for _, tc := range []struct {
		name string
		data string
		ok   bool
	}{
		{"old acknowledge-only stub", `{"synced_ids":[42],"count":1}`, false},
		{"wrong item", `{"zatca_sync_version":1,"synced_ids":[43]}`, false},
		{"missing item", `{"zatca_sync_version":1,"synced_ids":[]}`, false},
		{"persisted item", `{"zatca_sync_version":1,"synced_ids":[42]}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer test-token" || r.Header.Get("x-tenant-id") != "test-tenant" {
					t.Error("missing sync authentication")
				}
				_, _ = w.Write([]byte(`{"data":` + tc.data + `}`))
			}))
			defer server.Close()
			client := &ZatcaCloud{baseURL: server.URL, tenant: "test-tenant", token: "test-token", http: server.Client()}
			err := client.Push(context.Background(), OutboxItem{ID: 42, EntityID: 1, EntityType: "zatca_device_configs", Action: "UPDATE"}, repository.ZatcaSyncSnapshot{Device: json.RawMessage(`{"store_id":1}`)})
			if (err == nil) != tc.ok {
				t.Fatalf("durability check: %v", err)
			}
		})
	}
}
