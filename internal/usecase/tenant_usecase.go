package usecase

import (
	"context"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type TenantUseCase struct {
	repo *repository.Queries
}

// NewTenantUseCase creates a new tenant use case
// Repository will be injected per request via SetRepository
func NewTenantUseCase() *TenantUseCase {
	return &TenantUseCase{}
}

// SetRepository sets the repository for this request
func (uc *TenantUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// CreateTenant creates a new tenant
func (uc *TenantUseCase) CreateTenant(ctx context.Context, req repository.CreateTenantParams) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	tenant, err := uc.repo.CreateTenant(ctx, req)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "tenant created successfully", tenant)
}

// GetTenantBySlug returns only active tenant
func (uc *TenantUseCase) GetTenantBySlug(ctx context.Context, slug string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	tenant, err := uc.repo.GetTenantBySlug(ctx, slug)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "tenant not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "tenant fetched successfully", tenant)
}

// GetTenantBySlugAny returns tenant regardless of active status
func (uc *TenantUseCase) GetTenantBySlugAny(ctx context.Context, slug string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	tenant, err := uc.repo.GetTenantBySlugAny(ctx, slug)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "tenant not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "tenant fetched successfully", tenant)
}

// ListActiveTenants returns all active tenants
func (uc *TenantUseCase) ListActiveTenants(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	tenants, err := uc.repo.ListActiveTenants(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "active tenants fetched successfully", tenants)
}

// ListAllTenants returns all tenants (admin use)
func (uc *TenantUseCase) ListAllTenants(ctx context.Context) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	tenants, err := uc.repo.ListAllTenants(ctx)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "tenants fetched successfully", tenants)
}

// UpdateTenant updates tenant fields
func (uc *TenantUseCase) UpdateTenant(
	ctx context.Context,
	id uuid.UUID,
	req repository.UpdateTenantParams,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	req.ID = id

	tenant, err := uc.repo.UpdateTenant(ctx, req)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "tenant updated successfully", tenant)
}

// DeactivateTenant is a soft-disable helper
func (uc *TenantUseCase) DeactivateTenant(ctx context.Context, id uuid.UUID) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	tenant, err := uc.repo.UpdateTenant(ctx, repository.UpdateTenantParams{
		ID:       id,
		IsActive: pgtype.Bool{Bool: false, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "tenant deactivated successfully", tenant)
}
