package usecase

import (
	"context"
	"encoding/json"
	"strconv"

	"NEMBUS/internal/repository"
	"NEMBUS/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

// CustomerOutput is the response shape for customer APIs
type CustomerOutput struct {
	ID                 int32            `json:"id"`
	OrganizationID     int32            `json:"organization_id"`
	CustomerCode       string           `json:"customer_code"`
	Name               string           `json:"name"`
	Email              pgtype.Text      `json:"email"`
	Phone              pgtype.Text      `json:"phone"`
	Address            pgtype.Text      `json:"address"`
	CustomerType       pgtype.Text      `json:"customer_type"`
	PriceListID        pgtype.Int4      `json:"price_list_id"`
	CreditLimit        pgtype.Numeric   `json:"credit_limit"`
	OutstandingBalance pgtype.Numeric   `json:"outstanding_balance"`
	LoyaltyPoints      pgtype.Numeric   `json:"loyalty_points"`
	IsActive           pgtype.Bool      `json:"is_active"`
	Metadata           json.RawMessage  `json:"metadata"`
	CreatedAt          pgtype.Timestamp `json:"created_at"`
	UpdatedAt          pgtype.Timestamp `json:"updated_at"`
}

func customerToOutput(c repository.Customer) CustomerOutput {
	return CustomerOutput{
		ID:                 c.ID,
		OrganizationID:     c.OrganizationID,
		CustomerCode:       c.CustomerCode,
		Name:               c.Name,
		Email:              c.Email,
		Phone:              c.Phone,
		Address:            c.Address,
		CustomerType:       c.CustomerType,
		PriceListID:        c.PriceListID,
		CreditLimit:        c.CreditLimit,
		OutstandingBalance: c.OutstandingBalance,
		LoyaltyPoints:      c.LoyaltyPoints,
		IsActive:           c.IsActive,
		Metadata:           utils.BytesToJSONRawMessage(c.Metadata),
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
	}
}

type CustomerUseCase struct {
	repo *repository.Queries
}

func NewCustomerUseCase() *CustomerUseCase {
	return &CustomerUseCase{}
}

func (uc *CustomerUseCase) SetRepository(repo *repository.Queries) {
	uc.repo = repo
}

func zeroNumeric() pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan("0")
	return n
}

func (uc *CustomerUseCase) CreateCustomer(
	ctx context.Context,
	organizationID int32,
	customerCode string,
	name string,
	email *string,
	phone *string,
	address *string,
	customerType *string,
	priceListID *int32,
	creditLimit *pgtype.Numeric,
	outstandingBalance *pgtype.Numeric,
	isActive *bool,
	metadata []byte,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if organizationID == 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}
	if customerCode == "" {
		return utils.NewResponse(utils.CodeBadReq, "customer_code is required", nil)
	}
	if name == "" {
		return utils.NewResponse(utils.CodeBadReq, "name is required", nil)
	}

	_, err := uc.repo.GetCustomerByCode(ctx, repository.GetCustomerByCodeParams{
		OrganizationID: organizationID,
		CustomerCode:   customerCode,
	})
	if err == nil {
		return utils.NewResponse(utils.CodeBadReq, "customer code already exists", nil)
	}

	emailVal := pgtype.Text{Valid: false}
	if email != nil && *email != "" {
		emailVal = pgtype.Text{String: *email, Valid: true}
	}

	phoneVal := pgtype.Text{Valid: false}
	if phone != nil && *phone != "" {
		phoneVal = pgtype.Text{String: *phone, Valid: true}
	}

	addressVal := pgtype.Text{Valid: false}
	if address != nil && *address != "" {
		addressVal = pgtype.Text{String: *address, Valid: true}
	}

	customerTypeVal := pgtype.Text{String: "regular", Valid: true}
	if customerType != nil && *customerType != "" {
		customerTypeVal = pgtype.Text{String: *customerType, Valid: true}
	}

	priceListVal := pgtype.Int4{Valid: false}
	if priceListID != nil {
		priceListVal = pgtype.Int4{Int32: *priceListID, Valid: true}
	}

	creditLimitVal := zeroNumeric()
	if creditLimit != nil {
		creditLimitVal = *creditLimit
	}

	outstandingBalanceVal := zeroNumeric()
	if outstandingBalance != nil {
		outstandingBalanceVal = *outstandingBalance
	}

	isActiveVal := pgtype.Bool{Bool: true, Valid: true}
	if isActive != nil {
		isActiveVal = pgtype.Bool{Bool: *isActive, Valid: true}
	}

	metaBytes := []byte("{}")
	if metadata != nil {
		metaBytes = metadata
	}

	customer, err := uc.repo.CreateCustomer(ctx, repository.CreateCustomerParams{
		OrganizationID:     organizationID,
		CustomerCode:       customerCode,
		Name:               name,
		Email:              emailVal,
		Phone:              phoneVal,
		Address:            addressVal,
		CustomerType:       customerTypeVal,
		PriceListID:        priceListVal,
		CreditLimit:        creditLimitVal,
		OutstandingBalance: outstandingBalanceVal,
		IsActive:           isActiveVal,
		Metadata:           metaBytes,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeCreated, "customer created successfully", customerToOutput(customer))
}

