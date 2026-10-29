package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/NasTecSol/nembus-core/middleware"
	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/usecase"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// DeviceOnboarder abstracts device onboarding functionality across cloud server.
type DeviceOnboarder interface {
	OnboardDevice(ctx context.Context, orgID int32, storeID *int32, posTerminalID *int32, serialNumber string, deviceType string, otp string, environment string) (*repository.ZatcaDeviceConfig, error)
}

type PosZatcaSubmitter interface {
	SubmitPosTransaction(context.Context, int32) (*repository.ZatcaDocumentChain, error)
	CheckPosTransaction(context.Context, int32) (json.RawMessage, error)
}

type SalesOrderZatcaSubmitter interface {
	CheckSalesOrder(context.Context, uuid.UUID) (json.RawMessage, error)
	SubmitSalesOrder(context.Context, uuid.UUID) (*repository.ZatcaDocumentChain, error)
}

// @Summary      Generate or reuse the signed sales-order invoice
// @Description  Reads persisted sales_orders_v2 and its complete lines. No extra completion status or request body is required. Signs once and saves XML/hash/UUID/QR locally; does not call ZATCA. Requires organization metadata.zatca.enabled=true. Reuses historical linked transaction invoices to prevent duplicate issuance.
// @Tags         zatca
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        id path string true "sales_orders_v2.id UUID (not order number or POS transaction ID)"
// @Success      200 {object} ZatcaDocumentAPIResponse
// @Failure      400 {object} ZatcaAPIError
// @Failure      409 {object} ZatcaAPIError
// @Failure      422 {object} ZatcaAPIError
// @Failure      500 {object} ZatcaAPIError
// @Failure      401 {object} ErrorResponse
// @Router       /api/zatca/orders/{id}/generate [post]
func (h *ZatcaHandler) GenerateSalesOrder(c *gin.Context) {
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
	uc := usecase.NewZatcaUseCase(h.zatcaUC.GetConfig())
	uc.SetRepository(repo)
	entry, err := uc.EnsureSalesOrderDocument(c.Request.Context(), id)
	if err != nil {
		status := 422
		if errors.Is(err, usecase.ErrZatcaDisabled) {
			status = 409
		}
		c.JSON(status, utils.NewResponse(status, err.Error(), nil))
		return
	}
	c.JSON(200, utils.NewResponse(200, "Sales order ZATCA document ready; invoice has not been submitted by this call", entry))
}

// @Summary      Check a sales-order invoice with ZATCA
// @Description  Generates the saved order document if needed, then calls compliance/invoices with its original device compliance credentials/environment. This does not report the invoice or change its reporting status. Fully onboarded devices may return Submitted before; use submit for actual reporting. No body.
// @Tags         zatca
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        id path string true "Sales order UUID"
// @Success      200 {object} ZatcaCheckAPIResponse
// @Failure      400 {object} ZatcaAPIError
// @Failure      409 {object} ZatcaAPIError
// @Failure      501 {object} ZatcaAPIError
// @Failure      502 {object} ZatcaAPIError
// @Failure      401 {object} ErrorResponse
// @Router       /api/zatca/orders/{id}/check [post]
func (h *ZatcaHandler) CheckSalesOrder(c *gin.Context) { h.salesOrderGateway(c, false) }

