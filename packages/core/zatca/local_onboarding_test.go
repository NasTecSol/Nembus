package zatca

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NasTecSol/nembus-core/handler"
	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	router "github.com/NasTecSol/nembus-core/routing"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type localOnboardingDB struct{ onboardingFlowDB }

func (db *localOnboardingDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if strings.Contains(sql, "GetStore") {
		return submissionRow{repository.Store{ID: 17, OrganizationID: 1}}
	}
	return db.onboardingFlowDB.QueryRow(ctx, sql, args...)
}

func TestLocalOnboardingHTTPPersistsCredentialsInRequestTenant(t *testing.T) {
	ca, err := usecase.GenerateECDSAKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	db := &localOnboardingDB{onboardingFlowDB: onboardingFlowDB{completionDB: completionDB{org: repository.Organization{
		ID: 1, Name: "Local Seller", TaxID: pgtype.Text{String: "310123456700003", Valid: true},
		Metadata: []byte(`{"zatca":{"enabled":true,"seller":{"commercial_registration_number":"1010123456","street":"Street","building_number":"1234","district":"District","city":"Riyadh","postal_code":"12345"}}}`),
	}}}}
	var compliance, production string
	credentials, checks, exchanges := 0, 0, 0
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/compliance":
			credentials++
			if r.Header.Get("OTP") != "654321" {
				t.Error("OTP was not forwarded to ZATCA")
			}
			var body ComplianceCSIDRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
				w.WriteHeader(400)
				return
			}
			pub := subjectFromOnboardingCSR(t, body.CSR)
			compliance = completionCertificateWithSubject(t, pub, ca, 1)
			production = completionCertificateWithSubject(t, pub, ca, 2)
			_ = json.NewEncoder(w).Encode(map[string]any{"requestID": 12345, "binarySecurityToken": compliance, "secret": "local-compliance-secret"})
		case "/compliance/invoices":
			checks++
			if db.creates != 1 {
				t.Error("checks ran before credentials were persisted locally")
			}
			_, _ = w.Write([]byte(`{"validationResults":{"status":"PASS","errorMessages":[]}}`))
		case "/production/csids":
			exchanges++
			if checks != 6 {
				t.Error("reporting credentials requested before all checks")
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"binarySecurityToken": production, "secret": "local-production-secret"})
		default:
			t.Error("unexpected gateway call", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer gateway.Close()
	master := &onboardingLookupDB{}
	cfg := &usecase.ZatcaConfig{Env: "production"}
	service := testOnboardingService(repository.New(master), NewZatcaClient(gateway.URL), cfg)
	h := handler.NewZatcaHandler(usecase.NewZatcaUseCase(cfg))
	h.SetOnboarder(service)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set(string(middleware.OrgIDKey), int32(1))
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), middleware.RepoKey, repository.New(db)))
	})
	router.RegisterZatcaRoutes(engine.Group("/api"), h)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/api/zatca/onboard", strings.NewReader(`{"organization_id":1,"store_id":17,"pos_terminal_id":18,"device_serial":"LOCAL-POS-18","device_type":"pos","otp":"654321","zatca_env":"simulation"}`))
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(recorder, request)
	if recorder.Code != 200 || !strings.Contains(recorder.Body.String(), `"reporting_ready":true`) {
		t.Fatalf("local onboarding failed: HTTP %d %s", recorder.Code, recorder.Body)
	}
	if credentials != 1 || checks != 6 || exchanges != 1 || db.creates != 1 || db.writes != 1 || master.lookups != 0 {
		t.Fatalf("incorrect local persistence or gateway calls: credentials=%d checks=%d exchanges=%d creates=%d writes=%d master=%d", credentials, checks, exchanges, db.creates, db.writes, master.lookups)
	}
	if db.device.PosTerminalID.Int32 != 18 || db.device.StoreID.Int32 != 17 || db.device.ZatcaEnv.String != "simulation" || db.device.ProductionCsid.String != production || !strings.Contains(string(db.device.Metadata), "local-production-secret") {
		t.Fatal("local device credentials or location were not persisted")
	}
	if strings.Contains(recorder.Body.String(), "PRIVATE KEY") || strings.Contains(recorder.Body.String(), "local-production-secret") {
		t.Fatal("private credentials leaked into HTTP response")
	}
	// Recovery is available on the same local backend and reuses its saved device.
	resume := httptest.NewRecorder()
	engine.ServeHTTP(resume, httptest.NewRequest("POST", "/api/zatca/devices/3/complete-onboarding", nil))
	if resume.Code != 200 || credentials != 1 || checks != 6 || exchanges != 1 {
		t.Fatalf("local completion did not reuse saved credentials: HTTP %d %s", resume.Code, resume.Body)
	}
}
