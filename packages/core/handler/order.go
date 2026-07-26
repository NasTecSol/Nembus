package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrderHandler struct {
	useCase *usecase.OrderUseCase
}

func NewOrderHandler(uc *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{useCase: uc}
}

func (h *OrderHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// GetOrder handles GET /api/orders/:id
// @Summary      Get order by ID
// @Description  Retrieve a specific order and its lines by order ID
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Order ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	resp := h.useCase.GetOrder(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// ListOrders handles GET /api/orders
// @Summary      List orders
// @Description  List orders for an organization with optional status filter
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true  "Tenant identifier"
// @Param        Authorization    header    string  true  "Bearer token"
// @Param        organization_id  query     int     true  "Organization ID"
// @Param        status           query     string  false "Order status filter"
// @Success      200              {object}  SuccessResponse
// @Failure      400              {object}  ErrorResponse
// @Failure      401              {object}  ErrorResponse
// @Failure      500              {object}  ErrorResponse
// @Router       /api/orders [get]
func (h *OrderHandler) ListOrders(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid organization_id", nil))
		return
	}

	status := c.Query("status")

	resp := h.useCase.ListOrders(c.Request.Context(), int32(orgID), status)
	c.JSON(resp.StatusCode, resp)
}

// GetOrderByNumber handles GET /api/orders/by-number/:order_number
// @Summary      Get order by order number
// @Description  Retrieve a specific order and its lines by order number
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        order_number   path      string  true  "Order number"
// @Success      200            {object}  SuccessResponse
// @Failure      400            {object}  ErrorResponse
// @Failure      401            {object}  ErrorResponse
// @Failure      404            {object}  ErrorResponse
// @Failure      500            {object}  ErrorResponse
// @Router       /api/orders/by-number/{order_number} [get]
func (h *OrderHandler) GetOrderByNumber(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderNumber := c.Param("order_number")
	if orderNumber == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "order_number is required", nil))
		return
	}

	resp := h.useCase.GetOrderByNumber(c.Request.Context(), orderNumber)
	c.JSON(resp.StatusCode, resp)
}

