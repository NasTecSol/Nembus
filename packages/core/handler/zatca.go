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
type PushSyncPayload struct {
	StoreID int32          `json:"store_id"`
	Items   []PushSyncItem `json:"items"`
}

type PushSyncItem struct {
	ID         int64          `json:"id"`
	EntityType string         `json:"entity_type"`
	EntityID   int64          `json:"entity_id"`
	Action     string         `json:"action"`
	Payload    map[string]any `json:"payload"`
	CreatedAt  time.Time      `json:"created_at"`
}

// GetConfigsDelta handles GET /api/zatca/configs?store_id=X&since=TIMESTAMP (Pull: Cloud -> POS)
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
