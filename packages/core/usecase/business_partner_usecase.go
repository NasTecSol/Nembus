package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type BusinessPartnerOutput struct {
	ID                 int32                       `json:"id"`
	OrganizationID     int32                       `json:"organization_id"`
	Code               string                      `json:"code"`
	Name               string                      `json:"name"`
	PartnerRole        string                      `json:"partner_role"`
	TaxID              pgtype.Text                 `json:"tax_id"`
	CurrencyCode       pgtype.Text                 `json:"currency_code"`
	CreditLimit        pgtype.Numeric              `json:"credit_limit"`
	OutstandingBalance pgtype.Numeric              `json:"outstanding_balance"`
	PaymentTermsID     pgtype.Int4                 `json:"payment_terms_id"`
	SalesRepUserID     pgtype.Int4                 `json:"sales_rep_user_id"`
	IsActive           pgtype.Bool                 `json:"is_active"`
	Metadata           json.RawMessage             `json:"metadata"`
	CreatedAt          pgtype.Timestamp            `json:"created_at"`
	UpdatedAt          pgtype.Timestamp            `json:"updated_at"`
	Addresses          []repository.PartnerAddress `json:"addresses,omitempty"`
	Contacts           []repository.PartnerContact `json:"contacts,omitempty"`
}

func partnerToOutput(bp repository.BusinessPartner, addresses []repository.PartnerAddress, contacts []repository.PartnerContact) BusinessPartnerOutput {
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
		IsActive:           bp.IsActive,
		Metadata:           utils.BytesToJSONRawMessage(bp.Metadata),
		CreatedAt:          bp.CreatedAt,
		UpdatedAt:          bp.UpdatedAt,
		Addresses:          addresses,
		Contacts:           contacts,
	}
}

type PartnerAddressInput struct {
	AddressName string `json:"address_name"`
	AddressType string `json:"address_type"` // bill_to, ship_to, both
	Street      string `json:"street"`
	City        string `json:"city"`
	State       string `json:"state"`
	ZipCode     string `json:"zip_code"`
	CountryCode string `json:"country_code"`
	IsDefault   bool   `json:"is_default"`
}

type PartnerContactInput struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Position  string `json:"position"`
	IsPrimary bool   `json:"is_primary"`
}

type CreateBusinessPartnerInput struct {
	OrganizationID int32                 `json:"organization_id"`
	Code           string                `json:"code"`
	Name           string                `json:"name"`
	PartnerRole    string                `json:"partner_role"`
	TaxID          string                `json:"tax_id"`
	CurrencyCode   string                `json:"currency_code"`
	CreditLimit    float64               `json:"credit_limit"`
	PaymentTermsID *int32                `json:"payment_terms_id"`
	SalesRepUserID *int32                `json:"sales_rep_user_id"`
	IsActive       bool                  `json:"is_active"`
	Metadata       json.RawMessage       `json:"metadata"`
	Addresses      []PartnerAddressInput `json:"addresses"`
	Contacts       []PartnerContactInput `json:"contacts"`
}

type UpdateBusinessPartnerInput struct {
	Code           string          `json:"code"`
	Name           string          `json:"name"`
	PartnerRole    string          `json:"partner_role"`
	TaxID          string          `json:"tax_id"`
	CurrencyCode   string          `json:"currency_code"`
	CreditLimit    float64         `json:"credit_limit"`
	PaymentTermsID *int32          `json:"payment_terms_id"`
	SalesRepUserID *int32          `json:"sales_rep_user_id"`
	IsActive       bool            `json:"is_active"`
	Metadata       json.RawMessage `json:"metadata"`
}

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

