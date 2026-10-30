package zatca

import (
	"context"
	"errors"
	"testing"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestDisabledOrganizationStopsGatewayCallsWithoutDeletingDeviceOrDocument(t *testing.T) {
	db := &submissionDB{organizationMetadata: []byte(`{"zatca":{"enabled":false}}`), device: repository.ZatcaDeviceConfig{ID: 3}}
	service := NewService(nil, &usecase.ZatcaConfig{Enabled: true})
	calls := 0
	service.clients = func(string) (*ZatcaClient, error) { calls++; return nil, errors.New("must not dispatch") }
	entry := &repository.ZatcaDocumentChain{ID: 1, OrganizationID: 1, DeviceConfigID: 3, SignedXml: pgtype.Text{String: "saved XML", Valid: true}}
	db.chain = *entry
	for _, operation := range []func() error{
		func() error {
			_, err := service.checkDocument(context.Background(), repository.New(db), entry)
			return err
		},
		func() error {
			_, err := service.submitDocument(context.Background(), repository.New(db), entry)
			return err
		},
	} {
		if err := operation(); !errors.Is(err, usecase.ErrZatcaDisabled) {
			t.Fatal("expected organization-disabled error", err)
		}
	}
	if calls != 0 || db.updates != 0 || entry.SignedXml.String != "saved XML" {
		t.Fatal("disabled operation contacted gateway or changed documents")
	}
}

func TestOrganizationDisabledBlocksOnboardingBeforeOTPAndRenewal(t *testing.T) {
	db := &completionDB{org: repository.Organization{ID: 1, Metadata: []byte(`{"zatca":{"enabled":false}}`)}, device: repository.ZatcaDeviceConfig{ID: 3, OrganizationID: 1}}
	repo := repository.New(db)
	ctx := context.WithValue(context.Background(), middleware.RepoKey, repo)
	service := NewOnboardingService(nil, &usecase.ZatcaConfig{Enabled: true})
	calls := 0
	service.clients = func(string) (*ZatcaClient, error) { calls++; return nil, errors.New("must not dispatch") }
	if _, err := service.OnboardDevice(ctx, 1, nil, nil, "device", "pos", "otp", "simulation"); !errors.Is(err, usecase.ErrZatcaDisabled) {
		t.Fatal("disabled organization reached onboarding", err)
	}
	if _, err := service.CompleteDeviceOnboarding(ctx, 3); !errors.Is(err, usecase.ErrZatcaDisabled) {
		t.Fatal("disabled organization completed onboarding", err)
	}
	if _, err := service.RenewDeviceCSID(ctx, 3); !errors.Is(err, usecase.ErrZatcaDisabled) {
		t.Fatal("disabled organization renewed credentials", err)
	}
	if calls != 0 || db.writes != 0 {
		t.Fatal("disabled organization used OTP or changed credentials")
	}
}
