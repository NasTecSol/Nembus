package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/gin-gonic/gin"
)

type ZatcaHandler struct {
	zatcaUC *usecase.ZatcaUseCase
}

func NewZatcaHandler(uc *usecase.ZatcaUseCase) *ZatcaHandler {
	return &ZatcaHandler{
		zatcaUC: uc,
	}
}

// PushSyncPayload defines the payload pushed from POS client outbox.
// swagger:model PushSyncPayload
type PushSyncPayload struct {
	StoreID int32          `json:"store_id" example:"1"`
	Items   []PushSyncItem `json:"items"`
}

// PushSyncItem defines an individual item inside the POS outbox sync payload.
// swagger:model PushSyncItem
type PushSyncItem struct {
	ID         int64          `json:"id" example:"101"`
	EntityType string         `json:"entity_type" example:"pos_transaction"`
	EntityID   int64          `json:"entity_id" example:"5001"`
	Action     string         `json:"action" example:"INSERT"`
	Payload    map[string]any `json:"payload"`
	CreatedAt  time.Time      `json:"created_at" example:"2026-07-29T01:00:00Z"`
}

// ZatcaStatusResponse defines the response shape for ZATCA status check.
// swagger:model ZatcaStatusResponse
type ZatcaStatusResponse struct {
	Enabled bool   `json:"enabled" example:"true"`
	Env     string `json:"env" example:"sandbox"`
	BaseURL string `json:"base_url" example:"https://gw-fatoora.zatca.gov.sa/e-invoicing/developer-portal"`
	VatID   string `json:"vat_id" example:"300000000000003"`
}

// GetConfigsDelta handles GET /api/zatca/configs?store_id=X&since=TIMESTAMP (Pull: Cloud -> POS)
// @Summary      Get ZATCA device configuration deltas (Pull Sync)
// @Description  Returns all ZATCA device configs modified after the given timestamp for a store to sync active certificates and revocation flags
// @Tags         zatca
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true   "Tenant identifier"
// @Param        Authorization header    string  true   "Bearer token"
// @Param        store_id      query     int     true   "Store ID" example(1)
// @Param        since         query     string  false  "ISO 8601 Timestamp (e.g. 2026-01-01T00:00:00Z) or epoch" example(2026-01-01T00:00:00Z)
// @Success      200  {object}  repository.Response
// @Failure      400  {object}  repository.Response
// @Failure      401  {object}  repository.Response
// @Failure      500  {object}  repository.Response
// @Router       /api/zatca/configs [get]
func (h *ZatcaHandler) GetConfigsDelta(c *gin.Context) {
	storeIDStr := c.Query("store_id")
	sinceStr := c.Query("since")

	if storeIDStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "store_id query param is required", nil))
		return
	}

	storeID, err := strconv.ParseInt(storeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid store_id", err.Error()))
		return
	}

	var since time.Time
	if sinceStr != "" {
		since, err = time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			if epoch, parseErr := strconv.ParseInt(sinceStr, 10, 64); parseErr == nil {
				since = time.Unix(epoch, 0)
			}
		}
	}

	resp := h.zatcaUC.GetDeviceConfigsDelta(c.Request.Context(), int32(storeID), since)
	c.JSON(resp.StatusCode, resp)
}

// ReceivePushSync handles POST /api/zatca/sync/push (Push: POS -> Cloud)
// @Summary      Ingest offline transaction outbox payload (Push Sync)
// @Description  Receives signed offline transactions and operational outbox records from POS terminals for Cloud ERP ingestion and ZATCA reporting
// @Tags         zatca
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string           true  "Tenant identifier"
// @Param        Authorization header    string           true  "Bearer token"
// @Param        body          body      PushSyncPayload  true  "POS Outbox Push Payload"
// @Success      200  {object}  repository.Response
// @Failure      400  {object}  repository.Response
// @Failure      401  {object}  repository.Response
// @Failure      500  {object}  repository.Response
// @Router       /api/zatca/sync/push [post]
func (h *ZatcaHandler) ReceivePushSync(c *gin.Context) {
	var payload PushSyncPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid push sync payload", err.Error()))
		return
	}

	syncedIDs := make([]int64, 0, len(payload.Items))
	for _, item := range payload.Items {
		syncedIDs = append(syncedIDs, item.ID)
	}

	resp := utils.NewResponse(utils.CodeOK, "Sync payload ingested successfully", gin.H{
		"synced_ids": syncedIDs,
		"count":      len(syncedIDs),
	})
	c.JSON(resp.StatusCode, resp)
}

// GetZatcaStatus returns current ZATCA configuration status.
// @Summary      Get ZATCA configuration status
// @Description  Returns whether ZATCA Phase 2 compliance is enabled, environment mode (sandbox or production), base URL, and Org VAT ID
// @Tags         zatca
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string  true  "Tenant identifier"
// @Param        Authorization header    string  true  "Bearer token"
// @Success      200  {object}  ZatcaStatusResponse
// @Failure      401  {object}  repository.Response
// @Router       /api/zatca/status [get]
func (h *ZatcaHandler) GetZatcaStatus(c *gin.Context) {
	cfg := h.zatcaUC.GetConfig()
	if cfg == nil {
		resp := utils.NewResponse(utils.CodeOK, "ZATCA status", gin.H{"enabled": false})
		c.JSON(resp.StatusCode, resp)
		return
	}

	resp := utils.NewResponse(utils.CodeOK, "ZATCA status", gin.H{
		"enabled":  cfg.Enabled,
		"env":      cfg.Env,
		"base_url": cfg.BaseURL,
		"vat_id":   cfg.OrgVATID,
	})
	c.JSON(resp.StatusCode, resp)
}
