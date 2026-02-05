package usecase

import (
	"context"
	"encoding/json"
	"strconv"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// OrganizationOutput is the response shape for organization APIs. Metadata is json.RawMessage so JSONB marshals as JSON.
type OrganizationOutput struct {
	ID                int32            `json:"id"`
	Name              string           `json:"name"`
	Code              string           `json:"code"`
	LegalName         pgtype.Text      `json:"legal_name"`
	TaxID             pgtype.Text      `json:"tax_id"`
	CurrencyCode      pgtype.Text      `json:"currency_code"`
	FiscalYearVariant pgtype.Text      `json:"fiscal_year_variant"`
	IsActive          pgtype.Bool      `json:"is_active"`
	Metadata          json.RawMessage  `json:"metadata"`
	CreatedAt         pgtype.Timestamp `json:"created_at"`
	UpdatedAt         pgtype.Timestamp `json:"updated_at"`
}

func orgToOutput(o repository.Organization) OrganizationOutput {
	return OrganizationOutput{
		ID:                o.ID,
		Name:              o.Name,
		Code:              o.Code,
		LegalName:         o.LegalName,
		TaxID:             o.TaxID,
		CurrencyCode:      o.CurrencyCode,
		FiscalYearVariant: o.FiscalYearVariant,
		IsActive:          o.IsActive,
		Metadata:          utils.BytesToJSONRawMessage(o.Metadata),
		CreatedAt:         o.CreatedAt,
		UpdatedAt:         o.UpdatedAt,
	}
}

type OrganizationUseCase struct {
	repo *repository.Queries
}

// NewOrganizationUseCase creates a new use case without a repository
// Repository will be injected per request via SetRepository
func NewOrganizationUseCase() *OrganizationUseCase {
	return &OrganizationUseCase{}
}

// SetRepository sets the repository for this request
func (uc *OrganizationUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

// CreateOrganization creates a new organization
func (uc *OrganizationUseCase) CreateOrganization(ctx context.Context, name, code string, legalName, taxID, currencyCode, fiscalYearVariant *string, isActive bool) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if name == "" {
		return utils.NewResponse(utils.CodeBadReq, "name cannot be empty", nil)
	}
	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "code cannot be empty", nil)
	}

	// Prepare optional fields
	var legalNameText pgtype.Text
	if legalName != nil && *legalName != "" {
		legalNameText = pgtype.Text{String: *legalName, Valid: true}
	}

	var taxIDText pgtype.Text
	if taxID != nil && *taxID != "" {
		taxIDText = pgtype.Text{String: *taxID, Valid: true}
	}

	var currencyCodeText pgtype.Text
	if currencyCode != nil && *currencyCode != "" {
		currencyCodeText = pgtype.Text{String: *currencyCode, Valid: true}
	} else {
		// Default to USD if not provided
		currencyCodeText = pgtype.Text{String: "USD", Valid: true}
	}

	var fiscalYearVariantText pgtype.Text
	if fiscalYearVariant != nil && *fiscalYearVariant != "" {
		fiscalYearVariantText = pgtype.Text{String: *fiscalYearVariant, Valid: true}
	}

	// Call repository
	org, err := uc.repo.CreateOrganization(ctx, repository.CreateOrganizationParams{
		Name:              name,
		Code:              code,
		LegalName:         legalNameText,
		TaxID:             taxIDText,
		CurrencyCode:      currencyCodeText,
		FiscalYearVariant: fiscalYearVariantText,
		IsActive:          pgtype.Bool{Bool: isActive, Valid: true},
		Metadata:          []byte("{}"),
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "organization created successfully", orgToOutput(org))
}

// GetOrganization gets an organization by ID
func (uc *OrganizationUseCase) GetOrganization(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	// Parse ID as int32
	orgID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "invalid organization ID", nil)
	}

	org, err := uc.repo.GetOrganization(ctx, int32(orgID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "organization fetched successfully", orgToOutput(org))
}

// GetOrganizationByCode gets an organization by code
func (uc *OrganizationUseCase) GetOrganizationByCode(ctx context.Context, code string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if code == "" {
		return utils.NewResponse(utils.CodeError, "code cannot be empty", nil)
	}

	// Call repository to get organization by code
	org, err := uc.repo.GetOrganizationByCode(ctx, code)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "organization fetched successfully", orgToOutput(org))
}

// ListOrganizations lists all organizations
func (uc *OrganizationUseCase) ListOrganizations(ctx context.Context, limit, offset int32, isActive *bool) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	// Set default limit if not provided
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var isActiveBool pgtype.Bool
	if isActive != nil {
		isActiveBool = pgtype.Bool{Bool: *isActive, Valid: true}
	} else {
		isActiveBool = pgtype.Bool{Valid: false} // Don't filter by is_active
	}

	// Call repository
	orgs, err := uc.repo.ListOrganizations(ctx, repository.ListOrganizationsParams{
		Limit:    limit,
		Offset:   offset,
		IsActive: isActiveBool,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]OrganizationOutput, len(orgs))
	for i := range orgs {
		out[i] = orgToOutput(orgs[i])
	}
	return utils.NewResponse(utils.CodeOK, "organizations fetched successfully", out)
}

// UpdateOrganization updates an organization
func (uc *OrganizationUseCase) UpdateOrganization(ctx context.Context, id string, name, legalName, taxID, currencyCode, fiscalYearVariant *string, isActive *bool) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	// Parse ID as int32
	orgID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "invalid organization ID", nil)
	}

	// Prepare optional fields
	var nameText pgtype.Text
	if name != nil && *name != "" {
		nameText = pgtype.Text{String: *name, Valid: true}
	}

	var legalNameText pgtype.Text
	if legalName != nil {
		legalNameText = pgtype.Text{String: *legalName, Valid: true}
	}

	var taxIDText pgtype.Text
	if taxID != nil {
		taxIDText = pgtype.Text{String: *taxID, Valid: true}
	}

	var currencyCodeText pgtype.Text
	if currencyCode != nil {
		currencyCodeText = pgtype.Text{String: *currencyCode, Valid: true}
	}

	var fiscalYearVariantText pgtype.Text
	if fiscalYearVariant != nil {
		fiscalYearVariantText = pgtype.Text{String: *fiscalYearVariant, Valid: true}
	}

	var isActiveBool pgtype.Bool
	if isActive != nil {
		isActiveBool = pgtype.Bool{Bool: *isActive, Valid: true}
	}

	// Call repository
	org, err := uc.repo.UpdateOrganization(ctx, repository.UpdateOrganizationParams{
		ID:                int32(orgID),
		Name:              nameText,
		LegalName:         legalNameText,
		TaxID:             taxIDText,
		CurrencyCode:      currencyCodeText,
		FiscalYearVariant: fiscalYearVariantText,
		IsActive:          isActiveBool,
		Metadata:          []byte("{}"), // Default metadata
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "organization updated successfully", orgToOutput(org))
}

// DeleteOrganization deletes an organization
func (uc *OrganizationUseCase) DeleteOrganization(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	// Parse ID as int32
	orgID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeError, "invalid organization ID", nil)
	}

	err = uc.repo.DeleteOrganization(ctx, int32(orgID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "organization deleted successfully", nil)
}
