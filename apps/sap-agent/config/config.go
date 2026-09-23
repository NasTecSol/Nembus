package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// SAP Service Layer
	SAPServiceLayerURL string
	SAPCompanyDB       string
	SAPUsername        string
	SAPPassword        string
	SAPInsecureSkipSSL bool

	// SAP Direct MSSQL
	SAPMSSQLEnabled  bool
	SAPMSSQLHost     string
	SAPMSSQLPort     int
	SAPMSSQLDatabase string
	SAPMSSQLUser     string
	SAPMSSQLPassword string

	// Nembus
	NembusDBURL             string
	NembusAPIURL            string
	NembusAPIToken          string
	NembusOrganizationID    int32
	NembusStoreID           int32
	NembusPriceListID       int32
	NembusTaxCategoryID     int32
	NembusDefaultCustomer   string
	NembusDefaultWarehouse  string

	// Sync Controls
	DownstreamIntervalSeconds int
	UpstreamOutboxPollMs      int
	BatchSize                 int
	DryRun                    bool
}

func LoadConfig() (*Config, error) {
	// Try to load .env if available (do not fail if absent)
	_ = godotenv.Load(".env")
	_ = godotenv.Load()

	cfg := &Config{
		SAPServiceLayerURL:        getEnv("SAP_SL_URL", "https://127.0.0.1:50000/b1s/v1"),
		SAPCompanyDB:              getEnv("SAP_COMPANY_DB", ""),
		SAPUsername:               getEnv("SAP_USERNAME", ""),
		SAPPassword:               getEnv("SAP_PASSWORD", ""),
		SAPInsecureSkipSSL:        getEnvBool("SAP_INSECURE_SKIP_VERIFY", true),

		SAPMSSQLEnabled:           getEnvBool("SAP_MSSQL_ENABLED", false),
		SAPMSSQLHost:              getEnv("SAP_MSSQL_HOST", "127.0.0.1"),
		SAPMSSQLPort:              getEnvInt("SAP_MSSQL_PORT", 1433),
		SAPMSSQLDatabase:          getEnv("SAP_MSSQL_DATABASE", ""),
		SAPMSSQLUser:              getEnv("SAP_MSSQL_USER", "sa"),
		SAPMSSQLPassword:          getEnv("SAP_MSSQL_PASSWORD", ""),

		NembusDBURL:               getEnv("NEMBUS_DB_URL", "postgres://postgres:password@localhost:5433/nembus?sslmode=disable"),
		NembusAPIURL:              getEnv("NEMBUS_API_URL", "http://localhost:8080/api/v1"),
		NembusAPIToken:            getEnv("NEMBUS_API_TOKEN", ""),
		NembusOrganizationID:      int32(getEnvInt("NEMBUS_ORGANIZATION_ID", 1)),
		NembusStoreID:             int32(getEnvInt("NEMBUS_STORE_ID", 1)),
		NembusPriceListID:         int32(getEnvInt("NEMBUS_PRICE_LIST_ID", 1)),
		NembusTaxCategoryID:       int32(getEnvInt("NEMBUS_TAX_CATEGORY_ID", 1)),
		NembusDefaultCustomer:     getEnv("NEMBUS_DEFAULT_CUSTOMER_CODE", "C000001"),
		NembusDefaultWarehouse:    getEnv("NEMBUS_DEFAULT_WAREHOUSE_CODE", "01"),

		DownstreamIntervalSeconds: getEnvInt("DOWNSTREAM_INTERVAL_SECONDS", 300),
		UpstreamOutboxPollMs:      getEnvInt("UPSTREAM_OUTBOX_POLL_MS", 3000),
		BatchSize:                 getEnvInt("BATCH_SIZE", 100),
		DryRun:                    getEnvBool("DRY_RUN", false),
	}

	return cfg, nil
}

func (c *Config) MSSQLConnectionString() string {
	return fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;encrypt=disable",
		c.SAPMSSQLHost, c.SAPMSSQLUser, c.SAPMSSQLPassword, c.SAPMSSQLPort, c.SAPMSSQLDatabase)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultVal
}
