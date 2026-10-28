package zatca

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type submissionDB struct {
	organizationMetadata json.RawMessage
	repository.DBTX
	chain    repository.ZatcaDocumentChain
	device   repository.ZatcaDeviceConfig
	updates  int
	status   repository.NullZatcaDocStatus
	response json.RawMessage
}

type submissionTx struct {
	pgx.Tx
	db *submissionDB
}

func (db *submissionDB) Begin(context.Context) (pgx.Tx, error) {
	return &submissionTx{db: db}, nil
}
func (tx *submissionTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return tx.db.QueryRow(ctx, sql, args...)
}
func (tx *submissionTx) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return tx.db.Exec(ctx, sql, args...)
}
func (tx *submissionTx) Commit(context.Context) error   { return nil }
func (tx *submissionTx) Rollback(context.Context) error { return nil }

type submissionRow struct{ value any }

func (r submissionRow) Scan(dest ...interface{}) error {
	v := reflect.ValueOf(r.value)
	if v.NumField() != len(dest) {
		return fmt.Errorf("unexpected row shape")
	}
	for i := range dest {
		reflect.ValueOf(dest[i]).Elem().Set(v.Field(i))
	}
	return nil
}

func (db *submissionDB) QueryRow(_ context.Context, sql string, args ...interface{}) pgx.Row {
	if strings.Contains(sql, "GetOrganization") {
		metadata := db.organizationMetadata
		if metadata == nil {
			metadata = []byte(`{"zatca":{"enabled":true}}`)
		}
		return submissionRow{repository.Organization{ID: args[0].(int32), Metadata: metadata}}
	}
	if strings.Contains(sql, "GetChainEntryByEntity") {
		return submissionRow{db.chain}
	}
	if strings.Contains(sql, "GetZatcaDeviceConfig") {
		return submissionRow{db.device}
	}
	panic("unexpected query")
}

func (db *submissionDB) Exec(_ context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	if strings.Contains(sql, "LockZatcaChainEntry") {
		return pgconn.NewCommandTag("SELECT 1"), nil
	}
	if !strings.Contains(sql, "UpdateChainEntryStatus") {
		panic("unexpected update")
	}
	db.updates++
	db.status = args[1].(repository.NullZatcaDocStatus)
	db.response = args[2].(json.RawMessage)
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func TestZatcaSubmissionPersistsGatewayResult(t *testing.T) {
	for _, test := range []struct {
		name       string
		httpStatus int
		body       string
		want       repository.ZatcaDocStatus
		wantErr    bool
	}{
		{"reported", 202, `{"reportingStatus":"REPORTED","validationResults":{"status":"WARNING","warningMessages":[]}}`, repository.ZatcaDocStatusReported, false},
		{"rejected", 400, `{"reportingStatus":"NOT_REPORTED","validationResults":{"status":"ERROR","errorMessages":[{"code":"invalid-signature"}]}}`, repository.ZatcaDocStatusRejected, true},
		{"gateway failure", 503, `upstream unavailable`, repository.ZatcaDocStatusFailed, true},
		{"empty unauthorized response", 401, ``, repository.ZatcaDocStatusFailed, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			id := uuid.New()
			calls := 0
			gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != "/invoices/reporting/single" || r.Header.Get("Accept-Version") != "V2" || r.Header.Get("Clearance-Status") != "0" {
					t.Error("incorrect reporting request")
				}
				user, password, ok := r.BasicAuth()
				if !ok || user != "saved-certificate" || password != "saved-device-secret" {
					t.Error("request did not use saved credentials")
				}
				var payload InvoiceSubmissionRequest
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Error(err)
				}
				if payload.UUID != id.String() || payload.InvoiceHash != "saved-hash" {
					t.Error("request did not reuse checkout document")
				}
				w.WriteHeader(test.httpStatus)
				_, _ = w.Write([]byte(test.body))
			}))
			defer gateway.Close()
			db := &submissionDB{
				chain: repository.ZatcaDocumentChain{
					ID: 12, EntityType: "pos_transaction", EntityID: "77", OrganizationID: 1, DeviceConfigID: 5,
					ZatcaUuid: id, XmlHash: "saved-hash", SignedXml: pgtype.Text{String: "<Invoice/>", Valid: true},
				},
				device: repository.ZatcaDeviceConfig{
					ID: 5, OrganizationID: 1, IsActive: pgtype.Bool{Bool: true, Valid: true},
					ProductionCsid: pgtype.Text{String: "saved-certificate", Valid: true},
					ZatcaEnv:       pgtype.Text{String: "sandbox", Valid: true},
					CsidExpiry:     pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
					Metadata:       json.RawMessage(`{"secret":"old-compliance-secret","production_secret":"saved-device-secret"}`),
				},
			}
			ctx := context.WithValue(context.Background(), middleware.RepoKey, repository.New(db))
			service := testService(nil, &usecase.ZatcaConfig{Enabled: true, Env: "sandbox", BaseURL: gateway.URL})
			entry, err := service.SubmitPosTransaction(ctx, 77)
			if (err != nil) != test.wantErr {
				t.Fatalf("error = %v; wantErr %v", err, test.wantErr)
			}
			if entry == nil || entry.ZatcaStatus.ZatcaDocStatus != test.want || db.updates != 1 || db.status.ZatcaDocStatus != test.want {
				t.Fatalf("result was not persisted correctly: %+v", entry)
			}
			if !json.Valid(db.response) {
				t.Fatal("stored response is not valid JSON")
			}
			if json.Valid([]byte(test.body)) && string(db.response) != test.body {
				t.Fatal("gateway validation details were lost")
			}
			if test.httpStatus == 401 && (err == nil || !strings.Contains(err.Error(), "authentication failed") || strings.Contains(err.Error(), "decode")) {
				t.Fatalf("expected actionable authentication error, got %v", err)
			}
			if calls != 1 {
				t.Fatalf("gateway calls = %d; want 1", calls)
			}
			if test.want == repository.ZatcaDocStatusReported {
				db.chain = *entry
				if _, err := service.SubmitPosTransaction(ctx, 77); err != nil {
					t.Fatal(err)
				}
				if calls != 1 || db.updates != 1 {
					t.Fatal("already reported document was submitted twice")
				}
			}
		})
	}
}

