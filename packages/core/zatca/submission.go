package zatca

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/jackc/pgx/v5/pgtype"
)

// CheckPosTransaction uses the compliance credential pair, never the reporting
// pair. Its response is returned without changing the document's reporting state.
func (s *Service) CheckPosTransaction(ctx context.Context, transactionID int32) (json.RawMessage, error) {
	if s.cfg == nil {
		return nil, fmt.Errorf("ZATCA configuration unavailable")
	}
	repo, ok := ctx.Value(middleware.RepoKey).(*repository.Queries)
	if !ok || repo == nil {
		return nil, fmt.Errorf("tenant repository is required for ZATCA compliance checking")
	}
	entry, err := repo.GetChainEntryByEntity(ctx, repository.GetChainEntryByEntityParams{EntityType: "pos_transaction", EntityID: strconv.Itoa(int(transactionID))})
	if err != nil {
		return nil, fmt.Errorf("checkout document not found: %w", err)
	}
	return s.checkDocument(ctx, repo, &entry)
}

func (s *Service) checkDocument(ctx context.Context, repo *repository.Queries, entry *repository.ZatcaDocumentChain) (json.RawMessage, error) {
	if err := usecase.RequireOrganizationZatca(ctx, repo, entry.OrganizationID); err != nil {
		return nil, err
	}
	device, err := repo.GetZatcaDeviceConfig(ctx, entry.DeviceConfigID)
	if err != nil {
		return nil, fmt.Errorf("ZATCA device not found: %w", err)
	}
	if device.OrganizationID != entry.OrganizationID || !device.IsActive.Bool || device.IsRevoked.Bool {
		return nil, fmt.Errorf("device must be active and belong to the document organization")
	}
	environment, err := usecase.ZatcaDeviceEnvironment(&device)
	if err != nil {
		return nil, err
	}
	client, err := s.clients(environment)
	if err != nil {
		return nil, err
	}
	var metadata struct {
		ComplianceSecret string `json:"compliance_secret"`
		Secret           string `json:"secret"`
	}
	if err := json.Unmarshal(device.Metadata, &metadata); err != nil {
		return nil, fmt.Errorf("device credentials could not be read")
	}
	secret := metadata.ComplianceSecret
	// Older onboarding saved the compliance secret under this legacy key.
	if secret == "" && (device.ProductionCsid.String == "" || device.ProductionCsid.String == device.ComplianceCsid.String) {
		secret = metadata.Secret
	}
	if secret == "" || device.ComplianceCsid.String == "" || strings.HasPrefix(device.ComplianceCsid.String, "SIMULATED_") {
		return nil, fmt.Errorf("a compliance CSID and matching compliance secret are required")
	}
	if entry.SignedXml.String == "" || entry.XmlHash == "" {
		return nil, fmt.Errorf("checkout document XML or hash is missing")
	}
	response, err := client.SubmitComplianceInvoice(ctx, device.ComplianceCsid.String, secret, entry.XmlHash, entry.ZatcaUuid.String(), entry.SignedXml.String)
	if response == nil {
		return nil, err
	}
	return response.RawResponse, err
}

// SubmitPosTransaction reuses the checkout document's UUID, counter and XML on retries.
func (s *Service) SubmitPosTransaction(ctx context.Context, transactionID int32) (*repository.ZatcaDocumentChain, error) {
	if s.cfg == nil {
		return nil, fmt.Errorf("ZATCA configuration unavailable")
	}
	repo, ok := ctx.Value(middleware.RepoKey).(*repository.Queries)
	if !ok || repo == nil {
		return nil, fmt.Errorf("tenant repository is required for ZATCA submission")
	}
	entry, err := repo.GetChainEntryByEntity(ctx, repository.GetChainEntryByEntityParams{
		EntityType: "pos_transaction", EntityID: strconv.Itoa(int(transactionID)),
	})
	if err != nil {
		return nil, fmt.Errorf("no ZATCA document for transaction %d; create a checkout with ZATCA enabled: %w", transactionID, err)
	}
	return s.submitDocument(ctx, repo, &entry)
}

func (s *Service) submitDocument(ctx context.Context, repo *repository.Queries, entry *repository.ZatcaDocumentChain) (*repository.ZatcaDocumentChain, error) {
	var result *repository.ZatcaDocumentChain
	var submissionErr error
	err := repo.InZatcaTransaction(ctx, func(txRepo *repository.Queries) error {
		if err := txRepo.LockZatcaChainEntry(ctx, entry.ID); err != nil {
			return err
		}
		current, err := txRepo.GetChainEntryByEntity(ctx, repository.GetChainEntryByEntityParams{EntityType: entry.EntityType, EntityID: entry.EntityID})
		if err != nil {
			return err
		}
		if current.ID != entry.ID {
			return fmt.Errorf("ZATCA chain entry changed during submission")
		}
		// Commit the saved gateway result even when ZATCA rejects the invoice.
		result, submissionErr = s.submitDocumentLocked(ctx, txRepo, &current)
		return nil
	})
	if err != nil {
		return result, fmt.Errorf("could not serialize ZATCA reporting: %w", err)
	}
	return result, submissionErr
}

