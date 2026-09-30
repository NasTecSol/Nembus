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

// StockCountsHandler handles stock counts endpoints.
type StockCountsHandler struct {
	useCase *usecase.StockCountsUseCase
}

// NewStockCountsHandler creates a new StockCountsHandler instance.
func NewStockCountsHandler(uc *usecase.StockCountsUseCase) *StockCountsHandler {
	return &StockCountsHandler{
		useCase: uc,
	}
}

func (h *StockCountsHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "repository not found in context", nil))
		c.Abort()
		return nil
	}
	return repo
}

// CreateStockCount handles POST /api/stock-counts
// @Summary      Create Stock Count
// @Description  Creates a new stock count session with planned status and optional line items
// @Tags         stock-counts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                   true  "Tenant identifier"
// @Param        Authorization header    string                   true  "Bearer token"
// @Param        body          body      CreateStockCountRequest  true  "Stock count creation payload"
// @Success      200           {object}  StockCountResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-counts [post]
func (h *StockCountsHandler) CreateStockCount(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req usecase.CreateStockCountInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("invalid request body: %v", err), nil))
		return
	}

	resp := h.useCase.CreateStockCount(c.Request.Context(), req)
	c.JSON(resp.StatusCode, resp)
}

// ListStockCounts handles GET /api/stock-counts
// @Summary      List Stock Counts
// @Description  List stock counts with filtering by store, status, count_type, date range, search, and pagination
// @Tags         stock-counts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        store_id      query     int     false  "Store ID"
// @Param        status        query     string  false  "Status (planned, in_progress, completed, approved, reconciled, cancelled)"
// @Param        count_type    query     string  false  "Count type (full, cycle, spot, annual)"
// @Param        from_date     query     string  false  "From date (YYYY-MM-DD)"
// @Param        to_date       query     string  false  "To date (YYYY-MM-DD)"
// @Param        search        query     string  false  "Search count number, store name, or counter name"
// @Param        page          query     int     false  "Page number" default(1)
// @Param        limit         query     int     false  "Items per page" default(20)
// @Success      200           {object}  StockCountListResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-counts [get]
func (h *StockCountsHandler) ListStockCounts(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	filter := usecase.StockCountFilter{}

	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		if storeID, err := strconv.ParseInt(storeIDStr, 10, 32); err == nil && storeID > 0 {
			sID := int32(storeID)
			filter.StoreID = &sID
		}
	}

	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	if countType := c.Query("count_type"); countType != "" {
		filter.CountType = &countType
	}

	if fromDate := c.Query("from_date"); fromDate != "" {
		filter.FromDate = &fromDate
	}

	if toDate := c.Query("to_date"); toDate != "" {
		filter.ToDate = &toDate
	}

	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	filter.Page = int32(page)

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	filter.Limit = int32(limit)

	resp := h.useCase.ListStockCounts(c.Request.Context(), filter)
	c.JSON(resp.StatusCode, resp)
}

// GetStockCount handles GET /api/stock-counts/:id
// @Summary      Get Stock Count by ID
// @Description  Get detailed stock count header with line items and variance calculations
// @Tags         stock-counts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Stock Count ID"
// @Success      200           {object}  StockCountResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-counts/{id} [get]
func (h *StockCountsHandler) GetStockCount(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid stock count id", nil))
		return
	}

	resp := h.useCase.GetStockCount(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// UpdateStockCount handles PUT /api/stock-counts/:id
// @Summary      Update Stock Count Header
// @Description  Updates stock count header properties like location, scheduled date, count type, and metadata
// @Tags         stock-counts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                   true  "Tenant identifier"
// @Param        Authorization header    string                   true  "Bearer token"
// @Param        id            path      int                      true  "Stock Count ID"
// @Param        body          body      UpdateStockCountRequest  true  "Stock count update payload"
// @Success      200           {object}  StockCountResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-counts/{id} [put]
func (h *StockCountsHandler) UpdateStockCount(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid stock count id", nil))
		return
	}

	var req usecase.UpdateStockCountInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("invalid request body: %v", err), nil))
		return
	}

	resp := h.useCase.UpdateStockCount(c.Request.Context(), int32(id), req)
	c.JSON(resp.StatusCode, resp)
}

