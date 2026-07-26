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

type TransferRequestsHandler struct {
	useCase *usecase.TransferRequestsUseCase
}

func NewTransferRequestsHandler(uc *usecase.TransferRequestsUseCase) *TransferRequestsHandler {
	return &TransferRequestsHandler{
		useCase: uc,
	}
}

func (h *TransferRequestsHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.NewResponse(utils.CodeError, "repository not found in context", nil))
		c.Abort()
		return nil
	}
	return repo
}

// CreateTransferRequest handles POST /api/transfer-requests
// @Summary      Create transfer request
// @Description  Create a new inter-store/warehouse transfer request in draft status
// @Tags         transfer-requests
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string                              true   "Tenant identifier"
// @Param        Authorization header    string                              true   "Bearer token"
// @Param        body          body      usecase.CreateTransferRequestInput  true   "Transfer request payload"
// @Success      200           {object}  usecase.TransferRequestOutput
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/transfer-requests [post]
func (h *TransferRequestsHandler) CreateTransferRequest(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req usecase.CreateTransferRequestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body", nil))
		return
	}

	resp := h.useCase.CreateTransferRequest(c.Request.Context(), req)
	c.JSON(resp.StatusCode, resp)
}

// GetTransferRequest handles GET /api/transfer-requests/:id
// @Summary      Get transfer request by ID
// @Description  Get detailed transfer request with itemized lines
// @Tags         transfer-requests
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Transfer Request ID"
// @Success      200           {object}  usecase.TransferRequestOutput
// @Failure      400           {object}  ErrorResponse
// @Failure      404           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/transfer-requests/{id} [get]
func (h *TransferRequestsHandler) GetTransferRequest(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid transfer request id", nil))
		return
	}

	resp := h.useCase.GetTransferRequest(c.Request.Context(), int32(id))
	c.JSON(resp.StatusCode, resp)
}

// ListTransferRequests handles GET /api/transfer-requests
// @Summary      List transfer requests
// @Description  List transfer requests filtered by organization ID
// @Tags         transfer-requests
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id     header    string  true   "Tenant identifier"
// @Param        Authorization   header    string  true   "Bearer token"
// @Param        organization_id query     int     true   "Organization ID"
// @Param        limit           query     int     false  "Limit" default(50)
// @Param        offset          query     int     false  "Offset" default(0)
// @Success      200             {array}   usecase.TransferRequestOutput
// @Failure      400             {object}  ErrorResponse
// @Failure      500             {object}  ErrorResponse
// @Router       /api/transfer-requests [get]
func (h *TransferRequestsHandler) ListTransferRequests(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	orgIDStr := c.Query("organization_id")
	orgID, err := strconv.ParseInt(orgIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid organization_id", nil))
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.ParseInt(limitStr, 10, 32)
	offset, _ := strconv.ParseInt(offsetStr, 10, 32)

	resp := h.useCase.ListTransferRequestsByOrganization(c.Request.Context(), int32(orgID), int32(limit), int32(offset))
	c.JSON(resp.StatusCode, resp)
}

// ApproveTransferRequest handles POST /api/transfer-requests/:id/approve
// @Summary      Approve transfer request
// @Description  Approve transfer request lifecycle state transition (Draft/Pending -> Approved)
// @Tags         transfer-requests
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Transfer Request ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/transfer-requests/{id}/approve [post]
func (h *TransferRequestsHandler) ApproveTransferRequest(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid transfer request id", nil))
		return
	}

	var req struct {
		ApprovedBy int32 `json:"approved_by"`
	}
	_ = c.ShouldBindJSON(&req)

	resp := h.useCase.ApproveTransferRequest(c.Request.Context(), int32(id), req.ApprovedBy)
	c.JSON(resp.StatusCode, resp)
}

// ShipTransferRequest handles POST /api/transfer-requests/:id/ship
// @Summary      Ship / dispatch transfer request
// @Description  Deducts stock at source store, places stock in-transit at target store, transitions status to Shipped
// @Tags         transfer-requests
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Transfer Request ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/transfer-requests/{id}/ship [post]
func (h *TransferRequestsHandler) ShipTransferRequest(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid transfer request id", nil))
		return
	}

	var req struct {
		ShippedBy int32 `json:"shipped_by"`
	}
	_ = c.ShouldBindJSON(&req)

	resp := h.useCase.ShipTransferRequest(c.Request.Context(), int32(id), req.ShippedBy)
	c.JSON(resp.StatusCode, resp)
}

// ReceiveTransferRequest handles POST /api/transfer-requests/:id/receive
// @Summary      Receive transfer request
// @Description  Clears in-transit stock and adds to available stock at target store, transitions status to Received
// @Tags         transfer-requests
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Param        id            path      int     true  "Transfer Request ID"
// @Success      200           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/transfer-requests/{id}/receive [post]
func (h *TransferRequestsHandler) ReceiveTransferRequest(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid transfer request id", nil))
		return
	}

	var req struct {
		ReceivedBy int32 `json:"received_by"`
	}
	_ = c.ShouldBindJSON(&req)

	resp := h.useCase.ReceiveTransferRequest(c.Request.Context(), int32(id), req.ReceivedBy)
	c.JSON(resp.StatusCode, resp)
}
