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

type GoodsReceiptNotesHandler struct {
	useCase *usecase.GoodsReceiptNotesUseCase
}

func NewGoodsReceiptNotesHandler(uc *usecase.GoodsReceiptNotesUseCase) *GoodsReceiptNotesHandler {
	return &GoodsReceiptNotesHandler{
		useCase: uc,
	}
}

func (h *GoodsReceiptNotesHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "repository not found in context", nil))
		c.Abort()
		return nil
	}
	return repo
}

// CreateGoodsReceiptNote handles POST /api/goods-receipt-notes
// @Summary      Create Goods Receipt Note (GRN)
// @Description  Creates a new dock delivery receipt for incoming purchase orders or direct deliveries
// @Tags         goods-receipt-notes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string            true   "Tenant identifier"
// @Param        Authorization header    string            true   "Bearer token"
// @Param        body          body      CreateGRNRequest  true   "GRN creation payload"
// @Success      200           {object}  GRNResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/goods-receipt-notes [post]
func (h *GoodsReceiptNotesHandler) CreateGoodsReceiptNote(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req usecase.CreateGoodsReceiptNoteInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("invalid request body: %v", err), nil))
		return
	}

	resp := h.useCase.CreateGoodsReceiptNote(c.Request.Context(), req)
	c.JSON(resp.StatusCode, resp)
}

// ListGoodsReceiptNotes handles GET /api/goods-receipt-notes
// @Summary      List Goods Receipt Notes (GRN)
// @Description  List Goods Receipt Notes with filtering by organization, store, supplier, PO, status, date range, search, and pagination
// @Tags         goods-receipt-notes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id        header    string  true   "Tenant identifier"
// @Param        Authorization      header    string  true   "Bearer token"
// @Param        organization_id    query     int     false  "Organization ID"
// @Param        store_id           query     int     false  "Store ID"
// @Param        supplier_id        query     int     false  "Supplier / Partner ID"
// @Param        purchase_order_id  query     int     false  "Purchase Order ID"
// @Param        status             query     string  false  "Status (draft, posted, cancelled)"
// @Param        from_date          query     string  false  "From date (YYYY-MM-DD)"
// @Param        to_date            query     string  false  "To date (YYYY-MM-DD)"
// @Param        search             query     string  false  "Search GRN number, DN number, PO number, or supplier name"
// @Param        page               query     int     false  "Page number" default(1)
// @Param        limit              query     int     false  "Items per page" default(20)
// @Success      200                {object}  GRNListResponse
// @Failure      400                {object}  ErrorResponse
// @Failure      500                {object}  ErrorResponse
// @Router       /api/goods-receipt-notes [get]
func (h *GoodsReceiptNotesHandler) ListGoodsReceiptNotes(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	filter := usecase.GoodsReceiptNoteFilter{}

	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		if orgID, err := strconv.ParseInt(orgIDStr, 10, 32); err == nil && orgID > 0 {
			oID := int32(orgID)
			filter.OrganizationID = &oID
		}
	}

	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		if storeID, err := strconv.ParseInt(storeIDStr, 10, 32); err == nil {
			sID := int32(storeID)
			filter.StoreID = &sID
		}
	}

	if supplierIDStr := c.Query("supplier_id"); supplierIDStr != "" {
		if supplierID, err := strconv.ParseInt(supplierIDStr, 10, 32); err == nil {
			suppID := int32(supplierID)
			filter.SupplierID = &suppID
		}
	}

	if poIDStr := c.Query("purchase_order_id"); poIDStr != "" {
		if poID, err := strconv.ParseInt(poIDStr, 10, 32); err == nil {
			pID := int32(poID)
			filter.PurchaseOrderID = &pID
		}
	}

	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	if fromDate := c.Query("from_date"); fromDate != "" {
		filter.FromDate = &fromDate
	}

	if toDate := c.Query("to_date"); toDate != "" {
		filter.ToDate = &toDate
	}

	page := int32(1)
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.ParseInt(pageStr, 10, 32); err == nil && p > 0 {
			page = int32(p)
		}
	}
	filter.Page = page

	pageSize := int32(20)
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil && l > 0 {
			pageSize = int32(l)
		}
	}
	filter.PageSize = pageSize

	resp := h.useCase.ListGoodsReceiptNotes(c.Request.Context(), filter)
	c.JSON(resp.StatusCode, resp)
}

// GetGoodsReceiptNote handles GET /api/goods-receipt-notes/:id
// @Summary      Get Goods Receipt Note by ID
// @Description  Get detailed Goods Receipt Note with itemized lines
// @Tags         goods-receipt-notes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "GRN ID"
// @Success      200           {object}  GRNResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/goods-receipt-notes/{id} [get]
func (h *GoodsReceiptNotesHandler) GetGoodsReceiptNote(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid goods receipt note id", nil))
		return
	}

	resp := h.useCase.GetGoodsReceiptNote(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// PostGoodsReceiptNote handles POST /api/goods-receipt-notes/:id/post
// @Summary      Post / Complete Goods Receipt Note
// @Description  Posts dock GRN, updating PO received quantities, PO status, inventory stock, and stock movements
// @Tags         goods-receipt-notes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "GRN ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/goods-receipt-notes/{id}/post [post]
func (h *GoodsReceiptNotesHandler) PostGoodsReceiptNote(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid goods receipt note id", nil))
		return
	}

	resp := h.useCase.PostGoodsReceiptNote(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// UpdateGoodsReceiptNote handles PUT /api/goods-receipt-notes/:id
// @Summary      Update Goods Receipt Note (GRN)
// @Description  Updates header and item lines for a draft Goods Receipt Note
// @Tags         goods-receipt-notes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string            true   "Tenant identifier"
// @Param        Authorization header    string            true   "Bearer token"
// @Param        id            path      int               true   "GRN ID"
// @Param        body          body      UpdateGRNRequest  true   "GRN update payload"
// @Success      200           {object}  GRNResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/goods-receipt-notes/{id} [put]
func (h *GoodsReceiptNotesHandler) UpdateGoodsReceiptNote(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid goods receipt note id", nil))
		return
	}

	var req usecase.UpdateGoodsReceiptNoteInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("invalid request body: %v", err), nil))
		return
	}

	resp := h.useCase.UpdateGoodsReceiptNote(c.Request.Context(), int32(id), req)
	c.JSON(resp.StatusCode, resp)
}

// DeleteGoodsReceiptNote handles DELETE /api/goods-receipt-notes/:id
// @Summary      Delete Goods Receipt Note
// @Description  Deletes a draft Goods Receipt Note
// @Tags         goods-receipt-notes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "GRN ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/goods-receipt-notes/{id} [delete]
func (h *GoodsReceiptNotesHandler) DeleteGoodsReceiptNote(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid goods receipt note id", nil))
		return
	}

	resp := h.useCase.DeleteGoodsReceiptNote(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

