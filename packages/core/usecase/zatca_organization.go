package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/NasTecSol/nembus-core/repository"
)

var ErrZatcaDisabled = errors.New("ZATCA is disabled for this organization; enable organizations.metadata.zatca.enabled")

// Missing flags are disabled. Never fall back to a process environment value.
func OrganizationZatcaEnabled(org *repository.Organization) (bool, error) {
	if org == nil {
		return false, fmt.Errorf("organization is required")
	}
	var metadata struct {
		Zatca struct {
			Enabled bool `json:"enabled"`
		} `json:"zatca"`
	}
	if len(org.Metadata) == 0 {
		return false, nil
	}
	if err := json.Unmarshal(org.Metadata, &metadata); err != nil {
		return false, fmt.Errorf("invalid organization ZATCA metadata: %w", err)
	}
	return metadata.Zatca.Enabled, nil
}

func OrganizationZatcaEnabledByID(ctx context.Context, repo *repository.Queries, id int32) (bool, error) {
	org, err := repo.GetOrganization(ctx, id)
	if err != nil {
		return false, fmt.Errorf("organization not found: %w", err)
	}
	return OrganizationZatcaEnabled(&org)
}

func RequireOrganizationZatca(ctx context.Context, repo *repository.Queries, id int32) error {
	enabled, err := OrganizationZatcaEnabledByID(ctx, repo, id)
	if err != nil {
		return err
	}
	if !enabled {
		return ErrZatcaDisabled
	}
	return nil
}
