package zatca

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestOnboardingComplianceFailureDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		name, environment, vat, fallback, template string
		gatewayStatus, responseStatus              int
	}{
		{"simulation organization VAT", "simulation", "310123456700003", "399999999900003", "PREZATCA-Code-Signing", 400, 400},
		{"simulation fallback VAT", "simulation", "", "399999999900003", "PREZATCA-Code-Signing", 400, 400},
		{"production organization VAT", "production", "310123456700003", "", "ZATCA-Code-Signing", 400, 400},
		{"gateway unavailable", "simulation", "310123456700003", "", "PREZATCA-Code-Signing", 503, 502},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wantVAT := tc.vat
			if wantVAT == "" {
				wantVAT = tc.fallback
			}
			db := &onboardingFlowDB{completionDB: completionDB{org: repository.Organization{
				ID: 1, Name: "Test", TaxID: pgtype.Text{String: tc.vat, Valid: tc.vat != ""}, Metadata: []byte(`{"zatca":{"enabled":true}}`),
			}}}
			calls := 0
			gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != "/compliance" || r.Header.Get("OTP") != "654321" || r.Header.Get("Accept-Version") != "V2" {
					t.Error("incorrect compliance request")
				}
				var request ComplianceCSIDRequest
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
				}
				decoded, err := base64.StdEncoding.DecodeString(request.CSR)
				if err != nil {
					t.Error(err)
				}
				block, _ := pem.Decode(decoded)
				if block == nil || !bytes.Contains(block.Bytes, []byte(wantVAT)) || !bytes.Contains(block.Bytes, []byte(tc.template)) {
					t.Error("diagnostics do not match the wire CSR")
				}
				w.WriteHeader(tc.gatewayStatus)
				fmt.Fprint(w, `{"errors":[{"code":"Invalid-OTP","message":"The provided OTP is invalid"}]}`)
			}))
			defer gateway.Close()
			service := testOnboardingService(nil, NewZatcaClient(gateway.URL), &usecase.ZatcaConfig{OrgVATID: tc.fallback})
			ctx := context.WithValue(context.Background(), middleware.RepoKey, repository.New(db))
			device, err := service.OnboardDevice(ctx, 1, nil, nil, "EGS-IOS-TEST", "pos", "654321", tc.environment)
			var failure *onboardingComplianceError
			var upstream *ComplianceAPIError
			if device != nil || !errors.As(err, &failure) || !errors.As(err, &upstream) {
				t.Fatalf("missing gateway diagnostics: %v", err)
			}
			if calls != 1 || db.creates != 0 || db.writes != 0 {
				t.Fatalf("onboarding retried OTP or saved failed device: calls=%d creates=%d writes=%d", calls, db.creates, db.writes)
			}
			if failure.OnboardingHTTPStatus() != tc.responseStatus || upstream.StatusCode != tc.gatewayStatus {
				t.Fatal("incorrect error status")
			}
			var details struct {
				Environment string `json:"zatca_env"`
				VAT         string `json:"csr_organization_identifier"`
				Template    string `json:"csr_certificate_template"`
				Endpoint    string `json:"gateway_endpoint"`
				DeviceSaved bool   `json:"device_saved"`
			}
			raw := failure.OnboardingFailureDetails()
			if err := json.Unmarshal(raw, &details); err != nil {
				t.Fatal(err)
			}
			if details.Environment != tc.environment || details.VAT != wantVAT || details.Template != tc.template || details.Endpoint != gateway.URL+"/compliance" || details.DeviceSaved {
				t.Fatalf("incorrect diagnostics: %s", raw)
			}
			for _, secret := range []string{"654321", "PRIVATE KEY", "CERTIFICATE REQUEST", `"compliance_csid":`, `"production_csid":`, `"otp":`, `"private_key_pem":`} {
				if strings.Contains(string(raw), secret) {
					t.Fatalf("request secrets in diagnostics: %s", secret)
				}
			}
		})
	}
}
