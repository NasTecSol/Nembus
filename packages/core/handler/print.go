package handler

import (
	"net/http"

	"github.com/NasTecSol/nembus-client/utils"
	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"

	"github.com/gin-gonic/gin"
)

// PrintHandler handles ESC/POS receipt printing requests.
type PrintHandler struct {
	useCase *usecase.PrintUseCase
}

// NewPrintHandler creates a new PrintHandler.
func NewPrintHandler(uc *usecase.PrintUseCase) *PrintHandler {
	return &PrintHandler{useCase: uc}
}

func (h *PrintHandler) getRepositoryFromContext(c *gin.Context) *repository.Queries {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "repository not found in context"})
		c.Abort()
		return nil
	}
	return repo
}

// PrintReceipt handles POST /api/print/receipt
// @Summary      Print a receipt
// @Description  Builds and prints an ESC/POS receipt. For a ZATCA-enabled organization, the saved sale determines its store/terminal registration. Unregistered POS sales print normally with qr_status=not_applicable. Reported invoices retain their saved signed QR; registered sales with pending or failed reporting print without a QR so checkout is not blocked. Caller-supplied qr_code_base64 is ignored. The response includes qr_status.
// @Tags         print
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id  header  string                    true  "Tenant identifier"
// @Param        Authorization header string                    true  "Bearer token"
// @Param        body         body    PrintReceiptRequest       true  "Print request payload"
// @Success      200          {object} SuccessResponse
// @Failure      400          {object} ErrorResponse
// @Failure      401          {object} ErrorResponse
// @Failure      404          {object} ErrorResponse
// @Failure      500          {object} ErrorResponse
// @Router       /api/print/receipt [post]
func (h *PrintHandler) PrintReceipt(c *gin.Context) {
	repo := h.getRepositoryFromContext(c)
	if repo == nil {
		return
	}
	h.useCase.SetRepository(repo)

	var req usecase.PrintReceiptInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid request: "+err.Error(), nil))
		return
	}

	if req.OrgID <= 0 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "org_id is required", nil))
		return
	}

	resp := h.useCase.PrintReceipt(c.Request.Context(), req)
	c.JSON(resp.StatusCode, resp)
}