// @Summary      Report a sales-order invoice to ZATCA
// @Description  Generates the order document if missing, then reports the saved XML using its device reporting CSID, matching secret and saved zatca_env. Repeated calls reuse the saved identity; already-reported invoices are not sent again. Requires organization enablement. No body.
// @Tags         zatca
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        id path string true "Sales order UUID"
// @Success      200 {object} ZatcaDocumentAPIResponse
// @Failure      400 {object} ZatcaAPIError
// @Failure      409 {object} ZatcaAPIError
// @Failure      501 {object} ZatcaAPIError
// @Failure      502 {object} ZatcaAPIError
// @Failure      401 {object} ErrorResponse
// @Router       /api/zatca/orders/{id}/submit [post]
func (h *ZatcaHandler) SubmitSalesOrder(c *gin.Context) { h.salesOrderGateway(c, true) }
func (h *ZatcaHandler) salesOrderGateway(c *gin.Context, submit bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, utils.NewResponse(400, "invalid sales order UUID", nil))
		return
	}
	if submit {
		service := h.orderReporter
		if service == nil {
			service, _ = h.submitter.(OrderZatcaReporter)
		}
		if service == nil {
			c.JSON(501, utils.NewResponse(501, "order reporting is unavailable on this backend", nil))
			return
		}
		entry, err := service.SubmitSalesOrder(c.Request.Context(), id)
		if err != nil {
			status := zatcaGatewayStatus(err)
			c.JSON(status, utils.NewResponse(status, err.Error(), entry))
			return
		}
		c.JSON(200, utils.NewResponse(200, "ZATCA sales order invoice reported successfully", entry))
		return
	}
	service, ok := h.submitter.(SalesOrderZatcaSubmitter)
	if !ok {
		c.JSON(501, utils.NewResponse(501, "order compliance checking is available on the cloud-server backend", nil))
		return
	}
	response, err := service.CheckSalesOrder(c.Request.Context(), id)
	if err != nil {
		status := zatcaGatewayStatus(err)
		c.JSON(status, utils.NewResponse(status, err.Error(), response))
		return
	}
	c.JSON(200, utils.NewResponse(200, "ZATCA order compliance check response; this call does not report the invoice", response))
}

type ZatcaHandler struct {
	zatcaUC       *usecase.ZatcaUseCase
	onboarder     DeviceOnboarder
	submitter     PosZatcaSubmitter
	orderReporter OrderZatcaReporter
}

func (h *ZatcaHandler) SetOrderReporter(reporter OrderZatcaReporter) {
	h.orderReporter = reporter
}

func NewZatcaHandler(uc *usecase.ZatcaUseCase) *ZatcaHandler {
	return &ZatcaHandler{
		zatcaUC: uc,
	}
}

func (h *ZatcaHandler) SetOnboarder(onboarder DeviceOnboarder) {
	h.onboarder = onboarder
}

func (h *ZatcaHandler) SetSubmitter(submitter PosZatcaSubmitter) {
	h.submitter = submitter
}

// @Summary      Resume incomplete device onboarding
// @Description  Recovery only. Reuses saved compliance credentials, request ID, key and device zatca_env, completes six checks, and obtains/saves reporting credentials. No OTP or body. Already-ready devices return without another gateway call. Normal onboarding now completes this flow within POST /onboard.
// @Tags         zatca
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        id path int true "zatca_device_configs.id (not terminal ID)"
// @Success      200 {object} ZatcaCompletionAPIResponse
// @Failure      400 {object} ZatcaAPIError
// @Failure      409 {object} ZatcaAPIError
// @Failure      501 {object} ZatcaAPIError
// @Failure      502 {object} ZatcaAPIError
// @Failure      401 {object} ErrorResponse
// @Router       /api/zatca/devices/{id}/complete-onboarding [post]
func (h *ZatcaHandler) CompleteDeviceOnboarding(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid device configuration id", nil))
		return
	}
	if repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries); ok && repo != nil {
		device, err := repo.GetZatcaDeviceConfig(c.Request.Context(), int32(id))
		if err != nil {
			c.JSON(http.StatusNotFound, utils.NewResponse(http.StatusNotFound, "device not found", nil))
			return
		}
		if caller, hasOrg := middleware.GetOrganizationIDFromContext(c); hasOrg && caller > 0 && caller != device.OrganizationID {
			c.JSON(http.StatusForbidden, utils.NewResponse(http.StatusForbidden, "device organization does not match authenticated organization", nil))
			return
		}
	}
	completer, ok := h.onboarder.(DeviceOnboardingCompleter)
	if !ok {
		c.JSON(http.StatusNotImplemented, utils.NewResponse(http.StatusNotImplemented, "device onboarding completion is available on the cloud-server backend", nil))
		return
	}
	response, err := completer.CompleteDeviceOnboardingResponse(c.Request.Context(), int32(id))
	if err != nil {
		status := zatcaGatewayStatus(err)
		c.JSON(status, utils.NewResponse(status, err.Error(), response))
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(utils.CodeOK, "Device onboarding complete; reporting credentials saved", response))
}

type DeviceOnboardingCompleter interface {
	CompleteDeviceOnboardingResponse(context.Context, int32) (json.RawMessage, error)
}

