package usecase

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/NasTecSol/nembus-core/repository"
)

func ZatcaEnvironmentBaseURL(environment string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(environment)) {
	case "sandbox":
		return "https://gw-fatoora.zatca.gov.sa/e-invoicing/developer-portal", nil
	case "simulation":
		return "https://gw-fatoora.zatca.gov.sa/e-invoicing/simulation", nil
	case "production":
		return "https://gw-fatoora.zatca.gov.sa/e-invoicing/core", nil
	default:
		return "", fmt.Errorf("zatca_env must be sandbox, simulation, or production")
	}
}

// DeviceEnvironment treats the persisted environment as authoritative. The
// credential binding detects accidental environment-only edits on new devices.
func ZatcaDeviceEnvironment(device *repository.ZatcaDeviceConfig) (string, error) {
	environment := strings.ToLower(strings.TrimSpace(device.ZatcaEnv.String))
	if _, err := ZatcaEnvironmentBaseURL(environment); err != nil {
		return "", err
	}
	var metadata struct {
		Environment string `json:"credential_environment"`
	}
	if len(device.Metadata) > 0 {
		if err := json.Unmarshal(device.Metadata, &metadata); err != nil {
			return "", fmt.Errorf("invalid device metadata")
		}
	}
	if metadata.Environment != "" && metadata.Environment != environment {
		return "", fmt.Errorf("device environment differs from its credential environment; onboard separate credentials instead of relabeling this device")
	}
	return environment, nil
}
