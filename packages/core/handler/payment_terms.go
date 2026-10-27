package handler

import (
	"net/http"
	"strconv"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/gin-gonic/gin"
)

type PaymentTermsHandler struct {
	useCase *usecase.PaymentTermsUseCase
}

func NewPaymentTermsHandler(uc *usecase.PaymentTermsUseCase) *PaymentTermsHandler {
	return &PaymentTermsHandler{
		useCase: uc,
	}
}

func (h *PaymentTermsHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreatePaymentTerm handles POST /api/payment-terms
// @Summary      Create a new payment term
// @Description  Create a new payment term with code, name, and due days
// @Tags         payment-terms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        body      body      CreatePaymentTermRequest  true  "Payment term payload"
// @Success      201  {object}  PaymentTermResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/payment-terms [post]
func (h *PaymentTermsHandler) CreatePaymentTerm(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreatePaymentTermRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body: "+err.Error(), nil))
		return
	}

	resp := h.useCase.CreatePaymentTerm(c.Request.Context(), usecase.CreatePaymentTermInput{
		OrganizationID:     req.OrganizationID,
		Code:               req.Code,
		Name:               req.Name,
		DueDays:            req.DueDays,
		DiscountDays:       req.DiscountDays,
		DiscountPercentage: req.DiscountPercentage,
		LateFeePercentage:  req.LateFeePercentage,
		IsActive:           req.IsActive,
	})

	c.JSON(resp.StatusCode, resp)
}

// GetPaymentTerm handles GET /api/payment-terms/:id
// @Summary      Get payment term by ID
// @Description  Retrieve a specific payment term by its ID
// @Tags         payment-terms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        id            path      string  true  "Payment Term ID"
// @Success      200  {object}  PaymentTermResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/payment-terms/{id} [get]
func (h *PaymentTermsHandler) GetPaymentTerm(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.GetPaymentTerm(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}

// ListPaymentTerms handles GET /api/payment-terms
// @Summary      List payment terms
// @Description  Retrieve a list of payment terms for an organization
// @Tags         payment-terms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        organization_id  query     int     true  "Organization ID"
// @Param        limit         query     int     false "Limit results"
// @Param        offset        query     int     false "Offset results"
// @Param        is_active     query     bool    false "Filter by active status"
// @Success      200  {array}   PaymentTermResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /api/payment-terms [get]
func (h *PaymentTermsHandler) ListPaymentTerms(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id query parameter is required", nil))
		return
	}

	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid organization_id", nil))
		return
	}

	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")
	isActiveStr := c.Query("is_active")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		limit = 100
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil {
		offset = 0
	}

	var isActive *bool
	if isActiveStr != "" {
		active, err := strconv.ParseBool(isActiveStr)
		if err == nil {
			isActive = &active
		}
	}

	resp := h.useCase.ListPaymentTerms(c.Request.Context(), int32(orgID), int32(limit), int32(offset), isActive)
	c.JSON(resp.StatusCode, resp)
}

// UpdatePaymentTerm handles PUT /api/payment-terms/:id
// @Summary      Update payment term
// @Description  Update details of an existing payment term
// @Tags         payment-terms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        id            path      string  true  "Payment Term ID"
// @Param        body          body      UpdatePaymentTermRequest  true  "Payment term payload"
// @Success      200  {object}  PaymentTermResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      440  {object}  ErrorResponse
// @Router       /api/payment-terms/{id} [put]
func (h *PaymentTermsHandler) UpdatePaymentTerm(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")

	var req UpdatePaymentTermRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body: "+err.Error(), nil))
		return
	}

	resp := h.useCase.UpdatePaymentTerm(c.Request.Context(), id, usecase.UpdatePaymentTermInput{
		Code:               req.Code,
		Name:               req.Name,
		DueDays:            req.DueDays,
		DiscountDays:       req.DiscountDays,
		DiscountPercentage: req.DiscountPercentage,
		LateFeePercentage:  req.LateFeePercentage,
		IsActive:           req.IsActive,
	})

	c.JSON(resp.StatusCode, resp)
}

// DeletePaymentTerm handles DELETE /api/payment-terms/:id
// @Summary      Delete payment term
// @Description  Delete a payment term by its ID
// @Tags         payment-terms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header    string  true  "Tenant identifier"
// @Param        Authorization  header    string  true  "Bearer token"
// @Param        id            path      string  true  "Payment Term ID"
// @Success      200  {object}  SuccessResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /api/payment-terms/{id} [delete]
func (h *PaymentTermsHandler) DeletePaymentTerm(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	id := c.Param("id")
	resp := h.useCase.DeletePaymentTerm(c.Request.Context(), id)
	c.JSON(resp.StatusCode, resp)
}
