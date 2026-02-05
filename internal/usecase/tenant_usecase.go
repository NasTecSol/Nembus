package usecase

import (
	"context"
	"encoding/json"
	"time"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// TenantOutput is the response shape for tenant APIs. Settings is json.RawMessage
// so JSONB from DB marshals as embedded JSON instead of bytes.
type TenantOutput struct {
	ID         uuid.UUID        `json:"id"`
	TenantName string           `json:"tenant_name"`
	Slug       string           `json:"slug"`
	DbConnStr  string           `json:"db_conn_str"`
	IsActive   pgtype.Bool      `json:"is_active"`
	Settings   json.RawMessage  `json:"settings"`
	CreatedAt  pgtype.Timestamp `json:"created_at"`
	UpdatedAt  pgtype.Timestamp `json:"updated_at"`
}

// tenantToOutput converts repository.Tenant to TenantOutput with Settings as JSON.
func tenantToOutput(t repository.Tenant) TenantOutput {
	return TenantOutput{
		ID:         t.ID,
		TenantName: t.TenantName,
		Slug:       t.Slug,
		DbConnStr:  t.DbConnStr,
		IsActive:   t.IsActive,
		Settings:   utils.BytesToJSONRawMessage(t.Settings),
		CreatedAt:  t.CreatedAt,
		UpdatedAt:  t.UpdatedAt,
	}
}

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
	// --- Ensure is_active is set properly ---
	if !req.IsActive.Valid {
		req.IsActive = pgtype.Bool{
			Bool:  true,
			Valid: true,
		}
	}

	tenant, err := uc.repo.CreateTenant(ctx, req)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	// --- Decode settings ---
	var decodedSettings any
	if len(tenant.Settings) > 0 {
		if err := json.Unmarshal(tenant.Settings, &decodedSettings); err != nil {
			decodedSettings = string(tenant.Settings)
		}
	}

	// --- Convert pgtype / uuid to Go-native ---
	isActive := tenant.IsActive.Bool
	createdAt := tenant.CreatedAt.Time
	updatedAt := tenant.UpdatedAt.Time

	responseTenant := struct {
		ID         string    `json:"id"`
		TenantName string    `json:"tenant_name"`
		Slug       string    `json:"slug"`
		DbConnStr  string    `json:"db_conn_str"`
		IsActive   bool      `json:"is_active"`
		Settings   any       `json:"settings"`
		CreatedAt  time.Time `json:"created_at"`
		UpdatedAt  time.Time `json:"updated_at"`
	}{
		ID:         tenant.ID.String(),
		TenantName: tenant.TenantName,
		Slug:       tenant.Slug,
		DbConnStr:  tenant.DbConnStr,
		IsActive:   isActive,
		Settings:   decodedSettings,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}

	return utils.NewResponse(utils.CodeCreated, "tenant created successfully", responseTenant)
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

	return utils.NewResponse(utils.CodeOK, "tenant fetched successfully", tenantToOutput(tenant))
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

	return utils.NewResponse(utils.CodeOK, "tenant fetched successfully", tenantToOutput(tenant))
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
	out := make([]TenantOutput, len(tenants))
	for i := range tenants {
		out[i] = tenantToOutput(tenants[i])
	}
	return utils.NewResponse(utils.CodeOK, "active tenants fetched successfully", out)
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
	out := make([]TenantOutput, len(tenants))
	for i := range tenants {
		out[i] = tenantToOutput(tenants[i])
	}
	return utils.NewResponse(utils.CodeOK, "tenants fetched successfully", out)
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

	// --- Mark IsActive as valid if user sent a value ---
	req.IsActive.Valid = true

	req.ID = id
	tenant, err := uc.repo.UpdateTenant(ctx, req)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	// Optional: decode settings for response
	var decodedSettings any
	if len(tenant.Settings) > 0 {
		if err := json.Unmarshal(tenant.Settings, &decodedSettings); err != nil {
			decodedSettings = string(tenant.Settings)
		}
	}

	responseTenant := struct {
		ID         string    `json:"id"`
		TenantName string    `json:"tenant_name"`
		Slug       string    `json:"slug"`
		DbConnStr  string    `json:"db_conn_str"`
		IsActive   bool      `json:"is_active"`
		Settings   any       `json:"settings"`
		CreatedAt  time.Time `json:"created_at"`
		UpdatedAt  time.Time `json:"updated_at"`
	}{
		ID:         tenant.ID.String(),
		TenantName: tenant.TenantName,
		Slug:       tenant.Slug,
		DbConnStr:  tenant.DbConnStr,
		IsActive:   tenant.IsActive.Bool,
		Settings:   decodedSettings,
		CreatedAt:  tenant.CreatedAt.Time,
		UpdatedAt:  tenant.UpdatedAt.Time,
	}

	return utils.NewResponse(utils.CodeOK, "tenant updated successfully", responseTenant)
}

func (uc *TenantUseCase) DeactivateTenant(ctx context.Context, slug string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	// 1️⃣ Fetch current tenant by slug (active or inactive)
	currentResp := uc.GetTenantBySlugAny(ctx, slug)
	if currentResp.StatusCode != utils.CodeOK {
		return utils.NewResponse(utils.CodeError, "tenant not found", nil)
	}

	// 2️⃣ Cast to TenantOutput (safe)
	tenantOutput := currentResp.Data.(TenantOutput)

	// 3️⃣ Prepare update params: only deactivate
	update := repository.UpdateTenantParams{
		ID:         tenantOutput.ID,
		TenantName: tenantOutput.TenantName,
		Slug:       tenantOutput.Slug,
		DbConnStr:  tenantOutput.DbConnStr,
		IsActive:   pgtype.Bool{Bool: false, Valid: true}, // deactivate
		Settings:   tenantOutput.Settings,
	}

	// 4️⃣ Call repository to update
	tenant, err := uc.repo.UpdateTenant(ctx, update)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	// 5️⃣ Return updated tenant as TenantOutput
	return utils.NewResponse(utils.CodeOK, "tenant deactivated successfully", tenantToOutput(tenant))
}