func (uc *BusinessPartnerUseCase) CreateBusinessPartner(ctx context.Context, input CreateBusinessPartnerInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	if input.OrganizationID == 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}
	if input.Code == "" {
		return utils.NewResponse(utils.CodeBadReq, "code is required", nil)
	}
	if input.Name == "" {
		return utils.NewResponse(utils.CodeBadReq, "name is required", nil)
	}
	if input.PartnerRole == "" {
		return utils.NewResponse(utils.CodeBadReq, "partner_role is required", nil)
	}

	// Validate role
	roleValid := false
	for _, r := range []string{"supplier", "vendor", "special_customer", "corporate_group"} {
		if input.PartnerRole == r {
			roleValid = true
			break
		}
	}
	if !roleValid {
		return utils.NewResponse(utils.CodeBadReq, "invalid partner_role, must be one of: supplier, vendor, special_customer, corporate_group", nil)
	}

	// Check duplicate code
	_, err := uc.repo.GetBusinessPartnerByCode(ctx, repository.GetBusinessPartnerByCodeParams{
		OrganizationID: input.OrganizationID,
		Code:           input.Code,
	})
	if err == nil {
		return utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("business partner with code '%s' already exists", input.Code), nil)
	}

	var metadataBytes []byte
	if len(input.Metadata) > 0 {
		metadataBytes = input.Metadata
	} else {
		metadataBytes = []byte("{}")
	}

	bp, err := uc.repo.CreateBusinessPartner(ctx, repository.CreateBusinessPartnerParams{
		OrganizationID:     input.OrganizationID,
		Code:               input.Code,
		Name:               input.Name,
		PartnerRole:        input.PartnerRole,
		TaxID:              utils.StringToPgText(&input.TaxID),
		CurrencyCode:       utils.StringToPgText(&input.CurrencyCode),
		CreditLimit:        utils.Float64ToPgNumeric(input.CreditLimit),
		OutstandingBalance: utils.Float64ToPgNumeric(0.0),
		PaymentTermsID:     utils.Int32ToPgInt4(input.PaymentTermsID),
		SalesRepUserID:     utils.Int32ToPgInt4(input.SalesRepUserID),
		IsActive:           pgtype.Bool{Bool: input.IsActive, Valid: true},
		Metadata:           metadataBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	var addresses []repository.PartnerAddress
	for _, addr := range input.Addresses {
		a, err := uc.repo.CreatePartnerAddress(ctx, repository.CreatePartnerAddressParams{
			PartnerID:   bp.ID,
			AddressName: addr.AddressName,
			AddressType: addr.AddressType,
			Street:      utils.StringToPgText(&addr.Street),
			City:        utils.StringToPgText(&addr.City),
			State:       utils.StringToPgText(&addr.State),
			ZipCode:     utils.StringToPgText(&addr.ZipCode),
			CountryCode: utils.StringToPgText(&addr.CountryCode),
			IsDefault:   pgtype.Bool{Bool: addr.IsDefault, Valid: true},
		})
		if err == nil {
			addresses = append(addresses, a)
			if addr.IsDefault {
				_ = uc.repo.ClearDefaultPartnerAddresses(ctx, repository.ClearDefaultPartnerAddressesParams{
					PartnerID: bp.ID,
					ID:        a.ID,
				})
			}
		}
	}

	var contacts []repository.PartnerContact
	for _, ctc := range input.Contacts {
		c, err := uc.repo.CreatePartnerContact(ctx, repository.CreatePartnerContactParams{
			PartnerID: bp.ID,
			FirstName: ctc.FirstName,
			LastName:  utils.StringToPgText(&ctc.LastName),
			Email:     utils.StringToPgText(&ctc.Email),
			Phone:     utils.StringToPgText(&ctc.Phone),
			Position:  utils.StringToPgText(&ctc.Position),
			IsPrimary: pgtype.Bool{Bool: ctc.IsPrimary, Valid: true},
		})
		if err == nil {
			contacts = append(contacts, c)
			if ctc.IsPrimary {
				_ = uc.repo.ClearPrimaryPartnerContacts(ctx, repository.ClearPrimaryPartnerContactsParams{
					PartnerID: bp.ID,
					ID:        c.ID,
				})
			}
		}
	}

	return utils.NewResponse(utils.CodeCreated, "business partner created successfully", partnerToOutput(bp, addresses, contacts))
}

func (uc *BusinessPartnerUseCase) GetBusinessPartnerByID(ctx context.Context, idStr string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid business partner ID", nil)
	}

	bp, err := uc.repo.GetBusinessPartner(ctx, int32(id))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "business partner not found", nil)
	}

	addresses, _ := uc.repo.ListPartnerAddresses(ctx, bp.ID)
	contacts, _ := uc.repo.ListPartnerContacts(ctx, bp.ID)

	return utils.NewResponse(utils.CodeOK, "business partner fetched successfully", partnerToOutput(bp, addresses, contacts))
}

func (uc *BusinessPartnerUseCase) ListBusinessPartners(ctx context.Context, orgIDStr string, roleFilter string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil || orgID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid or missing organization_id", nil)
	}

	var bps []repository.BusinessPartner
	if roleFilter != "" {
		bps, err = uc.repo.ListBusinessPartnersByRole(ctx, repository.ListBusinessPartnersByRoleParams{
			OrganizationID: int32(orgID),
			PartnerRole:    roleFilter,
		})
	} else {
		bps, err = uc.repo.ListBusinessPartners(ctx, int32(orgID))
	}

	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	outputs := make([]BusinessPartnerOutput, len(bps))
	for i, bp := range bps {
		addresses, _ := uc.repo.ListPartnerAddresses(ctx, bp.ID)
		contacts, _ := uc.repo.ListPartnerContacts(ctx, bp.ID)
		outputs[i] = partnerToOutput(bp, addresses, contacts)
	}

	return utils.NewResponse(utils.CodeOK, "business partners fetched successfully", outputs)
}

