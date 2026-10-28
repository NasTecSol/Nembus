package zatca

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/jackc/pgx/v5/pgtype"
)

type OnboardingService struct {
	repo    *repository.Queries
	clients environmentClientFactory
	cfg     *usecase.ZatcaConfig
}

func NewOnboardingService(repo *repository.Queries, cfg *usecase.ZatcaConfig) *OnboardingService {
	return &OnboardingService{
		repo:    repo,
		clients: newEnvironmentClient,
		cfg:     cfg,
	}
}

// createComplianceDevice persists credentials before the remaining gateway
// calls so incomplete onboarding can resume without consuming another OTP.
func (s *OnboardingService) createComplianceDevice(ctx context.Context, orgID int32, storeID *int32, posTerminalID *int32, serialNumber string, deviceType string, otp string, environment string) (*repository.ZatcaDeviceConfig, error) {
	// HTTP tenant middleware selects the database for this request. Keep it
	// local so concurrent onboarding requests cannot change each other's repo.
	repo := s.repo
	if tenantRepo, ok := ctx.Value(middleware.RepoKey).(*repository.Queries); ok && tenantRepo != nil {
		repo = tenantRepo
	}
	if repo == nil {
		return nil, fmt.Errorf("repository not set")
	}

	org, err := repo.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}
	csrSerial, err := csrDeviceSerial(serialNumber, deviceType)
	if err != nil {
		return nil, err
	}

	// 1. Generate Keypair
	privKey, err := usecase.GenerateECDSAKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	privKeyPEM, err := usecase.EncodePrivateKeyPEM(privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode private key: %w", err)
	}

	client, err := s.clients(environment)
	if err != nil {
		return nil, err
	}
	isSandbox := environment != "production"

	cn := fmt.Sprintf("TST-%d-%s", orgID, serialNumber)
	csrParams := usecase.ZatcaCSRParams{
		CommonName:             cn,
		SerialNumber:           csrSerial,
		OrganizationIdentifier: org.TaxID.String,
		OrgName:                org.LegalName.String,
		OrgUnit:                "Head Office",
		CountryCode:            "SA",
		InvoiceType:            "1100",
		Location:               "Riyadh",
		Industry:               "Retail",
		IsSandbox:              isSandbox,
	}
	if csrParams.OrgName == "" {
		csrParams.OrgName = org.Name
	}
	if csrParams.OrganizationIdentifier == "" {
		csrParams.OrganizationIdentifier = s.cfg.OrgVATID
	}

	csrPEM, err := usecase.GenerateCSR(privKey, csrParams)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CSR: %w", err)
	}

	// 2. Request Compliance CSID
	ccsidResp, err := client.RequestComplianceCSID(ctx, csrPEM, otp)
	if err != nil {
		return nil, fmt.Errorf("compliance CSID request failed: %w", err)
	}

	requestID := ccsidResp.RequestID.String()
	if requestID == "" || requestID == "-1" {
		return nil, fmt.Errorf("compliance response did not include a valid requestID")
	}
	complianceToken := ccsidBinaryToken(ccsidResp)
	if complianceToken == "" || ccsidResp.Secret == "" {
		return nil, fmt.Errorf("compliance response must include a certificate and matching secret")
	}
	expiry, err := usecase.ZatcaCertificateExpiry(complianceToken)
	if err != nil {
		return nil, fmt.Errorf("invalid compliance certificate: %w", err)
	}

	var sID pgtype.Int4
	if storeID != nil {
		sID = pgtype.Int4{Int32: *storeID, Valid: true}
	}
	var tID pgtype.Int4
	if posTerminalID != nil {
		tID = pgtype.Int4{Int32: *posTerminalID, Valid: true}
	}

	metaJSON, _ := json.Marshal(map[string]interface{}{
		"secret":                 ccsidResp.Secret,
		"compliance_secret":      ccsidResp.Secret,
		"production_secret":      "",
		"compliance_request_id":  requestID,
		"onboarding_stage":       "compliance",
		"invoice_type":           csrParams.InvoiceType,
		"credential_environment": environment,
	})

	params := repository.CreateZatcaDeviceConfigParams{
		OrganizationID: orgID,
		StoreID:        sID,
		PosTerminalID:  tID,
		DeviceSerial:   serialNumber,
		DeviceType:     deviceType,
		CsrPem:         pgtype.Text{String: csrPEM, Valid: true},
		PrivateKeyPem:  privKeyPEM,
		ComplianceCsid: pgtype.Text{String: complianceToken, Valid: true},
		ProductionCsid: pgtype.Text{},
		CsidExpiry:     pgtype.Timestamptz{Time: expiry, Valid: true},
		ZatcaEnv:       pgtype.Text{String: environment, Valid: true},
		IsActive:       pgtype.Bool{Bool: true, Valid: true},
		Metadata:       metaJSON,
	}

	deviceConfig, err := repo.CreateZatcaDeviceConfig(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to save device config to DB: %w", err)
	}

	return &deviceConfig, nil
}

