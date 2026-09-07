package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// PromotionHandler handles HTTP requests for the promotions and coupons module.
type PromotionHandler struct {
	useCase *usecase.PromotionUseCase
}

// NewPromotionHandler creates a new PromotionHandler.
func NewPromotionHandler(uc *usecase.PromotionUseCase) *PromotionHandler {
	return &PromotionHandler{useCase: uc}
}

func (h *PromotionHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// ─── Admin CRUD ────────────────────────────────────────────────────────────────

// CreatePromotion handles POST /api/promotions
// @Summary      Create promotion
// @Description  Create a new promotion or coupon (all types: percentage_discount, fixed_discount, buy_x_get_y, happy_hour, points_multiplier, bundle_price, free_item)
// @Tags         promotions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                 true  "Tenant identifier"
// @Param        Authorization header    string                 true  "Bearer token"
// @Param        body          body      CreatePromotionRequest true  "Promotion payload"
// @Success      201           {object}  PromotionResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/promotions [post]
func (h *PromotionHandler) CreatePromotion(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	actionMeta, _ := json.Marshal(req.ActionMetadata)
	scheduleMeta, _ := json.Marshal(req.ScheduleJson)
	metaBytes, _ := json.Marshal(req.Metadata)

	arg := repository.CreatePromotionParams{
		OrganizationID:    req.OrganizationID,
		Code:              req.Code,
		Name:              req.Name,
		Description:       promoTextOpt(req.Description),
		PromotionType:     req.PromotionType,
		ActionMetadata:    actionMeta,
		ScheduleJson:      scheduleMeta,
		AppliesTo:         promoTextOpt(req.AppliesTo),
		TargetProductIds:  req.TargetProductIds,
		TargetCategoryIds: req.TargetCategoryIds,
		CouponCode:        promoTextOpt(req.CouponCode),
		IsStackable:       promoBoolOpt(req.IsStackable),
		IsActive:          promoBoolOpt(req.IsActive),
		StoreIds:          req.StoreIds,
		Metadata:          metaBytes,
	}

	if req.CreatedBy != nil {
		arg.CreatedBy = pgtype.Int4{Int32: *req.CreatedBy, Valid: true}
	} else if userIDStr, ok := middleware.GetUserIDFromContext(c); ok {
		if uid, err := strconv.ParseInt(userIDStr, 10, 32); err == nil {
			arg.CreatedBy = pgtype.Int4{Int32: int32(uid), Valid: true}
		}
	}

	if req.MinOrderAmount != nil {
		n := pgtype.Numeric{}
		_ = n.Scan(*req.MinOrderAmount)
		arg.MinOrderAmount = n
	}
	if req.MinQuantity != nil {
		n := pgtype.Numeric{}
		_ = n.Scan(*req.MinQuantity)
		arg.MinQuantity = n
	}
	if req.DiscountValue != nil {
		n := pgtype.Numeric{}
		_ = n.Scan(*req.DiscountValue)
		arg.DiscountValue = n
	}
	if req.UsageLimit != nil {
		arg.UsageLimit = pgtype.Int4{Int32: *req.UsageLimit, Valid: true}
	}
	if req.UsagePerCustomer != nil {
		arg.UsagePerCustomer = pgtype.Int4{Int32: *req.UsagePerCustomer, Valid: true}
	}
	if req.ValidFrom != nil {
		vf, err := parsePromotionTimestamp(req.ValidFrom)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid valid_from: "+err.Error(), nil))
			return
		}
		arg.ValidFrom = vf
	}
	if req.ValidTo != nil {
		vt, err := parsePromotionTimestamp(req.ValidTo)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid valid_to: "+err.Error(), nil))
			return
		}
		arg.ValidTo = vt
	}

	resp := h.useCase.CreatePromotion(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// GetPromotion handles GET /api/promotions/:id
// @Summary      Get promotion by ID
// @Description  Retrieve a promotion by its numeric ID
// @Tags         promotions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Promotion ID"
// @Success      200           {object}  PromotionResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Router       /api/promotions/{id} [get]
func (h *PromotionHandler) GetPromotion(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id64, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid promotion id", nil))
		return
	}

	resp := h.useCase.GetPromotion(c.Request.Context(), int32(id64))
	c.JSON(resp.StatusCode, resp)
}

