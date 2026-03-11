package handler

import (
	"net/http"
	"strconv"
	"time"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"
	"NEMBUS/utils"

	"github.com/gin-gonic/gin"
)

// StockMovementHandler handles stock movement HTTP requests.
type StockMovementHandler struct {
	useCase *usecase.StockMovementsUseCase
}

// NewStockMovementHandler creates a new handler instance.
func NewStockMovementHandler(uc *usecase.StockMovementsUseCase) *StockMovementHandler {
	return &StockMovementHandler{
		useCase: uc,
	}
}

func (h *StockMovementHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "repository not found in context", nil))
		c.Abort()
		return nil
	}
	return repo
}

// CreateStockMovement handles POST /api/stock-movements
// @Summary      Create stock movement
// @Description  Register a new stock movement (in, out, transfer, adjustment, etc.)
// @Tags         stock-movements
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                     true   "Tenant identifier"
// @Param        Authorization header    string                     true   "Bearer token"
// @Param        body          body      usecase.CreateStockMovementInput true "Stock movement payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-movements [post]
func (h *StockMovementHandler) CreateStockMovement(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req usecase.CreateStockMovementInput
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request", nil))
		return
	}

	resp := h.useCase.CreateStockMovement(c.Request.Context(), &req)
	c.JSON(resp.StatusCode, resp)
}

// GetStockMovement handles GET /api/stock-movements/:id
// @Summary      Get stock movement by ID
// @Description  Retrieve detailed information about a specific stock movement
// @Tags         stock-movements
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Stock movement ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-movements/{id} [get]
func (h *StockMovementHandler) GetStockMovement(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetStockMovement(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// ListStockMovements handles GET /api/stock-movements
// @Summary      List stock movements
// @Description  Get paginated list of all stock movements
// @Tags         stock-movements
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        limit         query     int     false  "Number of records per page" default(10)
// @Param        offset        query     int     false  "Offset for pagination" default(0)
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-movements [get]
func (h *StockMovementHandler) ListStockMovements(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.ParseInt(limitStr, 10, 32)
	offset, _ := strconv.ParseInt(offsetStr, 10, 32)

	resp := h.useCase.ListStockMovements(c.Request.Context(), int32(limit), int32(offset))
	c.JSON(resp.StatusCode, resp)
}

// ListStockMovementsByProduct handles GET /api/stock-movements/product/:productID
// @Summary      List stock movements for a product
// @Description  Get paginated stock movements history for a specific product
// @Tags         stock-movements
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        productID     path      string  true   "Product ID"
// @Param        limit         query     int     false  "Number of records per page" default(10)
// @Param        offset        query     int     false  "Offset for pagination" default(0)
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-movements/product/{productID} [get]
func (h *StockMovementHandler) ListStockMovementsByProduct(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	productID := c.Param("productID")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.ParseInt(limitStr, 10, 32)
	offset, _ := strconv.ParseInt(offsetStr, 10, 32)

	resp := h.useCase.ListStockMovementsByProduct(c.Request.Context(), productID, int32(limit), int32(offset))
	c.JSON(resp.StatusCode, resp)
}

// ListStockMovementsByDateRange handles GET /api/stock-movements/daterange
// @Summary      List stock movements in date range
// @Description  Get stock movements filtered by date range (inclusive)
// @Tags         stock-movements
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        start         query     string  true   "Start date (RFC3339)" example(2026-03-01T00:00:00Z)
// @Param        end           query     string  true   "End date (RFC3339)"   example(2026-03-10T23:59:59Z)
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-movements/daterange [get]
func (h *StockMovementHandler) ListStockMovementsByDateRange(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	startStr := c.Query("start")
	endStr := c.Query("end")
	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "start and end dates are required", nil))
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid start date format", nil))
		return
	}
	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid end date format", nil))
		return
	}

	resp := h.useCase.ListStockMovementsByDateRange(c.Request.Context(), &start, &end)
	c.JSON(resp.StatusCode, resp)
}

// UpdateStockMovementStatus handles PATCH /api/stock-movements/:id/status
// @Summary      Update stock movement status
// @Description  Change status of a stock movement (e.g. pending → confirmed, cancelled, etc.)
// @Tags         stock-movements
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        id            path      string  true   "Stock movement ID"
// @Param        body          body      object  true   "Status update payload"
// @Param        body.status   body      string  true   "New status" enum(pending,confirmed,cancelled,completed)
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-movements/{id}/status [patch]
func (h *StockMovementHandler) UpdateStockMovementStatus(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := c.BindJSON(&body); err != nil || body.Status == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "status is required", nil))
		return
	}

	resp := h.useCase.UpdateStockMovementStatus(c.Request.Context(), id, body.Status)
	c.JSON(resp.StatusCode, resp)
}

// ListStockMovementsByProductWithDateRange handles GET /api/stock-movements/product/:productID/daterange
// @Summary      List stock movements for product in date range
// @Description  Get paginated stock movements for a specific product, optionally filtered by date range
// @Tags         stock-movements
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header      string  true   "Tenant identifier"
// @Param        Authorization header      string  true   "Bearer token"
// @Param        productID     path        string  true   "Product ID"
// @Param        start_date    query       string  false  "Start date (RFC3339)" example(2026-03-01T00:00:00Z)
// @Param        end_date      query       string  false  "End date (RFC3339)"   example(2026-03-10T23:59:59Z)
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/stock-movements/product/{productID}/daterange [get]
func (h *StockMovementHandler) ListStockMovementsByProductWithDateRange(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	productID := c.Param("productID")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate *time.Time
	if startDateStr != "" {
		t, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid start_date format", nil))
			return
		}
		startDate = &t
	}
	if endDateStr != "" {
		t, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid end_date format", nil))
			return
		}
		endDate = &t
	}

	resp := h.useCase.ListStockMovementsByProductWithDateRange(c.Request.Context(), productID, startDate, endDate)
	c.JSON(resp.StatusCode, resp)
}
