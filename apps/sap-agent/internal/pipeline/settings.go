package pipeline

import (
	"github.com/NasTecSol/nembus-sap-agent/internal/config"
)

// Settings holds runtime sync parameters (from agent_config.json).
type Settings struct {
	OrganizationID       int32
	BatchSize            int
	DryRun               bool
	PriceListID          int32
	TaxCategoryID        int32
	DefaultCustomerCode  string
	DefaultWarehouseCode string
	UpstreamOutboxPollMs int
}

// FromAgent maps agent_config.json into pipeline sync settings.
func FromAgent(cfg *config.AgentConfig) *Settings {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	pollMs := cfg.Sync.UpstreamOutboxPollMs
	if pollMs <= 0 {
		pollMs = 3000
	}
	batch := cfg.BatchSize
	if batch <= 0 {
		batch = 500
	}
	orgID := int32(cfg.Cloud.OrganizationID)
	if orgID <= 0 {
		orgID = 1
	}
	return &Settings{
		OrganizationID:       orgID,
		BatchSize:            batch,
		DryRun:               cfg.Sync.DryRun,
		PriceListID:          cfg.Sync.PriceListID,
		TaxCategoryID:        cfg.Sync.TaxCategoryID,
		DefaultCustomerCode:  cfg.Sync.DefaultCustomerCode,
		DefaultWarehouseCode: cfg.Sync.DefaultWarehouseCode,
		UpstreamOutboxPollMs: pollMs,
	}
}
