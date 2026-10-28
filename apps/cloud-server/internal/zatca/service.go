package zatca

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/middleware/manager"
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

// StartReportingWorker starts a background worker that periodically polls for pending B2C reporting entries
// and submits them to ZATCA's Reporting API within the required 24-hour window.
func (s *Service) StartReportingWorker(ctx context.Context, tenants *manager.Manager) {

	ticker := time.NewTicker(2 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.processPendingReportsForTenants(ctx, tenants)
			}
		}
	}()
	log.Println("✅ ZATCA B2C Reporting Worker started")
}

func (s *Service) processPendingReportsForTenants(ctx context.Context, tenants *manager.Manager) {
	if s.repo == nil || tenants == nil {
		return
	}
	active, err := s.repo.ListActiveTenants(ctx)
	if err != nil {
		log.Printf("ZATCA worker could not list active tenants: %v", err)
		return
	}
	for _, tenant := range active {
		pool, err := tenants.GetPool(ctx, tenant.Slug)
		if err != nil {
			log.Printf("ZATCA worker could not open tenant %s: %v", tenant.Slug, err)
			continue
		}
		s.processPendingReports(ctx, repository.New(pool))
	}
}

func (s *Service) processPendingReports(ctx context.Context, repo *repository.Queries) {
	entries, err := repo.ListEnabledPendingZatcaDocuments(ctx, 20)
	if err != nil || len(entries) == 0 {
		if err != nil {
			log.Printf("ZATCA worker could not list pending reports: %v", err)
		}
		return
	}

	for _, entry := range entries {
		// Standard B2B invoices use clearance; do not send them to reporting.
		if entry.EntityType != usecase.ZatcaSalesOrderEntity && entry.EntityType != "pos_transaction" && entry.EntityType != "sales_return" {
			continue
		}
		// Reuse the same credential validation and status persistence as manual submit.
		if _, err := s.submitDocument(ctx, repo, &entry); err != nil && !errors.Is(err, usecase.ErrZatcaDisabled) {
			log.Printf("ZATCA reporting failed for document %d: %v", entry.ID, err)
		}
	}
}