// DeleteStockCount handles DELETE /api/stock-counts/:id
// @Summary      Delete Stock Count
// @Description  Deletes a stock count and its lines (cannot delete completed or reconciled counts)
// @Tags         stock-counts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Stock Count ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-counts/{id} [delete]
func (h *StockCountsHandler) DeleteStockCount(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid stock count id", nil))
		return
	}

	resp := h.useCase.DeleteStockCount(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// StartStockCount handles POST /api/stock-counts/:id/start
// @Summary      Start Stock Count
// @Description  Transitions a stock count status from planned to in_progress and records started timestamp
// @Tags         stock-counts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Stock Count ID"
// @Success      200           {object}  StockCountResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-counts/{id}/start [post]
func (h *StockCountsHandler) StartStockCount(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid stock count id", nil))
		return
	}

	resp := h.useCase.StartStockCount(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// CompleteStockCount handles POST /api/stock-counts/:id/complete
// @Summary      Complete Stock Count
// @Description  Transitions a stock count status from in_progress to completed and records completed timestamp
// @Tags         stock-counts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Stock Count ID"
// @Success      200           {object}  StockCountResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-counts/{id}/complete [post]
func (h *StockCountsHandler) CompleteStockCount(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid stock count id", nil))
		return
	}

	resp := h.useCase.CompleteStockCount(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// ApproveStockCount handles POST /api/stock-counts/:id/approve
// @Summary      Approve Stock Count
// @Description  Approves a completed stock count session
// @Tags         stock-counts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                    true  "Tenant identifier"
// @Param        Authorization header    string                    true  "Bearer token"
// @Param        id            path      int                       true  "Stock Count ID"
// @Param        body          body      ApproveStockCountRequest  false "Approver payload"
// @Success      200           {object}  StockCountResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-counts/{id}/approve [post]
func (h *StockCountsHandler) ApproveStockCount(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid stock count id", nil))
		return
	}

	var req usecase.ApproveStockCountInput
	_ = c.ShouldBindJSON(&req)

	resp := h.useCase.ApproveStockCount(c.Request.Context(), int32(id), req.ApprovedBy)
	c.JSON(resp.StatusCode, resp)
}

// ReconcileStockCount handles POST /api/stock-counts/:id/reconcile
// @Summary      Reconcile Stock Count
// @Description  Reconciles counted inventory with system stock, applies adjustments to inventory_stock, and logs stock_movements
// @Tags         stock-counts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Stock Count ID"
// @Success      200           {object}  StockCountResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-counts/{id}/reconcile [post]
func (h *StockCountsHandler) ReconcileStockCount(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid stock count id", nil))
		return
	}

	resp := h.useCase.ReconcileStockCount(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// GetStockCountSummary handles GET /api/stock-counts/:id/summary
// @Summary      Get Stock Count Variance Summary
// @Description  Calculates total lines, variance counts, positive/negative variance values
// @Tags         stock-counts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Stock Count ID"
// @Success      200           {object}  StockCountSummaryResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-counts/{id}/summary [get]
func (h *StockCountsHandler) GetStockCountSummary(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid stock count id", nil))
		return
	}

	resp := h.useCase.GetStockCountSummary(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// AddStockCountLine handles POST /api/stock-counts/:id/lines
// @Summary      Add Line Item to Stock Count
// @Description  Adds a product/variant count line to an existing stock count
// @Tags         stock-counts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string             true  "Tenant identifier"
// @Param        Authorization header    string             true  "Bearer token"
// @Param        id            path      int                true  "Stock Count ID"
// @Param        body          body      StockCountLineDTO  true  "Line item creation payload"
// @Success      201           {object}  StockCountLineResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-counts/{id}/lines [post]
func (h *StockCountsHandler) AddStockCountLine(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	countIDStr := c.Param("id")
	countID, err := strconv.ParseInt(countIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid stock count id", nil))
		return
	}

	var req usecase.StockCountLineInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("invalid request body: %v", err), nil))
		return
	}

	resp := h.useCase.AddStockCountLine(c.Request.Context(), int32(countID), req)
	c.JSON(resp.StatusCode, resp)
}

// ListStockCountLines handles GET /api/stock-counts/:id/lines
// @Summary      List Line Items for Stock Count
// @Description  Returns all line items with product and variance details for a specific stock count
// @Tags         stock-counts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Stock Count ID"
// @Success      200           {array}   StockCountLineResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-counts/{id}/lines [get]
func (h *StockCountsHandler) ListStockCountLines(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	countIDStr := c.Param("id")
	countID, err := strconv.ParseInt(countIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid stock count id", nil))
		return
	}

	resp := h.useCase.ListStockCountLines(c.Request.Context(), int32(countID))
	c.JSON(resp.StatusCode, resp)
}

// GetStockCountLine handles GET /api/stock-counts/:id/lines/:line_id
// @Summary      Get Stock Count Line by ID
// @Description  Get a single stock count line item by its ID
// @Tags         stock-counts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Stock Count ID"
// @Param        line_id       path      int     true  "Line Item ID"
// @Success      200           {object}  StockCountLineResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-counts/{id}/lines/{line_id} [get]
func (h *StockCountsHandler) GetStockCountLine(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	lineIDStr := c.Param("line_id")
	lineID, err := strconv.ParseInt(lineIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid stock count line id", nil))
		return
	}

	resp := h.useCase.GetStockCountLine(c.Request.Context(), int32(lineID))
	c.JSON(resp.StatusCode, resp)
}

// UpdateStockCountLine handles PUT /api/stock-counts/:id/lines/:line_id
// @Summary      Update Stock Count Line
// @Description  Updates counted quantity, batch, serial number, or metadata for a single line item
// @Tags         stock-counts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                       true  "Tenant identifier"
// @Param        Authorization header    string                       true  "Bearer token"
// @Param        id            path      int                          true  "Stock Count ID"
// @Param        line_id       path      int                          true  "Line Item ID"
// @Param        body          body      UpdateStockCountLineRequest  true  "Line item update payload"
// @Success      200           {object}  StockCountLineResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-counts/{id}/lines/{line_id} [put]
func (h *StockCountsHandler) UpdateStockCountLine(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	lineIDStr := c.Param("line_id")
	lineID, err := strconv.ParseInt(lineIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid stock count line id", nil))
		return
	}

	var req usecase.UpdateStockCountLineInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("invalid request body: %v", err), nil))
		return
	}

	resp := h.useCase.UpdateStockCountLine(c.Request.Context(), int32(lineID), req)
	c.JSON(resp.StatusCode, resp)
}