// GetPromotionByCode handles GET /api/promotions/code/:code
// @Summary      Get promotion by internal code
// @Description  Retrieve a promotion by its internal code and organization ID
// @Tags         promotions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true  "Tenant identifier"
// @Param        Authorization    header    string  true  "Bearer token"
// @Param        code             path      string  true  "Internal promotion code (e.g. PROMO-SUMMER20)"
// @Param        organization_id  query     int     true  "Organization ID"
// @Success      200              {object}  PromotionResponse
// @Failure      400              {object}  ErrorResponse
// @Failure      401              {object}  ErrorResponse
// @Failure      404              {object}  ErrorResponse
// @Router       /api/promotions/code/{code} [get]
func (h *PromotionHandler) GetPromotionByCode(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	code := c.Param("code")
	orgIDStr := c.Query("organization_id")
	orgID64, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid organization_id", nil))
		return
	}

	resp := h.useCase.GetPromotionByCode(c.Request.Context(), code, int32(orgID64))
	c.JSON(resp.StatusCode, resp)
}

// ListActivePromotions handles GET /api/promotions/active
// @Summary      List active promotions
// @Description  List all currently active promotions for an organisation
// @Tags         promotions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true  "Tenant identifier"
// @Param        Authorization    header    string  true  "Bearer token"
// @Param        organization_id  query     int     true  "Organization ID"
// @Success      200              {array}   PromotionResponse
// @Failure      400              {object}  ErrorResponse
// @Failure      401              {object}  ErrorResponse
// @Failure      500              {object}  ErrorResponse
// @Router       /api/promotions/active [get]
func (h *PromotionHandler) ListActivePromotions(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	orgID64, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid organization_id", nil))
		return
	}

	resp := h.useCase.ListActivePromotions(c.Request.Context(), int32(orgID64))
	c.JSON(resp.StatusCode, resp)
}

// ListAllPromotions handles GET /api/promotions
// @Summary      List all promotions (paginated)
// @Description  List all promotions (active and inactive) for an organisation with pagination
// @Tags         promotions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true  "Tenant identifier"
// @Param        Authorization    header    string  true  "Bearer token"
// @Param        organization_id  query     int     true  "Organization ID"
// @Param        limit            query     int     false "Limit (default 50)"
// @Param        offset           query     int     false "Offset (default 0)"
// @Success      200              {array}   PromotionResponse
// @Failure      400              {object}  ErrorResponse
// @Failure      401              {object}  ErrorResponse
// @Failure      500              {object}  ErrorResponse
// @Router       /api/promotions [get]
func (h *PromotionHandler) ListAllPromotions(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	orgID64, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid organization_id", nil))
		return
	}

	limit64, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 32)
	offset64, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 32)

	resp := h.useCase.ListAllPromotions(c.Request.Context(), repository.ListAllPromotionsParams{
		OrganizationID: int32(orgID64),
		Limit:          int32(limit64),
		Offset:         int32(offset64),
	})
	c.JSON(resp.StatusCode, resp)
}

// UpdatePromotion handles PUT /api/promotions/:id
// @Summary      Update promotion
// @Description  Update an existing promotion's details
// @Tags         promotions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                 true  "Tenant identifier"
// @Param        Authorization header    string                 true  "Bearer token"
// @Param        id            path      int                    true  "Promotion ID"
// @Param        body          body      UpdatePromotionRequest true  "Update payload"
// @Success      200           {object}  PromotionResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/promotions/{id} [put]
func (h *PromotionHandler) UpdatePromotion(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id64, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid promotion id", nil))
		return
	}

	var req UpdatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	actionMeta, _ := json.Marshal(req.ActionMetadata)
	scheduleMeta, _ := json.Marshal(req.ScheduleJson)

	arg := repository.UpdatePromotionParams{
		ID:                int32(id64),
		Name:              promoStrDeref(req.Name),
		Description:       promoTextOpt(req.Description),
		ActionMetadata:    actionMeta,
		ScheduleJson:      scheduleMeta,
		AppliesTo:         promoTextOpt(req.AppliesTo),
		TargetProductIds:  req.TargetProductIds,
		TargetCategoryIds: req.TargetCategoryIds,
		IsStackable:       promoBoolOpt(req.IsStackable),
		StoreIds:          req.StoreIds,
	}
	if req.MinOrderAmount != nil {
		n := pgtype.Numeric{}
		_ = n.Scan(*req.MinOrderAmount)
		arg.MinOrderAmount = n
	}
	if req.MinQuantity != nil {
		n := pgtype.Numeric{}
		_ = n.Scan(*req.MinQuantity)
		arg.MinQuantity = n
	}
	if req.DiscountValue != nil {
		n := pgtype.Numeric{}
		_ = n.Scan(*req.DiscountValue)
		arg.DiscountValue = n
	}
	if req.UsageLimit != nil {
		arg.UsageLimit = pgtype.Int4{Int32: *req.UsageLimit, Valid: true}
	}
	if req.UsagePerCustomer != nil {
		arg.UsagePerCustomer = pgtype.Int4{Int32: *req.UsagePerCustomer, Valid: true}
	}
	if req.ValidFrom != nil {
		vf, err := parsePromotionTimestamp(req.ValidFrom)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid valid_from: "+err.Error(), nil))
			return
		}
		arg.ValidFrom = vf
	}
	if req.ValidTo != nil {
		vt, err := parsePromotionTimestamp(req.ValidTo)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid valid_to: "+err.Error(), nil))
			return
		}
		arg.ValidTo = vt
	}

	resp := h.useCase.UpdatePromotion(c.Request.Context(), arg)
	c.JSON(resp.StatusCode, resp)
}

