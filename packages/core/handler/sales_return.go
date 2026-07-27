package handler

import (
	"net/http"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"

	"github.com/gin-gonic/gin"
)

type SalesReturnHandler struct {
	useCase *usecase.SalesReturnUseCase
}

func NewSalesReturnHandler(uc *usecase.SalesReturnUseCase) *SalesReturnHandler {
	return &SalesReturnHandler{useCase: uc}
}

func (h *SalesReturnHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// ProcessReturn handles POST /api/pos/returns
// @Summary      Process a sales return
// @Description  Creates a return record, decrements drawer expected_balance by refund amount, and optionally returns stock
// @Tags         sales-returns
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        body          body      ProcessReturnRequest true "Return payload"
// @Success      201           {object}  SuccessResponse
// @Failure      400           {object}  ErrorResponse
// @Failure      401           {object}  ErrorResponse
// @Failure      500           {object}  ErrorResponse
// @Router       /api/pos/returns [post]
func (h *SalesReturnHandler) ProcessReturn(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req ProcessReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	lines := make([]usecase.ProcessReturnLineInput, len(req.Lines))
	for i, l := range req.Lines {
		lines[i] = usecase.ProcessReturnLineInput{
			ProductID:        l.ProductID,
			ProductVariantID: l.ProductVariantID,
			OriginalLineID:   l.OriginalLineID,
			Quantity:         l.Quantity,
			UnitPrice:        l.UnitPrice,
			RefundAmount:     l.RefundAmount,
			ReturnToStock:    l.ReturnToStock,
			SerialNumber:     l.SerialNumber,
			BatchNumber:      l.BatchNumber,
			Condition:        l.Condition,
		}
	}
	in := usecase.ProcessReturnInput{
		StoreID:               req.StoreID,
		CashierID:             req.CashierID,
		SessionID:             req.SessionID,
		OriginalTransactionID: req.OriginalTransactionID,
		CustomerID:            req.CustomerID,
		ReturnReason:          req.ReturnReason,
		Subtotal:              req.Subtotal,
		TaxAmount:             req.TaxAmount,
		TotalRefundAmount:     req.TotalRefundAmount,
		RefundMethod:          req.RefundMethod,
		RefundReference:       req.RefundReference,
		Lines:                 lines,
	}

	resp := h.useCase.ProcessSalesReturn(c.Request.Context(), in)
	c.JSON(resp.StatusCode, resp)
}
