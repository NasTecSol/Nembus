package zatca

import (
	"context"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/jackc/pgx/v5"
)

type onboardingLookupDB struct {
	repository.DBTX
	lookups int
}

func TestOnboardingCSRSerialInCompliancePayload(t *testing.T) {
	for _, test := range []struct{ serial, deviceType, want string }{
		{"32164879", "pos", "1-Nembus|2-POS|3-32164879"},
		{"32164879", "cloud", "1-Nembus|2-Cloud|3-32164879"},
		{"1-Provider|2-Model|3-32164879", "pos", "1-Provider|2-Model|3-32164879"},
	} {
		t.Run(test.want, func(t *testing.T) {
			serial, err := csrDeviceSerial(test.serial, test.deviceType)
			if err != nil {
				t.Fatal(err)
			}
			key, err := usecase.GenerateECDSAKeyPair()
			if err != nil {
				t.Fatal(err)
			}
			csr, err := usecase.GenerateCSR(key, usecase.ZatcaCSRParams{
				CommonName: "POS-01", SerialNumber: serial, OrgName: "Nembus",
				OrgUnit: "Branch", CountryCode: "SA", OrganizationIdentifier: "310123456700003",
				InvoiceType: "1100", Location: "Riyadh", Industry: "Retail", IsSandbox: true,
			})
			if err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var payload ComplianceCSIDRequest
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Error(err)
					w.WriteHeader(400)
					return
				}
				decoded, err := base64.StdEncoding.DecodeString(payload.CSR)
				if err != nil {
					t.Error(err)
					w.WriteHeader(400)
					return
				}
				block, _ := pem.Decode(decoded)
				if block == nil {
					t.Error("request CSR is not Base64 PEM")
					w.WriteHeader(400)
					return
				}
				var request struct {
					Info struct {
						Version    int
						Subject    asn1.RawValue
						PublicKey  asn1.RawValue
						Attributes []struct {
							Type   asn1.ObjectIdentifier
							Values [][]pkix.Extension `asn1:"set"`
						} `asn1:"tag:0"`
					}
					Algorithm pkix.AlgorithmIdentifier
					Signature asn1.BitString
				}
				if _, err := asn1.Unmarshal(block.Bytes, &request); err != nil {
					t.Error(err)
					w.WriteHeader(400)
					return
				}
				found := false
				for _, attribute := range request.Info.Attributes {
					for _, extensions := range attribute.Values {
						for _, extension := range extensions {
							if !extension.Id.Equal(asn1.ObjectIdentifier{2, 5, 29, 17}) {
								continue
							}
							var names []asn1.RawValue
							if _, err := asn1.Unmarshal(extension.Value, &names); err != nil {
								t.Error(err)
								continue
							}
							for _, name := range names {
								var rdns pkix.RDNSequence
								if _, err := asn1.Unmarshal(name.Bytes, &rdns); err != nil {
									t.Error(err)
									continue
								}
								for _, set := range rdns {
									for _, field := range set {
										if field.Type.Equal(asn1.ObjectIdentifier{2, 5, 4, 4}) {
											found = true
											if field.Value != test.want {
												t.Errorf("wire CSR serial = %v; want %s", field.Value, test.want)
											}
										}
									}
								}
							}
						}
					}
				}
				if !found {
					t.Error("wire CSR has no SAN SN field")
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"binarySecurityToken":"test-token","secret":"test-secret"}`))
			}))
			defer server.Close()
			if _, err := NewZatcaClient(server.URL).RequestComplianceCSID(context.Background(), csr, "test-otp"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestOnboardingRejectsMalformedCSRSerial(t *testing.T) {
	for _, serial := range []string{"", " ", "1-Nembus", "1-Nembus|2-POS", "1-Nembus|2-|3-123", "1-Nembus|2-POS|3-", "1-Nembus|3-POS|2-123"} {
		if _, err := csrDeviceSerial(serial, "pos"); err == nil {
			t.Errorf("expected rejection of malformed serial %q", serial)
		}
	}
}

func (db *onboardingLookupDB) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	db.lookups++
	return missingOrganizationRow{}
}

type missingOrganizationRow struct{}

func (missingOrganizationRow) Scan(...interface{}) error { return pgx.ErrNoRows }

func TestOnboardingUsesRequestTenantRepository(t *testing.T) {
	masterDB := &onboardingLookupDB{}
	tenantADB := &onboardingLookupDB{}
	tenantBDB := &onboardingLookupDB{}
	masterRepo := repository.New(masterDB)
	service := testOnboardingService(masterRepo, nil, nil)

	for _, tenantDB := range []*onboardingLookupDB{tenantADB, tenantBDB} {
		ctx := context.WithValue(context.Background(), middleware.RepoKey, repository.New(tenantDB))
		_, err := service.createComplianceDevice(ctx, 1, nil, nil, "test-device", "pos", "test-otp", "sandbox")
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("expected organization lookup error, got %v", err)
		}
		if tenantDB.lookups != 1 {
			t.Fatalf("tenant repository received %d lookups; want 1", tenantDB.lookups)
		}
	}
	if masterDB.lookups != 0 {
		t.Fatalf("master database received %d tenant lookups", masterDB.lookups)
	}
	if service.repo != masterRepo {
		t.Fatal("request changed the service repository")
	}
}
