package zatca

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetBaseURLForEnv(t *testing.T) {
	tests := []struct {
		env      string
		expected string
	}{
		{"production", "https://gw-fatoora.zatca.gov.sa/e-invoicing/core"},
		{"prod", "https://gw-fatoora.zatca.gov.sa/e-invoicing/core"},
		{"simulation", "https://gw-fatoora.zatca.gov.sa/e-invoicing/simulation"},
		{"sim", "https://gw-fatoora.zatca.gov.sa/e-invoicing/simulation"},
		{"sandbox", "https://gw-fatoora.zatca.gov.sa/e-invoicing/developer-portal"},
		{"unknown", "https://gw-fatoora.zatca.gov.sa/e-invoicing/developer-portal"},
	}

	for _, tt := range tests {
		url := GetBaseURLForEnv(tt.env)
		if url != tt.expected {
			t.Errorf("GetBaseURLForEnv(%s) = %s; want %s", tt.env, url, tt.expected)
		}
	}
}

func TestRenewProductionCSID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/production/csids/renewal" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"issuedCertificate": "RENEWED_CERT",
			"binarySecurityToken": "RENEWED_TOKEN",
			"secret": "NEW_SECRET"
		}`))
	}))
	defer ts.Close()

	client := NewZatcaClient(ts.URL)
	resp, err := client.RenewProductionCSID(context.Background(), "CURR_TOKEN", "SECRET", "DUMMY_CSR")
	if err != nil {
		t.Fatalf("RenewProductionCSID failed: %v", err)
	}

	if resp.BinarySecurityToken != "RENEWED_TOKEN" {
		t.Errorf("Expected RENEWED_TOKEN, got %s", resp.BinarySecurityToken)
	}
}
