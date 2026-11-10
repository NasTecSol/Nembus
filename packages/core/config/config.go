package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	Env          string
	Port         string
	GRPCPort     string // gRPC backup service port (default: 50051)
	MasterDBURL  string
	JWTSecret    string
	DevUserID    string
	DevUserLogin string
	LogLevel     string
	PGDumpPath   string // full path to pg_dump binary

	// ZATCA Phase 2 Configuration
	ZatcaEnabled  bool   // Legacy configuration; runtime enablement uses organization metadata
	ZatcaEnv      string // ZATCA_ENV — "sandbox" or "production"
	ZatcaBaseURL  string // Derived from ZatcaEnv
	ZatcaOrgVATID string // ZATCA_ORG_VAT_ID — 15-digit VAT registration number

	// Release & Update Configuration
	ReleaseAPISecret   string // Secret token for CI/CD or admin to upload new builds
	AppUpdateClientKey string // Secret token embedded in POS to check/download updates
	UpdatesStorageDir  string // Local filesystem directory where update binaries are stored
}

// LoadConfig loads configuration from environment file based on environment
func LoadConfig(env string) *Config {
	// Determine which config file to load
	var envFile string
	switch env {
	case "development", "dev":
		envFile = ".env.dev"
	case "staging", "stg":
		envFile = ".env.stg"
	default:
		envFile = ".env"
	}

	// Load files in order of precedence (.env.local overrides .env.dev, which overrides .env)
	_ = godotenv.Load(".env.local")
	if envFile != "" {
		_ = godotenv.Load(envFile + ".local")
		if err := godotenv.Load(envFile); err != nil && envFile != ".env" {
			_ = godotenv.Load()
		}
	} else {
		_ = godotenv.Load()
	}

	// Override with system environment variables if set
	envValue := os.Getenv("ENV")
	if envValue == "" {
		envValue = env
	}

	zatcaEnv := getEnv("ZATCA_ENV", "sandbox")
	var zatcaBaseURL string
	switch zatcaEnv {
	case "production":
		zatcaBaseURL = "https://gw-fatoora.zatca.gov.sa/e-invoicing/core"
	case "simulation":
		zatcaBaseURL = "https://gw-fatoora.zatca.gov.sa/e-invoicing/simulation"
	default:
		zatcaBaseURL = "https://gw-fatoora.zatca.gov.sa/e-invoicing/developer-portal"
	}

	return &Config{
		Env:                getEnv("ENV", envValue),
		Port:               getEnv("PORT", "8080"),
		GRPCPort:           getEnv("GRPC_PORT", "50051"),
		MasterDBURL:        getEnv("MASTER_DB_URL", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		DevUserID:          getEnv("DEV_USER_ID", "00000000-0000-0000-0000-000000000000"),
		DevUserLogin:       getEnv("DEV_USER_LOGIN", "dev_user"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		PGDumpPath:         getEnv("PG_DUMP_PATH", "pg_dump"),
		ZatcaEnabled:       getEnv("ZATCA_ENABLED", "false") == "true",
		ZatcaEnv:           zatcaEnv,
		ZatcaBaseURL:       zatcaBaseURL,
		ZatcaOrgVATID:      getEnv("ZATCA_ORG_VAT_ID", ""),
		ReleaseAPISecret:   getEnv("RELEASE_API_SECRET", "nembus-release-secret-change-in-prod"),
		AppUpdateClientKey: getEnv("APP_UPDATE_CLIENT_KEY", "nembus-pos-update-client-key-default"),
		UpdatesStorageDir:  getEnv("UPDATES_STORAGE_DIR", "uploads/releases"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
