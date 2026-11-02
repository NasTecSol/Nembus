package handler

import (
	"errors"
	"strconv"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetSalesOrderDocumentStatus reads the saved result without issuing or reporting an invoice.
// @Summary      Read a sales order's saved ZATCA status and response
// @Description  Read-only lookup of zatca_document_chain for a sales_orders_v2 UUID. Returns zatca_status (pending, reported, cleared, warning, rejected, or failed), the saved zatca_response, QR, document identity and timestamps. A pending document may have no gateway response. qr_code_base64 is the saved document QR and does not by itself prove reporting success or that ZATCA returned it. Never generates, checks, retries or submits an invoice. Also resolves historical linked POS documents and remains available when ZATCA is disabled. Returns 404 if no saved document exists.
// @Tags         zatca
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        id path string true "sales_orders_v2.id UUID (not order number or POS transaction ID)"
// @Success      200 {object} ZatcaDocumentAPIResponse
// @Failure      400 {object} ZatcaAPIError
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ZatcaAPIError
// @Failure      404 {object} ZatcaAPIError
// @Failure      500 {object} ZatcaAPIError
// @Router       /api/zatca/orders/{id}/status [get]
func (h *ZatcaHandler) GetSalesOrderDocumentStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, utils.NewResponse(400, "invalid sales order UUID", nil))
		return
	}
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok || repo == nil {
		c.JSON(500, utils.NewResponse(500, "tenant repository is required", nil))
		return
	}
	ctx := c.Request.Context()
	entry, err := repo.GetChainEntryByEntity(ctx, repository.GetChainEntryByEntityParams{
		EntityType: usecase.ZatcaSalesOrderEntity, EntityID: id.String(),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		var transactionID int32
		transactionID, err = repo.GetLegacyZatcaOrderTransaction(ctx, id)
		if err == nil {
			entry, err = repo.GetChainEntryByEntity(ctx, repository.GetChainEntryByEntityParams{
				EntityType: "pos_transaction", EntityID: strconv.Itoa(int(transactionID)),
			})
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(404, utils.NewResponse(404, "saved ZATCA document not found for this sales order", nil))
		return
	}
	if err != nil {
		c.JSON(500, utils.NewResponse(500, "failed to load saved ZATCA document", nil))
		return
	}
	if orgID, hasOrg := middleware.GetOrganizationIDFromContext(c); hasOrg && orgID > 0 && orgID != entry.OrganizationID {
		c.JSON(403, utils.NewResponse(403, "document organization does not match authenticated organization", nil))
		return
	}
	c.JSON(200, utils.NewResponse(200, "Saved sales order ZATCA status and response", entry))
}
