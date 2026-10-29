package zatca

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
)

type ComplianceStep struct {
	DocumentType     string          `json:"document_type"`
	Response         json.RawMessage `json:"response"`
	AlreadyCompleted bool            `json:"already_completed,omitempty"`
}

// A retry can encounter checks completed by an earlier /check or completion
// call. Only recognize the gateway's specific already-completed response for
// the requested sample type; unrelated 406s and mixed validation errors fail.
func complianceAlreadyCompleted(response *ReportingResponse, sampleName string) bool {
	if response == nil || response.HTTPStatus != 406 {
		return false
	}
	var body struct {
		Errors     []json.RawMessage `json:"errors"`
		Validation struct {
			Status string `json:"status"`
			Errors []struct {
				Code     string `json:"code"`
				Category string `json:"category"`
				Message  string `json:"message"`
			} `json:"errorMessages"`
		} `json:"validationResults"`
	}
	if json.Unmarshal(response.RawResponse, &body) != nil || len(body.Errors) != 0 || body.Validation.Status != "ERROR" || len(body.Validation.Errors) != 1 {
		return false
	}
	expected := map[string]string{
		"standard-invoice": "STANDARD", "standard-credit-note": "STANDARD CREDIT NOTE", "standard-debit-note": "STANDARD DEBIT NOTE",
		"simplified-invoice": "SIMPLIFIED", "simplified-credit-note": "SIMPLIFIED CREDIT NOTE", "simplified-debit-note": "SIMPLIFIED DEBIT NOTE",
	}[sampleName]
	if expected == "" {
		return false
	}
	problem := body.Validation.Errors[0]
	message := strings.TrimSpace(problem.Message)
	const prefix = "Compliance check already completed for "
	if problem.Code != "Submitted before" || problem.Category != "Compliance-Check" || !strings.HasPrefix(message, prefix) || !strings.HasSuffix(message, ".") {
		return false
	}
	document := strings.TrimSuffix(strings.TrimPrefix(message, prefix), ".")
	document = strings.NewReplacer("_", " ", "-", " ").Replace(document)
	return document == expected
}

// OnboardingCompletion deliberately excludes certificates, keys, and secrets.
type OnboardingCompletion struct {
	DeviceID       int32            `json:"device_id"`
	Environment    string           `json:"zatca_env"`
	Stage          string           `json:"onboarding_stage"`
	ReportingReady bool             `json:"reporting_ready"`
	Checks         []ComplianceStep `json:"compliance_checks"`
}

func complianceAccepted(response *ReportingResponse) bool {
	if response == nil || response.HTTPStatus < 200 || response.HTTPStatus >= 300 {
		return false
	}
	var body struct {
		Errors     []json.RawMessage `json:"errors"`
		Validation struct {
			Status string            `json:"status"`
			Errors []json.RawMessage `json:"errorMessages"`
		} `json:"validationResults"`
	}
	if json.Unmarshal(response.RawResponse, &body) != nil {
		return false
	}
	return (body.Validation.Status == "PASS" || body.Validation.Status == "WARNING") && len(body.Validation.Errors) == 0 && len(body.Errors) == 0
}

