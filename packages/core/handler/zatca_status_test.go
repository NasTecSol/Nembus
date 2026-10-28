package handler

import (
	"context"
	"encoding/json"
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
	device repository.ZatcaDeviceConfig
}
type statusDeviceRow struct{ value interface{} }

func (db *statusDeviceDB) QueryRow(_ context.Context, sql string, args ...interface{}) pgx.Row {
	if strings.Contains(sql, "GetOrganization") {
		return statusDeviceRow{repository.Organization{ID: 1, Metadata: []byte(`{"zatca":{"enabled":true}}`)}}
	}
	return statusDeviceRow{db.device}
}
func (row statusDeviceRow) Scan(dest ...interface{}) error {
	value := reflect.ValueOf(row.value)
	for i, target := range dest {
		reflect.ValueOf(target).Elem().Set(value.Field(i))
	}
	return nil
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