// @Summary      Read the saved invoice QR payload
// @Description  Provide sales_order_id or transaction_id. Returns the exact saved Base64 TLV text, not a PNG. This read-only operation never generates or submits an invoice and remains available when the organization switch is disabled.
// @Tags         zatca
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        sales_order_id query string false "Sales order UUID; provide this or transaction_id"
// @Param        transaction_id query int false "POS transaction ID; takes precedence if both are supplied"
// @Success      200 {object} ZatcaReceiptQRAPIResponse
// @Failure      400 {object} ZatcaAPIError
// @Failure      404 {object} ZatcaAPIError
// @Failure      500 {object} ZatcaAPIError
// @Failure      401 {object} ErrorResponse
// @Router       /api/zatca/receipt-qr [get]
func (h *ZatcaHandler) GetReceiptQR(c *gin.Context) {
	repo, ok := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	if !ok || repo == nil {
		c.JSON(500, utils.NewResponse(500, "tenant repository is required", nil))
		return
	}
	transactionID := int64(0)
	orderID := c.Query("sales_order_id")
	if raw := c.Query("transaction_id"); raw != "" {
		var err error
		transactionID, err = strconv.ParseInt(raw, 10, 32)
		if err != nil || transactionID <= 0 {
			c.JSON(400, utils.NewResponse(400, "invalid transaction_id", nil))
			return
		}
	} else if _, err := uuid.Parse(orderID); err != nil {
		c.JSON(400, utils.NewResponse(400, "transaction_id or a valid sales_order_id is required", nil))
		return
	}
	var qr string
	var err error
	if transactionID == 0 {
		qr, err = repo.GetZatcaSalesOrderQR(c.Request.Context(), orderID)
	} else {
		qr, err = repo.GetZatcaReceiptQR(c.Request.Context(), int32(transactionID), "")
	}
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(404, utils.NewResponse(404, "receipt transaction not found", nil))
		return
	}
	if err != nil {
		c.JSON(500, utils.NewResponse(500, "failed to load receipt QR", nil))
		return
	}
	if qr == "" {
		c.JSON(404, utils.NewResponse(404, "saved receipt QR not found for this order or transaction", nil))
		return
	}
	c.JSON(200, utils.NewResponse(200, "Saved receipt QR", gin.H{"qr_code_base64": qr}))
}

// @Summary      Check an existing historical POS invoice
// @Description  Legacy transaction-based lookup. Uses a previously saved pos_transaction document; new sales-order invoices must use /orders/{id}/check. Calls ZATCA compliance only, without reporting or changing local reporting status. No body.
// @Tags         zatca
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        id path int true "Historical POS transaction ID"
// @Success      200 {object} ZatcaCheckAPIResponse
// @Failure      400 {object} ZatcaAPIError
// @Failure      409 {object} ZatcaAPIError
// @Failure      501 {object} ZatcaAPIError
// @Failure      502 {object} ZatcaAPIError
// @Deprecated
// @Failure      401 {object} ErrorResponse
// @Router       /api/zatca/transactions/{id}/check [post]
func (h *ZatcaHandler) CheckPosTransaction(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid transaction id", nil))
		return
	}
	if h.submitter == nil {
		c.JSON(http.StatusNotImplemented, utils.NewResponse(http.StatusNotImplemented, "ZATCA compliance checking is available on the cloud-server backend", nil))
		return
	}
	response, err := h.submitter.CheckPosTransaction(c.Request.Context(), int32(id))
	if err != nil {
		status := zatcaGatewayStatus(err)
		c.JSON(status, utils.NewResponse(status, err.Error(), response))
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(utils.CodeOK, "ZATCA compliance check response received; inspect validationResults; invoice has not been reported", response))
}