// CreateOrder handles POST /api/orders
// @Summary      Create sales order
// @Description  Create a new sales order (v2 schema)
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                     true  "Tenant identifier"
// @Param        Authorization header    string                     true  "Bearer token"
// @Param        body          body      CreateSalesOrderV2Request  true  "Order payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateSalesOrderV2Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	shipAddrBytes, err := bytesFromMap(req.ShippingAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid shipping_address", nil))
		return
	}
	billAddrBytes, err := bytesFromMap(req.BillingAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid billing_address", nil))
		return
	}
	metaBytes, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}

	orderDate, err := timestampPtrFromRFC3339(req.OrderDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order_date", nil))
		return
	}
	expectedDate, err := datePtrFromYMD(req.ExpectedDeliveryDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid expected_delivery_date", nil))
		return
	}
	paymentDueDate, err := datePtrFromYMD(req.PaymentDueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid payment_due_date", nil))
		return
	}

	sourceCartID, err := uuidPtrFromString(req.SourceCartID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid source_cart_id", nil))
		return
	}

	arg := repository.CreateSalesOrderV2Params{
		OrderNumber:          req.OrderNumber,
		OrganizationID:       req.OrganizationID,
		StoreID:              int4Ptr(req.StoreID),
		CustomerID:           int4Ptr(req.CustomerID),
		CustomerName:         textPtr(req.CustomerName),
		CustomerEmail:        textPtr(req.CustomerEmail),
		CustomerPhone:        textPtr(req.CustomerPhone),
		OrderType:            repository.OrderType(req.OrderType),
		OrderStatus:          repository.OrderStatusV2(req.OrderStatus),
		PaymentStatus:        repository.PaymentStatus(req.PaymentStatus),
		FulfillmentStatus:    repository.FulfillmentStatus(req.FulfillmentStatus),
		SalesChannel:         textPtr(req.SalesChannel),
		OrderSource:          textPtr(req.OrderSource),
		ReferralSource:       textPtr(req.ReferralSource),
		SourceCartID:         sourceCartID,
		CreatedByUserID:      int4Ptr(req.CreatedByUserID),
		AssignedToUserID:     int4Ptr(req.AssignedToUserID),
		OrderDate:            orderDate,
		ExpectedDeliveryDate: expectedDate,
		ShippingAddress:      json.RawMessage(shipAddrBytes),
		BillingAddress:       json.RawMessage(billAddrBytes),
		ShippingMethod:       textPtr(req.ShippingMethod),
		PaymentMethod:        textPtr(req.PaymentMethod),
		PaymentGateway:       textPtr(req.PaymentGateway),
		PaymentTerms:         textPtr(req.PaymentTerms),
		PaymentDueDate:       paymentDueDate,
		PosTerminalID:        int4Ptr(req.PosTerminalID),
		CashierID:            int4Ptr(req.CashierID),
		IsGift:               boolPtr(req.IsGift),
		GiftMessage:          textPtr(req.GiftMessage),
		SpecialInstructions:  textPtr(req.SpecialInstructions),
		InternalNotes:        textPtr(req.InternalNotes),
		Tags:                 req.Tags,
		Priority:             textPtr(req.Priority),
		Metadata:             metaBytes,
	}

	resp := h.useCase.CreateOrder(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// UpdateOrder handles PUT /api/orders/:id
// @Summary      Update sales order
// @Description  Update order customer, delivery, and metadata fields
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                      true  "Tenant identifier"
// @Param        Authorization header    string                      true  "Bearer token"
// @Param        id            path      string                      true  "Order ID (UUID)"
// @Param        body          body      UpdateSalesOrderV2Request   true  "Order update payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id} [put]
func (h *OrderHandler) UpdateOrder(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	var req UpdateSalesOrderV2Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	shipAddrBytes, err := bytesFromMap(req.ShippingAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid shipping_address", nil))
		return
	}
	billAddrBytes, err := bytesFromMap(req.BillingAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid billing_address", nil))
		return
	}
	metaBytes, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}

	expectedDate, err := datePtrFromYMD(req.ExpectedDeliveryDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid expected_delivery_date", nil))
		return
	}

	arg := repository.UpdateSalesOrderV2Params{
		ID:                   orderID,
		CustomerID:           int4Ptr(req.CustomerID),
		CustomerName:         textPtr(req.CustomerName),
		CustomerEmail:        textPtr(req.CustomerEmail),
		CustomerPhone:        textPtr(req.CustomerPhone),
		ExpectedDeliveryDate: expectedDate,
		ShippingAddress:      json.RawMessage(shipAddrBytes),
		BillingAddress:       json.RawMessage(billAddrBytes),
		ShippingMethod:       textPtr(req.ShippingMethod),
		PaymentMethod:        textPtr(req.PaymentMethod),
		PaymentGateway:       textPtr(req.PaymentGateway),
		SpecialInstructions:  textPtr(req.SpecialInstructions),
		InternalNotes:        textPtr(req.InternalNotes),
		Tags:                 req.Tags,
		Priority:             textPtr(req.Priority),
		Metadata:             metaBytes,
	}

	resp := h.useCase.UpdateOrder(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// UpdateOrderStatus handles PUT /api/orders/:id/status
// @Summary      Update order status
// @Description  Update the high-level order status
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                   true  "Tenant identifier"
// @Param        Authorization header    string                   true  "Bearer token"
// @Param        id            path      string                   true  "Order ID (UUID)"
// @Param        body          body      UpdateOrderStatusRequest true  "Order status payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id}/status [put]
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	arg := repository.UpdateOrderStatusParams{
		ID:          orderID,
		OrderStatus: repository.OrderStatusV2(req.OrderStatus),
	}

	resp := h.useCase.UpdateOrderStatus(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// UpdateOrderPaymentStatus handles PUT /api/orders/:id/payment-status
// @Summary      Update order payment status
// @Description  Update payment status and paid amount of an order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                          true  "Tenant identifier"
// @Param        Authorization header    string                          true  "Bearer token"
// @Param        id            path      string                          true  "Order ID (UUID)"
// @Param        body          body      UpdateOrderPaymentStatusRequest true  "Payment status payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id}/payment-status [put]
func (h *OrderHandler) UpdateOrderPaymentStatus(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	var req UpdateOrderPaymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	paidAmount, err := numericFromString(req.PaidAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid paid_amount", nil))
		return
	}

	arg := repository.UpdateOrderPaymentStatusParams{
		ID:             orderID,
		PaymentStatus:  repository.PaymentStatus(req.PaymentStatus),
		PaidAmount:     paidAmount,
		PaymentMethod:  textPtr(req.PaymentMethod),
		PaymentGateway: textPtr(req.PaymentGateway),
	}

	resp := h.useCase.UpdateOrderPaymentStatus(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// UpdateOrderFulfillmentStatus handles PUT /api/orders/:id/fulfillment-status
// @Summary      Update order fulfillment status
// @Description  Update the fulfillment status of an order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                             true  "Tenant identifier"
// @Param        Authorization header    string                             true  "Bearer token"
// @Param        id            path      string                             true  "Order ID (UUID)"
// @Param        body          body      UpdateOrderFulfillmentStatusRequest true  "Fulfillment status payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id}/fulfillment-status [put]
func (h *OrderHandler) UpdateOrderFulfillmentStatus(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	var req UpdateOrderFulfillmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	arg := repository.UpdateOrderFulfillmentStatusParams{
		ID:                orderID,
		FulfillmentStatus: repository.FulfillmentStatus(req.FulfillmentStatus),
	}

	resp := h.useCase.UpdateOrderFulfillmentStatus(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// UpdateOrderTotals handles PUT /api/orders/:id/totals
// @Summary      Update order totals
// @Description  Update monetary totals of an order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                     true  "Tenant identifier"
// @Param        Authorization header    string                     true  "Bearer token"
// @Param        id            path      string                     true  "Order ID (UUID)"
// @Param        body          body      UpdateOrderTotalsRequest   true  "Totals payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id}/totals [put]
func (h *OrderHandler) UpdateOrderTotals(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	var req UpdateOrderTotalsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	subtotal, err := numericFromString(req.Subtotal)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid subtotal", nil))
		return
	}
	discount, err := numericFromString(req.DiscountAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid discount_amount", nil))
		return
	}
	tax, err := numericFromString(req.TaxAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid tax_amount", nil))
		return
	}
	shipping, err := numericFromString(req.ShippingAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid shipping_amount", nil))
		return
	}
	adjustment, err := numericFromString(req.AdjustmentAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid adjustment_amount", nil))
		return
	}
	total, err := numericFromString(req.TotalAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid total_amount", nil))
		return
	}

	arg := repository.UpdateOrderTotalsParams{
		ID:               orderID,
		Subtotal:         subtotal,
		DiscountAmount:   discount,
		TaxAmount:        tax,
		ShippingAmount:   shipping,
		AdjustmentAmount: adjustment,
		TotalAmount:      total,
	}

	resp := h.useCase.UpdateOrderTotals(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// UpdateOrderDelivery handles PUT /api/orders/:id/delivery
// @Summary      Update order delivery info
// @Description  Update delivery tracking information of an order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                       true  "Tenant identifier"
// @Param        Authorization header    string                       true  "Bearer token"
// @Param        id            path      string                       true  "Order ID (UUID)"
// @Param        body          body      UpdateOrderDeliveryRequest   true  "Delivery payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id}/delivery [put]
func (h *OrderHandler) UpdateOrderDelivery(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	var req UpdateOrderDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	actualDeliveryDate, err := datePtrFromYMD(req.ActualDeliveryDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid actual_delivery_date", nil))
		return
	}

	arg := repository.UpdateOrderDeliveryParams{
		ID:                 orderID,
		ShippingCarrier:    textPtr(req.ShippingCarrier),
		TrackingNumber:     textPtr(req.TrackingNumber),
		TrackingUrl:        textPtr(req.TrackingUrl),
		ActualDeliveryDate: actualDeliveryDate,
	}

	resp := h.useCase.UpdateOrderDelivery(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// AssignOrder handles PUT /api/orders/:id/assign
// @Summary      Assign order to user
// @Description  Assign an order to a specific user
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                 true  "Tenant identifier"
// @Param        Authorization header    string                 true  "Bearer token"
// @Param        id            path      string                 true  "Order ID (UUID)"
// @Param        body          body      AssignOrderRequest     true  "Assignment payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id}/assign [put]
func (h *OrderHandler) AssignOrder(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	var req AssignOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	arg := repository.AssignOrderParams{
		ID:               orderID,
		AssignedToUserID: int4Ptr(&req.AssignedToUserID),
	}

	resp := h.useCase.AssignOrder(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// CancelOrder handles POST /api/orders/:id/cancel
// @Summary      Cancel order
// @Description  Cancel an order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Order ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id}/cancel [post]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	resp := h.useCase.CancelOrder(c.Request.Context(), orderID)
	c.JSON(resp.StatusCode, resp)
}

// DeleteOrder handles DELETE /api/orders/:id
// @Summary      Delete order
// @Description  Permanently delete an order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Order ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id} [delete]
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	resp := h.useCase.DeleteOrder(c.Request.Context(), orderID)
	c.JSON(resp.StatusCode, resp)
}

// CreateOrderLine handles POST /api/orders/:id/lines
// @Summary      Create order line
// @Description  Create a new sales order line
// @Tags         order-lines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                        true  "Tenant identifier"
// @Param        Authorization header    string                        true  "Bearer token"
// @Param        id            path      string                        true  "Order ID (UUID)"
// @Param        body          body      CreateSalesOrderLineV2Request true  "Order line payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id}/lines [post]
func (h *OrderHandler) CreateOrderLine(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	var req CreateSalesOrderLineV2Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	qtyOrdered, err := numericFromString(req.QuantityOrdered)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid quantity_ordered", nil))
		return
	}
	unitPrice, err := numericFromString(req.UnitPrice)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid unit_price", nil))
		return
	}
	lineTotal, err := numericFromString(req.LineTotal)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid line_total", nil))
		return
	}

	discountStr := "0"
	if req.DiscountAmount != nil {
		discountStr = *req.DiscountAmount
	}
	discount, err := numericFromString(discountStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid discount_amount", nil))
		return
	}

	discountPctStr := "0"
	if req.DiscountPercentage != nil {
		discountPctStr = *req.DiscountPercentage
	}
	discountPct, err := numericFromString(discountPctStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid discount_percentage", nil))
		return
	}

	taxAmountStr := "0"
	if req.TaxAmount != nil {
		taxAmountStr = *req.TaxAmount
	}
	taxAmount, err := numericFromString(taxAmountStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid tax_amount", nil))
		return
	}

	taxRateStr := "0"
	if req.TaxRate != nil {
		taxRateStr = *req.TaxRate
	}
	taxRate, err := numericFromString(taxRateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid tax_rate", nil))
		return
	}

	unitCostStr := "0"
	if req.UnitCost != nil {
		unitCostStr = *req.UnitCost
	}
	unitCost, err := numericFromString(unitCostStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid unit_cost", nil))
		return
	}

	expiryDate, err := datePtrFromYMD(req.ExpiryDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid expiry_date", nil))
		return
	}

	customBytes, err := bytesFromMap(req.Customization)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid customization_details", nil))
		return
	}
	metaBytes, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}

	arg := repository.CreateSalesOrderLineV2Params{
		SalesOrderID:         orderID,
		OrganizationID:       req.OrganizationID,
		LineNumber:           req.LineNumber,
		ProductID:            req.ProductID,
		ProductVariantID:     int4Ptr(req.ProductVariantID),
		ProductName:          req.ProductName,
		ProductSku:           textPtr(req.ProductSku),
		QuantityOrdered:      qtyOrdered,
		UomID:                int4Ptr(req.UomID),
		UnitPrice:            unitPrice,
		DiscountAmount:       discount,
		DiscountPercentage:   discountPct,
		TaxAmount:            taxAmount,
		LineTotal:            lineTotal,
		TaxCategoryID:        int4Ptr(req.TaxCategoryID),
		TaxRate:              taxRate,
		BatchNumber:          textPtr(req.BatchNumber),
		SerialNumbers:        req.SerialNumbers,
		ExpiryDate:           expiryDate,
		LineStatus:           textPtr(req.LineStatus),
		CustomizationDetails: customBytes,
		UnitCost:             unitCost,
		Notes:                textPtr(req.Notes),
		Metadata:             metaBytes,
	}

	resp := h.useCase.CreateOrderLine(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// GetOrderLine handles GET /api/order-lines/:line_id
// @Summary      Get order line
// @Description  Retrieve a single order line by its ID
// @Tags         order-lines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        line_id       path      string  true  "Order line ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/order-lines/{line_id} [get]
func (h *OrderHandler) GetOrderLine(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	lineID, err := uuid.Parse(c.Param("line_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid line_id", nil))
		return
	}

	resp := h.useCase.GetOrderLine(c.Request.Context(), lineID)
	c.JSON(resp.StatusCode, resp)
}

// ListOrderLines handles GET /api/orders/:id/lines
// @Summary      List order lines
// @Description  List all lines for a given order
// @Tags         order-lines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Order ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id}/lines [get]
func (h *OrderHandler) ListOrderLines(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	resp := h.useCase.ListOrderLines(c.Request.Context(), orderID)
	c.JSON(resp.StatusCode, resp)
}

// UpdateOrderLine handles PUT /api/order-lines/:line_id
// @Summary      Update order line
// @Description  Update pricing and metadata of an order line
// @Tags         order-lines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                        true  "Tenant identifier"
// @Param        Authorization header    string                        true  "Bearer token"
// @Param        line_id       path      string                        true  "Order line ID (UUID)"
// @Param        body          body      UpdateSalesOrderLineV2Request true  "Order line update payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/order-lines/{line_id} [put]
func (h *OrderHandler) UpdateOrderLine(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	lineID, err := uuid.Parse(c.Param("line_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid line_id", nil))
		return
	}

	var req UpdateSalesOrderLineV2Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	qtyOrdered, err := numericFromString(req.QuantityOrdered)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid quantity_ordered", nil))
		return
	}
	unitPrice, err := numericFromString(req.UnitPrice)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid unit_price", nil))
		return
	}
	discount, err := numericFromString(req.DiscountAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid discount_amount", nil))
		return
	}
	discountPct, err := numericFromString(req.DiscountPercentage)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid discount_percentage", nil))
		return
	}
	taxAmount, err := numericFromString(req.TaxAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid tax_amount", nil))
		return
	}
	lineTotal, err := numericFromString(req.LineTotal)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid line_total", nil))
		return
	}
	metaBytes, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}

	arg := repository.UpdateSalesOrderLineV2Params{
		ID:                 lineID,
		QuantityOrdered:    qtyOrdered,
		UnitPrice:          unitPrice,
		DiscountAmount:     discount,
		DiscountPercentage: discountPct,
		TaxAmount:          taxAmount,
		LineTotal:          lineTotal,
		Notes:              textPtr(req.Notes),
		Metadata:           metaBytes,
	}

	resp := h.useCase.UpdateOrderLine(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// UpdateOrderLineFulfillment handles PATCH /api/order-lines/:line_id/fulfillment
// @Summary      Update order line fulfillment quantity
// @Description  Update the quantity fulfilled for an order line
// @Tags         order-lines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                           true  "Tenant identifier"
// @Param        Authorization header    string                           true  "Bearer token"
// @Param        line_id       path      string                           true  "Order line ID (UUID)"
// @Param        body          body      UpdateOrderLineFulfillmentRequest true  "Fulfillment payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/order-lines/{line_id}/fulfillment [patch]
func (h *OrderHandler) UpdateOrderLineFulfillment(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	lineID, err := uuid.Parse(c.Param("line_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid line_id", nil))
		return
	}

	var req UpdateOrderLineFulfillmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	qtyFulfilled, err := numericFromString(req.QuantityFulfilled)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid quantity_fulfilled", nil))
		return
	}

	arg := repository.UpdateOrderLineFulfillmentParams{
		ID:                lineID,
		QuantityFulfilled: qtyFulfilled,
	}

	resp := h.useCase.UpdateOrderLineFulfillment(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// UpdateOrderLineStatus handles PATCH /api/order-lines/:line_id/status
// @Summary      Update order line status
// @Description  Update the status of an order line
// @Tags         order-lines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                       true  "Tenant identifier"
// @Param        Authorization header    string                       true  "Bearer token"
// @Param        line_id       path      string                       true  "Order line ID (UUID)"
// @Param        body          body      UpdateOrderLineStatusRequest true  "Status payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/order-lines/{line_id}/status [patch]
func (h *OrderHandler) UpdateOrderLineStatus(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	lineID, err := uuid.Parse(c.Param("line_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid line_id", nil))
		return
	}

	var req UpdateOrderLineStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	arg := repository.UpdateOrderLineStatusParams{
		ID:         lineID,
		LineStatus: textPtr(&req.LineStatus),
	}

	resp := h.useCase.UpdateOrderLineStatus(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// DeleteOrderLine handles DELETE /api/order-lines/:line_id
// @Summary      Delete order line
// @Description  Delete an order line
// @Tags         order-lines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        line_id       path      string  true  "Order line ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/order-lines/{line_id} [delete]
func (h *OrderHandler) DeleteOrderLine(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	lineID, err := uuid.Parse(c.Param("line_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid line_id", nil))
		return
	}

	resp := h.useCase.DeleteOrderLine(c.Request.Context(), lineID)
	c.JSON(resp.StatusCode, resp)
}

// GetOrderLineTotals handles GET /api/orders/:id/lines/totals
// @Summary      Get order line totals
// @Description  Get aggregated totals for order lines
// @Tags         order-lines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Order ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id}/lines/totals [get]
func (h *OrderHandler) GetOrderLineTotals(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	resp := h.useCase.GetOrderLineTotals(c.Request.Context(), orderID)
	c.JSON(resp.StatusCode, resp)
}

// GetOrderLineMargin handles GET /api/orders/:id/lines/margin
// @Summary      Get order line margin
// @Description  Get margin analytics for order lines
// @Tags         order-lines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Order ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id}/lines/margin [get]
func (h *OrderHandler) GetOrderLineMargin(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	resp := h.useCase.GetOrderLineMargin(c.Request.Context(), orderID)
	c.JSON(resp.StatusCode, resp)
}

// CreateOrderStatusHistory handles POST /api/orders/:id/status-history
// @Summary      Create order status history entry
// @Description  Append a new status history entry for an order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                           true  "Tenant identifier"
// @Param        Authorization header    string                           true  "Bearer token"
// @Param        id            path      string                           true  "Order ID (UUID)"
// @Param        body          body      CreateOrderStatusHistoryRequest  true  "Status history payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id}/status-history [post]
func (h *OrderHandler) CreateOrderStatusHistory(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	var req CreateOrderStatusHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	fromStatus := repository.NullOrderStatusV2{Valid: false}
	if req.FromStatus != nil && *req.FromStatus != "" {
		fromStatus.Valid = true
		fromStatus.OrderStatusV2 = repository.OrderStatusV2(*req.FromStatus)
	}

	arg := repository.CreateOrderStatusHistoryParams{
		SalesOrderID:    orderID,
		OrganizationID:  req.OrganizationID,
		FromStatus:      fromStatus,
		ToStatus:        repository.OrderStatusV2(req.ToStatus),
		Reason:          textPtr(req.Reason),
		Notes:           textPtr(req.Notes),
		ChangedByUserID: int4Ptr(req.ChangedByUserID),
	}

	resp := h.useCase.CreateOrderStatusHistory(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// ListOrderStatusHistory handles GET /api/orders/:id/status-history
// @Summary      List order status history
// @Description  List all status history entries for an order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Order ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id}/status-history [get]
func (h *OrderHandler) ListOrderStatusHistory(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	resp := h.useCase.ListOrderStatusHistory(c.Request.Context(), orderID)
	c.JSON(resp.StatusCode, resp)
}

// CreateOrderFulfillment handles POST /api/orders/:id/fulfillments
// @Summary      Create order fulfillment
// @Description  Create a new fulfillment record for an order
// @Tags         fulfillments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                        true  "Tenant identifier"
// @Param        Authorization header    string                        true  "Bearer token"
// @Param        id            path      string                        true  "Order ID (UUID)"
// @Param        body          body      CreateOrderFulfillmentRequest true  "Fulfillment payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id}/fulfillments [post]
func (h *OrderHandler) CreateOrderFulfillment(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	var req CreateOrderFulfillmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	metaBytes, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}

	estimatedDate, err := datePtrFromYMD(req.EstimatedDeliveryDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid estimated_delivery_date", nil))
		return
	}

	arg := repository.CreateOrderFulfillmentParams{
		SalesOrderID:          orderID,
		OrganizationID:        req.OrganizationID,
		FulfillmentNumber:     req.FulfillmentNumber,
		FulfillmentStatus:     textPtr(req.FulfillmentStatus),
		ShipmentStatus:        textPtr(req.ShipmentStatus),
		FulfillmentStoreID:    int4Ptr(req.FulfillmentStoreID),
		ShippingCarrier:       textPtr(req.ShippingCarrier),
		ShippingMethod:        textPtr(req.ShippingMethod),
		TrackingNumber:        textPtr(req.TrackingNumber),
		TrackingUrl:           textPtr(req.TrackingUrl),
		EstimatedDeliveryDate: estimatedDate,
		Notes:                 textPtr(req.Notes),
		Metadata:              metaBytes,
	}

	resp := h.useCase.CreateOrderFulfillment(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// GetOrderFulfillment handles GET /api/order-fulfillments/:id
// @Summary      Get order fulfillment
// @Description  Retrieve a fulfillment by its ID
// @Tags         fulfillments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Fulfillment ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/order-fulfillments/{id} [get]
func (h *OrderHandler) GetOrderFulfillment(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	fulfillmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid fulfillment id", nil))
		return
	}

	resp := h.useCase.GetOrderFulfillment(c.Request.Context(), fulfillmentID)
	c.JSON(resp.StatusCode, resp)
}

// GetOrderFulfillmentByNumber handles GET /api/order-fulfillments/by-number/:fulfillment_number
// @Summary      Get fulfillment by number
// @Description  Retrieve a fulfillment by its number
// @Tags         fulfillments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id         header    string  true  "Tenant identifier"
// @Param        Authorization       header    string  true  "Bearer token"
// @Param        fulfillment_number  path      string  true  "Fulfillment number"
// @Success      200                 {object}  SuccessResponse
// @Failure      400                 {object}  ErrorResponse
// @Failure      401                 {object}  ErrorResponse
// @Failure      404                 {object}  ErrorResponse
// @Failure      500                 {object}  ErrorResponse
// @Router       /api/order-fulfillments/by-number/{fulfillment_number} [get]
func (h *OrderHandler) GetOrderFulfillmentByNumber(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	number := c.Param("fulfillment_number")
	if number == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "fulfillment_number is required", nil))
		return
	}

	resp := h.useCase.GetOrderFulfillmentByNumber(c.Request.Context(), number)
	c.JSON(resp.StatusCode, resp)
}

// ListOrderFulfillments handles GET /api/orders/:id/fulfillments
// @Summary      List order fulfillments
// @Description  List all fulfillments for an order
// @Tags         fulfillments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Order ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/orders/{id}/fulfillments [get]
func (h *OrderHandler) ListOrderFulfillments(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order id", nil))
		return
	}

	resp := h.useCase.ListOrderFulfillments(c.Request.Context(), orderID)
	c.JSON(resp.StatusCode, resp)
}

