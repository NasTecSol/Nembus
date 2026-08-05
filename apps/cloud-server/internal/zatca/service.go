package zatca

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	repo       *repository.Queries
	client     *ZatcaClient
	usecase    *usecase.ZatcaUseCase
	onboarding *OnboardingService
	cfg        *usecase.ZatcaConfig
}

func NewService(repo *repository.Queries, cfg *usecase.ZatcaConfig) *Service {
	client := NewZatcaClient(cfg.BaseURL)
	uc := usecase.NewZatcaUseCase(cfg)
	uc.SetRepository(repo)
	onboarding := NewOnboardingService(repo, client, cfg)

	return &Service{
		repo:       repo,
		client:     client,
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
	if !s.cfg.Enabled {
		return nil
	}

	// 1. Get active cloud device config
	device, err := s.repo.GetActiveCloudDevice(ctx, org.ID)
	if err != nil {
		return fmt.Errorf("no active ZATCA cloud device found for org %d: %w", org.ID, err)
	}

	// 2. Sign document and record chain entry
	signedDoc, chainEntry, err := s.usecase.SignInvoice(ctx, invoice, lines, org, store, &device)
	if err != nil {
		return fmt.Errorf("failed to sign B2B invoice: %w", err)
	}

	// 3. Call ZATCA Clearance API synchronously
	resp, err := s.client.ClearStandardInvoice(
		ctx,
		device.ProductionCsid.String,
		"secret", // From metadata
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
		_ = s.repo.UpdateChainEntryStatus(ctx, repository.UpdateChainEntryStatusParams{
			ID:            chainEntry.ID,
			ZatcaStatus:   repository.NullZatcaDocStatus{ZatcaDocStatus: status, Valid: true},
			ZatcaResponse: respBytes,
			SubmittedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		})
		return fmt.Errorf("ZATCA clearance failed: %v", err)
	}

	// Clearance successful
	_ = s.repo.UpdateChainEntryStatus(ctx, repository.UpdateChainEntryStatusParams{
		ID:            chainEntry.ID,
		ZatcaStatus:   repository.NullZatcaDocStatus{ZatcaDocStatus: repository.ZatcaDocStatusCleared, Valid: true},
		ZatcaResponse: respBytes,
		SubmittedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		ClearedAt:     pgtype.Timestamptz{Time: now, Valid: true},
	})

	if resp.QrCode != "" {
		_ = s.repo.UpdateChainEntryQRCode(ctx, repository.UpdateChainEntryQRCodeParams{
			ID:           chainEntry.ID,
			QrCodeBase64: pgtype.Text{String: resp.QrCode, Valid: true},
		})
	}

	return nil
}

// StartReportingWorker starts a background worker that periodically polls for pending B2C reporting entries
// and submits them to ZATCA's Reporting API within the required 24-hour window.
func (s *Service) StartReportingWorker(ctx context.Context) {
	if !s.cfg.Enabled {
		return
	}

	ticker := time.NewTicker(2 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.processPendingReports(ctx)
			}
		}
	}()
	log.Println("✅ ZATCA B2C Reporting Worker started")
}

func (s *Service) processPendingReports(ctx context.Context) {
	entries, err := s.repo.ListPendingChainEntries(ctx, 20)
	if err != nil || len(entries) == 0 {
		return
	}

	for _, entry := range entries {
		device, err := s.repo.GetZatcaDeviceConfig(ctx, entry.DeviceConfigID)
		if err != nil {
			continue
		}

		resp, err := s.client.ReportSimplifiedInvoice(
			ctx,
			device.ProductionCsid.String,
			"secret",
			entry.XmlHash,
			entry.ZatcaUuid.String(),
			entry.SignedXml.String,
		)

		now := time.Now()
		respBytes, _ := json.Marshal(resp)

		if err != nil || (resp != nil && resp.ReportingStatus != "REPORTED") {
			_ = s.repo.UpdateChainEntryStatus(ctx, repository.UpdateChainEntryStatusParams{
				ID:            entry.ID,
				ZatcaStatus:   repository.NullZatcaDocStatus{ZatcaDocStatus: repository.ZatcaDocStatusFailed, Valid: true},
				ZatcaResponse: respBytes,
				SubmittedAt:   pgtype.Timestamptz{Time: now, Valid: true},
			})
			continue
		}

		_ = s.repo.UpdateChainEntryStatus(ctx, repository.UpdateChainEntryStatusParams{
			ID:            entry.ID,
			ZatcaStatus:   repository.NullZatcaDocStatus{ZatcaDocStatus: repository.ZatcaDocStatusReported, Valid: true},
			ZatcaResponse: respBytes,
			SubmittedAt:   pgtype.Timestamptz{Time: now, Valid: true},
			ClearedAt:     pgtype.Timestamptz{Time: now, Valid: true},
		})
	}
}
