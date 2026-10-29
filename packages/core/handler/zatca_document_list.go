package handler

import (
	"strconv"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/gin-gonic/gin"
)

// ListZatcaDocuments returns saved documents without contacting ZATCA.
// @Summary      List saved ZATCA documents with pagination
// @Description  Returns all saved document types and statuses, newest document ID first, from the current tenant and authenticated organization. Each document includes its saved status, gateway response, QR, identity and timestamps. Defaults to page=1 and limit=50; larger positive limits are accepted. An empty or out-of-range page returns documents=[] with HTTP 200. This read-only endpoint never generates or submits documents and works when ZATCA is disabled. The saved QR does not by itself prove successful reporting or that ZATCA returned it.
// @Tags         zatca
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        page query int false "Page number (starts at 1)" default(1) minimum(1) maximum(2147483647)
// @Param        limit query int false "Records per page; values above 50 are supported" default(50) minimum(1) maximum(2147483647)
// @Success      200 {object} ZatcaDocumentListAPIResponse
// @Failure      400 {object} ZatcaAPIError
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ZatcaAPIError
// @Router       /api/zatca/documents [get]
func (h *ZatcaHandler) ListZatcaDocuments(c *gin.Context) {
	page, pageErr := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	limit, limitErr := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 32)
	if pageErr != nil || limitErr != nil || page < 1 || limit < 1 {
		c.JSON(400, utils.NewResponse(400, "page and limit must be positive integers no greater than 2147483647", nil))
		return
	}
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok || repo == nil {
		c.JSON(500, utils.NewResponse(500, "tenant repository is required", nil))
		return
	}
	orgID, _ := middleware.GetOrganizationIDFromContext(c)
	documents, total, err := repo.ListZatcaDocumentsPage(c.Request.Context(), orgID, int32(limit), (page-1)*limit)
	if err != nil {
		c.JSON(500, utils.NewResponse(500, "failed to load saved ZATCA documents", nil))
		return
	}
	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}
	c.JSON(200, utils.NewResponse(200, "Saved ZATCA documents", gin.H{
		"documents": documents, "page": page, "limit": limit,
		"total": total, "total_pages": totalPages, "has_next_page": page < totalPages,
	}))
}