func (uc *CustomerUseCase) GetCustomerByID(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	customerID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid customer id", nil)
	}

	customer, err := uc.repo.GetCustomer(ctx, int32(customerID))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "customer not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "customer fetched successfully", customerToOutput(customer))
}

func (uc *CustomerUseCase) GetCustomerByCode(ctx context.Context, organizationID int32, code string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if organizationID == 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}
	if code == "" {
		return utils.NewResponse(utils.CodeBadReq, "customer code is required", nil)
	}

	customer, err := uc.repo.GetCustomerByCode(ctx, repository.GetCustomerByCodeParams{
		OrganizationID: organizationID,
		CustomerCode:   code,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "customer not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "customer fetched successfully", customerToOutput(customer))
}

func (uc *CustomerUseCase) ListCustomers(ctx context.Context, organizationID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if organizationID == 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}

	customers, err := uc.repo.ListCustomers(ctx, organizationID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]CustomerOutput, len(customers))
	for i := range customers {
		out[i] = customerToOutput(customers[i])
	}

	return utils.NewResponse(utils.CodeOK, "customers fetched successfully", out)
}

func (uc *CustomerUseCase) ListActiveCustomers(ctx context.Context, organizationID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if organizationID == 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}

	customers, err := uc.repo.ListActiveCustomers(ctx, organizationID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]CustomerOutput, len(customers))
	for i := range customers {
		out[i] = customerToOutput(customers[i])
	}

	return utils.NewResponse(utils.CodeOK, "active customers fetched successfully", out)
}

func (uc *CustomerUseCase) ListCustomersByType(ctx context.Context, organizationID int32, customerType string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if organizationID == 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}
	if customerType == "" {
		return utils.NewResponse(utils.CodeBadReq, "customer_type is required", nil)
	}

	customers, err := uc.repo.ListCustomersByType(ctx, repository.ListCustomersByTypeParams{
		OrganizationID: organizationID,
		CustomerType:   pgtype.Text{String: customerType, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]CustomerOutput, len(customers))
	for i := range customers {
		out[i] = customerToOutput(customers[i])
	}

	return utils.NewResponse(utils.CodeOK, "customers fetched successfully", out)
}

func (uc *CustomerUseCase) SearchCustomers(ctx context.Context, organizationID int32, q string, limit int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if organizationID == 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}
	if q == "" {
		return utils.NewResponse(utils.CodeBadReq, "search query is required", nil)
	}
	if limit <= 0 {
		limit = 10
	}

	customers, err := uc.repo.SearchCustomers(ctx, repository.SearchCustomersParams{
		OrganizationID: organizationID,
		Name:           "%" + q + "%",
		Limit:          limit,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]CustomerOutput, len(customers))
	for i := range customers {
		out[i] = customerToOutput(customers[i])
	}

	return utils.NewResponse(utils.CodeOK, "customers fetched successfully", out)
}