func (s *Service) submitDocumentLocked(ctx context.Context, repo *repository.Queries, entry *repository.ZatcaDocumentChain) (*repository.ZatcaDocumentChain, error) {
	if err := usecase.RequireOrganizationZatca(ctx, repo, entry.OrganizationID); err != nil {
		return nil, err
	}
	if entry.ZatcaStatus.Valid && entry.ZatcaStatus.ZatcaDocStatus == repository.ZatcaDocStatusReported {
		return entry, nil
	}
	device, err := repo.GetZatcaDeviceConfig(ctx, entry.DeviceConfigID)
	if err != nil {
		return nil, fmt.Errorf("ZATCA device not found: %w", err)
	}
	if device.OrganizationID != entry.OrganizationID || !device.IsActive.Bool || device.IsRevoked.Bool {
		return nil, fmt.Errorf("ZATCA device is inactive, revoked, or belongs to a different organization")
	}
	environment, err := usecase.ZatcaDeviceEnvironment(&device)
	if err != nil {
		return nil, err
	}
	client, err := s.clients(environment)
	if err != nil {
		return nil, err
	}

	if device.CsidExpiry.Valid && !device.CsidExpiry.Time.After(time.Now()) {
		return nil, fmt.Errorf("ZATCA device certificate has expired")
	}
	var credentials struct {
		Secret           string `json:"secret"`
		ProductionSecret string `json:"production_secret"`
		OnboardingStage  string `json:"onboarding_stage"`
	}
	if err := json.Unmarshal(device.Metadata, &credentials); err != nil {
		return nil, fmt.Errorf("ZATCA device credentials could not be read")
	}
	if credentials.OnboardingStage == "compliance" || strings.TrimSpace(device.ProductionCsid.String) == "" || (device.ComplianceCsid.String != "" && device.ProductionCsid.String == device.ComplianceCsid.String) {
		return nil, fmt.Errorf("device has only a compliance CSID; complete compliance checks and obtain a production CSID with its matching secret before reporting")
	}
	if credentials.ProductionSecret != "" {
		credentials.Secret = credentials.ProductionSecret
	}
	if strings.TrimSpace(credentials.Secret) == "" {
		return nil, fmt.Errorf("ZATCA device certificate or secret is missing")
	}
	if strings.HasPrefix(device.ProductionCsid.String, "SIMULATED_") {
		return nil, fmt.Errorf("simulated device credentials cannot be submitted to ZATCA")
	}
	if strings.TrimSpace(entry.SignedXml.String) == "" || entry.XmlHash == "" {
		return nil, fmt.Errorf("checkout document XML or hash is missing")
	}
	response, submitErr := client.ReportSimplifiedInvoice(ctx, device.ProductionCsid.String,
		credentials.Secret, entry.XmlHash, entry.ZatcaUuid.String(), entry.SignedXml.String)
	status := repository.ZatcaDocStatusFailed
	var responseJSON json.RawMessage
	if response != nil {
		responseJSON = response.RawResponse
		if response.HTTPStatus >= 400 && response.HTTPStatus < 500 && response.HTTPStatus != 429 && response.HTTPStatus != 401 && response.HTTPStatus != 403 {
			status = repository.ZatcaDocStatusRejected
		}
		if submitErr == nil && response.ReportingStatus == "REPORTED" {
			status = repository.ZatcaDocStatusReported
		}
	}
	if len(responseJSON) == 0 {
		responseJSON, _ = json.Marshal(map[string]string{"error": fmt.Sprint(submitErr)})
	}
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	if err := repo.UpdateChainEntryStatus(ctx, repository.UpdateChainEntryStatusParams{
		ID: entry.ID, ZatcaStatus: repository.NullZatcaDocStatus{ZatcaDocStatus: status, Valid: true},
		ZatcaResponse: responseJSON, SubmittedAt: now,
	}); err != nil {
		return nil, fmt.Errorf("ZATCA submission attempted but saving the result failed: %w", err)
	}
	entry.ZatcaStatus = repository.NullZatcaDocStatus{ZatcaDocStatus: status, Valid: true}
	entry.ZatcaResponse = responseJSON
	entry.SubmittedAt = now
	if submitErr != nil {
		return entry, submitErr
	}
	if status != repository.ZatcaDocStatusReported {
		return entry, fmt.Errorf("ZATCA did not report the invoice; inspect data.zatca_response for validation details")
	}
	return entry, nil
}
