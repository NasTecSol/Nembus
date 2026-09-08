package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/gin-gonic/gin"
)

type PurchaseOrdersHandler struct {
	useCase *usecase.PurchaseOrdersUseCase
}

func NewPurchaseOrdersHandler(uc *usecase.PurchaseOrdersUseCase) *PurchaseOrdersHandler {
	return &PurchaseOrdersHandler{
		useCase: uc,
	}
}

func (h *PurchaseOrdersHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "repository not found in context", nil))
		c.Abort()
		return nil
	}
	return repo
}

// CreatePurchaseOrder handles POST /api/purchase-orders
// @Summary      Create Purchase Order
// @Description  Creates a new purchase order with itemized lines in draft status
// @Tags         purchase-orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                      true   "Tenant identifier"
// @Param        Authorization header    string                      true   "Bearer token"
// @Param        body          body      CreatePurchaseOrderRequest  true   "Purchase order creation payload"
// @Success      200           {object}  PurchaseOrderResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/purchase-orders [post]
func (h *PurchaseOrdersHandler) CreatePurchaseOrder(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req usecase.CreatePurchaseOrderInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("invalid request body: %v", err), nil))
		return
	}

	resp := h.useCase.CreatePurchaseOrder(c.Request.Context(), req)
	c.JSON(resp.StatusCode, resp)
}

// GetPurchaseOrder handles GET /api/purchase-orders/:id
// @Summary      Get Purchase Order by ID
// @Description  Get detailed purchase order with supplier, store, user details and line items
// @Tags         purchase-orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Purchase Order ID"
// @Success      200           {object}  PurchaseOrderResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/purchase-orders/{id} [get]
func (h *PurchaseOrdersHandler) GetPurchaseOrder(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid purchase order id", nil))
		return
	}

	resp := h.useCase.GetPurchaseOrder(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// ListPurchaseOrders handles GET /api/purchase-orders
// @Summary      List Purchase Orders
// @Description  List purchase orders with filtering by organization, store, supplier, status, date range, search, and pagination
// @Tags         purchase-orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true   "Tenant identifier"
// @Param        Authorization   header    string  true   "Bearer token"
// @Param        organization_id query     int     true   "Organization ID"
// @Param        store_id        query     int     false  "Store ID"
// @Param        partner_id      query     int     false  "Supplier / Business Partner ID"
// @Param        status          query     string  false  "Status (draft, submitted, approved, partially_received, received, cancelled, closed)"
// @Param        from_date       query     string  false  "From date (YYYY-MM-DD)"
// @Param        to_date         query     string  false  "To date (YYYY-MM-DD)"
// @Param        search          query     string  false  "Search PO number or supplier name"
// @Param        page            query     int     false  "Page number" default(1)
// @Param        limit           query     int     false  "Items per page" default(20)
// @Success      200             {object}  PurchaseOrderListResponse
// @Failure      400             {object}  ErrorResponse
// @Failure      500             {object}  ErrorResponse
// @Router       /api/purchase-orders [get]
func (h *PurchaseOrdersHandler) ListPurchaseOrders(c *gin.Context) {
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

	filter := usecase.PurchaseOrderFilter{
		OrganizationID: int32(orgID),
	}

	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		if storeID, err := strconv.ParseInt(storeIDStr, 10, 32); err == nil {
			sID := int32(storeID)
			filter.StoreID = &sID
		}
	}

	if partnerIDStr := c.Query("partner_id"); partnerIDStr != "" {
		if partnerID, err := strconv.ParseInt(partnerIDStr, 10, 32); err == nil {
			pID := int32(partnerID)
			filter.PartnerID = &pID
		}
	}

	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	if fromDateStr := c.Query("from_date"); fromDateStr != "" {
		if fromDate, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			filter.FromDate = &fromDate
		}
	}

	if toDateStr := c.Query("to_date"); toDateStr != "" {
		if toDate, err := time.Parse("2006-01-02", toDateStr); err == nil {
			filter.ToDate = &toDate
		}
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	page, _ := strconv.ParseInt(pageStr, 10, 32)
	limit, _ := strconv.ParseInt(limitStr, 10, 32)
	filter.Page = int32(page)
	filter.Limit = int32(limit)

	resp := h.useCase.ListPurchaseOrders(c.Request.Context(), filter)
	c.JSON(resp.StatusCode, resp)
}

// UpdatePurchaseOrder handles PUT /api/purchase-orders/:id
// @Summary      Update Purchase Order
// @Description  Updates header and line items of a draft purchase order
// @Tags         purchase-orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                      true   "Tenant identifier"
// @Param        Authorization header    string                      true   "Bearer token"
// @Param        id            path      int                         true   "Purchase Order ID"
// @Param        body          body      UpdatePurchaseOrderRequest  true   "Purchase order update payload"
// @Success      200           {object}  PurchaseOrderResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/purchase-orders/{id} [put]
func (h *PurchaseOrdersHandler) UpdatePurchaseOrder(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid purchase order id", nil))
		return
	}

	var req usecase.UpdatePurchaseOrderInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("invalid request body: %v", err), nil))
		return
	}

	resp := h.useCase.UpdatePurchaseOrder(c.Request.Context(), int32(id), req)
	c.JSON(resp.StatusCode, resp)
}

// UpdatePurchaseOrderStatus handles PATCH /api/purchase-orders/:id/status
// @Summary      Update Purchase Order Status
// @Description  Update purchase order lifecycle status (draft, submitted, approved, cancelled, closed)
// @Tags         purchase-orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                            true   "Tenant identifier"
// @Param        Authorization header    string                            true   "Bearer token"
// @Param        id            path      int                               true   "Purchase Order ID"
// @Param        body          body      UpdatePurchaseOrderStatusRequest  true   "Status update payload"
// @Success      200           {object}  PurchaseOrderResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/purchase-orders/{id}/status [patch]
func (h *PurchaseOrdersHandler) UpdatePurchaseOrderStatus(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid purchase order id", nil))
		return
	}

	var req usecase.UpdatePurchaseOrderStatusInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("invalid request body: %v", err), nil))
		return
	}

	resp := h.useCase.UpdatePurchaseOrderStatus(c.Request.Context(), int32(id), req)
	c.JSON(resp.StatusCode, resp)
}

// ApprovePurchaseOrder handles POST /api/purchase-orders/:id/approve
// @Summary      Approve Purchase Order
// @Description  Approves a submitted purchase order
// @Tags         purchase-orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                            true   "Tenant identifier"
// @Param        Authorization header    string                            true   "Bearer token"
// @Param        id            path      int                               true   "Purchase Order ID"
// @Param        body          body      UpdatePurchaseOrderStatusRequest  false  "Optional approval payload"
// @Success      200           {object}  PurchaseOrderResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/purchase-orders/{id}/approve [post]
func (h *PurchaseOrdersHandler) ApprovePurchaseOrder(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid purchase order id", nil))
		return
	}

	var req usecase.UpdatePurchaseOrderStatusInput
	_ = c.ShouldBindJSON(&req)
	req.Status = "approved"

	resp := h.useCase.UpdatePurchaseOrderStatus(c.Request.Context(), int32(id), req)
	c.JSON(resp.StatusCode, resp)
}

// DeletePurchaseOrder handles DELETE /api/purchase-orders/:id
// @Summary      Delete Purchase Order
// @Description  Deletes a draft or cancelled purchase order
// @Tags         purchase-orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Purchase Order ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/purchase-orders/{id} [delete]
func (h *PurchaseOrdersHandler) DeletePurchaseOrder(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid purchase order id", nil))
		return
	}

	resp := h.useCase.DeletePurchaseOrder(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}