// @Summary      Report an existing historical POS invoice
// @Description  Legacy transaction-based lookup. Reports saved XML with its original device credentials/environment; does not create an invoice. Use /orders/{id}/submit for new sales. Already-reported documents are not sent again. No body.
// @Tags         zatca
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        id path int true "Historical POS transaction ID"
// @Success      200 {object} ZatcaDocumentAPIResponse
// @Failure      400 {object} ZatcaAPIError
// @Failure      409 {object} ZatcaAPIError
// @Failure      501 {object} ZatcaAPIError
// @Failure      502 {object} ZatcaAPIError
// @Deprecated
// @Failure      401 {object} ErrorResponse
// @Router       /api/zatca/transactions/{id}/submit [post]
func (h *ZatcaHandler) SubmitPosTransaction(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid transaction id", nil))
		return
	}
	if h.submitter == nil {
		c.JSON(http.StatusNotImplemented, utils.NewResponse(http.StatusNotImplemented, "ZATCA submission is available on the cloud-server backend", nil))
		return
	}
	entry, err := h.submitter.SubmitPosTransaction(c.Request.Context(), int32(id))
	if err != nil {
		status := zatcaGatewayStatus(err)
		c.JSON(status, utils.NewResponse(status, err.Error(), entry))
		return
	}
	c.JSON(http.StatusOK, utils.NewResponse(utils.CodeOK, "ZATCA document reported successfully", entry))
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
	Env     string `json:"env,omitempty" example:"simulation"`
	BaseURL string `json:"base_url,omitempty" example:"https://gw-fatoora.zatca.gov.sa/e-invoicing/developer-portal"`
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
// @Success      200  {object}  ZatcaGenericAPIResponse
// @Failure      400  {object}  ZatcaAPIError
// @Failure      401  {object}  ZatcaAPIError
// @Failure      500  {object}  ZatcaAPIError
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
// @Description  Acknowledges received outbox item IDs only. This route does not generate or submit invoices to ZATCA.
// @Tags         zatca
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id   header    string           true  "Tenant identifier"
// @Param        Authorization header    string           true  "Bearer token"
// @Param        body          body      PushSyncPayload  true  "POS Outbox Push Payload"
// @Success      200  {object}  ZatcaGenericAPIResponse
// @Failure      400  {object}  ZatcaAPIError
// @Failure      401  {object}  ZatcaAPIError
// @Failure      500  {object}  ZatcaAPIError
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

// @Summary      Read organization enablement and device gateway status
// @Description  Enablement comes from organizations.metadata.zatca.enabled, not ZATCA_ENABLED. Supply device_id to include its saved environment/gateway, or organization_id to read the organization switch. With no scope and no organization in authentication, enabled is null and organization_required is true.
// @Tags         zatca
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        device_id query int false "Saved device configuration ID; determines organization and gateway"
// @Param        organization_id query int false "Organization ID when device_id is omitted"
// @Success      200 {object} ZatcaStatusAPIResponse
// @Failure      400 {object} ZatcaAPIError
// @Failure      403 {object} ZatcaAPIError
// @Failure      404 {object} ZatcaAPIError
// @Failure      422 {object} ZatcaAPIError
// @Failure      500 {object} ZatcaAPIError
// @Failure      401 {object} ErrorResponse
// @Router       /api/zatca/status [get]
func (h *ZatcaHandler) GetZatcaStatus(c *gin.Context) {
	data := gin.H{"enabled": nil, "enabled_source": "organization_metadata", "environment_source": "device_config"}
	repo, _ := c.Request.Context().Value(middleware.RepoKey).(*repository.Queries)
	var orgID int32
	var device *repository.ZatcaDeviceConfig
	if raw := c.Query("device_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || id <= 0 {
			c.JSON(400, utils.NewResponse(400, "invalid device_id", nil))
			return
		}
		if repo == nil {
			c.JSON(500, utils.NewResponse(500, "tenant repository is required", nil))
			return
		}
		config, err := repo.GetZatcaDeviceConfig(c.Request.Context(), int32(id))
		if err != nil {
			c.JSON(404, utils.NewResponse(404, "device not found", nil))
			return
		}
		device = &config
		orgID = config.OrganizationID
		environment, err := usecase.ZatcaDeviceEnvironment(device)
		if err != nil {
			c.JSON(422, utils.NewResponse(422, err.Error(), nil))
			return
		}
		baseURL, _ := usecase.ZatcaEnvironmentBaseURL(environment)
		data["device_id"], data["env"], data["base_url"] = device.ID, environment, baseURL
	} else if raw := c.Query("organization_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || id <= 0 {
			c.JSON(400, utils.NewResponse(400, "invalid organization_id", nil))
			return
		}
		orgID = int32(id)
	} else if id, ok := middleware.GetOrganizationIDFromContext(c); ok {
		orgID = id
	}
	if caller, ok := middleware.GetOrganizationIDFromContext(c); ok && caller > 0 && orgID > 0 && caller != orgID {
		c.JSON(403, utils.NewResponse(403, "organization does not match authenticated organization", nil))
		return
	}
	if orgID > 0 {
		if repo == nil {
			c.JSON(500, utils.NewResponse(500, "tenant repository is required", nil))
			return
		}
		org, err := repo.GetOrganization(c.Request.Context(), orgID)
		if err != nil {
			c.JSON(404, utils.NewResponse(404, "organization not found", nil))
			return
		}
		enabled, err := usecase.OrganizationZatcaEnabled(&org)
		if err != nil {
			c.JSON(422, utils.NewResponse(422, err.Error(), nil))
			return
		}
		data["organization_id"], data["enabled"], data["vat_id"] = orgID, enabled, org.TaxID.String
		if device != nil {
			var credentials struct {
				Secret string `json:"production_secret"`
			}
			_ = json.Unmarshal(device.Metadata, &credentials)
			data["reporting_ready"] = enabled && device.IsActive.Bool && !device.IsRevoked.Bool && device.ProductionCsid.String != "" && device.ProductionCsid.String != device.ComplianceCsid.String && credentials.Secret != "" && (!device.CsidExpiry.Valid || device.CsidExpiry.Time.After(time.Now()))
			data["is_active"] = device.IsActive.Bool && !device.IsRevoked.Bool
		}
	} else {
		data["organization_required"] = true
	}
	c.JSON(200, utils.NewResponse(utils.CodeOK, "ZATCA status", data))
}