// UpdateOrderFulfillment handles PUT /api/order-fulfillments/:id
// @Summary      Update order fulfillment
// @Description  Update fulfillment status, shipment and metadata
// @Tags         fulfillments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                       true  "Tenant identifier"
// @Param        Authorization header    string                       true  "Bearer token"
// @Param        id            path      string                       true  "Fulfillment ID (UUID)"
// @Param        body          body      UpdateOrderFulfillmentRequest true  "Fulfillment update payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/order-fulfillments/{id} [put]
func (h *OrderHandler) UpdateOrderFulfillment(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	fulfillmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid fulfillment id", nil))
		return
	}

	var req UpdateOrderFulfillmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	metaBytes, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}

	estimatedDate, err := datePtrFromYMD(req.EstimatedDeliveryDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid estimated_delivery_date", nil))
		return
	}
	actualDate, err := datePtrFromYMD(req.ActualDeliveryDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid actual_delivery_date", nil))
		return
	}

	arg := repository.UpdateOrderFulfillmentParams{
		ID:                    fulfillmentID,
		FulfillmentStatus:     textPtr(req.FulfillmentStatus),
		ShipmentStatus:        textPtr(req.ShipmentStatus),
		ShippingCarrier:       textPtr(req.ShippingCarrier),
		TrackingNumber:        textPtr(req.TrackingNumber),
		TrackingUrl:           textPtr(req.TrackingUrl),
		EstimatedDeliveryDate: estimatedDate,
		ActualDeliveryDate:    actualDate,
		Notes:                 textPtr(req.Notes),
		Metadata:              metaBytes,
	}

	resp := h.useCase.UpdateOrderFulfillment(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// UpdateFulfillmentShipment handles PUT /api/order-fulfillments/:id/shipment
// @Summary      Update fulfillment shipment status
// @Description  Update shipment status of a fulfillment
// @Tags         fulfillments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                           true  "Tenant identifier"
// @Param        Authorization header    string                           true  "Bearer token"
// @Param        id            path      string                           true  "Fulfillment ID (UUID)"
// @Param        body          body      UpdateFulfillmentShipmentRequest true  "Shipment payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/order-fulfillments/{id}/shipment [put]
func (h *OrderHandler) UpdateFulfillmentShipment(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	fulfillmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid fulfillment id", nil))
		return
	}

	var req UpdateFulfillmentShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	arg := repository.UpdateFulfillmentShipmentParams{
		ID:             fulfillmentID,
		ShipmentStatus: textPtr(&req.ShipmentStatus),
	}

	resp := h.useCase.UpdateFulfillmentShipment(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// UpdateFulfillmentPickPack handles PUT /api/order-fulfillments/:id/pick-pack
// @Summary      Update fulfillment pick/pack info
// @Description  Update pick and pack timestamps and users
// @Tags         fulfillments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                         true  "Tenant identifier"
// @Param        Authorization header    string                         true  "Bearer token"
// @Param        id            path      string                         true  "Fulfillment ID (UUID)"
// @Param        body          body      UpdateFulfillmentPickPackRequest true  "Pick/pack payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/order-fulfillments/{id}/pick-pack [put]
func (h *OrderHandler) UpdateFulfillmentPickPack(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	fulfillmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid fulfillment id", nil))
		return
	}

	var req UpdateFulfillmentPickPackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	pickedAt, err := timestampPtrFromRFC3339(req.PickedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid picked_at", nil))
		return
	}
	packedAt, err := timestampPtrFromRFC3339(req.PackedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid packed_at", nil))
		return
	}

	arg := repository.UpdateFulfillmentPickPackParams{
		ID:             fulfillmentID,
		PickedAt:       pickedAt,
		PackedAt:       packedAt,
		PickedByUserID: int4Ptr(req.PickedByUserID),
		PackedByUserID: int4Ptr(req.PackedByUserID),
	}

	resp := h.useCase.UpdateFulfillmentPickPack(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// DeleteOrderFulfillment handles DELETE /api/order-fulfillments/:id
// @Summary      Delete order fulfillment
// @Description  Delete a fulfillment record
// @Tags         fulfillments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Fulfillment ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/order-fulfillments/{id} [delete]
func (h *OrderHandler) DeleteOrderFulfillment(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	fulfillmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid fulfillment id", nil))
		return
	}

	resp := h.useCase.DeleteOrderFulfillment(c.Request.Context(), fulfillmentID)
	c.JSON(resp.StatusCode, resp)
}

// CreateOrderFulfillmentItem handles POST /api/order-fulfillments/:id/items
// @Summary      Create fulfillment item
// @Description  Create a new item under a fulfillment
// @Tags         fulfillments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                            true  "Tenant identifier"
// @Param        Authorization header    string                            true  "Bearer token"
// @Param        id            path      string                            true  "Fulfillment ID (UUID)"
// @Param        body          body      CreateOrderFulfillmentItemRequest true  "Fulfillment item payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/order-fulfillments/{id}/items [post]
func (h *OrderHandler) CreateOrderFulfillmentItem(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	fulfillmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid fulfillment id", nil))
		return
	}

	var req CreateOrderFulfillmentItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	orderLineID, err := uuidFromString(req.OrderLineID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid order_line_id", nil))
		return
	}

	qty, err := numericFromString(req.QuantityFulfilled)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid quantity_fulfilled", nil))
		return
	}

	arg := repository.CreateOrderFulfillmentItemParams{
		FulfillmentID:     fulfillmentID,
		OrderLineID:       orderLineID,
		OrganizationID:    req.OrganizationID,
		QuantityFulfilled: qty,
		BatchNumber:       textPtr(req.BatchNumber),
		SerialNumbers:     req.SerialNumbers,
	}

	resp := h.useCase.CreateOrderFulfillmentItem(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// ListOrderFulfillmentItems handles GET /api/order-fulfillments/:id/items
// @Summary      List fulfillment items
// @Description  List all items for a fulfillment
// @Tags         fulfillments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Fulfillment ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/order-fulfillments/{id}/items [get]
func (h *OrderHandler) ListOrderFulfillmentItems(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	fulfillmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid fulfillment id", nil))
		return
	}

	resp := h.useCase.ListOrderFulfillmentItems(c.Request.Context(), fulfillmentID)
	c.JSON(resp.StatusCode, resp)
}

// DeleteOrderFulfillmentItem handles DELETE /api/order-fulfillment-items/:item_id
// @Summary      Delete fulfillment item
// @Description  Delete a single fulfillment item
// @Tags         fulfillments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        item_id       path      string  true  "Fulfillment item ID (UUID)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/order-fulfillment-items/{item_id} [delete]
func (h *OrderHandler) DeleteOrderFulfillmentItem(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid item_id", nil))
		return
	}

	resp := h.useCase.DeleteOrderFulfillmentItem(c.Request.Context(), itemID)
	c.JSON(resp.StatusCode, resp)
}
