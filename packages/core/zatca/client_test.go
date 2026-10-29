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

func TestAllZatcaRequestsSendRequiredHeaders(t *testing.T) {
	tests := []struct {
		path string
		call func(*ZatcaClient) error
	}{
		{"/compliance", func(c *ZatcaClient) error {
			_, err := c.RequestComplianceCSID(context.Background(), "CSR", "test-otp")
			return err
		}},
		{"/compliance/invoices", func(c *ZatcaClient) error {
			_, err := c.SubmitComplianceInvoice(context.Background(), "token", "secret", "hash", "uuid", "xml")
			return err
		}},
		{"/production/csids", func(c *ZatcaClient) error {
			_, err := c.RequestProductionCSID(context.Background(), "token", "secret", "request-id")
			return err
		}},
		{"/invoices/clearance/single", func(c *ZatcaClient) error {
			_, err := c.ClearStandardInvoice(context.Background(), "token", "secret", "hash", "uuid", "xml")
			return err
		}},
		{"/invoices/reporting/single", func(c *ZatcaClient) error {
			_, err := c.ReportSimplifiedInvoice(context.Background(), "token", "secret", "hash", "uuid", "xml")
			return err
		}},
		{"/production/csids/renewal", func(c *ZatcaClient) error {
			_, err := c.RenewProductionCSID(context.Background(), "token", "secret", "CSR")
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			called := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				if r.Method != http.MethodPost || r.URL.Path != test.path {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				for header, want := range map[string]string{
					"Accept-Version": "V2", "Accept": "application/json",
					"Content-Type": "application/json", "Accept-Language": "en",
				} {
					if got := r.Header.Get(header); got != want {
						t.Errorf("%s = %q; want %q", header, got, want)
						w.WriteHeader(http.StatusNotAcceptable)
						return
					}
				}
				if test.path == "/compliance" {
					if r.Header.Get("OTP") != "test-otp" {
						t.Error("missing OTP")
					}
				} else if user, password, ok := r.BasicAuth(); !ok || user != "token" || password != "secret" {
					t.Error("missing or changed Basic authentication")
				}
				if test.path == "/invoices/clearance/single" && r.Header.Get("Clearance-Status") != "1" {
					t.Error("missing clearance status")
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			}))
			defer server.Close()
			if err := test.call(NewZatcaClient(server.URL)); err != nil {
				t.Fatal(err)
			}
			if !called {
				t.Fatal("expected an HTTP request")
			}
		})
	}
}
