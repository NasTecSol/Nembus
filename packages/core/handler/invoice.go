package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/gin-gonic/gin"
)

type InvoiceHandler struct {
	useCase *usecase.InvoiceUseCase
}

func NewInvoiceHandler(uc *usecase.InvoiceUseCase) *InvoiceHandler {
	return &InvoiceHandler{
		useCase: uc,
	}
}

func (h *InvoiceHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "repository not found in context", nil))
		c.Abort()
		return nil
	}
	return repo
}

// CreateInvoice handles POST /api/invoices
// @Summary      Create Invoice
// @Description  Create a new commercial/customer invoice with line items
// @Tags         invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                     true  "Tenant identifier"
// @Param        Authorization header    string                     true  "Bearer token"
// @Param        body          body      usecase.CreateInvoiceInput true  "Invoice creation payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/invoices [post]
func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req usecase.CreateInvoiceInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("invalid request body: %v", err), nil))
		return
	}

	resp := h.useCase.CreateInvoice(c.Request.Context(), req)
	c.JSON(resp.StatusCode, resp)
}

// CreateInvoiceFromOrder handles POST /api/invoices/from-order/:order_id
// @Summary      Create Invoice from Order
// @Description  Automatically generates an invoice from an existing sales order
// @Tags         invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                 true  "Tenant identifier"
// @Param        Authorization header    string                 true  "Bearer token"
// @Param        order_id      path      string                 true  "Sales Order UUID"
// @Param        body          body      object                 true  "Payload with invoice_number"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/invoices/from-order/{order_id} [post]
func (h *InvoiceHandler) CreateInvoiceFromOrder(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID := c.Param("order_id")
	var body struct {
		InvoiceNumber string `json:"invoice_number" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invoice_number is required in body", nil))
		return
	}

	resp := h.useCase.CreateInvoiceFromOrder(c.Request.Context(), orderID, body.InvoiceNumber)
	c.JSON(resp.StatusCode, resp)
}

// GetInvoice handles GET /api/invoices/:id
// @Summary      Get Invoice by ID
// @Description  Retrieve full invoice details with lines and payment history
// @Tags         invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Invoice UUID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/invoices/{id} [get]
func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetInvoice(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// GetInvoiceByNumber handles GET /api/invoices/by-number/:invoice_number
// @Summary      Get Invoice by Number
// @Description  Retrieve invoice by unique invoice number
// @Tags         invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        invoice_number  path      string  true  "Invoice Number"
// @Success      200             {object}  SuccessResponse
// @Failure      400             {object}  ErrorResponse
// @Failure      404             {object}  ErrorResponse
// @Failure      500             {object}  ErrorResponse
// @Router       /api/invoices/by-number/{invoice_number} [get]
func (h *InvoiceHandler) GetInvoiceByNumber(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	invNumber := c.Param("invoice_number")
	resp := h.useCase.GetInvoiceByNumber(c.Request.Context(), invNumber)
	c.JSON(resp.StatusCode, resp)
}

// ListInvoices handles GET /api/invoices
// @Summary      List Invoices
// @Description  Get paginated list of invoices with filters
// @Tags         invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true   "Tenant identifier"
// @Param        Authorization   header    string  true   "Bearer token"
// @Param        organization_id query     int     true   "Organization ID"
// @Param        store_id        query     int     false  "Store ID"
// @Param        customer_id     query     int     false  "Customer ID"
// @Param        status          query     string  false  "Invoice status (draft, sent, viewed, partially_paid, paid, overdue, cancelled, refunded)"
// @Param        from_date       query     string  false  "From date (YYYY-MM-DD)"
// @Param        to_date         query     string  false  "To date (YYYY-MM-DD)"
// @Param        page            query     int     false  "Page number" default(1)
// @Param        limit           query     int     false  "Number of records per page" default(50)
// @Success      200             {object}  SuccessResponse
// @Failure      400             {object}  ErrorResponse
// @Failure      500             {object}  ErrorResponse
// @Router       /api/invoices [get]
func (h *InvoiceHandler) ListInvoices(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil || orgID <= 0 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil))
		return
	}

	filter := usecase.InvoiceFilter{
		OrganizationID: int32(orgID),
	}

	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		if sID, err := strconv.ParseInt(storeIDStr, 10, 32); err == nil && sID > 0 {
			s := int32(sID)
			filter.StoreID = &s
		}
	}

	if custIDStr := c.Query("customer_id"); custIDStr != "" {
		if cID, err := strconv.ParseInt(custIDStr, 10, 32); err == nil && cID > 0 {
			cu := int32(cID)
			filter.CustomerID = &cu
		}
	}

	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	if fromDate := c.Query("from_date"); fromDate != "" {
		filter.FromDate = &fromDate
	}

	if toDate := c.Query("to_date"); toDate != "" {
		filter.ToDate = &toDate
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")
	page, _ := strconv.ParseInt(pageStr, 10, 32)
	limit, _ := strconv.ParseInt(limitStr, 10, 32)
	filter.Page = int32(page)
	filter.Limit = int32(limit)

	resp := h.useCase.ListInvoices(c.Request.Context(), filter)
	c.JSON(resp.StatusCode, resp)
}

// ListCustomerInvoices handles GET /api/invoices/customer/:customer_id
// @Summary      List Customer Invoices
// @Description  Get paginated invoices for a specific customer
// @Tags         invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true   "Tenant identifier"
// @Param        Authorization   header    string  true   "Bearer token"
// @Param        customer_id     path      int     true   "Customer ID"
// @Param        organization_id query     int     true   "Organization ID"
// @Param        page            query     int     false  "Page number" default(1)
// @Param        limit           query     int     false  "Number of records per page" default(50)
// @Success      200             {object}  SuccessResponse
// @Failure      400             {object}  ErrorResponse
// @Failure      500             {object}  ErrorResponse
// @Router       /api/invoices/customer/{customer_id} [get]
func (h *InvoiceHandler) ListCustomerInvoices(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	custIDStr := c.Param("customer_id")
	custID, err := strconv.ParseInt(custIDStr, 10, 32)
	if err != nil || custID <= 0 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid customer_id", nil))
		return
	}

	orgIDStr := c.Query("organization_id")
	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil || orgID <= 0 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil))
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")
	page, _ := strconv.ParseInt(pageStr, 10, 32)
	limit, _ := strconv.ParseInt(limitStr, 10, 32)

	resp := h.useCase.ListCustomerInvoices(c.Request.Context(), int32(custID), int32(orgID), int32(page), int32(limit))
	c.JSON(resp.StatusCode, resp)
}

// ListOverdueInvoices handles GET /api/invoices/overdue
// @Summary      List Overdue Invoices
// @Description  Get paginated overdue invoices for a store
// @Tags         invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        store_id      query     int     true   "Store ID"
// @Param        page          query     int     false  "Page number" default(1)
// @Param        limit         query     int     false  "Number of records per page" default(50)
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/invoices/overdue [get]
func (h *InvoiceHandler) ListOverdueInvoices(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	storeIDStr := c.Query("store_id")
	storeID, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil || storeID <= 0 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "store_id is required", nil))
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")
	page, _ := strconv.ParseInt(pageStr, 10, 32)
	limit, _ := strconv.ParseInt(limitStr, 10, 32)

	resp := h.useCase.ListOverdueInvoices(c.Request.Context(), int32(storeID), int32(page), int32(limit))
	c.JSON(resp.StatusCode, resp)
}

// UpdateInvoice handles PUT /api/invoices/:id
// @Summary      Update Invoice
// @Description  Update details of an existing invoice
// @Tags         invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                     true  "Tenant identifier"
// @Param        Authorization header    string                     true  "Bearer token"
// @Param        id            path      string                     true  "Invoice UUID"
// @Param        body          body      usecase.UpdateInvoiceInput true  "Invoice update payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/invoices/{id} [put]
func (h *InvoiceHandler) UpdateInvoice(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	var req usecase.UpdateInvoiceInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("invalid request body: %v", err), nil))
		return
	}

	resp := h.useCase.UpdateInvoice(c.Request.Context(), id, req)
	c.JSON(resp.StatusCode, resp)
}

// UpdateInvoiceStatus handles PATCH /api/invoices/:id/status
// @Summary      Update Invoice Status
// @Description  Change the status of an invoice (e.g. sent, paid, cancelled) and record audit log
// @Tags         invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Invoice UUID"
// @Param        body          body      object  true  "Status update payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/invoices/{id}/status [patch]
func (h *InvoiceHandler) UpdateInvoiceStatus(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	var body struct {
		Status          string  `json:"status" binding:"required"`
		ChangedByUserID *int32  `json:"changed_by_user_id"`
		Reason          *string `json:"reason"`
		Notes           *string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "status is required in body", nil))
		return
	}

	resp := h.useCase.UpdateInvoiceStatus(c.Request.Context(), id, body.Status, body.ChangedByUserID, body.Reason, body.Notes)
	c.JSON(resp.StatusCode, resp)
}

// DeleteInvoice handles DELETE /api/invoices/:id
// @Summary      Delete Invoice
// @Description  Delete an invoice by ID
// @Tags         invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Invoice UUID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/invoices/{id} [delete]
func (h *InvoiceHandler) DeleteInvoice(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.DeleteInvoice(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// GetInvoiceStats handles GET /api/invoices/stats
// @Summary      Get Invoice Stats
// @Description  Get aggregated invoice metrics (total invoices, paid, overdue, total amounts)
// @Tags         invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true  "Tenant identifier"
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        organization_id query     int     true  "Organization ID"
// @Success      200             {object}  SuccessResponse
// @Failure      400             {object}  ErrorResponse
// @Failure      500             {object}  ErrorResponse
// @Router       /api/invoices/stats [get]
func (h *InvoiceHandler) GetInvoiceStats(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil || orgID <= 0 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil))
		return
	}

	var storeIDPtr *int32
	if sIDStr := c.Query("store_id"); sIDStr != "" {
		if sID, err := strconv.ParseInt(sIDStr, 10, 32); err == nil {
			s32 := int32(sID)
			storeIDPtr = &s32
		}
	}
	var fromDatePtr, toDatePtr *string
	if fd := c.Query("from_date"); fd != "" {
		fromDatePtr = &fd
	}
	if td := c.Query("to_date"); td != "" {
		toDatePtr = &td
	}

	resp := h.useCase.GetInvoiceStats(c.Request.Context(), int32(orgID), storeIDPtr, fromDatePtr, toDatePtr)
	c.JSON(resp.StatusCode, resp)
}

// ListInvoiceLines handles GET /api/invoices/:id/lines
// @Summary      List Invoice Lines
// @Description  Get all item lines for an invoice
// @Tags         invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Invoice UUID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/invoices/{id}/lines [get]
func (h *InvoiceHandler) ListInvoiceLines(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.ListInvoiceLines(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// CreateInvoiceLine handles POST /api/invoices/:id/lines
// @Summary      Add Invoice Line
// @Description  Add a new line item to an existing invoice
// @Tags         invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                         true  "Tenant identifier"
// @Param        Authorization header    string                         true  "Bearer token"
// @Param        id            path      string                         true  "Invoice UUID"
// @Param        body          body      usecase.CreateInvoiceLineInput true  "Line payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/invoices/{id}/lines [post]
func (h *InvoiceHandler) CreateInvoiceLine(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	var req usecase.CreateInvoiceLineInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("invalid request body: %v", err), nil))
		return
	}

	resp := h.useCase.CreateInvoiceLine(c.Request.Context(), id, req)
	c.JSON(resp.StatusCode, resp)
}

// DeleteInvoiceLine handles DELETE /api/invoices/lines/:line_id
// @Summary      Delete Invoice Line
// @Description  Delete an invoice line item by UUID
// @Tags         invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        line_id       path      string  true  "Invoice Line UUID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/invoices/lines/{line_id} [delete]
func (h *InvoiceHandler) DeleteInvoiceLine(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	lineID := c.Param("line_id")
	resp := h.useCase.DeleteInvoiceLine(c.Request.Context(), lineID)
	c.JSON(resp.StatusCode, resp)
}

// CreateInvoicePayment handles POST /api/invoices/:id/payments
// @Summary      Record Invoice Payment
// @Description  Record a payment transaction against an invoice
// @Tags         invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                            true  "Tenant identifier"
// @Param        Authorization header    string                            true  "Bearer token"
// @Param        id            path      string                            true  "Invoice UUID"
// @Param        body          body      usecase.RecordInvoicePaymentInput true  "Payment payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/invoices/{id}/payments [post]
func (h *InvoiceHandler) CreateInvoicePayment(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	var req usecase.RecordInvoicePaymentInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("invalid request body: %v", err), nil))
		return
	}

	resp := h.useCase.CreateInvoicePayment(c.Request.Context(), id, req)
	c.JSON(resp.StatusCode, resp)
}

// ListInvoicePayments handles GET /api/invoices/:id/payments
// @Summary      List Invoice Payments
// @Description  Get all payment transactions for an invoice
// @Tags         invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        id            path      string  true   "Invoice UUID"
// @Param        page          query     int     false  "Page number" default(1)
// @Param        limit         query     int     false  "Number of records per page" default(50)
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/invoices/{id}/payments [get]
func (h *InvoiceHandler) ListInvoicePayments(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.ListInvoicePayments(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}