// DeleteStockCountLine handles DELETE /api/stock-counts/:id/lines/:line_id
// @Summary      Delete Stock Count Line
// @Description  Deletes a single line item from a stock count
// @Tags         stock-counts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Stock Count ID"
// @Param        line_id       path      int     true  "Line Item ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-counts/{id}/lines/{line_id} [delete]
func (h *StockCountsHandler) DeleteStockCountLine(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	lineIDStr := c.Param("line_id")
	lineID, err := strconv.ParseInt(lineIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid stock count line id", nil))
		return
	}

	resp := h.useCase.DeleteStockCountLine(c.Request.Context(), int32(lineID))
	c.JSON(resp.StatusCode, resp)
}

// BulkUpdateStockCountLines handles PUT /api/stock-counts/:id/lines
// @Summary      Bulk Update Stock Count Lines
// @Description  Submits counted quantities for multiple line items in a stock count at once
// @Tags         stock-counts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                            true  "Tenant identifier"
// @Param        Authorization header    string                            true  "Bearer token"
// @Param        id            path      int                               true  "Stock Count ID"
// @Param        body          body      BulkUpdateStockCountLinesRequest  true  "Bulk update payload"
// @Success      200           {object}  StockCountResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-counts/{id}/lines [put]
func (h *StockCountsHandler) BulkUpdateStockCountLines(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	countIDStr := c.Param("id")
	countID, err := strconv.ParseInt(countIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid stock count id", nil))
		return
	}

	var req struct {
		Lines []usecase.BulkStockCountLineItem `json:"lines" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, fmt.Sprintf("invalid request body: %v", err), nil))
		return
	}

	resp := h.useCase.BulkUpdateStockCountLines(c.Request.Context(), int32(countID), req.Lines)
	c.JSON(resp.StatusCode, resp)
}
