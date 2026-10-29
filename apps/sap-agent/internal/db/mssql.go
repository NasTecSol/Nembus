package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/NasTecSol/nembus-sap-agent/internal/config"
)


type MSSQLClient struct {
	DB     *sql.DB
	Config config.MSSQLConfig
}

// BuildMSSQLConnectionString constructs standard ADO.NET / TDS connection URL for go-mssqldb
func BuildMSSQLConnectionString(cfg config.MSSQLConfig) string {
	timeout := cfg.ConnectionTimeout
	if timeout <= 0 {
		timeout = 30
	}

	encryptStr := "disable"
	if cfg.Encrypt {
		encryptStr = "true"
	}
	trustCertStr := "true"
	if !cfg.TrustServerCertificate {
		trustCertStr = "false"
	}

	authPart := ""
	if cfg.User == "" {
		authPart = "trusted_connection=yes;"
	} else {
		authPart = fmt.Sprintf("user id=%s;password=%s;", cfg.User, cfg.Password)
	}

	// Support named instances (e.g. "localhost\SQLEXPRESS" or "SERVER\INSTANCE")
	if strings.Contains(cfg.Host, `\`) {
		parts := strings.SplitN(cfg.Host, `\`, 2)
		serverName := parts[0]
		instanceName := parts[1]
		if serverName == "." || serverName == "localhost" {
			serverName = "127.0.0.1"
		}
		return fmt.Sprintf("server=%s;instance=%s;database=%s;%sencrypt=%s;trustservercertificate=%s;connection timeout=%d",
			serverName, instanceName, cfg.Database, authPart, encryptStr, trustCertStr, timeout)
	}

	host := cfg.Host
	if host == "." || host == "" {
		host = "127.0.0.1"
	}

	port := cfg.Port
	if port <= 0 {
		port = 1433
	}

	return fmt.Sprintf("server=%s;port=%d;database=%s;%sencrypt=%s;trustservercertificate=%s;connection timeout=%d",
		host, port, cfg.Database, authPart, encryptStr, trustCertStr, timeout)
}



// NewMSSQLClient creates and validates a connection pool to SAP SQL Server
func NewMSSQLClient(cfg config.MSSQLConfig) (*MSSQLClient, error) {
	connStr := BuildMSSQLConnectionString(cfg)
	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlserver driver: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(10 * time.Minute)

	timeout := cfg.ConnectionTimeout
	if timeout <= 0 {
		timeout = 15
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()


	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping SAP SQL Server at %s:%d: %w", cfg.Host, cfg.Port, err)
	}

	return &MSSQLClient{
		DB:     db,
		Config: cfg,
	}, nil
}

// Ping checks if database is reachable
func (c *MSSQLClient) Ping(ctx context.Context) error {
	if c.DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	return c.DB.PingContext(ctx)
}

// Close closes database connection pool
func (c *MSSQLClient) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}