func (uc *BusinessPartnerUseCase) SearchBusinessPartners(ctx context.Context, orgIDStr string, query string, limit int32) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil || orgID <= 0 {
		return utils.NewResponse(utils.CodeBadReq, "invalid or missing organization_id", nil)
	}

	if limit <= 0 {
		limit = 10
	}

	bps, err := uc.repo.SearchBusinessPartners(ctx, repository.SearchBusinessPartnersParams{
		OrganizationID: int32(orgID),
		Name:           "%" + query + "%",
		Limit:          limit,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	outputs := make([]BusinessPartnerOutput, len(bps))
	for i, bp := range bps {
		addresses, _ := uc.repo.ListPartnerAddresses(ctx, bp.ID)
		contacts, _ := uc.repo.ListPartnerContacts(ctx, bp.ID)
		outputs[i] = partnerToOutput(bp, addresses, contacts)
	}

	return utils.NewResponse(utils.CodeOK, "business partners searched successfully", outputs)
}

func (uc *BusinessPartnerUseCase) UpdateBusinessPartner(ctx context.Context, idStr string, input UpdateBusinessPartnerInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid business partner ID", nil)
	}

	bp, err := uc.repo.GetBusinessPartner(ctx, int32(id))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "business partner not found", nil)
	}

	// Validate role if updated
	if input.PartnerRole != "" {
		roleValid := false
		for _, r := range []string{"supplier", "vendor", "special_customer", "corporate_group"} {
			if input.PartnerRole == r {
				roleValid = true
				break
			}
		}
		if !roleValid {
			return utils.NewResponse(utils.CodeBadReq, "invalid partner_role", nil)
		}
	} else {
		input.PartnerRole = bp.PartnerRole
	}

	if input.Code == "" {
		input.Code = bp.Code
	}
	if input.Name == "" {
		input.Name = bp.Name
	}

	var metadataBytes []byte
	if len(input.Metadata) > 0 {
		metadataBytes = input.Metadata
	} else {
		metadataBytes = bp.Metadata
	}

	updatedBp, err := uc.repo.UpdateBusinessPartner(ctx, repository.UpdateBusinessPartnerParams{
		ID:             bp.ID,
		Code:           input.Code,
		Name:           input.Name,
		PartnerRole:    input.PartnerRole,
		TaxID:          utils.StringToPgText(&input.TaxID),
		CurrencyCode:   utils.StringToPgText(&input.CurrencyCode),
		CreditLimit:    utils.Float64ToPgNumeric(input.CreditLimit),
		PaymentTermsID: utils.Int32ToPgInt4(input.PaymentTermsID),
		SalesRepUserID: utils.Int32ToPgInt4(input.SalesRepUserID),
		IsActive:       pgtype.Bool{Bool: input.IsActive, Valid: true},
		Metadata:       metadataBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	addresses, _ := uc.repo.ListPartnerAddresses(ctx, updatedBp.ID)
	contacts, _ := uc.repo.ListPartnerContacts(ctx, updatedBp.ID)

	return utils.NewResponse(utils.CodeOK, "business partner updated successfully", partnerToOutput(updatedBp, addresses, contacts))
}

func (uc *BusinessPartnerUseCase) DeleteBusinessPartner(ctx context.Context, idStr string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid business partner ID", nil)
	}

	err = uc.repo.DeleteBusinessPartner(ctx, int32(id))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "business partner deleted successfully", nil)
}

func (uc *BusinessPartnerUseCase) ToggleBusinessPartnerActive(ctx context.Context, idStr string, isActive bool) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid business partner ID", nil)
	}

	bp, err := uc.repo.ToggleBusinessPartnerActive(ctx, repository.ToggleBusinessPartnerActiveParams{
		ID:       int32(id),
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "business partner status toggled successfully", partnerToOutput(bp, nil, nil))
}

// Granular Address operations

