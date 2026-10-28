package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

// SelectZatcaDevice refuses to choose an environment by recency when multiple
// registrations exist. An order can explicitly bind metadata.zatca_device_config_id.
func (uc *ZatcaUseCase) SelectZatcaDevice(ctx context.Context, orgID int32, terminalID pgtype.Int4, metadata json.RawMessage) (repository.ZatcaDeviceConfig, error) {
	var choice struct {
		ID *int32 `json:"zatca_device_config_id"`
	}
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &choice); err != nil {
			return repository.ZatcaDeviceConfig{}, fmt.Errorf("invalid order metadata")
		}
	}
	var device repository.ZatcaDeviceConfig
	var err error
	if choice.ID != nil {
		if *choice.ID <= 0 {
			return device, fmt.Errorf("zatca_device_config_id must be positive")
		}
		device, err = uc.repo.GetZatcaDeviceConfig(ctx, *choice.ID)
		if err != nil {
			return device, err
		}
		if device.DeviceType == "pos" && (!terminalID.Valid || device.PosTerminalID != terminalID) {
			return device, fmt.Errorf("selected ZATCA device belongs to a different terminal")
		}
	} else {
		count := int64(0)
		if terminalID.Valid {
			count, err = uc.repo.CountActiveZatcaDevices(ctx, orgID, terminalID, "pos")
			if err != nil {
				return device, err
			}
			if count > 1 {
				return device, fmt.Errorf("multiple active ZATCA devices for this terminal; select metadata.zatca_device_config_id explicitly")
			}
		}
		if count == 1 {
			device, err = uc.repo.GetActivePosDevice(ctx, repository.GetActivePosDeviceParams{OrganizationID: orgID, PosTerminalID: terminalID})
		} else {
			count, err = uc.repo.CountActiveZatcaDevices(ctx, orgID, pgtype.Int4{}, "cloud")
			if err != nil {
				return device, err
			}
			if count != 1 {
				return device, fmt.Errorf("expected one active cloud ZATCA device, found %d; select metadata.zatca_device_config_id explicitly", count)
			}
			device, err = uc.repo.GetActiveCloudDevice(ctx, orgID)
		}
		if err != nil {
			return device, err
		}
	}
	if device.OrganizationID != orgID || !device.IsActive.Bool || device.IsRevoked.Bool {
		return device, fmt.Errorf("ZATCA device must be active, not revoked, and belong to the order organization")
	}
	if _, err := ZatcaDeviceEnvironment(&device); err != nil {
		return device, err
	}
	return device, nil
}
