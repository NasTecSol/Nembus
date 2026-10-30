package zatca

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type credentialSaveDB struct {
	repository.DBTX
	saved repository.ZatcaDeviceConfig
}

func (db *credentialSaveDB) QueryRow(_ context.Context, sql string, args ...interface{}) pgx.Row {
	if strings.Contains(sql, "GetOrganization") {
		return submissionRow{repository.Organization{ID: 1, Name: "Nembus", TaxID: pgtype.Text{String: "310123456700003", Valid: true}}}
	}
	if strings.Contains(sql, "CreateZatcaDeviceConfig") {
		db.saved = repository.ZatcaDeviceConfig{
			ID: 1, OrganizationID: args[0].(int32), ComplianceCsid: args[7].(pgtype.Text),
			ProductionCsid: args[8].(pgtype.Text), Metadata: args[12].(json.RawMessage),
		}
		return submissionRow{db.saved}
	}
	panic("unexpected query")
}

func TestOnboardingSavesComplianceBeforeProductionExchange(t *testing.T) {
	for _, env := range []string{"simulation", "production"} {
		t.Run(env, func(t *testing.T) {
			key, _ := usecase.GenerateECDSAKeyPair()
			token := completionCertificate(t, key, 1)
			calls := 0
			gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != "/compliance" {
					t.Error("reporting exchange attempted before compliance suite", r.URL.Path)
					w.WriteHeader(400)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"requestID": 123456789, "binarySecurityToken": token, "secret": "compliance-secret"})
			}))
			defer gateway.Close()
			db := &credentialSaveDB{}
			service := testOnboardingService(repository.New(db), NewZatcaClient(gateway.URL), &usecase.ZatcaConfig{Env: env})
			device, err := service.createComplianceDevice(context.Background(), 1, nil, nil, "32164879", "pos", "test-otp", env)
			if err != nil {
				t.Fatal(err)
			}
			var metadata map[string]string
			if err := json.Unmarshal(device.Metadata, &metadata); err != nil {
				t.Fatal(err)
			}
			if calls != 1 || device.ComplianceCsid.String != token || metadata["compliance_request_id"] != "123456789" || metadata["compliance_secret"] != "compliance-secret" {
				t.Fatal("compliance credentials/request ID lost")
			}
			if _, ok := metadata["otp"]; ok {
				t.Fatal("OTP must not be persisted")
			}
			if device.ProductionCsid.Valid || metadata["production_secret"] != "" || metadata["onboarding_stage"] != "compliance" || metadata["invoice_type"] != "1100" {
				t.Fatal("onboarding must save compliance stage before reporting")
			}
		})
	}
}
