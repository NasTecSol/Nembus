package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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

	// Stage 2C product enrichment execution. Disabled by default so SAP
	// synchronization remains independent of provider availability.
	EnrichmentEnabled        bool
	EnrichmentProvider       string
	OpenAIAPIKey             string
	OpenAIEnrichmentModel    string
	DeepSeekAPIKey           string
	DeepSeekBaseURL          string
	DeepSeekEnrichmentModel  string
	OpenAIEnrichmentTimeout  time.Duration
	EnrichmentWorkerInterval time.Duration
	EnrichmentBatchSize      int
	EnrichmentMaxRetries     int
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

	return &Config{
		Env:                      getEnv("ENV", envValue),
		Port:                     getEnv("PORT", "8080"),
		GRPCPort:                 getEnv("GRPC_PORT", "50051"),
		MasterDBURL:              getEnv("MASTER_DB_URL", ""),
		JWTSecret:                getEnv("JWT_SECRET", ""),
		DevUserID:                getEnv("DEV_USER_ID", "00000000-0000-0000-0000-000000000000"),
		DevUserLogin:             getEnv("DEV_USER_LOGIN", "dev_user"),
		LogLevel:                 getEnv("LOG_LEVEL", "info"),
		PGDumpPath:               getEnv("PG_DUMP_PATH", "pg_dump"),
		ZatcaEnabled:             getEnv("ZATCA_ENABLED", "false") == "true",
		ZatcaEnv:                 zatcaEnv,
		ZatcaBaseURL:             zatcaBaseURL,
		ZatcaOrgVATID:            getEnv("ZATCA_ORG_VAT_ID", ""),
		EnrichmentEnabled:        getEnv("ENRICHMENT_ENABLED", "false") == "true",
		EnrichmentProvider:       getEnv("ENRICHMENT_PROVIDER", "openai"),
		OpenAIAPIKey:             getEnv("OPENAI_API_KEY", ""),
		OpenAIEnrichmentModel:    getEnv("OPENAI_ENRICHMENT_MODEL", "gpt-5.6-terra"),
		DeepSeekAPIKey:           getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekBaseURL:          getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
		DeepSeekEnrichmentModel:  getEnv("DEEPSEEK_MODEL", "deepseek-v4-flash"),
		OpenAIEnrichmentTimeout:  parseDurationEnv("OPENAI_ENRICHMENT_TIMEOUT", 45*time.Second),
		EnrichmentWorkerInterval: parseDurationEnv("OPENAI_ENRICHMENT_WORKER_INTERVAL", 30*time.Second),
		EnrichmentBatchSize:      parseIntEnv("OPENAI_ENRICHMENT_BATCH_SIZE", 5),
		EnrichmentMaxRetries:     parseIntEnv("OPENAI_ENRICHMENT_MAX_RETRIES", 3),
	}
}

// ValidateEnrichmentConfig validates only the optional Stage 2C settings. It
// is intentionally called by server wiring only when enrichment is enabled.
func (c *Config) ValidateEnrichmentConfig() error {
	if c == nil || !c.EnrichmentEnabled {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(c.EnrichmentProvider)) {
	case "openai":
		if c.OpenAIAPIKey == "" {
			return fmt.Errorf("OPENAI_API_KEY is required when enrichment is enabled")
		}
		if c.OpenAIEnrichmentModel == "" {
			return fmt.Errorf("OPENAI_ENRICHMENT_MODEL is required when enrichment is enabled")
		}
	case "deepseek":
		if c.DeepSeekAPIKey == "" {
			return fmt.Errorf("DEEPSEEK_API_KEY is required when enrichment provider is deepseek")
		}
		if c.DeepSeekEnrichmentModel == "" {
			return fmt.Errorf("DEEPSEEK_MODEL is required when enrichment provider is deepseek")
		}
	default:
		return fmt.Errorf("unsupported enrichment provider %q", c.EnrichmentProvider)
	}
	if c.OpenAIEnrichmentTimeout <= 0 || c.EnrichmentWorkerInterval <= 0 || c.EnrichmentBatchSize <= 0 || c.EnrichmentMaxRetries <= 0 {
		return fmt.Errorf("enrichment timeout, interval, batch size, and max retries must be positive")
	}
	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDurationEnv(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return defaultValue
	}
	return parsed
}

func parseIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return defaultValue
	}
	return parsed
}
