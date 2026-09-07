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
)

// AccountingHandler exposes the double-entry accounting engine over HTTP.
// All handlers are thin: they parse input, delegate to AccountingUseCase, and write the response.
type AccountingHandler struct {
	useCase *usecase.AccountingUseCase
}

func NewAccountingHandler(uc *usecase.AccountingUseCase) *AccountingHandler {
	return &AccountingHandler{useCase: uc}
}

func (h *AccountingHandler) getRepo(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError,
			utils.NewResponse(utils.CodeError, "repository not found in context", nil))
		c.Abort()
		return nil
	}
	return repo
}

// parseOrgID reads organization_id from the gin context (set by tenant middleware).
func parseOrgID(c *gin.Context) (int32, bool) {
	raw, exists := c.Get("organization_id")
	if !exists {
		c.JSON(http.StatusUnauthorized,
			utils.NewResponse(utils.CodeError, "organization_id missing from context", nil))
		return 0, false
	}
	switch v := raw.(type) {
	case int32:
		return v, true
	case int:
		return int32(v), true
	case float64:
		return int32(v), true
	}
	c.JSON(http.StatusInternalServerError,
		utils.NewResponse(utils.CodeError, "invalid organization_id type in context", nil))
	return 0, false
}

// =====================================================
// FISCAL YEARS
// =====================================================

