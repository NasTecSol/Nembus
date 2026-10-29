package zatca

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestEnvironmentClientUsesAllowlistedGatewayIgnoringGlobalOverride(t *testing.T) {
	service := NewService(nil, &usecase.ZatcaConfig{Enabled: true, Env: "production", BaseURL: "https://wrong-global-gateway.invalid"})
	for env, suffix := range map[string]string{"sandbox": "developer-portal", "simulation": "simulation", "production": "core"} {
		client, err := service.clients(env)
		if err != nil || client.baseURL != "https://gw-fatoora.zatca.gov.sa/e-invoicing/"+suffix {
			t.Fatalf("%s routed incorrectly: %v", env, err)
		}
	}
	if _, err := service.clients("unknown"); err == nil {
		t.Fatal("unknown environment fell back to a gateway")
	}
}

func TestSubmissionUsesDocumentDeviceEnvironmentInsteadOfGlobalSetting(t *testing.T) {
	for _, env := range []string{"simulation", "production"} {
		t.Run(env, func(t *testing.T) {
			gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				user, secret, _ := r.BasicAuth()
				if user != env+"-token" || secret != env+"-secret" {
					t.Error("mixed credential pair")
				}
				_, _ = w.Write([]byte(`{"reportingStatus":"REPORTED","validationResults":{"status":"PASS"}}`))
			}))
			defer gateway.Close()
			db := &submissionDB{device: repository.ZatcaDeviceConfig{ID: 3, OrganizationID: 1, IsActive: pgtype.Bool{Bool: true, Valid: true}, ZatcaEnv: pgtype.Text{String: env, Valid: true}, ProductionCsid: pgtype.Text{String: env + "-token", Valid: true}, Metadata: []byte(`{"production_secret":"` + env + `-secret","credential_environment":"` + env + `"}`)}}
			service := NewService(nil, &usecase.ZatcaConfig{Enabled: false, Env: "sandbox"})
			routed := ""
			service.clients = func(environment string) (*ZatcaClient, error) {
				routed = environment
				return NewZatcaClient(gateway.URL), nil
			}
			entry := &repository.ZatcaDocumentChain{ID: 1, OrganizationID: 1, DeviceConfigID: 3, XmlHash: "hash", SignedXml: pgtype.Text{String: "saved XML", Valid: true}}
			db.chain = *entry
			if _, err := service.submitDocument(context.Background(), repository.New(db), entry); err != nil {
				t.Fatal(err)
			}
			if routed != env || db.status.ZatcaDocStatus != repository.ZatcaDocStatusReported {
				t.Fatal("ignored saved device environment")
			}
		})
	}
}

type renewalEnvironmentDB struct {
	completionDB
	expectedToken string
	updates       int
}

func (db *renewalEnvironmentDB) Exec(_ context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	if !strings.Contains(sql, "SaveRenewedZatcaCredentials") {
		panic("unexpected renewal update")
	}
	db.updates++
	db.expectedToken = args[1].(string)
	db.device.ProductionCsid = pgtype.Text{String: args[2].(string), Valid: true}
	db.device.CsidExpiry = pgtype.Timestamptz{Time: args[3].(time.Time), Valid: true}
	db.device.Metadata = args[4].(json.RawMessage)
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func TestRenewalRoutesByDeviceAndSavesRotatedSecretAtomically(t *testing.T) {
	key, _ := usecase.GenerateECDSAKeyPair()
	pem, _ := usecase.EncodePrivateKeyPEM(key)
	oldToken := completionCertificate(t, key, 1)
	newToken := completionCertificate(t, key, 2)
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, secret, _ := r.BasicAuth()
		if r.URL.Path != "/production/csids/renewal" || user != oldToken || secret != "old-secret" {
			t.Error("wrong renewal endpoint or credential pair")
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"binarySecurityToken": newToken, "secret": "new-secret"})
	}))
	defer gateway.Close()
	db := &renewalEnvironmentDB{completionDB: completionDB{org: repository.Organization{ID: 1, Metadata: []byte(`{"zatca":{"enabled":true}}`)}, device: repository.ZatcaDeviceConfig{ID: 3, OrganizationID: 1, PrivateKeyPem: pem, IsActive: pgtype.Bool{Bool: true, Valid: true}, ZatcaEnv: pgtype.Text{String: "simulation", Valid: true}, ProductionCsid: pgtype.Text{String: oldToken, Valid: true}, Metadata: []byte(`{"production_secret":"old-secret","credential_environment":"simulation"}`)}}}
	service := NewOnboardingService(nil, &usecase.ZatcaConfig{Enabled: true, Env: "production"})
	routed := ""
	service.clients = func(env string) (*ZatcaClient, error) { routed = env; return NewZatcaClient(gateway.URL), nil }
	ctx := context.WithValue(context.Background(), middleware.RepoKey, repository.New(db))
	updated, err := service.RenewDeviceCSID(ctx, 3)
	if err != nil {
		t.Fatal(err)
	}
	if routed != "simulation" || updated.ProductionCsid.String != newToken || db.expectedToken != oldToken || db.updates != 1 || !strings.Contains(string(updated.Metadata), `"production_secret":"new-secret"`) {
		t.Fatal("renewal lost environment or matching secret")
	}
}
