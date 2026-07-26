package handler

import (
	"net/http"
	"strconv"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type CustomerHandler struct {
	useCase *usecase.CustomerUseCase
}

func NewCustomerHandler(uc *usecase.CustomerUseCase) *CustomerHandler {
	return &CustomerHandler{
		useCase: uc,
	}
}

func (h *CustomerHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateCustomer handles POST /customers
// @Summary      Create a new customer
// @Description  Create a customer record
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        customer      body      CreateCustomerRequest  true  "Customer data"
// @Success      201           {object}  CustomerResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/customers [post]
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateCustomerRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	var creditLimit *pgtype.Numeric
	if req.CreditLimit != nil {
		n, err := numericFromString(*req.CreditLimit)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid credit_limit", nil))
			return
		}
		creditLimit = &n
	}

	var outstandingBalance *pgtype.Numeric
	if req.OutstandingBalance != nil {
		n, err := numericFromString(*req.OutstandingBalance)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid outstanding_balance", nil))
			return
		}
		outstandingBalance = &n
	}

	metaBytes, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}

	resp := h.useCase.CreateCustomer(
		c.Request.Context(),
		req.OrganizationID,
		req.CustomerCode,
		req.Name,
		req.Email,
		req.Phone,
		req.Address,
		req.CustomerType,
		req.PriceListID,
		creditLimit,
		outstandingBalance,
		req.IsActive,
		metaBytes,
	)
	c.JSON(resp.StatusCode, resp)
}

// GetCustomerByID handles GET /customers/:id
// @Summary      Get customer by ID
// @Description  Retrieve a customer by ID
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Customer ID"
// @Success      200           {object}  CustomerResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/customers/{id} [get]
func (h *CustomerHandler) GetCustomerByID(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetCustomerByID(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// GetCustomerByCode handles GET /customers/code/:code
// @Summary      Get customer by code
// @Description  Retrieve a customer by organization and code
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true  "Tenant identifier"
// @Param        Authorization    header    string  true  "Bearer token"
// @Param        code             path      string  true  "Customer code"
// @Param        organization_id  query     int     true  "Organization ID"
// @Success      200              {object}  CustomerResponse
// @Failure      400              {object}  ErrorResponse
// @Failure      401              {object}  ErrorResponse
// @Failure      404              {object}  ErrorResponse
// @Router       /api/customers/code/{code} [get]
func (h *CustomerHandler) GetCustomerByCode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil))
		return
	}

	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid organization_id", nil))
		return
	}

	resp := h.useCase.GetCustomerByCode(c.Request.Context(), int32(orgID), c.Param("code"))
	c.JSON(resp.StatusCode, resp)
}

// ListCustomers handles GET /customers
// @Summary      List customers
// @Description  Retrieve all customers for an organization
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true  "Tenant identifier"
// @Param        Authorization    header    string  true  "Bearer token"
// @Param        organization_id  query     int     true  "Organization ID"
// @Success      200              {array}   CustomerResponse
// @Failure      400              {object}  ErrorResponse
// @Failure      401              {object}  ErrorResponse
// @Failure      500              {object}  ErrorResponse
// @Router       /api/customers [get]
func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil))
		return
	}

	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid organization_id", nil))
		return
	}

	resp := h.useCase.ListCustomers(c.Request.Context(), int32(orgID))
	c.JSON(resp.StatusCode, resp)
}

// ListActiveCustomers handles GET /customers/active
// @Summary      List active customers
// @Description  Retrieve active customers for an organization
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true  "Tenant identifier"
// @Param        Authorization    header    string  true  "Bearer token"
// @Param        organization_id  query     int     true  "Organization ID"
// @Success      200              {array}   CustomerResponse
// @Failure      400              {object}  ErrorResponse
// @Failure      401              {object}  ErrorResponse
// @Failure      500              {object}  ErrorResponse
// @Router       /api/customers/active [get]
func (h *CustomerHandler) ListActiveCustomers(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil))
		return
	}

	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid organization_id", nil))
		return
	}

	resp := h.useCase.ListActiveCustomers(c.Request.Context(), int32(orgID))
	c.JSON(resp.StatusCode, resp)
}

// ListCustomersByType handles GET /customers/type/:customer_type
// @Summary      List customers by type
// @Description  Retrieve customers by type for an organization
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true  "Tenant identifier"
// @Param        Authorization    header    string  true  "Bearer token"
// @Param        customer_type    path      string  true  "Customer type"
// @Param        organization_id  query     int     true  "Organization ID"
// @Success      200              {array}   CustomerResponse
// @Failure      400              {object}  ErrorResponse
// @Failure      401              {object}  ErrorResponse
// @Failure      500              {object}  ErrorResponse
// @Router       /api/customers/type/{customer_type} [get]
func (h *CustomerHandler) ListCustomersByType(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil))
		return
	}

	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid organization_id", nil))
		return
	}

	resp := h.useCase.ListCustomersByType(c.Request.Context(), int32(orgID), c.Param("customer_type"))
	c.JSON(resp.StatusCode, resp)
}