func (uc *BusinessPartnerUseCase) AddPartnerAddress(ctx context.Context, partnerIDStr string, input PartnerAddressInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	partnerID, err := strconv.ParseInt(partnerIDStr, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid business partner ID", nil)
	}

	a, err := uc.repo.CreatePartnerAddress(ctx, repository.CreatePartnerAddressParams{
		PartnerID:   int32(partnerID),
		AddressName: input.AddressName,
		AddressType: input.AddressType,
		Street:      utils.StringToPgText(&input.Street),
		City:        utils.StringToPgText(&input.City),
		State:       utils.StringToPgText(&input.State),
		ZipCode:     utils.StringToPgText(&input.ZipCode),
		CountryCode: utils.StringToPgText(&input.CountryCode),
		IsDefault:   pgtype.Bool{Bool: input.IsDefault, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	if input.IsDefault {
		_ = uc.repo.ClearDefaultPartnerAddresses(ctx, repository.ClearDefaultPartnerAddressesParams{
			PartnerID: int32(partnerID),
			ID:        a.ID,
		})
	}

	return utils.NewResponse(utils.CodeCreated, "partner address added successfully", a)
}

func (uc *BusinessPartnerUseCase) UpdatePartnerAddress(ctx context.Context, addressIDStr string, input PartnerAddressInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	addrID, err := strconv.ParseInt(addressIDStr, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid address ID", nil)
	}

	a, err := uc.repo.GetPartnerAddress(ctx, int32(addrID))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "address not found", nil)
	}

	if input.AddressName == "" {
		input.AddressName = a.AddressName
	}
	if input.AddressType == "" {
		input.AddressType = a.AddressType
	}

	updated, err := uc.repo.UpdatePartnerAddress(ctx, repository.UpdatePartnerAddressParams{
		ID:          a.ID,
		AddressName: input.AddressName,
		AddressType: input.AddressType,
		Street:      utils.StringToPgText(&input.Street),
		City:        utils.StringToPgText(&input.City),
		State:       utils.StringToPgText(&input.State),
		ZipCode:     utils.StringToPgText(&input.ZipCode),
		CountryCode: utils.StringToPgText(&input.CountryCode),
		IsDefault:   pgtype.Bool{Bool: input.IsDefault, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	if input.IsDefault {
		_ = uc.repo.ClearDefaultPartnerAddresses(ctx, repository.ClearDefaultPartnerAddressesParams{
			PartnerID: a.PartnerID,
			ID:        a.ID,
		})
	}

	return utils.NewResponse(utils.CodeOK, "partner address updated successfully", updated)
}

func (uc *BusinessPartnerUseCase) DeletePartnerAddress(ctx context.Context, addressIDStr string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	addrID, err := strconv.ParseInt(addressIDStr, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid address ID", nil)
	}

	err = uc.repo.DeletePartnerAddress(ctx, int32(addrID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "partner address deleted successfully", nil)
}

// Granular Contact operations

func (uc *BusinessPartnerUseCase) AddPartnerContact(ctx context.Context, partnerIDStr string, input PartnerContactInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	partnerID, err := strconv.ParseInt(partnerIDStr, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid business partner ID", nil)
	}

	c, err := uc.repo.CreatePartnerContact(ctx, repository.CreatePartnerContactParams{
		PartnerID: int32(partnerID),
		FirstName: input.FirstName,
		LastName:  utils.StringToPgText(&input.LastName),
		Email:     utils.StringToPgText(&input.Email),
		Phone:     utils.StringToPgText(&input.Phone),
		Position:  utils.StringToPgText(&input.Position),
		IsPrimary: pgtype.Bool{Bool: input.IsPrimary, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	if input.IsPrimary {
		_ = uc.repo.ClearPrimaryPartnerContacts(ctx, repository.ClearPrimaryPartnerContactsParams{
			PartnerID: int32(partnerID),
			ID:        c.ID,
		})
	}

	return utils.NewResponse(utils.CodeCreated, "partner contact added successfully", c)
}

func (uc *BusinessPartnerUseCase) UpdatePartnerContact(ctx context.Context, contactIDStr string, input PartnerContactInput) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	ctcID, err := strconv.ParseInt(contactIDStr, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid contact ID", nil)
	}

	c, err := uc.repo.GetPartnerContact(ctx, int32(ctcID))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "contact not found", nil)
	}

	if input.FirstName == "" {
		input.FirstName = c.FirstName
	}

	updated, err := uc.repo.UpdatePartnerContact(ctx, repository.UpdatePartnerContactParams{
		ID:        c.ID,
		FirstName: input.FirstName,
		LastName:  utils.StringToPgText(&input.LastName),
		Email:     utils.StringToPgText(&input.Email),
		Phone:     utils.StringToPgText(&input.Phone),
		Position:  utils.StringToPgText(&input.Position),
		IsPrimary: pgtype.Bool{Bool: input.IsPrimary, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	if input.IsPrimary {
		_ = uc.repo.ClearPrimaryPartnerContacts(ctx, repository.ClearPrimaryPartnerContactsParams{
			PartnerID: c.PartnerID,
			ID:        c.ID,
		})
	}

	return utils.NewResponse(utils.CodeOK, "partner contact updated successfully", updated)
}

func (uc *BusinessPartnerUseCase) DeletePartnerContact(ctx context.Context, contactIDStr string) *repository.Response {
	if resp := uc.repoOrErr(); resp != nil {
		return resp
	}

	ctcID, err := strconv.ParseInt(contactIDStr, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid contact ID", nil)
	}

	err = uc.repo.DeletePartnerContact(ctx, int32(ctcID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "partner contact deleted successfully", nil)
}
