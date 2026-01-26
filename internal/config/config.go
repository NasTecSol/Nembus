package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	Env          string
	Port         string
	MasterDBURL  string
	JWTSecret    string
	DevUserID    string
	DevUserLogin string
	LogLevel     string
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

	return &Config{
		Env:          getEnv("ENV", envValue),
		Port:         getEnv("PORT", "8080"),
		MasterDBURL:  getEnv("MASTER_DB_URL", ""),
		JWTSecret:    getEnv("JWT_SECRET", ""),
		DevUserID:    getEnv("DEV_USER_ID", "00000000-0000-0000-0000-000000000000"),
		DevUserLogin: getEnv("DEV_USER_LOGIN", "dev_user"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
