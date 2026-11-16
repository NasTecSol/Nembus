package transport_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/NasTecSol/nembus-sap/mappings"
	"github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/transport"
)

func TestFilterDownstreamPriceItems_PolicyMissing404DoesNotSeed(t *testing.T) {
	var putCount atomic.Int32
	var submitCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/auth/login":
			_ = json.NewEncoder(w).Encode(map[string]any{"token": "test-jwt"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/approvals/policies/catalog":
			http.NotFound(w, r)
		case r.Method == http.MethodPut:
			putCount.Add(1)
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/approvals/submit":
			submitCount.Add(1)
			w.WriteHeader(http.StatusCreated)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := transport.NewCloudClient(config.CloudConfig{
		BaseURL:        server.URL,
		APIKey:         "test-key",
		OrganizationID: 1,
		TimeoutSeconds: 5,
	})

	items := []mappings.CanonicalPriceListItem{
		{ProductSKU: "SKU-1", PriceListCode: "PL-1", Price: 80},
	}

	direct, held := client.FilterDownstreamPriceItems(context.Background(), items)
	if putCount.Load() != 0 {
		t.Fatalf("expected 0 PUT calls on 404 policy lookup, got %d", putCount.Load())
	}
	if submitCount.Load() != 0 {
		t.Fatalf("expected 0 approval submits when policy missing, got %d", submitCount.Load())
	}
	if len(held) != 0 {
		t.Fatalf("expected no held notices, got %v", held)
	}
	if len(direct) != 1 || direct[0].ProductSKU != "SKU-1" {
		t.Fatalf("expected SKU-1 in directItems, got %+v", direct)
	}
}

func TestFilterDownstreamPriceItems_PolicyDisabledOrNoApproval(t *testing.T) {
	cases := []struct {
		name             string
		isEnabled        bool
		requiresApproval bool
	}{
		{name: "disabled", isEnabled: false, requiresApproval: true},
		{name: "requires_approval_false", isEnabled: true, requiresApproval: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var submitCount atomic.Int32
			var searchCount atomic.Int32

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodPost && r.URL.Path == "/api/auth/login":
					_ = json.NewEncoder(w).Encode(map[string]any{"token": "test-jwt"})
				case r.Method == http.MethodGet && r.URL.Path == "/api/approvals/policies/catalog":
					_ = json.NewEncoder(w).Encode(map[string]any{
						"is_enabled":        tc.isEnabled,
						"requires_approval": tc.requiresApproval,
						"rules": map[string]any{
							"price_drop_requires_approval": true,
							"below_cost_requires_approval": true,
						},
					})
				case r.Method == http.MethodGet && r.URL.Path == "/api/products/master-catalog/search":
					searchCount.Add(1)
					http.NotFound(w, r)
				case r.Method == http.MethodPost && r.URL.Path == "/api/approvals/submit":
					submitCount.Add(1)
					w.WriteHeader(http.StatusCreated)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()

			client := transport.NewCloudClient(config.CloudConfig{
				BaseURL:        server.URL,
				APIKey:         "test-key",
				OrganizationID: 1,
				TimeoutSeconds: 5,
			})

			items := []mappings.CanonicalPriceListItem{
				{ProductSKU: "SKU-2", PriceListCode: "PL-1", Price: 50},
			}

			direct, held := client.FilterDownstreamPriceItems(context.Background(), items)
			if searchCount.Load() != 0 || submitCount.Load() != 0 {
				t.Fatalf("expected bypass without search/submit, got search=%d submit=%d", searchCount.Load(), submitCount.Load())
			}
			if len(held) != 0 {
				t.Fatalf("expected no held notices, got %v", held)
			}
			if len(direct) != 1 || direct[0].ProductSKU != "SKU-2" {
				t.Fatalf("expected SKU-2 in directItems, got %+v", direct)
			}
		})
	}
}

func TestFilterDownstreamPriceItems_BelowCostRuleWithCostLookup(t *testing.T) {
	var submittedBodies []map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/auth/login":
			_ = json.NewEncoder(w).Encode(map[string]any{"token": "test-jwt"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/approvals/policies/catalog":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"is_enabled":        true,
				"requires_approval": true,
				"rules": map[string]any{
					"price_drop_requires_approval": false,
					"below_cost_requires_approval": true,
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/products/master-catalog/search":
			q := r.URL.Query().Get("q")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"data": []map[string]any{
						{
							"product_id": 101,
							"sku":        q,
							"name":       "Test Item " + q,
							"prices": []map[string]any{
								{
									"price_list_id":   1,
									"price_list_code": "PL-1",
									"price_list_name": "Retail",
									"price":           100.0,
									"is_active":       true,
								},
							},
						},
					},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/product-prices/product/101/with-pricing":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"cost_price": 60.0,
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/approvals":
			_ = json.NewEncoder(w).Encode([]any{})
		case r.Method == http.MethodPost && r.URL.Path == "/api/approvals/submit":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode submit payload: %v", err)
			}
			submittedBodies = append(submittedBodies, body)
			w.WriteHeader(http.StatusCreated)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := transport.NewCloudClient(config.CloudConfig{
		BaseURL:        server.URL,
		APIKey:         "test-key",
		OrganizationID: 1,
		TimeoutSeconds: 5,
	})

	items := []mappings.CanonicalPriceListItem{
		{ProductSKU: "SKU-ABOVE-COST", PriceListCode: "PL-1", Price: 80},
		{ProductSKU: "SKU-BELOW-COST", PriceListCode: "PL-1", Price: 50},
	}

	direct, held := client.FilterDownstreamPriceItems(context.Background(), items)
	if len(direct) != 1 || direct[0].ProductSKU != "SKU-ABOVE-COST" {
		t.Fatalf("expected only SKU-ABOVE-COST in directItems, got %+v", direct)
	}
	if len(held) != 1 {
		t.Fatalf("expected 1 held notice, got %v", held)
	}
	if len(submittedBodies) != 1 {
		t.Fatalf("expected 1 submitted approval, got %d", len(submittedBodies))
	}
	payload, _ := submittedBodies[0]["payload"].(map[string]any)
	if payload == nil || payload["sku"] != "SKU-BELOW-COST" {
		t.Fatalf("expected held SKU-BELOW-COST payload, got %+v", submittedBodies[0])
	}
	if payload["old_cost_price"] != 60.0 {
		t.Fatalf("expected old_cost_price=60.0 from cost lookup, got %v", payload["old_cost_price"])
	}
}
