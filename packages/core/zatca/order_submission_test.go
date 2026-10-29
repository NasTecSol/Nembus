package zatca

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type orderSubmissionDB struct {
	submissionDB
	order repository.SalesOrdersV2
}
type orderSubmissionTx struct {
	pgx.Tx
	db *orderSubmissionDB
}

func (db *orderSubmissionDB) Begin(context.Context) (pgx.Tx, error) {
	return &orderSubmissionTx{db: db}, nil
}
func (db *orderSubmissionDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if strings.Contains(sql, "GetSalesOrderV2") {
		return submissionRow{db.order}
	}
	return db.submissionDB.QueryRow(ctx, sql, args...)
}
func (tx *orderSubmissionTx) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	if strings.Contains(sql, "LockZatcaOrder") {
		return pgconn.NewCommandTag("SELECT 1"), nil
	}
	return tx.db.Exec(ctx, sql, args...)
}
func (tx *orderSubmissionTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return tx.db.QueryRow(ctx, sql, args...)
}
func (tx *orderSubmissionTx) Commit(context.Context) error   { return nil }
func (tx *orderSubmissionTx) Rollback(context.Context) error { return nil }

func TestOrderCheckAndSubmitUseSavedDocumentAndMatchingCredentialPairs(t *testing.T) {
	id := uuid.New()
	docID := uuid.New()
	calls := 0
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		user, secret, _ := r.BasicAuth()
		if r.URL.Path == "/compliance/invoices" {
			if user != "ccsid" || secret != "cc-secret" {
				t.Error("wrong compliance credentials")
			}
		} else if r.URL.Path == "/invoices/reporting/single" {
			if user != "pcsid" || secret != "pc-secret" {
				t.Error("wrong reporting credentials")
			}
		} else {
			t.Error("unexpected endpoint", r.URL.Path)
		}
		var payload InvoiceSubmissionRequest
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if payload.InvoiceHash != "saved-order-hash" || payload.UUID != docID.String() {
			t.Error("saved order identity changed")
		}
		_, _ = w.Write([]byte(`{"reportingStatus":"REPORTED","validationResults":{"status":"PASS","errorMessages":[]}}`))
	}))
	defer gateway.Close()
	db := &orderSubmissionDB{order: repository.SalesOrdersV2{ID: id}, submissionDB: submissionDB{
		chain:  repository.ZatcaDocumentChain{ID: 1, EntityType: usecase.ZatcaSalesOrderEntity, EntityID: id.String(), OrganizationID: 1, DeviceConfigID: 3, ZatcaUuid: docID, XmlHash: "saved-order-hash", SignedXml: pgtype.Text{String: "saved-order-xml", Valid: true}},
		device: repository.ZatcaDeviceConfig{ID: 3, OrganizationID: 1, IsActive: pgtype.Bool{Bool: true, Valid: true}, ZatcaEnv: pgtype.Text{String: "simulation", Valid: true}, ComplianceCsid: pgtype.Text{String: "ccsid", Valid: true}, ProductionCsid: pgtype.Text{String: "pcsid", Valid: true}, Metadata: []byte(`{"compliance_secret":"cc-secret","production_secret":"pc-secret","onboarding_stage":"production"}`)},
	}}
	repo := repository.New(db)
	ctx := context.WithValue(context.Background(), middleware.RepoKey, repo)
	service := testService(repo, &usecase.ZatcaConfig{Enabled: true, Env: "simulation", BaseURL: gateway.URL})
	if _, err := service.CheckSalesOrder(ctx, id); err != nil {
		t.Fatal(err)
	}
	if db.updates != 0 {
		t.Fatal("check changed reporting status")
	}
	result, err := service.SubmitSalesOrder(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if result.ZatcaStatus.ZatcaDocStatus != repository.ZatcaDocStatusReported || db.status.ZatcaDocStatus != repository.ZatcaDocStatusReported || calls != 2 {
		t.Fatal("order was not reported/persisted")
	}
	db.chain = *result
	if _, err := service.SubmitSalesOrder(ctx, id); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatal("reported order was submitted again")
	}
}
