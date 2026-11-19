package server

import (
	"bytes"
	"embed"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/NasTecSol/nembus-sap-agent/internal/config"
	"github.com/NasTecSol/nembus-sap/contracts"
	"github.com/gin-gonic/gin"
)

func TestDownstreamMasterDomainsIncludesInventory(t *testing.T) {
	for _, want := range []contracts.DomainType{contracts.DomainUOM, contracts.DomainUOMGroups, contracts.DomainInventory} {
		if !slices.Contains(DownstreamMasterDomains, want) {
			t.Fatalf("expected DownstreamMasterDomains to include %q, got %v", want, DownstreamMasterDomains)
		}
	}
	uomIdx := slices.Index(DownstreamMasterDomains, contracts.DomainUOM)
	groupsIdx := slices.Index(DownstreamMasterDomains, contracts.DomainUOMGroups)
	productsIdx := slices.Index(DownstreamMasterDomains, contracts.DomainProducts)
	if !(uomIdx < groupsIdx && groupsIdx < productsIdx) {
		t.Fatalf("expected uom < uom_groups < products ordering, got %v", DownstreamMasterDomains)
	}
}

func TestBuildPipeConfigUsesConfiguredStoreID(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "agent_config.json")
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	s := &Server{}
	cfg.NembusStoreID = 42
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	if got := s.buildPipeConfig().NembusStoreID; got != 42 {
		t.Fatalf("configured NembusStoreID: got %d want 42", got)
	}

	cfg.NembusStoreID = 0
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	t.Setenv("NEMBUS_STORE_ID", "7")
	if got := s.buildPipeConfig().NembusStoreID; got != 7 {
		t.Fatalf("env fallback NembusStoreID: got %d want 7", got)
	}
}

func TestBuildPipeConfigUsesConfiguredDefaultCustomer(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "agent_config.json")
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	s := &Server{}
	cfg.DefaultCustomerCode = "C-WALKIN-99"
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	if got := s.buildPipeConfig().NembusDefaultCustomer; got != "C-WALKIN-99" {
		t.Fatalf("configured DefaultCustomerCode: got %q want C-WALKIN-99", got)
	}

	cfg.DefaultCustomerCode = ""
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	t.Setenv("NEMBUS_DEFAULT_CUSTOMER_CODE", "C-ENV-02")
	if got := s.buildPipeConfig().NembusDefaultCustomer; got != "C-ENV-02" {
		t.Fatalf("env fallback NembusDefaultCustomer: got %q want C-ENV-02", got)
	}
}

func TestNextScheduledRunTime(t *testing.T) {
	runStart := time.Date(2026, 11, 13, 7, 0, 0, 0, time.UTC)
	interval := 5 * time.Minute

	// Run finishes in 28s -> next run is 07:05:00 (not 07:05:28)
	now := runStart.Add(28 * time.Second)
	want := time.Date(2026, 11, 13, 7, 5, 0, 0, time.UTC)
	if got := nextScheduledRunTime(runStart, now, interval); !got.Equal(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}

	// Run takes longer than interval -> next run clamps to now
	overrun := runStart.Add(6 * time.Minute)
	if got := nextScheduledRunTime(runStart, overrun, interval); !got.Equal(overrun) {
		t.Fatalf("expected %v on overrun, got %v", overrun, got)
	}
}

func TestBuildPipeConfigSAPCompanyDB(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "agent_config.json")
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	s := &Server{}

	// Stale default "Qadsiya" should defer to configured MSSQL.Database
	cfg.SAPCompanyDB = "Qadsiya"
	cfg.MSSQL.Database = "QITAFALAELAHLTD"
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	if got := s.buildPipeConfig().SAPCompanyDB; got != "QITAFALAELAHLTD" {
		t.Fatalf("stale default fallback: got %q want %q", got, "QITAFALAELAHLTD")
	}

	// Explicit non-default SAPCompanyDB should be respected
	cfg.SAPCompanyDB = "EBTEKAR_SAP"
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	if got := s.buildPipeConfig().SAPCompanyDB; got != "EBTEKAR_SAP" {
		t.Fatalf("explicit SAPCompanyDB: got %q want %q", got, "EBTEKAR_SAP")
	}

	// Env var overrides config
	t.Setenv("SAP_COMPANY_DB", "ENV_SAP_DB")
	if got := s.buildPipeConfig().SAPCompanyDB; got != "ENV_SAP_DB" {
		t.Fatalf("env override SAPCompanyDB: got %q want %q", got, "ENV_SAP_DB")
	}
}

