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

// LoyaltyHandler handles HTTP requests for loyalty redemption rules and customer loyalty points
type LoyaltyHandler struct {
	useCase *usecase.LoyaltyUseCase
}

func NewLoyaltyHandler(uc *usecase.LoyaltyUseCase) *LoyaltyHandler {
	return &LoyaltyHandler{useCase: uc}
}

func (h *LoyaltyHandler) getRepository(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// ── Loyalty Rules ────────────────────────────────────────────────────────────

// CreateLoyaltyRule handles POST /loyalty-rules
// @Summary      Create a loyalty redemption rule
// @Description  Create a new loyalty points earning/redemption rule for an organization
// @Tags         loyalty-rules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                   true  "Tenant identifier"
// @Param        Authorization header    string                   true  "Bearer token"
// @Param        rule          body      CreateLoyaltyRuleRequest true  "Loyalty rule data"
// @Success      201           {object}  LoyaltyRuleResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/loyalty-rules [post]
func (h *LoyaltyHandler) CreateLoyaltyRule(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateLoyaltyRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	in := usecase.CreateLoyaltyRuleInput{
		OrganizationID:       req.OrganizationID,
		RuleName:             req.RuleName,
		EligibleProductTypes: req.EligibleProductTypes,
		ExpiryDays:           req.ExpiryDays,
		IsActive:             req.IsActive,
		ValidFrom:            req.ValidFrom,
		ValidTo:              req.ValidTo,
	}

	if req.PointsEarningRate != nil {
		n, err := numericFromString(*req.PointsEarningRate)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid points_earning_rate", nil))
			return
		}
		in.PointsEarningRate = &n
	}
	if req.PointsRedemptionRate != nil {
		n, err := numericFromString(*req.PointsRedemptionRate)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid points_redemption_rate", nil))
			return
		}
		in.PointsRedemptionRate = &n
	}
	if req.MinPointsToRedeem != nil {
		n, err := numericFromString(*req.MinPointsToRedeem)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid min_points_to_redeem", nil))
			return
		}
		in.MinPointsToRedeem = &n
	}
	if req.MaxPointsPerTxn != nil {
		n, err := numericFromString(*req.MaxPointsPerTxn)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid max_points_per_txn", nil))
			return
		}
		in.MaxPointsPerTxn = &n
	}
	if req.MaxRedemptionPercent != nil {
		n, err := numericFromString(*req.MaxRedemptionPercent)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid max_redemption_percent", nil))
			return
		}
		in.MaxRedemptionPercent = &n
	}
	metaBytes, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}
	in.Metadata = metaBytes

	resp := h.useCase.CreateLoyaltyRule(c.Request.Context(), in)
	c.JSON(resp.StatusCode, resp)
}

// ListLoyaltyRules handles GET /loyalty-rules
// @Summary      List loyalty rules
// @Description  Retrieve all loyalty rules for an organization
// @Tags         loyalty-rules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true  "Tenant identifier"
// @Param        Authorization    header    string  true  "Bearer token"
// @Param        organization_id  query     int     true  "Organization ID"
// @Success      200              {array}   LoyaltyRuleResponse
// @Failure      400              {object}  ErrorResponse
// @Failure      401              {object}  ErrorResponse
// @Router       /api/loyalty-rules [get]
func (h *LoyaltyHandler) ListLoyaltyRules(c *gin.Context) {
	repo := h.getRepository(c)
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
	resp := h.useCase.ListLoyaltyRules(c.Request.Context(), int32(orgID))
	c.JSON(resp.StatusCode, resp)
}

// GetActiveLoyaltyRule handles GET /loyalty-rules/active
// @Summary      Get active loyalty rule
// @Description  Get the currently active loyalty rule for an organization (used at POS checkout)
// @Tags         loyalty-rules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true  "Tenant identifier"
// @Param        Authorization    header    string  true  "Bearer token"
// @Param        organization_id  query     int     true  "Organization ID"
// @Success      200              {object}  LoyaltyRuleResponse
// @Failure      400              {object}  ErrorResponse
// @Failure      401              {object}  ErrorResponse
// @Failure      404              {object}  ErrorResponse
// @Router       /api/loyalty-rules/active [get]
func (h *LoyaltyHandler) GetActiveLoyaltyRule(c *gin.Context) {
	repo := h.getRepository(c)
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
	resp := h.useCase.GetActiveLoyaltyRule(c.Request.Context(), int32(orgID))
	c.JSON(resp.StatusCode, resp)
}

// GetLoyaltyRule handles GET /loyalty-rules/:id
// @Summary      Get loyalty rule by ID
// @Description  Retrieve a specific loyalty rule
// @Tags         loyalty-rules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Rule ID"
// @Success      200           {object}  LoyaltyRuleResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/loyalty-rules/{id} [get]
func (h *LoyaltyHandler) GetLoyaltyRule(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)
	resp := h.useCase.GetLoyaltyRule(c.Request.Context(), c.Param("id"))
	c.JSON(resp.StatusCode, resp)
}