// CompleteDeviceOnboarding reuses saved credentials; no new OTP, keypair, sale,
// or business-chain entry is created. Failed checks never promote the device.
func (s *OnboardingService) CompleteDeviceOnboarding(ctx context.Context, configID int32) (*OnboardingCompletion, error) {
	if s.cfg == nil {
		return nil, fmt.Errorf("ZATCA configuration unavailable")
	}
	// Never fall back to the master database for a device ID supplied over HTTP.
	repo, ok := ctx.Value(middleware.RepoKey).(*repository.Queries)
	if !ok || repo == nil {
		return nil, fmt.Errorf("tenant repository is required for device onboarding completion")
	}
	device, err := repo.GetZatcaDeviceConfig(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("device config not found: %w", err)
	}
	if err := usecase.RequireOrganizationZatca(ctx, repo, device.OrganizationID); err != nil {
		return nil, err
	}
	result := &OnboardingCompletion{DeviceID: device.ID, Environment: device.ZatcaEnv.String, Stage: "compliance", Checks: []ComplianceStep{}}
	if !device.IsActive.Bool || device.IsRevoked.Bool {
		return result, fmt.Errorf("device must be active and not revoked")
	}
	environment, err := usecase.ZatcaDeviceEnvironment(&device)
	if err != nil {
		return result, err
	}
	client, err := s.clients(environment)
	if err != nil {
		return result, err
	}
	var metadata struct {
		ComplianceSecret string `json:"compliance_secret"`
		ProductionSecret string `json:"production_secret"`
		Secret           string `json:"secret"`
		RequestID        string `json:"compliance_request_id"`
	}
	if err := json.Unmarshal(device.Metadata, &metadata); err != nil {
		return result, fmt.Errorf("device credentials could not be read")
	}
	if device.ProductionCsid.String != "" && device.ProductionCsid.String != device.ComplianceCsid.String {
		if metadata.ProductionSecret == "" {
			return result, fmt.Errorf("reporting certificate exists without its matching production_secret")
		}
		expiry, err := usecase.ZatcaCertificateExpiry(device.ProductionCsid.String)
		if err != nil || !expiry.After(time.Now()) {
			return result, fmt.Errorf("existing reporting certificate is invalid or expired; renew the device")
		}
		if err := usecase.ValidateZatcaDeviceCertificate(device.ProductionCsid.String, device.PrivateKeyPem); err != nil {
			return result, err
		}
		result.Stage, result.ReportingReady = "production", true
		return result, nil
	}
	if metadata.ComplianceSecret == "" {
		metadata.ComplianceSecret = metadata.Secret
	}
	if metadata.ComplianceSecret == "" || device.ComplianceCsid.String == "" || strings.HasPrefix(device.ComplianceCsid.String, "SIMULATED_") || metadata.RequestID == "" || metadata.RequestID == "-1" {
		return result, fmt.Errorf("saved compliance CSID, matching secret, and compliance_request_id are required")
	}
	org, err := repo.GetOrganization(ctx, device.OrganizationID)
	if err != nil {
		return result, fmt.Errorf("organization not found: %w", err)
	}
	store := repository.Store{}
	if device.StoreID.Valid {
		store, err = repo.GetStore(ctx, device.StoreID.Int32)
		if err != nil {
			return result, fmt.Errorf("store not found: %w", err)
		}
		if store.OrganizationID != org.ID {
			return result, fmt.Errorf("device store belongs to a different organization")
		}
	}
	samples, err := usecase.BuildZatcaComplianceSamples(&org, &store, &device, s.cfg.OrgVATID)
	if err != nil {
		return result, err
	}
	for _, sample := range samples {
		doc := sample.Document
		response, err := client.SubmitComplianceInvoice(ctx, device.ComplianceCsid.String, metadata.ComplianceSecret, doc.XMLHash, doc.ZatcaUUID, string(doc.RawXML))
		step := ComplianceStep{DocumentType: sample.Name}
		if response != nil {
			step.Response = response.RawResponse
		}
		step.AlreadyCompleted = complianceAlreadyCompleted(response, sample.Name)
		result.Checks = append(result.Checks, step)
		if step.AlreadyCompleted {
			continue
		}
		if err != nil {
			return result, fmt.Errorf("%s compliance check failed: %w", sample.Name, err)
		}
		if !complianceAccepted(response) {
			return result, fmt.Errorf("%s did not pass compliance; inspect compliance_checks", sample.Name)
		}
	}
	reporting, err := client.RequestProductionCSID(ctx, device.ComplianceCsid.String, metadata.ComplianceSecret, metadata.RequestID)
	if err != nil {
		return result, fmt.Errorf("all sample checks passed but reporting CSID exchange failed: %w", err)
	}
	if reporting == nil || reporting.BinarySecurityToken == "" || reporting.Secret == "" || reporting.BinarySecurityToken == device.ComplianceCsid.String {
		return result, fmt.Errorf("reporting CSID response must contain a distinct certificate and matching secret")
	}
	expiry, err := usecase.ZatcaCertificateExpiry(reporting.BinarySecurityToken)
	if err != nil || !expiry.After(time.Now()) {
		return result, fmt.Errorf("gateway returned an invalid or expired reporting certificate")
	}
	// Verify the issued certificate belongs to this device's private key too.
	if err := usecase.ValidateZatcaDeviceCertificate(reporting.BinarySecurityToken, device.PrivateKeyPem); err != nil {
		return result, err
	}
	patch, _ := json.Marshal(map[string]any{
		"secret": reporting.Secret, "production_secret": reporting.Secret,
		"compliance_secret": metadata.ComplianceSecret, "onboarding_stage": "production",
		"compliance_completed_at": time.Now().UTC(), "compliance_checks": result.Checks, "credential_environment": environment,
	})
	if err := repo.CompleteZatcaOnboarding(ctx, device.ID, device.ComplianceCsid.String, reporting.BinarySecurityToken, expiry, patch, device.ZatcaEnv.String); err != nil {
		return result, fmt.Errorf("reporting CSID obtained but saving credentials failed: %w", err)
	}
	result.Stage, result.ReportingReady = "production", true
	return result, nil
}

// JSON adapter keeps the shared handler independent of gateway types.
func (s *OnboardingService) CompleteDeviceOnboardingResponse(ctx context.Context, id int32) (json.RawMessage, error) {
	result, err := s.CompleteDeviceOnboarding(ctx, id)
	if result == nil {
		return nil, err
	}
	data, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return nil, marshalErr
	}
	return data, err
}
