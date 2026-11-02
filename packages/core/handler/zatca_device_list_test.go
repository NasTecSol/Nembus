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
	"github.com/jackc/pgx/v5/pgtype"
)

type registeredDeviceDB struct {
	repository.DBTX
	args    []interface{}
	sql     string
	err     error
	devices []repository.ZatcaRegisteredDevice
}

func (db *registeredDeviceDB) Query(_ context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	db.sql, db.args = sql, args
	return &registeredDeviceRows{devices: db.devices}, db.err
}

type registeredDeviceRows struct {
	pgx.Rows
	devices []repository.ZatcaRegisteredDevice
	index   int
}

func (r *registeredDeviceRows) Close()     {}
func (r *registeredDeviceRows) Err() error { return nil }
func (r *registeredDeviceRows) Next() bool { return r.index < len(r.devices) }
func (r *registeredDeviceRows) Scan(dest ...interface{}) error {
	value := reflect.ValueOf(r.devices[r.index])
	for i, target := range dest {
		reflect.ValueOf(target).Elem().Set(value.Field(i))
	}
	r.index++
	return nil
}

func deviceListEngine(db *registeredDeviceDB, orgID int32) *gin.Engine {
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		if orgID > 0 {
			c.Set(string(middleware.OrgIDKey), orgID)
		}
		if db != nil {
			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), middleware.RepoKey, repository.New(db)))
		}
	})
	engine.GET("/devices", NewZatcaHandler(nil).ListZatcaDevices)
	return engine
}

func TestZatcaDeviceListScopeAndSafeResponse(t *testing.T) {
	for _, terminal := range []int32{0, 8} {
		db := &registeredDeviceDB{devices: []repository.ZatcaRegisteredDevice{{
			ID: 3, OrganizationID: 7, StoreID: 2, StoreName: "Main Store",
			PosTerminalID: pgtype.Int4{Int32: 8, Valid: true}, TerminalName: pgtype.Text{String: "Counter 1", Valid: true},
			DeviceSerial: "EGS-POS-08", DeviceType: "pos", ReportingReady: true,
		}}}
		query := "/devices?store_id=2"
		if terminal > 0 {
			query += "&pos_terminal_id=8"
		}
		recorder := httptest.NewRecorder()
		deviceListEngine(db, 7).ServeHTTP(recorder, httptest.NewRequest("GET", query, nil))
		if recorder.Code != 200 {
			t.Fatalf("code=%d body=%s", recorder.Code, recorder.Body)
		}
		if !reflect.DeepEqual(db.args, []interface{}{int32(7), int32(2), terminal}) {
			t.Fatalf("wrong scope: %v", db.args)
		}
		for _, required := range []string{"d.organization_id=$1 AND d.store_id=$2", "($3::int=0 OR d.pos_terminal_id=$3)", "t.store_id=d.store_id"} {
			if !strings.Contains(db.sql, required) {
				t.Fatalf("missing SQL scope %s", required)
			}
		}
		projection := strings.Split(db.sql, " FROM zatca_device_configs")[0]
		for _, forbidden := range []string{"private_key", "csr_pem", "d.*"} {
			if strings.Contains(projection, forbidden) {
				t.Fatalf("unsafe projection: %s", forbidden)
			}
		}
		var body struct {
			Data struct {
				Devices []repository.ZatcaRegisteredDevice `json:"devices"`
				Total   int                                `json:"total"`
			} `json:"data"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Data.Total != 1 || len(body.Data.Devices) != 1 || body.Data.Devices[0].TerminalName.String != "Counter 1" || !body.Data.Devices[0].ReportingReady {
			t.Fatalf("wrong response: %s", recorder.Body)
		}
		for _, forbidden := range []string{"private_key", "production_csid", "compliance_csid", "production_secret", "metadata", "csr_pem"} {
			if strings.Contains(recorder.Body.String(), forbidden) {
				t.Fatalf("credential exposed: %s", forbidden)
			}
		}
	}
}

func TestZatcaDeviceListEmptyAndFailures(t *testing.T) {
	for _, tc := range []struct {
		name, query string
		db          *registeredDeviceDB
		org         int32
		code        int
	}{
		{"empty", "?store_id=2", &registeredDeviceDB{}, 7, 200},
		{"missing store", "", nil, 7, 400},
		{"bad store", "?store_id=0", nil, 7, 400},
		{"duplicate store", "?store_id=1&store_id=2", nil, 7, 400},
		{"bad terminal", "?store_id=2&pos_terminal_id=abc", nil, 7, 400},
		{"empty terminal", "?store_id=2&pos_terminal_id=", nil, 7, 400},
		{"overflow", "?store_id=2147483648", nil, 7, 400},
		{"missing organization", "?store_id=2", nil, 0, 401},
		{"missing repository", "?store_id=2", nil, 7, 500},
		{"database failure", "?store_id=2", &registeredDeviceDB{err: errors.New("secret database failure")}, 7, 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			deviceListEngine(tc.db, tc.org).ServeHTTP(recorder, httptest.NewRequest("GET", "/devices"+tc.query, nil))
			if recorder.Code != tc.code {
				t.Fatalf("code=%d body=%s", recorder.Code, recorder.Body)
			}
			if tc.code == 200 && !strings.Contains(recorder.Body.String(), `"devices":[]`) {
				t.Fatal("empty list must be []")
			}
			if strings.Contains(recorder.Body.String(), "secret database failure") {
				t.Fatal("database error leaked")
			}
		})
	}
}
