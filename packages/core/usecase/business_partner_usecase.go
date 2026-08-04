package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// ============================================================
// INPUT / OUTPUT TYPES
// ============================================================

// CreateBusinessPartnerInput is the payload for creating a B2B partner.
type CreateBusinessPartnerInput struct {
	OrganizationID  int32   `json:"organization_id"`
	Code            string  `json:"code"`
	Name            string  `json:"name"`
	PartnerRole     string  `json:"partner_role"` // supplier | vendor | special_customer | corporate_group
	TaxID           *string `json:"tax_id,omitempty"`
	CurrencyCode    *string `json:"currency_code,omitempty"`
	CreditLimit     float64 `json:"credit_limit"`
	PaymentTermsID  *int32  `json:"payment_terms_id,omitempty"`
	SalesRepUserID  *int32  `json:"sales_rep_user_id,omitempty"`
	IsActive        *bool   `json:"is_active,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateBusinessPartnerInput is the payload for updating a B2B partner.
type UpdateBusinessPartnerInput struct {
	ID             int32   `json:"id"`
	Name           *string `json:"name,omitempty"`
	TaxID          *string `json:"tax_id,omitempty"`
	CurrencyCode   *string `json:"currency_code,omitempty"`
	CreditLimit    *float64 `json:"credit_limit,omitempty"`
	PaymentTermsID *int32  `json:"payment_terms_id,omitempty"`
	SalesRepUserID *int32  `json:"sales_rep_user_id,omitempty"`
	IsActive       *bool   `json:"is_active,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// CreatePartnerAddressInput is the payload for adding an address to a business partner.
type CreatePartnerAddressInput struct {
	PartnerID   int32  `json:"partner_id"`
	AddressName string `json:"address_name"`
	AddressType string `json:"address_type"` // bill_to | ship_to | both
	Street      *string `json:"street,omitempty"`
	City        *string `json:"city,omitempty"`
	State       *string `json:"state,omitempty"`
	ZipCode     *string `json:"zip_code,omitempty"`
	CountryCode *string `json:"country_code,omitempty"`
	IsDefault   bool   `json:"is_default"`
}

// CreatePartnerContactInput is the payload for adding a contact to a business partner.
type CreatePartnerContactInput struct {
	PartnerID int32   `json:"partner_id"`
	FirstName string  `json:"first_name"`
	LastName  *string `json:"last_name,omitempty"`
	Email     *string `json:"email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Position  *string `json:"position,omitempty"`
	IsPrimary bool    `json:"is_primary"`
}

// BusinessPartnerOutput is the API response for a business partner.
type BusinessPartnerOutput struct {
	ID                int32            `json:"id"`
	OrganizationID    int32            `json:"organization_id"`
	Code              string           `json:"code"`
	Name              string           `json:"name"`
	PartnerRole       string           `json:"partner_role"`
	TaxID             pgtype.Text      `json:"tax_id"`
	CurrencyCode      pgtype.Text      `json:"currency_code"`
	CreditLimit       pgtype.Numeric   `json:"credit_limit"`
	OutstandingBalance pgtype.Numeric  `json:"outstanding_balance"`
	PaymentTermsID    pgtype.Int4      `json:"payment_terms_id"`
	SalesRepUserID    pgtype.Int4      `json:"sales_rep_user_id"`
	IsActive          bool             `json:"is_active"`
	Metadata          json.RawMessage  `json:"metadata"`
	CreatedAt         pgtype.Timestamp `json:"created_at"`
	UpdatedAt         pgtype.Timestamp `json:"updated_at"`
}

func bpToOutput(bp repository.BusinessPartner) BusinessPartnerOutput {
	return BusinessPartnerOutput{
		ID:                 bp.ID,
		OrganizationID:     bp.OrganizationID,
		Code:               bp.Code,
		Name:               bp.Name,
		PartnerRole:        bp.PartnerRole,
		TaxID:              bp.TaxID,
		CurrencyCode:       bp.CurrencyCode,
		CreditLimit:        bp.CreditLimit,
		OutstandingBalance: bp.OutstandingBalance,
		PaymentTermsID:     bp.PaymentTermsID,
		SalesRepUserID:     bp.SalesRepUserID,
		IsActive:           bp.IsActive.Bool,
		Metadata:           utils.BytesToJSONRawMessage(bp.Metadata),
		CreatedAt:          bp.CreatedAt,
		UpdatedAt:          bp.UpdatedAt,
	}
}

// ============================================================
// USE CASE
// ============================================================

// BusinessPartnerUseCase handles all B2B partner management operations,
// superseding the legacy supplier model.
type BusinessPartnerUseCase struct {
	repo *repository.Queries
}

func NewBusinessPartnerUseCase() *BusinessPartnerUseCase {
	return &BusinessPartnerUseCase{}
}

func (uc *BusinessPartnerUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func (uc *BusinessPartnerUseCase) repoOrErr() *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	return nil
}

// CreateBusinessPartner creates a new B2B business partner (supplier, vendor, corporate, etc.).
func (uc *BusinessPartnerUseCase) CreateBusinessPartner(
	ctx context.Context,
	input CreateBusinessPartnerInput,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if input.OrganizationID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}
	if input.Code == "" {
		return utils.NewResponse(utils.CodeBadReq, "code is required", nil)
	}
	if input.Name == "" {
		return utils.NewResponse(utils.CodeBadReq, "name is required", nil)
	}
	validRoles := map[string]bool{"supplier": true, "vendor": true, "special_customer": true, "corporate_group": true}
	if !validRoles[input.PartnerRole] {
		return utils.NewResponse(utils.CodeBadReq, "partner_role must be one of: supplier, vendor, special_customer, corporate_group", nil)
	}

	metaBytes, _ := json.Marshal(input.Metadata)
	if metaBytes == nil {
		metaBytes = []byte("{}")
	}

	creditLimit := pgtype.Numeric{}
	_ = creditLimit.Scan(fmt.Sprintf("%.2f", input.CreditLimit))

	params := repository.CreateBusinessPartnerParams{
		OrganizationID:     input.OrganizationID,
		Code:               input.Code,
		Name:               input.Name,
		PartnerRole:        input.PartnerRole,
		TaxID:              utils.StringToPgText(input.TaxID),
		CurrencyCode:       utils.StringToPgText(input.CurrencyCode),
		CreditLimit:        creditLimit,
		OutstandingBalance: pgtype.Numeric{},
		PaymentTermsID:     utils.Int32ToPgInt4(input.PaymentTermsID),
		SalesRepUserID:     utils.Int32ToPgInt4(input.SalesRepUserID),
		IsActive:           pgtype.Bool{Bool: true, Valid: true},
		Metadata:           metaBytes,
	}
	if input.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *input.IsActive, Valid: true}
	}

	bp, err := uc.repo.CreateBusinessPartner(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "business partner created", bpToOutput(bp))
}


// GetBusinessPartner retrieves a business partner by ID.
func (uc *BusinessPartnerUseCase) GetBusinessPartner(ctx context.Context, id int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	bp, err := uc.repo.GetBusinessPartner(ctx, id)
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "business partner not found", nil)
	}
	return utils.NewResponse(utils.CodeOK, "business partner retrieved", bpToOutput(bp))
}

// ListBusinessPartners lists business partners for an org, optionally filtered by role.
func (uc *BusinessPartnerUseCase) ListBusinessPartners(
	ctx context.Context,
	organizationID int32,
	partnerRole *string,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	rows, err := uc.repo.ListActiveBusinessPartners(ctx, repository.ListActiveBusinessPartnersParams{
		OrganizationID: organizationID,
		PartnerRole:    utils.StringToPgText(partnerRole),
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]BusinessPartnerOutput, 0, len(rows))
	for _, bp := range rows {
		out = append(out, bpToOutput(bp))
	}
	return utils.NewResponse(utils.CodeOK, "business partners retrieved", out)
}

// UpdateBusinessPartner updates a business partner's fields.
func (uc *BusinessPartnerUseCase) UpdateBusinessPartner(
	ctx context.Context,
	input UpdateBusinessPartnerInput,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	var creditLimitNum pgtype.Numeric
	if input.CreditLimit != nil {
		_ = creditLimitNum.Scan(fmt.Sprintf("%.2f", *input.CreditLimit))
	}

	var metaBytes []byte
	if input.Metadata != nil {
		metaBytes, _ = json.Marshal(input.Metadata)
	}

	params := repository.UpdateBusinessPartnerParams{
		ID:             input.ID,
		Name:           utils.StringToPgText(input.Name),
		TaxID:          utils.StringToPgText(input.TaxID),
		CurrencyCode:   utils.StringToPgText(input.CurrencyCode),
		CreditLimit:    creditLimitNum,
		PaymentTermsID: utils.Int32ToPgInt4(input.PaymentTermsID),
		SalesRepUserID: utils.Int32ToPgInt4(input.SalesRepUserID),
		IsActive:       pgtype.Bool{},
		Metadata:       metaBytes,
	}
	if input.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *input.IsActive, Valid: true}
	}
	bp, err := uc.repo.UpdateBusinessPartner(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "business partner updated", bpToOutput(bp))
}

// ToggleBusinessPartnerActive activates or deactivates a business partner.
func (uc *BusinessPartnerUseCase) ToggleBusinessPartnerActive(
	ctx context.Context,
	id int32,
	isActive bool,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	bp, err := uc.repo.ToggleBusinessPartnerActive(ctx, repository.ToggleBusinessPartnerActiveParams{
		ID:       id,
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "business partner status updated", bpToOutput(bp))
}


// ListForSupplierContext returns active supplier/vendor partners (GRN & PO screens).
func (uc *BusinessPartnerUseCase) ListForSupplierContext(ctx context.Context, organizationID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	rows, err := uc.repo.ListSuppliersAsBusinessPartners(ctx, organizationID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]BusinessPartnerOutput, 0, len(rows))
	for _, bp := range rows {
		out = append(out, bpToOutput(bp))
	}
	return utils.NewResponse(utils.CodeOK, "supplier partners retrieved", out)
}

// AddAddress adds a billing/shipping address to a business partner.
func (uc *BusinessPartnerUseCase) AddAddress(ctx context.Context, input CreatePartnerAddressInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	countryCode := "SA"
	if input.CountryCode != nil {
		countryCode = *input.CountryCode
	}
	addr, err := uc.repo.CreatePartnerAddress(ctx, repository.CreatePartnerAddressParams{
		PartnerID:   input.PartnerID,
		AddressName: input.AddressName,
		AddressType: input.AddressType,
		Street:      utils.StringToPgText(input.Street),
		City:        utils.StringToPgText(input.City),
		State:       utils.StringToPgText(input.State),
		ZipCode:     utils.StringToPgText(input.ZipCode),
		CountryCode: pgtype.Text{String: countryCode, Valid: true},
		IsDefault:   pgtype.Bool{Bool: input.IsDefault, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "partner address added", addr)
}

// ListAddresses lists all addresses for a business partner.
func (uc *BusinessPartnerUseCase) ListAddresses(ctx context.Context, partnerID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	rows, err := uc.repo.ListPartnerAddresses(ctx, partnerID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "partner addresses retrieved", rows)
}

// AddContact adds a contact person to a business partner.
func (uc *BusinessPartnerUseCase) AddContact(ctx context.Context, input CreatePartnerContactInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	if input.FirstName == "" {
		return utils.NewResponse(utils.CodeBadReq, "first_name is required", nil)
	}
	contact, err := uc.repo.CreatePartnerContact(ctx, repository.CreatePartnerContactParams{
		PartnerID: input.PartnerID,
		FirstName: input.FirstName,
		LastName:  utils.StringToPgText(input.LastName),
		Email:     utils.StringToPgText(input.Email),
		Phone:     utils.StringToPgText(input.Phone),
		Position:  utils.StringToPgText(input.Position),
		IsPrimary: pgtype.Bool{Bool: input.IsPrimary, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeCreated, "partner contact added", contact)
}

// ListContacts lists all contacts for a business partner.
func (uc *BusinessPartnerUseCase) ListContacts(ctx context.Context, partnerID int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	rows, err := uc.repo.ListPartnerContacts(ctx, partnerID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	return utils.NewResponse(utils.CodeOK, "partner contacts retrieved", rows)
}

// UpdateOutstandingBalance updates the partner's outstanding AP balance.
func (uc *BusinessPartnerUseCase) UpdateOutstandingBalance(
	ctx context.Context,
	partnerID int32,
	delta float64,
) error {
	if uc.repo == nil {
		return fmt.Errorf("repository not set")
	}
	var deltaNum pgtype.Numeric
	_ = deltaNum.Scan(fmt.Sprintf("%.2f", delta))
	_, err := uc.repo.UpdateBusinessPartnerBalance(ctx, repository.UpdateBusinessPartnerBalanceParams{
		ID:                 partnerID,
		OutstandingBalance: deltaNum,
	})
	return err
}

// SearchBusinessPartners performs a fuzzy name/code search for a business partner.
func (uc *BusinessPartnerUseCase) SearchBusinessPartners(
	ctx context.Context,
	organizationID int32,
	query string,
	partnerRole *string,
	limit int32,
) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}
	rows, err := uc.repo.SearchBusinessPartners(ctx, repository.SearchBusinessPartnersParams{
		OrganizationID: organizationID,
		Name:           "%" + query + "%",
		PartnerRole:    utils.StringToPgText(partnerRole),
		Limit:          limit,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}
	out := make([]BusinessPartnerOutput, 0, len(rows))
	for _, bp := range rows {
		out = append(out, bpToOutput(bp))
	}
	return utils.NewResponse(utils.CodeOK, "search results", out)
}

// helper: format entry number for journal entries
func formatJournalEntryNumber(prefix, ref string) string {
	return fmt.Sprintf("%s-%s-%d", prefix, ref, time.Now().UnixNano()%100000)
}
