package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type documentStatusDB struct {
	repository.DBTX
	entry  repository.ZatcaDocumentChain
	err    error
	legacy bool
	calls  int
}

type documentStatusRow struct {
	entry  repository.ZatcaDocumentChain
	err    error
	legacy bool
}

func (r documentStatusRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	if r.legacy {
		*dest[0].(*int32) = 6
		return nil
	}
	value := reflect.ValueOf(r.entry)
	for i := range dest {
		reflect.ValueOf(dest[i]).Elem().Set(value.Field(i))
	}
	return nil
}

func (db *documentStatusDB) QueryRow(_ context.Context, sql string, args ...interface{}) pgx.Row {
	db.calls++
	if !strings.Contains(strings.ToUpper(sql), "SELECT") {
		panic("status lookup must be read-only")
	}
	if db.err != nil {
		return documentStatusRow{err: db.err}
	}
	if strings.Contains(sql, "GetLegacyZatcaOrderTransaction") {
		return documentStatusRow{legacy: true}
	}
	if db.legacy && args[0] == "sales_order_v2" {
		return documentStatusRow{err: pgx.ErrNoRows}
	}
	return documentStatusRow{entry: db.entry}
}

func TestSalesOrderDocumentStatusReadsSavedTenantResult(t *testing.T) {
	const orderID = "00000000-0000-4000-8000-000000000006"
	for _, tc := range []struct {
		name   string
		status repository.ZatcaDocStatus
		legacy bool
		err    error
		code   int
	}{
		{"pending", repository.ZatcaDocStatusPending, false, nil, 200},
		{"reported", repository.ZatcaDocStatusReported, false, nil, 200},
		{"failed", repository.ZatcaDocStatusFailed, false, nil, 200},
		{"rejected", repository.ZatcaDocStatusRejected, false, nil, 200},
		{"legacy", repository.ZatcaDocStatusReported, true, nil, 200},
		{"missing", "", false, pgx.ErrNoRows, 404},
		{"database failure", "", false, errors.New("database unavailable"), 500},
		{"wrong organization", repository.ZatcaDocStatusReported, false, nil, 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := &documentStatusDB{legacy: tc.legacy, err: tc.err, entry: repository.ZatcaDocumentChain{
				ID: 9, OrganizationID: 1, EntityType: "sales_order_v2", EntityID: orderID,
				ZatcaStatus: repository.NullZatcaDocStatus{ZatcaDocStatus: tc.status, Valid: true},
			}}
			if tc.status == repository.ZatcaDocStatusReported {
				db.entry.ZatcaResponse = []byte(`{"reportingStatus":"REPORTED"}`)
			}
			if tc.legacy {
				db.entry.EntityType, db.entry.EntityID = "pos_transaction", "6"
			}
			engine := gin.New()
			engine.Use(func(c *gin.Context) {
				ctx := context.WithValue(c.Request.Context(), middleware.RepoKey, repository.New(db))
				if tc.name == "wrong organization" {
					c.Set(string(middleware.OrgIDKey), int32(2))
				}
				c.Request = c.Request.WithContext(ctx)
			})
			engine.GET("/orders/:id/status", NewZatcaHandler(nil).GetSalesOrderDocumentStatus)
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, httptest.NewRequest("GET", "/orders/"+orderID+"/status", nil))
			if recorder.Code != tc.code {
				t.Fatalf("code=%d body=%s", recorder.Code, recorder.Body)
			}
			if tc.code == 200 {
				var body struct {
					Data repository.ZatcaDocumentChain `json:"data"`
				}
				if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				wantResponse := string(db.entry.ZatcaResponse)
				if wantResponse == "" {
					wantResponse = "null"
				}
				if body.Data.ZatcaStatus != db.entry.ZatcaStatus || string(body.Data.ZatcaResponse) != wantResponse || body.Data.EntityType != db.entry.EntityType {
					t.Fatalf("saved result changed: %s", recorder.Body)
				}
			}
		})
	}
}

func TestSalesOrderDocumentStatusRejectsInvalidIDAndMissingTenant(t *testing.T) {
	engine := gin.New()
	engine.GET("/orders/:id/status", NewZatcaHandler(nil).GetSalesOrderDocumentStatus)
	for _, tc := range []struct {
		id   string
		code int
	}{
		{"invalid", 400},
		{"00000000-0000-4000-8000-000000000006", 500},
	} {
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, httptest.NewRequest("GET", "/orders/"+tc.id+"/status", nil))
		if recorder.Code != tc.code {
			t.Fatalf("code=%d body=%s", recorder.Code, recorder.Body)
		}
	}
}
