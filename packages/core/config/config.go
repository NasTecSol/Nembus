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
	ZatcaEnabled  bool   // ZATCA_ENABLED — false for non-Saudi deployments
	ZatcaEnv      string // ZATCA_ENV — "sandbox" or "production"
	ZatcaBaseURL  string // Derived from ZatcaEnv
	ZatcaOrgVATID string // ZATCA_ORG_VAT_ID — 15-digit VAT registration number

	// gRPC Sync Security & TLS
	SyncJWTSecret   string // SYNC_JWT_SECRET (min 32 bytes)
	SyncTLSEnabled  bool   // SYNC_TLS_ENABLED
	SyncTLSCertFile string // SYNC_TLS_CERT_FILE
	SyncTLSKeyFile  string // SYNC_TLS_KEY_FILE
	SyncEnforceAuth bool   // SYNC_ENFORCE_AUTH
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
	default:
		zatcaBaseURL = "https://gw-fatoora.zatca.gov.sa/e-invoicing/developer-portal"
	}

	syncSecret := getEnv("SYNC_JWT_SECRET", "")
	if syncSecret == "" {
		// Fallback to JWT_SECRET if long enough
		if jwtSec := getEnv("JWT_SECRET", ""); len(jwtSec) >= 32 {
			syncSecret = jwtSec
		}
	}

	return &Config{
		Env:             getEnv("ENV", envValue),
		Port:            getEnv("PORT", "8080"),
		GRPCPort:        getEnv("GRPC_PORT", "50051"),
		MasterDBURL:     getEnv("MASTER_DB_URL", ""),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		DevUserID:       getEnv("DEV_USER_ID", "00000000-0000-0000-0000-000000000000"),
		DevUserLogin:    getEnv("DEV_USER_LOGIN", "dev_user"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		PGDumpPath:      getEnv("PG_DUMP_PATH", "pg_dump"),
		ZatcaEnabled:    getEnv("ZATCA_ENABLED", "false") == "true",
		ZatcaEnv:        zatcaEnv,
		ZatcaBaseURL:    zatcaBaseURL,
		ZatcaOrgVATID:   getEnv("ZATCA_ORG_VAT_ID", ""),
		SyncJWTSecret:   syncSecret,
		SyncTLSEnabled:  getEnv("SYNC_TLS_ENABLED", "false") == "true",
		SyncTLSCertFile: getEnv("SYNC_TLS_CERT_FILE", ""),
		SyncTLSKeyFile:  getEnv("SYNC_TLS_KEY_FILE", ""),
		SyncEnforceAuth: getEnv("SYNC_ENFORCE_AUTH", "false") == "true" || syncSecret != "",
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
