package zatca

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type onboardingFlowDB struct {
	completionDB
	creates int
}

func (db *onboardingFlowDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if strings.Contains(sql, "CreateZatcaDeviceConfig") {
		db.creates++
		db.device = repository.ZatcaDeviceConfig{ID: 3, OrganizationID: args[0].(int32), StoreID: args[1].(pgtype.Int4), PosTerminalID: args[2].(pgtype.Int4), DeviceSerial: args[3].(string), DeviceType: args[4].(string), CsrPem: args[5].(pgtype.Text), PrivateKeyPem: args[6].(string), ComplianceCsid: args[7].(pgtype.Text), ProductionCsid: args[8].(pgtype.Text), CsidExpiry: args[9].(pgtype.Timestamptz), ZatcaEnv: args[10].(pgtype.Text), IsActive: args[11].(pgtype.Bool), Metadata: args[12].(json.RawMessage)}
		return submissionRow{db.device}
	}
	return db.completionDB.QueryRow(ctx, sql, args...)
}

func subjectFromOnboardingCSR(t *testing.T, encoded string) *ecdsa.PublicKey {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		t.Fatal("expected PEM CSR")
	}
	var csr struct {
		Info struct {
			Version int
			Subject asn1.RawValue
			SPKI    struct {
				Algorithm pkix.AlgorithmIdentifier
				Key       asn1.BitString
			}
			Attributes asn1.RawValue
		}
		Algorithm pkix.AlgorithmIdentifier
		Signature asn1.BitString
	}
	if _, err := asn1.Unmarshal(block.Bytes, &csr); err != nil {
		t.Fatal(err)
	}
	x, y := elliptic.Unmarshal(usecase.Secp256k1(), csr.Info.SPKI.Key.Bytes)
	if x == nil {
		t.Fatal("CSR public key invalid")
	}
	return &ecdsa.PublicKey{Curve: usecase.Secp256k1(), X: x, Y: y}
}

func TestSingleOnboardCallCompletesAllStepsAndPreservesRetryCredentials(t *testing.T) {
	for _, scenario := range []string{"success", "production-success", "sample-failure", "exchange-failure"} {
		t.Run(scenario, func(t *testing.T) {
			ca, _ := usecase.GenerateECDSAKeyPair()
			var compliance, production string
			credentials, checks, exchanges := 0, 0, 0
			requestedEnvironment, globalEnvironment := "simulation", "production"
			if scenario == "production-success" {
				requestedEnvironment, globalEnvironment = "production", "simulation"
			}
			fail := scenario != "success" && scenario != "production-success"
			db := &onboardingFlowDB{completionDB: completionDB{org: repository.Organization{ID: 1, Name: "Seller", TaxID: pgtype.Text{String: "310123456700003", Valid: true}, Metadata: []byte(`{"zatca":{"enabled":true,"seller":{"commercial_registration_number":"1010123456","street":"Street","building_number":"1234","district":"District","city":"Riyadh","postal_code":"12345"}}}`)}}}
			gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/compliance":
					credentials++
					var req ComplianceCSIDRequest
					if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
						t.Error(err)
						w.WriteHeader(400)
						return
					}
					pub := subjectFromOnboardingCSR(t, req.CSR)
					compliance = completionCertificateWithSubject(t, pub, ca, 1)
					production = completionCertificateWithSubject(t, pub, ca, 2)
					_ = json.NewEncoder(w).Encode(map[string]any{"requestID": 12345, "binarySecurityToken": compliance, "secret": "cc-secret"})
				case "/compliance/invoices":
					checks++
					if db.creates != 1 {
						t.Error("compliance credentials were not saved before checking")
					}
					user, secret, _ := r.BasicAuth()
					if user != compliance || secret != "cc-secret" {
						t.Error("wrong compliance credentials")
					}
					if fail && scenario == "sample-failure" {
						w.WriteHeader(400)
						_, _ = w.Write([]byte(`{"validationResults":{"status":"ERROR","errorMessages":[{"code":"test-failure"}]}}`))
						return
					}
					_, _ = w.Write([]byte(`{"validationResults":{"status":"PASS","errorMessages":[]}}`))
				case "/production/csids":
					exchanges++
					if checks < 6 {
						t.Error("exchange happened before all checks")
					}
					if fail && scenario == "exchange-failure" {
						w.WriteHeader(400)
						_, _ = w.Write([]byte(`{"errors":["Missing-ComplianceSteps"]}`))
						return
					}
					_ = json.NewEncoder(w).Encode(map[string]string{"binarySecurityToken": production, "secret": "pc-secret"})
				default:
					t.Error("unexpected gateway endpoint", r.URL.Path)
					w.WriteHeader(404)
				}
			}))
			defer gateway.Close()
			repo := repository.New(db)
			ctx := context.WithValue(context.Background(), middleware.RepoKey, repo)
			service := testOnboardingService(nil, NewZatcaClient(gateway.URL), &usecase.ZatcaConfig{Enabled: false, Env: globalEnvironment})
			factory := service.clients
			service.clients = func(environment string) (*ZatcaClient, error) {
				if environment != requestedEnvironment {
					t.Error("wrong environment", environment)
				}
				return factory(environment)
			}
			device, err := service.OnboardDevice(ctx, 1, nil, nil, "NEW-SINGLE-CALL-DEVICE", "pos", "test-otp", requestedEnvironment)
			if scenario != "success" && scenario != "production-success" {
				if err == nil || device == nil || db.device.ProductionCsid.Valid || db.writes != 0 {
					t.Fatal("incomplete onboarding was reported ready")
				}
				var progress interface{ OnboardingFailureDetails() json.RawMessage }
				if !errors.As(err, &progress) {
					t.Fatal("saved device ID/progress missing")
				}
				data := string(progress.OnboardingFailureDetails())
				if !strings.Contains(data, `"device_id":3`) || !strings.Contains(data, `"reporting_ready":false`) || strings.Contains(data, "cc-secret") || strings.Contains(data, "PRIVATE KEY") {
					t.Fatal("unsafe or incomplete error details", data)
				}
				// Resume the saved registration, without obtaining another OTP/key/CCSID.
				fail = false
				if device, err := service.OnboardDevice(ctx, 1, nil, nil, "NEW-SINGLE-CALL-DEVICE", "pos", "", requestedEnvironment); err != nil || device.ProductionCsid.String == "" {
					t.Fatal("resume failed", err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if credentials != 1 || checks != 6 || exchanges != 1 || db.creates != 1 || db.writes != 1 {
					t.Fatal("single call did not execute the entire onboarding sequence")
				}
				if device.ProductionCsid.String != production || !strings.Contains(string(device.Metadata), `"production_secret":"pc-secret"`) {
					t.Fatal("returned stale compliance-only config")
				}
			}
			if db.device.ZatcaEnv.String != requestedEnvironment {
				t.Error("request environment was not saved")
			}
			if credentials != 1 || db.creates != 1 {
				t.Fatal("retry recreated device credentials")
			}
		})
	}
}
