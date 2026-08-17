package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

type CloudConfig struct {
	BaseURL        string `json:"base_url"`
	APIKey         string `json:"api_key"`
	OrganizationID int    `json:"organization_id"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

type AgentConfig struct {
	Port                 int         `json:"port"`
	SQLitePath           string      `json:"sqlite_path"`
	BatchSize            int         `json:"batch_size"`
	MaxConcurrency       int         `json:"max_concurrency"`
	// DefaultStoreCode is assigned to imported cashiers when SAP has no store mapping.
	// Defaults to "01" if empty.
	DefaultStoreCode     string      `json:"default_store_code"`
	// CashierDrawerLimit is the default cash drawer limit for imported cashiers. Defaults to 5000.
	CashierDrawerLimit   float64     `json:"cashier_drawer_limit"`
	// CashierDiscountLimit is the default max discount % for imported cashiers. Defaults to 20.
	CashierDiscountLimit float64     `json:"cashier_discount_limit"`
	MSSQL                MSSQLConfig `json:"mssql"`
	Cloud                CloudConfig `json:"cloud"`
}

var (
	globalConfig *AgentConfig
	configMutex  sync.RWMutex
	configPath   = "agent_config.json"
)

func DefaultConfig() *AgentConfig {
	return &AgentConfig{
		Port:           17890,
		SQLitePath:     "agent.db",
		BatchSize:      500,
		MaxConcurrency: 4,
		MSSQL: MSSQLConfig{
			Host:                   "192.168.18.77",
			Port:                   1433,
			User:                   "admin",
			Password:               "nastecsol",
			Database:               "Qadsiya",
			Encrypt:                false,
			TrustServerCertificate: true,
			ConnectionTimeout:      15,
		},
		Cloud: CloudConfig{
			BaseURL:        "http://127.0.0.1:8080",
			APIKey:         "",
			OrganizationID: 1,
			TimeoutSeconds: 60,
		},
	}
}

func LoadConfig(customPath ...string) (*AgentConfig, error) {
	configMutex.Lock()
	defer configMutex.Unlock()

	if len(customPath) > 0 && customPath[0] != "" {
		configPath = customPath[0]
	}

	cfg := DefaultConfig()

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

func Get() *AgentConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	if globalConfig == nil {
		return DefaultConfig()
	}
	return globalConfig
}
