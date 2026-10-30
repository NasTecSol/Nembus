package zatca

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NasTecSol/nembus-core/usecase"
	"strings"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type onboardingRequestError struct {
	message string
	status  int
}

func (e onboardingRequestError) Error() string             { return e.message }
func (e onboardingRequestError) OnboardingHTTPStatus() int { return e.status }

func matchOnboardingDevice(device repository.ZatcaDeviceConfig, storeID, terminalID *int32, deviceType, environment string) error {
	if !device.IsActive.Bool || device.IsRevoked.Bool {
		return onboardingRequestError{"This device registration is inactive or revoked; an administrator must review it before onboarding.", 409}
	}
	savedEnvironment, err := usecase.ZatcaDeviceEnvironment(&device)
	if err != nil || savedEnvironment != environment || device.DeviceType != deviceType {
		return onboardingRequestError{"This serial is already registered with a different device type or ZATCA environment. Select the existing registration's settings to resume it.", 409}
	}
	if (storeID != nil && (!device.StoreID.Valid || device.StoreID.Int32 != *storeID)) ||
		(terminalID != nil && (!device.PosTerminalID.Valid || device.PosTerminalID.Int32 != *terminalID)) {
		return onboardingRequestError{"This serial is already registered to a different store or terminal. Select its existing store and terminal to resume onboarding.", 409}
	}
	return nil
}

// onboardingIncomplete exposes progress without returning any saved secrets.
type onboardingIncomplete struct {
	cause  error
	result *OnboardingCompletion
}

func (e *onboardingIncomplete) Error() string { return e.cause.Error() }
func (e *onboardingIncomplete) Unwrap() error { return e.cause }
func (e *onboardingIncomplete) OnboardingFailureDetails() json.RawMessage {
	result := map[string]any{"id": e.result.DeviceID, "device_id": e.result.DeviceID, "zatca_env": e.result.Environment, "onboarding_stage": e.result.Stage, "reporting_ready": e.result.ReportingReady, "compliance_checks": e.result.Checks,
		"completion_endpoint": fmt.Sprintf("/api/zatca/devices/%d/complete-onboarding", e.result.DeviceID)}
	data, _ := json.Marshal(result)
	return data
}

// OnboardDevice is the complete public onboarding operation: obtain compliance
// credentials, save them, check all six samples, exchange and save reporting
// credentials, then reload the finished device configuration.
func (s *OnboardingService) OnboardDevice(ctx context.Context, orgID int32, storeID, posTerminalID *int32, serialNumber, deviceType, otp, environment string) (*repository.ZatcaDeviceConfig, error) {
	if s.cfg == nil {
		return nil, fmt.Errorf("ZATCA configuration unavailable")
	}
	repo, ok := ctx.Value(middleware.RepoKey).(*repository.Queries)
	if !ok || repo == nil {
		return nil, fmt.Errorf("tenant repository is required for device onboarding")
	}
	if err := usecase.RequireOrganizationZatca(ctx, repo, orgID); err != nil {
		return nil, err
	}
	environment = strings.ToLower(strings.TrimSpace(environment))
	if _, err := usecase.ZatcaEnvironmentBaseURL(environment); err != nil {
		return nil, err
	}
	serialNumber = strings.TrimSpace(serialNumber)
	deviceType = strings.ToLower(strings.TrimSpace(deviceType))
	if deviceType == "" {
		deviceType = "pos"
	}
	if serialNumber == "" || (deviceType != "pos" && deviceType != "cloud") {
		return nil, onboardingRequestError{"A device serial and device_type of pos or cloud are required.", 400}
	}
	lookup := func() (repository.ZatcaDeviceConfig, error) {
		return repo.GetZatcaDeviceConfigBySerial(ctx, repository.GetZatcaDeviceConfigBySerialParams{OrganizationID: orgID, DeviceSerial: serialNumber})
	}
	existing, err := lookup()
	var device *repository.ZatcaDeviceConfig
	if err == nil {
		device = &existing
	} else if errors.Is(err, pgx.ErrNoRows) {
		if strings.TrimSpace(otp) == "" {
			return nil, onboardingRequestError{"Enter a fresh OTP from the matching ZATCA portal to register this device for the first time. Existing registrations can resume without an OTP.", 400}
		}
		device, err = s.createComplianceDevice(ctx, orgID, storeID, posTerminalID, serialNumber, deviceType, otp, environment)
		// Another request may have saved this serial after our initial lookup.
		// Reuse its credentials rather than replacing the key or returning SQL.
		var conflict *pgconn.PgError
		if errors.As(err, &conflict) && conflict.Code == "23505" && conflict.ConstraintName == "zatca_device_configs_organization_id_device_serial_key" {
			existing, err = lookup()
			if err == nil {
				device = &existing
			}
		}
	}
	if err != nil {
		return nil, err
	}
	if err := matchOnboardingDevice(*device, storeID, posTerminalID, deviceType, environment); err != nil {
		return nil, err
	}
	if err := usecase.RequireZatcaSigningOwner(device, s.cfg); err != nil {
		return nil, onboardingRequestError{message: err.Error(), status: 409}
	}
	result, err := s.CompleteDeviceOnboarding(ctx, device.ID)
	if result == nil {
		result = &OnboardingCompletion{DeviceID: device.ID, Environment: device.ZatcaEnv.String, Stage: "compliance", Checks: []ComplianceStep{}}
	}
	if err != nil {
		return device, &onboardingIncomplete{cause: fmt.Errorf("compliance credentials saved; %w", err), result: result}
	}
	updated, err := repo.GetZatcaDeviceConfig(ctx, device.ID)
	if err != nil {
		return device, &onboardingIncomplete{cause: fmt.Errorf("reporting credentials saved but device reload failed: %w", err), result: result}
	}
	return &updated, nil
}