// OpenFiscalYear handles POST /api/accounting/fiscal-years
// @Summary      Open a new fiscal year (creates 12 posting periods)
// @Tags         accounting
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header  string  true  "Tenant identifier"
// @Param        body          body    object  true  "{ year: int }"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /api/accounting/fiscal-years [post]
func (h *AccountingHandler) OpenFiscalYear(c *gin.Context) {
	orgID, ok := parseOrgID(c)
	if !ok {
		return
	}

	var req struct {
		Year int `json:"year" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "year is required", nil))
		return
	}
	if req.Year < 2000 || req.Year > 2100 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid year value", nil))
		return
	}

	resp := h.useCase.OpenFiscalYear(c.Request.Context(), orgID, req.Year, "SAR")
	c.JSON(resp.StatusCode, resp)
}

// ListFiscalYears handles GET /api/accounting/fiscal-years
// @Summary      List fiscal years for the organization
// @Tags         accounting
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header  string  true  "Tenant identifier"
// @Success      200  {array}   object
// @Router       /api/accounting/fiscal-years [get]
func (h *AccountingHandler) ListFiscalYears(c *gin.Context) {
	repo := h.getRepo(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgID, ok := parseOrgID(c)
	if !ok {
		return
	}

	resp := h.useCase.ListFiscalYears(c.Request.Context(), orgID)
	c.JSON(resp.StatusCode, resp)
}

// =====================================================
// POSTING PERIODS
// =====================================================

// ListPostingPeriods handles GET /api/accounting/posting-periods
// @Summary      List posting periods, optionally filtered by fiscal year
// @Tags         accounting
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header  string  true   "Tenant identifier"
// @Param        fiscal_year_id query   int     false  "Filter by fiscal year ID"
// @Success      200  {array}   object
// @Router       /api/accounting/posting-periods [get]
func (h *AccountingHandler) ListPostingPeriods(c *gin.Context) {
	repo := h.getRepo(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgID, ok := parseOrgID(c)
	if !ok {
		return
	}

	var fyID *int32
	if fyStr := c.Query("fiscal_year_id"); fyStr != "" {
		v, err := strconv.ParseInt(fyStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid fiscal_year_id", nil))
			return
		}
		v32 := int32(v)
		fyID = &v32
	}

	resp := h.useCase.ListPostingPeriods(c.Request.Context(), orgID, fyID)
	c.JSON(resp.StatusCode, resp)
}

// ClosePeriod handles PUT /api/accounting/posting-periods/:id/close
// @Summary      Close a posting period
// @Tags         accounting
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header  string  true  "Tenant identifier"
// @Param        id           path    int     true  "Period ID"
// @Success      200  {object}  SuccessResponse
// @Router       /api/accounting/posting-periods/{id}/close [put]
func (h *AccountingHandler) ClosePeriod(c *gin.Context) {
	orgID, ok := parseOrgID(c)
	if !ok {
		return
	}

	periodID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid period id", nil))
		return
	}

	// closedBy from JWT claims (set by auth middleware as user_id).
	closedBy := int32(0)
	if uid, exists := c.Get("user_id"); exists {
		switch v := uid.(type) {
		case int32:
			closedBy = v
		case float64:
			closedBy = int32(v)
		}
	}

	resp := h.useCase.RunPeriodClose(c.Request.Context(), orgID, int32(periodID), closedBy)
	c.JSON(resp.StatusCode, resp)
}

// =====================================================
// JOURNAL ENTRIES
// =====================================================

// ListJournalEntries handles GET /api/accounting/journal-entries
// @Summary      List journal entries with optional filters
// @Tags         accounting
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id    header  string  true   "Tenant identifier"
// @Param        period_id      query   int     false  "Posting period ID"
// @Param        from           query   string  false  "From date (YYYY-MM-DD)"
// @Param        to             query   string  false  "To date (YYYY-MM-DD)"
// @Param        status         query   string  false  "draft | posted | void | reversed"
// @Param        source_type    query   string  false  "Filter by source type"
// @Param        limit          query   int     false  "Page size (default 50)"
// @Param        offset         query   int     false  "Page offset (default 0)"
// @Success      200  {array}   object
// @Router       /api/accounting/journal-entries [get]
func (h *AccountingHandler) ListJournalEntries(c *gin.Context) {
	repo := h.getRepo(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgID, ok := parseOrgID(c)
	if !ok {
		return
	}

	var periodID *int32
	if pStr := c.Query("period_id"); pStr != "" {
		v, err := strconv.ParseInt(pStr, 10, 32)
		if err == nil {
			v32 := int32(v)
			periodID = &v32
		}
	}

	var from, to *time.Time
	if fStr := c.Query("from"); fStr != "" {
		if t, err := time.Parse("2006-01-02", fStr); err == nil {
			from = &t
		}
	}
	if tStr := c.Query("to"); tStr != "" {
		if t, err := time.Parse("2006-01-02", tStr); err == nil {
			to = &t
		}
	}

	var status, sourceType *string
	if s := c.Query("status"); s != "" {
		status = &s
	}
	if st := c.Query("source_type"); st != "" {
		sourceType = &st
	}

	limit := int32(50)
	offset := int32(0)
	if lStr := c.Query("limit"); lStr != "" {
		if v, err := strconv.ParseInt(lStr, 10, 32); err == nil {
			limit = int32(v)
		}
	}
	if oStr := c.Query("offset"); oStr != "" {
		if v, err := strconv.ParseInt(oStr, 10, 32); err == nil {
			offset = int32(v)
		}
	}

	resp := h.useCase.ListJournalEntries(c.Request.Context(), orgID,
		periodID, from, to, status, sourceType, limit, offset)
	c.JSON(resp.StatusCode, resp)
}

// ReverseJournalEntry handles POST /api/accounting/journal-entries/:id/reverse
// @Summary      Reverse a posted journal entry
// @Tags         accounting
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header  string  true  "Tenant identifier"
// @Param        id           path    int     true  "Journal entry ID"
// @Param        body         body    object  true  "{ reason: string, reversal_date: string }"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /api/accounting/journal-entries/{id}/reverse [post]
func (h *AccountingHandler) ReverseJournalEntry(c *gin.Context) {
	orgID, ok := parseOrgID(c)
	if !ok {
		return
	}

	entryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid journal entry id", nil))
		return
	}

	var req struct {
		Reason        string `json:"reason" binding:"required"`
		ReversalDate  string `json:"reversal_date"` // YYYY-MM-DD; defaults to today
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "reason is required", nil))
		return
	}

	reversalDate := time.Now()
	if req.ReversalDate != "" {
		if t, parseErr := time.Parse("2006-01-02", req.ReversalDate); parseErr == nil {
			reversalDate = t
		}
	}

	// Get reversedBy user from JWT context.
	var reversedBy *int32
	if uid, exists := c.Get("user_id"); exists {
		switch v := uid.(type) {
		case int32:
			reversedBy = &v
		case float64:
			v32 := int32(v)
			reversedBy = &v32
		}
	}

	// Reversal needs a transaction — obtain pool from usecase via context approach.
	// For now the usecase's pool is used internally. We call through the usecase directly.
	// NOTE: ReverseJournalEntry uses the internal pool; it does not require a handler-level tx.
	_ = orgID // orgID used inside usecase call below

	// We need to call within a transaction from the pool.
	// The usecase exposes ReverseJournalEntry which creates its own tx from the pool.
	// For the handler we create a thin wrapper that calls it.
	resp := h.useCase.ReverseJournalEntryFromHandler(c.Request.Context(), orgID, entryID,
		req.Reason, reversalDate, reversedBy)
	c.JSON(resp.StatusCode, resp)
}

// =====================================================
// FINANCIAL REPORTS
// =====================================================

// GetTrialBalance handles GET /api/accounting/trial-balance
// @Summary      Trial balance for a posting period
// @Tags         accounting
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header  string  true  "Tenant identifier"
// @Param        period_id    query   int     true  "Posting period ID"
// @Success      200  {array}   object
// @Router       /api/accounting/trial-balance [get]
func (h *AccountingHandler) GetTrialBalance(c *gin.Context) {
	repo := h.getRepo(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgID, ok := parseOrgID(c)
	if !ok {
		return
	}

	pStr := c.Query("period_id")
	if pStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "period_id is required", nil))
		return
	}
	periodID, err := strconv.ParseInt(pStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid period_id", nil))
		return
	}

	resp := h.useCase.GetTrialBalance(c.Request.Context(), orgID, int32(periodID))
	c.JSON(resp.StatusCode, resp)
}

// GetProfitAndLoss handles GET /api/accounting/profit-loss
// @Summary      Profit & Loss statement for a date range
// @Tags         accounting
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header  string  true  "Tenant identifier"
// @Param        from         query   string  true  "From date (YYYY-MM-DD)"
// @Param        to           query   string  true  "To date (YYYY-MM-DD)"
// @Success      200  {object}  object
// @Router       /api/accounting/profit-loss [get]
func (h *AccountingHandler) GetProfitAndLoss(c *gin.Context) {
	repo := h.getRepo(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgID, ok := parseOrgID(c)
	if !ok {
		return
	}

	from, err := time.Parse("2006-01-02", c.Query("from"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "from date is required (YYYY-MM-DD)", nil))
		return
	}
	to, err := time.Parse("2006-01-02", c.Query("to"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "to date is required (YYYY-MM-DD)", nil))
		return
	}

	resp := h.useCase.GetProfitAndLoss(c.Request.Context(), orgID, from, to)
	c.JSON(resp.StatusCode, resp)
}

// GetBalanceSheet handles GET /api/accounting/balance-sheet
// @Summary      Balance sheet as of end of a posting period
// @Tags         accounting
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header  string  true  "Tenant identifier"
// @Param        period_id    query   int     true  "Posting period ID"
// @Success      200  {array}   object
// @Router       /api/accounting/balance-sheet [get]
func (h *AccountingHandler) GetBalanceSheet(c *gin.Context) {
	repo := h.getRepo(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgID, ok := parseOrgID(c)
	if !ok {
		return
	}

	pStr := c.Query("period_id")
	if pStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "period_id is required", nil))
		return
	}
	periodID, err := strconv.ParseInt(pStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid period_id", nil))
		return
	}

	resp := h.useCase.GetBalanceSheet(c.Request.Context(), orgID, int32(periodID))
	c.JSON(resp.StatusCode, resp)
}

// GetAccountLedger handles GET /api/accounting/ledger/:account_id
// @Summary      Account ledger — transaction history for one GL account
// @Tags         accounting
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header  string  true  "Tenant identifier"
// @Param        account_id   path    int     true  "Chart of accounts ID"
// @Param        from         query   string  true  "From date (YYYY-MM-DD)"
// @Param        to           query   string  true  "To date (YYYY-MM-DD)"
// @Param        limit        query   int     false "Page size (default 100)"
// @Param        offset       query   int     false "Page offset (default 0)"
// @Success      200  {array}  object
// @Router       /api/accounting/ledger/{account_id} [get]
func (h *AccountingHandler) GetAccountLedger(c *gin.Context) {
	repo := h.getRepo(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgID, ok := parseOrgID(c)
	if !ok {
		return
	}

	accountID, err := strconv.ParseInt(c.Param("account_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid account_id", nil))
		return
	}

	from, err := time.Parse("2006-01-02", c.Query("from"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "from date required (YYYY-MM-DD)", nil))
		return
	}
	to, err := time.Parse("2006-01-02", c.Query("to"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "to date required (YYYY-MM-DD)", nil))
		return
	}

	limit := int32(100)
	offset := int32(0)
	if lStr := c.Query("limit"); lStr != "" {
		if v, parseErr := strconv.ParseInt(lStr, 10, 32); parseErr == nil {
			limit = int32(v)
		}
	}
	if oStr := c.Query("offset"); oStr != "" {
		if v, parseErr := strconv.ParseInt(oStr, 10, 32); parseErr == nil {
			offset = int32(v)
		}
	}

	resp := h.useCase.GetAccountLedger(c.Request.Context(), orgID, int32(accountID), from, to, limit, offset)
	c.JSON(resp.StatusCode, resp)
}

// =====================================================
// GL POSTING RULES
// =====================================================

// ListGLPostingRules handles GET /api/accounting/gl-posting-rules
// @Summary      List all GL posting rules for the organization
// @Tags         accounting
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header  string  true  "Tenant identifier"
// @Success      200  {array}  object
// @Router       /api/accounting/gl-posting-rules [get]
func (h *AccountingHandler) ListGLPostingRules(c *gin.Context) {
	repo := h.getRepo(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgID, ok := parseOrgID(c)
	if !ok {
		return
	}

	resp := h.useCase.ListGLPostingRules(c.Request.Context(), orgID)
	c.JSON(resp.StatusCode, resp)
}

// UpsertGLPostingRule handles PUT /api/accounting/gl-posting-rules/:posting_type
// @Summary      Create or update a GL posting rule
// @Tags         accounting
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header  string  true  "Tenant identifier"
// @Param        posting_type  path    string  true  "Posting type key"
// @Param        body          body    object  true  "GL posting rule fields"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /api/accounting/gl-posting-rules/{posting_type} [put]
func (h *AccountingHandler) UpsertGLPostingRule(c *gin.Context) {
	repo := h.getRepo(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgID, ok := parseOrgID(c)
	if !ok {
		return
	}

	postingType := c.Param("posting_type")
	if postingType == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "posting_type is required", nil))
		return
	}

	var req usecase.UpsertGLPostingRuleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}
	req.PostingType = postingType

	resp := h.useCase.UpsertGLPostingRule(c.Request.Context(), orgID, req)
	c.JSON(resp.StatusCode, resp)
}

// =====================================================
// DOCUMENT SERIES
// =====================================================

// ListDocumentSeries handles GET /api/accounting/document-series
func (h *AccountingHandler) ListDocumentSeries(c *gin.Context) {
	repo := h.getRepo(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgID, ok := parseOrgID(c)
	if !ok {
		return
	}

	var docType *string
	if dt := c.Query("doc_type"); dt != "" {
		docType = &dt
	}

	resp := h.useCase.ListDocumentSeries(c.Request.Context(), orgID, docType)
	c.JSON(resp.StatusCode, resp)
}

// =====================================================
// PROFIT CENTERS
// =====================================================

// ListProfitCenters handles GET /api/accounting/profit-centers
func (h *AccountingHandler) ListProfitCenters(c *gin.Context) {
	repo := h.getRepo(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgID, ok := parseOrgID(c)
	if !ok {
		return
	}

	resp := h.useCase.ListProfitCenters(c.Request.Context(), orgID)
	c.JSON(resp.StatusCode, resp)
}

// =====================================================
// WHT ENTRIES
// =====================================================

// ListWHTEntries handles GET /api/accounting/wht-entries
// @Summary      List withholding tax entries
// @Tags         accounting
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header  string  true   "Tenant identifier"
// @Param        from         query   string  false  "From date (YYYY-MM-DD)"
// @Param        to           query   string  false  "To date (YYYY-MM-DD)"
// @Param        is_remitted  query   bool    false  "Filter by remittance status"
// @Success      200  {array}  object
// @Router       /api/accounting/wht-entries [get]
func (h *AccountingHandler) ListWHTEntries(c *gin.Context) {
	repo := h.getRepo(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgID, ok := parseOrgID(c)
	if !ok {
		return
	}

	var from, to *time.Time
	if fStr := c.Query("from"); fStr != "" {
		if t, err := time.Parse("2006-01-02", fStr); err == nil {
			from = &t
		}
	}
	if tStr := c.Query("to"); tStr != "" {
		if t, err := time.Parse("2006-01-02", tStr); err == nil {
			to = &t
		}
	}

	var isRemitted *bool
	if irStr := c.Query("is_remitted"); irStr != "" {
		v := irStr == "true"
		isRemitted = &v
	}

	resp := h.useCase.ListWHTEntries(c.Request.Context(), orgID, from, to, isRemitted, 100, 0)
	c.JSON(resp.StatusCode, resp)
}
