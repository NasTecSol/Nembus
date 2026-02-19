package handler

import (
	"net/http"

	"NEMBUS/internal/middleware"
	"NEMBUS/internal/repository"
	"NEMBUS/internal/usecase"
	"NEMBUS/utils"

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
// @Description  Creates a return record and updates drawer balance
// @Tags         pos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        body          body      usecase.ProcessReturnInput true "Return payload"
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

	var req usecase.ProcessReturnInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, err.Error(), nil))
		return
	}

	resp := h.useCase.ProcessSalesReturn(c.Request.Context(), req)
	c.JSON(resp.StatusCode, resp)
}
