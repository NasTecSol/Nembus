package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	Env             string
	Port            string
	MasterDBURL     string
	JWTSecret       string
	DevUserID       string
	DevUserLogin    string
	LogLevel        string
	IsDesktop       bool
	CloudURL        string
	GRPCAddr        string // bare host:port for the cloud gRPC backup server
	BackupAuthToken string // Auth token for downloading tenant backups
	GithubToken     string // GitHub token for private repo release updates
	GithubRepo      string // Repository owner/name (default: NasTecSol/Nembus)

	// ZATCA Phase 2 Configuration
	ZatcaEnabled  bool   // ZATCA_ENABLED — false for non-Saudi deployments
	ZatcaEnv      string // ZATCA_ENV — "sandbox" or "production"
	ZatcaBaseURL  string // Derived from ZatcaEnv
	ZatcaOrgVATID string // ZATCA_ORG_VAT_ID — 15-digit VAT registration number
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

	// Try to load the environment-specific file
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Note: %s not found, trying .env file", envFile)
		// Fallback to .env if environment-specific file doesn't exist
		if err := godotenv.Load(); err != nil {
			log.Println("Note: .env file not found, using system environment variables")
		}
	}

	// Override with system environment variables if set
	envValue := os.Getenv("ENV")
	if envValue == "" {
		envValue = env
	}

	ghToken := os.Getenv("GITHUB_TOKEN")
	if ghToken == "" {
		ghToken = os.Getenv("GH_TOKEN")
	}

	zatcaEnv := getEnv("ZATCA_ENV", "sandbox")
	var zatcaBaseURL string
	switch zatcaEnv {
	case "production":
		zatcaBaseURL = "https://gw-fatoora.zatca.gov.sa/e-invoicing/core"
	default:
		zatcaBaseURL = "https://gw-fatoora.zatca.gov.sa/e-invoicing/developer-portal"
	}

	cfg := &Config{
		Env:             getEnv("ENV", envValue),
		Port:            getEnv("PORT", "8080"),
		MasterDBURL:     getEnv("MASTER_DB_URL", ""),
		JWTSecret:       getEnv("JWT_SECRET", "nembus-desktop-jwt-secret-2024"),
		DevUserID:       getEnv("DEV_USER_ID", "00000000-0000-0000-0000-000000000000"),
		DevUserLogin:    getEnv("DEV_USER_LOGIN", "dev_user"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		IsDesktop:       true, // Default to true for this Wails build
		CloudURL:        getEnv("CLOUD_URL", "https://nembus.nashrms.com"),
		GRPCAddr:        getEnv("GRPC_SERVER_ADDR", "nembus.nashrms.com:50051"),
		BackupAuthToken: getEnv("BACKUP_AUTH_TOKEN", "nembus"),
		GithubToken:     ghToken,
		GithubRepo:      getEnv("GITHUB_REPO", "NasTecSol/Nembus"),
		ZatcaEnabled:    getEnv("ZATCA_ENABLED", "false") == "true",
		ZatcaEnv:        zatcaEnv,
		ZatcaBaseURL:    zatcaBaseURL,
		ZatcaOrgVATID:   getEnv("ZATCA_ORG_VAT_ID", ""),
	}

	// Propagate resolved secrets back into the OS environment so that any
	// code that reads os.Getenv directly (e.g. JWTAuthMiddleware in auth.go)
	// always finds a non-empty value even in production builds where .env
	// files are not bundled alongside the binary.
	if cfg.JWTSecret != "" {
		os.Setenv("JWT_SECRET", cfg.JWTSecret)
	}

	return cfg
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