// SearchCustomers handles GET /customers/search
// @Summary      Search customers
// @Description  Search customers by name or code
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true  "Tenant identifier"
// @Param        Authorization    header    string  true  "Bearer token"
// @Param        organization_id  query     int     true  "Organization ID"
// @Param        q                query     string  true  "Search query"
// @Param        limit            query     int     false "Limit"
// @Success      200              {array}   CustomerResponse
// @Failure      400              {object}  ErrorResponse
// @Failure      401              {object}  ErrorResponse
// @Failure      500              {object}  ErrorResponse
// @Router       /api/customers/search [get]
func (h *CustomerHandler) SearchCustomers(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil))
		return
	}
	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid organization_id", nil))
		return
	}

	limit := int64(10)
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid limit", nil))
			return
		}
		limit = parsedLimit
	}

	resp := h.useCase.SearchCustomers(c.Request.Context(), int32(orgID), c.Query("q"), int32(limit))
	c.JSON(resp.StatusCode, resp)
}

// GetCustomersWithOutstandingBalance handles GET /customers/outstanding
// @Summary      List customers with outstanding balance
// @Description  Retrieve active customers with outstanding balance
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true  "Tenant identifier"
// @Param        Authorization    header    string  true  "Bearer token"
// @Param        organization_id  query     int     true  "Organization ID"
// @Success      200              {array}   CustomerResponse
// @Failure      400              {object}  ErrorResponse
// @Failure      401              {object}  ErrorResponse
// @Failure      500              {object}  ErrorResponse
// @Router       /api/customers/outstanding [get]
func (h *CustomerHandler) GetCustomersWithOutstandingBalance(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil))
		return
	}

	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid organization_id", nil))
		return
	}

	resp := h.useCase.GetCustomersWithOutstandingBalance(c.Request.Context(), int32(orgID))
	c.JSON(resp.StatusCode, resp)
}

// GetCustomerCreditStatus handles GET /customers/:id/credit-status
// @Summary      Get customer credit status
// @Description  Retrieve credit limit, outstanding balance and available credit
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Customer ID"
// @Success      200           {object}  CustomerCreditStatusResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/customers/{id}/credit-status [get]
func (h *CustomerHandler) GetCustomerCreditStatus(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.GetCustomerCreditStatus(c.Request.Context(), c.Param("id"))
	c.JSON(resp.StatusCode, resp)
}

// UpdateCustomer handles PUT /customers/:id
// @Summary      Update customer
// @Description  Update a customer record
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Customer ID"
// @Param        customer      body      UpdateCustomerRequest  true  "Customer update data"
// @Success      200           {object}  CustomerResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/customers/{id} [put]
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req UpdateCustomerRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	var creditLimit *pgtype.Numeric
	if req.CreditLimit != nil {
		n, err := numericFromString(*req.CreditLimit)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid credit_limit", nil))
			return
		}
		creditLimit = &n
	}

	metaBytes, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}

	resp := h.useCase.UpdateCustomer(
		c.Request.Context(),
		c.Param("id"),
		req.Name,
		req.Email,
		req.Phone,
		req.Address,
		req.CustomerType,
		req.PriceListID,
		creditLimit,
		req.IsActive,
		metaBytes,
	)
	c.JSON(resp.StatusCode, resp)
}

// ToggleCustomerActive handles PATCH /customers/:id/active
// @Summary      Toggle customer active status
// @Description  Update active status for a customer
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Customer ID"
// @Param        payload       body      ToggleCustomerActiveRequest  true  "Active status"
// @Success      200           {object}  CustomerResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/customers/{id}/active [patch]
func (h *CustomerHandler) ToggleCustomerActive(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req ToggleCustomerActiveRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	resp := h.useCase.ToggleCustomerActive(c.Request.Context(), c.Param("id"), req.IsActive)
	c.JSON(resp.StatusCode, resp)
}

// UpdateCustomerBalance handles PATCH /customers/:id/balance
// @Summary      Update customer outstanding balance
// @Description  Add amount to customer outstanding balance
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Customer ID"
// @Param        payload       body      UpdateCustomerBalanceRequest  true  "Balance adjustment"
// @Success      200           {object}  CustomerResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/customers/{id}/balance [patch]
func (h *CustomerHandler) UpdateCustomerBalance(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req UpdateCustomerBalanceRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	amount, err := numericFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid amount", nil))
		return
	}

	resp := h.useCase.UpdateCustomerBalance(c.Request.Context(), c.Param("id"), amount)
	c.JSON(resp.StatusCode, resp)
}

// DeleteCustomer handles DELETE /customers/:id
// @Summary      Delete customer
// @Description  Delete customer by ID
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Customer ID"
// @Success      200           {object}  SuccessResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/customers/{id} [delete]
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	resp := h.useCase.DeleteCustomer(c.Request.Context(), c.Param("id"))
	c.JSON(resp.StatusCode, resp)
}