func (uc *CustomerUseCase) UpdateCustomer(
	ctx context.Context,
	id string,
	name *string,
	email *string,
	phone *string,
	address *string,
	customerType *string,
	priceListID *int32,
	creditLimit *pgtype.Numeric,
	isActive *bool,
	metadata []byte,
) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	customerID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid customer id", nil)
	}

	current, err := uc.repo.GetCustomer(ctx, int32(customerID))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "customer not found", nil)
	}

	params := repository.UpdateCustomerParams{
		ID: int32(customerID),
	}

	if name != nil && *name != "" {
		params.Name = *name
	} else {
		params.Name = current.Name
	}

	if email != nil {
		if *email == "" {
			params.Email = pgtype.Text{Valid: false}
		} else {
			params.Email = pgtype.Text{String: *email, Valid: true}
		}
	} else {
		params.Email = current.Email
	}

	if phone != nil {
		if *phone == "" {
			params.Phone = pgtype.Text{Valid: false}
		} else {
			params.Phone = pgtype.Text{String: *phone, Valid: true}
		}
	} else {
		params.Phone = current.Phone
	}

	if address != nil {
		if *address == "" {
			params.Address = pgtype.Text{Valid: false}
		} else {
			params.Address = pgtype.Text{String: *address, Valid: true}
		}
	} else {
		params.Address = current.Address
	}

	if customerType != nil && *customerType != "" {
		params.CustomerType = pgtype.Text{String: *customerType, Valid: true}
	} else {
		params.CustomerType = current.CustomerType
	}

	if priceListID != nil {
		params.PriceListID = pgtype.Int4{Int32: *priceListID, Valid: true}
	} else {
		params.PriceListID = current.PriceListID
	}

	if creditLimit != nil {
		params.CreditLimit = *creditLimit
	} else {
		params.CreditLimit = current.CreditLimit
	}

	if isActive != nil {
		params.IsActive = pgtype.Bool{Bool: *isActive, Valid: true}
	} else {
		params.IsActive = current.IsActive
	}

	if metadata != nil {
		params.Metadata = metadata
	} else {
		params.Metadata = current.Metadata
	}

	customer, err := uc.repo.UpdateCustomer(ctx, params)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "customer updated successfully", customerToOutput(customer))
}

func (uc *CustomerUseCase) ToggleCustomerActive(ctx context.Context, id string, isActive bool) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	customerID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid customer id", nil)
	}

	customer, err := uc.repo.ToggleCustomerActive(ctx, repository.ToggleCustomerActiveParams{
		ID:       int32(customerID),
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "customer status updated successfully", customerToOutput(customer))
}

func (uc *CustomerUseCase) UpdateCustomerBalance(ctx context.Context, id string, amount pgtype.Numeric) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	customerID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid customer id", nil)
	}

	customer, err := uc.repo.UpdateCustomerBalance(ctx, repository.UpdateCustomerBalanceParams{
		ID:                 int32(customerID),
		OutstandingBalance: amount,
	})
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "customer balance updated successfully", customerToOutput(customer))
}

func (uc *CustomerUseCase) DeleteCustomer(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	customerID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid customer id", nil)
	}

	err = uc.repo.DeleteCustomer(ctx, int32(customerID))
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	return utils.NewResponse(utils.CodeOK, "customer deleted successfully", nil)
}

func (uc *CustomerUseCase) GetCustomersWithOutstandingBalance(ctx context.Context, organizationID int32) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}
	if organizationID == 0 {
		return utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil)
	}

	customers, err := uc.repo.GetCustomersWithOutstandingBalance(ctx, organizationID)
	if err != nil {
		return utils.NewResponse(utils.CodeError, err.Error(), nil)
	}

	out := make([]CustomerOutput, len(customers))
	for i := range customers {
		out[i] = customerToOutput(customers[i])
	}

	return utils.NewResponse(utils.CodeOK, "customers fetched successfully", out)
}

func (uc *CustomerUseCase) GetCustomerCreditStatus(ctx context.Context, id string) *repository.Response {
	if uc.repo == nil {
		return utils.NewResponse(utils.CodeError, "repository not set", nil)
	}

	customerID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		return utils.NewResponse(utils.CodeBadReq, "invalid customer id", nil)
	}

	data, err := uc.repo.GetCustomerCreditStatus(ctx, int32(customerID))
	if err != nil {
		return utils.NewResponse(utils.CodeNotFound, "customer not found", nil)
	}

	return utils.NewResponse(utils.CodeOK, "customer credit status fetched successfully", data)
}
