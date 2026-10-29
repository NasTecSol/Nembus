package zatca

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/NasTecSol/nembus-core/usecase"
	"strings"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
)

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
	device, err := s.createComplianceDevice(ctx, orgID, storeID, posTerminalID, serialNumber, deviceType, otp, environment)
	if err != nil {
		return nil, err
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
