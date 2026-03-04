package db

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
)

type DBManager struct {
	postgres *embeddedpostgres.EmbeddedPostgres
	config   Config
}

type Config struct {
	Username string
	Password string
	Database string
	Port     uint32
	DataPath string
}

func NewDBManager(cfg Config) *DBManager {
	if cfg.DataPath == "" {
		// Use app data directory or local folder
		home, _ := os.UserHomeDir()
		cfg.DataPath = filepath.Join(home, ".nembus", "data")
	}
	return &DBManager{
		config: cfg,
	}
}

func (m *DBManager) Start() error {
	binPath := filepath.Join(m.config.DataPath, "bin")
	runtimePath := filepath.Join(m.config.DataPath, "runtime")

	cfg := embeddedpostgres.
		DefaultConfig().
		Username(m.config.Username).
		Password(m.config.Password).
		Database(m.config.Database).
		Port(m.config.Port).
		Logger(io.Discard).
		RuntimePath(runtimePath).
		BinariesPath(binPath)

	m.postgres = embeddedpostgres.NewDatabase(cfg)

	if err := m.postgres.Start(); err != nil {
		// Workaround for occasional "failed to clean up run time directory" errors:
		// remove the runtime directory and retry once.
		if strings.Contains(err.Error(), "failed to clean up run time directory") {
			_ = os.RemoveAll(runtimePath)
			m.postgres = embeddedpostgres.NewDatabase(cfg)
			if retryErr := m.postgres.Start(); retryErr == nil {
				log.Printf("Embedded PostgreSQL started on port %d (after cleaning runtime dir)", m.config.Port)
				return nil
			}
		}
		return fmt.Errorf("failed to start embedded postgres: %w", err)
	}

	log.Printf("Embedded PostgreSQL started on port %d", m.config.Port)
	return nil
}

func (m *DBManager) Stop() error {
	if m.postgres != nil {
		if err := m.postgres.Stop(); err != nil {
			return fmt.Errorf("failed to stop embedded postgres: %w", err)
		}
	}
	return nil
}

func (m *DBManager) GetConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@localhost:%d/%s?sslmode=disable",
		m.config.Username, m.config.Password, m.config.Port, m.config.Database)
}
