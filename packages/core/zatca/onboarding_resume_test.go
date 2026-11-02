package zatca

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestExistingOnboardingChecksScopeBeforeUsingOTP(t *testing.T) {
	key, _ := usecase.GenerateECDSAKeyPair()
	pem, _ := usecase.EncodePrivateKeyPEM(key)
	certificate := completionCertificate(t, key, 2)
	for _, tc := range []struct {
		name, environment, kind      string
		store, terminal              int32
		inactive, revoked, newDevice bool
		status                       int
	}{
		{"ready", "simulation", "pos", 3, 8, false, false, false, 200},
		{"wrong store", "simulation", "pos", 4, 8, false, false, false, 409},
		{"wrong terminal", "simulation", "pos", 3, 9, false, false, false, 409},
		{"wrong environment", "production", "pos", 3, 8, false, false, false, 409},
		{"wrong type", "simulation", "cloud", 3, 8, false, false, false, 409},
		{"inactive", "simulation", "pos", 3, 8, true, false, false, 409},
		{"revoked", "simulation", "pos", 3, 8, false, true, false, 409},
		{"new without OTP", "simulation", "pos", 3, 8, false, false, true, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := &onboardingFlowDB{completionDB: completionDB{
				org: repository.Organization{ID: 1, Metadata: []byte(`{"zatca":{"enabled":true}}`)},
				device: repository.ZatcaDeviceConfig{ID: 3, OrganizationID: 1, DeviceSerial: "POS-08", DeviceType: "pos",
					StoreID: pgtype.Int4{Int32: 3, Valid: true}, PosTerminalID: pgtype.Int4{Int32: 8, Valid: true},
					ZatcaEnv: pgtype.Text{String: "simulation", Valid: true}, IsActive: pgtype.Bool{Bool: !tc.inactive, Valid: true}, IsRevoked: pgtype.Bool{Bool: tc.revoked, Valid: true},
					PrivateKeyPem: pem, ComplianceCsid: pgtype.Text{String: "saved-compliance", Valid: true}, ProductionCsid: pgtype.Text{String: certificate, Valid: true},
					Metadata: []byte(`{"production_secret":"saved-secret","credential_environment":"simulation"}`)},
			}}
			if tc.newDevice {
				db.device = repository.ZatcaDeviceConfig{}
			}
			service := NewOnboardingService(nil, &usecase.ZatcaConfig{})
			gatewayCalls := 0
			service.clients = func(string) (*ZatcaClient, error) { gatewayCalls++; return nil, fmt.Errorf("must not call gateway") }
			// A ready-device validation constructs a client, but sends no requests.
			if tc.status == 200 {
				service.clients = newEnvironmentClient
			}
			ctx := context.WithValue(context.Background(), middleware.RepoKey, repository.New(db))
			device, err := service.OnboardDevice(ctx, 1, &tc.store, &tc.terminal, "POS-08", tc.kind, "", tc.environment)
			if tc.status == 200 {
				if err != nil || device.ID != 3 || device.PrivateKeyPem != pem || device.ProductionCsid.String != certificate {
					t.Fatalf("ready registration was not reused: %v", err)
				}
			} else {
				var problem interface{ OnboardingHTTPStatus() int }
				if !errors.As(err, &problem) || problem.OnboardingHTTPStatus() != tc.status {
					t.Fatalf("expected HTTP %d, got %v", tc.status, err)
				}
			}
			if db.creates != 0 || db.writes != 0 || gatewayCalls != 0 {
				t.Fatal("retry replaced credentials or contacted the gateway before scope validation")
			}
		})
	}
}

type concurrentOnboardingDB struct {
	completionDB
	lookups, creates int
}

type duplicateRegistrationRow struct{}

func (duplicateRegistrationRow) Scan(...interface{}) error {
	return &pgconn.PgError{Code: "23505", ConstraintName: "zatca_device_configs_organization_id_device_serial_key"}
}

func (db *concurrentOnboardingDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if strings.Contains(sql, "GetZatcaDeviceConfigBySerial") {
		db.lookups++
		if db.lookups == 1 {
			return missingOrganizationRow{}
		}
		return submissionRow{db.device}
	}
	if strings.Contains(sql, "CreateZatcaDeviceConfig") {
		db.creates++
		return duplicateRegistrationRow{}
	}
	return db.completionDB.QueryRow(ctx, sql, args...)
}

func TestConcurrentRegistrationReusesWinningDeviceWithoutReplacingKeys(t *testing.T) {
	key, _ := usecase.GenerateECDSAKeyPair()
	pem, _ := usecase.EncodePrivateKeyPEM(key)
	compliance := completionCertificate(t, key, 1)
	production := completionCertificate(t, key, 2)
	db := &concurrentOnboardingDB{completionDB: completionDB{
		org: repository.Organization{ID: 1, Name: "Seller", TaxID: pgtype.Text{String: "310123456700003", Valid: true}, Metadata: []byte(`{"zatca":{"enabled":true}}`)},
		device: repository.ZatcaDeviceConfig{ID: 3, OrganizationID: 1, DeviceSerial: "POS-08", DeviceType: "pos",
			ZatcaEnv: pgtype.Text{String: "simulation", Valid: true}, IsActive: pgtype.Bool{Bool: true, Valid: true}, PrivateKeyPem: pem,
			ComplianceCsid: pgtype.Text{String: compliance, Valid: true}, ProductionCsid: pgtype.Text{String: production, Valid: true},
			Metadata: []byte(`{"production_secret":"saved-secret","credential_environment":"simulation"}`)},
	}}
	requests := 0
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/compliance" {
			t.Error("unexpected gateway call", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"requestID": 12345, "binarySecurityToken": compliance, "secret": "new-unused-secret"})
	}))
	defer gateway.Close()
	service := testOnboardingService(nil, NewZatcaClient(gateway.URL), &usecase.ZatcaConfig{})
	ctx := context.WithValue(context.Background(), middleware.RepoKey, repository.New(db))
	device, err := service.OnboardDevice(ctx, 1, nil, nil, "POS-08", "pos", "654321", "simulation")
	if err != nil || device == nil || device.PrivateKeyPem != pem || device.ComplianceCsid.String != compliance || device.ProductionCsid.String != production {
		t.Fatalf("concurrent registration did not preserve the existing device: %v", err)
	}
	if requests != 1 || db.lookups != 2 || db.creates != 1 || db.writes != 0 {
		t.Fatal("unexpected writes or extra onboarding calls")
	}
}
