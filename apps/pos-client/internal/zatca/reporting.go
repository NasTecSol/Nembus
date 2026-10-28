package zatca

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Reporter submits the locally signed sales order using the device's saved
// environment and production credentials. It never depends on a cloud worker.
type Reporter struct {
	cfg        *usecase.ZatcaConfig
	httpClient *http.Client
}

func NewReporter(cfg *usecase.ZatcaConfig) *Reporter {
	return &Reporter{cfg: cfg, httpClient: &http.Client{Timeout: 30 * time.Second}}
}

func (r *Reporter) SubmitSalesOrder(ctx context.Context, id uuid.UUID) (*repository.ZatcaDocumentChain, error) {
	repo, ok := ctx.Value(middleware.RepoKey).(*repository.Queries)
	if !ok || repo == nil {
		return nil, fmt.Errorf("local tenant repository is required for ZATCA reporting")
	}
	uc := usecase.NewZatcaUseCase(r.cfg)
	uc.SetRepository(repo)
	entry, err := uc.EnsureSalesOrderDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	var result *repository.ZatcaDocumentChain
	var reportErr error
	err = repo.InZatcaTransaction(ctx, func(txRepo *repository.Queries) error {
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
		// Commit the gateway response even when the gateway rejects the invoice.
		result, reportErr = r.reportLocked(ctx, txRepo, &current)
		return nil
	})
	if err != nil {
		return result, fmt.Errorf("could not serialize ZATCA reporting: %w", err)
	}
	return result, reportErr
}

func (r *Reporter) reportLocked(ctx context.Context, repo *repository.Queries, entry *repository.ZatcaDocumentChain) (*repository.ZatcaDocumentChain, error) {
	if err := usecase.RequireOrganizationZatca(ctx, repo, entry.OrganizationID); err != nil {
		return nil, err
	}
	if entry.ZatcaStatus.Valid && entry.ZatcaStatus.ZatcaDocStatus == repository.ZatcaDocStatusReported {
		return entry, nil
	}
	device, err := repo.GetZatcaDeviceConfig(ctx, entry.DeviceConfigID)
	if err != nil {
		return nil, fmt.Errorf("local ZATCA device not found: %w", err)
	}
	if device.OrganizationID != entry.OrganizationID || !device.IsActive.Bool || device.IsRevoked.Bool {
		return nil, fmt.Errorf("ZATCA device is inactive, revoked, or belongs to another organization")
	}
	environment, err := usecase.ZatcaDeviceEnvironment(&device)
	if err != nil {
		return nil, err
	}
	baseURL, err := usecase.ZatcaEnvironmentBaseURL(environment)
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
		return nil, fmt.Errorf("local ZATCA device credentials could not be read")
	}
	if credentials.OnboardingStage == "compliance" || strings.TrimSpace(device.ProductionCsid.String) == "" || (device.ComplianceCsid.String != "" && device.ProductionCsid.String == device.ComplianceCsid.String) {
		return nil, fmt.Errorf("device has only a compliance CSID; complete onboarding and sync the production CSID to POS")
	}
	if credentials.ProductionSecret != "" {
		credentials.Secret = credentials.ProductionSecret
	}
	if strings.TrimSpace(credentials.Secret) == "" {
		return nil, fmt.Errorf("matching production secret is missing from the local ZATCA device")
	}
	if strings.HasPrefix(device.ProductionCsid.String, "SIMULATED_") {
		return nil, fmt.Errorf("simulated device credentials cannot be submitted to ZATCA")
	}
	if strings.TrimSpace(entry.SignedXml.String) == "" || entry.XmlHash == "" {
		return nil, fmt.Errorf("sales order signed XML or hash is missing")
	}
	response, httpStatus, reportingStatus, reportErr := r.report(ctx, baseURL, device.ProductionCsid.String, credentials.Secret, entry)
	status := repository.ZatcaDocStatusFailed
	if httpStatus >= 400 && httpStatus < 500 && httpStatus != 401 && httpStatus != 403 && httpStatus != 429 {
		status = repository.ZatcaDocStatusRejected
	}
	if reportErr == nil && reportingStatus == "REPORTED" {
		status = repository.ZatcaDocStatusReported
	}
	if len(response) == 0 {
		response, _ = json.Marshal(map[string]string{"error": fmt.Sprint(reportErr)})
	}
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	if err := repo.UpdateChainEntryStatus(ctx, repository.UpdateChainEntryStatusParams{
		ID: entry.ID, ZatcaStatus: repository.NullZatcaDocStatus{ZatcaDocStatus: status, Valid: true},
		ZatcaResponse: response, SubmittedAt: now,
	}); err != nil {
		return nil, fmt.Errorf("ZATCA submission attempted but saving the local result failed: %w", err)
	}
	entry.ZatcaStatus = repository.NullZatcaDocStatus{ZatcaDocStatus: status, Valid: true}
	entry.ZatcaResponse = response
	entry.SubmittedAt = now
	if reportErr != nil {
		return entry, reportErr
	}
	if status != repository.ZatcaDocStatusReported {
		return entry, fmt.Errorf("ZATCA did not report the invoice; inspect zatca_response")
	}
	return entry, nil
}

func (r *Reporter) report(ctx context.Context, baseURL, csid, secret string, entry *repository.ZatcaDocumentChain) (json.RawMessage, int, string, error) {
	body, err := json.Marshal(map[string]string{
		"invoiceHash": entry.XmlHash,
		"uuid":        entry.ZatcaUuid.String(),
		"invoice":     base64.StdEncoding.EncodeToString([]byte(entry.SignedXml.String)),
	})
	if err != nil {
		return nil, 0, "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/invoices/reporting/single", bytes.NewReader(body))
	if err != nil {
		return nil, 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en")
	req.Header.Set("Accept-Version", "V2")
	req.Header.Set("Clearance-Status", "0")
	req.SetBasicAuth(csid, secret)
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, 0, "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, "", fmt.Errorf("failed to read reporting response: %w", err)
	}
	if !json.Valid(raw) {
		raw, _ = json.Marshal(map[string]any{"http_status": resp.StatusCode, "body": string(raw)})
	}
	var parsed struct {
		ReportingStatus string `json:"reportingStatus"`
	}
	_ = json.Unmarshal(raw, &parsed)
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return raw, resp.StatusCode, parsed.ReportingStatus, fmt.Errorf("ZATCA authentication failed (HTTP %d): check local production CSID, secret and environment", resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return raw, resp.StatusCode, parsed.ReportingStatus, fmt.Errorf("ZATCA reporting API error (HTTP %d)", resp.StatusCode)
	}
	if parsed.ReportingStatus == "" {
		return raw, resp.StatusCode, "", fmt.Errorf("ZATCA reporting response has no reportingStatus")
	}
	return raw, resp.StatusCode, parsed.ReportingStatus, nil
}
