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
// @Param        x-tenant-id   header    string                                true   "Tenant identifier"
// @Param        Authorization header    string                                true   "Bearer token"
// @Param        body          body      usecase.CreateGoodsReceiptNoteInput   true   "GRN creation payload"
// @Success      200           {object}  usecase.GoodsReceiptNoteOutput
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
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request body", nil))
		return
	}

	resp := h.useCase.CreateGoodsReceiptNote(c.Request.Context(), req)
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
// @Success      200           {object}  usecase.GoodsReceiptNoteOutput
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