func zatcaGatewayStatus(err error) int {
	if errors.Is(err, usecase.ErrZatcaDisabled) {
		return http.StatusConflict
	}
	return http.StatusBadGateway
}

// ZatcaOnboardRequest defines payload to onboard a device with ZATCA Phase 2.
// swagger:model ZatcaOnboardRequest
type ZatcaOnboardRequest struct {
	OrganizationID *int32 `json:"organization_id"`
	StoreID        *int32 `json:"store_id"`
	PosTerminalID  *int32 `json:"pos_terminal_id"`
	DeviceSerial   string `json:"device_serial" binding:"required"`
	DeviceType     string `json:"device_type" enums:"pos,cloud" example:"pos"` // "pos" or "cloud", defaults to "pos"
	OTP            string `json:"otp" binding:"required" example:"OTP_FROM_MATCHING_PORTAL"`
	ZatcaEnv       string `json:"zatca_env" binding:"required,oneof=sandbox simulation production" enums:"sandbox,simulation,production" example:"simulation"`
}

// @Summary      Onboard a device completely in one call
// @Description  Requires organization metadata.zatca.enabled=true and explicit zatca_env. Uses an OTP from the matching portal, creates/saves compliance credentials, runs six checks, then saves the reporting CSID and matching secret. HTTP 200 means reporting_ready=true. Allow up to 240 seconds; partial failures include saved device ID and completion_endpoint for recovery without a new OTP. Environment-file flags do not select enablement or gateway.
// @Tags         zatca
// @Produce      json
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Accept       json
// @Param        body body ZatcaOnboardRequest true "Device onboarding payload; zatca_env is required"
// @Success      200 {object} ZatcaOnboardAPIResponse
// @Failure      400 {object} ZatcaAPIError
// @Failure      409 {object} ZatcaAPIError
// @Failure      500 {object} ZatcaAPIError
// @Failure      501 {object} ZatcaAPIError
// @Failure      502 {object} ZatcaAPIError
// @Failure      401 {object} ErrorResponse
// @Router       /api/zatca/onboard [post]
func (h *ZatcaHandler) OnboardDevice(c *gin.Context) {
	var req ZatcaOnboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid onboarding payload", err.Error()))
		return
	}

	var orgID int32
	if req.OrganizationID != nil && *req.OrganizationID > 0 {
		orgID = *req.OrganizationID
	} else if ctxOrgID, ok := middleware.GetOrganizationIDFromContext(c); ok && ctxOrgID > 0 {
		orgID = ctxOrgID
	} else {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "organization_id is required", nil))
		return
	}
	if caller, ok := middleware.GetOrganizationIDFromContext(c); ok && caller > 0 && caller != orgID {
		c.JSON(http.StatusForbidden, utils.NewResponse(http.StatusForbidden, "organization does not match authenticated organization", nil))
		return
	}

	deviceType := req.DeviceType
	if deviceType == "" {
		deviceType = "pos"
	}

	if h.onboarder != nil {
		config, err := h.onboarder.OnboardDevice(c.Request.Context(), orgID, req.StoreID, req.PosTerminalID, req.DeviceSerial, deviceType, req.OTP, req.ZatcaEnv)
		if err != nil {
			var progress interface{ OnboardingFailureDetails() json.RawMessage }
			if errors.As(err, &progress) {
				status := zatcaGatewayStatus(err)
				var responseStatus interface{ OnboardingHTTPStatus() int }
				if errors.As(err, &responseStatus) {
					status = responseStatus.OnboardingHTTPStatus()
				}
				c.JSON(status, utils.NewResponse(status, "ZATCA onboarding incomplete: "+err.Error(), progress.OnboardingFailureDetails()))
				return
			}
			status := http.StatusInternalServerError
			if errors.Is(err, usecase.ErrZatcaDisabled) {
				status = 409
			}
			c.JSON(status, utils.NewResponse(status, "ZATCA onboarding failed: "+err.Error(), nil))
			return
		}

		reportingReady := config.ProductionCsid.Valid && config.ProductionCsid.String != "" && config.ProductionCsid.String != config.ComplianceCsid.String
		stage := "compliance"
		message := "Compliance credentials saved; compliance checks and production onboarding are incomplete"
		if reportingReady {
			stage = "production"
			message = "Device onboarded with ZATCA successfully"
		}
		if !reportingReady {
			c.JSON(http.StatusBadGateway, utils.NewResponse(http.StatusBadGateway, "ZATCA onboarding incomplete; reporting credentials were not obtained", gin.H{"id": config.ID, "onboarding_stage": stage, "reporting_ready": false, "completion_endpoint": "/api/zatca/devices/" + strconv.Itoa(int(config.ID)) + "/complete-onboarding"}))
			return
		}
		c.JSON(http.StatusOK, utils.NewResponse(utils.CodeOK, message, gin.H{
			"completion_endpoint": "/api/zatca/devices/" + strconv.Itoa(int(config.ID)) + "/complete-onboarding",
			"onboarding_stage":    stage,
			"reporting_ready":     reportingReady,
			"id":                  config.ID,
			"device_serial":       config.DeviceSerial,
			"device_type":         config.DeviceType,
			"compliance_csid":     config.ComplianceCsid.String,
			"production_csid":     config.ProductionCsid.String,
			"csid_expiry":         config.CsidExpiry.Time,
			"zatca_env":           config.ZatcaEnv.String,
			"is_active":           config.IsActive.Bool,
		}))
		return
	}

	c.JSON(http.StatusNotImplemented, utils.NewResponse(http.StatusNotImplemented, "Device onboarding requires the cloud-server gateway service", nil))
}

