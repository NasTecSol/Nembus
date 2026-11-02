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
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type orgZatcaTestDB struct {
	repository.DBTX
	org     repository.Organization
	writes  int
	missing bool
}
type orgZatcaTestRow struct {
	value any
	err   error
}

func (row orgZatcaTestRow) Scan(dest ...interface{}) error {
	if row.err != nil {
		return row.err
	}
	if metadata, ok := row.value.(json.RawMessage); ok {
		*dest[0].(*json.RawMessage) = metadata
		return nil
	}
	value := reflect.ValueOf(row.value)
	for i, target := range dest {
		reflect.ValueOf(target).Elem().Set(value.Field(i))
	}
	return nil
}
func (db *orgZatcaTestDB) QueryRow(_ context.Context, sql string, args ...interface{}) pgx.Row {
	if db.missing {
		return orgZatcaTestRow{err: pgx.ErrNoRows}
	}
	if strings.Contains(sql, "SetOrganizationZatcaEnabled") {
		db.writes++
		// The repository passes only id/bool, never a replacement metadata object.
		if len(args) != 2 {
			panic("toggle must not replace organization metadata")
		}
		var metadata map[string]any
		_ = json.Unmarshal(db.org.Metadata, &metadata)
		if metadata == nil {
			metadata = map[string]any{}
		}
		zatca, _ := metadata["zatca"].(map[string]any)
		if zatca == nil {
			zatca = map[string]any{}
		}
		zatca["enabled"] = args[1].(bool)
		metadata["zatca"] = zatca
		db.org.Metadata, _ = json.Marshal(metadata)
		return orgZatcaTestRow{value: db.org.Metadata}
	}
	return orgZatcaTestRow{value: db.org}
}

func TestOrganizationZatcaEndpointReadsAndUpdatesBooleanOnly(t *testing.T) {
	db := &orgZatcaTestDB{org: repository.Organization{ID: 1, Metadata: []byte(`{"other":"preserve","zatca":{"seller":{"street":"Original street"},"custom":"keep"}}`)}}
	handler := NewOrganizationHandler(nil)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(string(middleware.OrgIDKey), int32(1))
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), middleware.RepoKey, repository.New(db)))
	})
	engine.GET("/organizations/:id/zatca", handler.GetOrganizationZatca)
	engine.PUT("/organizations/:id/zatca", handler.UpdateOrganizationZatca)
	call := func(method, body, path string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(recorder, request)
		return recorder
	}
	if response := call("GET", "", "/organizations/1/zatca"); response.Code != 200 || !strings.Contains(response.Body.String(), `"enabled":false`) {
		t.Fatal("missing flag must read disabled", response.Body)
	}
	for _, enabled := range []string{"true", "false"} {
		response := call("PUT", `{"enabled":`+enabled+`}`, "/organizations/1/zatca")
		if response.Code != 200 || !strings.Contains(response.Body.String(), `"enabled":`+enabled) {
			t.Fatal("boolean update failed", response.Body)
		}
		if !strings.Contains(string(db.org.Metadata), "Original street") || !strings.Contains(string(db.org.Metadata), `"other":"preserve"`) || !strings.Contains(string(db.org.Metadata), `"custom":"keep"`) {
			t.Fatal("unrelated metadata was lost")
		}
	}
	if db.writes != 2 {
		t.Fatal("expected one write per toggle")
	}
	for _, body := range []string{`{}`, `{"enabled":null}`, `{"enabled":"false"}`, `{"enabled":1}`} {
		if response := call("PUT", body, "/organizations/1/zatca"); response.Code != 400 {
			t.Fatal("invalid boolean was accepted", body, response.Body)
		}
	}
	if response := call("PUT", `{"enabled":true}`, "/organizations/2/zatca"); response.Code != 403 {
		t.Fatal("cross-organization update allowed")
	}
	if db.writes != 2 {
		t.Fatal("invalid requests changed metadata")
	}
	db.missing = true
	if response := call("GET", "", "/organizations/1/zatca"); response.Code != 404 {
		t.Fatal("missing organization was not rejected")
	}
}