// csrDeviceSerial keeps the application's hardware identifier separate from
// the structured EGS identifier required inside the certificate request.
func csrDeviceSerial(serialNumber, deviceType string) (string, error) {
	serial := strings.TrimSpace(serialNumber)
	if serial == "" {
		return "", fmt.Errorf("device_serial is required")
	}
	if strings.Contains(serial, "|") || strings.HasPrefix(serial, "1-") {
		parts := strings.Split(serial, "|")
		if len(parts) != 3 {
			return "", fmt.Errorf("device_serial must use 1-Provider|2-Model|3-SerialNumber")
		}
		for i, prefix := range []string{"1-", "2-", "3-"} {
			if !strings.HasPrefix(parts[i], prefix) || strings.TrimSpace(strings.TrimPrefix(parts[i], prefix)) == "" {
				return "", fmt.Errorf("device_serial must use 1-Provider|2-Model|3-SerialNumber with all three values")
			}
		}
		return serial, nil
	}
	model := "POS"
	if deviceType == "cloud" {
		model = "Cloud"
	}
	return fmt.Sprintf("1-Nembus|2-%s|3-%s", model, serial), nil
}

func ccsidBinaryToken(resp *ComplianceCSIDResponse) string {
	if resp.BinarySecurityToken != "" {
		return resp.BinarySecurityToken
	}
	return resp.IssuedCertificate
}

// RenewDeviceCSID executes CSID renewal for an active device configuration near expiration.
func (s *OnboardingService) RenewDeviceCSID(ctx context.Context, configID int32) (*repository.ZatcaDeviceConfig, error) {
	if s.cfg == nil {
		return nil, fmt.Errorf("ZATCA configuration unavailable")
	}
	repo := s.repo
	if tenantRepo, ok := ctx.Value(middleware.RepoKey).(*repository.Queries); ok && tenantRepo != nil {
		repo = tenantRepo
	}
	if repo == nil {
		return nil, fmt.Errorf("repository not set")
	}
	device, err := repo.GetZatcaDeviceConfig(ctx, configID)
	if err != nil {
		return nil, err
	}
	if err := usecase.RequireOrganizationZatca(ctx, repo, device.OrganizationID); err != nil {
		return nil, err
	}
	if !device.IsActive.Bool || device.IsRevoked.Bool {
		return nil, fmt.Errorf("device is inactive or revoked")
	}
	environment, err := usecase.ZatcaDeviceEnvironment(&device)
	if err != nil {
		return nil, err
	}
	client, err := s.clients(environment)
	if err != nil {
		return nil, err
	}
	var credentials struct {
		ProductionSecret string `json:"production_secret"`
		Secret           string `json:"secret"`
	}
	if json.Unmarshal(device.Metadata, &credentials) != nil {
		return nil, fmt.Errorf("invalid device credentials")
	}
	if credentials.ProductionSecret == "" {
		credentials.ProductionSecret = credentials.Secret
	}
	if credentials.ProductionSecret == "" || device.ProductionCsid.String == "" || device.ProductionCsid.String == device.ComplianceCsid.String {
		return nil, fmt.Errorf("reporting credentials are required for renewal")
	}
	response, err := client.RenewProductionCSID(ctx, device.ProductionCsid.String, credentials.ProductionSecret, device.CsrPem.String)
	if err != nil {
		return nil, fmt.Errorf("CSID renewal failed: %w", err)
	}
	if response == nil || response.Secret == "" || response.BinarySecurityToken == "" {
		return nil, fmt.Errorf("renewal did not return a certificate and matching secret")
	}
	if err := usecase.ValidateZatcaDeviceCertificate(response.BinarySecurityToken, device.PrivateKeyPem); err != nil {
		return nil, err
	}
	expiry, err := usecase.ZatcaCertificateExpiry(response.BinarySecurityToken)
	if err != nil {
		return nil, err
	}
	patch, _ := json.Marshal(map[string]any{"production_secret": response.Secret, "secret": response.Secret, "credential_environment": environment, "onboarding_stage": "production"})
	if err := repo.SaveRenewedZatcaCredentials(ctx, device.ID, device.ProductionCsid.String, response.BinarySecurityToken, expiry, patch, device.ZatcaEnv.String); err != nil {
		return nil, err
	}
	updated, err := repo.GetZatcaDeviceConfig(ctx, device.ID)
	return &updated, err
}
