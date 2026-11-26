package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type MSSQLConfig struct {
	Host                   string `json:"host"`
	Port                   int    `json:"port"`
	User                   string `json:"user"`
	Password               string `json:"password"`
	Database               string `json:"database"`
	Encrypt                bool   `json:"encrypt"`
	TrustServerCertificate bool   `json:"trust_server_certificate"`
	ConnectionTimeout      int    `json:"connection_timeout_seconds"`
}

type ServiceLayerConfig struct {
	URL             string `json:"url"`
	CompanyDB       string `json:"company_db"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	InsecureSkipSSL bool   `json:"insecure_skip_ssl"`
}

type CloudConfig struct {
	BaseURL            string `json:"base_url"`
	APIKey             string `json:"api_key"`
	AppUpdateClientKey string `json:"app_update_client_key,omitempty"`
	OrganizationID     int    `json:"organization_id"`
	TimeoutSeconds     int    `json:"timeout_seconds"`
	DatabaseURL        string `json:"database_url,omitempty"`
}

type AgentConfig struct {
	Version              string      `json:"version,omitempty"`
	Port                 int         `json:"port"`
	SQLitePath           string      `json:"sqlite_path"`
	BatchSize            int         `json:"batch_size"`
	MaxConcurrency       int         `json:"max_concurrency"`
	// DefaultStoreCode is assigned to imported cashiers when SAP has no store mapping.
	// Defaults to "01" if empty.
	DefaultStoreCode     string      `json:"default_store_code"`
	// DefaultCustomerCode is the fallback SAP CardCode used for walk-in or unmapped customers.
	// Defaults to "C000001" if empty.
	DefaultCustomerCode  string      `json:"default_customer_code"`
	// NembusStoreID identifies the Nembus retail store polled for POS upstream transactions.
	// Defaults to 1 if unset.
	NembusStoreID        int         `json:"nembus_store_id"`
	// CashierDrawerLimit is the default cash drawer limit for imported cashiers. Defaults to 5000.
	CashierDrawerLimit   float64     `json:"cashier_drawer_limit"`
	// CashierDiscountLimit is the default max discount % for imported cashiers. Defaults to 20.
	CashierDiscountLimit float64     `json:"cashier_discount_limit"`
	// InvoiceStartDate is the starting date for the invoice window loop (format "YYYY-MM-DD").
	// If empty, defaults to "2000-01-01". Set e.g. to "2023-01-01" to resume an interrupted migration.
	InvoiceStartDate       string      `json:"invoice_start_date,omitempty"`
	// StockMovementStartDate is the starting date for the stock movement window loop (format "YYYY-MM-DD").
	// If empty, defaults to "2024-01-01". Set e.g. to "2024-03-01" to resume an interrupted migration.
	StockMovementStartDate string      `json:"stock_movement_start_date,omitempty"`

	// Onboarding & Multi-Tenant Setup
	OnboardingCompleted bool   `json:"onboarding_completed"`
	TenantSlug          string `json:"tenant_slug"`
	TenantName          string `json:"tenant_name"`
	SAPCompanyDB        string `json:"sap_company_db"`

	// Security & System User Mapping
	MigratorUser     string `json:"migrator_user"`     // Default: "Migrator"
	MigratorPassword string `json:"migrator_password"` // Default: "Migrator"
	MigratorRole     string `json:"migrator_role"`     // Default: "auto-user"
	MigratorToken    string `json:"migrator_token"`    // JWT or API Key token
	SAPUserMapping      string `json:"sap_user_mapping"`       // SAP Operator mapped to Migrator
	SAPPassword         string `json:"sap_password,omitempty"` // SAP Service Layer password (falls back to MSSQL password if empty)
	SAPBranchID         int    `json:"sap_branch_id,omitempty"`
	SAPInvoiceSeries    int    `json:"sap_invoice_series,omitempty"`
	SAPCreditMemoSeries int    `json:"sap_credit_memo_series,omitempty"`
	SAPPaymentSeries    int    `json:"sap_payment_series,omitempty"`

	// Pipeline Scheduling & Automation
	DownstreamIntervalSec     int  `json:"downstream_interval_sec"`     // Default 300s (5m)
	UpstreamIntervalSec       int  `json:"upstream_interval_sec"`       // Default 60s (1m)
	DownstreamAutoActive      bool `json:"downstream_auto_active"`      // Persisted auto-sync toggle
	UpstreamAutoActive        bool `json:"upstream_auto_active"`        // Persisted auto-sync toggle
	ReconciliationIntervalSec int  `json:"reconciliation_interval_sec"` // Default 3600s (1h)
	ReconciliationAutoRun     bool `json:"reconciliation_auto_run"`

	MSSQL        MSSQLConfig        `json:"mssql"`
	ServiceLayer ServiceLayerConfig `json:"service_layer"`
	Cloud        CloudConfig        `json:"cloud"`
}

var (
	globalConfig *AgentConfig
	configMutex  sync.RWMutex
	configPath   = "agent_config.json"
)

func DefaultConfig() *AgentConfig {
	return &AgentConfig{
		Port:                 17890,
		SQLitePath:           "agent.db",
		BatchSize:            500,
		MaxConcurrency:       4,
		DefaultStoreCode:     "01",
		DefaultCustomerCode:  "C000001",
		NembusStoreID:        1,
		CashierDrawerLimit:   5000,
		CashierDiscountLimit: 20,
		InvoiceStartDate:     "2000-01-01",
		OnboardingCompleted:  false,
		TenantSlug:          "default",
		TenantName:          "Default Tenant",
		SAPCompanyDB:        "",
		MigratorUser:        "Migrator",
		MigratorPassword:    "Migrator",
		MigratorRole:        "auto-user",
		MigratorToken:       "",
		SAPUserMapping:      "manager",
		DownstreamIntervalSec:     300,
		UpstreamIntervalSec:       60,
		ReconciliationIntervalSec: 3600,
		ReconciliationAutoRun:     true,
		MSSQL: MSSQLConfig{
			Host:                   "",
			Port:                   1433,
			User:                   "",
			Password:               "",
			Database:               "",
			Encrypt:                false,
			TrustServerCertificate: true,
			ConnectionTimeout:      15,
		},
		ServiceLayer: ServiceLayerConfig{
			InsecureSkipSSL: true,
		},
		Cloud: CloudConfig{
			BaseURL:            "http://127.0.0.1:8080",
			APIKey:             "",
			AppUpdateClientKey: "",
			OrganizationID:     1,
			TimeoutSeconds:     60,
		},
	}
}

// ResolveDataDir returns the persistent OS-level directory for SAP Agent config,
// SQLite state, and logs so that installing/updating binaries never resets settings.
func ResolveDataDir(isService bool) string {
	if envDir := os.Getenv("NEMBUS_SAP_AGENT_DATA_DIR"); envDir != "" {
		return envDir
	}
	if runtime.GOOS == "windows" {
		if isService {
			if pd := os.Getenv("ProgramData"); pd != "" {
				return filepath.Join(pd, "Nembus", "SAPAgent")
			}
		}
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "Nembus", "SAPAgent")
		}
	}
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, "Nembus", "SAPAgent")
	}
	return "."
}

// DefaultConfigPath returns the persistent agent_config.json path.
func DefaultConfigPath(isService bool) string {
	return filepath.Join(ResolveDataDir(isService), "agent_config.json")
}

// DefaultSQLitePath returns the persistent agent.db path.
func DefaultSQLitePath(isService bool) string {
	return filepath.Join(ResolveDataDir(isService), "agent.db")
}

// DefaultLogPath returns the persistent rotating log file path.
func DefaultLogPath(isService bool) string {
	return filepath.Join(ResolveDataDir(isService), "logs", "sap-agent.log")
}

// ConfigPath returns the currently loaded config file path.
func ConfigPath() string {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return configPath
}

// MigrateLegacyFileIfNeeded copies a legacy working-directory file (e.g. ./agent_config.json
// or ./agent.db) into the persistent AppData path once, if the target does not exist yet.
func MigrateLegacyFileIfNeeded(legacyPath, targetPath string) {
	if legacyPath == "" || targetPath == "" || legacyPath == targetPath {
		return
	}
	if _, err := os.Stat(targetPath); err == nil {
		return // already exists at persistent location
	}
	info, err := os.Stat(legacyPath)
	if err != nil || info.IsDir() {
		return
	}
	_ = os.MkdirAll(filepath.Dir(targetPath), 0755)
	src, err := os.Open(legacyPath)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer dst.Close()
	_, _ = io.Copy(dst, src)
}

func LoadConfig(customPath ...string) (*AgentConfig, error) {
	configMutex.Lock()
	defer configMutex.Unlock()

	if len(customPath) > 0 && customPath[0] != "" {
		configPath = customPath[0]
	}

	_ = os.MkdirAll(filepath.Dir(configPath), 0755)
	cfg := DefaultConfig()
	if dir := filepath.Dir(configPath); dir != "" && dir != "." {
		cfg.SQLitePath = filepath.Join(dir, "agent.db")
	}

	// If file doesn't exist, create default
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err == nil {
			_ = os.WriteFile(configPath, data, 0644)
		}
		globalConfig = cfg
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Anchor relative "agent.db" inside the persistent config directory so upgrades never split state.
	if (cfg.SQLitePath == "" || cfg.SQLitePath == "agent.db") && filepath.Dir(configPath) != "." {
		cfg.SQLitePath = filepath.Join(filepath.Dir(configPath), "agent.db")
	}

	globalConfig = cfg
	return cfg, nil
}

func SaveConfig(cfg *AgentConfig) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	dir := filepath.Dir(configPath)
	if dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save config to %s: %w", configPath, err)
	}

	globalConfig = cfg
	return nil
}

// MaskedSecret is the sentinel value returned by Sanitized() for configured secrets.
const MaskedSecret = "********"

// Sanitized returns a shallow copy of cfg with all sensitive credentials masked
// for safe exposure over HTTP API endpoints.
func (c *AgentConfig) Sanitized() *AgentConfig {
	if c == nil {
		return nil
	}
	cp := *c
	mask := func(v string) string {
		if v == "" {
			return ""
		}
		return MaskedSecret
	}
	cp.MigratorPassword = mask(cp.MigratorPassword)
	cp.MigratorToken = mask(cp.MigratorToken)
	cp.SAPPassword = mask(cp.SAPPassword)
	cp.MSSQL.Password = mask(cp.MSSQL.Password)
	cp.ServiceLayer.Password = mask(cp.ServiceLayer.Password)
	cp.Cloud.APIKey = mask(cp.Cloud.APIKey)
	cp.Cloud.AppUpdateClientKey = mask(cp.Cloud.AppUpdateClientKey)
	cp.Cloud.DatabaseURL = mask(cp.Cloud.DatabaseURL)
	return &cp
}

func Get() *AgentConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	if globalConfig == nil {
		return DefaultConfig()
	}
	return globalConfig
}
