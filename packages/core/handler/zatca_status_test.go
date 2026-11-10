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
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type statusDeviceDB struct {
	repository.DBTX
	device     repository.ZatcaDeviceConfig
	enabled    *bool
	err        error
	serialArgs []interface{}
}
type statusDeviceRow struct {
	value interface{}
	err   error
}

func (db *statusDeviceDB) QueryRow(_ context.Context, sql string, args ...interface{}) pgx.Row {
	if strings.Contains(sql, "GetOrganization") {
		if db.enabled != nil && !*db.enabled {
			return statusDeviceRow{value: repository.Organization{ID: 1, Metadata: []byte(`{"zatca":{"enabled":false}}`)}}
		}
		return statusDeviceRow{value: repository.Organization{ID: 1, Metadata: []byte(`{"zatca":{"enabled":true}}`)}}
	}
	if strings.Contains(sql, "GetZatcaDeviceConfigBySerial") {
		db.serialArgs = args
	}
	return statusDeviceRow{value: db.device, err: db.err}
}
func (row statusDeviceRow) Scan(dest ...interface{}) error {
	if row.err != nil {
		return row.err
	}
	value := reflect.ValueOf(row.value)
	for i, target := range dest {
		reflect.ValueOf(target).Elem().Set(value.Field(i))
	}
	return nil
}

func TestStatusLooksUpSerialWithinOrganization(t *testing.T) {
	for _, scenario := range []struct {
		name, query    string
		enabled, ready bool
		err            error
		code           int
	}{
		{"ready", "?organization_id=1&device_serial=%20POS-01%20", true, true, nil, 200},
		{"incomplete", "?organization_id=1&device_serial=POS-01", true, false, nil, 200},
		{"disabled", "?organization_id=1&device_serial=POS-01", false, true, nil, 200},
		{"missing", "?organization_id=1&device_serial=POS-01", true, false, pgx.ErrNoRows, 404},
		{"database failure", "?organization_id=1&device_serial=POS-01", true, false, errors.New("database unavailable"), 500},
		{"organization required", "?device_serial=POS-01", true, false, nil, 400},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			device := repository.ZatcaDeviceConfig{ID: 3, OrganizationID: 1, DeviceSerial: "POS-01", ZatcaEnv: pgtype.Text{String: "simulation", Valid: true}, IsActive: pgtype.Bool{Bool: true, Valid: true}, ComplianceCsid: pgtype.Text{String: "compliance", Valid: true}, Metadata: []byte(`{"production_secret":"do-not-leak"}`)}
			if scenario.ready {
				device.ProductionCsid = pgtype.Text{String: "production", Valid: true}
			}
			db := &statusDeviceDB{device: device, enabled: &scenario.enabled, err: scenario.err}
			h := NewZatcaHandler(usecase.NewZatcaUseCase(&usecase.ZatcaConfig{}))
			engine := gin.New()
			engine.Use(func(c *gin.Context) {
				c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), middleware.RepoKey, repository.New(db)))
			})
			engine.GET("/status", h.GetZatcaStatus)
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, httptest.NewRequest("GET", "/status"+scenario.query, nil))
			if recorder.Code != scenario.code {
				t.Fatalf("code=%d body=%s", recorder.Code, recorder.Body)
			}
			if scenario.code != 400 && !reflect.DeepEqual(db.serialArgs, []interface{}{int32(1), "POS-01"}) {
				t.Fatalf("unscoped lookup: %v", db.serialArgs)
			}
			if strings.Contains(recorder.Body.String(), "do-not-leak") {
				t.Fatal("credential leaked")
			}
			if scenario.code == 200 {
				var response struct {
					Data struct {
						Enabled        bool `json:"enabled"`
						ReportingReady bool `json:"reporting_ready"`
						DeviceID       int  `json:"device_id"`
					} `json:"data"`
				}
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Fatal(err)
				}
				if response.Data.Enabled != scenario.enabled || response.Data.ReportingReady != (scenario.enabled && scenario.ready) || response.Data.DeviceID != 3 {
					t.Fatalf("unexpected status: %s", recorder.Body)
				}
			}
		})
	}
}

func TestStatusUsesRequestedDeviceEnvironmentAndNeverGlobalGateway(t *testing.T) {
	for _, scenario := range []struct {
		query, environment string
		code               int
	}{
		{"", "simulation", 200}, {"?device_id=3", "simulation", 200}, {"?device_id=bad", "simulation", 400}, {"?device_id=3", "staging", 422},
	} {
		t.Run(scenario.query+scenario.environment, func(t *testing.T) {
			cfg := &usecase.ZatcaConfig{Enabled: true, Env: "production", BaseURL: "https://wrong-global.invalid"}
			db := &statusDeviceDB{device: repository.ZatcaDeviceConfig{ID: 3, OrganizationID: 1, ZatcaEnv: pgtype.Text{String: scenario.environment, Valid: true}, PrivateKeyPem: "private-key-must-not-leak", Metadata: []byte(`{"production_secret":"secret-must-not-leak"}`)}}
			handler := NewZatcaHandler(usecase.NewZatcaUseCase(cfg))
			engine := gin.New()
			engine.Use(func(c *gin.Context) {
				c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), middleware.RepoKey, repository.New(db)))
			})
			engine.GET("/status", handler.GetZatcaStatus)
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, httptest.NewRequest("GET", "/status"+scenario.query, nil))
			if recorder.Code != scenario.code {
				t.Fatalf("code=%d body=%s", recorder.Code, recorder.Body)
			}
			for _, forbidden := range []string{"wrong-global.invalid", "private-key-must-not-leak", "secret-must-not-leak"} {
				if strings.Contains(recorder.Body.String(), forbidden) {
					t.Fatal("global gateway or credentials leaked")
				}
			}
			if scenario.code == 200 {
				var response struct {
					Data map[string]any `json:"data"`
				}
				_ = json.Unmarshal(recorder.Body.Bytes(), &response)
				if response.Data["environment_source"] != "device_config" {
					t.Fatal("wrong environment source")
				}
				if scenario.query == "" {
					if _, ok := response.Data["env"]; ok {
						t.Fatal("global status pretended to select a device")
					}
				} else if response.Data["env"] != "simulation" || response.Data["base_url"] != "https://gw-fatoora.zatca.gov.sa/e-invoicing/simulation" {
					t.Fatal("wrong device gateway")
				}
			}
		})
	}
}
