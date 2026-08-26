package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap-agent/internal/transport"
	"github.com/NasTecSol/nembus-sap-agent/ui"
)

func TestReviewProxyRoutesUseCloudClientAndDoNotExposeToken(t *testing.T) {
	cloud := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/product-enrichment/suggestions" {
			t.Fatalf("unexpected cloud path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer private-token" {
			t.Fatalf("expected private token to be attached by CloudClient")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"items":[]}}`))
	}))
	defer cloud.Close()

	cfg := config.DefaultConfig()
	cfg.Cloud = config.CloudConfig{BaseURL: cloud.URL, M2MToken: "private-token", TenantSlug: "tenant-a", OrganizationID: 1, TimeoutSeconds: 5}
	srv := NewServer(cfg, nil, nil, nil, transport.NewCloudClient(cfg.Cloud), ui.FS)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/product-enrichment/suggestions?status=in_review", nil)
	res := httptest.NewRecorder()
	srv.router.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"items"`) {
		t.Fatalf("unexpected review proxy response: status=%d body=%s", res.Code, res.Body.String())
	}

	configReq := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
	configRes := httptest.NewRecorder()
	srv.router.ServeHTTP(configRes, configReq)
	if strings.Contains(configRes.Body.String(), "private-token") || strings.Contains(configRes.Body.String(), "m2m_token") {
		t.Fatalf("cloud bearer leaked through config endpoint: %s", configRes.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(configRes.Body.Bytes(), &payload); err != nil {
		t.Fatalf("config response is not JSON: %v", err)
	}
}
