package handler

import (
	"net/http"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/gin-gonic/gin"
)

// BPPriceContractHandler handles HTTP requests for business partner price contracts.
type BPPriceContractHandler struct {
	useCase *usecase.BPPriceContractUseCase
}

// NewBPPriceContractHandler creates a new BPPriceContractHandler instance.
func NewBPPriceContractHandler(uc *usecase.BPPriceContractUseCase) *BPPriceContractHandler {
	return &BPPriceContractHandler{
		useCase: uc,
	}
}

func (h *BPPriceContractHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// CreateBPPriceContract handles POST /api/bp-price-contracts
// @Summary      Create a new business partner price contract
// @Description  Create a price contract for a business partner and product/variant
// @Tags         bp-price-contracts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                        true  "Tenant identifier"
// @Param        Authorization header    string                        true  "Bearer token"
// @Param        contract      body      CreateBPPriceContractRequest  true  "Price contract data"
// @Success      201           {object}  BPPriceContractResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/bp-price-contracts [post]
func (h *BPPriceContractHandler) CreateBPPriceContract(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req CreateBPPriceContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.CreateBPPriceContract(
		c.Request.Context(),
		usecase.CreateBPPriceContractInput{
			OrganizationID:     req.OrganizationID,
			PartnerID:          req.PartnerID,
			ProductID:          req.ProductID,
			ProductVariantID:   req.ProductVariantID,
			ContractPrice:      req.ContractPrice,
			DiscountPercentage: req.DiscountPercentage,
			MinQuantity:        req.MinQuantity,
			ValidFrom:          req.ValidFrom,
			ValidTo:            req.ValidTo,
			IsActive:           req.IsActive,
			Notes:              req.Notes,
		},
	)
	c.JSON(resp.StatusCode, resp)
}

// GetBPPriceContract handles GET /api/bp-price-contracts/:id
// @Summary      Get price contract by ID
// @Description  Retrieve a business partner price contract by its ID
// @Tags         bp-price-contracts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Price contract ID"
// @Success      200           {object}  BPPriceContractResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/bp-price-contracts/{id} [get]
func (h *BPPriceContractHandler) GetBPPriceContract(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	resp := h.useCase.GetBPPriceContract(c.Request.Context(), idStr)
	c.JSON(resp.StatusCode, resp)
}

// ListBPPriceContracts handles GET /api/bp-price-contracts
// @Summary      List price contracts
// @Description  List business partner price contracts filtered by organization, partner, product, or active state
// @Tags         bp-price-contracts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id      header    string  true   "Tenant identifier"
// @Param        Authorization    header    string  true   "Bearer token"
// @Param        organization_id  query     int     true   "Organization ID"
// @Param        partner_id       query     int     false  "Partner ID filter"
// @Param        product_id       query     int     false  "Product ID filter"
// @Param        is_active        query     bool    false  "Active status filter"
// @Success      200              {array}   BPPriceContractResponse
// @Failure      400              {object}  ErrorResponse
// @Failure      401              {object}  ErrorResponse
// @Failure      500              {object}  ErrorResponse
// @Router       /api/bp-price-contracts [get]
func (h *BPPriceContractHandler) ListBPPriceContracts(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	partnerIDStr := c.Query("partner_id")
	productIDStr := c.Query("product_id")
	isActiveStr := c.Query("is_active")

	resp := h.useCase.ListBPPriceContracts(c.Request.Context(), orgIDStr, partnerIDStr, productIDStr, isActiveStr)
	c.JSON(resp.StatusCode, resp)
}

// ListBPPriceContractsByPartner handles GET /api/bp-price-contracts/partner/:partner_id
// @Summary      List price contracts by partner
// @Description  List all price contracts for a specific business partner
// @Tags         bp-price-contracts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        partner_id    path      string  true  "Business partner ID"
// @Success      200           {array}   BPPriceContractResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/bp-price-contracts/partner/{partner_id} [get]
func (h *BPPriceContractHandler) ListBPPriceContractsByPartner(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	partnerIDStr := c.Param("partner_id")
	resp := h.useCase.ListBPPriceContractsByPartner(c.Request.Context(), partnerIDStr)
	c.JSON(resp.StatusCode, resp)
}

// GetEffectiveBPPriceContract handles GET /api/bp-price-contracts/effective
// @Summary      Get effective price contract
// @Description  Get the currently active contract price for a partner, product, and optional variant/quantity
// @Tags         bp-price-contracts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id          header    string  true   "Tenant identifier"
// @Param        Authorization        header    string  true   "Bearer token"
// @Param        partner_id           query     int     true   "Business partner ID"
// @Param        product_id           query     int     true   "Product ID"
// @Param        product_variant_id   query     int     false  "Product variant ID"
// @Param        quantity             query     number  false  "Quantity (default 1)"
// @Success      200                  {object}  BPPriceContractResponse
// @Failure      400                  {object}  ErrorResponse
// @Failure      401                  {object}  ErrorResponse
// @Failure      404                  {object}  ErrorResponse
// @Failure      500                  {object}  ErrorResponse
// @Router       /api/bp-price-contracts/effective [get]
func (h *BPPriceContractHandler) GetEffectiveBPPriceContract(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	partnerIDStr := c.Query("partner_id")
	productIDStr := c.Query("product_id")
	variantIDStr := c.Query("product_variant_id")
	quantityStr := c.Query("quantity")

	resp := h.useCase.GetEffectiveBPPriceContract(c.Request.Context(), partnerIDStr, productIDStr, variantIDStr, quantityStr)
	c.JSON(resp.StatusCode, resp)
}

// UpdateBPPriceContract handles PUT /api/bp-price-contracts/:id
// @Summary      Update price contract
// @Description  Update fields of an existing business partner price contract
// @Tags         bp-price-contracts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                        true  "Tenant identifier"
// @Param        Authorization header    string                        true  "Bearer token"
// @Param        id            path      string                        true  "Price contract ID"
// @Param        contract      body      UpdateBPPriceContractRequest  true  "Updated contract data"
// @Success      200           {object}  BPPriceContractResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/bp-price-contracts/{id} [put]
func (h *BPPriceContractHandler) UpdateBPPriceContract(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")

	var req UpdateBPPriceContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.UpdateBPPriceContract(
		c.Request.Context(),
		idStr,
		usecase.UpdateBPPriceContractInput{
			ContractPrice:      req.ContractPrice,
			DiscountPercentage: req.DiscountPercentage,
			MinQuantity:        req.MinQuantity,
			ValidFrom:          req.ValidFrom,
			ValidTo:            req.ValidTo,
			IsActive:           req.IsActive,
			Notes:              req.Notes,
		},
	)
	c.JSON(resp.StatusCode, resp)
}

// DeleteBPPriceContract handles DELETE /api/bp-price-contracts/:id
// @Summary      Delete price contract
// @Description  Delete a business partner price contract by ID
// @Tags         bp-price-contracts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      string  true  "Price contract ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/bp-price-contracts/{id} [delete]
func (h *BPPriceContractHandler) DeleteBPPriceContract(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	resp := h.useCase.DeleteBPPriceContract(c.Request.Context(), idStr)
	c.JSON(resp.StatusCode, resp)
}

// ToggleBPPriceContractActive handles PATCH /api/bp-price-contracts/:id/toggle
// @Summary      Toggle price contract active status
// @Description  Toggle the is_active status of a business partner price contract
// @Tags         bp-price-contracts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                              true  "Tenant identifier"
// @Param        Authorization header    string                              true  "Bearer token"
// @Param        id            path      string                              true  "Price contract ID"
// @Param        body          body      ToggleBPPriceContractActiveRequest  true  "Active state payload"
// @Success      200           {object}  BPPriceContractResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/bp-price-contracts/{id}/toggle [patch]
func (h *BPPriceContractHandler) ToggleBPPriceContractActive(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	var req ToggleBPPriceContractActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	var isActive bool
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	resp := h.useCase.ToggleBPPriceContractActive(c.Request.Context(), idStr, isActive)
	c.JSON(resp.StatusCode, resp)
}
