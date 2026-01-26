package usecase

import (
	"context"
	"errors"
	"strconv"

	"NEMBUS/internal/repository"

	"github.com/jackc/pgx/v5/pgtype"
)

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
func (uc *OrganizationUseCase) CreateOrganization(ctx context.Context, name, code string, legalName, taxID, currencyCode, fiscalYearVariant *string, isActive bool) (repository.Organization, error) {
	if uc.repo == nil {
		return repository.Organization{}, errors.New("repository not set")
	}
	if name == "" {
		return repository.Organization{}, errors.New("name cannot be empty")
	}
	if code == "" {
		return repository.Organization{}, errors.New("code cannot be empty")
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

	return uc.repo.CreateOrganization(ctx, repository.CreateOrganizationParams{
		Name:              name,
		Code:              code,
		LegalName:         legalNameText,
		TaxID:             taxIDText,
		CurrencyCode:      currencyCodeText,
		FiscalYearVariant: fiscalYearVariantText,
		IsActive:          pgtype.Bool{Bool: isActive, Valid: true},
		Metadata:          []byte("{}"),
	})
}

// GetOrganization gets an organization by ID
func (uc *OrganizationUseCase) GetOrganization(ctx context.Context, id string) (repository.Organization, error) {
	if uc.repo == nil {
		return repository.Organization{}, errors.New("repository not set")
	}

	// Parse ID as int32
	orgID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return repository.Organization{}, errors.New("invalid organization ID")
	}

	return uc.repo.GetOrganization(ctx, int32(orgID))
}

// GetOrganizationByCode gets an organization by code
func (uc *OrganizationUseCase) GetOrganizationByCode(ctx context.Context, code string) (repository.Organization, error) {
	if uc.repo == nil {
		return repository.Organization{}, errors.New("repository not set")
	}
	if code == "" {
		return repository.Organization{}, errors.New("code cannot be empty")
	}

	return uc.repo.GetOrganizationByCode(ctx, code)
}

// ListOrganizations lists all organizations
func (uc *OrganizationUseCase) ListOrganizations(ctx context.Context, limit, offset int32, isActive *bool) ([]repository.Organization, error) {
	if uc.repo == nil {
		return nil, errors.New("repository not set")
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

	return uc.repo.ListOrganizations(ctx, repository.ListOrganizationsParams{
		Limit:    limit,
		Offset:   offset,
		IsActive: isActiveBool,
	})
}

// UpdateOrganization updates an organization
func (uc *OrganizationUseCase) UpdateOrganization(ctx context.Context, id string, name, legalName, taxID, currencyCode, fiscalYearVariant *string, isActive *bool) (repository.Organization, error) {
	if uc.repo == nil {
		return repository.Organization{}, errors.New("repository not set")
	}

	// Parse ID as int32
	orgID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return repository.Organization{}, errors.New("invalid organization ID")
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

	return uc.repo.UpdateOrganization(ctx, repository.UpdateOrganizationParams{
		Name:              nameText,
		LegalName:         legalNameText,
		TaxID:             taxIDText,
		CurrencyCode:      currencyCodeText,
		FiscalYearVariant: fiscalYearVariantText,
		IsActive:          isActiveBool,
		Metadata:          []byte("{}"), // Keep existing metadata or set default
		ID:                int32(orgID),
	})
}

// DeleteOrganization deletes an organization
func (uc *OrganizationUseCase) DeleteOrganization(ctx context.Context, id string) error {
	if uc.repo == nil {
		return errors.New("repository not set")
	}

	// Parse ID as int32
	orgID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return errors.New("invalid organization ID")
	}

	return uc.repo.DeleteOrganization(ctx, int32(orgID))
}
