package zatca

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	clientsync "github.com/NasTecSol/nembus-client/internal/sync"
	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	gateway "github.com/NasTecSol/nembus-core/zatca"
	"github.com/jackc/pgx/v5"
)

type LocalOnboarder struct {
	service  *gateway.OnboardingService
	cloudURL string
}

func NewOnboarder(cfg *usecase.ZatcaConfig, cloudURL string) *LocalOnboarder {
	if cfg != nil {
		localConfig := *cfg
		localConfig.LocalPOS = true
		cfg = &localConfig
	}
	return &LocalOnboarder{service: gateway.NewOnboardingService(nil, cfg), cloudURL: cloudURL}
}

// After a fresh clone, the normal onboarding action first recovers this serial
// from its cloud backup. It never asks for a new OTP to replace a backed-up key.
func (o *LocalOnboarder) OnboardDevice(ctx context.Context, orgID int32, storeID, terminalID *int32, serial, kind, otp, environment string) (*repository.ZatcaDeviceConfig, error) {
	repo, _ := ctx.Value(middleware.RepoKey).(*repository.Queries)
	if repo == nil {
		return nil, fmt.Errorf("local tenant repository is required")
	}
	if err := usecase.RequireOrganizationZatca(ctx, repo, orgID); err != nil {
		return nil, err
	}
	serial = strings.TrimSpace(serial)
	_, err := repo.GetZatcaDeviceConfigBySerial(ctx, repository.GetZatcaDeviceConfigBySerialParams{OrganizationID: orgID, DeviceSerial: serial})
	if errors.Is(err, pgx.ErrNoRows) && serial != "" && o.cloudURL != "" {
		tenant, ok := ctx.Value(middleware.TenantSlugKey).(string)
		if !ok || tenant == "" {
			return nil, fmt.Errorf("tenant identity is required for ZATCA recovery")
		}
		if err := clientsync.RecoverZatcaDevice(ctx, repo, o.cloudURL, tenant, serial, orgID); err != nil && !errors.Is(err, clientsync.ErrZatcaBackupMissing) {
			return nil, fmt.Errorf("recover this device's cloud registration before onboarding: %w", err)
		}
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	return o.service.OnboardDevice(ctx, orgID, storeID, terminalID, serial, kind, otp, environment)
}

func (o *LocalOnboarder) CompleteDeviceOnboardingResponse(ctx context.Context, id int32) (json.RawMessage, error) {
	return o.service.CompleteDeviceOnboardingResponse(ctx, id)
}