// UpdateLoyaltyRule handles PUT /loyalty-rules/:id
// @Summary      Update loyalty rule
// @Description  Update a loyalty redemption rule
// @Tags         loyalty-rules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                   true  "Tenant identifier"
// @Param        Authorization header    string                   true  "Bearer token"
// @Param        id            path      int                      true  "Rule ID"
// @Param        rule          body      UpdateLoyaltyRuleRequest true  "Rule update data"
// @Success      200           {object}  LoyaltyRuleResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/loyalty-rules/{id} [put]
func (h *LoyaltyHandler) UpdateLoyaltyRule(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req UpdateLoyaltyRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	in := usecase.UpdateLoyaltyRuleInput{
		RuleName:  req.RuleName,
		IsActive:  req.IsActive,
		ValidFrom: req.ValidFrom,
		ValidTo:   req.ValidTo,
	}
	parseAndAssign := func(s *string, dst **pgtype.Numeric, field string) bool {
		if s == nil {
			return true
		}
		n, err := numericFromString(*s)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid "+field, nil))
			return false
		}
		*dst = &n
		return true
	}
	if !parseAndAssign(req.PointsEarningRate, &in.PointsEarningRate, "points_earning_rate") {
		return
	}
	if !parseAndAssign(req.PointsRedemptionRate, &in.PointsRedemptionRate, "points_redemption_rate") {
		return
	}
	if !parseAndAssign(req.MinPointsToRedeem, &in.MinPointsToRedeem, "min_points_to_redeem") {
		return
	}
	if !parseAndAssign(req.MaxPointsPerTxn, &in.MaxPointsPerTxn, "max_points_per_txn") {
		return
	}
	if !parseAndAssign(req.MaxRedemptionPercent, &in.MaxRedemptionPercent, "max_redemption_percent") {
		return
	}
	metaBytes, err := bytesFromMap(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid metadata", nil))
		return
	}
	in.Metadata = metaBytes

	resp := h.useCase.UpdateLoyaltyRule(c.Request.Context(), c.Param("id"), in)
	c.JSON(resp.StatusCode, resp)
}

// ToggleLoyaltyRuleActive handles PATCH /loyalty-rules/:id/active
// @Summary      Toggle loyalty rule active status
// @Description  Activate or deactivate a loyalty rule
// @Tags         loyalty-rules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                          true  "Tenant identifier"
// @Param        Authorization header    string                          true  "Bearer token"
// @Param        id            path      int                             true  "Rule ID"
// @Param        payload       body      ToggleLoyaltyRuleActiveRequest  true  "Active flag"
// @Success      200           {object}  LoyaltyRuleResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/loyalty-rules/{id}/active [patch]
func (h *LoyaltyHandler) ToggleLoyaltyRuleActive(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req ToggleLoyaltyRuleActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}
	resp := h.useCase.ToggleLoyaltyRuleActive(c.Request.Context(), c.Param("id"), req.IsActive)
	c.JSON(resp.StatusCode, resp)
}

// DeleteLoyaltyRule handles DELETE /loyalty-rules/:id
// @Summary      Delete loyalty rule
// @Description  Permanently delete a loyalty rule
// @Tags         loyalty-rules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Rule ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/loyalty-rules/{id} [delete]
func (h *LoyaltyHandler) DeleteLoyaltyRule(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)
	resp := h.useCase.DeleteLoyaltyRule(c.Request.Context(), c.Param("id"))
	c.JSON(resp.StatusCode, resp)
}

// ── Customer Loyalty Points ───────────────────────────────────────────────────

// GetCustomerLoyaltyBalance handles GET /customers/:id/loyalty-balance
// @Summary      Get customer loyalty balance
// @Description  Lightweight fetch of customer points balance for POS validation
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Customer ID"
// @Success      200           {object}  CustomerLoyaltyBalanceResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/customers/{id}/loyalty-balance [get]
func (h *LoyaltyHandler) GetCustomerLoyaltyBalance(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)
	resp := h.useCase.GetCustomerLoyaltyBalance(c.Request.Context(), c.Param("id"))
	c.JSON(resp.StatusCode, resp)
}

// AdjustCustomerLoyaltyPoints handles PATCH /customers/:id/loyalty-points
// @Summary      Adjust customer loyalty points
// @Description  Add or deduct loyalty points for a customer (pass negative value to deduct/redeem)
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                            true  "Tenant identifier"
// @Param        Authorization header    string                            true  "Bearer token"
// @Param        id            path      int                               true  "Customer ID"
// @Param        payload       body      AdjustLoyaltyPointsRequest        true  "Points adjustment (use negative value for redemption)"
// @Success      200           {object}  CustomerResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/customers/{id}/loyalty-points [patch]
func (h *LoyaltyHandler) AdjustCustomerLoyaltyPoints(c *gin.Context) {
	repo := h.getRepository(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req AdjustLoyaltyPointsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	points, err := numericFromString(req.Points)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid points value", nil))
		return
	}

	resp := h.useCase.AdjustCustomerLoyaltyPoints(c.Request.Context(), c.Param("id"), points)
	c.JSON(resp.StatusCode, resp)
}