// UpdatePromotionStatus handles PATCH /api/promotions/:id/status
// @Summary      Toggle promotion status
// @Description  Activate or deactivate a promotion
// @Tags         promotions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                       true  "Tenant identifier"
// @Param        Authorization header    string                       true  "Bearer token"
// @Param        id            path      int                          true  "Promotion ID"
// @Param        body          body      UpdatePromotionStatusRequest true  "Status payload"
// @Success      200           {object}  PromotionResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/promotions/{id}/status [patch]
func (h *PromotionHandler) UpdatePromotionStatus(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id64, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid promotion id", nil))
		return
	}

	var req UpdatePromotionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	var isActive bool
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	resp := h.useCase.UpdatePromotionStatus(c.Request.Context(), repository.UpdatePromotionStatusParams{
		ID:       int32(id64),
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
	})
	c.JSON(resp.StatusCode, resp)
}

// DeletePromotion handles DELETE /api/promotions/:id
// @Summary      Delete promotion
// @Description  Permanently delete a promotion by ID
// @Tags         promotions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Promotion ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/promotions/{id} [delete]
func (h *PromotionHandler) DeletePromotion(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id64, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid promotion id", nil))
		return
	}

	resp := h.useCase.DeletePromotion(c.Request.Context(), int32(id64))
	c.JSON(resp.StatusCode, resp)
}

// ─── Coupon / Discount Application ───────────────────────────────────────────

// ApplyCoupon handles POST /api/promotions/apply-coupon
// @Summary      Apply coupon to cart
// @Description  Apply a coupon code to a cart. Validates constraints (min_order_amount, min_quantity, happy_hour schedule, buy_x_get_y thresholds) and applies the appropriate discount. Supported types: percentage_discount, fixed_discount, buy_x_get_y, happy_hour, points_multiplier, bundle_price, free_item.
// @Tags         promotions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                  true  "Tenant identifier"
// @Param        Authorization header    string                  true  "Bearer token"
// @Param        body          body      PromotionCouponRequest  true  "Coupon application payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/promotions/apply-coupon [post]
func (h *PromotionHandler) ApplyCoupon(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req PromotionCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	cartID, err := uuid.Parse(req.CartID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart_id", nil))
		return
	}

	resp := h.useCase.ApplyCoupon(c.Request.Context(), usecase.ApplyCouponInput{
		CartID:         cartID,
		CouponCode:     req.CouponCode,
		OrganizationID: req.OrganizationID,
	})
	c.JSON(resp.StatusCode, resp)
}

// ValidateCoupon handles POST /api/promotions/validate-coupon
// @Summary      Validate coupon (read-only)
// @Description  Check whether a coupon code is valid for a cart without applying it. Returns validation errors and the calculated discount amount.
// @Tags         promotions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                  true  "Tenant identifier"
// @Param        Authorization header    string                  true  "Bearer token"
// @Param        body          body      PromotionCouponRequest  true  "Coupon validation payload"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/promotions/validate-coupon [post]
func (h *PromotionHandler) ValidateCoupon(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req PromotionCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	cartID, err := uuid.Parse(req.CartID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid cart_id", nil))
		return
	}

	resp := h.useCase.ValidateCoupon(c.Request.Context(), usecase.ApplyCouponInput{
		CartID:         cartID,
		CouponCode:     req.CouponCode,
		OrganizationID: req.OrganizationID,
	})
	c.JSON(resp.StatusCode, resp)
}

// ─── Private helpers ──────────────────────────────────────────────────────────

func promoTextOpt(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func promoBoolOpt(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{Valid: false}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

func promoStrDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func parsePromotionTimestamp(s *string) (pgtype.Timestamp, error) {
	if s == nil || *s == "" {
		return pgtype.Timestamp{Valid: false}, nil
	}
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, fmtStr := range formats {
		if t, err := time.Parse(fmtStr, *s); err == nil {
			return pgtype.Timestamp{Time: t, Valid: true}, nil
		}
	}
	return pgtype.Timestamp{}, fmt.Errorf("invalid timestamp '%s'", *s)
}