func TestZatcaSubmissionRequiresTenant(t *testing.T) {
	service := testService(nil, &usecase.ZatcaConfig{Enabled: true})
	if _, err := service.SubmitPosTransaction(context.Background(), 77); err == nil || !strings.Contains(err.Error(), "tenant repository") {
		t.Fatalf("expected tenant repository error, got %v", err)
	}
}

func TestZatcaComplianceOnlyDeviceChecksWithoutReporting(t *testing.T) {
	for _, legacy := range []bool{true, false} {
		t.Run(fmt.Sprintf("legacy=%v", legacy), func(t *testing.T) {
			calls := 0
			gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != "/compliance/invoices" {
					t.Errorf("wrong endpoint: %s", r.URL.Path)
				}
				user, secret, ok := r.BasicAuth()
				if !ok || user != "compliance-token" || secret != "compliance-secret" {
					t.Error("wrong compliance credentials")
				}
				_, _ = w.Write([]byte(`{"validationResults":{"status":"ERROR","errorMessages":[{"code":"invalid-signature"}]}}`))
			}))
			defer gateway.Close()
			device := repository.ZatcaDeviceConfig{
				OrganizationID: 1, IsActive: pgtype.Bool{Bool: true, Valid: true},
				ZatcaEnv:       pgtype.Text{String: "sandbox", Valid: true},
				ComplianceCsid: pgtype.Text{String: "compliance-token", Valid: true},
				Metadata:       json.RawMessage(`{"compliance_secret":"compliance-secret","onboarding_stage":"compliance"}`),
			}
			if legacy {
				device.ProductionCsid = device.ComplianceCsid
				device.Metadata = json.RawMessage(`{"secret":"compliance-secret"}`)
			}
			db := &submissionDB{device: device, chain: repository.ZatcaDocumentChain{
				OrganizationID: 1, ZatcaUuid: uuid.New(), XmlHash: "hash", SignedXml: pgtype.Text{String: "<Invoice/>", Valid: true},
			}}
			ctx := context.WithValue(context.Background(), middleware.RepoKey, repository.New(db))
			service := testService(nil, &usecase.ZatcaConfig{Enabled: true, Env: "sandbox", BaseURL: gateway.URL})
			if _, err := service.SubmitPosTransaction(ctx, 2); err == nil || !strings.Contains(err.Error(), "only a compliance CSID") {
				t.Fatalf("reporting should stop before network call: %v", err)
			}
			response, err := service.CheckPosTransaction(ctx, 2)
			if err != nil || !strings.Contains(string(response), "invalid-signature") {
				t.Fatalf("compliance validation details missing: %s, %v", response, err)
			}
			if calls != 1 || db.updates != 0 {
				t.Fatal("compliance check must not submit reporting or change the reporting status")
			}
		})
	}
}