// GetTerminalZatcaConfig handles GET /api/zatca/terminals/:id
// @Summary      Get active ZATCA device config for a POS terminal
// @Description  Legacy terminal configuration lookup. For organization enablement and the authoritative saved device environment, use GET /api/zatca/status?device_id=ID.
// @Tags         zatca
// @Produce      json
// @Security     BearerAuth
// @Security     BearerAuth
// @Param        x-tenant-id header string true "Tenant identifier"
// @Param        id path int true "POS Terminal ID"
// @Success      200 {object} ZatcaGenericAPIResponse
// @Failure      404 {object} ZatcaAPIError
// @Failure      401 {object} ErrorResponse
// @Router       /api/zatca/terminals/{id} [get]
func (h *ZatcaHandler) GetTerminalZatcaConfig(c *gin.Context) {
	terminalIDStr := c.Param("id")
	terminalID, err := strconv.ParseInt(terminalIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewResponse(utils.CodeBadReq, "invalid terminal id", err.Error()))
		return
	}

	orgID, ok := middleware.GetOrganizationIDFromContext(c)
	if !ok || orgID <= 0 {
		orgID = 1 // Default fallback for local single-tenant store
	}

	if h.zatcaUC != nil {
		// Try to query repo if set
		configs := h.zatcaUC.GetDeviceConfigsDelta(c.Request.Context(), int32(terminalID), time.Time{})
		if configs != nil && configs.StatusCode == utils.CodeOK {
			c.JSON(http.StatusOK, configs)
			return
		}
	}

	c.JSON(http.StatusOK, utils.NewResponse(utils.CodeOK, "Terminal ZATCA status", gin.H{
		"pos_terminal_id": terminalID,
		"status":          "configured",
	}))
}