func TestSaveConfigAndOnboardingSyncSAPCompanyDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfgPath := filepath.Join(t.TempDir(), "agent_config.json")
	if _, err := config.LoadConfig(cfgPath); err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	s := &Server{}
	r := gin.New()
	r.POST("/config", s.handleSaveConfig)
	r.POST("/onboarding/complete", s.handleOnboardingComplete)

	// 1. SaveConfig without explicit sap_company_db should sync from mssql.database
	req := httptest.NewRequest(http.MethodPost, "/config", bytes.NewBufferString(`{"mssql":{"host":"192.168.1.90","port":1433,"user":"sa","password":"p","database":"EBTEKAR"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("handleSaveConfig status = %d, body = %s", w.Code, w.Body.String())
	}
	if got := config.Get().SAPCompanyDB; got != "EBTEKAR" {
		t.Fatalf("after handleSaveConfig: SAPCompanyDB = %q, want %q", got, "EBTEKAR")
	}

	// 2. OnboardingComplete without explicit sap_company_db should sync from mssql.database
	req2 := httptest.NewRequest(http.MethodPost, "/onboarding/complete", bytes.NewBufferString(`{"mssql":{"host":"192.168.1.90","port":1433,"user":"sa","password":"p","database":"QITAFALAELAHLTD"}}`))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("handleOnboardingComplete status = %d, body = %s", w2.Code, w2.Body.String())
	}
	if got := config.Get().SAPCompanyDB; got != "QITAFALAELAHLTD" {
		t.Fatalf("after handleOnboardingComplete: SAPCompanyDB = %q, want %q", got, "QITAFALAELAHLTD")
	}
}

func TestNewServerRestoresPersistedAutoSyncState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var emptyFS embed.FS

	cfg := config.DefaultConfig()
	cfg.OnboardingCompleted = true
	cfg.DownstreamAutoActive = true
	cfg.UpstreamAutoActive = false
	s := NewServer(cfg, nil, nil, nil, nil, emptyFS)
	if !s.isDownstreamAutoActive() {
		t.Fatalf("expected downstreamAutoActive=true when OnboardingCompleted && DownstreamAutoActive")
	}
	if s.isUpstreamAutoActive() {
		t.Fatalf("expected upstreamAutoActive=false when UpstreamAutoActive=false")
	}

	cfg.OnboardingCompleted = false
	cfg.UpstreamAutoActive = true
	s2 := NewServer(cfg, nil, nil, nil, nil, emptyFS)
	if s2.isDownstreamAutoActive() || s2.isUpstreamAutoActive() {
		t.Fatalf("expected auto-sync flags false when OnboardingCompleted=false")
	}
}

func TestBuildPipeConfigUsesServiceLayerConfig(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "agent_config.json")
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	cfg.MSSQL.Host = "10.0.0.10"
	cfg.MSSQL.Password = "mssql-secret"
	cfg.ServiceLayer = config.ServiceLayerConfig{
		URL:             "https://sap-sl.internal:50000/b1s/v1",
		CompanyDB:       "PROD_SAP",
		Username:        "b1_api_user",
		Password:        "b1-secret",
		InsecureSkipSSL: false,
	}
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	s := &Server{}
	got := s.buildPipeConfig()
	if got.SAPServiceLayerURL != "https://sap-sl.internal:50000/b1s/v1" {
		t.Fatalf("SAPServiceLayerURL = %q", got.SAPServiceLayerURL)
	}
	if got.SAPCompanyDB != "PROD_SAP" {
		t.Fatalf("SAPCompanyDB = %q", got.SAPCompanyDB)
	}
	if got.SAPUsername != "b1_api_user" || got.SAPPassword != "b1-secret" {
		t.Fatalf("unexpected SAP credentials: user=%q pass=%q", got.SAPUsername, got.SAPPassword)
	}
}

func TestOnboardingFailsExplicitlyWhenCloudUnreachable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfgPath := filepath.Join(t.TempDir(), "agent_config.json")
	if _, err := config.LoadConfig(cfgPath); err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	// Closed test server guarantees connection refused
	closedSrv := httptest.NewServer(http.NewServeMux())
	closedURL := closedSrv.URL
	closedSrv.Close()

	s := &Server{}
	r := gin.New()
	r.POST("/onboarding/tenants", s.handleOnboardingTenants)
	r.POST("/onboarding/login", s.handleOnboardingLogin)

	req1 := httptest.NewRequest(http.MethodPost, "/onboarding/tenants", bytes.NewBufferString(`{"base_url":"`+closedURL+`"}`))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code == http.StatusOK {
		t.Fatalf("expected non-200 when cloud unreachable for tenants, got 200: %s", w1.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/onboarding/login", bytes.NewBufferString(`{"base_url":"`+closedURL+`","tenant_slug":"t1","user_login":"Migrator","password":"p"}`))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code == http.StatusOK {
		t.Fatalf("expected non-200 when cloud unreachable for login, got 200: %s", w2.Body.String())
	}
}

func TestHandleOnboardingSAPEntities_MissingHost(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "agent_config.json")
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	cfg.MSSQL.Host = ""
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	s := &Server{cfg: cfg}
	r.POST("/api/v1/onboarding/sap-entities", s.handleOnboardingSAPEntities)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/onboarding/sap-entities", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Success || resp.Error == "" {
		t.Fatalf("expected failure with error message, got %+v", resp)
	}
}