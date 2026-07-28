package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type CashierSessionHandler struct {
	useCase *usecase.CashierSessionUseCase
}

func NewCashierSessionHandler(uc *usecase.CashierSessionUseCase) *CashierSessionHandler {
	return &CashierSessionHandler{useCase: uc}
}

func (h *CashierSessionHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// OpenCashierSession handles POST /api/cashier-sessions
// @Summary      Open cashier session
// @Description  Open a new cashier session
// @Tags         cashier-sessions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                     true  "Tenant identifier"
// @Param        Authorization header    string                     true  "Bearer token"
// @Param        body          body      OpenCashierSessionRequest  true  "Session payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashier-sessions [post]
func (h *CashierSessionHandler) OpenCashierSession(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req OpenCashierSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	openingBalance, err := numericFromString(req.OpeningBalance)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid opening_balance", nil))
		return
	}

	arg := repository.OpenCashierSessionParams{
		CashierID:      req.CashierID,
		PosTerminalID:  req.PosTerminalID,
		SessionNumber:  req.SessionNumber,
		OpeningBalance: openingBalance,
	}

	resp := h.useCase.OpenCashierSession(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// GetSessionByID handles GET /api/cashier-sessions/:id
// @Summary      Get cashier session by ID
// @Description  Returns a cashier session by ID (open or closed)
// @Tags         cashier-sessions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Session ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashier-sessions/{id} [get]
func (h *CashierSessionHandler) GetSessionByID(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid session id", nil))
		return
	}

	resp := h.useCase.GetSessionByID(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// GetActiveCashierSession handles GET /api/cashier-sessions/active/:cashier_id
// @Summary      Get active cashier session
// @Description  Get the active session for a specific cashier
// @Tags         cashier-sessions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        cashier_id    path      int     true  "Cashier ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashier-sessions/active/{cashier_id} [get]
func (h *CashierSessionHandler) GetActiveCashierSession(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	cashierIDStr := c.Param("cashier_id")
	cashierID, err := strconv.Atoi(cashierIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cashier_id", nil))
		return
	}

	resp := h.useCase.GetActiveCashierSession(c.Request.Context(), int32(cashierID))
	c.JSON(resp.StatusCode, resp)
}

// CloseCashierSession handles PUT /api/cashier-sessions/:id/close
// @Summary      Close cashier session
// @Description  Close an existing cashier session
// @Tags         cashier-sessions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                      true  "Tenant identifier"
// @Param        Authorization header    string                      true  "Bearer token"
// @Param        id            path      int                         true  "Session ID"
// @Param        body          body      CloseCashierSessionRequest  true  "Close session payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashier-sessions/{id}/close [put]
func (h *CashierSessionHandler) CloseCashierSession(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid session id", nil))
		return
	}

	var req CloseCashierSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	closingBalance, err := numericFromString(req.ClosingBalance)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid closing_balance", nil))
		return
	}
	// expected_balance and variance are optional; usecase reconciles using DB expected_balance and computes variance
	expectedBalance, _ := numericFromString(req.ExpectedBalance)
	variance, _ := numericFromString(req.Variance)

	arg := repository.CloseCashierSessionParams{
		ID:              int32(id),
		ClosingBalance:  closingBalance,
		ExpectedBalance: expectedBalance,
		Variance:        variance,
		Column5:         req.ClosingNote,
		Column6:         req.ClosedBy,
	}

	resp := h.useCase.CloseCashierSession(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// GetSessionSummary handles GET /api/cashier-sessions/:id/summary
// @Summary      Get session summary
// @Description  Get summary stats for a cashier session
// @Tags         cashier-sessions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Session ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashier-sessions/{id}/summary [get]
func (h *CashierSessionHandler) GetSessionSummary(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid session id", nil))
		return
	}

	resp := h.useCase.GetSessionSummary(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// GetCashierSessions handles GET /api/cashier-sessions/list/:cashier_id
// @Summary      Get cashier sessions with optional date and status filters
// @Description  Get sessions for a specific cashier. Status can be 'all', 'open', or 'closed'.
// @Tags         cashier-sessions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        cashier_id    path      int     true   "Cashier ID"
// @Param        status        query     string  false  "Status filter: all, open, closed" default(all)
// @Param        start_date    query     string  false  "Start date (e.g. 2006-01-02)"
// @Param        end_date      query     string  false  "End date (e.g. 2006-01-02)"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/cashier-sessions/list/{cashier_id} [get]
func (h *CashierSessionHandler) GetCashierSessions(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	cashierIDStr := c.Param("cashier_id")
	cashierID, err := strconv.Atoi(cashierIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cashier_id", nil))
		return
	}

	statusFilter := c.Query("status")
	if statusFilter == "" {
		statusFilter = "all"
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var pgStart, pgEnd pgtype.Timestamp
	if startDateStr != "" {
		parsedStart, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			parsedStart, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid start_date format", nil))
				return
			}
		}
		pgStart = pgtype.Timestamp{Time: parsedStart, Valid: true}
	}

	if endDateStr != "" {
		parsedEnd, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			parsedEnd, err = time.Parse("2006-01-02", endDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid end_date format", nil))
				return
			}
		}
		pgEnd = pgtype.Timestamp{Time: parsedEnd, Valid: true}
	}

	arg := repository.GetCashierSessionsParams{
		CashierID:    int32(cashierID),
		StatusFilter: statusFilter,
		StartDate:    pgStart,
		EndDate:      pgEnd,
	}

	resp := h.useCase.GetCashierSessions(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}
