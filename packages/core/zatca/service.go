package zatca

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	repo       *repository.Queries
	clients    environmentClientFactory
	usecase    *usecase.ZatcaUseCase
	onboarding *OnboardingService
	cfg        *usecase.ZatcaConfig
}

func NewService(repo *repository.Queries, cfg *usecase.ZatcaConfig) *Service {

	uc := usecase.NewZatcaUseCase(cfg)
	uc.SetRepository(repo)
	onboarding := NewOnboardingService(repo, cfg)

	return &Service{
		repo:       repo,
		clients:    newEnvironmentClient,
		usecase:    uc,
		onboarding: onboarding,
		cfg:        cfg,
	}
}

func (s *Service) GetUseCase() *usecase.ZatcaUseCase {
	return s.usecase
}

func (s *Service) GetOnboardingService() *OnboardingService {
	return s.onboarding
}

// ClearB2BInvoice performs real-time synchronous clearance for B2B Standard Invoices via ZATCA API.
func (s *Service) ClearB2BInvoice(ctx context.Context, invoice *repository.Invoice, lines []repository.InvoiceLine, org *repository.Organization, store *repository.Store) error {
	enabled, err := usecase.OrganizationZatcaEnabled(org)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	// 1. Get active cloud device config
	repo := s.repo
	if tenantRepo, ok := ctx.Value(middleware.RepoKey).(*repository.Queries); ok && tenantRepo != nil {
		repo = tenantRepo
	}
	selector := usecase.NewZatcaUseCase(s.cfg)
	selector.SetRepository(repo)
	device, err := selector.SelectZatcaDevice(ctx, org.ID, pgtype.Int4{}, invoice.Metadata)
	if err != nil {
		return fmt.Errorf("no active ZATCA cloud device found for org %d: %w", org.ID, err)
	}

	// 2. Sign document and record chain entry
	signedDoc, chainEntry, err := selector.SignInvoice(ctx, invoice, lines, org, store, &device)
	if err != nil {
		return fmt.Errorf("failed to sign B2B invoice: %w", err)
	}

	// 3. Route using the selected device, not a process-wide gateway.
	environment, err := usecase.ZatcaDeviceEnvironment(&device)
	if err != nil {
		return err
	}
	client, err := s.clients(environment)
	if err != nil {
		return err
	}
	var credentials struct {
		Secret string `json:"production_secret"`
	}
	if json.Unmarshal(device.Metadata, &credentials) != nil || credentials.Secret == "" {
		return fmt.Errorf("matching production_secret is required")
	}
	// Call ZATCA Clearance API synchronously
	resp, err := client.ClearStandardInvoice(
		ctx,
		device.ProductionCsid.String,
		credentials.Secret,
		signedDoc.XMLHash,
		signedDoc.ZatcaUUID,
		string(signedDoc.RawXML),
	)

	now := time.Now()
	respBytes, _ := json.Marshal(resp)

	if err != nil || (resp != nil && resp.ClearanceStatus != "CLEARED") {
		// Document failed clearance. Status set to rejected/failed, BUT chain entry is kept intact.
		status := repository.ZatcaDocStatusRejected
		if err != nil {
			status = repository.ZatcaDocStatusFailed
		}
		_ = repo.UpdateChainEntryStatus(ctx, repository.UpdateChainEntryStatusParams{
			ID:            chainEntry.ID,
			ZatcaStatus:   repository.NullZatcaDocStatus{ZatcaDocStatus: status, Valid: true},
			ZatcaResponse: respBytes,
			SubmittedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		})
		return fmt.Errorf("ZATCA clearance failed: %v", err)
	}

	// Clearance successful
	_ = repo.UpdateChainEntryStatus(ctx, repository.UpdateChainEntryStatusParams{
		ID:            chainEntry.ID,
		ZatcaStatus:   repository.NullZatcaDocStatus{ZatcaDocStatus: repository.ZatcaDocStatusCleared, Valid: true},
		ZatcaResponse: respBytes,
		SubmittedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		ClearedAt:     pgtype.Timestamptz{Time: now, Valid: true},
	})

	if resp.QrCode != "" {
		_ = repo.UpdateChainEntryQRCode(ctx, repository.UpdateChainEntryQRCodeParams{
			ID:           chainEntry.ID,
			QrCodeBase64: pgtype.Text{String: resp.QrCode, Valid: true},
		})
	}

	return nil
}
